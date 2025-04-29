package services

import (
	"errors"
	"ppo/internal/entities"
	"ppo/internal/services"
	"testing"
)

type mockNotifRepo struct {
	createErr    error
	updateErr    error
	findErr      error
	deleteErr    error
	findByIDResp *entities.Notification
	byUser       []*entities.Notification
}

func (m *mockNotifRepo) Delete(id int64) error {
	return m.deleteErr
}

func (m *mockNotifRepo) Create(n *entities.Notification) error { return m.createErr }
func (m *mockNotifRepo) Update(n *entities.Notification) error { return m.updateErr }
func (m *mockNotifRepo) FindByID(id int64) (*entities.Notification, error) {
	return m.findByIDResp, m.findErr
}
func (m *mockNotifRepo) FindByUserID(id int64) ([]*entities.Notification, error) {
	return m.byUser, m.findErr
}

// stub subscriptionRepo

type mockSubRepo struct {
	subs []int64
	err  error
}

func (m *mockSubRepo) GetSubscribers(datasetID int64) ([]int64, error)       { return m.subs, m.err }
func (m *mockSubRepo) Subscribe(userID, datasetID int64) error               { return nil }
func (m *mockSubRepo) IsSubscribed(userID, datasetID int64) (bool, error)    { return true, nil }
func (m *mockSubRepo) FindByUser(id int64) ([]*entities.Subscription, error) { return nil, nil }
func (m *mockSubRepo) Unsubscribe(userID, datasetID int64) error             { return nil }

func TestNotificationService_NotifyAndRead(t *testing.T) {
	// NotifySubscribers: ошибка создания уведомления
	nrErr := &mockNotifRepo{createErr: errors.New("сбой создания уведомления")}
	srStub := &mockSubRepo{subs: []int64{1, 2}}
	svc := services.NewNotificationService(nrErr, srStub)
	err := svc.NotifySubscribers(10, "новая версия v1.1")
	if err == nil {
		t.Error("ожидается ошибка при рассылке уведомлений")
	}

	// NotifySubscribers: успешная рассылка
	nrOK := &mockNotifRepo{}
	svc = services.NewNotificationService(nrOK, srStub)
	err = svc.NotifySubscribers(10, "новая версия v1.1")
	if err != nil {
		t.Errorf("неожиданная ошибка при успешной рассылке: %v", err)
	}

	// GetNotificationsByUser: ошибка получения
	nrErrFetch := &mockNotifRepo{findErr: errors.New("сбой получения уведомлений")}
	svc = services.NewNotificationService(nrErrFetch, srStub)
	_, err = svc.GetNotificationsByUser(1)
	if err == nil {
		t.Error("ожидается ошибка при получении уведомлений")
	}
}
