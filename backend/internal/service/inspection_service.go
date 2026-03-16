package service

import (
	"strconv"
	"time"

	"ops-inspection/internal/model"
	"ops-inspection/internal/repository"
)

type InspectionService struct {
	reportRepo  *repository.ReportRepository
	ruleRepo    *repository.RuleRepository
	projectRepo *repository.ProjectRepository
	prometheus  *PrometheusService
}

func NewInspectionService(
	reportRepo *repository.ReportRepository,
	ruleRepo *repository.RuleRepository,
	projectRepo *repository.ProjectRepository,
	prometheus *PrometheusService,
) *InspectionService {
	return &InspectionService{
		reportRepo:  reportRepo,
		ruleRepo:    ruleRepo,
		projectRepo: projectRepo,
		prometheus:  prometheus,
	}
}

// StartInspection 启动巡检
func (s *InspectionService) StartInspection(projectID uint, inspector string) (*model.InspectionReport, error) {
	// 获取项目信息
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, err
	}

	// 创建报告记录（不先保存到数据库）
	report := &model.InspectionReport{
		ProjectID:   projectID,
		ProjectName: project.Name,
		Inspector:   inspector,
		StartTime:   time.Now(),
		Status:      "running",
	}

	// 获取适用的规则
	rules, err := s.ruleRepo.GetByProjectScope(project.Name)
	if err != nil {
		return nil, err
	}

	// 执行巡检（report.ID 为 0，后续会由数据库自动生成）
	items := s.executeInspection(rules, project)

	// 更新报告统计
	report.EndTime = time.Now()
	report.Status = "completed"
	report.TotalItems = len(items)

	for _, item := range items {
		if item.Status == "critical" {
			report.CriticalCount++
		} else if item.Status == "warning" {
			report.WarningCount++
		}
	}

	// 使用事务一次性保存报告和巡检项
	if err := s.reportRepo.CreateWithItems(report, items); err != nil {
		return nil, err
	}

	return report, nil
}

// executeInspection 执行巡检采集
func (s *InspectionService) executeInspection(rules []model.Rule, project *model.Project) []model.InspectionItem {
	var items []model.InspectionItem

	for _, rule := range rules {
		// 查询 Prometheus 数据
		results, err := s.prometheus.Query(project.PrometheusURL, project.Token, rule.Query)
		if err != nil {
			continue
		}

		// 处理查询结果
		for _, result := range results {
			item := model.InspectionItem{
				RuleID:      rule.ID,
				GroupID:     rule.GroupID,
				GroupName:   rule.Group.Name,
				RuleName:    rule.Name,
				Instance:    result.Instance,
				Value:       result.Value,
				Status:      s.evaluateStatus(result.Value, rule),
				ShowInTable: rule.ShowInTable,
				Labels:      result.Labels,
				Unit:        rule.Unit,
			}

			// 获取趋势数据
			if rule.TrendQuery != "" {
				trendData, _ := s.prometheus.QueryRange(project.PrometheusURL, project.Token, rule.TrendQuery)
				if trendData != "" {
					item.TrendData = trendData
				}
			}

			items = append(items, item)
		}
	}

	return items
}

// evaluateStatus 评估状态
func (s *InspectionService) evaluateStatus(value float64, rule model.Rule) string {
	// 如果不是告警规则，直接返回正常
	if !rule.Type || rule.Threshold == nil {
		return "normal"
	}

	threshold := *rule.Threshold
	warningThreshold := threshold * 0.9

	switch rule.ThresholdType {
	case model.ThresholdGreater:
		// 值 > 阈值 = critical
		if value > threshold {
			return "critical"
		} else if value > warningThreshold {
			return "warning"
		}

	case model.ThresholdGreaterEqual:
		// 值 >= 阈值 = critical
		if value >= threshold {
			return "critical"
		} else if value >= warningThreshold {
			return "warning"
		}

	case model.ThresholdLess:
		// 值 < 阈值 = normal
		if value < threshold {
			return "normal"
		} else if value < threshold*1.1 {
			return "warning"
		}
		return "critical"

	case model.ThresholdLessEqual:
		// 值 <= 阈值 = normal
		if value <= threshold {
			return "normal"
		} else if value <= threshold*1.1 {
			return "warning"
		}
		return "critical"

	case model.ThresholdEqual:
		// 值 == 阈值 = normal
		if value == threshold {
			return "normal"
		}
		return "warning"

	case model.ThresholdAtLeast:
		// 值 >= 阈值 = normal
		if value >= threshold {
			return "normal"
		} else if value >= warningThreshold {
			return "warning"
		}
		return "critical"
	}

	return "normal"
}

func (s *InspectionService) GetReportByID(id uint) (*model.InspectionReport, []model.InspectionItem, error) {
	return s.reportRepo.GetByIDWithItems(id)
}

func (s *InspectionService) GetAllReports() ([]model.InspectionReport, error) {
	return s.reportRepo.GetAll()
}

// GetReportList 分页获取报告列表
func (s *InspectionService) GetReportList(keyword string, page, pageSize int) (*repository.ListResult, error) {
	return s.reportRepo.GetList(repository.ListParams{
		Keyword:  keyword,
		Page:     page,
		PageSize: pageSize,
	})
}

func (s *InspectionService) UpdateSummary(id uint, summary, remark string) error {
	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		return err
	}

	report.Summary = summary
	report.Remark = remark
	return s.reportRepo.Update(report)
}

// CleanupOldReports 清理过期报告
func (s *InspectionService) CleanupOldReports(days int) (int64, error) {
	return s.reportRepo.DeleteOlderThan(days)
}

// ParseInt 辅助函数：解析整数
func ParseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return val
}
