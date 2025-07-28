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
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"html/template"
	"io"
	"log"
	"math/big"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		log.Println(err)
	}
}

func GetIp() (ip string) {
	ips, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return ip
	}

	for _, a := range ips {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip = ipNet.IP.String()
				if strings.HasPrefix(ip, "10") {
					return
				}
				if strings.HasPrefix(ip, "172") {
					return
				}
				if strings.HasPrefix(ip, "192.168") {
					return
				}
				ip = ""
			}
		}
	}
	return
}

var sizeKB = 1024
var sizeMB = sizeKB * 1024
var sizeGB = sizeMB * 1024

func Bytes2Size(num int64) string {
	numStr := ""
	unit := "B"
	if num/int64(sizeGB) > 1 {
		numStr = fmt.Sprintf("%.2f", float64(num)/float64(sizeGB))
		unit = "GB"
	} else if num/int64(sizeMB) > 1 {
		numStr = fmt.Sprintf("%d", int(float64(num)/float64(sizeMB)))
		unit = "MB"
	} else if num/int64(sizeKB) > 1 {
		numStr = fmt.Sprintf("%d", int(float64(num)/float64(sizeKB)))
		unit = "KB"
	} else {
		numStr = fmt.Sprintf("%d", num)
	}
	return numStr + " " + unit
}

func Seconds2Time(num int) (time string) {
	if num/31104000 > 0 {
		time += strconv.Itoa(num/31104000) + " 年 "
		num %= 31104000
	}
	if num/2592000 > 0 {
		time += strconv.Itoa(num/2592000) + " 个月 "
		num %= 2592000
	}
	if num/86400 > 0 {
		time += strconv.Itoa(num/86400) + " 天 "
		num %= 86400
	}
	if num/3600 > 0 {
		time += strconv.Itoa(num/3600) + " 小时 "
		num %= 3600
	}
	if num/60 > 0 {
		time += strconv.Itoa(num/60) + " 分钟 "
		num %= 60
	}
	time += strconv.Itoa(num) + " 秒"
	return
}

func Interface2String(inter interface{}) string {
	switch inter.(type) {
	case string:
		return inter.(string)
	case int:
		return fmt.Sprintf("%d", inter.(int))
	case float64:
		return fmt.Sprintf("%f", inter.(float64))
	}
	return "Not Implemented"
}

func UnescapeHTML(x string) interface{} {
	return template.HTML(x)
}

func IntMax(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func IsIP(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil
}

func GetUUID() string {
	code := uuid.New().String()
	code = strings.Replace(code, "-", "", -1)
	return code
}

const keyChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GenerateRandomCharsKey(length int) (string, error) {
	b := make([]byte, length)
	maxI := big.NewInt(int64(len(keyChars)))

	for i := range b {
		n, err := crand.Int(crand.Reader, maxI)
		if err != nil {
			return "", err
		}
		b[i] = keyChars[n.Int64()]
	}

	return string(b), nil
}

func GenerateRandomKey(length int) (string, error) {
	bytes := make([]byte, length*3/4) // 对于48位的输出，这里应该是36
	if _, err := crand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func GenerateKey() (string, error) {
	//rand.Seed(time.Now().UnixNano())
	return GenerateRandomCharsKey(48)
}

func GetRandomInt(max int) int {
	//rand.Seed(time.Now().UnixNano())
	return rand.Intn(max)
}

// RandomInt returns a random integer in the range [0, max)
func RandomInt(max int) int {
	return rand.Intn(max)
}

func GetTimestamp() int64 {
	return time.Now().Unix()
}

func GetTimeString() string {
	now := time.Now()
	return fmt.Sprintf("%s%d", now.Format("20060102150405"), now.UnixNano()%1e9)
}

func Max(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func MessageWithRequestId(message string, id string) string {
	return fmt.Sprintf("%s (request id: %s)", message, id)
}

func RandomSleep() {
	// Sleep for 0-3000 ms
	time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
}

func GetPointer[T any](v T) *T {
	return &v
}

func Any2Type[T any](data any) (T, error) {
	var zero T
	bytes, err := json.Marshal(data)
	if err != nil {
		return zero, err
	}
	var res T
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return zero, err
	}
	return res, nil
}

// SaveTmpFile saves data to a temporary file. The filename would be apppended with a random string.
func SaveTmpFile(filename string, data io.Reader) (string, error) {
	f, err := os.CreateTemp(os.TempDir(), filename)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create temporary file %s", filename)
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	if err != nil {
		return "", errors.Wrapf(err, "failed to copy data to temporary file %s", filename)
	}

	return f.Name(), nil
}

// GetAudioDuration returns the duration of an audio file in seconds.
func GetAudioDuration(ctx context.Context, filename string) (float64, error) {
	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 {{input}}
	c := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filename)
	output, err := c.Output()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get audio duration")
	}

	return strconv.ParseFloat(string(bytes.TrimSpace(output)), 64)
}

// GetClientIP detects the client IP address based on reverse proxy configuration
// This function handles different proxy configurations (Cloudflare, Nginx) and falls back to direct connection IP
func GetClientIP(c *gin.Context) string {
	// If reverse proxy is not enabled, use direct connection IP
	if !ReverseProxyEnabled {
		return c.ClientIP()
	}

	var ip string
	
	// Handle different proxy providers
	switch ReverseProxyProvider {
	case "cloudflare":
		// Cloudflare uses CF-Connecting-IP header
		ip = c.GetHeader("CF-Connecting-IP")
		if ip != "" && IsValidIP(ip) {
			return SanitizeIP(ip)
		}
	case "nginx":
		// Nginx/OpenResty typically uses X-Real-IP header first
		ip = c.GetHeader("X-Real-IP")
		if ip != "" && IsValidIP(ip) {
			return SanitizeIP(ip)
		}
		
		// Fallback to X-Forwarded-For header (get first IP in chain)
		forwardedFor := c.GetHeader("X-Forwarded-For")
		if forwardedFor != "" {
			// X-Forwarded-For can contain multiple IPs separated by commas
			// The first IP is typically the original client IP
			ips := strings.Split(forwardedFor, ",")
			if len(ips) > 0 {
				firstIP := strings.TrimSpace(ips[0])
				if IsValidIP(firstIP) {
					return SanitizeIP(firstIP)
				}
			}
		}
	}
	
	// Fallback to connection IP if proxy headers are missing or invalid
	return c.ClientIP()
}

// DetectProxyHeaders analyzes request headers to determine proxy configuration
// Returns the detected provider and the IP address found
func DetectProxyHeaders(c *gin.Context) (provider string, ip string) {
	// Check for Cloudflare headers first
	if cfIP := c.GetHeader("CF-Connecting-IP"); cfIP != "" && IsValidIP(cfIP) {
		return "cloudflare", SanitizeIP(cfIP)
	}
	
	// Check for Nginx/OpenResty headers
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" && IsValidIP(realIP) {
		return "nginx", SanitizeIP(realIP)
	}
	
	// Check X-Forwarded-For as secondary option for Nginx
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			firstIP := strings.TrimSpace(ips[0])
			if IsValidIP(firstIP) {
				return "nginx", SanitizeIP(firstIP)
			}
		}
	}
	
	// No proxy headers detected
	return "", ""
}

// IsValidIP validates if a string is a valid IP address (IPv4 or IPv6)
func IsValidIP(ip string) bool {
	if ip == "" {
		return false
	}
	
	// Remove any surrounding whitespace
	ip = strings.TrimSpace(ip)
	
	// Use net.ParseIP to validate the IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// Additional validation: reject obviously invalid IPs
	// Reject localhost/loopback addresses in proxy context
	if parsedIP.IsLoopback() {
		return false
	}
	
	// Reject unspecified addresses (0.0.0.0 or ::)
	if parsedIP.IsUnspecified() {
		return false
	}
	
	return true
}

// SanitizeIP cleans and validates an IP address string
func SanitizeIP(ip string) string {
	if ip == "" {
		return ""
	}
	
	// Remove surrounding whitespace
	ip = strings.TrimSpace(ip)
	
	// Parse and reformat the IP to ensure it's in standard format
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ""
	}
	
	// Return the standardized string representation
	return parsedIP.String()
}
