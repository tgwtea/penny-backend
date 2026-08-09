# Contributing

Read `AGENTS.md`, the relevant architecture-baseline sections, the current
OpenAPI contract, owning module, SQL/migrations, and tests before editing.

Keep changes within one feature or platform boundary. State API, data,
authorisation, financial-history, background-job, and deployment effects in the
pull request template. Stop at unresolved open-question gates.

For public API changes:

1. Change `openapi.yaml` first and preserve `/api/v1` compatibility.
2. Run `go generate ./...`.
3. Implement thin transport mapping in the owning feature.
4. Add contract and behaviour tests.
5. Run `make check-contract` and `make check-generated`.

For schema changes, add a new expand-and-contract migration. Never edit a
migration already applied in any shared environment. SQL is the source of truth;
regenerate sqlc output instead of editing generated files.

Before handoff run the smallest relevant checks, followed by `make verify` and
applicable PostgreSQL integration tests. Do not include unrelated refactors,
dependency updates, or formatting churn.

