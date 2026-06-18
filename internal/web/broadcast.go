package web

// ukBroadcaster maps a provider match id to the UK free-to-air channel
// showing that game — "BBC" or "ITV". All 104 matches are split between the
// two, but only the 72 group-stage games have an announced channel; the
// knockout picks (Round of 32 onward) are not published until later, so they
// are absent here and render without a badge. England's games are simulcast
// on both networks; the listed lead channel is used. Presentation metadata
// like the team flags, sourced from the published UK TV listings; each entry
// is annotated with its fixture and date for auditing.
var ukBroadcaster = map[string]string{
	"537327": "ITV", // Mexico vs South Africa, 2026-06-11
	"537328": "ITV", // South Korea vs Czechia, 2026-06-12
	"537333": "BBC", // Canada vs Bosnia-Herzegovina, 2026-06-12
	"537345": "BBC", // United States vs Paraguay, 2026-06-13
	"537334": "ITV", // Qatar vs Switzerland, 2026-06-13
	"537339": "BBC", // Brazil vs Morocco, 2026-06-13
	"537340": "BBC", // Haiti vs Scotland, 2026-06-14
	"537346": "ITV", // Australia vs Turkey, 2026-06-14
	"537351": "ITV", // Germany vs Curaçao, 2026-06-14
	"537357": "ITV", // Netherlands vs Japan, 2026-06-14
	"537352": "BBC", // Ivory Coast vs Ecuador, 2026-06-14
	"537358": "ITV", // Sweden vs Tunisia, 2026-06-15
	"537369": "ITV", // Spain vs Cape Verde Islands, 2026-06-15
	"537363": "BBC", // Belgium vs Egypt, 2026-06-15
	"537370": "ITV", // Saudi Arabia vs Uruguay, 2026-06-15
	"537364": "BBC", // Iran vs New Zealand, 2026-06-16
	"537391": "BBC", // France vs Senegal, 2026-06-16
	"537392": "BBC", // Iraq vs Norway, 2026-06-16
	"537397": "ITV", // Argentina vs Algeria, 2026-06-17
	"537398": "BBC", // Austria vs Jordan, 2026-06-17
	"537403": "BBC", // Portugal vs Congo DR, 2026-06-17
	"537409": "ITV", // England vs Croatia, 2026-06-17
	"537410": "ITV", // Ghana vs Panama, 2026-06-17
	"537404": "BBC", // Uzbekistan vs Colombia, 2026-06-18
	"537329": "BBC", // Czechia vs South Africa, 2026-06-18
	"537335": "ITV", // Switzerland vs Bosnia-Herzegovina, 2026-06-18
	"537336": "ITV", // Canada vs Qatar, 2026-06-18
	"537330": "BBC", // Mexico vs South Korea, 2026-06-19
	"537348": "BBC", // United States vs Australia, 2026-06-19
	"537342": "ITV", // Scotland vs Morocco, 2026-06-19
	"537341": "ITV", // Brazil vs Haiti, 2026-06-20
	"537347": "ITV", // Turkey vs Paraguay, 2026-06-20
	"537359": "BBC", // Netherlands vs Sweden, 2026-06-20
	"537353": "ITV", // Germany vs Ivory Coast, 2026-06-20
	"537354": "BBC", // Ecuador vs Curaçao, 2026-06-21
	"537360": "BBC", // Tunisia vs Japan, 2026-06-21
	"537371": "BBC", // Spain vs Saudi Arabia, 2026-06-21
	"537365": "ITV", // Belgium vs Iran, 2026-06-21
	"537372": "BBC", // Uruguay vs Cape Verde Islands, 2026-06-21
	"537366": "ITV", // New Zealand vs Egypt, 2026-06-22
	"537399": "BBC", // Argentina vs Austria, 2026-06-22
	"537393": "BBC", // France vs Iraq, 2026-06-22
	"537394": "ITV", // Norway vs Senegal, 2026-06-23
	"537400": "ITV", // Jordan vs Algeria, 2026-06-23
	"537405": "ITV", // Portugal vs Uzbekistan, 2026-06-23
	"537411": "BBC", // England vs Ghana, 2026-06-23
	"537412": "BBC", // Panama vs Croatia, 2026-06-23
	"537406": "ITV", // Colombia vs Congo DR, 2026-06-24
	"537338": "ITV", // Bosnia-Herzegovina vs Qatar, 2026-06-24
	"537337": "ITV", // Switzerland vs Canada, 2026-06-24
	"537344": "BBC", // Morocco vs Haiti, 2026-06-24
	"537343": "BBC", // Scotland vs Brazil, 2026-06-24
	"537331": "BBC", // Czechia vs Mexico, 2026-06-25
	"537332": "BBC", // South Africa vs South Korea, 2026-06-25
	"537356": "BBC", // Curaçao vs Ivory Coast, 2026-06-25
	"537355": "BBC", // Ecuador vs Germany, 2026-06-25
	"537362": "BBC", // Japan vs Sweden, 2026-06-25
	"537361": "BBC", // Tunisia vs Netherlands, 2026-06-25
	"537350": "ITV", // Paraguay vs Australia, 2026-06-26
	"537349": "ITV", // Turkey vs United States, 2026-06-26
	"537395": "ITV", // Norway vs France, 2026-06-26
	"537396": "ITV", // Senegal vs Iraq, 2026-06-26
	"537374": "ITV", // Cape Verde Islands vs Saudi Arabia, 2026-06-27
	"537373": "ITV", // Uruguay vs Spain, 2026-06-27
	"537368": "BBC", // Egypt vs Iran, 2026-06-27
	"537367": "BBC", // New Zealand vs Belgium, 2026-06-27
	"537414": "ITV", // Croatia vs Ghana, 2026-06-27
	"537413": "ITV", // Panama vs England, 2026-06-27
	"537407": "BBC", // Colombia vs Portugal, 2026-06-27
	"537408": "BBC", // Congo DR vs Uzbekistan, 2026-06-27
	"537402": "BBC", // Algeria vs Austria, 2026-06-28
	"537401": "BBC", // Jordan vs Argentina, 2026-06-28
}

// broadcasterFor returns the UK channel showing a match, or "" when none is
// known (knockout games until their channel is announced).
func broadcasterFor(providerMatchID string) string {
	return ukBroadcaster[providerMatchID]
}
