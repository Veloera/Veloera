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
  Popconfirm,
  Row,
  Space,
  Spin,
} from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
  verifyJSON,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function ModelRatioSettings(props) {
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    ModelPrice: '',
    ModelRatio: '',
    CacheRatio: '',
    CompletionRatio: '',
    fallback_pricing_enabled: false,
    fallback_single_price: '',
    fallback_input_ratio: '',
    fallback_completion_ratio: '',
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);
  const { t } = useTranslation();

  const handleFallbackPriceChange = (type, value) => {
    const newInputs = {...inputs};
    
    if (type === 'single') {
      newInputs.fallback_single_price = value;
      if (value) {
        // Clear ratio fields when single price is entered
        newInputs.fallback_input_ratio = '';
        newInputs.fallback_completion_ratio = '';
      }
    } else {
      newInputs[`fallback_${type}_ratio`] = value;
      if (value) {
        // Clear single price when either ratio field is entered
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
    
    return true; // All empty is valid
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

  async function onSubmit() {
    try {
      await refForm.current
        .validate()
        .then(() => {
          const updateArray = compareObjects(inputs, inputsRow);
          if (!updateArray.length)
            return showWarning(t('你似乎并没有修改什么'));

          const requestQueue = updateArray.map((item) => {
            const value =
              typeof inputs[item.key] === 'boolean'
                ? String(inputs[item.key])
                : inputs[item.key];
            return API.put('/api/option/', { key: item.key, value });
          });

          setLoading(true);
          Promise.all(requestQueue)
            .then((res) => {
              if (res.includes(undefined)) {
                return showError(
                  requestQueue.length > 1
                    ? t('部分保存失败，请重试')
                    : t('保存失败'),
                );
              }

              for (let i = 0; i < res.length; i++) {
                if (!res[i].data.success) {
                  return showError(res[i].data.message);
                }
              }

              showSuccess(t('保存成功'));
              props.refresh();
            })
            .catch((error) => {
              console.error('Unexpected error:', error);
              showError(t('保存失败，请重试'));
            })
            .finally(() => {
              setLoading(false);
            });
        })
        .catch(() => {
          showError(t('请检查输入'));
        });
    } catch (error) {
      showError(t('请检查输入'));
      console.error(error);
    }
  }

  async function resetModelRatio() {
    try {
      let res = await API.post(`/api/option/rest_model_ratio`);
      if (res.data.success) {
        showSuccess(res.data.message);
        props.refresh();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error);
    }
  }

  useEffect(() => {
    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        if (key === 'fallback_pricing_enabled') {
          // Convert string to boolean for the toggle switch
          currentInputs[key] = props.options[key] === 'true';
        } else {
          currentInputs[key] = props.options[key];
        }
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);
  }, [props.options]);

  return (
    <Spin spinning={loading}>
      <Form
        values={inputs}
        getFormApi={(formAPI) => (refForm.current = formAPI)}
        style={{ marginBottom: 15 }}
      >
        <Form.Section>
          <Row gutter={16}>
            <Col span={24}>
              <Form.Switch
                label={t('启用兜底倍率')}
                field="fallback_pricing_enabled"
                onChange={(checked) => setInputs({...inputs, fallback_pricing_enabled: checked})}
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
                  onChange={(value) => handleFallbackPriceChange('single', value)}
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
                  onChange={(value) => handleFallbackPriceChange('input', value)}
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
                  onChange={(value) => handleFallbackPriceChange('completion', value)}
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
        <Form.Section>
          <Row gutter={16}>
            <Col xs={24} sm={16}>
              <Form.TextArea
                label={t('模型固定价格')}
                extraText={t('一次调用消耗多少刀，优先级大于模型倍率')}
                placeholder={t(
                  '为一个 JSON 文本，键为模型名称，值为一次调用消耗多少刀，比如 "gpt-4-gizmo-*": 0.1，一次消耗0.1刀',
                )}
                field={'ModelPrice'}
                autosize={{ minRows: 6, maxRows: 12 }}
                trigger='blur'
                stopValidateWithError
                rules={[
                  {
                    validator: (rule, value) => verifyJSON(value),
                    message: '不是合法的 JSON 字符串',
                  },
                ]}
                onChange={(value) =>
                  setInputs({ ...inputs, ModelPrice: value })
                }
              />
            </Col>
          </Row>
          <Row gutter={16}>
            <Col xs={24} sm={16}>
              <Form.TextArea
                label={t('模型倍率')}
                placeholder={t('为一个 JSON 文本，键为模型名称，值为倍率')}
                field={'ModelRatio'}
                autosize={{ minRows: 6, maxRows: 12 }}
                trigger='blur'
                stopValidateWithError
                rules={[
                  {
                    validator: (rule, value) => verifyJSON(value),
                    message: '不是合法的 JSON 字符串',
                  },
                ]}
                onChange={(value) =>
                  setInputs({ ...inputs, ModelRatio: value })
                }
              />
            </Col>
          </Row>
          <Row gutter={16}>
            <Col xs={24} sm={16}>
              <Form.TextArea
                label={t('提示缓存倍率')}
                placeholder={t('为一个 JSON 文本，键为模型名称，值为倍率')}
                field={'CacheRatio'}
                autosize={{ minRows: 6, maxRows: 12 }}
                trigger='blur'
                stopValidateWithError
                rules={[
                  {
                    validator: (rule, value) => verifyJSON(value),
                    message: '不是合法的 JSON 字符串',
                  },
                ]}
                onChange={(value) =>
                  setInputs({ ...inputs, CacheRatio: value })
                }
              />
            </Col>
          </Row>
          <Row gutter={16}>
            <Col xs={24} sm={16}>
              <Form.TextArea
                label={t('模型补全倍率（仅对自定义模型有效）')}
                extraText={t('仅对自定义模型有效')}
                placeholder={t('为一个 JSON 文本，键为模型名称，值为倍率')}
                field={'CompletionRatio'}
                autosize={{ minRows: 6, maxRows: 12 }}
                trigger='blur'
                stopValidateWithError
                rules={[
                  {
                    validator: (rule, value) => verifyJSON(value),
                    message: '不是合法的 JSON 字符串',
                  },
                ]}
                onChange={(value) =>
                  setInputs({ ...inputs, CompletionRatio: value })
                }
              />
            </Col>
          </Row>
        </Form.Section>
      </Form>
      <Space>
        <Button onClick={onSubmit}>{t('保存模型倍率设置')}</Button>
        <Popconfirm
          title={t('确定重置模型倍率吗？')}
          content={t('此修改将不可逆')}
          okType={'danger'}
          position={'top'}
          onConfirm={resetModelRatio}
        >
          <Button type={'danger'}>{t('重置模型倍率')}</Button>
        </Popconfirm>
      </Space>
    </Spin>
  );
}
