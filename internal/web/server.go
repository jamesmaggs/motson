// Package web serves Motson's HTTP boundary: the fixtures page, the
// calendar feed and the staleness-aware health check.
package web

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jamesmaggs/motson/internal/feed"
	"github.com/jamesmaggs/motson/internal/fixtures"
)

// NewHandler routes the public surfaces. host anchors feed UIDs; clock
// supplies "now" so staleness is testable.
func NewHandler(store fixtures.Store, host string, clock func() time.Time) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", page(store, host))
	mux.HandleFunc("GET /healthz", healthz(store, clock))
	mux.HandleFunc("GET /calendar.ics", calendar(store, host))
	mux.Handle("GET /static/", http.FileServerFS(staticFS))
	return mux
}

func healthz(store fixtures.Store, clock func() time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := store.SyncState(r.Context())
		if err != nil {
			http.Error(w, "sync state unavailable", http.StatusServiceUnavailable)
			return
		}
		lastSynced := "never"
		if state.LastSyncedAt != nil {
			lastSynced = state.LastSyncedAt.UTC().Format(time.RFC3339)
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if state.IsStale(clock()) {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "stale: last synced %s\n", lastSynced)
			return
		}
		fmt.Fprintf(w, "ok: last synced %s\n", lastSynced)
	}
}

func calendar(store fixtures.Store, host string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matches, err := store.Matches(r.Context())
		if err != nil {
			http.Error(w, "fixtures unavailable", http.StatusInternalServerError)
			return
		}
		body, err := feed.Render(host, matches)
		if err != nil {
			slog.Error("rendering feed", "error", err)
			http.Error(w, "feed unavailable", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		fmt.Fprint(w, body)
	}
}
