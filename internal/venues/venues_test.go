package venues

import (
	"strings"
	"testing"

	"github.com/jamesmaggs/motson/internal/fixtures"
)

func TestEnrichFillsMissingVenue(t *testing.T) {
	// 537390 is the final — MetLife Stadium.
	got := Enrich([]fixtures.Match{{ProviderMatchID: "537390"}})
	if got[0].Venue != "MetLife Stadium, New York New Jersey" {
		t.Errorf("venue not filled from the table, got %q", got[0].Venue)
	}
}

func TestEnrichNeverOverridesProviderVenue(t *testing.T) {
	got := Enrich([]fixtures.Match{{ProviderMatchID: "537390", Venue: "Provider Stadium"}})
	if got[0].Venue != "Provider Stadium" {
		t.Errorf("provider-supplied venue must win, got %q", got[0].Venue)
	}
}

func TestEnrichLeavesUnknownMatchesUntouched(t *testing.T) {
	got := Enrich([]fixtures.Match{{ProviderMatchID: "does-not-exist"}})
	if got[0].Venue != "" {
		t.Errorf("unknown match should keep its empty venue, got %q", got[0].Venue)
	}
}

// The table covers the whole 104-match tournament and every entry names a
// real venue; this guards against an empty/typo'd value slipping in.
func TestVenueTableIsCompleteAndValid(t *testing.T) {
	if len(byID) != 104 {
		t.Errorf("expected 104 venue entries, got %d", len(byID))
	}
	known := map[string]bool{}
	for _, v := range []string{
		"Estadio Azteca, Mexico City", "Estadio Akron, Guadalajara",
		"Estadio BBVA, Monterrey", "BC Place, Vancouver", "BMO Field, Toronto",
		"Lumen Field, Seattle", "Levi's Stadium, San Francisco Bay Area",
		"SoFi Stadium, Los Angeles", "Arrowhead Stadium, Kansas City",
		"AT&T Stadium, Dallas", "NRG Stadium, Houston",
		"Mercedes-Benz Stadium, Atlanta", "Hard Rock Stadium, Miami",
		"Lincoln Financial Field, Philadelphia", "MetLife Stadium, New York New Jersey",
		"Gillette Stadium, Boston",
	} {
		known[v] = true
	}
	for id, v := range byID {
		if strings.TrimSpace(v) == "" {
			t.Errorf("match %s has an empty venue", id)
		}
		if !known[v] {
			t.Errorf("match %s has an unrecognised venue %q", id, v)
		}
	}
}
