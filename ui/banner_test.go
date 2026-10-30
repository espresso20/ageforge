package ui

import (
	"testing"

	"github.com/espresso20/ageforge/rules"
)

// The banner fits its grid, every glyph of its line is in the banner's face
// (a change to the number of ages may need a letter added), and it comes out
// the size its cells say.
func TestBannerImage(t *testing.T) {
	g, _ := bannerGrid()
	if len(g.clipped) > 0 {
		t.Fatalf("banner text cut at the edge: %q", g.clipped)
	}
	for _, r := range menuTagline(len(rules.Core().AgeKeys())) {
		if _, ok := bannerFont[r]; !ok {
			t.Errorf("the banner's face has no %q", r)
		}
	}
	img := BannerImage(10, 20)
	if b := img.Bounds(); b.Dx() != bannerW*10 || b.Dy() != bannerH*20 {
		t.Errorf("banner is %dx%d, want %dx%d", b.Dx(), b.Dy(), bannerW*10, bannerH*20)
	}
}
