package middleware

import (
	"context"
	"net/http"
	"strings"

	"library-backend/internal/handlers"
)

func AuthMiddleware(h *handlers.Handlers, secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), "user", "demo-user")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
