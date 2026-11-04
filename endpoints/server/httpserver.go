package server

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/OpenNHP/opennhp/nhp/audit"
	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/plugins"
	"github.com/OpenNHP/opennhp/nhp/version"
)

type HttpServer struct {
	id         string
	udpServer  *UdpServer
	httpServer *http.Server
	ginEngine  *gin.Engine
	listenAddr *net.TCPAddr
	config     *HttpConfig
	storage    *storageState

	// trustedProxies is the resolved HttpConfig.TrustedProxies list, kept
	// so knockSourceIP can tell "no proxy is configured" from "a proxy is"
	// without re-reading gin's engine state. Empty means the forwarding
	// headers are never consulted. Written once in Start, before the
	// listener goroutine starts, and read-only afterwards.
	trustedProxies []string

	wg      sync.WaitGroup
	running atomic.Bool

	// signals
	signals struct {
		stop chan struct{}
	}
}

// Note HttpServer must be started after starting UdpServer, when log and config have been setup
func (hs *HttpServer) Start(us *UdpServer, hc *HttpConfig) error {
	hs.id = time.Now().Format("2006-01-02 15:04:05")
	log.Info("==================================================")
	log.Info("===  HttpServer (%s) started  ===", hs.id)
	log.Info("==================================================")

	hs.udpServer = us
	hs.config = hc

	ipStr := hc.HttpListenIp
	var netIP net.IP
	if len(ipStr) > 0 {
		netIP = net.ParseIP(ipStr)
		if netIP == nil {
			log.Error("http listen ip address is incorrect! using udp listening ip")
			netIP = us.listenAddr.IP
		}
	} else {
		netIP = net.IPv4zero // will both listen on ipv4 0.0.0.0:port and ipv6 [::]:port
	}

	listenPort := us.listenAddr.Port // use the same port as udp server if HttpListenPort is not specified
	if hc.HttpListenPort > 0 && hc.HttpListenPort < 65536 {
		listenPort = hc.HttpListenPort
	}
	hs.listenAddr = &net.TCPAddr{
		IP:   netIP,
		Port: listenPort,
	}

	hs.signals.stop = make(chan struct{})

	gin.SetMode(gin.ReleaseMode)
	hs.ginEngine = gin.New()
	if err := hs.applyTrustedProxies(hc.TrustedProxies); err != nil {
		log.Error("refusing to start the HTTP knock listener: %v", err)
		return err
	}
	store := cookie.NewStore([]byte("nhpstore"))
	hs.ginEngine.Use(sessions.Sessions("nhpsessions", store))
	hs.ginEngine.Use(corsMiddleware())
	hs.ginEngine.Use(gin.LoggerWithWriter(us.log.Writer()))
	hs.ginEngine.Use(gin.Recovery())

	hs.initRouter()

	hs.httpServer = &http.Server{
		Addr:         hs.listenAddr.String(),
		Handler:      hs.ginEngine,
		ReadTimeout:  time.Duration(hc.ReadTimeoutMs) * time.Millisecond,
		WriteTimeout: time.Duration(hc.WriteTimeoutMs) * time.Millisecond,
		IdleTimeout:  time.Duration(hc.IdleTimeoutMs) * time.Millisecond,
	}

	// Mark the server running before either listener goroutine starts, not
	// after -- the TLS branch below returns from inside its own `if`, so a
	// store placed at the end of the function is reached only on the plain
	// HTTP path. Everything that manages a live listener keys off this flag:
	// Stop returns early without it, updateHttpConfig reads IsRunning to
	// decide whether to restart, and trustedProxiesPending uses it to decide
	// whether a proxy-list edit is still pending. Leaving it unset for a TLS
	// listener meant EnableHttp = false never stopped it, each reload bound a
	// second listener onto a port already in use, and the "TrustedProxies
	// changed" warning never fired -- for the configuration the shipped
	// http.toml files actually ship (EnableTLS = true).
	//
	// The flag means "this HttpServer owns a listener goroutine", which is
	// true from here on in all three paths (TLS, TLS with missing cert files
	// falling through, plain HTTP). The store also orders the trustedProxies
	// write above ahead of any cross-goroutine read of it.
	// wg.Add before the flag, not after: Stop observes running, then blocks on
	// wg.Wait(). Storing the flag first leaves a window where a Stop from the
	// etcd watcher could sail past a counter this goroutine has not yet
	// incremented. The order below makes that impossible.
	hs.wg.Add(1)
	hs.running.Store(true)

	if hc.EnableTLS {
		certFilePath := filepath.Join(ExeDirPath, hc.TLSCertFile)
		keyFilePath := filepath.Join(ExeDirPath, hc.TLSKeyFile)
		_, err1 := os.Stat(certFilePath)
		_, err2 := os.Stat(keyFilePath)
		if err1 == nil && err2 == nil {
			go func() {
				defer hs.wg.Done()
				log.Info("Listening https on %s", hs.listenAddr.String())
				var err = hs.httpServer.ListenAndServeTLS(certFilePath, keyFilePath)
				if err != nil && err != http.ErrServerClosed {
					log.Error("https server close error: %v", err)
					hs.listenerFailed(err)
				}
			}()

			return nil
		}
	}

	go func() {
		defer hs.wg.Done()
		log.Info("Listening http on %s", hs.listenAddr.String())
		var err = hs.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error("http server close error: %v", err)
			hs.listenerFailed(err)
		}
	}()

	return nil
}

// Stop stops the HttpServer by setting the running flag to false,
// closing the stop channel, shutting down the underlying http server,
// waiting for all goroutines to finish, and logging a message indicating
// that the HttpServer has been stopped.
func (hs *HttpServer) Stop() {
	if !hs.running.Load() {
		// already stopped, do nothing
		return
	}

	hs.running.Store(false)
	close(hs.signals.stop)
	ctx, cancel := context.WithTimeout(context.Background(), 5500*time.Millisecond)
	_ = hs.httpServer.Shutdown(ctx)

	hs.wg.Wait()
	cancel()
	cancel = nil
	log.Info("==================================================")
	log.Info("===  HttpServer (%s) stopped  ===", hs.id)
	log.Info("==================================================")
}

func (hs *HttpServer) IsRunning() bool {
	return hs.running.Load()
}

// listenerFailed clears the running flag after ListenAndServe{,TLS} returns
// something other than a clean shutdown -- a port already in use, an
// unreadable key, a fatal accept error.
//
// Without this, a bind that fails asynchronously leaves an HttpServer that
// claims to be running and serves nothing: updateHttpConfig sees IsRunning
// and never builds a replacement, so the daemon stays dark until someone
// cycles EnableHttp. Clearing it lets the next http.toml reload retry the
// bind, so freeing the port is enough to recover.
//
// Only reached on the error path. A Stop-initiated shutdown returns
// http.ErrServerClosed and must not land here, or Stop's own bookkeeping
// would race with this store.
func (hs *HttpServer) listenerFailed(err error) {
	hs.running.Store(false)
	log.Error("HttpServer (%s) is no longer listening; the next http config reload will retry: %v", hs.id, err)
}

// applyTrustedProxies pins which network origins nhp-serverd will believe
// X-Forwarded-For / X-Real-IP from, and must run before any handler does.
//
// gin.New() leaves Engine.trustedProxies at the package default of
// 0.0.0.0/0 + ::/0 with ForwardedByClientIP on, so ctx.ClientIP() returns a
// *caller-supplied* address for every request that carries a forwarding
// header. That value is not cosmetic here: the plugin routes put it in
// HttpKnockRequest.SrcIp and handleHttpOpenResource hands it to the AC as the
// source address of the firewall rule it opens, so trusting every peer lets
// any caller that can authenticate to a plugin open the resource for an
// address of their choosing and file the access against it (CWE-348).
//
// The default is therefore to trust nobody: with the list empty, ClientIP()
// degenerates to the TCP peer and the headers are inert. An operator who
// actually terminates TLS in a reverse proxy names that proxy in
// HttpConfig.TrustedProxies, and the proxy must *overwrite* the forwarding
// headers rather than append to them (nginx: `proxy_set_header
// X-Forwarded-For $remote_addr`, which deploy/nginx/server.conf.template
// does) — $proxy_add_x_forwarded_for would carry the client's own value
// through and hand the spoof back.
//
// A list that does not parse is a hard failure rather than a partial trust:
// gin's SetTrustedProxies installs the prefix it managed to parse *before*
// the bad entry and returns the error alongside it, so an engine left in that
// state would silently trust whichever proxies happened to sort first. We
// reset to "trust nobody" and refuse to bind at all, which the operator sees
// immediately instead of discovering it from a mis-attributed AC rule.
func (hs *HttpServer) applyTrustedProxies(proxies []string) error {
	trimmed := normalizeTrustedProxies(proxies)

	if len(trimmed) == 0 {
		hs.trustedProxies = nil
		// nil (not an empty slice) is gin's own "no proxy is trusted" value.
		if err := hs.ginEngine.SetTrustedProxies(nil); err != nil {
			return fmt.Errorf("failed to disable proxy header trust: %w", err)
		}
		log.Info("http knock: TrustedProxies is empty, forwarding headers are ignored and the TCP peer address is used as the knock source")
		return nil
	}

	if err := hs.ginEngine.SetTrustedProxies(trimmed); err != nil {
		// Undo the partially-installed prefix before reporting.
		if resetErr := hs.ginEngine.SetTrustedProxies(nil); resetErr != nil {
			log.Error("http knock: failed to reset proxy header trust after a bad TrustedProxies list: %v", resetErr)
		}
		hs.trustedProxies = nil
		return fmt.Errorf("invalid TrustedProxies entry in http.toml: %w", err)
	}

	hs.trustedProxies = trimmed
	log.Info("http knock: trusting forwarding headers from %d configured proxy origin(s)", len(trimmed))
	return nil
}

// normalizeTrustedProxies drops blank entries and surrounding whitespace so
// a configured list and a running list compare equal when they say the same
// thing. It does not validate: that is gin's job in SetTrustedProxies, which
// is the only place an entry is accepted or refused.
func normalizeTrustedProxies(proxies []string) []string {
	trimmed := make([]string, 0, len(proxies))
	for _, proxy := range proxies {
		proxy = strings.TrimSpace(proxy)
		if len(proxy) == 0 {
			continue
		}
		trimmed = append(trimmed, proxy)
	}
	return trimmed
}

// trustedProxyPrefixes resolves a list the way gin's prepareTrustedCIDRs does
// -- a bare address becomes a host prefix -- and returns the result as a set.
//
// A set of prefixes, rather than the strings, is what "the same trust" means:
// "127.0.0.1" and "127.0.0.1/32" admit exactly the same peer, and reordering
// a list changes nothing, because isTrustedProxy tests every entry. Comparing
// raw strings would announce a pending change for either edit.
//
// ok is false when an entry does not parse. Such a list is not equal to
// anything -- gin would refuse it outright -- so callers treat that as a
// difference rather than guessing.
func trustedProxyPrefixes(proxies []string) (map[netip.Prefix]struct{}, bool) {
	set := make(map[netip.Prefix]struct{}, len(proxies))
	for _, proxy := range normalizeTrustedProxies(proxies) {
		if strings.Contains(proxy, "/") {
			prefix, err := netip.ParsePrefix(proxy)
			if err != nil {
				return nil, false
			}
			set[prefix.Masked()] = struct{}{}
			continue
		}
		addr, err := netip.ParseAddr(proxy)
		if err != nil {
			return nil, false
		}
		set[netip.PrefixFrom(addr, addr.BitLen())] = struct{}{}
	}
	return set, true
}

// trustedProxiesPending reports whether a running listener is still enforcing
// a different set of proxy origins than the one now configured.
//
// updateHttpConfig does not restart a listener that is already running, so an
// edit to TrustedProxies alone changes s.httpConfig while the live engine
// keeps the list it started with. For most fields that lag is merely
// confusing; this one grants trust, so an operator who deletes a proxy to
// stop trusting it would be reading a config that says the proxy is gone from
// a process that still believes it. Callers log that rather than let it pass
// silently; applying it takes the documented EnableHttp false/true cycle.
//
// The comparison is on resolved prefixes so the warning only fires when the
// set of admitted peers actually differs. A line that cried wolf over
// "127.0.0.1" becoming "127.0.0.1/32" would be worth less when it fires for
// real, and it repeats on every reload for as long as the mismatch stands.
//
// Safe to call from another goroutine: trustedProxies is written by
// applyTrustedProxies before Start stores the running flag, and that atomic
// store orders the write ahead of the Load below.
func (hs *HttpServer) trustedProxiesPending(configured []string) bool {
	if hs == nil || !hs.running.Load() {
		return false
	}
	want, wantOK := trustedProxyPrefixes(configured)
	// The running list parsed once already, in applyTrustedProxies, so a
	// failure here would mean the two disagree about what gin accepts.
	running, runningOK := trustedProxyPrefixes(hs.trustedProxies)
	if !wantOK || !runningOK {
		return true
	}
	if len(want) != len(running) {
		return true
	}
	for prefix := range want {
		if _, found := running[prefix]; !found {
			return true
		}
	}
	return false
}

// knockSourceIP returns the address a knock is attributed to: the address the
// AC rule is opened for, the one the plugin authenticates, and the one the
// audit ledger records.
//
// With no trusted proxy configured this is the TCP peer, unconditionally —
// ctx.RemoteIP() rather than ctx.ClientIP(), so the answer does not depend on
// gin's engine still being in the state applyTrustedProxies left it in. Only
// once an operator has named a proxy does ClientIP()'s forwarding-header walk
// come into play, and then only for requests that actually arrive from one of
// those origins.
func (hs *HttpServer) knockSourceIP(ctx *gin.Context) string {
	if len(hs.trustedProxies) == 0 {
		return ctx.RemoteIP()
	}
	return ctx.ClientIP()
}

// LoadFilesRecursively loads HTML and template files recursively from the specified directory and adds them to the given gin.Engine.
// It walks through the directory and its subdirectories, and for each file with a .html or .tmpl extension, it reads the file content,
// creates a new template with the file path as the template name, and parses the content into the template.
// The loaded templates are set as the HTML templates for the gin.Engine.
// The directory path should be a clean absolute path.
// If any error occurs during the file loading or template parsing, the function returns the error.
func LoadFilesRecursively(g *gin.Engine, dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// dir does not exist
		return
	}

	cleanRootDir := filepath.Clean(dir)
	rootTmpl := template.New("").Funcs(g.FuncMap)
	f := os.DirFS(cleanRootDir)

	err = fs.WalkDir(f, ".", func(path string, info fs.DirEntry, walkErr error) error {
		// add *.html and *.tmpl files
		if !info.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".tmpl")) {
			if walkErr != nil {
				return walkErr
			}

			absPath := filepath.Join(cleanRootDir, path)
			content, fileErr := os.ReadFile(absPath)
			if fileErr != nil {
				return fileErr
			}

			t := rootTmpl.New(path) // template name is relative path separated by slash on all platforms
			_, fileErr = t.Parse(string(content))
			if fileErr != nil {
				return fileErr
			}
			log.Info("gin load file %s from %s", path, cleanRootDir)
			g.SetHTMLTemplate(t)
		}

		return nil
	})

	if err != nil {
		log.Error("load files to web engine failed. %v", err)
		return
	}
}

// init gin engine. Must be called at initialization
func (hs *HttpServer) initRouter() {
	g := hs.ginEngine

	// load templates. won't trigger panic if file does not exist
	staticPath := filepath.Join(ExeDirPath, "static")
	LoadFilesRecursively(g, staticPath)
	templatePath := filepath.Join(ExeDirPath, "templates")
	LoadFilesRecursively(g, templatePath)

	pluginGrp := g.Group("plugins")
	// display login page with templates
	pluginGrp.GET("/:aspid", func(ctx *gin.Context) {
		var err error
		aspId := ctx.Param("aspid")
		log.Info("get plugins request. aspId: %s, query: %v", aspId, ctx.Request.URL.RawQuery)

		if len(aspId) == 0 {
			err = common.ErrUrlPathInvalid
			log.Error("path error: %v", err)
			ctx.String(http.StatusOK, "{\"errMsg\": \"path error: %v\"}", err)
			return
		}

		req := &common.HttpKnockRequest{
			AuthServiceId: aspId,
			DeviceId:      ctx.Request.UserAgent(),
			SrcIp:         hs.knockSourceIP(ctx),
			Url:           ctx.Request.URL,
		}

		hs.authWithAspPlugin(ctx, req)
	})

	// legacy api
	pluginGrp.GET("/:aspid/:resid/valid", func(ctx *gin.Context) {
		// parse url parameters
		aspId := ctx.Param("aspid")
		resId := ctx.Param("resid")
		log.Info("get plugins request. aspId: %s, resId: %s, query: %v", aspId, resId, ctx.Request.URL.RawQuery)

		if len(aspId) == 0 {
			log.Error("no aspId provided")
			ctx.String(http.StatusOK, "{\"errMsg\": \"no aspId provided\"}")
			return
		}

		if len(resId) == 0 {
			log.Error("no resId provided")
			ctx.String(http.StatusOK, "{\"errMsg\": \"no resId provided\"}")
			return
		}

		req := &common.HttpKnockRequest{
			AuthServiceId: aspId,
			ResourceId:    resId,
			DeviceId:      ctx.Request.UserAgent(),
			SrcIp:         hs.knockSourceIP(ctx),
			Url:           ctx.Request.URL,
		}
		hs.legacyAuthWithAspPlugin(ctx, req)
	})

	hs.initStorageRouter()

	hs.initKbsRouter()

	/*
		refreshGrp := g.Group("refresh")
		refreshGrp.GET("/:token", func(ctx *gin.Context) {
			var err error
			token := ctx.Param("token")
			log.Info("get refresh request. aspId: %s, query: %v", token, ctx.Request.URL.RawQuery)

			if len(token) == 0 {
				err = common.ErrUrlPathInvalid
				log.Error("path error: %v", err)
				ctx.String(http.StatusOK, "{\"errMsg\": \"path error: %v\"}", err)
				return
			}

			// NOTE before re-enabling: a `srcip` query parameter is caller-
			// supplied in exactly the way X-Forwarded-For was (CWE-348), and it
			// ends up as the source of an AC rule. Take the address from
			// hs.knockSourceIP(ctx) and treat any value in the URL as a request
			// to be authorized against the token, never as the answer.
			req := &common.HttpRefreshRequest{
				Token: token,
				SrcIp: ctx.Query("srcip"),
			}

			hs.handleRefreshResource()
		})
	*/

}

// corsMiddleware is a middleware function that adds CORS headers to the HTTP response.
// It allows cross-origin resource sharing, specifies allowed methods, exposes headers, and sets maximum age.
// If the request method is OPTIONS, PUT, or DELETE, it aborts the request with a 204 status code.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// HTTP headers for CORS
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")                   // allow cross-origin resource sharing
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, POST") // methods
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Type, Content-Length, Set-Cookie")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, X-NHP-Ver, Cookie")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "300")
		// NHP headers
		c.Writer.Header().Set("Access-Control-NHP-Ver", version.Version+"/"+version.CommitId)

		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusOK)
			return
		}

		if c.Request.Method == "DELETE" || c.Request.Method == "PUT" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (hs *HttpServer) handleHttpOpenResource(req *common.HttpKnockRequest, res *common.ResourceData) (ack *common.ServerKnockAckMsg, err error) {
	defer hs.wg.Done()
	hs.wg.Add(1)
	s := hs.udpServer
	srcIp := req.SrcIp

	knkMsg := &common.AgentKnockMsg{
		UserId:         req.UserId,
		DeviceId:       req.DeviceId,
		OrganizationId: req.OrganizationId,
		AuthServiceId:  req.AuthServiceId,
		ResourceId:     res.ResourceId,
	}

	if req.Command == "exit" {
		knkMsg.HeaderType = core.NHP_EXT
	}

	// Declared before the audit defer below so the closure can read the
	// populated error code off it. The named return `ack` stays nil on every
	// naked `return` (resource-not-found, all-AC-failed), so reading `ack`
	// there would drop errCode on exactly the denials that carry it.
	ackMsg := &common.ServerKnockAckMsg{
		AuthProviderToken: req.Token,
		AgentAddr:         srcIp,
		OpenTime:          res.OpenTime,
	}

	// Record the HTTP access decision in the same ledger as the UDP knock
	// path — the browser / js-agent flow opens AC rules too, so leaving it
	// out would give the trail a blind spot exactly where the demo stack
	// operates. Deferred so every return path (resource-not-found, no AC
	// connection, all-AC-failed, success) is covered from the final err/ack.
	defer func() {
		if s == nil || s.auditLedger == nil {
			return
		}
		// Grant means a nil error AND the explicit success code, so neither a
		// soft denial nor an unset code reads as "granted". See decisionGranted.
		severity, result := audit.SeverityWarn, "denied"
		if decisionGranted(err, ackMsg == nil, ackMsg.ErrCode) {
			severity, result = audit.SeverityInfo, "granted"
		}
		op := "open"
		if knkMsg.HeaderType == core.NHP_EXT {
			op = "close"
		}
		fields := map[string]string{
			"user":   knkMsg.UserId,
			"device": knkMsg.DeviceId,
			"src":    srcIp,
			"aspId":  knkMsg.AuthServiceId,
			"resId":  knkMsg.ResourceId,
			"op":     op,
			"via":    "http",
			"result": result,
		}
		// ackMsg always carries the code set on the failure paths; the named
		// return ack only mirrors it on success (return ackMsg, nil).
		if ackMsg.ErrCode != "" {
			fields["errCode"] = ackMsg.ErrCode
		}
		if err != nil {
			fields["reason"] = err.Error()
		}
		s.auditEvent("knock", severity, fields)
	}()

	// The source address is what the AC turns into a firewall rule, so it has
	// to be an address. knockSourceIP returns "" when RemoteAddr does not split
	// into host/port, and a plugin is free to write req.SrcIp itself; either way
	// an unparsable value would reach the AC as the source of an ipset entry.
	// Refuse here so the denial is audited with a reason rather than handed on.
	//
	// net.ParseIP also rejects a zone-scoped address such as fe80::1%eth0,
	// which RemoteIP() can return for a link-local peer. That is deliberate:
	// the zone names an interface on *this* host and means nothing to the AC,
	// which is a different machine, so the rule it built from one would at
	// best not parse and at worst match something else. A link-local client
	// has no route to a remote AC's protected resources anyway.
	if net.ParseIP(srcIp) == nil {
		err = common.ErrInvalidIpAddress
		ackMsg.ErrCode = common.ErrInvalidIpAddress.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		log.Error("httpserver-agent(%s#%s)[handleHttpOpenResource] refusing to open a resource for an unparsable source address %q", knkMsg.UserId, knkMsg.DeviceId, srcIp)
		return
	}

	if len(res.Resources) == 0 {
		err = common.ErrResourceNotFound
		ackMsg.ErrCode = common.ErrResourceNotFound.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return
	}

	// PART II: determine knock src ip address and resource dst ip addresses
	srcAddr := &common.NetAddress{Ip: srcIp}

	acDstIpMap := make(map[string][]*common.NetAddress)
	for resName, info := range res.Resources {
		addrs, exist := acDstIpMap[resName]
		if exist {
			addrs = append(addrs, info.Addr)
			acDstIpMap[resName] = addrs
		} else {
			acDstIpMap[resName] = []*common.NetAddress{info.Addr}
		}
	}

	// PART III: request ac operation for each resource and block for response
	var acWg sync.WaitGroup
	var artMsgsMutex sync.Mutex
	artMsgs := make(map[string]*common.ACOpsResultMsg)
	ackMsg.ResourceHost = make(map[string]string)
	ackMsg.ACTokens = make(map[string]string)
	ackMsg.PreAccessActions = make(map[string]*common.PreAccessInfo)

	for resName, addrs := range acDstIpMap {
		resInfo := res.Resources[resName]
		if resInfo == nil {
			continue
		}
		acId := resInfo.ACId
		s.acConnectionMapMutex.Lock()
		acConn, found := s.acConnectionMap[acId]
		s.acConnectionMapMutex.Unlock()
		if !found {
			log.Warning("httpserver-agent(%s#%s@%s)-ac(@%s)[HandleHttpKnockRequest] no ac connection is available", knkMsg.UserId, knkMsg.DeviceId, srcIp, acId)
			artMsg := &common.ACOpsResultMsg{}
			err = common.ErrACConnectionNotFound
			artMsg.ErrCode = common.ErrACConnectionNotFound.ErrorCode()
			artMsg.ErrMsg = err.Error()
			artMsgsMutex.Lock()
			artMsgs[resName] = artMsg
			artMsgsMutex.Unlock()
			continue
		}

		acWg.Add(1)
		go func(name string, info *common.ResourceInfo, dstAddrs []*common.NetAddress) {
			defer acWg.Done()

			openTime := res.OpenTime
			if knkMsg.HeaderType == core.NHP_EXT {
				openTime = 1 // timeout in 1 second
			}
			artMsg, opErr := s.processACOperation(knkMsg, acConn, srcAddr, dstAddrs, openTime)
			artMsgsMutex.Lock()
			artMsgs[name] = artMsg
			if opErr == nil {
				ackMsg.ResourceHost[name] = info.DestHost()
				ackMsg.ACTokens[name] = artMsg.ACToken
				ackMsg.PreAccessActions[name] = artMsg.PreAccessAction
			}
			artMsgsMutex.Unlock()
		}(resName, resInfo, addrs)
	}
	acWg.Wait()

	var successCount int
	for _, artMsg := range artMsgs {
		if artMsg.ErrCode == common.ErrSuccess.ErrorCode() {
			successCount++
		}
	}

	if successCount == 0 {
		log.Info("httpserver-agent(%s#%s@%s)[handleHttpOpenResource] failed: %+v", knkMsg.UserId, knkMsg.DeviceId, srcIp, artMsgs)
		err = common.ErrServerACOpsFailed
		ackMsg.ErrCode = common.ErrServerACOpsFailed.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return
	}

	log.Info("httpserver-agent(%s#%s@%s)[handleHttpOpenResource] succeed", knkMsg.UserId, knkMsg.DeviceId, srcIp)
	ackMsg.ErrCode = common.ErrSuccess.ErrorCode()
	ackMsg.ErrMsg = common.ErrSuccess.Error()

	return ackMsg, nil
}

func (hs *HttpServer) NewHttpServerHelper() *plugins.HttpServerPluginHelper {
	h := &plugins.HttpServerPluginHelper{}
	h.StopSignal = hs.signals.stop

	h.AuthWithHttpCallbackFunc = func(req *common.HttpKnockRequest, res *common.ResourceData) (*common.ServerKnockAckMsg, error) {
		return hs.handleHttpOpenResource(req, res)
	}

	// Session helper functions for plugins
	h.SessionGet = func(ctx *gin.Context, key string) interface{} {
		session := sessions.Default(ctx)
		return session.Get(key)
	}
	h.SessionSet = func(ctx *gin.Context, key string, val interface{}) {
		session := sessions.Default(ctx)
		session.Set(key, val)
	}
	h.SessionSave = func(ctx *gin.Context) error {
		session := sessions.Default(ctx)
		return session.Save()
	}
	h.SessionClear = func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		session.Clear()
	}

	return h
}

// FindPluginHandler returns the plugin handler for the given ASP ID
// It delegates the task to the underlying UDP server's FindPluginHandler method.
func (hs *HttpServer) FindPluginHandler(aspId string) plugins.PluginHandler {
	return hs.udpServer.FindPluginHandler(aspId)
}

func (hs *HttpServer) handleRefreshResource(token string) (err error) {
	// to do
	return nil
}
