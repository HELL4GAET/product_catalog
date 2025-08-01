package auth

import (
	"net/http"
	"product-catalog/internal/errors"
	"strings"
)

type Middleware struct {
	jwtManager *Manager
}

func NewMiddleware(jwtManager *Manager) *Middleware {
	return &Middleware{jwtManager: jwtManager}
}

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, errors.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := m.jwtManager.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, errors.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		ctx := WithUserContext(r.Context(), claims.UserID, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
