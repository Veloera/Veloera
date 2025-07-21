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
package xai

var ModelList = []string{
	// grok-3
	"grok-3-beta", "grok-3-mini-beta",
	// grok-3 mini
	"grok-3-fast-beta", "grok-3-mini-fast-beta",
	// extend grok-3-mini reasoning
	"grok-3-mini-beta-high", "grok-3-mini-beta-low", "grok-3-mini-beta-medium",
	"grok-3-mini-fast-beta-high", "grok-3-mini-fast-beta-low", "grok-3-mini-fast-beta-medium",
	// image model
	"grok-2-image",
	// legacy models
	"grok-2", "grok-2-vision",
	"grok-beta", "grok-vision-beta",
}

var ChannelName = "xai"
