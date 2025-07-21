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
import { Badge, Button, Tooltip } from '@douyinfe/semi-ui';
import { IconMail } from '@douyinfe/semi-icons';
import { useNavigate } from 'react-router-dom';
import { API } from '../helpers';
import { useTranslation } from 'react-i18next';

const InboxIcon = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);

  const loadUnreadCount = async () => {
    if (loading) return;
    
    setLoading(true);
    try {
      const res = await API.get('/api/user/messages/unread_count');
      const { success, data } = res.data;
      if (success) {
        setUnreadCount(data.unread_count || 0);
      }
    } catch (error) {
      // Silently fail - unread count is not critical
      console.error('Failed to load unread count:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleClick = () => {
    navigate('/app/inbox');
  };

  useEffect(() => {
    loadUnreadCount();
    
    // Poll for unread count every 30 seconds
    const interval = setInterval(loadUnreadCount, 30000);
    
    return () => clearInterval(interval);
  }, []);

  return (
    <Tooltip content={t('收件箱')} position="bottom">
      <div style={{ position: 'relative', display: 'inline-block' }}>
        <Button
          icon={<IconMail />}
          theme="borderless"
          type="tertiary"
          onClick={handleClick}
          style={{
            color: 'var(--semi-color-text-2)',
            fontSize: '16px',
            padding: '8px',
            minWidth: 'auto',
            height: 'auto'
          }}
        />
        {unreadCount > 0 && (
          <Badge
            count={unreadCount > 99 ? '99+' : unreadCount}
            style={{
              position: 'absolute',
              top: '-2px',
              right: '-2px',
              fontSize: '10px',
              minWidth: '16px',
              height: '16px',
              lineHeight: '16px',
              padding: '0 4px',
              borderRadius: '8px',
              backgroundColor: 'var(--semi-color-danger)',
              color: 'white',
              border: '1px solid var(--semi-color-bg-1)',
              boxShadow: '0 1px 2px rgba(0, 0, 0, 0.1)'
            }}
          />
        )}
      </div>
    </Tooltip>
  );
};

export default InboxIcon;