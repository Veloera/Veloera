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
import React, { useEffect, useState, useRef } from 'react';
import {
  Button,
  Table,
  Modal,
  Form,
  Input,
  InputNumber,
  Space,
  Popconfirm,
  Typography,
  Banner,
  Card,
  Spin,
  Tag,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess, showWarning } from '../../../helpers';
import { useTranslation } from 'react-i18next';
import {
  IconPlus,
  IconEdit,
  IconDelete,
  IconRefresh,
} from '@douyinfe/semi-icons';

const { Title, Text } = Typography;

export default function SettingModelMapping() {
  const { t } = useTranslation();

  const [loading, setLoading] = useState(false);
  const [mappings, setMappings] = useState({});
  const [modalVisible, setModalVisible] = useState(false);
  const [editingMapping, setEditingMapping] = useState(null);
  const [models, setModels] = useState([{ model: '', priorities: 0, weight: 0 }]); // 初始化添加权重字段
  const [virtualModelError, setVirtualModelError] = useState('');
  const [modelErrors, setModelErrors] = useState([]);
  const formApiRef = useRef();

  // 获取模型映射配置
  const fetchMappings = async () => {
    try {
      setLoading(true);
      const res = await API.get('/api/model_mapping/');
      if (res.data.success) {
        setMappings(res.data.data.mapping || {});
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError('获取模型映射配置失败：' + error.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMappings();
  }, []);

  // 当models变化时自动验证
  useEffect(() => {
    validateModels();
  }, [models]);

  // 实时验证虚拟模型名
  const validateVirtualModel = (value) => {
    const trimmedValue = value.trim();
    if (!trimmedValue) {
      setVirtualModelError('');
      return;
    }

    if (!editingMapping && mappings[trimmedValue]) {
      setVirtualModelError(`虚拟模型名 '${trimmedValue}' 已存在`);
    } else if (editingMapping && editingMapping !== trimmedValue && mappings[trimmedValue]) {
      setVirtualModelError(`虚拟模型名 '${trimmedValue}' 已存在`);
    } else {
      setVirtualModelError('');
    }
  };

  // 实时验证实际模型名
  const validateModels = () => {
    const errors = [];
    const modelNames = new Set();
    const duplicates = new Set();

    models.forEach((model, index) => {
      const trimmedModel = model.model.trim();
      if (trimmedModel) {
        if (modelNames.has(trimmedModel)) {
          duplicates.add(trimmedModel);
          errors[index] = `模型名重复: ${trimmedModel}`;
        } else {
          modelNames.add(trimmedModel);
          errors[index] = '';
        }
      } else {
        errors[index] = '';
      }
    });

    // 为重复的模型设置错误信息
    models.forEach((model, index) => {
      const trimmedModel = model.model.trim();
      if (duplicates.has(trimmedModel) && !errors[index]) {
        errors[index] = `模型名重复: ${trimmedModel}`;
      }
    });

    setModelErrors(errors);
  };

  // 转换数据为表格格式
  const getTableData = () => {
    return Object.entries(mappings).map(([virtualModel, items]) => ({
      key: virtualModel,
      virtualModel,
      models: items || [],
    }));
  };

  // 添加/编辑映射
  const handleSaveMapping = async () => {
    try {
      const values = await formApiRef.current.validate();
      const { virtualModel } = values;

      // 处理模型列表数据
      const processedModels = models
        .map((model) => ({
          model: model.model.trim(),
          priorities: parseInt(model.priorities) || 0,
          weight: parseInt(model.weight) || 0,  // 添加权重字段
        }))
        .filter((model) => model.model); // 过滤空模型

      // 前端验证：检查虚拟模型名是否重复
      const trimmedVirtualModel = virtualModel.trim();
      if (!editingMapping && mappings[trimmedVirtualModel]) {
        showError(`虚拟模型名 '${trimmedVirtualModel}' 已存在，请使用不同的名称`);
        return;
      }

      // 如果是编辑模式但虚拟模型名已存在且不是当前编辑的项目
      if (editingMapping && editingMapping !== trimmedVirtualModel && mappings[trimmedVirtualModel]) {
        showError(`虚拟模型名 '${trimmedVirtualModel}' 已存在，请使用不同的名称`);
        return;
      }

      // 前端验证：检查同一虚拟模型内实际模型名是否重复
      const modelNames = new Set();
      const duplicateModels = [];

      for (let i = 0; i < processedModels.length; i++) {
        const modelName = processedModels[i].model;
        if (modelNames.has(modelName)) {
          duplicateModels.push(modelName);
        } else {
          modelNames.add(modelName);
        }
      }

      if (duplicateModels.length > 0) {
        showError(`实际模型名重复: ${duplicateModels.join(', ')}`);
        return;
      }

      // 检查是否有空的实际模型名
      if (processedModels.length === 0) {
        showError('至少需要配置一个实际模型');
        return;
      }

      const newMappings = { ...mappings };

      // 如果是编辑模式且虚拟模型名发生了变化，需要删除原来的映射
      if (editingMapping && editingMapping !== trimmedVirtualModel) {
        delete newMappings[editingMapping];
      }

      newMappings[trimmedVirtualModel] = processedModels;

      const res = await API.put('/api/model_mapping/', {
        mapping: newMappings,
      });
      if (res.data.success) {
        showSuccess(res.data.message || '保存成功');
        setMappings(newMappings);
        setModalVisible(false);
        if (formApiRef.current) {
          formApiRef.current.reset();
        }
        setEditingMapping(null);
        setModels([{ model: '', priorities: 0, weight: 0 }]); // 重置时添加权重字段
        setVirtualModelError('');
        setModelErrors([]);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      if (error.errorFields) {
        showWarning('请检查表单输入');
      } else {
        showError('保存失败：' + error.message);
      }
    }
  };

  // 删除映射
  const handleDeleteMapping = async (virtualModel) => {
    try {
      const newMappings = { ...mappings };
      delete newMappings[virtualModel];

      const res = await API.put('/api/model_mapping/', {
        mapping: newMappings,
      });
      if (res.data.success) {
        showSuccess('删除成功');
        setMappings(newMappings);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError('删除失败：' + error.message);
    }
  };

  // 重新加载配置
  const handleReload = async () => {
    try {
      const res = await API.post('/api/model_mapping/reload');
      if (res.data.success) {
        showSuccess(res.data.message);
        await fetchMappings();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError('重新加载失败：' + error.message);
    }
  };

  // 打开编辑模态框
  const openEditModal = (virtualModel = null) => {
    setEditingMapping(virtualModel);
    setModalVisible(true);
    setVirtualModelError('');
    setModelErrors([]);

    // 延迟设置表单值，确保模态框已渲染
    setTimeout(() => {
      if (formApiRef.current) {
        if (virtualModel && mappings[virtualModel]) {
          formApiRef.current.setValues({
            virtualModel,
          });
          setModels(mappings[virtualModel] || [{ model: '', priorities: 0, weight: 0 }]);
        } else {
          formApiRef.current.reset();
          setModels([{ model: '', priorities: 0, weight: 0 }]); // 新增时添加权重字段
        }
      }
    }, 100);
  };

  // 添加模型
  const addModel = () => {
    setModels([...models, { model: '', priorities: 0, weight: 0 }]); // 添加权重字段
  };

  // 删除模型
  const removeModel = (index) => {
    if (models.length > 1) {
      const newModels = models.filter((_, i) => i !== index);
      setModels(newModels);
    }
  };

  // 更新模型
  const updateModel = (index, field, value) => {
    const newModels = [...models];
    newModels[index][field] = value;
    setModels(newModels);
  };

  // 复制虚拟模型名
  const copyVirtualModel = (virtualModel) => {
    navigator.clipboard
      .writeText(virtualModel)
      .then(() => {
        showSuccess(`已复制: ${virtualModel}`);
      })
      .catch(() => {
        // 备用方法
        const textarea = document.createElement('textarea');
        textarea.value = virtualModel;
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
        showSuccess(`已复制: ${virtualModel}`);
      });
  };

  const columns = [
    {
      title: '虚拟模型名',
      dataIndex: 'virtualModel',
      key: 'virtualModel',
      width: 200,
      render: (virtualModel) => (
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Tag
            color='violet'
            size='large'
            style={{
              fontSize: '13px',
              fontWeight: '500',
              padding: '4px 12px',
              borderRadius: '6px',
              cursor: 'pointer',
              userSelect: 'none',
              transition: 'all 0.2s',
            }}
            onClick={() => copyVirtualModel(virtualModel)}
            onMouseEnter={(e) => {
              e.target.style.transform = 'scale(1.02)';
              e.target.style.boxShadow = '0 2px 8px rgba(0,0,0,0.15)';
            }}
            onMouseLeave={(e) => {
              e.target.style.transform = 'scale(1)';
              e.target.style.boxShadow = 'none';
            }}
            title='点击复制模型名'
          >
            {virtualModel}
          </Tag>
        </div>
      ),
    },
    {
      title: '实际模型映射',
      dataIndex: 'models',
      key: 'models',
      render: (models) => {
        // 找到最高优先级的值
        const maxPriority = Math.max(...models.map(model => model.priorities));

        return (
          <Space wrap>
            {models.map((model, index) => {
              // 判断是否为最高优先级
              const isHighestPriority = model.priorities === maxPriority;

              return (
                <Tag
                  key={index}
                  color={
                    isHighestPriority
                      ? 'green'
                      : model.priorities >= 8
                        ? 'blue'
                        : model.priorities >= 5
                          ? 'orange'
                          : 'grey'
                  }
                  style={
                    isHighestPriority
                      ? { fontWeight: 'bold' }
                      : {}
                  }
                >
                  {model.model} (优先级: {model.priorities}, 权重: {model.weight || 0}) {/* 显示权重信息 */}
                </Tag>
              );
            })}
          </Space>
        );
      },
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button
            icon={<IconEdit />}
            size='small'
            onClick={() => openEditModal(record.virtualModel)}
          >
            编辑
          </Button>
          <Popconfirm
            title='确定要删除这个模型映射吗？'
            onConfirm={() => handleDeleteMapping(record.virtualModel)}
            position='leftTop'
          >
            <Button icon={<IconDelete />} type='danger' size='small'>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Spin spinning={loading}>
        <Form style={{ marginBottom: 15 }}>
          <Form.Section text='模型映射配置'>
            <Banner
              type='info'
              description='配置虚拟模型名到实际模型的映射关系。支持多个实际模型，按优先级和轮询策略选择。'
              style={{ marginBottom: 16 }}
            />

            <Space style={{ marginBottom: 16 }}>
              <Button
                type='primary'
                icon={<IconPlus />}
                onClick={() => openEditModal()}
              >
                添加映射
              </Button>
              <Button icon={<IconRefresh />} onClick={handleReload}>
                重新加载
              </Button>
            </Space>

            <div style={{ position: 'relative', overflow: 'visible' }}>
              <Table
                columns={columns}
                dataSource={getTableData()}
                pagination={false}
                size='small'
                empty={
                  <div style={{ textAlign: 'center', padding: '20px' }}>
                    <Text type='secondary'>暂无模型映射配置</Text>
                  </div>
                }
              />
            </div>
          </Form.Section>
        </Form>
      </Spin>

      {/* 添加/编辑模态框 */}
      <Modal
        title={editingMapping ? '编辑模型映射' : '添加模型映射'}
        visible={modalVisible}
        onOk={handleSaveMapping}
        onCancel={() => {
          setModalVisible(false);
          setEditingMapping(null);
          setModels([{ model: '', priorities: 0, weight: 0 }]);  // 关闭时重置权重字段
          setVirtualModelError('');
          setModelErrors([]);
          if (formApiRef.current) {
            formApiRef.current.reset();
          }
        }}
        width={600}
      >
        <Form
          getFormApi={(formAPI) => (formApiRef.current = formAPI)}
          labelPosition='left'
          labelWidth={120}
        >
          <Form.Input
            field='virtualModel'
            label='虚拟模型名'
            placeholder='输入虚拟模型名，如: gpt-4'
            rules={[
              { required: true, message: '请输入虚拟模型名' },
              { type: 'string', message: '虚拟模型名必须是字符串' },
            ]}
            disabled={!!editingMapping}  // 编辑模式下禁用
            onChange={validateVirtualModel}
            validateStatus={virtualModelError ? 'error' : ''}
            helpText={virtualModelError}
          />
          <div style={{ marginBottom: 16 }}>
            <Text strong>实际模型列表:</Text>
            <Text type="secondary" style={{ marginLeft: 8 }}>
              (优先级决定模型选择顺序，权重决定同优先级模型的选择概率) {/* 添加说明 */}
            </Text>
            {models.map((model, index) => (
              <div
                key={index}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  marginTop: 8,
                  marginBottom: 8,
                }}
              >
                <div style={{ display: 'flex', flexDirection: 'column', marginRight: 8 }}>
                  <Input
                    value={model.model}
                    placeholder='实际模型名'
                    style={{
                      width: 300,
                      borderColor: modelErrors[index] ? '#ff4d4f' : undefined
                    }}
                    onChange={(value) => updateModel(index, 'model', value)}
                  />
                  {modelErrors[index] && (
                    <Text
                      type="danger"
                      size="small"
                      style={{ fontSize: '12px', marginTop: '2px' }}
                    >
                      {modelErrors[index]}
                    </Text>
                  )}
                </div>
                <InputNumber
                  value={model.priorities}
                  placeholder='优先级'
                  min={0}
                  style={{ width: 100, marginRight: 8, alignSelf: 'flex-start' }}
                  onChange={(value) => updateModel(index, 'priorities', value)}
                />
                <InputNumber
                  value={model.weight || 0} // 添加权重输入框
                  placeholder='权重'
                  min={0}
                  style={{ width: 100, marginRight: 8, alignSelf: 'flex-start' }}
                  onChange={(value) => updateModel(index, 'weight', value)}
                />
                <Button
                  type='danger'
                  size='small'
                  onClick={() => removeModel(index)}
                  disabled={models.length === 1}
                  style={{ alignSelf: 'flex-start' }}
                >
                  删除
                </Button>
              </div>
            ))}
            <Button
              onClick={addModel}
              icon={<IconPlus />}
              size='small'
              style={{ marginTop: 8 }}
            >
              添加模型
            </Button>
          </div>
        </Form>
      </Modal>
    </>
  );
}
