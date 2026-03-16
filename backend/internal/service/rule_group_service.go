package service

import (
	"errors"
	"ops-inspection/internal/model"
	"ops-inspection/internal/repository"
)

type RuleGroupService struct {
	repo *repository.RuleGroupRepository
}

func NewRuleGroupService(repo *repository.RuleGroupRepository) *RuleGroupService {
	return &RuleGroupService{repo: repo}
}

func (s *RuleGroupService) Create(group *model.RuleGroup) error {
	// 检查 Code 是否已存在
	existing, _ := s.repo.GetByCode(group.Code)
	if existing != nil {
		return errors.New("规则组标识已存在")
	}
	return s.repo.Create(group)
}

func (s *RuleGroupService) Update(group *model.RuleGroup) error {
	// 检查规则组是否存在
	existing, err := s.repo.GetByID(group.ID)
	if err != nil {
		return errors.New("规则组不存在")
	}

	// 如果修改了 Code，检查新 Code 是否已被使用
	if existing.Code != group.Code {
		duplicate, _ := s.repo.GetByCode(group.Code)
		if duplicate != nil {
			return errors.New("规则组标识已存在")
		}
	}

	return s.repo.Update(group)
}

func (s *RuleGroupService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *RuleGroupService) GetByID(id uint) (*model.RuleGroup, error) {
	return s.repo.GetByID(id)
}

func (s *RuleGroupService) GetAll() ([]model.RuleGroup, error) {
	return s.repo.GetAll()
}
