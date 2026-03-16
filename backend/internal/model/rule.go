package model

import (
	"time"
)

// ThresholdType 阈值类型常量
const (
	ThresholdGreater      = "greater"        // 值 > 阈值 = critical
	ThresholdGreaterEqual = "greater_equal"  // 值 >= 阈值 = critical
	ThresholdLess         = "less"           // 值 < 阈值 = normal
	ThresholdLessEqual    = "less_equal"     // 值 <= 阈值 = normal
	ThresholdEqual        = "equal"          // 值 == 阈值 = normal
	ThresholdAtLeast      = "at_least"       // 值 >= 阈值 = normal
)

// Rule 巡检规则表
type Rule struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	GroupID       uint      `gorm:"index;not null" json:"group_id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Type          bool      `gorm:"default:true" json:"type"` // true=告警规则, false=仅展示
	ShowInTable   bool      `gorm:"default:false" json:"show_in_table"`
	Description   string    `gorm:"size:500" json:"description"`
	Query         string    `gorm:"type:text;not null" json:"query"`          // 即时查询 PromQL
	TrendQuery    string    `gorm:"type:text" json:"trend_query"`             // 趋势查询 PromQL
	Threshold     *float64  `json:"threshold"`                                // 阈值，可为空
	Unit          string    `gorm:"size:20" json:"unit"`                      // 单位
	Labels        string    `gorm:"type:text" json:"labels"`                  // JSON 格式的标签别名映射
	ThresholdType string    `gorm:"size:20;default:'greater'" json:"threshold_type"`
	ProjectScope  string    `gorm:"size:100;default:'*'" json:"project_scope"` // 适用项目，* 表示全部
	Enabled       bool      `gorm:"default:true" json:"enabled"`
	SortOrder     int       `gorm:"default:0" json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联
	Group *RuleGroup `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}

func (Rule) TableName() string {
	return "rules"
}
