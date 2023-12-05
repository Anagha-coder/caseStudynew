package main

import (
	"net/http"

	handlers "casestudy/handlers"

	"github.com/gin-gonic/gin"
)

func main() {

	// Create a new Gin router
	r := gin.Default()

	// Enable CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5500")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Define the endpoint for creating a user
	r.POST("/api/create-user", handlers.CreateUser)

	r.POST("api/login", handlers.Login)

	r.POST("api/upload", handlers.UploadFile)

	// Run the Gin router
	r.Run(":8087")
}
