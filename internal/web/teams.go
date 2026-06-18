package web

import (
	"net/http"
	"sort"
	"strings"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

// slugify addresses a team page: lowercased, spaces to hyphens
// ("South Korea" -> "south-korea"), per the OnePagePerTeam guarantee.
func slugify(team string) string {
	return strings.ToLower(strings.ReplaceAll(team, " ", "-"))
}

// teamURL links a named team's page; unnamed sides get no link
// (TeamNamesLink guarantee).
func teamURL(team string) string {
	if team == "" {
		return ""
	}
	return "/teams/" + slugify(team)
}

type teamEntry struct {
	Name string
	Flag string
	URL  string
}

// collectTeams lists every named team exactly once, ordered by FIFA world
// ranking (the MenuOrder guarantee). Teams with no known ranking follow the
// ranked ones, alphabetically among themselves.
func collectTeams(matches []fixtures.Match) []teamEntry {
	seen := map[string]bool{}
	var teams []teamEntry
	for _, m := range matches {
		for _, name := range []string{m.HomeTeam, m.AwayTeam} {
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			teams = append(teams, teamEntry{Name: name, Flag: flagFor(name), URL: teamURL(name)})
		}
	}
	sort.Slice(teams, func(i, j int) bool {
		ri, iok := rankOf(teams[i].Name)
		rj, jok := rankOf(teams[j].Name)
		if iok != jok {
			return iok // ranked teams sort ahead of unranked ones
		}
		if iok && ri != rj {
			return ri < rj
		}
		return teams[i].Name < teams[j].Name
	})
	return teams
}

type teamDetailData struct {
	TeamName        string
	Flag            string
	GroupName       string
	GroupURL        string
	Standings       []standingRow
	Days            []dayGroup
	LastSyncedUTC   string
	LastSyncedLabel string
	AssetVersion    string
	Nav             navData
}

// teamDetail renders the TeamDetailPage surface: the team's group
// standings, then every fixture it appears in across all stages.
func teamDetail(store fixtures.Store, host string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("team")

		matches, state, ok := loadFixtures(w, r, store)
		if !ok {
			return
		}

		all := collectTeams(matches)
		var team teamEntry
		for _, t := range all {
			if slugify(t.Name) == slug {
				team = t
				break
			}
		}
		if team.Name == "" {
			renderError(w, http.StatusNotFound, "Team not found",
				"We couldn't find that team. Browse all teams from the menu.")
			return
		}

		data := teamDetailData{
			TeamName:        team.Name,
			Flag:            team.Flag,
			LastSyncedUTC:   lastSynced(state),
			LastSyncedLabel: syncedLabel(state),
			AssetVersion:    assetVersion,
			Nav:             buildNav(matches, host),
		}
		var own, groupMatches []fixtures.Match
		for _, m := range matches {
			if m.HomeTeam == team.Name || m.AwayTeam == team.Name {
				own = append(own, m)
				if m.Stage == fixtures.StageGroup && data.GroupName == "" {
					data.GroupName = m.GroupName
				}
			}
		}
		if data.GroupName != "" {
			for _, m := range matches {
				if m.Stage == fixtures.StageGroup && m.GroupName == data.GroupName {
					groupMatches = append(groupMatches, m)
				}
			}
			data.Standings = standingRows(groupMatches)
			for i := range data.Standings {
				data.Standings[i].Current = data.Standings[i].Team == team.Name
			}
			if letter, ok := strings.CutPrefix(data.GroupName, "Group "); ok {
				data.GroupURL = "/groups/" + letter
			}
		}
		data.Days = groupByDay(own)
		// Highlight this team — and its group — in the nav.
		data.Nav.ActiveTeamURL = teamURL(team.Name)
		data.Nav.ActiveGroupURL = data.GroupURL

		render(w, "teamdetail.html.tmpl", data)
	}
}
