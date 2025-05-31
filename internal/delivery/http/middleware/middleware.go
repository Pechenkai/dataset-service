package middleware

import (
	"errors"
	"go.uber.org/zap"
	"net/http"
	"time"

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

func LoggingMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			// Можно обернуть w, чтобы узнать точный статус-код ответа
			ww := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			logger.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.status),
				zap.Duration("duration", duration),
				zap.String("remote", r.RemoteAddr),
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
