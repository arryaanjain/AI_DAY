package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	HTTPAddress             string
	DatabaseURL             string
	AuthMode                string
	PricePerGenerationPaise int64
	Currency                string
	LogLevel                string
	LogFormat               string
	CORSAllowedOrigins      []string
	// Storage provider settings
	StorageProvider   string
	StorageDisabled   bool
	FilesystemBaseDir string
	S3Endpoint        string
	S3AccessKey       string
	S3SecretKey       string
	S3Bucket          string
	S3Region          string
	S3UseSSL          bool
	// Queue provider settings
	QueueProvider string
	RedisAddr     string
	// AI provider settings
	AIProvider  string
	OpenAIKey   string
	OpenAIModel string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddress:             value("HTTP_ADDRESS", ":8080"),
		DatabaseURL:             value("DATABASE_URL", "postgres://ai_day:ai_day@localhost:5432/ai_day?sslmode=disable"),
		AuthMode:                value("AUTH_MODE", "phone"),
		PricePerGenerationPaise: 2000,
		Currency:                value("CURRENCY", "INR"),
		LogLevel:                value("LOG_LEVEL", "info"),
		LogFormat:               value("LOG_FORMAT", "json"),
		CORSAllowedOrigins:      csv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174"),
		StorageProvider:         value("STORAGE_PROVIDER", "minio"),
		StorageDisabled:         value("STORAGE_DISABLED", "false") == "true",
		FilesystemBaseDir:       value("FILESYSTEM_BASE_DIR", "./storage"),
		S3Endpoint:              value("S3_ENDPOINT", "localhost:9000"),
		S3AccessKey:             value("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey:             value("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:                value("S3_BUCKET", "ai-day"),
		S3Region:                value("S3_REGION", "us-east-1"),
		S3UseSSL:                value("S3_USE_SSL", "false") == "true",
		QueueProvider:           value("QUEUE_PROVIDER", "redis"),
		RedisAddr:               value("REDIS_ADDR", "localhost:6379"),
		AIProvider:              value("AI_PROVIDER", "mock"),
		OpenAIKey:               value("OPENAI_API_KEY", ""),
		OpenAIModel:             value("OPENAI_IMAGE_MODEL", "dall-e-3"),
	}

	if cfg.AuthMode != "phone" && cfg.AuthMode != "microsoft" {
		return Config{}, fmt.Errorf("AUTH_MODE must be phone or microsoft")
	}
	if cfg.Currency != "INR" {
		return Config{}, fmt.Errorf("CURRENCY must be INR")
	}
	if len(cfg.CORSAllowedOrigins) == 0 {
		return Config{}, fmt.Errorf("CORS_ALLOWED_ORIGINS must contain at least one origin")
	}

	return cfg, nil
}

func NewLogger(cfg Config) *slog.Logger {
	level := new(slog.LevelVar)
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		level.Set(slog.LevelInfo)
	}
	options := &slog.HandlerOptions{Level: level}
	if cfg.LogFormat == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, options))
}

func value(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func csv(key, fallback string) []string {
	raw := value(key, fallback)
	values := strings.Split(raw, ",")
	origins := make([]string, 0, len(values))
	for _, item := range values {
		if origin := strings.TrimRight(strings.TrimSpace(item), "/"); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
