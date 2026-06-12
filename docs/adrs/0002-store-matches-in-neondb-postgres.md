# 0002. Store matches in NeonDB Postgres

- Status: Accepted
- Date: 2026-06-12

## Context and drivers

The Allium spec defines a small persistent store: the `Match` collection
(~104 rows for World Cup 2026) and a `SyncSchedule` singleton. The
`StalenessTolerated` guarantee requires Motson to keep serving the last
successfully synced data when the provider is unavailable — including
across service restarts. The project brief (CLAUDE.md) nominates NeonDB
if a relational database is needed.

Drivers:

- Survive restarts that coincide with provider outages
- Tiny data volume; cost and operational burden should stay near zero
- Railway deployment (redeploys replace instances)

## Considered options

- **NeonDB Postgres** — durable across restarts and outages; free tier
  easily covers the data volume; matches the project brief
- **In-memory + fetch on boot** — no database at all; simplest, but a
  restart during a provider outage leaves nothing to serve
- **SQLite on a Railway volume** — durable and dependency-free, but
  volumes pin the service to one instance and complicate redeploys

## Decision

Use NeonDB Postgres as the match store.

## Consequences

- `StalenessTolerated` holds unconditionally, including restart during
  provider outage
- One external dependency (Neon) and a `DATABASE_URL` secret on Railway
- Go service needs a Postgres driver (e.g. pgx) and a minimal schema
  (matches table + sync state); migration tooling to be decided
- Slight overkill for ~104 rows — accepted for zero operational surprises
