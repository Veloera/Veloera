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

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"veloera/model"
	"veloera/service"
)

func main() {
	// 初始化模型映射服务
	if err := service.InitializeModelMappingService(); err != nil {
		log.Fatalf("Failed to initialize model mapping service: %v", err)
	}

	// 创建示例模型映射配置
	exampleMapping := &model.GlobalModelMapping{
		Mapping: map[string][]model.ModelMappingItem{
			"gpt-3.5-turbo": {
				{Model: "gpt-3.5-turbo-0613", Priorities: 10, Weight: 30},
				{Model: "gpt-3.5-turbo-0301", Priorities: 10, Weight: 20},
				{Model: "gpt-3.5-turbo-16k", Priorities: 5, Weight: 10},
			},
			"gpt-4": {
				{Model: "gpt-4-0613", Priorities: 10, Weight: 40},
				{Model: "gpt-4-0314", Priorities: 8, Weight: 20},
			},
			"claude-instant": {
				{Model: "claude-instant-1.2", Priorities: 10, Weight: 30},
				{Model: "claude-instant-1.1", Priorities: 5, Weight: 15},
			},
		},
	}

	// 更新全局模型映射配置
	if err := service.UpdateGlobalModelMapping(exampleMapping); err != nil {
		log.Fatalf("Failed to update global model mapping: %v", err)
	}

	// 获取实际模型名
	testModels := []string{"gpt-3.5-turbo", "gpt-4", "claude-instant", "non-existent-model"}

	fmt.Println("模型映射测试:")
	for _, virtualModel := range testModels {
		actualModel, err := service.GetActualModel(virtualModel)
		if err != nil {
			log.Printf("Error getting actual model for %s: %v", virtualModel, err)
			continue
		}
		fmt.Printf("虚拟模型: %-20s -> 实际模型: %s\n", virtualModel, actualModel)
	}

	// 获取当前配置
	currentConfig := service.GetGlobalModelMapping()
	fmt.Println("\n当前全局模型映射配置:")
	jsonData, _ := json.MarshalIndent(currentConfig, "", "  ")
	fmt.Println(string(jsonData))

	// 重新加载配置
	if err := service.ReloadModelMapping(); err != nil {
		log.Printf("Failed to reload model mapping: %v", err)
	} else {
		fmt.Println("\n配置重新加载成功")
	}
}
