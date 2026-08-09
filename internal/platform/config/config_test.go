package config_test

import (
	"errors"
	"testing"
	"time"

	"github.com/tgwtea/penny-backend/internal/platform/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PENNY_DATABASE_URL", "postgres://example.invalid/penny")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddress != ":8080" {
		t.Errorf("HTTPAddress = %q, want :8080", cfg.HTTPAddress)
	}
	if cfg.DatabaseHealthTimeout != 2*time.Second {
		t.Errorf("DatabaseHealthTimeout = %s, want 2s", cfg.DatabaseHealthTimeout)
	}
	if cfg.DatabaseMaxConnections != 10 || cfg.DatabaseMinConnections != 1 {
		t.Errorf("connection bounds = %d/%d, want 1/10", cfg.DatabaseMinConnections, cfg.DatabaseMaxConnections)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("PENNY_DATABASE_URL", "")

	_, err := config.Load()
	if !errors.Is(err, config.ErrDatabaseURLRequired) {
		t.Fatalf("Load() error = %v, want ErrDatabaseURLRequired", err)
	}
}

func TestLoadRejectsInvalidBounds(t *testing.T) {
	t.Setenv("PENNY_DATABASE_URL", "postgres://example.invalid/penny")
	t.Setenv("PENNY_DATABASE_MIN_CONNECTIONS", "11")
	t.Setenv("PENNY_DATABASE_MAX_CONNECTIONS", "10")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid connection bounds")
	}
}
