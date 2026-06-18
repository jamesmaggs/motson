// Package venues enriches matches with their stadium when the provider
// omits it. The football-data.org free tier does not supply venues, but the
// 2026 World Cup venue for every match is fixed and public, so we keep a
// static lookup keyed by the provider's match id (the same id stored as
// Match.ProviderMatchID). Sourced from the official FIFA 2026 schedule;
// each entry is annotated with its fixture and date for auditing.
package venues

import "github.com/jamesmaggs/motson/internal/fixtures"

var byID = map[string]string{
	"537327": "Estadio Azteca, Mexico City",            // Mexico vs South Africa, 2026-06-11
	"537328": "Estadio Akron, Guadalajara",             // South Korea vs Czechia, 2026-06-12
	"537333": "BMO Field, Toronto",                     // Canada vs Bosnia-Herzegovina, 2026-06-12
	"537345": "SoFi Stadium, Los Angeles",              // United States vs Paraguay, 2026-06-13
	"537334": "Levi's Stadium, San Francisco Bay Area", // Qatar vs Switzerland, 2026-06-13
	"537339": "MetLife Stadium, New York New Jersey",   // Brazil vs Morocco, 2026-06-13
	"537340": "Gillette Stadium, Boston",               // Haiti vs Scotland, 2026-06-14
	"537346": "BC Place, Vancouver",                    // Australia vs Turkey, 2026-06-14
	"537351": "NRG Stadium, Houston",                   // Germany vs Curaçao, 2026-06-14
	"537357": "AT&T Stadium, Dallas",                   // Netherlands vs Japan, 2026-06-14
	"537352": "Lincoln Financial Field, Philadelphia",  // Ivory Coast vs Ecuador, 2026-06-14
	"537358": "Estadio BBVA, Monterrey",                // Sweden vs Tunisia, 2026-06-15
	"537369": "Mercedes-Benz Stadium, Atlanta",         // Spain vs Cape Verde Islands, 2026-06-15
	"537363": "Lumen Field, Seattle",                   // Belgium vs Egypt, 2026-06-15
	"537370": "Hard Rock Stadium, Miami",               // Saudi Arabia vs Uruguay, 2026-06-15
	"537364": "SoFi Stadium, Los Angeles",              // Iran vs New Zealand, 2026-06-16
	"537391": "MetLife Stadium, New York New Jersey",   // France vs Senegal, 2026-06-16
	"537392": "Gillette Stadium, Boston",               // Iraq vs Norway, 2026-06-16
	"537397": "Arrowhead Stadium, Kansas City",         // Argentina vs Algeria, 2026-06-17
	"537398": "Levi's Stadium, San Francisco Bay Area", // Austria vs Jordan, 2026-06-17
	"537403": "NRG Stadium, Houston",                   // Portugal vs Congo DR, 2026-06-17
	"537409": "AT&T Stadium, Dallas",                   // England vs Croatia, 2026-06-17
	"537410": "BMO Field, Toronto",                     // Ghana vs Panama, 2026-06-17
	"537404": "Estadio Azteca, Mexico City",            // Uzbekistan vs Colombia, 2026-06-18
	"537329": "Mercedes-Benz Stadium, Atlanta",         // Czechia vs South Africa, 2026-06-18
	"537335": "SoFi Stadium, Los Angeles",              // Switzerland vs Bosnia-Herzegovina, 2026-06-18
	"537336": "BC Place, Vancouver",                    // Canada vs Qatar, 2026-06-18
	"537330": "Estadio Akron, Guadalajara",             // Mexico vs South Korea, 2026-06-19
	"537348": "Lumen Field, Seattle",                   // United States vs Australia, 2026-06-19
	"537342": "Gillette Stadium, Boston",               // Scotland vs Morocco, 2026-06-19
	"537341": "Lincoln Financial Field, Philadelphia",  // Brazil vs Haiti, 2026-06-20
	"537347": "Levi's Stadium, San Francisco Bay Area", // Turkey vs Paraguay, 2026-06-20
	"537359": "NRG Stadium, Houston",                   // Netherlands vs Sweden, 2026-06-20
	"537353": "BMO Field, Toronto",                     // Germany vs Ivory Coast, 2026-06-20
	"537354": "Arrowhead Stadium, Kansas City",         // Ecuador vs Curaçao, 2026-06-21
	"537360": "Estadio BBVA, Monterrey",                // Tunisia vs Japan, 2026-06-21
	"537371": "Mercedes-Benz Stadium, Atlanta",         // Spain vs Saudi Arabia, 2026-06-21
	"537365": "SoFi Stadium, Los Angeles",              // Belgium vs Iran, 2026-06-21
	"537372": "Hard Rock Stadium, Miami",               // Uruguay vs Cape Verde Islands, 2026-06-21
	"537366": "BC Place, Vancouver",                    // New Zealand vs Egypt, 2026-06-22
	"537399": "AT&T Stadium, Dallas",                   // Argentina vs Austria, 2026-06-22
	"537393": "Lincoln Financial Field, Philadelphia",  // France vs Iraq, 2026-06-22
	"537394": "MetLife Stadium, New York New Jersey",   // Norway vs Senegal, 2026-06-23
	"537400": "Levi's Stadium, San Francisco Bay Area", // Jordan vs Algeria, 2026-06-23
	"537405": "NRG Stadium, Houston",                   // Portugal vs Uzbekistan, 2026-06-23
	"537411": "Gillette Stadium, Boston",               // England vs Ghana, 2026-06-23
	"537412": "BMO Field, Toronto",                     // Panama vs Croatia, 2026-06-23
	"537406": "Estadio Akron, Guadalajara",             // Colombia vs Congo DR, 2026-06-24
	"537338": "Lumen Field, Seattle",                   // Bosnia-Herzegovina vs Qatar, 2026-06-24
	"537337": "BC Place, Vancouver",                    // Switzerland vs Canada, 2026-06-24
	"537344": "Mercedes-Benz Stadium, Atlanta",         // Morocco vs Haiti, 2026-06-24
	"537343": "Hard Rock Stadium, Miami",               // Scotland vs Brazil, 2026-06-24
	"537331": "Estadio Azteca, Mexico City",            // Czechia vs Mexico, 2026-06-25
	"537332": "Estadio BBVA, Monterrey",                // South Africa vs South Korea, 2026-06-25
	"537356": "Lincoln Financial Field, Philadelphia",  // Curaçao vs Ivory Coast, 2026-06-25
	"537355": "MetLife Stadium, New York New Jersey",   // Ecuador vs Germany, 2026-06-25
	"537362": "AT&T Stadium, Dallas",                   // Japan vs Sweden, 2026-06-25
	"537361": "Arrowhead Stadium, Kansas City",         // Tunisia vs Netherlands, 2026-06-25
	"537350": "Levi's Stadium, San Francisco Bay Area", // Paraguay vs Australia, 2026-06-26
	"537349": "SoFi Stadium, Los Angeles",              // Turkey vs United States, 2026-06-26
	"537395": "Gillette Stadium, Boston",               // Norway vs France, 2026-06-26
	"537396": "BMO Field, Toronto",                     // Senegal vs Iraq, 2026-06-26
	"537374": "NRG Stadium, Houston",                   // Cape Verde Islands vs Saudi Arabia, 2026-06-27
	"537373": "Estadio Akron, Guadalajara",             // Uruguay vs Spain, 2026-06-27
	"537368": "Lumen Field, Seattle",                   // Egypt vs Iran, 2026-06-27
	"537367": "BC Place, Vancouver",                    // New Zealand vs Belgium, 2026-06-27
	"537414": "Lincoln Financial Field, Philadelphia",  // Croatia vs Ghana, 2026-06-27
	"537413": "MetLife Stadium, New York New Jersey",   // Panama vs England, 2026-06-27
	"537407": "Hard Rock Stadium, Miami",               // Colombia vs Portugal, 2026-06-27
	"537408": "Mercedes-Benz Stadium, Atlanta",         // Congo DR vs Uzbekistan, 2026-06-27
	"537402": "Arrowhead Stadium, Kansas City",         // Algeria vs Austria, 2026-06-28
	"537401": "AT&T Stadium, Dallas",                   // Jordan vs Argentina, 2026-06-28
	"537417": "SoFi Stadium, Los Angeles",              // Round of 32, 2026-06-28
	"537423": "NRG Stadium, Houston",                   // Round of 32, 2026-06-29
	"537415": "Gillette Stadium, Boston",               // Round of 32, 2026-06-29
	"537418": "Estadio BBVA, Monterrey",                // Round of 32, 2026-06-30
	"537424": "AT&T Stadium, Dallas",                   // Round of 32, 2026-06-30
	"537416": "MetLife Stadium, New York New Jersey",   // Round of 32, 2026-06-30
	"537425": "Estadio Azteca, Mexico City",            // Round of 32, 2026-07-01
	"537426": "Mercedes-Benz Stadium, Atlanta",         // Round of 32, 2026-07-01
	"537422": "Lumen Field, Seattle",                   // Round of 32, 2026-07-01
	"537421": "Levi's Stadium, San Francisco Bay Area", // Round of 32, 2026-07-02
	"537420": "SoFi Stadium, Los Angeles",              // Round of 32, 2026-07-02
	"537419": "BMO Field, Toronto",                     // Round of 32, 2026-07-02
	"537429": "BC Place, Vancouver",                    // Round of 32, 2026-07-03
	"537428": "AT&T Stadium, Dallas",                   // Round of 32, 2026-07-03
	"537427": "Hard Rock Stadium, Miami",               // Round of 32, 2026-07-03
	"537430": "Arrowhead Stadium, Kansas City",         // Round of 32, 2026-07-04
	"537376": "NRG Stadium, Houston",                   // Round of 16, 2026-07-04
	"537375": "Lincoln Financial Field, Philadelphia",  // Round of 16, 2026-07-04
	"537377": "MetLife Stadium, New York New Jersey",   // Round of 16, 2026-07-05
	"537378": "Estadio Azteca, Mexico City",            // Round of 16, 2026-07-06
	"537379": "AT&T Stadium, Dallas",                   // Round of 16, 2026-07-06
	"537380": "Lumen Field, Seattle",                   // Round of 16, 2026-07-07
	"537381": "Mercedes-Benz Stadium, Atlanta",         // Round of 16, 2026-07-07
	"537382": "BC Place, Vancouver",                    // Round of 16, 2026-07-07
	"537383": "Gillette Stadium, Boston",               // Quarter-final, 2026-07-09
	"537384": "SoFi Stadium, Los Angeles",              // Quarter-final, 2026-07-10
	"537385": "Hard Rock Stadium, Miami",               // Quarter-final, 2026-07-11
	"537386": "Arrowhead Stadium, Kansas City",         // Quarter-final, 2026-07-12
	"537387": "AT&T Stadium, Dallas",                   // Semi-final, 2026-07-14
	"537388": "Mercedes-Benz Stadium, Atlanta",         // Semi-final, 2026-07-15
	"537389": "Hard Rock Stadium, Miami",               // Third place, 2026-07-18
	"537390": "MetLife Stadium, New York New Jersey",   // Final, 2026-07-19
}

// Enrich fills each match's Venue from the static table when the provider
// left it empty; a provider-supplied venue always wins, and matches with no
// known venue are left untouched (surfaces omit venues gracefully).
func Enrich(matches []fixtures.Match) []fixtures.Match {
	for i := range matches {
		if matches[i].Venue == "" {
			if v, ok := byID[matches[i].ProviderMatchID]; ok {
				matches[i].Venue = v
			}
		}
	}
	return matches
}
