package web

import (
	"net/http"
	"sort"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

type groupsData struct {
	Groups        []groupSection
	LastSyncedUTC string
	AssetVersion  string
	HasVenues     bool
}

type groupSection struct {
	Name    string
	Matches []matchView
}

// groups renders the GroupFixturesPage surface: group-stage matches
// only, one section per group, groups in alphabetical order, kickoff
// order within each (preserved from the store's ordering).
func groups(store fixtures.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		data := groupsData{AssetVersion: assetVersion, LastSyncedUTC: lastSynced(state)}
		index := map[string]int{}
		for _, m := range matches {
			if m.Stage != fixtures.StageGroup {
				continue
			}
			i, ok := index[m.GroupName]
			if !ok {
				i = len(data.Groups)
				index[m.GroupName] = i
				data.Groups = append(data.Groups, groupSection{Name: m.GroupName})
			}
			data.Groups[i].Matches = append(data.Groups[i].Matches, viewOf(m))
			if m.Venue != "" {
				data.HasVenues = true
			}
		}
		sort.Slice(data.Groups, func(i, j int) bool { return data.Groups[i].Name < data.Groups[j].Name })

		render(w, "groups.html.tmpl", data)
	}
}
