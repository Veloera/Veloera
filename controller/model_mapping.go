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
	"encoding/json"
	"net/http"
	"veloera/model"
	"veloera/service"

	"github.com/gin-gonic/gin"
)

// GetGlobalModelMapping 获取当前全局模型映射配置
func GetGlobalModelMapping(c *gin.Context) {
	mapping := service.GetGlobalModelMapping()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    mapping,
	})
}

// UpdateGlobalModelMapping 更新全局模型映射配置
func UpdateGlobalModelMapping(c *gin.Context) {
	var mapping model.GlobalModelMapping

	if err := c.ShouldBindJSON(&mapping); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数解析失败: " + err.Error(),
		})
		return
	}

	if err := service.UpdateGlobalModelMapping(&mapping); err != nil {
		// 将StatusInternalServerError改为StatusBadRequest，并更新错误信息
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "更新全局模型映射配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "全局模型映射配置更新成功", // 更新成功消息
		"data":    mapping,        // 返回更新后的配置
	})
}

// GetModelMappingConfig 获取模型映射配置的JSON字符串形式
func GetModelMappingConfig(c *gin.Context) {
	configStr := service.GlobalModelMappingToJSONString()

	// 尝试解析JSON字符串以验证格式
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "配置格式错误: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    config,
	})
}

// UpdateModelMappingConfig 通过JSON字符串更新模型映射配置
func UpdateModelMappingConfig(c *gin.Context) {
	// 读取请求体
	jsonStr, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "读取请求体失败: " + err.Error(),
		})
		return
	}

	// 如果请求体为空，设置为空字符串
	if len(jsonStr) == 0 {
		jsonStr = []byte("{}")
	}

	// 更新配置
	if err := service.UpdateGlobalModelMappingFromJSONString(string(jsonStr)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新全局模型映射配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "全局模型映射配置更新成功",
	})
}

// ReloadModelMapping 重新加载模型映射配置
func ReloadModelMapping(c *gin.Context) {
	if err := service.ReloadModelMapping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "重新加载模型映射配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "模型映射配置重新加载成功",
	})
}
