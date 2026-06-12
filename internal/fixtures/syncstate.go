package fixtures

import "time"

// SyncState is the singleton pacing the hourly refresh
// (the spec's SyncSchedule entity).
type SyncState struct {
	NextRunAt    time.Time
	LastSyncedAt *time.Time // nil until the first successful sync
}

// IsStale reports whether the last successful sync is missing or at
// least StalenessThreshold old — the health boundary's signal.
func (s SyncState) IsStale(now time.Time) bool {
	return s.LastSyncedAt == nil || !s.LastSyncedAt.Add(StalenessThreshold).After(now)
}
