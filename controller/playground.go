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
package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
	"veloera/common"
	"veloera/constant"
	"veloera/dto"
	"veloera/middleware"
	"veloera/model"
	"veloera/service"
	"veloera/setting"
)

func Playground(c *gin.Context) {
	var openaiErr *dto.OpenAIErrorWithStatusCode

	defer func() {
		if openaiErr != nil {
			c.JSON(openaiErr.StatusCode, gin.H{
				"error": openaiErr.Error,
			})
		}
	}()

	useAccessToken := c.GetBool("use_access_token")
	if useAccessToken {
		openaiErr = service.OpenAIErrorWrapperLocal(errors.New("暂不支持使用 access token"), "access_token_not_supported", http.StatusBadRequest)
		return
	}

	playgroundRequest := &dto.PlayGroundRequest{}
	err := common.UnmarshalBodyReusable(c, playgroundRequest)
	if err != nil {
		openaiErr = service.OpenAIErrorWrapperLocal(err, "unmarshal_request_failed", http.StatusBadRequest)
		return
	}

	if playgroundRequest.Model == "" {
		openaiErr = service.OpenAIErrorWrapperLocal(errors.New("请选择模型"), "model_required", http.StatusBadRequest)
		return
	}
	c.Set("original_model", playgroundRequest.Model)
	group := playgroundRequest.Group
	userGroup := c.GetString("group")

	if group == "" {
		group = userGroup
	} else {
		if !setting.GroupInUserUsableGroups(group) && group != userGroup {
			openaiErr = service.OpenAIErrorWrapperLocal(errors.New("无权访问该分组"), "group_not_allowed", http.StatusForbidden)
			return
		}
		c.Set("group", group)
	}
	c.Set("token_name", "playground-"+group)

	// Handle model prefix for channel selection (similar to middleware/distributor.go)
	originalModel := playgroundRequest.Model
	modelToQuery := playgroundRequest.Model
	var channel *model.Channel

	// Check if the model has a prefix that should be used for routing
	modelPrefix := ""
	for prefix := range middleware.GetPrefixChannels(group) {
		if prefix != "" && strings.HasPrefix(originalModel, prefix) {
			modelPrefix = prefix
			// Strip the prefix for channel selection
			modelToQuery = strings.TrimPrefix(originalModel, prefix)
			break
		}
	}

	// Select channel based on whether we found a prefix
	if modelPrefix != "" {
		// Use prefix-based channel selection
		channel, err = middleware.SelectChannelByPrefix(group, modelPrefix, modelToQuery)
	} else {
		// Use normal channel selection
		channel, err = model.CacheGetRandomSatisfiedChannel(group, modelToQuery, 0)
	}

	if err != nil {
		message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", group, originalModel)
		openaiErr = service.OpenAIErrorWrapperLocal(errors.New(message), "get_playground_channel_failed", http.StatusInternalServerError)
		return
	}
	middleware.SetupContextForSelectedChannel(c, channel, modelToQuery)
	c.Set(constant.ContextKeyRequestStartTime, time.Now())
	Relay(c)
}
