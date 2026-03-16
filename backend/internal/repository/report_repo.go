package repository

import (
	"ops-inspection/internal/model"
	"time"

	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(report *model.InspectionReport) error {
	return r.db.Create(report).Error
}

func (r *ReportRepository) CreateWithItems(report *model.InspectionReport, items []model.InspectionItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(report).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ReportID = report.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ReportRepository) Update(report *model.InspectionReport) error {
	return r.db.Save(report).Error
}

func (r *ReportRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先删除关联的巡检项
		if err := tx.Where("report_id = ?", id).Delete(&model.InspectionItem{}).Error; err != nil {
			return err
		}
		// 再删除报告
		return tx.Delete(&model.InspectionReport{}, id).Error
	})
}

func (r *ReportRepository) GetByID(id uint) (*model.InspectionReport, error) {
	var report model.InspectionReport
	err := r.db.Preload("Items").First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepository) GetAll() ([]model.InspectionReport, error) {
	var reports []model.InspectionReport
	err := r.db.Order("created_at desc").Find(&reports).Error
	return reports, err
}

// ListParams 列表查询参数
type ListParams struct {
	Keyword    string // 搜索关键词（项目名称）
	Page       int    // 页码
	PageSize   int    // 每页数量
	ProjectID  uint   // 项目ID筛选
	StartTime  string // 开始时间
	EndTime    string // 结束时间
}

// ListResult 列表查询结果
type ListResult struct {
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
	List     []model.InspectionReport `json:"list"`
}

// GetList 分页查询报告列表
func (r *ReportRepository) GetList(params ListParams) (*ListResult, error) {
	var result ListResult
	var reports []model.InspectionReport

	// 设置默认值
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	result.Page = params.Page
	result.PageSize = params.PageSize

	// 构建查询
	query := r.db.Model(&model.InspectionReport{})

	if params.Keyword != "" {
		query = query.Where("project_name LIKE ?", "%"+params.Keyword+"%")
	}
	if params.ProjectID > 0 {
		query = query.Where("project_id = ?", params.ProjectID)
	}
	if params.StartTime != "" {
		query = query.Where("created_at >= ?", params.StartTime)
	}
	if params.EndTime != "" {
		query = query.Where("created_at <= ?", params.EndTime)
	}

	// 查询总数
	if err := query.Count(&result.Total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(params.PageSize).Find(&reports).Error; err != nil {
		return nil, err
	}

	result.List = reports
	return &result, nil
}

func (r *ReportRepository) GetByProjectID(projectID uint) ([]model.InspectionReport, error) {
	var reports []model.InspectionReport
	err := r.db.Where("project_id = ?", projectID).Order("created_at desc").Find(&reports).Error
	return reports, err
}

func (r *ReportRepository) GetByIDWithItems(id uint) (*model.InspectionReport, []model.InspectionItem, error) {
	var report model.InspectionReport
	var items []model.InspectionItem

	err := r.db.First(&report, id).Error
	if err != nil {
		return nil, nil, err
	}

	err = r.db.Where("report_id = ?", id).Order("group_id asc, id asc").Find(&items).Error
	if err != nil {
		return nil, nil, err
	}

	return &report, items, nil
}

// DeleteOlderThan 删除指定天数之前的报告
func (r *ReportRepository) DeleteOlderThan(days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	var count int64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 获取需要删除的报告ID
		var reportIDs []uint
		if err := tx.Model(&model.InspectionReport{}).
			Where("created_at < ?", cutoffTime).
			Pluck("id", &reportIDs).Error; err != nil {
			return err
		}

		if len(reportIDs) == 0 {
			return nil
		}

		// 删除关联的巡检项
		if err := tx.Where("report_id IN ?", reportIDs).Delete(&model.InspectionItem{}).Error; err != nil {
			return err
		}

		// 删除报告
		if err := tx.Where("id IN ?", reportIDs).Delete(&model.InspectionReport{}).Error; err != nil {
			return err
		}

		count = int64(len(reportIDs))
		return nil
	})

	return count, err
}
