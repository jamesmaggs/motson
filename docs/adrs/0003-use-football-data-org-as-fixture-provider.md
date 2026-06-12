# 0003. Use football-data.org as the fixture provider

- Status: Accepted
- Date: 2026-06-12

## Context and drivers

The Allium spec keeps the data provider abstract behind the
`FixtureSource` contract (full tournament snapshot per fetch). The
implementation needs a concrete provider. Hourly sync means ~24
requests/day, well inside every candidate's free tier, so the decision
rests on data quality and how cleanly the provider's shape maps onto the
spec.

Drivers:

- Status model must map onto `scheduled | in_play | finished | postponed
  | cancelled`
- Full-time and penalty-shootout scores as distinct fields
- Knockout placeholders ("Winner Group A") delivered as team names, per
  the pure-mirror decision
- Free or near-free at hourly polling rates

## Considered options

- **football-data.org** — free tier includes the World Cup; statuses map
  almost 1:1; full-time and penalty scores are distinct fields; simple
  token auth
- **API-Football (api-sports.io)** — richer data (lineups, events) the
  spec doesn't need; 100 req/day free tier still fits; heavier response
  shapes to map
- **TheSportsDB** — free and simple but community-maintained; weaker
  reliability for live tournament corrections

## Decision

Implement `FixtureSource` against football-data.org's competition
matches endpoint, authenticated with its API token.

## Consequences

- One secret (`FOOTBALL_DATA_TOKEN`) alongside `DATABASE_URL` on Railway
- A thin adapter maps the provider's match JSON onto `ProviderFixture`;
  provider statuses (including TIMED/PAUSED variants) need an explicit
  mapping table onto the spec's five statuses
- The provider remains swappable: only the adapter implements the
  `FixtureSource` contract
- Free-tier rate limits (10 req/min) are irrelevant at hourly cadence
