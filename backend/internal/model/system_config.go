package model

import "gorm.io/gorm"

// SystemConfig 系统配置表
type SystemConfig struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Key       string         `gorm:"uniqueIndex;size:100;not null" json:"key"`   // 配置键
	Value     string         `gorm:"size:500" json:"value"`                       // 配置值
	Comment   string         `gorm:"size:200" json:"comment"`                     // 配置说明
	CreatedAt int64          `json:"created_at"`
	UpdatedAt int64          `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 默认配置键
const (
	ConfigReportRetentionDays = "report_retention_days" // 报告保留天数
)

// 默认配置值
var defaultConfigs = map[string]struct {
	Value   string
	Comment string
}{
	ConfigReportRetentionDays: {
		Value:   "30",
		Comment: "历史报告保留天数，超过此天数的报告将被自动删除，0表示不删除",
	},
}

// InitDefaultConfigs 初始化默认系统配置
func InitDefaultConfigs() error {
	for key, cfg := range defaultConfigs {
		var count int64
		DB.Model(&SystemConfig{}).Where("key = ?", key).Count(&count)
		if count == 0 {
			config := SystemConfig{
				Key:     key,
				Value:   cfg.Value,
				Comment: cfg.Comment,
			}
			if err := DB.Create(&config).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// GetConfigValue 获取配置值
func GetConfigValue(key string) string {
	var config SystemConfig
	if err := DB.Where("key = ?", key).First(&config).Error; err != nil {
		return ""
	}
	return config.Value
}

// SetConfigValue 设置配置值
func SetConfigValue(key, value string) error {
	return DB.Model(&SystemConfig{}).Where("key = ?", key).Update("value", value).Error
}
