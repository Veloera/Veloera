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
package zhipu

import (
	"time"
	"veloera/dto"
)

//	type ZhipuMessage struct {
//		Role       string `json:"role,omitempty"`
//		Content    string `json:"content,omitempty"`
//		ToolCalls  any    `json:"tool_calls,omitempty"`
//		ToolCallId any    `json:"tool_call_id,omitempty"`
//	}
//
//	type ZhipuRequest struct {
//		Model       string         `json:"model"`
//		Stream      bool           `json:"stream,omitempty"`
//		Messages    []ZhipuMessage `json:"messages"`
//		Temperature float64        `json:"temperature,omitempty"`
//		TopP        float64        `json:"top_p,omitempty"`
//		MaxTokens   int            `json:"max_tokens,omitempty"`
//		Stop        []string       `json:"stop,omitempty"`
//		RequestId   string         `json:"request_id,omitempty"`
//		Tools       any            `json:"tools,omitempty"`
//		ToolChoice  any            `json:"tool_choice,omitempty"`
//	}
//
//	type ZhipuTextResponseChoice struct {
//		Index        int `json:"index"`
//		ZhipuMessage `json:"message"`
//		FinishReason string `json:"finish_reason"`
//	}
type ZhipuResponse struct {
	Id                  string                         `json:"id"`
	Created             int64                          `json:"created"`
	Model               string                         `json:"model"`
	TextResponseChoices []dto.OpenAITextResponseChoice `json:"choices"`
	Usage               dto.Usage                      `json:"usage"`
	Error               dto.OpenAIError                `json:"error"`
}

//
//type ZhipuStreamResponseChoice struct {
//	Index        int          `json:"index,omitempty"`
//	Delta        ZhipuMessage `json:"delta"`
//	FinishReason *string      `json:"finish_reason,omitempty"`
//}

type ZhipuStreamResponse struct {
	Id      string                                    `json:"id"`
	Created int64                                     `json:"created"`
	Choices []dto.ChatCompletionsStreamResponseChoice `json:"choices"`
	Usage   dto.Usage                                 `json:"usage"`
}

type tokenData struct {
	Token      string
	ExpiryTime time.Time
}
