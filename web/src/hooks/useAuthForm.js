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
import { useState, useEffect, useContext } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { UserContext } from '../context/User';
import { API, showError, showInfo, showSuccess, updateAPI } from '../helpers';
import { setUserData } from '../helpers/data.js';
import { TELEGRAM_LOGIN_FIELDS, VALIDATION_MESSAGES } from '../utils/authConstants';

export const useAuthForm = (initialInputs = {}) => {
  const [inputs, setInputs] = useState(initialInputs);
  const [searchParams] = useSearchParams();
  const [userState, userDispatch] = useContext(UserContext);
  const [status, setStatus] = useState({});
  const [showWeChatLoginModal, setShowWeChatLoginModal] = useState(false);
  const navigate = useNavigate();

  // Handle input changes
  const handleChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  // Load status from localStorage and listen for updates
  useEffect(() => {
    const updateStatus = () => {
      let status = localStorage.getItem('status');
      if (status) {
        status = JSON.parse(status);
        setStatus(status);
      }
    };

    // Initial status load
    updateStatus();

    // Listen for storage changes to handle async status loading
    const handleStorageChange = (e) => {
      if (e.key === 'status') {
        updateStatus();
      }
    };

    // Listen for custom status update events
    const handleStatusUpdate = () => {
      updateStatus();
    };

    window.addEventListener('storage', handleStorageChange);
    window.addEventListener('statusUpdated', handleStatusUpdate);

    return () => {
      window.removeEventListener('storage', handleStorageChange);
      window.removeEventListener('statusUpdated', handleStatusUpdate);
    };
  }, []);

  // Handle AFF code processing
  const processAffCode = () => {
    let affCode = null;
    let statusFromStorage = localStorage.getItem('status');
    if (statusFromStorage) {
      statusFromStorage = JSON.parse(statusFromStorage);
      if (statusFromStorage.aff_enabled === true) {
        affCode = new URLSearchParams(window.location.search).get('aff');
        if (affCode) {
          localStorage.setItem('aff', affCode);
        }
      }
    }
    return affCode;
  };

  // Re-process AFF code when status changes
  useEffect(() => {
    processAffCode();
  }, [status]);

  // WeChat login handlers
  const onWeChatLoginClicked = () => {
    setShowWeChatLoginModal(true);
  };

  const onSubmitWeChatVerificationCode = async (turnstileToken) => {
    if (status.turnstile_check && turnstileToken === '') {
      showInfo(VALIDATION_MESSAGES.TURNSTILE_WAIT);
      return;
    }
    const res = await API.get(
      `/api/oauth/wechat?code=${inputs.wechat_verification_code}`,
    );
    const { success, message, data } = res.data;
    if (success) {
      userDispatch({ type: 'login', payload: data });
      localStorage.setItem('user', JSON.stringify(data));
      setUserData(data);
      updateAPI();
      navigate(searchParams.get('returnTo') || '/');
      showSuccess(VALIDATION_MESSAGES.LOGIN_SUCCESS);
      setShowWeChatLoginModal(false);
    } else {
      showError(message);
    }
  };

  // Telegram login handler
  const onTelegramLoginClicked = async (response) => {
    const params = {};
    TELEGRAM_LOGIN_FIELDS.forEach((field) => {
      if (response[field]) {
        params[field] = response[field];
      }
    });
    const res = await API.get(`/api/oauth/telegram/login`, { params });
    const { success, message, data } = res.data;
    if (success) {
      userDispatch({ type: 'login', payload: data });
      localStorage.setItem('user', JSON.stringify(data));
      showSuccess(VALIDATION_MESSAGES.LOGIN_SUCCESS);
      setUserData(data);
      updateAPI();
      navigate(searchParams.get('returnTo') || '/');
    } else {
      showError(message);
    }
  };

  return {
    inputs,
    setInputs,
    handleChange,
    status,
    setStatus,
    showWeChatLoginModal,
    setShowWeChatLoginModal,
    userDispatch,
    navigate,
    searchParams,
    processAffCode,
    onWeChatLoginClicked,
    onSubmitWeChatVerificationCode,
    onTelegramLoginClicked,
  };
};