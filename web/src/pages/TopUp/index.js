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
  const [topUpCode, setTopUpCode] = useState('');
  const [topUpCount, setTopUpCount] = useState(0);
  const [minTopupCount, setMinTopUpCount] = useState(1);
  const [amount, setAmount] = useState(0.0);
  const [minTopUp, setMinTopUp] = useState(1);
  const [stripeTopUpCount, setStripeTopUpCount] = useState(0);
  const [stripeAmount, setStripeAmount] = useState(0.0);
  const [stripeMinTopUp, setStripeMinTopUp] = useState(1);
  const [topUpLink, setTopUpLink] = useState('');
  const [enableOnlineTopUp, setEnableOnlineTopUp] = useState(false);
  const [enableStripeTopUp, setEnableStripeTopUp] = useState(false);
  const [userQuota, setUserQuota] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [open, setOpen] = useState(false);
  const [stripeOpen, setStripeOpen] = useState(false);
  const [payWay, setPayWay] = useState('');
  const [isPaying, setIsPaying] = useState(false);

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

  const preTopUp = async (payment) => {
    if (!enableOnlineTopUp) {
      showError(t('管理员未开启在线充值！'));
      return;
    }
    await getAmount();
    if (topUpCount < minTopUp) {
      showError(t('充值数量不能小于') + minTopUp);
      return;
    }
    setPayWay(payment);
    setOpen(true);
  };

  const onlineTopUp = async () => {
    if (amount === 0) {
      await getAmount();
    }
    if (topUpCount < minTopUp) {
      showError('充值数量不能小于' + minTopUp);
      return;
    }
    setOpen(false);
    try {
      const res = await API.post('/api/user/pay', {
        amount: parseInt(topUpCount),
        top_up_code: topUpCode,
        payment_method: payWay,
      });
      if (res !== undefined) {
        const { message, data } = res.data;
        // showInfo(message);
        if (message === 'success') {
          let params = data;
          let url = res.data.url;
          let form = document.createElement('form');
          form.action = url;
          form.method = 'POST';
          // 判断是否为safari浏览器
          let isSafari =
            navigator.userAgent.indexOf('Safari') > -1 &&
            navigator.userAgent.indexOf('Chrome') < 1;
          if (!isSafari) {
            form.target = '_blank';
          }
          for (let key in params) {
            let input = document.createElement('input');
            input.type = 'hidden';
            input.name = key;
            input.value = params[key];
            form.appendChild(input);
          }
          document.body.appendChild(form);
          form.submit();
          document.body.removeChild(form);
        } else {
          showError(data);
          // setTopUpCount(parseInt(res.data.count));
          // setAmount(parseInt(data));
        }
      } else {
        showError(res);
      }
    } catch (err) {
      console.log(err);
    } finally {
    }
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
    const getStatus = async () => {
      const res = await API.get('/api/status');
      const { success, data } = res.data;
      if (success) {
        if (data.min_topup) {
          setMinTopUp(data.min_topup);
        }
        if (data.stripe_min_topup) {
          setStripeMinTopUp(data.stripe_min_topup)
        }
        if (data.enable_online_topup) {
          setEnableOnlineTopUp(data.enable_online_topup);
        }
        if (data.enable_stripe_topup) {
          setEnableStripeTopUp(data.enable_stripe_topup)
        }
      }
    }
    getStatus().then();
    getUserQuota().then();
  }, []);

  const renderAmount = () => {
    // console.log(amount);
    return amount + ' ' + t('元');
  };

  const getAmount = async (value) => {
    if (value === undefined) {
      value = topUpCount;
    }
    try {
      const res = await API.post('/api/user/amount', {
        amount: parseFloat(value),
        top_up_code: topUpCode,
      });
      if (res !== undefined) {
        const { message, data } = res.data;
        // showInfo(message);
        if (message === 'success') {
          setAmount(parseFloat(data));
        } else {
          setAmount(0);
          Toast.error({ content: '错误：' + data, id: 'getAmount' });
          // setTopUpCount(parseInt(res.data.count));
          // setAmount(parseInt(data));
        }
      } else {
        showError(res);
      }
    } catch (err) {
      console.log(err);
    } finally {
    }
  };

  const handleCancel = () => {
    setOpen(false);
  };
const stripePreTopUp = async () => {
  if (!enableStripeTopUp) {
    showError(t('管理员未开启在线充值！'));
    return;
  }
  await getStripeAmount();
  if (stripeTopUpCount < stripeMinTopUp) {
    showError(t('充值数量不能小于') + stripeMinTopUp);
    return;
  }
  setStripeOpen(true);
}

const onlineStripeTopUp = async () => {
  if (stripeAmount === 0) {
    await getStripeAmount();
  }
  if (stripeTopUpCount < stripeMinTopUp) {
    showError('充值数量不能小于' + stripeMinTopUp);
    return;
  }
  setStripeOpen(false);
  try {
    setIsPaying(true);
    const res = await API.post('/api/user/stripe/pay', {
      amount: parseInt(stripeTopUpCount),
      payment_method: "stripe",
    });
    if (res !== undefined) {
      const { message, data } = res.data;
      if (message === 'success') {
        processStripeCallback(data)
      } else {
        showError(data);
      }
    } else {
      showError(res);
    }
  } catch (err) {
    console.log(err);
  } finally {
    setIsPaying(false);
  }
}

const processStripeCallback = (data) => {
  location.href = data.pay_link;
};

const renderStripeAmount = () => {
  return stripeAmount + '元';
};

const getStripeAmount = async (value) => {
  if (value === undefined) {
    value = stripeTopUpCount
  }
  try {
    const res = await API.post('/api/user/stripe/amount', {
      amount: parseFloat(value),
    });
    if (res !== undefined) {
      const { message, data } = res.data;
      // showInfo(message);
      if (message === 'success') {
        setStripeAmount(parseFloat(data))
      } else {
        setStripeAmount( 0)
        Toast.error({ content: '错误：' + data, id: 'getAmount' });
      }
    } else {
      showError(res);
    }
  } catch (err) {
    console.log(err);
  } finally {
  }
}

const handleStripeCancel = () => {
  setStripeOpen(false);
};

  return (
    <div>
      <Layout>
        <Layout.Header>
          <h3>{t('我的钱包')}</h3>
        </Layout.Header>
        <Layout.Content>
          <Modal
            title={t('确定要充值吗')}
            visible={open}
            onOk={onlineTopUp}
            onCancel={handleCancel}
            maskClosable={false}
            size={'small'}
            centered={true}
          >
            <p>
              {t('充值数量')}：{topUpCount}
            </p>
            <p>
              {t('实付金额')}：{renderAmount()}
            </p>
            <p>{t('是否确认充值？')}</p>
          </Modal>
          <Modal
              title={t('确定要充值吗')}
              visible={stripeOpen}
              onOk={onlineStripeTopUp}
              onCancel={handleStripeCancel}
              maskClosable={false}
              size={'small'}
              centered={true}
          >
            <p>
              {t('充值数量')}：{stripeTopUpCount}
            </p>
            <p>
              {t('实付金额')}：{renderStripeAmount()}
            </p>
            <p>{t('是否确认充值？')}</p>
          </Modal>
          <div
            style={{ marginTop: 20, display: 'flex', justifyContent: 'center' }}
          >
            <Card
              style={{
                width: '100%',
                maxWidth: '800px',
                padding: '20px',
              }}
            >
              <Typography.Title heading={3}>
                {t('充值额度')}
              </Typography.Title>
              <div style={{ display: 'flex', justifyContent: 'center' }}>
                <Text>{t('当前剩余额度：') + renderQuota(userQuota)}</Text>
              </div>
              <Divider />
              <div style={{ display: 'flex', justifyContent: 'center' }}>
                <Typography.Title heading={5}>
                  {t('使用兑换码')}
                </Typography.Title>
              </div>
              <Form>
                <Form.Input
                  field={'redemptionCode'}
                  label={t('兑换码')}
                  placeholder={t('请输入兑换码')}
                  name='redemptionCode'
                  value={redemptionCode}
                  onChange={(value) => setRedemptionCode(value)}
                />
                <Space>
                  <Button
                    style={{backgroundColor: '#b161fe'}}
                    type={'primary'}
                    theme={'solid'}
                    onClick={async () => {
                      await topUp();
                    }}
                  >
                    {t('兑换')}
                  </Button>
                </Space>
              </Form>
              {enableOnlineTopUp ? (
                <div>
                  <Divider />
                  <div style={{ display: 'flex', justifyContent: 'center' }}>
                    <Typography.Title heading={5}>
                      {t('在线充值')}
                    </Typography.Title>
                  </div>
                  <Form>
                    <Form.Input
                      disabled={!enableOnlineTopUp}
                      field={'redemptionCount'}
                      label={t('实付金额：') + ' ' + renderAmount()}
                      placeholder={t('充值数量，最低 ') + minTopUp + '元'}
                      name='redemptionCount'
                      type={'number'}
                      value={topUpCount}
                      suffix={t('元')}
                      min={minTopUp}
                      defaultValue={minTopUp}
                      max={100000}
                      onChange={async (value) => {
                        if (value < 1) {
                          value = 1;
                        }
                        if (value > 100000) {
                          value = 100000;
                        }
                        setTopUpCount(value);
                        await getAmount(value);
                      }}
                    />
                    <Space>
                      <Button
                        style={{backgroundColor: '#b161fe'}}
                        type={'primary'}
                        disabled={isPaying}
                        theme={'solid'}
                        onClick={async () => {
                          preTopUp('zfb');
                        }}
                      >
                        {isPaying ? t('支付中...') : t('去支付')}
                      </Button>
                    </Space>
                  </Form>
                </div>
              ) : (
                <></>
              )}
                {enableStripeTopUp ? (
                    <div>
                      <Form>
                        <Form.Input
                            disabled={!enableStripeTopUp}
                            field={'redemptionCount'}
                            label={t('实付金额：') + ' ' + renderStripeAmount()}
                            placeholder={t('充值数量，最低 ') + stripeMinTopUp + '$'}
                            name='redemptionCount'
                            type={'number'}
                            value={stripeTopUpCount}
                            suffix={'$'}
                            min={stripeMinTopUp}
                            defaultValue={stripeMinTopUp}
                            max={100000}
                            onChange={async (value) => {
                              if (value < 1) {
                                value = 1;
                              }
                              if (value > 100000) {
                                value = 100000;
                              }
                              setStripeTopUpCount(value);
                              await getStripeAmount( value);
                            }}
                        />
                        <Space>
                          <Button
                              style={{backgroundColor: '#b161fe'}}
                              type={'primary'}
                              disabled={isPaying}
                              theme={'solid'}
                              onClick={async () => {
                                stripePreTopUp();
                              }}
                          >
                            {isPaying ? '支付中...' : '去支付'}
                          </Button>
                        </Space>
                      </Form>
                    </div>
                ) : (
                    <></>
                )}
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
