package config

import (
	"os"
	"time"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	TokenTTL    time.Duration
}

func Load() *Config {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:Elemisb00%23@localhost:5432/stack%2Ddata?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "super_secret_jwt_key_change_me")

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
		TokenTTL:    7 * 24 * time.Hour,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
