package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	sendDailyFacts()

	r := gin.Default()
	r.Static("/.well-known", "./.well-known")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Run(":8080")

	result, err := getGeminiResponse("give me some history facts that happened today", nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.Text())
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
