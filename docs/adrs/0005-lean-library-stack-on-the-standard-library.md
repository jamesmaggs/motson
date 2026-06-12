# 0005. Lean library stack on the standard library

- Status: Accepted
- Date: 2026-06-12

## Context and drivers

With Go decided (ADR 0001), Motson needs an HTTP layer, HTML rendering,
Postgres access and iCalendar serialisation. The surface area is tiny:
two public GET routes, two tables, ~6 queries, one ICS document. Every
dependency is a maintenance liability for a service that should run
untouched through a six-week tournament.

Drivers:

- Two routes, select/upsert/delete queries, one feed document
- The `AppleCalendarCompatible` guarantee rests on RFC 5545 subtleties
  (75-octet line folding, CRLF, text escaping, stable UIDs,
  STATUS:CANCELLED)
- Prefer the standard library where it is genuinely sufficient

## Considered options

- **HTTP**: stdlib `net/http` (Go 1.22+ ServeMux) vs chi vs Echo/Gin
- **Rendering**: stdlib `html/template` vs a template engine
- **Postgres**: pgx with hand-written SQL vs sqlc+pgx vs
  `database/sql`+lib/pq (maintenance mode) vs GORM
- **ICS**: `github.com/arran4/golang-ical` vs hand-rolled serialiser

## Decision

Standard library for HTTP routing and HTML templating. Two dependencies
where the stdlib is not sufficient:

- **pgx** as the Postgres driver, with hand-written SQL in one
  repository file — explicit control over the pure-mirror upsert
- **golang-ical** for ICS serialisation — a tested implementation of
  RFC 5545's sharp edges rather than owning them ourselves

## Consequences

- `go.mod` stays at roughly two direct dependencies (pgx, golang-ical)
- Handlers are trivially testable with `httptest`; no framework idiom
  to learn or upgrade
- Feed correctness depends on golang-ical's RFC compliance; an
  Apple-Calendar smoke test of real output is still warranted
- If route count or query complexity ever grows substantially, chi or
  sqlc can be introduced without unwinding anything
