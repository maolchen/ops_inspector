package model

import (
	"time"
)

// RuleGroup 规则组表
type RuleGroup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Code        string    `gorm:"uniqueIndex;size:50;not null" json:"code"` // 唯一标识码
	Description string    `gorm:"size:500" json:"description"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"` // 排序
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (RuleGroup) TableName() string {
	return "rule_groups"
}
