package web

import (
	"sort"
	"strings"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

// navData is the site-wide sidebar: group links then every team,
// alphabetically, with flags. Rebuilt per request from the store.
type navData struct {
	FeedHost string // anchors the webcal calendar link (scheme is literal in the template)
	Groups   []navLink
	Teams    []teamEntry
}

type navLink struct {
	Name   string
	Letter string
	URL    string
}

// buildNav derives the sidebar from the current matches. host anchors
// the webcal calendar link.
func buildNav(matches []fixtures.Match, host string) navData {
	seen := map[string]bool{}
	var groups []navLink
	for _, m := range matches {
		if m.Stage != fixtures.StageGroup || m.GroupName == "" || seen[m.GroupName] {
			continue
		}
		seen[m.GroupName] = true
		if letter, ok := strings.CutPrefix(m.GroupName, "Group "); ok {
			groups = append(groups, navLink{Name: m.GroupName, Letter: letter, URL: "/groups/" + letter})
		}
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Name < groups[j].Name })
	return navData{
		FeedHost: host,
		Groups:   groups,
		Teams:    collectTeams(matches),
	}
}
