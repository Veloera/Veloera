package dto

type GeminiCompatGenerateContentRequest struct {
	Contents          []GeminiCompatContent         `json:"contents"`
	Tools             []GeminiCompatTool            `json:"tools,omitempty"`
	ToolConfig        *GeminiCompatToolConfig       `json:"toolConfig,omitempty"`
	SafetySettings    []GeminiCompatSafetySetting   `json:"safetySettings,omitempty"`
	SystemInstruction *GeminiCompatContent          `json:"systemInstruction,omitempty"`
	GenerationConfig  *GeminiCompatGenerationConfig `json:"generationConfig,omitempty"`
	CachedContent     string                        `json:"cachedContent,omitempty"`
}

type GeminiCompatContent struct {
	Role  string             `json:"role,omitempty"`
	Parts []GeminiCompatPart `json:"parts"`
}

type GeminiCompatPart struct {
	Text             string                        `json:"text,omitempty"`
	InlineData       *GeminiCompatInlineData       `json:"inlineData,omitempty"`
	FileData         *GeminiCompatFileData         `json:"fileData,omitempty"`
	FunctionCall     *GeminiCompatFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *GeminiCompatFunctionResponse `json:"functionResponse,omitempty"`
}

type GeminiCompatInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type GeminiCompatFileData struct {
	MimeType string `json:"mimeType,omitempty"`
	FileURI  string `json:"fileUri,omitempty"`
}

type GeminiCompatFunctionCall struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"args"`
}

type GeminiCompatFunctionResponse struct {
	Name     string      `json:"name"`
	Response interface{} `json:"response"`
}

type GeminiCompatTool struct {
	FunctionDeclarations  []GeminiCompatFunctionDeclaration `json:"functionDeclarations,omitempty"`
	GoogleSearch          interface{}                       `json:"googleSearch,omitempty"`
	GoogleSearchRetrieval interface{}                       `json:"googleSearchRetrieval,omitempty"`
	CodeExecution         interface{}                       `json:"codeExecution,omitempty"`
}

type GeminiCompatFunctionDeclaration struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

type GeminiCompatToolConfig struct {
	FunctionCallingConfig *GeminiCompatFunctionCallingConfig `json:"functionCallingConfig,omitempty"`
}

type GeminiCompatFunctionCallingConfig struct {
	Mode                 string   `json:"mode,omitempty"`
	AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
}

type GeminiCompatSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GeminiCompatGenerationConfig struct {
	StopSequences    []string `json:"stopSequences,omitempty"`
	ResponseMimeType string   `json:"responseMimeType,omitempty"`
	CandidateCount   int      `json:"candidateCount,omitempty"`
	MaxOutputTokens  int      `json:"maxOutputTokens,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"topP,omitempty"`
	TopK             *int     `json:"topK,omitempty"`
}

type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type GeminiErrorResponse struct {
	Error GeminiError `json:"error"`
}

type GeminiModel struct {
	Name                       string   `json:"name"`
	BaseModelID                string   `json:"baseModelId"`
	Version                    string   `json:"version"`
	DisplayName                string   `json:"displayName"`
	Description                string   `json:"description"`
	InputTokenLimit            int      `json:"inputTokenLimit"`
	OutputTokenLimit           int      `json:"outputTokenLimit"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	Temperature                *float64 `json:"temperature,omitempty"`
	MaxTemperature             *float64 `json:"maxTemperature,omitempty"`
	TopP                       *float64 `json:"topP,omitempty"`
	TopK                       *int     `json:"topK,omitempty"`
}

type GeminiModelListResponse struct {
	Models        []GeminiModel `json:"models"`
	NextPageToken string        `json:"nextPageToken,omitempty"`
}
