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
import { Link } from 'react-router-dom';
import { API, showError, showInfo, showSuccess } from '../helpers';
import { Button, Form } from '@douyinfe/semi-ui';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';
import { useTranslation } from 'react-i18next';
import { VALIDATION_MESSAGES } from '../utils/authConstants';

// Shared components and hooks
import { useAuthForm } from '../hooks/useAuthForm';
import { useTurnstile } from '../hooks/useTurnstile';
import AuthFormLayout from './shared/AuthFormLayout';
import ThirdPartyAuth from './shared/ThirdPartyAuth';
import WeChatLoginModal from './shared/WeChatLoginModal';
import TurnstileWrapper from './shared/TurnstileWrapper';

const RegisterForm = () => {
  const { t } = useTranslation();
  const [showEmailVerification, setShowEmailVerification] = useState(false);
  const [loading, setLoading] = useState(false);

  // Use shared hooks
  const {
    inputs,
    setInputs,
    handleChange,
    status,
    setStatus,
    showWeChatLoginModal,
    setShowWeChatLoginModal,
    navigate,
    processAffCode,
    onWeChatLoginClicked,
    onSubmitWeChatVerificationCode,
    onTelegramLoginClicked,
  } = useAuthForm({
    username: '',
    password: '',
    password2: '',
    email: '',
    verification_code: '',
    wechat_verification_code: '',
  });

  const {
    turnstileEnabled,
    turnstileSiteKey,
    turnstileToken,
    setTurnstileToken,
    validateTurnstile,
  } = useTurnstile();

  const { username, password, password2 } = inputs;

  // Update email verification when status changes
  useEffect(() => {
    setShowEmailVerification(status.email_verification);
  }, [status]);

  const handleSubmit = async (e) => {
    if (password.length < 8) {
      showInfo(VALIDATION_MESSAGES.PASSWORD_TOO_SHORT);
      return;
    }
    if (password !== password2) {
      showInfo(VALIDATION_MESSAGES.PASSWORD_MISMATCH);
      return;
    }
    if (username && password) {
      if (!validateTurnstile()) {
        showInfo(VALIDATION_MESSAGES.TURNSTILE_WAIT);
        return;
      }
      setLoading(true);
      
      // Handle AFF code if enabled
      let statusFromStorage = localStorage.getItem('status');
      if (statusFromStorage) {
        statusFromStorage = JSON.parse(statusFromStorage);
        if (statusFromStorage.aff_enabled === true) {
          let affCode = localStorage.getItem('aff');
          if (affCode) {
            setInputs(prev => ({ ...prev, aff_code: affCode }));
          }
        }
      }
      
      const res = await API.post(
        `/api/user/register?turnstile=${turnstileToken}`,
        inputs,
      );
      const { success, message } = res.data;
      if (success) {
        navigate('/login');
        showSuccess(VALIDATION_MESSAGES.REGISTER_SUCCESS);
      } else {
        showError(message);
      }
      setLoading(false);
    }
  };

  const sendVerificationCode = async () => {
    if (inputs.email === '') return;
    if (!validateTurnstile()) {
      showInfo(VALIDATION_MESSAGES.TURNSTILE_WAIT);
      return;
    }
    setLoading(true);
    const res = await API.get(
      `/api/verification?email=${inputs.email}&turnstile=${turnstileToken}`,
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess(VALIDATION_MESSAGES.VERIFICATION_CODE_SENT);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const handleWeChatSubmit = () => {
    onSubmitWeChatVerificationCode(turnstileToken);
  };

  return (
    <AuthFormLayout title={t('新用户注册')}>
      <Form size='large'>
        <Form.Input
          field={'username'}
          label={t('用户名')}
          placeholder={t('用户名')}
          name='username'
          onChange={(value) => handleChange('username', value)}
        />
        <Form.Input
          field={'password'}
          label={t('密码')}
          placeholder={t('输入密码，最短 8 位，最长 20 位')}
          name='password'
          type='password'
          onChange={(value) => handleChange('password', value)}
        />
        <Form.Input
          field={'password2'}
          label={t('确认密码')}
          placeholder={t('确认密码')}
          name='password2'
          type='password'
          onChange={(value) => handleChange('password2', value)}
        />
        {showEmailVerification && (
          <>
            <Form.Input
              field={'email'}
              label={t('邮箱')}
              placeholder={t('输入邮箱地址')}
              onChange={(value) => handleChange('email', value)}
              name='email'
              type='email'
              suffix={
                <Button
                  onClick={sendVerificationCode}
                  disabled={loading}
                >
                  {t('获取验证码')}
                </Button>
              }
            />
            <Form.Input
              field={'verification_code'}
              label={t('验证码')}
              placeholder={t('输入验证码')}
              onChange={(value) =>
                handleChange('verification_code', value)
              }
              name='verification_code'
            />
          </>
        )}
        <Button
          theme='solid'
          style={{ width: '100%' }}
          type={'primary'}
          size='large'
          htmlType={'submit'}
          onClick={handleSubmit}
        >
          {t('注册')}
        </Button>
      </Form>
      
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          marginTop: 20,
        }}
      >
        <Text>
          {t('已有账户？')}
          <Link to='/login'>{t('点击登录')}</Link>
        </Text>
      </div>

      <ThirdPartyAuth
        status={status}
        onWeChatLoginClicked={onWeChatLoginClicked}
        onTelegramLoginClicked={onTelegramLoginClicked}
      />

      <WeChatLoginModal
        visible={showWeChatLoginModal}
        onOk={handleWeChatSubmit}
        onCancel={() => setShowWeChatLoginModal(false)}
        status={status}
        inputs={inputs}
        handleChange={handleChange}
      />

      <TurnstileWrapper
        enabled={turnstileEnabled}
        siteKey={turnstileSiteKey}
        onVerify={setTurnstileToken}
      />
    </AuthFormLayout>
  );
};

export default RegisterForm;
