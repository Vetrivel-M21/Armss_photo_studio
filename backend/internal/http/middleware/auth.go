package middleware

import (
	"context"
	"net/http"

	"armss-poster-studio/internal/auth"
	"armss-poster-studio/internal/http/respond"
	"armss-poster-studio/internal/store/mysql"
)

type userCtxKey struct{}

// RequireAuth resolves the session cookie to a user and stores it on the request
// context. Missing or invalid sessions get a 401, never a panic downstream.
func RequireAuth(manager *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(manager.CookieName())
			if err != nil {
				respond.Error(w, http.StatusUnauthorized, "You need to be logged in for this.")
				return
			}
			user, err := manager.UserFromToken(r.Context(), cookie.Value)
			if err != nil {
				respond.Error(w, http.StatusUnauthorized, "Your session has expired. Please log in again.")
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin must run after RequireAuth. It rejects staff users from admin-only routes.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil || user.Role != mysql.RoleAdmin {
			respond.Error(w, http.StatusForbidden, "This action requires an admin account.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserFromContext returns the authenticated user set by RequireAuth, or nil.
func UserFromContext(ctx context.Context) *mysql.User {
	u, _ := ctx.Value(userCtxKey{}).(*mysql.User)
	return u
}
