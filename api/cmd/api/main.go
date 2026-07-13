package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arryaanjain/AI_DAY/internal/ai"
	"github.com/arryaanjain/AI_DAY/internal/config"
	"github.com/arryaanjain/AI_DAY/internal/database"
	"github.com/arryaanjain/AI_DAY/internal/httpapi"
	"github.com/arryaanjain/AI_DAY/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // loads .env from current working directory

	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	logger := config.NewLogger(cfg)

	// Initialize database
	db, err := database.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Warn("database unavailable; readiness will remain unavailable", "error", err)
	}
	if db != nil {
		defer db.Close()
	}

	// Initialize storage provider
	var storageProvider storage.Provider
	if cfg.StorageDisabled {
		logger.Info("storage provider disabled", "provider", "disabled")
	} else if cfg.StorageProvider == "filesystem" {
		storageProvider, err = storage.NewFilesystemProvider(cfg.FilesystemBaseDir)
		if err != nil {
			logger.Error("failed to initialize filesystem storage provider", "error", err)
			os.Exit(1)
		}
		logger.Info("storage provider initialized", "provider", "filesystem", "basedir", cfg.FilesystemBaseDir)
	} else if cfg.StorageProvider == "noop" {
		storageProvider = storage.NewNoopProvider()
		logger.Info("storage provider initialized", "provider", "noop")
	} else if cfg.StorageProvider == "minio" || cfg.StorageProvider == "s3" {
		storageProvider, err = storage.NewMinIOProvider(
			cfg.S3Endpoint,
			cfg.S3AccessKey,
			cfg.S3SecretKey,
			cfg.S3Bucket,
			cfg.S3Region,
			cfg.S3UseSSL,
		)
		if err != nil {
			logger.Error("failed to initialize storage provider", "error", err)
			os.Exit(1)
		}
		logger.Info("storage provider initialized", "provider", cfg.StorageProvider)
	} else {
		logger.Warn("unknown storage provider, defaulting to noop", "provider", cfg.StorageProvider)
		storageProvider = storage.NewNoopProvider()
	}

	// Initialize AI provider
	var aiProvider ai.Provider
	if cfg.AIProvider == "openai" && cfg.OpenAIKey != "" {
		aiProvider = ai.NewOpenAIProvider(cfg.OpenAIKey, cfg.OpenAIModel, cfg.OpenAIChatModel)
		logger.Info("ai provider initialized", "provider", "openai", "chatModel", cfg.OpenAIChatModel, "imageModel", cfg.OpenAIModel)
	} else {
		aiProvider = ai.NewMockProvider()
		logger.Info("ai provider initialized", "provider", "mock")
	}

	// Create HTTP server
	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           httpapi.NewRouter(cfg, db, logger, storageProvider, aiProvider),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("api listening", "address", cfg.HTTPAddress)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
