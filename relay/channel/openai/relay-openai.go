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
package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"veloera/common"
	"veloera/constant"
	"veloera/dto"
	relaycommon "veloera/relay/common"
	"veloera/relay/helper"
	"veloera/service"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func sendStreamData(c *gin.Context, info *relaycommon.RelayInfo, data string, forceFormat bool, thinkToContent bool) error {
	if data == "" {
		return nil
	}

	if !forceFormat && !thinkToContent {
		return helper.StringData(c, data)
	}

	var lastStreamResponse dto.ChatCompletionsStreamResponse
	if err := common.DecodeJsonStr(data, &lastStreamResponse); err != nil {
		return err
	}

	if !thinkToContent {
		return helper.ObjectData(c, lastStreamResponse)
	}

	hasThinkingContent := false
	hasContent := false
	var thinkingContent strings.Builder
	for _, choice := range lastStreamResponse.Choices {
		if len(choice.Delta.GetReasoningContent()) > 0 {
			hasThinkingContent = true
			thinkingContent.WriteString(choice.Delta.GetReasoningContent())
		}
		if len(choice.Delta.GetContentString()) > 0 {
			hasContent = true
		}
	}

	// Handle think to content conversion
	if info.ThinkingContentInfo.IsFirstThinkingContent {
		if hasThinkingContent {
			response := lastStreamResponse.Copy()
			for i := range response.Choices {
				// send `think` tag with thinking content
				response.Choices[i].Delta.SetContentString("<think>\n" + thinkingContent.String())
				response.Choices[i].Delta.ReasoningContent = nil
				response.Choices[i].Delta.Reasoning = nil
			}
			info.ThinkingContentInfo.IsFirstThinkingContent = false
			info.ThinkingContentInfo.HasSentThinkingContent = true
			return helper.ObjectData(c, response)
		}
	}

	if lastStreamResponse.Choices == nil || len(lastStreamResponse.Choices) == 0 {
		return helper.ObjectData(c, lastStreamResponse)
	}

	// Process each choice
	for i, choice := range lastStreamResponse.Choices {
		// Handle transition from thinking to content
		// only send `</think>` tag when previous thinking content has been sent
		if hasContent && !info.ThinkingContentInfo.SendLastThinkingContent && info.ThinkingContentInfo.HasSentThinkingContent {
			response := lastStreamResponse.Copy()
			for j := range response.Choices {
				response.Choices[j].Delta.SetContentString("\n</think>\n")
				response.Choices[j].Delta.ReasoningContent = nil
				response.Choices[j].Delta.Reasoning = nil
			}
			info.ThinkingContentInfo.SendLastThinkingContent = true
			helper.ObjectData(c, response)
		}

		// Convert reasoning content to regular content if any
		if len(choice.Delta.GetReasoningContent()) > 0 {
			lastStreamResponse.Choices[i].Delta.SetContentString(choice.Delta.GetReasoningContent())
			lastStreamResponse.Choices[i].Delta.ReasoningContent = nil
			lastStreamResponse.Choices[i].Delta.Reasoning = nil
		} else if !hasThinkingContent && !hasContent {
			// flush thinking content
			lastStreamResponse.Choices[i].Delta.ReasoningContent = nil
			lastStreamResponse.Choices[i].Delta.Reasoning = nil
		}
	}

	return helper.ObjectData(c, lastStreamResponse)
}

func OaiStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	if resp == nil || resp.Body == nil {
		common.LogError(c, "invalid response or response body")
		return service.OpenAIErrorWrapper(fmt.Errorf("invalid response"), "invalid_response", http.StatusInternalServerError), nil
	}

	containStreamUsage := false
	var responseId string
	var createAt int64 = 0
	var systemFingerprint string
	model := info.UpstreamModelName

	var responseTextBuilder strings.Builder
	var toolCount int
	var usage = &dto.Usage{}
	var streamItems []string // store stream items
	var forceFormat bool
	var thinkToContent bool

	if forceFmt, ok := info.ChannelSetting[constant.ForceFormat].(bool); ok {
		forceFormat = forceFmt
	}

	if think2Content, ok := info.ChannelSetting[constant.ChannelSettingThinkingToContent].(bool); ok {
		thinkToContent = think2Content
	}

	var (
		lastStreamData string
	)

	helper.StreamScannerHandler(c, resp, info, func(data string) bool {
		if lastStreamData != "" {
			err := handleStreamFormat(c, info, lastStreamData, forceFormat, thinkToContent)
			if err != nil {
				common.SysError("error handling stream format: " + err.Error())
			}
		}
		lastStreamData = data
		streamItems = append(streamItems, data)
		return true
	})

	shouldSendLastResp := true
	var lastStreamResponse dto.ChatCompletionsStreamResponse
	err := common.DecodeJsonStr(lastStreamData, &lastStreamResponse)
	if err == nil {
		responseId = lastStreamResponse.Id
		createAt = lastStreamResponse.Created
		systemFingerprint = lastStreamResponse.GetSystemFingerprint()
		model = lastStreamResponse.Model
		if service.ValidUsage(lastStreamResponse.Usage) {
			containStreamUsage = true
			usage = lastStreamResponse.Usage
			if !info.ShouldIncludeUsage {
				shouldSendLastResp = false
			}
		}
		for _, choice := range lastStreamResponse.Choices {
			if choice.FinishReason != nil {
				shouldSendLastResp = true
			}
		}
	}

	if shouldSendLastResp {
		sendStreamData(c, info, lastStreamData, forceFormat, thinkToContent)
		//err = handleStreamFormat(c, info, lastStreamData, forceFormat, thinkToContent)
	}

	// 处理token计算
	if err := processTokens(info.RelayMode, streamItems, &responseTextBuilder, &toolCount); err != nil {
		common.SysError("error processing tokens: " + err.Error())
	}

	// 检查是否为空回复或只有空格的回复，如果是则不计费
	responseText := responseTextBuilder.String()
	// 保存到 info 中，供日志记录使用
	if info.Other == nil {
		info.Other = make(map[string]interface{})
	}

	// 提取输入内容（最后一条用户消息）和上下文（其他所有非system消息）
	if messages, ok := info.PromptMessages.([]interface{}); ok && len(messages) > 0 {
		var systemPrompt string
		var contextMessages []interface{}
		var userMessage interface{}
		var lastUserMessageIndex = -1

		// 先找出最后一条user消息的索引
		for i := len(messages) - 1; i >= 0; i-- {
			if msgMap, ok := messages[i].(map[string]interface{}); ok {
				if role, exists := msgMap["role"]; exists && role == "user" {
					lastUserMessageIndex = i
					break
				}
			}
		}

		// 再遍历处理所有消息
		for i, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				if role, exists := msgMap["role"]; exists {
					if role == "system" {
						// 如果是system消息，保存其内容
						if content, hasContent := msgMap["content"]; hasContent && content != nil {
							if systemPrompt == "" {
								systemPrompt = fmt.Sprintf("%v", content)
							} else {
								systemPrompt += "\n" + fmt.Sprintf("%v", content)
							}
						}
					} else if i == lastUserMessageIndex {
						// 如果是最后一条user消息
						userMessage = msgMap
					} else {
						// 其他非system消息作为上下文
						contextMessages = append(contextMessages, msgMap)
					}
				}
			}
		}

		// 保存处理后的数据
		if systemPrompt != "" {
			info.Other["system_prompt"] = systemPrompt
		}
		info.Other["context"] = contextMessages
		if userMessage != nil {
			info.Other["input_content"] = userMessage
		} else {
			// 如果没有找到user消息，保存最后一条消息作为输入
			if len(messages) > 0 {
				info.Other["input_content"] = messages[len(messages)-1]
			}
		}
	} else {
		info.Other["input_content"] = info.PromptMessages // 备用方案，保存全部输入内容
	}

	info.Other["output_content"] = responseText // 保存输出内容

	if common.IsEmptyOrWhitespace(responseText) && toolCount == 0 {
		// 空回复或全是空格不计费，返回零使用量（而不是只设置CompletionTokens为0）
		zeroUsage := &dto.Usage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		}
		// 直接返回空使用量，结束处理
		handleFinalResponse(c, info, lastStreamData, responseId, createAt, model, systemFingerprint, zeroUsage, false)
		return nil, zeroUsage
	}

	if !containStreamUsage {
		usage, _ = service.ResponseText2Usage(responseText, info.UpstreamModelName, info.PromptTokens)
		usage.CompletionTokens += toolCount * 7
	} else {
		if info.ChannelType == common.ChannelTypeDeepSeek {
			if usage.PromptCacheHitTokens != 0 {
				usage.PromptTokensDetails.CachedTokens = usage.PromptCacheHitTokens
			}
		}
	}

	handleFinalResponse(c, info, lastStreamData, responseId, createAt, model, systemFingerprint, usage, containStreamUsage)

	return nil, usage
}

func OpenaiHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	var simpleResponse dto.OpenAITextResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = common.DecodeJson(responseBody, &simpleResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if simpleResponse.Error != nil && simpleResponse.Error.Type != "" {
		return &dto.OpenAIErrorWithStatusCode{
			Error:      *simpleResponse.Error,
			StatusCode: resp.StatusCode,
		}, nil
	}

	// 保存输入输出内容到 info 中，供日志记录使用
	if info.Other == nil {
		info.Other = make(map[string]interface{})
	}

	// 提取输入内容（最后一条用户消息）和上下文（其他所有非system消息）
	if messages, ok := info.PromptMessages.([]interface{}); ok && len(messages) > 0 {
		var systemPrompt string
		var contextMessages []interface{}
		var userMessage interface{}
		var lastUserMessageIndex = -1

		// 先找出最后一条user消息的索引
		for i := len(messages) - 1; i >= 0; i-- {
			if msgMap, ok := messages[i].(map[string]interface{}); ok {
				if role, exists := msgMap["role"]; exists && role == "user" {
					lastUserMessageIndex = i
					break
				}
			}
		}

		// 再遍历处理所有消息
		for i, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				if role, exists := msgMap["role"]; exists {
					if role == "system" {
						// 如果是system消息，保存其内容
						if content, hasContent := msgMap["content"]; hasContent && content != nil {
							if systemPrompt == "" {
								systemPrompt = fmt.Sprintf("%v", content)
							} else {
								systemPrompt += "\n" + fmt.Sprintf("%v", content)
							}
						}
					} else if i == lastUserMessageIndex {
						// 如果是最后一条user消息
						userMessage = msgMap
					} else {
						// 其他非system消息作为上下文
						contextMessages = append(contextMessages, msgMap)
					}
				}
			}
		}

		// 保存处理后的数据
		if systemPrompt != "" {
			info.Other["system_prompt"] = systemPrompt
		}
		info.Other["context"] = contextMessages
		if userMessage != nil {
			info.Other["input_content"] = userMessage
		} else {
			// 如果没有找到user消息，保存最后一条消息作为输入
			if len(messages) > 0 {
				info.Other["input_content"] = messages[len(messages)-1]
			}
		}
	} else {
		info.Other["input_content"] = info.PromptMessages // 备用方案，保存全部输入内容
	}

	// 提取输出内容
	var outputContent string
	for _, choice := range simpleResponse.Choices {
		outputContent += choice.Message.StringContent() + choice.Message.ReasoningContent + choice.Message.Reasoning
	}
	info.Other["output_content"] = outputContent // 保存输出内容

	switch info.RelayFormat {
	case relaycommon.RelayFormatOpenAI:
		break
	case relaycommon.RelayFormatClaude:
		claudeResp := service.ResponseOpenAI2Claude(&simpleResponse, info)
		claudeRespStr, err := json.Marshal(claudeResp)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
		}
		responseBody = claudeRespStr
	}

	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		//return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
		common.SysError("error copying response body: " + err.Error())
	}
	resp.Body.Close()
	// 检查响应是否为空或只包含空格
	isEmptyResponse := true
	for _, choice := range simpleResponse.Choices {
		content := choice.Message.StringContent() + choice.Message.ReasoningContent + choice.Message.Reasoning
		if !common.IsEmptyOrWhitespace(content) {
			isEmptyResponse = false
			break
		}
	}

	// 如果响应是空的或只有空格，返回零使用量
	if isEmptyResponse {
		zeroUsage := &dto.Usage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		}
		return nil, zeroUsage
	}

	if simpleResponse.Usage.TotalTokens == 0 || (simpleResponse.Usage.PromptTokens == 0 && simpleResponse.Usage.CompletionTokens == 0) {
		completionTokens := 0
		for _, choice := range simpleResponse.Choices {
			ctkm, _ := service.CountTextToken(choice.Message.StringContent()+choice.Message.ReasoningContent+choice.Message.Reasoning, info.UpstreamModelName)
			completionTokens += ctkm
		}
		simpleResponse.Usage = dto.Usage{
			PromptTokens:     info.PromptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      info.PromptTokens + completionTokens,
		}
	}
	return nil, &simpleResponse.Usage
}

func OpenaiTTSHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	usage := &dto.Usage{}
	usage.PromptTokens = info.PromptTokens
	usage.TotalTokens = info.PromptTokens
	return nil, usage
}

func OpenaiSTTHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo, responseFormat string) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	// count tokens by audio file duration
	audioTokens, err := countAudioTokens(c)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "count_audio_tokens_failed", http.StatusInternalServerError), nil
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	resp.Body.Close()

	usage := &dto.Usage{}
	usage.PromptTokens = audioTokens
	usage.CompletionTokens = 0
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return nil, usage
}

func OpenaiResponsesHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	// read response body
	var responsesResponse dto.OpenAIResponsesResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = common.DecodeJson(responseBody, &responsesResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if responsesResponse.Error != nil {
		return &dto.OpenAIErrorWithStatusCode{
			Error: dto.OpenAIError{
				Message: responsesResponse.Error.Message,
				Type:    "openai_error",
				Code:    responsesResponse.Error.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	// reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	// copy response body
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		common.SysError("error copying response body: " + err.Error())
	}
	resp.Body.Close()
	// compute usage
	usage := dto.Usage{}
	usage.PromptTokens = responsesResponse.Usage.InputTokens
	usage.CompletionTokens = responsesResponse.Usage.OutputTokens
	usage.TotalTokens = responsesResponse.Usage.TotalTokens
	return nil, &usage
}

func OaiResponsesStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	if resp == nil || resp.Body == nil {
		common.LogError(c, "invalid response or response body")
		return service.OpenAIErrorWrapper(fmt.Errorf("invalid response"), "invalid_response", http.StatusInternalServerError), nil
	}

	var usage = &dto.Usage{}
	var responseTextBuilder strings.Builder

	helper.StreamScannerHandler(c, resp, info, func(data string) bool {

		// 检查当前数据是否包含 completed 状态和 usage 信息
		var streamResponse dto.ResponsesStreamResponse
		if err := common.DecodeJsonStr(data, &streamResponse); err == nil {
			sendResponsesStreamData(c, streamResponse, data)
			switch streamResponse.Type {
			case "response.completed":
				usage.PromptTokens = streamResponse.Response.Usage.InputTokens
				usage.CompletionTokens = streamResponse.Response.Usage.OutputTokens
				usage.TotalTokens = streamResponse.Response.Usage.TotalTokens
			case "response.output_text.delta":
				// 处理输出文本
				responseTextBuilder.WriteString(streamResponse.Delta)

			}
		}
		return true
	})

	helper.Done(c)

	if usage.CompletionTokens == 0 {
		// 计算输出文本的 token 数量
		tempStr := responseTextBuilder.String()
		if len(tempStr) > 0 {
			// 非正常结束，使用输出文本的 token 数量
			completionTokens, _ := service.CountTextToken(tempStr, info.UpstreamModelName)
			usage.CompletionTokens = completionTokens
		}
	}

	return nil, usage
}

func sendResponsesStreamData(c *gin.Context, streamResponse dto.ResponsesStreamResponse, data string) {
	if data == "" {
		return
	}
	helper.ResponseChunkData(c, streamResponse, data)
}

func countAudioTokens(c *gin.Context) (int, error) {
	body, err := common.GetRequestBody(c)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	var reqBody struct {
		File *multipart.FileHeader `form:"file" binding:"required"`
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if err = c.ShouldBind(&reqBody); err != nil {
		return 0, errors.WithStack(err)
	}

	reqFp, err := reqBody.File.Open()
	if err != nil {
		return 0, errors.WithStack(err)
	}

	tmpFp, err := os.CreateTemp("", "audio-*")
	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer os.Remove(tmpFp.Name())

	_, err = io.Copy(tmpFp, reqFp)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	if err = tmpFp.Close(); err != nil {
		return 0, errors.WithStack(err)
	}

	duration, err := common.GetAudioDuration(c.Request.Context(), tmpFp.Name())
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return int(math.Round(math.Ceil(duration) / 60.0 * 1000)), nil // 1 minute 相当于 1k tokens
}

func OpenaiRealtimeHandler(c *gin.Context, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.RealtimeUsage) {
	if info == nil || info.ClientWs == nil || info.TargetWs == nil {
		return service.OpenAIErrorWrapper(fmt.Errorf("invalid websocket connection"), "invalid_connection", http.StatusBadRequest), nil
	}

	info.IsStream = true
	clientConn := info.ClientWs
	targetConn := info.TargetWs

	clientClosed := make(chan struct{})
	targetClosed := make(chan struct{})
	sendChan := make(chan []byte, 100)
	receiveChan := make(chan []byte, 100)
	errChan := make(chan error, 2)

	usage := &dto.RealtimeUsage{}
	localUsage := &dto.RealtimeUsage{}
	sumUsage := &dto.RealtimeUsage{}

	gopool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic in client reader: %v", r)
			}
		}()
		for {
			select {
			case <-c.Done():
				return
			default:
				_, message, err := clientConn.ReadMessage()
				if err != nil {
					if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						errChan <- fmt.Errorf("error reading from client: %v", err)
					}
					close(clientClosed)
					return
				}

				realtimeEvent := &dto.RealtimeEvent{}
				err = json.Unmarshal(message, realtimeEvent)
				if err != nil {
					errChan <- fmt.Errorf("error unmarshalling message: %v", err)
					return
				}

				if realtimeEvent.Type == dto.RealtimeEventTypeSessionUpdate {
					if realtimeEvent.Session != nil {
						if realtimeEvent.Session.Tools != nil {
							info.RealtimeTools = realtimeEvent.Session.Tools
						}
					}
				}

				textToken, audioToken, err := service.CountTokenRealtime(info, *realtimeEvent, info.UpstreamModelName)
				if err != nil {
					errChan <- fmt.Errorf("error counting text token: %v", err)
					return
				}
				common.LogInfo(c, fmt.Sprintf("type: %s, textToken: %d, audioToken: %d", realtimeEvent.Type, textToken, audioToken))
				localUsage.TotalTokens += textToken + audioToken
				localUsage.InputTokens += textToken + audioToken
				localUsage.InputTokenDetails.TextTokens += textToken
				localUsage.InputTokenDetails.AudioTokens += audioToken

				err = helper.WssString(c, targetConn, string(message))
				if err != nil {
					errChan <- fmt.Errorf("error writing to target: %v", err)
					return
				}

				select {
				case sendChan <- message:
				default:
				}
			}
		}
	})

	gopool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic in target reader: %v", r)
			}
		}()
		for {
			select {
			case <-c.Done():
				return
			default:
				_, message, err := targetConn.ReadMessage()
				if err != nil {
					if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
						errChan <- fmt.Errorf("error reading from target: %v", err)
					}
					close(targetClosed)
					return
				}
				info.SetFirstResponseTime()
				realtimeEvent := &dto.RealtimeEvent{}
				err = json.Unmarshal(message, realtimeEvent)
				if err != nil {
					errChan <- fmt.Errorf("error unmarshalling message: %v", err)
					return
				}

				if realtimeEvent.Type == dto.RealtimeEventTypeResponseDone {
					realtimeUsage := realtimeEvent.Response.Usage
					if realtimeUsage != nil {
						usage.TotalTokens += realtimeUsage.TotalTokens
						usage.InputTokens += realtimeUsage.InputTokens
						usage.OutputTokens += realtimeUsage.OutputTokens
						usage.InputTokenDetails.AudioTokens += realtimeUsage.InputTokenDetails.AudioTokens
						usage.InputTokenDetails.CachedTokens += realtimeUsage.InputTokenDetails.CachedTokens
						usage.InputTokenDetails.TextTokens += realtimeUsage.InputTokenDetails.TextTokens
						usage.OutputTokenDetails.AudioTokens += realtimeUsage.OutputTokenDetails.AudioTokens
						usage.OutputTokenDetails.TextTokens += realtimeUsage.OutputTokenDetails.TextTokens
						err := preConsumeUsage(c, info, usage, sumUsage)
						if err != nil {
							errChan <- fmt.Errorf("error consume usage: %v", err)
							return
						}
						// 本次计费完成，清除
						usage = &dto.RealtimeUsage{}

						localUsage = &dto.RealtimeUsage{}
					} else {
						textToken, audioToken, err := service.CountTokenRealtime(info, *realtimeEvent, info.UpstreamModelName)
						if err != nil {
							errChan <- fmt.Errorf("error counting text token: %v", err)
							return
						}
						common.LogInfo(c, fmt.Sprintf("type: %s, textToken: %d, audioToken: %d", realtimeEvent.Type, textToken, audioToken))
						localUsage.TotalTokens += textToken + audioToken
						info.IsFirstRequest = false
						localUsage.InputTokens += textToken + audioToken
						localUsage.InputTokenDetails.TextTokens += textToken
						localUsage.InputTokenDetails.AudioTokens += audioToken
						err = preConsumeUsage(c, info, localUsage, sumUsage)
						if err != nil {
							errChan <- fmt.Errorf("error consume usage: %v", err)
							return
						}
						// 本次计费完成，清除
						localUsage = &dto.RealtimeUsage{}
						// print now usage
					}
					//common.LogInfo(c, fmt.Sprintf("realtime streaming sumUsage: %v", sumUsage))
					//common.LogInfo(c, fmt.Sprintf("realtime streaming localUsage: %v", localUsage))
					//common.LogInfo(c, fmt.Sprintf("realtime streaming localUsage: %v", localUsage))

				} else if realtimeEvent.Type == dto.RealtimeEventTypeSessionUpdated || realtimeEvent.Type == dto.RealtimeEventTypeSessionCreated {
					realtimeSession := realtimeEvent.Session
					if realtimeSession != nil {
						// update audio format
						info.InputAudioFormat = common.GetStringIfEmpty(realtimeSession.InputAudioFormat, info.InputAudioFormat)
						info.OutputAudioFormat = common.GetStringIfEmpty(realtimeSession.OutputAudioFormat, info.OutputAudioFormat)
					}
				} else {
					textToken, audioToken, err := service.CountTokenRealtime(info, *realtimeEvent, info.UpstreamModelName)
					if err != nil {
						errChan <- fmt.Errorf("error counting text token: %v", err)
						return
					}
					common.LogInfo(c, fmt.Sprintf("type: %s, textToken: %d, audioToken: %d", realtimeEvent.Type, textToken, audioToken))
					localUsage.TotalTokens += textToken + audioToken
					localUsage.OutputTokens += textToken + audioToken
					localUsage.OutputTokenDetails.TextTokens += textToken
					localUsage.OutputTokenDetails.AudioTokens += audioToken
				}

				err = helper.WssString(c, clientConn, string(message))
				if err != nil {
					errChan <- fmt.Errorf("error writing to client: %v", err)
					return
				}

				select {
				case receiveChan <- message:
				default:
				}
			}
		}
	})

	select {
	case <-clientClosed:
	case <-targetClosed:
	case err := <-errChan:
		//return service.OpenAIErrorWrapper(err, "realtime_error", http.StatusInternalServerError), nil
		common.LogError(c, "realtime error: "+err.Error())
	case <-c.Done():
	}

	if usage.TotalTokens != 0 {
		_ = preConsumeUsage(c, info, usage, sumUsage)
	}

	if localUsage.TotalTokens != 0 {
		_ = preConsumeUsage(c, info, localUsage, sumUsage)
	}

	// check usage total tokens, if 0, use local usage

	return nil, sumUsage
}

func preConsumeUsage(ctx *gin.Context, info *relaycommon.RelayInfo, usage *dto.RealtimeUsage, totalUsage *dto.RealtimeUsage) error {
	if usage == nil || totalUsage == nil {
		return fmt.Errorf("invalid usage pointer")
	}

	totalUsage.TotalTokens += usage.TotalTokens
	totalUsage.InputTokens += usage.InputTokens
	totalUsage.OutputTokens += usage.OutputTokens
	totalUsage.InputTokenDetails.CachedTokens += usage.InputTokenDetails.CachedTokens
	totalUsage.InputTokenDetails.TextTokens += usage.InputTokenDetails.TextTokens
	totalUsage.InputTokenDetails.AudioTokens += usage.InputTokenDetails.AudioTokens
	totalUsage.OutputTokenDetails.TextTokens += usage.OutputTokenDetails.TextTokens
	totalUsage.OutputTokenDetails.AudioTokens += usage.OutputTokenDetails.AudioTokens
	// clear usage
	err := service.PreWssConsumeQuota(ctx, info, usage)
	return err
}

func OpenaiPseudoStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	simpleResponse, errResp := parseOpenAITextResponse(resp)
	if errResp != nil {
		return errResp, nil
	}

	fillOpenAIUsage(simpleResponse, info)

	helper.SetEventStreamHeaders(c)
	info.SetFirstResponseTime()

	streamResp := BuildStreamChunkFromTextResponse(simpleResponse)
	_ = helper.ObjectData(c, streamResp)
	if info.ShouldIncludeUsage {
		final := helper.GenerateFinalUsageResponse(helper.GetResponseID(c), common.GetTimestamp(), info.UpstreamModelName, simpleResponse.Usage)
		_ = helper.ObjectData(c, final)
	}
	helper.Done(c)
	return nil, &simpleResponse.Usage
}

func parseOpenAITextResponse(resp *http.Response) (*dto.OpenAITextResponse, *dto.OpenAIErrorWithStatusCode) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	if err = resp.Body.Close(); err != nil {
		return nil, service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError)
	}
	var res dto.OpenAITextResponse
	if err = common.DecodeJson(body, &res); err != nil {
		return nil, service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
	}
	if res.Error != nil && res.Error.Type != "" {
		return nil, &dto.OpenAIErrorWithStatusCode{Error: *res.Error, StatusCode: resp.StatusCode}
	}
	return &res, nil
}

func fillOpenAIUsage(resp *dto.OpenAITextResponse, info *relaycommon.RelayInfo) {
	if resp.Usage.TotalTokens != 0 && (resp.Usage.PromptTokens != 0 || resp.Usage.CompletionTokens != 0) {
		return
	}
	completionTokens := 0
	for _, choice := range resp.Choices {
		ctkm, _ := service.CountTextToken(choice.Message.StringContent()+choice.Message.ReasoningContent+choice.Message.Reasoning, info.UpstreamModelName)
		completionTokens += ctkm
	}
	resp.Usage = dto.Usage{
		PromptTokens:     info.PromptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      info.PromptTokens + completionTokens,
	}
}

func BuildStreamChunkFromTextResponse(simpleResponse *dto.OpenAITextResponse) dto.ChatCompletionsStreamResponse {
	var streamResp dto.ChatCompletionsStreamResponse
	streamResp.Id = simpleResponse.Id
	streamResp.Object = "chat.completion.chunk"
	streamResp.Created = simpleResponse.Created
	streamResp.Model = simpleResponse.Model
	streamResp.Choices = make([]dto.ChatCompletionsStreamResponseChoice, len(simpleResponse.Choices))
	for i, ch := range simpleResponse.Choices {
		var choice dto.ChatCompletionsStreamResponseChoice
		choice.Index = ch.Index
		finishReason := ch.FinishReason
		choice.FinishReason = &finishReason
		choice.Delta.Role = "assistant"
		choice.Delta.SetContentString(ch.Message.StringContent())
		streamResp.Choices[i] = choice
	}
	return streamResp
}
