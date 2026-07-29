package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/response"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type TokenValidator func(ctx context.Context, token string) (string, error)

func Auth(validate TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				response.Error(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}
			token := strings.TrimPrefix(auth, "Bearer ")

			userID, err := validate(r.Context(), token)
			if err != nil || userID == "" {
				response.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
