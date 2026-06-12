package fixtures_test

import (
	"testing"
	"time"

	"github.com/Jazzatola/motson/internal/fixtures"
)

// Obligations: config-default.sync_interval, config-default.group_match_duration,
// config-default.knockout_match_duration, config-default.staleness_threshold
func TestConfigDefaults(t *testing.T) {
	if got, want := fixtures.SyncInterval, time.Hour; got != want {
		t.Errorf("SyncInterval = %v, want %v", got, want)
	}
	if got, want := fixtures.GroupMatchDuration, 2*time.Hour; got != want {
		t.Errorf("GroupMatchDuration = %v, want %v", got, want)
	}
	if got, want := fixtures.KnockoutMatchDuration, 2*time.Hour+45*time.Minute; got != want {
		t.Errorf("KnockoutMatchDuration = %v, want %v", got, want)
	}
	if got, want := fixtures.StalenessThreshold, 3*time.Hour; got != want {
		t.Errorf("StalenessThreshold = %v, want %v", got, want)
	}
}

// Obligation: derived.Match.ends_at — kickoff + 2h for group matches,
// kickoff + 2h45 for every knockout stage.
func TestMatchEndsAt(t *testing.T) {
	kickoff := time.Date(2026, 6, 13, 18, 0, 0, 0, time.UTC)

	cases := []struct {
		stage fixtures.Stage
		want  time.Time
	}{
		{fixtures.StageGroup, kickoff.Add(2 * time.Hour)},
		{fixtures.StageRoundOf32, kickoff.Add(2*time.Hour + 45*time.Minute)},
		{fixtures.StageRoundOf16, kickoff.Add(2*time.Hour + 45*time.Minute)},
		{fixtures.StageQuarterFinal, kickoff.Add(2*time.Hour + 45*time.Minute)},
		{fixtures.StageSemiFinal, kickoff.Add(2*time.Hour + 45*time.Minute)},
		{fixtures.StageThirdPlace, kickoff.Add(2*time.Hour + 45*time.Minute)},
		{fixtures.StageFinal, kickoff.Add(2*time.Hour + 45*time.Minute)},
	}
	for _, c := range cases {
		m := fixtures.Match{Stage: c.stage, KickoffAt: kickoff}
		if got := m.EndsAt(); !got.Equal(c.want) {
			t.Errorf("stage %s: EndsAt() = %v, want %v", c.stage, got, c.want)
		}
	}
}
