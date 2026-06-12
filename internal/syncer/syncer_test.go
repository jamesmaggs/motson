package syncer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Jazzatola/motson/internal/fixtures"
	"github.com/Jazzatola/motson/internal/store"
	"github.com/Jazzatola/motson/internal/syncer"
)

type fakeSource struct {
	matches []fixtures.Match
	err     error
	calls   int
}

func (f *fakeSource) FetchFixtures(ctx context.Context) ([]fixtures.Match, error) {
	f.calls++
	return f.matches, f.err
}

func match(id string) fixtures.Match {
	return fixtures.Match{
		ProviderMatchID: id,
		HomeTeam:        "Canada",
		AwayTeam:        "Mexico",
		KickoffAt:       time.Date(2026, 6, 13, 18, 0, 0, 0, time.UTC),
		Stage:           fixtures.StageGroup,
		GroupName:       "Group A",
		Status:          fixtures.StatusScheduled,
	}
}

func setup() (*fakeSource, *store.Memory, *syncer.Syncer) {
	src := &fakeSource{matches: []fixtures.Match{match("wc-1")}}
	st := store.NewMemory()
	return src, st, syncer.New(src, st)
}

var now = time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)

// Obligation: temporal.SyncDue — fires at the deadline, not before.
func TestSyncDoesNothingBeforeDeadline(t *testing.T) {
	ctx := context.Background()
	src, st, s := setup()
	if err := st.ScheduleNextSync(ctx, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	if err := s.RunDue(ctx, now); err != nil {
		t.Fatal(err)
	}

	if src.calls != 0 {
		t.Errorf("fetch calls = %d, want 0 before deadline", src.calls)
	}
	matches, _ := st.Matches(ctx)
	if len(matches) != 0 {
		t.Errorf("got %d matches, want 0", len(matches))
	}
}

// Obligations: rule-success.SyncDue, rule-success.FixtureDataApplied —
// at the deadline the schedule advances by sync_interval and the
// snapshot is applied.
func TestSyncAtDeadlineAppliesSnapshotAndReschedules(t *testing.T) {
	ctx := context.Background()
	src, st, s := setup()
	if err := st.ScheduleNextSync(ctx, now); err != nil { // exactly due
		t.Fatal(err)
	}

	if err := s.RunDue(ctx, now); err != nil {
		t.Fatal(err)
	}

	if src.calls != 1 {
		t.Fatalf("fetch calls = %d, want 1", src.calls)
	}
	matches, _ := st.Matches(ctx)
	if len(matches) != 1 || matches[0].ProviderMatchID != "wc-1" {
		t.Errorf("matches = %v, want [wc-1]", matches)
	}
	state, _ := st.SyncState(ctx)
	if want := now.Add(fixtures.SyncInterval); !state.NextRunAt.Equal(want) {
		t.Errorf("NextRunAt = %v, want %v", state.NextRunAt, want)
	}
	if state.LastSyncedAt == nil || !state.LastSyncedAt.Equal(now) {
		t.Errorf("LastSyncedAt = %v, want %v", state.LastSyncedAt, now)
	}
}

// A fresh store (zero NextRunAt) is due immediately: the boot sync.
func TestSyncRunsImmediatelyOnFreshStore(t *testing.T) {
	ctx := context.Background()
	src, _, s := setup()
	if err := s.RunDue(ctx, now); err != nil {
		t.Fatal(err)
	}
	if src.calls != 1 {
		t.Errorf("fetch calls = %d, want 1 on fresh store", src.calls)
	}
}

// Obligation: temporal.SyncDue — does not re-fire once handled.
func TestSyncDoesNotRefireWithinInterval(t *testing.T) {
	ctx := context.Background()
	src, _, s := setup()

	for range 3 {
		if err := s.RunDue(ctx, now); err != nil {
			t.Fatal(err)
		}
	}

	if src.calls != 1 {
		t.Errorf("fetch calls = %d, want 1 (no re-fire before next deadline)", src.calls)
	}
}

// Guarantee: StalenessTolerated — a failed fetch leaves stored data
// untouched and the next attempt waits for the advanced deadline.
func TestFailedFetchLeavesStoreUntouched(t *testing.T) {
	ctx := context.Background()
	src, st, s := setup()
	if err := s.RunDue(ctx, now); err != nil { // seed a good snapshot
		t.Fatal(err)
	}

	src.err = errors.New("provider down")
	later := now.Add(fixtures.SyncInterval)
	if err := s.RunDue(ctx, later); err == nil {
		t.Error("want error reported when fetch fails")
	}

	matches, _ := st.Matches(ctx)
	if len(matches) != 1 {
		t.Errorf("got %d matches, want 1 preserved", len(matches))
	}
	state, _ := st.SyncState(ctx)
	if state.LastSyncedAt == nil || !state.LastSyncedAt.Equal(now) {
		t.Errorf("LastSyncedAt = %v, want unchanged %v", state.LastSyncedAt, now)
	}
	if want := later.Add(fixtures.SyncInterval); !state.NextRunAt.Equal(want) {
		t.Errorf("NextRunAt = %v, want %v (retry at next interval)", state.NextRunAt, want)
	}
}
