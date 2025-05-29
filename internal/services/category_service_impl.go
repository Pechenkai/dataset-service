package services

import (
	"context"
	"errors"
	"fmt"
	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type categoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) CreateCategory(ctx context.Context, cmd CreateCategoryCmd) (uint64, error) {
	cat, err := entities.NewCategory(cmd.Name, cmd.Description)
	if err != nil {
		return 0, fmt.Errorf("invalid category data: %w", err)
	}

	if err := s.repo.Create(ctx, cat); err != nil {
		if errors.Is(err, repositories.ErrCategoryAlreadyExists) {
			return 0, ErrCategoryExists
		}
		return 0, fmt.Errorf("create category: %w", err)
	}
	return cat.ID, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, cmd UpdateCategoryCmd) error {
	cat, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("fetch category: %w", err)
	}
	if cat == nil {
		return repositories.ErrCategoryNotFound
	}

	updated, err := entities.NewCategory(cmd.Name, cmd.Description)
	if err != nil {
		return fmt.Errorf("invalid category data: %w", err)
	}
	updated.ID = cmd.ID

	if err := s.repo.Update(ctx, updated); err != nil {
		if errors.Is(err, repositories.ErrCategoryAlreadyExists) {
			return ErrCategoryExists
		}
		return fmt.Errorf("update category: %w", err)
	}
	return nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, repositories.ErrCategoryNotEmpty):
			return ErrCategoryNotEmpty
		case errors.Is(err, repositories.ErrCategoryNotFound):
			return repositories.ErrCategoryNotFound
		default:
			return fmt.Errorf("delete category: %w", err)
		}
	}
	return nil
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id uint64) (*entities.Category, error) {
	cat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}
	if cat == nil || errors.Is(err, repositories.ErrCategoryNotFound) {
		return nil, ErrCategoryNotFound
	}
	return cat, nil
}

func (s *categoryService) ListCategories(ctx context.Context) ([]*entities.Category, error) {
	cats, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return cats, nil
}
