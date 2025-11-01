package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"google.golang.org/genai"
)

func TaskHandler(ctx *gin.Context) {
	var req A2ARequest

	//Decode the response into a taskRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		log.Print(errorResponse(err))

		jsonError := JsonRPCError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
		errorResponse := A2AResponseError{
			JsonRPC: "2.0",
			Id:      req.Id,
			Error:   jsonError,
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Get the last 20 messages
	messages := req.Params.Message
	if len(messages.Parts) < 2 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing required parts"})
		return
	}

	prompt := messages.Parts[0].Text

	dataParts := messages.Parts[1].Data
	var history []*genai.Content
	for i, dp := range dataParts {
		if i%2 == 1 {
			history = append(history, genai.NewContentFromText(dp.Text, genai.RoleUser))
		} else {
			history = append(history, genai.NewContentFromText(dp.Text, genai.RoleModel))
		}
	}

	getGeminiResponse(prompt, history)
}

func sendDailyFacts() (*string, error) {
	events := getHistoricalEvents()
	result := getGeminiSummary(events)

	resultKind := "text"
	resultString := result.Text()

	textPart := Part{
		Kind: resultKind,
		Text: resultString,
	}

	partList := []Part{textPart}

	data := A2AResponseSuccess{
		JsonRPC: "2.0",
		Id:      1,
		Result: Message{
			Id:    "msg --1",
			Role:  "agent",
			Parts: partList,
			Kind:  "message",
		},
	}

	body, _ := json.Marshal(data)
	url := "http://localhost:8080/api/daily-task"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Println("❌ Error creating request:", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("API_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("❌ Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	return &resultString, nil
}

func retrieveFact() {
	fact, err := sendDailyFacts()
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	fmt.Print(fact)
}

func startScheduler() {
	c := cron.New()
	// "@daily" means once every day at midnight (00:00)
	// You can change to "0 9 * * *" for 9 AM daily
	_, err := c.AddFunc("0 7 * * *", retrieveFact)
	if err != nil {
		log.Fatal("Error scheduling job:", err)
	}
	c.Start()
	fmt.Println("🕓 Daily job scheduled successfully.")
}
