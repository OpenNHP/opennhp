package server

import (
	"net/http"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/gin-gonic/gin"
)

// doAuthWithPlugin is the common implementation for plugin-based authentication.
// Both authWithAspPlugin and legacyAuthWithAspPlugin delegate to this function.
func (hs *HttpServer) doAuthWithPlugin(c *gin.Context, req *common.HttpKnockRequest) {
	handler := hs.FindPluginHandler(req.AuthServiceId)
	if handler == nil {
		log.Error("no auth handler provided for aspId: %s", req.AuthServiceId)
		c.String(http.StatusOK, "{\"errMsg\": \"no auth handler provided\"}")
		return
	}

	helper := hs.NewHttpServerHelper()
	_, err := handler.AuthWithHttp(c, req, helper)
	if err != nil {
		log.Info("auth error: %v", err)
		if !c.Writer.Written() {
			c.String(http.StatusOK, "{\"errMsg\": \"auth error: %v\"}", err)
		}
	} else {
		log.Info("auth completed successfully")
	}
}

// authWithAspPlugin handles authentication using the ASP plugin system.
func (hs *HttpServer) authWithAspPlugin(c *gin.Context, req *common.HttpKnockRequest) {
	hs.doAuthWithPlugin(c, req)
}

// legacyAuthWithAspPlugin is the legacy authentication handler.
// Kept for backward compatibility; delegates to the common implementation.
func (hs *HttpServer) legacyAuthWithAspPlugin(c *gin.Context, req *common.HttpKnockRequest) {
	hs.doAuthWithPlugin(c, req)
}
