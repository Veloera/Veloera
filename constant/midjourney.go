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

const (
	MjErrorUnknown = 5
	MjRequestError = 4
)

const (
	MjActionImagine       = "IMAGINE"
	MjActionDescribe      = "DESCRIBE"
	MjActionBlend         = "BLEND"
	MjActionUpscale       = "UPSCALE"
	MjActionVariation     = "VARIATION"
	MjActionReRoll        = "REROLL"
	MjActionInPaint       = "INPAINT"
	MjActionModal         = "MODAL"
	MjActionZoom          = "ZOOM"
	MjActionCustomZoom    = "CUSTOM_ZOOM"
	MjActionShorten       = "SHORTEN"
	MjActionHighVariation = "HIGH_VARIATION"
	MjActionLowVariation  = "LOW_VARIATION"
	MjActionPan           = "PAN"
	MjActionSwapFace      = "SWAP_FACE"
	MjActionUpload        = "UPLOAD"
)

var MidjourneyModel2Action = map[string]string{
	"mj_imagine":        MjActionImagine,
	"mj_describe":       MjActionDescribe,
	"mj_blend":          MjActionBlend,
	"mj_upscale":        MjActionUpscale,
	"mj_variation":      MjActionVariation,
	"mj_reroll":         MjActionReRoll,
	"mj_modal":          MjActionModal,
	"mj_inpaint":        MjActionInPaint,
	"mj_zoom":           MjActionZoom,
	"mj_custom_zoom":    MjActionCustomZoom,
	"mj_shorten":        MjActionShorten,
	"mj_high_variation": MjActionHighVariation,
	"mj_low_variation":  MjActionLowVariation,
	"mj_pan":            MjActionPan,
	"swap_face":         MjActionSwapFace,
	"mj_upload":         MjActionUpload,
}
