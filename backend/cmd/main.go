package main

import (
	"fmt"
	"log"
	"ops-inspection/internal/config"
	"ops-inspection/internal/model"
	"ops-inspection/internal/repository"
	"ops-inspection/internal/router"
	"ops-inspection/internal/service"
	"ops-inspection/pkg/scheduler"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// 获取工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("获取工作目录失败:", err)
	}

	// 切换到 backend 目录
	backendDir := workDir + "/backend"
	if err := os.Chdir(backendDir); err != nil {
		// 如果 backend 目录不存在，尝试直接使用当前目录
		configPath := "config.yaml"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Fatal("找不到配置文件:", configPath)
		}
	} else {
		workDir = backendDir
	}

	// 加载配置
	configPath := workDir + "/config.yaml"
	if _, err := os.Stat("config.yaml"); err == nil {
		configPath = "config.yaml"
	}

	if err := config.Init(configPath); err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 初始化数据库
	if err := model.InitDB(&config.GlobalConfig.Database); err != nil {
		log.Fatal("初始化数据库失败:", err)
	}

	// 初始化定时清理任务
	reportRepo := repository.NewReportRepository(model.DB)
	ruleRepo := repository.NewRuleRepository(model.DB)
	projectRepo := repository.NewProjectRepository(model.DB)
	prometheusService := service.NewPrometheusService()
	inspectionService := service.NewInspectionService(reportRepo, ruleRepo, projectRepo, prometheusService)
	
	cleanupScheduler := scheduler.NewCleanupScheduler(inspectionService)
	cleanupScheduler.Start()
	defer cleanupScheduler.Stop()

	// 设置运行模式
	gin.SetMode(config.GlobalConfig.Server.Mode)

	// 设置路由
	r := router.SetupRouter()

	// 启动服务器
	port := config.GlobalConfig.Server.Port
	addr := fmt.Sprintf(":%d", port)

	log.Printf("服务器启动在端口 %d", port)
	if err := r.Run(addr); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}
