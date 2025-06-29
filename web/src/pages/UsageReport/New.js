import React, { useState } from 'react';
import { Form, Button } from '@douyinfe/semi-ui';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { API, showSuccess } from '../../helpers';

const NewReport = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (values) => {
    setSubmitting(true);
    try {
      const res = await API.post('/api/usage-report/', { name: values.name });
      if (res.data.success) {
        showSuccess(t('生成成功'));
        navigate('/usage-report');
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Form onSubmit={handleSubmit} style={{ maxWidth: 400 }}>
      <Form.Input field='name' label={t('名称')} required />
      <Button htmlType='submit' loading={submitting} type='primary'>
        {t('生成新报告')}
      </Button>
    </Form>
  );
};

export default NewReport;
