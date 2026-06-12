package footballdata_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jamesmaggs/motson/internal/fixtures"
	"github.com/jamesmaggs/motson/internal/footballdata"
)

const sampleResponse = `{
  "matches": [
    {
      "id": 501,
      "utcDate": "2026-06-13T18:00:00Z",
      "status": "TIMED",
      "stage": "GROUP_STAGE",
      "group": "GROUP_A",
      "venue": "Estadio Azteca",
      "homeTeam": {"name": "Canada"},
      "awayTeam": {"name": "Mexico"},
      "score": {"winner": null, "duration": "REGULAR",
                "fullTime": {"home": null, "away": null}}
    },
    {
      "id": 502,
      "utcDate": "2026-06-13T21:00:00Z",
      "status": "IN_PLAY",
      "stage": "GROUP_STAGE",
      "group": "GROUP_B",
      "venue": "MetLife Stadium",
      "homeTeam": {"name": "Spain"},
      "awayTeam": {"name": "France"},
      "score": {"winner": null, "duration": "REGULAR",
                "fullTime": {"home": 1, "away": 0}}
    },
    {
      "id": 503,
      "utcDate": "2026-07-19T20:00:00Z",
      "status": "FINISHED",
      "stage": "FINAL",
      "group": null,
      "venue": "MetLife Stadium",
      "homeTeam": {"name": "Winner SF1"},
      "awayTeam": {"name": "Winner SF2"},
      "score": {"winner": "HOME_TEAM", "duration": "PENALTY_SHOOTOUT",
                "fullTime": {"home": 3, "away": 3},
                "penalties": {"home": 4, "away": 2}}
    }
  ]
}`

func client(t *testing.T, handler http.HandlerFunc) *footballdata.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return footballdata.New(srv.URL, "test-token", "WC")
}

func fetch(t *testing.T) []fixtures.Match {
	t.Helper()
	c := client(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Auth-Token"); got != "test-token" {
			t.Errorf("X-Auth-Token = %q, want test-token", got)
		}
		if r.URL.Path != "/v4/competitions/WC/matches" {
			t.Errorf("path = %q, want /v4/competitions/WC/matches", r.URL.Path)
		}
		w.Write([]byte(sampleResponse))
	})
	matches, err := c.FetchFixtures(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return matches
}

// Obligation: contract-signature.FixtureSource.fetch_fixtures — one
// call returns the full snapshot mapped onto the domain model
// (entity-fields.ProviderFixture).
func TestFetchMapsScheduledMatch(t *testing.T) {
	m := fetch(t)[0]

	want := fixtures.Match{
		ProviderMatchID: "501",
		HomeTeam:        "Canada",
		AwayTeam:        "Mexico",
		KickoffAt:       time.Date(2026, 6, 13, 18, 0, 0, 0, time.UTC),
		Venue:           "Estadio Azteca",
		Stage:           fixtures.StageGroup,
		GroupName:       "Group A",
		Status:          fixtures.StatusScheduled,
	}
	assertEqual(t, m, want)
}

// In-play matches carry no scores: the spec exposes scores only once
// finished, whatever the provider reports mid-match.
func TestFetchMapsInPlayMatchWithoutScores(t *testing.T) {
	m := fetch(t)[1]

	if m.Status != fixtures.StatusInPlay {
		t.Errorf("Status = %s, want in_play", m.Status)
	}
	if m.HomeScore != nil || m.AwayScore != nil {
		t.Errorf("in-play match must carry no scores, got %v-%v", m.HomeScore, m.AwayScore)
	}
}

// Knockout placeholders pass through verbatim (pure mirror) and
// shootout results map onto the penalty fields.
func TestFetchMapsFinishedFinalWithShootout(t *testing.T) {
	m := fetch(t)[2]

	if m.HomeTeam != "Winner SF1" || m.AwayTeam != "Winner SF2" {
		t.Errorf("placeholder names not mirrored: %s vs %s", m.HomeTeam, m.AwayTeam)
	}
	if m.Stage != fixtures.StageFinal || m.GroupName != "" {
		t.Errorf("Stage/Group = %s/%q, want final/empty", m.Stage, m.GroupName)
	}
	if m.Status != fixtures.StatusFinished {
		t.Fatalf("Status = %s, want finished", m.Status)
	}
	if m.HomeScore == nil || *m.HomeScore != 3 || m.AwayScore == nil || *m.AwayScore != 3 {
		t.Errorf("scores = %v-%v, want 3-3", m.HomeScore, m.AwayScore)
	}
	if m.HomePenalties == nil || *m.HomePenalties != 4 || m.AwayPenalties == nil || *m.AwayPenalties != 2 {
		t.Errorf("penalties = %v-%v, want 4-2", m.HomePenalties, m.AwayPenalties)
	}
}

// ADR 0003: every provider status maps onto the spec's five statuses.
func TestStatusMapping(t *testing.T) {
	cases := map[string]fixtures.Status{
		"SCHEDULED": fixtures.StatusScheduled,
		"TIMED":     fixtures.StatusScheduled,
		"IN_PLAY":   fixtures.StatusInPlay,
		"PAUSED":    fixtures.StatusInPlay,
		"FINISHED":  fixtures.StatusFinished,
		"AWARDED":   fixtures.StatusFinished,
		"POSTPONED": fixtures.StatusPostponed,
		"SUSPENDED": fixtures.StatusPostponed,
		"CANCELLED": fixtures.StatusCancelled,
	}
	for providerStatus, want := range cases {
		c := client(t, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"matches": [{"id": 1, "utcDate": "2026-06-13T18:00:00Z",
				"status": "` + providerStatus + `", "stage": "GROUP_STAGE", "group": "GROUP_A",
				"homeTeam": {"name": "A"}, "awayTeam": {"name": "B"},
				"score": {"fullTime": {"home": null, "away": null}}}]}`))
		})
		matches, err := c.FetchFixtures(context.Background())
		if err != nil {
			t.Fatalf("%s: %v", providerStatus, err)
		}
		if got := matches[0].Status; got != want {
			t.Errorf("provider status %s mapped to %s, want %s", providerStatus, got, want)
		}
	}
}

func TestStageMapping(t *testing.T) {
	cases := map[string]fixtures.Stage{
		"GROUP_STAGE":    fixtures.StageGroup,
		"LAST_32":        fixtures.StageRoundOf32,
		"LAST_16":        fixtures.StageRoundOf16,
		"QUARTER_FINALS": fixtures.StageQuarterFinal,
		"SEMI_FINALS":    fixtures.StageSemiFinal,
		"THIRD_PLACE":    fixtures.StageThirdPlace,
		"FINAL":          fixtures.StageFinal,
	}
	for providerStage, want := range cases {
		c := client(t, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"matches": [{"id": 1, "utcDate": "2026-06-13T18:00:00Z",
				"status": "TIMED", "stage": "` + providerStage + `", "group": null,
				"homeTeam": {"name": "A"}, "awayTeam": {"name": "B"},
				"score": {"fullTime": {"home": null, "away": null}}}]}`))
		})
		matches, err := c.FetchFixtures(context.Background())
		if err != nil {
			t.Fatalf("%s: %v", providerStage, err)
		}
		if got := matches[0].Stage; got != want {
			t.Errorf("provider stage %s mapped to %s, want %s", providerStage, got, want)
		}
	}
}

func TestFetchReportsHTTPErrors(t *testing.T) {
	c := client(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	})
	if _, err := c.FetchFixtures(context.Background()); err == nil {
		t.Error("want error on non-200 response")
	}
}

func assertEqual(t *testing.T, got, want fixtures.Match) {
	t.Helper()
	if got.ProviderMatchID != want.ProviderMatchID ||
		got.HomeTeam != want.HomeTeam ||
		got.AwayTeam != want.AwayTeam ||
		!got.KickoffAt.Equal(want.KickoffAt) ||
		got.Venue != want.Venue ||
		got.Stage != want.Stage ||
		got.GroupName != want.GroupName ||
		got.Status != want.Status {
		t.Errorf("match = %+v, want %+v", got, want)
	}
}
