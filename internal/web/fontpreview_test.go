package web

import (
	"strings"
	"testing"
)

// The preview gallery offers exactly ten alternatives, none of which is
// Inter (the whole point is to compare against it), and each is fully
// specified so the iframe and stylesheet links render.
func TestPreviewFontsAreTenDistinctNonInterFaces(t *testing.T) {
	if len(previewFonts) != 10 {
		t.Fatalf("want 10 preview typefaces, got %d", len(previewFonts))
	}
	seen := map[string]bool{}
	for _, f := range previewFonts {
		if strings.EqualFold(f.Name, "Inter") {
			t.Errorf("preview set must exclude Inter, found %q", f.Name)
		}
		if f.Slug == "" || f.Name == "" || f.Href == "" {
			t.Errorf("incomplete preview font: %+v", f)
		}
		if seen[f.Slug] {
			t.Errorf("duplicate slug %q", f.Slug)
		}
		seen[f.Slug] = true
		if !strings.HasPrefix(string(f.Href), "https://fonts.googleapis.com/css2?family=") {
			t.Errorf("font %q has an unexpected stylesheet URL: %s", f.Slug, f.Href)
		}
	}
}

func TestFontBySlugResolvesKnownAndRejectsUnknown(t *testing.T) {
	if f, ok := fontBySlug("sora"); !ok || f.Name != "Sora" {
		t.Errorf(`fontBySlug("sora") = %+v, %v; want Sora, true`, f, ok)
	}
	if _, ok := fontBySlug("not-a-font"); ok {
		t.Error(`fontBySlug("not-a-font") = true; want false`)
	}
}
