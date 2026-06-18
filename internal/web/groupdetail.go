package web

import (
	"net/http"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

type standingRow struct {
	fixtures.GroupStanding
	URL     string
	Flag    string
	Current bool // the team whose page this is — highlighted on TeamDetailPage
}

// standingRows computes a group's standings with team-page links
// (TeamNamesLink guarantee) and national flags.
func standingRows(groupMatches []fixtures.Match) []standingRow {
	standings := fixtures.Standings(groupMatches)
	rows := make([]standingRow, len(standings))
	for i, s := range standings {
		rows[i] = standingRow{GroupStanding: s, URL: teamURL(s.Team), Flag: flagFor(s.Team)}
	}
	return rows
}

type groupDetailData struct {
	GroupName       string
	Standings       []standingRow
	Days            []dayGroup
	LastSyncedUTC   string
	LastSyncedLabel string
	AssetVersion    string
	Nav             navData
}

// groupDetail renders the GroupDetailPage surface: one group's
// standings table followed by its results and fixtures in kickoff
// order. Unknown groups are not found.
func groupDetail(store fixtures.Store, host string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupName := "Group " + r.PathValue("group")

		matches, state, ok := loadFixtures(w, r, store)
		if !ok {
			return
		}

		var group []fixtures.Match
		for _, m := range matches {
			if m.Stage == fixtures.StageGroup && m.GroupName == groupName {
				group = append(group, m)
			}
		}
		if len(group) == 0 {
			renderError(w, http.StatusNotFound, "Group not found",
				"We couldn't find that group — the World Cup runs in groups A to L. Pick one from the menu.")
			return
		}

		data := groupDetailData{
			GroupName:       groupName,
			Standings:       standingRows(group),
			AssetVersion:    assetVersion,
			LastSyncedUTC:   lastSynced(state),
			LastSyncedLabel: syncedLabel(state),
			Nav:             buildNav(matches, host),
		}
		data.Nav.ActiveGroupURL = "/groups/" + r.PathValue("group") // highlight this group in the nav
		data.Days = groupByDay(group)
		// On a group's own page the group is implied, so drop the
		// self-referential group pill from each card (as knockout cards
		// carry none). The accessible label, computed in viewOf, keeps
		// the group context.
		for i := range data.Days {
			for j := range data.Days[i].Matches {
				data.Days[i].Matches[j].GroupName = ""
				data.Days[i].Matches[j].GroupURL = ""
			}
		}

		render(w, "groupdetail.html.tmpl", data)
	}
}
