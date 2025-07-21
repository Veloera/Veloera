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
package middleware

import (
	"compress/gzip"
	"github.com/andybalholm/brotli"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func DecompressRequestMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.Method == http.MethodGet {
			c.Next()
			return
		}
		switch c.GetHeader("Content-Encoding") {
		case "gzip":
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()

			// Replace the request body with the decompressed data
			c.Request.Body = io.NopCloser(gzipReader)
			c.Request.Header.Del("Content-Encoding")
		case "br":
			reader := brotli.NewReader(c.Request.Body)
			c.Request.Body = io.NopCloser(reader)
			c.Request.Header.Del("Content-Encoding")
		}

		// Continue processing the request
		c.Next()
	}
}
