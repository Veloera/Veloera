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
  Layout, 
  List, 
  Typography, 
  Tag, 
  Empty, 
  Spin, 
  Pagination,
  Space,
  Card
} from '@douyinfe/semi-ui';
import { useNavigate, useParams } from 'react-router-dom';
import { API, showError } from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';
import { useTranslation } from 'react-i18next';
import MessageDetail from './MessageDetail';

const { Title, Text } = Typography;

const Inbox = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams();
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [currentPage, setCurrentPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [selectedMessage, setSelectedMessage] = useState(null);

  const loadMessages = async (page = 1) => {
    setLoading(true);
    try {
      const res = await API.get(`/api/user/messages?p=${page}&page_size=${ITEMS_PER_PAGE}`);
      const { success, data } = res.data;
      if (success) {
        setMessages(data.items || []);
        setTotal(data.total || 0);
        setCurrentPage(page);
      } else {
        showError(t('加载消息失败'));
      }
    } catch (error) {
      showError(t('加载消息失败'));
    } finally {
      setLoading(false);
    }
  };

  const markAsRead = async (messageId) => {
    try {
      await API.put(`/api/user/messages/${messageId}/read`);
      // Update the message in the list to mark it as read
      setMessages(prevMessages => 
        prevMessages.map(msg => 
          msg.id === messageId 
            ? { ...msg, is_read: true, read_at: new Date().toISOString() }
            : msg
        )
      );
    } catch (error) {
      // Silently fail - read status is not critical
      console.error('Failed to mark message as read:', error);
    }
  };

  const handleMessageClick = (message) => {
    if (!message.is_read) {
      markAsRead(message.id);
    }
    navigate(`/app/inbox/${message.id}`);
  };

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInHours = (now - date) / (1000 * 60 * 60);
    
    if (diffInHours < 24) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } else if (diffInHours < 24 * 7) {
      return date.toLocaleDateString([], { weekday: 'short', hour: '2-digit', minute: '2-digit' });
    } else {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
    }
  };

  const getContentPreview = (content, format) => {
    if (!content) return '';
    
    // Remove HTML tags and markdown formatting for preview
    let preview = content
      .replace(/<[^>]*>/g, '') // Remove HTML tags
      .replace(/[#*_`~]/g, '') // Remove markdown formatting
      .replace(/\n/g, ' ') // Replace newlines with spaces
      .trim();
    
    // Limit preview length
    if (preview.length > 100) {
      preview = preview.substring(0, 97) + '...';
    }
    
    return preview;
  };

  useEffect(() => {
    if (id) {
      // If there's an ID in the URL, we need to find the message in our list
      const message = messages.find(msg => msg.id === parseInt(id));
      if (message) {
        setSelectedMessage(message);
        if (!message.is_read) {
          markAsRead(message.id);
        }
      } else if (messages.length > 0) {
        // Messages are loaded but message not found
        showError(t('消息不存在'));
        navigate('/app/inbox');
      }
      // If messages.length === 0, we'll wait for them to load
    }
  }, [id, messages, t, navigate]);

  useEffect(() => {
    // Always load messages when component mounts
    loadMessages();
  }, []);

  // If viewing a specific message, show the detail view
  if (id && selectedMessage) {
    return <MessageDetail message={selectedMessage} onBack={() => navigate('/app/inbox')} />;
  }

  return (
    <Layout>
      <Layout.Header>
        <h3>{t('收件箱')}</h3>
      </Layout.Header>
      <Layout.Content style={{ padding: '0 24px', display: 'flex', justifyContent: 'flex-start' }}>
        <div style={{ width: '90%' }}>
          {loading ? (
            <div style={{ textAlign: 'center', padding: '50px' }}>
              <Spin size="large" />
            </div>
          ) : messages.length === 0 ? (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              title={t('暂无消息')}
              description={t('您还没有收到任何消息')}
              style={{ padding: '50px', width: '100%' }}
            />
          ) : (
            <Space vertical style={{ width: '100%' }} spacing="medium">
              <List
                style={{ width: '100%', padding: '24px 0px' }}
                dataSource={messages}
                renderItem={(message) => (
                  <List.Item
                    style={{
                      cursor: 'pointer',
                      padding: '16px',
                      borderRadius: '8px',
                      marginBottom: '8px',
                      backgroundColor: message.is_read 
                        ? 'var(--semi-color-bg-1)' 
                        : 'var(--semi-color-primary-light-default)',
                      border: '1px solid var(--semi-color-border)',
                      transition: 'all 0.2s ease',
                      width: '100%',
                    }}
                    onClick={() => handleMessageClick(message)}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.backgroundColor = message.is_read 
                        ? 'var(--semi-color-fill-0)' 
                        : 'var(--semi-color-primary-light-hover)';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.backgroundColor = message.is_read 
                        ? 'var(--semi-color-bg-1)' 
                        : 'var(--semi-color-primary-light-default)';
                    }}
                  >
                  <div style={{ width: '100%' }}>
                    <div style={{ 
                      display: 'flex', 
                      justifyContent: 'space-between', 
                      alignItems: 'flex-start',
                      marginBottom: '8px'
                    }}>
                      <h3
                        style={{ 
                          margin: 0, 
                          fontWeight: 'bold',
                          color: message.is_read ? 'var(--semi-color-text-1)' : 'var(--semi-color-primary)'
                          
                        }}
                      >
                        {message.title}
                      </h3>
                      <Space>
                        {!message.is_read && (
                          <Tag color="red" size="small">{t('未读')}</Tag>
                        )}
                        <Text 
                          type="tertiary" 
                          size="small"
                          style={{ whiteSpace: 'nowrap' }}
                        >
                          {formatTimestamp(message.created_at)}
                        </Text>
                      </Space>
                    </div>
                    <Text 
                      type="secondary" 
                      size="small"
                      style={{ 
                        display: 'block',
                        lineHeight: '1.4',
                        color: message.is_read ? 'var(--semi-color-text-2)' : 'var(--semi-color-text-1)'
                      }}
                    >
                      {getContentPreview(message.content, message.format)}
                    </Text>
                  </div>
                </List.Item>
              )}
            />
            
            {total > ITEMS_PER_PAGE && (
              <div style={{ textAlign: 'center', marginTop: '24px' }}>
                <Pagination
                  current={currentPage}
                  total={total}
                  pageSize={ITEMS_PER_PAGE}
                  onChange={(page) => loadMessages(page)}
                  showSizeChanger={false}
                  showQuickJumper
                  showTotal={(total, range) => 
                    t('第 {{start}}-{{end}} 条，共 {{total}} 条', {
                      start: range[0],
                      end: range[1],
                      total: total
                    })
                  }
                />
              </div>
              )}
            </Space>
          )}
        </div>
      </Layout.Content>
    </Layout>
  );
};

export default Inbox;