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
	"encoding/json"
)

var TopupGroupRatio = map[string]float64{
	"default": 1,
	"vip":     1,
	"svip":    1,
}

func TopupGroupRatio2JSONString() string {
	jsonBytes, err := json.Marshal(TopupGroupRatio)
	if err != nil {
		SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateTopupGroupRatioByJSONString(jsonStr string) error {
	TopupGroupRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &TopupGroupRatio)
}

func GetTopupGroupRatio(name string) float64 {
	ratio, ok := TopupGroupRatio[name]
	if !ok {
		SysError("topup group ratio not found: " + name)
		return 1
	}
	return ratio
}
