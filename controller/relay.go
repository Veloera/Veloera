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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"veloera/common"
	"veloera/dto"
	"veloera/middleware"
	"veloera/model"
	"veloera/relay"
	relaycommon "veloera/relay/common"
	"veloera/relay/constant"
	relayconstant "veloera/relay/constant"
	"veloera/relay/helper"
	"veloera/service"
	"veloera/setting/model_setting"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func relayHandler(c *gin.Context, relayMode int) *dto.OpenAIErrorWithStatusCode {
	var err *dto.OpenAIErrorWithStatusCode
	switch relayMode {
	case relayconstant.RelayModeImagesGenerations:
		err = relay.ImageHelper(c)
	case relayconstant.RelayModeAudioSpeech:
		fallthrough
	case relayconstant.RelayModeAudioTranslation:
		fallthrough
	case relayconstant.RelayModeAudioTranscription:
		err = relay.AudioHelper(c)
	case relayconstant.RelayModeRerank:
		err = relay.RerankHelper(c, relayMode)
	case relayconstant.RelayModeEmbeddings:
		err = relay.EmbeddingHelper(c)
	case relayconstant.RelayModeResponses:
		err = relay.ResponsesHelper(c)
	case relayconstant.RelayModeTokenCount:
		err = relay.TokenCountHelper(c)
	default:
		err = relay.TextHelper(c)
	}

	if err != nil && common.LogErrorEnabled { // If error log is enabled
		// 保存错误日志到mysql中
		userId := c.GetInt("id")
		tokenName := c.GetString("token_name")
		modelName := c.GetString("original_model")
		tokenId := c.GetInt("token_id")
		userGroup := c.GetString("group")
		channelId := c.GetInt("channel_id")
		other := make(map[string]interface{})
		other["error_type"] = err.Error.Type
		other["error_code"] = err.Error.Code
		other["status_code"] = err.StatusCode
		other["channel_id"] = channelId
		other["channel_name"] = c.GetString("channel_name")
		other["channel_type"] = c.GetInt("channel_type")

		model.RecordErrorLog(c, userId, channelId, modelName, tokenName, err.Error.Message, tokenId, 0, false, userGroup, other)
	}

	return err
}

func Relay(c *gin.Context) {
	relayMode := constant.Path2RelayMode(c.Request.URL.Path)
	requestId := c.GetString(common.RequestIdKey)
	group := c.GetString("group")
	originalModel := c.GetString("original_model")
	var openaiErr *dto.OpenAIErrorWithStatusCode

	if model_setting.GetGlobalSettings().BlockBrowserExtensionEnabled &&
		relayMode == relayconstant.RelayModeChatCompletions {
		origin := c.Request.Header.Get("Origin")
		if strings.Contains(origin, "extension") && strings.Contains(origin, "://") {
			openaiErr = service.OpenAIErrorWrapperLocal(fmt.Errorf("browser extension blocked"), "browser_extension_blocked", http.StatusForbidden)
			c.JSON(openaiErr.StatusCode, gin.H{"error": openaiErr.Error})
			return
		}
	}

	autoRetryCount := model_setting.GetAutoRetryCount()
	maxRetries := autoRetryCount

	for i := 0; i <= maxRetries; i++ {
		// 如果启用了强制切换渠道，且不是第一次尝试，则增加重试索引以获取不同渠道
		retryIndex := i
		if i > 0 && model_setting.ShouldForceChannelSwitch() {
			retryIndex = i + 1 // 强制获取不同的渠道
		}

		channel, err := getChannel(c, group, originalModel, retryIndex)
		if err != nil {
			common.LogError(c, err.Error())
			openaiErr = service.OpenAIErrorWrapperLocal(err, "get_channel_failed", http.StatusInternalServerError)
			break
		}

		// 记录响应前的状态，用于检测空回复
		c.Set("response_written", false)
		openaiErr = relayRequest(c, relayMode, channel)

		// 检测空回复的情况
		if openaiErr == nil {
			// 检查是否实际写入了响应内容
			responseWritten := c.GetBool("response_written")
			if !responseWritten && model_setting.GetGlobalSettings().AutoRetryEnabled {
				// 创建一个表示空回复的错误，以便触发重试
				openaiErr = service.OpenAIErrorWrapperLocal(
					fmt.Errorf("empty response from upstream"),
					"empty_response",
					http.StatusInternalServerError,
				)
				common.LogWarn(c, fmt.Sprintf("detected empty response from channel #%d, will retry", channel.Id))
			} else {
				return // 成功处理请求，直接返回
			}
		}

		go processChannelError(c, channel.Id, channel.Type, channel.Name, channel.GetAutoBan(), openaiErr)

		if !shouldRetryWithAutoConfig(c, openaiErr, maxRetries-i) {
			break
		}
	}
	useChannel := c.GetStringSlice("use_channel")
	if len(useChannel) > 1 {
		retryLogStr := fmt.Sprintf("重试：%s", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(useChannel)), "->"), "[]"))
		common.LogInfo(c, retryLogStr)
	}

	if openaiErr != nil {
		if openaiErr.StatusCode == http.StatusTooManyRequests {
			common.LogError(c, fmt.Sprintf("origin 429 error: %s", openaiErr.Error.Message))
			openaiErr.Error.Message = "当前分组上游负载已饱和，请稍后再试"
		}
		openaiErr.Error.Message = common.MessageWithRequestId(openaiErr.Error.Message, requestId)
		if c.GetString("relay_format") == relaycommon.RelayFormatGemini {
			geminiErr := service.OpenAIErrorToGeminiResponse(openaiErr)
			c.JSON(openaiErr.StatusCode, geminiErr)
		} else {
			c.JSON(openaiErr.StatusCode, gin.H{
				"error": openaiErr.Error,
			})
		}
	}
}

var upgrader = websocket.Upgrader{
	Subprotocols: []string{"realtime"}, // WS 握手支持的协议，如果有使用 Sec-WebSocket-Protocol，则必须在此声明对应的 Protocol TODO add other protocol
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

func WssRelay(c *gin.Context) {
	// 将 HTTP 连接升级为 WebSocket 连接

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	defer ws.Close()

	if err != nil {
		openaiErr := service.OpenAIErrorWrapper(err, "get_channel_failed", http.StatusInternalServerError)
		helper.WssError(c, ws, openaiErr.Error)
		return
	}

	relayMode := constant.Path2RelayMode(c.Request.URL.Path)
	requestId := c.GetString(common.RequestIdKey)
	group := c.GetString("group")
	//wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview-2024-10-01
	originalModel := c.GetString("original_model")
	var openaiErr *dto.OpenAIErrorWithStatusCode

	for i := 0; i <= common.RetryTimes; i++ {
		channel, err := getChannel(c, group, originalModel, i)
		if err != nil {
			common.LogError(c, err.Error())
			openaiErr = service.OpenAIErrorWrapperLocal(err, "get_channel_failed", http.StatusInternalServerError)
			break
		}

		openaiErr = wssRequest(c, ws, relayMode, channel)

		if openaiErr == nil {
			return // 成功处理请求，直接返回
		}

		go processChannelError(c, channel.Id, channel.Type, channel.Name, channel.GetAutoBan(), openaiErr)

		if !shouldRetry(c, openaiErr, common.RetryTimes-i) {
			break
		}
	}
	useChannel := c.GetStringSlice("use_channel")
	if len(useChannel) > 1 {
		retryLogStr := fmt.Sprintf("重试：%s", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(useChannel)), "->"), "[]"))
		common.LogInfo(c, retryLogStr)
	}

	if openaiErr != nil {
		if openaiErr.StatusCode == http.StatusTooManyRequests {
			openaiErr.Error.Message = "当前分组上游负载已饱和，请稍后再试"
		}
		openaiErr.Error.Message = common.MessageWithRequestId(openaiErr.Error.Message, requestId)
		helper.WssError(c, ws, openaiErr.Error)
	}
}

func RelayClaude(c *gin.Context) {
	//relayMode := constant.Path2RelayMode(c.Request.URL.Path)
	requestId := c.GetString(common.RequestIdKey)
	group := c.GetString("group")
	originalModel := c.GetString("original_model")
	var claudeErr *dto.ClaudeErrorWithStatusCode

	for i := 0; i <= common.RetryTimes; i++ {
		channel, err := getChannel(c, group, originalModel, i)
		if err != nil {
			common.LogError(c, err.Error())
			claudeErr = service.ClaudeErrorWrapperLocal(err, "get_channel_failed", http.StatusInternalServerError)
			break
		}

		claudeErr = claudeRequest(c, channel)

		if claudeErr == nil {
			return // 成功处理请求，直接返回
		}

		openaiErr := service.ClaudeErrorToOpenAIError(claudeErr)

		go processChannelError(c, channel.Id, channel.Type, channel.Name, channel.GetAutoBan(), openaiErr)

		if !shouldRetry(c, openaiErr, common.RetryTimes-i) {
			break
		}
	}
	useChannel := c.GetStringSlice("use_channel")
	if len(useChannel) > 1 {
		retryLogStr := fmt.Sprintf("重试：%s", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(useChannel)), "->"), "[]"))
		common.LogInfo(c, retryLogStr)
	}

	if claudeErr != nil {
		claudeErr.Error.Message = common.MessageWithRequestId(claudeErr.Error.Message, requestId)
		c.JSON(claudeErr.StatusCode, gin.H{
			"type":  "error",
			"error": claudeErr.Error,
		})
	}
}

func relayRequest(c *gin.Context, relayMode int, channel *model.Channel) *dto.OpenAIErrorWithStatusCode {
	addUsedChannel(c, channel.Id)
	requestBody, _ := common.GetRequestBody(c)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return relayHandler(c, relayMode)
}

func wssRequest(c *gin.Context, ws *websocket.Conn, relayMode int, channel *model.Channel) *dto.OpenAIErrorWithStatusCode {
	addUsedChannel(c, channel.Id)
	requestBody, _ := common.GetRequestBody(c)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return relay.WssHelper(c, ws)
}

func claudeRequest(c *gin.Context, channel *model.Channel) *dto.ClaudeErrorWithStatusCode {
	addUsedChannel(c, channel.Id)
	requestBody, _ := common.GetRequestBody(c)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return relay.ClaudeHelper(c)
}

func addUsedChannel(c *gin.Context, channelId int) {
	useChannel := c.GetStringSlice("use_channel")
	useChannel = append(useChannel, fmt.Sprintf("%d", channelId))
	c.Set("use_channel", useChannel)
}

func getChannel(c *gin.Context, group, originalModel string, retryCount int) (*model.Channel, error) {
	if retryCount == 0 {
		autoBan := c.GetBool("auto_ban")
		autoBanInt := 1
		if !autoBan {
			autoBanInt = 0
		}
		return &model.Channel{
			Id:      c.GetInt("channel_id"),
			Type:    c.GetInt("channel_type"),
			Name:    c.GetString("channel_name"),
			AutoBan: &autoBanInt,
		}, nil
	}
	channel, err := model.CacheGetRandomSatisfiedChannel(group, originalModel, retryCount)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("获取重试渠道失败: %s", err.Error()))
	}
	middleware.SetupContextForSelectedChannel(c, channel, originalModel)
	return channel, nil
}

func shouldRetry(c *gin.Context, openaiErr *dto.OpenAIErrorWithStatusCode, retryTimes int) bool {
	if openaiErr == nil {
		return false
	}
	if openaiErr.LocalError {
		return false
	}
	if retryTimes <= 0 {
		return false
	}
	if _, ok := c.Get("specific_channel_id"); ok {
		return false
	}
	if openaiErr.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if openaiErr.StatusCode == 307 {
		return true
	}
	if openaiErr.StatusCode/100 == 5 {
		// 超时不重试
		if openaiErr.StatusCode == 504 || openaiErr.StatusCode == 524 {
			return false
		}
		return true
	}
	if openaiErr.StatusCode == http.StatusBadRequest {
		channelType := c.GetInt("channel_type")
		if channelType == common.ChannelTypeAnthropic {
			return true
		}
		return false
	}
	if openaiErr.StatusCode == 408 {
		// azure处理超时不重试
		return false
	}
	if openaiErr.StatusCode/100 == 2 {
		return false
	}
	return true
}

// shouldRetryWithAutoConfig 使用自动重试配置判断是否应该重试
func shouldRetryWithAutoConfig(c *gin.Context, openaiErr *dto.OpenAIErrorWithStatusCode, retryTimes int) bool {
	if openaiErr == nil {
		return false
	}
	if openaiErr.LocalError {
		return false
	}
	if retryTimes <= 0 {
		return false
	}
	if _, ok := c.Get("specific_channel_id"); ok {
		return false
	}

	// 检查是否启用了自动重试
	if !model_setting.GetGlobalSettings().AutoRetryEnabled {
		return shouldRetry(c, openaiErr, retryTimes) // 回退到原有逻辑
	}

	// 使用配置的状态码判断是否重试
	if !model_setting.ShouldRetryForStatusCode(openaiErr.StatusCode) {
		return false
	}

	// 对于空回复的情况，也应该重试
	if openaiErr.Error.Message == "" || strings.Contains(strings.ToLower(openaiErr.Error.Message), "empty") {
		return true
	}

	// 超时不重试
	if openaiErr.StatusCode == 504 || openaiErr.StatusCode == 524 || openaiErr.StatusCode == 408 {
		return false
	}

	// 成功状态码不重试
	if openaiErr.StatusCode/100 == 2 {
		return false
	}

	return true
}

func processChannelError(c *gin.Context, channelId int, channelType int, channelName string, autoBan bool, err *dto.OpenAIErrorWithStatusCode) {
	// 不要使用context获取渠道信息，异步处理时可能会出现渠道信息不一致的情况
	// do not use context to get channel info, there may be inconsistent channel info when processing asynchronously
	common.LogError(c, fmt.Sprintf("relay error (channel #%d, status code: %d): %s", channelId, err.StatusCode, err.Error.Message))
	if service.ShouldDisableChannel(channelType, err) && autoBan {
		service.DisableChannel(channelId, channelName, err.Error.Message)
	}
}

func RelayMidjourney(c *gin.Context) {
	relayMode := c.GetInt("relay_mode")
	var err *dto.MidjourneyResponse
	switch relayMode {
	case relayconstant.RelayModeMidjourneyNotify:
		err = relay.RelayMidjourneyNotify(c)
	case relayconstant.RelayModeMidjourneyTaskFetch, relayconstant.RelayModeMidjourneyTaskFetchByCondition:
		err = relay.RelayMidjourneyTask(c, relayMode)
	case relayconstant.RelayModeMidjourneyTaskImageSeed:
		err = relay.RelayMidjourneyTaskImageSeed(c)
	case relayconstant.RelayModeSwapFace:
		err = relay.RelaySwapFace(c)
	default:
		err = relay.RelayMidjourneySubmit(c, relayMode)
	}
	//err = relayMidjourneySubmit(c, relayMode)
	log.Println(err)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Code == 30 {
			err.Result = "当前分组负载已饱和，请稍后再试，或升级账户以提升服务质量。"
			statusCode = http.StatusTooManyRequests
		}
		c.JSON(statusCode, gin.H{
			"description": fmt.Sprintf("%s %s", err.Description, err.Result),
			"type":        "upstream_error",
			"code":        err.Code,
		})
		channelId := c.GetInt("channel_id")
		common.LogError(c, fmt.Sprintf("relay error (channel #%d, status code %d): %s", channelId, statusCode, fmt.Sprintf("%s %s", err.Description, err.Result)))
	}
}

func RelayNotImplemented(c *gin.Context) {
	err := dto.OpenAIError{
		Message: "API not implemented",
		Type:    "veloera_error",
		Param:   "",
		Code:    "api_not_implemented",
	}
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": err,
	})
}

func RelayNotFound(c *gin.Context) {
	err := dto.OpenAIError{
		Message: fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path),
		Type:    "invalid_request_error",
		Param:   "",
		Code:    "",
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": err,
	})
}

func RelayTokenCount(c *gin.Context) {
	requestId := c.GetString(common.RequestIdKey)

	// Get request body and parse it
	requestBody, err := common.GetRequestBody(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "Failed to read request body",
			},
		})
		return
	}

	// Parse Claude request
	var claudeRequest dto.ClaudeRequest
	err = json.Unmarshal(requestBody, &claudeRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": "Invalid request format",
			},
		})
		return
	}

	// Count tokens locally without needing channels
	tokenCount, err := service.CountTokenClaudeRequest(claudeRequest, claudeRequest.Model)
	if err != nil {
		common.LogError(c, fmt.Sprintf("Token counting failed: %s", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "internal_server_error",
				"message": common.MessageWithRequestId("Token counting failed", requestId),
			},
		})
		return
	}

	// Return token count in Claude API format
	c.JSON(http.StatusOK, gin.H{
		"input_tokens": tokenCount,
	})
}

func RelayTask(c *gin.Context) {
	retryTimes := model_setting.GetAutoRetryCount()
	channelId := c.GetInt("channel_id")
	relayMode := c.GetInt("relay_mode")
	group := c.GetString("group")
	originalModel := c.GetString("original_model")
	c.Set("use_channel", []string{fmt.Sprintf("%d", channelId)})
	taskErr := taskRelayHandler(c, relayMode)
	if taskErr == nil {
		retryTimes = 0
	}
	for i := 0; shouldRetryTaskRelay(c, channelId, taskErr, retryTimes) && i < retryTimes; i++ {
		channel, err := model.CacheGetRandomSatisfiedChannel(group, originalModel, i)
		if err != nil {
			common.LogError(c, fmt.Sprintf("CacheGetRandomSatisfiedChannel failed: %s", err.Error()))
			break
		}
		channelId = channel.Id
		useChannel := c.GetStringSlice("use_channel")
		useChannel = append(useChannel, fmt.Sprintf("%d", channelId))
		c.Set("use_channel", useChannel)
		common.LogInfo(c, fmt.Sprintf("using channel #%d to retry (remain times %d)", channel.Id, i))
		middleware.SetupContextForSelectedChannel(c, channel, originalModel)

		requestBody, err := common.GetRequestBody(c)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		taskErr = taskRelayHandler(c, relayMode)
	}
	useChannel := c.GetStringSlice("use_channel")
	if len(useChannel) > 1 {
		retryLogStr := fmt.Sprintf("重试：%s", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(useChannel)), "->"), "[]"))
		common.LogInfo(c, retryLogStr)
	}
	if taskErr != nil {
		if taskErr.StatusCode == http.StatusTooManyRequests {
			taskErr.Message = "当前分组上游负载已饱和，请稍后再试"
		}
		c.JSON(taskErr.StatusCode, taskErr)
	}
}

func taskRelayHandler(c *gin.Context, relayMode int) *dto.TaskError {
	var err *dto.TaskError
	switch relayMode {
	case relayconstant.RelayModeSunoFetch, relayconstant.RelayModeSunoFetchByID:
		err = relay.RelayTaskFetch(c, relayMode)
	default:
		err = relay.RelayTaskSubmit(c, relayMode)
	}
	return err
}

func shouldRetryTaskRelay(c *gin.Context, channelId int, taskErr *dto.TaskError, retryTimes int) bool {
	if taskErr == nil {
		return false
	}
	if retryTimes <= 0 {
		return false
	}
	if _, ok := c.Get("specific_channel_id"); ok {
		return false
	}
	if taskErr.StatusCode == http.StatusTooManyRequests {
		return true
	}
	if taskErr.StatusCode == 307 {
		return true
	}
	if taskErr.StatusCode/100 == 5 {
		// 超时不重试
		if taskErr.StatusCode == 504 || taskErr.StatusCode == 524 {
			return false
		}
		return true
	}
	if taskErr.StatusCode == http.StatusBadRequest {
		return false
	}
	if taskErr.StatusCode == 408 {
		// azure处理超时不重试
		return false
	}
	if taskErr.LocalError {
		return false
	}
	if taskErr.StatusCode/100 == 2 {
		return false
	}
	return true
}
