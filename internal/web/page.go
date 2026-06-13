package web

import (
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

// assetVersion fingerprints static asset URLs per build so edge
// caches cannot serve a previous build's assets after a deploy.
var assetVersion = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, kv := range info.Settings {
			if kv.Key == "vcs.revision" && len(kv.Value) >= 12 {
				return kv.Value[:12]
			}
		}
	}
	return fmt.Sprintf("dev%d", time.Now().Unix())
}()

//go:embed templates
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

var templates = template.Must(template.ParseFS(templateFS, "templates/*.tmpl"))

type pageData struct {
	Matches       []matchView
	LastSyncedUTC string
	FeedHost      string
	HasVenues     bool
	AssetVersion  string
	Nav           navData
	FontFamily    string       // preview ?font= override: family name
	FontHref      template.URL // preview ?font= override: stylesheet URL
}

type matchView struct {
	HomeTeam    string
	AwayTeam    string
	HomeURL     string // team page link; empty for unnamed sides
	AwayURL     string
	HomeFlag    string
	AwayFlag    string
	KickoffUTC  string
	Venue       string
	StageLabel  string // table stage column: group name, or stage for knockouts
	StageURL    string // link to the group detail page; empty for knockouts
	StageName   string // card top-left: always the stage label ("Group stage")
	GroupName   string // card group pill: "Group A", empty for knockouts
	GroupURL    string // card group pill link
	Score       string // combined "2 – 1" for table views; empty until finished
	HomeGoals   string // card score: home goals, empty until finished
	AwayGoals   string // card score: away goals
	Pens        string // "4–2 pens" for a finished shootout, else empty
	StatusLabel string // "In play", "Postponed", "Cancelled" or empty
	Live        bool   // match in progress — shown with a LIVE badge
	AriaLabel   string // accessible name summarising the card for screen readers
}

var stateLabels = map[fixtures.Status]string{
	fixtures.StatusInPlay:    "In play",
	fixtures.StatusPostponed: "Postponed",
	fixtures.StatusCancelled: "Cancelled",
}

func page(store fixtures.Store, host string) http.HandlerFunc {
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

		data := pageData{FeedHost: host, AssetVersion: assetVersion, LastSyncedUTC: lastSynced(state), Nav: buildNav(matches, host)}
		data.Matches, data.HasVenues = buildViews(matches)
		if f, ok := fontBySlug(r.URL.Query().Get("font")); ok {
			data.FontFamily, data.FontHref = f.Name, f.Href // preview override
		}

		render(w, "index.html.tmpl", data)
	}
}

func lastSynced(state fixtures.SyncState) string {
	if state.LastSyncedAt == nil {
		return ""
	}
	return state.LastSyncedAt.UTC().Format(time.RFC3339)
}

func buildViews(matches []fixtures.Match) ([]matchView, bool) {
	views := make([]matchView, len(matches))
	hasVenues := false
	for i, m := range matches {
		views[i] = viewOf(m)
		if m.Venue != "" {
			hasVenues = true
		}
	}
	return views, hasVenues
}

func render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("rendering page", "template", name, "error", err)
	}
}

func nameOrTBC(team string) string {
	if team == "" {
		return "TBC"
	}
	return team
}

func viewOf(m fixtures.Match) matchView {
	v := matchView{
		HomeTeam:    nameOrTBC(m.HomeTeam),
		AwayTeam:    nameOrTBC(m.AwayTeam),
		HomeURL:     teamURL(m.HomeTeam),
		AwayURL:     teamURL(m.AwayTeam),
		HomeFlag:    flagFor(m.HomeTeam),
		AwayFlag:    flagFor(m.AwayTeam),
		KickoffUTC:  m.KickoffAt.UTC().Format(time.RFC3339),
		Venue:       m.Venue,
		StageLabel:  m.GroupName,
		StageName:   m.Stage.Label(),
		StatusLabel: stateLabels[m.Status],
		Live:        m.Status == fixtures.StatusInPlay,
	}
	if m.Stage != fixtures.StageGroup {
		v.StageLabel = m.Stage.Label()
	} else if letter, ok := strings.CutPrefix(m.GroupName, "Group "); ok {
		v.StageURL = "/groups/" + letter
		v.GroupName = m.GroupName
		v.GroupURL = "/groups/" + letter
	}
	if m.Status == fixtures.StatusFinished && m.HomeScore != nil && m.AwayScore != nil {
		v.Score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
		v.HomeGoals = fmt.Sprintf("%d", *m.HomeScore)
		v.AwayGoals = fmt.Sprintf("%d", *m.AwayScore)
		if m.HomePenalties != nil && m.AwayPenalties != nil {
			v.Pens = fmt.Sprintf("%d–%d pens", *m.HomePenalties, *m.AwayPenalties)
		}
	}
	v.AriaLabel = ariaLabel(m, v)
	return v
}

// ariaLabel is a concise accessible name for a match card, so screen
// reader users get a one-line summary instead of an unlabelled article.
func ariaLabel(m fixtures.Match, v matchView) string {
	middle := "versus"
	if v.Score != "" {
		middle = strings.ReplaceAll(v.Score, "–", "to") // "2 to 1"
	}
	label := fmt.Sprintf("%s %s %s", v.HomeTeam, middle, v.AwayTeam)
	if v.Pens != "" {
		label += ", " + strings.Replace(v.Pens, "–", " to ", 1)
	}
	context := m.Stage.Label()
	if v.GroupName != "" {
		context = v.GroupName
	}
	label += ", " + context
	switch {
	case v.Live:
		label += ", in progress"
	case m.Status == fixtures.StatusFinished:
		label += ", full time"
	case m.Status == fixtures.StatusPostponed:
		label += ", postponed"
	case m.Status == fixtures.StatusCancelled:
		label += ", cancelled"
	}
	return label
}
