import React, { useEffect, useState } from 'react';
import { API, showError } from '../../helpers';
import { Layout, Card, Table, Row, Col, Typography } from '@douyinfe/semi-ui';
import {
  IconHistogram,
  IconBarChartHStroked,
  IconAlertTriangle,
} from '@douyinfe/semi-icons';
import { renderNumber } from '../../helpers/render';
import { useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const Metric = ({ icon, title, value }) => (
  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
    {icon}
    <div>
      <Typography.Text type='secondary'>{title}</Typography.Text>
      <Typography.Title heading={5} style={{ margin: 0 }}>
        {value}
      </Typography.Title>
    </div>
  </div>
);

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

  const parsed = report ? JSON.parse(report.data || '{}') : {};
  const overview = parsed.overview || {};

  const renderTable = (title, dataSource, columns) => (
    <Card style={{ marginTop: 24 }} title={title}>
      <Table pagination={false} dataSource={dataSource} columns={columns} />
    </Card>
  );

  return (
    <Layout>
      <Layout.Header>{t('报表')}</Layout.Header>
      <Layout.Content>
        {report && (
          <>
            <Card title={t('总览')}>
              <Row gutter={16}>
                <Col span={6}>
                  <Metric
                    icon={<IconBarChartHStroked />}
                    title={t('总消耗Tokens')}
                    value={renderNumber(overview.total_tokens || 0)}
                  />
                </Col>
                <Col span={6}>
                  <Metric
                    icon={<IconHistogram />}
                    title={t('总请求次数')}
                    value={renderNumber(overview.total_requests || 0)}
                  />
                </Col>
                <Col span={6}>
                  <Metric
                    icon={<IconAlertTriangle />}
                    title='429'
                    value={renderNumber(overview.error_429 || 0)}
                  />
                </Col>
                <Col span={6}>
                  <Metric
                    icon={<IconAlertTriangle />}
                    title={t('普通错误')}
                    value={renderNumber(overview.normal_error || 0)}
                  />
                </Col>
              </Row>
            </Card>

            {parsed.channel &&
              renderTable(t('渠道统计'), parsed.channel, [
                {
                  title: t('名称'),
                  render: (text, record) =>
                    `${record.ChannelName}(${record.ChannelID})`,
                },
                {
                  title: t('请求次数'),
                  dataIndex: 'RequestCount',
                  render: renderNumber,
                },
                {
                  title: t('消耗Tokens'),
                  dataIndex: 'TotalTokens',
                  render: renderNumber,
                },
              ])}

            {parsed.user &&
              renderTable(t('用户统计'), parsed.user, [
                {
                  title: t('名称'),
                  render: (text, record) =>
                    `${record.Username}(${record.UserID})`,
                },
                {
                  title: t('请求次数'),
                  dataIndex: 'RequestCount',
                  render: renderNumber,
                },
                {
                  title: t('消耗Tokens'),
                  dataIndex: 'TotalTokens',
                  render: renderNumber,
                },
              ])}

            {parsed.token &&
              renderTable(t('令牌统计'), parsed.token, [
                {
                  title: t('名称'),
                  render: (text, record) =>
                    `${record.TokenName}(${record.TokenID})`,
                },
                {
                  title: t('请求次数'),
                  dataIndex: 'RequestCount',
                  render: renderNumber,
                },
                {
                  title: t('消耗Tokens'),
                  dataIndex: 'TotalTokens',
                  render: renderNumber,
                },
              ])}

            {parsed.model &&
              renderTable(t('模型统计'), parsed.model, [
                {
                  title: t('名称'),
                  dataIndex: 'ModelName',
                },
                {
                  title: t('请求次数'),
                  dataIndex: 'RequestCount',
                  render: renderNumber,
                },
                {
                  title: t('消耗Tokens'),
                  dataIndex: 'TotalTokens',
                  render: renderNumber,
                },
              ])}

            {parsed.ip &&
              renderTable(t('IP统计'), parsed.ip, [
                {
                  title: t('名称'),
                  dataIndex: 'IP',
                },
                {
                  title: t('请求次数'),
                  dataIndex: 'RequestCount',
                  render: renderNumber,
                },
                {
                  title: t('消耗Tokens'),
                  dataIndex: 'TotalTokens',
                  render: renderNumber,
                },
              ])}
          </>
        )}
      </Layout.Content>
    </Layout>
  );
};

export default ViewReport;
