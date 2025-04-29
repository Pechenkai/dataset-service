package services

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"ppo/internal/entities"
	"ppo/internal/services"

	"testing"
)

type mockUserRepo struct {
	createErr   error
	findErr     error
	findResp    *entities.User
	updateErr   error
	deleteErr   error
	findAllResp []*entities.User
}

func (m *mockUserRepo) FindAll() ([]*entities.User, error) {
	return m.findAllResp, m.findErr
}

func (m *mockUserRepo) Create(u *entities.User) error { return m.createErr }
func (m *mockUserRepo) FindByEmail(email string) (*entities.User, error) {
	return m.findResp, m.findErr
}
func (m *mockUserRepo) Update(u *entities.User) error             { return m.updateErr }
func (m *mockUserRepo) Delete(id int64) error                     { return m.deleteErr }
func (m *mockUserRepo) FindByID(id int64) (*entities.User, error) { return m.findResp, m.findErr }

func TestUserService_RegisterAndAuthenticate(t *testing.T) {
	// Регистрация: nil-параметр
	svc := services.NewUserService(&mockUserRepo{})
	err := svc.Register(nil)
	if !errors.Is(err, services.ErrNilUser) {
		t.Error("ожидается ошибка при регистрации nil-пользователя")
	}

	// Регистрация: дублирование email
	repoDup := &mockUserRepo{findResp: &entities.User{Email: "a"}}
	svc = services.NewUserService(repoDup)
	err = svc.Register(&entities.User{Email: "a", Password: "pwd"})
	if !errors.Is(err, services.ErrUserExists) {
		t.Error("ожидается ошибка при дублировании email")
	}

	// Регистрация: ошибка при сохранении
	repoErr := &mockUserRepo{findResp: nil, createErr: errors.New("сбой создания")}
	svc = services.NewUserService(repoErr)
	err = svc.Register(&entities.User{Email: "b", Password: "pwd"})
	if err == nil || err.Error() != "сбой создания" {
		t.Errorf("ожидается 'сбой создания', получено: %v", err)
	}

	// Аутентификация: пользователь не найден
	repoNotFound := &mockUserRepo{findResp: nil, findErr: errors.New("нет такого пользователя")}
	svc = services.NewUserService(repoNotFound)
	_, err = svc.Authenticate("no@mail", "pwd")
	if err == nil {
		t.Error("ожидается ошибка при аутентификации несуществующего пользователя")
	}

	// Проверка успешной аутентификации
	hashed, _ := bcrypt.GenerateFromPassword([]byte("goodpass"), bcrypt.DefaultCost)
	repoSuccess := &mockUserRepo{
		findResp: &entities.User{
			Email:    "valid@mail",
			Password: string(hashed),
		},
		findErr: nil,
	}
	svc = services.NewUserService(repoSuccess)
	user, err := svc.Authenticate("valid@mail", "goodpass")
	if err != nil {
		t.Errorf("неожиданная ошибка при успешной аутентификации: %v", err)
	}
	if user.Email != "valid@mail" {
		t.Errorf("ожидается email valid@mail, получено %s", user.Email)
	}
}
