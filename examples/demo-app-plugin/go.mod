module github.com/OpenNHP/opennhp/examples/demo-app-plugin

go 1.26

require (
	github.com/OpenNHP/opennhp/nhp v0.6.0
	github.com/gin-gonic/gin v1.12.0
	github.com/pelletier/go-toml/v2 v2.4.3
)

replace github.com/OpenNHP/opennhp/nhp v0.6.0 => ../../nhp
