package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"veloera/common"
	"veloera/dto"
	"veloera/model"
	relaycommon "veloera/relay/common"
	"veloera/service"
)

func RelayGemini(c *gin.Context) {
	modelParam := c.Param("model")
	action := "generateContent"
	modelName := modelParam
	if idx := strings.Index(modelParam, ":"); idx >= 0 {
		action = modelParam[idx+1:]
		modelName = modelParam[:idx]
	}

	stream := false
	switch action {
	case "generateContent":
		stream = false
	case "streamGenerateContent":
		stream = true
	default:
		respondGeminiError(c, http.StatusNotFound, "unsupported action")
		return
	}

	if modelName == "" {
		respondGeminiError(c, http.StatusBadRequest, "model is required")
		return
	}

	bodyBytes, err := common.GetRequestBody(c)
	if err != nil {
		respondGeminiError(c, http.StatusBadRequest, "failed to read request body")
		return
	}

	var geminiReq dto.GeminiCompatGenerateContentRequest
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &geminiReq); err != nil {
			respondGeminiError(c, http.StatusBadRequest, "invalid JSON body")
			return
		}
	}

	openaiReq, err := service.ConvertGeminiCompatRequestToOpenAI(&geminiReq, modelName, stream)
	if err != nil {
		respondGeminiError(c, http.StatusBadRequest, err.Error())
		return
	}

	convertedBody, err := json.Marshal(openaiReq)
	if err != nil {
		respondGeminiError(c, http.StatusInternalServerError, "failed to encode internal request")
		return
	}

	c.Set(common.KeyRequestBody, convertedBody)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(convertedBody))
	c.Request.ContentLength = int64(len(convertedBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("relay_format", relaycommon.RelayFormatGemini)

	Relay(c)
}

func ListGeminiModels(c *gin.Context) {
	userID := c.GetInt("id")
	group, err := model.GetUserGroup(userID, true)
	if err != nil {
		respondGeminiError(c, http.StatusInternalServerError, "failed to load user group")
		return
	}

	tokenGroup := c.GetString("token_group")
	if tokenGroup != "" {
		group = tokenGroup
	}

	models := model.GetGroupModels(group)
	response := dto.GeminiModelListResponse{Models: make([]dto.GeminiModel, 0, len(models))}

	for _, name := range models {
		response.Models = append(response.Models, dto.GeminiModel{
			Name:                       "models/" + name,
			BaseModelID:                name,
			Version:                    "v1beta",
			DisplayName:                name,
			Description:                "",
			InputTokenLimit:            32768,
			OutputTokenLimit:           8192,
			SupportedGenerationMethods: []string{"generateContent", "streamGenerateContent"},
		})
	}

	c.JSON(http.StatusOK, response)
}

func RetrieveGeminiModel(c *gin.Context) {
	modelName := c.Param("model")
	if modelName == "" {
		respondGeminiError(c, http.StatusNotFound, "model not found")
		return
	}

	userID := c.GetInt("id")
	group, err := model.GetUserGroup(userID, true)
	if err != nil {
		respondGeminiError(c, http.StatusInternalServerError, "failed to load user group")
		return
	}
	tokenGroup := c.GetString("token_group")
	if tokenGroup != "" {
		group = tokenGroup
	}

	models := model.GetGroupModels(group)
	found := false
	for _, name := range models {
		if name == modelName {
			found = true
			break
		}
	}

	if !found {
		respondGeminiError(c, http.StatusNotFound, "model not found")
		return
	}

	c.JSON(http.StatusOK, dto.GeminiModel{
		Name:                       "models/" + modelName,
		BaseModelID:                modelName,
		Version:                    "v1beta",
		DisplayName:                modelName,
		Description:                "",
		InputTokenLimit:            32768,
		OutputTokenLimit:           8192,
		SupportedGenerationMethods: []string{"generateContent", "streamGenerateContent"},
	})
}

func respondGeminiError(c *gin.Context, status int, message string) {
	resp := service.BuildGeminiErrorResponse(status, message)
	c.JSON(status, resp)
}
