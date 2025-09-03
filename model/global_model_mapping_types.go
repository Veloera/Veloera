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

package model

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ModelMappingItem 模型映射项
type ModelMappingItem struct {
	Model      string `json:"model" binding:"required"`   // 实际模型名
	Priorities int    `json:"priorities" binding:"min=0"` // 优先级(非负整数)
	Weight     int    `json:"weight" binding:"min=0"`     // 权重(非负整数，0权重会默认设置为10)
}

// GlobalModelMapping 全局模型映射配置
type GlobalModelMapping struct {
	Mapping map[string][]ModelMappingItem `json:"mapping"` // 虚拟模型名 -> 实际模型映射列表
}

// GetActualModel 根据虚拟模型名获取实际模型名（考虑优先级和权重）
func GetActualModel(virtualModel string) (string, error) {
	ModelMappingMutex.RLock()
	defer ModelMappingMutex.RUnlock()

	if globalModelMapping == nil || len(globalModelMapping.Mapping) == 0 {
		// 如果没有配置映射，直接返回原模型名
		return virtualModel, nil
	}

	mappingItems, exists := globalModelMapping.Mapping[virtualModel]
	if !exists || len(mappingItems) == 0 {
		// 如果虚拟模型没有映射配置，直接返回原模型名
		return virtualModel, nil
	}

	// 按优先级分组
	priorityGroups := make(map[int][]ModelMappingItem)
	for _, item := range mappingItems {
		if item.Priorities < 0 {
			// 跳过无效的优先级（负数）
			continue
		}
		// 确保权重至少为10（避免默认0的权重无法选取）
		if item.Weight < 10 {
			item.Weight = 10
		}
		priorityGroups[item.Priorities] = append(priorityGroups[item.Priorities], item)
	}

	if len(priorityGroups) == 0 {
		return virtualModel, nil
	}

	// 获取最高优先级
	var maxPriority int = -1
	for priority := range priorityGroups {
		if priority > maxPriority {
			maxPriority = priority
		}
	}

	highestPriorityItems := priorityGroups[maxPriority]
	if len(highestPriorityItems) == 0 {
		return virtualModel, nil
	}

	// 如果只有一个最高优先级项目，直接返回
	if len(highestPriorityItems) == 1 {
		return highestPriorityItems[0].Model, nil
	}

	// 根据权重随机选择
	return selectModelByWeight(highestPriorityItems)
}

// selectModelByWeight 根据权重随机选择模型
func selectModelByWeight(items []ModelMappingItem) (string, error) {
	// 计算总权重
	totalWeight := 0
	for _, item := range items {
		totalWeight += item.Weight
	}

	if totalWeight == 0 {
		// 如果所有权重都为0，返回第一个模型或错误
		if len(items) > 0 {
			return items[0].Model, nil
		}
		return "", fmt.Errorf("没有可用的模型进行权重选择")
	}

	// 生成随机数
	rand.Seed(time.Now().UnixNano())
	randomWeight := rand.Intn(totalWeight)

	// 根据权重选择模型
	currentWeight := 0
	for _, item := range items {
		currentWeight += item.Weight
		if randomWeight < currentWeight {
			return item.Model, nil
		}
	}

	// 理论上不应到达此处，作为备用返回最后一个
	if len(items) > 0 {
		return items[len(items)-1].Model, nil
	}
	return "", fmt.Errorf("未能根据权重选择模型")
}

// ValidateModelMapping 验证模型映射配置
func ValidateModelMapping(mapping *GlobalModelMapping) error {
	if mapping == nil {
		return fmt.Errorf("映射配置不能为空")
	}

	if mapping.Mapping == nil {
		return fmt.Errorf("映射字典不能为空")
	}

	for virtualModel, items := range mapping.Mapping {
		if strings.TrimSpace(virtualModel) == "" {
			return fmt.Errorf("虚拟模型名不能为空")
		}

		if len(items) == 0 {
			return fmt.Errorf("虚拟模型 '%s' 的映射项不能为空", virtualModel)
		}

		// 用于检查同一虚拟模型内实际模型名的重复
		modelSet := make(map[string]bool)

		for i, item := range items {
			if strings.TrimSpace(item.Model) == "" {
				return fmt.Errorf("虚拟模型 '%s' 的第 %d 个映射项的实际模型名不能为空", virtualModel, i+1)
			}

			if item.Priorities < 0 {
				return fmt.Errorf("虚拟模型 '%s' 的实际模型 '%s' 的优先级必须为非负整数", virtualModel, item.Model)
			}

			if item.Weight < 0 {
				return fmt.Errorf("虚拟模型 '%s' 的实际模型 '%s' 的权重必须为非负整数", virtualModel, item.Model)
			}

			// 检查同一虚拟模型内实际模型名的重复
			trimmedModel := strings.TrimSpace(item.Model)
			if modelSet[trimmedModel] {
				return fmt.Errorf("虚拟模型 '%s' 中存在重复的实际模型名 '%s'", virtualModel, trimmedModel)
			}
			modelSet[trimmedModel] = true
		}
	}

	return nil
}
