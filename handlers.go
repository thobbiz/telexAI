package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	prompt := messages.Parts[0].Text

	var result *genai.GenerateContentResponse
	var err *error

	// convert them to gemini History
	if len(messages.Parts) >= 2 {
		dataParts := messages.Parts[1].Data
		var history []*genai.Content
		for i, dp := range dataParts {
			if i%2 == 1 {
				history = append(history, genai.NewContentFromText(dp.Text, genai.RoleUser))
			} else {
				history = append(history, genai.NewContentFromText(dp.Text, genai.RoleModel))
			}
		}

		result, err = getGeminiResponse(prompt, history)
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
	} else {
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
			Parts: partArray,
			Kind:  "message",
		},
	}

	ctx.JSON(http.StatusOK, response)
}
