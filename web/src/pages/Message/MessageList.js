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
import { API, showError, showSuccess } from '../../helpers';
import {
  Button,
  Form,
  Popconfirm,
  Space,
  Table,
  Tag,
  Tooltip,
} from '@douyinfe/semi-ui';
import { ITEMS_PER_PAGE } from '../../constants';
import CreateMessage from './CreateMessage';
import EditMessage from './EditMessage';
import { useTranslation } from 'react-i18next';

const MessageList = () => {
  const { t } = useTranslation();

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: t('标题'),
      dataIndex: 'title',
      render: (text) => (
        <div style={{ maxWidth: 200, wordBreak: 'break-word' }}>
          {text}
        </div>
      ),
    },
    {
      title: t('格式'),
      dataIndex: 'format',
      width: 100,
      render: (text) => (
        <Tag color={text === 'html' ? 'blue' : 'green'} size='large'>
          {text?.toUpperCase()}
        </Tag>
      ),
    },
    {
      title: t('收件人数'),
      dataIndex: 'stats',
      width: 100,
      render: (stats) => (
        <Tag color='white' size='large'>
          {stats?.total_recipients || 0}
        </Tag>
      ),
    },
    {
      title: t('创建时间'),
      dataIndex: 'created_at',
      width: 180,
      render: (text) => {
        return new Date(text).toLocaleString();
      },
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      width: 200,
      render: (_, record) => (
        <div>
          <Button
            theme='light'
            type='tertiary'
            size='small'
            style={{ marginRight: 8 }}
            onClick={() => {
              setEditingMessage(record);
              setShowEditMessage(true);
            }}
          >
            {t('编辑')}
          </Button>
          <Popconfirm
            title={t('确定删除此消息？')}
            content={t('此操作将删除消息及所有用户的接收记录，不可恢复')}
            okType={'danger'}
            position={'left'}
            onConfirm={() => deleteMessage(record.id)}
          >
            <Button theme='light' type='danger' size='small'>
              {t('删除')}
            </Button>
          </Popconfirm>
        </div>
      ),
    },
  ];

  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [dateRange, setDateRange] = useState([]);
  const [messageCount, setMessageCount] = useState(ITEMS_PER_PAGE);
  const [showCreateMessage, setShowCreateMessage] = useState(false);
  const [showEditMessage, setShowEditMessage] = useState(false);
  const [editingMessage, setEditingMessage] = useState({
    id: undefined,
  });

  const setMessageFormat = (messages) => {
    for (let i = 0; i < messages.length; i++) {
      messages[i].key = messages[i].id;
    }
    setMessages(messages);
  };

  const loadMessages = async (startIdx, pageSize) => {
    const res = await API.get(`/api/admin/messages?p=${startIdx}&page_size=${pageSize}`);
    const { success, message, data } = res.data;
    if (success) {
      const newPageData = data.items || [];
      setActivePage(data.page || 1);
      setMessageCount(data.total || 0);
      setMessageFormat(newPageData);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    loadMessages(0, pageSize)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  const deleteMessage = async (messageId) => {
    const res = await API.delete(`/api/admin/messages/${messageId}`);
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('消息删除成功'));
      // Remove the deleted message from the list
      setMessages(messages.filter(msg => msg.id !== messageId));
      setMessageCount(messageCount - 1);
    } else {
      showError(message);
    }
  };

  const searchMessages = async (
    startIdx,
    pageSize,
    searchKeyword,
    dateRange,
  ) => {
    if (searchKeyword === '' && dateRange.length === 0) {
      // if keyword is blank and no date range, load messages instead.
      await loadMessages(startIdx, pageSize);
      return;
    }
    setSearching(true);
    
    let params = new URLSearchParams({
      p: startIdx,
      page_size: pageSize,
    });
    
    if (searchKeyword) {
      params.append('keyword', searchKeyword);
    }
    
    if (dateRange.length === 2) {
      params.append('start_date', dateRange[0].toISOString());
      params.append('end_date', dateRange[1].toISOString());
    }
    
    const res = await API.get(`/api/admin/messages/search?${params.toString()}`);
    const { success, message, data } = res.data;
    if (success) {
      const newPageData = data.items || [];
      setActivePage(data.page || 1);
      setMessageCount(data.total || 0);
      setMessageFormat(newPageData);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (value) => {
    setSearchKeyword(value.trim());
  };

  const handlePageChange = (page) => {
    setActivePage(page);
    if (searchKeyword === '' && dateRange.length === 0) {
      loadMessages(page, pageSize).then();
    } else {
      searchMessages(page, pageSize, searchKeyword, dateRange).then();
    }
  };

  const closeCreateMessage = () => {
    setShowCreateMessage(false);
  };

  const closeEditMessage = () => {
    setShowEditMessage(false);
    setEditingMessage({
      id: undefined,
    });
  };

  const refresh = async () => {
    setActivePage(1);
    if (searchKeyword === '' && dateRange.length === 0) {
      await loadMessages(1, pageSize);
    } else {
      await searchMessages(1, pageSize, searchKeyword, dateRange);
    }
  };

  const handlePageSizeChange = async (size) => {
    localStorage.setItem('message-page-size', size + '');
    setPageSize(size);
    setActivePage(1);
    loadMessages(1, size)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  };

  return (
    <>
      <CreateMessage
        refresh={refresh}
        visible={showCreateMessage}
        handleClose={closeCreateMessage}
      />
      <EditMessage
        refresh={refresh}
        visible={showEditMessage}
        handleClose={closeEditMessage}
        editingMessage={editingMessage}
      />
      <Form
        onSubmit={() => {
          searchMessages(1, pageSize, searchKeyword, dateRange);
        }}
        labelPosition='left'
      >
        <div style={{ display: 'flex', marginBottom: 16 }}>
          <Space>
            <Tooltip content={t('支持搜索消息标题和内容')}>
              <Form.Input
                label={t('搜索关键字')}
                icon='search'
                field='keyword'
                iconPosition='left'
                placeholder={t('搜索消息标题或内容')}
                value={searchKeyword}
                loading={searching}
                onChange={(value) => handleKeywordChange(value)}
                style={{ width: 200 }}
              />
            </Tooltip>

            <Form.DatePicker
              field='dateRange'
              label={t('创建时间')}
              type='dateRange'
              placeholder={[t('开始日期'), t('结束日期')]}
              onChange={(value) => {
                setDateRange(value || []);
              }}
              style={{ width: 300 }}
            />
            
            <Button
              label={t('查询')}
              type='primary'
              htmlType='submit'
              className='btn-margin-right'
            >
              {t('查询')}
            </Button>
            <Button
              theme='light'
              type='primary'
              onClick={() => {
                setShowCreateMessage(true);
              }}
            >
              {t('创建消息')}
            </Button>
          </Space>
        </div>
      </Form>

      <Table
        columns={columns}
        dataSource={messages}
        pagination={{
          formatPageText: (page) =>
            t('第 {{start}} - {{end}} 条，共 {{total}} 条', {
              start: page.currentStart,
              end: page.currentEnd,
              total: messageCount,
            }),
          currentPage: activePage,
          pageSize: pageSize,
          total: messageCount,
          pageSizeOpts: [10, 20, 50, 100],
          showSizeChanger: true,
          onPageSizeChange: (size) => {
            handlePageSizeChange(size);
          },
          onPageChange: handlePageChange,
        }}
        loading={loading}
      />
    </>
  );
};

export default MessageList;