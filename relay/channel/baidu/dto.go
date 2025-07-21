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
package baidu

import (
	"time"
	"veloera/dto"
)

type BaiduMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type BaiduChatRequest struct {
	Messages        []BaiduMessage `json:"messages"`
	Temperature     *float64       `json:"temperature,omitempty"`
	TopP            float64        `json:"top_p,omitempty"`
	PenaltyScore    float64        `json:"penalty_score,omitempty"`
	Stream          bool           `json:"stream,omitempty"`
	System          string         `json:"system,omitempty"`
	DisableSearch   bool           `json:"disable_search,omitempty"`
	EnableCitation  bool           `json:"enable_citation,omitempty"`
	MaxOutputTokens *int           `json:"max_output_tokens,omitempty"`
	UserId          string         `json:"user_id,omitempty"`
}

type Error struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

type BaiduChatResponse struct {
	Id               string    `json:"id"`
	Object           string    `json:"object"`
	Created          int64     `json:"created"`
	Result           string    `json:"result"`
	IsTruncated      bool      `json:"is_truncated"`
	NeedClearHistory bool      `json:"need_clear_history"`
	Usage            dto.Usage `json:"usage"`
	Error
}

type BaiduChatStreamResponse struct {
	BaiduChatResponse
	SentenceId int  `json:"sentence_id"`
	IsEnd      bool `json:"is_end"`
}

type BaiduEmbeddingRequest struct {
	Input []string `json:"input"`
}

type BaiduEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type BaiduEmbeddingResponse struct {
	Id      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Data    []BaiduEmbeddingData `json:"data"`
	Usage   dto.Usage            `json:"usage"`
	Error
}

type BaiduAccessToken struct {
	AccessToken      string    `json:"access_token"`
	Error            string    `json:"error,omitempty"`
	ErrorDescription string    `json:"error_description,omitempty"`
	ExpiresIn        int64     `json:"expires_in,omitempty"`
	ExpiresAt        time.Time `json:"-"`
}

type BaiduTokenResponse struct {
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}
