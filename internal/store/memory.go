// Package store provides fixtures.Store implementations: an in-memory
// fake for tests and a Postgres store for production.
package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Jazzatola/motson/internal/fixtures"
)

// Memory is the in-memory fixtures.Store used by behavioural tests.
type Memory struct {
	mu      sync.RWMutex
	matches map[string]fixtures.Match
	state   fixtures.SyncState
}

func NewMemory() *Memory {
	return &Memory{matches: make(map[string]fixtures.Match)}
}

func (m *Memory) ReplaceAll(_ context.Context, matches []fixtures.Match, syncedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.matches = make(map[string]fixtures.Match, len(matches))
	for _, match := range matches {
		m.matches[match.ProviderMatchID] = match
	}
	ts := syncedAt
	m.state.LastSyncedAt = &ts
	return nil
}

func (m *Memory) Matches(_ context.Context) ([]fixtures.Match, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]fixtures.Match, 0, len(m.matches))
	for _, match := range m.matches {
		out = append(out, match)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].KickoffAt.Equal(out[j].KickoffAt) {
			return out[i].KickoffAt.Before(out[j].KickoffAt)
		}
		return out[i].ProviderMatchID < out[j].ProviderMatchID
	})
	return out, nil
}

func (m *Memory) SyncState(_ context.Context) (fixtures.SyncState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state, nil
}

func (m *Memory) ScheduleNextSync(_ context.Context, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.NextRunAt = at
	return nil
}
