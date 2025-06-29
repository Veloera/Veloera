import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Card } from '@douyinfe/semi-ui';
import { API } from '../../helpers';

const ReportDetail = () => {
  const { id } = useParams();
  const [report, setReport] = useState(null);

  useEffect(() => {
    const fetchDetail = async () => {
      const res = await API.get(`/api/usage-report/${id}`);
      if (res.data.success) {
        setReport(res.data.data);
      }
    };
    fetchDetail();
  }, [id]);

  if (!report) return null;

  let content;
  try {
    const data = JSON.parse(report.data || '{}');
    content = (
      <pre style={{ whiteSpace: 'pre-wrap' }}>
        {JSON.stringify(data, null, 2)}
      </pre>
    );
  } catch (e) {
    content = <pre>{report.data}</pre>;
  }

  return (
    <Card title={report.name} style={{ maxWidth: 600 }}>
      {content}
    </Card>
  );
};

export default ReportDetail;
