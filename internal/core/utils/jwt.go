package utils

import (
	"fmt"

	"superaib/internal/core/config"
	"superaib/internal/models"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateJWT creates a new JWT token for a given user
func GenerateJWT(user *models.User, cfg *config.Config) (string, error) {
	claims := jwt.StandardClaims{
		Subject:   user.ID.String(),
		ExpiresAt: time.Now().Add(time.Minute * time.Duration(cfg.AccessTokenExpireMinutes)).Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "superaib-backend",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}
