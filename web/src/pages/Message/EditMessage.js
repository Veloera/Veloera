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
import React, { useState, useEffect } from 'react';
import { API, isMobile, showError, showSuccess } from '../../helpers';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import { 
  Button, 
  Input, 
  SideSheet, 
  Space, 
  Spin, 
  TextArea,
  RadioGroup,
  Radio,
  Table,
  Tag,
  Descriptions
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

const EditMessage = (props) => {
  const { t } = useTranslation();
  
  const [inputs, setInputs] = useState({
    title: '',
    content: '',
    format: 'markdown',
  });
  
  const [loading, setLoading] = useState(false);
  const [messageDetails, setMessageDetails] = useState(null);
  const [recipients, setRecipients] = useState([]);
  const [recipientLoading, setRecipientLoading] = useState(false);
  
  const { title, content, format } = inputs;

  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const loadMessageDetails = async (messageId) => {
    if (!messageId) return;
    
    setLoading(true);
    try {
      const res = await API.get(`/api/admin/messages/${messageId}`);
      const { success, message, data } = res.data;
      if (success) {
        setMessageDetails(data);
        setInputs({
          title: data.title || '',
          content: data.content || '',
          format: data.format || 'markdown',
        });
        loadRecipients(messageId);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setLoading(false);
  };

  const loadRecipients = async (messageId) => {
    setRecipientLoading(true);
    try {
      const res = await API.get(`/api/admin/messages/${messageId}/recipients`);
      const { success, message, data } = res.data;
      if (success) {
        setRecipients(data || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setRecipientLoading(false);
  };

  useEffect(() => {
    if (props.visible && props.editingMessage?.id) {
      loadMessageDetails(props.editingMessage.id);
    }
  }, [props.visible, props.editingMessage?.id]);

  const recipientColumns = [
    {
      title: 'ID',
      dataIndex: 'user_id',
      width: 80,
    },
    {
      title: t('用户名'),
      dataIndex: 'username',
    },
    {
      title: t('显示名称'),
      dataIndex: 'display_name',
      render: (text) => text || '-',
    },
    {
      title: t('阅读状态'),
      dataIndex: 'read_at',
      width: 120,
      render: (text) => (
        <Tag color={text ? 'green' : 'orange'} size='small'>
          {text ? t('已读') : t('未读')}
        </Tag>
      ),
    },
    {
      title: t('阅读时间'),
      dataIndex: 'read_at',
      width: 180,
      render: (text) => {
        return text ? new Date(text).toLocaleString() : '-';
      },
    },
  ];

  const submit = async () => {
    setLoading(true);
    
    if (!inputs.title.trim()) {
      showError(t('请输入消息标题'));
      setLoading(false);
      return;
    }
    
    if (!inputs.content.trim()) {
      showError(t('请输入消息内容'));
      setLoading(false);
      return;
    }

    const messageData = {
      title: inputs.title.trim(),
      content: inputs.content.trim(),
      format: inputs.format,
    };

    try {
      const res = await API.put(`/api/admin/messages/${props.editingMessage.id}`, messageData);
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('消息更新成功'));
        props.refresh();
        props.handleClose();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setLoading(false);
  };

  const handleCancel = () => {
    setInputs({
      title: '',
      content: '',
      format: 'markdown',
    });
    setMessageDetails(null);
    setRecipients([]);
    props.handleClose();
  };

  return (
    <>
      <SideSheet
        placement={'right'}
        title={<Title level={3}>{t('编辑消息')}</Title>}
        headerStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        bodyStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        visible={props.visible}
        footer={
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Space>
              <Button theme='solid' size={'large'} onClick={submit} loading={loading}>
                {t('更新消息')}
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
        width={isMobile() ? '100%' : 900}
      >
        <Spin spinning={loading}>
          <div style={{ padding: '20px 0' }}>
            {messageDetails && (
              <div style={{ marginBottom: 24 }}>
                <Descriptions
                  data={[
                    { key: t('消息ID'), value: messageDetails.id },
                    { key: t('创建时间'), value: new Date(messageDetails.created_at).toLocaleString() },
                    { key: t('收件人数'), value: recipients.length },
                    { key: t('已读人数'), value: recipients.filter(r => r.read_at).length },
                  ]}
                  row
                  size="small"
                />
              </div>
            )}

            <Input
              label={t('消息标题')}
              placeholder={t('请输入消息标题')}
              value={title}
              onChange={(value) => handleInputChange('title', value)}
              style={{ marginBottom: 20 }}
            />

            <div style={{ marginBottom: 20 }}>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 600 }}>
                {t('内容格式')}
              </label>
              <RadioGroup
                type="button"
                value={format}
                onChange={(e) => handleInputChange('format', e.target.value)}
              >
                <Radio value="markdown">Markdown</Radio>
                <Radio value="html">HTML</Radio>
              </RadioGroup>
            </div>

            <div style={{ marginBottom: 20 }}>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 600 }}>
                {t('消息内容')}
              </label>
              <TextArea
                placeholder={format === 'html' ? t('请输入HTML格式的消息内容') : t('请输入Markdown格式的消息内容')}
                value={content}
                onChange={(value) => handleInputChange('content', value)}
                rows={8}
                style={{ fontFamily: 'monospace' }}
              />
            </div>

            <div>
              <label style={{ display: 'block', marginBottom: 8, fontWeight: 600 }}>
                {t('收件人列表')}
              </label>
              
              <Table
                columns={recipientColumns}
                dataSource={recipients}
                loading={recipientLoading}
                pagination={{
                  pageSize: 10,
                  showSizeChanger: false,
                  size: 'small',
                }}
                size="small"
                style={{ maxHeight: 400, overflow: 'auto' }}
              />
            </div>
          </div>
        </Spin>
      </SideSheet>
    </>
  );
};

export default EditMessage;