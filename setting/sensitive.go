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
package setting

import (
	"regexp"
	"strings"
)

var CheckSensitiveEnabled = true
var CheckSensitiveOnPromptEnabled = true

var SafeCheckExemptEnabled = false
var SafeCheckExemptGroup = "nsfw-ok"

//var CheckSensitiveOnCompletionEnabled = true

// StopOnSensitiveEnabled 如果检测到敏感词，是否立刻停止生成，否则替换敏感词
var StopOnSensitiveEnabled = true

// StreamCacheQueueLength 流模式缓存队列长度，0表示无缓存
var StreamCacheQueueLength = 0

// SensitiveWords 普通屏蔽词
var SensitiveWords = []string{
	"test_sensitive",
}

// RegexSensitiveWords 正则表达式屏蔽词
var RegexSensitiveWords = []string{}

// 判断字符串是否是有效的正则表达式
func isValidRegex(pattern string) bool {
	_, err := regexp.Compile(pattern)
	return err == nil
}

func SensitiveWordsToString() string {
	var builder strings.Builder

	// Add normal words
	for _, word := range SensitiveWords {
		builder.WriteString(word)
		builder.WriteString("\n")
	}

	// Add regex patterns with "regex:" prefix
	for _, pattern := range RegexSensitiveWords {
		builder.WriteString("regex:")
		builder.WriteString(pattern)
		builder.WriteString("\n")
	}

	return builder.String()
}

func SensitiveWordsFromString(s string) {
	SensitiveWords = []string{}
	RegexSensitiveWords = []string{}

	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line starts with "regex:"
		if strings.HasPrefix(line, "regex:") {
			pattern := strings.TrimPrefix(line, "regex:")
			if isValidRegex(pattern) {
				RegexSensitiveWords = append(RegexSensitiveWords, pattern)
			}
		} else {
			SensitiveWords = append(SensitiveWords, line)
		}
	}
}

func ShouldCheckPromptSensitive() bool {
	return CheckSensitiveEnabled && CheckSensitiveOnPromptEnabled
}

func ShouldCheckPromptSensitiveWithGroup(group string) bool {
	if SafeCheckExemptEnabled && group == SafeCheckExemptGroup {
		return false
	}
	return ShouldCheckPromptSensitive()
}

//func ShouldCheckCompletionSensitive() bool {
//	return CheckSensitiveEnabled && CheckSensitiveOnCompletionEnabled
//}
