package middleware

import (
	"net/http"
)

// RequireRole принимает список ролей, которым разрешено войти.
// Если роль из контекста не совпала ни с одной из needRoles → 403 Forbidden.
func RequireRole(needRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(needRoles))
	for _, r := range needRoles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, role := FromContext(r.Context())
			if !allowed[role] {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
