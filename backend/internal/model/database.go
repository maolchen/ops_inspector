package model

import (
	"fmt"
	"ops-inspection/internal/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库
func InitDB(cfg *config.DatabaseConfig) error {
	var err error

	// 配置 GORM 日志
	logLevel := logger.Info
	if config.GlobalConfig.Server.Mode == "release" {
		logLevel = logger.Warn
	}

	// 连接数据库
	DB, err = gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移
	if err = autoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 初始化默认数据
	if err = initDefaultData(); err != nil {
		return fmt.Errorf("failed to init default data: %w", err)
	}

	return nil
}

// autoMigrate 自动迁移表结构
func autoMigrate() error {
	return DB.AutoMigrate(
		&Project{},
		&RuleGroup{},
		&Rule{},
		&InspectionReport{},
		&InspectionItem{},
	)
}

// initDefaultData 初始化默认数据
func initDefaultData() error {
	// 检查是否已有规则组
	var count int64
	DB.Model(&RuleGroup{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 创建默认规则组
	defaultGroups := []RuleGroup{
		{Name: "基础资源组", Code: "basic_resources", Description: "服务器基础资源监控", SortOrder: 1},
		{Name: "K8S容器组", Code: "k8s_container", Description: "Kubernetes容器资源监控", SortOrder: 2},
		{Name: "进程资源组", Code: "process_resources", Description: "进程资源监控", SortOrder: 3},
		{Name: "其他分组", Code: "others", Description: "其他自定义指标", SortOrder: 4},
	}

	return DB.Create(&defaultGroups).Error
}
