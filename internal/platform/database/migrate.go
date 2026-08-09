package database

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var ErrMigrationFailed = errors.New("database migration failed")

// MigrateUp applies pending forward migrations. No down/force operation is
// exposed by the Penny executable, preventing automatic destructive rollback.
func MigrateUp(sourceURL, databaseURL string) error {
	migrator, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return ErrMigrationFailed
	}

	migrationErr := migrator.Up()
	sourceErr, databaseErr := migrator.Close()
	if migrationErr != nil && !errors.Is(migrationErr, migrate.ErrNoChange) {
		return ErrMigrationFailed
	}
	if sourceErr != nil || databaseErr != nil {
		return ErrMigrationFailed
	}
	return nil
}
