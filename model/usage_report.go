package model

import "veloera/common"

type UsageReport struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Options     string `json:"options"`
	Data        string `json:"data" gorm:"type:text"`
	CreatedTime int64  `json:"created_time" gorm:"bigint"`
}

func (r *UsageReport) Insert() error {
	if r.CreatedTime == 0 {
		r.CreatedTime = common.GetTimestamp()
	}
	return DB.Create(r).Error
}

func GetUsageReports(offset, limit int) (reports []UsageReport, total int64, err error) {
	err = DB.Model(&UsageReport{}).Count(&total).Error
	if err != nil {
		return
	}
	err = DB.Order("id desc").Limit(limit).Offset(offset).Find(&reports).Error
	return
}

func GetUsageReportById(id int) (report *UsageReport, err error) {
	report = &UsageReport{}
	err = DB.First(report, id).Error
	return
}
