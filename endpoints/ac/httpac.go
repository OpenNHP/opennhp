package ac

import (
	"context"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/log"
)

type HttpAC struct {
	id         string
	ua         *UdpAC
	httpServer *http.Server
	ginEngine  *gin.Engine
	listenAddr *net.TCPAddr

	wg      sync.WaitGroup
	running atomic.Bool

	// signals
	signals struct {
		stop chan struct{}
	}
}

// Note HttpServer must be started after starting UdpAC, when log and config have been setup
func (hs *HttpAC) Start(uac *UdpAC, hc *HttpConfig) error {
	hs.id = time.Now().Format("2006-01-02 15:04:05")
	log.Info("==================================================")
	log.Info("===  HttpServer (%s) started  ===", hs.id)
	log.Info("==================================================")

	hs.ua = uac

	port := hc.HttpListenPort
	if hc.HttpListenPort == 0 {
		port = 62206
	}
	// only listen to localhost for security reason.
	hs.listenAddr = &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: port,
	}

	hs.signals.stop = make(chan struct{})

	gin.SetMode(gin.ReleaseMode)
	hs.ginEngine = gin.New()
	hs.ginEngine.Use(gin.LoggerWithWriter(uac.log.Writer()))
	hs.ginEngine.Use(gin.Recovery())

	hs.initRouter()

	hs.httpServer = &http.Server{
		Addr:         hs.listenAddr.String(),
		Handler:      hs.ginEngine,
		ReadTimeout:  4500 * time.Millisecond,
		WriteTimeout: 4000 * time.Millisecond,
		IdleTimeout:  5000 * time.Millisecond,
	}

	// Set before either listener starts: the HTTPS branch returns early, and
	// Stop does nothing while running is false.
	hs.running.Store(true)
	hs.wg.Add(1)
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
					log.Error("https server close error: %v\n", err)
					//panic(err)
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
			log.Error("http server close error: %v\n", err)
			//panic(err)
		}
	}()

	return nil
}

// Stop stops the HttpServer by setting the running flag to false,
// closing the stop channel, shutting down the underlying http server,
// waiting for all goroutines to finish, and logging a message indicating
// that the HttpServer has been stopped.
func (hs *HttpAC) Stop() {
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

func (hs *HttpAC) IsRunning() bool {
	return hs.running.Load()
}

// init gin engine. Must be called at initialization
func (ha *HttpAC) initRouter() {
	g := ha.ginEngine

	refreshGrp := g.Group("refresh")
	refreshGrp.GET("/:token", func(ctx *gin.Context) {
		var err error
		token := ctx.Param("token")

		if len(token) == 0 {
			err = common.ErrUrlPathInvalid
			log.Error("path error: %v", err)
			ctx.String(http.StatusOK, "{\"errMsg\": \"path error: %v\"}", err)
			return
		}

		if token, err = url.QueryUnescape(token); err != nil {
			err = common.ErrUrlPathInvalid
			log.Error("token error: %v", err)
			ctx.String(http.StatusOK, "{\"errMsg\": \"token error: %v\"}", err)
			return
		}

		// Access tokens are bearer credentials. Log only a short correlation
		// prefix after URL decoding so percent escapes do not dilute the prefix.
		log.Info("get refresh request. token: %s, query: %v", common.RedactToken(token), ctx.Request.URL.RawQuery)

		// srcip stays a parameter, deliberately.
		//
		// This listener binds 127.0.0.1 only (see Start), so every caller
		// is a local process — a reverse proxy or a portal on the AC host
		// — and the connection's own address is always loopback, while an
		// access entry holds the agent's real address from the knock.
		// Deriving the source from the connection here would refuse every
		// legitimate refresh. The parameter is how the local front end
		// passes the real client address along.
		//
		// Accepting it is safe because of what HandleHttpRefreshOperations
		// does with it: a refresh can only re-apply a grant the entry
		// already covers, and the entry's addresses were authenticated by
		// the NHP knock. Naming an address is no longer a way to acquire
		// one. That was the defect; see GHSA-7g2g-wvgm-35qx.
		//
		// What the parameter still decides is which of an entry's
		// already-authorized addresses gets re-applied. Any process on
		// this host -- or a front end that copies a client-controlled
		// header such as X-Forwarded-For into srcip -- can choose that,
		// and can keep a grant alive for an address it does not hold.
		// Narrowing it further would mean authenticating the local front
		// end, which this endpoint does not do.
		req := &common.HttpRefreshRequest{
			Token: token,
			SrcIp: ctx.Query("srcip"),
		}

		ha.HandleHttpRefreshOperations(ctx, req)
	})
}

func (ha *HttpAC) HandleHttpRefreshOperations(c *gin.Context, req *common.HttpRefreshRequest) {
	if len(req.SrcIp) == 0 {
		log.Error("empty source ip")
		c.String(http.StatusOK, "{\"errMsg\": \"empty source ip\"}")
		return
	}

	netIp := net.ParseIP(req.SrcIp)
	if netIp == nil {
		log.Error("invalid source ip")
		c.String(http.StatusOK, "{\"errMsg\": \"invalid source ip\"}")
		return
	}

	buf, err := base64.StdEncoding.DecodeString(req.Token)
	if err != nil || len(buf) != 32 {
		log.Error("invalid token format")
		c.String(http.StatusOK, "{\"errMsg\": \"invalid token format\"}")
		return
	}

	// Looked up without renewing: this handler can still refuse below,
	// and a refusal that had already reset the expiry would keep the
	// grant alive for a request it rejected. The token is renewed on the
	// path that re-applies the grant, at the end.
	entry := ha.ua.LookupAccessToken(req.Token)
	if entry == nil {
		log.Error("token verification failed")
		c.String(http.StatusOK, "{\"errMsg\": \"token verification failed\"}")
		return
	}

	// Refresh extends an existing grant; it does not create one.
	//
	// This used to add the address to the entry when it was not already
	// there, and then program the filter for it. Combined with the source
	// address coming from a query parameter, one access token opened the
	// protected destination to any address its holder cared to name.
	// Enrolment belongs to the knock, which the NHP protocol
	// authenticates; a refresh that arrives from somewhere the entry does
	// not cover is refused, and the client knocks again.
	//
	// Compared as addresses, not as text. The two spellings come from
	// different places -- the entry's from the knock, the request's from
	// whoever calls this listener -- so the same address can be written
	// differently on each side: ::ffff:10.0.0.5 against 10.0.0.5, or a
	// zero-compressed IPv6 address against an expanded one. While a
	// mismatch quietly enrolled the address that cost nothing; now that
	// it refuses, a spelling difference would turn away a client the
	// entry does cover.
	var found bool
	for _, addr := range entry.SrcAddrs {
		if entryIp := net.ParseIP(addr.Ip); entryIp != nil && entryIp.Equal(netIp) {
			found = true
			break
		}
	}
	if !found {
		log.Error("refresh from %s is not covered by the token's access entry; refusing (knock again to authorize this address)", req.SrcIp)
		c.String(http.StatusOK, "{\"errMsg\": \"source address not authorized for this token\"}")
		return
	}

	_, err = ha.ua.HandleAccessControl(entry.User, entry.SrcAddrs, entry.DstAddrs, entry.OpenTime, nil)
	if err != nil {
		log.Error("HandleAccessControl failed: %v", err)
		c.String(http.StatusOK, "{\"errMsg\": \"%s\"}", err)
		return
	}

	// The grant has been re-applied, so the token gets a full window.
	ha.ua.RenewAccessToken(entry)

	c.JSON(http.StatusOK, entry)
}
