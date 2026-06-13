package web_test

import (
	"strings"
	"testing"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

func navSpread(t *testing.T) fixtures.Store {
	t.Helper()
	groupB := withID(match("wc-b"), "wc-b")
	groupB.GroupName = "Group B"
	groupB.HomeTeam, groupB.AwayTeam = "Spain", "Argentina"
	return seeded(t, match("wc-a1"), groupB)
}

// The sidebar appears on every page, listing each group (linked) and
// every team alphabetically with a flag, linked to its page.
func TestSidebarListsGroupsAndTeams(t *testing.T) {
	for _, path := range []string{"/", "/groups/A", "/teams/canada"} {
		body := get(t, navSpread(t), now, path).Body.String()

		if !strings.Contains(body, `class="sidebar"`) {
			t.Errorf("%s: sidebar missing", path)
			continue
		}
		if !strings.Contains(body, `<div class="nav-group-grid">`) || !strings.Contains(body, `<a class="group-badge" href="/groups/A">A</a>`) {
			t.Errorf("%s: group letter badges missing from sidebar", path)
		}
		if !strings.Contains(body, `<ul class="nav-teams">`) ||
			!strings.Contains(body, `<a href="/teams/argentina"><span class="flag">🇦🇷</span>Argentina</a>`) {
			t.Errorf("%s: team link with flag missing from sidebar: %s", path, body)
		}
	}
}

// The sidebar's first entry is an "Add to Calendar" link (with an
// icon) to the webcal feed, above the groups.
func TestSidebarHasAddToCalendar(t *testing.T) {
	for _, path := range []string{"/", "/groups/A", "/teams/canada"} {
		body := get(t, navSpread(t), now, path).Body.String()

		if !strings.Contains(body, `href="webcal://motson.jamesmaggs.com/calendar.ics"`) {
			t.Errorf("%s: Add to Calendar link missing", path)
		}
		if !strings.Contains(body, "Add to Calendar") {
			t.Errorf("%s: Add to Calendar text missing", path)
		}
		cal := strings.Index(body, "Add to Calendar")
		groups := strings.Index(body, `class="nav-heading">Groups`)
		if cal < 0 || groups < 0 || cal > groups {
			t.Errorf("%s: Add to Calendar should sit above the Groups heading", path)
		}
		// An icon precedes the link text.
		entry := body[strings.Index(body, `class="add-cal"`):]
		if i := strings.Index(entry, "Add to Calendar"); i < 0 || !strings.Contains(entry[:i], "<svg") {
			t.Errorf("%s: calendar icon missing before the link text", path)
		}
	}
}

// The sidebar has a collapse toggle (for small screens) that starts
// collapsed and controls the nav.
func TestSidebarHasCollapseToggle(t *testing.T) {
	body := get(t, navSpread(t), now, "/").Body.String()

	if !strings.Contains(body, `class="nav-toggle"`) {
		t.Errorf("menu toggle button missing")
	}
	if !strings.Contains(body, `aria-controls="sidebar-nav"`) || !strings.Contains(body, `id="sidebar-nav"`) {
		t.Errorf("toggle does not control the nav by id")
	}
	if !strings.Contains(body, `aria-expanded="false"`) {
		t.Errorf("menu should start collapsed (aria-expanded=false)")
	}
}

// Sidebar teams are alphabetical and unnamed sides are excluded.
func TestSidebarTeamsAlphabeticalAndNamed(t *testing.T) {
	tbc := withID(match("wc-final"), "wc-final")
	tbc.Stage, tbc.GroupName = fixtures.StageFinal, ""
	tbc.HomeTeam, tbc.AwayTeam = "", ""
	body := get(t, seeded(t, match("wc-a1"), tbc), now, "/").Body.String()

	sidebar := body[strings.Index(body, `class="nav-teams"`):]
	sidebar = sidebar[:strings.Index(sidebar, "</ul>")]
	if canada, mexico := strings.Index(sidebar, "Canada"), strings.Index(sidebar, "Mexico"); canada < 0 || mexico < 0 || canada > mexico {
		t.Errorf("sidebar teams not alphabetical: %s", sidebar)
	}
	if strings.Contains(sidebar, "TBC") {
		t.Errorf("sidebar must not list unnamed sides")
	}
}
