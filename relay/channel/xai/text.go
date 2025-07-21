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
package xai

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"veloera/common"
	"veloera/dto"
	"veloera/relay/channel/openai"
	relaycommon "veloera/relay/common"
	"veloera/relay/helper"
	"veloera/service"
)

func streamResponseXAI2OpenAI(xAIResp *dto.ChatCompletionsStreamResponse, usage *dto.Usage) *dto.ChatCompletionsStreamResponse {
	if xAIResp == nil {
		return nil
	}
	if xAIResp.Usage != nil {
		xAIResp.Usage.CompletionTokens = usage.CompletionTokens
	}
	openAIResp := &dto.ChatCompletionsStreamResponse{
		Id:      xAIResp.Id,
		Object:  xAIResp.Object,
		Created: xAIResp.Created,
		Model:   xAIResp.Model,
		Choices: xAIResp.Choices,
		Usage:   xAIResp.Usage,
	}

	return openAIResp
}

func xAIStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	usage := &dto.Usage{}
	var responseTextBuilder strings.Builder
	var toolCount int
	var containStreamUsage bool

	helper.SetEventStreamHeaders(c)

	helper.StreamScannerHandler(c, resp, info, func(data string) bool {
		var xAIResp *dto.ChatCompletionsStreamResponse
		err := json.Unmarshal([]byte(data), &xAIResp)
		if err != nil {
			common.SysError("error unmarshalling stream response: " + err.Error())
			return true
		}

		// 把 xAI 的usage转换为 OpenAI 的usage
		if xAIResp.Usage != nil {
			containStreamUsage = true
			usage.PromptTokens = xAIResp.Usage.PromptTokens
			usage.TotalTokens = xAIResp.Usage.TotalTokens
			usage.CompletionTokens = usage.TotalTokens - usage.PromptTokens
		}

		openaiResponse := streamResponseXAI2OpenAI(xAIResp, usage)
		_ = openai.ProcessStreamResponse(*openaiResponse, &responseTextBuilder, &toolCount)
		err = helper.ObjectData(c, openaiResponse)
		if err != nil {
			common.SysError(err.Error())
		}
		return true
	})

	if !containStreamUsage {
		usage, _ = service.ResponseText2Usage(responseTextBuilder.String(), info.UpstreamModelName, info.PromptTokens)
		usage.CompletionTokens += toolCount * 7
	}

	helper.Done(c)
	err := resp.Body.Close()
	if err != nil {
		//return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
		common.SysError("close_response_body_failed: " + err.Error())
	}
	return nil, usage
}

func xAIHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	var response *dto.TextResponse
	err = common.DecodeJson(responseBody, &response)
	if err != nil {
		common.SysError("error unmarshalling stream response: " + err.Error())
		return nil, nil
	}
	response.Usage.CompletionTokens = response.Usage.TotalTokens - response.Usage.PromptTokens
	response.Usage.CompletionTokenDetails.TextTokens = response.Usage.CompletionTokens - response.Usage.CompletionTokenDetails.ReasoningTokens

	// new body
	encodeJson, err := common.EncodeJson(response)
	if err != nil {
		common.SysError("error marshalling stream response: " + err.Error())
		return nil, nil
	}

	// set new body
	resp.Body = io.NopCloser(bytes.NewBuffer(encodeJson))

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

	return nil, &response.Usage
}
