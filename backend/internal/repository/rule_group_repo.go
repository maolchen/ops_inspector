package repository

import (
	"ops-inspection/internal/model"

	"gorm.io/gorm"
)

type RuleGroupRepository struct {
	db *gorm.DB
}

func NewRuleGroupRepository(db *gorm.DB) *RuleGroupRepository {
	return &RuleGroupRepository{db: db}
}

func (r *RuleGroupRepository) Create(group *model.RuleGroup) error {
	return r.db.Create(group).Error
}

func (r *RuleGroupRepository) Update(group *model.RuleGroup) error {
	return r.db.Save(group).Error
}

func (r *RuleGroupRepository) Delete(id uint) error {
	// 检查是否有关联的规则
	var count int64
	r.db.Model(&model.Rule{}).Where("group_id = ?", id).Count(&count)
	if count > 0 {
		return gorm.ErrRecordNotFound // 使用错误表示有关联数据
	}
	return r.db.Delete(&model.RuleGroup{}, id).Error
}

func (r *RuleGroupRepository) GetByID(id uint) (*model.RuleGroup, error) {
	var group model.RuleGroup
	err := r.db.First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *RuleGroupRepository) GetAll() ([]model.RuleGroup, error) {
	var groups []model.RuleGroup
	err := r.db.Order("sort_order asc, id asc").Find(&groups).Error
	return groups, err
}

func (r *RuleGroupRepository) GetByCode(code string) (*model.RuleGroup, error) {
	var group model.RuleGroup
	err := r.db.Where("code = ?", code).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}
