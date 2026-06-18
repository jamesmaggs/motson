// Package syncer realises the spec's SyncDue temporal rule: when the
// persisted next_run_at passes, advance the schedule and apply the
// provider's snapshot.
package syncer

import (
	"context"
	"fmt"
	"time"

	"github.com/jamesmaggs/motson/internal/fixtures"
	"github.com/jamesmaggs/motson/internal/venues"
)

// Source is the FixtureSource contract: one call returns the complete
// tournament snapshot.
type Source interface {
	FetchFixtures(ctx context.Context) ([]fixtures.Match, error)
}

type Syncer struct {
	source Source
	store  fixtures.Store
}

func New(source Source, store fixtures.Store) *Syncer {
	return &Syncer{source: source, store: store}
}

// RunDue fires the sync if next_run_at has passed, advancing the
// schedule first so a failed fetch waits for the next interval rather
// than retrying hot. A failure leaves stored matches untouched
// (StalenessTolerated).
func (s *Syncer) RunDue(ctx context.Context, now time.Time) error {
	state, err := s.store.SyncState(ctx)
	if err != nil {
		return fmt.Errorf("reading sync state: %w", err)
	}
	if state.NextRunAt.After(now) {
		return nil
	}
	if err := s.store.ScheduleNextSync(ctx, now.Add(fixtures.SyncInterval)); err != nil {
		return fmt.Errorf("scheduling next sync: %w", err)
	}
	matches, err := s.source.FetchFixtures(ctx)
	if err != nil {
		return fmt.Errorf("fetching fixtures: %w", err)
	}
	// The provider's free tier omits venues; fill them from static data.
	matches = venues.Enrich(matches)
	if err := s.store.ReplaceAll(ctx, matches, now); err != nil {
		return fmt.Errorf("applying snapshot: %w", err)
	}
	return nil
}
