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
package ollama

import "veloera/dto"

type OllamaRequest struct {
	Model            string                `json:"model,omitempty"`
	Messages         []dto.Message         `json:"messages,omitempty"`
	Stream           bool                  `json:"stream,omitempty"`
	Temperature      *float64              `json:"temperature,omitempty"`
	Seed             float64               `json:"seed,omitempty"`
	Topp             float64               `json:"top_p,omitempty"`
	TopK             int                   `json:"top_k,omitempty"`
	Stop             any                   `json:"stop,omitempty"`
	MaxTokens        uint                  `json:"max_tokens,omitempty"`
	Tools            []dto.ToolCallRequest `json:"tools,omitempty"`
	ResponseFormat   any                   `json:"response_format,omitempty"`
	FrequencyPenalty float64               `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64               `json:"presence_penalty,omitempty"`
	Suffix           any                   `json:"suffix,omitempty"`
	StreamOptions    *dto.StreamOptions    `json:"stream_options,omitempty"`
	Prompt           any                   `json:"prompt,omitempty"`
}

type Options struct {
	Seed             int      `json:"seed,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	TopK             int      `json:"top_k,omitempty"`
	TopP             float64  `json:"top_p,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64  `json:"presence_penalty,omitempty"`
	NumPredict       int      `json:"num_predict,omitempty"`
	NumCtx           int      `json:"num_ctx,omitempty"`
}

type OllamaEmbeddingRequest struct {
	Model   string   `json:"model,omitempty"`
	Input   []string `json:"input"`
	Options *Options `json:"options,omitempty"`
}

type OllamaEmbeddingResponse struct {
	Error     string      `json:"error,omitempty"`
	Model     string      `json:"model"`
	Embedding [][]float64 `json:"embeddings,omitempty"`
}
