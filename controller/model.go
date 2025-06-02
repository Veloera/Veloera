package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"veloera/common"
	"veloera/constant"
	"veloera/dto"
	"veloera/middleware"
	"veloera/model"
	"veloera/relay"
	"veloera/relay/channel/ai360"
	"veloera/relay/channel/lingyiwanwu"
	"veloera/relay/channel/minimax"
	"veloera/relay/channel/moonshot"
	relaycommon "veloera/relay/common"
	relayconstant "veloera/relay/constant"
)

// https://platform.openai.com/docs/api-reference/models/list

var openAIModels []dto.OpenAIModels
var openAIModelsMap map[string]dto.OpenAIModels
var channelId2Models map[int][]string

func getPermission() []dto.OpenAIModelPermission {
	var permission []dto.OpenAIModelPermission
	permission = append(permission, dto.OpenAIModelPermission{
		Id:                 "modelperm-LwHkVFn8AcMItP432fKKDIKJ",
		Object:             "model_permission",
		Created:            1626777600,
		AllowCreateEngine:  true,
		AllowSampling:      true,
		AllowLogprobs:      true,
		AllowSearchIndices: false,
		AllowView:          true,
		AllowFineTuning:    false,
		Organization:       "*",
		Group:              nil,
		IsBlocking:         false,
	})
	return permission
}

func init() {
	// https://platform.openai.com/docs/models/model-endpoint-compatibility
	permission := getPermission()
	for i := 0; i < relayconstant.APITypeDummy; i++ {
		if i == relayconstant.APITypeAIProxyLibrary {
			continue
		}
		adaptor := relay.GetAdaptor(i)
		channelName := adaptor.GetChannelName()
		modelNames := adaptor.GetModelList()
		for _, modelName := range modelNames {
			openAIModels = append(openAIModels, dto.OpenAIModels{
				Id:         modelName,
				Object:     "model",
				Created:    1626777600,
				OwnedBy:    channelName,
				Permission: permission,
				Root:       modelName,
				Parent:     nil,
			})
		}
	}
	for _, modelName := range ai360.ModelList {
		openAIModels = append(openAIModels, dto.OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    ai360.ChannelName,
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	for _, modelName := range moonshot.ModelList {
		openAIModels = append(openAIModels, dto.OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    moonshot.ChannelName,
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	for _, modelName := range lingyiwanwu.ModelList {
		openAIModels = append(openAIModels, dto.OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    lingyiwanwu.ChannelName,
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	for _, modelName := range minimax.ModelList {
		openAIModels = append(openAIModels, dto.OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    minimax.ChannelName,
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	for modelName := range constant.MidjourneyModel2Action {
		openAIModels = append(openAIModels, dto.OpenAIModels{
			Id:         modelName,
			Object:     "model",
			Created:    1626777600,
			OwnedBy:    "midjourney",
			Permission: permission,
			Root:       modelName,
			Parent:     nil,
		})
	}
	openAIModelsMap = make(map[string]dto.OpenAIModels)
	for _, aiModel := range openAIModels {
		openAIModelsMap[aiModel.Id] = aiModel
	}
	channelId2Models = make(map[int][]string)
	for i := 1; i <= common.ChannelTypeDummy; i++ {
		apiType, success := relayconstant.ChannelType2APIType(i)
		if !success || apiType == relayconstant.APITypeAIProxyLibrary {
			continue
		}
		meta := &relaycommon.RelayInfo{ChannelType: i}
		adaptor := relay.GetAdaptor(apiType)
		adaptor.Init(meta)
		channelId2Models[i] = adaptor.GetModelList()
	}
}

func ListModels(c *gin.Context) {
	userOpenAiModels := make([]dto.OpenAIModels, 0)
	permission := getPermission()
	modelPrefixMap := make(map[string][]string) // Map to store all prefixed versions for each model

	// Get the user group to look up available channels
	userId := c.GetInt("id")
	userGroup, err := model.GetUserGroup(userId, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "get user group failed",
		})
		return
	}
	group := userGroup
	tokenGroup := c.GetString("token_group")
	if tokenGroup != "" {
		group = tokenGroup
	}

	// Get channels with prefixes
	prefixChannels := middleware.GetPrefixChannels(group)
	for prefix, channels := range prefixChannels {
		if prefix == "" {
			continue // Skip channels without prefixes
		}

		// For each channel with a prefix, add its models to the prefix map
		for _, channel := range channels {
			for _, modelName := range channel.GetModels() {
				if modelPrefixMap[modelName] == nil {
					modelPrefixMap[modelName] = make([]string, 0)
				}
				prefixedModel := prefix + modelName
				// Check if this prefixed model already exists to avoid duplicates
				exists := false
				for _, existing := range modelPrefixMap[modelName] {
					if existing == prefixedModel {
						exists = true
						break
					}
				}
				if !exists {
					modelPrefixMap[modelName] = append(modelPrefixMap[modelName], prefixedModel)
				}
			}
		}
	}

	modelLimitEnable := c.GetBool("token_model_limit_enabled")
	if modelLimitEnable {
		s, ok := c.Get("token_model_limit")
		var tokenModelLimit map[string]bool
		if ok {
			tokenModelLimit = s.(map[string]bool)
		} else {
			tokenModelLimit = map[string]bool{}
		}

		for allowModel := range tokenModelLimit {
			// Check if this model has prefix mappings
			prefixedModels := modelPrefixMap[allowModel]
			
			// Add base model first
			if _, ok := openAIModelsMap[allowModel]; ok {
				modelInfo := openAIModelsMap[allowModel]
				userOpenAiModels = append(userOpenAiModels, modelInfo)
			} else {
				userOpenAiModels = append(userOpenAiModels, dto.OpenAIModels{
					Id:         allowModel,
					Object:     "model",
					Created:    1626777600,
					OwnedBy:    "custom",
					Permission: permission,
					Root:       allowModel,
					Parent:     nil,
				})
			}

			// Add all prefixed versions
			for _, prefixedModel := range prefixedModels {
				if _, ok := openAIModelsMap[allowModel]; ok {
					modelInfo := openAIModelsMap[allowModel]
					// Create a copy and replace ID with prefixed model
					prefixedModelInfo := modelInfo
					prefixedModelInfo.Id = prefixedModel
					userOpenAiModels = append(userOpenAiModels, prefixedModelInfo)
				} else {
					userOpenAiModels = append(userOpenAiModels, dto.OpenAIModels{
						Id:         prefixedModel,
						Object:     "model",
						Created:    1626777600,
						OwnedBy:    "custom",
						Permission: permission,
						Root:       allowModel, // Keep the original model name as root
						Parent:     nil,
					})
				}
			}
		}
	} else {
		models := model.GetGroupModels(group)
		addedModels := make(map[string]bool)

		// First add all models with prefixes
		for baseModel, prefixedModels := range modelPrefixMap {
			// Check if the base model is in the allowed models for this group
			isAllowed := false
			for _, groupModel := range models {
				if groupModel == baseModel {
					isAllowed = true
					break
				}
			}

			if isAllowed {
				// Add all prefixed versions of this model
				for _, prefixedModel := range prefixedModels {
					if _, ok := openAIModelsMap[baseModel]; ok {
						modelInfo := openAIModelsMap[baseModel]
						// Create a copy and replace ID with prefixed model
						prefixedModelInfo := modelInfo
						prefixedModelInfo.Id = prefixedModel
						userOpenAiModels = append(userOpenAiModels, prefixedModelInfo)
					} else {
						userOpenAiModels = append(userOpenAiModels, dto.OpenAIModels{
							Id:         prefixedModel,
							Object:     "model",
							Created:    1626777600,
							OwnedBy:    "custom",
							Permission: permission,
							Root:       baseModel,
							Parent:     nil,
						})
					}
				}
				addedModels[baseModel] = true
			}
		}

		// Then add remaining models without prefixes
		for _, s := range models {
			// Skip if already added with a prefix
			if addedModels[s] {
				continue
			}

			if _, ok := openAIModelsMap[s]; ok {
				userOpenAiModels = append(userOpenAiModels, openAIModelsMap[s])
			} else {
				userOpenAiModels = append(userOpenAiModels, dto.OpenAIModels{
					Id:         s,
					Object:     "model",
					Created:    1626777600,
					OwnedBy:    "custom",
					Permission: permission,
					Root:       s,
					Parent:     nil,
				})
			}
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    userOpenAiModels,
	})
}

func ChannelListModels(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    openAIModels,
	})
}

func DashboardListModels(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    channelId2Models,
	})
}

func EnabledListModels(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    model.GetEnabledModels(),
	})
}

func RetrieveModel(c *gin.Context) {
	modelId := c.Param("model")

	// Check if the model ID has a prefix
	userId := c.GetInt("id")
	userGroup, err := model.GetUserGroup(userId, true)
	if err == nil {
		group := userGroup
		tokenGroup := c.GetString("token_group")
		if tokenGroup != "" {
			group = tokenGroup
		}

		prefixChannels := middleware.GetPrefixChannels(group)
		for prefix := range prefixChannels {
			if prefix != "" && strings.HasPrefix(modelId, prefix) {
				// We found a model with a prefix, try to retrieve the base model
				baseModelId := strings.TrimPrefix(modelId, prefix)
				if aiModel, ok := openAIModelsMap[baseModelId]; ok {
					modelCopy := aiModel
					modelCopy.Id = modelId // Use the prefixed model ID
					c.JSON(200, modelCopy)
					return
				}
			}
		}
	}

	// No prefix found or base model doesn't exist, try the original model ID
	if aiModel, ok := openAIModelsMap[modelId]; ok {
		c.JSON(200, aiModel)
	} else {
		openAIError := dto.OpenAIError{
			Message: fmt.Sprintf("The model '%s' does not exist", modelId),
			Type:    "invalid_request_error",
			Param:   "model",
			Code:    "model_not_found",
		}
		c.JSON(200, gin.H{
			"error": openAIError,
		})
	}
}
