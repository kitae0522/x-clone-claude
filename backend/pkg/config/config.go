package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL  string
	JWTSecret    string
	JWTExpiryHours int
}

func Load() *Config {
	expiryHours := 24
	if v := os.Getenv("JWT_EXPIRY_HOURS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			expiryHours = parsed
		}
	}

	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/twitter_clone?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production"),
		JWTExpiryHours: expiryHours,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
