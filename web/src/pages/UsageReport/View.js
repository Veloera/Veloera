import React, { useEffect, useState } from 'react';
import { API, showError } from '../../helpers';
import { Layout, Card, Row, Col, Statistic, Table } from '@douyinfe/semi-ui';
import {
  IconHistogram,
  IconSend,
  IconAlertTriangle,
  IconInfoCircle,
} from '@douyinfe/semi-icons';
import { renderNumber } from '../../helpers/render';
import { useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const ViewReport = () => {
  const { t } = useTranslation();
  const { id } = useParams();
  const [data, setData] = useState(null);
  const [parsed, setParsed] = useState({});

  const fetchData = async () => {
    const res = await API.get('/api/usage-report/' + id);
    if (res.data.success) {
      setData(res.data.data);
      try {
        setParsed(JSON.parse(res.data.data.data || '{}'));
      } catch (e) {
        setParsed({});
      }
    } else {
      showError(res.data.message);
    }
  };

  useEffect(() => {
    fetchData();
  }, [id]);

  const overview = parsed.overview || {};
  const tables = ['channel', 'user', 'token', 'model', 'ip'];

  const columnsMap = {
    channel: [
      {
        title: t('渠道'),
        dataIndex: 'channel_name',
        render: (v, r) => `${r.channel_name} (#${r.channel_id})`,
      },
      {
        title: t('请求次数'),
        dataIndex: 'request_count',
        render: renderNumber,
      },
      {
        title: t('消耗Tokens'),
        dataIndex: 'total_tokens',
        render: renderNumber,
      },
    ],
    user: [
      {
        title: t('用户'),
        dataIndex: 'username',
        render: (v, r) => `${r.username} (#${r.user_id})`,
      },
      {
        title: t('请求次数'),
        dataIndex: 'request_count',
        render: renderNumber,
      },
      {
        title: t('消耗Tokens'),
        dataIndex: 'total_tokens',
        render: renderNumber,
      },
    ],
    token: [
      {
        title: t('Token'),
        dataIndex: 'token_name',
        render: (v, r) => `${r.token_name} (#${r.token_id})`,
      },
      {
        title: t('请求次数'),
        dataIndex: 'request_count',
        render: renderNumber,
      },
      {
        title: t('消耗Tokens'),
        dataIndex: 'total_tokens',
        render: renderNumber,
      },
    ],
    model: [
      { title: t('模型'), dataIndex: 'model_name' },
      {
        title: t('请求次数'),
        dataIndex: 'request_count',
        render: renderNumber,
      },
      {
        title: t('消耗Tokens'),
        dataIndex: 'total_tokens',
        render: renderNumber,
      },
    ],
    ip: [
      { title: t('IP'), dataIndex: 'ip' },
      {
        title: t('请求次数'),
        dataIndex: 'request_count',
        render: renderNumber,
      },
      {
        title: t('消耗Tokens'),
        dataIndex: 'total_tokens',
        render: renderNumber,
      },
    ],
  };

  return (
    <Layout>
      <Layout.Header>{data ? data.name : t('报表')}</Layout.Header>
      <Layout.Content>
        {data && (
          <>
            <Card style={{ marginBottom: 20 }}>
              <Row gutter={16} type='flex'>
                <Col span={6}>
                  <Statistic
                    title={t('总消耗Tokens')}
                    value={renderNumber(overview.total_tokens || 0)}
                    prefix={<IconHistogram />}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title={t('总请求次数')}
                    value={renderNumber(overview.total_requests || 0)}
                    prefix={<IconSend />}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title='429'
                    value={renderNumber(overview.error_429 || 0)}
                    prefix={<IconAlertTriangle />}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title={t('普通错误')}
                    value={renderNumber(overview.normal_error || 0)}
                    prefix={<IconInfoCircle />}
                  />
                </Col>
              </Row>
            </Card>
            {tables.map((key) =>
              parsed[key] && parsed[key].length > 0 ? (
                <Card style={{ marginTop: 20 }} key={key}>
                  <h3>{t(key)}</h3>
                  <Table
                    columns={columnsMap[key]}
                    dataSource={parsed[key]}
                    pagination={false}
                  />
                </Card>
              ) : null,
            )}
          </>
        )}
      </Layout.Content>
    </Layout>
  );
};

export default ViewReport;
