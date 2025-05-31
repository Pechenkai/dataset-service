package services

import (
	"context"
	"errors"
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
		return 0, ErrNilCategory
	}

	if err := s.repo.Create(ctx, cat); err != nil {
		switch {
		case errors.Is(err, repositories.ErrCategoryAlreadyExists):
			return 0, ErrCategoryExists
		case errors.Is(err, repositories.ErrCategoryCreate):
			return 0, ErrCategoryCreate
		default:
			return 0, err
		}
	}
	return cat.ID, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, cmd UpdateCategoryCmd) error {
	cat, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil || cat == nil {
		return ErrCategoryNotFound
	}

	updated, err := entities.NewCategory(cmd.Name, cmd.Description)
	if err != nil {
		return ErrNilCategory
	}
	updated.ID = cmd.ID

	if err := s.repo.Update(ctx, updated); err != nil {
		switch {
		case errors.Is(err, repositories.ErrCategoryAlreadyExists):
			return ErrCategoryExists
		case errors.Is(err, repositories.ErrCategoryNotFound):
			return ErrCategoryNotFound
		default:
			return err
		}
	}
	return nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, repositories.ErrCategoryNotEmpty):
			return ErrCategoryNotEmpty
		case errors.Is(err, repositories.ErrCategoryNotFound):
			return ErrCategoryNotFound
		default:
			return err
		}
	}
	return nil
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id uint64) (*entities.Category, error) {
	cat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return cat, nil
}

func (s *categoryService) ListCategories(ctx context.Context) ([]*entities.Category, error) {
	cats, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return cats, nil
}
