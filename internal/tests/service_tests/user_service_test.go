package services_test

import (
	"context"
	"ppo/internal/dataaccess/repositories/postgres"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"ppo/internal/entities"
	"ppo/internal/services"
	"ppo/internal/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeClock struct{ now time.Time }

func (f fakeClock) Now() time.Time { return f.now }

// === Register ===

func TestRegister_Success(t *testing.T) {
	repo := new(mocks.UserRepository)
	clk := fakeClock{now: time.Date(2025, 5, 24, 10, 0, 0, 0, time.UTC)}
	svc := services.NewUserService(repo, clk)

	cmd := services.RegisterUserCmd{
		Username: "alice",
		Email:    "a@b.com",
		Password: "secret",
		Country:  "NL",
		Role:     entities.RoleUser,
	}

	// 1) FindByEmail -> no existing
	repo.On("FindByEmail", mock.Anything, "a@b.com").Return((*entities.User)(nil), nil)
	// 2) Create -> assign ID
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *entities.User) bool {
		// пароль уже захеширован
		err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("secret"))
		return err == nil &&
			u.Username == "alice" &&
			u.Email == "a@b.com" &&
			u.Country == "NL" &&
			u.Role == entities.RoleUser
	})).Run(func(args mock.Arguments) {
		args.Get(1).(*entities.User).ID = 77
	}).Return(nil)

	id, err := svc.Register(context.Background(), cmd)
	assert.NoError(t, err)
	assert.Equal(t, uint64(77), id)

	repo.AssertExpectations(t)
}

func TestRegister_EmptyUsername(t *testing.T) {
	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})

	_, err := svc.Register(context.Background(), services.RegisterUserCmd{
		Username: "   ",
		Email:    "a@b.com",
		Password: "pass",
		Country:  "US",
		Role:     entities.RoleUser,
	})
	assert.ErrorIs(t, err, entities.ErrEmptyUsername)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})

	repo.On("FindByEmail", mock.Anything, "dup@em.com").
		Return(&entities.User{ID: 1}, nil)

	_, err := svc.Register(context.Background(), services.RegisterUserCmd{
		Username: "bob",
		Email:    "dup@em.com",
		Password: "p",
		Country:  "",
		Role:     entities.RoleUser,
	})
	assert.ErrorIs(t, err, services.ErrUserExists)
}

// === Authenticate ===

func TestAuthenticate_Success(t *testing.T) {
	// подготовим юзера с захешированным паролем
	hash, _ := bcrypt.GenerateFromPassword([]byte("mypwd"), bcrypt.DefaultCost)
	stored := &entities.User{ID: 5, Email: "x@y", Password: string(hash)}

	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})

	repo.On("FindByEmail", mock.Anything, "x@y").Return(stored, nil)

	u, err := svc.Authenticate(context.Background(), services.AuthenticateUserCmd{
		Email:    "x@y",
		Password: "mypwd",
	})
	assert.NoError(t, err)
	assert.Equal(t, stored, u)
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.DefaultCost)
	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})
	repo.On("FindByEmail", mock.Anything, "x@y").Return(&entities.User{Password: string(hash)}, nil)

	_, err := svc.Authenticate(context.Background(), services.AuthenticateUserCmd{
		Email:    "x@y",
		Password: "badpwd",
	})
	assert.ErrorIs(t, err, services.ErrInvalidCredentials)
}

func TestAuthenticate_NotFound(t *testing.T) {
	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})
	repo.On("FindByEmail", mock.Anything, "none@x").Return(nil, nil)

	_, err := svc.Authenticate(context.Background(), services.AuthenticateUserCmd{
		Email:    "none@x",
		Password: "anything",
	})
	assert.ErrorIs(t, err, services.ErrUserNotFound)
}

// === UpdateUser ===

func TestUpdateUser_Success(t *testing.T) {
	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})

	existing := &entities.User{
		ID:               10,
		Username:         "old",
		Email:            "a@a",
		Password:         "hash",
		Country:          "oldC",
		IsBlocked:        false,
		Role:             entities.RoleUser,
		RegistrationDate: time.Now(),
	}
	repo.On("FindByID", mock.Anything, uint64(10)).Return(existing, nil)
	// email не меняем; password update
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entities.User) bool {
		matchPwd := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("newpass")) == nil
		return u.ID == 10 && u.Username == "new" && matchPwd && u.Country == "newC" &&
			u.IsBlocked && u.Role == entities.RoleAdmin
	})).Return(nil)

	err := svc.UpdateUser(context.Background(), services.UpdateUserCmd{
		ID:        10,
		Username:  "new",
		Password:  "newpass",
		Country:   "newC",
		IsBlocked: true,
		Role:      entities.RoleAdmin,
	})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateUser_NotFound(t *testing.T) {
	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})
	repo.On("FindByID", mock.Anything, uint64(99)).Return(nil, nil)

	err := svc.UpdateUser(context.Background(), services.UpdateUserCmd{ID: 99})
	assert.ErrorIs(t, err, services.ErrUserNotFound)
}

func TestUpdateUser_DuplicateEmail(t *testing.T) {
	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})

	existing := &entities.User{ID: 1, Email: "old@e"}
	repo.On("FindByID", mock.Anything, uint64(1)).Return(existing, nil)
	repo.On("FindByEmail", mock.Anything, "new@e").Return(&entities.User{ID: 2}, nil)

	err := svc.UpdateUser(context.Background(), services.UpdateUserCmd{
		ID:    1,
		Email: "new@e",
	})
	assert.ErrorIs(t, err, services.ErrUserExists)
}

// === DeleteUser ===

func TestDeleteUser_Success(t *testing.T) {
	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})

	repo.On("Delete", mock.Anything, uint64(3)).Return(nil)
	err := svc.DeleteUser(context.Background(), 3)
	assert.NoError(t, err)
}

func TestDeleteUser_NotFound(t *testing.T) {
	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})

	repo.On("Delete", mock.Anything, uint64(4)).Return(postgres.ErrUserNotFound)
	err := svc.DeleteUser(context.Background(), 4)
	assert.ErrorIs(t, err, services.ErrUserNotFound)
}

// === GetUserByID ===

func TestGetUserByID_Success(t *testing.T) {
	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})

	expected := &entities.User{ID: 5}
	repo.On("FindByID", mock.Anything, uint64(5)).Return(expected, nil)

	u, err := svc.GetUserByID(context.Background(), 5)
	assert.NoError(t, err)
	assert.Equal(t, expected, u)
}

func TestGetUserByID_NotFound(t *testing.T) {
	repo := new(mocks.UserRepository)
	svc := services.NewUserService(repo, fakeClock{now: time.Now()})

	repo.On("FindByID", mock.Anything, uint64(6)).Return(nil, nil)

	_, err := svc.GetUserByID(context.Background(), 6)
	assert.ErrorIs(t, err, services.ErrUserNotFound)
}
