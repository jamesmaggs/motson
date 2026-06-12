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
}

type matchView struct {
	HomeTeam   string
	AwayTeam   string
	KickoffUTC string
	Venue      string
	StageLabel string
	Score      string // empty until the match has finished
	StateLabel string // "In play", "Postponed", "Cancelled" or empty
}

var stageLabels = map[fixtures.Stage]string{
	fixtures.StageRoundOf32:    "Round of 32",
	fixtures.StageRoundOf16:    "Round of 16",
	fixtures.StageQuarterFinal: "Quarter-final",
	fixtures.StageSemiFinal:    "Semi-final",
	fixtures.StageThirdPlace:   "Third place",
	fixtures.StageFinal:        "Final",
}

var stateLabels = map[fixtures.Status]string{
	fixtures.StatusInPlay:    "In play",
	fixtures.StatusPostponed: "Postponed",
	fixtures.StatusCancelled: "Cancelled",
}

func page(store fixtures.Store) http.HandlerFunc {
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

		data := pageData{Matches: make([]matchView, len(matches))}
		if state.LastSyncedAt != nil {
			data.LastSyncedUTC = state.LastSyncedAt.UTC().Format(time.RFC3339)
		}
		for i, m := range matches {
			data.Matches[i] = viewOf(m)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := indexTemplate.Execute(w, data); err != nil {
			slog.Error("rendering page", "error", err)
		}
	}
}

func viewOf(m fixtures.Match) matchView {
	v := matchView{
		HomeTeam:   m.HomeTeam,
		AwayTeam:   m.AwayTeam,
		KickoffUTC: m.KickoffAt.UTC().Format(time.RFC3339),
		Venue:      m.Venue,
		StageLabel: m.GroupName,
		StateLabel: stateLabels[m.Status],
	}
	if m.Stage != fixtures.StageGroup {
		v.StageLabel = stageLabels[m.Stage]
	}
	if m.Status == fixtures.StatusFinished && m.HomeScore != nil && m.AwayScore != nil {
		v.Score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
		if m.HomePenalties != nil && m.AwayPenalties != nil {
			v.StateLabel = fmt.Sprintf("%d–%d pens", *m.HomePenalties, *m.AwayPenalties)
		}
	}
	return v
}
