package fixtures

import (
	"context"
	"time"
)

// Store is the port between the sync/serving logic and storage.
// Implementations must satisfy the shared contract suite in
// internal/store/storetest (ADR 0007).
type Store interface {
	// ReplaceAll applies a full provider snapshot as a pure mirror:
	// matches present are created or updated verbatim, matches absent
	// are removed, and the sync time is recorded — atomically.
	ReplaceAll(ctx context.Context, matches []Match, syncedAt time.Time) error

	// Matches returns every stored match in kickoff order.
	Matches(ctx context.Context) ([]Match, error)

	// SyncState returns the sync schedule singleton.
	SyncState(ctx context.Context) (SyncState, error)

	// ScheduleNextSync persists the time the next sync is due.
	ScheduleNextSync(ctx context.Context, at time.Time) error
}
