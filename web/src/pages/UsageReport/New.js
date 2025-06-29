import React, { useState } from 'react';
import { API, showError, showSuccess } from '../../helpers';
import { Button, Form, Layout, Select } from '@douyinfe/semi-ui';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const OPTION_LIST = [
  { label: '渠道', value: 'channel' },
  { label: '用户', value: 'user' },
  { label: '令牌', value: 'token' },
  { label: '模型', value: 'model' },
  { label: 'IP地址', value: 'ip' },
];

const NewReport = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [items, setItems] = useState([]);

  const handleSubmit = async () => {
    const res = await API.post('/api/usage-report', { name, items });
    if (res.data.success) {
      showSuccess('');
      navigate('/usage-report');
    } else {
      showError(res.data.message);
    }
  };

  return (
    <Layout>
      <Layout.Header>{t('生成报表')}</Layout.Header>
      <Layout.Content>
        <Form>
          <Form.Input
            field='name'
            label={t('报表名称')}
            value={name}
            onChange={setName}
          />
          <Form.Field label={t('统计项目')}>
            <Select
              multiple
              optionList={OPTION_LIST}
              value={items}
              onChange={setItems}
            />
          </Form.Field>
          <Button
            type='primary'
            onClick={handleSubmit}
            style={{ marginTop: 16 }}
          >
            {t('生成报表')}
          </Button>
        </Form>
      </Layout.Content>
    </Layout>
  );
};

export default NewReport;
