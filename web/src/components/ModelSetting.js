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
import React, { useEffect, useState } from 'react';
import { Card, Spin, Tabs } from '@douyinfe/semi-ui';

import { API, showError, showSuccess } from '../helpers';
import { useTranslation } from 'react-i18next';
import SettingGeminiModel from '../pages/Setting/Model/SettingGeminiModel.js';
import SettingClaudeModel from '../pages/Setting/Model/SettingClaudeModel.js';
import SettingGlobalModel from '../pages/Setting/Model/SettingGlobalModel.js';

const ModelSetting = () => {
  const { t } = useTranslation();
  let [inputs, setInputs] = useState({
    'gemini.safety_settings': '',
    'gemini.version_settings': '',
    'gemini.supported_imagine_models': '',
    'claude.model_headers_settings': '',
    'claude.thinking_adapter_enabled': true,
    'claude.default_max_tokens': '',
    'claude.thinking_adapter_budget_tokens_percentage': 0.8,
    'global.pass_through_request_enabled': false,
    'global.hide_upstream_error_enabled': false,
    'global.block_browser_extension_enabled': false,
    'global.rate_limit_exempt_enabled': false,
    'global.rate_limit_exempt_group': 'bulk-ok',
    'global.safe_check_exempt_enabled': false,
    'global.safe_check_exempt_group': 'nsfw-ok',
    'general_setting.ping_interval_enabled': false,
    'general_setting.ping_interval_seconds': 60,
    'gemini.thinking_adapter_enabled': false,
    'gemini.thinking_adapter_budget_tokens_percentage': 0.6,
    'gemini.models_supported_thinking_budget': '',
  });

  let [loading, setLoading] = useState(false);

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (
          item.key === 'gemini.safety_settings' ||
          item.key === 'gemini.version_settings' ||
          item.key === 'claude.model_headers_settings' ||
          item.key === 'claude.default_max_tokens' ||
          item.key === 'gemini.supported_imagine_models' ||
          item.key === 'gemini.models_supported_thinking_budget'
        ) {
          item.value = JSON.stringify(JSON.parse(item.value), null, 2);
        }
        if (item.key.endsWith('Enabled') || item.key.endsWith('enabled')) {
          newInputs[item.key] = item.value === 'true' ? true : false;
        } else {
          newInputs[item.key] = item.value;
        }
      });

      setInputs(newInputs);
    } else {
      showError(message);
    }
  };
  async function onRefresh() {
    try {
      setLoading(true);
      await getOptions();
      // showSuccess('刷新成功');
    } catch (error) {
      showError('刷新失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    onRefresh();
  }, []);

  return (
    <>
      <Spin spinning={loading} size='large'>
        {/* OpenAI */}
        <Card style={{ marginTop: '10px' }}>
          <SettingGlobalModel options={inputs} refresh={onRefresh} />
        </Card>
        {/* Gemini */}
        <Card style={{ marginTop: '10px' }}>
          <SettingGeminiModel options={inputs} refresh={onRefresh} />
        </Card>
        {/* Claude */}
        <Card style={{ marginTop: '10px' }}>
          <SettingClaudeModel options={inputs} refresh={onRefresh} />
        </Card>
      </Spin>
    </>
  );
};

export default ModelSetting;
