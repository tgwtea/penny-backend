# Local development

## Configuration

Copy `.env.example` for reference, but load variables using your shell or local
environment manager. The application deliberately does not parse `.env` files,
which avoids introducing environment-specific secret-loading behaviour.

`PENNY_DATABASE_URL` is required by all executables. Never use Pilot credentials
locally. The application does not log the URL or raw PostgreSQL errors.

## PostgreSQL connection check

```powershell
$env:PENNY_DATABASE_URL = 'postgres://penny:penny@localhost:5432/penny?sslmode=disable'
go run ./cmd/dbcheck
```

Success emits one JSON `database_check_succeeded` log. Failure emits only the
safe category `database_unavailable` and exits non-zero.

## API and worker

```powershell
go run ./cmd/api
go run ./cmd/worker
```

The worker currently claims no jobs. Both processes validate PostgreSQL during
startup and shut down on interrupt/termination. The API's liveness check remains
healthy during a transient PostgreSQL outage while readiness becomes `503`.

## Tests

Unit tests do not require PostgreSQL. To exercise the real pgx/sqlc path:

```powershell
$env:PENNY_TEST_DATABASE_URL = 'postgres://penny:penny@localhost:5432/penny_test?sslmode=disable'
go test -race -tags=integration ./...
```

Use a dedicated disposable test database. The current integration test is
read-only and executes `SELECT 1`; future feature suites may manage fixtures and
transactions but must never target shared Staging or Pilot data.

