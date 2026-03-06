package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Env             string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiryHours  int
	MediaServiceURL string
	InternalAPIKey  string
}

func Load() (*Config, error) {
	env := getEnv("APP_ENV", "development")

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		if env == "production" {
			return nil, fmt.Errorf("JWT_SECRET environment variable is required in production")
		}
		jwtSecret = "dev-jwt-secret-change-in-production"
	}

	expiryHours := 24
	if v := os.Getenv("JWT_EXPIRY_HOURS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			expiryHours = parsed
		}
	}

	return &Config{
		Env:             env,
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/twitter_clone?sslmode=disable"),
		JWTSecret:       jwtSecret,
		JWTExpiryHours:  expiryHours,
		MediaServiceURL: getEnv("MEDIA_SERVICE_URL", "http://localhost:8081"),
		InternalAPIKey:  getEnv("INTERNAL_API_KEY", "dev-internal-api-key-change-in-production"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
