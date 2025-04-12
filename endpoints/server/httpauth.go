package server

import (
	"net/http"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/gin-gonic/gin"
)

func (hs *HttpServer) authWithAspPlugin(c *gin.Context, req *common.HttpKnockRequest) {
	var err error
	aspId := req.AuthServiceId

	handler := hs.FindPluginHandler(aspId)
	if handler == nil {
		err = common.ErrAuthHandlerNotFound
		log.Error("no auth handler provided")
		c.String(http.StatusOK, "{\"errMsg\": \"no auth handler provided\"}")
		return
	}

	helper := hs.NewHttpServerHelper()
	ackMsg, err := handler.AuthWithHttp(c, req, helper)
	_ = ackMsg
	if err != nil {
		log.Info("auth error: %v", err)
		if !c.Writer.Written() {
			c.String(http.StatusOK, "{\"errMsg\": \"auth error: %v\"}", err)
		}
	} else {
		log.Info("auth completed successfully")
	}
}

func (hs *HttpServer) legacyAuthWithAspPlugin(c *gin.Context, req *common.HttpKnockRequest) {
	handler := hs.FindPluginHandler(req.AuthServiceId)
	if handler == nil {
		log.Error("no auth handler provided")
		c.String(http.StatusOK, "{\"errMsg\": \"no auth handler provided\"}")
		return
	}

	helper := hs.NewHttpServerHelper()
	ackMsg, err := handler.AuthWithHttp(c, req, helper)
	_ = ackMsg
	if err != nil {
		log.Info("auth error: %v", err)
		if !c.Writer.Written() {
			c.String(http.StatusOK, "{\"errMsg\": \"auth error: %v\"}", err)
		}
	} else {
		log.Info("auth completed successfully")
	}
}
