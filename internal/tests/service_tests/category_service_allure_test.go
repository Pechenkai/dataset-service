package service_tests

import (
    "context"
    "errors"
    "testing"

    "github.com/ozontech/allure-go/pkg/framework"
    "github.com/ozontech/allure-go/pkg/framework/provider"
    "github.com/ozontech/allure-go/pkg/framework/suite"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "go.uber.org/zap"

    "ppo/internal/entities"
    "ppo/internal/repositories"
    "ppo/internal/services"
    "ppo/internal/tests/mocks"
    "ppo/internal/tests/testdata"
)

type categoryRepoStub struct {
    createFn  func(ctx context.Context, c *entities.Category) error
    updateFn  func(ctx context.Context, c *entities.Category) error
    deleteFn  func(ctx context.Context, id uint64) error
    findIDFn  func(ctx context.Context, id uint64) (*entities.Category, error)
    findAllFn func(ctx context.Context) ([]*entities.Category, error)
}

func (s *categoryRepoStub) Create(ctx context.Context, c *entities.Category) error {
    if s.createFn != nil {
        return s.createFn(ctx, c)
    }
    return nil
}

func (s *categoryRepoStub) Update(ctx context.Context, c *entities.Category) error {
    if s.updateFn != nil {
        return s.updateFn(ctx, c)
    }
    return nil
}

func (s *categoryRepoStub) Delete(ctx context.Context, id uint64) error {
    if s.deleteFn != nil {
        return s.deleteFn(ctx, id)
    }
    return nil
}

func (s *categoryRepoStub) FindByID(ctx context.Context, id uint64) (*entities.Category, error) {
    if s.findIDFn != nil {
        return s.findIDFn(ctx, id)
    }
    return nil, nil
}

func (s *categoryRepoStub) FindAll(ctx context.Context) ([]*entities.Category, error) {
    if s.findAllFn != nil {
        return s.findAllFn(ctx)
    }
    return nil, nil
}

type CategoryServiceSuite struct {
    suite.Suite
    logger *zap.Logger
}

func TestCategoryServiceSuite(t *testing.T) {
    framework.Run(t, new(CategoryServiceSuite))
}

func (s *CategoryServiceSuite) BeforeEach(t provider.T) {
    s.logger = zap.NewNop()
}

func (s *CategoryServiceSuite) TestCreateCategory_Success_Classical(t provider.T) {
    // Arrange
    repoStub := &categoryRepoStub{createFn: func(ctx context.Context, c *entities.Category) error {
        c.ID = 42
        return nil
    }}
    svc := services.NewCategoryService(repoStub, s.logger)
    cmd := services.CreateCategoryCmd{Name: "ML", Description: "Machine learning"}

    // Act
    id, err := svc.CreateCategory(context.Background(), cmd)

    // Assert
    t.Require().NoError(err)
    t.Assert().Equal(uint64(42), id)
}

func (s *CategoryServiceSuite) TestCreateCategory_Exists_Error(t provider.T) {
    // Arrange
    repoStub := &categoryRepoStub{createFn: func(ctx context.Context, c *entities.Category) error {
        return repositories.ErrCategoryAlreadyExists
    }}
    svc := services.NewCategoryService(repoStub, s.logger)
    cmd := services.CreateCategoryCmd{Name: "ML", Description: "Machine learning"}

    // Act
    id, err := svc.CreateCategory(context.Background(), cmd)

    // Assert
    t.Assert().Equal(uint64(0), id)
    t.Require().Error(err)
    t.Assert().ErrorIs(err, services.ErrCategoryExists)
}

func (s *CategoryServiceSuite) TestCreateCategory_CallsRepository_London(t provider.T) {
    // Arrange
    repoMock := new(mocks.CategoryRepository)
    svc := services.NewCategoryService(repoMock, s.logger)
    cmd := services.CreateCategoryCmd{Name: "AI", Description: "Artificial intelligence"}
    repoMock.On("Create", mock.Anything, mock.AnythingOfType("*entities.Category")).Return(nil)

    // Act
    _, err := svc.CreateCategory(context.Background(), cmd)

    // Assert
    t.Require().NoError(err)
    t.Assert().Len(repoMock.Calls, 1)
    if len(repoMock.Calls) > 0 {
        t.Assert().Equal("Create", repoMock.Calls[0].Method)
    }
}

func (s *CategoryServiceSuite) TestUpdateCategory_Success(t provider.T) {
    // Arrange
    existing := testdata.NewCategoryBuilder().WithName("old").WithDescription("legacy").Build()
    existing.ID = 77
    repoStub := &categoryRepoStub{
        findIDFn: func(ctx context.Context, id uint64) (*entities.Category, error) {
            return existing, nil
        },
        updateFn: func(ctx context.Context, c *entities.Category) error {
            return nil
        },
    }
    svc := services.NewCategoryService(repoStub, s.logger)
    cmd := services.UpdateCategoryCmd{ID: existing.ID, Name: "new-name", Description: "updated"}

    // Act
    err := svc.UpdateCategory(context.Background(), cmd)

    // Assert
    t.Require().NoError(err)
}

func (s *CategoryServiceSuite) TestUpdateCategory_NotFound(t provider.T) {
    // Arrange
    repoStub := &categoryRepoStub{
        findIDFn: func(ctx context.Context, id uint64) (*entities.Category, error) {
            return nil, repositories.ErrCategoryNotFound
        },
    }
    svc := services.NewCategoryService(repoStub, s.logger)
    cmd := services.UpdateCategoryCmd{ID: 999, Name: "none", Description: ""}

    // Act
    err := svc.UpdateCategory(context.Background(), cmd)

    // Assert
    t.Require().Error(err)
    t.Assert().ErrorIs(err, services.ErrCategoryNotFound)
}

func (s *CategoryServiceSuite) TestDeleteCategory_Success(t provider.T) {
    // Arrange
    repoStub := &categoryRepoStub{
        deleteFn: func(ctx context.Context, id uint64) error { return nil },
    }
    svc := services.NewCategoryService(repoStub, s.logger)

    // Act
    err := svc.DeleteCategory(context.Background(), 11)

    // Assert
    t.Require().NoError(err)
}

func (s *CategoryServiceSuite) TestDeleteCategory_NotEmpty(t provider.T) {
    // Arrange
    repoStub := &categoryRepoStub{
        deleteFn: func(ctx context.Context, id uint64) error { return repositories.ErrCategoryNotEmpty },
    }
    svc := services.NewCategoryService(repoStub, s.logger)

    // Act
    err := svc.DeleteCategory(context.Background(), 11)

    // Assert
    t.Require().Error(err)
    t.Assert().ErrorIs(err, services.ErrCategoryNotEmpty)
}

func (s *CategoryServiceSuite) TestGetCategoryByID_Success(t provider.T) {
    // Arrange
    expected := testdata.NewCategoryBuilder().WithName("vision").Build()
    expected.ID = 5
    repoStub := &categoryRepoStub{
        findIDFn: func(ctx context.Context, id uint64) (*entities.Category, error) {
            return expected, nil
        },
    }
    svc := services.NewCategoryService(repoStub, s.logger)

    // Act
    cat, err := svc.GetCategoryByID(context.Background(), expected.ID)

    // Assert
    t.Require().NoError(err)
    t.Assert().Equal(expected, cat)
}

func (s *CategoryServiceSuite) TestGetCategoryByID_NotFound(t provider.T) {
    // Arrange
    repoStub := &categoryRepoStub{
        findIDFn: func(ctx context.Context, id uint64) (*entities.Category, error) {
            return nil, repositories.ErrCategoryNotFound
        },
    }
    svc := services.NewCategoryService(repoStub, s.logger)

    // Act
    cat, err := svc.GetCategoryByID(context.Background(), 404)

    // Assert
    t.Require().Error(err)
    t.Assert().ErrorIs(err, services.ErrCategoryNotFound)
    t.Assert().Nil(cat)
}

func (s *CategoryServiceSuite) TestListCategories_Success(t provider.T) {
    // Arrange
    cats := []*entities.Category{
        testdata.NewCategoryBuilder().WithName("vision").Build(),
        testdata.NewCategoryBuilder().WithName("nlp").Build(),
    }
    repoStub := &categoryRepoStub{
        findAllFn: func(ctx context.Context) ([]*entities.Category, error) {
            return cats, nil
        },
    }
    svc := services.NewCategoryService(repoStub, s.logger)

    // Act
    result, err := svc.ListCategories(context.Background())

    // Assert
    t.Require().NoError(err)
    t.Assert().Len(result, 2)
}

func (s *CategoryServiceSuite) TestListCategories_Error(t provider.T) {
    // Arrange
    repoStub := &categoryRepoStub{
        findAllFn: func(ctx context.Context) ([]*entities.Category, error) {
            return nil, errors.New("db unavailable")
        },
    }
    svc := services.NewCategoryService(repoStub, s.logger)

    // Act
    result, err := svc.ListCategories(context.Background())

    // Assert
    t.Require().Error(err)
    t.Assert().Nil(result)
}
