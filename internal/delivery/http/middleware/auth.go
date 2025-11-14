package middleware

import (
	"context"
	"net/http"
	"strings"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/entities"
	"ppo/internal/services"
)

type contextKey string

const userContextKey contextKey = "auth_user"

func BearerAuth(tokenSvc services.TokenService, userSvc services.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r.Header.Get("Authorization"))
			if token == "" {
				dto.WriteStatusError(w, http.StatusUnauthorized, services.ErrInvalidCredentials)
				return
			}

			record, err := tokenSvc.GetToken(r.Context(), token)
			if err != nil {
				dto.WriteStatusError(w, http.StatusUnauthorized, err)
				return
			}

			user, err := userSvc.GetUserByID(r.Context(), record.UserID)
			if err != nil {
				dto.WriteStatusError(w, http.StatusUnauthorized, err)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func CurrentUser(ctx context.Context) *entities.User {
	if ctx == nil {
		return nil
	}
	if user, ok := ctx.Value(userContextKey).(*entities.User); ok {
		return user
	}
	return nil
}
