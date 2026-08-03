package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port              int
	CORSAllowedOrigin []string
	DatabaseURL       string
	B2Endpoint        string
	B2KeyID           string
	B2AppKey          string
	B2Bucket          string
	B2Region          string

	N8NBaseURL      string
	N8NChatPath     string
	N8NChatUsername string
	N8NChatPassword string
	N8NNexxaPath    string
	N8NNexxaSecret  string
	N8NCvPath       string
	N8NCvSecret     string
	N8NTimeout      time.Duration
	AIRateLimit     int
}

func Load() *Config {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080
	}
	timeout, _ := strconv.Atoi(os.Getenv("N8N_TIMEOUT"))
	if timeout == 0 {
		timeout = 115
	}
	rateLimit, _ := strconv.Atoi(os.Getenv("AI_RATE_LIMIT"))
	if rateLimit == 0 {
		rateLimit = 10
	}
	origins := strings.FieldsFunc(getEnv("CORS_ALLOWED_ORIGINS", "*"), func(r rune) bool { return r == ',' })
	for i, o := range origins {
		origins[i] = strings.TrimSpace(o)
	}
	return &Config{
		Port:              port,
		CORSAllowedOrigin: origins,
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://localhost:5432/jhic?sslmode=disable"),
		B2Endpoint:        getEnv("B2_ENDPOINT", "s3.eu-central-003.backblazeb2.com"),
		B2KeyID:           getEnv("B2_KEY_ID", ""),
		B2AppKey:          getEnv("B2_APP_KEY", ""),
		B2Bucket:          getEnv("B2_BUCKET", "jhic-berita-images"),
		B2Region:          getEnv("B2_REGION", ""),

		N8NBaseURL:      getEnv("N8N_BASE_URL", "https://n8n-b0wow8osw0okkcwc0g0gog4o.dev.usbypkp.ac.id"),
		N8NChatPath:     getEnv("N8N_CHAT_PATH", "/webhook/d1b0712b-8783-46ee-8add-5a386132f460/chat"),
		N8NChatUsername: getEnv("N8N_CHAT_USERNAME", ""),
		N8NChatPassword: getEnv("N8N_CHAT_PASSWORD", ""),
		N8NNexxaPath:    getEnv("N8N_NEXXA_PATH", "/webhook/e44f0376-40ef-42f4-980b-ec38e8390592"),
		N8NNexxaSecret:  getEnv("N8N_NEXXA_SECRET", ""),
		N8NCvPath:       getEnv("N8N_CV_PATH", ""),
		N8NCvSecret:     getEnv("N8N_CV_SECRET", ""),
		N8NTimeout:      time.Duration(timeout) * time.Second,
		AIRateLimit:     rateLimit,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
