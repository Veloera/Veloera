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
package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"veloera/common"
	constant2 "veloera/constant"
	"veloera/dto"
	"veloera/relay/channel"
	"veloera/relay/channel/ai360"
	"veloera/relay/channel/lingyiwanwu"
	"veloera/relay/channel/minimax"
	"veloera/relay/channel/moonshot"
	"veloera/relay/channel/openrouter"
	"veloera/relay/channel/github"
	"veloera/relay/channel/xinference"
	relaycommon "veloera/relay/common"
	"veloera/relay/common_handler"
	"veloera/relay/constant"
	"veloera/service"

	"github.com/gin-gonic/gin"
)

type Adaptor struct {
	ChannelType    int
	ResponseFormat string
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	if !strings.Contains(request.Model, "claude") {
		return nil, fmt.Errorf("you are using openai channel type with path /v1/messages, only claude model supported convert, but got %s", request.Model)
	}
	aiRequest, err := service.ClaudeToOpenAIRequest(*request, info)
	if err != nil {
		return nil, err
	}
	if info.SupportStreamOptions {
		aiRequest.StreamOptions = &dto.StreamOptions{
			IncludeUsage: true,
		}
	}
	return a.ConvertOpenAIRequest(c, info, aiRequest)
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	a.ChannelType = info.ChannelType

	// initialize ThinkingContentInfo when thinking_to_content is enabled
	if think2Content, ok := info.ChannelSetting[constant2.ChannelSettingThinkingToContent].(bool); ok && think2Content {
		info.ThinkingContentInfo = relaycommon.ThinkingContentInfo{
			IsFirstThinkingContent:  true,
			SendLastThinkingContent: false,
			HasSentThinkingContent:  false,
		}
	}
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	if info.RelayFormat == relaycommon.RelayFormatClaude || info.RelayMode == constant.RelayModeResponses {
		var suffixPath string
		if info.RelayFormat == relaycommon.RelayFormatClaude {
			suffixPath = "chat/completions"
		} else {
			suffixPath = "responses"
		}

		// 统一检查 BaseUrl 是否以 "/" 结尾
		if strings.HasSuffix(info.BaseUrl, "/") {
			return info.BaseUrl + suffixPath, nil
		}
		return fmt.Sprintf("%s/v1/%s", info.BaseUrl, suffixPath), nil
	}
	if info.RelayMode == constant.RelayModeRealtime {
		if strings.HasPrefix(info.BaseUrl, "https://") {
			baseUrl := strings.TrimPrefix(info.BaseUrl, "https://")
			baseUrl = "wss://" + baseUrl
			info.BaseUrl = baseUrl
		} else if strings.HasPrefix(info.BaseUrl, "http://") {
			baseUrl := strings.TrimPrefix(info.BaseUrl, "http://")
			baseUrl = "ws://" + baseUrl
			info.BaseUrl = baseUrl
		}
	}
	switch info.ChannelType {
	case common.ChannelTypeAzure:
		apiVersion := info.ApiVersion
		if apiVersion == "" {
			apiVersion = constant2.AzureDefaultAPIVersion
		}
		// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/chatgpt-quickstart?pivots=rest-api&tabs=command-line#rest-api
		requestURL := strings.Split(info.RequestURLPath, "?")[0]
		requestURL = fmt.Sprintf("%s?api-version=%s", requestURL, apiVersion)
		task := strings.TrimPrefix(requestURL, "/v1/")
		model_ := info.UpstreamModelName
		// 2025年5月10日后创建的渠道不移除.
		if info.ChannelCreateTime < constant2.AzureNoRemoveDotTime {
			model_ = strings.Replace(model_, ".", "", -1)
		}
		// https://github.com/songquanpeng/veloera/issues/67
		requestURL = fmt.Sprintf("/openai/deployments/%s/%s", model_, task)
		if info.RelayMode == constant.RelayModeRealtime {
			requestURL = fmt.Sprintf("/openai/realtime?deployment=%s&api-version=%s", model_, apiVersion)
		}
		return relaycommon.GetFullRequestURL(info.BaseUrl, requestURL, info.ChannelType), nil
	case common.ChannelTypeMiniMax:
		return minimax.GetRequestURL(info)
	case common.ChannelTypeCustom:
		url := info.BaseUrl
		url = strings.Replace(url, "{model}", info.UpstreamModelName, -1)
		return url, nil
	default:
		return relaycommon.GetFullRequestURL(info.BaseUrl, info.RequestURLPath, info.ChannelType), nil
	}
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, header *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, header)
	if info.ChannelType == common.ChannelTypeAzure {
		header.Set("api-key", info.ApiKey)
		return nil
	}
	if info.ChannelType == common.ChannelTypeOpenAI && "" != info.Organization {
		header.Set("OpenAI-Organization", info.Organization)
	}
	if info.RelayMode == constant.RelayModeRealtime {
		swp := c.Request.Header.Get("Sec-WebSocket-Protocol")
		if swp != "" {
			items := []string{
				"realtime",
				"openai-insecure-api-key." + info.ApiKey,
				"openai-beta.realtime-v1",
			}
			header.Set("Sec-WebSocket-Protocol", strings.Join(items, ","))
			//req.Header.Set("Sec-WebSocket-Key", c.Request.Header.Get("Sec-WebSocket-Key"))
			//req.Header.Set("Sec-Websocket-Extensions", c.Request.Header.Get("Sec-Websocket-Extensions"))
			//req.Header.Set("Sec-Websocket-Version", c.Request.Header.Get("Sec-Websocket-Version"))
		} else {
			header.Set("openai-beta", "realtime=v1")
			header.Set("Authorization", "Bearer "+info.ApiKey)
		}
	} else {
		header.Set("Authorization", "Bearer "+info.ApiKey)
	}
	if info.ChannelType == common.ChannelTypeOpenRouter {
		header.Set("HTTP-Referer", "https://github.com/Veloera/Veloera")
		header.Set("X-Title", "Veloera")
	}
	return nil
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	if info.ChannelType != common.ChannelTypeOpenAI && info.ChannelType != common.ChannelTypeAzure {
		request.StreamOptions = nil
	}
	if strings.HasPrefix(request.Model, "o1") || strings.HasPrefix(request.Model, "o3") {
		if request.MaxCompletionTokens == 0 && request.MaxTokens != 0 {
			request.MaxCompletionTokens = request.MaxTokens
			request.MaxTokens = 0
		}
		if strings.HasPrefix(request.Model, "o3") || strings.HasPrefix(request.Model, "o1") {
			request.Temperature = nil
		}
		if strings.HasSuffix(request.Model, "-high") {
			request.ReasoningEffort = "high"
			request.Model = strings.TrimSuffix(request.Model, "-high")
		} else if strings.HasSuffix(request.Model, "-low") {
			request.ReasoningEffort = "low"
			request.Model = strings.TrimSuffix(request.Model, "-low")
		} else if strings.HasSuffix(request.Model, "-medium") {
			request.ReasoningEffort = "medium"
			request.Model = strings.TrimSuffix(request.Model, "-medium")
		}
		info.ReasoningEffort = request.ReasoningEffort
		info.UpstreamModelName = request.Model
	}
	if request.Model == "o1" || request.Model == "o1-2024-12-17" || strings.HasPrefix(request.Model, "o3") {
		//修改第一个Message的内容，将system改为developer
		if len(request.Messages) > 0 && request.Messages[0].Role == "system" {
			request.Messages[0].Role = "developer"
		}
	}

	return request, nil
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	a.ResponseFormat = request.ResponseFormat
	if info.RelayMode == constant.RelayModeAudioSpeech {
		jsonData, err := json.Marshal(request)
		if err != nil {
			return nil, fmt.Errorf("error marshalling object: %w", err)
		}
		return bytes.NewReader(jsonData), nil
	} else {
		var requestBody bytes.Buffer
		writer := multipart.NewWriter(&requestBody)

		writer.WriteField("model", request.Model)

		// 获取所有表单字段
		formData := c.Request.PostForm

		// 遍历表单字段并打印输出
		for key, values := range formData {
			if key == "model" {
				continue
			}
			for _, value := range values {
				writer.WriteField(key, value)
			}
		}

		// 添加文件字段
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			return nil, errors.New("file is required")
		}
		defer file.Close()

		part, err := writer.CreateFormFile("file", header.Filename)
		if err != nil {
			return nil, errors.New("create form file failed")
		}
		if _, err := io.Copy(part, file); err != nil {
			return nil, errors.New("copy file failed")
		}

		// 关闭 multipart 编写器以设置分界线
		writer.Close()
		c.Request.Header.Set("Content-Type", writer.FormDataContentType())
		return &requestBody, nil
	}
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	// 模型后缀转换 reasoning effort
	if strings.HasSuffix(request.Model, "-high") {
		request.Reasoning.Effort = "high"
		request.Model = strings.TrimSuffix(request.Model, "-high")
	} else if strings.HasSuffix(request.Model, "-low") {
		request.Reasoning.Effort = "low"
		request.Model = strings.TrimSuffix(request.Model, "-low")
	} else if strings.HasSuffix(request.Model, "-medium") {
		request.Reasoning.Effort = "medium"
		request.Model = strings.TrimSuffix(request.Model, "-medium")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	if info.RelayMode == constant.RelayModeAudioTranscription || info.RelayMode == constant.RelayModeAudioTranslation {
		return channel.DoFormRequest(a, c, info, requestBody)
	} else if info.RelayMode == constant.RelayModeRealtime {
		return channel.DoWssRequest(a, c, info, requestBody)
	} else {
		return channel.DoApiRequest(a, c, info, requestBody)
	}
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *dto.OpenAIErrorWithStatusCode) {
	switch info.RelayMode {
	case constant.RelayModeRealtime:
		err, usage = OpenaiRealtimeHandler(c, info)
	case constant.RelayModeAudioSpeech:
		err, usage = OpenaiTTSHandler(c, resp, info)
	case constant.RelayModeAudioTranslation:
		fallthrough
	case constant.RelayModeAudioTranscription:
		err, usage = OpenaiSTTHandler(c, resp, info, a.ResponseFormat)
	case constant.RelayModeImagesGenerations:
		err, usage = OpenaiTTSHandler(c, resp, info)
	case constant.RelayModeRerank:
		err, usage = common_handler.RerankHandler(c, info, resp)
	case constant.RelayModeResponses:
		if info.IsStream {
			err, usage = OaiResponsesStreamHandler(c, resp, info)
		} else {
			err, usage = OpenaiResponsesHandler(c, resp, info)
		}
	default:
		if info.IsStream {
			err, usage = OaiStreamHandler(c, resp, info)
		} else {
			err, usage = OpenaiHandler(c, resp, info)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	switch a.ChannelType {
	case common.ChannelType360:
		return ai360.ModelList
	case common.ChannelTypeMoonshot:
		return moonshot.ModelList
	case common.ChannelTypeLingYiWanWu:
		return lingyiwanwu.ModelList
	case common.ChannelTypeMiniMax:
		return minimax.ModelList
	case common.ChannelTypeXinference:
		return xinference.ModelList
	case common.ChannelTypeOpenRouter:
		return openrouter.ModelList
	case common.ChannelTypeGitHub:
		return github.ModelList
	default:
		return ModelList
	}
}

func (a *Adaptor) GetChannelName() string {
	switch a.ChannelType {
	case common.ChannelType360:
		return ai360.ChannelName
	case common.ChannelTypeMoonshot:
		return moonshot.ChannelName
	case common.ChannelTypeLingYiWanWu:
		return lingyiwanwu.ChannelName
	case common.ChannelTypeMiniMax:
		return minimax.ChannelName
	case common.ChannelTypeXinference:
		return xinference.ChannelName
	case common.ChannelTypeOpenRouter:
		return openrouter.ChannelName
	case common.ChannelTypeGitHub:
		return github.ChannelName
	default:
		return ChannelName
	}
}
