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
package dto

import "time"

type CreateMessageRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Format  string `json:"format"`
	UserIds []int  `json:"user_ids" binding:"required"`
}

type UpdateMessageRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Format  string `json:"format"`
}

type MessageResponse struct {
	Id         int       `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Format     string    `json:"format"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	CreatedBy  int       `json:"created_by"`
	IsRead     bool      `json:"is_read,omitempty"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
}

type AdminMessageResponse struct {
	Id               int                    `json:"id"`
	Title            string                 `json:"title"`
	Content          string                 `json:"content"`
	Format           string                 `json:"format"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	CreatedBy        int                    `json:"created_by"`
	Stats            map[string]interface{} `json:"stats,omitempty"`
}