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
import React, { useEffect, useState, useRef } from 'react';
import {
  Button,
  Col,
  Form,
  Row,
  Spin,
  Modal,
  Select,
  Space,
  Typography,
} from '@douyinfe/semi-ui';
import {
  API,
  showError,
  showSuccess,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

export default function ModelCommonRatioSettings(props) {
  const [loading, setLoading] = useState(false);
  const [presetModalVisible, setPresetModalVisible] = useState(false);
  const [resetModalVisible, setResetModalVisible] = useState(false);
  const [presetLoading, setPresetLoading] = useState(false);
  const [resetLoading, setResetLoading] = useState(false);
  const [selectedPresetSource, setSelectedPresetSource] = useState('flexible');
  const [selectedResetSource, setSelectedResetSource] = useState('flexible');
  const [inputs, setInputs] = useState({
    fallback_pricing_enabled: false,
    fallback_single_price: '',
    fallback_input_ratio: '',
    fallback_completion_ratio: '',
  });
  const { t } = useTranslation();
  const refForm = useRef();

  function handleFieldChange(fieldName) {
    return (value) => {
      if (fieldName === 'fallback_single_price') {
        setInputs((inputs) => ({
          ...inputs,
          fallback_single_price: typeof value === 'number' ? String(value) : value,
          fallback_input_ratio: value ? '' : inputs.fallback_input_ratio,
          fallback_completion_ratio: value ? '' : inputs.fallback_completion_ratio,
        }));
      } else if (fieldName === 'fallback_input_ratio' || fieldName === 'fallback_completion_ratio') {
        setInputs((inputs) => ({
          ...inputs,
          [fieldName]: typeof value === 'number' ? String(value) : value,
          fallback_single_price: value ? '' : inputs.fallback_single_price,
        }));
      } else {
        setInputs((inputs) => ({ 
          ...inputs, 
          [fieldName]: typeof value === 'number' ? String(value) : value 
        }));
      }
    };
  }

  const presetSources = [
    { value: 'flexible', label: '通用（默认，推荐）' },
    { value: 'openrouter', label: 'OpenRouter' },
    { value: 'legacy', label: '传统（不推荐）' },
  ];

  const resetPresetRatios = async () => {
    setResetLoading(true);
    try {
      let promptUrl, completionUrl;
      
      if (selectedResetSource === 'legacy') {
        completionUrl = 'https://public-assets.veloera.org/defaults/model-ratios/completion.json';
        promptUrl = null;
      } else {
        promptUrl = `https://public-assets.veloera.org/defaults/model-ratios/${selectedResetSource}/prompt.json`;
        completionUrl = `https://public-assets.veloera.org/defaults/model-ratios/${selectedResetSource}/completion.json`;
      }

      const requests = [];
      if (promptUrl) {
        requests.push(fetch(promptUrl).then(res => res.json()).catch(() => ({})));
      }
      requests.push(fetch(completionUrl).then(res => res.json()).catch(() => ({})));

      const [promptRatios, completionRatios] = promptUrl ? await Promise.all(requests) : [null, await requests[0]];

      const currentModelRatio = JSON.parse(props.options.ModelRatio || '{}');
      const currentCompletionRatio = JSON.parse(props.options.CompletionRatio || '{}');

      const filteredModelRatio = { ...currentModelRatio };
      const filteredCompletionRatio = { ...currentCompletionRatio };

      if (promptRatios) {
        Object.keys(promptRatios).forEach(model => {
          delete filteredModelRatio[model];
        });
      }

      if (completionRatios) {
        Object.keys(completionRatios).forEach(model => {
          delete filteredCompletionRatio[model];
        });
      }

      const requestQueue = [
        API.put('/api/option/', {
          key: 'ModelRatio',
          value: JSON.stringify(filteredModelRatio, null, 2)
        }),
        API.put('/api/option/', {
          key: 'CompletionRatio', 
          value: JSON.stringify(filteredCompletionRatio, null, 2)
        })
      ];

      const results = await Promise.all(requestQueue);
      
      if (results.includes(undefined)) {
        return showError(t('重置预设倍率失败，请重试'));
      }

      for (const res of results) {
        if (!res.data.success) {
          return showError(res.data.message);
        }
      }

      showSuccess(t('预设倍率重置成功'));
      setResetModalVisible(false);
      props.refresh();
    } catch (error) {
      console.error('重置预设倍率失败:', error);
      showError(t('重置预设倍率失败，请重试'));
    } finally {
      setResetLoading(false);
    }
  };

  const fetchPresetRatios = async () => {
    setPresetLoading(true);
    try {
      let promptUrl, completionUrl;
      
      if (selectedPresetSource === 'legacy') {
        completionUrl = 'https://public-assets.veloera.org/defaults/model-ratios/completion.json';
        promptUrl = null;
      } else {
        promptUrl = `https://public-assets.veloera.org/defaults/model-ratios/${selectedPresetSource}/prompt.json`;
        completionUrl = `https://public-assets.veloera.org/defaults/model-ratios/${selectedPresetSource}/completion.json`;
      }

      const requests = [];
      if (promptUrl) {
        requests.push(fetch(promptUrl).then(res => res.json()).catch(() => ({})));
      }
      requests.push(fetch(completionUrl).then(res => res.json()).catch(() => ({})));

      const [promptRatios, completionRatios] = promptUrl ? await Promise.all(requests) : [null, await requests[0]];

      const currentModelRatio = JSON.parse(props.options.ModelRatio || '{}');
      const currentCompletionRatio = JSON.parse(props.options.CompletionRatio || '{}');

      const mergedModelRatio = { ...currentModelRatio };
      const mergedCompletionRatio = { ...currentCompletionRatio };

      if (promptRatios) {
        Object.keys(promptRatios).forEach(model => {
          mergedModelRatio[model] = promptRatios[model];
        });
      }

      if (completionRatios) {
        Object.keys(completionRatios).forEach(model => {
          mergedCompletionRatio[model] = completionRatios[model];
        });
      }

      const requestQueue = [
        API.put('/api/option/', {
          key: 'ModelRatio',
          value: JSON.stringify(mergedModelRatio, null, 2)
        }),
        API.put('/api/option/', {
          key: 'CompletionRatio', 
          value: JSON.stringify(mergedCompletionRatio, null, 2)
        })
      ];

      const results = await Promise.all(requestQueue);
      
      if (results.includes(undefined)) {
        return showError(t('获取预设倍率失败，请重试'));
      }

      for (const res of results) {
        if (!res.data.success) {
          return showError(res.data.message);
        }
      }

      showSuccess(t('预设倍率获取成功'));
      setPresetModalVisible(false);
      props.refresh();
      
      await saveFallbackPricing();
    } catch (error) {
      console.error('获取预设倍率失败:', error);
      showError(t('获取预设倍率失败，请重试'));
    } finally {
      setPresetLoading(false);
    }
  };

  const handleFallbackPriceChange = (type, value) => {
    const newInputs = {...inputs};
    
    if (type === 'single') {
      newInputs.fallback_single_price = value;
      if (value) {
        newInputs.fallback_input_ratio = '';
        newInputs.fallback_completion_ratio = '';
      }
    } else {
      newInputs[`fallback_${type}_ratio`] = value;
      if (value) {
        newInputs.fallback_single_price = '';
      }
    }
    
    setInputs(newInputs);
  };

  const validateFallbackPricing = () => {
    const { fallback_single_price, fallback_input_ratio, fallback_completion_ratio } = inputs;
    
    if (fallback_single_price) {
      return !fallback_input_ratio && !fallback_completion_ratio;
    }
    
    if (fallback_input_ratio || fallback_completion_ratio) {
      return fallback_input_ratio && fallback_completion_ratio && !fallback_single_price;
    }
    
    return true;
  };

  const saveFallbackPricing = async () => {
    if (!validateFallbackPricing()) {
      showError(t('请检查兜底倍率配置：使用单次价格时不能设置倍率，使用倍率时需要同时设置输入和补全倍率'));
      return;
    }

    const fallbackOptions = [
      { key: 'fallback_pricing_enabled', value: String(inputs.fallback_pricing_enabled) },
      { key: 'fallback_single_price', value: String(inputs.fallback_single_price || '') },
      { key: 'fallback_input_ratio', value: String(inputs.fallback_input_ratio || '') },
      { key: 'fallback_completion_ratio', value: String(inputs.fallback_completion_ratio || '') }
    ];

    try {
      setLoading(true);
      const requestQueue = fallbackOptions.map((option) => 
        API.put('/api/option/', option)
      );

      const res = await Promise.all(requestQueue);
      
      if (res.includes(undefined)) {
        return showError(t('保存失败，请重试'));
      }

      for (let i = 0; i < res.length; i++) {
        if (!res[i].data.success) {
          return showError(res[i].data.message);
        }
      }

      showSuccess(t('兜底倍率保存成功'));
      props.refresh();
    } catch (error) {
      console.error('Unexpected error:', error);
      showError(t('保存失败，请重试'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const currentInputs = {
      fallback_pricing_enabled: false,
      fallback_single_price: '',
      fallback_input_ratio: '',
      fallback_completion_ratio: '',
    };
    
    for (let key in props.options) {
      if (Object.keys(currentInputs).includes(key)) {
        if (key === 'fallback_pricing_enabled') {
          currentInputs[key] = props.options[key] === 'true';
        } else {
          currentInputs[key] = props.options[key] || '';
        }
      }
    }
    setInputs(currentInputs);
    if (refForm.current) {
      refForm.current.setValues(currentInputs);
    }
  }, [props.options]);

  return (
    <Spin spinning={loading}>
      <Form
        values={inputs}
        getFormApi={(formAPI) => (refForm.current = formAPI)}
        style={{ marginBottom: 15 }}
      >
        <Form.Section text={t('其他倍率设置')}>
          <Row gutter={16} style={{ marginBottom: 16 }}>
            <Col span={12}>
              <Button 
                type='primary'
                onClick={() => setPresetModalVisible(true)}
              >
                {t('获取预设倍率')}
              </Button>
            </Col>
            <Col span={12}>
              <Button 
                type='tertiary'
                onClick={() => setResetModalVisible(true)}
              >
                {t('重置预设(移除预设中模型)')}
              </Button>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={24}>
              <Form.Switch
                label={t('启用兜底倍率')}
                field="fallback_pricing_enabled"
                onChange={handleFieldChange('fallback_pricing_enabled')}
              />
            </Col>
          </Row>
          {inputs.fallback_pricing_enabled && (
            <Row gutter={16} style={{ marginTop: 16 }}>
              <Col span={6}>
                <Form.InputNumber
                  label={t('单次价格')}
                  field="fallback_single_price"
                  placeholder="0.01"
                  min={0}
                  step={0.001}
                  disabled={inputs.fallback_input_ratio || inputs.fallback_completion_ratio}
                  onChange={handleFieldChange('fallback_single_price')}
                />
              </Col>
              <Col span={6}>
                <Form.InputNumber
                  label={t('模型输入倍率')}
                  field="fallback_input_ratio"
                  placeholder="1.0"
                  min={0}
                  step={0.1}
                  disabled={inputs.fallback_single_price}
                  onChange={handleFieldChange('fallback_input_ratio')}
                />
              </Col>
              <Col span={6}>
                <Form.InputNumber
                  label={t('模型补全倍率')}
                  field="fallback_completion_ratio"
                  placeholder="2.0"
                  min={0}
                  step={0.1}
                  disabled={inputs.fallback_single_price}
                  onChange={handleFieldChange('fallback_completion_ratio')}
                />
              </Col>
              <Col span={6}>
                <Button style={{ marginTop: 30 }} onClick={saveFallbackPricing} loading={loading}>
                  {t('保存兜底倍率')}
                </Button>
              </Col>
            </Row>
          )}
        </Form.Section>
      </Form>
      
      <Modal
        title={t('获取预设倍率')}
        visible={presetModalVisible}
        onCancel={() => setPresetModalVisible(false)}
        onOk={fetchPresetRatios}
        okText={t('保存')}
        confirmLoading={presetLoading}
        width={500}
      >
        <Space vertical style={{ width: '100%' }} spacing={16}>
          <div>
            <Text strong>{t('选择预设源：')}</Text>
            <Select
              value={selectedPresetSource}
              onChange={setSelectedPresetSource}
              style={{ width: '100%', marginTop: 8 }}
              optionList={presetSources}
              position='bottomLeft'
            />
          </div>
          
          <div style={{ 
            padding: '12px', 
            backgroundColor: 'var(--semi-color-fill-0)', 
            borderRadius: '6px',
            border: '1px solid var(--semi-color-border)'
          }}>
            <Text type='warning' strong>{t('注意事项：')}</Text>
            <ul style={{ margin: '8px 0', paddingLeft: '20px' }}>
              <li><Text type='secondary'>{t('将覆盖已有模型倍率（如果本地有同名模型则覆盖，否则保留）')}</Text></li>
              <li><Text type='secondary'>{t('请确认网络连接正常')}</Text></li>
              <li><Text type='secondary'>{t('建议在操作前备份当前配置')}</Text></li>
            </ul>
          </div>
          
          <div style={{ 
            padding: '8px', 
            borderTop: '1px solid var(--semi-color-border)', 
            marginTop: '16px' 
          }}>
            <Text size='small' type='tertiary'>
              {t('由 Veloera Public Assets 提供支持，')} 
              <a 
                href='https://public-assets.veloera.org/terms' 
                target='_blank'
                rel='noopener noreferrer'
                style={{ color: 'var(--semi-color-primary)', textDecoration: 'none' }}
              >
                {t('条款和条件')}
              </a>
              {t('适用。')}
            </Text>
          </div>
        </Space>
      </Modal>
      
      <Modal
        title={t('重置预设(移除预设中模型)')}
        visible={resetModalVisible}
        onCancel={() => setResetModalVisible(false)}
        onOk={resetPresetRatios}
        okText={t('重置')}
        confirmLoading={resetLoading}
        width={500}
      >
        <Space vertical style={{ width: '100%' }} spacing={16}>
          <div>
            <Text strong>{t('选择要重置的预设源：')}</Text>
            <Select
              value={selectedResetSource}
              onChange={setSelectedResetSource}
              style={{ width: '100%', marginTop: 8 }}
              optionList={presetSources}
              position='bottomLeft'
            />
          </div>
          
          <div style={{ 
            padding: '12px', 
            backgroundColor: 'var(--semi-color-fill-0)', 
            borderRadius: '6px',
            border: '1px solid var(--semi-color-border)'
          }}>
            <Text type='danger' strong>{t('警告：')}</Text>
            <ul style={{ margin: '8px 0', paddingLeft: '20px' }}>
              <li><Text type='secondary'>{t('将从本地配置中移除选定预设源中包含的所有模型')}</Text></li>
              <li><Text type='secondary'>{t('此操作不可撤销，请谨慎操作')}</Text></li>
              <li><Text type='secondary'>{t('建议在操作前备份当前配置')}</Text></li>
            </ul>
          </div>
          
          <div style={{ 
            padding: '8px', 
            borderTop: '1px solid var(--semi-color-border)', 
            marginTop: '16px' 
          }}>
            <Text size='small' type='tertiary'>
              {t('由 Veloera Public Assets 提供支持，')} 
              <a 
                href='https://public-assets.veloera.org/terms' 
                target='_blank'
                rel='noopener noreferrer'
                style={{ color: 'var(--semi-color-primary)', textDecoration: 'none' }}
              >
                {t('条款和条件')}
              </a>
              {t('适用。')}
            </Text>
          </div>
        </Space>
      </Modal>
    </Spin>
  );
}