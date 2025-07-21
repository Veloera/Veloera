// Copyright (c) 2025 Tethys Plex
//
// This file is part of Veloera.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package gemini

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"veloera/common"
	"veloera/constant"
	"veloera/dto"
	openaichannel "veloera/relay/channel/openai"
	relaycommon "veloera/relay/common"
	"veloera/relay/helper"
	"veloera/service"
	"veloera/setting/model_setting"
)

func extractGeminiInputContent(info *relaycommon.RelayInfo) {
	if info.Other == nil {
		info.Other = make(map[string]interface{})
	}

	if messages, ok := info.PromptMessages.([]interface{}); ok && len(messages) > 0 {
		var systemPrompt string
		var contextMessages []interface{}
		var userMessage interface{}
		var lastUserMessageIndex = -1

		for i := len(messages) - 1; i >= 0; i-- {
			if msgMap, ok := messages[i].(map[string]interface{}); ok {
				if role, exists := msgMap["role"]; exists && role == "user" {
					lastUserMessageIndex = i
					break
				}
			}
		}

		for i, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				if role, exists := msgMap["role"]; exists {
					if role == "system" {
						if content, hasContent := msgMap["content"]; hasContent && content != nil {
							if systemPrompt == "" {
								systemPrompt = fmt.Sprintf("%v", content)
							} else {
								systemPrompt += "\n" + fmt.Sprintf("%v", content)
							}
						}
					} else if i == lastUserMessageIndex {
						userMessage = msgMap
					} else {
						contextMessages = append(contextMessages, msgMap)
					}
				}
			}
		}

		if systemPrompt != "" {
			info.Other["system_prompt"] = systemPrompt
		}
		info.Other["context"] = contextMessages
		if userMessage != nil {
			info.Other["input_content"] = userMessage
		} else {
			if len(messages) > 0 {
				info.Other["input_content"] = messages[len(messages)-1]
			}
		}
	} else {
		info.Other["input_content"] = info.PromptMessages
	}
}

func extractGeminiOutputContent(info *relaycommon.RelayInfo, response *GeminiChatResponse) {
	if info.Other == nil {
		info.Other = make(map[string]interface{})
	}

	if response == nil || len(response.Candidates) == 0 {
		return
	}

	var outputTexts []string
	var thinkingTexts []string
	var functionCalls []interface{}
	var safetyRatings []interface{}
	var codeExecutions []interface{}
	var multimodalSummary = map[string]interface{}{
		"image_count": 0,
		"audio_count": 0,
		"file_count":  0,
	}

	for _, candidate := range response.Candidates {
		if len(candidate.SafetyRatings) > 0 {
			for _, rating := range candidate.SafetyRatings {
				safetyRatings = append(safetyRatings, map[string]interface{}{
					"category":    rating.Category,
					"probability": rating.Probability,
				})
			}
		}

		for _, part := range candidate.Content.Parts {
			if part.FunctionCall != nil {
				functionCalls = append(functionCalls, map[string]interface{}{
					"name":      part.FunctionCall.FunctionName,
					"arguments": part.FunctionCall.Arguments,
				})
			} else if part.FunctionResponse != nil {
				functionCalls = append(functionCalls, map[string]interface{}{
					"type":     "response",
					"name":     part.FunctionResponse.Name,
					"response": part.FunctionResponse.Response,
				})
			} else if part.Thought {
				thinkingTexts = append(thinkingTexts, part.Text)
			} else if part.ExecutableCode != nil {
				codeExecutions = append(codeExecutions, map[string]interface{}{
					"type":     "code",
					"language": part.ExecutableCode.Language,
					"code":     part.ExecutableCode.Code,
				})
			} else if part.CodeExecutionResult != nil {
				codeExecutions = append(codeExecutions, map[string]interface{}{
					"type":    "result",
					"outcome": part.CodeExecutionResult.Outcome,
					"output":  part.CodeExecutionResult.Output,
				})
			} else if part.InlineData != nil {
				if strings.HasPrefix(part.InlineData.MimeType, "image") {
					multimodalSummary["image_count"] = multimodalSummary["image_count"].(int) + 1
				} else if strings.HasPrefix(part.InlineData.MimeType, "audio") {
					multimodalSummary["audio_count"] = multimodalSummary["audio_count"].(int) + 1
				} else {
					multimodalSummary["file_count"] = multimodalSummary["file_count"].(int) + 1
				}
			} else if part.Text != "" && part.Text != "\n" {
				outputTexts = append(outputTexts, part.Text)
			}
		}

		if candidate.FinishReason != nil {
			info.Other["finish_reason"] = *candidate.FinishReason
		}
	}

	info.Other["output_content"] = strings.Join(outputTexts, "\n")

	if len(thinkingTexts) > 0 {
		info.Other["thinking_content"] = strings.Join(thinkingTexts, "\n")
	}

	if len(functionCalls) > 0 {
		info.Other["function_calls"] = functionCalls
	}

	if len(safetyRatings) > 0 {
		info.Other["safety_ratings"] = safetyRatings
	}

	if len(codeExecutions) > 0 {
		info.Other["code_execution"] = codeExecutions
	}

	totalMultimodal := multimodalSummary["image_count"].(int) + multimodalSummary["audio_count"].(int) + multimodalSummary["file_count"].(int)
	if totalMultimodal > 0 {
		multimodalSummary["has_multimodal"] = true
		info.Other["multimodal_summary"] = multimodalSummary
	}

	if response.UsageMetadata.TotalTokenCount > 0 {
		info.Other["usage_metadata"] = map[string]interface{}{
			"prompt_tokens":     response.UsageMetadata.PromptTokenCount,
			"candidates_tokens": response.UsageMetadata.CandidatesTokenCount,
			"total_tokens":      response.UsageMetadata.TotalTokenCount,
			"thoughts_tokens":   response.UsageMetadata.ThoughtsTokenCount,
		}
	}
}

func extractGeminiStreamContent(info *relaycommon.RelayInfo, accumulatedContent string, accumulatedThinking string, accumulatedFunctionCalls []interface{}, accumulatedSafetyRatings []interface{}, accumulatedCodeExecutions []interface{}, multimodalSummary map[string]interface{}) {
	if info.Other == nil {
		info.Other = make(map[string]interface{})
	}

	info.Other["output_content"] = accumulatedContent

	if accumulatedThinking != "" {
		info.Other["thinking_content"] = accumulatedThinking
	}

	if len(accumulatedFunctionCalls) > 0 {
		info.Other["function_calls"] = accumulatedFunctionCalls
	}

	if len(accumulatedSafetyRatings) > 0 {
		info.Other["safety_ratings"] = accumulatedSafetyRatings
	}

	if len(accumulatedCodeExecutions) > 0 {
		info.Other["code_execution"] = accumulatedCodeExecutions
	}

	totalMultimodal := multimodalSummary["image_count"].(int) + multimodalSummary["audio_count"].(int) + multimodalSummary["file_count"].(int)
	if totalMultimodal > 0 {
		multimodalSummary["has_multimodal"] = true
		info.Other["multimodal_summary"] = multimodalSummary
	}
}

// Setting safety to the lowest possible values since Gemini is already powerless enough
func CovertGemini2OpenAI(textRequest dto.GeneralOpenAIRequest, info *relaycommon.RelayInfo) (*GeminiChatRequest, error) {

	geminiRequest := GeminiChatRequest{
		Contents: make([]GeminiChatContent, 0, len(textRequest.Messages)),
		GenerationConfig: GeminiChatGenerationConfig{
			Temperature:     textRequest.Temperature,
			TopP:            textRequest.TopP,
			MaxOutputTokens: textRequest.MaxTokens,
			Seed:            int64(textRequest.Seed),
		},
	}

	// Check if the current model supports thinking budget
	if isModelSupportedThinkingBudget(info.OriginModelName, model_setting.GetGeminiSettings().ModelsSupportedThinkingBudget) {
		geminiRequest.GenerationConfig.ThinkingConfig = &GeminiThinkingConfig{
			IncludeThoughts: true,
		}
	}

	if model_setting.IsGeminiModelSupportImagine(info.UpstreamModelName) {
		geminiRequest.GenerationConfig.ResponseModalities = []string{
			"TEXT",
			"IMAGE",
		}
	}

	if model_setting.GetGeminiSettings().ThinkingAdapterEnabled {
		if geminiRequest.GenerationConfig.ThinkingConfig != nil { // Only apply if ThinkingConfig was instantiated
			if strings.HasSuffix(info.OriginModelName, "-thinking") {
				budgetTokens := model_setting.GetGeminiSettings().ThinkingAdapterBudgetTokensPercentage * float64(geminiRequest.GenerationConfig.MaxOutputTokens)
				if budgetTokens == 0 || budgetTokens > 24576 {
					budgetTokens = 24576
				}
				geminiRequest.GenerationConfig.ThinkingConfig.SetThinkingBudget(int(budgetTokens))
			} else if strings.HasSuffix(info.OriginModelName, "-nothinking") {
				geminiRequest.GenerationConfig.ThinkingConfig.SetThinkingBudget(0)
			}
		}
	}

	safetySettings := make([]GeminiChatSafetySettings, 0, len(SafetySettingList))
	for _, category := range SafetySettingList {
		safetySettings = append(safetySettings, GeminiChatSafetySettings{
			Category:  category,
			Threshold: model_setting.GetGeminiSafetySetting(category),
		})
	}
	geminiRequest.SafetySettings = safetySettings

	// openaiContent.FuncToToolCalls()
	if textRequest.Tools != nil {
		functions := make([]dto.FunctionRequest, 0, len(textRequest.Tools))
		googleSearch := false
		codeExecution := false
		for _, tool := range textRequest.Tools {
			if tool.Function.Name == "googleSearch" {
				googleSearch = true
				continue
			}
			if tool.Function.Name == "codeExecution" {
				codeExecution = true
				continue
			}
			if tool.Function.Parameters != nil {

				params, ok := tool.Function.Parameters.(map[string]interface{})
				if ok {
					if props, hasProps := params["properties"].(map[string]interface{}); hasProps {
						if len(props) == 0 {
							tool.Function.Parameters = nil
						}
					}
				}
			}
			// Clean the parameters before appending
			cleanedParams := cleanFunctionParameters(tool.Function.Parameters)
			tool.Function.Parameters = cleanedParams
			functions = append(functions, tool.Function)
		}
		if codeExecution {
			geminiRequest.Tools = append(geminiRequest.Tools, GeminiChatTool{
				CodeExecution: make(map[string]string),
			})
		}
		if googleSearch {
			geminiRequest.Tools = append(geminiRequest.Tools, GeminiChatTool{
				GoogleSearch: make(map[string]string),
			})
		}
		if len(functions) > 0 {
			geminiRequest.Tools = append(geminiRequest.Tools, GeminiChatTool{
				FunctionDeclarations: functions,
			})
		}
		// common.SysLog("tools: " + fmt.Sprintf("%+v", geminiRequest.Tools))
		// json_data, _ := json.Marshal(geminiRequest.Tools)
		// common.SysLog("tools_json: " + string(json_data))
	} else if textRequest.Functions != nil {
		//geminiRequest.Tools = []GeminiChatTool{
		//	{
		//		FunctionDeclarations: textRequest.Functions,
		//	},
		//}
	}

	if textRequest.ResponseFormat != nil && (textRequest.ResponseFormat.Type == "json_schema" || textRequest.ResponseFormat.Type == "json_object") {
		geminiRequest.GenerationConfig.ResponseMimeType = "application/json"

		if textRequest.ResponseFormat.JsonSchema != nil && textRequest.ResponseFormat.JsonSchema.Schema != nil {
			cleanedSchema := removeAdditionalPropertiesWithDepth(textRequest.ResponseFormat.JsonSchema.Schema, 0)
			geminiRequest.GenerationConfig.ResponseSchema = cleanedSchema
		}
	}
	tool_call_ids := make(map[string]string)
	var system_content []string
	//shouldAddDummyModelMessage := false
	for _, message := range textRequest.Messages {
		if message.Role == "system" {
			system_content = append(system_content, message.StringContent())
			continue
		} else if message.Role == "tool" || message.Role == "function" {
			if len(geminiRequest.Contents) == 0 || geminiRequest.Contents[len(geminiRequest.Contents)-1].Role == "model" {
				geminiRequest.Contents = append(geminiRequest.Contents, GeminiChatContent{
					Role: "user",
				})
			}
			var parts = &geminiRequest.Contents[len(geminiRequest.Contents)-1].Parts
			name := ""
			if message.Name != nil {
				name = *message.Name
			} else if val, exists := tool_call_ids[message.ToolCallId]; exists {
				name = val
			}
			content := common.StrToMap(message.StringContent())
			functionResp := &FunctionResponse{
				Name: name,
				Response: GeminiFunctionResponseContent{
					Name:    name,
					Content: content,
				},
			}
			if content == nil {
				functionResp.Response.Content = message.StringContent()
			}
			*parts = append(*parts, GeminiPart{
				FunctionResponse: functionResp,
			})
			continue
		}
		var parts []GeminiPart
		content := GeminiChatContent{
			Role: message.Role,
		}
		// isToolCall := false
		if message.ToolCalls != nil {
			// message.Role = "model"
			// isToolCall = true
			for _, call := range message.ParseToolCalls() {
				args := map[string]interface{}{}
				if call.Function.Arguments != "" {
					if json.Unmarshal([]byte(call.Function.Arguments), &args) != nil {
						return nil, fmt.Errorf("invalid arguments for function %s, args: %s", call.Function.Name, call.Function.Arguments)
					}
				}
				toolCall := GeminiPart{
					FunctionCall: &FunctionCall{
						FunctionName: call.Function.Name,
						Arguments:    args,
					},
				}
				parts = append(parts, toolCall)
				tool_call_ids[call.ID] = call.Function.Name
			}
		}

		openaiContent := message.ParseContent()
		imageNum := 0
		for _, part := range openaiContent {
			if part.Type == dto.ContentTypeText {
				if part.Text == "" {
					continue
				}
				parts = append(parts, GeminiPart{
					Text: part.Text,
				})
			} else if part.Type == dto.ContentTypeImageURL {
				imageNum += 1

				if constant.GeminiVisionMaxImageNum != -1 && imageNum > constant.GeminiVisionMaxImageNum {
					return nil, fmt.Errorf("too many images in the message, max allowed is %d", constant.GeminiVisionMaxImageNum)
				}
				// 判断是否是url
				if strings.HasPrefix(part.GetImageMedia().Url, "http") {
					// 是url，获取图片的类型和base64编码的数据
					fileData, err := service.GetFileBase64FromUrl(part.GetImageMedia().Url)
					if err != nil {
						return nil, fmt.Errorf("get file base64 from url failed: %s", err.Error())
					}
					parts = append(parts, GeminiPart{
						InlineData: &GeminiInlineData{
							MimeType: fileData.MimeType,
							Data:     fileData.Base64Data,
						},
					})
				} else {
					format, base64String, err := service.DecodeBase64FileData(part.GetImageMedia().Url)
					if err != nil {
						return nil, fmt.Errorf("decode base64 image data failed: %s", err.Error())
					}
					parts = append(parts, GeminiPart{
						InlineData: &GeminiInlineData{
							MimeType: format,
							Data:     base64String,
						},
					})
				}
			} else if part.Type == dto.ContentTypeFile {
				if part.GetFile().FileId != "" {
					return nil, fmt.Errorf("only base64 file is supported in gemini")
				}
				format, base64String, err := service.DecodeBase64FileData(part.GetFile().FileData)
				if err != nil {
					return nil, fmt.Errorf("decode base64 file data failed: %s", err.Error())
				}
				parts = append(parts, GeminiPart{
					InlineData: &GeminiInlineData{
						MimeType: format,
						Data:     base64String,
					},
				})
			} else if part.Type == dto.ContentTypeInputAudio {
				if part.GetInputAudio().Data == "" {
					return nil, fmt.Errorf("only base64 audio is supported in gemini")
				}
				format, base64String, err := service.DecodeBase64FileData(part.GetInputAudio().Data)
				if err != nil {
					return nil, fmt.Errorf("decode base64 audio data failed: %s", err.Error())
				}
				parts = append(parts, GeminiPart{
					InlineData: &GeminiInlineData{
						MimeType: format,
						Data:     base64String,
					},
				})
			}
		}

		content.Parts = parts

		// there's no assistant role in gemini and API shall vomit if Role is not user or model
		if content.Role == "assistant" {
			content.Role = "model"
		}
		if len(content.Parts) > 0 {
			geminiRequest.Contents = append(geminiRequest.Contents, content)
		}
	}

	if len(system_content) > 0 {
		geminiRequest.SystemInstructions = &GeminiChatContent{
			Parts: []GeminiPart{
				{
					Text: strings.Join(system_content, "\n"),
				},
			},
		}
	}

	return &geminiRequest, nil
}

// isModelSupportedThinkingBudget checks if a given model is in the list of models supporting thinking budget.
func isModelSupportedThinkingBudget(model string, supportedModels []string) bool {
	for _, supportedModel := range supportedModels {
		if supportedModel == model {
			return true
		}
	}
	return false
}

// cleanFunctionParameters recursively removes unsupported fields from Gemini function parameters.
func cleanFunctionParameters(params interface{}) interface{} {
	if params == nil {
		return nil
	}

	switch v := params.(type) {
	case map[string]interface{}:
		// Create a copy to avoid modifying the original
		cleanedMap := make(map[string]interface{})
		for k, val := range v {
			cleanedMap[k] = val
		}

		// Remove unsupported root-level fields
		delete(cleanedMap, "default")
		delete(cleanedMap, "exclusiveMaximum")
		delete(cleanedMap, "exclusiveMinimum")
		delete(cleanedMap, "$schema")
		delete(cleanedMap, "additionalProperties")

		// Check and clean 'format' for string types
		if propType, typeExists := cleanedMap["type"].(string); typeExists && propType == "string" {
			if formatValue, formatExists := cleanedMap["format"].(string); formatExists {
				if formatValue != "enum" && formatValue != "date-time" {
					delete(cleanedMap, "format")
				}
			}
		}

		// Clean properties
		if props, ok := cleanedMap["properties"].(map[string]interface{}); ok && props != nil {
			cleanedProps := make(map[string]interface{})
			for propName, propValue := range props {
				cleanedProps[propName] = cleanFunctionParameters(propValue)
			}
			cleanedMap["properties"] = cleanedProps
		}

		// Recursively clean items in arrays
		if items, ok := cleanedMap["items"].(map[string]interface{}); ok && items != nil {
			cleanedMap["items"] = cleanFunctionParameters(items)
		}
		// Also handle items if it's an array of schemas
		if itemsArray, ok := cleanedMap["items"].([]interface{}); ok {
			cleanedItemsArray := make([]interface{}, len(itemsArray))
			for i, item := range itemsArray {
				cleanedItemsArray[i] = cleanFunctionParameters(item)
			}
			cleanedMap["items"] = cleanedItemsArray
		}

		// Recursively clean other schema composition keywords
		for _, field := range []string{"allOf", "anyOf", "oneOf"} {
			if nested, ok := cleanedMap[field].([]interface{}); ok {
				cleanedNested := make([]interface{}, len(nested))
				for i, item := range nested {
					cleanedNested[i] = cleanFunctionParameters(item)
				}
				cleanedMap[field] = cleanedNested
			}
		}

		// Recursively clean patternProperties
		if patternProps, ok := cleanedMap["patternProperties"].(map[string]interface{}); ok {
			cleanedPatternProps := make(map[string]interface{})
			for pattern, schema := range patternProps {
				cleanedPatternProps[pattern] = cleanFunctionParameters(schema)
			}
			cleanedMap["patternProperties"] = cleanedPatternProps
		}

		// Recursively clean definitions
		if definitions, ok := cleanedMap["definitions"].(map[string]interface{}); ok {
			cleanedDefinitions := make(map[string]interface{})
			for defName, defSchema := range definitions {
				cleanedDefinitions[defName] = cleanFunctionParameters(defSchema)
			}
			cleanedMap["definitions"] = cleanedDefinitions
		}

		// Recursively clean $defs (newer JSON Schema draft)
		if defs, ok := cleanedMap["$defs"].(map[string]interface{}); ok {
			cleanedDefs := make(map[string]interface{})
			for defName, defSchema := range defs {
				cleanedDefs[defName] = cleanFunctionParameters(defSchema)
			}
			cleanedMap["$defs"] = cleanedDefs
		}

		// Clean conditional keywords
		for _, field := range []string{"if", "then", "else", "not"} {
			if nested, ok := cleanedMap[field]; ok {
				cleanedMap[field] = cleanFunctionParameters(nested)
			}
		}

		return cleanedMap

	case []interface{}:
		// Handle arrays of schemas
		cleanedArray := make([]interface{}, len(v))
		for i, item := range v {
			cleanedArray[i] = cleanFunctionParameters(item)
		}
		return cleanedArray

	default:
		// Not a map or array, return as is (e.g., could be a primitive)
		return params
	}
}

func removeAdditionalPropertiesWithDepth(schema interface{}, depth int) interface{} {
	if depth >= 5 {
		return schema
	}

	v, ok := schema.(map[string]interface{})
	if !ok || len(v) == 0 {
		return schema
	}
	// 删除所有的title字段
	delete(v, "title")
	// 如果type不为object和array，则直接返回
	if typeVal, exists := v["type"]; !exists || (typeVal != "object" && typeVal != "array") {
		return schema
	}
	switch v["type"] {
	case "object":
		delete(v, "additionalProperties")
		// 处理 properties
		if properties, ok := v["properties"].(map[string]interface{}); ok {
			for key, value := range properties {
				properties[key] = removeAdditionalPropertiesWithDepth(value, depth+1)
			}
		}
		for _, field := range []string{"allOf", "anyOf", "oneOf"} {
			if nested, ok := v[field].([]interface{}); ok {
				for i, item := range nested {
					nested[i] = removeAdditionalPropertiesWithDepth(item, depth+1)
				}
			}
		}
	case "array":
		if items, ok := v["items"].(map[string]interface{}); ok {
			v["items"] = removeAdditionalPropertiesWithDepth(items, depth+1)
		}
	}

	return v
}

func unescapeString(s string) (string, error) {
	var result []rune
	escaped := false
	i := 0

	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:]) // 正确解码UTF-8字符
		if r == utf8.RuneError {
			return "", fmt.Errorf("invalid UTF-8 encoding")
		}

		if escaped {
			// 如果是转义符后的字符，检查其类型
			switch r {
			case '"':
				result = append(result, '"')
			case '\\':
				result = append(result, '\\')
			case '/':
				result = append(result, '/')
			case 'b':
				result = append(result, '\b')
			case 'f':
				result = append(result, '\f')
			case 'n':
				result = append(result, '\n')
			case 'r':
				result = append(result, '\r')
			case 't':
				result = append(result, '\t')
			case '\'':
				result = append(result, '\'')
			default:
				// 如果遇到一个非法的转义字符，直接按原样输出
				result = append(result, '\\', r)
			}
			escaped = false
		} else {
			if r == '\\' {
				escaped = true // 记录反斜杠作为转义符
			} else {
				result = append(result, r)
			}
		}
		i += size // 移动到下一个字符
	}

	return string(result), nil
}
func unescapeMapOrSlice(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		for k, val := range v {
			v[k] = unescapeMapOrSlice(val)
		}
	case []interface{}:
		for i, val := range v {
			v[i] = unescapeMapOrSlice(val)
		}
	case string:
		if unescaped, err := unescapeString(v); err != nil {
			return v
		} else {
			return unescaped
		}
	}
	return data
}

func getResponseToolCall(item *GeminiPart) *dto.ToolCallResponse {
	var argsBytes []byte
	var err error
	if result, ok := item.FunctionCall.Arguments.(map[string]interface{}); ok {
		argsBytes, err = json.Marshal(unescapeMapOrSlice(result))
	} else {
		argsBytes, err = json.Marshal(item.FunctionCall.Arguments)
	}

	if err != nil {
		return nil
	}
	return &dto.ToolCallResponse{
		ID:   fmt.Sprintf("call_%s", common.GetUUID()),
		Type: "function",
		Function: dto.FunctionResponse{
			Arguments: string(argsBytes),
			Name:      item.FunctionCall.FunctionName,
		},
	}
}

func responseGeminiChat2OpenAI(response *GeminiChatResponse) *dto.OpenAITextResponse {
	fullTextResponse := dto.OpenAITextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
		Choices: make([]dto.OpenAITextResponseChoice, 0, len(response.Candidates)),
	}
	content, _ := json.Marshal("")
	isToolCall := false
	for _, candidate := range response.Candidates {
		choice := dto.OpenAITextResponseChoice{
			Index: int(candidate.Index),
			Message: dto.Message{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: constant.FinishReasonStop,
		}
		if len(candidate.Content.Parts) > 0 {
			var texts []string
			var toolCalls []dto.ToolCallResponse
			for _, part := range candidate.Content.Parts {
				if part.FunctionCall != nil {
					choice.FinishReason = constant.FinishReasonToolCalls
					if call := getResponseToolCall(&part); call != nil {
						toolCalls = append(toolCalls, *call)
					}
				} else if part.Thought {
					choice.Message.ReasoningContent = part.Text
				} else {
					if part.ExecutableCode != nil {
						texts = append(texts, "```"+part.ExecutableCode.Language+"\n"+part.ExecutableCode.Code+"\n```")
					} else if part.CodeExecutionResult != nil {
						texts = append(texts, "```output\n"+part.CodeExecutionResult.Output+"\n```")
					} else {
						// 过滤掉空行
						if part.Text != "\n" {
							texts = append(texts, part.Text)
						}
					}
				}
			}
			if len(toolCalls) > 0 {
				choice.Message.SetToolCalls(toolCalls)
				isToolCall = true
			}

			choice.Message.SetStringContent(strings.Join(texts, "\n"))

		}
		if candidate.FinishReason != nil {
			switch *candidate.FinishReason {
			case "STOP":
				choice.FinishReason = constant.FinishReasonStop
			case "MAX_TOKENS":
				choice.FinishReason = constant.FinishReasonLength
			default:
				choice.FinishReason = constant.FinishReasonContentFilter
			}
		}
		if isToolCall {
			choice.FinishReason = constant.FinishReasonToolCalls
		}

		fullTextResponse.Choices = append(fullTextResponse.Choices, choice)
	}
	return &fullTextResponse
}

func streamResponseGeminiChat2OpenAI(geminiResponse *GeminiChatResponse) (*dto.ChatCompletionsStreamResponse, bool, bool) {
	choices := make([]dto.ChatCompletionsStreamResponseChoice, 0, len(geminiResponse.Candidates))
	isStop := false
	hasImage := false
	for _, candidate := range geminiResponse.Candidates {
		if candidate.FinishReason != nil && *candidate.FinishReason == "STOP" {
			isStop = true
			candidate.FinishReason = nil
		}
		choice := dto.ChatCompletionsStreamResponseChoice{
			Index: int(candidate.Index),
			Delta: dto.ChatCompletionsStreamResponseChoiceDelta{
				Role: "assistant",
			},
		}
		var texts []string
		isTools := false
		isThought := false
		if candidate.FinishReason != nil {
			// p := GeminiConvertFinishReason(*candidate.FinishReason)
			switch *candidate.FinishReason {
			case "STOP":
				choice.FinishReason = &constant.FinishReasonStop
			case "MAX_TOKENS":
				choice.FinishReason = &constant.FinishReasonLength
			default:
				choice.FinishReason = &constant.FinishReasonContentFilter
			}
		}
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil {
				if strings.HasPrefix(part.InlineData.MimeType, "image") {
					imgText := "![image](data:" + part.InlineData.MimeType + ";base64," + part.InlineData.Data + ")"
					texts = append(texts, imgText)
					hasImage = true
				}
			} else if part.FunctionCall != nil {
				isTools = true
				if call := getResponseToolCall(&part); call != nil {
					call.SetIndex(len(choice.Delta.ToolCalls))
					choice.Delta.ToolCalls = append(choice.Delta.ToolCalls, *call)
				}
			} else if part.Thought {
				isThought = true
				texts = append(texts, part.Text)
			} else {
				if part.ExecutableCode != nil {
					texts = append(texts, "```"+part.ExecutableCode.Language+"\n"+part.ExecutableCode.Code+"\n```\n")
				} else if part.CodeExecutionResult != nil {
					texts = append(texts, "```output\n"+part.CodeExecutionResult.Output+"\n```\n")
				} else {
					if part.Text != "\n" {
						texts = append(texts, part.Text)
					}
				}
			}
		}
		if isThought {
			choice.Delta.SetReasoningContent(strings.Join(texts, "\n"))
		} else {
			choice.Delta.SetContentString(strings.Join(texts, "\n"))
		}
		if isTools {
			choice.FinishReason = &constant.FinishReasonToolCalls
		}
		choices = append(choices, choice)
	}

	var response dto.ChatCompletionsStreamResponse
	response.Object = "chat.completion.chunk"
	response.Choices = choices
	return &response, isStop, hasImage
}

func GeminiChatStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	extractGeminiInputContent(info)

	id := fmt.Sprintf("chatcmpl-%s", common.GetUUID())
	createAt := common.GetTimestamp()
	var usage = &dto.Usage{}
	var imageCount int

	var accumulatedContent strings.Builder
	var accumulatedThinking strings.Builder
	var accumulatedFunctionCalls []interface{}
	var accumulatedSafetyRatings []interface{}
	var accumulatedCodeExecutions []interface{}
	var multimodalSummary = map[string]interface{}{
		"image_count": 0,
		"audio_count": 0,
		"file_count":  0,
	}

	helper.StreamScannerHandler(c, resp, info, func(data string) bool {
		var geminiResponse GeminiChatResponse
		err := common.DecodeJsonStr(data, &geminiResponse)
		if err != nil {
			common.LogError(c, "error unmarshalling stream response: "+err.Error())
			return false
		}

		for _, candidate := range geminiResponse.Candidates {
			if len(candidate.SafetyRatings) > 0 {
				for _, rating := range candidate.SafetyRatings {
					accumulatedSafetyRatings = append(accumulatedSafetyRatings, map[string]interface{}{
						"category":    rating.Category,
						"probability": rating.Probability,
					})
				}
			}

			for _, part := range candidate.Content.Parts {
				if part.FunctionCall != nil {
					accumulatedFunctionCalls = append(accumulatedFunctionCalls, map[string]interface{}{
						"name":      part.FunctionCall.FunctionName,
						"arguments": part.FunctionCall.Arguments,
					})
				} else if part.FunctionResponse != nil {
					accumulatedFunctionCalls = append(accumulatedFunctionCalls, map[string]interface{}{
						"type":     "response",
						"name":     part.FunctionResponse.Name,
						"response": part.FunctionResponse.Response,
					})
				} else if part.Thought {
					accumulatedThinking.WriteString(part.Text)
				} else if part.ExecutableCode != nil {
					accumulatedCodeExecutions = append(accumulatedCodeExecutions, map[string]interface{}{
						"type":     "code",
						"language": part.ExecutableCode.Language,
						"code":     part.ExecutableCode.Code,
					})
				} else if part.CodeExecutionResult != nil {
					accumulatedCodeExecutions = append(accumulatedCodeExecutions, map[string]interface{}{
						"type":    "result",
						"outcome": part.CodeExecutionResult.Outcome,
						"output":  part.CodeExecutionResult.Output,
					})
				} else if part.InlineData != nil {
					if strings.HasPrefix(part.InlineData.MimeType, "image") {
						multimodalSummary["image_count"] = multimodalSummary["image_count"].(int) + 1
					} else if strings.HasPrefix(part.InlineData.MimeType, "audio") {
						multimodalSummary["audio_count"] = multimodalSummary["audio_count"].(int) + 1
					} else {
						multimodalSummary["file_count"] = multimodalSummary["file_count"].(int) + 1
					}
				} else if part.Text != "" && part.Text != "\n" {
					accumulatedContent.WriteString(part.Text)
				}
			}
		}

		response, isStop, hasImage := streamResponseGeminiChat2OpenAI(&geminiResponse)
		if hasImage {
			imageCount++
		}
		response.Id = id
		response.Created = createAt
		response.Model = info.UpstreamModelName
		if geminiResponse.UsageMetadata.TotalTokenCount != 0 {
			usage.PromptTokens = geminiResponse.UsageMetadata.PromptTokenCount
			usage.CompletionTokens = geminiResponse.UsageMetadata.CandidatesTokenCount
			usage.CompletionTokenDetails.ReasoningTokens = geminiResponse.UsageMetadata.ThoughtsTokenCount
		}
		err = helper.ObjectData(c, response)
		if err != nil {
			common.LogError(c, err.Error())
		}
		if isStop {
			response := helper.GenerateStopResponse(id, createAt, info.UpstreamModelName, constant.FinishReasonStop)
			helper.ObjectData(c, response)
		}
		return true
	})

	extractGeminiStreamContent(info, accumulatedContent.String(), accumulatedThinking.String(), accumulatedFunctionCalls, accumulatedSafetyRatings, accumulatedCodeExecutions, multimodalSummary)

	var response *dto.ChatCompletionsStreamResponse

	if imageCount != 0 {
		if usage.CompletionTokens == 0 {
			usage.CompletionTokens = imageCount * 258
		}
	}

	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	usage.PromptTokensDetails.TextTokens = usage.PromptTokens
	//usage.CompletionTokenDetails.TextTokens = usage.CompletionTokens

	if info.ShouldIncludeUsage {
		response = helper.GenerateFinalUsageResponse(id, createAt, info.UpstreamModelName, *usage)
		err := helper.ObjectData(c, response)
		if err != nil {
			common.SysError("send final response failed: " + err.Error())
		}
	}
	helper.Done(c)
	//resp.Body.Close()
	return nil, usage
}

func GeminiChatHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	extractGeminiInputContent(info)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	var geminiResponse GeminiChatResponse
	err = json.Unmarshal(responseBody, &geminiResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if len(geminiResponse.Candidates) == 0 {
		return &dto.OpenAIErrorWithStatusCode{
			Error: dto.OpenAIError{
				Message: "No candidates returned",
				Type:    "server_error",
				Param:   "",
				Code:    500,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	extractGeminiOutputContent(info, &geminiResponse)

	fullTextResponse := responseGeminiChat2OpenAI(&geminiResponse)
	fullTextResponse.Model = info.UpstreamModelName
	usage := &dto.Usage{
		PromptTokens:     geminiResponse.UsageMetadata.PromptTokenCount,
		CompletionTokens: geminiResponse.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      geminiResponse.UsageMetadata.TotalTokenCount,
	}
	if geminiResponse.UsageMetadata.ThoughtsTokenCount > 0 {
		usage.CompletionTokenDetails.ReasoningTokens = geminiResponse.UsageMetadata.ThoughtsTokenCount
	}
	fullTextResponse.Usage = *usage
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, usage
}

func GeminiEmbeddingHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *dto.OpenAIErrorWithStatusCode) {
	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, service.OpenAIErrorWrapper(readErr, "read_response_body_failed", http.StatusInternalServerError)
	}
	_ = resp.Body.Close()

	var geminiResponse GeminiEmbeddingResponse
	if jsonErr := json.Unmarshal(responseBody, &geminiResponse); jsonErr != nil {
		return nil, service.OpenAIErrorWrapper(jsonErr, "unmarshal_response_body_failed", http.StatusInternalServerError)
	}

	// convert to openai format response
	openAIResponse := dto.OpenAIEmbeddingResponse{
		Object: "list",
		Data: []dto.OpenAIEmbeddingResponseItem{
			{
				Object:    "embedding",
				Embedding: geminiResponse.Embedding.Values,
				Index:     0,
			},
		},
		Model: info.UpstreamModelName,
	}

	// calculate usage
	// https://ai.google.dev/gemini-api/docs/pricing?hl=zh-cn#text-embedding-004
	// Google has not yet clarified how embedding models will be billed
	// refer to openai billing method to use input tokens billing
	// https://platform.openai.com/docs/guides/embeddings#what-are-embeddings
	usage = &dto.Usage{
		PromptTokens:     info.PromptTokens,
		CompletionTokens: 0,
		TotalTokens:      info.PromptTokens,
	}
	openAIResponse.Usage = *usage.(*dto.Usage)

	jsonResponse, jsonErr := json.Marshal(openAIResponse)
	if jsonErr != nil {
		return nil, service.OpenAIErrorWrapper(jsonErr, "marshal_response_failed", http.StatusInternalServerError)
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)

	return usage, nil
}

func GeminiChatPseudoStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	extractGeminiInputContent(info)

	geminiResp, errResp := parseGeminiChatResponse(resp)
	if errResp != nil {
		return errResp, nil
	}

	extractGeminiOutputContent(info, geminiResp)

	fullTextResponse := responseGeminiChat2OpenAI(geminiResp)
	fullTextResponse.Model = info.UpstreamModelName
	usage := buildGeminiUsage(geminiResp)
	fullTextResponse.Usage = *usage

	helper.SetEventStreamHeaders(c)
	info.SetFirstResponseTime()

	streamResp := openaichannel.BuildStreamChunkFromTextResponse(fullTextResponse)
	_ = helper.ObjectData(c, streamResp)
	if info.ShouldIncludeUsage {
		final := helper.GenerateFinalUsageResponse(helper.GetResponseID(c), common.GetTimestamp(), info.UpstreamModelName, *usage)
		_ = helper.ObjectData(c, final)
	}
	helper.Done(c)
	return nil, usage
}

func parseGeminiChatResponse(resp *http.Response) (*GeminiChatResponse, *dto.OpenAIErrorWithStatusCode) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	if err = resp.Body.Close(); err != nil {
		return nil, service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}
	var geminiResponse GeminiChatResponse
	if err = json.Unmarshal(body, &geminiResponse); err != nil {
		return nil, service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
	}
	if len(geminiResponse.Candidates) == 0 {
		return nil, &dto.OpenAIErrorWithStatusCode{
			Error:      dto.OpenAIError{Message: "No candidates returned", Type: "server_error", Code: 500},
			StatusCode: resp.StatusCode,
		}
	}
	return &geminiResponse, nil
}

func buildGeminiUsage(gResp *GeminiChatResponse) *dto.Usage {
	usage := &dto.Usage{
		PromptTokens:     gResp.UsageMetadata.PromptTokenCount,
		CompletionTokens: gResp.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      gResp.UsageMetadata.TotalTokenCount,
	}
	if gResp.UsageMetadata.ThoughtsTokenCount > 0 {
		usage.CompletionTokenDetails.ReasoningTokens = gResp.UsageMetadata.ThoughtsTokenCount
	}
	return usage
}
