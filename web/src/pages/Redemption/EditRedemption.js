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
import { useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  API,
  downloadTextAsFile,
  isMobile,
  showError,
  showSuccess,
} from '../../helpers';
import {
  getQuotaPerUnit,
  renderQuota,
  renderQuotaWithPrompt,
} from '../../helpers/render';
import {
  AutoComplete,
  Button,
  DatePicker,
  Input,
  Modal,
  SideSheet,
  Space,
  Spin,
  Typography,
  Switch,
} from '@douyinfe/semi-ui';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import { Divider } from 'semantic-ui-react';

const EditRedemption = (props) => {
  const { t } = useTranslation();
  const isEdit = props.editingRedemption.id !== undefined;
  const [loading, setLoading] = useState(isEdit);

  const params = useParams();
  const navigate = useNavigate();
  const originInputs = {
    name: '',
    quota: 100000,
    count: 1,
    is_gift: false,
    max_uses: -1,  // -1 means unlimited
    valid_from: 0,  // 0 means immediately effective
    valid_until: 0  // 0 means never expires
  };
  const [inputs, setInputs] = useState(originInputs);
  const [redemptionKey, setRedemptionKey] = useState(''); // New state for optional key
  const { name, quota, count, is_gift, max_uses, valid_from, valid_until } = inputs;

  const handleCancel = () => {
    props.handleClose();
  };

  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const loadRedemption = async () => {
    setLoading(true);
    let res = await API.get(`/api/redemption/${props.editingRedemption.id}`);
    const { success, message, data } = res.data;
    if (success) {
      setInputs(data);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    if (isEdit) {
      loadRedemption().then(() => {
        // console.log(inputs);
      });
    } else {
      setInputs(originInputs);
    }
  }, [props.editingRedemption.id]);

  const submit = async () => {
    let name = inputs.name;
    if (!isEdit && inputs.name === '') {
      name = renderQuota(quota);
    }
    
    // 验证时间范围
    if (valid_from > 0 && valid_until > 0 && valid_from >= valid_until) {
      showError(t('生效时间必须早于过期时间'));
      return;
    }
    setLoading(true);
    let localInputs = inputs;
    localInputs.count = parseInt(localInputs.count);
    localInputs.quota = parseInt(localInputs.quota);
    localInputs.max_uses = parseInt(localInputs.max_uses);
    localInputs.valid_from = parseInt(localInputs.valid_from);
    localInputs.valid_until = parseInt(localInputs.valid_until);
    localInputs.name = name;
    let res;
    if (isEdit) {
      res = await API.put(`/api/redemption/`, {
        ...localInputs,
        id: parseInt(props.editingRedemption.id),
      });
    } else {
      res = await API.post(`/api/redemption/`, {
        ...localInputs,
        // Add the optional key if provided
        key: redemptionKey.trim() === '' ? undefined : redemptionKey.trim(),
      });
    }
    const { success, message, data } = res.data;
    if (success) {
      if (isEdit) {
        showSuccess(t('兑换码更新成功！'));
        props.refresh();
        props.handleClose();
      } else {
        showSuccess(t('兑换码创建成功！'));
        setInputs(originInputs);
        props.refresh();
        props.handleClose();
      }
    } else {
      showError(message);
    }
    if (!isEdit && data) {
      let text = '';
      for (let i = 0; i < data.length; i++) {
        text += data[i] + '\n';
      }
      Modal.confirm({
        title: t('兑换码创建成功'),
        content: (
          <div>
            <p>{t('兑换码创建成功，是否下载兑换码？')}</p>
            <p>{t('兑换码将以文本文件的形式下载，文件名为兑换码的名称。')}</p>
          </div>
        ),
        onOk: () => {
          downloadTextAsFile(text, `${inputs.name}.txt`);
        },
      });
    }
    setLoading(false);
  };

  return (
    <>
      <SideSheet
        placement={isEdit ? 'right' : 'left'}
        title={
          <Title level={3}>
            {isEdit ? t('更新兑换码信息') : t('创建新的兑换码')}
          </Title>
        }
        headerStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        bodyStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        visible={props.visiable}
        footer={
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Space>
              <Button theme='solid' size={'large'} onClick={submit}>
                {t('提交')}
              </Button>
              <Button
                theme='solid'
                size={'large'}
                type={'tertiary'}
                onClick={handleCancel}
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
          <Input
            style={{ marginTop: 20 }}
            label={t('名称')}
            name='name'
            placeholder={t('请输入名称')}
            onChange={(value) => handleInputChange('name', value)}
            value={name}
            autoComplete='new-password'
            required={!isEdit}
          />
          <Divider />
          <div style={{ marginTop: 20 }}>
            <Typography.Text>
              {t('额度') + renderQuotaWithPrompt(quota)}
            </Typography.Text>
          </div>
          <AutoComplete
            style={{ marginTop: 8 }}
            name='quota'
            placeholder={t('请输入额度')}
            onChange={(value) => handleInputChange('quota', value)}
            value={quota}
            autoComplete='new-password'
            type='number'
            position={'bottom'}
            data={[
              { value: 500000, label: '1$' },
              { value: 5000000, label: '10$' },
              { value: 25000000, label: '50$' },
              { value: 50000000, label: '100$' },
              { value: 250000000, label: '500$' },
              { value: 500000000, label: '1000$' },
            ]}
          />
          {!isEdit && (
            <>
              <Divider />
              <Typography.Text>{t('生成数量')}</Typography.Text>
              <Input
                style={{ marginTop: 8 }}
                label={t('生成数量')}
                name='count'
                placeholder={t('请输入生成数量')}
                onChange={(value) => handleInputChange('count', value)}
                value={count}
                autoComplete='new-password'
                type='number'
              />
            </>
          )}

          {!isEdit && (
            <>
              <Divider />
              <Input
                style={{ marginTop: 8 }}
                label={t('兑换码内容 (可选)')}
                name='key'
                placeholder={t('留空则自动生成')}
                onChange={(value) => setRedemptionKey(value)}
                value={redemptionKey}
                autoComplete='new-password'
              />
            </>
          )}

          <Divider />
          <Typography.Text>{t('兑换码类型')}</Typography.Text>
          <div style={{ marginTop: 8 }}>
            <Space>
              <Typography.Text>{t('礼品码')}</Typography.Text>
              <Switch
                checked={inputs.is_gift}
                onChange={(checked) => handleInputChange('is_gift', checked)}
              />
            </Space>
          </div>

          {inputs.is_gift && (
            <>
              <Divider />
              <Typography.Text>{t('最大使用次数')}</Typography.Text>
              <Input
                style={{ marginTop: 8 }}
                label={t('最大使用次数')}
                name='max_uses'
                placeholder={t('留空或-1表示无限制')}
                onChange={(value) => handleInputChange('max_uses', value)}
                value={inputs.max_uses}
                type='number'
              />
            </>
          )}

          <Divider />
          <Typography.Text>{t('生效时间')}</Typography.Text>
          <DatePicker
            style={{ marginTop: 8, width: '100%' }}
            type='dateTime'
            placeholder={t('留空表示立即生效')}
            value={valid_from > 0 ? new Date(valid_from * 1000) : null}
            onChange={(date) => {
              const timestamp = date ? Math.floor(date.getTime() / 1000) : 0;
              handleInputChange('valid_from', timestamp);
            }}
            format='yyyy-MM-dd HH:mm:ss'
          />

          <Divider />
          <Typography.Text>{t('过期时间')}</Typography.Text>
          <DatePicker
            style={{ marginTop: 8, width: '100%' }}
            type='dateTime'
            placeholder={t('留空表示永不过期')}
            value={valid_until > 0 ? new Date(valid_until * 1000) : null}
            onChange={(date) => {
              const timestamp = date ? Math.floor(date.getTime() / 1000) : 0;
              handleInputChange('valid_until', timestamp);
            }}
            format='yyyy-MM-dd HH:mm:ss'
          />
        </Spin>
      </SideSheet>
    </>
  );
};

export default EditRedemption;
