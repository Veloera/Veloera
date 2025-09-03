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
	"encoding/json"
	"fmt"
	"sync"

	"veloera/common"
)

// 全局变量
var (
	globalModelMapping *GlobalModelMapping
	// roundRobinCounter  *RoundRobinCounter // 移除轮询计数器，因为现在使用权重随机选择
	ModelMappingMutex sync.RWMutex
)

// init 包初始化函数
func init() {
	// 初始化全局变量，实际配置将在InitializeModelMapping中加载
	globalModelMapping = &GlobalModelMapping{
		Mapping: make(map[string][]ModelMappingItem),
	}
	// roundRobinCounter = NewRoundRobinCounter() // 移除轮询计数器初始化
}

// LoadModelMappingConfig 从数据库加载配置
func LoadModelMappingConfig() error {
	ModelMappingMutex.Lock()
	defer ModelMappingMutex.Unlock()

	return loadModelMappingConfigUnsafe()
}

// loadModelMappingConfigUnsafe 从数据库加载配置（内部调用，不加锁）
func loadModelMappingConfigUnsafe() error {
	// 从数据库加载配置
	common.OptionMapRWMutex.RLock()
	modelMappingStr, ok := common.OptionMap["global_model_mapping"]
	common.OptionMapRWMutex.RUnlock()

	if !ok || modelMappingStr == "" || modelMappingStr == "{}" {
		// 如果没有配置或配置为空，使用默认配置
		globalModelMapping = &GlobalModelMapping{
			Mapping: make(map[string][]ModelMappingItem),
		}
		return nil
	}

	// 解析JSON配置
	var mapping GlobalModelMapping
	if err := json.Unmarshal([]byte(modelMappingStr), &mapping); err != nil {
		common.SysError(fmt.Sprintf("Failed to parse global model mapping configuration: %v", err))
		return err
	}

	globalModelMapping = &mapping
	return nil
}

// GetGlobalModelMapping 获取全局模型映射配置的副本
func GetGlobalModelMapping() *GlobalModelMapping {
	ModelMappingMutex.RLock()
	defer ModelMappingMutex.RUnlock()

	if globalModelMapping == nil {
		return &GlobalModelMapping{
			Mapping: make(map[string][]ModelMappingItem),
		}
	}

	// 返回配置的副本以避免外部修改
	configCopy := &GlobalModelMapping{
		Mapping: make(map[string][]ModelMappingItem),
	}

	for k, v := range globalModelMapping.Mapping {
		itemsCopy := make([]ModelMappingItem, len(v))
		copy(itemsCopy, v)
		configCopy.Mapping[k] = itemsCopy
	}

	return configCopy
}

// GlobalModelMappingToJSONString 将全局模型映射配置转换为JSON字符串
func GlobalModelMappingToJSONString() string {
	ModelMappingMutex.RLock()
	defer ModelMappingMutex.RUnlock()

	if globalModelMapping == nil {
		return "{}"
	}

	jsonBytes, err := json.Marshal(globalModelMapping)
	if err != nil {
		common.SysError("Failed to serialize global model mapping configuration: " + err.Error())
		return "{}"
	}
	return string(jsonBytes)
}

// UpdateGlobalModelMappingFromJSONString 从JSON字符串更新全局模型映射配置
func UpdateGlobalModelMappingFromJSONString(jsonStr string) error {
	if jsonStr == "" || jsonStr == "{}" {
		// 如果配置为空，使用默认配置
		defaultMapping := &GlobalModelMapping{
			Mapping: make(map[string][]ModelMappingItem),
		}
		return UpdateGlobalModelMapping(defaultMapping)
	}

	// 解析JSON配置
	var mapping GlobalModelMapping
	if err := json.Unmarshal([]byte(jsonStr), &mapping); err != nil {
		common.SysError(fmt.Sprintf("Failed to parse global model mapping configuration: %v", err))
		return err
	}

	return UpdateGlobalModelMapping(&mapping)
}

// UpdateGlobalModelMapping 更新全局模型映射配置
func UpdateGlobalModelMapping(mapping *GlobalModelMapping) error {
	ModelMappingMutex.Lock()
	defer ModelMappingMutex.Unlock()

	// 验证配置格式
	if err := ValidateModelMapping(mapping); err != nil {
		return err
	}

	// 序列化配置为JSON
	jsonData, err := json.Marshal(mapping)
	if err != nil {
		common.SysError(fmt.Sprintf("Failed to serialize global model mapping configuration: %v", err))
		return err
	}

	// 保存配置到数据库
	if err := UpdateOption("global_model_mapping", string(jsonData)); err != nil {
		common.SysError(fmt.Sprintf("Failed to save global model mapping configuration to the database.: %v", err))
		return err
	}

	globalModelMapping = mapping

	// 重置轮询计数器 (已移除，无需操作)
	// roundRobinCounter.ClearAllCounters()

	common.SysLog("Global model mapping configuration updated successfully")
	return nil
}

// InitializeModelMapping 初始化模型映射
func InitializeModelMapping() error {
	ModelMappingMutex.Lock()
	defer ModelMappingMutex.Unlock()

	// 初始化轮询计数器 (已移除，无需操作)
	// roundRobinCounter = NewRoundRobinCounter()

	// 直接调用内部加载逻辑，避免重复加锁
	if err := loadModelMappingConfigUnsafe(); err != nil {
		common.SysError(fmt.Sprintf("Failed to load model mapping configuration: %v", err))
		return err
	}

	common.SysLog("Model mapping initialized successfully")
	return nil
}

// ReloadModelMapping 重新加载模型映射配置
func ReloadModelMapping() error {
	ModelMappingMutex.Lock()
	defer ModelMappingMutex.Unlock()

	// 重新加载配置
	if err := loadModelMappingConfigUnsafe(); err != nil {
		common.SysError(fmt.Sprintf("Failed to reload model mapping configuration: %v", err))
		return err
	}

	// 重置轮询计数器 (已移除，无需操作)
	// roundRobinCounter.ClearAllCounters()

	common.SysLog("Model mapping configuration reload successful")
	return nil
}

// GetAllVirtualModels 获取所有已配置的虚拟模型名称
func GetAllVirtualModels() []string {
	ModelMappingMutex.RLock()
	defer ModelMappingMutex.RUnlock()

	if globalModelMapping == nil || len(globalModelMapping.Mapping) == 0 {
		return []string{}
	}

	virtualModels := make([]string, 0, len(globalModelMapping.Mapping))
	for virtualModel := range globalModelMapping.Mapping {
		virtualModels = append(virtualModels, virtualModel)
	}

	return virtualModels
}
