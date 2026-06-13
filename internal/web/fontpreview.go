package web

import (
	"html/template"
	"net/http"
)

// PREVIEW TOOLING (impeccable branch only — remove before merging to
// main). The deterministic slop detector flags Inter as an over-used
// AI-default body face, so this serves the index rendered with ten
// alternative typefaces side by side at /fonts, plus a ?font=<slug>
// override on the index itself. Production (no ?font) stays on Inter.

// fontOption is one alternative body/data typeface for the gallery.
type fontOption struct {
	Slug string       // ?font= value and iframe target
	Name string       // family name, used in the CSS stack and as a label
	Note string       // one-line description of the face's character
	Href template.URL // Google Fonts stylesheet (preview-only network dep)
}

// previewFonts are ten non-Inter faces chosen for legibility at small
// sizes and tabular data, with a little more character than Inter. Each
// supports the weights the design uses (300/400/600/700) on Google Fonts.
var previewFonts = []fontOption{
	{"hanken-grotesk", "Hanken Grotesk", "Warm geometric grotesque, friendly but neutral", googleFont("Hanken+Grotesk")},
	{"sora", "Sora", "Technical, faintly squared — quietly sporty", googleFont("Sora")},
	{"manrope", "Manrope", "Modern semi-rounded, soft terminals", googleFont("Manrope")},
	{"outfit", "Outfit", "Clean geometric, even and unfussy", googleFont("Outfit")},
	{"ibm-plex-sans", "IBM Plex Sans", "Humanist with engineered detail", googleFont("IBM+Plex+Sans")},
	{"public-sans", "Public Sans", "Neutral, plainspoken, highly legible", googleFont("Public+Sans")},
	{"archivo", "Archivo", "Sturdy grotesque with a broadcast feel", googleFont("Archivo")},
	{"chivo", "Chivo", "Grotesque with sporting, scoreboard energy", googleFont("Chivo")},
	{"saira", "Saira", "Slightly narrow, technical and modern", googleFont("Saira")},
	{"figtree", "Figtree", "Friendly geometric, rounded and open", googleFont("Figtree")},
}

// googleFont builds a css2 stylesheet URL for the weights the design
// uses. family is the '+'-encoded family name.
func googleFont(family string) template.URL {
	return template.URL("https://fonts.googleapis.com/css2?family=" + family + ":wght@300;400;600;700&display=swap")
}

func fontBySlug(slug string) (fontOption, bool) {
	for _, f := range previewFonts {
		if f.Slug == slug {
			return f, true
		}
	}
	return fontOption{}, false
}

type fontsPageData struct {
	AssetVersion string
	Fonts        []fontOption
}

// fontGallery renders the side-by-side comparison page at /fonts.
func fontGallery() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, "fonts.html.tmpl", fontsPageData{AssetVersion: assetVersion, Fonts: previewFonts})
	}
}
