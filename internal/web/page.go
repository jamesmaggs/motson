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
	Matches         []matchView
	LastSyncedUTC   string
	LastSyncedLabel string
	HasVenues       bool
	AssetVersion    string
	Nav             navData
}

// errorData backs the styled error page (404s and 500s).
type errorData struct {
	AssetVersion string
	Status       int
	Title        string
	Message      string
}

// utcLabel is a human-readable UTC time shown as the no-JS fallback for a
// <time> element; client JS replaces it with the visitor's local time.
func utcLabel(t time.Time) string {
	return t.UTC().Format("Mon 2 Jan, 15:04") + " UTC"
}

func syncedLabel(state fixtures.SyncState) string {
	if state.LastSyncedAt == nil {
		return ""
	}
	return utcLabel(*state.LastSyncedAt)
}

// renderError writes a styled error page with the given HTTP status.
func renderError(w http.ResponseWriter, status int, title, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	data := errorData{AssetVersion: assetVersion, Status: status, Title: title, Message: message}
	if err := templates.ExecuteTemplate(w, "error.html.tmpl", data); err != nil {
		slog.Error("rendering error page", "error", err)
	}
}

// errFixturesUnavailable renders the 500 shown when the store can't be read.
func errFixturesUnavailable(w http.ResponseWriter) {
	renderError(w, http.StatusInternalServerError, "Something went wrong",
		"We couldn't load the fixtures just now. Please try again in a moment.")
}

// loadFixtures reads the matches and sync state every HTML page needs,
// rendering the styled 500 and returning ok=false if the store fails.
func loadFixtures(w http.ResponseWriter, r *http.Request, store fixtures.Store) (matches []fixtures.Match, state fixtures.SyncState, ok bool) {
	matches, err := store.Matches(r.Context())
	if err != nil {
		errFixturesUnavailable(w)
		return nil, fixtures.SyncState{}, false
	}
	state, err = store.SyncState(r.Context())
	if err != nil {
		errFixturesUnavailable(w)
		return nil, fixtures.SyncState{}, false
	}
	return matches, state, true
}

type matchView struct {
	HomeTeam     string
	AwayTeam     string
	HomeURL      string // team page link; empty for unnamed sides
	AwayURL      string
	HomeFlag     string
	AwayFlag     string
	KickoffUTC   string // machine-readable ISO 8601 UTC (the <time> datetime)
	KickoffLabel string // human-readable UTC fallback shown until JS localises it
	Venue        string
	StageName    string // card top-left: always the stage label ("Group stage")
	GroupName    string // card group pill: "Group A", empty for knockouts
	GroupURL     string // card group pill link
	Score        string // "2 – 1" once finished; used to derive AriaLabel (not rendered)
	HomeGoals    string // card score: home goals, empty until finished
	AwayGoals    string // card score: away goals
	Pens         string // "4–2 pens" for a finished shootout, else empty
	StatusLabel  string // "In play", "Postponed", "Cancelled" or empty
	Live         bool   // match in progress — shown with a LIVE badge
	AriaLabel    string // accessible name summarising the card for screen readers
}

var stateLabels = map[fixtures.Status]string{
	fixtures.StatusInPlay:    "In play",
	fixtures.StatusPostponed: "Postponed",
	fixtures.StatusCancelled: "Cancelled",
}

func page(store fixtures.Store, host string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matches, state, ok := loadFixtures(w, r, store)
		if !ok {
			return
		}

		data := pageData{AssetVersion: assetVersion, LastSyncedUTC: lastSynced(state), LastSyncedLabel: syncedLabel(state), Nav: buildNav(matches, host)}
		data.Matches, data.HasVenues = buildViews(matches)

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
		HomeTeam:     nameOrTBC(m.HomeTeam),
		AwayTeam:     nameOrTBC(m.AwayTeam),
		HomeURL:      teamURL(m.HomeTeam),
		AwayURL:      teamURL(m.AwayTeam),
		HomeFlag:     flagFor(m.HomeTeam),
		AwayFlag:     flagFor(m.AwayTeam),
		KickoffUTC:   m.KickoffAt.UTC().Format(time.RFC3339),
		KickoffLabel: utcLabel(m.KickoffAt),
		Venue:        m.Venue,
		StageName:    m.Stage.Label(),
		StatusLabel:  stateLabels[m.Status],
		Live:         m.Status == fixtures.StatusInPlay,
	}
	// Group-stage cards carry a pill linking to the group page; knockout
	// cards (no group_name) carry none.
	if letter, ok := strings.CutPrefix(m.GroupName, "Group "); ok && m.Stage == fixtures.StageGroup {
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
