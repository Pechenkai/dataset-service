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
	"golang.org/x/crypto/bcrypt"

	"ppo/internal/entities"
	"ppo/internal/repositories"
	"ppo/internal/services"
	"ppo/internal/tests/mocks"
	"ppo/internal/tests/testdata"
)

type UserServiceSuite struct {
	suite.Suite
}

type userMocks struct {
	repo *mocks.UserRepository
	svc  services.UserService
}

func newUserMocks(t provider.T) userMocks {
	t.Helper()
	repo := &mocks.UserRepository{}
	svc := services.NewUserService(repo, zap.NewNop())
	return userMocks{repo: repo, svc: svc}
}

func (m userMocks) AssertExpectations(t provider.T) {
	t.Helper()
	m.repo.AssertExpectations(t)
}

func (s *UserServiceSuite) TestRegister_Success(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.RegisterUserCommand()
	m := newUserMocks(t)

	originalPassword := cmd.Password
	expectedID := uint64(777)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, cmd.Email).Return((*entities.User)(nil), repositories.ErrUserNotFound)
		m.repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.User")).Run(func(args mock.Arguments) {
			user := args.Get(1).(*entities.User)
			assert.NotEqual(t, originalPassword, user.Password)
			require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(originalPassword)))
			user.ID = expectedID
		}).Return(nil)
	})

	var (
		id  uint64
		err error
	)

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		id, err = m.svc.Register(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, expectedID, id)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestRegister_InvalidData(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.InvalidRegisterUserCommand()
	m := newUserMocks(t)

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.Register(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user data")
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestRegister_FindByEmailError(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.RegisterUserCommand()
	m := newUserMocks(t)
	expected := errors.New("db down")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, cmd.Email).Return((*entities.User)(nil), expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.Register(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestRegister_UserExists(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.RegisterUserCommand()
	m := newUserMocks(t)

	dup := testdata.NewUserBuilder().WithEmail(cmd.Email).Build()

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, cmd.Email).Return(dup, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.Register(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrUserExists)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestRegister_CreateError(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.RegisterUserCommand()
	m := newUserMocks(t)
	expected := errors.New("insert fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, cmd.Email).Return((*entities.User)(nil), repositories.ErrUserNotFound)
		m.repo.On("Create", mock.Anything, mock.Anything).Return(expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.Register(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestRegister_CreateDuplicate(t provider.T) {
	fabric := testdata.NewFabric()
	cmd := fabric.RegisterUserCommand()
	m := newUserMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, cmd.Email).Return((*entities.User)(nil), repositories.ErrUserNotFound)
		m.repo.On("Create", mock.Anything, mock.Anything).Return(repositories.ErrEmailAlreadyExists)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.Register(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrUserExists)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestAuthenticate_Success(t provider.T) {
	m := newUserMocks(t)
	cmd := services.AuthenticateUserCmd{Email: "user@example.com", Password: "secret"}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	stored := testdata.NewUserBuilder().WithEmail(cmd.Email).Build()
	stored.Password = string(hashed)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, cmd.Email).Return(stored, nil)
	})

	var (
		user *entities.User
		err  error
	)

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		user, err = m.svc.Authenticate(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, stored.ID, user.ID)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestAuthenticate_UserNotFound(t provider.T) {
	m := newUserMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, "missing@example.com").Return((*entities.User)(nil), repositories.ErrUserNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.Authenticate(context.Background(), services.AuthenticateUserCmd{Email: "missing@example.com", Password: "pwd"})
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrInvalidCredentials)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestAuthenticate_RepoError(t provider.T) {
	m := newUserMocks(t)
	expected := errors.New("lookup fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, "user@example.com").Return((*entities.User)(nil), expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.Authenticate(context.Background(), services.AuthenticateUserCmd{Email: "user@example.com", Password: "pwd"})
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestAuthenticate_UserBlocked(t provider.T) {
	m := newUserMocks(t)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	blocked := testdata.NewUserBuilder().WithEmail("user@example.com").Build()
	blocked.Password = string(hashed)
	blocked.IsBlocked = true

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, blocked.Email).Return(blocked, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.Authenticate(context.Background(), services.AuthenticateUserCmd{Email: blocked.Email, Password: "secret"})
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrUserBlocked)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestAuthenticate_InvalidPassword(t provider.T) {
	m := newUserMocks(t)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	stored := testdata.NewUserBuilder().WithEmail("user@example.com").Build()
	stored.Password = string(hashed)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, stored.Email).Return(stored, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.Authenticate(context.Background(), services.AuthenticateUserCmd{Email: stored.Email, Password: "wrong"})
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrInvalidCredentials)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestUpdateUser_Success(t provider.T) {
	m := newUserMocks(t)
	stored := testdata.NewUserBuilder().WithID(10).Build()
	hashed, _ := bcrypt.GenerateFromPassword([]byte("old"), bcrypt.DefaultCost)
	stored.Password = string(hashed)
	cmd := testdata.NewUpdateUserCmdBuilder().WithID(stored.ID).WithEmail("new@example.com").WithPassword("newpass").Build()

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, stored.ID).Return(stored, nil)
		m.repo.On("FindByEmail", mock.Anything, cmd.Email).Return((*entities.User)(nil), repositories.ErrUserNotFound)
		m.repo.On("Update", mock.Anything, mock.AnythingOfType("*entities.User")).Run(func(args mock.Arguments) {
			updated := args.Get(1).(*entities.User)
			require.NoError(t, bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte(cmd.Password)))
			assert.Equal(t, cmd.Email, updated.Email)
		}).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateUser(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestUpdateUser_FindByIDError(t provider.T) {
	m := newUserMocks(t)
	expected := errors.New("fetch fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, uint64(2)).Return((*entities.User)(nil), expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateUser(context.Background(), services.UpdateUserCmd{ID: 2})
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestUpdateUser_NotFound(t provider.T) {
	m := newUserMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, uint64(3)).Return((*entities.User)(nil), nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateUser(context.Background(), services.UpdateUserCmd{ID: 3})
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrUserNotFound)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestUpdateUser_EmailExists(t provider.T) {
	m := newUserMocks(t)
	stored := testdata.NewUserBuilder().WithID(5).WithEmail("old@example.com").Build()
	cmd := services.UpdateUserCmd{ID: stored.ID, Email: "taken@example.com"}

	dup := testdata.NewUserBuilder().WithEmail("taken@example.com").Build()

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, stored.ID).Return(stored, nil)
		m.repo.On("FindByEmail", mock.Anything, cmd.Email).Return(dup, nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateUser(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrUserExists)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestUpdateUser_UpdateError(t provider.T) {
	m := newUserMocks(t)
	stored := testdata.NewUserBuilder().WithID(11).Build()
	cmd := services.UpdateUserCmd{ID: stored.ID}
	expected := errors.New("update fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, stored.ID).Return(stored, nil)
		m.repo.On("Update", mock.Anything, mock.Anything).Return(expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateUser(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestUpdateUser_UpdateNotFound(t provider.T) {
	m := newUserMocks(t)
	stored := testdata.NewUserBuilder().WithID(12).Build()
	cmd := services.UpdateUserCmd{ID: stored.ID}

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, stored.ID).Return(stored, nil)
		m.repo.On("Update", mock.Anything, mock.Anything).Return(repositories.ErrUserNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.UpdateUser(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrUserNotFound)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestDeleteUser_Success(t provider.T) {
	m := newUserMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Delete", mock.Anything, uint64(42)).Return(nil)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.DeleteUser(context.Background(), 42)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestDeleteUser_NotFound(t provider.T) {
	m := newUserMocks(t)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Delete", mock.Anything, uint64(42)).Return(repositories.ErrUserNotFound)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.DeleteUser(context.Background(), 42)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, services.ErrUserNotFound)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestDeleteUser_Error(t provider.T) {
	m := newUserMocks(t)
	expected := errors.New("delete fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("Delete", mock.Anything, uint64(42)).Return(expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		err = m.svc.DeleteUser(context.Background(), 42)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestGetUserByID_Success(t provider.T) {
	m := newUserMocks(t)
	user := testdata.NewUserBuilder().WithID(5).Build()

	var (
		result *entities.User
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetUserByID(context.Background(), user.ID)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, user.ID, result.ID)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestGetUserByID_NotFound(t provider.T) {
	m := newUserMocks(t)
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, uint64(6)).Return((*entities.User)(nil), nil)
	})

	var (
		result *entities.User
		err    error
	)
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetUserByID(context.Background(), 6)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, services.ErrUserNotFound)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestGetUserByID_Error(t provider.T) {
	m := newUserMocks(t)
	expected := errors.New("fetch fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByID", mock.Anything, uint64(6)).Return((*entities.User)(nil), expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.GetUserByID(context.Background(), 6)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestGetUserByEmail_Success(t provider.T) {
	m := newUserMocks(t)
	user := testdata.NewUserBuilder().WithEmail("user@example.com").Build()

	var (
		result *entities.User
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetUserByEmail(context.Background(), user.Email)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, user.Email, result.Email)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestGetUserByEmail_NotFound(t provider.T) {
	m := newUserMocks(t)
	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, "missing@example.com").Return((*entities.User)(nil), repositories.ErrUserNotFound)
	})

	var (
		result *entities.User
		err    error
	)
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.GetUserByEmail(context.Background(), "missing@example.com")
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, services.ErrUserNotFound)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestGetUserByEmail_Error(t provider.T) {
	m := newUserMocks(t)
	expected := errors.New("lookup fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindByEmail", mock.Anything, "user@example.com").Return((*entities.User)(nil), expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.GetUserByEmail(context.Background(), "user@example.com")
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestListAllUsers_Success(t provider.T) {
	m := newUserMocks(t)
	users := []*entities.User{testdata.NewUserBuilder().Build(), testdata.NewUserBuilder().Build()}

	var (
		result []*entities.User
		err    error
	)

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindAll", mock.Anything).Return(users, nil)
	})
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		result, err = m.svc.ListAllUsers(context.Background())
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, users, result)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestListAllUsers_Error(t provider.T) {
	m := newUserMocks(t)
	expected := errors.New("list fail")

	t.WithNewStep("Arrange mocks", func(ctx provider.StepCtx) {
		m.repo.On("FindAll", mock.Anything).Return(nil, expected)
	})

	var err error
	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		_, err = m.svc.ListAllUsers(context.Background())
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
		m.AssertExpectations(t)
	})
}

func (s *UserServiceSuite) TestRegister_ClassicStyle(t provider.T) {
	repo := newInMemoryUserRepo()
	svc := services.NewUserService(repo, zap.NewNop())
	cmd := services.RegisterUserCmd{Username: "alice", Email: "alice@example.com", Password: "secret", Country: "RU", Role: entities.RoleUser}

	var (
		id  uint64
		err error
	)

	t.WithNewStep("Act", func(ctx provider.StepCtx) {
		id, err = svc.Register(context.Background(), cmd)
	})
	t.WithNewStep("Assert", func(ctx provider.StepCtx) {
		require.NoError(t, err)
		assert.Equal(t, uint64(1), id)
		stored, _ := repo.FindByEmail(context.Background(), cmd.Email)
		require.NotNil(t, stored)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(cmd.Password)))
	})
}

func TestUserServiceSuite(t *testing.T) {
	suite.RunSuite(t, new(UserServiceSuite))
}

type inMemoryUserRepo struct {
	nextID  uint64
	users   map[uint64]*entities.User
	byEmail map[string]*entities.User
}

func newInMemoryUserRepo() *inMemoryUserRepo {
	return &inMemoryUserRepo{nextID: 1, users: make(map[uint64]*entities.User), byEmail: make(map[string]*entities.User)}
}

func (r *inMemoryUserRepo) Create(ctx context.Context, u *entities.User) error {
	if _, ok := r.byEmail[u.Email]; ok {
		return repositories.ErrEmailAlreadyExists
	}
	u.ID = r.nextID
	r.nextID++
	r.users[u.ID] = u
	r.byEmail[u.Email] = u
	return nil
}

func (r *inMemoryUserRepo) Update(ctx context.Context, u *entities.User) error {
	if _, ok := r.users[u.ID]; !ok {
		return repositories.ErrUserNotFound
	}
	r.users[u.ID] = u
	r.byEmail[u.Email] = u
	return nil
}

func (r *inMemoryUserRepo) Delete(ctx context.Context, id uint64) error {
	u, ok := r.users[id]
	if !ok {
		return repositories.ErrUserNotFound
	}
	delete(r.users, id)
	delete(r.byEmail, u.Email)
	return nil
}

func (r *inMemoryUserRepo) FindByID(ctx context.Context, id uint64) (*entities.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, repositories.ErrUserNotFound
	}
	return u, nil
}

func (r *inMemoryUserRepo) FindAll(ctx context.Context) ([]*entities.User, error) {
	result := make([]*entities.User, 0, len(r.users))
	for _, u := range r.users {
		result = append(result, u)
	}
	return result, nil
}

func (r *inMemoryUserRepo) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	u, ok := r.byEmail[email]
	if !ok {
		return nil, repositories.ErrUserNotFound
	}
	return u, nil
}

var _ repositories.UserRepository = (*inMemoryUserRepo)(nil)
