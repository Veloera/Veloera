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

type ZhipuMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ZhipuRequest struct {
	Prompt      []ZhipuMessage `json:"prompt"`
	Temperature *float64       `json:"temperature,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
	RequestId   string         `json:"request_id,omitempty"`
	Incremental bool           `json:"incremental,omitempty"`
}

type ZhipuResponseData struct {
	TaskId     string         `json:"task_id"`
	RequestId  string         `json:"request_id"`
	TaskStatus string         `json:"task_status"`
	Choices    []ZhipuMessage `json:"choices"`
	dto.Usage  `json:"usage"`
}

type ZhipuResponse struct {
	Code    int               `json:"code"`
	Msg     string            `json:"msg"`
	Success bool              `json:"success"`
	Data    ZhipuResponseData `json:"data"`
}

type ZhipuStreamMetaResponse struct {
	RequestId  string `json:"request_id"`
	TaskId     string `json:"task_id"`
	TaskStatus string `json:"task_status"`
	dto.Usage  `json:"usage"`
}

type zhipuTokenData struct {
	Token      string
	ExpiryTime time.Time
}
