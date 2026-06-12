package web_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

func groupStageSpread(t *testing.T) fixtures.Store {
	t.Helper()
	groupB := withID(match("wc-b"), "wc-b")
	groupB.GroupName = "Group B"
	groupB.HomeTeam, groupB.AwayTeam = "Spain", "France"

	laterA := withID(match("wc-a2"), "wc-a2")
	laterA.HomeTeam, laterA.AwayTeam = "Honduras", "Jamaica"
	laterA.KickoffAt = laterA.KickoffAt.Add(48 * time.Hour)

	knockout := withID(match("wc-sf"), "wc-sf")
	knockout.Stage, knockout.GroupName = fixtures.StageSemiFinal, ""
	knockout.HomeTeam, knockout.AwayTeam = "Brazil", "Germany"

	return seeded(t, match("wc-a1"), laterA, groupB, knockout)
}

// Guarantee: GroupStageOnly — one section per group; knockout
// fixtures are absent.
func TestGroupsViewShowsOnlyGroupGamesBySection(t *testing.T) {
	rec := get(t, groupStageSpread(t), now, "/groups")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Group A", "Group B", "Canada", "Spain", "Honduras"} {
		if !strings.Contains(body, want) {
			t.Errorf("groups view missing %q", want)
		}
	}
	if strings.Contains(body, "Brazil") || strings.Contains(body, "Semi-final") {
		t.Errorf("knockout fixture leaked into groups view: %s", body)
	}
}

// Groups are sections in alphabetical order; within a group, matches
// keep kickoff order.
func TestGroupsAreOrderedAndMatchesKeepKickoffOrder(t *testing.T) {
	body := get(t, groupStageSpread(t), now, "/groups").Body.String()

	posA, posB := strings.Index(body, "Group A"), strings.Index(body, "Group B")
	if posA < 0 || posB < 0 || posA > posB {
		t.Errorf("groups out of order: A at %d, B at %d", posA, posB)
	}
	if first, second := strings.Index(body, "Canada"), strings.Index(body, "Honduras"); first > second {
		t.Errorf("matches within Group A out of kickoff order")
	}
}

// The view shares the fixture row treatment (flags hugging a
// dedicated score cell) and shows the last-synced time.
func TestGroupsViewSharesFixtureTreatment(t *testing.T) {
	body := get(t, groupStageSpread(t), now, "/groups").Body.String()

	if !strings.Contains(body, `Canada</a> 🇨🇦`) || !strings.Contains(body, `🇲🇽 <a`) {
		t.Errorf("flag treatment missing on groups view: %s", body)
	}
	if got := strings.Count(body, `<td class="score">`); got != 3 {
		t.Errorf("got %d score cells, want 3", got)
	}
	if !strings.Contains(body, `datetime="2026-06-13T12:00:00Z"`) {
		t.Errorf("last-synced time missing from groups view")
	}
}

// Group section headings link to their group's detail page.
func TestGroupHeadingsLinkToGroupPages(t *testing.T) {
	body := get(t, groupStageSpread(t), now, "/groups").Body.String()

	for _, want := range []string{`<h2><a href="/groups/A">Group A</a></h2>`, `<h2><a href="/groups/B">Group B</a></h2>`} {
		if !strings.Contains(body, want) {
			t.Errorf("groups view missing linked heading %q", want)
		}
	}
}

// The two views link to each other (spec: related surfaces).
func TestViewsLinkToEachOther(t *testing.T) {
	st := groupStageSpread(t)
	if body := get(t, st, now, "/").Body.String(); !strings.Contains(body, `href="/groups"`) {
		t.Errorf("main view missing link to groups view")
	}
	if body := get(t, st, now, "/groups").Body.String(); !strings.Contains(body, `href="/"`) {
		t.Errorf("groups view missing link back to all fixtures")
	}
}
