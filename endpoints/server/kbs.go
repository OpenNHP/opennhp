package server

import (
	kbsAttest "github.com/OpenNHP/opennhp/endpoints/server/kbs/attest"
	kbsAuth "github.com/OpenNHP/opennhp/endpoints/server/kbs/auth"
	kbsResource "github.com/OpenNHP/opennhp/endpoints/server/kbs/resource"
)

func (hs *HttpServer) initKbsRouter() {
	g := hs.ginEngine.Group("/kbs/v0")

	g.POST("/auth", kbsAuth.Auth)
	g.POST("/attest", kbsAttest.Attest)
	g.GET("/resource/*path", kbsResource.GetResource)
}
