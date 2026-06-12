package web

import (
	"net/http"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

type groupDetailData struct {
	GroupName     string
	Standings     []fixtures.GroupStanding
	Matches       []matchView
	LastSyncedUTC string
	AssetVersion  string
	HasVenues     bool
}

// groupDetail renders the GroupDetailPage surface: one group's
// standings table followed by its results and fixtures in kickoff
// order. Unknown groups are not found.
func groupDetail(store fixtures.Store) http.HandlerFunc {
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
			Standings:     fixtures.Standings(group),
			AssetVersion:  assetVersion,
			LastSyncedUTC: lastSynced(state),
		}
		data.Matches, data.HasVenues = buildViews(group)

		render(w, "groupdetail.html.tmpl", data)
	}
}
