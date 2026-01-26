module github.com/OpenNHP/opennhp/examples/server_plugin

go 1.23

require (
	github.com/OpenNHP/opennhp/nhp v0.0.0
	github.com/coreos/go-oidc/v3 v3.11.0
	github.com/gin-contrib/sessions v1.0.1
	github.com/gin-gonic/gin v1.10.0
	golang.org/x/oauth2 v0.23.0
)

replace github.com/OpenNHP/opennhp/nhp v0.0.0 => ../../../nhp
