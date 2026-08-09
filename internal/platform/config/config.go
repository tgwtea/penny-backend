package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPAddress         = ":8080"
	defaultLogLevel            = "info"
	defaultMaxConnections      = int32(10)
	defaultMinConnections      = int32(1)
	defaultConnectTimeout      = 5 * time.Second
	defaultHealthTimeout       = 2 * time.Second
	defaultWorkerCheckInterval = 30 * time.Second
	defaultShutdownTimeout     = 15 * time.Second
	defaultMigrationsSource    = "file://db/migrations"
)

var ErrDatabaseURLRequired = errors.New("PENNY_DATABASE_URL is required")

// Config contains process configuration shared by the API, worker, migration,
// and database-check executables. It deliberately excludes unresolved CORS,
// provider, routing, and secret-management choices.
type Config struct {
	HTTPAddress            string
	LogLevel               string
	DatabaseURL            string
	DatabaseMaxConnections int32
	DatabaseMinConnections int32
	DatabaseConnectTimeout time.Duration
	DatabaseHealthTimeout  time.Duration
	WorkerHealthInterval   time.Duration
	ShutdownTimeout        time.Duration
	MigrationsSource       string
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		HTTPAddress:            valueOrDefault("PENNY_HTTP_ADDRESS", defaultHTTPAddress),
		LogLevel:               valueOrDefault("PENNY_LOG_LEVEL", defaultLogLevel),
		DatabaseURL:            os.Getenv("PENNY_DATABASE_URL"),
		DatabaseMaxConnections: defaultMaxConnections,
		DatabaseMinConnections: defaultMinConnections,
		DatabaseConnectTimeout: defaultConnectTimeout,
		DatabaseHealthTimeout:  defaultHealthTimeout,
		WorkerHealthInterval:   defaultWorkerCheckInterval,
		ShutdownTimeout:        defaultShutdownTimeout,
		MigrationsSource:       valueOrDefault("PENNY_MIGRATIONS_SOURCE", defaultMigrationsSource),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, ErrDatabaseURLRequired
	}

	var err error
	if cfg.DatabaseMaxConnections, err = positiveInt32("PENNY_DATABASE_MAX_CONNECTIONS", defaultMaxConnections); err != nil {
		return Config{}, err
	}
	if cfg.DatabaseMinConnections, err = nonNegativeInt32("PENNY_DATABASE_MIN_CONNECTIONS", defaultMinConnections); err != nil {
		return Config{}, err
	}
	if cfg.DatabaseMinConnections > cfg.DatabaseMaxConnections {
		return Config{}, errors.New("PENNY_DATABASE_MIN_CONNECTIONS must not exceed PENNY_DATABASE_MAX_CONNECTIONS")
	}

	if cfg.DatabaseConnectTimeout, err = positiveDuration("PENNY_DATABASE_CONNECT_TIMEOUT", defaultConnectTimeout); err != nil {
		return Config{}, err
	}
	if cfg.DatabaseHealthTimeout, err = positiveDuration("PENNY_DATABASE_HEALTH_TIMEOUT", defaultHealthTimeout); err != nil {
		return Config{}, err
	}
	if cfg.WorkerHealthInterval, err = positiveDuration("PENNY_WORKER_HEALTH_INTERVAL", defaultWorkerCheckInterval); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = positiveDuration("PENNY_SHUTDOWN_TIMEOUT", defaultShutdownTimeout); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func valueOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func positiveDuration(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s must be a positive Go duration", name)
	}
	return duration, nil
}

func positiveInt32(name string, fallback int32) (int32, error) {
	value, err := nonNegativeInt32(name, fallback)
	if err != nil {
		return 0, err
	}
	if value == 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}
	return value, nil
}

func nonNegativeInt32(name string, fallback int32) (int32, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative 32-bit integer", name)
	}
	return int32(value), nil
}
