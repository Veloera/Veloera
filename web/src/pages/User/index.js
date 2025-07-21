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
import UsersTable from '../../components/UsersTable';
import { Layout } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

const User = () => {
  const { t } = useTranslation();
  return (
    <>
      <Layout>
        <Layout.Header>
          <h3>{t('管理用户')}</h3>
        </Layout.Header>
        <Layout.Content>
          <UsersTable />
        </Layout.Content>
      </Layout>
    </>
  );
};

export default User;
