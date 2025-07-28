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

// GeminiSettings 定义Gemini模型的配置
type GeminiSettings struct {
	SafetySettings                        map[string]string `json:"safety_settings"`
	VersionSettings                       map[string]string `json:"version_settings"`
	SupportedImagineModels                []string          `json:"supported_imagine_models"`
	ThinkingAdapterEnabled                bool              `json:"thinking_adapter_enabled"`
	ThinkingAdapterBudgetTokensPercentage float64           `json:"thinking_adapter_budget_tokens_percentage"`
	ModelsSupportedThinkingBudget         []string          `json:"models_supported_thinking_budget"`
}

// 默认配置
var defaultGeminiSettings = GeminiSettings{
	SafetySettings: map[string]string{
		"default":                         "OFF",
		"HARM_CATEGORY_HARASSMENT":        "BLOCK_NONE",
		"HARM_CATEGORY_HATE_SPEECH":       "BLOCK_NONE",
		"HARM_CATEGORY_SEXUALLY_EXPLICIT": "BLOCK_NONE",
		"HARM_CATEGORY_DANGEROUS_CONTENT": "BLOCK_NONE",
	},
	VersionSettings: map[string]string{
		"default":        "v1beta",
		"gemini-1.0-pro": "v1beta",
	},
	SupportedImagineModels: []string{
		"gemini-2.0-flash-exp-image-generation",
		"gemini-2.0-flash-exp",
	},
	ThinkingAdapterEnabled:                false,
	ThinkingAdapterBudgetTokensPercentage: 0.6,
	ModelsSupportedThinkingBudget: []string{
		"gemini-2.5-flash-preview-05-20",
		"gemini-2.5-flash-preview-04-17",
		"gemini-2.5-pro-preview-06-05",
		"gemini-2.5-pro",
	},
}

// 全局实例
var geminiSettings = defaultGeminiSettings

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("gemini", &geminiSettings)
}

// GetGeminiSettings 获取Gemini配置
func GetGeminiSettings() *GeminiSettings {
	return &geminiSettings
}

// GetGeminiSafetySetting 获取安全设置
func GetGeminiSafetySetting(key string) string {
	if value, ok := geminiSettings.SafetySettings[key]; ok {
		return value
	}
	return geminiSettings.SafetySettings["default"]
}

// GetGeminiVersionSetting 获取版本设置
func GetGeminiVersionSetting(key string) string {
	if value, ok := geminiSettings.VersionSettings[key]; ok {
		return value
	}
	return geminiSettings.VersionSettings["default"]
}

func IsGeminiModelSupportImagine(model string) bool {
	for _, v := range geminiSettings.SupportedImagineModels {
		if v == model {
			return true
		}
	}
	return false
}
