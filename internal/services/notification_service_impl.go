package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ppo/internal/entities"
	"ppo/internal/repositories"
)

type notificationService struct {
	notifRepo repositories.NotificationRepository
	subRepo   repositories.SubscriptionRepository
}

// NewNotificationService создаёт экземпляр NotificationService.
func NewNotificationService(
	notifRepo repositories.NotificationRepository,
	subRepo repositories.SubscriptionRepository,
) NotificationService {
	return &notificationService{
		notifRepo: notifRepo,
		subRepo:   subRepo,
	}
}

// NotifySubscribers рассылает сообщение всем подписчикам указанного датасета.
func (s *notificationService) NotifySubscribers(ctx context.Context, cmd NotifySubscribersCmd) (int, error) {
	// 1) Получаем список userID подписчиков из репозитория
	subscribers, err := s.subRepo.GetSubscribers(ctx, cmd.DatasetID)
	if err != nil {
		// Если репозиторий вернул «свою» ошибку, пробуем её распознать:
		// (например, никакой специальной ошибки «нет подписчиков» на уровне repozitory).
		// Здесь просто возвращаем 0, обернув в fmt.Errorf, чтобы точно понять источник.
		return 0, fmt.Errorf("fetch subscribers: %w", err)
	}

	if len(subscribers) == 0 {
		return 0, ErrNoSubscribers
	}

	count := 0
	for _, userID := range subscribers {
		// 2) Создаём новую сущность Notification
		notif, err := entities.NewNotification(userID, cmd.DatasetID, cmd.Message, time.Now().UTC())
		if err != nil {
			// Ошибка валидации модели Notification
			return count, fmt.Errorf("invalid notification for user %d: %w", userID, err)
		}

		// 3) Пытаемся сохранить уведомление через репозиторий
		if err := s.notifRepo.Create(ctx, notif); err != nil {
			// Если репозиторий вернул ErrNotificationCreate (или какую-либо иную известную ошибку),
			// мы можем либо преобразовать её в собственную ошибку, либо обернуть:
			// Например, если сохранить не удалось из-за каких-то проблем, мы возвращаем ошибку,
			// но атаку «notification create» отражаем минимально:
			return count, fmt.Errorf("create notification for user %d: %w", userID, err)
		}

		count++
	}
	return count, nil
}

// GetNotificationsByUser возвращает все уведомления для данного пользователя.
func (s *notificationService) GetNotificationsByUser(ctx context.Context, userID uint64) ([]*entities.Notification, error) {
	notifs, err := s.notifRepo.FindByUserID(ctx, userID)
	if err != nil {
		// Если репозиторий вернул ErrNotificationQueryBuild / ErrNotificationScanRow / ErrNotificationIterateRows,
		// здесь мы оборачиваем их, чтобы контроллер знал, что это «внутренняя» ошибка чтения.
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	return notifs, nil
}

// MarkAsRead устанавливает флаг «прочитано» для конкретного уведомления.
func (s *notificationService) MarkAsRead(ctx context.Context, notificationID uint64) error {
	// 1) Сначала пытаемся получить уведомление по ID
	notif, err := s.notifRepo.FindByID(ctx, notificationID)
	if err != nil {
		// Если репозиторий вернул ErrNotificationNotFound, преобразуем его в сервисную ошибку
		if errors.Is(err, repositories.ErrNotificationNotFound) {
			return ErrNotificationNotFound
		}
		// Иначе – это какая-то иная внутреняя ошибка чтения
		return fmt.Errorf("fetch notification: %w", err)
	}
	if notif == nil {
		// Теоретически FindByID никогда не вернёт (nil, nil) – если нотификация не найдена, репозиторий вернёт ErrNotificationNotFound.
		return ErrNotificationNotFound
	}

	// 2) Отмечаем «прочитано»
	notif.IsRead = true
	if err := s.notifRepo.Update(ctx, notif); err != nil {
		// Если при обновлении репозиторий вернул ErrNotificationNotFound – снова возвращаем сервисную ErrNotificationNotFound
		if errors.Is(err, repositories.ErrNotificationNotFound) {
			return ErrNotificationNotFound
		}
		// Иначе – какая-то внутренняя ошибка при обновлении
		return fmt.Errorf("mark as read: %w", err)
	}
	return nil
}
