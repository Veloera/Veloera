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
package service

import (
	"fmt"
	"net/http"
	"strings"
	"veloera/common"
	"veloera/dto"
	"veloera/model"
	"veloera/setting/operation_setting"
)

func formatNotifyType(channelId int, status int) string {
	return fmt.Sprintf("%s_%d_%d", dto.NotifyTypeChannelUpdate, channelId, status)
}

// disable & notify
func DisableChannel(channelId int, channelName string, reason string) {
	success := model.UpdateChannelStatusById(channelId, common.ChannelStatusAutoDisabled, reason)
	if success {
		subject := fmt.Sprintf("通道「%s」（#%d）已被禁用", channelName, channelId)
		content := fmt.Sprintf("通道「%s」（#%d）已被禁用，原因：%s", channelName, channelId, reason)
		NotifyRootUser(formatNotifyType(channelId, common.ChannelStatusAutoDisabled), subject, content)
	}
}

func EnableChannel(channelId int, channelName string) {
	success := model.UpdateChannelStatusById(channelId, common.ChannelStatusEnabled, "")
	if success {
		subject := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
		content := fmt.Sprintf("通道「%s」（#%d）已被启用", channelName, channelId)
		NotifyRootUser(formatNotifyType(channelId, common.ChannelStatusEnabled), subject, content)
	}
}

func ShouldDisableChannel(channelType int, err *dto.OpenAIErrorWithStatusCode) bool {
	if !common.AutomaticDisableChannelEnabled {
		return false
	}
	if err == nil {
		return false
	}
	if err.LocalError {
		return false
	}
	if err.StatusCode == http.StatusUnauthorized {
		return true
	}
	if err.StatusCode == http.StatusForbidden {
		switch channelType {
		case common.ChannelTypeGemini:
			return true
		}
	}
	switch err.Error.Code {
	case "invalid_api_key":
		return true
	case "account_deactivated":
		return true
	case "billing_not_active":
		return true
	}
	switch err.Error.Type {
	case "insufficient_quota":
		return true
	case "insufficient_user_quota":
		return true
	// https://docs.anthropic.com/claude/reference/errors
	case "authentication_error":
		return true
	case "permission_error":
		return true
	case "forbidden":
		return true
	}

	lowerMessage := strings.ToLower(err.Error.Message)
	search, _ := AcSearch(lowerMessage, operation_setting.AutomaticDisableKeywords, true)
	if search {
		return true
	}

	return false
}

func ShouldEnableChannel(err error, openaiWithStatusErr *dto.OpenAIErrorWithStatusCode, status int) bool {
	if !common.AutomaticEnableChannelEnabled {
		return false
	}
	if err != nil {
		return false
	}
	if openaiWithStatusErr != nil {
		return false
	}
	if status != common.ChannelStatusAutoDisabled {
		return false
	}
	return true
}
