// Copyright (c) 2025 Tethys Plex
//
// This file is part of Veloera.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package constant

import "veloera/common"

var (
	TokenCacheSeconds         = common.SyncFrequency
	UserId2GroupCacheSeconds  = common.SyncFrequency
	UserId2QuotaCacheSeconds  = common.SyncFrequency
	UserId2StatusCacheSeconds = common.SyncFrequency
)

// Cache keys
const (
	UserGroupKeyFmt    = "user_group:%d"
	UserQuotaKeyFmt    = "user_quota:%d"
	UserEnabledKeyFmt  = "user_enabled:%d"
	UserUsernameKeyFmt = "user_name:%d"
)

const (
	TokenFiledRemainQuota = "RemainQuota"
	TokenFieldGroup       = "Group"
)
