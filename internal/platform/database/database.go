package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tgwtea/penny-backend/internal/platform/database/dbgen"
)

var (
	ErrInvalidConfiguration = errors.New("invalid PostgreSQL configuration")
	ErrUnavailable          = errors.New("PostgreSQL unavailable")
)

// Config defines bounded pool and startup behaviour without exposing provider
// or deployment details.
type Config struct {
	URL            string
	MaxConnections int32
	MinConnections int32
	ConnectTimeout time.Duration
}

// DB owns the shared pgx pool and generated decision-neutral SQL queries.
type DB struct {
	pool    *pgxpool.Pool
	queries *dbgen.Queries
}

// Open creates and verifies a PostgreSQL pool. It intentionally returns safe,
// stable errors so a connection URL can never be copied into logs.
func Open(ctx context.Context, cfg Config) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, ErrInvalidConfiguration
	}
	poolConfig.MaxConns = cfg.MaxConnections
	poolConfig.MinConns = cfg.MinConnections

	connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, ErrUnavailable
	}

	db := &DB{pool: pool, queries: dbgen.New(pool)}
	if err := db.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, ErrUnavailable
	}
	return db, nil
}

// Ping executes a generated, read-only SQL query against PostgreSQL.
func (db *DB) Ping(ctx context.Context) error {
	if _, err := db.queries.Ping(ctx); err != nil {
		return ErrUnavailable
	}
	return nil
}

// Close releases all PostgreSQL connections.
func (db *DB) Close() {
	db.pool.Close()
}
