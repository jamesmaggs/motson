// Package storetest is the shared contract suite for fixtures.Store
// implementations (ADR 0007). Every implementation — the in-memory
// fake and the Postgres store — must pass exactly this suite, so the
// two cannot drift apart.
package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/Jazzatola/motson/internal/fixtures"
)

func intp(i int) *int { return &i }

func sampleMatch(id string) fixtures.Match {
	return fixtures.Match{
		ProviderMatchID: id,
		HomeTeam:        "Canada",
		AwayTeam:        "Mexico",
		KickoffAt:       time.Date(2026, 6, 13, 18, 0, 0, 0, time.UTC),
		Venue:           "Estadio Azteca, Mexico City",
		Stage:           fixtures.StageGroup,
		GroupName:       "Group A",
		Status:          fixtures.StatusScheduled,
	}
}

// Run exercises the Store contract against the implementation produced
// by newStore, which must return an empty store each call.
func Run(t *testing.T, newStore func(t *testing.T) fixtures.Store) {
	ctx := context.Background()
	syncedAt := time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)

	t.Run("empty store has no matches and a never-synced state", func(t *testing.T) {
		s := newStore(t)
		matches, err := s.Matches(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 0 {
			t.Errorf("got %d matches, want 0", len(matches))
		}
		state, err := s.SyncState(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if state.LastSyncedAt != nil {
			t.Errorf("LastSyncedAt = %v, want nil", state.LastSyncedAt)
		}
	})

	// Obligation: rule-entity-creation.FixtureDataApplied.1 — new
	// fixtures create matches carrying every specified field.
	t.Run("snapshot creates matches with all fields", func(t *testing.T) {
		s := newStore(t)
		want := sampleMatch("wc-1")
		if err := s.ReplaceAll(ctx, []fixtures.Match{want}, syncedAt); err != nil {
			t.Fatal(err)
		}
		matches, err := s.Matches(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 1 {
			t.Fatalf("got %d matches, want 1", len(matches))
		}
		assertMatchEqual(t, matches[0], want)
	})

	// Obligation: rule-success.FixtureDataApplied — existing matches are
	// updated field-for-field; scores appear once finished
	// (when-presence.Match.home_score / away_score / *_penalties).
	t.Run("snapshot updates an existing match in place", func(t *testing.T) {
		s := newStore(t)
		before := sampleMatch("wc-1")
		if err := s.ReplaceAll(ctx, []fixtures.Match{before}, syncedAt); err != nil {
			t.Fatal(err)
		}

		after := before
		after.Status = fixtures.StatusFinished
		after.HomeScore, after.AwayScore = intp(2), intp(1)
		after.KickoffAt = before.KickoffAt.Add(30 * time.Minute) // provider correction
		if err := s.ReplaceAll(ctx, []fixtures.Match{after}, syncedAt.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}

		matches, err := s.Matches(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 1 {
			t.Fatalf("got %d matches, want 1 (update must not duplicate)", len(matches))
		}
		assertMatchEqual(t, matches[0], after)
	})

	// Pure mirror: a provider correction that un-finishes a match takes
	// its score with it (when-presence — scores absent unless finished).
	t.Run("snapshot applies regressions verbatim", func(t *testing.T) {
		s := newStore(t)
		finished := sampleMatch("wc-1")
		finished.Status = fixtures.StatusFinished
		finished.HomeScore, finished.AwayScore = intp(2), intp(1)
		if err := s.ReplaceAll(ctx, []fixtures.Match{finished}, syncedAt); err != nil {
			t.Fatal(err)
		}

		reverted := sampleMatch("wc-1") // back to scheduled, no scores
		if err := s.ReplaceAll(ctx, []fixtures.Match{reverted}, syncedAt.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}

		matches, err := s.Matches(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assertMatchEqual(t, matches[0], reverted)
	})

	// Shootout results roundtrip (entity-optional.Match.*_penalties).
	t.Run("snapshot preserves penalty shootout results", func(t *testing.T) {
		s := newStore(t)
		m := sampleMatch("wc-final")
		m.Stage = fixtures.StageFinal
		m.GroupName = ""
		m.Status = fixtures.StatusFinished
		m.HomeScore, m.AwayScore = intp(3), intp(3)
		m.HomePenalties, m.AwayPenalties = intp(4), intp(2)
		if err := s.ReplaceAll(ctx, []fixtures.Match{m}, syncedAt); err != nil {
			t.Fatal(err)
		}
		matches, err := s.Matches(ctx)
		if err != nil {
			t.Fatal(err)
		}
		assertMatchEqual(t, matches[0], m)
	})

	// Obligation: invariant.NonNegativeScores — a snapshot carrying a
	// negative score is rejected whole, leaving the store unchanged
	// (the invariant holds after every state change; StalenessTolerated
	// covers serving the previous data).
	t.Run("snapshot with a negative score is rejected whole", func(t *testing.T) {
		s := newStore(t)
		good := sampleMatch("wc-1")
		if err := s.ReplaceAll(ctx, []fixtures.Match{good}, syncedAt); err != nil {
			t.Fatal(err)
		}

		bad := sampleMatch("wc-2")
		bad.Status = fixtures.StatusFinished
		bad.HomeScore, bad.AwayScore = intp(-1), intp(2)
		if err := s.ReplaceAll(ctx, []fixtures.Match{good, bad}, syncedAt.Add(time.Hour)); err == nil {
			t.Fatal("want error for negative score")
		}

		matches, err := s.Matches(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 1 || matches[0].ProviderMatchID != "wc-1" {
			t.Errorf("store changed by rejected snapshot: %v", ids(matches))
		}
		state, err := s.SyncState(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if state.LastSyncedAt == nil || !state.LastSyncedAt.Equal(syncedAt) {
			t.Errorf("LastSyncedAt = %v, want unchanged %v", state.LastSyncedAt, syncedAt)
		}
	})

	// Obligation: rule-success.VanishedMatchesRemoved — matches absent
	// from a snapshot are removed.
	t.Run("snapshot removes vanished matches", func(t *testing.T) {
		s := newStore(t)
		a, b := sampleMatch("wc-1"), sampleMatch("wc-2")
		b.KickoffAt = a.KickoffAt.Add(3 * time.Hour)
		if err := s.ReplaceAll(ctx, []fixtures.Match{a, b}, syncedAt); err != nil {
			t.Fatal(err)
		}
		if err := s.ReplaceAll(ctx, []fixtures.Match{b}, syncedAt.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
		matches, err := s.Matches(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 1 || matches[0].ProviderMatchID != "wc-2" {
			t.Errorf("got %v, want only wc-2", ids(matches))
		}
	})

	t.Run("matches are returned in kickoff order", func(t *testing.T) {
		s := newStore(t)
		first, second, third := sampleMatch("wc-1"), sampleMatch("wc-2"), sampleMatch("wc-3")
		second.KickoffAt = first.KickoffAt.Add(3 * time.Hour)
		third.KickoffAt = first.KickoffAt.Add(6 * time.Hour)
		if err := s.ReplaceAll(ctx, []fixtures.Match{third, first, second}, syncedAt); err != nil {
			t.Fatal(err)
		}
		matches, err := s.Matches(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := ids(matches), []string{"wc-1", "wc-2", "wc-3"}; !equal(got, want) {
			t.Errorf("order = %v, want %v", got, want)
		}
	})

	// FixtureDataApplied also ensures last_synced_at = now.
	t.Run("applying a snapshot records the sync time", func(t *testing.T) {
		s := newStore(t)
		if err := s.ReplaceAll(ctx, []fixtures.Match{sampleMatch("wc-1")}, syncedAt); err != nil {
			t.Fatal(err)
		}
		state, err := s.SyncState(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if state.LastSyncedAt == nil || !state.LastSyncedAt.Equal(syncedAt) {
			t.Errorf("LastSyncedAt = %v, want %v", state.LastSyncedAt, syncedAt)
		}
	})

	// SyncDue persists the advanced next_run_at.
	t.Run("scheduling the next sync persists next run time", func(t *testing.T) {
		s := newStore(t)
		next := syncedAt.Add(time.Hour)
		if err := s.ScheduleNextSync(ctx, next); err != nil {
			t.Fatal(err)
		}
		state, err := s.SyncState(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if !state.NextRunAt.Equal(next) {
			t.Errorf("NextRunAt = %v, want %v", state.NextRunAt, next)
		}
	})
}

func assertMatchEqual(t *testing.T, got, want fixtures.Match) {
	t.Helper()
	if got.ProviderMatchID != want.ProviderMatchID ||
		got.HomeTeam != want.HomeTeam ||
		got.AwayTeam != want.AwayTeam ||
		!got.KickoffAt.Equal(want.KickoffAt) ||
		got.Venue != want.Venue ||
		got.Stage != want.Stage ||
		got.GroupName != want.GroupName ||
		got.Status != want.Status ||
		!intpEqual(got.HomeScore, want.HomeScore) ||
		!intpEqual(got.AwayScore, want.AwayScore) ||
		!intpEqual(got.HomePenalties, want.HomePenalties) ||
		!intpEqual(got.AwayPenalties, want.AwayPenalties) {
		t.Errorf("match = %+v, want %+v", got, want)
	}
}

func intpEqual(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func ids(matches []fixtures.Match) []string {
	out := make([]string, len(matches))
	for i, m := range matches {
		out[i] = m.ProviderMatchID
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
