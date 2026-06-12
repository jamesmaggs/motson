package web

import (
	"net/http"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

type standingRow struct {
	fixtures.GroupStanding
	URL string
}

// standingRows computes a group's standings with team-page links
// (TeamNamesLink guarantee).
func standingRows(groupMatches []fixtures.Match) []standingRow {
	standings := fixtures.Standings(groupMatches)
	rows := make([]standingRow, len(standings))
	for i, s := range standings {
		rows[i] = standingRow{GroupStanding: s, URL: teamURL(s.Team)}
	}
	return rows
}

type groupDetailData struct {
	GroupName     string
	Standings     []standingRow
	Matches       []matchView
	LastSyncedUTC string
	AssetVersion  string
	HasVenues     bool
	Nav           navData
}

// groupDetail renders the GroupDetailPage surface: one group's
// standings table followed by its results and fixtures in kickoff
// order. Unknown groups are not found.
func groupDetail(store fixtures.Store, host string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupName := "Group " + r.PathValue("group")

		matches, err := store.Matches(r.Context())
		if err != nil {
			http.Error(w, "fixtures unavailable", http.StatusInternalServerError)
			return
		}
		state, err := store.SyncState(r.Context())
		if err != nil {
			http.Error(w, "fixtures unavailable", http.StatusInternalServerError)
			return
		}

		var group []fixtures.Match
		for _, m := range matches {
			if m.Stage == fixtures.StageGroup && m.GroupName == groupName {
				group = append(group, m)
			}
		}
		if len(group) == 0 {
			http.NotFound(w, r)
			return
		}

		data := groupDetailData{
			GroupName:     groupName,
			Standings:     standingRows(group),
			AssetVersion:  assetVersion,
			LastSyncedUTC: lastSynced(state),
			Nav:           buildNav(matches, host),
		}
		data.Matches, data.HasVenues = buildViews(group)

		render(w, "groupdetail.html.tmpl", data)
	}
}
