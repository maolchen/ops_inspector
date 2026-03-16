package scheduler

import (
	"log"
	"ops-inspection/internal/model"
	"ops-inspection/internal/service"
	"time"
)

// CleanupScheduler 报告清理调度器
type CleanupScheduler struct {
	inspectionService *service.InspectionService
	ticker            *time.Ticker
	stopChan          chan struct{}
}

// NewCleanupScheduler 创建清理调度器
func NewCleanupScheduler(inspectionService *service.InspectionService) *CleanupScheduler {
	return &CleanupScheduler{
		inspectionService: inspectionService,
		stopChan:          make(chan struct{}),
	}
}

// Start 启动定时清理任务
func (s *CleanupScheduler) Start() {
	// 每天凌晨3点执行一次清理
	s.ticker = time.NewTicker(24 * time.Hour)

	// 启动时先执行一次检查
	s.doCleanup()

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.doCleanup()
			case <-s.stopChan:
				return
			}
		}
	}()

	log.Println("报告清理定时任务已启动")
}

// Stop 停止定时清理任务
func (s *CleanupScheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopChan)
	log.Println("报告清理定时任务已停止")
}

// doCleanup 执行清理
func (s *CleanupScheduler) doCleanup() {
	// 获取保留天数配置
	retentionDays := s.getRetentionDays()
	if retentionDays <= 0 {
		log.Println("报告保留天数为0，跳过清理")
		return
	}

	log.Printf("开始清理%d天前的历史报告...\n", retentionDays)
	count, err := s.inspectionService.CleanupOldReports(retentionDays)
	if err != nil {
		log.Printf("清理历史报告失败: %v\n", err)
		return
	}

	if count > 0 {
		log.Printf("成功清理%d条历史报告\n", count)
	}
}

// getRetentionDays 获取保留天数配置
func (s *CleanupScheduler) getRetentionDays() int {
	value := model.GetConfigValue(model.ConfigReportRetentionDays)
	if value == "" {
		return 30 // 默认30天
	}

	days := 0
	if _, err := time.ParseDuration(value + "h"); err == nil {
		// 如果是小时格式，转换为天数
	} else {
		// 尝试解析为天数
		var d int
		if _, err := time.ParseDuration(value + "h"); err != nil {
			// 直接解析数字
			for i, c := range value {
				if c >= '0' && c <= '9' {
					d = d*10 + int(c-'0')
				} else {
					break
				}
				if i > 10 {
					break
				}
			}
		}
		days = d
	}

	return days
}
