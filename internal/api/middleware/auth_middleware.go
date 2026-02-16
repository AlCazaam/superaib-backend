package middleware

import (
	"context"
	"net/http"
	"strings"
	"superaib/internal/api/response"
	"superaib/internal/core/config"
	"superaib/internal/core/security"

	"github.com/google/uuid"
)

type AuthMiddleware struct {
	cfg *config.Config
}

func NewAuthMiddleware(cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg}
}

// 1. Authenticate: Waxaa loo isticmaalaa Dashboard-ka (JWT Token)
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(w, http.StatusUnauthorized, "Invalid format (Bearer <token>)")
			return
		}

		claims, err := security.ValidateJWT(parts[1])
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		userIDString, _ := claims["user_id"].(string)
		if _, err := uuid.Parse(userIDString); err != nil {
			response.Error(w, http.StatusUnauthorized, "Invalid user ID format")
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userIDString)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
