package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"google.golang.org/genai"
)

func TaskHandler(ctx *gin.Context) {
	var req A2ARequest

	//Decode the response into a taskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Print(errorResponse(err))

		idStr := fmt.Sprintf("%v", req.Id)

		jsonError := JsonRPCError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
		errorResponse := A2AResponseError{
			JsonRPC: "2.0",
			Id:      idStr,
			Error:   jsonError,
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	if req.JsonRPC != "2.0" {
		jsonError := JsonRPCError{
			Code:    -32600,
			Message: "Invalid Request: jsonrpc must be \"2.0\" and id is required",
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
		jsonError := JsonRPCError{
			Code:    -32603,
			Message: "Invalid Request: missing required part",
		}
		errorResponse := A2AResponseError{
			JsonRPC: "2.0",
			Id:      req.Id,
			Error:   jsonError,
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// get the user's prompt
	prompt := messages.Parts[0].Text

	// convert them to gemini History
	dataParts := messages.Parts[1].Data
	var history []*genai.Content
	for i, dp := range dataParts {
		if i%2 == 1 {
			history = append(history, genai.NewContentFromText(dp.Text, genai.RoleUser))
		} else {
			history = append(history, genai.NewContentFromText(dp.Text, genai.RoleModel))
		}
	}

	result, err := getGeminiResponse(prompt, history)
	if err != nil || result == nil {
		jsonError := JsonRPCError{
			Code:    -32603,
			Message: "Internal Error",
		}
		errorResponse := A2AResponseError{
			JsonRPC: "2.0",
			Id:      req.Id,
			Error:   jsonError,
		}
		ctx.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	parts := Part{
		Kind: "text",
		Text: result.Text(),
	}

	partArray := []Part{parts}

	response := A2AResponseSuccess{
		JsonRPC: "2.0",
		Id:      req.Id,
		Result: Message{
			Id:    uuid.New().String(),
			Role:  "agent",
			Kind:  "message",
			Parts: partArray,
		},
	}

	fmt.Println(response)

	ctx.JSON(http.StatusOK, response)
}

func sendDailyFacts() {
	c := cron.New()

	// Run every day at 7:00 AM
	c.AddFunc("0 7 * * *", func() {
		// Move ALL the logic INSIDE the cron function
		result, err := getGeminiResponse("give me some history facts that happened today", nil)
		if err != nil {
			log.Printf("Error getting Gemini response: %v", err)
			return
		}

		rand.NewSource(time.Now().UnixNano())
		num := rand.Intn(20) + 1 //
		strNum := strconv.Itoa(num)

		resultKind := "text"
		resultString := result.Text()
		textPart := Part{
			Kind: resultKind,
			Text: resultString,
		}
		partList := []Part{textPart}

		data := A2AResponseSuccess{
			JsonRPC: "2.0",
			Id:      strNum,
			Result: Message{
				Id:    "msg --1",
				Role:  "agent",
				Parts: partList,
				Kind:  "message",
			},
		}

		if err := makePostRequest("https://api.example.com/daily", data); err != nil {
			log.Printf("Cron job error: %v", err)
		} else {
			log.Println("Daily facts sent successfully!")
		}
	})

	c.Start()
	log.Println("Daily facts scheduler started (runs at 9:00 AM)")
}

func makePostRequest(url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("POST to %s - Status: %s", url, resp.Status)
	return nil
}
