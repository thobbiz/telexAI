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
	// Load the app.env
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

	// Make a tool that can use a local function to get historical events
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

	// Create an array of genai tool and add the dailyfactstool variable to it
	tools := []*genai.Tool{dailyFactTools}

	// Create a promptPart variable from prompt String
	promptPart := genai.Part{Text: prompt}

	// Create an array of genai part and add the promptPart variable to it
	promptPartArray := []*genai.Part{&promptPart}

	// Create a genai content variable and add the promptpart variable to it
	content := &genai.Content{
		Parts: promptPartArray,
		Role:  "user",
	}
	// Create an array of genai content and add the content variable to it
	contentArray := []*genai.Content{
		content,
	}

	// Create a config variabe that defines the behaviour of our AI
	config := &genai.GenerateContentConfig{
		Tools:             tools,
		SystemInstruction: genai.NewContentFromText("You are a helpful history assistant that provides accurate historical fact everyday.Your primary function is to give users historical facts(you can use emojis). When responding: \n- Always ask if the user wants more \n- Include relevant details like relevant people's name and date of birth if death occurs \n- Keep responses concise but informative \n- If the user asks for location and provides the current day country. \nUse the history tool to fetch historical facts", genai.RoleUser),
	}

	// Create a chat using the history parameter from the function definition
	chat, _ := client.Chats.Create(ctx, "gemini-2.5-flash", config, history)

	// Generate Response to the prompt
	response, err := chat.GenerateContent(ctx, "gemini-2.5-flash", contentArray, config)
	if err != nil {
		log.Fatal(err)
		return nil, &err
	}

	// variable that indicated if the response called the getHistory function
	hasFunctionCall := false
	// define a variable to get the details of the function called
	var functionCall *genai.FunctionCall

	// check if the response has candidates and checks if that first candidate has content parts
	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		// Loop through all parts in that candidate
		for _, part := range response.Candidates[0].Content.Parts {
			// If a part contains a function call
			if part.FunctionCall != nil {
				// If found save the function call to the functionCall variable we defined earlier
				// And we mark the hasFunction call as true
				hasFunctionCall = true
				functionCall = part.FunctionCall
				break
			}
		}
	}

	// If functionCall was found
	if hasFunctionCall && functionCall != nil {
		// If the functionCAll name was our get_Historical_event function
		if functionCall.Name == "get_historical_event" {
			// We call our function to get an Historical Event
			result := getHistoricalEvents()

			// If result was empty
			if result == "" {
				result = "No historical events found for today"
			}

			// use the result to Create a new functionResponse variable
			functionResponsePart := genai.NewPartFromFunctionResponse(functionCall.Name, map[string]any{"result": result})
			// Create an array of genai functionPart and add the functionResponsePart variable to it
			functionPartArray := []*genai.Part{
				functionResponsePart,
			}
			// Add the response candidate to the contentArray variable from earlier
			contentArray = append(contentArray, response.Candidates[0].Content)
			// Add the functionPartArray to the contentArray variable
			contentArray = append(contentArray, &genai.Content{
				Role:  "user",
				Parts: functionPartArray,
			})
			// Generate response to the HistoryEvent
			finalResponse, err := chat.GenerateContent(ctx, "gemini-2.5-flash", contentArray, config)
			if err != nil {
				log.Fatal(err)
				return nil, &err
			}
			// return the final response
			return finalResponse, nil
		}
	}
	// return the first response if function was never called
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
	rand.NewSource(time.Now().UnixNano())

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
