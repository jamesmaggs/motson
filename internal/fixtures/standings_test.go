package fixtures_test

import (
	"testing"
	"time"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

func intp(i int) *int { return &i }

func groupMatch(home, away string, homeGoals, awayGoals int) fixtures.Match {
	return fixtures.Match{
		HomeTeam: home, AwayTeam: away,
		KickoffAt: time.Date(2026, 6, 13, 18, 0, 0, 0, time.UTC),
		Stage:     fixtures.StageGroup, GroupName: "Group A",
		Status:    fixtures.StatusFinished,
		HomeScore: intp(homeGoals), AwayScore: intp(awayGoals),
	}
}

func upcoming(home, away string) fixtures.Match {
	m := groupMatch(home, away, 0, 0)
	m.Status = fixtures.StatusScheduled
	m.HomeScore, m.AwayScore = nil, nil
	return m
}

// Spec value GroupStanding: played/won/drawn/lost, goals, derived
// goal_difference and points (3 for a win, 1 for a draw).
func TestStandingsComputeRecordAndPoints(t *testing.T) {
	standings := fixtures.Standings([]fixtures.Match{
		groupMatch("Canada", "Mexico", 2, 0),
		groupMatch("Mexico", "Honduras", 1, 1),
		groupMatch("Canada", "Honduras", 0, 3),
	})

	want := map[string]fixtures.GroupStanding{
		"Canada":   {Team: "Canada", Played: 2, Won: 1, Drawn: 0, Lost: 1, GoalsFor: 2, GoalsAgainst: 3},
		"Mexico":   {Team: "Mexico", Played: 2, Won: 0, Drawn: 1, Lost: 1, GoalsFor: 1, GoalsAgainst: 3},
		"Honduras": {Team: "Honduras", Played: 2, Won: 1, Drawn: 1, Lost: 0, GoalsFor: 4, GoalsAgainst: 1},
	}
	for _, s := range standings {
		w, ok := want[s.Team]
		if !ok {
			t.Errorf("unexpected team %q", s.Team)
			continue
		}
		if s.Played != w.Played || s.Won != w.Won || s.Drawn != w.Drawn || s.Lost != w.Lost ||
			s.GoalsFor != w.GoalsFor || s.GoalsAgainst != w.GoalsAgainst {
			t.Errorf("%s = %+v, want %+v", s.Team, s, w)
		}
	}
	byTeam := func(team string) fixtures.GroupStanding {
		for _, s := range standings {
			if s.Team == team {
				return s
			}
		}
		t.Fatalf("team %s missing", team)
		return fixtures.GroupStanding{}
	}
	if got := byTeam("Honduras").Points(); got != 4 {
		t.Errorf("Honduras points = %d, want 4", got)
	}
	if got := byTeam("Honduras").GoalDifference(); got != 3 {
		t.Errorf("Honduras GD = %d, want 3", got)
	}
}

// StandingsOrder guarantee: points, then goal difference, then goals
// scored, then team name.
func TestStandingsOrdering(t *testing.T) {
	standings := fixtures.Standings([]fixtures.Match{
		groupMatch("Canada", "Mexico", 2, 0),    // Canada 3pts GD+2
		groupMatch("Honduras", "Jamaica", 3, 1), // Honduras 3pts GD+2 GF3
	})

	got := make([]string, len(standings))
	for i, s := range standings {
		got[i] = s.Team
	}
	// Honduras beats Canada on goals scored; losers tie on all
	// criteria so Jamaica precedes Mexico alphabetically... but
	// Jamaica has GF1 vs Mexico GF0, so goals scored decides.
	want := []string{"Honduras", "Canada", "Jamaica", "Mexico"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

// Only finished matches contribute, but every named team appears —
// with zeroes before it has played. Unnamed (TBC) sides are ignored.
func TestStandingsIncludeUnplayedTeamsAndIgnoreUnfinished(t *testing.T) {
	inPlay := groupMatch("Canada", "Mexico", 1, 0)
	inPlay.Status = fixtures.StatusInPlay
	inPlay.HomeScore, inPlay.AwayScore = nil, nil

	standings := fixtures.Standings([]fixtures.Match{
		inPlay,
		upcoming("Honduras", "Jamaica"),
		upcoming("", "Canada"), // hypothetical unnamed side
	})

	if len(standings) != 4 {
		t.Fatalf("got %d standings, want 4 named teams", len(standings))
	}
	for _, s := range standings {
		if s.Played != 0 || s.Points() != 0 {
			t.Errorf("%s should be zeroed before playing: %+v", s.Team, s)
		}
	}
}
