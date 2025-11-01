package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// prompt := "sup"
	// res := getGeminiSummary(prompt, nil, nil)
	// if len(res.Candidates) > 0 {
	// 	fmt.Println(res.Candidates[0].Content.Parts[0].Text)
	// }
	startScheduler()

	r := gin.Default()
	r.Static("/.well-known", "./.well-known")

	r.GET("/send-now", func(ctx *gin.Context) {
		fact, err := sendDailyFacts()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"status": "sent",
			"fact":   fact,
		})
	})

	// Health check route
	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Gin app running with daily job"})
	})

	r.Run(":8080")
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
