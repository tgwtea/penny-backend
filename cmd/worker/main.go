package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := database.Open(ctx, database.Config{
		URL:            cfg.DatabaseURL,
		MaxConnections: cfg.DatabaseMaxConnections,
		MinConnections: cfg.DatabaseMinConnections,
		ConnectTimeout: cfg.DatabaseConnectTimeout,
	})
	if err != nil {
		logger.Error("worker_start_failed", "error_category", "database_unavailable")
		return 1
	}
	defer db.Close()

	info := buildinfo.Current()
	logger.Info("worker_started",
		"version", info.Version,
		"commit", info.Commit,
		"artifact_digest", info.ArtifactDigest,
		"contract_version", info.ContractVersion,
		"product_milestone", info.ProductMilestone,
		"claiming_enabled", false,
	)

	// The skeleton worker does not claim business jobs. It only keeps its
	// PostgreSQL dependency observable until job handlers are implemented.
	ticker := time.NewTicker(cfg.WorkerHealthInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Info("worker_stopped")
			return 0
		case <-ticker.C:
			checkCtx, cancel := context.WithTimeout(ctx, cfg.DatabaseHealthTimeout)
			err := db.Ping(checkCtx)
			cancel()
			if err != nil {
				logger.Warn("worker_dependency_unavailable", "dependency", "postgresql")
			}
		}
	}
}
