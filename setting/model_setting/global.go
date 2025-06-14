package model_setting

import (
	"veloera/setting/config"
)

type GlobalSettings struct {
	PassThroughRequestEnabled    bool `json:"pass_through_request_enabled"`
	HideUpstreamErrorEnabled     bool `json:"hide_upstream_error_enabled"`
	BlockBrowserExtensionEnabled bool `json:"block_browser_extension_enabled"`
}

// 默认配置
var defaultOpenaiSettings = GlobalSettings{
	PassThroughRequestEnabled:    false,
	HideUpstreamErrorEnabled:     false,
	BlockBrowserExtensionEnabled: false,
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
