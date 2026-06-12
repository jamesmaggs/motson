# 0004. Single service with in-process sync ticker

- Status: Accepted
- Date: 2026-06-12

## Context and drivers

Motson has two responsibilities: serving the public surfaces (iCalendar
feed, fixtures page) and running the hourly provider sync. The Allium
spec models sync as a temporal rule (`SyncDue`) firing when the
persisted `SyncSchedule.next_run_at` passes. Railway can host one or
many services and offers cron triggers.

Drivers:

- Minimise moving parts for a single-tournament, single-purpose app
- Keep the sync cadence in the spec's `config` block, not platform config
- The spec's temporal rule should map directly onto the implementation

## Considered options

- **One service, in-process ticker** — a single Go binary serves HTTP and
  runs a ticker that checks the persisted `next_run_at`; syncs on boot if
  due
- **Web service + Railway cron** — cron hits a sync endpoint or job
  hourly; schedule lives in Railway config, two moving parts
- **Two services (web + worker)** — clean separation, sharing the
  database; overkill at this scale

## Decision

Ship one Go binary. It serves HTTP and runs an in-process ticker that
fires the sync whenever `SyncSchedule.next_run_at <= now`, including
immediately at boot when due.

## Consequences

- One Railway service, one deploy pipeline, one log stream
- Sync cadence stays in application config, mirroring the spec
- A crashed sync loop takes the web server down with it (and vice
  versa) — acceptable: Railway restarts the service and the store
  preserves last-synced data
- If Railway ever runs multiple replicas, the persisted `next_run_at`
  check makes concurrent syncs harmless (idempotent upsert) but a
  single replica is the assumption
