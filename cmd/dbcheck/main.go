package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/tgwtea/penny-backend/internal/platform/config"
	"github.com/tgwtea/penny-backend/internal/platform/database"
	"github.com/tgwtea/penny-backend/internal/platform/logging"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("configuration_invalid", "error", err.Error())
		return 1
	}
	logger, err := logging.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("configuration_invalid", "error", err.Error())
		return 1
	}

	db, err := database.Open(context.Background(), database.Config{
		URL:            cfg.DatabaseURL,
		MaxConnections: cfg.DatabaseMaxConnections,
		MinConnections: cfg.DatabaseMinConnections,
		ConnectTimeout: cfg.DatabaseConnectTimeout,
	})
	if err != nil {
		logger.Error("database_check_failed", "error_category", "database_unavailable")
		return 1
	}
	db.Close()
	logger.Info("database_check_succeeded")
	return 0
}
