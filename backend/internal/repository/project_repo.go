package repository

import (
	"ops-inspection/internal/model"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(project *model.Project) error {
	return r.db.Create(project).Error
}

func (r *ProjectRepository) Update(project *model.Project) error {
	return r.db.Save(project).Error
}

func (r *ProjectRepository) Delete(id uint) error {
	return r.db.Delete(&model.Project{}, id).Error
}

func (r *ProjectRepository) GetByID(id uint) (*model.Project, error) {
	var project model.Project
	err := r.db.First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) GetAll() ([]model.Project, error) {
	var projects []model.Project
	err := r.db.Order("created_at desc").Find(&projects).Error
	return projects, err
}

func (r *ProjectRepository) GetByName(name string) (*model.Project, error) {
	var project model.Project
	err := r.db.Where("name = ?", name).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}
