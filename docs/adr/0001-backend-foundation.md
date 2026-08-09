# ADR-0001: Backend foundation implementation choices

- Status: Accepted
- Date: 2026-08-09
- Scope: Decision-neutral repository and application foundation

## Context

The approved baseline fixes Go, PostgreSQL, pgx/sqlc, backend-owned OpenAPI,
separate API/worker processes, forward-compatible migrations, structured logs,
Kubernetes deployment, and build-once promotion. It does not select the small
implementation tools needed to make those confirmed boundaries executable.

## Decision

- Use Go 1.26.5 and the module path `github.com/tgwtea/penny-backend`.
- Use the standard `net/http` server and `log/slog` JSON handler, avoiding a web
  framework before route requirements justify one.
- Use pgx v5, sqlc, and explicit SQL. A generated `SELECT 1` query powers the
  initial readiness check.
- Use oapi-codegen for committed Go transport models and server interfaces.
- Use golang-migrate behind a Penny-owned forward-only executable. API and
  worker startup never migrate, and the executable exposes no down/force path.
- Build separate minimal, non-root API, worker, and migration OCI targets from
  one source revision. Release CI records their immutable registry digests.
- Supply an image digest as deployment configuration because an image cannot
  embed its own content digest without changing that digest.

All exact versions are pinned in `go.mod`, `go.sum`, and the Docker build
arguments. Dependabot may propose changes, but normal review and verification
remain required.

## Consequences

Generated transport and SQL code is committed and checked for drift. Runtime
dependencies remain small. PostgreSQL integration behaviour can be tested
against a real server without SQLite substitutions. The migration image is
available before the first schema change, reducing release-flow rework.

This ADR does not resolve any architecture open question. In particular, it
does not choose database placement, backup tooling, cluster secret management,
tunnel/TLS placement, domains/CORS, or a PDF library.

## Compatibility, recovery, and tests

This is the first backend artifact and contract, so there is no previous API or
schema to migrate. Future schema changes must remain compatible with the
previous API and worker. Application rollback uses previous image digests;
database recovery uses reviewed forward correction rather than automatic down
migrations.

CI regenerates code, compares released `/api/v1` contracts, lints, runs unit
tests with the race detector, scans dependencies, builds executables, and builds
all OCI targets. PostgreSQL connectivity has an opt-in integration test.
