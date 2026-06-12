# 0012. Deploy from GitHub Actions with the Railway CLI

- Status: Accepted
- Date: 2026-06-12

## Context and drivers

ADR 0009 assumed Railway's GitHub integration would auto-deploy `main`
after CI passed. In practice the integration requires installing
Railway's GitHub App, which the owner prefers not to do; pushes were
producing no deployments. The owner has deployed to Railway from CI
before and asked for a GitHub Actions step instead.

Drivers:

- Deploys must still be gated by the full test suite (including the
  Postgres contract suite)
- No additional third-party app installations on the GitHub account
- Keep the pipeline in one visible place (the workflow file)

## Considered options

- **Railway GitHub App auto-deploy** (ADR 0009) — dashboard-managed,
  but requires installing the GitHub App
- **Deploy job in GitHub Actions running the Railway CLI** — `railway
  up` with a project-scoped `RAILWAY_TOKEN` secret, running only on
  `main` pushes after the test job succeeds
- **Local `railway up`** — bypasses CI gating; ruled out

## Decision

Replace the GitHub App integration with a `deploy` job in the CI
workflow: on `main` pushes, after the test job passes, run
`railway up --service motson --ci` authenticated by a project token
stored as the `RAILWAY_TOKEN` repository secret.

This supersedes the deploy mechanism of ADR 0009; its CI content
(vet, gofmt, tests against real Postgres) is unchanged.

## Consequences

- The deploy is explicitly sequenced after green tests in one workflow
  file; no Railway-side "wait for CI" configuration needed
- One repository secret (`RAILWAY_TOKEN`, project-scoped) to manage
- Railway builds from the pushed source via the CLI rather than its
  GitHub fetcher; the service's attached repo source is vestigial
- Rotating the project token invalidates deploys until the secret is
  updated
