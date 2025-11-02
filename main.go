package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()

	r.POST("/history_agent", requestHandler)
	r.Static("/.well-known", "./.well-known")

	r.Run(":" + port)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
