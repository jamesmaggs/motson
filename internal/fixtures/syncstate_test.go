package fixtures_test

import (
	"testing"
	"time"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

// Obligation: derived.SyncSchedule.is_stale — stale when never synced,
// or when the last successful sync is at least staleness_threshold old.
func TestSyncStateIsStale(t *testing.T) {
	now := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	at := func(d time.Duration) *time.Time { ts := now.Add(d); return &ts }

	cases := []struct {
		name       string
		lastSynced *time.Time
		want       bool
	}{
		{"never synced", nil, true},
		{"synced just now", at(0), false},
		{"synced an hour ago", at(-time.Hour), false},
		{"synced just under threshold", at(-3*time.Hour + time.Second), false},
		{"synced exactly threshold ago", at(-3 * time.Hour), true},
		{"synced well past threshold", at(-24 * time.Hour), true},
	}
	for _, c := range cases {
		s := fixtures.SyncState{LastSyncedAt: c.lastSynced}
		if got := s.IsStale(now); got != c.want {
			t.Errorf("%s: IsStale = %v, want %v", c.name, got, c.want)
		}
	}
}
