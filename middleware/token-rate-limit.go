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
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"veloera/common"
	"veloera/setting/model_setting"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	TokenRateLimitCountMark        = "TRRL"
	TokenRateLimitSuccessCountMark = "TRRLS"
)

func tokenCheckRedisRateLimit(ctx context.Context, rdb *redis.Client, key string, maxCount int, duration int64) (bool, error) {
	if maxCount == 0 {
		return true, nil
	}
	length, err := rdb.LLen(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if length < int64(maxCount) {
		return true, nil
	}
	oldTimeStr, _ := rdb.LIndex(ctx, key, -1).Result()
	oldTime, err := time.Parse(timeFormat, oldTimeStr)
	if err != nil {
		return false, err
	}
	nowTimeStr := time.Now().Format(timeFormat)
	nowTime, err := time.Parse(timeFormat, nowTimeStr)
	if err != nil {
		return false, err
	}
	if int64(nowTime.Sub(oldTime).Seconds()) < duration {
		rdb.Expire(ctx, key, common.RateLimitKeyExpirationDuration)
		return false, nil
	}
	return true, nil
}

func tokenRecordRedisRequest(ctx context.Context, rdb *redis.Client, key string, maxCount int) {
	if maxCount == 0 {
		return
	}
	now := time.Now().Format(timeFormat)
	rdb.LPush(ctx, key, now)
	rdb.LTrim(ctx, key, 0, int64(maxCount-1))
	rdb.Expire(ctx, key, common.RateLimitKeyExpirationDuration)
}

func tokenRedisRateLimitHandler(duration int64, totalMaxCount, successMaxCount int) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenId := strconv.Itoa(c.GetInt("token_id"))
		ctx := context.Background()
		rdb := common.RDB
		totalKey := fmt.Sprintf("rateLimit:%s:%s", TokenRateLimitCountMark, tokenId)
		allowed, err := tokenCheckRedisRateLimit(ctx, rdb, totalKey, totalMaxCount, duration)
		if err != nil {
			fmt.Println("check token rate limit failed:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "token_rate_limit_check_failed")
			return
		}
		if !allowed {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "Key-level rate limit exceed.")
			return
		}
		successKey := fmt.Sprintf("rateLimit:%s:%s", TokenRateLimitSuccessCountMark, tokenId)
		allowed, err = tokenCheckRedisRateLimit(ctx, rdb, successKey, successMaxCount, duration)
		if err != nil {
			fmt.Println("check token rate limit failed:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "token_rate_limit_check_failed")
			return
		}
		if !allowed {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, "Key-level rate limit exceed.")
			return
		}
		tokenRecordRedisRequest(ctx, rdb, totalKey, totalMaxCount)
		c.Next()
		if c.Writer.Status() < 400 {
			tokenRecordRedisRequest(ctx, rdb, successKey, successMaxCount)
		}
	}
}

func tokenMemoryRateLimitHandler(duration int64, totalMaxCount, successMaxCount int) gin.HandlerFunc {
	inMemoryRateLimiter.Init(common.RateLimitKeyExpirationDuration)
	return func(c *gin.Context) {
		tokenId := strconv.Itoa(c.GetInt("token_id"))
		totalKey := TokenRateLimitCountMark + tokenId
		successKey := TokenRateLimitSuccessCountMark + tokenId
		if totalMaxCount > 0 && !inMemoryRateLimiter.Request(totalKey, totalMaxCount, duration) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}
		if !inMemoryRateLimiter.Request(successKey, successMaxCount, duration) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}
		c.Next()
		if c.Writer.Status() < 400 {
			inMemoryRateLimiter.Request(successKey, successMaxCount, duration)
		}
	}
}

func TokenRateLimit() func(c *gin.Context) {
	return func(c *gin.Context) {
		if !c.GetBool("token_rate_limit_enabled") {
			c.Next()
			return
		}
		tokenGroup := c.GetString("token_group")
		if model_setting.ShouldBypassRateLimit(tokenGroup) {
			c.Next()
			return
		}
		duration := int64(c.GetInt("token_rate_limit_period"))
		totalMaxCount := c.GetInt("token_rate_limit_count")
		successMaxCount := c.GetInt("token_rate_limit_success")
		if common.RedisEnabled {
			tokenRedisRateLimitHandler(duration, totalMaxCount, successMaxCount)(c)
		} else {
			tokenMemoryRateLimitHandler(duration, totalMaxCount, successMaxCount)(c)
		}
	}
}
