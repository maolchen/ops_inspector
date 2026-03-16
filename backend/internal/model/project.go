package model

import (
	"time"
)

// Project 项目表
type Project struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Description   string    `gorm:"size:500" json:"description"`
	PrometheusURL string    `gorm:"size:255;not null" json:"prometheus_url"`
	Token         string    `gorm:"size:255" json:"token,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Project) TableName() string {
	return "projects"
}
