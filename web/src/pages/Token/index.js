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
import TokensTable from '../../components/TokensTable';
import { Banner, Layout } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
const Token = () => {
  const { t } = useTranslation();
  return (
    <>
      <Layout>
        <Layout.Header>
          <Banner
            type='warning'
            description={t(
              '令牌无法精确控制使用额度，只允许自用，请勿直接将令牌分发给他人。',
            )}
          />
        </Layout.Header>
        <Layout.Content>
          <TokensTable />
        </Layout.Content>
      </Layout>
    </>
  );
};

export default Token;
