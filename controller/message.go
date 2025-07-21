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
package controller

import (
	"net/http"
	"strconv"
	"time"
	"veloera/common"
	"veloera/dto"
	"veloera/model"

	"github.com/gin-gonic/gin"
)

// GetUserMessages retrieves paginated messages for the authenticated user
func GetUserMessages(c *gin.Context) {
	userId := c.GetInt("id")
	p, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	
	if p < 1 {
		p = 1
	}
	if pageSize <= 0 {
		pageSize = common.ItemsPerPage
	} else if pageSize > 100 {
		pageSize = 100
	}

	startIdx := (p - 1) * pageSize
	userMessages, total, err := model.GetUserMessages(userId, startIdx, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Transform the data to include read status
	var responseData []gin.H
	for _, userMessage := range userMessages {
		responseData = append(responseData, gin.H{
			"id":         userMessage.Message.Id,
			"title":      userMessage.Message.Title,
			"content":    userMessage.Message.Content,
			"format":     userMessage.Message.Format,
			"created_at": userMessage.Message.CreatedAt,
			"is_read":    userMessage.ReadAt != nil,
			"read_at":    userMessage.ReadAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":     responseData,
			"total":     total,
			"page":      p,
			"page_size": pageSize,
		},
	})
}

// MarkMessageAsRead marks a specific message as read for the authenticated user
func MarkMessageAsRead(c *gin.Context) {
	userId := c.GetInt("id")
	messageId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid message ID",
		})
		return
	}

	// Check if the user message exists
	userMessage, err := model.GetUserMessageById(userId, messageId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Message not found or access denied",
		})
		return
	}

	// Mark as read if not already read
	if userMessage.ReadAt == nil {
		err = model.MarkUserMessageAsRead(userId, messageId)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message marked as read",
	})
}

// GetUnreadCount returns the count of unread messages for the authenticated user
func GetUnreadCount(c *gin.Context) {
	userId := c.GetInt("id")
	
	count, err := model.GetUnreadMessageCount(userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"unread_count": count,
		},
	})
}

// Admin Message Management Endpoints

// GetAllMessages retrieves all messages with search and filter capabilities for admin users
func GetAllMessages(c *gin.Context) {
	// Check admin authorization
	role := c.GetInt("role")
	if role < common.RoleAdminUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Access denied: admin privileges required",
		})
		return
	}

	// Parse pagination parameters
	p, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	keyword := c.Query("keyword")

	if p < 1 {
		p = 1
	}
	if pageSize <= 0 {
		pageSize = common.ItemsPerPage
	} else if pageSize > 100 {
		pageSize = 100
	}

	startIdx := (p - 1) * pageSize
	messages, total, err := model.GetAllMessages(startIdx, pageSize, keyword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Transform messages to admin response format
	var responseData []dto.AdminMessageResponse
	for _, message := range messages {
		// Get message statistics
		stats, err := model.GetMessageStats(message.Id)
		if err != nil {
			// Log error but continue with empty stats
			stats = map[string]interface{}{
				"total_recipients": 0,
				"read_count":      0,
				"unread_count":    0,
				"read_rate":       0.0,
			}
		}

		responseData = append(responseData, dto.AdminMessageResponse{
			Id:        message.Id,
			Title:     message.Title,
			Content:   message.Content,
			Format:    message.Format,
			CreatedAt: message.CreatedAt,
			UpdatedAt: message.UpdatedAt,
			CreatedBy: message.CreatedBy,
			Stats:     stats,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":     responseData,
			"total":     total,
			"page":      p,
			"page_size": pageSize,
		},
	})
}

// GetMessage retrieves a specific message by ID for admin users
func GetMessage(c *gin.Context) {
	// Check admin authorization
	role := c.GetInt("role")
	if role < common.RoleAdminUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Access denied: admin privileges required",
		})
		return
	}

	messageId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid message ID",
		})
		return
	}

	message, err := model.GetMessageById(messageId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Message not found",
		})
		return
	}

	// Get message statistics
	stats, err := model.GetMessageStats(messageId)
	if err != nil {
		stats = map[string]interface{}{
			"total_recipients": 0,
			"read_count":      0,
			"unread_count":    0,
			"read_rate":       0.0,
		}
	}

	response := dto.AdminMessageResponse{
		Id:        message.Id,
		Title:     message.Title,
		Content:   message.Content,
		Format:    message.Format,
		CreatedAt: message.CreatedAt,
		UpdatedAt: message.UpdatedAt,
		CreatedBy: message.CreatedBy,
		Stats:     stats,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    response,
	})
}

// CreateMessage creates and sends a new message to specified users
func CreateMessage(c *gin.Context) {
	// Check admin authorization
	role := c.GetInt("role")
	if role < common.RoleAdminUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Access denied: admin privileges required",
		})
		return
	}

	var req dto.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
		return
	}

	// Validate format
	if req.Format == "" {
		req.Format = "markdown"
	}
	if req.Format != "markdown" && req.Format != "html" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid format: must be 'markdown' or 'html'",
		})
		return
	}

	// Validate user IDs
	if len(req.UserIds) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "At least one recipient must be specified",
		})
		return
	}

	// Create message
	message := &model.Message{
		Title:     req.Title,
		Content:   req.Content,
		Format:    req.Format,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: c.GetInt("id"),
	}

	// Create message and associate with users
	err := model.CreateMessageForUsers(message, req.UserIds)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to create message: " + err.Error(),
		})
		return
	}

	// Get message statistics for response
	stats, err := model.GetMessageStats(message.Id)
	if err != nil {
		stats = map[string]interface{}{
			"total_recipients": len(req.UserIds),
			"read_count":      0,
			"unread_count":    len(req.UserIds),
			"read_rate":       0.0,
		}
	}

	response := dto.AdminMessageResponse{
		Id:        message.Id,
		Title:     message.Title,
		Content:   message.Content,
		Format:    message.Format,
		CreatedAt: message.CreatedAt,
		UpdatedAt: message.UpdatedAt,
		CreatedBy: message.CreatedBy,
		Stats:     stats,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message created and sent successfully",
		"data":    response,
	})
}

// UpdateMessage updates an existing message
func UpdateMessage(c *gin.Context) {
	// Check admin authorization
	role := c.GetInt("role")
	if role < common.RoleAdminUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Access denied: admin privileges required",
		})
		return
	}

	messageId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid message ID",
		})
		return
	}

	var req dto.UpdateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
		return
	}

	// Validate format
	if req.Format == "" {
		req.Format = "markdown"
	}
	if req.Format != "markdown" && req.Format != "html" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid format: must be 'markdown' or 'html'",
		})
		return
	}

	// Get existing message
	message, err := model.GetMessageById(messageId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Message not found",
		})
		return
	}

	// Update message fields
	message.Title = req.Title
	message.Content = req.Content
	message.Format = req.Format
	message.UpdatedAt = time.Now()

	// Save updated message
	err = message.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to update message: " + err.Error(),
		})
		return
	}

	// Get message statistics for response
	stats, err := model.GetMessageStats(messageId)
	if err != nil {
		stats = map[string]interface{}{
			"total_recipients": 0,
			"read_count":      0,
			"unread_count":    0,
			"read_rate":       0.0,
		}
	}

	response := dto.AdminMessageResponse{
		Id:        message.Id,
		Title:     message.Title,
		Content:   message.Content,
		Format:    message.Format,
		CreatedAt: message.CreatedAt,
		UpdatedAt: message.UpdatedAt,
		CreatedBy: message.CreatedBy,
		Stats:     stats,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message updated successfully",
		"data":    response,
	})
}

// DeleteMessage deletes a message and all associated user message records
func DeleteMessage(c *gin.Context) {
	// Check admin authorization
	role := c.GetInt("role")
	if role < common.RoleAdminUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Access denied: admin privileges required",
		})
		return
	}

	messageId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid message ID",
		})
		return
	}

	// Get existing message to verify it exists
	message, err := model.GetMessageById(messageId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Message not found",
		})
		return
	}

	// Delete message and all associated user message records
	err = message.Delete()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to delete message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message deleted successfully",
	})
}

// SearchMessages searches messages with keyword and date range filters for admin users
func SearchMessages(c *gin.Context) {
	// Check admin authorization
	role := c.GetInt("role")
	if role < common.RoleAdminUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Access denied: admin privileges required",
		})
		return
	}

	// Parse pagination parameters
	p, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	keyword := c.Query("keyword")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if p < 1 {
		p = 1
	}
	if pageSize <= 0 {
		pageSize = common.ItemsPerPage
	} else if pageSize > 100 {
		pageSize = 100
	}

	startIdx := (p - 1) * pageSize
	messages, total, err := model.SearchMessages(startIdx, pageSize, keyword, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Transform messages to admin response format
	var responseData []dto.AdminMessageResponse
	for _, message := range messages {
		// Get message statistics
		stats, err := model.GetMessageStats(message.Id)
		if err != nil {
			// Log error but continue with empty stats
			stats = map[string]interface{}{
				"total_recipients": 0,
				"read_count":      0,
				"unread_count":    0,
				"read_rate":       0.0,
			}
		}

		responseData = append(responseData, dto.AdminMessageResponse{
			Id:        message.Id,
			Title:     message.Title,
			Content:   message.Content,
			Format:    message.Format,
			CreatedAt: message.CreatedAt,
			UpdatedAt: message.UpdatedAt,
			CreatedBy: message.CreatedBy,
			Stats:     stats,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":     responseData,
			"total":     total,
			"page":      p,
			"page_size": pageSize,
		},
	})
}

// GetMessageRecipients retrieves all recipients for a specific message
func GetMessageRecipients(c *gin.Context) {
	// Check admin authorization
	role := c.GetInt("role")
	if role < common.RoleAdminUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Access denied: admin privileges required",
		})
		return
	}

	messageId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid message ID",
		})
		return
	}

	// Verify message exists
	_, err = model.GetMessageById(messageId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Message not found",
		})
		return
	}

	// Get message recipients
	recipients, err := model.GetMessageRecipients(messageId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to retrieve recipients: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    recipients,
	})
}