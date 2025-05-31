package services

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type categoryService struct {
	repo   repositories.CategoryRepository
	logger *zap.Logger
}

func NewCategoryService(repo repositories.CategoryRepository, logger *zap.Logger) CategoryService {
	return &categoryService{repo: repo, logger: logger}
}

func (s *categoryService) CreateCategory(ctx context.Context, cmd CreateCategoryCmd) (uint64, error) {
	cat, err := entities.NewCategory(cmd.Name, cmd.Description)
	if err != nil {
		s.logger.Error("Nil category while creating", zap.Error(err))
		return 0, ErrNilCategory
	}

	s.logger.Info("Category created", zap.String("name", cat.Name), zap.String("description", cat.Description))

	if err := s.repo.Create(ctx, cat); err != nil {
		switch {
		case errors.Is(err, repositories.ErrCategoryAlreadyExists):
			s.logger.Warn("Category already exists", zap.String("name", cat.Name), zap.String("description", cat.Description))
			return 0, ErrCategoryExists
		case errors.Is(err, repositories.ErrCategoryCreate):
			s.logger.Error("Failed to create category", zap.String("name", cat.Name), zap.String("description", cat.Description))
			return 0, ErrCategoryCreate
		default:
			s.logger.Error("Failed to create category: No info", zap.String("name", cat.Name), zap.String("description", cat.Description))
			return 0, err
		}
	}
	return cat.ID, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, cmd UpdateCategoryCmd) error {
	cat, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil || cat == nil {
		s.logger.Warn("Failed to find category", zap.Error(err))
		return ErrCategoryNotFound
	}

	updated, err := entities.NewCategory(cmd.Name, cmd.Description)
	if err != nil {
		s.logger.Error("Failed to update category", zap.Error(err))
		return ErrNilCategory
	}
	updated.ID = cmd.ID

	if err := s.repo.Update(ctx, updated); err != nil {
		switch {
		case errors.Is(err, repositories.ErrCategoryAlreadyExists):
			s.logger.Warn("Category already exists", zap.String("name", updated.Name), zap.String("description", updated.Description))
			return ErrCategoryExists
		case errors.Is(err, repositories.ErrCategoryNotFound):
			s.logger.Warn("Category not found", zap.Error(err))
			return ErrCategoryNotFound
		default:
			return err
		}
	}

	s.logger.Info("Category updated", zap.String("name", cat.Name), zap.String("description", cat.Description))
	return nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, repositories.ErrCategoryNotEmpty):
			s.logger.Warn("Category not empty", zap.Uint64("id", id))
			return ErrCategoryNotEmpty
		case errors.Is(err, repositories.ErrCategoryNotFound):
			s.logger.Warn("Category not found", zap.Uint64("id", id))
			return ErrCategoryNotFound
		default:
			s.logger.Error("Failed to delete category", zap.Uint64("id", id))
			return err
		}
	}

	s.logger.Info("Category deleted", zap.Uint64("id", id))
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
		s.logger.Error("Failed to list all categories", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Category listed", zap.Int("count", len(cats)))
	return cats, nil
}
