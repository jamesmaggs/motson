package web_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

func groupA(t *testing.T) fixtures.Store {
	t.Helper()
	finished := match("wc-a1")
	finished.Status = fixtures.StatusFinished
	finished.HomeScore, finished.AwayScore = intp(2), intp(0)

	upcoming := withID(match("wc-a2"), "wc-a2")
	upcoming.HomeTeam, upcoming.AwayTeam = "Honduras", "Jamaica"

	other := withID(match("wc-b"), "wc-b")
	other.GroupName = "Group B"
	other.HomeTeam, other.AwayTeam = "Spain", "France"

	return seeded(t, finished, upcoming, other)
}

// Guarantee: OnePagePerGroup — the page shows its own group's
// standings then matches; other groups are absent.
func TestGroupDetailShowsStandingsThenMatches(t *testing.T) {
	rec := get(t, groupA(t), now, "/groups/A")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	for _, want := range []string{"Group A", "Pts", "Canada", "Mexico", "Honduras", "Jamaica"} {
		if !strings.Contains(body, want) {
			t.Errorf("group page missing %q", want)
		}
	}
	if main := mainContent(body); strings.Contains(main, "Spain") || strings.Contains(main, "Group B") {
		t.Errorf("other group leaked into /groups/A: %s", main)
	}
	// The standings table comes first in source order (spec:
	// OnePagePerGroup), even though on wide screens it sits to the
	// right of the matches.
	if standings, matches := strings.Index(body, "Pts"), strings.Index(body, `<article class="card`); standings < 0 || matches < 0 || standings > matches {
		t.Errorf("standings table must precede the match list in source order")
	}
}

// Group matches use the shared card layout (as on the index), and the
// standings sit in a matching card container.
func TestGroupDetailMatchesAreCardsAndTableIsCarded(t *testing.T) {
	body := get(t, groupA(t), now, "/groups/A").Body.String()

	if got := strings.Count(body, `<article class="card`); got != 2 {
		t.Errorf("got %d match cards, want one per group-A match (2): %s", got, body)
	}
	if !strings.Contains(body, `class="standings-card"`) {
		t.Errorf("standings should sit in a card container: %s", body)
	}
	// Standings list each team with its national flag beside the name.
	if !strings.Contains(body, `<a href="/teams/canada"><span class="flag">🇨🇦</span> Canada</a>`) {
		t.Errorf("standings row should carry the team flag: %s", body)
	}
	// The group's own page does not repeat the redundant "Group A" pill
	// on each card (it links to the page you're already on).
	if strings.Contains(mainContent(body), `class="group"`) {
		t.Errorf("group-page cards should not carry a self-linking group pill")
	}
}

// StandingsOrder on the page: Canada (3pts) tops the table; the
// winner's row carries its record.
func TestGroupDetailStandingsContent(t *testing.T) {
	body := get(t, groupA(t), now, "/groups/A").Body.String()

	// Scope to the standings card: the nav menu also names both teams, in
	// FIFA-ranking order, which is not the standings order under test.
	standings := body[strings.Index(body, `class="standings-card"`):]
	canada, mexico := strings.Index(standings, ">Canada"), strings.Index(standings, ">Mexico")
	if canada < 0 || mexico < 0 || canada > mexico {
		t.Errorf("Canada (3pts) should precede Mexico in standings")
	}
}

// Wayfinding: on a group page, that group's nav circle is marked active.
func TestGroupPageHighlightsActiveGroupInNav(t *testing.T) {
	body := get(t, groupA(t), now, "/groups/A").Body.String()

	if !strings.Contains(body, `class="group-badge active" href="/groups/A" aria-current="page"`) {
		t.Errorf("active group circle not highlighted in nav: %s", body)
	}
	// Only the current group is active.
	if strings.Count(body, "group-badge active") != 1 {
		t.Errorf("expected exactly one active group circle")
	}
}

func TestUnknownGroupIsNotFound(t *testing.T) {
	if rec := get(t, groupA(t), now, "/groups/Z"); rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 for unknown group", rec.Code)
	}
}

// The group pill on each card links to that group's page; knockout
// cards carry no group pill.
func TestGroupPillLinksToGroupPage(t *testing.T) {
	knockout := withID(match("wc-sf"), "wc-sf")
	knockout.Stage, knockout.GroupName = fixtures.StageSemiFinal, ""
	knockout.HomeTeam, knockout.AwayTeam = "Brazil", "Germany"

	body := get(t, seeded(t, match("wc-a1"), knockout), now, "/").Body.String()

	if !strings.Contains(body, `<a class="group" href="/groups/A">Group A</a>`) {
		t.Errorf("group pill not linked: %s", body)
	}
	if strings.Contains(body, `href="/groups/"`) {
		t.Errorf("knockout card must not link a group")
	}
}
