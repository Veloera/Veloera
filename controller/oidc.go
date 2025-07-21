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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"veloera/common"
	"veloera/model"
	"veloera/setting"
	"veloera/setting/system_setting"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type OidcResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

type OidcUser struct {
	OpenID            string `json:"sub"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Picture           string `json:"picture"`
}

func getOidcUserInfoByCode(code string) (*OidcUser, error) {
	if code == "" {
		return nil, errors.New("无效的参数")
	}

	values := url.Values{}
	values.Set("client_id", system_setting.GetOIDCSettings().ClientId)
	values.Set("client_secret", system_setting.GetOIDCSettings().ClientSecret)
	values.Set("code", code)
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", fmt.Sprintf("%s/oauth/oidc", setting.ServerAddress))
	formData := values.Encode()
	req, err := http.NewRequest("POST", system_setting.GetOIDCSettings().TokenEndpoint, strings.NewReader(formData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		common.SysLog(err.Error())
		return nil, errors.New("无法连接至 OIDC 服务器，请稍后重试！")
	}
	defer res.Body.Close()
	var oidcResponse OidcResponse
	err = json.NewDecoder(res.Body).Decode(&oidcResponse)
	if err != nil {
		return nil, err
	}

	if oidcResponse.AccessToken == "" {
		common.SysError("OIDC 获取 Token 失败，请检查设置！")
		return nil, errors.New("OIDC 获取 Token 失败，请检查设置！")
	}

	req, err = http.NewRequest("GET", system_setting.GetOIDCSettings().UserInfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+oidcResponse.AccessToken)
	res2, err := client.Do(req)
	if err != nil {
		common.SysLog(err.Error())
		return nil, errors.New("无法连接至 OIDC 服务器，请稍后重试！")
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		common.SysError("OIDC 获取用户信息失败！请检查设置！")
		return nil, errors.New("OIDC 获取用户信息失败！请检查设置！")
	}

	var oidcUser OidcUser
	err = json.NewDecoder(res2.Body).Decode(&oidcUser)
	if err != nil {
		return nil, err
	}
	if oidcUser.OpenID == "" || oidcUser.Email == "" {
		common.SysError("OIDC 获取用户信息为空！请检查设置！")
		return nil, errors.New("OIDC 获取用户信息为空！请检查设置！")
	}
	return &oidcUser, nil
}

func OidcAuth(c *gin.Context) {
	session := sessions.Default(c)
	state := c.Query("state")
	if state == "" || session.Get("oauth_state") == nil || state != session.Get("oauth_state").(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "state is empty or not same",
		})
		return
	}
	username := session.Get("username")
	if username != nil {
		OidcBind(c)
		return
	}
	if !system_setting.GetOIDCSettings().Enabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 OIDC 登录以及注册",
		})
		return
	}
	code := c.Query("code")
	oidcUser, err := getOidcUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user := model.User{
		OidcId: oidcUser.OpenID,
	}
	if model.IsOidcIdAlreadyTaken(user.OidcId) {
		err := user.FillUserByOidcId()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	} else {
		if common.RegisterEnabled {
			user.Email = oidcUser.Email
			// Initialize username
			if oidcUser.PreferredUsername != "" {
				user.Username = oidcUser.PreferredUsername
			} else {
				user.Username = "oidc_" + strconv.Itoa(model.GetMaxUserId()+1)
			}

			if oidcUser.Name != "" {
				user.DisplayName = oidcUser.Name
			} else {
				user.DisplayName = "OIDC User"
			}

			// Try to insert user, handle username uniqueness constraint
			var err error
			// If custom username provided, try it first, then fallback to incremental ids
			isCustomUsername := oidcUser.PreferredUsername != ""
			baseUserId := model.GetMaxUserId() + 1

			for i := 0; i < 5; i++ {
				if i > 0 || !isCustomUsername {
					// If custom username failed on first attempt or if using generated username
					user.Username = "oidc_" + strconv.Itoa(baseUserId+i-1)
				}

				err = user.Insert(0)
				if err == nil {
					break
				}
				// Check if error is about username uniqueness
				if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
					continue
				}
				// If it's another error, break the loop
				break
			}

			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "管理员关闭了新用户注册",
			})
			return
		}
	}

	if user.Status != common.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{
			"message": "用户已被封禁",
			"success": false,
		})
		return
	}
	setupLogin(&user, c)
}

func OidcBind(c *gin.Context) {
	if !system_setting.GetOIDCSettings().Enabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "管理员未开启通过 OIDC 登录以及注册",
		})
		return
	}
	code := c.Query("code")
	oidcUser, err := getOidcUserInfoByCode(code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user := model.User{
		OidcId: oidcUser.OpenID,
	}
	if model.IsOidcIdAlreadyTaken(user.OidcId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该 OIDC 账户已被绑定",
		})
		return
	}
	session := sessions.Default(c)
	id := session.Get("id")
	// id := c.GetInt("id")  // critical bug!
	user.Id = id.(int)
	err = user.FillUserById()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	user.OidcId = oidcUser.OpenID
	err = user.Update(false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "bind",
	})
	return
}
