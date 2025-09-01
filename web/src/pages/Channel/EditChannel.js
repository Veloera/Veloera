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
import React, { useEffect, useRef, useState, useMemo, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  API,
  isMobile,
  showError,
  showInfo,
  showSuccess,
  showWarning,
  verifyJSON,
} from '../../helpers';
import { CHANNEL_OPTIONS } from '../../constants';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import {
  SideSheet,
  Space,
  Spin,
  Button,
  Tooltip,
  Input,
  Typography,
  Select,
  TextArea,
  Checkbox,
  Banner,
  Modal,
  Checkbox as SemiCheckbox,
  Row,
  Col,
} from '@douyinfe/semi-ui';
import { getChannelModels, loadChannelModels } from '../../components/utils.js';
import {
  IconEyeOpened,
  IconEyeClosedSolid,
  IconRefresh,
  IconPlusCircle,
  IconMinusCircle,
  IconCopy
} from '@douyinfe/semi-icons';

// ModelSelector component for advanced model selection
const ModelSelector = ({ channelId, type, apiKey, baseUrl, isEdit, selectedModels, onSelect }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [localSelectedModels, setLocalSelectedModels] = useState([...selectedModels]);
  const [search, setSearch] = useState('');
  const [availableModels, setAvailableModels] = useState([]);

  // Create options from available models
  const allOptions = availableModels.map(model => ({
    label: model,
    value: model
  }));

  // Filter models based on search input
  const filteredOptions = allOptions.filter(option =>
    option.label.toLowerCase().includes(search.toLowerCase())
  );

  // Handle check/uncheck of individual model
  const handleCheckboxChange = (value) => {
    if (localSelectedModels.includes(value)) {
      setLocalSelectedModels(localSelectedModels.filter(model => model !== value));
    } else {
      setLocalSelectedModels([...localSelectedModels, value]);
    }
  };

  // Select all visible models
  const handleSelectAll = () => {
    const allValues = filteredOptions.map(option => option.value);
    // Merge with existing selection
    const newSelection = [...new Set([...localSelectedModels, ...allValues])];
    setLocalSelectedModels(newSelection);
  };

  // Invert selection for all visible models
  const handleDeselectAll = () => {
    const visibleValues = new Set(filteredOptions.map(option => option.value));
    const newSelection = [...localSelectedModels];

    // For each visible option, toggle its selection state
    filteredOptions.forEach(option => {
      const value = option.value;
      const index = newSelection.indexOf(value);
      if (index === -1) {
        // If not selected, add it
        newSelection.push(value);
      } else {
        // If selected, remove it
        newSelection.splice(index, index + 1); // Use index + 1 for correct splice
      }
    });

    setLocalSelectedModels(newSelection);
  };

  // Fetch models from API - using the same logic as fetchUpstreamModelList
  const fetchModels = async () => {
    try {
      setLoading(true);
      const models = [...localSelectedModels]; // Keep existing selections
      let res;

      if (isEdit && channelId) {
        // 如果在编辑模式且有channelId，使用后端根据channelId获取已保存渠道的模型列表
        // res = await API.post('/api/channel/fetch_models', {
        //   base_url: baseUrl,
        //   type: type,
        //   key: apiKey.split(',')[0].trim(),
        // });
        res = await API.get(`/api/channel/fetch_models/${channelId}`);
      } else {
        // 如果在创建模式，使用提供的凭据
        if (!apiKey) {
          showError(t('请填写密钥'));
          setLoading(false);
          return;
        }

        res = await API.post('/api/channel/fetch_models', {
          base_url: baseUrl,
          type: type,
          key: apiKey.split(',')[0].trim(),
        });
      }

      if (res.data && res.data.success) {
        // Get models from the response
        let fetchedModels = [];

        if (Array.isArray(res.data.data)) {
          fetchedModels = res.data.data;
        } else if (res.data.data && Array.isArray(res.data.data.data)) {
          fetchedModels = res.data.data.data;
        }

        // Update available models
        setAvailableModels(fetchedModels);

        // Show success message
        showSuccess(t('获取模型列表成功'));
      } else {
        showError(t('获取模型列表失败'));
      }
    } catch (error) {
      showError(error.message || t('获取模型列表失败'));
    } finally {
      setLoading(false);
    }
  };

  // Apply selection and close modal
  const applySelection = () => {
    onSelect(localSelectedModels);
    Modal.destroyAll();
  };

  // Load models when component mounts
  useEffect(() => {
    fetchModels();
  }, []);

  return (
    <div>
      <div style={{ display: 'flex', marginBottom: 16, alignItems: 'center' }}>
        <Input
          placeholder={t('搜索模型')}
          value={search}
          onChange={setSearch}
          style={{ flex: 1 }}
          showClear
        />
        <Button
          icon={<IconRefresh />}
          onClick={fetchModels}
          loading={loading}
          style={{ marginLeft: 8 }}
        />
        <Button onClick={handleSelectAll} style={{ marginLeft: 8 }}>{t('全选')}</Button>
        <Button onClick={handleDeselectAll} style={{ marginLeft: 8 }}>{t('反选')}</Button>
      </div>

      <div style={{ maxHeight: '350px', overflowY: 'auto', border: '1px solid var(--semi-color-border)', padding: 8 }}>
        <Row>
          {filteredOptions.map((option) => (
            <Col span={6} key={option.value} style={{ marginBottom: 8 }}>
              <SemiCheckbox
                checked={localSelectedModels.includes(option.value)}
                onChange={() => handleCheckboxChange(option.value)}
                style={{ width: '100%' }}
              >
                <Typography.Text
                  ellipsis={{ showTooltip: true }}
                  style={{
                    maxWidth: '100%',
                    wordBreak: 'break-word',
                    whiteSpace: 'normal',
                    lineHeight: '1.2'
                  }}
                >
                  {option.label}
                </Typography.Text>
              </SemiCheckbox>
            </Col>
          ))}
        </Row>
      </div>

      <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 16, marginBottom: 8 }}>
        <Button type='primary' onClick={applySelection}>{t('确定')}</Button>
        <Button style={{ marginLeft: 8 }} onClick={() => Modal.destroyAll()}>{t('取消')}</Button>
      </div>
    </div>
  );
};

const MODEL_MAPPING_EXAMPLE = {
  'gpt-3.5-turbo': 'gpt-3.5-turbo-0125',
};

// ModelMappingEditor component for visual key-value editing
const ModelMappingEditor = ({ value, onChange, placeholder, onRealtimeChange }) => {
  const { t } = useTranslation();
  const [mappingPairs, setMappingPairs] = useState([]);
  const [mode, setMode] = useState('visual'); // 'visual' or 'json'
  const [jsonValue, setJsonValue] = useState('');
  const [jsonError, setJsonError] = useState('');

  // 用于标记是否是内部更新，避免循环更新
  const isInternalUpdateRef = useRef(false);

  // Parse JSON value to key-value pairs
  const parseJsonToMappings = (jsonStr) => {
    if (!jsonStr || jsonStr.trim() === '') {
      return [];
    }
    try {
      const parsed = JSON.parse(jsonStr);
      if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
        const unsafe = new Set(['__proto__', 'prototype', 'constructor']);
        return Object.entries(parsed)
          .filter(([k]) => !unsafe.has(k))
          .map(([key, value]) => ({
          id: Date.now() + Math.random(),
          key,
          value
        }));
      }
      return [];
    } catch {
      return [];
    }
  };

  // Convert key-value pairs to JSON string
  const mappingsToJson = (pairs) => {
    if (!pairs || pairs.length === 0) {
      return '';
    }
    const obj = Object.create(null);
    const unsafe = new Set(['__proto__', 'prototype', 'constructor']);
    pairs.forEach(pair => {
      if (pair.key && pair.key.trim() !== '') {
        const k = pair.key.trim();
        if (!unsafe.has(k)) obj[k] = pair.value || '';
      }
    });
    return Object.keys(obj).length > 0 ? JSON.stringify(obj, null, 2) : '';
  };

  // Initialize component state from value prop
  useEffect(() => {
    // 只有在非内部更新时才同步外部状态
    if (!isInternalUpdateRef.current) {
      const pairs = parseJsonToMappings(value);
      setMappingPairs(pairs.length > 0 ? pairs : [{ id: Date.now() + Math.random(), key: '', value: '' }]);
      setJsonValue(value || '');
      setJsonError('');
    }
    isInternalUpdateRef.current = false;
  }, [value]);

  // Add new mapping pair
  const addMappingPair = () => {
    const newPairs = [...mappingPairs, { id: Date.now() + Math.random(), key: '', value: '' }];
    setMappingPairs(newPairs);
  };

  // Remove mapping pair
  const removeMappingPair = (index) => {
    const newPairs = mappingPairs.filter((_, i) => i !== index);
    const finalPairs = newPairs.length > 0 ? newPairs : [{ id: Date.now() + Math.random(), key: '', value: '' }];
    setMappingPairs(finalPairs);

    // 立即更新父组件
    const jsonStr = mappingsToJson(finalPairs);
    isInternalUpdateRef.current = true;
    onChange(jsonStr);

    // 触发实时同步
    if (onRealtimeChange) {
      onRealtimeChange(jsonStr);
    }
  };

  // Update mapping pair - 只更新本地状态，不触发父组件更新
  const updateMappingPair = (index, field, value) => {
    const newPairs = [...mappingPairs];
    newPairs[index] = { ...newPairs[index], [field]: value };
    setMappingPairs(newPairs);

    // 触发实时同步回调，但不更新父组件的 model_mapping 状态
    if (onRealtimeChange) {
      const jsonStr = mappingsToJson(newPairs);
      onRealtimeChange(jsonStr);
    }
  };

  // 在失去焦点时同步到父组件
  const handleInputBlur = () => {
    const jsonStr = mappingsToJson(mappingPairs);
    isInternalUpdateRef.current = true;
    onChange(jsonStr);
  };

  // Handle mode switch
  const switchMode = (newMode) => {
    if (newMode === 'json' && mode === 'visual') {
      // Switching from visual to JSON
      const jsonStr = mappingsToJson(mappingPairs);
      setJsonValue(jsonStr);
      setJsonError('');
    } else if (newMode === 'visual' && mode === 'json') {
      // Switching from JSON to visual
      try {
        if (jsonValue.trim() === '') {
          setMappingPairs([{ id: Date.now() + Math.random(), key: '', value: '' }]);
          setJsonError('');
        } else {
          const parsed = JSON.parse(jsonValue);
          if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
            const unsafe = new Set(['__proto__', 'prototype', 'constructor']);
              const pairs = Object.entries(parsed)
                .filter(([k]) => !unsafe.has(k))
                .map(([key, value]) => ({
                id: Date.now() + Math.random(),
                key,
                value
            }));
            setMappingPairs(pairs.length > 0 ? pairs : [{ id: Date.now() + Math.random(), key: '', value: '' }]);
            setJsonError('');
          } else {
            setJsonError(t('请输入有效的JSON对象格式'));
            return;
          }
        }
      } catch (error) {
        setJsonError(t('JSON格式错误: {{message}}', { message: error.message }));
        return;
      }
    }
    setMode(newMode);
  };

  // Handle JSON input change
  const handleJsonChange = (newValue) => {
    setJsonValue(newValue);

    // Validate JSON
    if (newValue.trim() === '') {
      setJsonError('');
      // 触发实时同步
      if (onRealtimeChange) {
        onRealtimeChange('');
      }
      return;
    }

    try {
      const parsed = JSON.parse(newValue);
      if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
        setJsonError('');
        // 触发实时同步
        if (onRealtimeChange) {
          onRealtimeChange(newValue);
        }
      } else {
        setJsonError(t('请输入有效的JSON对象格式'));
      }
    } catch (error) {
      setJsonError(t('JSON格式错误: {{message}}', { message: error.message }));
    }
  };

  // 在 JSON 模式失去焦点时同步到父组件
  const handleJsonBlur = () => {
    if (!jsonError && jsonValue.trim() !== '') {
      try {
        const parsed = JSON.parse(jsonValue);
        if (typeof parsed === 'object' && parsed !== null && !Array.isArray(parsed)) {
          isInternalUpdateRef.current = true;
          onChange(jsonValue);
        }
      } catch (error) {
        // 忽略错误，不更新父组件
      }
    } else if (jsonValue.trim() === '') {
      isInternalUpdateRef.current = true;
      onChange('');
    }
  };

  // Fill template
  const fillTemplate = () => {
    const templateJson = JSON.stringify(MODEL_MAPPING_EXAMPLE, null, 2);
    if (mode === 'visual') {
      const pairs = parseJsonToMappings(templateJson);
      setMappingPairs(pairs);
    } else {
      setJsonValue(templateJson);
    }

    // 立即更新父组件和触发实时同步
    isInternalUpdateRef.current = true;
    onChange(templateJson);
    if (onRealtimeChange) {
      onRealtimeChange(templateJson);
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', marginBottom: 10 }}>
        <div style={{ display: 'flex', marginRight: 16 }}>
          <Button
            type={mode === 'visual' ? 'primary' : 'tertiary'}
            onClick={() => switchMode('visual')}
            style={{
              borderRadius: '6px 0 0 6px'
            }}
          >
            {t('可视化编辑')}
          </Button>
          <Button
            type={mode === 'json' ? 'primary' : 'tertiary'}
            onClick={() => switchMode('json')}
            style={{
              borderRadius: '0 6px 6px 0'
            }}
          >
            {t('JSON编辑')}
          </Button>
        </div>
        <Typography.Text
          style={{
            color: 'rgba(var(--semi-blue-5), 1)',
            userSelect: 'none',
            cursor: 'pointer',
          }}
          onClick={fillTemplate}
        >
          {t('填入模板')}
        </Typography.Text>
      </div>

      {mode === 'visual' ? (
        <div>
          <div style={{ marginBottom: 10 }}>
            <Typography.Text type="secondary">{placeholder}</Typography.Text>
          </div>

          {mappingPairs.map((pair, index) => (
            <div key={pair.id} style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
              <Input
                placeholder={t('目标模型名称')}
                value={pair.key}
                onChange={(value) => updateMappingPair(index, 'key', value)}
                onBlur={() => handleInputBlur(index, 'key')}
                style={{ flex: 1, marginRight: 8 }}
              />
              <Typography.Text style={{ margin: '0 8px' }}>→</Typography.Text>
              <Input
                placeholder={t('实际模型名称')}
                value={pair.value}
                onChange={(value) => updateMappingPair(index, 'value', value)}
                onBlur={() => handleInputBlur(index, 'value')}
                style={{ flex: 1, marginRight: 8 }}
              />
              <Button
                type="danger"
                icon={<IconMinusCircle />}
                size="small"
                onClick={() => removeMappingPair(index)}
                style={{ marginLeft: 4 }}
              />
            </div>
          ))}

          <Button
            type="tertiary"
            icon={<IconPlusCircle />}
            onClick={addMappingPair}
            style={{ width: '100%', marginTop: 8 }}
          >
            {t('添加映射')}
          </Button>
        </div>
      ) : (
        <div>
          <TextArea
            placeholder={
              t(
                '此项可选，用于修改请求体中的模型名称，为一个 JSON 字符串，键为请求中模型名称，值为要替换的模型名称，例如：',
              ) + `\n${JSON.stringify(MODEL_MAPPING_EXAMPLE, null, 2)}`
            }
            value={jsonValue}
            onChange={handleJsonChange}
            onBlur={handleJsonBlur}
            autosize
            autoComplete='new-password'
          />
          {jsonError && (
            <Typography.Text type="danger" style={{ marginTop: 4, display: 'block' }}>
              {jsonError}
            </Typography.Text>
          )}
        </div>
      )}
    </div>
  );
};

const STATUS_CODE_MAPPING_EXAMPLE = {
  400: '500',
};

const REGION_EXAMPLE = {
  default: 'us-central1',
  'claude-3-5-sonnet-20240620': 'europe-west1',
};

function type2secretPrompt(type) {
  switch (type) {
    case 15:
      return '按照如下格式输入：APIKey|SecretKey，多个密钥使用英文逗号分隔';
    case 18:
      return '按照如下格式输入：APPID|APISecret|APIKey，多个密钥使用英文逗号分隔';
    case 22:
      return '按照如下格式输入：APIKey-AppId，例如：fastgpt-0sp2gtvfdgyi4k30jwlgwf1i-64f335d84283f05518e9e041，多个密钥使用英文逗号分隔';
    case 23:
      return '按照如下格式输入：AppId|SecretId|SecretKey，多个密钥使用英文逗号分隔';
    case 33:
      return '按照如下格式输入：Ak|Sk|Region，多个密钥使用英文逗号分隔';
    default:
      return '请输入渠道对应的鉴权密钥，多个密钥使用英文逗号分隔';
  }
}

// Function to check if a channel type supports multi-key view (excluding type 41 which is textarea)
const supportsMultiKeyView = (type) => {
  // Multi-key view is supported for all types except type 41 (which uses a textarea)
  return type !== 41;
};

const EditChannel = (props) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const channelId = props.editingChannel.id;
  const isEdit = channelId !== undefined;
  const [loading, setLoading] = useState(isEdit);
  const [showKey, setShowKey] = useState(false);
  const [initialKey, setInitialKey] = useState('');
  const [keyList, setKeyList] = useState([]);
  const [useKeyListMode, setUseKeyListMode] = useState(false);
  const [disableMultiKeyView, setDisableMultiKeyView] = useState(false);

  // 缓存有效密钥的计算结果
  const validKeys = useMemo(() => {
    return keyList.filter(key => key && key.trim());
  }, [keyList]);

  // Ref to store the input element that triggered the switch to list mode
  const singleKeyInputRef = useRef(null);

  const handleCancel = () => {
    props.handleClose();
  };

  const originInputs = {
    name: '',
    type: 1,
    key: '',
    openai_organization: '',
    max_input_tokens: 0,
    base_url: '',
    other: '',
    model_mapping: '',
    status_code_mapping: '',
    models: [],
    auto_ban: 1,
    test_model: '',
    groups: ['default'],
    priority: 0,
    weight: 0,
    tag: '',
    model_prefix: '', // Added model_prefix
    setting: '', // Added setting
    param_override: '', // Added param_override
    system_prompt: '', // Added system_prompt
  };

  const [batch, setBatch] = useState(false);
  const [autoBan, setAutoBan] = useState(true);
  const [inputs, setInputs] = useState(originInputs);
  const [originalModelMapping, setOriginalModelMapping] = useState(''); // Save the original model_mapping data
  const [componentResetKey, setComponentResetKey] = useState(0); // Used to force component reset
  const [originModelOptions, setOriginModelOptions] = useState([]);
  const [modelOptions, setModelOptions] = useState([]);
  const [groupOptions, setGroupOptions] = useState([]);
  const [basicModels, setBasicModels] = useState([]);
  const [fullModels, setFullModels] = useState([]);
  const [customModel, setCustomModel] = useState('');

  // 用于追踪模型的原始名称映射关系 { displayName: originalName }
  const [modelOriginalMapping, setModelOriginalMapping] = useState({});

  // 解析模型映射配置的工具函数
  const parseModelMapping = (mappingValue) => {
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
  };

  // 获取当前模型列表的工具函数
  const getCurrentModels = () => {
    return inputs.models || [];
  };

  // 更新模型列表的统一方法
  const updateModelsList = (newModels, newMapping) => {
    const uniqueModels = Array.from(new Set(newModels.filter(model => model && model.trim())));

    setInputs((prevInputs) => ({ ...prevInputs, models: uniqueModels }));
    setModelOriginalMapping(newMapping);
  };

  // 应用模型映射的核心逻辑
  const applyModelMapping = (mapping, currentModels, currentMapping) => {
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

        // 建立映射关系：displayName -> originalName
        newMapping[displayNameTrimmed] = originalNameTrimmed;
      }
    });

    const result = {
      updatedModels: Array.from(updatedModels),
      newMapping,
      hasChanges
    };
    return result;
  };

  // 实时同步模型重定向到模型配置的函数（用于实时预览）
  const syncModelMappingToModels = useCallback((mappingValue) => {
    const mapping = parseModelMapping(mappingValue);

    if (!mapping || Object.keys(mapping).length === 0) {
      // 当映射为空时，恢复为原始名称
      const currentModels = getCurrentModels();

      // 模型列表为空，则不需要恢复
      if (!currentModels || currentModels.length === 0) {
        return;
      }

      // 没有映射关系，则不需要恢复
      if (!modelOriginalMapping || Object.keys(modelOriginalMapping).length === 0) {
        updateModelsList(currentModels, {});
        return;
      }

      // 当删除映射时，需要将显示名称（key）替换为原始名称（value）
      const restoredModels = currentModels.map(model => {
        if (modelOriginalMapping[model]) {
          return modelOriginalMapping[model];
        }
        return model;
      });

      // 去重
      const uniqueRestoredModels = Array.from(new Set(restoredModels));

      // 清空原始映射关系，因为已经恢复完成
      updateModelsList(uniqueRestoredModels, {});
      return;
    }

    const currentModels = getCurrentModels();

    // 在应用新映射之前，先恢复当前模型到原始名称
    const restoredModels = currentModels.map(model => {
      // 如果模型在原始映射中，恢复为原始名称
      if (modelOriginalMapping[model]) {
        return modelOriginalMapping[model];
      }
      return model;
    });
    const { updatedModels, newMapping, hasChanges } = applyModelMapping(
      mapping,
      restoredModels,
      {}
    );

      if (hasChanges || JSON.stringify([...currentModels].sort()) !== JSON.stringify([...updatedModels].sort())) {
        updateModelsList(updatedModels, newMapping);
    }
  }, [modelOriginalMapping, inputs.models]);


  // Handle changes to the key list
  const updateKeyListToInput = (newKeyList) => {
    // Filter out empty strings before joining
    const filteredKeyList = newKeyList.filter(key => key.trim().length > 0);
    setKeyList(filteredKeyList);
    const combinedKey = filteredKeyList.join(',');

    // If only one valid key remains, switch back to single input mode
    if (filteredKeyList.length <= 1 && supportsMultiKeyView(inputs.type)) {
      setUseKeyListMode(false);
      // When switching back, ensure the single input shows the remaining key
      setInputs(inputs => ({ ...inputs, key: combinedKey }));
    } else {
      // Otherwise, update the main inputs.key based on the list
      setInputs(inputs => ({ ...inputs, key: combinedKey }));
    }
  };

  // Add a new key input box
  const addKeyInput = (initialValue = '') => {
    const newKeyList = [...keyList, initialValue];
    setKeyList(newKeyList);
    // Focus on the new input after it's rendered
    setTimeout(() => {
      const inputs = document.querySelectorAll('.key-input-item input');
      if (inputs.length > 0) {
        inputs[inputs.length - 1].focus();
      }
    }, 0);
  };

  // Remove a key input box
  const removeKeyInput = (index) => {
    const newKeyList = [...keyList];
    newKeyList.splice(index, 1);
    updateKeyListToInput(newKeyList);
  };

  // Update specific index key value
  const updateKeyAtIndex = (index, value) => {
    const newKeyList = [...keyList];
    newKeyList[index] = value;
    setKeyList(newKeyList); // Update local state immediately for smooth typing
    // Delay updating the main inputs.key state to avoid performance issues during typing
    // The main inputs.key will be updated by updateKeyListToInput when adding/removing or on submit
  };

  const handleInputChange = (name, value) => {
    if (name === 'base_url' && value.endsWith('/v1')) {
      Modal.confirm({
        title: '警告',
        content:
          '不需要在末尾加/v1，Veloera会自动处理，添加后可能导致请求失败，是否继续？',
        onOk: () => {
          setInputs((inputs) => ({ ...inputs, [name]: value }));
        },
      });
      return;
    }

    // 处理模型重定向变更时只更新状态，不触发实时同步
    if (name === 'model_mapping') {
      setInputs((inputs) => ({ ...inputs, [name]: value }));
      return;
    }

    // Special handling for key input when not in key list mode and not type 41
    if (name === 'key' && !useKeyListMode && inputs.type !== 41 && supportsMultiKeyView(inputs.type)) {
      // Check if the new value contains comma or newline
      if (value.includes(',') || value.includes('\n')) {
        // Switch to list mode
        setUseKeyListMode(true);
        setShowKey(true); // Switch to plain text display

        // Process the input value to create the initial key list
        const keys = value
          .split(/[,\n]/)
          .map(k => k.trim())
          .filter(k => k.length > 0);

        // If there are keys, set the list and focus the first input
        if (keys.length > 0) {
          setKeyList(keys);
          // Focus the first input after switching to list mode
          setTimeout(() => {
            const inputs = document.querySelectorAll('.key-input-item input');
            if (inputs.length > 0) {
              inputs[0].focus();
            }
          }, 0);
        } else {
          // If splitting resulted in no keys, stay in single mode but update input value
          setInputs((inputs) => ({ ...inputs, [name]: value }));
          setUseKeyListMode(false); // Ensure we don't switch to list mode with empty list
        }

        // The main inputs.key will be updated by updateKeyListToInput based on the list state
        return; // Prevent updating inputs.key directly here
      }
    }

    setInputs((inputs) => ({ ...inputs, [name]: value }));

    if (name === 'type') {
      // Reset key list mode when type changes, unless it's type 41
      if (value === 41 || !supportsMultiKeyView(value)) {
        setUseKeyListMode(false); // Type 41 uses a single textarea or doesn't support multi-key
        setKeyList([]); // Clear keyList if switching to single input mode
      } else if (inputs.type === 41 && value !== 41 && supportsMultiKeyView(value)) {
        // If switching from type 41 to another type that supports multi-key, check if the key contains commas/newlines
        if (inputs.key && (inputs.key.includes(',') || inputs.key.includes('\n'))) {
          setUseKeyListMode(true);
          setShowKey(true);
          const keys = inputs.key
            .split(/[,\n]/)
            .map(k => k.trim())
            .filter(k => k.length > 0);
          setKeyList(keys);
        } else {
          setUseKeyListMode(false);
          setKeyList([]); // Clear keyList if switching from type 41 to single mode
        }
      } else if (value !== 41 && inputs.key && (inputs.key.includes(',') || inputs.key.includes('\n')) && supportsMultiKeyView(value)) {
        // If changing type between non-41 types that support multi-key, and key already contains multi-keys
        setUseKeyListMode(true);
        setShowKey(true);
        const keys = inputs.key
          .split(/[,\n]/)
          .map(k => k.trim())
          .filter(k => k.length > 0);
        setKeyList(keys);
      } else {
        setUseKeyListMode(false);
        setKeyList([]); // Clear keyList if switching to single mode
      }


      // 更新基础模型选项供参考，但不自动填充到 Select 组件
      let localModels = [];
      switch (value) {
        case 2:
          localModels = [
            'mj_imagine',
            'mj_variation',
            'mj_reroll',
            'mj_blend',
            'mj_upscale',
            'mj_describe',
            'mj_uploads',
          ];
          break;
        case 5:
          localModels = [
            'swap_face',
            'mj_imagine',
            'mj_variation',
            'mj_reroll',
            'mj_blend',
            'mj_upscale',
            'mj_describe',
            'mj_zoom',
            'mj_shorten',
            'mj_modal',
            'mj_inpaint',
            'mj_custom_zoom',
            'mj_high_variation',
            'mj_low_variation',
            'mj_pan',
            'mj_uploads',
          ];
          break;
        case 36:
          localModels = ['suno_music', 'suno_lyrics'];
          break;
        default:
          localModels = getChannelModels(value);
          break;
      }
      setBasicModels(localModels);
    }
  };

  const loadChannel = async () => {
    setLoading(true);
    let res = await API.get(`/api/channel/${channelId}`);
    if (res === undefined) {
      setLoading(false);
      return;
    }
    const { success, message, data } = res.data;
    if (success) {
      if (data.models === '') {
        data.models = [];
      } else {
        data.models = data.models.split(',');
      }
      if (data.group === '') {
        data.groups = [];
      } else {
        data.groups = data.group.split(',');
      }
      if (data.model_mapping !== '') {
        data.model_mapping = JSON.stringify(
          JSON.parse(data.model_mapping),
          null,
          2,
        );
      }
      if (data.setting !== '' && data.setting !== null) { // Handle null setting
        try { // Add try-catch in case it's not valid JSON
          data.setting = JSON.stringify(
            JSON.parse(data.setting),
            null,
            2,
          );
        } catch (e) {
          console.error("Failed to parse channel setting:", data.setting, e);
          data.setting = data.setting; // Keep as is if invalid JSON
        }
      } else {
        data.setting = ''; // Ensure it's an empty string if null
      }
      if (data.param_override !== '' && data.param_override !== null) { // Handle null param_override
        try { // Add try-catch in case it's not valid JSON
          data.param_override = JSON.stringify(
            JSON.parse(data.param_override),
            null,
            2,
          );
        } catch (e) {
          console.error("Failed to parse channel param_override:", data.param_override, e);
          data.param_override = data.param_override; // Keep as is if invalid JSON
        }
      } else {
        data.param_override = ''; // Ensure it's an empty string if null
      }

      if (data.system_prompt !== '' && data.system_prompt !== null) { // Handle null system_prompt
        // Keep as is since it's already a string, no need to parse JSON
      } else {
        data.system_prompt = ''; // Ensure it's an empty string if null
      }


      // Handle the key
      if (data.key && supportsMultiKeyView(data.type)) {
        const keys = data.key.split(',').map(k => k.trim()).filter(k => k.length > 0);
        if (keys.length > 1) {
          setUseKeyListMode(true);
          setShowKey(true); // Ensure showKey is true for list mode
          setKeyList(keys);
        } else {
          setUseKeyListMode(false);
          setKeyList([]); // Clear keyList if not in list mode
        }
      } else {
        setUseKeyListMode(false);
        setKeyList([]);
      }
      setInitialKey(data.key); // Store initial key for single input mode placeholder

      // Save the original model_mapping data
      setOriginalModelMapping(data.model_mapping);

      // 初始化模型原始映射关系
      const mapping = parseModelMapping(data.model_mapping);
      if (mapping && Object.keys(mapping).length > 0) {
        const initialMapping = {};
        // 根据当前的模型映射和模型列表，建立原始映射关系
        Object.entries(mapping).forEach(([displayName, originalName]) => {
          const displayNameTrimmed = displayName.trim();
          const originalNameTrimmed = originalName?.trim();
          if (displayNameTrimmed && originalNameTrimmed && data.models.includes(displayNameTrimmed)) {
            initialMapping[displayNameTrimmed] = originalNameTrimmed;
          }
        });
        setModelOriginalMapping(initialMapping);
      } else {
        setModelOriginalMapping({});
      }

      setInputs(data);
      if (data.auto_ban === 0) {
        setAutoBan(false);
      } else {
        setAutoBan(true);
      }
      setBasicModels(getChannelModels(data.type));
    } else {
      showError(message);
    }
    setLoading(false);
  };


  const fetchModels = async () => {
    try {
      let res = await API.get(`/api/channel/models`);
      let localModelOptions = res.data.data.map((model) => ({
        label: model.id,
        value: model.id,
      }));
      setOriginModelOptions(localModelOptions);
      setFullModels(res.data.data.map((model) => model.id));
      setBasicModels(
        res.data.data
          .filter((model) => {
            return model.id.startsWith('gpt-') || model.id.startsWith('text-') || model.id.startsWith('claude-'); // Added claude for basic
          })
          .map((model) => model.id),
      );
    } catch (error) {
      console.error('Failed to parse model mapping JSON:', jsonStr, error);
      return [];
    }
  };

  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      if (res === undefined) {
        return;
      }
      setGroupOptions(
        res.data.data.map((group) => ({
          label: group,
          value: group,
        })),
      );
    } catch (error) {
      showError(error.message);
    }
  };

  useEffect(() => {
    let localModelOptions = [...originModelOptions];
    inputs.models.forEach((model) => {
      if (!localModelOptions.find((option) => option.label === model)) {
        localModelOptions.push({
          label: model,
          value: model,
        });
      }
    });
    setModelOptions(localModelOptions);
  }, [originModelOptions, inputs.models]);

  useEffect(() => {
    fetchModels().then();
    fetchGroups().then();
    if (isEdit) {
      loadChannel().then(() => {
        // Update the reset key after data loading is complete to force component reset
        setComponentResetKey(prev => prev + 1);
      });
    } else {
      setInputs(originInputs);
      setOriginalModelMapping(''); // Initialize as an empty string
      // 重置模型原始映射关系
      setModelOriginalMapping({});
      let localModels = getChannelModels(originInputs.type); // Use originInputs.type for initial state
      setBasicModels(localModels);
      setComponentResetKey(prev => prev + 1);
    }
  }, [props.editingChannel.id]);

  useEffect(() => {
    // When switching back from list mode to single mode, ensure the single input is focused
    if (!useKeyListMode && singleKeyInputRef.current) {
      singleKeyInputRef.current.focus();
    }
  }, [useKeyListMode]);

  // 在组件卸载时清理资源
  useEffect(() => {
    return () => {
      setModelOriginalMapping({});
    };
  }, []);


  const submit = async () => {
    // Update inputs.key from keyList before submitting if in list mode
    let finalKey = inputs.key;
    if (useKeyListMode) {
      // Filter out empty strings before joining
      const filteredKeyList = keyList.filter(key => key.trim().length > 0);
      finalKey = filteredKeyList.join(',');
    }

    if (!isEdit && (inputs.name === '' || finalKey === '')) {
      showInfo(t('请填写渠道名称和渠道密钥！'));
      return;
    }
    if (inputs.models.length === 0) {
      showInfo(t('请至少选择一个模型！'));
      return;
    }
    if (inputs.model_mapping !== '' && !verifyJSON(inputs.model_mapping)) {
      showInfo(t('模型映射必须是合法的 JSON格式！'));
      return;
    }
    if (inputs.setting !== '' && !verifyJSON(inputs.setting)) {
      showInfo(t('渠道额外设置必须是合法的 JSON 格式！'));
      return;
    }
    if (inputs.param_override !== '' && !verifyJSON(inputs.param_override)) {
      showInfo(t('参数覆盖必须是合法的 JSON 格式！'));
      return;
    }
    if (inputs.other !== '' && inputs.type === 41) {
      // For type 41, check if it's JSON only if it starts with {
      if (inputs.other.trim().startsWith('{') && !verifyJSON(inputs.other)) {
        showInfo(t('部署地区必须是合法的 JSON 格式或纯文本！'));
        return;
      }
    }


    let localInputs = { ...inputs };
    // Use the finalKey from multi-key mode if applicable
    if (useKeyListMode) {
      localInputs.key = finalKey;
    }
    if (localInputs.base_url && localInputs.base_url.endsWith('/')) {
      localInputs.base_url = localInputs.base_url.slice(
        0,
        localInputs.base_url.length - 1,
      );
    }
    if (localInputs.type === 18 && localInputs.other === '') {
      localInputs.other = 'v2.1';
    }
    let res;
    if (!Array.isArray(localInputs.models)) {
      showError(t('提交失败，请勿重复提交！'));
      handleCancel();
      return;
    }
    localInputs.auto_ban = autoBan ? 1 : 0;
    localInputs.models = localInputs.models.join(',');
    localInputs.group = localInputs.groups.join(',');

    // Ensure other is string for type 41 if it was JSON
    if (localInputs.type === 41 && typeof localInputs.other !== 'string') {
      localInputs.other = JSON.stringify(localInputs.other);
    }

    // Ensure setting and param_override are strings if they are objects (parsed from JSON)
    if (typeof localInputs.setting !== 'string') {
      localInputs.setting = JSON.stringify(localInputs.setting);
    }
    if (typeof localInputs.param_override !== 'string') {
      localInputs.param_override = JSON.stringify(localInputs.param_override);
    }


    if (isEdit) {
      res = await API.put(`/api/channel/`, {
        ...localInputs,
        id: parseInt(channelId),
      });
    } else {
      res = await API.post(`/api/channel/`, localInputs);
    }
    const { success, message } = res.data;
    if (success) {
      if (isEdit) {
        showSuccess(t('渠道更新成功！'));
      } else {
        showSuccess(t('渠道创建成功！'));
        setInputs(originInputs); // Reset form for new creation
        setKeyList([]); // Clear keyList state
        setUseKeyListMode(false); // Reset list mode
        setShowKey(false); // Reset showKey
        setDisableMultiKeyView(false); // Reset disable multi-key view
      }
      props.refresh();
      props.handleClose();
    } else {
      showError(message);
    }
  };

  const addCustomModels = () => {
    if (customModel.trim() === '') return;
    // Split by comma, newline, or space
    const modelArray = customModel.split(/[\s,]+/).map((model) => model.trim()).filter(model => model.length > 0);


    let localModels = [...inputs.models];
    let localModelOptions = [...modelOptions];
    let hasError = false;
    let addedCount = 0;

    modelArray.forEach((model) => {
      if (model && !localModels.includes(model)) {
        localModels.push(model);
        localModelOptions.push({
          label: model, // Use label/value for Semi Select
          value: model,
        });
        addedCount++;
      } else if (model) {
        hasError = true;
      }
    });

    if (hasError) {
      showWarning(t('某些模型已存在，已忽略！'));
    }

    if (addedCount > 0) {
      setModelOptions(localModelOptions);
      handleInputChange('models', localModels);
      setCustomModel(''); // Clear input only if something was added
    }
  };

  // Handle key down event for key list input
  const handleKeyInputKeyDown = (e, index) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault(); // Prevent default newline or comma
      const currentValue = keyList[index].trim();
      if (currentValue.length > 0) {
        // If the current input has content, ensure it's in the list (handled by updateKeyAtIndex)
        // Then add a new empty input below
        addKeyInput();
      } else if (e.key === 'Enter') {
        // If Enter is pressed on an empty input, just add a new empty one
        addKeyInput();
      }
      // If it's a comma on an empty input, do nothing (just prevent default)
    } else if (e.key === 'Backspace' && keyList[index] === '' && keyList.length > 1 && index > 0) {
      // If backspace is pressed on an empty input and there are other inputs before it
      e.preventDefault(); // Prevent default backspace
      const prevInput = document.querySelectorAll('.key-input-item input')[index - 1];
      removeKeyInput(index);
      // Focus on the previous input
      if (prevInput) {
        prevInput.focus();
      }
    }
  };

  // Handle paste event for key list input
  const handleKeyInputPaste = (e, index) => {
    const clipboardData = e.clipboardData || window.clipboardData;
    const pastedData = clipboardData.getData('Text');

    // Check if pasted data contains newline or comma
    if (pastedData.includes('\n') || pastedData.includes(',')) {
      e.preventDefault(); // Prevent default paste behavior

      // Get current value in the input where pasting
      const currentValue = keyList[index];

      // Process the pasted data and the current value
      const combinedValue = currentValue + pastedData;
      const newKeys = combinedValue
        .split(/[,\n]/) // Split by comma or newline
        .map(k => k.trim())
        .filter(k => k.length > 0); // Filter out empty strings

      // Update the key list state
      const newKeyList = [...keyList];
      // Remove the original key at the current index
      newKeyList.splice(index, 1);
      // Insert the new keys at the current index
      newKeyList.splice(index, 0, ...newKeys);

      updateKeyListToInput(newKeyList); // Update state and sync with inputs.key

      // Focus on the last inserted input if new ones were added
      if (newKeys.length > 0) {
        setTimeout(() => {
          const inputs = document.querySelectorAll('.key-input-item input');
          if (inputs.length >= index + newKeys.length) {
            inputs[index + newKeys.length - 1].focus();
          }
        }, 0);
      } else {
        // If pasted content resulted in no valid keys, focus on the previous input or the first if at index 0
        setTimeout(() => {
          const inputs = document.querySelectorAll('.key-input-item input');
          if (inputs.length > 0) {
            const targetIndex = index > 0 ? index - 1 : 0;
            if (inputs[targetIndex]) {
              inputs[targetIndex].focus();
            }
          }
        }, 0);
      }
    }
    // If no newline or comma, allow default paste
  };

  // Toggle multi-key view disable state
  const switchToSingleKeyMode = () => {
    setUseKeyListMode(false);
    // When switching back to single mode, combine existing keys back into one string
    const combinedKey = keyList.join(',');
    setInputs(inputs => ({ ...inputs, key: combinedKey }));
    setKeyList([]); // Clear key list state
  };

  // 复制功能
  const copyToClipboard = async (text, successMessage) => {
    try {
      await navigator.clipboard.writeText(text);
      showSuccess(successMessage || t('已复制到剪贴板'));
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
      showError(t('复制失败'));
    }
  };

  // 复制单个密钥
  const copyKey = async (key) => {
    if (!key || !key.trim()) {
      showWarning(t('密钥为空，无法复制'));
      return;
    }
    await copyToClipboard(key.trim(), t('密钥已复制'));
  };

  // 复制所有密钥（一行一个）
  const copyAllKeys = async () => {
    if (validKeys.length === 0) {
      showWarning(t('没有有效的密钥可复制'));
      return;
    }
    const allKeysText = validKeys.join('\n');
    await copyToClipboard(allKeysText, t('已复制全部密钥（{{count}}个）', { count: validKeys.length }));
  };


  // 渲染密钥输入组件
  const renderKeyInput = () => {
    // 多行文本框类型的渠道 (type 41)
    if (inputs.type === 41) {
      return (
        <div>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginTop: 10, marginBottom: 8 }}>
            <Typography.Text style={{ fontSize: 14, fontWeight: 600 }}>{t('密钥')}：</Typography.Text>
            <Button
              icon={<IconCopy />}
              onClick={() => copyKey(inputs.key)}
              size="small"
              theme="borderless"
              disabled={!inputs.key || !inputs.key.trim()}
            >
              {t('复制')}
            </Button>
          </div>
          <TextArea
            name='key'
            required
            placeholder={t(type2secretPrompt(inputs.type))}
            onChange={(value) => {
              handleInputChange('key', value);
            }}
            value={inputs.key}
            autoComplete='new-password'
            autosize={{ minRows: 2 }}
          />
        </div>
      );
    }

    // 使用列表模式显示多个密钥
    if (useKeyListMode && supportsMultiKeyView(inputs.type)) {
      return (
        <div>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginTop: 10, marginBottom: 8 }}>
            <Typography.Text style={{ fontSize: 14, fontWeight: 600 }}>{t('密钥')}：</Typography.Text>
            <Button
              size="small"
              theme="solid"
              onClick={switchToSingleKeyMode}
            >
              {t('切换为单密钥模式')}
            </Button>
          </div>

          <div style={{
            maxHeight: '50vh',
            overflowY: 'auto',
            border: '1px solid var(--semi-color-border)',
            borderRadius: '6px',
            padding: '12px',
            backgroundColor: 'var(--semi-color-fill-0)'
          }}>
            {keyList.map((key, index) => (
              <div key={index} style={{ display: 'flex', marginBottom: '5px', alignItems: 'center' }} className="key-input-item">
                <Typography.Text
                  style={{
                    minWidth: '30px',
                    textAlign: 'center',
                    color: 'var(--semi-color-text-2)',
                    fontSize: 12,
                    marginRight: 8
                  }}
                >
                  {index + 1}
                </Typography.Text>
                <Input
                  style={{ flex: 1 }}
                  value={key}
                  onChange={(value) => updateKeyAtIndex(index, value)}
                  onKeyDown={(e) => handleKeyInputKeyDown(e, index)}
                  onPaste={(e) => handleKeyInputPaste(e, index)}
                  placeholder={t('请输入第 {{index}} 个密钥', { index: index + 1 })}
                  size="small"
                />
                <Button
                  icon={<IconCopy />}
                  theme="borderless"
                  size="small"
                  onClick={() => copyKey(key)}
                  style={{
                    marginLeft: '4px',
                    minWidth: '28px',
                    opacity: key && key.trim() ? 1 : 0.3
                  }}
                  disabled={!key || !key.trim()}
                />
                <Button
                  icon={<IconMinusCircle />}
                  type="danger"
                  theme="borderless"
                  size="small"
                  onClick={() => removeKeyInput(index)}
                  style={{
                    marginLeft: '8px',
                    minWidth: '28px',
                    opacity: keyList.length <= 1 ? 0.3 : 1
                  }}
                  disabled={keyList.length <= 1}
                />
              </div>
            ))}
          </div>

          <div style={{ marginTop: 12, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <Button
                icon={<IconPlusCircle />}
                onClick={() => addKeyInput()}
                size="small"
                theme="solid"
              >
                {t('添加密钥')}
              </Button>
              <Button
                icon={<IconCopy />}
                onClick={() => copyAllKeys()}
                size="small"
                theme="solid"
                disabled={validKeys.length === 0}
              >
                {t('复制全部')}
              </Button>
            </div>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              {t('总计: {{count}} 个密钥', { count: keyList.length })}
            </Typography.Text>
          </div>

          <Typography.Text type="tertiary" style={{ fontSize: 12, marginTop: 8, display: 'block' }}>
            {t('💡 提示: 输入逗号或回车可快速添加新密钥')}
          </Typography.Text>
        </div>
      );
    }

    // 默认单行密钥输入
    return (
      <div>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginTop: 10, marginBottom: 8 }}>
          <Typography.Text style={{ fontSize: 14, fontWeight: 600 }}>{t('密钥')}：</Typography.Text>
          {supportsMultiKeyView(inputs.type) && (
            <Button
              size="small"
              theme="solid"
              onClick={() => {
                // 切换到多密钥模式，保留当前密钥作为第一个
                const currentKey = inputs.key || '';
                const keys = currentKey.split(/[\n,]/).map(k => k.trim()).filter(Boolean);
                const keysToSet = keys.length === 0 ? ['', ''] : [...keys, ''];
                setKeyList(keysToSet);
                setUseKeyListMode(true);
                setShowKey(true);
                setTimeout(() => {
                  const keyInputs = document.querySelectorAll('.key-input-item input');
                  if (keyInputs.length > 0) {
                    // 聚焦到最后一个（新添加的空）输入框
                    keyInputs[keyInputs.length - 1].focus();
                  }
                }, 0);
              }}
            >
              {t('切换为多密钥模式')}
            </Button>
          )}
        </div>

        <Input
          ref={singleKeyInputRef}
          name='key'
          required
          type={showKey ? 'text' : 'password'}
          placeholder={t(type2secretPrompt(inputs.type))}
          onChange={(value) => {
            handleInputChange('key', value);
          }}
          onPaste={(e) => {
            // Handle paste for single input mode to switch to list mode, if supported
            if (supportsMultiKeyView(inputs.type)) {
              const clipboardData = e.clipboardData || window.clipboardData;
              const pastedData = clipboardData.getData('Text');

              // Check if pasted data contains newline or comma
              if (pastedData.includes('\n') || pastedData.includes(',')) {
                e.preventDefault(); // Prevent default paste

                // Prepend existing key if any
                const combinedData = (inputs.key || '') + pastedData;

                // Process the pasted data to switch to list mode
                const keys = combinedData
                  .split(/[,\n]/)
                  .map(k => k.trim())
                  .filter(k => k.length > 0);

                if (keys.length > 0) {
                  setUseKeyListMode(true);
                  setShowKey(true);
                  setKeyList(keys);
                  // Update the main inputs.key state based on the new list
                  handleInputChange('key', keys.join(','));

                  // Focus the first input after switching to list mode
                  setTimeout(() => {
                    const inputs = document.querySelectorAll('.key-input-item input');
                    if (inputs.length > 0) {
                      inputs[0].focus();
                    }
                  }, 0);

                } else {
                  // If splitting resulted in no valid keys, just update the input value (which is empty after split)
                  handleInputChange('key', '');
                }
              }
              // If no newline or comma, allow default paste (handled by onChange)
            }
            // If multi-key view not supported, allow default paste (handled by onChange)
          }}
          value={inputs.key}
          autoComplete='new-password'
          addonAfter={
            <Space>
              <Button
                theme="borderless"
                icon={<IconCopy />}
                onClick={() => copyKey(inputs.key)}
                style={{ padding: '0 4px' }}
                disabled={!inputs.key || !inputs.key.trim()}
              />
              <Button
                theme="borderless"
                icon={showKey ? <IconEyeClosedSolid /> : <IconEyeOpened />}
                onClick={() => setShowKey(!showKey)}
                style={{ padding: '0 4px' }}
              />
            </Space>
          }
        />

        {/* 清空按钮 */}
        {inputs.key && (
          <Button
            type='danger'
            theme='borderless'
            onClick={() => {
              Modal.confirm({
                title: t('确认清空密钥'),
                content: t('您确定要清空密钥输入框的内容吗？'),
                onOk: () => {
                  handleInputChange('key', '');
                  showSuccess(t('密钥已清空'));
                },
              });
            }}
            style={{ marginTop: 8 }}
          >
            {t('清空')}
          </Button>
        )}
      </div>
    );
  };

  return (
    <>
      <SideSheet
        maskClosable={false}
        placement={isEdit ? 'right' : 'left'}
        title={
          <Title level={3}>
            {isEdit ? t('更新渠道信息') : t('创建新的渠道')}
          </Title>
        }
        headerStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        bodyStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        visible={props.visible}
        footer={
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Space>
              <Button theme='solid' size={'large'} onClick={submit} loading={loading}>
                {t('提交')}
              </Button>
              <Button
                theme='solid'
                size={'large'}
                type={'tertiary'}
                onClick={handleCancel}
                disabled={loading}
              >
                {t('取消')}
              </Button>
            </Space>
          </div>
        }
        closeIcon={null}
        onCancel={() => handleCancel()}
        width={isMobile() ? '100%' : 600}
      >
        <Spin spinning={loading}>
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>{t('类型')}：</Typography.Text>
          </div>
          <Select
            name='type'
            required
            optionList={CHANNEL_OPTIONS}
            value={inputs.type}
            onChange={(value) => handleInputChange('type', value)}
            style={{ width: '50%' }}
            filter
            searchPosition='dropdown'
            placeholder={t('请选择渠道类型')}
          />
          {inputs.type === 40 && (
            <div style={{ marginTop: 10 }}>
              <Banner
                type='info'
                description={
                  <div>
                    <Typography.Text strong>{t('邀请链接')}:</Typography.Text>
                    <Typography.Text
                      link
                      underline
                      style={{ marginLeft: 8 }}
                      onClick={() =>
                        window.open('https://cloud.siliconflow.cn/i/hij0YNTZ')
                      }
                    >
                      https://cloud.siliconflow.cn/i/hij0YNTZ
                    </Typography.Text>
                  </div>
                }
              />
            </div>
          )}
          {inputs.type === 3 && (
            <>
              <div style={{ marginTop: 10 }}>
                <Banner
                  type={'warning'}
                  description={
                    <>
                      {t('2025年5月10日后添加的渠道，不需要再在部署的时候移除模型名称中的"."')}
                      {/*<br />*/}
                      {/*<Typography.Text*/}
                      {/*  style={{*/}
                      {/*    color: 'rgba(var(--semi-blue-5), 1)',*/}
                      {/*    userSelect: 'none',*/}
                      {/*    cursor: 'pointer',*/}
                      {/*  }}*/}
                      {/*  onClick={() => {*/}
                      {/*    setModalImageUrl(*/}
                      {/*      '/azure_model_name.png',*/}
                      {/*    );*/}
                      {/*    setIsModalOpenurl(true)*/}

                      {/*  }}*/}
                      {/*>*/}
                      {/*  {t('查看示例')}*/}
                      {/*</Typography.Text>*/}
                    </>
                  }
                ></Banner>
              </div>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>
                  AZURE_OPENAI_ENDPOINT：
                </Typography.Text>
              </div>
              <Input
                label='AZURE_OPENAI_ENDPOINT'
                name='azure_base_url'
                placeholder={t(
                  '请输入 AZURE_OPENAI_ENDPOINT，例如：https://docs-test-001.openai.azure.com',
                )}
                onChange={(value) => {
                  handleInputChange('base_url', value);
                }}
                value={inputs.base_url}
                autoComplete='new-password'
              />
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>{t('默认 API 版本')}：</Typography.Text>
              </div>
              <Input
                label={t('默认 API 版本')}
                name='azure_other'
                placeholder={t('请输入默认 API 版本，例如：2024-12-01-preview')}
                onChange={(value) => {
                  handleInputChange('other', value);
                }}
                value={inputs.other}
                autoComplete='new-password'
              />
            </>
          )}
          {inputs.type === 8 && (
            <>
              <div style={{ marginTop: 10 }}>
                <Banner
                  type={'warning'}
                  description={t(
                    '如果你对接的是上游One API或者New API等转发项目，请使用OpenAI类型，不要使用此类型，除非你知道你在做什么。',
                  )}
                ></Banner>
              </div>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>
                  {t('完整的 Base URL，支持变量{model}')}：
                </Typography.Text>
              </div>
              <Input
                name='base_url'
                placeholder={t(
                  '请输入完整的URL，例如：https://api.openai.com/v1/chat/completions',
                )}
                onChange={(value) => {
                  handleInputChange('base_url', value);
                }}
                value={inputs.base_url}
                autoComplete='new-password'
              />
            </>
          )}
          {inputs.type === 37 && (
            <>
              <div style={{ marginTop: 10 }}>
                <Banner
                  type={'warning'}
                  description={t(
                    'Dify渠道只适配chatflow和agent，并且agent不支持图片！',
                  )}
                ></Banner>
              </div>
            </>
          )}
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>{t('名称')}：</Typography.Text>
          </div>
          <Input
            required
            name='name'
            placeholder={t('请为渠道命名')}
            onChange={(value) => {
              handleInputChange('name', value);
            }}
            value={inputs.name}
            autoComplete='new-password'
          />
          {inputs.type !== 3 && inputs.type !== 8 && inputs.type !== 22 && inputs.type !== 36 && inputs.type !== 45 && (
            <>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>{t('API地址')}：</Typography.Text>
              </div>
              <Tooltip content={t('对于官方渠道，Veloera已经内置地址，除非是第三方代理站点或者Azure的特殊接入地址，否则不需要填写')}>
                <Input
                  label={t('API地址')}
                  name="base_url"
                  placeholder={t('此项可选，用于通过自定义API地址来进行 API 调用，末尾带 / 以不使用默认 /v1 前缀')}
                  onChange={(value) => {
                    handleInputChange('base_url', value);
                  }}
                  value={inputs.base_url}
                  autoComplete="new-password"
                />
              </Tooltip>
            </>
          )}
          {renderKeyInput()}
          {inputs.type === 22 && (
            <>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>{t('私有部署地址')}：</Typography.Text>
              </div>
              <Input
                name='base_url'
                placeholder={t(
                  '请输入私有部署地址，格式为：https://fastgpt.run/api/openapi',
                )}
                onChange={(value) => {
                  handleInputChange('base_url', value);
                }}
                value={inputs.base_url}
                autoComplete='new-password'
              />
            </>
          )}
          {inputs.type === 36 && (
            <>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>
                  {t(
                    '注意非Chat API，请务必填写正确的API地址，否则可能导致无法使用',
                  )}
                </Typography.Text>
              </div>
              <Input
                name='base_url'
                placeholder={t(
                  '请输入到 /suno 前的路径，通常就是域名，例如：https://api.example.com',
                )}
                onChange={(value) => {
                  handleInputChange('base_url', value);
                }}
                value={inputs.base_url}
                autoComplete='new-password'
              />
            </>
          )}
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>{t('分组')}：</Typography.Text>
          </div>
          <Select
            placeholder={t('请选择可以使用该渠道的分组')}
            name='groups'
            required
            multiple
            selection
            allowAdditions
            additionLabel={t('请在系统设置页面编辑分组倍率以添加新的分组：')}
            onChange={(value) => {
              handleInputChange('groups', value);
            }}
            value={inputs.groups}
            autoComplete='new-password'
            optionList={groupOptions}
          />
          {inputs.type === 18 && (
            <>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>模型版本：</Typography.Text>
              </div>
              <Input
                name='other'
                placeholder={
                  '请输入星火大模型版本，注意是接口地址中的版本号，例如：v2.1'
                }
                onChange={(value) => {
                  handleInputChange('other', value);
                }}
                value={inputs.other}
                autoComplete='new-password'
              />
            </>
          )}
          {inputs.type === 41 && (
            <>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>{t('部署地区')}：</Typography.Text>
              </div>
              <TextArea
                name='other'
                placeholder={t(
                  '请输入部署地区，例如：us-central1\n支持使用模型映射格式\n' +
                  '{\n' +
                  '    "default": "us-central1",\n' +
                  '    "claude-3-5-sonnet-20240620": "europe-west1"\n' +
                  '}',
                )}
                autosize={{ minRows: 2 }}
                onChange={(value) => {
                  handleInputChange('other', value);
                }}
                value={inputs.other}
                autoComplete='new-password'
              />
              <Typography.Text
                style={{
                  color: 'rgba(var(--semi-blue-5), 1)',
                  userSelect: 'none',
                  cursor: 'pointer',
                }}
                onClick={() => {
                  handleInputChange(
                    'other',
                    JSON.stringify(REGION_EXAMPLE, null, 2),
                  );
                }}
              >
                {t('填入模板')}
              </Typography.Text>
            </>
          )}
          {inputs.type === 21 && (
            <>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>知识库 ID：</Typography.Text>
              </div>
              <Input
                label='知识库 ID'
                name='other'
                placeholder={'请输入知识库 ID，例如：123456'}
                onChange={(value) => {
                  handleInputChange('other', value);
                }}
                value={inputs.other}
                autoComplete='new-password'
              />
            </>
          )}
          {inputs.type === 39 && (
            <>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>Account ID：</Typography.Text>
              </div>
              <Input
                name='other'
                placeholder={
                  '请输入Account ID，例如：d6b5da8hk1awo8nap34ube6gh'
                }
                onChange={(value) => {
                  handleInputChange('other', value);
                }}
                value={inputs.other}
                autoComplete='new-password'
              />
            </>
          )}
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>{t('模型')}：</Typography.Text>
          </div>
          <Select
            placeholder={'请选择该渠道所支持的模型'}
            name='models'
            required
            multiple
            selection
            filter
            searchPosition='dropdown'
            onChange={(value) => {
              // 对模型列表进行去重处理（区分大小写）
              const deduplicatedModels = [...new Set(value.filter(model => model && typeof model === 'string' && model.trim() !== ''))];
              handleInputChange('models', deduplicatedModels);
            }}
            value={inputs.models}
            autoComplete='new-password'
            optionList={modelOptions}
            renderSelectedItem={(optionNode) => {
              const modelName = String(optionNode?.value ?? '');

              const handleCopy = async (e) => {
                e.stopPropagation();
                try {
                  await navigator.clipboard.writeText(modelName);
                  showSuccess(t('已复制：{{name}}', { name: modelName }));
                } catch (error) {
                  console.error('Failed to copy to clipboard:', error);
                  showError(t('复制失败'));
                }
              };

              return {
                isRenderInTag: true,
                content: (
                  <span
                    className="cursor-pointer select-none"
                    role="button"
                    tabIndex={0}
                    title={t('点击复制模型名称')}
                    onClick={handleCopy}
                  >
                    {optionNode.label || modelName}
                  </span>
                ),
              };
            }}
          />
          <div style={{ lineHeight: '40px', marginBottom: '12px' }}>
            <Space>
              <Button
                type='primary'
                onClick={() => {
                  handleInputChange('models', basicModels);
                }}
              >
                {t('填入相关模型')}
              </Button>
              <Button
                type='secondary'
                onClick={() => {
                  handleInputChange('models', fullModels);
                }}
              >
                {t('填入所有模型')}
              </Button>
              <Button
                style={{ marginLeft: 8 }}
                type='tertiary'
                onClick={() => {
                  // Using Modal.info for better customization
                  Modal.info({
                    title: t('高级模型选择'),
                    content: (
                      <div>
                        <ModelSelector
                          channelId={isEdit ? channelId : null}
                          type={inputs.type}
                          apiKey={inputs.key}
                          baseUrl={inputs.base_url}
                          isEdit={isEdit}
                          selectedModels={inputs.models}
                          onSelect={(selectedModels) => {
                            handleInputChange('models', selectedModels);
                          }}
                        />
                      </div>
                    ),
                    footer: null,
                    width: 800,
                    mask: true,
                    maskClosable: false,
                    closable: true,
                  });
                }}
              >
                {t('模型选择')}
              </Button>
              <Button
                type='warning'
                onClick={() => {
                  handleInputChange('models', []);
                }}
              >
                {t('清除所有')}
              </Button>
              <Button
                type='tertiary'
                onClick={() => {
                  if (inputs.models && inputs.models.length > 0) {
                    const modelsText = inputs.models.join(',');
                    navigator.clipboard.writeText(modelsText).then(() => {
                      showSuccess(t('已复制到剪贴板'));
                    }).catch(() => {
                      showError(t('复制失败'));
                    });
                  } else {
                    showWarning(t('没有模型可复制'));
                  }
                }}
              >
                {t('复制所有')}
              </Button>
            </Space>
            <Input
              addonAfter={
                <Button type='primary' onClick={addCustomModels}>
                  {t('填入')}
                </Button>
              }
              placeholder={t('输入自定义模型名称，多个用逗号或空格分隔')}
              value={customModel}
              onChange={(value) => {
                setCustomModel(value); // Allow space and comma for input
              }}
              onPressEnter={addCustomModels} // Add on press enter
            />
          </div>
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>{t('模型重定向')}：</Typography.Text>
          </div>
          <ModelMappingEditor
            key={`model-mapping-${componentResetKey}`}
            value={originalModelMapping || inputs.model_mapping}
            onChange={(value) => handleInputChange('model_mapping', value)}
            onRealtimeChange={syncModelMappingToModels}
            placeholder={t('此项可选，用于修改请求体中的模型名称')}
          />
          <div style={{ marginTop: 8 }}>
            <Typography.Text type="tertiary" style={{ fontSize: 12 }}>
              {t('💡 提示：设置重定向后，系统自动将“模型配置”中对应的“值”替换为“键”')}
            </Typography.Text>
          </div>
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>{t('渠道标签')}</Typography.Text>
          </div>
          <Input
            label={t('渠道标签')}
            name='tag'
            placeholder={t('渠道标签')}
            onChange={(value) => {
              handleInputChange('tag', value);
            }}
            value={inputs.tag}
            autoComplete='new-password'
          />
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>{t('渠道优先级')}</Typography.Text>
          </div>
          <Input
            label={t('渠道优先级')}
            name='priority'
            placeholder={t('渠道优先级')}
            onChange={(value) => {
              const number = parseInt(value);
              if (isNaN(number)) {
                handleInputChange('priority', value);
              } else {
                handleInputChange('priority', number);
              }
            }}
            value={inputs.priority}
            autoComplete='new-password'
          />
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>{t('渠道权重')}</Typography.Text>
          </div>
          <Input
            label={t('渠道权重')}
            name='weight'
            placeholder={t('渠道权重')}
            onChange={(value) => {
              const number = parseInt(value);
              if (isNaN(number)) {
                handleInputChange('weight', value);
              } else {
                handleInputChange('weight', number);
              }
            }}
            value={inputs.weight}
            autoComplete='new-password'
          />
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>{t('渠道前缀')}</Typography.Text>
          </div>
          <Tooltip content={t('可选项，设置后此渠道的所有模型向客户显示时都会自动添加此前缀，同时可以根据前缀将请求路由到该渠道')}>
            <Input
              label={t('渠道前缀')}
              name='model_prefix'
              placeholder={t('例如: cursor-')}
              onChange={(value) => {
                handleInputChange('model_prefix', value);
              }}
              value={inputs.model_prefix}
              autoComplete='new-password'
            />
          </Tooltip>
          <>
            <div style={{ marginTop: 10 }}>
              <Typography.Text strong>{t('渠道额外设置')}：</Typography.Text>
            </div>
            <TextArea
              placeholder={
                t(
                  '此项可选，用于配置渠道特定设置，为一个 JSON 字符串，例如：',
                ) + '\n{\n  "force_format": true\n}'
              }
              name='setting'
              onChange={(value) => {
                handleInputChange('setting', value);
              }}
              autosize
              value={inputs.setting}
              autoComplete='new-password'
            />
            <Space>
              <Typography.Text
                style={{
                  color: 'rgba(var(--semi-blue-5), 1)',
                  userSelect: 'none',
                  cursor: 'pointer',
                }}
                onClick={() => {
                  handleInputChange(
                    'setting',
                    JSON.stringify(
                      {
                        force_format: true,
                      },
                      null,
                      2,
                    ),
                  );
                }}
              >
                {t('填入模板')}
              </Typography.Text>
              <Typography.Text
                style={{
                  color: 'rgba(var(--semi-blue-5), 1)',
                  userSelect: 'none',
                  cursor: 'pointer',
                }}
                onClick={() => {
                  window.open(
                    'https://github.com/Veloera/Veloera/blob/main/docs/channel/other_setting.md',
                  );
                }}
              >
                {t('设置说明')}
              </Typography.Text>
            </Space>
          </>
          <>
            <div style={{ marginTop: 10 }}>
              <Typography.Text strong>{t('参数覆盖')}：</Typography.Text>
            </div>
            <TextArea
              placeholder={
                t(
                  '此项可选，用于覆盖请求参数。不支持覆盖 stream 参数。为一个 JSON 字符串，例如：',
                ) + '\n{\n  "temperature": 0\n}'
              }
              name='param_override'
              onChange={(value) => {
                handleInputChange('param_override', value);
              }}
              autosize
              value={inputs.param_override}
              autoComplete='new-password'
            />
          </>
          {inputs.type === 1 && (
            <>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>{t('组织')}：</Typography.Text>
              </div>
              <Input
                label={t('组织，可选，不填则为默认组织')}
                name='openai_organization'
                placeholder={t('请输入组织org-xxx')}
                onChange={(value) => {
                  handleInputChange('openai_organization', value);
                }}
                value={inputs.openai_organization}
              />
            </>
          )}
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>{t('默认测试模型')}：</Typography.Text>
          </div>
          <Input
            name='test_model'
            placeholder={t('不填则为模型列表第一个')}
            onChange={(value) => {
              handleInputChange('test_model', value);
            }}
            value={inputs.test_model}
          />
          <div style={{ marginTop: 10, display: 'flex' }}>
            <Space>
              <Checkbox
                name='auto_ban'
                checked={autoBan}
                onChange={() => {
                  setAutoBan(!autoBan);
                }}
              />
              <Typography.Text strong>
                {t(
                  '是否自动禁用（仅当自动禁用开启时有效），关闭后不会自动禁用该渠道：',
                )}
              </Typography.Text>
            </Space>
          </div>
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>
              {t('状态码复写（仅影响本地判断，不修改返回到上游的状态码）')}：
            </Typography.Text>
          </div>
          <TextArea
            placeholder={
              t(
                '此项可选，用于复写返回的状态码，比如将claude渠道的400错误复写为500（用于重试），请勿滥用该功能，例如：',
              ) +
              '\n' +
              JSON.stringify(STATUS_CODE_MAPPING_EXAMPLE, null, 2)
            }
            name='status_code_mapping'
            onChange={(value) => {
              handleInputChange('status_code_mapping', value);
            }}
            autosize
            value={inputs.status_code_mapping}
            autoComplete='new-password'
          />
          <Typography.Text
            style={{
              color: 'rgba(var(--semi-blue-5), 1)',
              userSelect: 'none',
              cursor: 'pointer',
            }}
            onClick={() => {
              handleInputChange(
                'status_code_mapping',
                JSON.stringify(STATUS_CODE_MAPPING_EXAMPLE, null, 2),
              );
            }}
          >
            {t('填入模板')}
          </Typography.Text>
          <div style={{ marginTop: 10 }}>
            <Typography.Text strong>{t('系统提示词设置')}：</Typography.Text>
          </div>
          <TextArea
            placeholder={t('此项可选，用于为该渠道设置系统提示词，将会在用户请求的系统提示词前追加')}
            name='system_prompt'
            onChange={(value) => {
              handleInputChange('system_prompt', value);
            }}
            autosize={{ minRows: 2 }}
            value={inputs.system_prompt}
            autoComplete='new-password'
          />
        </Spin>
      </SideSheet>
    </>
  );
};

export default EditChannel;
