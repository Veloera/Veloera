import React, { useEffect, useState } from 'react';
import { API, showError, timestamp2string } from '../../helpers';
import { Button, Table, Layout } from '@douyinfe/semi-ui';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const UsageReport = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [items, setItems] = useState([]);

  const fetchData = async () => {
    const res = await API.get('/api/usage-report');
    if (res.data.success) {
      setItems(res.data.data);
    } else {
      showError(res.data.message);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const columns = [
    {
      title: t('报表名称'),
      dataIndex: 'name',
    },
    {
      title: t('创建时间'),
      dataIndex: 'created_time',
      render: (v) => timestamp2string(v),
    },
    {
      title: '',
      render: (_, record) => (
        <Button
          size='small'
          onClick={() => navigate('/usage-report/' + record.id)}
        >
          {t('查看')}
        </Button>
      ),
    },
  ];

  return (
    <Layout>
      <Layout.Header>
        <Button onClick={() => navigate('/usage-report/new')} type='primary'>
          {t('生成报表')}
        </Button>
      </Layout.Header>
      <Layout.Content>
        <Table columns={columns} dataSource={items} pagination={false} />
      </Layout.Content>
    </Layout>
  );
};

export default UsageReport;
