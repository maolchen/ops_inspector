package service

import (
	"ops-inspection/internal/model"
	"ops-inspection/internal/repository"
)

type RuleService struct {
	repo      *repository.RuleRepository
	groupRepo *repository.RuleGroupRepository
}

func NewRuleService(repo *repository.RuleRepository, groupRepo *repository.RuleGroupRepository) *RuleService {
	return &RuleService{repo: repo, groupRepo: groupRepo}
}

func (s *RuleService) Create(rule *model.Rule) error {
	// 验证规则组是否存在
	_, err := s.groupRepo.GetByID(rule.GroupID)
	if err != nil {
		return err
	}

	// 设置默认阈值类型
	if rule.ThresholdType == "" {
		rule.ThresholdType = model.ThresholdGreater
	}

	// 设置默认项目范围
	if rule.ProjectScope == "" {
		rule.ProjectScope = "*"
	}

	return s.repo.Create(rule)
}

func (s *RuleService) Update(rule *model.Rule) error {
	// 验证规则组是否存在
	_, err := s.groupRepo.GetByID(rule.GroupID)
	if err != nil {
		return err
	}

	return s.repo.Update(rule)
}

func (s *RuleService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *RuleService) GetByID(id uint) (*model.Rule, error) {
	return s.repo.GetByID(id)
}

func (s *RuleService) GetAll() ([]model.Rule, error) {
	return s.repo.GetAll()
}

func (s *RuleService) GetByGroupID(groupID uint) ([]model.Rule, error) {
	return s.repo.GetByGroupID(groupID)
}

func (s *RuleService) GetEnabled() ([]model.Rule, error) {
	return s.repo.GetEnabled()
}

func (s *RuleService) GetByProjectScope(projectName string) ([]model.Rule, error) {
	return s.repo.GetByProjectScope(projectName)
}

// ToggleEnabled 切换规则启用状态
func (s *RuleService) ToggleEnabled(id uint) error {
	rule, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	rule.Enabled = !rule.Enabled
	return s.repo.Update(rule)
}
