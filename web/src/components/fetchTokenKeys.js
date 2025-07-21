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
// src/hooks/useTokenKeys.js
import { useEffect, useState } from 'react';
import { API, showError } from '../helpers';

async function fetchTokenKeys() {
  try {
    const response = await API.get('/api/token/?p=0&size=100');
    const { success, data } = response.data;
    if (success) {
      const activeTokens = data.filter((token) => token.status === 1);
      return activeTokens.map((token) => token.key);
    } else {
      throw new Error('Failed to fetch token keys');
    }
  } catch (error) {
    console.error('Error fetching token keys:', error);
    return [];
  }
}

function getServerAddress() {
  let status = localStorage.getItem('status');
  let serverAddress = '';

  if (status) {
    try {
      status = JSON.parse(status);
      serverAddress = status.server_address || '';
    } catch (error) {
      console.error('Failed to parse status from localStorage:', error);
    }
  }

  if (!serverAddress) {
    serverAddress = window.location.origin;
  }

  return serverAddress;
}

export function useTokenKeys(id) {
  const [keys, setKeys] = useState([]);
  // const [chatLink, setChatLink] = useState('');
  const [serverAddress, setServerAddress] = useState('');
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadAllData = async () => {
      const fetchedKeys = await fetchTokenKeys();
      if (fetchedKeys.length === 0) {
        showError('当前没有可用的启用令牌，请确认是否有令牌处于启用状态！');
        setTimeout(() => {
          window.location.href = '/app/tokens';
        }, 1500); // 延迟 1.5 秒后跳转
      }
      setKeys(fetchedKeys);
      setIsLoading(false);
      // setChatLink(link);

      const address = getServerAddress();
      setServerAddress(address);
    };

    loadAllData();
  }, []);

  return { keys, serverAddress, isLoading };
}
