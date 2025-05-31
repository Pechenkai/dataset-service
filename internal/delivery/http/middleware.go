package http

import (
	"errors"
	"net/http"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/services"
)

type HandlerWithError func(w http.ResponseWriter, r *http.Request) error

func mapErrorToStatus(err error) int {
	switch {
	// ================================
	// === 400 Bad Request: ошибки валидации ===
	// ================================
	case errors.Is(err, services.ErrNilReview),
		errors.Is(err, services.ErrInvalidRating),
		errors.Is(err, services.ErrNilUser),
		errors.Is(err, services.ErrNilDataset),
		errors.Is(err, services.ErrInvalidMetadata),
		errors.Is(err, services.ErrNilCategory):
		return http.StatusBadRequest

	// ==================================================
	// === 401 Unauthorized: неверные данные аутентификации ===
	// ==================================================
	case errors.Is(err, services.ErrInvalidCredentials),
		errors.Is(err, services.ErrInvalidPassword):
		return http.StatusUnauthorized

	// =======================================================
	// === 404 Not Found: когда чего-то не существует       ===
	// =======================================================
	case errors.Is(err, services.ErrReviewNotFound),
		errors.Is(err, services.ErrUserNotFound),
		errors.Is(err, services.ErrDatasetNotFound),
		errors.Is(err, services.ErrVersionNotFound),
		errors.Is(err, services.ErrCategoryNotFound),
		errors.Is(err, services.ErrNotificationNotFound),
		errors.Is(err, services.ErrNotSubscribed):
		return http.StatusNotFound

	// ================================================================
	// === 409 Conflict: когда мы пытаемся создать «то, что уже есть» ===
	// ================================================================
	case errors.Is(err, services.ErrReviewNotFound): // (смотри примечание ниже)
		// на самом деле, ErrReviewNotFound здесь не нужна – оставлено для примера
		return http.StatusNotFound
	case errors.Is(err, services.ErrCategoryExists),
		errors.Is(err, services.ErrUserExists),
		errors.Is(err, services.ErrAlreadySubscribed):
		return http.StatusConflict

	// ======================================================
	// === 424 Failed Dependency: нет подписчиков к оповещению ===
	// ======================================================
	case errors.Is(err, services.ErrNoSubscribers):
		return http.StatusFailedDependency

	default:
		return http.StatusInternalServerError
	}
}
func WrapHandler(h HandlerWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			status := mapErrorToStatus(err)
			dto.WriteStatusError(w, status, err)
		}
	}
}
