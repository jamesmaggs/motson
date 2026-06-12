# Motson

Live World Cup 2026 fixtures and scores, served as an Apple
Calendar-compatible feed and a web page. Fixture data is mirrored
every 10 minutes from [football-data.org](https://www.football-data.org)
into Postgres and served even when the provider is down.

- **Web page**: `/` — fixtures and scores, kickoffs in your local time
- **Calendar feed**: `/calendar.ics` — subscribe in Apple Calendar;
  event titles update with final scores
- **Health**: `/healthz` — unhealthy when the last successful sync is
  older than 3 hours

## How it's specified

Behaviour is specified in [Allium](docs/allium/motson.allium) and
architecture decisions are recorded as [ADRs](docs/adrs/README.md).
The implementation was test-driven from the spec's obligations
(`allium plan`); spec-to-code alignment is audited with the Allium
weed workflow.

## Development

Toolchain is pinned with [mise](https://mise.jdx.dev):

```sh
mise install
go test ./...
```

The store contract suite (`internal/store/storetest`) always runs
against the in-memory fake. Set `TEST_DATABASE_URL` to also run it
against real Postgres, which CI does on every push:

```sh
TEST_DATABASE_URL=postgres://localhost:5432/motson_test go test ./internal/store/
```

## Running

```sh
DATABASE_URL=postgres://... \
FOOTBALL_DATA_TOKEN=... \
go run ./cmd/motson
```

| Variable | Required | Default |
|----------|----------|---------|
| `DATABASE_URL` | yes | — |
| `FOOTBALL_DATA_TOKEN` | yes | — |
| `FOOTBALL_DATA_URL` | no | `https://api.football-data.org` |
| `COMPETITION` | no | `WC` |
| `PORT` | no | `8080` |
| `FEED_HOST` | no | `motson.jamesmaggs.com` |

The schema is provisioned idempotently at boot; a fresh database needs
no setup. The first sync runs immediately, then every 10 minutes.

## Deployment

Pushes to `main` run CI (vet, gofmt, tests including the Postgres
contract suite) and Railway auto-deploys on green, building the
Dockerfile. Production database is NeonDB Postgres 17. The service is
monitored externally via `/healthz`.
