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

func MapErrorToStatus(err error) int {
	if isBadRequestError(err) || matchesAny(err, badRequestErrors) {
		return http.StatusBadRequest
	}
	if matchesAny(err, unauthorizedErrors) {
		return http.StatusUnauthorized
	}
	if matchesAny(err, forbiddenErrors) {
		return http.StatusForbidden
	}
	if matchesAny(err, notFoundErrors) {
		return http.StatusNotFound
	}
	if matchesAny(err, conflictErrors) {
		return http.StatusConflict
	}
	if matchesAny(err, failedDependencyErrors) {
		return http.StatusFailedDependency
	}
	return http.StatusInternalServerError
}

func WrapHandler(h HandlerWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			status := MapErrorToStatus(err)
			dto.WriteStatusError(w, status, err)
		}
	}
}

func isBadRequestError(err error) bool {
	if err == nil {
		return false
	}
	var badReq *dto.BadRequestError
	return errors.As(err, &badReq)
}

var badRequestErrors = []error{
	services.ErrNilReview,
	services.ErrInvalidRating,
	services.ErrNilUser,
	services.ErrNilDataset,
	services.ErrInvalidMetadata,
	services.ErrNilCategory,
	services.ErrBadRequest,
}

var unauthorizedErrors = []error{
	services.ErrInvalidCredentials,
	services.ErrInvalidPassword,
	services.ErrInvalidTwoFACode,
	services.ErrTwoFAExpired,
	services.ErrTokenInvalid,
}

var forbiddenErrors = []error{
	services.ErrUserBlocked,
	services.ErrTwoFADebugDisabled,
	services.ErrRequestForbidden,
}

var notFoundErrors = []error{
	services.ErrReviewNotFound,
	services.ErrUserNotFound,
	services.ErrDatasetNotFound,
	services.ErrVersionNotFound,
	services.ErrCategoryNotFound,
	services.ErrNotificationNotFound,
	services.ErrNotSubscribed,
	services.ErrRequestNotFound,
	services.ErrTokenNotFound,
	services.ErrTwoFAChallengeNotFound,
}

var conflictErrors = []error{
	services.ErrCategoryExists,
	services.ErrUserExists,
	services.ErrAlreadySubscribed,
	services.ErrRequestAlreadyExists,
}

var failedDependencyErrors = []error{
	services.ErrNoSubscribers,
}

func matchesAny(err error, targets []error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
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
