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
  copy,
  showError,
  showSuccess,
  timestamp2string,
} from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { renderQuota } from '../helpers/render';
import {
  Button,
  Divider,
  Form,
  Modal,
  Popconfirm,
  Popover,
  Table,
  Tag,
} from '@douyinfe/semi-ui';
import EditRedemption from '../pages/Redemption/EditRedemption';
import { useTranslation } from 'react-i18next';

function renderTimestamp(timestamp) {
  return <>{timestamp2string(timestamp)}</>;
}

// 批量删除兑换码模态框组件
const BatchDeleteByNameModal = ({ visible, onClose, onConfirm }) => {
  const { t } = useTranslation();
  const [name, setName] = useState('');
  const [count, setCount] = useState(0);
  const [confirming, setConfirming] = useState(false);
  const [loading, setLoading] = useState(false);
  
  // 重置状态
  const resetState = () => {
    setName('');
    setCount(0);
    setConfirming(false);
    setLoading(false);
  };
  
  // 关闭模态框
  const handleClose = () => {
    resetState();
    onClose();
  };

  // 当模态框显示状态变化时重置状态
  useEffect(() => {
    if (!visible) {
      resetState();
    }
  }, [visible]);

  // 根据名称查询兑换码数量
  const checkRedemptionCount = async () => {
    if (!name.trim()) return;
    
    setLoading(true);
    try {
      const res = await API.get(`/api/redemption/count-by-name?name=${encodeURIComponent(name.trim())}`);
      if (res.data.success) {
        setCount(res.data.data);
        setConfirming(true);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  // 确认删除
  const handleConfirm = async () => {
    if (count === 0) return;
    
    setLoading(true);
    try {
      const res = await API.delete(`/api/redemption/delete-by-name?name=${encodeURIComponent(name.trim())}`);
      if (res.data.success) {
        showSuccess(t('已成功删除 {{count}} 个兑换码', { count }));
        onConfirm();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      handleClose();
    }
  };

  return (
    <Modal
      title={t('按名称批量删除兑换码')}
      visible={visible}
      onCancel={handleClose}
      footer={
        confirming ? (
          <>
            <Button onClick={handleClose}>{t('取消')}</Button>
            <Button type="danger" loading={loading} onClick={handleConfirm}>
              {t('确认删除 {{count}} 个兑换码', { count })}
            </Button>
          </>
        ) : (
          <>
            <Button onClick={handleClose}>{t('取消')}</Button>
            <Button type="primary" loading={loading} onClick={checkRedemptionCount}>
              {t('查询数量')}
            </Button>
          </>
        )
      }
    >
      {confirming ? (
        <div>
          <p>{t('将删除名称为 "{{name}}" 的所有兑换码，共 {{count}} 个', { name, count })}</p>
          <p style={{ color: 'red' }}>{t('此操作不可逆，请确认')}</p>
        </div>
      ) : (
        <Form>
          <Form.Input
            field="name"
            label={t('兑换码名称')}
            placeholder={t('请输入要删除的兑换码名称')}
            value={name}
            onChange={setName}
          />
        </Form>
      )}
    </Modal>
  );
};

const RedemptionsTable = () => {
  const { t } = useTranslation();

  const renderStatus = (status) => {
    switch (status) {
      case 1:
        return (
          <Tag color='green' size='large'>
            {t('未使用')}
          </Tag>
        );
      case 2:
        return (
          <Tag color='red' size='large'>
            {t('已禁用')}
          </Tag>
        );
      case 3:
        return (
          <Tag color='grey' size='large'>
            {t('已使用')}
          </Tag>
        );
      default:
        return (
          <Tag color='black' size='large'>
            {t('未知状态')}
          </Tag>
        );
    }
  };

  const renderType = (isGift) => {
    return isGift ? 
      <Tag color="purple">{t('礼品码')}</Tag> : 
      <Tag color="blue">{t('兑换码')}</Tag>;
  };

  const renderMaxUses = (maxUses) => {
    return maxUses === -1 ? t('无限制') : maxUses;
  };

  const columns = [
    {
      title: t('ID'),
      dataIndex: 'id',
    },
    {
      title: t('名称'),
      dataIndex: 'name',
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      key: 'status',
      render: (text, record, index) => {
        return <div>{renderStatus(text)}</div>;
      },
    },
    {
      title: t('额度'),
      dataIndex: 'quota',
      render: (text, record, index) => {
        return <div>{renderQuota(parseInt(text))}</div>;
      },
    },
    {
      title: t('创建时间'),
      dataIndex: 'created_time',
      render: (text, record, index) => {
        return <div>{renderTimestamp(text)}</div>;
      },
    },
    {
      title: t('兑换人ID'),
      dataIndex: 'used_user_id',
      render: (text, record, index) => {
        return <div>{text === 0 ? t('无') : text}</div>;
      },
    },
    {
      title: t('类型'),
      dataIndex: 'is_gift',
      render: (text, record) => renderType(record.is_gift),
    },
    {
      title: t('最大使用次数'),
      dataIndex: 'max_uses',
      render: (text, record) => record.is_gift ? renderMaxUses(record.max_uses) : '-',
    },
    {
      title: t('已使用次数'),
      dataIndex: 'used_count',
      render: (text, record) => record.is_gift ? text : '-',
    },
    {
      title: '',
      dataIndex: 'operate',
      render: (text, record, index) => (
        <div>
          <Popover content={record.key} style={{ padding: 20 }} position='top'>
            <Button theme='light' type='tertiary' style={{ marginRight: 1 }}>
              {t('查看')}
            </Button>
          </Popover>
          <Button
            theme='light'
            type='secondary'
            style={{ marginRight: 1 }}
            onClick={async (text) => {
              await copyText(record.key);
            }}
          >
            {t('复制')}
          </Button>
          <Popconfirm
            title={t('确定是否要删除此兑换码？')}
            content={t('此修改将不可逆')}
            okType={'danger'}
            position={'left'}
            onConfirm={() => {
              manageRedemption(record.id, 'delete', record).then(() => {
                removeRecord(record.key);
              });
            }}
          >
            <Button theme='light' type='danger' style={{ marginRight: 1 }}>
              {t('删除')}
            </Button>
          </Popconfirm>
          {record.status === 1 ? (
            <Button
              theme='light'
              type='warning'
              style={{ marginRight: 1 }}
              onClick={async () => {
                manageRedemption(record.id, 'disable', record);
              }}
            >
              {t('禁用')}
            </Button>
          ) : (
            <Button
              theme='light'
              type='secondary'
              style={{ marginRight: 1 }}
              onClick={async () => {
                manageRedemption(record.id, 'enable', record);
              }}
              disabled={record.status === 3}
            >
              {t('启用')}
            </Button>
          )}
          <Button
            theme='light'
            type='tertiary'
            style={{ marginRight: 1 }}
            onClick={() => {
              setEditingRedemption(record);
              setShowEdit(true);
            }}
            disabled={record.status !== 1}
          >
            {t('编辑')}
          </Button>
        </div>
      ),
    },
  ];

  const [redemptions, setRedemptions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [tokenCount, setTokenCount] = useState(ITEMS_PER_PAGE);
  const [selectedKeys, setSelectedKeys] = useState([]);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [editingRedemption, setEditingRedemption] = useState({
    id: undefined,
  });
  const [showEdit, setShowEdit] = useState(false);
  const [showBatchDeleteByName, setShowBatchDeleteByName] = useState(false);
  
  const closeEdit = () => {
    setShowEdit(false);
  };
  
  // 批量禁用所选兑换码
  const disableSelectedRedemptions = async () => {
    if (selectedKeys.length === 0) {
      showError(t('请至少选择一个兑换码！'));
      return;
    }
    
    setLoading(true);
    try {
      // 获取所有选中的 ID
      const ids = selectedKeys.map(item => item.id);
      const res = await API.put('/api/redemption/batch-disable', { ids });
      
      if (res.data.success) {
        showSuccess(t('成功禁用 {{count}} 个兑换码！', { count: res.data.data }));
        await refresh();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };
  
  // 删除所有已禁用的兑换码
  const deleteAllDisabledRedemptions = async () => {
    Modal.confirm({
      title: t('删除所有已禁用的兑换码'),
      content: t('此操作将删除所有已禁用/已使用的兑换码，不可恢复，是否继续？'),
      okType: 'danger',
      onOk: async () => {
        setLoading(true);
        try {
          const res = await API.delete('/api/redemption/delete-disabled');
          
          if (res.data.success) {
            showSuccess(t('成功删除 {{count}} 个已禁用的兑换码！', { count: res.data.data }));
            await refresh();
          } else {
            showError(res.data.message);
          }
        } catch (error) {
          showError(error.message);
        } finally {
          setLoading(false);
        }
      }
    });
  };

  const setRedemptionFormat = (redeptions) => {
    setRedemptions(redeptions);
  };

  const loadRedemptions = async (startIdx, pageSize) => {
    const res = await API.get(
      `/api/redemption/?p=${startIdx}&page_size=${pageSize}`,
    );
    const { success, message, data } = res.data;
    if (success) {
      const newPageData = data.items;
      setActivePage(data.page);
      setTokenCount(data.total);
      setRedemptionFormat(newPageData);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const removeRecord = (key) => {
    let newDataSource = [...redemptions];
    if (key != null) {
      let idx = newDataSource.findIndex((data) => data.key === key);

      if (idx > -1) {
        newDataSource.splice(idx, 1);
        setRedemptions(newDataSource);
      }
    }
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess(t('已复制到剪贴板！'));
    } else {
      // setSearchKeyword(text);
      Modal.error({ title: t('无法复制到剪贴板，请手动复制'), content: text });
    }
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(redemptions.length / pageSize) + 1) {
        await loadRedemptions(activePage - 1, pageSize);
      }
      setActivePage(activePage);
    })();
  };

  useEffect(() => {
    loadRedemptions(0, pageSize)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  const refresh = async () => {
    await loadRedemptions(activePage - 1, pageSize);
  };

  const manageRedemption = async (id, action, record) => {
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/redemption/${id}/`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/redemption/?status_only=true', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/redemption/?status_only=true', data);
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('操作成功完成！'));
      let redemption = res.data.data;
      let newRedemptions = [...redemptions];
      // let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      if (action === 'delete') {
      } else {
        record.status = redemption.status;
      }
      setRedemptions(newRedemptions);
    } else {
      showError(message);
    }
  };

  const searchRedemptions = async (keyword, page, pageSize) => {
    if (searchKeyword === '') {
      await loadRedemptions(page, pageSize);
      return;
    }
    setSearching(true);
    const res = await API.get(
      `/api/redemption/search?keyword=${keyword}&p=${page}&page_size=${pageSize}`,
    );
    const { success, message, data } = res.data;
    if (success) {
      const newPageData = data.items;
      setActivePage(data.page);
      setTokenCount(data.total);
      setRedemptionFormat(newPageData);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (value) => {
    setSearchKeyword(value.trim());
  };

  const sortRedemption = (key) => {
    if (redemptions.length === 0) return;
    setLoading(true);
    let sortedRedemptions = [...redemptions];
    sortedRedemptions.sort((a, b) => {
      return ('' + a[key]).localeCompare(b[key]);
    });
    if (sortedRedemptions[0].id === redemptions[0].id) {
      sortedRedemptions.reverse();
    }
    setRedemptions(sortedRedemptions);
    setLoading(false);
  };

  const handlePageChange = (page) => {
    setActivePage(page);
    if (searchKeyword === '') {
      loadRedemptions(page, pageSize).then();
    } else {
      searchRedemptions(searchKeyword, page, pageSize).then();
    }
  };

  let pageData = redemptions;
  const rowSelection = {
    onSelect: (record, selected) => {},
    onSelectAll: (selected, selectedRows) => {},
    onChange: (selectedRowKeys, selectedRows) => {
      setSelectedKeys(selectedRows);
    },
  };

  const handleRow = (record, index) => {
    if (record.status !== 1 || (record.is_gift && record.used_count >= record.max_uses && record.max_uses !== -1)) {
      return {
        style: {
          background: 'var(--semi-color-disabled-border)',
        },
      };
    } else {
      return {};
    }
  };

  return (
    <>
      <EditRedemption
        refresh={refresh}
        editingRedemption={editingRedemption}
        visiable={showEdit}
        handleClose={closeEdit}
      ></EditRedemption>
      
      {/* 批量删除兑换码模态框 */}
      <BatchDeleteByNameModal 
        visible={showBatchDeleteByName} 
        onClose={() => setShowBatchDeleteByName(false)} 
        onConfirm={refresh}
      />
      
      <Form
        onSubmit={() => {
          searchRedemptions(searchKeyword, activePage, pageSize).then();
        }}
      >
        <Form.Input
          label={t('搜索关键字')}
          field='keyword'
          icon='search'
          iconPosition='left'
          placeholder={t('关键字(id或者名称)')}
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>
      <Divider style={{ margin: '5px 0 15px 0' }} />
      <div>
        <Button
          theme='light'
          type='primary'
          style={{ marginRight: 8 }}
          onClick={() => {
            setEditingRedemption({
              id: undefined,
            });
            setShowEdit(true);
          }}
        >
          {t('添加兑换码')}
        </Button>
        <Button
          label={t('复制所选兑换码')}
          type='warning'
          style={{ marginRight: 8 }}
          onClick={async () => {
            if (selectedKeys.length === 0) {
              showError(t('请至少选择一个兑换码！'));
              return;
            }
            let keys = '';
            for (let i = 0; i < selectedKeys.length; i++) {
              keys +=
                selectedKeys[i].name + '    ' + selectedKeys[i].key + '\n';
            }
            await copyText(keys);
          }}
        >
          {t('复制所选兑换码到剪贴板')}
        </Button>
        <Button
          type='warning'
          style={{ marginRight: 8 }}
          onClick={disableSelectedRedemptions}
        >
          {t('禁用所选兑换码')}
        </Button>
        <Button
          type='danger'
          style={{ marginRight: 8 }}
          onClick={() => setShowBatchDeleteByName(true)}
        >
          {t('按名称批量删除兑换码')}
        </Button>
        <Button
          type='danger'
          onClick={deleteAllDisabledRedemptions}
        >
          {t('删除所有已禁用的兑换码')}
        </Button>
      </div>

      <Table
        style={{ marginTop: 20 }}
        columns={columns}
        dataSource={pageData}
        pagination={{
          currentPage: activePage,
          pageSize: pageSize,
          total: tokenCount,
          showSizeChanger: true,
          pageSizeOpts: [10, 20, 50, 100],
          formatPageText: (page) =>
            t('第 {{start}} - {{end}} 条，共 {{total}} 条', {
              start: page.currentStart,
              end: page.currentEnd,
              total: tokenCount,
            }),
          onPageSizeChange: (size) => {
            setPageSize(size);
            setActivePage(1);
            if (searchKeyword === '') {
              loadRedemptions(1, size).then();
            } else {
              searchRedemptions(searchKeyword, 1, size).then();
            }
          },
          onPageChange: handlePageChange,
        }}
        loading={loading}
        rowSelection={rowSelection}
        onRow={handleRow}
      ></Table>
    </>
  );
};

export default RedemptionsTable;
