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

var (
	UserSettingNotifyType            = "notify_type"                    // QuotaWarningType 额度预警类型
	UserSettingQuotaWarningThreshold = "quota_warning_threshold"        // QuotaWarningThreshold 额度预警阈值
	UserSettingWebhookUrl            = "webhook_url"                    // WebhookUrl webhook地址
	UserSettingWebhookSecret         = "webhook_secret"                 // WebhookSecret webhook密钥
	UserSettingNotificationEmail     = "notification_email"             // NotificationEmail 通知邮箱地址
	UserAcceptUnsetRatioModel        = "accept_unset_model_ratio_model" // AcceptUnsetRatioModel 是否接受未设置价格的模型
	UserSettingShowIPInLogs          = "show_ip_in_logs"                // ShowIPInLogs 是否在消费日志中显示IP
)

var (
	NotifyTypeEmail   = "email"   // Email 邮件
	NotifyTypeWebhook = "webhook" // Webhook
)
