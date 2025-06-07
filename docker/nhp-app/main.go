package main

import "github.com/gin-gonic/gin"

func main() {
	//Create a default Gin router
	r := gin.Default()

	//Set the router to use the default middleware
	r.GET("/", func(c *gin.Context) {
		//Handle the request and respond with a JSON message
		c.JSON(200, gin.H{
			"message": "Hello World!",
		})
	})
	// Set the router to listen on port 8080
	r.Run()
}
