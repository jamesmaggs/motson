# 0001. Build Motson in Go

- Status: Accepted
- Date: 2026-06-12

## Context and drivers

Motson (specified in `docs/allium/motson.allium`) is a small, long-running
service: an hourly sync mirroring fixture data from an external provider,
plus two public read-only surfaces (an iCalendar feed and a fixtures web
page). It deploys to Railway. The runtime choice constrains frameworks,
libraries, testing tools, and how the hourly sync runs.

Drivers:

- Long-running process with a ticker-based hourly sync
- Small memory footprint and fast cold starts on Railway
- Simple single-service deployment
- Must be supported by Railway's build system

## Considered options

- **TypeScript on Node** — first-class Railway support, mature ICS and
  Postgres libraries, vitest for spec-generated tests
- **Go** — single static binary, tiny footprint, excellent fit for a
  long-running service with a ticker; less library convenience for
  templating/ICS
- **Python** — FastAPI/Flask plus icalendar; quick to write, weaker typing
- **Kotlin/JVM** — strong typing and ecosystem; heavy footprint and slow
  cold starts for a tiny service

## Decision

Build Motson in Go. Railway support is confirmed: Railpack/Nixpacks
auto-detects `go.mod` and builds a static binary (a Dockerfile remains an
option if more build control is needed).

## Consequences

- The service ships as one static binary; hourly sync can run in-process
  on a ticker rather than requiring an external scheduler
- Low memory/CPU usage on Railway
- ICS generation and HTML templating will use Go's smaller ecosystem
  (stdlib `html/template`; ICS via a small library or hand-rolled
  serialiser — to be decided separately)
- Test generation from the Allium spec targets Go's `testing` package
