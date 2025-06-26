package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
	"veloera/common"
	"veloera/model"
)

func TokenRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenId := c.GetInt("token_id")
		if tokenId == 0 {
			c.Next()
			return
		}

		token, err := model.GetTokenById(tokenId)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		if !token.RateLimitEnabled {
			c.Next()
			return
		}

		// Key for Redis
		key := fmt.Sprintf("token_rate_limit:%d", tokenId)

		// Get current counts from Redis
		currentRequestsStr, _ := common.RedisGet(key + ":requests")
		currentRequests, _ := strconv.ParseInt(currentRequestsStr, 10, 64)
		currentSuccessfulRequestsStr, _ := common.RedisGet(key + ":successful_requests")
		currentSuccessfulRequests, _ := strconv.ParseInt(currentSuccessfulRequestsStr, 10, 64)

		// Check if the key exists in Redis. If not, set it with expiration.
		if common.RedisExists(key + ":requests").Val() == 0 {
			common.RedisSet(key+":requests", "0", time.Duration(token.Frequency)*time.Second)
			common.RedisSet(key+":successful_requests", "0", time.Duration(token.Frequency)*time.Second)
		}

		// Increment request count
		newRequests, _ := common.RedisIncr(key + ":requests").Result()

		// Check total request limit
		if newRequests > token.Limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Key-level rate limit exceed.",
			})
			c.Abort()
			return
		}

		// Store the original writer to capture status code
		w := &common.ResponseWriter{ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		// After the request is handled, check if it was successful
		if w.Status() >= 200 && w.Status() < 300 {
			newSuccessfulRequests, _ := common.RedisIncr(key + ":successful_requests").Result()
			if newSuccessfulRequests > token.SuccessfulLimit {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"success": false,
					"message": "Key-level rate limit exceed.",
				})
				c.Abort()
				return
			}
		}
	}
}
