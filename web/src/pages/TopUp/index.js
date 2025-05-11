import React, { useEffect, useState } from 'react';
import { API, isMobile, showError, showInfo, showSuccess } from '../../helpers';
import {
  renderNumber,
  renderQuota,
  renderQuotaWithAmount,
} from '../../helpers/render';
import {
  Col,
  Layout,
  Row,
  Typography,
  Card,
  Button,
  Form,
  Divider,
  Space,
  Modal,
  Toast,
} from '@douyinfe/semi-ui';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const TopUp = () => {
  const { t } = useTranslation();
  const [redemptionCode, setRedemptionCode] = useState('');
  const [topUpLink, setTopUpLink] = useState('');
  const [userQuota, setUserQuota] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const topUp = async () => {
    if (redemptionCode === '') {
      showInfo(t('请输入兑换码！'));
      return;
    }
    setIsSubmitting(true);
    try {
      const res = await API.post('/api/user/topup', {
        key: redemptionCode,
      });
      const { success, message, data } = res.data;
      if (success) {
        let successMessage = t('兑换成功！');
        if (data.is_gift) {
          successMessage = t('礼品码兑换成功！');
        }
        showSuccess(successMessage);

        // 确保 quota 是数字并且正确渲染
        const quotaAmount = parseInt(data.quota, 10);
        Modal.success({
          title: successMessage,
          content: t('成功兑换额度：') + renderQuotaWithAmount(quotaAmount),
          centered: true,
        });

        setUserQuota((quota) => quota + quotaAmount);
        setRedemptionCode('');
      } else {
        showError(message);
      }
    } catch (err) {
      showError(t('请求失败'));
    } finally {
      setIsSubmitting(false);
    }
  };

  const openTopUpLink = () => {
    if (!topUpLink) {
      showError(t('超级管理员未设置充值链接！'));
      return;
    }
    window.open(topUpLink, '_blank');
  };

  const getUserQuota = async () => {
    let res = await API.get(`/api/user/self`);
    const { success, message, data } = res.data;
    if (success) {
      setUserQuota(data.quota);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      if (status.top_up_link) {
        setTopUpLink(status.top_up_link);
      }
    }
    getUserQuota().then();
  }, []);

  return (
    <div>
      <Layout>
        <Layout.Header>
          <h3>{t('我的钱包')}</h3>
        </Layout.Header>
        <Layout.Content>
          <div
            style={{ marginTop: 20, display: 'flex', justifyContent: 'center' }}
          >
            <Card style={{ width: '500px', padding: '20px' }}>
              <Title level={3} style={{ textAlign: 'center' }}>
                {t('余额')} {renderQuota(userQuota)}
              </Title>
              <div style={{ marginTop: 20 }}>
                <Divider>{t('兑换余额')}</Divider>
                <Form>
                  <Form.Input
                    field={'redemptionCode'}
                    label={t('兑换码')}
                    placeholder={t('兑换码')}
                    name='redemptionCode'
                    value={redemptionCode}
                    onChange={(value) => {
                      setRedemptionCode(value);
                    }}
                  />
                  <Space>
                    {topUpLink ? (
                      <Button
                        type={'primary'}
                        theme={'solid'}
                        onClick={openTopUpLink}
                      >
                        {t('获取兑换码')}
                      </Button>
                    ) : null}
                    <Button
                      type={'warning'}
                      theme={'solid'}
                      onClick={topUp}
                      disabled={isSubmitting}
                    >
                      {isSubmitting ? t('兑换中...') : t('兑换')}
                    </Button>
                  </Space>
                </Form>
              </div>
              {/*<div style={{ marginTop: 20 }}>*/}
              {/*  <Divider>{t('在线充值')}</Divider>*/}
              {/*  <Form>*/}
              {/*    <Form.Input*/}
              {/*      disabled={!enableOnlineTopUp}*/}
              {/*      field={'redemptionCount'}*/}
              {/*      label={t('实付金额：') + ' ' + renderAmount()}*/}
              {/*      placeholder={*/}
              {/*        t('充值数量，最低 ') + renderQuotaWithAmount(minTopUp)*/}
              {/*      }*/}
              {/*      name='redemptionCount'*/}
              {/*      type={'number'}*/}
              {/*      value={topUpCount}*/}
              {/*      onChange={async (value) => {*/}
              {/*        if (value < 1)*/}
              {/*          value = 1;*/}
              {/*        setTopUpCount(value);*/}
              {/*        await getAmount(value);*/}
              {/*      }}*/}
              {/*    />*/}
              {/*    <Space>*/}
              {/*      <Button*/}
              {/*        type={'primary'}*/}
              {/*        theme={'solid'}*/}
              {/*        onClick={async () => {*/}
              {/*          preTopUp('zfb');*/}
              {/*        }}*/}
              {/*      >*/}
              {/*        {t('支付宝')}*/}
              {/*      </Button>*/}
              {/*      <Button*/}
              {/*        style={{*/}
              {/*          backgroundColor: 'rgba(var(--semi-green-5), 1)',*/}
              {/*        }}*/}
              {/*        type={'primary'}*/}
              {/*        theme={'solid'}*/}
              {/*        onClick={async () => {*/}
              {/*          preTopUp('wx');*/}
              {/*        }}*/}
              {/*      >*/}
              {/*        {t('微信')}*/}
              {/*      }*/}
              {/*      </Button>*/}
              {/*    </Space>*/}
              {/*  </Form>*/}
              {/*</div>*/}
              {/*<div style={{ display: 'flex', justifyContent: 'right' }}>*/}
              {/*    <Text>*/}
              {/*        <Link onClick={*/}
              {/*            async () => {*/}
              {/*                window.location.href = '/topup/history'*/}
              {/*            }*/}
              {/*        }>充值记录</Link>*/}
              {/*    </Text>*/}
              {/*</div>*/}
            </Card>
          </div>
        </Layout.Content>
      </Layout>
    </div>
  );
};

export default TopUp;
