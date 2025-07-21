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
package cloudflare

import "veloera/dto"

type CfRequest struct {
	Messages    []dto.Message `json:"messages,omitempty"`
	Lora        string        `json:"lora,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Prompt      string        `json:"prompt,omitempty"`
	Raw         bool          `json:"raw,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
}

type CfAudioResponse struct {
	Result CfSTTResult `json:"result"`
}

type CfSTTResult struct {
	Text string `json:"text"`
}
