package web

// teamRank maps provider team names (including the same variants the flag
// table uses) to the team's FIFA men's world ranking. This is a fixed
// pre-tournament snapshot — the official FIFA/Coca-Cola ranking of
// 11 June 2026 — and is deliberately NOT updated during the tournament; it
// only orders the navigation menu (see the MenuOrder guarantee). Names not
// present here — knockout placeholders, or any nation missing from the
// snapshot — have no rank and sort after the ranked teams (rankOf reports
// ok=false).
var teamRank = map[string]int{
	"Argentina": 1, "Spain": 2, "France": 3, "England": 4, "Portugal": 5,
	"Brazil": 6, "Morocco": 7, "Netherlands": 8, "Belgium": 9, "Germany": 10,
	"Croatia": 11, "Italy": 12, "Colombia": 13, "Mexico": 14, "Senegal": 15,
	"Uruguay": 16,
	"USA":     17, "United States": 17,
	"Japan":       18,
	"Switzerland": 19,
	"Iran":        20, "IR Iran": 20,
	"Denmark": 21,
	"Türkiye": 22, "Turkey": 22,
	"Ecuador": 23, "Austria": 24,
	"South Korea": 25, "Korea Republic": 25,
	"Nigeria": 26, "Australia": 27, "Algeria": 28, "Egypt": 29, "Canada": 30,
	"Norway": 31, "Ukraine": 32,
	"Ivory Coast": 33, "Côte d'Ivoire": 33,
	"Panama": 34, "Poland": 36, "Wales": 37, "Sweden": 38, "Hungary": 39,
	"Czechia": 40, "Czech Republic": 40,
	"Paraguay": 41, "Scotland": 42, "Serbia": 43, "Cameroon": 44, "Tunisia": 45,
	"DR Congo": 46, "Congo DR": 46,
	"Slovakia": 47, "Greece": 48, "Venezuela": 49, "Uzbekistan": 50,
	"Chile": 51, "Peru": 52, "Costa Rica": 53, "Romania": 54, "Mali": 55,
	"Qatar": 56, "Iraq": 57, "Slovenia": 59, "South Africa": 60,
	"Saudi Arabia": 61, "Burkina Faso": 62, "Jordan": 63,
	"Bosnia-Herzegovina": 64, "Bosnia and Herzegovina": 64,
	"Honduras": 65, "Albania": 66,
	"Cape Verde": 67, "Cape Verde Islands": 67,
	"United Arab Emirates": 68, "North Macedonia": 69,
	"Jamaica": 71, "Georgia": 72, "Ghana": 73, "Iceland": 74, "Finland": 75,
	"Bolivia": 77, "Montenegro": 80,
	"Curaçao": 82, "Haiti": 83, "New Zealand": 85, "China": 91,
	"Guatemala": 97, "El Salvador": 100,
	"Trinidad and Tobago": 102,
}

// rankOf returns the team's FIFA ranking and whether one is known.
func rankOf(team string) (int, bool) {
	r, ok := teamRank[team]
	return r, ok
}
