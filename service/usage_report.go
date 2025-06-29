package service

import (
	"time"
	"veloera/model"
)

type Overview struct {
	TotalTokens   int64 `json:"total_tokens"`
	TotalRequests int64 `json:"total_requests"`
	Error429      int64 `json:"error_429"`
	NormalError   int64 `json:"normal_error"`
}

func GenerateUsageReport(options []string) (map[string]interface{}, error) {
	start := time.Now().Add(-24 * time.Hour).Unix()
	db := model.LOG_DB
	result := make(map[string]interface{})

	var ov Overview
	db.Table("logs").
		Select("COALESCE(SUM(prompt_tokens+completion_tokens),0) as total_tokens, COUNT(id) as total_requests, "+
			"SUM(CASE WHEN (type = 5 OR (type = 2 AND completion_tokens = 0 AND content LIKE '%超时%')) AND content LIKE '%429%' THEN 1 ELSE 0 END) as error_429, "+
			"SUM(CASE WHEN (type = 5 OR (type = 2 AND completion_tokens = 0 AND content LIKE '%超时%')) AND content NOT LIKE '%429%' THEN 1 ELSE 0 END) as normal_error").
		Where("created_at >= ?", start).
		Scan(&ov)
	result["overview"] = ov

	for _, op := range options {
		switch op {
		case "channel":
			var items []struct {
				ChannelID    int
				ChannelName  string
				RequestCount int64
				TotalTokens  int64
			}
			db.Table("logs l").Select("l.channel_id as channel_id, c.name as channel_name, COUNT(l.id) as request_count, COALESCE(SUM(l.prompt_tokens + l.completion_tokens),0) as total_tokens").
				Joins("LEFT JOIN channels c ON l.channel_id = c.id").
				Where("l.created_at >= ? AND l.channel_id != 0", start).
				Group("l.channel_id, c.name").Order("request_count desc").Limit(5).Scan(&items)
			result["channel"] = items
		case "user":
			var items []struct {
				UserID       int
				Username     string
				RequestCount int64
				TotalTokens  int64
			}
			db.Table("logs").Select("user_id, username, COUNT(id) as request_count, COALESCE(SUM(prompt_tokens + completion_tokens),0) as total_tokens").
				Where("created_at >= ?", start).
				Group("user_id, username").Order("request_count desc").Limit(5).Scan(&items)
			result["user"] = items
		case "token":
			var items []struct {
				TokenID      int
				TokenName    string
				RequestCount int64
				TotalTokens  int64
			}
			db.Table("logs").Select("token_id, token_name, COUNT(id) as request_count, COALESCE(SUM(prompt_tokens + completion_tokens),0) as total_tokens").
				Where("created_at >= ? AND token_name != ''", start).
				Group("token_id, token_name").Order("request_count desc").Limit(5).Scan(&items)
			result["token"] = items
		case "model":
			var items []struct {
				ModelName    string
				RequestCount int64
				TotalTokens  int64
			}
			db.Table("logs").Select("model_name, COUNT(id) as request_count, COALESCE(SUM(prompt_tokens + completion_tokens),0) as total_tokens").
				Where("created_at >= ? AND model_name != ''", start).
				Group("model_name").Order("request_count desc").Limit(5).Scan(&items)
			result["model"] = items
		case "ip":
			var items []struct {
				IP           string
				RequestCount int64
				TotalTokens  int64
			}
			db.Table("logs").Select("ip, COUNT(id) as request_count, COALESCE(SUM(prompt_tokens + completion_tokens),0) as total_tokens").
				Where("created_at >= ? AND ip != ''", start).
				Group("ip").Order("request_count desc").Limit(5).Scan(&items)
			result["ip"] = items
		}
	}
	return result, nil
}
