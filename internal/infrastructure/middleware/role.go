package middleware

import (
	"context"
	"net/http"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/response"
)

type RoleChecker func(ctx context.Context, userID string) (string, error)

func RequireRole(allowedRoles ...string) func(RoleChecker) func(http.Handler) http.Handler {
	return func(check RoleChecker) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID, ok := r.Context().Value(UserIDKey).(string)
				if !ok || userID == "" {
					response.Error(w, http.StatusUnauthorized, "unauthorized")
					return
				}

				role, err := check(r.Context(), userID)
				if err != nil {
					response.Error(w, http.StatusInternalServerError, "failed to verify role")
					return
				}

				for _, allowed := range allowedRoles {
					if role == allowed {
						next.ServeHTTP(w, r)
						return
					}
				}

				response.Error(w, http.StatusForbidden, "insufficient permissions")
			})
		}
	}
}
