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
		if !strings.Contains(body, `<ul class="nav-groups">`) || !strings.Contains(body, `<a href="/groups/A">Group A</a>`) {
			t.Errorf("%s: group links missing from sidebar", path)
		}
		if !strings.Contains(body, `<ul class="nav-teams">`) ||
			!strings.Contains(body, `<a href="/teams/argentina"><span class="flag">🇦🇷</span>Argentina</a>`) {
			t.Errorf("%s: team link with flag missing from sidebar: %s", path, body)
		}
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
