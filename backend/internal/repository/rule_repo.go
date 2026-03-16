package repository

import (
	"ops-inspection/internal/model"

	"gorm.io/gorm"
)

type RuleRepository struct {
	db *gorm.DB
}

func NewRuleRepository(db *gorm.DB) *RuleRepository {
	return &RuleRepository{db: db}
}

func (r *RuleRepository) Create(rule *model.Rule) error {
	return r.db.Create(rule).Error
}

func (r *RuleRepository) Update(rule *model.Rule) error {
	return r.db.Save(rule).Error
}

func (r *RuleRepository) Delete(id uint) error {
	return r.db.Delete(&model.Rule{}, id).Error
}

func (r *RuleRepository) GetByID(id uint) (*model.Rule, error) {
	var rule model.Rule
	err := r.db.Preload("Group").First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *RuleRepository) GetAll() ([]model.Rule, error) {
	var rules []model.Rule
	err := r.db.Preload("Group").Order("group_id asc, sort_order asc, id asc").Find(&rules).Error
	return rules, err
}

func (r *RuleRepository) GetByGroupID(groupID uint) ([]model.Rule, error) {
	var rules []model.Rule
	err := r.db.Where("group_id = ?", groupID).Order("sort_order asc, id asc").Find(&rules).Error
	return rules, err
}

func (r *RuleRepository) GetEnabled() ([]model.Rule, error) {
	var rules []model.Rule
	err := r.db.Preload("Group").Where("enabled = ?", true).Order("group_id asc, sort_order asc, id asc").Find(&rules).Error
	return rules, err
}

func (r *RuleRepository) GetByProjectScope(projectName string) ([]model.Rule, error) {
	var rules []model.Rule
	err := r.db.Preload("Group").
		Where("enabled = ? AND (project_scope = ? OR project_scope = ?)", true, "*", projectName).
		Order("group_id asc, sort_order asc, id asc").
		Find(&rules).Error
	return rules, err
}
