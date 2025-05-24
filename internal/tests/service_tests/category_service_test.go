package services_test

import (
	"context"
	"ppo/internal/entities"
	"ppo/internal/services"
	"ppo/internal/tests/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCategoryService_CreateCategory_Success(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Category")).Run(func(args mock.Arguments) {
		cat := args.Get(1).(*entities.Category)
		cat.ID = 1
	}).Return(nil)

	id, err := svc.CreateCategory(context.Background(), "ML", "Machine Learning")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), id)
	repo.AssertExpectations(t)
}

func TestCategoryService_CreateCategory_ValidationError(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	id, err := svc.CreateCategory(context.Background(), "   ", "description")
	assert.ErrorIs(t, err, entities.ErrEmptyCategoryName)
	assert.Equal(t, uint64(0), id)
}

func TestCategoryService_CreateCategory_DuplicateError(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	repo.On("Create", mock.Anything, mock.Anything).Return(services.ErrCategoryExists)

	id, err := svc.CreateCategory(context.Background(), "ML", "description")
	assert.ErrorIs(t, err, services.ErrCategoryExists)
	assert.Equal(t, uint64(0), id)
}

func TestCategoryService_UpdateCategory_Success(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	existing := &entities.Category{ID: 1, Name: "Old", Description: "Desc"}
	repo.On("FindByID", mock.Anything, uint64(1)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Category")).Return(nil)

	err := svc.UpdateCategory(context.Background(), 1, "New", "Updated")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCategoryService_UpdateCategory_NotFound(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	repo.On("FindByID", mock.Anything, uint64(999)).Return((*entities.Category)(nil), nil)

	err := svc.UpdateCategory(context.Background(), 999, "New", "Updated")
	assert.Error(t, err)
}

func TestCategoryService_UpdateCategory_ValidationError(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	existing := &entities.Category{ID: 1, Name: "Old", Description: "Desc"}
	repo.On("FindByID", mock.Anything, uint64(1)).Return(existing, nil)

	err := svc.UpdateCategory(context.Background(), 1, "", "Updated")
	assert.ErrorIs(t, err, entities.ErrEmptyCategoryName)
}

func TestCategoryService_DeleteCategory_Success(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	repo.On("Delete", mock.Anything, uint64(1)).Return(nil)

	err := svc.DeleteCategory(context.Background(), 1)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCategoryService_DeleteCategory_NotFound(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	repo.On("Delete", mock.Anything, uint64(2)).Return(services.ErrCategoryNotFound)

	err := svc.DeleteCategory(context.Background(), 2)
	assert.ErrorIs(t, err, services.ErrCategoryNotFound)
}

func TestCategoryService_DeleteCategory_NotEmpty(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	repo.On("Delete", mock.Anything, uint64(3)).Return(services.ErrCategoryNotEmpty)

	err := svc.DeleteCategory(context.Background(), 3)
	assert.ErrorIs(t, err, services.ErrCategoryNotEmpty)
}

func TestCategoryService_GetCategoryByID_Success(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	expected := &entities.Category{ID: 1, Name: "Test"}
	repo.On("FindByID", mock.Anything, uint64(1)).Return(expected, nil)

	cat, err := svc.GetCategoryByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, expected, cat)
}

func TestCategoryService_GetCategoryByID_NotFound(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	repo.On("FindByID", mock.Anything, uint64(2)).Return(nil, nil)

	cat, err := svc.GetCategoryByID(context.Background(), 2)
	assert.Error(t, err)
	assert.Nil(t, cat)
}

func TestCategoryService_ListCategories_Success(t *testing.T) {
	repo := new(mocks.CategoryRepository)
	svc := services.NewCategoryService(repo)

	expected := []*entities.Category{
		{ID: 1, Name: "Cat1"},
		{ID: 2, Name: "Cat2"},
	}
	repo.On("FindAll", mock.Anything).Return(expected, nil)

	result, err := svc.ListCategories(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
