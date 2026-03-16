package model

import (
	"time"
)

// InspectionReport 巡检报告表
type InspectionReport struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProjectID     uint      `gorm:"index;not null" json:"project_id"`
	ProjectName   string    `gorm:"size:100" json:"project_name"`
	Inspector     string    `gorm:"size:50" json:"inspector"` // 巡检人
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `gorm:"size:20" json:"status"` // running, completed, failed
	TotalItems    int       `json:"total_items"`           // 总巡检项
	WarningCount  int       `json:"warning_count"`         // 告警数量
	CriticalCount int       `json:"critical_count"`        // 严重数量
	Summary       string    `gorm:"type:text" json:"summary"` // 巡检总结
	Remark        string    `gorm:"type:text" json:"remark"`  // 备注
	CreatedAt     time.Time `json:"created_at"`

	// 关联
	Items []InspectionItem `gorm:"foreignKey:ReportID" json:"items,omitempty"`
}

func (InspectionReport) TableName() string {
	return "inspection_reports"
}

// InspectionItem 巡检项详情表
type InspectionItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ReportID    uint      `gorm:"index;not null" json:"report_id"`
	RuleID      uint      `json:"rule_id"`
	GroupID     uint      `json:"group_id"`
	GroupName   string    `gorm:"size:100" json:"group_name"`
	RuleName    string    `gorm:"size:100" json:"rule_name"`
	Instance    string    `gorm:"size:100" json:"instance"` // 实例标识
	Value       float64   `json:"value"`
	Status      string    `gorm:"size:20" json:"status"` // normal, warning, critical
	ShowInTable bool      `json:"show_in_table"`
	TrendData   string    `gorm:"type:text" json:"trend_data"` // JSON 格式的趋势数据
	Labels      string    `gorm:"type:text" json:"labels"`     // JSON 格式的标签数据
	Unit        string    `gorm:"size:20" json:"unit"`
	CreatedAt   time.Time `json:"created_at"`
}

func (InspectionItem) TableName() string {
	return "inspection_items"
}
