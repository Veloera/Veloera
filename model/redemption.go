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
package model

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
	"veloera/common"

	"gorm.io/gorm"
)

type Redemption struct {
	Id           int            `json:"id"`
	UserId       int            `json:"user_id"`
	Key          string         `json:"key" gorm:"type:char(32);uniqueIndex"`
	Status       int            `json:"status" gorm:"default:1"`
	Name         string         `json:"name" gorm:"index"`
	Quota        int            `json:"quota" gorm:"default:100"`
	CreatedTime  int64          `json:"created_time" gorm:"bigint"`
	RedeemedTime int64          `json:"redeemed_time" gorm:"bigint"`
	ValidFrom    int64          `json:"valid_from" gorm:"bigint;default:0"`    // 生效时间，0表示立即生效
	ValidUntil   int64          `json:"valid_until" gorm:"bigint;default:0"`   // 过期时间，0表示永不过期
	Count        int            `json:"count" gorm:"-:all"` // only for api request
	UsedUserId   int            `json:"used_user_id"`
	IsGift       bool           `json:"is_gift" gorm:"default:false"`
	MaxUses      int            `json:"max_uses" gorm:"default:-1"` // -1 means unlimited
	UsedCount    int            `json:"used_count" gorm:"default:0"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// RedemptionLog 记录礼品码的使用记录
type RedemptionLog struct {
	Id           int   `json:"id"`
	RedemptionId int   `json:"redemption_id" gorm:"index"`
	UserId       int   `json:"user_id" gorm:"index"`
	UsedTime     int64 `json:"used_time" gorm:"bigint"`
}

func GetAllRedemptions(startIdx int, num int) (redemptions []*Redemption, total int64, err error) {
	// 开始事务
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取总数
	err = tx.Model(&Redemption{}).Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// 获取分页数据
	err = tx.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return redemptions, total, nil
}

func SearchRedemptions(keyword string, startIdx int, num int) (redemptions []*Redemption, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Build query based on keyword type
	query := tx.Model(&Redemption{})

	// Only try to convert to ID if the string represents a valid integer
	if id, err := strconv.Atoi(keyword); err == nil {
		query = query.Where("id = ? OR name LIKE ?", id, keyword+"%")
	} else {
		query = query.Where("name LIKE ?", keyword+"%")
	}

	// Get total count
	err = query.Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Get paginated data
	err = query.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return redemptions, total, nil
}

func GetRedemptionById(id int) (*Redemption, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	redemption := Redemption{Id: id}
	var err error = nil
	err = DB.First(&redemption, "id = ?", id).Error
	return &redemption, err
}

// GetRedemptionByKey 根据兑换码内容获取兑换码
func GetRedemptionByKey(key string) (*Redemption, error) {
	if key == "" {
		return nil, errors.New("兑换码内容为空！")
	}
	redemption := Redemption{}
	var err error = nil
	err = DB.First(&redemption, "`key` = ?", key).Error
	return &redemption, err
}

func Redeem(key string, userId int) (quota int, isGift bool, err error) {
	if key == "" {
		return 0, false, errors.New("未提供兑换码")
	}
	if userId == 0 {
		return 0, false, errors.New("无效的 user id")
	}
	redemption := &Redemption{}

	keyCol := "`key`"
	if common.UsingPostgreSQL {
		keyCol = `"key"`
	}

	if common.RedisEnabled && common.RDB != nil {
		lockKey := "rloc:" + key // Using a shorter prefix "rloc:" for brevity
		locked, redisErr := common.RDB.SetNX(context.Background(), lockKey, "1", 10*time.Second).Result()
		if redisErr != nil {
			common.SysError("Redis lock acquisition error: " + redisErr.Error())
			return 0, false, errors.New("系统暂时繁忙，请稍后再试") // System temporarily busy, please try again later
		}
		if !locked {
			return 0, false, errors.New("操作过于频繁，请稍后再试") // Operation too frequent, please try again later
		}
		defer common.RDB.Del(context.Background(), lockKey)
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where(keyCol+" = ?", key).First(redemption).Error
		if err != nil {
			return errors.New("无效的兑换码")
		}

		// 检查兑换码时间范围
		currentTime := common.GetTimestamp()
		if redemption.ValidFrom > 0 && currentTime < redemption.ValidFrom {
			return errors.New("兑换码尚未生效")
		}
		if redemption.ValidUntil > 0 && currentTime > redemption.ValidUntil {
			return errors.New("兑换码已过期")
		}

		if !redemption.IsGift {
			// 普通兑换码逻辑
			if redemption.Status != common.RedemptionCodeStatusEnabled {
				return errors.New("该兑换码已被使用")
			}
			err = tx.Model(&User{}).Where("id = ?", userId).Update("quota", gorm.Expr("quota + ?", redemption.Quota)).Error
			if err != nil {
				return err
			}
			redemption.RedeemedTime = common.GetTimestamp()
			redemption.Status = common.RedemptionCodeStatusUsed
			redemption.UsedUserId = userId
		} else {
			// 礼品码逻辑
			if redemption.MaxUses != -1 && redemption.UsedCount >= redemption.MaxUses {
				return errors.New("该礼品码已达到最大使用次数")
			}
			// 检查用户是否已经使用过这个礼品码
			var usageCount int64
			err = tx.Model(&RedemptionLog{}).Where("redemption_id = ? AND user_id = ?", redemption.Id, userId).Count(&usageCount).Error
			if err != nil {
				return err
			}
			if usageCount > 0 {
				return errors.New("您已经使用过这个礼品码")
			}

			err = tx.Model(&User{}).Where("id = ?", userId).Update("quota", gorm.Expr("quota + ?", redemption.Quota)).Error
			if err != nil {
				return err
			}

			// 记录使用日志
			log := RedemptionLog{
				RedemptionId: redemption.Id,
				UserId:       userId,
				UsedTime:     common.GetTimestamp(),
			}
			if err = tx.Create(&log).Error; err != nil {
				return err
			}

			redemption.UsedCount++
			if redemption.MaxUses != -1 && redemption.UsedCount >= redemption.MaxUses {
				redemption.Status = common.RedemptionCodeStatusUsed
			}
		}

		err = tx.Save(redemption).Error
		return err
	})
	if err != nil {
		return 0, false, errors.New("兑换失败，" + err.Error())
	}
	RecordLog(userId, LogTypeTopup, fmt.Sprintf("通过%s充值 %s，兑换码ID %d",
		map[bool]string{true: "礼品码", false: "兑换码"}[redemption.IsGift],
		common.LogQuota(redemption.Quota),
		redemption.Id))

	// 处理返佣逻辑
	rebateType := map[bool]string{true: "礼品码", false: "兑换码"}[redemption.IsGift]
	err = ProcessRebate(userId, redemption.Quota, rebateType)
	if err != nil {
		common.SysError("处理兑换码返佣失败: " + err.Error())
	}

	return redemption.Quota, redemption.IsGift, nil
}

func (redemption *Redemption) Insert() error {
	var err error
	err = DB.Create(redemption).Error
	return err
}

func (redemption *Redemption) SelectUpdate() error {
	// This can update zero values
	return DB.Model(redemption).Select("redeemed_time", "status").Updates(redemption).Error
}

// Update Make sure your token's fields is completed, because this will update non-zero values
func (redemption *Redemption) Update() error {
	var err error
	err = DB.Model(redemption).Select("name", "status", "quota", "redeemed_time", "valid_from", "valid_until").Updates(redemption).Error
	return err
}

func (redemption *Redemption) Delete() error {
	var err error
	err = DB.Delete(redemption).Error
	return err
}

func DeleteRedemptionById(id int) (err error) {
	if id == 0 {
		return errors.New("id 为空！")
	}
	redemption := Redemption{Id: id}
	err = DB.Where(redemption).First(&redemption).Error
	if err != nil {
		return err
	}
	return redemption.Delete()
}

// CountRedemptionsByName 根据名称统计兑换码数量
func CountRedemptionsByName(name string) (count int64, err error) {
	if name == "" {
		return 0, errors.New("名称不能为空")
	}
	err = DB.Model(&Redemption{}).Where("name = ?", name).Count(&count).Error
	return count, err
}

// DeleteRedemptionsByName 根据名称批量删除兑换码
func DeleteRedemptionsByName(name string) (count int64, err error) {
	if name == "" {
		return 0, errors.New("名称不能为空")
	}

	// 先计算数量
	count, err = CountRedemptionsByName(name)
	if err != nil {
		return 0, err
	}

	// 没有找到匹配的兑换码
	if count == 0 {
		return 0, nil
	}

	// 执行删除
	result := DB.Where("name = ?", name).Delete(&Redemption{})
	return result.RowsAffected, result.Error
}

// BatchDisableRedemptions 批量禁用指定ID的兑换码
func BatchDisableRedemptions(ids []int) (count int64, err error) {
	if len(ids) == 0 {
		return 0, errors.New("ID列表不能为空")
	}

	result := DB.Model(&Redemption{}).Where("id IN ?", ids).Update("status", common.RedemptionCodeStatusDisabled)
	return result.RowsAffected, result.Error
}

// DeleteDisabledRedemptions 删除所有已禁用的兑换码
func DeleteDisabledRedemptions() (count int64, err error) {
    // Delete both disabled and already used redemption codes
    result := DB.Where("status IN ?", []int{common.RedemptionCodeStatusDisabled, common.RedemptionCodeStatusUsed}).Delete(&Redemption{})
    return result.RowsAffected, result.Error
}
