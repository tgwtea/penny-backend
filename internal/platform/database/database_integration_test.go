//go:build integration

package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/tgwtea/penny-backend/internal/platform/database"
)

func TestOpenAndPingPostgreSQL(t *testing.T) {
	databaseURL := os.Getenv("PENNY_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("PENNY_TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Open(ctx, database.Config{
		URL:            databaseURL,
		MaxConnections: 2,
		MinConnections: 0,
		ConnectTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
}
