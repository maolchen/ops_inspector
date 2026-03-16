package model

import (
	"time"
)

// User 用户表
type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password    string    `gorm:"size:255;not null" json:"-"` // 密码哈希，不返回给前端
	DisplayName string    `gorm:"size:100" json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
