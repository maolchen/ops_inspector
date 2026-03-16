package repository

import (
	"ops-inspection/internal/model"

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
