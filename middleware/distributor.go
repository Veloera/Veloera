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
package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"veloera/common"
	"veloera/constant"
	"veloera/dto"
	"veloera/model"
	relayconstant "veloera/relay/constant"
	"veloera/service"
	"veloera/setting"

	"github.com/gin-gonic/gin"
)

type ModelRequest struct {
	Model string `json:"model"`
}

// Cache for storing channels by prefix
var (
	prefixChannelsMutex       sync.RWMutex
	prefixChannelsCache       = make(map[string]map[string][]*model.Channel) // group -> prefix -> channels
	prefixChannelsCacheExpiry = make(map[string]int64)                       // group -> expiry timestamp

	// Round-robin key selection state
	channelKeysMutex sync.Mutex
	channelKeysIndex = make(map[int]int)    // channel_id -> current key index
	channelKeysHash  = make(map[int]string) // channel_id -> hash of keys
)

// getPrefixChannels returns a map of prefix -> channels for a given group
// The cache is refreshed every hour
func getPrefixChannels(group string) map[string][]*model.Channel {
	prefixChannelsMutex.RLock()
	expiry, ok := prefixChannelsCacheExpiry[group]
	prefixChannelsMutex.RUnlock()

	currentTime := time.Now().Unix()
	if !ok || expiry < currentTime {
		// Cache expired, refresh it
		return refreshPrefixChannelsCache(group)
	}

	prefixChannelsMutex.RLock()
	defer prefixChannelsMutex.RUnlock()
	prefixMap, ok := prefixChannelsCache[group]
	if !ok {
		return make(map[string][]*model.Channel)
	}

	return prefixMap
}

// GetPrefixChannels is the exported version of getPrefixChannels for use by other packages
func GetPrefixChannels(group string) map[string][]*model.Channel {
	return getPrefixChannels(group)
}

// SelectChannelByPrefix is the exported version of selectChannelByPrefix for use by other packages
func SelectChannelByPrefix(group, prefix, originalModel string) (*model.Channel, error) {
	return selectChannelByPrefix(group, prefix, originalModel)
}

// ResetChannelKeyIndex resets the round-robin key index for a specific channel
// This can be called when a channel's keys are updated
func ResetChannelKeyIndex(channelId int) {
	channelKeysMutex.Lock()
	defer channelKeysMutex.Unlock()

	// Reset the key index and hash to force recalculation next time
	delete(channelKeysIndex, channelId)
	delete(channelKeysHash, channelId)
}

// RefreshPrefixChannelsCache refreshes prefix cache for one or multiple groups.
// Groups should be a comma separated string, empty entries are ignored.
func RefreshPrefixChannelsCache(groups string) {
	for _, g := range strings.Split(groups, ",") {
		g = strings.TrimSpace(g)
		if g == "" {
			continue
		}
		refreshPrefixChannelsCache(g)
	}
}

// refreshPrefixChannelsCache refreshes the prefix channels cache for a given group
func refreshPrefixChannelsCache(group string) map[string][]*model.Channel {
	var channels []*model.Channel

	// Get channels for this group from the database
	db := model.DB.Model(&model.Channel{}).Where("status = ?", common.ChannelStatusEnabled)
	if group != "" {
		var condition string
		groupCol := "`group`"
		if common.UsingPostgreSQL {
			groupCol = "\"group\""
		}
		if common.UsingMySQL {
			condition = fmt.Sprintf("CONCAT(',', %s, ',') LIKE '%%,%s,%%'", groupCol, group)
		} else {
			// sqlite, PostgreSQL
			condition = fmt.Sprintf("(',' || %s || ',') LIKE '%%,%s,%%'", groupCol, group)
		}
		db = db.Where(condition)
	}

	db.Order("priority desc").Find(&channels)
	prefixMap := make(map[string][]*model.Channel)

	// Group channels by prefix
	for _, channel := range channels {
		prefix := ""
		if channel.ModelPrefix != nil {
			prefix = *channel.ModelPrefix
		}
		prefixMap[prefix] = append(prefixMap[prefix], channel)
	}

	// Store in cache
	prefixChannelsMutex.Lock()
	prefixChannelsCache[group] = prefixMap
	prefixChannelsCacheExpiry[group] = time.Now().Add(1 * time.Hour).Unix()
	prefixChannelsMutex.Unlock()

	return prefixMap
}

// selectChannelByPrefix selects a channel based on the model prefix
func selectChannelByPrefix(group, prefix, originalModel string) (*model.Channel, error) {
	prefixMap := getPrefixChannels(group)

	channels, ok := prefixMap[prefix]
	if !ok || len(channels) == 0 {
		return nil, fmt.Errorf("no channels found for prefix %s", prefix)
	}

	// Filter channels that support the model (without prefix)
	var compatibleChannels []*model.Channel
	for _, channel := range channels {
		// Check if the channel supports the model
		for _, model := range channel.GetModels() {
			if model == originalModel {
				compatibleChannels = append(compatibleChannels, channel)
				break
			}
		}
	}

	if len(compatibleChannels) == 0 {
		return nil, fmt.Errorf("no channels supporting model %s found for prefix %s", originalModel, prefix)
	}

	// Select a random channel based on weight
	totalWeight := 0
	for _, channel := range compatibleChannels {
		totalWeight += channel.GetWeight()
	}

	if totalWeight <= 0 {
		// If no weight or all zero weights, select randomly
		return compatibleChannels[common.GetRandomInt(len(compatibleChannels))], nil
	}

	// Weighted random selection
	randWeight := common.GetRandomInt(totalWeight)
	currentWeight := 0

	for _, channel := range compatibleChannels {
		currentWeight += channel.GetWeight()
		if randWeight < currentWeight {
			return channel, nil
		}
	}

	// Fallback - should not happen unless there's a calculation error
	return compatibleChannels[0], nil
}

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		allowIpsMap := c.GetStringMap("allow_ips")
		if len(allowIpsMap) != 0 {
			clientIp := c.ClientIP()
			if _, ok := allowIpsMap[clientIp]; !ok {
				abortWithOpenAiMessage(c, http.StatusForbidden, "您的 IP 不在令牌允许访问的列表中")
				return
			}
		}
		var channel *model.Channel
		channelId, ok := c.Get("specific_channel_id")
		modelRequest, shouldSelectChannel, err := getModelRequest(c)
		if err != nil {
			abortWithOpenAiMessage(c, http.StatusBadRequest, "Invalid request, "+err.Error())
			return
		}
		userGroup := c.GetString(constant.ContextKeyUserGroup)
		tokenGroup := c.GetString("token_group")
		if tokenGroup != "" {
			// check common.UserUsableGroups[userGroup]
			if _, ok := setting.GetUserUsableGroups(userGroup)[tokenGroup]; !ok {
				abortWithOpenAiMessage(c, http.StatusForbidden, fmt.Sprintf("令牌分组 %s 已被禁用", tokenGroup))
				return
			}
			// check group in common.GroupRatio
			if !setting.ContainsGroupRatio(tokenGroup) {
				abortWithOpenAiMessage(c, http.StatusForbidden, fmt.Sprintf("分组 %s 已被弃用", tokenGroup))
				return
			}
			userGroup = tokenGroup
		}
		c.Set("group", userGroup)

		// Check if the model has a prefix, which is used for routing
		originalModel := modelRequest.Model
		prefixedModel, hasPrefixedModel := c.Get("prefixed_model")
		modelPrefix := ""
		prefixedModelStr := ""
		if hasPrefixedModel {
			prefixedModelStr = prefixedModel.(string)
			// Extract prefix from the model name if it exists
			for prefix := range getPrefixChannels(userGroup) {
				if prefix != "" && strings.HasPrefix(prefixedModelStr, prefix) {
					modelPrefix = prefix
					// Update the model name to strip the prefix for channel selection
					modelRequest.Model = strings.TrimPrefix(prefixedModelStr, prefix)
					break
				}
			}
		}

		if ok {
			id, err := strconv.Atoi(channelId.(string))
			if err != nil {
				abortWithOpenAiMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			channel, err = model.GetChannelById(id, true)
			if err != nil {
				abortWithOpenAiMessage(c, http.StatusBadRequest, "无效的渠道 Id")
				return
			}
			if channel.Status != common.ChannelStatusEnabled {
				abortWithOpenAiMessage(c, http.StatusForbidden, "该渠道已被禁用")
				return
			}
		} else {
			// Select a channel for the user
			// check token model mapping
			modelLimitEnable := c.GetBool("token_model_limit_enabled")
			if modelLimitEnable {
				s, ok := c.Get("token_model_limit")
				var tokenModelLimit map[string]bool
				if ok {
					tokenModelLimit = s.(map[string]bool)
				} else {
					tokenModelLimit = map[string]bool{}
				}
				if tokenModelLimit != nil {
					// Check access against the original (prefixed) model name
					if _, ok := tokenModelLimit[originalModel]; !ok {
						abortWithOpenAiMessage(c, http.StatusForbidden, "该令牌无权访问模型 "+originalModel)
						return
					}
				} else {
					// token model limit is empty, all models are not allowed
					abortWithOpenAiMessage(c, http.StatusForbidden, "该令牌无权访问任何模型")
					return
				}
			}

			if shouldSelectChannel {
				// If we have a model prefix, use it to select among specific channels
				if modelPrefix != "" {
					channel, err = selectChannelByPrefix(userGroup, modelPrefix, modelRequest.Model)
				} else {
					channel, err = model.CacheGetRandomSatisfiedChannel(userGroup, modelRequest.Model, 0)
				}

				if err != nil {
					message := fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道", userGroup, originalModel)
					// 如果错误，但是渠道不为空，说明是数据库一致性问题
					if channel != nil {
						common.SysError(fmt.Sprintf("渠道不存在：%d", channel.Id))
						message = "数据库一致性已被破坏，请联系管理员"
					}
					// 如果错误，而且渠道为空，说明是没有可用渠道
					abortWithOpenAiMessage(c, http.StatusServiceUnavailable, message)
					return
				}
				if channel == nil {
					abortWithOpenAiMessage(c, http.StatusServiceUnavailable, fmt.Sprintf("当前分组 %s 下对于模型 %s 无可用渠道（数据库一致性已被破坏）", userGroup, originalModel))
					return
				}
			}
		}
		c.Set(constant.ContextKeyRequestStartTime, time.Now())
		SetupContextForSelectedChannel(c, channel, modelRequest.Model)
		c.Next()
	}
}

func getModelRequest(c *gin.Context) (*ModelRequest, bool, error) {
	var modelRequest ModelRequest
	shouldSelectChannel := true
	var err error
	if strings.Contains(c.Request.URL.Path, "/mj/") {
		relayMode := relayconstant.Path2RelayModeMidjourney(c.Request.URL.Path)
		if relayMode == relayconstant.RelayModeMidjourneyTaskFetch ||
			relayMode == relayconstant.RelayModeMidjourneyTaskFetchByCondition ||
			relayMode == relayconstant.RelayModeMidjourneyNotify ||
			relayMode == relayconstant.RelayModeMidjourneyTaskImageSeed {
			shouldSelectChannel = false
		} else {
			midjourneyRequest := dto.MidjourneyRequest{}
			err = common.UnmarshalBodyReusable(c, &midjourneyRequest)
			if err != nil {
				return nil, false, err
			}
			midjourneyModel, mjErr, success := service.GetMjRequestModel(relayMode, &midjourneyRequest)
			if mjErr != nil {
				return nil, false, fmt.Errorf(mjErr.Description)
			}
			if midjourneyModel == "" {
				if !success {
					return nil, false, fmt.Errorf("无效的请求, 无法解析模型")
				} else {
					// task fetch, task fetch by condition, notify
					shouldSelectChannel = false
				}
			}
			modelRequest.Model = midjourneyModel
		}
		c.Set("relay_mode", relayMode)
	} else if strings.Contains(c.Request.URL.Path, "/suno/") {
		relayMode := relayconstant.Path2RelaySuno(c.Request.Method, c.Request.URL.Path)
		if relayMode == relayconstant.RelayModeSunoFetch ||
			relayMode == relayconstant.RelayModeSunoFetchByID {
			shouldSelectChannel = false
		} else {
			modelName := service.CoverTaskActionToModelName(constant.TaskPlatformSuno, c.Param("action"))
			modelRequest.Model = modelName
		}
		c.Set("platform", string(constant.TaskPlatformSuno))
		c.Set("relay_mode", relayMode)
	} else if !strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") {
		err = common.UnmarshalBodyReusable(c, &modelRequest)
	}
	if err != nil {
		return nil, false, errors.New("无效的请求, " + err.Error())
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/realtime") {
		//wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview-2024-10-01
		modelRequest.Model = c.Query("model")
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "text-moderation-stable"
		}
	}
	if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
		if modelRequest.Model == "" {
			modelRequest.Model = c.Param("model")
		}
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
		modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "dall-e")
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/audio") {
		relayMode := relayconstant.RelayModeAudioSpeech
		if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/speech") {
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "tts-1")
		} else if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/translations") {
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, c.PostForm("model"))
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "whisper-1")
			relayMode = relayconstant.RelayModeAudioTranslation
		} else if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") {
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, c.PostForm("model"))
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "whisper-1")
			relayMode = relayconstant.RelayModeAudioTranscription
		}
		c.Set("relay_mode", relayMode)
	}

	// Check if the model name has a prefix that needs to be used for routing
	// We save both the original model name (with prefix) and the model name without prefix
	if modelRequest.Model != "" {
		c.Set("prefixed_model", modelRequest.Model) // Store the original model name for later reference
	}

	return &modelRequest, shouldSelectChannel, nil
}

func SetupContextForSelectedChannel(c *gin.Context, channel *model.Channel, modelName string) {
	c.Set("original_model", modelName) // for retry
	if channel == nil {
		return
	}
	c.Set("channel_id", channel.Id)
	c.Set("channel_name", channel.Name)
	c.Set("channel_type", channel.Type)
	c.Set("channel_create_time", channel.CreatedTime)
	c.Set("channel_setting", channel.GetSetting())
	c.Set("param_override", channel.GetParamOverride())

	// Set model prefix if available
	if channel.ModelPrefix != nil && *channel.ModelPrefix != "" {
		c.Set("model_prefix", *channel.ModelPrefix)
	}

	if nil != channel.OpenAIOrganization && "" != *channel.OpenAIOrganization {
		c.Set("channel_organization", *channel.OpenAIOrganization)
	}
	c.Set("auto_ban", channel.GetAutoBan())
	c.Set("model_mapping", channel.GetModelMapping())
	c.Set("status_code_mapping", channel.GetStatusCodeMapping())

	// 如果key包含逗号，使用轮询方式选择一个key
	// 对于渠道类型41，不处理逗号分隔的多key机制
	if channel.Type == 41 {
		c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
	} else if strings.Contains(channel.Key, ",") {
		keys := strings.Split(channel.Key, ",")

		// Get current index for this channel using round-robin
		channelKeysMutex.Lock()

		// Check if keys have changed by comparing with stored hash
		currentHash := common.GetMD5Hash(channel.Key)
		storedHash, hashExists := channelKeysHash[channel.Id]

		// Reset index if keys have changed or index doesn't exist
		index := 0
		if hashExists && storedHash == currentHash {
			// Keys haven't changed, use stored index
			storedIndex, exists := channelKeysIndex[channel.Id]
			if exists && storedIndex < len(keys) {
				index = storedIndex
			}
		} else {
			// Keys have changed or this is first use, update hash and reset index
			channelKeysHash[channel.Id] = currentHash
		}

		// Select key at current index
		selectedKey := keys[index]

		// Update index for next use
		channelKeysIndex[channel.Id] = (index + 1) % len(keys)
		channelKeysMutex.Unlock()

		c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", strings.TrimSpace(selectedKey)))
	} else {
		c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
	}

	c.Set("base_url", channel.GetBaseURL())
	// TODO: api_version统一
	switch channel.Type {
	case common.ChannelTypeAzure:
		c.Set("api_version", channel.Other)
	case common.ChannelTypeVertexAi:
		c.Set("region", channel.Other)
	case common.ChannelTypeXunfei:
		c.Set("api_version", channel.Other)
	case common.ChannelTypeGemini:
		c.Set("api_version", channel.Other)
	case common.ChannelTypeAli:
		c.Set("plugin", channel.Other)
	case common.ChannelCloudflare:
		c.Set("api_version", channel.Other)
	case common.ChannelTypeMokaAI:
		c.Set("api_version", channel.Other)
	}
}
