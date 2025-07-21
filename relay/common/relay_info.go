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
package common

import (
	"strings"
	"time"
	"veloera/common"
	"veloera/constant"
	"veloera/dto"
	relayconstant "veloera/relay/constant"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ThinkingContentInfo struct {
	IsFirstThinkingContent  bool
	SendLastThinkingContent bool
	HasSentThinkingContent  bool
}

const (
	LastMessageTypeNone     = "none"
	LastMessageTypeText     = "text"
	LastMessageTypeTools    = "tools"
	LastMessageTypeThinking = "thinking"
)

type ClaudeConvertInfo struct {
	LastMessagesType string
	Index            int
	Usage            *dto.Usage
	FinishReason     string
	Done             bool
}

const (
	RelayFormatOpenAI = "openai"
	RelayFormatClaude = "claude"
)

type RerankerInfo struct {
	Documents       []any
	ReturnDocuments bool
}

type RelayInfo struct {
	ChannelType       int
	ChannelId         int
	TokenId           int
	TokenKey          string
	UserId            int
	Group             string
	TokenUnlimited    bool
	StartTime         time.Time
	FirstResponseTime time.Time
	isFirstResponse   bool
	//SendLastReasoningResponse bool
	ApiType           int
	IsStream          bool
	IsPlayground      bool
	UsePrice          bool
	RelayMode         int
	UpstreamModelName string
	OriginModelName   string
	//RecodeModelName      string
	RequestURLPath       string
	ApiVersion           string
	PromptTokens         int
	ApiKey               string
	Organization         string
	BaseUrl              string
	SupportStreamOptions bool
	ShouldIncludeUsage   bool
	IsModelMapped        bool
	ClientWs             *websocket.Conn
	TargetWs             *websocket.Conn
	InputAudioFormat     string
	OutputAudioFormat    string
	RealtimeTools        []dto.RealTimeTool
	IsFirstRequest       bool
	AudioUsage           bool
	ReasoningEffort      string
	ChannelSetting       map[string]interface{}
	ParamOverride        map[string]interface{}
	UserSetting          map[string]interface{}
	UserEmail            string
	UserQuota            int
	RelayFormat          string
	SendResponseCount    int
	ChannelCreateTime    int64
	PromptMessages       interface{}            // 保存请求的消息内容
	Other                map[string]interface{} // 用于存储额外信息，如输入输出内容
	ThinkingContentInfo
	*ClaudeConvertInfo
	*RerankerInfo
}

// 定义支持流式选项的通道类型
var streamSupportedChannels = map[int]bool{
	common.ChannelTypeOpenAI:     true,
	common.ChannelTypeAnthropic:  true,
	common.ChannelTypeAws:        true,
	common.ChannelTypeGemini:     true,
	common.ChannelCloudflare:     true,
	common.ChannelTypeAzure:      true,
	common.ChannelTypeVolcEngine: true,
	common.ChannelTypeOllama:     true,
	common.ChannelTypeXai:        true,
}

func GenRelayInfoWs(c *gin.Context, ws *websocket.Conn) *RelayInfo {
	info := GenRelayInfo(c)
	info.ClientWs = ws
	info.InputAudioFormat = "pcm16"
	info.OutputAudioFormat = "pcm16"
	info.IsFirstRequest = true
	return info
}

func GenRelayInfoClaude(c *gin.Context) *RelayInfo {
	info := GenRelayInfo(c)
	info.RelayFormat = RelayFormatClaude
	info.ShouldIncludeUsage = false
	info.ClaudeConvertInfo = &ClaudeConvertInfo{
		LastMessagesType: LastMessageTypeNone,
	}
	return info
}

func GenRelayInfoRerank(c *gin.Context, req *dto.RerankRequest) *RelayInfo {
	info := GenRelayInfo(c)
	info.RelayMode = relayconstant.RelayModeRerank
	info.RerankerInfo = &RerankerInfo{
		Documents:       req.Documents,
		ReturnDocuments: req.GetReturnDocuments(),
	}
	return info
}

func GenRelayInfo(c *gin.Context) *RelayInfo {
	channelType := c.GetInt("channel_type")
	channelId := c.GetInt("channel_id")
	channelSetting := c.GetStringMap("channel_setting")
	paramOverride := c.GetStringMap("param_override")

	tokenId := c.GetInt("token_id")
	tokenKey := c.GetString("token_key")
	userId := c.GetInt("id")
	group := c.GetString("group")
	tokenUnlimited := c.GetBool("token_unlimited_quota")
	startTime := c.GetTime(constant.ContextKeyRequestStartTime)
	// firstResponseTime = time.Now() - 1 second

	apiType, _ := relayconstant.ChannelType2APIType(channelType)

	// Get the original model name (with prefix if any)
	prefixedModel := c.GetString("prefixed_model")
	originalModel := c.GetString("original_model")

	// If we have a prefixed model, use it as the origin model name for display
	// but use the unprefixed model name for the upstream
	if prefixedModel != "" {
		originalModel = prefixedModel
	}

	info := &RelayInfo{
		UserQuota:         c.GetInt(constant.ContextKeyUserQuota),
		UserSetting:       c.GetStringMap(constant.ContextKeyUserSetting),
		UserEmail:         c.GetString(constant.ContextKeyUserEmail),
		isFirstResponse:   true,
		RelayMode:         relayconstant.Path2RelayMode(c.Request.URL.Path),
		BaseUrl:           c.GetString("base_url"),
		RequestURLPath:    c.Request.URL.String(),
		ChannelType:       channelType,
		ChannelId:         channelId,
		TokenId:           tokenId,
		TokenKey:          tokenKey,
		UserId:            userId,
		Group:             group,
		TokenUnlimited:    tokenUnlimited,
		StartTime:         startTime,
		FirstResponseTime: startTime.Add(-time.Second),
		OriginModelName:   originalModel,                 // Use the prefixed model name for display
		UpstreamModelName: c.GetString("original_model"), // Use the unprefixed model name for upstream
		//RecodeModelName:   c.GetString("original_model"),
		IsModelMapped:     false,
		ApiType:           apiType,
		ApiVersion:        c.GetString("api_version"),
		ApiKey:            strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "),
		Organization:      c.GetString("channel_organization"),
		ChannelSetting:    channelSetting,
		ChannelCreateTime: c.GetInt64("channel_create_time"),
		ParamOverride:     paramOverride,
		Other:             make(map[string]interface{}),
		RelayFormat:       RelayFormatOpenAI,
		ThinkingContentInfo: ThinkingContentInfo{
			IsFirstThinkingContent:  true,
			SendLastThinkingContent: false,
		},
	}
	if strings.HasPrefix(c.Request.URL.Path, "/pg") {
		info.IsPlayground = true
		info.RequestURLPath = strings.TrimPrefix(info.RequestURLPath, "/pg")
		info.RequestURLPath = "/v1" + info.RequestURLPath
	}
	if info.BaseUrl == "" {
		info.BaseUrl = common.ChannelBaseURLs[channelType]
	}
	if info.ChannelType == common.ChannelTypeAzure {
		info.ApiVersion = GetAPIVersion(c)
	}
	if info.ChannelType == common.ChannelTypeVertexAi {
		info.ApiVersion = c.GetString("region")
	}
	if streamSupportedChannels[info.ChannelType] {
		info.SupportStreamOptions = true
	}
	// responses 模式不支持 StreamOptions
	if relayconstant.RelayModeResponses == info.RelayMode {
		info.SupportStreamOptions = false
	}
	return info
}

func (info *RelayInfo) SetPromptTokens(promptTokens int) {
	info.PromptTokens = promptTokens
}

func (info *RelayInfo) SetIsStream(isStream bool) {
	info.IsStream = isStream
}

func (info *RelayInfo) SetFirstResponseTime() {
	if info.isFirstResponse {
		info.FirstResponseTime = time.Now()
		info.isFirstResponse = false
	}
}

func (info *RelayInfo) HasSendResponse() bool {
	return info.FirstResponseTime.After(info.StartTime)
}

type TaskRelayInfo struct {
	*RelayInfo
	Action       string
	OriginTaskID string

	ConsumeQuota bool
}

func GenTaskRelayInfo(c *gin.Context) *TaskRelayInfo {
	info := &TaskRelayInfo{
		RelayInfo: GenRelayInfo(c),
	}
	return info
}
