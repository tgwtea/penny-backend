# Database migrations

Place ordered, SQL-first migrations here as paired files:

```text
000001_short_description.up.sql
000001_short_description.down.sql
```

There is intentionally no initial business schema. Applied migrations are
immutable. Production rollbacks use a previous compatible application artifact
or a reviewed forward corrective migration; the deployment workflow must never
run automatic down-migrations against Pilot data.

The dedicated `migrate` executable applies only `up` migrations and is intended
to run once in a Kubernetes Job before API and worker rollout.

