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
	"github.com/gin-gonic/gin"
	"veloera/middleware"
	"veloera/model"
	"veloera/setting"
	"veloera/setting/operation_setting"
)

func GetPricing(c *gin.Context) {
	pricing := model.GetPricing()
	userId, exists := c.Get("id")
	usableGroup := map[string]string{}
	groupRatio := map[string]float64{}
	for s, f := range setting.GetGroupRatioCopy() {
		groupRatio[s] = f
	}
	var group string
	if exists {
		user, err := model.GetUserCache(userId.(int))
		if err == nil {
			group = user.Group
		}
	}

	usableGroup = setting.GetUserUsableGroups(group)
	// check groupRatio contains usableGroup
	for group := range setting.GetGroupRatioCopy() {
		if _, ok := usableGroup[group]; !ok {
			delete(groupRatio, group)
		}
	}

	// Now add models with prefixes to the pricing data
	pricingWithPrefixes := enhancePricingWithPrefixes(pricing, group)

	c.JSON(200, gin.H{
		"success":      true,
		"data":         pricingWithPrefixes,
		"group_ratio":  groupRatio,
		"usable_group": usableGroup,
	})
}

// enhancePricingWithPrefixes adds models with prefixes to the pricing data
// and removes non-prefixed models that are also available with prefixes
func enhancePricingWithPrefixes(pricing []model.Pricing, group string) []model.Pricing {
	// Get all channels with prefixes for this group
	prefixChannels := middleware.GetPrefixChannels(group)

	// Track which models have prefixed versions
	modelsWithPrefix := make(map[string]bool)

	// First, create a list of prefixed models
	var prefixedPricingModels []model.Pricing

	for prefix, channels := range prefixChannels {
		if prefix == "" {
			continue // Skip channels without prefixes
		}

		// For each channel with a prefix, add models with that prefix
		for _, channel := range channels {
			for _, channelModel := range channel.GetModels() {
				// Look up the base model in the pricing data
				for _, baseModelPricing := range pricing {
					if baseModelPricing.ModelName == channelModel {
						// Create a new pricing entry for the prefixed model
						prefixedPricing := baseModelPricing
						prefixedPricing.ModelName = prefix + channelModel

						// Add the new prefixed model to the result
						prefixedPricingModels = append(prefixedPricingModels, prefixedPricing)

						// Mark this model as having a prefixed version
						modelsWithPrefix[channelModel] = true
						break
					}
				}
			}
		}
	}

	// Now create the result with:
	// 1. All prefixed models
	// 2. Only those non-prefixed models that aren't also available with prefixes
	// 3. Or are available both with and without prefixes
	var result []model.Pricing

	// First add all prefixed models
	result = append(result, prefixedPricingModels...)

	// Check if non-prefixed models are available from non-prefixed channels
	nonPrefixedChannels := prefixChannels[""]
	modelsFromNonPrefixedChannels := make(map[string]bool)

	for _, channel := range nonPrefixedChannels {
		for _, model := range channel.GetModels() {
			modelsFromNonPrefixedChannels[model] = true
		}
	}

	// Now add original models that either:
	// - Don't have a prefixed version
	// - AND are also available directly (without prefix)
	for _, modelPricing := range pricing {
		modelName := modelPricing.ModelName

		// Only add original model if it doesn't have a prefixed version AND is available from non-prefixed channels
		// This prevents duplicate display when a model has both prefixed and non-prefixed versions
		if !modelsWithPrefix[modelName] && modelsFromNonPrefixedChannels[modelName] {
			result = append(result, modelPricing)
		}
	}

	return result
}

func ResetModelRatio(c *gin.Context) {
	defaultStr := operation_setting.DefaultModelRatio2JSONString()
	err := model.UpdateOption("ModelRatio", defaultStr)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = operation_setting.UpdateModelRatioByJSONString(defaultStr)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "重置模型倍率成功",
	})
}
