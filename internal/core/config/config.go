package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config holds all application configuration
type Config struct {
	ServerPort string
	ServerHost string
	AppEnv     string

	DatabaseURL              string
	JWTSecret                string
	AccessTokenExpireMinutes int
}

// LoadConfig loads configuration from .env file or environment variables
func LoadConfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil && !os.IsNotExist(err) {
		logrus.Errorf("Error loading .env file: %v", err)
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	jwtExpire, err := strconv.Atoi(getEnv("ACCESS_TOKEN_EXPIRE_MINUTES", "60"))
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_EXPIRE_MINUTES in .env: %w", err)
	}

	return &Config{
		ServerPort:               getEnv("SERVER_PORT", "8080"),
		ServerHost:               getEnv("SERVER_HOST", "0.0.0.0"),
		AppEnv:                   getEnv("APP_ENV", "development"),
		DatabaseURL:              getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/superaib_db?sslmode=disable"),
		JWTSecret:                getEnv("JWT_SECRET", "supersecretjwtsecretkey"),
		AccessTokenExpireMinutes: jwtExpire,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
