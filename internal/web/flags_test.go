package web

import "testing"

// The flag map must key on the provider's exact team names.
func TestFlagForProviderNames(t *testing.T) {
	cases := map[string]string{
		"Cape Verde Islands": "🇨🇻",
		"Congo DR":           "🇨🇩",
		"Canada":             "🇨🇦",
		"Winner SF1":         "", // placeholder, no flag
	}
	for team, want := range cases {
		if got := flagFor(team); got != want {
			t.Errorf("flagFor(%q) = %q, want %q", team, got, want)
		}
	}
}
