package web_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

// Obligation: surface-exposure.FixturesPage — every match exposes
// stage, group, teams, kickoff and venue.
func TestPageListsFixtures(t *testing.T) {
	rec := get(t, seeded(t, match("wc-1")), now, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body := rec.Body.String()
	for _, want := range []string{"Canada", "Mexico", "Estadio Azteca, Mexico City", "Group A"} {
		if !strings.Contains(body, want) {
			t.Errorf("page missing %q", want)
		}
	}
}

// Kickoff times are emitted as machine-readable UTC for browser-local
// conversion (ADR 0010).
func TestPageEmitsKickoffAsUTCDatetime(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()
	if !strings.Contains(body, `datetime="2026-06-13T18:00:00Z"`) {
		t.Errorf("page missing UTC datetime attribute for kickoff: %s", body)
	}
}

// Scores appear only once a match is finished.
func TestPageShowsScoreOnlyWhenFinished(t *testing.T) {
	finished := match("wc-1")
	finished.Status = fixtures.StatusFinished
	finished.HomeScore, finished.AwayScore = intp(2), intp(1)

	inPlay := withID(match("wc-2"), "wc-2")
	inPlay.HomeTeam, inPlay.AwayTeam = "Spain", "France"
	inPlay.Status = fixtures.StatusInPlay

	body := get(t, seeded(t, finished, inPlay), now, "/").Body.String()

	if !strings.Contains(body, "2&nbsp;–&nbsp;1") && !strings.Contains(body, "2 – 1") {
		t.Errorf("finished match score missing from page: %s", body)
	}
	if strings.Contains(body, "Spain 0") || strings.Contains(body, "France 0") {
		t.Errorf("in-play match must not show a score")
	}
	if !strings.Contains(body, "In play") {
		t.Errorf("in-play match should be labelled, page: %s", body)
	}
}

func TestPageShowsShootoutResult(t *testing.T) {
	m := match("wc-1")
	m.Stage, m.GroupName = fixtures.StageFinal, ""
	m.Status = fixtures.StatusFinished
	m.HomeScore, m.AwayScore = intp(3), intp(3)
	m.HomePenalties, m.AwayPenalties = intp(4), intp(2)

	body := get(t, seeded(t, m), now, "/").Body.String()
	if !strings.Contains(body, "4–2 pens") {
		t.Errorf("shootout result missing from page: %s", body)
	}
}

// Obligation: surface-exposure.FixturesPage — exposes last_synced_at.
func TestPageShowsLastSyncedTime(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()
	if !strings.Contains(body, `datetime="2026-06-13T12:00:00Z"`) {
		t.Errorf("page missing last-synced timestamp: %s", body)
	}
}

// Spec: venue exposed only when present. The column appears when any
// match has a venue and disappears entirely when none do.
func TestVenueColumnOmittedWhenProviderSendsNoVenues(t *testing.T) {
	noVenue := match("wc-1")
	noVenue.Venue = ""

	body := get(t, seeded(t, noVenue), now, "/").Body.String()
	if strings.Contains(body, "Venue") {
		t.Errorf("venue column shown despite no venue data: %s", body)
	}
}

func TestVenueColumnShownWhenVenuesAvailable(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()
	if !strings.Contains(body, "Venue") || !strings.Contains(body, "Estadio Azteca, Mexico City") {
		t.Errorf("venue column missing despite venue data: %s", body)
	}
}

// Guarantee: TeamFlags — the home team's flag precedes its name, the
// away team's flag follows its name.
func TestPageShowsFlagsAroundTeamNames(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()

	if !strings.Contains(body, "🇨🇦 Canada") {
		t.Errorf("home flag missing before home team name: %s", body)
	}
	if !strings.Contains(body, "Mexico 🇲🇽") {
		t.Errorf("away flag missing after away team name: %s", body)
	}
}

// Guarantee: TeamFlags — placeholder names render without a flag
// rather than with a wrong or broken one.
func TestPlaceholderTeamsRenderWithoutFlags(t *testing.T) {
	m := match("wc-final")
	m.Stage, m.GroupName = fixtures.StageFinal, ""
	m.HomeTeam, m.AwayTeam = "Winner SF1", "Winner SF2"

	body := get(t, seeded(t, m), now, "/").Body.String()
	if !strings.Contains(body, "Winner SF1") || !strings.Contains(body, "Winner SF2") {
		t.Fatalf("placeholder names missing: %s", body)
	}
	for _, r := range body {
		if r >= 0x1F1E6 && r <= 0x1F1FF { // regional indicator symbols
			t.Errorf("placeholder page contains a flag: %s", body)
			break
		}
	}
}

// The subscribe link must use the webcal scheme so calendar clients
// subscribe (and auto-update) rather than import a one-off copy.
func TestPageSubscribeLinkUsesWebcalScheme(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()
	if !strings.Contains(body, `href="webcal://motson.jamesmaggs.com/calendar.ics"`) {
		t.Errorf("page missing webcal subscribe link: %s", body)
	}
}

func TestPageMarksCancelledAndPostponedMatches(t *testing.T) {
	cancelled := match("wc-1")
	cancelled.Status = fixtures.StatusCancelled
	postponed := withID(match("wc-2"), "wc-2")
	postponed.Status = fixtures.StatusPostponed

	body := get(t, seeded(t, cancelled, postponed), now, "/").Body.String()
	if !strings.Contains(body, "Cancelled") {
		t.Errorf("cancelled match not marked: %s", body)
	}
	if !strings.Contains(body, "Postponed") {
		t.Errorf("postponed match not marked: %s", body)
	}
}
