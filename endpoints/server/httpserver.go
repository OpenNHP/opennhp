package server

import (
	"context"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

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
					log.Error("https server close error: %v", err)
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
			log.Error("http server close error: %v", err)
			//panic(err)
		}
	}()

	hs.running.Store(true)
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
			content, err := os.ReadFile(absPath)
			if err != nil {
				return err
			}

			t := rootTmpl.New(path) // template name is relative path separated by slash on all platforms
			_, err = t.Parse(string(content))
			if err != nil {
				return err
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
			SrcIp:         ctx.ClientIP(),
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
			SrcIp:         ctx.ClientIP(),
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

	ackMsg := &common.ServerKnockAckMsg{
		AuthProviderToken: req.Token,
		AgentAddr:         srcIp,
		OpenTime:          res.OpenTime,
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
			artMsg, err := s.processACOperation(knkMsg, acConn, srcAddr, dstAddrs, openTime)
			artMsgsMutex.Lock()
			artMsgs[name] = artMsg
			if err == nil {
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
