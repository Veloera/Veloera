package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"veloera/common"
	"veloera/dto"
)

func ConvertGeminiCompatRequestToOpenAI(req *dto.GeminiCompatGenerateContentRequest, model string, stream bool) (*dto.GeneralOpenAIRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}

	messages := make([]dto.Message, 0, len(req.Contents)+1)

	if req.SystemInstruction != nil {
		if systemText := extractGeminiCompatText(req.SystemInstruction.Parts); systemText != "" {
			sysMsg := dto.Message{Role: "system"}
			sysMsg.SetStringContent(systemText)
			messages = append(messages, sysMsg)
		}
	}

	toolCallIndex := 0

	for _, content := range req.Contents {
		role := strings.ToLower(content.Role)
		switch role {
		case "model", "assistant":
			role = "assistant"
		case "user", "":
			if role == "" {
				role = "user"
			}
		case "tool":
			// leave as tool
		default:
			role = "user"
		}

		msg := dto.Message{Role: role}
		mediaContents := make([]dto.MediaContent, 0, len(content.Parts))
		toolCalls := make([]dto.ToolCallRequest, 0)
		toolMessages := make([]dto.Message, 0)

		for _, part := range content.Parts {
			switch {
			case part.FunctionCall != nil:
				argumentsBytes, err := json.Marshal(part.FunctionCall.Arguments)
				if err != nil {
					return nil, fmt.Errorf("marshal function call arguments failed: %w", err)
				}
				tc := dto.ToolCallRequest{
					ID:   fmt.Sprintf("call_%d", toolCallIndex),
					Type: "function",
					Function: dto.FunctionRequest{
						Name:      part.FunctionCall.Name,
						Arguments: string(argumentsBytes),
					},
				}
				toolCalls = append(toolCalls, tc)
				toolCallIndex++
			case part.FunctionResponse != nil:
				responseBytes, err := json.Marshal(part.FunctionResponse.Response)
				if err != nil {
					return nil, fmt.Errorf("marshal function response failed: %w", err)
				}
				toolMsg := dto.Message{Role: "tool"}
				toolMsg.Name = common.GetPointer(part.FunctionResponse.Name)
				toolMsg.SetStringContent(string(responseBytes))
				toolMessages = append(toolMessages, toolMsg)
			case part.InlineData != nil:
				if part.InlineData.Data == "" {
					continue
				}
				dataURI := fmt.Sprintf("data:%s;base64,%s", part.InlineData.MimeType, part.InlineData.Data)
				mediaContents = append(mediaContents, dto.MediaContent{
					Type: dto.ContentTypeImageURL,
					ImageUrl: &dto.MessageImageUrl{
						Url:    dataURI,
						Detail: "auto",
					},
				})
			case part.Text != "":
				mediaContents = append(mediaContents, dto.MediaContent{
					Type: dto.ContentTypeText,
					Text: part.Text,
				})
			}
		}

		if len(toolCalls) > 0 {
			msg.SetToolCalls(toolCalls)
		}

		appendMsg := true
		switch {
		case len(mediaContents) == 0:
			if len(toolCalls) == 0 {
				appendMsg = false
			}
		case len(mediaContents) == 1 && mediaContents[0].Type == dto.ContentTypeText:
			msg.SetStringContent(mediaContents[0].Text)
		default:
			msg.SetMediaContent(mediaContents)
		}

		if appendMsg {
			messages = append(messages, msg)
		}
		if len(toolMessages) > 0 {
			messages = append(messages, toolMessages...)
		}
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no valid messages in request")
	}

	openaiReq := &dto.GeneralOpenAIRequest{
		Model:    model,
		Stream:   stream,
		Messages: messages,
	}

	if req.GenerationConfig != nil {
		openaiReq.Stop = req.GenerationConfig.StopSequences
		if req.GenerationConfig.MaxOutputTokens > 0 {
			openaiReq.MaxTokens = uint(req.GenerationConfig.MaxOutputTokens)
		}
		if req.GenerationConfig.CandidateCount > 0 {
			openaiReq.N = req.GenerationConfig.CandidateCount
		}
		if req.GenerationConfig.ResponseMimeType != "" {
			if strings.EqualFold(req.GenerationConfig.ResponseMimeType, "application/json") {
				openaiReq.ResponseFormat = &dto.ResponseFormat{Type: "json_object"}
			}
		}
		if req.GenerationConfig.Temperature != nil {
			openaiReq.Temperature = req.GenerationConfig.Temperature
		}
		if req.GenerationConfig.TopP != nil {
			openaiReq.TopP = *req.GenerationConfig.TopP
		}
		if req.GenerationConfig.TopK != nil {
			openaiReq.TopK = *req.GenerationConfig.TopK
		}
	}

	if len(req.Tools) > 0 {
		openaiReq.Tools = convertGeminiTools(req.Tools)
	}

	if req.ToolConfig != nil && req.ToolConfig.FunctionCallingConfig != nil {
		switch strings.ToUpper(req.ToolConfig.FunctionCallingConfig.Mode) {
		case "NONE":
			openaiReq.ToolChoice = "none"
		case "AUTO":
			// leave nil to allow auto
		case "ANY":
			if len(req.ToolConfig.FunctionCallingConfig.AllowedFunctionNames) == 1 {
				openaiReq.ToolChoice = map[string]interface{}{
					"type": "function",
					"function": map[string]string{
						"name": req.ToolConfig.FunctionCallingConfig.AllowedFunctionNames[0],
					},
				}
			}
		}
	}

	return openaiReq, nil
}

func extractGeminiCompatText(parts []dto.GeminiCompatPart) string {
	var builder strings.Builder
	for _, part := range parts {
		if part.Text == "" {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(part.Text)
	}
	return builder.String()
}

func convertGeminiTools(tools []dto.GeminiCompatTool) []dto.ToolCallRequest {
	results := make([]dto.ToolCallRequest, 0)
	for _, tool := range tools {
		for _, fn := range tool.FunctionDeclarations {
			results = append(results, dto.ToolCallRequest{
				Type: "function",
				Function: dto.FunctionRequest{
					Name:        fn.Name,
					Description: fn.Description,
					Parameters:  fn.Parameters,
				},
			})
		}
	}
	return results
}

func HTTPStatusToGeminiStatus(code int) string {
	switch code {
	case 400:
		return "INVALID_ARGUMENT"
	case 401:
		return "UNAUTHENTICATED"
	case 403:
		return "PERMISSION_DENIED"
	case 404:
		return "NOT_FOUND"
	case 429:
		return "RESOURCE_EXHAUSTED"
	case 500:
		return "INTERNAL"
	case 503:
		return "UNAVAILABLE"
	default:
		return "UNKNOWN"
	}
}

func BuildGeminiErrorResponse(statusCode int, message string) dto.GeminiErrorResponse {
	status := HTTPStatusToGeminiStatus(statusCode)
	if message == "" {
		message = httpStatusText(statusCode)
	}
	return dto.GeminiErrorResponse{
		Error: dto.GeminiError{
			Code:    statusCode,
			Message: message,
			Status:  status,
		},
	}
}

func httpStatusText(code int) string {
	switch code {
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 429:
		return "Too Many Requests"
	case 500:
		return "Internal Server Error"
	case 503:
		return "Service Unavailable"
	default:
		return "Unexpected error"
	}
}

func OpenAIErrorToGeminiResponse(err *dto.OpenAIErrorWithStatusCode) dto.GeminiErrorResponse {
	if err == nil {
		return BuildGeminiErrorResponse(500, "Internal Server Error")
	}
	message := err.Error.Message
	if message == "" {
		message = httpStatusText(err.StatusCode)
	}
	return dto.GeminiErrorResponse{
		Error: dto.GeminiError{
			Code:    err.StatusCode,
			Message: message,
			Status:  HTTPStatusToGeminiStatus(err.StatusCode),
		},
	}
}
