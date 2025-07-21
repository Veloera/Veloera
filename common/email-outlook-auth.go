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
package common

import (
	"errors"
	"net/smtp"
	"strings"
)

type outlookAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &outlookAuth{username, password}
}

func (a *outlookAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *outlookAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unknown fromServer")
		}
	}
	return nil, nil
}

func isOutlookServer(server string) bool {
	// 兼容多地区的outlook邮箱和ofb邮箱
	// 其实应该加一个Option来区分是否用LOGIN的方式登录
	// 先临时兼容一下
	return strings.Contains(server, "outlook") || strings.Contains(server, "onmicrosoft")
}
