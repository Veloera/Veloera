package model

import (
	"encoding/json"
	"veloera/common"
)

type UsageReport struct {
	Id        int    `json:"id"`
	Name      string `json:"name" gorm:"size:128"`
	CreatedAt int64  `json:"created_at" gorm:"bigint"`
	Data      string `json:"data" gorm:"type:text"`
}

func CreateUsageReport(report *UsageReport) error {
	return DB.Create(report).Error
}

func GetAllUsageReports() ([]*UsageReport, error) {
	var reports []*UsageReport
	err := DB.Order("id desc").Find(&reports).Error
	return reports, err
}

func GetUsageReport(id int) (*UsageReport, error) {
	var r UsageReport
	err := DB.First(&r, id).Error
	return &r, err
}

// Simple statistics example
func GenerateUsageReport(name string) (*UsageReport, error) {
	stat := struct {
		TotalRequests int64 `json:"total_requests"`
		TotalTokens   int64 `json:"total_tokens"`
	}{}
	err := LOG_DB.Model(&Log{}).Count(&stat.TotalRequests).Error
	if err != nil {
		return nil, err
	}
	err = LOG_DB.Model(&Log{}).Select("COALESCE(SUM(prompt_tokens+completion_tokens),0)").Scan(&stat.TotalTokens).Error
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(stat)
	data := string(b)
	report := &UsageReport{
		Name:      name,
		CreatedAt: common.GetTimestamp(),
		Data:      data,
	}
	return report, CreateUsageReport(report)
}
