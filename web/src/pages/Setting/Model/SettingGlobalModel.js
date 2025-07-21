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
import { Button, Col, Form, Row, Spin, Banner } from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
  verifyJSON,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingGlobalModel(props) {
  const { t } = useTranslation();

  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    'global.pass_through_request_enabled': false,
    'global.hide_upstream_error_enabled': false,
    'global.block_browser_extension_enabled': false,
    'global.rate_limit_exempt_enabled': false,
    'global.rate_limit_exempt_group': 'bulk-ok',
    'global.safe_check_exempt_enabled': false,
    'global.safe_check_exempt_group': 'nsfw-ok',
    'general_setting.ping_interval_enabled': false,
    'general_setting.ping_interval_seconds': 60,
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  function onSubmit() {
    const updateArray = compareObjects(inputs, inputsRow);
    if (!updateArray.length) return showWarning(t('你似乎并没有修改什么'));
    const requestQueue = updateArray.map((item) => {
      let value = String(inputs[item.key]);

      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1) {
          if (res.includes(undefined)) return;
        } else if (requestQueue.length > 1) {
          if (res.includes(undefined))
            return showError(t('部分保存失败，请重试'));
        }
        showSuccess(t('保存成功'));
        props.refresh();
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
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);
  }, [props.options]);

  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('全局设置')}>
            <Row>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  label={t('启用请求透传')}
                  field={'global.pass_through_request_enabled'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'global.pass_through_request_enabled': value,
                    })
                  }
                  extraText={
                    '开启后，所有请求将直接透传给上游，不会进行任何处理（重定向和渠道适配也将失效）,请谨慎开启'
                  }
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  label={t('隐藏上游报错信息')}
                  field={'global.hide_upstream_error_enabled'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'global.hide_upstream_error_enabled': value,
                    })
                  }
                  extraText={'开启后，只返回统一的上游错误信息'}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  label={t('是否允许沉浸式翻译类插件')}
                  field={'global.block_browser_extension_enabled'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'global.block_browser_extension_enabled': value,
                    })
                  }
                  extraText={
                    '是否允许浏览器插件请求, 请注意此判断逻辑不可靠, 并可能误杀!'
                  }
                />
              </Col>
            </Row>

            <Form.Section text={t('规则豁免设置')}>
              <Row>
                <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                  <Form.Switch
                    label={t('启用速率限制豁免')}
                    field={'global.rate_limit_exempt_enabled'}
                    onChange={(value) =>
                      setInputs({
                        ...inputs,
                        'global.rate_limit_exempt_enabled': value,
                      })
                    }
                  />
                </Col>
                <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                  <Form.Input
                    label={t('豁免分组')}
                    field={'global.rate_limit_exempt_group'}
                    onChange={(value) =>
                      setInputs({
                        ...inputs,
                        'global.rate_limit_exempt_group': value,
                      })
                    }
                    disabled={!inputs['global.rate_limit_exempt_enabled']}
                  />
                </Col>
              </Row>
              <Row>
                <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                  <Form.Switch
                    label={t('启用安全审查豁免')}
                    field={'global.safe_check_exempt_enabled'}
                    onChange={(value) =>
                      setInputs({
                        ...inputs,
                        'global.safe_check_exempt_enabled': value,
                      })
                    }
                  />
                </Col>
                <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                  <Form.Input
                    label={t('豁免分组')}
                    field={'global.safe_check_exempt_group'}
                    onChange={(value) =>
                      setInputs({
                        ...inputs,
                        'global.safe_check_exempt_group': value,
                      })
                    }
                    disabled={!inputs['global.safe_check_exempt_enabled']}
                  />
                </Col>
              </Row>
            </Form.Section>

            <Form.Section text={t('连接保活设置')}>
              <Row style={{ marginTop: 10 }}>
                <Col span={24}>
                  <Banner
                    type='warning'
                    description='警告：启用保活后，如果已经写入保活数据后渠道出错，系统无法重试，如果必须开启，推荐设置尽可能大的Ping间隔'
                  />
                </Col>
              </Row>
              <Row>
                <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                  <Form.Switch
                    label={t('启用Ping间隔')}
                    field={'general_setting.ping_interval_enabled'}
                    onChange={(value) =>
                      setInputs({
                        ...inputs,
                        'general_setting.ping_interval_enabled': value,
                      })
                    }
                    extraText={'开启后，将定期发送ping数据保持连接活跃'}
                  />
                </Col>
                <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                  <Form.InputNumber
                    label={t('Ping间隔（秒）')}
                    field={'general_setting.ping_interval_seconds'}
                    onChange={(value) =>
                      setInputs({
                        ...inputs,
                        'general_setting.ping_interval_seconds': value,
                      })
                    }
                    min={1}
                    disabled={!inputs['general_setting.ping_interval_enabled']}
                  />
                </Col>
              </Row>
            </Form.Section>

            <Row>
              <Button size='default' onClick={onSubmit}>
                {t('保存')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
