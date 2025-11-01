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
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

type HistoricalEvent struct {
	Year  string `json:"year"`
	Month string `json:"month"`
	Day   string `json:"day"`
	Event string `json:"event"`
}

func getGeminiResponse(prompt string, history []*genai.Content) (*genai.GenerateContentResponse, *error) {
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
		return nil, &err
	}

	dailyFactTools := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        "get_historical_event",
				Description: "Get an historical events",
				Parameters: &genai.Schema{
					Type:       genai.TypeObject,
					Properties: map[string]*genai.Schema{},
					Required:   []string{},
				},
			},
		},
	}

	tools := []*genai.Tool{dailyFactTools}

	promptPart := genai.Part{Text: prompt}
	promptPartArray := []*genai.Part{&promptPart}

	content := &genai.Content{
		Parts: promptPartArray,
		Role:  "user",
	}
	contentArray := []*genai.Content{
		content,
	}

	config := &genai.GenerateContentConfig{
		Tools:             tools,
		SystemInstruction: genai.NewContentFromText("You are a helpful history assistant that provides accurate historical fact everyday.Your primary function is to give users historical facts(you can use emojis). When responding: \n- Always ask if the user wants more \n- Include relevant details like relevant people's name and date of birth if death occurs \n- Keep responses concise but informative \n- If the user asks for location and provides the current day country. \nUse the history tool to fetch historical facts", genai.RoleUser),
	}

	chat, _ := client.Chats.Create(ctx, "gemini-2.5-flash", config, history)

	response, err := chat.GenerateContent(ctx, "gemini-2.5-flash", contentArray, config)
	if err != nil {
		log.Fatal(err)
		return nil, &err
	}

	hasFunctionCall := false
	var functionCall *genai.FunctionCall

	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		for _, part := range response.Candidates[0].Content.Parts {
			if part.FunctionCall != nil {
				hasFunctionCall = true
				functionCall = part.FunctionCall
				break
			}
		}
	}

	if hasFunctionCall && functionCall != nil {
		if functionCall.Name == "get_historical_event" {
			result := getHistoricalEvents()

			if result == "" {
				result = "No historical events found for today"
			}

			functionResponsePart := genai.NewPartFromFunctionResponse(functionCall.Name, map[string]any{"result": result})
			functionPartArray := []*genai.Part{
				functionResponsePart,
			}
			contentArray = append(contentArray, response.Candidates[0].Content)
			contentArray = append(contentArray, &genai.Content{
				Role:  "user",
				Parts: functionPartArray,
			})

			finalResponse, err := chat.GenerateContent(ctx, "gemini-2.5-flash", contentArray, config)
			if err != nil {
				log.Fatal(err)
				return nil, &err
			}
			return finalResponse, nil
		}
	}
	return response, nil
}

func getHistoricalEvents() string {
	if err := godotenv.Load("./app.env"); err != nil {
		log.Fatalf("Error loading .env file : %v", err)
	}

	ninjasKey := os.Getenv("NINJAS_API_KEY")

	month, day := getRandomMonthAndDay()

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
	fmt.Println(response)
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
	}

	// Randomly pick 5 unique events
	rand.NewSource(time.Now().UnixNano())
	randomIndex := rand.Intn(len(events))

	selected := events[randomIndex]

	day, _ = strconv.Atoi(selected.Day)
	month, _ = strconv.Atoi(selected.Month)
	year, _ := strconv.Atoi(selected.Year)

	result := fmt.Sprintf("On %d/%d/%d(dd/mm/yy) in history: %s", day, month, year, selected.Event)

	return result
}

func getRandomMonthAndDay() (int, int) {
	rand.Seed(time.Now().UnixNano())

	// Generate random month (1-12)
	month := rand.Intn(12) + 1

	// Days in each month (non-leap year)
	daysInMonth := map[int]int{
		1: 31, 2: 28, 3: 31, 4: 30, 5: 31, 6: 30,
		7: 31, 8: 31, 9: 30, 10: 31, 11: 30, 12: 31,
	}

	// Generate random day based on the month
	day := rand.Intn(daysInMonth[month]) + 1

	return month, day
}
