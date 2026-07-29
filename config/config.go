package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	DatabaseURL string
}

func Load() *Config {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080
	}
	return &Config{
		Port:        port,
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/jhic?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
