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
package relay

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"veloera/common"
	"veloera/dto"
	relaycommon "veloera/relay/common"
	"veloera/relay/helper"
	"veloera/service"
	"veloera/setting"
	"veloera/setting/model_setting"
)

func getAndValidateResponsesRequest(c *gin.Context, relayInfo *relaycommon.RelayInfo) (*dto.OpenAIResponsesRequest, error) {
	request := &dto.OpenAIResponsesRequest{}
	err := common.UnmarshalBodyReusable(c, request)
	if err != nil {
		return nil, err
	}
	if request.Model == "" {
		return nil, errors.New("model is required")
	}
	if len(request.Input) == 0 {
		return nil, errors.New("input is required")
	}
	relayInfo.IsStream = request.Stream
	return request, nil
}

func checkInputSensitive(textRequest *dto.OpenAIResponsesRequest, info *relaycommon.RelayInfo) ([]string, error) {
	sensitiveWords, err := service.CheckSensitiveInput(textRequest.Input)
	return sensitiveWords, err
}

func getInputTokens(req *dto.OpenAIResponsesRequest, info *relaycommon.RelayInfo) (int, error) {
	inputTokens, err := service.CountTokenInput(req.Input, req.Model)
	info.PromptTokens = inputTokens
	return inputTokens, err
}

func ResponsesHelper(c *gin.Context) (openaiErr *dto.OpenAIErrorWithStatusCode) {
	relayInfo := relaycommon.GenRelayInfo(c)

	// Validate request and check sensitive content
	req, openaiErr := validateAndPrepareRequest(c, relayInfo)
	if openaiErr != nil {
		return openaiErr
	}

	// Handle model mapping and token counting
	openaiErr = handleModelAndTokens(c, relayInfo, req)
	if openaiErr != nil {
		return openaiErr
	}

	// Calculate price and pre-consume quota
	priceData, preConsumedQuota, userQuota, openaiErr := handleQuotaAndPricing(c, relayInfo, req)
	if openaiErr != nil {
		return openaiErr
	}

	// Return quota if there's an error
	defer func() {
		if openaiErr != nil {
			returnPreConsumedQuota(c, relayInfo, userQuota, preConsumedQuota)
		}
	}()

	// Prepare and send request
	httpResp, openaiErr := prepareAndSendRequest(c, relayInfo, req)
	if openaiErr != nil {
		return openaiErr
	}

	// Process response and handle quota consumption
	usage, openaiErr := processResponse(c, httpResp, relayInfo)
	if openaiErr != nil {
		return openaiErr
	}

	// Post-consume quota
	postProcessQuota(c, relayInfo, usage, preConsumedQuota, userQuota, priceData)

	return nil
}

func validateAndPrepareRequest(c *gin.Context, relayInfo *relaycommon.RelayInfo) (*dto.OpenAIResponsesRequest, *dto.OpenAIErrorWithStatusCode) {
	req, err := getAndValidateResponsesRequest(c, relayInfo)
	if err != nil {
		common.LogError(c, fmt.Sprintf("getAndValidateResponsesRequest error: %s", err.Error()))
		return nil, service.OpenAIErrorWrapperLocal(err, "invalid_responses_request", http.StatusBadRequest)
	}

	tokenGroup := c.GetString("token_group")
	if setting.ShouldCheckPromptSensitiveWithGroup(tokenGroup) {
		sensitiveWords, err := checkInputSensitive(req, relayInfo)
		if err != nil {
			common.LogWarn(c, fmt.Sprintf("user sensitive words detected: %s", strings.Join(sensitiveWords, ", ")))
			return nil, service.OpenAIErrorWrapperLocal(err, "check_request_sensitive_error", http.StatusBadRequest)
		}
	}

	return req, nil
}

func handleModelAndTokens(c *gin.Context, relayInfo *relaycommon.RelayInfo, req *dto.OpenAIResponsesRequest) *dto.OpenAIErrorWithStatusCode {
	err := helper.ModelMappedHelper(c, relayInfo)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "model_mapped_error", http.StatusBadRequest)
	}

	req.Model = relayInfo.UpstreamModelName

	if value, exists := c.Get("prompt_tokens"); exists {
		promptTokens := value.(int)
		relayInfo.SetPromptTokens(promptTokens)
	} else {
		promptTokens, err := getInputTokens(req, relayInfo)
		if err != nil {
			return service.OpenAIErrorWrapperLocal(err, "count_input_tokens_error", http.StatusBadRequest)
		}
		c.Set("prompt_tokens", promptTokens)
	}

	return nil
}

func handleQuotaAndPricing(c *gin.Context, relayInfo *relaycommon.RelayInfo, req *dto.OpenAIResponsesRequest) (helper.PriceData, int, int, *dto.OpenAIErrorWithStatusCode) {
	priceData, err := helper.ModelPriceHelper(c, relayInfo, relayInfo.PromptTokens, int(req.MaxOutputTokens))
	if err != nil {
		return helper.PriceData{}, 0, 0, service.OpenAIErrorWrapperLocal(err, "model_price_error", http.StatusInternalServerError)
	}

	// pre consume quota
	preConsumedQuota, userQuota, openaiErr := preConsumeQuota(c, priceData.ShouldPreConsumedQuota, relayInfo)
	if openaiErr != nil {
		return helper.PriceData{}, 0, 0, openaiErr
	}

	return priceData, preConsumedQuota, userQuota, nil
}

func prepareAndSendRequest(c *gin.Context, relayInfo *relaycommon.RelayInfo, req *dto.OpenAIResponsesRequest) (*http.Response, *dto.OpenAIErrorWithStatusCode) {
	adaptor := GetAdaptor(relayInfo.ApiType)
	if adaptor == nil {
		return nil, service.OpenAIErrorWrapperLocal(fmt.Errorf("invalid api type: %d", relayInfo.ApiType), "invalid_api_type", http.StatusBadRequest)
	}

	adaptor.Init(relayInfo)
	requestBody, openaiErr := prepareRequestBody(c, relayInfo, req, adaptor)
	if openaiErr != nil {
		return nil, openaiErr
	}

	resp, err := adaptor.DoRequest(c, relayInfo, requestBody)
	if err != nil {
		return nil, service.OpenAIErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	var httpResp *http.Response
	if resp != nil {
		httpResp = resp.(*http.Response)
		statusCodeMappingStr := c.GetString("status_code_mapping")

		if httpResp.StatusCode != http.StatusOK {
			openaiErr := service.RelayErrorHandler(httpResp, false)
			service.ResetStatusCode(openaiErr, statusCodeMappingStr)
			return nil, openaiErr
		}
	}

	return httpResp, nil
}

func prepareRequestBody(c *gin.Context, relayInfo *relaycommon.RelayInfo, req *dto.OpenAIResponsesRequest, adaptor interface{}) (io.Reader, *dto.OpenAIErrorWithStatusCode) {
	if model_setting.GetGlobalSettings().PassThroughRequestEnabled {
		body, err := common.GetRequestBody(c)
		if err != nil {
			return nil, service.OpenAIErrorWrapperLocal(err, "get_request_body_error", http.StatusInternalServerError)
		}
		return bytes.NewBuffer(body), nil
	}

	convertedRequest, err := adaptor.(interface {
		ConvertOpenAIResponsesRequest(*gin.Context, *relaycommon.RelayInfo, dto.OpenAIResponsesRequest) (interface{}, error)
	}).ConvertOpenAIResponsesRequest(c, relayInfo, *req)
	if err != nil {
		return nil, service.OpenAIErrorWrapperLocal(err, "convert_request_error", http.StatusBadRequest)
	}

	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		return nil, service.OpenAIErrorWrapperLocal(err, "marshal_request_error", http.StatusInternalServerError)
	}

	// Apply param override
	if len(relayInfo.ParamOverride) > 0 {
		jsonData, err = applyParamOverride(jsonData, relayInfo.ParamOverride)
		if err != nil {
			return nil, service.OpenAIErrorWrapperLocal(err, "param_override_failed", http.StatusInternalServerError)
		}
	}

	if common.DebugEnabled {
		println("requestBody: ", string(jsonData))
	}

	return bytes.NewBuffer(jsonData), nil
}

func applyParamOverride(jsonData []byte, paramOverride map[string]interface{}) ([]byte, error) {
	reqMap := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &reqMap)
	if err != nil {
		return nil, fmt.Errorf("param_override_unmarshal_failed: %w", err)
	}

	for key, value := range paramOverride {
		reqMap[key] = value
	}

	return json.Marshal(reqMap)
}

func processResponse(c *gin.Context, httpResp *http.Response, relayInfo *relaycommon.RelayInfo) (*dto.Usage, *dto.OpenAIErrorWithStatusCode) {
	adaptor := GetAdaptor(relayInfo.ApiType)
	usage, openaiErr := adaptor.DoResponse(c, httpResp, relayInfo)

	if openaiErr != nil {
		statusCodeMappingStr := c.GetString("status_code_mapping")
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return nil, openaiErr
	}

	return usage.(*dto.Usage), nil
}

func postProcessQuota(c *gin.Context, relayInfo *relaycommon.RelayInfo, usage *dto.Usage, preConsumedQuota int, userQuota int, priceData helper.PriceData) {
	if strings.HasPrefix(relayInfo.OriginModelName, "gpt-4o-audio") {
		service.PostAudioConsumeQuota(c, relayInfo, usage, preConsumedQuota, userQuota, priceData, "")
	} else {
		postConsumeQuota(c, relayInfo, usage, preConsumedQuota, userQuota, priceData, "")
	}
}
