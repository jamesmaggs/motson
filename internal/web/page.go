package web

import (
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

//go:embed templates/index.html.tmpl
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

var indexTemplate = template.Must(template.ParseFS(templateFS, "templates/index.html.tmpl"))

type pageData struct {
	Matches       []matchView
	LastSyncedUTC string
	FeedHost      string
	HasVenues     bool
}

type matchView struct {
	HomeTeam   string
	AwayTeam   string
	HomeFlag   string
	AwayFlag   string
	KickoffUTC string
	Venue      string
	StageLabel string
	Score      string // empty until the match has finished
	StateLabel string // "In play", "Postponed", "Cancelled" or empty
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

		data := pageData{Matches: make([]matchView, len(matches)), FeedHost: host}
		if state.LastSyncedAt != nil {
			data.LastSyncedUTC = state.LastSyncedAt.UTC().Format(time.RFC3339)
		}
		for i, m := range matches {
			data.Matches[i] = viewOf(m)
			if m.Venue != "" {
				data.HasVenues = true
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := indexTemplate.Execute(w, data); err != nil {
			slog.Error("rendering page", "error", err)
		}
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
		HomeTeam:   nameOrTBC(m.HomeTeam),
		AwayTeam:   nameOrTBC(m.AwayTeam),
		HomeFlag:   flagFor(m.HomeTeam),
		AwayFlag:   flagFor(m.AwayTeam),
		KickoffUTC: m.KickoffAt.UTC().Format(time.RFC3339),
		Venue:      m.Venue,
		StageLabel: m.GroupName,
		StateLabel: stateLabels[m.Status],
	}
	if m.Stage != fixtures.StageGroup {
		v.StageLabel = m.Stage.Label()
	}
	if m.Status == fixtures.StatusFinished && m.HomeScore != nil && m.AwayScore != nil {
		v.Score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
		if m.HomePenalties != nil && m.AwayPenalties != nil {
			v.StateLabel = fmt.Sprintf("%d–%d pens", *m.HomePenalties, *m.AwayPenalties)
		}
	}
	return v
}
