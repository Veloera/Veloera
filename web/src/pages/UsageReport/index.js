import React, { useEffect, useState } from 'react';
import { Button, Table } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { API, timestamp2string } from '../../helpers';
import { Link } from 'react-router-dom';

const UsageReportList = () => {
  const { t } = useTranslation();
  const [reports, setReports] = useState([]);

  const fetchReports = async () => {
    const res = await API.get('/api/usage-report/');
    if (res.data.success) {
      setReports(res.data.data);
    }
  };

  useEffect(() => {
    fetchReports();
  }, []);

  const columns = [
    { title: 'ID', dataIndex: 'id' },
    { title: t('名称'), dataIndex: 'name' },
    {
      title: t('创建时间'),
      dataIndex: 'created_at',
      render: (text) => timestamp2string(text),
    },
    {
      title: '',
      render: (_, record) => (
        <Link to={`/usage-report/${record.id}`}>{t('查看')}</Link>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Link to='/usage-report/new'>
          <Button type='primary'>{t('生成新报告')}</Button>
        </Link>
      </div>
      <Table columns={columns} dataSource={reports} pagination={false} />
    </div>
  );
};

export default UsageReportList;
