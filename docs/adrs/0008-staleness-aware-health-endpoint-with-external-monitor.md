# 0008. Staleness-aware health endpoint with external monitor

- Status: Accepted
- Date: 2026-06-12

## Context and drivers

The Allium spec left one open question: should anyone be alerted when
syncs keep failing, or is the visible last-synced time on the web page
enough? A quietly failing sync (provider errors, bad token, network
issues) would otherwise serve increasingly stale fixtures all
tournament without anyone noticing.

Drivers:

- Failures should reach a human without anyone having to look
- Zero or near-zero cost and operational overhead
- Railway can use a health check endpoint natively

## Considered options

- **Staleness-aware `/healthz` + free external monitor** — unhealthy
  when `last_synced_at` exceeds a threshold; monitor emails on failure
- **Railway logs/restart alerts only** — catches crashes, misses
  quietly failing syncs
- **No alerting** — accept silent staleness; page shows last-synced time

## Decision

Expose `/healthz`, returning healthy only when the database is
reachable and `last_synced_at` is within a staleness threshold
(default 3 hours — three consecutive missed syncs). A free external
monitor (e.g. UptimeRobot or healthchecks.io) pings it and emails on
failure. The same endpoint serves as Railway's health check.

## Consequences

- Sync failures surface within hours via email, not via someone
  noticing a stale page
- The staleness threshold joins application config; the Allium spec's
  open question is resolved and the spec gains the health boundary
- One manual setup step outside the repo: registering the monitor
- During the threshold window (first ~3h of failures) staleness is
  visible only on the page — accepted
