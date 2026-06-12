# 0006. Idempotent schema at boot

- Status: Accepted
- Date: 2026-06-12

## Context and drivers

The store (ADR 0002) needs two tables — `matches` and `sync_schedule` —
which will rarely if ever change during the tournament. Something must
create them on a fresh Neon database and keep deploys hands-free.

Drivers:

- Self-provisioning deploys (fresh database → working service)
- Avoid tooling for a schema this small
- Lean dependency stance (ADR 0005)

## Considered options

- **Idempotent schema at boot** — `CREATE TABLE IF NOT EXISTS` on
  startup; zero tooling
- **goose embedded migrations** — versioned migration history; one more
  dependency and ceremony for two tables
- **Manual psql** — human step on every fresh environment

## Decision

The service applies idempotent `CREATE TABLE IF NOT EXISTS` statements
at boot, before the ticker and HTTP server start.

## Consequences

- Fresh environments provision themselves; no deploy runbook
- No migration history; a future schema change needs a guarded `ALTER`
  or a revisit of this decision (goose can be adopted later without
  unwinding anything)
- Boot fails fast and loudly if the database is unreachable or the DDL
  fails
