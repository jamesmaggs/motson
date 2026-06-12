package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ics "github.com/arran4/golang-ical"

	"github.com/Jazzatola/motson/internal/fixtures"
	"github.com/Jazzatola/motson/internal/store"
	"github.com/Jazzatola/motson/internal/web"
)

var now = time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)

func intp(i int) *int { return &i }

func match(id string) fixtures.Match {
	return fixtures.Match{
		ProviderMatchID: id,
		HomeTeam:        "Canada",
		AwayTeam:        "Mexico",
		KickoffAt:       time.Date(2026, 6, 13, 18, 0, 0, 0, time.UTC),
		Venue:           "Estadio Azteca, Mexico City",
		Stage:           fixtures.StageGroup,
		GroupName:       "Group A",
		Status:          fixtures.StatusScheduled,
	}
}

func get(t *testing.T, st fixtures.Store, at time.Time, path string) *httptest.ResponseRecorder {
	t.Helper()
	h := web.NewHandler(st, "motson.jamesmaggs.com", func() time.Time { return at })
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func seeded(t *testing.T, matches ...fixtures.Match) fixtures.Store {
	t.Helper()
	st := store.NewMemory()
	if err := st.ReplaceAll(context.Background(), matches, now); err != nil {
		t.Fatal(err)
	}
	return st
}

// Guarantee: StalenessIsUnhealthy — healthy while fresh.
func TestHealthzHealthyWhenFresh(t *testing.T) {
	rec := get(t, seeded(t, match("wc-1")), now.Add(time.Hour), "/healthz")
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

// Guarantee: StalenessIsUnhealthy — unhealthy once the last sync is
// older than staleness_threshold, and when no sync has ever happened.
func TestHealthzUnhealthyWhenStale(t *testing.T) {
	rec := get(t, seeded(t, match("wc-1")), now.Add(3*time.Hour), "/healthz")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("stale: status = %d, want 503", rec.Code)
	}

	rec = get(t, store.NewMemory(), now, "/healthz")
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("never synced: status = %d, want 503", rec.Code)
	}
}

// Obligation: surface-exposure.HealthCheck — exposes last_synced_at.
func TestHealthzReportsLastSyncedTime(t *testing.T) {
	rec := get(t, seeded(t, match("wc-1")), now.Add(time.Hour), "/healthz")
	if !strings.Contains(rec.Body.String(), now.Format(time.RFC3339)) {
		t.Errorf("body %q does not report last synced time %s", rec.Body.String(), now.Format(time.RFC3339))
	}
}

// The feed endpoint serves the stored matches as iCalendar with the
// content type Apple Calendar expects.
func TestCalendarEndpointServesFeed(t *testing.T) {
	rec := get(t, seeded(t, match("wc-1"), withID(match("wc-2"), "wc-2")), now, "/calendar.ics")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/calendar") {
		t.Errorf("Content-Type = %q, want text/calendar", ct)
	}
	cal, err := ics.ParseCalendar(strings.NewReader(rec.Body.String()))
	if err != nil {
		t.Fatalf("body does not parse as iCalendar: %v", err)
	}
	if got := len(cal.Events()); got != 2 {
		t.Errorf("got %d events, want 2", got)
	}
}

func withID(m fixtures.Match, id string) fixtures.Match {
	m.ProviderMatchID = id
	return m
}
