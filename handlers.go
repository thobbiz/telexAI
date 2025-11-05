package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/genai"
)

// requesthandler handles incoming A2A (Agent-to-Agent) JSON-RPC requests.
// It decodes the JSON payload, validates it, forwards the prompt to Gemini,
// and returns a structured JSON-RPC response.
func requestHandler(ctx *gin.Context) {
	var req A2ARequest

	// Decode the incoming JSON payload into an A2ARequest struct.
	// If decoding fails, return a JSON-RPC compliant error response.
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Print(errorResponse(err)) // Log the error for debugging

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

	// Validate that the JSON-RPC version is correct.
	// According to spec, it must be exactly "2.0".
	if req.JsonRPC != "2.0" {
		jsonError := JsonRPCError{
			Code:    -32600, // JSON-RPC invalid request error code
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

	// Extract the last 20 messages from the request.
	messages := req.Params.Message

	// The first message part usually contains the system prompt or user query.
	prompt := messages.Parts[0].Text

	var result *genai.GenerateContentResponse
	var err *error
	var history []*genai.Content

	// If the request contains conversation history, convert it into Gemini format.
	if len(messages.Parts) >= 2 {
		dataParts := messages.Parts[1].Data

		// Iterate over message data parts and alternate between user and model roles.
		for i, dp := range dataParts {
			if i%2 == 1 {
				// Odd index: user message
				history = append(history, genai.NewContentFromText(dp.Text, genai.RoleUser))
			} else {
				// Even index: model (AI) message
				history = append(history, genai.NewContentFromText(dp.Text, genai.RoleModel))
			}
		}

		// Send the prompt and history to Gemini to generate a new response.
		result, err = getGeminiResponse(prompt, history)
		if err != nil || result == nil {
			// If Gemini fails or returns nothing, send an internal error.
			jsonError := JsonRPCError{
				Code:    -32603, // JSON-RPC internal error
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

	} else {
		// No conversation history — just send the prompt directly.
		result, err = getGeminiResponse(prompt, nil)
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
	}

	// Build the response "Part" object — Gemini’s text result is wrapped in one part.
	parts := Part{
		Kind: "text",
		Text: result.Text(),
	}

	partArray := []Part{parts}

	// Construct the successful JSON-RPC response.
	// It includes the generated message from the AI agent.
	response := A2AResponseSuccess{
		JsonRPC: "2.0",
		Id:      req.Id,
		Result:
		// Result{
		// Id:        "task-001",
		// ContextId: "ctx--uuid",
		// Status: Status{
		// 	State:     "completed",
		// 	TimeStamp: time.Now().UTC().Format(time.RFC3339Nano),
		// Message:
		Message{
			Id:    uuid.New().String(), // Generate a unique message ID
			Role:  "agent",             // Mark this as an agent (AI) message
			Parts: partArray,           // Include the AI-generated content
			Kind:  "message",           // Specify the message kind
			// 	},
			// },
			// Artifacts: []Artifact{
			// 	Artifact{
			// 		ArtifactId: "artifact-uuid",
			// 		Name:       "historyData",
			// 		Parts: []ArtifactDataPart{
			// 			ArtifactDataPart{
			// 				Kind: "data",
			// 				Data: result,
			// 			},
			// 		},
			// 	},
			// },
			// History: history,
			// Kind:    "task",
		},
	}

	// Send the successful response with HTTP 200 OK.
	ctx.JSON(http.StatusOK, response)
}
