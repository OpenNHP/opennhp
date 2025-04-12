package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenNHP/opennhp/endpoints/server"
	"github.com/gin-gonic/gin"
)

func TestGlobInit(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	pwd, _ := os.Getwd()
	staticPath := filepath.Join(pwd, "../release/nhp-server/static")
	server.LoadFilesRecursively(g, staticPath)
	templatePath := filepath.Join(pwd, "../release/nhp-server/templates")
	server.LoadFilesRecursively(g, templatePath)
}
