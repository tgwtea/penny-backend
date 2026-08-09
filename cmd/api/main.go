package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tgwtea/penny-backend/internal/api"
	"github.com/tgwtea/penny-backend/internal/platform/buildinfo"
	"github.com/tgwtea/penny-backend/internal/platform/config"
	"github.com/tgwtea/penny-backend/internal/platform/database"
	"github.com/tgwtea/penny-backend/internal/platform/httpserver"
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
		logger.Error("api_start_failed", "error_category", "database_unavailable")
		return 1
	}
	defer db.Close()

	info := buildinfo.Current()
	handler := api.NewHandler(api.NewServer(db, info, cfg.DatabaseHealthTimeout), logger)
	server := httpserver.New(cfg.HTTPAddress, handler)
	logger.Info("api_starting",
		"address", cfg.HTTPAddress,
		"version", info.Version,
		"commit", info.Commit,
		"artifact_digest", info.ArtifactDigest,
		"contract_version", info.ContractVersion,
		"product_milestone", info.ProductMilestone,
	)

	if err := httpserver.Run(ctx, server, cfg.ShutdownTimeout); err != nil {
		logger.Error("api_stopped_unexpectedly", "error_category", "http_server_failure")
		return 1
	}
	logger.Info("api_stopped")
	return 0
}
