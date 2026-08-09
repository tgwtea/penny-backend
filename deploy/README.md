# Deployment gate

Kubernetes is Penny's confirmed runtime, but deployable manifests are not yet
permitted by the architecture baseline. Section 17 requires OQ-01, OQ-02,
OQ-04, OQ-05, real domains/CORS, a rehearsed restore runbook, and a confirmed
Pages promotion mechanism before infrastructure scaffolding.

Accordingly, this directory contains no placeholder YAML with guessed storage
classes, operators, secret references, tunnel placement, TLS, host names,
namespaces, service accounts, resource sizes, or backup schedules.

Once the gates are resolved, the first deployment change should add, in reviewable
order:

1. environment overlays for Local/Staging/Pilot without secret values in Git;
2. migration Job consuming the exact released migration-image digest;
3. API Deployment/Service with `/health/live` and `/health/ready` probes,
   readiness-gated rolling updates, and `maxUnavailable: 0`;
4. worker Deployment with graceful termination and the exact compatible digest;
5. CronJobs that only initiate scans while PostgreSQL owns durable job state;
6. cluster-standard PostgreSQL and off-cluster backup/WAL resources selected by
   OQ-01/OQ-02;
7. cluster-standard secret references selected by OQ-04;
8. Cloudflare Tunnel, Traefik, TLS, and private-route resources selected by
   OQ-05/OQ-06; and
9. promotion/smoke/rollback automation that consumes existing digests and never
   rebuilds between Staging and Pilot.

The existing image-build workflow is intentionally separate from deployment so
future promotion can consume recorded immutable digests.

