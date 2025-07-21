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
package service

import (
	"encoding/base64"
	"fmt"
	"io"
	"veloera/constant"
	"veloera/dto"
)

func GetFileBase64FromUrl(url string) (*dto.LocalFileData, error) {
	var maxFileSize = constant.MaxFileDownloadMB * 1024 * 1024

	resp, err := DoDownloadRequest(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Always use LimitReader to prevent oversized downloads
	fileBytes, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxFileSize+1)))
	if err != nil {
		return nil, err
	}
	// Check actual size after reading
	if len(fileBytes) > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size: %dMB", constant.MaxFileDownloadMB)
	}

	// Convert to base64
	base64Data := base64.StdEncoding.EncodeToString(fileBytes)

	return &dto.LocalFileData{
		Base64Data: base64Data,
		MimeType:   resp.Header.Get("Content-Type"),
		Size:       int64(len(fileBytes)),
	}, nil
}
