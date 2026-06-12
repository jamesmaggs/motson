package web_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

// Obligation: surface-exposure.FixturesPage — every match exposes
// stage, group, teams, kickoff and venue.
func TestPageListsFixtures(t *testing.T) {
	rec := get(t, seeded(t, match("wc-1")), now, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body := rec.Body.String()
	for _, want := range []string{"Canada", "Mexico", "Estadio Azteca, Mexico City", "Group A"} {
		if !strings.Contains(body, want) {
			t.Errorf("page missing %q", want)
		}
	}
}

// Kickoff times are emitted as machine-readable UTC for browser-local
// conversion (ADR 0010).
func TestPageEmitsKickoffAsUTCDatetime(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()
	if !strings.Contains(body, `datetime="2026-06-13T18:00:00Z"`) {
		t.Errorf("page missing UTC datetime attribute for kickoff: %s", body)
	}
}

// Scores appear only once a match is finished; an in-play match shows
// "vs" and an "In play" label, not a score.
func TestPageShowsScoreOnlyWhenFinished(t *testing.T) {
	finished := match("wc-1")
	finished.Status = fixtures.StatusFinished
	finished.HomeScore, finished.AwayScore = intp(2), intp(1)

	inPlay := withID(match("wc-2"), "wc-2")
	inPlay.HomeTeam, inPlay.AwayTeam = "Spain", "France"
	inPlay.Status = fixtures.StatusInPlay

	body := get(t, seeded(t, finished, inPlay), now, "/").Body.String()

	if !strings.Contains(body, `2<span class="sep">:</span>1`) {
		t.Errorf("finished match score missing from page: %s", body)
	}
	if got := strings.Count(body, `class="score"`); got != 1 {
		t.Errorf("got %d score blocks, want 1 (only the finished match)", got)
	}
}

// In-progress matches are signalled by a green glow on the card (the
// "live" class), not by an "In play" label.
func TestInProgressCardGlowsAndDropsLabel(t *testing.T) {
	inPlay := match("wc-1")
	inPlay.Status = fixtures.StatusInPlay

	body := get(t, seeded(t, inPlay), now, "/").Body.String()

	if !strings.Contains(body, `class="card live"`) {
		t.Errorf("in-play card not marked live: %s", body)
	}
	if strings.Contains(body, "In play") {
		t.Errorf("in-play card should not carry an 'In play' label")
	}
	if !strings.Contains(body, ".card.live") {
		t.Errorf("stylesheet missing the live-card glow rule")
	}
}

func TestPageShowsShootoutResult(t *testing.T) {
	m := match("wc-1")
	m.Stage, m.GroupName = fixtures.StageFinal, ""
	m.Status = fixtures.StatusFinished
	m.HomeScore, m.AwayScore = intp(3), intp(3)
	m.HomePenalties, m.AwayPenalties = intp(4), intp(2)

	body := get(t, seeded(t, m), now, "/").Body.String()
	if !strings.Contains(body, "4–2 pens") {
		t.Errorf("shootout result missing from page: %s", body)
	}
}

// Obligation: surface-exposure.FixturesPage — exposes last_synced_at.
func TestPageShowsLastSyncedTime(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()
	if !strings.Contains(body, `datetime="2026-06-13T12:00:00Z"`) {
		t.Errorf("page missing last-synced timestamp: %s", body)
	}
}

// Spec: venue exposed only when present. The card shows it when the
// match has one and omits it entirely otherwise.
func TestVenueOmittedWhenAbsent(t *testing.T) {
	noVenue := match("wc-1")
	noVenue.Venue = ""

	body := get(t, seeded(t, noVenue), now, "/").Body.String()
	if strings.Contains(body, `class="venue"`) {
		t.Errorf("venue shown despite no venue data: %s", body)
	}
}

func TestVenueShownWhenPresent(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()
	if !strings.Contains(body, `class="venue"`) || !strings.Contains(body, "Estadio Azteca, Mexico City") {
		t.Errorf("venue missing despite venue data: %s", body)
	}
}

// A football leads the page title, and the favicon is the same
// emoji (served as an inline SVG so no icon asset is needed).
func TestPageHasFootballTitleAndFavicon(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()

	if !strings.Contains(body, "<h1>⚽ World Cup 2026</h1>") {
		t.Errorf("title missing leading football emoji: %s", body)
	}
	if !strings.Contains(body, `rel="icon"`) || !strings.Contains(body, "image/svg+xml") {
		t.Errorf("favicon link missing: %s", body)
	}
}

// Static asset URLs carry a per-build version so edge caches (e.g.
// Cloudflare ahead of the custom domain) can't serve a previous
// build's CSS or fonts after a deploy.
func TestStaticAssetLinksAreBuildVersioned(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()

	for _, asset := range []string{"pico.min.css?v=", "fonts.css?v="} {
		if !strings.Contains(body, asset) {
			t.Errorf("asset link missing build version: want %q in page", asset)
		}
	}
}

// Guarantee: UndeterminedFixtures — an unnamed side renders as "TBC"
// with no flag.
func TestUnnamedTeamsRenderAsTBC(t *testing.T) {
	m := match("wc-r32")
	m.Stage, m.GroupName = fixtures.StageRoundOf32, ""
	m.HomeTeam, m.AwayTeam = "", ""

	body := get(t, seeded(t, m), now, "/").Body.String()
	if got := strings.Count(body, "TBC"); got != 2 {
		t.Errorf("got %d TBC placeholders, want 2: %s", got, body)
	}
	for _, r := range body {
		if r >= 0x1F1E6 && r <= 0x1F1FF {
			t.Error("TBC side must not carry a flag")
			break
		}
	}
}

// Guarantee: TeamFlags — each side shows its flag with its (linked)
// name within the card.
func TestCardShowsFlagAndLinkedName(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()

	for _, want := range []string{
		`<span class="flag">🇨🇦</span>`, `<span class="name"><a href="/teams/canada">Canada</a></span>`,
		`<span class="flag">🇲🇽</span>`, `<span class="name"><a href="/teams/mexico">Mexico</a></span>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("card missing %q: %s", want, body)
		}
	}
}

// Each match is a card: stage + kickoff top-left, a linked group pill
// top-right, the score flanked by the two sides.
func TestPageRendersMatchesAsCards(t *testing.T) {
	finished := match("wc-1")
	finished.Status = fixtures.StatusFinished
	finished.HomeScore, finished.AwayScore = intp(2), intp(1)
	upcoming := withID(match("wc-2"), "wc-2")
	upcoming.HomeTeam, upcoming.AwayTeam = "Spain", "France"

	body := get(t, seeded(t, finished, upcoming), now, "/").Body.String()

	if got := strings.Count(body, `class="card"`); got != 2 {
		t.Errorf("got %d cards, want one per match (2): %s", got, body)
	}
	if got := strings.Count(body, `class="scoreline"`); got != 2 {
		t.Errorf("got %d scorelines, want 2", got)
	}
	if !strings.Contains(body, `<span class="stage">Group stage</span>`) {
		t.Errorf("card missing stage label: %s", body)
	}
}

// Guarantee: TeamFlags — placeholder names render without a flag
// rather than with a wrong or broken one.
func TestPlaceholderTeamsRenderWithoutFlags(t *testing.T) {
	m := match("wc-final")
	m.Stage, m.GroupName = fixtures.StageFinal, ""
	m.HomeTeam, m.AwayTeam = "Winner SF1", "Winner SF2"

	body := get(t, seeded(t, m), now, "/").Body.String()
	if !strings.Contains(body, "Winner SF1") || !strings.Contains(body, "Winner SF2") {
		t.Fatalf("placeholder names missing: %s", body)
	}
	for _, r := range body {
		if r >= 0x1F1E6 && r <= 0x1F1FF { // regional indicator symbols
			t.Errorf("placeholder page contains a flag: %s", body)
			break
		}
	}
}

// The subscribe link must use the webcal scheme so calendar clients
// subscribe (and auto-update) rather than import a one-off copy.
func TestPageSubscribeLinkUsesWebcalScheme(t *testing.T) {
	body := get(t, seeded(t, match("wc-1")), now, "/").Body.String()
	if !strings.Contains(body, `href="webcal://motson.jamesmaggs.com/calendar.ics"`) {
		t.Errorf("page missing webcal subscribe link: %s", body)
	}
}

func TestPageMarksCancelledAndPostponedMatches(t *testing.T) {
	cancelled := match("wc-1")
	cancelled.Status = fixtures.StatusCancelled
	postponed := withID(match("wc-2"), "wc-2")
	postponed.Status = fixtures.StatusPostponed

	body := get(t, seeded(t, cancelled, postponed), now, "/").Body.String()
	if !strings.Contains(body, "Cancelled") {
		t.Errorf("cancelled match not marked: %s", body)
	}
	if !strings.Contains(body, "Postponed") {
		t.Errorf("postponed match not marked: %s", body)
	}
}
