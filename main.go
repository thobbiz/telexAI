package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	go sendDailyFacts()

	r := gin.Default()

	r.POST("/history_agent", TaskHandler)

	r.Static("/.well-known", "./.well-known")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Run(":8080")
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
