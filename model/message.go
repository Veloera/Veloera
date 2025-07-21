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
package model

import (
	"errors"
	"time"
)

type Message struct {
	Id        int       `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" gorm:"not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Format    string    `json:"format" gorm:"default:markdown"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy int       `json:"created_by"`
}

type UserMessage struct {
	Id        int        `json:"id" gorm:"primaryKey"`
	UserId    int        `json:"user_id" gorm:"not null;index;uniqueIndex:idx_user_message"`
	MessageId int        `json:"message_id" gorm:"not null;index;uniqueIndex:idx_user_message"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
	Message   Message    `json:"message" gorm:"foreignKey:MessageId"`
}

// Message CRUD operations

func (message *Message) Insert() error {
	return DB.Create(message).Error
}

func (message *Message) Update() error {
	return DB.Save(message).Error
}

func (message *Message) Delete() error {
	if message.Id == 0 {
		return errors.New("message id is empty")
	}
	
	// Start transaction
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	// Delete all user_messages associated with this message
	if err := tx.Where("message_id = ?", message.Id).Delete(&UserMessage{}).Error; err != nil {
		return err
	}

	// Delete the message itself
	if err := tx.Delete(message).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}

func GetMessageById(id int) (*Message, error) {
	if id == 0 {
		return nil, errors.New("message id is empty")
	}
	var message Message
	err := DB.First(&message, "id = ?", id).Error
	return &message, err
}

func GetAllMessages(startIdx int, num int, keyword string) ([]*Message, int64, error) {
	var messages []*Message
	var total int64

	// Start transaction
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer tx.Rollback()

	// Build query
	query := tx.Model(&Message{})
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with newest first
	if err := query.Order("created_at desc").Limit(num).Offset(startIdx).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func SearchMessages(startIdx int, num int, keyword string, startDate string, endDate string) ([]*Message, int64, error) {
	var messages []*Message
	var total int64

	// Start transaction
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer tx.Rollback()

	// Build query
	query := tx.Model(&Message{})
	
	// Add keyword filter
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	
	// Add date range filter
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with newest first
	if err := query.Order("created_at desc").Limit(num).Offset(startIdx).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// UserMessage CRUD operations

func (userMessage *UserMessage) Insert() error {
	return DB.Create(userMessage).Error
}

func (userMessage *UserMessage) MarkAsRead() error {
	if userMessage.Id == 0 {
		return errors.New("user message id is empty")
	}
	now := time.Now()
	userMessage.ReadAt = &now
	return DB.Model(userMessage).Update("read_at", now).Error
}

func GetUserMessages(userId int, startIdx int, num int) ([]*UserMessage, int64, error) {
	if userId == 0 {
		return nil, 0, errors.New("user id is empty")
	}

	var userMessages []*UserMessage
	var total int64

	// Start transaction
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer tx.Rollback()

	// Get total count
	if err := tx.Model(&UserMessage{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with message details, newest first
	if err := tx.Preload("Message").Where("user_id = ?", userId).
		Order("created_at desc").Limit(num).Offset(startIdx).Find(&userMessages).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return userMessages, total, nil
}

func GetUserMessageById(userId int, messageId int) (*UserMessage, error) {
	if userId == 0 || messageId == 0 {
		return nil, errors.New("user id or message id is empty")
	}

	var userMessage UserMessage
	err := DB.Preload("Message").Where("user_id = ? AND message_id = ?", userId, messageId).First(&userMessage).Error
	return &userMessage, err
}

func GetUnreadMessageCount(userId int) (int64, error) {
	if userId == 0 {
		return 0, errors.New("user id is empty")
	}

	var count int64
	err := DB.Model(&UserMessage{}).Where("user_id = ? AND read_at IS NULL", userId).Count(&count).Error
	return count, err
}

func CreateMessageForUsers(message *Message, userIds []int) error {
	if len(userIds) == 0 {
		return errors.New("no users specified")
	}

	// Start transaction
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	// Create the message
	if err := tx.Create(message).Error; err != nil {
		return err
	}

	// Create user_message records for each user
	userMessages := make([]UserMessage, len(userIds))
	for i, userId := range userIds {
		userMessages[i] = UserMessage{
			UserId:    userId,
			MessageId: message.Id,
			CreatedAt: time.Now(),
		}
	}

	// Batch insert user messages
	if err := tx.Create(&userMessages).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}

func MarkUserMessageAsRead(userId int, messageId int) error {
	if userId == 0 || messageId == 0 {
		return errors.New("user id or message id is empty")
	}

	now := time.Now()
	return DB.Model(&UserMessage{}).
		Where("user_id = ? AND message_id = ?", userId, messageId).
		Update("read_at", now).Error
}

// Get message statistics for admin
func GetMessageStats(messageId int) (map[string]interface{}, error) {
	if messageId == 0 {
		return nil, errors.New("message id is empty")
	}

	var totalRecipients int64
	var readCount int64

	// Start transaction
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()

	// Get total recipients
	if err := tx.Model(&UserMessage{}).Where("message_id = ?", messageId).Count(&totalRecipients).Error; err != nil {
		return nil, err
	}

	// Get read count
	if err := tx.Model(&UserMessage{}).Where("message_id = ? AND read_at IS NOT NULL", messageId).Count(&readCount).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_recipients": totalRecipients,
		"read_count":      readCount,
		"unread_count":    totalRecipients - readCount,
		"read_rate":       float64(readCount) / float64(totalRecipients) * 100,
	}

	return stats, nil
}

// Get message recipients with user info for admin
func GetMessageRecipients(messageId int) ([]map[string]interface{}, error) {
	if messageId == 0 {
		return nil, errors.New("message id is empty")
	}

	var results []map[string]interface{}
	
	// Join user_messages with users table to get user info
	err := DB.Table("user_messages").
		Select("user_messages.user_id, users.username, users.display_name, user_messages.read_at").
		Joins("JOIN users ON user_messages.user_id = users.id").
		Where("user_messages.message_id = ?", messageId).
		Order("user_messages.created_at DESC").
		Scan(&results).Error
	
	return results, err
}