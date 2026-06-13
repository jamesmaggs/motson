# Motson

Live fixtures, scores and group standings for the 2026 FIFA World Cup. A
single Go service mirrors fixture data from football-data.org into Postgres
and serves it as a server-rendered web app and an iCalendar feed. Live at
[motson.jamesmaggs.com](https://motson.jamesmaggs.com); user-facing overview
in [README.md](README.md).

## Orientation

Two sources of truth — read the relevant one before changing behaviour:

- **What the app must do** → the Allium behavioural spec,
  [`docs/allium/motson.allium`](docs/allium/motson.allium). It defines every
  surface (FixturesPage, GroupDetailPage, TeamDetailPage, the feed, the
  health check), its rules and its guarantees. The implementation was
  test-driven from this spec's obligations.
- **Why the architecture is the way it is** → the ADRs in
  [`docs/adrs/`](docs/adrs/README.md) (12 records: Go, NeonDB Postgres,
  football-data.org, in-process sync ticker, lean stdlib stack, idempotent
  schema, fake repo + contract suite, staleness-aware health, CI/deploy,
  the server-rendered page, the custom domain).
- **What the UI is meant to feel like** → the product register and
  personality in [`PRODUCT.md`](PRODUCT.md).

## Stack

- Go 1.25, pinned via [mise](https://mise.jdx.dev) (`mise.toml`). Standard
  library throughout: `net/http`, `html/template`, `embed`.
- Lean dependencies (ADR 0005): `pgx` (Postgres), `golang-ical` (RFC 5545 feed).
- NeonDB Postgres 17 in production; an in-memory fake elsewhere.
- Deployed to Railway from a `Dockerfile`; Cloudflare fronts the domain.

## Layout

- `cmd/motson` — entrypoint: load config, provision schema, start the sync
  ticker, serve HTTP.
- `internal/fixtures` — the domain: `Match`, `Status`, `Stage`,
  `Standings()`, and the `Store` interface everything depends on.
- `internal/store` — Postgres and in-memory implementations of `Store`,
  with a shared contract suite in `internal/store/storetest` (ADR 0007).
- `internal/footballdata` — the provider client.
- `internal/syncer` — the 10-minute sync ticker (ADR 0004).
- `internal/feed` — the iCalendar feed.
- `internal/web` — the HTTP boundary: handlers + `html/template` views in
  `templates/` (chrome and CSS live in `shared.html.tmpl`), and vendored
  fonts/CSS/mascot in `static/`. Routes: `/` (index), `/groups/{letter}`,
  `/teams/{slug}`, `/calendar.ics`, `/healthz`, and a styled catch-all 404.
- `docs/` — the Allium spec and the ADRs.

## Working agreement

- TDD, red-green-refactor, driven by the Allium spec's obligations
  (`allium plan`). Tests assert behaviour, not implementation — never weaken
  a test to make it pass.
- Keep the spec aligned with the code: when behaviour changes, update
  [`docs/allium/motson.allium`](docs/allium/motson.allium) and run
  `allium check`. Spec-to-code drift is audited with the Allium weed workflow.
- Keep `go test ./...`, `gofmt -l` (empty) and `go vet ./...` green. Small,
  atomic commits.
- The web UI follows the "impeccable" design process; design changes are
  verified in a browser (Chrome DevTools / Lighthouse), not by eye.

## Runtime & configuration

Single service. The schema is provisioned idempotently at boot (ADR 0006),
so a fresh database needs no setup. The first sync runs immediately, then
every 10 minutes. `/healthz` is unhealthy when the last successful sync is
older than 3 hours (ADR 0008) and is monitored externally.

| Variable | Required | Default |
|---|---|---|
| `DATABASE_URL` | yes | — |
| `FOOTBALL_DATA_TOKEN` | yes | — |
| `FOOTBALL_DATA_URL` | no | `https://api.football-data.org` |
| `COMPETITION` | no | `WC` |
| `PORT` | no | `8080` |
| `FEED_HOST` | no | `motson.jamesmaggs.com` |

Run and test locally:

```sh
mise install
go test ./...
DATABASE_URL=postgres://... FOOTBALL_DATA_TOKEN=... go run ./cmd/motson
```

The store contract suite runs against the in-memory fake by default; set
`TEST_DATABASE_URL` to also exercise real Postgres (CI does).

## CI & deploy

Pushes to `main` run GitHub Actions ([`.github/workflows/ci.yml`](.github/workflows/ci.yml)):
gofmt, `go vet`, and `go test ./...` including the Postgres contract suite
against a service container. On green, the workflow deploys to Railway via
the Railway CLI (ADRs 0009, 0012). Static assets are fingerprinted with
`?v=<build>` and served with long-lived immutable cache headers.

## Design system

A "floodlit scoreboard" dark theme: a Bebas Neue display title over Sora
body and score type, an amber accent (`#ffd23f`), green group accents, and
all colours expressed as CSS custom properties (`--bg`, `--surface`,
`--ink`, `--accent`, …) in `shared.html.tmpl`. The three pages share one
card-based, three-column layout (menu · matches · standings) that stacks on
mobile with the standings first; the current group/team is highlighted in
amber as a wayfinding cue. Match cards come from a shared `matchCard`
partial. Fonts, CSS and the mascot are vendored under
`internal/web/static` (no external runtime dependencies); a deferred Umami
snippet provides analytics. Kick-off times render a UTC fallback that client
JS localises (ADR 0010), so the page still works without JavaScript.
