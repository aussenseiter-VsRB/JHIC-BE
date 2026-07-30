package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	DatabaseURL string
	B2Endpoint  string
	B2KeyID     string
	B2AppKey    string
	B2Bucket    string
	B2Region    string
}

func Load() *Config {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080
	}
	return &Config{
		Port:        port,
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost:5432/jhic?sslmode=disable"),
		B2Endpoint:  getEnv("B2_ENDPOINT", "s3.eu-central-003.backblazeb2.com"),
		B2KeyID:     getEnv("B2_KEY_ID", ""),
		B2AppKey:    getEnv("B2_APP_KEY", ""),
		B2Bucket:    getEnv("B2_BUCKET", "jhic-berita-images"),
		B2Region:    getEnv("B2_REGION", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
