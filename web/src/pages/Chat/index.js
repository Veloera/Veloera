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
import React, { useEffect } from 'react';
import { useTokenKeys } from '../../components/fetchTokenKeys';
import { Banner, Layout } from '@douyinfe/semi-ui';
import { useParams } from 'react-router-dom';

const ChatPage = () => {
  const { id } = useParams();
  const { keys, serverAddress, isLoading } = useTokenKeys(id);

  const comLink = (key) => {
    // console.log('chatLink:', chatLink);
    if (!serverAddress || !key) return '';
    let link = '';
    if (id) {
      let chats = localStorage.getItem('chats');
      if (chats) {
        chats = JSON.parse(chats);
        if (Array.isArray(chats) && chats.length > 0) {
          for (let k in chats[id]) {
            link = chats[id][k];
            link = link.replaceAll(
              '{address}',
              encodeURIComponent(serverAddress),
            );
            link = link.replaceAll('{key}', 'sk-' + key);
          }
        }
      }
    }
    return link;
  };

  const iframeSrc = keys.length > 0 ? comLink(keys[0]) : '';

  return !isLoading && iframeSrc ? (
    <iframe
      src={iframeSrc}
      style={{ width: '100%', height: '100%', border: 'none' }}
      title='Token Frame'
      allow='camera;microphone'
    />
  ) : (
    <div>
      <Layout>
        <Layout.Header>
          <Banner description={'正在跳转......'} type={'warning'} />
        </Layout.Header>
      </Layout>
    </div>
  );
};

export default ChatPage;
