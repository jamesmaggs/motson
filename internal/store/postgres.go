package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Jazzatola/motson/internal/fixtures"
)

// schema is applied idempotently at boot (ADR 0006). sync_state is a
// singleton row enforced by its always-true primary key.
const schema = `
CREATE TABLE IF NOT EXISTS matches (
    provider_match_id text PRIMARY KEY,
    home_team         text NOT NULL,
    away_team         text NOT NULL,
    kickoff_at        timestamptz NOT NULL,
    venue             text NOT NULL,
    stage             text NOT NULL,
    group_name        text NOT NULL DEFAULT '',
    status            text NOT NULL,
    home_score        int CHECK (home_score >= 0),
    away_score        int CHECK (away_score >= 0),
    home_penalties    int,
    away_penalties    int
);
CREATE TABLE IF NOT EXISTS sync_state (
    id             boolean PRIMARY KEY DEFAULT true CHECK (id),
    next_run_at    timestamptz NOT NULL DEFAULT 'epoch',
    last_synced_at timestamptz
);
INSERT INTO sync_state (id) VALUES (true) ON CONFLICT (id) DO NOTHING;
`

// Postgres is the production fixtures.Store backed by NeonDB (ADR 0002).
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres connects and provisions the schema.
func NewPostgres(ctx context.Context, databaseURL string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("provisioning schema: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() { p.pool.Close() }

// Reset empties the store; used by the contract suite between cases.
func (p *Postgres) Reset(ctx context.Context) error {
	_, err := p.pool.Exec(ctx,
		`DELETE FROM matches;
		 UPDATE sync_state SET next_run_at = 'epoch', last_synced_at = NULL;`)
	return err
}

func (p *Postgres) ReplaceAll(ctx context.Context, matches []fixtures.Match, syncedAt time.Time) error {
	for _, m := range matches {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Full snapshot replace: the pure mirror in SQL.
	if _, err := tx.Exec(ctx, `DELETE FROM matches`); err != nil {
		return err
	}
	for _, m := range matches {
		if _, err := tx.Exec(ctx,
			`INSERT INTO matches (provider_match_id, home_team, away_team, kickoff_at,
			                      venue, stage, group_name, status,
			                      home_score, away_score, home_penalties, away_penalties)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			m.ProviderMatchID, m.HomeTeam, m.AwayTeam, m.KickoffAt,
			m.Venue, m.Stage, m.GroupName, m.Status,
			m.HomeScore, m.AwayScore, m.HomePenalties, m.AwayPenalties); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE sync_state SET last_synced_at = $1`, syncedAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (p *Postgres) Matches(ctx context.Context) ([]fixtures.Match, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT provider_match_id, home_team, away_team, kickoff_at,
		        venue, stage, group_name, status,
		        home_score, away_score, home_penalties, away_penalties
		 FROM matches
		 ORDER BY kickoff_at, provider_match_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []fixtures.Match
	for rows.Next() {
		var m fixtures.Match
		if err := rows.Scan(&m.ProviderMatchID, &m.HomeTeam, &m.AwayTeam, &m.KickoffAt,
			&m.Venue, &m.Stage, &m.GroupName, &m.Status,
			&m.HomeScore, &m.AwayScore, &m.HomePenalties, &m.AwayPenalties); err != nil {
			return nil, err
		}
		m.KickoffAt = m.KickoffAt.UTC()
		matches = append(matches, m)
	}
	return matches, rows.Err()
}

func (p *Postgres) SyncState(ctx context.Context) (fixtures.SyncState, error) {
	var state fixtures.SyncState
	err := p.pool.QueryRow(ctx,
		`SELECT next_run_at, last_synced_at FROM sync_state`).
		Scan(&state.NextRunAt, &state.LastSyncedAt)
	if err == pgx.ErrNoRows {
		return fixtures.SyncState{}, nil
	}
	if err != nil {
		return fixtures.SyncState{}, err
	}
	state.NextRunAt = state.NextRunAt.UTC()
	if state.LastSyncedAt != nil {
		utc := state.LastSyncedAt.UTC()
		state.LastSyncedAt = &utc
	}
	return state, nil
}

func (p *Postgres) ScheduleNextSync(ctx context.Context, at time.Time) error {
	_, err := p.pool.Exec(ctx, `UPDATE sync_state SET next_run_at = $1`, at)
	return err
}
