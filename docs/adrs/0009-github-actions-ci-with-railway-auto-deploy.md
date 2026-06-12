# 0009. GitHub Actions CI with Railway auto-deploy

- Status: Superseded-by-0012 (deploy mechanism; CI content unchanged)
- Date: 2026-06-12

## Context and drivers

The repo is local-only and code needs a path to Railway. The contract
suite (ADR 0007) is only fully exercised when `TEST_DATABASE_URL`
points at a real Postgres, so the deploy path determines whether that
check actually gates releases.

Drivers:

- The Postgres-backed contract suite must run before anything deploys
- Minimal pipeline maintenance for a short-lived tournament app
- Railway integrates natively with GitHub

## Considered options

- **GitHub + Actions + Railway auto-deploy** — Actions runs vet and
  tests with a Postgres service container; Railway deploys `main` on
  green via its GitHub integration
- **GitHub + Railway auto-deploy, no CI** — every push deploys; tests
  only run locally
- **Railway CLI from local** — `railway up`; no remote, no history off
  the laptop, no gate

## Decision

Host on GitHub. A GitHub Actions workflow runs `go vet` and
`go test ./...` with a Postgres service container supplying
`TEST_DATABASE_URL`, so the shared contract suite runs against real
Postgres on every push. Railway's GitHub integration auto-deploys
`main` after checks pass.

## Consequences

- The drift guard from ADR 0007 gates every deploy
- One workflow file to maintain; Railway needs the wait-for-CI option
  enabled on the service
- Secrets split cleanly: CI needs none beyond its ephemeral Postgres;
  Railway holds `DATABASE_URL` and `FOOTBALL_DATA_TOKEN`
