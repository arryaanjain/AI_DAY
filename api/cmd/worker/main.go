package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/arryaanjain/AI_DAY/internal/ai"
	"github.com/arryaanjain/AI_DAY/internal/assets"
	"github.com/arryaanjain/AI_DAY/internal/config"
	"github.com/arryaanjain/AI_DAY/internal/database"
	"github.com/arryaanjain/AI_DAY/internal/storage"
	"github.com/arryaanjain/AI_DAY/internal/worker"
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
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

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
		aiProvider = ai.NewOpenAIProvider(ai.OpenAIConfig{
			APIKey:             cfg.OpenAIKey,
			ImageModel:         cfg.OpenAIModel,
			ChatModel:          cfg.OpenAIChatModel,
			ImageSize:          cfg.OpenAIImageSize,
			ImageQuality:       cfg.OpenAIImageQuality,
			OutputRequirements: cfg.OpenAIOutputRequirements,
			OrgID:              cfg.OpenAIOrgID,
			ProjectID:          cfg.OpenAIProjectID,
			BaseURL:            cfg.OpenAIBaseURL,
		})
		logger.Info("ai provider initialized", "provider", "openai", "chatModel", cfg.OpenAIChatModel, "imageModel", cfg.OpenAIModel)
	} else {
		aiProvider = ai.NewMockProvider()
		logger.Info("ai provider initialized", "provider", "mock")
	}

	// Create asset service
	assetService := assets.New(db, storageProvider, "ai-day", 10*1024*1024, []string{"image/jpeg", "image/png", "image/webp"})

	// Create worker
	w := worker.NewWorker(db, aiProvider, storageProvider, assetService, logger)

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("shutdown signal received")
		cancel()
	}()

	// Start processing jobs
	w.Start(ctx)
	logger.Info("worker shutdown complete")
}
