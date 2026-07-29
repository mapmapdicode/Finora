package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                    string
	Port                   string
	CorsOrigins            string
	StaticToken            string
	JWTSecret              string
	DatabaseURL            string
	RequestIDHeader        string
	LogLevel               string
	JWTTTL                 time.Duration
	SePayWebhookSecret     string
	SePayBankHubClientID   string
	SePayBankHubSecret     string
	SePayBankHubCompanyID  string
	SePayBankHubAPIKey     string
	SePayBankHubBaseURL    string
	SePayBankHubPilotBanks []string
	TelegramWebhookSecret  string
	HermesExecutorSecret   string
	SMTPHost               string
	SMTPPort               string
	SMTPUsername           string
	SMTPPassword           string
	SMTPFrom               string
	SMTPFromName           string
	RedisURL               string
	SeedDemoData           bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:                    os.Getenv("APP_ENV"),
		Port:                   os.Getenv("APP_PORT"),
		CorsOrigins:            os.Getenv("APP_CORS_ORIGINS"),
		StaticToken:            os.Getenv("APP_STATIC_TOKEN"),
		JWTSecret:              os.Getenv("APP_JWT_SECRET"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		RedisURL:               os.Getenv("REDIS_URL"),
		SeedDemoData:           strings.ToLower(strings.TrimSpace(os.Getenv("APP_SEED_DEMO"))) != "false",
		RequestIDHeader:        os.Getenv("REQUEST_ID_HEADER"),
		LogLevel:               os.Getenv("LOG_LEVEL"),
		SePayWebhookSecret:     os.Getenv("SEPAY_WEBHOOK_SECRET"),
		SePayBankHubClientID:   os.Getenv("SEPAY_BANKHUB_CLIENT_ID"),
		SePayBankHubSecret:     os.Getenv("SEPAY_BANKHUB_CLIENT_SECRET"),
		SePayBankHubCompanyID:  os.Getenv("SEPAY_BANKHUB_COMPANY_XID"),
		SePayBankHubAPIKey:     os.Getenv("SEPAY_BANKHUB_IPN_API_KEY"),
		SePayBankHubBaseURL:    os.Getenv("SEPAY_BANKHUB_BASE_URL"),
		SePayBankHubPilotBanks: splitCSV(os.Getenv("SEPAY_BANKHUB_PILOT_BANK_CODES")),
		TelegramWebhookSecret:  os.Getenv("TELEGRAM_WEBHOOK_SECRET"),
		HermesExecutorSecret:   os.Getenv("HERMES_EXECUTOR_SECRET"),
		SMTPHost:               os.Getenv("SMTP_HOST"),
		SMTPPort:               os.Getenv("SMTP_PORT"),
		SMTPUsername:           os.Getenv("SMTP_USERNAME"),
		SMTPPassword:           os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:               os.Getenv("SMTP_FROM"),
		SMTPFromName:           os.Getenv("SMTP_FROM_NAME"),
		JWTTTL:                 24 * time.Hour,
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
	if cfg.SePayBankHubBaseURL == "" {
		cfg.SePayBankHubBaseURL = "https://bankhub-api.sepay.vn"
	}
	if strings.EqualFold(cfg.Env, "production") && (strings.TrimSpace(cfg.SMTPHost) == "" || strings.TrimSpace(cfg.SMTPFrom) == "") {
		return nil, fmt.Errorf("SMTP_HOST and SMTP_FROM are required in production for email verification")
	}

	if ttlEnv := os.Getenv("APP_JWT_TTL_HOURS"); ttlEnv != "" {
		if ttl, err := strconv.Atoi(ttlEnv); err == nil {
			cfg.JWTTTL = time.Duration(ttl) * time.Hour
		}
	}

	return cfg, nil
}

func splitCSV(value string) []string {
	items := strings.Split(value, ",")
	out := make([]string, 0, len(items))
	for _, item := range items {
		if normalized := strings.ToUpper(strings.TrimSpace(item)); normalized != "" {
			out = append(out, normalized)
		}
	}
	return out
}
