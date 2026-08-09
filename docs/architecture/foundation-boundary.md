# Foundation skeleton boundary

## Included

This change establishes infrastructure-neutral application foundations:

- Go module and feature-based package layout;
- separate API and worker executables with graceful shutdown;
- a dedicated forward-only migration executable;
- PostgreSQL connection pooling, generated health SQL, and a connection checker;
- backend-owned OpenAPI 3 contract and pinned generated Go transport;
- RFC 9457 Problem Details schema and safe request-ID/error middleware;
- JSON logging that avoids URLs with query strings, bodies, and credentials;
- liveness independent from dependencies and readiness dependent on PostgreSQL;
- pinned sqlc, OpenAPI, lint, vulnerability, and compatibility tools;
- unit and opt-in real-PostgreSQL integration-test structure;
- reproducible non-root API, worker, and migration container targets;
- CI generation drift, compatibility, formatting, lint, race, vulnerability,
  build, and container checks; and
- a release workflow that publishes build-once images and records their digests.

## Feature and risk impact

| Concern | Foundation effect |
| --- | --- |
| API | Adds only `/health/live` and `/health/ready`; no `/api/v1` product route. |
| Data | Adds no business table or migration. The only SQL query is `SELECT 1` for readiness. |
| Authorisation | Adds no authenticated route, Firebase adapter, membership lookup, or workspace repository. Package ownership is reserved for future contract-first work. |
| Financial history | No invoice, payment, audit, money, numbering, snapshot, or PDF behaviour. |
| Background jobs | Worker starts and observes PostgreSQL but explicitly does not claim or execute jobs. |
| External side effects | No Firebase, Resend, PDF, webhook, or other network provider integration. |
| Deployment | Produces immutable OCI artifacts and provenance metadata only; no cluster resources or environment secret values. |

## Explicitly deferred decision gates

- OQ-01/OQ-02: no PostgreSQL StatefulSet/operator, storage, backup, WAL, or restore resources.
- OQ-04: no Kubernetes Secret or secret-store integration.
- OQ-05: no cloudflared, Traefik route, TLS, or certificate resources.
- OQ-06: no guessed domain names or CORS policy. Browser integration remains blocked.
- OQ-07: no PDF library or implementation.
- OQ-08: no cleanup or retention policy.
- OQ-09: no invoice issuance or presentation behaviour.
- OQ-10: no deletion or anonymisation behaviour.

The gated items require ADRs or a baseline update before implementation. The
container contract and environment-variable boundary are intended to remain
valid regardless of which permitted infrastructure options are selected.

