package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"veloera/common"
	"veloera/setting"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	ModelRequestRateLimitCountMark        = "MRRL"
	ModelRequestRateLimitSuccessCountMark = "MRRLS"
)

// 检查Redis中的请求限制
func checkRedisRateLimit(ctx context.Context, rdb *redis.Client, key string, maxCount int, duration int64) (bool, error) {
	// 如果maxCount为0，表示不限制
	if maxCount == 0 {
		return true, nil
	}

	// 获取当前计数
	length, err := rdb.LLen(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// 如果未达到限制，允许请求
	if length < int64(maxCount) {
		return true, nil
	}

	// 检查时间窗口
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
	// 如果在时间窗口内已达到限制，拒绝请求
	subTime := nowTime.Sub(oldTime).Seconds()
	if int64(subTime) < duration {
		rdb.Expire(ctx, key, time.Duration(setting.ModelRequestRateLimitDurationMinutes)*time.Minute)
		return false, nil
	}

	return true, nil
}

// 记录Redis请求
func recordRedisRequest(ctx context.Context, rdb *redis.Client, key string, maxCount int) {
	// 如果maxCount为0，不记录请求
	if maxCount == 0 {
		return
	}

	now := time.Now().Format(timeFormat)
	rdb.LPush(ctx, key, now)
	rdb.LTrim(ctx, key, 0, int64(maxCount-1))
	rdb.Expire(ctx, key, time.Duration(setting.ModelRequestRateLimitDurationMinutes)*time.Minute)
}

// Redis限流处理器
func redisRateLimitHandler(duration int64, totalMaxCount, successMaxCount int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := strconv.Itoa(c.GetInt("id"))
		tokenId := strconv.Itoa(c.GetInt("token_id"))
		ip := c.ClientIP()
		ctx := context.Background()
		rdb := common.RDB

		// 1. 检查总请求数限制（当totalMaxCount为0时会自动跳过）
		totalKeys := []string{
			fmt.Sprintf("rateLimit:%s:user:%s", ModelRequestRateLimitCountMark, userId),
			fmt.Sprintf("rateLimit:%s:ip:%s", ModelRequestRateLimitCountMark, ip),
			fmt.Sprintf("rateLimit:%s:token:%s", ModelRequestRateLimitCountMark, tokenId),
		}
		allowed, err := true, error(nil)
		for _, tk := range totalKeys {
			allowed, err = checkRedisRateLimit(ctx, rdb, tk, totalMaxCount, duration)
			if err != nil || !allowed {
				break
			}
		}
		if err != nil {
			fmt.Println("检查总请求数限制失败:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
			return
		}
		if !allowed {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到总请求数限制：%d分钟内最多请求%d次，包括失败次数，请检查您的请求是否正确", setting.ModelRequestRateLimitDurationMinutes, totalMaxCount))
		}

		// 2. 检查成功请求数限制
		successKeys := []string{
			fmt.Sprintf("rateLimit:%s:user:%s", ModelRequestRateLimitSuccessCountMark, userId),
			fmt.Sprintf("rateLimit:%s:ip:%s", ModelRequestRateLimitSuccessCountMark, ip),
			fmt.Sprintf("rateLimit:%s:token:%s", ModelRequestRateLimitSuccessCountMark, tokenId),
		}
		for _, sk := range successKeys {
			allowed, err = checkRedisRateLimit(ctx, rdb, sk, successMaxCount, duration)
			if err != nil || !allowed {
				break
			}
		}
		if err != nil {
			fmt.Println("检查成功请求数限制失败:", err.Error())
			abortWithOpenAiMessage(c, http.StatusInternalServerError, "rate_limit_check_failed")
			return
		}
		if !allowed {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests, fmt.Sprintf("您已达到请求数限制：%d分钟内最多请求%d次", setting.ModelRequestRateLimitDurationMinutes, successMaxCount))
			return
		}

		// 3. 记录总请求（当totalMaxCount为0时会自动跳过）
		for _, tk := range totalKeys {
			recordRedisRequest(ctx, rdb, tk, totalMaxCount)
		}

		// 4. 处理请求
		c.Next()

		// 5. 如果请求成功，记录成功请求
		if c.Writer.Status() < 400 {
			for _, sk := range successKeys {
				recordRedisRequest(ctx, rdb, sk, successMaxCount)
			}
		}
	}
}

// 内存限流处理器
func memoryRateLimitHandler(duration int64, totalMaxCount, successMaxCount int) gin.HandlerFunc {
	inMemoryRateLimiter.Init(time.Duration(setting.ModelRequestRateLimitDurationMinutes) * time.Minute)

	return func(c *gin.Context) {
		userId := strconv.Itoa(c.GetInt("id"))
		tokenId := strconv.Itoa(c.GetInt("token_id"))
		ip := c.ClientIP()
		totalKeys := []string{
			ModelRequestRateLimitCountMark + "user:" + userId,
			ModelRequestRateLimitCountMark + "ip:" + ip,
			ModelRequestRateLimitCountMark + "token:" + tokenId,
		}
		successKeys := []string{
			ModelRequestRateLimitSuccessCountMark + "user:" + userId,
			ModelRequestRateLimitSuccessCountMark + "ip:" + ip,
			ModelRequestRateLimitSuccessCountMark + "token:" + tokenId,
		}

		// 1. 检查总请求数限制（当totalMaxCount为0时跳过）
		if totalMaxCount > 0 {
			for _, tk := range totalKeys {
				if !inMemoryRateLimiter.Request(tk, totalMaxCount, duration) {
					c.Status(http.StatusTooManyRequests)
					c.Abort()
					return
				}
			}
		}

		// 2. 检查成功请求数限制
		// 使用一个临时key来检查限制，这样可以避免实际记录
		checkKeys := []string{}
		for _, sk := range successKeys {
			checkKeys = append(checkKeys, sk+"_check")
		}
		for _, ck := range checkKeys {
			if !inMemoryRateLimiter.Request(ck, successMaxCount, duration) {
				c.Status(http.StatusTooManyRequests)
				c.Abort()
				return
			}
		}

		// 3. 处理请求
		c.Next()

		// 4. 如果请求成功，记录到实际的成功请求计数中
		if c.Writer.Status() < 400 {
			for _, sk := range successKeys {
				inMemoryRateLimiter.Request(sk, successMaxCount, duration)
			}
		}
	}
}

// ModelRequestRateLimit 模型请求限流中间件
func ModelRequestRateLimit() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 在每个请求时检查是否启用限流
		if !setting.ModelRequestRateLimitEnabled {
			c.Next()
			return
		}

		// 计算限流参数
		duration := int64(setting.ModelRequestRateLimitDurationMinutes * 60)
		totalMaxCount := setting.ModelRequestRateLimitCount
		successMaxCount := setting.ModelRequestRateLimitSuccessCount

		// 根据存储类型选择并执行限流处理器
		if common.RedisEnabled {
			redisRateLimitHandler(duration, totalMaxCount, successMaxCount)(c)
		} else {
			memoryRateLimitHandler(duration, totalMaxCount, successMaxCount)(c)
		}
	}
}
