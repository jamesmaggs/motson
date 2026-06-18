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
	Featured        *matchView // the spotlighted live/next match, nil when none
	Days            []dayGroup
	LastSyncedUTC   string
	LastSyncedLabel string
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

func page(store fixtures.Store, host string, clock func() time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matches, state, ok := loadFixtures(w, r, store)
		if !ok {
			return
		}

		data := pageData{AssetVersion: assetVersion, LastSyncedUTC: lastSynced(state), LastSyncedLabel: syncedLabel(state), Nav: buildNav(matches, host)}
		// Spotlight the live or next match above the list; it also stays in
		// the kickoff-ordered list below (the FeaturedMatch guarantee).
		if fm, ok := featuredMatch(matches, clock()); ok {
			v := viewOf(fm)
			data.Featured = &v
		}
		data.Days = groupByDay(matches)

		render(w, "index.html.tmpl", data)
	}
}

// featuredMatch is the match the index spotlights above the list: the one
// in progress, or — when none is — the next still to kick off (the
// scheduled match with the earliest kickoff at or after now). The second
// return is false when neither exists.
func featuredMatch(matches []fixtures.Match, now time.Time) (fixtures.Match, bool) {
	var best fixtures.Match
	found := false
	// A match in progress is "current": prefer it, earliest kickoff first.
	for _, m := range matches {
		if m.Status == fixtures.StatusInPlay && (!found || m.KickoffAt.Before(best.KickoffAt)) {
			best, found = m, true
		}
	}
	if found {
		return best, true
	}
	// Otherwise the soonest scheduled match still to kick off.
	for _, m := range matches {
		if m.Status == fixtures.StatusScheduled && !m.KickoffAt.Before(now) && (!found || m.KickoffAt.Before(best.KickoffAt)) {
			best, found = m, true
		}
	}
	return best, found
}

func lastSynced(state fixtures.SyncState) string {
	if state.LastSyncedAt == nil {
		return ""
	}
	return state.LastSyncedAt.UTC().Format(time.RFC3339)
}

// dayGroup is one calendar day's fixtures, fronted by a date heading.
// Matches are grouped by their UTC date (the pages are server-rendered);
// the heading localises to the viewer's date via the same client script
// that localises kickoff times (ADR 0010), with a UTC date as the no-JS
// fallback.
type dayGroup struct {
	DateUTC   string // representative instant (noon UTC) for the <time> heading
	DateLabel string // human UTC fallback shown until the client localises it
	Matches   []matchView
}

// groupByDay splits kickoff-ordered matches into consecutive day groups,
// preserving order within and across days.
func groupByDay(matches []fixtures.Match) []dayGroup {
	var days []dayGroup
	lastKey := ""
	for _, m := range matches {
		y, mo, d := m.KickoffAt.UTC().Date()
		key := fmt.Sprintf("%04d-%02d-%02d", y, mo, d)
		if key != lastKey {
			// Noon UTC is a stable representative: it localises to the same
			// calendar date across every realistic timezone offset.
			noon := time.Date(y, mo, d, 12, 0, 0, 0, time.UTC)
			days = append(days, dayGroup{
				DateUTC:   noon.Format(time.RFC3339),
				DateLabel: noon.Format("Monday 2 January"),
			})
			lastKey = key
		}
		days[len(days)-1].Matches = append(days[len(days)-1].Matches, viewOf(m))
	}
	return days
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
