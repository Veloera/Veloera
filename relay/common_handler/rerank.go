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
package common_handler

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"veloera/common"
	"veloera/dto"
	"veloera/relay/channel/xinference"
	relaycommon "veloera/relay/common"
	"veloera/service"
)

func RerankHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	if common.DebugEnabled {
		println("reranker response body: ", string(responseBody))
	}
	var jinaResp dto.RerankResponse
	if info.ChannelType == common.ChannelTypeXinference {
		var xinRerankResponse xinference.XinRerankResponse
		err = common.DecodeJson(responseBody, &xinRerankResponse)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
		}
		jinaRespResults := make([]dto.RerankResponseResult, len(xinRerankResponse.Results))
		for i, result := range xinRerankResponse.Results {
			respResult := dto.RerankResponseResult{
				Index:          result.Index,
				RelevanceScore: result.RelevanceScore,
			}
			if info.ReturnDocuments {
				var document any
				if result.Document == "" {
					document = info.Documents[result.Index]
				} else {
					document = result.Document
				}
				respResult.Document = document
			}
			jinaRespResults[i] = respResult
		}
		jinaResp = dto.RerankResponse{
			Results: jinaRespResults,
			Usage: dto.Usage{
				PromptTokens: info.PromptTokens,
				TotalTokens:  info.PromptTokens,
			},
		}
	} else {
		err = common.DecodeJson(responseBody, &jinaResp)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
		}
		jinaResp.Usage.PromptTokens = jinaResp.Usage.TotalTokens
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.JSON(http.StatusOK, jinaResp)
	return nil, &jinaResp.Usage
}
