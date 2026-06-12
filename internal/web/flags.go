package web

// teamISO maps provider team names (including common variants) to
// ISO 3166-1 alpha-2 codes, from which flag emoji are derived. Names
// not present — knockout placeholders like "Winner Group A" — get no
// flag (TeamFlags guarantee).
var teamISO = map[string]string{
	"Canada": "CA", "Mexico": "MX",
	"USA": "US", "United States": "US",
	"Argentina": "AR", "Brazil": "BR", "Uruguay": "UY", "Colombia": "CO",
	"Ecuador": "EC", "Paraguay": "PY", "Chile": "CL", "Peru": "PE",
	"Bolivia": "BO", "Venezuela": "VE",
	"France": "FR", "Germany": "DE", "Spain": "ES", "Portugal": "PT",
	"Italy": "IT", "Netherlands": "NL", "Belgium": "BE", "Croatia": "HR",
	"Switzerland": "CH", "Austria": "AT", "Poland": "PL",
	"Czechia": "CZ", "Czech Republic": "CZ",
	"Slovakia": "SK", "Slovenia": "SI", "Serbia": "RS",
	"Bosnia-Herzegovina": "BA", "Bosnia and Herzegovina": "BA",
	"Albania": "AL", "North Macedonia": "MK", "Montenegro": "ME",
	"Norway": "NO", "Sweden": "SE", "Denmark": "DK", "Finland": "FI",
	"Iceland": "IS", "Ukraine": "UA", "Romania": "RO", "Hungary": "HU",
	"Greece": "GR", "Türkiye": "TR", "Turkey": "TR", "Georgia": "GE",
	"Morocco": "MA", "Senegal": "SN", "Tunisia": "TN", "Algeria": "DZ",
	"Egypt": "EG", "Ghana": "GH", "Nigeria": "NG", "Cameroon": "CM",
	"Ivory Coast": "CI", "Côte d'Ivoire": "CI",
	"South Africa": "ZA", "Cape Verde": "CV", "Cape Verde Islands": "CV", "Mali": "ML",
	"Burkina Faso": "BF", "DR Congo": "CD", "Congo DR": "CD",
	"Japan": "JP", "South Korea": "KR", "Korea Republic": "KR",
	"Australia": "AU", "Saudi Arabia": "SA", "Qatar": "QA",
	"Iran": "IR", "IR Iran": "IR", "Iraq": "IQ",
	"United Arab Emirates": "AE", "Uzbekistan": "UZ", "Jordan": "JO",
	"Indonesia": "ID", "China": "CN", "New Zealand": "NZ",
	"Panama": "PA", "Costa Rica": "CR", "Honduras": "HN",
	"Jamaica": "JM", "Haiti": "HT", "Curaçao": "CW", "Guatemala": "GT",
	"El Salvador": "SV", "Trinidad and Tobago": "TT", "Suriname": "SR",
}

// Subdivision flags have no ISO alpha-2 form; they use Unicode tag
// sequences instead.
var teamFlagSpecial = map[string]string{
	"England":  "\U0001F3F4\U000E0067\U000E0062\U000E0065\U000E006E\U000E0067\U000E007F",
	"Scotland": "\U0001F3F4\U000E0067\U000E0062\U000E0073\U000E0063\U000E0074\U000E007F",
	"Wales":    "\U0001F3F4\U000E0067\U000E0062\U000E0077\U000E006C\U000E0073\U000E007F",
}

// flagFor returns the team's flag emoji, or "" when the name does not
// identify a nation.
func flagFor(team string) string {
	if flag, ok := teamFlagSpecial[team]; ok {
		return flag
	}
	iso, ok := teamISO[team]
	if !ok {
		return ""
	}
	// Two regional indicator symbols spell the ISO code.
	out := make([]rune, 0, 2)
	for _, c := range iso {
		out = append(out, 0x1F1E6+(c-'A'))
	}
	return string(out)
}
