package main

import (
	"log/slog"
	"os"

	"github.com/tgwtea/penny-backend/internal/platform/buildinfo"
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

	info := buildinfo.Current()
	logger.Info("migration_starting", "version", info.Version, "commit", info.Commit)
	if err := database.MigrateUp(cfg.MigrationsSource, cfg.DatabaseURL); err != nil {
		logger.Error("migration_failed", "error_category", "migration_failed")
		return 1
	}
	logger.Info("migration_complete")
	return 0
}
