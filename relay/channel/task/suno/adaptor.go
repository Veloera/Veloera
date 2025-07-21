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
package suno

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"time"
	"veloera/common"
	"veloera/constant"
	"veloera/dto"
	"veloera/relay/channel"
	relaycommon "veloera/relay/common"
	"veloera/service"
)

type TaskAdaptor struct {
	ChannelType int
}

func (a *TaskAdaptor) Init(info *relaycommon.TaskRelayInfo) {
	a.ChannelType = info.ChannelType
}

func (a *TaskAdaptor) ValidateRequestAndSetAction(c *gin.Context, info *relaycommon.TaskRelayInfo) (taskErr *dto.TaskError) {
	action := strings.ToUpper(c.Param("action"))

	var sunoRequest *dto.SunoSubmitReq
	err := common.UnmarshalBodyReusable(c, &sunoRequest)
	if err != nil {
		taskErr = service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
		return
	}
	err = actionValidate(c, sunoRequest, action)
	if err != nil {
		taskErr = service.TaskErrorWrapperLocal(err, "invalid_request", http.StatusBadRequest)
		return
	}

	if sunoRequest.ContinueClipId != "" {
		if sunoRequest.TaskID == "" {
			taskErr = service.TaskErrorWrapperLocal(fmt.Errorf("task id is empty"), "invalid_request", http.StatusBadRequest)
			return
		}
		info.OriginTaskID = sunoRequest.TaskID
	}

	info.Action = action
	c.Set("task_request", sunoRequest)
	return nil
}

func (a *TaskAdaptor) BuildRequestURL(info *relaycommon.TaskRelayInfo) (string, error) {
	baseURL := info.BaseUrl
	fullRequestURL := fmt.Sprintf("%s%s", baseURL, "/suno/submit/"+info.Action)
	return fullRequestURL, nil
}

func (a *TaskAdaptor) BuildRequestHeader(c *gin.Context, req *http.Request, info *relaycommon.TaskRelayInfo) error {
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	req.Header.Set("Authorization", "Bearer "+info.ApiKey)
	return nil
}

func (a *TaskAdaptor) BuildRequestBody(c *gin.Context, info *relaycommon.TaskRelayInfo) (io.Reader, error) {
	sunoRequest, ok := c.Get("task_request")
	if !ok {
		err := common.UnmarshalBodyReusable(c, &sunoRequest)
		if err != nil {
			return nil, err
		}
	}
	data, err := json.Marshal(sunoRequest)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func (a *TaskAdaptor) DoRequest(c *gin.Context, info *relaycommon.TaskRelayInfo, requestBody io.Reader) (*http.Response, error) {
	return channel.DoTaskApiRequest(a, c, info, requestBody)
}

func (a *TaskAdaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.TaskRelayInfo) (taskID string, taskData []byte, taskErr *dto.TaskError) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
		return
	}
	var sunoResponse dto.TaskResponse[string]
	err = json.Unmarshal(responseBody, &sunoResponse)
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError)
		return
	}
	if !sunoResponse.IsSuccess() {
		taskErr = service.TaskErrorWrapper(fmt.Errorf(sunoResponse.Message), sunoResponse.Code, http.StatusInternalServerError)
		return
	}

	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)

	_, err = io.Copy(c.Writer, bytes.NewBuffer(responseBody))
	if err != nil {
		taskErr = service.TaskErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError)
		return
	}

	return sunoResponse.Data, nil, nil
}

func (a *TaskAdaptor) GetModelList() []string {
	return ModelList
}

func (a *TaskAdaptor) GetChannelName() string {
	return ChannelName
}

func (a *TaskAdaptor) FetchTask(baseUrl, key string, body map[string]any) (*http.Response, error) {
	requestUrl := fmt.Sprintf("%s/suno/fetch", baseUrl)
	byteBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(byteBody))
	if err != nil {
		common.SysError(fmt.Sprintf("Get Task error: %v", err))
		return nil, err
	}
	defer req.Body.Close()
	// 设置超时时间
	timeout := time.Second * 15
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// 使用带有超时的 context 创建新的请求
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := service.GetHttpClient().Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func actionValidate(c *gin.Context, sunoRequest *dto.SunoSubmitReq, action string) (err error) {
	switch action {
	case constant.SunoActionMusic:
		if sunoRequest.Mv == "" {
			sunoRequest.Mv = "chirp-v3-0"
		}
	case constant.SunoActionLyrics:
		if sunoRequest.Prompt == "" {
			err = fmt.Errorf("prompt_empty")
			return
		}
	default:
		err = fmt.Errorf("invalid_action")
	}
	return
}
