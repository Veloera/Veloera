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
import { API, showError } from '../helpers';

export async function getOAuthState() {
  let path = '/api/oauth/state';
  
  // Check AFF toggle status before processing AFF code
  let statusFromStorage = localStorage.getItem('status');
  let affEnabled = false;
  if (statusFromStorage) {
    statusFromStorage = JSON.parse(statusFromStorage);
    affEnabled = statusFromStorage.aff_enabled === true;
  }
  
  // Only append AFF code to OAuth state URL when aff_enabled is true
  if (affEnabled) {
    let affCode = localStorage.getItem('aff');
    if (affCode && affCode.length > 0) {
      path += `?aff=${affCode}`;
    }
  }
  
  const res = await API.get(path);
  const { success, message, data } = res.data;
  if (success) {
    return data;
  } else {
    showError(message);
    return '';
  }
}

export async function onOIDCClicked(auth_url, client_id, openInNewTab = false) {
  const state = await getOAuthState();
  if (!state) return;
  const redirect_uri = `${window.location.origin}/oauth/oidc`;
  const response_type = 'code';
  const scope = 'openid profile email';
  const url = `${auth_url}?client_id=${client_id}&redirect_uri=${redirect_uri}&response_type=${response_type}&scope=${scope}&state=${state}`;
  if (openInNewTab) {
    window.open(url);
  } else {
    window.location.href = url;
  }
}

export async function onGitHubOAuthClicked(github_client_id, openInNewTab = false) {
  const state = await getOAuthState();
  if (!state) return;
  const url = `https://github.com/login/oauth/authorize?client_id=${github_client_id}&state=${state}&scope=user:email`;
  if (openInNewTab) {
    window.open(url);
  } else {
    window.location.href = url;
  }
}

export async function onLinuxDOOAuthClicked(linuxdo_client_id, openInNewTab = false) {
  const state = await getOAuthState();
  if (!state) return;
  const url = `https://connect.linux.do/oauth2/authorize?response_type=code&client_id=${linuxdo_client_id}&state=${state}`;
  if (openInNewTab) {
    window.open(url);
  } else {
    window.location.href = url;
  }
}

let channelModels = undefined;
export async function loadChannelModels() {
  const res = await API.get('/api/models');
  const { success, data } = res.data;
  if (!success) {
    return;
  }
  channelModels = data;
  localStorage.setItem('channel_models', JSON.stringify(data));
}

export function getChannelModels(type) {
  if (channelModels !== undefined && type in channelModels) {
    if (!channelModels[type]) {
      return [];
    }
    return channelModels[type];
  }
  let models = localStorage.getItem('channel_models');
  if (!models) {
    return [];
  }
  channelModels = JSON.parse(models);
  if (type in channelModels) {
    return channelModels[type];
  }
  return [];
}
