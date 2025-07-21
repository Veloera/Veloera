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
import { Button, Col, Form, Row, Spin, Typography, Banner } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';

const { Text } = Typography;

export default function SettingsRebate(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    RebateEnabled: false,
    RebatePercentage: 0,
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  function handleFieldChange(fieldName) {
    return (value) => {
      setInputs((inputs) => ({ 
        ...inputs, 
        [fieldName]: typeof value === 'number' ? value : value 
      }));
    };
  }

  function onSubmit() {
    // 验证返佣百分比
    if (inputs.RebateEnabled && (inputs.RebatePercentage < 0 || inputs.RebatePercentage > 100)) {
      showError(t('返佣百分比必须在0到100之间'));
      return false;
    }

    const updateArray = compareObjects(inputs, inputsRow);
    if (!updateArray.length) {
      showWarning(t('你似乎并没有修改什么'));
      return false;
    }
    
    const requestQueue = updateArray.map((item) => {
      let value = '';
      if (typeof inputs[item.key] === 'boolean') {
        value = String(inputs[item.key]);
      } else {
        value = String(inputs[item.key]);
      }
      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });
    
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1) {
          if (res.includes(undefined)) {
            return false;
          }
        } else if (requestQueue.length > 1) {
          if (res.includes(undefined)) {
            showError(t('部分保存失败，请重试'));
            return false;
          }
        }
        showSuccess(t('保存成功'));
        props.refresh();
        return true;
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
  }

  useEffect(() => {
    const currentInputs = {};
    const validKeys = Object.keys(inputs);
    
    // 使用 Object.entries 和过滤来避免原型污染
    const safeOptions = Object.entries(props.options || {})
      .filter(([key]) => validKeys.includes(key))
      .reduce((acc, [key, value]) => {
        if (value !== null && value !== undefined) {
          // 使用 Object.defineProperty 来安全地设置属性
          Object.defineProperty(acc, key, {
            value: value,
            writable: true,
            enumerable: true,
            configurable: true
          });
        }
        return acc;
      }, Object.create(null));
    
    Object.assign(currentInputs, safeOptions);
    
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    if (refForm.current) {
      refForm.current.setValues(currentInputs);
    }
  }, [props.options]);

  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('返佣设置')}>
            <Banner
              type="info"
              description={
                <div>
                  <Text strong>返佣功能说明：</Text>
                  <br />
                  <Text>通过用户邀请下级用户在线充值 / 使用兑换码 进行返佣</Text>
                  <br />
                  <Text type="secondary">
                    示例：（返佣百分比为 50%）我邀请了一个用户，那个用户充值了1块钱，那么我可以返佣到0.5块钱
                  </Text>
                </div>
              }
              style={{ marginBottom: 16 }}
            />
            
            <Row gutter={16}>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  label={t('启用返佣功能')}
                  field={'RebateEnabled'}
                  onChange={handleFieldChange('RebateEnabled')}
                />
              </Col>
            </Row>
            
            <Row gutter={16}>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.InputNumber
                  label={t('返佣百分比')}
                  field={'RebatePercentage'}
                  step={0.1}
                  min={0}
                  max={100}
                  suffix={'%'}
                  extraText={t('0至100之间的数，可包含小数')}
                  placeholder={t('例如：50')}
                  onChange={handleFieldChange('RebatePercentage')}
                  disabled={!inputs.RebateEnabled}
                />
              </Col>
            </Row>

            <Row>
              <Button size='default' onClick={onSubmit}>
                {t('保存返佣设置')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
