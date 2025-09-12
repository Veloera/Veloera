/*
Copyright (c) 2025 Tethys Plex

This file is part of Veloera.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

/**
 * 模型映射工具类
 * 用于处理模型名称映射和转换逻辑
 */
export class ModelMappingUtils {
  /**
   * 解析模型映射配置
   * @param {string} mappingValue - JSON 格式的映射配置字符串
   * @returns {Object|null} 解析后的映射对象，解析失败返回 null
   */
  static parseModelMapping(mappingValue) {
    if (!mappingValue || typeof mappingValue !== 'string' || mappingValue.trim() === '') {
      return null;
    }

    try {
      const mapping = JSON.parse(mappingValue);
      if (typeof mapping !== 'object' || mapping === null) {
        return null;
      }
      return mapping;
    } catch (error) {
      console.warn('模型重定向 JSON 解析失败:', error);
      return null;
    }
  }

  /**
   * 应用模型映射的核心逻辑
   * @param {Object} mapping - 映射配置对象
   * @param {Array} currentModels - 当前模型列表
   * @returns {Object} 包含更新后的模型列表、新映射关系和是否有变更的对象
   */
  static applyModelMapping(mapping, currentModels) {
    if (!mapping || typeof mapping !== 'object') {
      return { updatedModels: currentModels, newMapping: {}, hasChanges: false };
    }

    const updatedModels = new Set(currentModels);
    const newMapping = Object.create(null);
    let hasChanges = false;

    // 处理新的映射关系
    Object.entries(mapping).forEach(([displayName, originalName]) => {
      const displayNameTrimmed = displayName.trim();
      const originalNameTrimmed = originalName?.trim();

      if (displayNameTrimmed && originalNameTrimmed) {
        // 添加显示名称到模型列表中（如果不存在）
        if (!updatedModels.has(displayNameTrimmed)) {
          updatedModels.add(displayNameTrimmed);
          hasChanges = true;
        }

        // 移除原始名称，用显示名称代替
        if (originalNameTrimmed !== displayNameTrimmed) {
          if (updatedModels.delete(originalNameTrimmed)) {
            hasChanges = true;
          }
        }

        // 建立映射关系：displayName -> originalName
        newMapping[displayNameTrimmed] = originalNameTrimmed;
      }
    });

    return {
      updatedModels: Array.from(updatedModels),
      newMapping,
      hasChanges
    };
  }

  /**
   * 恢复模型名称到原始名称
   * @param {Array} currentModels - 当前模型列表
   * @param {Object} originalMapping - 原始映射关系
   * @returns {Array} 恢复后的模型列表
   */
  static restoreModelsToOriginal(currentModels, originalMapping) {
    if (!currentModels || currentModels.length === 0) {
      return [];
    }

    if (!originalMapping || Object.keys(originalMapping).length === 0) {
      return currentModels;
    }

    // 使用无原型对象并安全访问属性，避免原型污染影响
    const safeMapping = Object.assign(Object.create(null), originalMapping);
    const restoredModels = currentModels.map((model) => {
      const key = typeof model === 'string' ? model.trim() : model;
      return Object.prototype.hasOwnProperty.call(safeMapping, key)
        ? safeMapping[key]
        : key;
    });

    // 去重
    return Array.from(new Set(restoredModels));
  }

  /**
   * 检查两个模型列表是否相同
   * @param {Array} models1 - 第一个模型列表
   * @param {Array} models2 - 第二个模型列表
   * @returns {boolean} 是否相同
   */
  static areModelsEqual(models1, models2) {
    if (!Array.isArray(models1) || !Array.isArray(models2)) {
      return false;
    }

    if (models1.length !== models2.length) {
      return false;
    }

    const sorted1 = [...models1].sort();
    const sorted2 = [...models2].sort();

    return JSON.stringify(sorted1) === JSON.stringify(sorted2);
  }

  /**
   * 过滤并去重模型列表
   * @param {Array} models - 模型列表
   * @returns {Array} 过滤后的模型列表
   */
  static filterAndDedupeModels(models) {
    if (!Array.isArray(models)) {
      return [];
    }

    return Array.from(new Set(models.filter((m) => typeof m === 'string').map((m) => m.trim()).filter(Boolean)));
  }
}