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
package constant

import (
	"veloera/common"
)

const (
	APITypeOpenAI = iota
	APITypeAnthropic
	APITypePaLM
	APITypeBaidu
	APITypeZhipu
	APITypeAli
	APITypeXunfei
	APITypeAIProxyLibrary
	APITypeTencent
	APITypeGemini
	APITypeOllama
	APITypePerplexity
	APITypeAws
	APITypeCohere
	APITypeDify
	APITypeJina
	APITypeCloudflare
	APITypeSiliconFlow
	APITypeVertexAi
	APITypeMistral
	APITypeDeepSeek
	APITypeMokaAI
	APITypeVolcEngine
	APITypeBaiduV2
	APITypeOpenRouter
	APITypeXinference
	APITypeXai
	APITypeGitHub
	APITypeDummy // this one is only for count, do not add any channel after this
)

func ChannelType2APIType(channelType int) (int, bool) {
	apiType := -1
	switch channelType {
	case common.ChannelTypeOpenAI:
		apiType = APITypeOpenAI
	case common.ChannelTypeAnthropic:
		apiType = APITypeAnthropic
	case common.ChannelTypeBaidu:
		apiType = APITypeBaidu
	case common.ChannelTypePaLM:
		apiType = APITypePaLM
	case common.ChannelTypeZhipu:
		apiType = APITypeZhipu
	case common.ChannelTypeAli:
		apiType = APITypeAli
	case common.ChannelTypeXunfei:
		apiType = APITypeXunfei
	case common.ChannelTypeAIProxyLibrary:
		apiType = APITypeAIProxyLibrary
	case common.ChannelTypeTencent:
		apiType = APITypeTencent
	case common.ChannelTypeGemini:
		apiType = APITypeGemini
	case common.ChannelTypeOllama:
		apiType = APITypeOllama
	case common.ChannelTypePerplexity:
		apiType = APITypePerplexity
	case common.ChannelTypeAws:
		apiType = APITypeAws
	case common.ChannelTypeCohere:
		apiType = APITypeCohere
	case common.ChannelTypeDify:
		apiType = APITypeDify
	case common.ChannelTypeJina:
		apiType = APITypeJina
	case common.ChannelCloudflare:
		apiType = APITypeCloudflare
	case common.ChannelTypeSiliconFlow:
		apiType = APITypeSiliconFlow
	case common.ChannelTypeVertexAi:
		apiType = APITypeVertexAi
	case common.ChannelTypeMistral:
		apiType = APITypeMistral
	case common.ChannelTypeDeepSeek:
		apiType = APITypeDeepSeek
	case common.ChannelTypeMokaAI:
		apiType = APITypeMokaAI
	case common.ChannelTypeVolcEngine:
		apiType = APITypeVolcEngine
	case common.ChannelTypeBaiduV2:
		apiType = APITypeBaiduV2
	case common.ChannelTypeOpenRouter:
		apiType = APITypeOpenRouter
	case common.ChannelTypeXinference:
		apiType = APITypeXinference
	case common.ChannelTypeGitHub:
		apiType = APITypeGitHub
	case common.ChannelTypeXai:
		apiType = APITypeXai
	}
	if apiType == -1 {
		return APITypeOpenAI, false
	}
	return apiType, true
}
