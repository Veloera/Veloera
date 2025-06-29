import React, { useEffect, useState } from 'react';
import { API, showError } from '../../helpers';
import {
  Layout,
  Row,
  Col,
  Card,
  Table,
  Statistic,
} from '@douyinfe/semi-ui';
import {
  IconHistogram,
  IconKey,
  IconAlertCircle,
} from '@douyinfe/semi-icons';
import { useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const ViewReport = () => {
  const { t } = useTranslation();
  const { id } = useParams();
  const [report, setReport] = useState(null);

  const fetchData = async () => {
    const res = await API.get('/api/usage-report/' + id);
    if (res.data.success) {
      setReport(res.data.data);
    } else {
      showError(res.data.message);
    }
  };

  useEffect(() => {
    fetchData();
  }, [id]);

  const data = report ? JSON.parse(report.data || '{}') : {};
  const overview = data.overview || {};

  const stats = [
    {
      title: '总请求',
      value: overview.total_requests || 0,
      icon: <IconHistogram size='extra-large' />,
    },
    {
      title: '总 Tokens',
      value: overview.total_tokens || 0,
      icon: <IconKey size='extra-large' />,
    },
    {
      title: '429 错误',
      value: overview.error_429 || 0,
      icon: <IconAlertCircle size='extra-large' />,
    },
    {
      title: '普通错误',
      value: overview.normal_error || 0,
      icon: <IconAlertCircle size='extra-large' />,
    },
  ];

  const renderStats = () => (
    <Row gutter={16} style={{ marginTop: 16 }}>
      {stats.map((s, idx) => (
        <Col span={6} key={idx}>
          <Card>
            <Statistic title={s.title} value={s.value} prefix={s.icon} />
          </Card>
        </Col>
      ))}
    </Row>
  );

  const columns = {
    channel: [
      {
        title: '渠道',
        render: (_, r) => `${r.channel_name}(${r.channel_id})`,
      },
      { title: '请求次数', dataIndex: 'request_count' },
      { title: 'Tokens', dataIndex: 'total_tokens' },
    ],
    user: [
      {
        title: '用户',
        render: (_, r) => `${r.username}(${r.user_id})`,
      },
      { title: '请求次数', dataIndex: 'request_count' },
      { title: 'Tokens', dataIndex: 'total_tokens' },
    ],
    token: [
      {
        title: '令牌',
        render: (_, r) => `${r.token_name}(${r.token_id})`,
      },
      { title: '请求次数', dataIndex: 'request_count' },
      { title: 'Tokens', dataIndex: 'total_tokens' },
    ],
    model: [
      { title: '模型', dataIndex: 'model_name' },
      { title: '请求次数', dataIndex: 'request_count' },
      { title: 'Tokens', dataIndex: 'total_tokens' },
    ],
    ip: [
      { title: 'IP', dataIndex: 'ip' },
      { title: '请求次数', dataIndex: 'request_count' },
      { title: 'Tokens', dataIndex: 'total_tokens' },
    ],
  };

  const renderTable = (key, title) => {
    const list = data[key];
    if (!list || list.length === 0) return null;
    return (
      <Card style={{ marginTop: 20 }} headerLine={false} title={title}>
        <Table columns={columns[key]} dataSource={list} pagination={false} />
      </Card>
    );
  };

  return (
    <Layout>
      <Layout.Header>{report ? report.name : t('报表')}</Layout.Header>
      <Layout.Content>
        {renderStats()}
        {renderTable('channel', '渠道')}
        {renderTable('user', '用户')}
        {renderTable('token', '令牌')}
        {renderTable('model', '模型')}
        {renderTable('ip', 'IP')}
      </Layout.Content>
    </Layout>
  );
};

export default ViewReport;
