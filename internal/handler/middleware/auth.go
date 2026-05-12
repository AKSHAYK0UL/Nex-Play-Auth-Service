package middleware

import (
	"context"
	"net/http"
	"nex_play_auth/github.com/pkg/jwt"
	"nex_play_auth/github.com/pkg/response"
	"strings"
)

type contextKey string

const (
	ContextKeyUserID    contextKey = "user_id"
	ContextKeyUserEmail contextKey = "user_email"
)

func RequiredAuth(jwt *jwt.Manager) Middleware {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")

			if !strings.HasPrefix(authHeader, "Bearer ") {

				response.Error(w, http.StatusUnauthorized, "missing authorization header")

				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := jwt.Verify(tokenStr)

			if err != nil {

				response.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			//Store claims in context so other handlers can read them
			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)

			ctx = context.WithValue(ctx, ContextKeyUserEmail, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Context accessors
func UserIDFromCtx(ctx context.Context) (int64, bool) {

	id, ok := ctx.Value(ContextKeyUserID).(int64)

	return id, ok
}

func UserEmailFromCtx(ctx context.Context) (string, bool) {

	email, ok := ctx.Value(ContextKeyUserEmail).(string)

	return email, ok
}
