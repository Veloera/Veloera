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
package xunfei

import "veloera/dto"

type XunfeiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type XunfeiChatRequest struct {
	Header struct {
		AppId string `json:"app_id"`
	} `json:"header"`
	Parameter struct {
		Chat struct {
			Domain      string   `json:"domain,omitempty"`
			Temperature *float64 `json:"temperature,omitempty"`
			TopK        int      `json:"top_k,omitempty"`
			MaxTokens   uint     `json:"max_tokens,omitempty"`
			Auditing    bool     `json:"auditing,omitempty"`
		} `json:"chat"`
	} `json:"parameter"`
	Payload struct {
		Message struct {
			Text []XunfeiMessage `json:"text"`
		} `json:"message"`
	} `json:"payload"`
}

type XunfeiChatResponseTextItem struct {
	Content string `json:"content"`
	Role    string `json:"role"`
	Index   int    `json:"index"`
}

type XunfeiChatResponse struct {
	Header struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Sid     string `json:"sid"`
		Status  int    `json:"status"`
	} `json:"header"`
	Payload struct {
		Choices struct {
			Status int                          `json:"status"`
			Seq    int                          `json:"seq"`
			Text   []XunfeiChatResponseTextItem `json:"text"`
		} `json:"choices"`
		Usage struct {
			//Text struct {
			//	QuestionTokens   string `json:"question_tokens"`
			//	PromptTokens     string `json:"prompt_tokens"`
			//	CompletionTokens string `json:"completion_tokens"`
			//	TotalTokens      string `json:"total_tokens"`
			//} `json:"text"`
			Text dto.Usage `json:"text"`
		} `json:"usage"`
	} `json:"payload"`
}
