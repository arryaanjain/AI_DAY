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

	"github.com/joho/godotenv"
	"github.com/arryaanjain/AI_DAY/internal/config"
	"github.com/arryaanjain/AI_DAY/internal/database"
	"github.com/arryaanjain/AI_DAY/internal/httpapi"
)

func main() {
	 _ = godotenv.Load() // loads .env from current working directory
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	logger := config.NewLogger(cfg)
	db, err := database.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Warn("database unavailable; readiness will remain unavailable", "error", err)
	}
	if db != nil {
		defer db.Close()
	}
	server := &http.Server{Addr: cfg.HTTPAddress, Handler: httpapi.NewRouter(cfg, db, logger), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
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
