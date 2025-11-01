package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

type HistoricalEvent struct {
	Year  string `json:"year"`
	Event string `json:"event"`
}

func getGeminiResponse(prompt string, history []*genai.Content) *genai.GenerateContentResponse {
	if err := godotenv.Load("./app.env"); err != nil {
		log.Fatalf("Error loading .env file : %v", err)
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")

	// Set the key so genai can use it
	os.Setenv("GEMINI_API_KEY", geminiKey)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	// // Get today’s date
	// today := time.Now()
	// month := int(today.Month())
	// day := today.Day()

	// // Randomly pick 5 unique events
	// rand.NewSource(time.Now().UnixNano())
	// rand.Shuffle(len(events), func(i, j int) { events[i], events[j] = events[j], events[i] })

	// count := 5
	// if len(events) < 5 {
	// 	count = len(events)
	// }
	// selected := events[:count]

	// prompt := "Summarize these historical events that happened on this "
	// prompt += fmt.Sprintf("day %d/%d/2025(dd/mm/yy) engagingly and briefly:\n", day, month)
	// for _, e := range selected {
	// 	prompt += fmt.Sprintf("- In the year %s: %s\n", e.Year, e.Event)
	// }

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("You are a \"know it all\" enthusiatic history teacher. Your name is Anthonei. You give historical facts whic maybe random sometimes. You don't do any other thing", genai.RoleUser),
	}

	chat, _ := client.Chats.Create(ctx, "gemini-2.5-flash", config, history)
	result, err := chat.SendMessage(ctx, genai.Part{Text: prompt})
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return result
}

func getGeminiSummary(events []HistoricalEvent) *genai.GenerateContentResponse {
	if err := godotenv.Load("./app.env"); err != nil {
		log.Fatalf("Error loading .env file : %v", err)
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")

	// Set the key so genai can use it
	os.Setenv("GEMINI_API_KEY", geminiKey)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	// Get today’s date
	today := time.Now()
	month := int(today.Month())
	day := today.Day()

	// Randomly pick 5 unique events
	rand.NewSource(time.Now().UnixNano())
	rand.Shuffle(len(events), func(i, j int) { events[i], events[j] = events[j], events[i] })

	count := 5
	if len(events) < 5 {
		count = len(events)
	}
	selected := events[:count]

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText("You are a history teacher. Your name is Anthonei. You give daily historical facts", genai.RoleUser),
	}

	prompt := "Summarize these historical events that happened on this "
	prompt += fmt.Sprintf("day %d/%d/2025(dd/mm/yy) casualy and very briefly. Make the facts properly spaced and paginated and use emojis sometimes:\n", day, month)
	for _, e := range selected {
		prompt += fmt.Sprintf("- In the year %s: %s\n", e.Year, e.Event)
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		config,
	)
	if err != nil {
		log.Fatalf("An error occured while generating the response: %v", err)
	}

	return result
}

func getHistoricalEvents() []HistoricalEvent {
	if err := godotenv.Load("./app.env"); err != nil {
		log.Fatalf("Error loading .env file : %v", err)
	}

	ninjasKey := os.Getenv("NINJAS_API_KEY")

	// Get today’s date
	today := time.Now()
	month := int(today.Month())
	day := today.Day()

	//	Format the url
	url := fmt.Sprintf("https://api.api-ninjas.com/v1/historicalevents?month=%d&day=%d", month, day)

	//Create a new request
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("X-Api-Key", ninjasKey)

	// Send the request
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("An error occured %v", err)
	}
	defer response.Body.Close()

	// Get the request body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("An error occured while decoding the response: %v", err)
	}

	// Decode the body into Historical Events List
	var events []HistoricalEvent
	if err := json.Unmarshal(body, &events); err != nil {
		log.Fatal(err)
	}

	// If list is empty
	if len(events) == 0 {
		fmt.Println("No events found for this date.")
		return nil
	}

	return events
}

func printGeminiResponse(resp *genai.GenerateContentResponse) {
	if resp == nil {
		fmt.Println("No response received.")
		return
	}
	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			fmt.Println(part)
		}
	}
}
