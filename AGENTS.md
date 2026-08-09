# AGENTS.md — Penny Backend

## Purpose

This file governs coding agents working in the `penny-backend` repository. Build the Penny Pilot backend described by:

- `PENNY_TECHNICAL_ARCHITECTURE_BASELINE.md` — approved technical source of truth.
- `MVP-FEATURES.md` — current product scope.

If a repository document uses a different path, locate it before changing code. Do not rely on remembered copies.

The backend is a Go feature-based modular monolith with separate API and worker executables, a versioned REST/JSON contract, PostgreSQL business state, and Kubernetes deployment. Optimise for correctness, auditability, workspace isolation, safe retries, and a small maintainable Pilot—not speculative scale.

## Instruction and decision precedence

Follow decisions in this order:

1. The current user request and explicit acceptance criteria.
2. This `AGENTS.md`.
3. Confirmed decisions in `PENNY_TECHNICAL_ARCHITECTURE_BASELINE.md`.
4. The backend-owned `openapi.yaml` for transport behaviour.
5. Existing ADRs, tests, and repository conventions.
6. `MVP-FEATURES.md` for product scope.

When sources conflict, stop and report the conflict. Do not silently select the easiest interpretation.

The architecture baseline uses binding markers:

- **Confirmed:** implement as written.
- **Open:** do not choose a default. Record or request the decision before affected work begins.
- **High impact / Breaking:** require an ADR, migration and compatibility plan, rollback or forward-recovery plan, affected-system list, and test strategy.
- **Superseded:** never reintroduce unless a new ADR explicitly reverses the decision.

## Mandatory working method

For every change:

1. Read the relevant baseline, MVP, OpenAPI, module, migration, and tests before editing.
2. State the feature boundary and identify affected API, data, authorisation, financial-history, background-job, and deployment concerns.
3. Check the open-question and high-impact registers. Stop at unresolved decision gates.
4. For API work, update the backend-owned OpenAPI contract first. Keep `/api/v1` backward compatible.
5. For schema work, design an expand-and-contract migration that remains compatible with the previous known-good API and worker.
6. Implement in the owning feature module. Keep transport, application, domain, and persistence responsibilities separated.
7. Add or update unit, PostgreSQL integration, contract, authorisation, idempotency, and worker tests as applicable.
8. Regenerate pinned generated code and verify the committed output has no drift. Never hand-edit generated files.
9. Run the smallest relevant checks during development, then the repository’s full required checks before handoff.
10. Report changed behaviour, migrations, contract impact, operational impact, tests run, and any unresolved risks.

Do not bundle unrelated refactors, dependency upgrades, formatting churn, infrastructure choices, or product expansion into a feature change.

## Fixed architecture

Do not replace or bypass these decisions without an approved ADR:

- Go API and Go worker, sharing business modules but using separate entry points.
- Feature-based modular monolith; no Pilot microservices.
- JSON REST API under `/api/v1`.
- Backend-owned OpenAPI specification and pinned code generation for Go and Dart.
- `sqlc` with `pgx`; reviewed SQL is the query source of truth.
- One private PostgreSQL database as the source of truth for all Penny business and authorisation data.
- Firebase Authentication with Google Sign-In only. Firebase proves identity; PostgreSQL grants workspace access.
- Cloudflare Pages for Flutter Web. Do not build or serve frontend assets from the backend.
- Cloudflare Tunnel and Traefik for external API routing. Never expose PostgreSQL publicly.
- Kubernetes for API, worker, migration Jobs, and CronJobs.
- Resend behind a small provider interface for transactional email only.
- PostgreSQL-backed durable jobs and transactional outbox with at-least-once processing.
- Backend-generated, on-demand invoice PDFs from immutable versioned snapshots.
- Build once, test in Staging, and promote the exact artifact to Pilot.

Do not introduce Firestore, email magic-link authentication, an application-managed session system, a payment gateway, object storage, RLS, a new queue service, a frontend pod, or a new external service unless explicitly authorised by a new decision. The MVP permits recording payments; it does not define automated payment collection or bank-proof verification.

## Repository shape and module ownership

Preserve this intended structure unless the repository already contains an equivalent documented structure:

```text
cmd/
  api/
  worker/

internal/
  auth/
  users/
  students/
  lessons/
  billing/
  invoices/
  payments/
  notifications/
  publicsummary/
  jobs/

internal/platform/
  database/
  logging/
  email/
  pdf/
  firebase/
```

Each feature owns its handlers, application services, domain rules, repository code, transport mapping, and tests. `internal/platform` contains technical adapters, not product rules.

Dependencies should point inward:

- HTTP handlers parse and validate transport input, call an application service, and map the result to OpenAPI responses.
- Application services define business transactions and orchestration.
- Domain code contains invariants and state transitions without HTTP, Firebase, Resend, Kubernetes, or generated-SQL concerns.
- Repositories enforce workspace-scoped persistence and expose purpose-specific methods.
- Provider adapters implement small interfaces owned by the consuming feature or application layer.

Do not place business logic in generated code, handlers, SQL row mappers, `cmd/`, CronJobs, or provider adapters. Do not duplicate business implementations between the API and worker.

## API and OpenAPI rules

`openapi.yaml` is authoritative for routes, schemas, validation, authentication, and expected errors.

- Preserve backward compatibility inside `/api/v1`.
- Do not remove or rename fields or endpoints, change field meaning/type/format, or strengthen authentication requirements silently.
- Make new request fields optional unless a versioned migration explicitly says otherwise.
- New response fields are allowed; design clients to tolerate unknown fields.
- Treat enum additions as compatibility work and define old-client behaviour.
- Keep deprecated fields until all supported clients have migrated.
- Use `/api/v2` or an explicit compatibility period for deliberate breaking changes.
- Commit generated Go output, mark it generated, and never edit it manually.
- Pin generators and commands. CI must regenerate and fail on drift.
- Compare changes with the latest released `/api/v1` contract and block accidental breaks.

Use RFC 9457 Problem Details for errors. Include:

- a stable Penny `code` for client behaviour;
- field-level `errors` for validation failures; and
- a request ID for unexpected failures, without stack traces or secrets.

Document expected authentication, authorisation, validation, conflict, rate-limit, and service-unavailable responses in OpenAPI.

Use cursor pagination for growing lessons, invoices, payments, jobs, and audit events. Ordering must be deterministic; cursors opaque; default and maximum limits documented. Do not expose raw offsets as durable cursors.

Use optimistic concurrency for mutable records, normally with a `version` field. A stale update returns `409 VERSION_CONFLICT`; never use last-write-wins for consequential state.

## Authentication and workspace authorisation

For every authenticated request:

1. Verify the Firebase ID token using the official server-side mechanism.
2. Extract the Firebase UID.
3. Resolve the Penny user and active workspace membership from PostgreSQL.
4. Authorise the requested action against that membership.
5. Pass a required workspace context into application and repository calls.

Non-negotiable isolation rules:

- Every student, lesson, invoice, payment, setting, job, and audit record is workspace-owned.
- Every workspace-owned query and mutation includes `workspace_id`.
- Repository APIs require workspace context; it must not be optional or inferred from an entity UUID.
- UUID knowledge never grants access.
- Use composite constraints or equivalent schema protections across workspace boundaries where appropriate.
- Never trust a client-supplied workspace ID without matching it to authenticated membership.
- Add negative integration tests that attempt cross-workspace reads, updates, deletes, associations, cursor reuse, and background-job access.
- Do not enable PostgreSQL RLS during the Pilot without an approved architecture change; keep schema and APIs compatible with adding it later.

Public summary endpoints are the only no-login business endpoints. They use token possession as narrowly scoped access and must never reuse authenticated repository paths without explicit public-safe projection.

## Data representation rules

Use these representations end to end:

| Concern | Required representation |
| --- | --- |
| Internal entity IDs | UUIDv7 in PostgreSQL `uuid`; opaque strings over APIs |
| Provider IDs | Separate provider-specific columns; never primary business IDs |
| Money | Integer minor units (`bigint` / Go `int64`) plus ISO 4217 currency |
| Precise time | UTC timestamps |
| Workspace time | Separate IANA time-zone name |
| Calendar dates | Date type for invoice issue and due dates |
| Recurrence | Local schedule time plus workspace zone |
| Lesson occurrence | Precise start and end instants |

Never use floating point for money. Use currency metadata for scale and rounding; do not assume every currency has two decimal places. The Flutter device time zone is never the billing source of truth.

Validate invariants both in Go and, when structurally expressible, with PostgreSQL constraints. Prefer explicit state transitions over arbitrary status writes.

## Financial-history rules

Treat invoice, payment, and audit records as financial history, not ordinary CRUD.

- Draft invoices may change. Finalisation is a one-way transition into an immutable payer-facing snapshot.
- A finalised snapshot contains approved lesson lines, totals, currency, issue and due dates, payment instructions, display details, invoice number, locale/formatting rules, and PDF template version.
- Never reconstruct or recalculate a historical invoice from mutable student, tutor, schedule, rate, or workspace settings.
- Finalised invoices cannot be deleted or casually edited. Corrections use explicit cancellation, credit/correction, or replacement flows defined by the product and contract.
- Recorded payments cannot be deleted. Mistakes use reversal or correction records.
- An issued invoice number is never reused.
- Students and recurring schedules are archived rather than erased when history depends on them.
- Delete an unused draft only when no dependent records exist.
- Audit events are append-only and contain limited non-secret context.
- Do not promise workspace deletion or anonymisation until OQ-10 is resolved.

Assign invoice numbers atomically at finalisation using a yearly workspace-scoped sequence such as `2026-0001`. Enforce uniqueness on `(workspace_id, invoice_year, invoice_sequence)`. The UUIDv7 remains the internal identifier. Concurrent finalisation must not create duplicate numbers; any stronger no-gap requirement depends on OQ-09’s legal and presentation review.

Consequential writes require `Idempotency-Key`, including:

- invoice finalisation;
- payment recording and reversal;
- email sending;
- reminder preparation and sending; and
- public-link rotation.

Bind an idempotency record to the authenticated workspace/actor, operation, key, and a canonical request fingerprint. Reusing a key with a different request must fail deterministically. Retain enough status and response data to return the original result for a safe retry. Database uniqueness remains the final duplicate-prevention layer. The retention period is unresolved by OQ-08; do not hard-code a production cleanup period without that decision.

## Public payment summaries

The public endpoint shape is:

```text
GET /api/v1/public/summaries/{token}
```

Implement public links as bearer secrets:

- Generate long tokens with a cryptographically secure random source.
- Store only a one-way hash suitable for exact lookup; never store or log the plaintext token.
- Never use an invoice UUID, sequential number, or guessable value as the credential.
- Show only an allowlisted projection of the immutable invoice snapshot.
- Exclude private notes, internal metadata, raw provider data, user identity internals, and tutor-only fields.
- Rotation atomically installs the new hash and immediately invalidates the old link.
- Revoked, invalid, missing, inaccessible, and malformed tokens receive the same generic unavailable response. Avoid behavioural or timing distinctions where practical.
- Apply application-level rate limits even when Cloudflare also rate-limits.
- Set caching and referrer behaviour conservatively so token-bearing URLs and responses do not leak through shared caches or analytics.

Public-link access does not constitute payment, tutor verification, signature, contractual acceptance, or legal proof unless a separately approved product/legal design defines those semantics. Do not invent such a workflow from the shareable-page MVP bullet.

## Transactions, outbox, jobs, and side effects

Start transactions at the application-service boundary when one business action changes multiple records. Keep transactions short and deterministic.

When a committed business action requires an external side effect, write the business state, audit event, and outbox event in the same PostgreSQL transaction. Never call Resend, Firebase, PDF clients, or other networks inside that transaction.

Outbox and job processing must assume at-least-once delivery:

- claim work safely for multiple workers;
- make each handler idempotent;
- use bounded exponential backoff with jitter;
- classify transient and permanent failures;
- record attempts and safe failure categories;
- move exhausted work to a visible dead-letter state;
- allow a safe, auditable manual retry after the cause is fixed; and
- retain processed events for a troubleshooting window, with cleanup deferred until OQ-08 is resolved.

Kubernetes CronJobs only initiate periodic scans. PostgreSQL owns durable job state. CronJobs must not directly implement billing, overdue, reminder, or cleanup business rules.

For Pilot reminders, the system may identify and prepare an overdue reminder, but a tutor must approve it before sending. PostgreSQL determines whether an invoice is due, unpaid, and eligible; Resend does not.

## Email and webhook rules

Keep Resend behind a small Go interface and make provider behaviour replaceable in tests. Email delivery is an asynchronous side effect driven through outbox/jobs.

For Resend webhooks:

- verify the official signature and timestamp against the untouched raw request body before parsing or acting;
- reject invalid or stale signatures;
- deduplicate using the provider event ID and a database constraint;
- acknowledge valid events quickly;
- persist only the minimal durable fields needed;
- hand slower work to the background worker; and
- do not retain raw payloads by default.

Never use Resend for authentication.

## PDF rules

- Generate the official invoice PDF in Go from the stored immutable snapshot.
- Keep Flutter preview state out of official document generation.
- Store and select a PDF template version for every finalised invoice.
- A design change creates a new template version; it must not alter historical output.
- Make output deterministic for the same snapshot and template version.
- Exclude private notes and internal metadata.
- Generate on demand for the Pilot; do not add persistent object storage without a decision.
- Do not select the PDF library until OQ-07 is resolved and its acceptance criteria are met.

Use golden or structural tests to prove historical reproducibility while avoiding brittle assertions on irrelevant binary metadata.

## SQL, migrations, and persistence

- Write explicit SQL and generate access code with pinned `sqlc` and `pgx` versions.
- Never edit `sqlc` output manually.
- Use PostgreSQL integration tests for query, constraint, locking, transaction, and concurrency behaviour. Mocks alone are insufficient.
- Keep applied migrations immutable. Add a new corrective migration instead of editing history.
- Keep seed and fixture data separate from migrations.
- Run migrations only as a dedicated Kubernetes Job, never automatically from API or worker startup.
- Use expand-and-contract: expand schema, deploy compatible code, backfill, migrate all clients/workers, then contract after the rollback window.
- Do not add a migration that prevents the previous known-good API or worker from running without explicit high-impact review.
- Do not use automatic down-migrations against Pilot data. Prefer application rollback plus a reviewed forward corrective migration.
- Design backfills to be restartable, observable, bounded, and safe under concurrent old/new code where applicable.

The exact PostgreSQL deployment mechanism and backup tool remain unresolved. Do not create production StatefulSet/operator or backup manifests until OQ-01 and OQ-02 are decided.

## Validation, security, and privacy

Treat Cloudflare as one layer, not as application security.

- Validate content type, shape, size, ranges, state transitions, and identifiers in the API.
- Apply endpoint-specific application rate limits to public lookups, authenticated traffic, invoice/reminder sending, PDF generation, and webhooks.
- Return `429 RATE_LIMITED` and `Retry-After` when useful.
- Use strict CORS allowlists for approved Local, Staging, Pilot, and explicitly authorised preview origins. Never combine credentialed requests with wildcard origins.
- Do not guess domains or preview policy; OQ-06 must be resolved before integration.
- Keep Pilot credentials away from preview deployments by default.
- Never place secrets in Git, frontend assets, images, logs, analytics, traces, error responses, or audit metadata.
- Separate Local, Staging, and Pilot credentials.
- Redact Firebase tokens, public-summary tokens, credentials, private notes, unnecessary emails, payment details, webhook signatures, and raw payloads.
- Use parameterised SQL only.
- Bound request bodies and expensive operations.
- Use constant-time comparison where comparing secret-derived values directly is unavoidable.
- Avoid exposing entity existence through authorisation or public-token error differences.
- Keep dependencies minimal and pinned through normal Go module controls; justify new runtime dependencies.

Do not claim a workflow is tamper-proof or legally binding. Implement verifiable records, immutable snapshots, append-only events, authenticated actions, timestamps, and hashes precisely, then defer legal meaning and market-specific invoice presentation to an approved legal/product decision. OQ-09 must be resolved before Pilot invoice issuance.

## Logging and operational behaviour

Emit structured JSON logs with consistent field names across API and worker.

- Include request, job, invoice, and delivery identifiers when relevant.
- Include workspace or actor identifiers only when necessary and safe; never include bearer secrets.
- Distinguish expected validation/business failures from unexpected application failures.
- Propagate request IDs through application and job boundaries.
- Make dead-letter jobs, retries, webhook failures, and backup failures visible.
- Use graceful API shutdown and worker shutdown. A worker stops claiming new jobs and completes or safely releases active work.

Implement health endpoints and probes deliberately:

- liveness means the process should be restarted;
- readiness means it can safely receive traffic;
- do not make liveness depend on a transient external provider;
- expose deployed version/commit/artifact metadata without exposing secrets.

Do not add Sentry or a new observability stack merely because the code is “ready” for one.

## Environments, deployment, and recovery

Maintain Local, Staging, and Pilot separation. Staging uses fake data and test integrations but resembles Pilot.

- CI/CD is GitHub Actions.
- Build API and worker images once and promote the exact tested digests.
- Never rebuild a branch specifically for Pilot.
- Run the migration Job once and stop deployment on failure.
- Roll out API and worker only after migration success.
- Use readiness-gated rolling updates; API `maxUnavailable` should normally be `0`.
- Record backend version, milestone, commit SHA, artifact digest, and compatible contract version/range.
- Preserve the previous known-good artifact references.
- A failed smoke test stops or rejects promotion.

Required Staging/Pilot smoke coverage includes API health through Cloudflare Tunnel and Traefik, PostgreSQL connectivity, Firebase-authenticated traffic, a public summary, a test worker job, and intended deployment digest/version verification. Frontend pipeline ownership covers the Pages application and direct SPA route, but backend changes must not break those checks.

PostgreSQL must support encrypted off-cluster backup/WAL archival, point-in-time recovery, approximately one-hour RPO, 30 days of normal recovery points, provisional monthly copies for 12 months, and monthly restore testing. Do not assert this is complete until OQ-01 through OQ-05 have named mechanisms/owners and a restore has produced recorded evidence.

## Open-question stop gates

Do not implement affected production decisions until these are resolved in an ADR or updated baseline:

| ID | Stop before |
| --- | --- |
| OQ-01 PostgreSQL deployment | Infrastructure scaffolding |
| OQ-02 backup tool/NAS integration | Pilot data entry; production backup manifests |
| OQ-03 recovery owner/evidence | Pilot launch |
| OQ-04 secrets-management tool | Staging deployment |
| OQ-05 tunnel placement/TLS path | External Staging access |
| OQ-06 domains/CORS/preview policy | Frontend/API integration |
| OQ-07 Go PDF library | PDF implementation |
| OQ-08 retention periods | Production cleanup jobs |
| OQ-09 invoice numbering/legal presentation | Pilot invoice issuance |
| OQ-10 deletion/anonymisation | Any user deletion promise or implementation |

When blocked, implement only decision-neutral interfaces, tests, or domain behaviour that will remain valid across the documented options. Clearly report the gate; do not encode a provisional production default.

## MVP scope discipline

Backend product scope currently supports:

- Firebase/Google tutor sign-in and first-time setup;
- tutor profile and workspace membership;
- students and profiles;
- recurring schedules plus create/edit/cancel lesson occurrences;
- home attention items and upcoming/week/calendar query data;
- payments list, recording, status, reversal/correction, and overdue derivation as defined by the contract;
- invoice drafts, atomic finalisation, immutable documents, and PDF output; and
- a public no-login summary showing approved tutee, schedule, and amount-owed data.

The MVP list is a scope summary, not permission to invent ambiguous financial semantics. Before adding acceptance signatures, proof-of-payment uploads, automated bank verification, payment processing, subscriptions, multi-tutor collaboration, destructive deletion, or automatic reminder sending, require explicit product, legal, security, data-retention, and API decisions as applicable.

## Testing requirements

Write tests at the lowest useful layer and keep critical invariants covered by more than mocks.

Every feature should consider:

- domain tests for state transitions, dates/time zones, money, and totals;
- handler/contract tests for validation, Problem Details, status codes, and authentication requirements;
- PostgreSQL integration tests for SQL, constraints, transactions, cursors, and locking;
- cross-workspace negative tests for every workspace-owned resource;
- idempotency tests for first request, identical retry, conflicting reuse, concurrent duplicates, and recovery after interruption;
- concurrency tests for invoice numbering, finalisation, payment recording/reversal, and job claiming;
- outbox tests proving business state and event commit or roll back together;
- worker tests for retry, jitter bounds, permanent failure, dead-letter, manual retry, and graceful claim release;
- public-summary tests for token hashing, rotation/revocation, generic failures, data minimisation, and rate limits;
- webhook tests using raw-body signature verification, replay/deduplication, stale timestamps, and fast acknowledgement;
- migration tests from the currently released schema and compatibility tests with the previous application version;
- PDF reproducibility and private-field exclusion tests; and
- smoke tests for health, authentication, public access, database, and a worker job.

Use realistic PostgreSQL tests rather than SQLite substitutes. Freeze clocks and randomness through injected interfaces where determinism matters. Do not weaken tests to accommodate a flawed implementation.

## Definition of done

A backend change is complete only when all applicable items are true:

- Behaviour and scope match the approved baseline and MVP.
- No open question was silently decided.
- OpenAPI changed first where the public contract changed, remains compatible, and documents errors/authentication.
- Generated code was regenerated with pinned tooling and has no drift.
- Workspace authorisation is explicit in every affected path.
- Money, time, identifiers, immutable history, and state-transition invariants are preserved.
- Consequential writes are transactionally safe, idempotent, and constraint-backed.
- External side effects occur after commit through outbox/jobs.
- Schema migrations are immutable, expand-and-contract compatible, and safe for rollback of application artifacts.
- Logs, errors, events, and audit metadata contain no secrets or unnecessary personal/financial data.
- Relevant unit, PostgreSQL integration, concurrency, negative-authorisation, contract, worker, and smoke tests pass.
- Operational behaviour, config, migration order, and rollout/rollback notes are documented.
- The handoff lists exactly what was verified and any remaining decision gates.

## Forbidden shortcuts

Never:

- trust entity UUIDs as authorisation;
- make `workspace_id` optional in workspace repositories;
- use floats for money;
- use device-local time as billing truth;
- mutate or regenerate a finalised invoice from current settings;
- delete recorded payments or issued invoice history;
- expose or log a public token, Firebase token, secret, or raw webhook body;
- perform an external call inside a business transaction;
- assume exactly-once job or webhook delivery;
- hand-edit OpenAPI-, `sqlc`-, or client-generated files;
- edit an already-applied migration;
- run migrations from API/worker startup;
- use automatic down-migrations on Pilot data;
- expose PostgreSQL publicly;
- use wildcard credentialed CORS;
- make Pilot credentials available to preview builds by default;
- rebuild artifacts between Staging and Pilot;
- treat a CronJob or email provider as the source of durable business state;
- implement a superseded decision; or
- resolve an Open item through an undocumented code choice.
