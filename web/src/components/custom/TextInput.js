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
import { Input, Typography } from '@douyinfe/semi-ui';
import React from 'react';

const TextInput = ({
  label,
  name,
  value,
  onChange,
  placeholder,
  type = 'text',
}) => {
  return (
    <>
      <div style={{ marginTop: 10 }}>
        <Typography.Text strong>{label}</Typography.Text>
      </div>
      <Input
        name={name}
        placeholder={placeholder}
        onChange={(value) => onChange(value)}
        value={value}
        autoComplete='new-password'
      />
    </>
  );
};

export default TextInput;
