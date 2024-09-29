package ac

import (
	"context"
	"encoding/base64"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/log"
	"github.com/gin-gonic/gin"
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

	hs.running.Store(true)
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
	hs.httpServer.Shutdown(ctx)

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

	pluginGrp := g.Group("refresh")
	// display login page with templates
	pluginGrp.GET("/:token", func(ctx *gin.Context) {
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

		ha.HandleHttpRefreshOperations(ctx, req)
	})
}

func (ha *HttpAC) HandleHttpRefreshOperations(c *gin.Context, req *common.HttpRefreshRequest) {
	if len(req.SrcIp) == 0 {
		c.String(http.StatusOK, "{\"errMsg\": \"empty source ip\"}")
		return
	}

	netIp := net.ParseIP(req.SrcIp)
	if netIp == nil {
		c.String(http.StatusOK, "{\"errMsg\": \"invalid source ip\"}")
		return
	}

	buf, err := base64.StdEncoding.DecodeString(req.Token)
	if err != nil || len(buf) != 32 {
		c.String(http.StatusOK, "{\"errMsg\": \"invalid token\"}")
		return
	}

	entry := ha.ua.VerifyAccessToken(req.Token)
	if entry == nil {
		c.String(http.StatusOK, "{\"errMsg\": \"token verification failed\"}")
		return
	}

	var found bool
	var newSrcAddr *common.NetAddress
	for _, addr := range entry.SrcAddrs {
		if addr.Ip == req.SrcIp {
			found = true
			break
		}
	}
	if !found {
		newSrcAddr = &common.NetAddress{
			Ip:       req.SrcIp,
			Port:     entry.SrcAddrs[0].Port,
			Protocol: entry.SrcAddrs[0].Protocol,
		}
		entry.SrcAddrs = append(entry.SrcAddrs, newSrcAddr)
	}

	_, err = ha.ua.HandleAccessControl(entry.AgentUser, entry.SrcAddrs, entry.DstAddrs, entry.OpenTime, nil)
	if err != nil {
		c.String(http.StatusOK, "{\"errMsg\": \"%s\"}", err)
		return
	}

	c.JSON(http.StatusOK, entry)
}
