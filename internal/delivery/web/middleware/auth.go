package middleware

import (
	"context"
	"net/http"

	"ppo/internal/delivery/web/session"
)

type ctxKey string

const (
	ctxKeyUserID ctxKey = "userID"
	ctxKeyRole   ctxKey = "role"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, _ := session.Get(r)
		role, ok := sess.Values["role"].(string)
		if !ok {
			role = "guest"
		}
		uid, ok := sess.Values["uid"].(uint64)
		if !ok {
			uid = 0
		}

		ctx := context.WithValue(r.Context(), ctxKeyRole, role)
		ctx = context.WithValue(ctx, ctxKeyUserID, uid)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FromContext(ctx context.Context) (userID uint64, role string) {
	roleVal, ok := ctx.Value(ctxKeyRole).(string)
	if !ok {
		roleVal = "guest"
	}
	uidVal, ok := ctx.Value(ctxKeyUserID).(uint64)
	if !ok {
		uidVal = 0
	}
	return uidVal, roleVal
}
