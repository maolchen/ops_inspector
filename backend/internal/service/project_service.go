package service

import (
	"errors"
	"ops-inspection/internal/model"
	"ops-inspection/internal/repository"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(project *model.Project) error {
	// 检查名称是否已存在
	existing, _ := s.repo.GetByName(project.Name)
	if existing != nil {
		return errors.New("项目名称已存在")
	}
	return s.repo.Create(project)
}

func (s *ProjectService) Update(project *model.Project) error {
	// 检查项目是否存在
	existing, err := s.repo.GetByID(project.ID)
	if err != nil {
		return errors.New("项目不存在")
	}

	// 如果修改了名称，检查新名称是否已被使用
	if existing.Name != project.Name {
		duplicate, _ := s.repo.GetByName(project.Name)
		if duplicate != nil {
			return errors.New("项目名称已存在")
		}
	}

	return s.repo.Update(project)
}

func (s *ProjectService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *ProjectService) GetByID(id uint) (*model.Project, error) {
	return s.repo.GetByID(id)
}

func (s *ProjectService) GetAll() ([]model.Project, error) {
	return s.repo.GetAll()
}

// MaskToken 隐藏 Token（用于展示）
func (s *ProjectService) MaskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 10 {
		return "**********"
	}
	return "**********"
}
