// Package feed renders the CalendarFeed surface: an RFC 5545 iCalendar
// document with one event per match, stable UIDs, and scores in the
// title once a match has finished.
package feed

import (
	"fmt"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

// Render serialises the matches as an iCalendar document. host anchors
// event UIDs (StableEventIdentity), e.g. "wc-1@motson.jamesmaggs.com".
func Render(host string, matches []fixtures.Match) (string, error) {
	cal := ics.NewCalendar()
	cal.SetProductId("-//Motson//World Cup 2026//EN")
	cal.SetName("World Cup 2026")
	// Subscribed clients re-poll on this hint, matching sync_interval.
	refresh := iso8601Duration(fixtures.SyncInterval)
	cal.SetRefreshInterval(refresh)
	cal.SetXPublishedTTL(refresh)

	for _, m := range matches {
		e := cal.AddEvent(fmt.Sprintf("%s@%s", m.ProviderMatchID, host))
		e.SetSummary(summary(m))
		e.SetStartAt(m.KickoffAt.UTC())
		e.SetEndAt(m.EndsAt().UTC())
		if m.Venue != "" {
			e.SetLocation(m.Venue)
		}
		e.SetStatus(eventStatus(m.Status))
		e.SetDescription(statusLabels[m.Status])
	}
	return cal.Serialize(), nil
}

// iso8601Duration formats a sync interval as an RFC 5545 duration:
// whole hours as "PT{n}H", otherwise minutes as "PT{n}M".
func iso8601Duration(d time.Duration) string {
	if d%time.Hour == 0 {
		return fmt.Sprintf("PT%dH", int(d/time.Hour))
	}
	return fmt.Sprintf("PT%dM", int(d/time.Minute))
}

func nameOrTBC(team string) string {
	if team == "" {
		return "TBC"
	}
	return team
}

// statusLabels expose the spec's provider_status on every event.
var statusLabels = map[fixtures.Status]string{
	fixtures.StatusScheduled: "Scheduled",
	fixtures.StatusInPlay:    "In play",
	fixtures.StatusFinished:  "Finished",
	fixtures.StatusPostponed: "Postponed",
	fixtures.StatusCancelled: "Cancelled",
}

// eventStatus maps provider_status onto iCalendar's event statuses:
// postponed is tentative, cancelled is cancelled, the rest confirmed.
func eventStatus(s fixtures.Status) ics.ObjectStatus {
	switch s {
	case fixtures.StatusCancelled:
		return ics.ObjectStatusCancelled
	case fixtures.StatusPostponed:
		return ics.ObjectStatusTentative
	default:
		return ics.ObjectStatusConfirmed
	}
}

// summary is the event title: "Home vs Away" until the match finishes,
// then "Home 2-1 Away", with "(4-2 pens)" appended after a shootout.
// A fixture with no named teams is titled with its stage; a single
// named team appears alongside "TBC" (UndeterminedFixtures).
func summary(m fixtures.Match) string {
	if m.HomeTeam == "" && m.AwayTeam == "" {
		return m.Stage.Label()
	}
	if m.Status != fixtures.StatusFinished || m.HomeScore == nil || m.AwayScore == nil {
		return fmt.Sprintf("%s vs %s", nameOrTBC(m.HomeTeam), nameOrTBC(m.AwayTeam))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s %d-%d %s", m.HomeTeam, *m.HomeScore, *m.AwayScore, m.AwayTeam)
	if m.HomePenalties != nil && m.AwayPenalties != nil {
		fmt.Fprintf(&b, " (%d-%d pens)", *m.HomePenalties, *m.AwayPenalties)
	}
	return b.String()
}
