package services

import (
	"errors"
	"testing"

	"ppo/internal/entities"
	"ppo/internal/services"
)

type mockCategoryRepo struct {
	createErr    error
	updateErr    error
	deleteErr    error
	findErr      error
	findAllErr   error
	findByIDResp *entities.Category
	findAllResp  []*entities.Category
}

func (m *mockCategoryRepo) Create(c *entities.Category) error { return m.createErr }
func (m *mockCategoryRepo) Update(c *entities.Category) error { return m.updateErr }
func (m *mockCategoryRepo) Delete(id int64) error             { return m.deleteErr }
func (m *mockCategoryRepo) FindByID(id int64) (*entities.Category, error) {
	return m.findByIDResp, m.findErr
}
func (m *mockCategoryRepo) FindAll() ([]*entities.Category, error) {
	return m.findAllResp, m.findAllErr
}

func TestCategoryService(t *testing.T) {
	// CreateCategory возвращает ошибку при передаче nil
	svc := services.NewCategoryService(&mockCategoryRepo{})
	err := svc.CreateCategory(nil)
	if !errors.Is(err, services.ErrNilCategory) {
		t.Error("ожидается ошибка при создании nil категории")
	}

	// UpdateCategory возвращает ошибку при передаче nil
	err = svc.UpdateCategory(nil)
	if !errors.Is(err, services.ErrNilCategory) {
		t.Error("ожидается ошибка при обновлении nil категории")
	}

	// Ошибка репозитория Create создает ошибку создания категории
	repoErr := &mockCategoryRepo{createErr: errors.New("ошибка создания")}
	svc = services.NewCategoryService(repoErr)
	err = svc.CreateCategory(&entities.Category{})
	if err == nil || err.Error() != "ошибка создания" {
		t.Errorf("ожидается 'ошибка создания', получено: %v", err)
	}

	// Ошибка обновления репозитория Update создает ошибку обновления категории
	repoErr = &mockCategoryRepo{updateErr: errors.New("ошибка обновления")}
	svc = services.NewCategoryService(repoErr)
	err = svc.UpdateCategory(&entities.Category{})
	if err == nil || err.Error() != "ошибка обновления" {
		t.Errorf("ожидается 'ошибка обновления', получено: %v", err)
	}

	// Ошибка удаления репозитория создает ошибку удаления категории
	repoErr = &mockCategoryRepo{deleteErr: errors.New("ошибка удаления")}
	svc = services.NewCategoryService(repoErr)
	err = svc.DeleteCategory(1)
	if err == nil || err.Error() != "ошибка удаления" {
		t.Errorf("ожидается 'ошибка удаления', получено: %v", err)
	}

	// Ошибка поиска по ID
	repoErr = &mockCategoryRepo{findErr: errors.New("ошибка поиска")}
	svc = services.NewCategoryService(repoErr)
	_, err = svc.GetCategoryByID(1)
	if err == nil || err.Error() != "ошибка поиска" {
		t.Errorf("ожидается 'ошибка поиска', получено: %v", err)
	}

	// Проверка поиска по ID
	cat := &entities.Category{ID: 42, Name: "Тест"}
	repoSuccess := &mockCategoryRepo{findByIDResp: cat}
	svc = services.NewCategoryService(repoSuccess)
	got, err := svc.GetCategoryByID(42)
	if err != nil {
		t.Errorf("неожиданная ошибка при GetCategoryByID: %v", err)
	}
	if got != cat {
		t.Errorf("ожидается %v, получено %v", cat, got)
	}

	// Проверка получения всех категорий
	list := []*entities.Category{{ID: 1}, {ID: 2}}
	repoList := &mockCategoryRepo{findAllResp: list}
	svc = services.NewCategoryService(repoList)
	all, err := svc.GetAllCategories()
	if err != nil {
		t.Errorf("неожиданная ошибка при GetAllCategories: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("ожидается 2 категории, получено %d", len(all))
	}
}
