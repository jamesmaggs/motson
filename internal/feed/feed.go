// Package feed renders the CalendarFeed surface: an RFC 5545 iCalendar
// document with one event per match, stable UIDs, and scores in the
// title once a match has finished.
package feed

import (
	"fmt"
	"strings"

	ics "github.com/arran4/golang-ical"

	"github.com/Jazzatola/motson/internal/fixtures"
)

// Render serialises the matches as an iCalendar document. host anchors
// event UIDs (StableEventIdentity), e.g. "wc-1@motson.jamesmaggs.com".
func Render(host string, matches []fixtures.Match) (string, error) {
	cal := ics.NewCalendar()
	cal.SetProductId("-//Motson//World Cup 2026//EN")
	cal.SetMethod(ics.MethodPublish)
	cal.SetName("World Cup 2026")

	for _, m := range matches {
		e := cal.AddEvent(fmt.Sprintf("%s@%s", m.ProviderMatchID, host))
		e.SetSummary(summary(m))
		e.SetStartAt(m.KickoffAt.UTC())
		e.SetEndAt(m.EndsAt().UTC())
		e.SetLocation(m.Venue)
		if m.Status == fixtures.StatusCancelled {
			e.SetStatus(ics.ObjectStatusCancelled)
		}
	}
	return cal.Serialize(), nil
}

// summary is the event title: "Home vs Away" until the match finishes,
// then "Home 2-1 Away", with "(4-2 pens)" appended after a shootout.
func summary(m fixtures.Match) string {
	if m.Status != fixtures.StatusFinished || m.HomeScore == nil || m.AwayScore == nil {
		return fmt.Sprintf("%s vs %s", m.HomeTeam, m.AwayTeam)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s %d-%d %s", m.HomeTeam, *m.HomeScore, *m.AwayScore, m.AwayTeam)
	if m.HomePenalties != nil && m.AwayPenalties != nil {
		fmt.Fprintf(&b, " (%d-%d pens)", *m.HomePenalties, *m.AwayPenalties)
	}
	return b.String()
}
