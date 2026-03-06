package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port       int
	JWTSecret  string
	APIKey     string
	S3Endpoint string
	S3Bucket   string
	S3Region   string
	S3KeyID    string
	S3Secret   string
	S3UseSSL   bool
	MaxWorkers int
	TempDir    string
}

func Load() (*Config, error) {
	port, _ := strconv.Atoi(getEnv("PORT", "8081"))
	maxWorkers, _ := strconv.Atoi(getEnv("MAX_WORKERS", "4"))
	useSSL, _ := strconv.ParseBool(getEnv("S3_USE_SSL", "false"))

	cfg := &Config{
		Port:       port,
		JWTSecret:  getEnv("JWT_SECRET", ""),
		APIKey:     getEnv("INTERNAL_API_KEY", ""),
		S3Endpoint: getEnv("S3_ENDPOINT", ""),
		S3Bucket:   getEnv("S3_BUCKET", ""),
		S3Region:   getEnv("S3_REGION", "us-east-1"),
		S3KeyID:    getEnv("S3_ACCESS_KEY_ID", ""),
		S3Secret:   getEnv("S3_SECRET_ACCESS_KEY", ""),
		S3UseSSL:   useSSL,
		MaxWorkers: maxWorkers,
		TempDir:    getEnv("TEMP_DIR", "/tmp/media-service"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.S3Endpoint == "" {
		return nil, fmt.Errorf("S3_ENDPOINT is required")
	}
	if cfg.S3Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required")
	}
	if cfg.S3KeyID == "" {
		return nil, fmt.Errorf("S3_ACCESS_KEY_ID is required")
	}
	if cfg.S3Secret == "" {
		return nil, fmt.Errorf("S3_SECRET_ACCESS_KEY is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("INTERNAL_API_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
