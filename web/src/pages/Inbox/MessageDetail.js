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
import React from 'react';
import { 
  Layout, 
  Typography, 
  Button, 
  Space, 
  Card,
  Tag,
  Divider,
  Tooltip
} from '@douyinfe/semi-ui';
import { IconArrowLeft } from '@douyinfe/semi-icons';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';

const { Title, Text } = Typography;

const MessageDetail = ({ message, onBack }) => {
  const { t } = useTranslation();

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleString();
  };

  const renderContent = (content, format) => {
    if (!content) return null;

    if (format === 'html') {
      return (
        <div 
          dangerouslySetInnerHTML={{ __html: content }}
          style={{
            lineHeight: '1.6',
            color: 'var(--semi-color-text-0)',
            fontSize: '14px'
          }}
        />
      );
    } else if (format === 'markdown') {
      // Configure marked options for security and styling
      marked.setOptions({
        breaks: true,
        gfm: true,
        sanitize: false, // We trust admin content, but in production you might want to sanitize
      });

      const htmlContent = marked(content);
      return (
        <div 
          dangerouslySetInnerHTML={{ __html: htmlContent }}
          style={{
            lineHeight: '1.6',
            color: 'var(--semi-color-text-0)',
            fontSize: '14px'
          }}
          className="markdown-content"
        />
      );
    } else {
      // Plain text
      return (
        <div style={{
          lineHeight: '1.6',
          color: 'var(--semi-color-text-0)',
          fontSize: '14px',
          whiteSpace: 'pre-wrap'
        }}>
          {content}
        </div>
      );
    }
  };

  return (
    <Layout>
      <Layout.Header>
        <Space>
          <h3>
            {t('消息详情')}
          </h3>
        </Space>
      </Layout.Header>
      <Layout.Content style={{ padding: '10px 24px' }}>
        <Card
          style={{
            width: '90%',
            backgroundColor: 'var(--semi-color-bg-1)'
          }}
          bodyStyle={{ padding: '24px' }}
        >
          <Space vertical style={{ width: '100%', alignItems: 'flex-start' }} spacing="large">
            {/* Message Header */}
            <div>
              <div style={{ 
                display: 'flex', 
                justifyContent: 'space-between', 
                alignItems: 'flex-start',
                marginBottom: '12px'
              }}>
                <Tooltip content={message.title}>
                  <Title level={4} style={{ 
                    margin: 0, 
                    flex: 1,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap'
                  }}>
                    {message.title}
                  </Title>
                </Tooltip>
                <Space>
                  {message.format && message.format !== 'markdown' && (
                    <Tag color="blue" size="small">
                      {message.format.toUpperCase()}
                    </Tag>
                  )}
                </Space>
              </div>
              
              <Space>
                <Text type="tertiary" size="small">
                  {t('发送时间')}: {formatTimestamp(message.created_at)}
                </Text>
                {message.read_at && (
                  <>
                    <Text type="tertiary" size="small">•</Text>
                    <Text type="tertiary" size="small">
                      {t('阅读时间')}: {formatTimestamp(message.read_at)}
                    </Text>
                  </>
                )}
              </Space>
            </div>

            <Divider />

            {/* Message Content */}
            <div style={{ minHeight: '200px' }}>
              {renderContent(message.content, message.format)}
            </div>
          </Space>
        </Card>
      </Layout.Content>
      
      {/* Add some CSS for markdown content styling */}
      <style jsx>{`
        .markdown-content h1,
        .markdown-content h2,
        .markdown-content h3,
        .markdown-content h4,
        .markdown-content h5,
        .markdown-content h6 {
          margin-top: 24px;
          margin-bottom: 12px;
          font-weight: 600;
          line-height: 1.25;
        }
        
        .markdown-content h1 { font-size: 24px; }
        .markdown-content h2 { font-size: 20px; }
        .markdown-content h3 { font-size: 18px; }
        .markdown-content h4 { font-size: 16px; }
        .markdown-content h5 { font-size: 14px; }
        .markdown-content h6 { font-size: 12px; }
        
        .markdown-content p {
          margin-bottom: 16px;
        }
        
        .markdown-content ul,
        .markdown-content ol {
          margin-bottom: 16px;
          padding-left: 24px;
        }
        
        .markdown-content li {
          margin-bottom: 4px;
        }
        
        .markdown-content blockquote {
          margin: 16px 0;
          padding: 8px 16px;
          border-left: 4px solid var(--semi-color-primary);
          background-color: var(--semi-color-fill-0);
          color: var(--semi-color-text-1);
        }
        
        .markdown-content code {
          background-color: var(--semi-color-fill-1);
          padding: 2px 4px;
          border-radius: 3px;
          font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
          font-size: 12px;
        }
        
        .markdown-content pre {
          background-color: var(--semi-color-fill-1);
          padding: 12px;
          border-radius: 6px;
          overflow-x: auto;
          margin: 16px 0;
        }
        
        .markdown-content pre code {
          background: none;
          padding: 0;
        }
        
        .markdown-content table {
          border-collapse: collapse;
          width: 100%;
          margin: 16px 0;
        }
        
        .markdown-content th,
        .markdown-content td {
          border: 1px solid var(--semi-color-border);
          padding: 8px 12px;
          text-align: left;
        }
        
        .markdown-content th {
          background-color: var(--semi-color-fill-0);
          font-weight: 600;
        }
        
        .markdown-content a {
          color: var(--semi-color-primary);
          text-decoration: none;
        }
        
        .markdown-content a:hover {
          text-decoration: underline;
        }
        
        .markdown-content hr {
          border: none;
          border-top: 1px solid var(--semi-color-border);
          margin: 24px 0;
        }
      `}</style>
    </Layout>
  );
};

export default MessageDetail;