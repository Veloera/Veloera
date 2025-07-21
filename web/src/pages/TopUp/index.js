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
import React, { useEffect, useState } from 'react';
import {
  API,
  isMobile,
  showError,
  showInfo,
  showSuccess,
} from '../../helpers';
import {
  renderNumber,
  renderQuota,
  renderQuotaWithAmount,
} from '../../helpers/render';
import {
  Layout,
  Card,
  Button,
  Form,
  Divider,
  Space,
  Modal,
  Toast,
  Banner,
} from '@douyinfe/semi-ui';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import { useTranslation } from 'react-i18next';

const TopUp = () => {
  const { t } = useTranslation();

  // --- 原有状态 ---
  const [redemptionCode, setRedemptionCode] = useState('');
  const [topUpCode, setTopUpCode] = useState('');
  const [topUpCount, setTopUpCount] = useState(0);
  const [minTopUpCount, setMinTopUpCount] = useState(1);
  const [amount, setAmount] = useState(0.0);
  const [minTopUp, setMinTopUp] = useState(1);
  const [topUpLink, setTopUpLink] = useState('');
  const [enableOnlineTopUp, setEnableOnlineTopUp] = useState(false);
  const [userQuota, setUserQuota] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [open, setOpen] = useState(false);
  const [payWay, setPayWay] = useState('');

  // --- 新增：标记 code 是否来源于 URL 且尚未重置 ---
  const [useUrlCode, setUseUrlCode] = useState(false);

  // 从 URL 读取 code、以及初始化管理员配置 & 用户余额
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const codeFromUrl = params.get('code');
    if (codeFromUrl) {
      setRedemptionCode(codeFromUrl);
      setUseUrlCode(true);
    }

    // 读取本地管理员配置
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      if (status.top_up_link) {
        setTopUpLink(status.top_up_link);
      }
      if (status.min_topup) {
        setMinTopUp(status.min_topup);
      }
      if (status.enable_online_topup) {
        setEnableOnlineTopUp(status.enable_online_topup);
      }
    }

    getUserQuota();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 点击“点我更改”
  const handleResetCode = () => {
    setRedemptionCode('');
    setUseUrlCode(false);
  };

  // 获取用户当前额度
  const getUserQuota = async () => {
    try {
      const res = await API.get('/api/user/self');
      const { success, message, data } = res.data;
      if (success) {
        setUserQuota(data.quota);
      } else {
        showError(message);
      }
    } catch (err) {
      showError(t('请求失败'));
    }
  };

  // 兑换余额
  const topUp = async () => {
    if (!redemptionCode) {
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

        const quotaAmount = parseInt(data.quota, 10);
        Modal.success({
          title: successMessage,
          content: t('成功兑换额度：') + renderQuotaWithAmount(quotaAmount),
          centered: true,
        });
        setUserQuota((q) => q + quotaAmount);

        // 兑换成功后清空，并恢复输入框
        setRedemptionCode('');
        setUseUrlCode(false);
      } else {
        showError(message);
      }
    } catch (err) {
      showError(t('请求失败'));
    } finally {
      setIsSubmitting(false);
    }
  };

  // 打开外部充值链接
  const openTopUpLink = () => {
    if (!topUpLink) {
      showError(t('超级管理员未设置充值链接！'));
      return;
    }
    window.open(topUpLink, '_blank');
  };

  // 预充值：校验后弹确认框
  const preTopUp = async (payment) => {
    if (!enableOnlineTopUp) {
      showError(t('管理员未开启在线充值！'));
      return;
    }
    await getAmount(topUpCount);
    if (topUpCount < minTopUp) {
      showError(t('充值数量不能小于') + minTopUp);
      return;
    }
    setPayWay(payment);
    setOpen(true);
  };

  // 执行在线充值
  const onlineTopUp = async () => {
    if (amount === 0) {
      await getAmount(topUpCount);
    }
    if (topUpCount < minTopUp) {
      showError(t('充值数量不能小于') + minTopUp);
      return;
    }
    setOpen(false);
    try {
      const res = await API.post('/api/user/pay', {
        amount: parseInt(topUpCount),
        top_up_code: topUpCode,
        payment_method: payWay,
      });
      const { message, data, url } = res.data;
      if (message === 'success') {
        // 动态构建表单并提交
        const form = document.createElement('form');
        form.action = url;
        form.method = 'POST';
        const isSafari =
          navigator.userAgent.indexOf('Safari') > -1 &&
          navigator.userAgent.indexOf('Chrome') < 0;
        if (!isSafari) {
          form.target = '_blank';
        }
        Object.keys(data).forEach((key) => {
          const input = document.createElement('input');
          input.type = 'hidden';
          input.name = key;
          input.value = data[key];
          form.appendChild(input);
        });
        document.body.appendChild(form);
        form.submit();
        document.body.removeChild(form);
      } else {
        showError(data);
      }
    } catch (err) {
      console.error(err);
      showError(t('支付请求失败'));
    }
  };

  // 获取实际需支付金额
  const getAmount = async (value) => {
    if (value === undefined) {
      value = topUpCount;
    }
    try {
      const res = await API.post('/api/user/amount', {
        amount: parseFloat(value),
        top_up_code: topUpCode,
      });
      const { message, data } = res.data;
      if (message === 'success') {
        setAmount(parseFloat(data));
      } else {
        setAmount(0);
        Toast.error({ content: '错误：' + data, id: 'getAmount' });
      }
    } catch (err) {
      console.error(err);
    }
  };

  const handleCancel = () => {
    setOpen(false);
  };

  // 渲染实付金额文字
  const renderAmountText = () => {
    return amount + ' ' + t('元');
  };

  return (
    <Layout>
      <Layout.Header>
        <h3>{t('我的钱包')}</h3>
      </Layout.Header>
      <Layout.Content>
        {/* 在线充值确认弹窗 */}
        <Modal
          title={t('确定要充值吗')}
          visible={open}
          onOk={onlineTopUp}
          onCancel={handleCancel}
          maskClosable={false}
          size="small"
          centered
        >
          <p>
            {t('充值数量')}：{topUpCount}
          </p>
          <p>
            {t('实付金额')}：{renderAmountText()}
          </p>
          <p>{t('是否确认充值？')}</p>
        </Modal>

        <div
          style={{
            marginTop: 20,
            display: 'flex',
            justifyContent: 'center',
          }}
        >
          <Card style={{ width: '500px', padding: '20px' }}>
            {/* 余额展示 */}
            <Title level={3} style={{ textAlign: 'center' }}>
              {t('余额')} {renderQuota(userQuota)}
            </Title>

            {/* 兑换余额 */}
            <div style={{ marginTop: 20 }}>
              <Divider>{t('兑换余额')}</Divider>
              <Form>
                {useUrlCode ? (
                  <p>
                    {t('已识别到兑换码')}:&nbsp;
                    <strong>{redemptionCode}</strong>。&nbsp;
                    <a
                      style={{ cursor: 'pointer', color: '#1890ff' }}
                      onClick={handleResetCode}
                    >
                      {t('点我更改')}
                    </a>
                  </p>
                ) : (
                  <Form.Input
                    field="redemptionCode"
                    label={t('兑换码')}
                    placeholder={t('请输入兑换码')}
                    value={redemptionCode}
                    onChange={(value) => setRedemptionCode(value)}
                  />
                )}

                <Space style={{ marginTop: 8 }}>
                  {topUpLink && (
                    <Button
                      type="primary"
                      theme="solid"
                      onClick={openTopUpLink}
                    >
                      {t('获取兑换码')}
                    </Button>
                  )}
                  <Button
                    type="warning"
                    theme="solid"
                    onClick={topUp}
                    disabled={isSubmitting}
                  >
                    {isSubmitting ? t('兑换中...') : t('兑换')}
                  </Button>
                </Space>
              </Form>
            </div>

            {/* 在线充值 */}
            <div style={{ marginTop: 20 }}>
              <Divider>{t('在线充值')}</Divider>
              {enableOnlineTopUp ? (
                <Form>
                  <Form.Input
                    disabled={!enableOnlineTopUp}
                    field="redemptionCount"
                    label={t('实付金额：') + ' ' + renderAmountText()}
                    placeholder={
                      t('充值数量，最低 ') +
                      renderQuotaWithAmount(minTopUp)
                    }
                    type="number"
                    value={topUpCount}
                    onChange={async (value) => {
                      if (value < 1) value = 1;
                      setTopUpCount(value);
                      await getAmount(value);
                    }}
                  />
                  <Space>
                    <Button
                      type="primary"
                      theme="solid"
                      onClick={() => preTopUp('zfb')}
                    >
                      {t('支付宝')}
                    </Button>
                    <Button
                      style={{
                        backgroundColor: 'rgba(var(--semi-green-5), 1)',
                      }}
                      type="primary"
                      theme="solid"
                      onClick={() => preTopUp('wx')}
                    >
                      {t('微信')}
                    </Button>
                  </Space>
                </Form>
              ) : (
                <Banner
                  fullMode={false}
                  type="info"
                  bordered
                  icon={null}
                  closeIcon={null}
                  title={
                    <div
                      style={{
                        fontWeight: 'bold',
                        fontSize: '14px',
                        lineHeight: '22px',
                      }}
                    >
                      {t('管理员已关闭在线充值')}
                    </div>
                  }
                />
              )}
            </div>
          </Card>
        </div>
      </Layout.Content>
    </Layout>
  );
};

export default TopUp;
