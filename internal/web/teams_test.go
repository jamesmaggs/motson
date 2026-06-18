package web_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

// DatedFixtures: a team page groups its fixtures across stages by matchday.
func TestTeamDetailGroupsFixturesByDate(t *testing.T) {
	group := match("wc-tg")
	group.KickoffAt = time.Date(2026, 6, 13, 18, 0, 0, 0, time.UTC)

	knockout := withID(match("wc-tko"), "wc-tko")
	knockout.Stage, knockout.GroupName = fixtures.StageSemiFinal, ""
	knockout.HomeTeam, knockout.AwayTeam = "Canada", "France"
	knockout.KickoffAt = time.Date(2026, 7, 14, 19, 0, 0, 0, time.UTC)

	body := get(t, seeded(t, group, knockout), now, "/teams/canada").Body.String()
	if got := strings.Count(body, `class="day-heading"`); got != 2 {
		t.Errorf("team page should group fixtures by date (2 headings), got %d: %s", got, body)
	}
}

// A spread where South Korea plays in Group F and Canada reaches a
// semi-final, plus an undetermined fixture.
func teamSpread(t *testing.T) fixtures.Store {
	t.Helper()
	groupF := withID(match("wc-f"), "wc-f")
	groupF.GroupName = "Group F"
	groupF.HomeTeam, groupF.AwayTeam = "South Korea", "Czechia"
	groupF.Status = fixtures.StatusFinished
	groupF.HomeScore, groupF.AwayScore = intp(2), intp(1)

	knockout := withID(match("wc-sf"), "wc-sf")
	knockout.Stage, knockout.GroupName = fixtures.StageSemiFinal, ""
	knockout.HomeTeam, knockout.AwayTeam = "Canada", "France"

	tbc := withID(match("wc-final"), "wc-final")
	tbc.Stage, tbc.GroupName = fixtures.StageFinal, ""
	tbc.HomeTeam, tbc.AwayTeam = "", ""

	return seeded(t, match("wc-a1"), groupF, knockout, tbc)
}

// Guarantee: OnePagePerTeam — hyphenated slug, flag beside the title,
// group standings, and every fixture across stages.
func TestTeamDetailPage(t *testing.T) {
	rec := get(t, teamSpread(t), now, "/teams/canada")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	if !strings.Contains(body, "🇨🇦 Canada") {
		t.Errorf("flag missing beside team title: %s", body)
	}
	if !strings.Contains(body, "Pts") || !strings.Contains(body, ">Mexico<") {
		t.Errorf("group standings missing from team page")
	}
	if !strings.Contains(body, "Semi-final") {
		t.Errorf("knockout fixture missing from team page")
	}
	// Exactly Canada's two fixtures as cards — no other team's matches
	// leak in (team names legitimately appear in the sidebar nav).
	if got := strings.Count(body, `<article class="card`); got != 2 {
		t.Errorf("got %d fixture cards, want 2 (Canada's only)", got)
	}
}

// Guarantee: TeamNavigation — links to the group page and every other
// team's page.
func TestTeamPageLinksToGroupAndOtherTeams(t *testing.T) {
	body := get(t, teamSpread(t), now, "/teams/canada").Body.String()

	if !strings.Contains(body, `href="/groups/A"`) {
		t.Errorf("team page missing link to its group page")
	}
	for _, want := range []string{`href="/teams/france"`, `href="/teams/south-korea"`, `href="/teams/mexico"`} {
		if !strings.Contains(body, want) {
			t.Errorf("team page missing other-team link %q", want)
		}
	}
}

// Wayfinding: on a team page, both the team's nav link and its group's
// circle are marked active.
func TestTeamPageHighlightsTeamAndGroupInNav(t *testing.T) {
	body := get(t, teamSpread(t), now, "/teams/canada").Body.String()

	if !strings.Contains(body, `href="/teams/canada" class="active" aria-current="page"`) {
		t.Errorf("active team not highlighted in nav: %s", body)
	}
	if !strings.Contains(body, `class="group-badge active" href="/groups/A"`) {
		t.Errorf("team's group circle not highlighted in nav: %s", body)
	}
}

func TestHyphenatedTeamSlug(t *testing.T) {
	if rec := get(t, teamSpread(t), now, "/teams/south-korea"); rec.Code != http.StatusOK {
		t.Errorf("/teams/south-korea status = %d, want 200", rec.Code)
	}
	if rec := get(t, teamSpread(t), now, "/teams/atlantis"); rec.Code != http.StatusNotFound {
		t.Errorf("unknown team status = %d, want 404", rec.Code)
	}
}

// Guarantee: TeamNamesLink — names link wherever they appear; TBC
// sides do not.
func TestTeamNamesLinkEverywhere(t *testing.T) {
	st := teamSpread(t)

	index := get(t, st, now, "/").Body.String()
	if !strings.Contains(index, `<a class="team" href="/teams/canada">`) {
		t.Errorf("index card missing team link: %s", index)
	}
	if strings.Contains(index, `href="/teams/"`) {
		t.Errorf("TBC side must not link")
	}

	groupPage := get(t, st, now, "/groups/A").Body.String()
	if !strings.Contains(groupPage, `<a href="/teams/mexico"><span class="flag">🇲🇽</span> Mexico</a>`) {
		t.Errorf("group standings missing flagged team link: %s", groupPage)
	}
}
