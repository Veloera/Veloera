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
	"encoding/json"
	"strings"
	"veloera/common"
	"veloera/dto"
	relaycommon "veloera/relay/common"
	relayconstant "veloera/relay/constant"
	"veloera/relay/helper"
	"veloera/service"

	"github.com/gin-gonic/gin"
)

// 辅助函数
func handleStreamFormat(c *gin.Context, info *relaycommon.RelayInfo, data string, forceFormat bool, thinkToContent bool) error {
	info.SendResponseCount++
	switch info.RelayFormat {
	case relaycommon.RelayFormatOpenAI:
		return sendStreamData(c, info, data, forceFormat, thinkToContent)
	case relaycommon.RelayFormatClaude:
		return handleClaudeFormat(c, data, info)
	}
	return nil
}

func handleClaudeFormat(c *gin.Context, data string, info *relaycommon.RelayInfo) error {
	var streamResponse dto.ChatCompletionsStreamResponse
	if err := json.Unmarshal(common.StringToByteSlice(data), &streamResponse); err != nil {
		return err
	}

	if streamResponse.Usage != nil {
		info.ClaudeConvertInfo.Usage = streamResponse.Usage
	}
	claudeResponses := service.StreamResponseOpenAI2Claude(&streamResponse, info)
	for _, resp := range claudeResponses {
		helper.ClaudeData(c, *resp)
	}
	return nil
}

func ProcessStreamResponse(streamResponse dto.ChatCompletionsStreamResponse, responseTextBuilder *strings.Builder, toolCount *int) error {
	for _, choice := range streamResponse.Choices {
		responseTextBuilder.WriteString(choice.Delta.GetContentString())
		responseTextBuilder.WriteString(choice.Delta.GetReasoningContent())
		if choice.Delta.ToolCalls != nil {
			if len(choice.Delta.ToolCalls) > *toolCount {
				*toolCount = len(choice.Delta.ToolCalls)
			}
			for _, tool := range choice.Delta.ToolCalls {
				responseTextBuilder.WriteString(tool.Function.Name)
				responseTextBuilder.WriteString(tool.Function.Arguments)
			}
		}
	}
	return nil
}

func processTokens(relayMode int, streamItems []string, responseTextBuilder *strings.Builder, toolCount *int) error {
	streamResp := "[" + strings.Join(streamItems, ",") + "]"

	switch relayMode {
	case relayconstant.RelayModeChatCompletions:
		return processChatCompletions(streamResp, streamItems, responseTextBuilder, toolCount)
	case relayconstant.RelayModeCompletions:
		return processCompletions(streamResp, streamItems, responseTextBuilder)
	}
	return nil
}

func processChatCompletions(streamResp string, streamItems []string, responseTextBuilder *strings.Builder, toolCount *int) error {
	var streamResponses []dto.ChatCompletionsStreamResponse
	if err := json.Unmarshal(common.StringToByteSlice(streamResp), &streamResponses); err != nil {
		// 一次性解析失败，逐个解析
		common.SysError("error unmarshalling stream response: " + err.Error())
		for _, item := range streamItems {
			var streamResponse dto.ChatCompletionsStreamResponse
			if err := json.Unmarshal(common.StringToByteSlice(item), &streamResponse); err != nil {
				return err
			}
			if err := ProcessStreamResponse(streamResponse, responseTextBuilder, toolCount); err != nil {
				common.SysError("error processing stream response: " + err.Error())
			}
		}
		return nil
	}

	// 批量处理所有响应
	for _, streamResponse := range streamResponses {
		for _, choice := range streamResponse.Choices {
			responseTextBuilder.WriteString(choice.Delta.GetContentString())
			responseTextBuilder.WriteString(choice.Delta.GetReasoningContent())
			if choice.Delta.ToolCalls != nil {
				if len(choice.Delta.ToolCalls) > *toolCount {
					*toolCount = len(choice.Delta.ToolCalls)
				}
				for _, tool := range choice.Delta.ToolCalls {
					responseTextBuilder.WriteString(tool.Function.Name)
					responseTextBuilder.WriteString(tool.Function.Arguments)
				}
			}
		}
	}
	return nil
}

func processCompletions(streamResp string, streamItems []string, responseTextBuilder *strings.Builder) error {
	var streamResponses []dto.CompletionsStreamResponse
	if err := json.Unmarshal(common.StringToByteSlice(streamResp), &streamResponses); err != nil {
		// 一次性解析失败，逐个解析
		common.SysError("error unmarshalling stream response: " + err.Error())
		for _, item := range streamItems {
			var streamResponse dto.CompletionsStreamResponse
			if err := json.Unmarshal(common.StringToByteSlice(item), &streamResponse); err != nil {
				continue
			}
			for _, choice := range streamResponse.Choices {
				responseTextBuilder.WriteString(choice.Text)
			}
		}
		return nil
	}

	// 批量处理所有响应
	for _, streamResponse := range streamResponses {
		for _, choice := range streamResponse.Choices {
			responseTextBuilder.WriteString(choice.Text)
		}
	}
	return nil
}

func handleLastResponse(lastStreamData string, responseId *string, createAt *int64,
	systemFingerprint *string, model *string, usage **dto.Usage,
	containStreamUsage *bool, info *relaycommon.RelayInfo,
	shouldSendLastResp *bool) error {

	var lastStreamResponse dto.ChatCompletionsStreamResponse
	if err := json.Unmarshal(common.StringToByteSlice(lastStreamData), &lastStreamResponse); err != nil {
		return err
	}

	*responseId = lastStreamResponse.Id
	*createAt = lastStreamResponse.Created
	*systemFingerprint = lastStreamResponse.GetSystemFingerprint()
	*model = lastStreamResponse.Model

	if service.ValidUsage(lastStreamResponse.Usage) {
		*containStreamUsage = true
		*usage = lastStreamResponse.Usage
		if !info.ShouldIncludeUsage {
			*shouldSendLastResp = false
		}
	}

	return nil
}

func handleFinalResponse(c *gin.Context, info *relaycommon.RelayInfo, lastStreamData string,
	responseId string, createAt int64, model string, systemFingerprint string,
	usage *dto.Usage, containStreamUsage bool) {

	switch info.RelayFormat {
	case relaycommon.RelayFormatOpenAI:
		if info.ShouldIncludeUsage && !containStreamUsage {
			response := helper.GenerateFinalUsageResponse(responseId, createAt, model, *usage)
			response.SetSystemFingerprint(systemFingerprint)
			helper.ObjectData(c, response)
		}
		helper.Done(c)

	case relaycommon.RelayFormatClaude:
		info.ClaudeConvertInfo.Done = true
		var streamResponse dto.ChatCompletionsStreamResponse
		if err := json.Unmarshal(common.StringToByteSlice(lastStreamData), &streamResponse); err != nil {
			common.SysError("error unmarshalling stream response: " + err.Error())
			return
		}

		info.ClaudeConvertInfo.Usage = usage

		claudeResponses := service.StreamResponseOpenAI2Claude(&streamResponse, info)
		for _, resp := range claudeResponses {
			helper.ClaudeData(c, *resp)
		}
	}
}
