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
}

func Load() (Config, error) {
	cfg := Config{HTTPAddress: value("HTTP_ADDRESS", ":8080"), DatabaseURL: value("DATABASE_URL", "postgres://ai_day:ai_day@localhost:5432/ai_day?sslmode=disable"), AuthMode: value("AUTH_MODE", "phone"), PricePerGenerationPaise: 2000, Currency: value("CURRENCY", "INR"), LogLevel: value("LOG_LEVEL", "info"), LogFormat: value("LOG_FORMAT", "json")}
	if cfg.AuthMode != "phone" && cfg.AuthMode != "microsoft" {
		return Config{}, fmt.Errorf("AUTH_MODE must be phone or microsoft")
	}
	if cfg.Currency != "INR" {
		return Config{}, fmt.Errorf("CURRENCY must be INR")
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
