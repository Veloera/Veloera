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
	commonconstant "veloera/constant"
	"veloera/relay/channel"
	"veloera/relay/channel/ali"
	"veloera/relay/channel/aws"
	"veloera/relay/channel/baidu"
	"veloera/relay/channel/baidu_v2"
	"veloera/relay/channel/claude"
	"veloera/relay/channel/cloudflare"
	"veloera/relay/channel/cohere"
	"veloera/relay/channel/deepseek"
	"veloera/relay/channel/dify"
	"veloera/relay/channel/gemini"
	"veloera/relay/channel/jina"
	"veloera/relay/channel/mistral"
	"veloera/relay/channel/mokaai"
	"veloera/relay/channel/ollama"
	"veloera/relay/channel/openai"
	"veloera/relay/channel/palm"
	"veloera/relay/channel/perplexity"
	"veloera/relay/channel/siliconflow"
	"veloera/relay/channel/task/suno"
	"veloera/relay/channel/tencent"
	"veloera/relay/channel/vertex"
	"veloera/relay/channel/volcengine"
	"veloera/relay/channel/xai"
	"veloera/relay/channel/xunfei"
	"veloera/relay/channel/zhipu"
	"veloera/relay/channel/zhipu_4v"
	"veloera/relay/constant"
)

func GetAdaptor(apiType int) channel.Adaptor {
	switch apiType {
	case constant.APITypeAli:
		return &ali.Adaptor{}
	case constant.APITypeAnthropic:
		return &claude.Adaptor{}
	case constant.APITypeBaidu:
		return &baidu.Adaptor{}
	case constant.APITypeGemini:
		return &gemini.Adaptor{}
	case constant.APITypeOpenAI:
		return &openai.Adaptor{}
	case constant.APITypePaLM:
		return &palm.Adaptor{}
	case constant.APITypeTencent:
		return &tencent.Adaptor{}
	case constant.APITypeXunfei:
		return &xunfei.Adaptor{}
	case constant.APITypeZhipu:
		return &zhipu.Adaptor{}
	case constant.APITypeZhipuV4:
		return &zhipu_4v.Adaptor{}
	case constant.APITypeOllama:
		return &ollama.Adaptor{}
	case constant.APITypePerplexity:
		return &perplexity.Adaptor{}
	case constant.APITypeAws:
		return &aws.Adaptor{}
	case constant.APITypeCohere:
		return &cohere.Adaptor{}
	case constant.APITypeDify:
		return &dify.Adaptor{}
	case constant.APITypeJina:
		return &jina.Adaptor{}
	case constant.APITypeCloudflare:
		return &cloudflare.Adaptor{}
	case constant.APITypeSiliconFlow:
		return &siliconflow.Adaptor{}
	case constant.APITypeVertexAi:
		return &vertex.Adaptor{}
	case constant.APITypeMistral:
		return &mistral.Adaptor{}
	case constant.APITypeDeepSeek:
		return &deepseek.Adaptor{}
	case constant.APITypeMokaAI:
		return &mokaai.Adaptor{}
	case constant.APITypeVolcEngine:
		return &volcengine.Adaptor{}
	case constant.APITypeBaiduV2:
		return &baidu_v2.Adaptor{}
	case constant.APITypeOpenRouter:
		return &openai.Adaptor{}
	case constant.APITypeXinference:
		return &openai.Adaptor{}
	case constant.APITypeGitHub:
		return &openai.Adaptor{}
	case constant.APITypeXai:
		return &xai.Adaptor{}
	}
	return nil
}

func GetTaskAdaptor(platform commonconstant.TaskPlatform) channel.TaskAdaptor {
	switch platform {
	//case constant.APITypeAIProxyLibrary:
	//	return &aiproxy.Adaptor{}
	case commonconstant.TaskPlatformSuno:
		return &suno.TaskAdaptor{}
	}
	return nil
}
