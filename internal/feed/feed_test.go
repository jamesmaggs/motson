package feed_test

import (
	"strings"
	"testing"
	"time"

	ics "github.com/arran4/golang-ical"

	"github.com/jamesmaggs/motson/internal/feed"
	"github.com/jamesmaggs/motson/internal/fixtures"
)

const host = "motson.jamesmaggs.com"

func intp(i int) *int { return &i }

func kickoff() time.Time { return time.Date(2026, 6, 13, 18, 0, 0, 0, time.UTC) }

func scheduled() fixtures.Match {
	return fixtures.Match{
		ProviderMatchID: "wc-1",
		HomeTeam:        "Canada",
		AwayTeam:        "Mexico",
		KickoffAt:       kickoff(),
		Venue:           "Estadio Azteca, Mexico City",
		Stage:           fixtures.StageGroup,
		GroupName:       "Group A",
		Status:          fixtures.StatusScheduled,
	}
}

func render(t *testing.T, matches ...fixtures.Match) *ics.Calendar {
	t.Helper()
	out, err := feed.Render(host, matches)
	if err != nil {
		t.Fatal(err)
	}
	cal, err := ics.ParseCalendar(strings.NewReader(out))
	if err != nil {
		t.Fatalf("rendered feed does not parse as iCalendar: %v", err)
	}
	return cal
}

func event(t *testing.T, cal *ics.Calendar) *ics.VEvent {
	t.Helper()
	events := cal.Events()
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	return events[0]
}

func prop(t *testing.T, e *ics.VEvent, p ics.ComponentProperty) string {
	t.Helper()
	if ianaProp := e.GetProperty(p); ianaProp != nil {
		return ianaProp.Value
	}
	return ""
}

// Obligation: surface-exposure.CalendarFeed — every match becomes an
// event exposing teams, kickoff, end time and venue.
func TestFeedExposesMatchAsEvent(t *testing.T) {
	e := event(t, render(t, scheduled()))

	if got, want := prop(t, e, ics.ComponentPropertySummary), "Canada vs Mexico"; got != want {
		t.Errorf("SUMMARY = %q, want %q", got, want)
	}
	start, err := e.GetStartAt()
	if err != nil || !start.Equal(kickoff()) {
		t.Errorf("DTSTART = %v (%v), want %v", start, err, kickoff())
	}
	end, err := e.GetEndAt()
	if err != nil || !end.Equal(kickoff().Add(2*time.Hour)) {
		t.Errorf("DTEND = %v (%v), want %v (group match: +2h)", end, err, kickoff().Add(2*time.Hour))
	}
	if got, want := prop(t, e, ics.ComponentPropertyLocation), "Estadio Azteca, Mexico City"; got != want {
		t.Errorf("LOCATION = %q, want %q", got, want)
	}
}

// Guarantee: StableEventIdentity — UID is keyed on provider_match_id
// and identical across renders.
func TestFeedEventIdentityIsStable(t *testing.T) {
	first := prop(t, event(t, render(t, scheduled())), ics.ComponentPropertyUniqueId)
	second := prop(t, event(t, render(t, scheduled())), ics.ComponentPropertyUniqueId)

	if first == "" {
		t.Fatal("event has no UID")
	}
	if first != second {
		t.Errorf("UID changed across renders: %q vs %q", first, second)
	}
	if !strings.Contains(first, "wc-1") {
		t.Errorf("UID %q not keyed on provider match id", first)
	}
}

// Scores appear in the title only once a match is finished.
func TestFinishedMatchShowsScoreInTitle(t *testing.T) {
	m := scheduled()
	m.Status = fixtures.StatusFinished
	m.HomeScore, m.AwayScore = intp(2), intp(1)

	e := event(t, render(t, m))
	if got, want := prop(t, e, ics.ComponentPropertySummary), "Canada 2-1 Mexico"; got != want {
		t.Errorf("SUMMARY = %q, want %q", got, want)
	}
}

func TestShootoutResultShownInTitle(t *testing.T) {
	m := scheduled()
	m.Stage, m.GroupName = fixtures.StageFinal, ""
	m.Status = fixtures.StatusFinished
	m.HomeScore, m.AwayScore = intp(3), intp(3)
	m.HomePenalties, m.AwayPenalties = intp(4), intp(2)

	e := event(t, render(t, m))
	if got, want := prop(t, e, ics.ComponentPropertySummary), "Canada 3-3 Mexico (4-2 pens)"; got != want {
		t.Errorf("SUMMARY = %q, want %q", got, want)
	}
	end, err := e.GetEndAt()
	if err != nil || !end.Equal(kickoff().Add(2*time.Hour+45*time.Minute)) {
		t.Errorf("DTEND = %v (%v), want kickoff+2h45 for knockout", end, err)
	}
}

// No in-play scores: the spec shows scores only when finished.
func TestInPlayMatchShowsNoScore(t *testing.T) {
	m := scheduled()
	m.Status = fixtures.StatusInPlay
	m.HomeScore, m.AwayScore = nil, nil

	e := event(t, render(t, m))
	if got, want := prop(t, e, ics.ComponentPropertySummary), "Canada vs Mexico"; got != want {
		t.Errorf("SUMMARY = %q, want %q", got, want)
	}
}

// Cancelled matches are marked cancelled rather than removed.
func TestCancelledMatchMarkedCancelled(t *testing.T) {
	m := scheduled()
	m.Status = fixtures.StatusCancelled

	e := event(t, render(t, m))
	if got := prop(t, e, ics.ComponentPropertyStatus); got != "CANCELLED" {
		t.Errorf("STATUS = %q, want CANCELLED", got)
	}
}

// Postponed matches are tentative, everything else confirmed: the
// closest iCalendar rendering of the spec's provider_status exposure.
func TestEventStatusReflectsMatchStatus(t *testing.T) {
	postponed := scheduled()
	postponed.Status = fixtures.StatusPostponed
	e := event(t, render(t, postponed))
	if got := prop(t, e, ics.ComponentPropertyStatus); got != "TENTATIVE" {
		t.Errorf("postponed: STATUS = %q, want TENTATIVE", got)
	}

	e = event(t, render(t, scheduled()))
	if got := prop(t, e, ics.ComponentPropertyStatus); got != "CONFIRMED" {
		t.Errorf("scheduled: STATUS = %q, want CONFIRMED", got)
	}
}

// Obligation: surface-exposure.CalendarFeed — provider_status is
// exposed on every event via its description.
func TestEventDescriptionExposesStatus(t *testing.T) {
	cases := map[fixtures.Status]string{
		fixtures.StatusScheduled: "Scheduled",
		fixtures.StatusInPlay:    "In play",
		fixtures.StatusFinished:  "Finished",
		fixtures.StatusPostponed: "Postponed",
		fixtures.StatusCancelled: "Cancelled",
	}
	for status, want := range cases {
		m := scheduled()
		m.Status = status
		e := event(t, render(t, m))
		if got := prop(t, e, ics.ComponentPropertyDescription); !strings.Contains(got, want) {
			t.Errorf("status %s: DESCRIPTION = %q, want it to contain %q", status, got, want)
		}
	}
}

// Subscribed clients should re-poll hourly, matching sync_interval.
func TestFeedAdvertisesHourlyRefresh(t *testing.T) {
	out, err := feed.Render(host, []fixtures.Match{scheduled()})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "REFRESH-INTERVAL;VALUE=DURATION:PT1H") {
		t.Error("feed missing REFRESH-INTERVAL:PT1H")
	}
	if !strings.Contains(out, "X-PUBLISHED-TTL:PT1H") {
		t.Error("feed missing X-PUBLISHED-TTL:PT1H")
	}
}

func TestFeedRendersOneEventPerMatch(t *testing.T) {
	a, b := scheduled(), scheduled()
	b.ProviderMatchID = "wc-2"
	b.HomeTeam, b.AwayTeam = "Spain", "France"

	cal := render(t, a, b)
	if got := len(cal.Events()); got != 2 {
		t.Errorf("got %d events, want 2", got)
	}
}
