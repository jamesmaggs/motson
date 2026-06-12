package web_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

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

// Guarantee: TeamDirectory — every named team once, alphabetical,
// flagged, linked; unnamed sides absent.
func TestTeamsDirectory(t *testing.T) {
	rec := get(t, teamSpread(t), now, "/teams")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`<a href="/teams/canada">`, `<a href="/teams/south-korea">`,
		`<a href="/teams/czechia">`, `<a href="/teams/france">`, `<a href="/teams/mexico">`,
		"🇰🇷",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("teams directory missing %q", want)
		}
	}
	main := mainContent(body)
	if got := strings.Count(main, `href="/teams/canada"`); got != 1 {
		t.Errorf("Canada listed %d times in the directory, want once", got)
	}
	if strings.Contains(main, "TBC") {
		t.Errorf("unnamed sides must not appear in the directory")
	}
	if canada, czechia := strings.Index(main, "/teams/canada"), strings.Index(main, "/teams/czechia"); canada > czechia {
		t.Errorf("directory not alphabetical")
	}
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
	// Exactly Canada's two fixtures — no other team's matches leak in
	// (team names legitimately appear in the other-teams navigation).
	if got := strings.Count(body, `<td class="score">`); got != 2 {
		t.Errorf("got %d fixture rows, want 2 (Canada's only)", got)
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
	if !strings.Contains(index, `<a href="/teams/canada">Canada</a>`) {
		t.Errorf("index match cells missing team link: %s", index)
	}
	if strings.Contains(index, `href="/teams/"`) {
		t.Errorf("TBC side must not link")
	}

	groupPage := get(t, st, now, "/groups/A").Body.String()
	if !strings.Contains(groupPage, `<a href="/teams/mexico">Mexico</a>`) {
		t.Errorf("group standings missing team link: %s", groupPage)
	}

	if !strings.Contains(index, `href="/teams"`) {
		t.Errorf("root page missing link to the teams directory")
	}
}
