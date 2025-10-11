package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"ppo/internal/entities"
	"ppo/internal/repositories"
	"ppo/internal/services"
	"ppo/internal/tests/mocks"
	"ppo/internal/tests/testdata"
)

type CategoryServiceSuite struct {
	suite.Suite
}

type categoryMocks struct {
	repo *mocks.CategoryRepository
	svc  services.CategoryService
}

func newCategoryMocks(t provider.T) categoryMocks {
	t.Helper()
	repo := &mocks.CategoryRepository{}
	svc := services.NewCategoryService(repo, zap.NewNop())
	return categoryMocks{repo: repo, svc: svc}
}

func (m categoryMocks) AssertExpectations(t provider.T) {
	t.Helper()
	m.repo.AssertExpectations(t)
}

func (s *CategoryServiceSuite) TestCreateCategory_Success(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.CreateCategoryCommand()
	m := newCategoryMocks(t)

	expectedID := uint64(501)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Category")).Run(func(args mock.Arguments) {
			cat := args.Get(1).(*entities.Category)
			cat.ID = expectedID
		}).Return(nil)
	})

	var (
		id  uint64
		err error
	)

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		id, err = m.svc.CreateCategory(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, expectedID, id)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestCreateCategory_InvalidPayload(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.InvalidCreateCategoryCommand()
	m := newCategoryMocks(t)

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.CreateCategory(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrNilCategory)
	})
}

func (s *CategoryServiceSuite) TestCreateCategory_AlreadyExists(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.CreateCategoryCommand()
	m := newCategoryMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Create", mock.Anything, mock.Anything).Return(repositories.ErrCategoryAlreadyExists)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.CreateCategory(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrCategoryExists)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestCreateCategory_RepoFailure(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.CreateCategoryCommand()
	m := newCategoryMocks(t)
	expectedErr := errors.New("db down")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Create", mock.Anything, mock.Anything).Return(expectedErr)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.CreateCategory(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestUpdateCategory_Success(t provider.T) {
	fabric := testdata.NewFabric()
	stored := fabric.Category()
	cmd := fabric.UpdateCategoryCommand(stored)
	m := newCategoryMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, stored.ID).Return(stored, nil)
		m.repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Category")).Run(func(args mock.Arguments) {
			updated := args.Get(1).(*entities.Category)
			assert.Equal(t, cmd.ID, updated.ID)
			assert.Equal(t, cmd.Name, updated.Name)
			assert.Equal(t, cmd.Description, updated.Description)
		}).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateCategory(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestUpdateCategory_NotFound(t provider.T) {
	fabric := testdata.NewFabric()
	stored := fabric.Category()
	cmd := fabric.UpdateCategoryCommand(stored)
	m := newCategoryMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, cmd.ID).Return((*entities.Category)(nil), repositories.ErrCategoryNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateCategory(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrCategoryNotFound)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestUpdateCategory_InvalidPayload(t provider.T) {
	fabric := testdata.NewFabric()
	stored := fabric.Category()
	cmd := fabric.InvalidUpdateCategoryCommand(stored)
	m := newCategoryMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, cmd.ID).Return(stored, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateCategory(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrNilCategory)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestUpdateCategory_RepoErrors(t provider.T) {
	fabric := testdata.NewFabric()
	stored := fabric.Category()
	cmd := fabric.UpdateCategoryCommand(stored)
	m := newCategoryMocks(t)

	repoErr := repositories.ErrCategoryAlreadyExists
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, cmd.ID).Return(stored, nil)
		m.repo.On("Update", mock.Anything, mock.Anything).Return(repoErr)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateCategory(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrCategoryExists)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestDeleteCategory_Success(t provider.T) {
	fabric := testdata.NewFabric()
	category := fabric.Category()
	m := newCategoryMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Delete", mock.Anything, category.ID).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.DeleteCategory(context.Background(), category.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestDeleteCategory_NotEmpty(t provider.T) {
	fabric := testdata.NewFabric()
	category := fabric.Category()
	m := newCategoryMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Delete", mock.Anything, category.ID).Return(repositories.ErrCategoryNotEmpty)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.DeleteCategory(context.Background(), category.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrCategoryNotEmpty)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestDeleteCategory_NotFound(t provider.T) {
	fabric := testdata.NewFabric()
	category := fabric.Category()
	m := newCategoryMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Delete", mock.Anything, category.ID).Return(repositories.ErrCategoryNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.DeleteCategory(context.Background(), category.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrCategoryNotFound)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestGetCategoryByID_Success(t provider.T) {
	fabric := testdata.NewFabric()
	category := fabric.Category()
	m := newCategoryMocks(t)

	var (
		result *entities.Category
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, category.ID).Return(category, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetCategoryByID(context.Background(), category.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, category.ID, result.ID)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestGetCategoryByID_NotFound(t provider.T) {
	m := newCategoryMocks(t)
	categoryID := uint64(77)

	var (
		result *entities.Category
		err    error
	)
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, categoryID).Return((*entities.Category)(nil), repositories.ErrCategoryNotFound)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetCategoryByID(context.Background(), categoryID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, services.ErrCategoryNotFound)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestGetCategoryByID_RepoError(t provider.T) {
	m := newCategoryMocks(t)
	categoryID := uint64(88)
	expectedErr := errors.New("boom")

	var err error
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, categoryID).Return((*entities.Category)(nil), expectedErr)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.GetCategoryByID(context.Background(), categoryID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func (s *CategoryServiceSuite) TestListCategories_Success(t provider.T) {
	fabric := testdata.NewFabric()
	repo := &inMemoryCategoryRepo{}
	repo.categories = []*entities.Category{fabric.Category(), fabric.Category()}
	svc := services.NewCategoryService(repo, zap.NewNop())

	var (
		result []*entities.Category
		err    error
	)
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = svc.ListCategories(context.Background())
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		require.Len(t, result, 2)
	})
}

func (s *CategoryServiceSuite) TestListCategories_RepoError(t provider.T) {
	m := newCategoryMocks(t)
	expectedErr := errors.New("storage fail")

	var (
		result []*entities.Category
		err    error
	)
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindAll", mock.Anything).Return(nil, expectedErr)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListCategories(context.Background())
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
		m.AssertExpectations(t)
	})
}

func TestCategoryServiceSuite(t *testing.T) {
	suite.RunSuite(t, new(CategoryServiceSuite))
}

type inMemoryCategoryRepo struct {
	categories []*entities.Category
}

func (r *inMemoryCategoryRepo) Create(ctx context.Context, c *entities.Category) error {
	r.categories = append(r.categories, c)
	if c.ID == 0 {
		c.ID = uint64(len(r.categories))
	}
	return nil
}

func (r *inMemoryCategoryRepo) Delete(ctx context.Context, id uint64) error {
	for i, cat := range r.categories {
		if cat.ID == id {
			r.categories = append(r.categories[:i], r.categories[i+1:]...)
			return nil
		}
	}
	return repositories.ErrCategoryNotFound
}

func (r *inMemoryCategoryRepo) Update(ctx context.Context, c *entities.Category) error {
	for i, cat := range r.categories {
		if cat.ID == c.ID {
			r.categories[i] = c
			return nil
		}
	}
	return repositories.ErrCategoryNotFound
}

func (r *inMemoryCategoryRepo) FindByID(ctx context.Context, id uint64) (*entities.Category, error) {
	for _, cat := range r.categories {
		if cat.ID == id {
			return cat, nil
		}
	}
	return nil, repositories.ErrCategoryNotFound
}

func (r *inMemoryCategoryRepo) FindAll(ctx context.Context) ([]*entities.Category, error) {
	return r.categories, nil
}
