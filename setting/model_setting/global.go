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
package model_setting

import (
	"veloera/setting/config"
)

type GlobalSettings struct {
	PassThroughRequestEnabled    bool   `json:"pass_through_request_enabled"`
	HideUpstreamErrorEnabled     bool   `json:"hide_upstream_error_enabled"`
	BlockBrowserExtensionEnabled bool   `json:"block_browser_extension_enabled"`
	RateLimitExemptEnabled       bool   `json:"rate_limit_exempt_enabled"`
	RateLimitExemptGroup         string `json:"rate_limit_exempt_group"`
	SafeCheckExemptEnabled       bool   `json:"safe_check_exempt_enabled"`
	SafeCheckExemptGroup         string `json:"safe_check_exempt_group"`
}

// 默认配置
var defaultOpenaiSettings = GlobalSettings{
	PassThroughRequestEnabled:    false,
	HideUpstreamErrorEnabled:     false,
	BlockBrowserExtensionEnabled: false,
	RateLimitExemptEnabled:       false,
	RateLimitExemptGroup:         "bulk-ok",
	SafeCheckExemptEnabled:       false,
	SafeCheckExemptGroup:         "nsfw-ok",
}

// 全局实例
var globalSettings = defaultOpenaiSettings

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("global", &globalSettings)
}

func GetGlobalSettings() *GlobalSettings {
	return &globalSettings
}

func ShouldBypassRateLimit(group string) bool {
	return globalSettings.RateLimitExemptEnabled && group == globalSettings.RateLimitExemptGroup
}

func ShouldBypassSafeCheck(group string) bool {
	return globalSettings.SafeCheckExemptEnabled && group == globalSettings.SafeCheckExemptGroup
}
