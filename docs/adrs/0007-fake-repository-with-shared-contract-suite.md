# 0007. Fake repository with a shared contract suite

- Status: Accepted
- Date: 2026-06-12

## Context and drivers

Test obligations will be derived from the Allium spec
(`/allium:propagate`), and the system's core behaviour — the pure-mirror
snapshot apply (upsert, removal, scores only when finished) — sits
between the sync logic and Postgres. Tests need to exercise that
behaviour fast and deterministically, without making every test run
depend on Docker or a network database.

Drivers:

- Spec-derived behavioural tests should run anywhere `go test` runs
- The pgx SQL must still be verified for real (ON CONFLICT, DELETE,
  null-score semantics)
- The fake and the real implementation must not drift apart

## Considered options

- **testcontainers Postgres** — full fidelity everywhere, but every
  test run needs Docker
- **Fake repository + SQL smoke tests** — behavioural tests against an
  in-memory fake; the SQL verified separately
- **Neon branch per CI run** — production-identical, but couples tests
  to Neon availability and credentials

## Decision

Define a repository interface between the sync/serving logic and
storage. Behavioural tests run against an in-memory fake. A single
shared contract suite exercises the repository interface and runs
against **both** implementations: always against the fake, and against
real Postgres whenever `TEST_DATABASE_URL` is set (locally and in CI).

## Consequences

- `go test ./...` is fast and dependency-free by default; CI gains a
  Postgres-backed job that sets `TEST_DATABASE_URL`
- The contract suite is the drift guard: any behaviour the fake
  exhibits must hold for pgx, or the suite fails
- The repository interface becomes a deliberate architectural seam
  (ports-and-adapters in miniature), which also keeps the provider
  adapter swappable per ADR 0003
- Bugs in Postgres-specific behaviour only surface in runs where
  `TEST_DATABASE_URL` is set — CI must always set it
