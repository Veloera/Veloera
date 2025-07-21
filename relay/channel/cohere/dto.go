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
package cohere

import "veloera/dto"

type CohereRequest struct {
	Model       string        `json:"model"`
	ChatHistory []ChatHistory `json:"chat_history"`
	Message     string        `json:"message"`
	Stream      bool          `json:"stream"`
	MaxTokens   int           `json:"max_tokens"`
	SafetyMode  string        `json:"safety_mode,omitempty"`
}

type ChatHistory struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

type CohereResponse struct {
	IsFinished   bool                  `json:"is_finished"`
	EventType    string                `json:"event_type"`
	Text         string                `json:"text,omitempty"`
	FinishReason string                `json:"finish_reason,omitempty"`
	Response     *CohereResponseResult `json:"response"`
}

type CohereResponseResult struct {
	ResponseId   string     `json:"response_id"`
	FinishReason string     `json:"finish_reason,omitempty"`
	Text         string     `json:"text"`
	Meta         CohereMeta `json:"meta"`
}

type CohereRerankRequest struct {
	Documents       []any  `json:"documents"`
	Query           string `json:"query"`
	Model           string `json:"model"`
	TopN            int    `json:"top_n"`
	ReturnDocuments bool   `json:"return_documents"`
}

type CohereRerankResponseResult struct {
	Results []dto.RerankResponseResult `json:"results"`
	Meta    CohereMeta                 `json:"meta"`
}

type CohereMeta struct {
	//Tokens CohereTokens `json:"tokens"`
	BilledUnits CohereBilledUnits `json:"billed_units"`
}

type CohereBilledUnits struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type CohereTokens struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
