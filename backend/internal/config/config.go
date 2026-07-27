package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                   string
	Port                  string
	CorsOrigins           string
	StaticToken           string
	JWTSecret             string
	DatabaseURL           string
	RequestIDHeader       string
	LogLevel              string
	JWTTTL                time.Duration
	SePayWebhookSecret    string
	TelegramWebhookSecret string
	HermesExecutorSecret  string
	RedisURL              string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:                   os.Getenv("APP_ENV"),
		Port:                  os.Getenv("APP_PORT"),
		CorsOrigins:           os.Getenv("APP_CORS_ORIGINS"),
		StaticToken:           os.Getenv("APP_STATIC_TOKEN"),
		JWTSecret:             os.Getenv("APP_JWT_SECRET"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		RedisURL:              os.Getenv("REDIS_URL"),
		RequestIDHeader:       os.Getenv("REQUEST_ID_HEADER"),
		LogLevel:              os.Getenv("LOG_LEVEL"),
		SePayWebhookSecret:    os.Getenv("SEPAY_WEBHOOK_SECRET"),
		TelegramWebhookSecret: os.Getenv("TELEGRAM_WEBHOOK_SECRET"),
		HermesExecutorSecret:  os.Getenv("HERMES_EXECUTOR_SECRET"),
		JWTTTL:                24 * time.Hour,
	}

	if cfg.Env == "" {
		cfg.Env = "development"
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.CorsOrigins == "" {
		cfg.CorsOrigins = "*"
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "local-dev-secret"
	}
	if cfg.RequestIDHeader == "" {
		cfg.RequestIDHeader = "X-Request-ID"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	if ttlEnv := os.Getenv("APP_JWT_TTL_HOURS"); ttlEnv != "" {
		if ttl, err := strconv.Atoi(ttlEnv); err == nil {
			cfg.JWTTTL = time.Duration(ttl) * time.Hour
		}
	}

	return cfg, nil
}
