package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	//Create a default Gin router
	r := gin.Default()

	// APP_NAME lets a single image back more than one protected resource
	// (e.g. the multi-cluster demo runs one instance per cluster). Defaults
	// to "Hello World!" so the single-cluster demo is unchanged.
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "Hello World!"
	}

	//Set the router to use the default middleware
	r.GET("/", func(c *gin.Context) {
		//Handle the request and respond with a JSON message
		c.JSON(200, gin.H{
			"message": appName,
		})
	})
	// Set the router to listen on port 8080
	r.Run()
}
