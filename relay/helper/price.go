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
package helper

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"veloera/common"
	constant2 "veloera/constant"
	relaycommon "veloera/relay/common"
	"veloera/setting"
	"veloera/setting/operation_setting"
)

type PriceData struct {
	ModelPrice             float64
	ModelRatio             float64
	CompletionRatio        float64
	CacheRatio             float64
	GroupRatio             float64
	UsePrice               bool
	CacheCreationRatio     float64
	ShouldPreConsumedQuota int
}

func (p PriceData) ToSetting() string {
	return fmt.Sprintf("ModelPrice: %f, ModelRatio: %f, CompletionRatio: %f, CacheRatio: %f, GroupRatio: %f, UsePrice: %t, CacheCreationRatio: %f, ShouldPreConsumedQuota: %d", p.ModelPrice, p.ModelRatio, p.CompletionRatio, p.CacheRatio, p.GroupRatio, p.UsePrice, p.CacheCreationRatio, p.ShouldPreConsumedQuota)
}

func ModelPriceHelper(c *gin.Context, info *relaycommon.RelayInfo, promptTokens int, completionTokens int) (PriceData, error) {
	// Extract the base model name if it has a prefix
	modelName := info.OriginModelName
	modelNameForPrice := modelName
	modelNameForRatio := modelName

	// Check if the model has a prefix by looking at the difference between OriginModelName and UpstreamModelName
	if info.OriginModelName != info.UpstreamModelName {
		modelNameForPrice = info.UpstreamModelName
		modelNameForRatio = info.UpstreamModelName
	}

	modelPrice, usePrice := operation_setting.GetModelPriceWithFallback(modelNameForPrice, false)
	groupRatio := setting.GetGroupRatio(info.Group)
	var preConsumedQuota int
	var modelRatio float64
	var completionRatio float64
	var cacheRatio float64
	var cacheCreationRatio float64
	if !usePrice {
		preConsumedTokens := common.PreConsumedQuota
		if completionTokens != 0 {
			preConsumedTokens = promptTokens + completionTokens
		}
		var success bool
		modelRatio, success = operation_setting.GetModelRatioWithFallback(modelNameForRatio)
		if !success {
			acceptUnsetRatio := false
			if accept, ok := info.UserSetting[constant2.UserAcceptUnsetRatioModel]; ok {
				b, ok := accept.(bool)
				if ok {
					acceptUnsetRatio = b
				}
			}
			if !acceptUnsetRatio {
				return PriceData{}, fmt.Errorf("模型 %s 倍率或价格未配置，请联系管理员设置或开始自用模式；Model %s ratio or price not set, please set or start self-use mode", info.OriginModelName, info.OriginModelName)
			}
		}
		completionRatio = operation_setting.GetCompletionRatioWithFallback(modelNameForRatio)
		cacheRatio, _ = operation_setting.GetCacheRatio(modelNameForRatio)
		cacheCreationRatio, _ = operation_setting.GetCreateCacheRatio(modelNameForRatio)
		ratio := modelRatio * groupRatio
		preConsumedQuota = int(float64(preConsumedTokens) * ratio)
	} else {
		preConsumedQuota = int(modelPrice * common.QuotaPerUnit * groupRatio)
	}

	priceData := PriceData{
		ModelPrice:             modelPrice,
		ModelRatio:             modelRatio,
		CompletionRatio:        completionRatio,
		GroupRatio:             groupRatio,
		UsePrice:               usePrice,
		CacheRatio:             cacheRatio,
		CacheCreationRatio:     cacheCreationRatio,
		ShouldPreConsumedQuota: preConsumedQuota,
	}

	if common.DebugEnabled {
		println(fmt.Sprintf("model_price_helper result: %s", priceData.ToSetting()))
	}

	return priceData, nil
}

func ContainPriceOrRatio(modelName string) bool {
	_, ok := operation_setting.GetModelPriceWithFallback(modelName, false)
	if ok {
		return true
	}
	_, ok = operation_setting.GetModelRatioWithFallback(modelName)
	if ok {
		return true
	}
	return false
}
