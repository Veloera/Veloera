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
package dify

import "veloera/dto"

type DifyChatRequest struct {
	Inputs           map[string]interface{} `json:"inputs"`
	Query            string                 `json:"query"`
	ResponseMode     string                 `json:"response_mode"`
	User             string                 `json:"user"`
	AutoGenerateName bool                   `json:"auto_generate_name"`
	Files            []DifyFile             `json:"files"`
}

type DifyFile struct {
	Type         string `json:"type"`
	TransferMode string `json:"transfer_mode"`
	URL          string `json:"url,omitempty"`
	UploadFileId string `json:"upload_file_id,omitempty"`
}

type DifyMetaData struct {
	Usage dto.Usage `json:"usage"`
}

type DifyData struct {
	WorkflowId string `json:"workflow_id"`
	NodeId     string `json:"node_id"`
	NodeType   string `json:"node_type"`
	Status     string `json:"status"`
}

type DifyChatCompletionResponse struct {
	ConversationId string       `json:"conversation_id"`
	Answer         string       `json:"answer"`
	CreateAt       int64        `json:"create_at"`
	MetaData       DifyMetaData `json:"metadata"`
}

type DifyChunkChatCompletionResponse struct {
	Event          string       `json:"event"`
	ConversationId string       `json:"conversation_id"`
	Answer         string       `json:"answer"`
	Data           DifyData     `json:"data"`
	MetaData       DifyMetaData `json:"metadata"`
}
