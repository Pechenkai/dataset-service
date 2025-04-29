package services

import (
	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type categoryService struct {
	categoryRepo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: repo}
}

func (s *categoryService) CreateCategory(category *entities.Category) error {
	if category == nil {
		return ErrNilCategory
	}
	return s.categoryRepo.Create(category)
}

func (s *categoryService) UpdateCategory(category *entities.Category) error {
	if category == nil {
		return ErrNilCategory
	}
	return s.categoryRepo.Update(category)
}

func (s *categoryService) DeleteCategory(id uint64) error {
	return s.categoryRepo.Delete(id)
}

func (s *categoryService) GetCategoryByID(id uint64) (*entities.Category, error) {
	return s.categoryRepo.FindByID(id)
}

func (s *categoryService) GetAllCategories() ([]*entities.Category, error) {
	return s.categoryRepo.FindAll()
}
