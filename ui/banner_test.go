package ui

import (
	"image/color"
	"testing"

	"github.com/espresso20/ageforge/rules"
)

// The banner fits its grid, every glyph of its line is in the banner's face
// (a change to the number of ages may need a letter added), and it comes out
// the size its cells say.
func TestBannerImage(t *testing.T) {
	g, _ := bannerGrid(len(rules.Core().AgeKeys()))
	if len(g.clipped) > 0 {
		t.Fatalf("banner text cut at the edge: %q", g.clipped)
	}
	for _, r := range menuTagline(len(rules.Core().AgeKeys())) {
		if _, ok := bannerFont[r]; !ok {
			t.Errorf("the banner's face has no %q", r)
		}
	}
	img := BannerImage(len(rules.Core().AgeKeys()), 10, 20)
	if b := img.Bounds(); b.Dx() != bannerW*10 || b.Dy() != bannerH*20 {
		t.Errorf("banner is %dx%d, want %dx%d", b.Dx(), b.Dy(), bannerW*10, bannerH*20)
	}
}

// The icon is 1024x1024, its wordmark colours vary (the gradient), and
// the corner pixels are the background.
func TestIconImage(t *testing.T) {
	img := IconImage(1024)
	b := img.Bounds()
	if b.Dx() != 1024 || b.Dy() != 1024 {
		t.Errorf("icon is %dx%d, want 1024x1024", b.Dx(), b.Dy())
	}

	// Check for gradient: collect unique colours (excluding the background).
	_, pal := iconGrid()
	bg := rgba(pal.bg)
	colours := make(map[color.RGBA]bool)
	for y := 0; y < 1024; y++ {
		for x := 0; x < 1024; x++ {
			c := img.At(x, y).(color.RGBA)
			if c != bg {
				colours[c] = true
			}
		}
	}
	if len(colours) < 2 {
		t.Errorf("icon has %d unique colours (excluding background), want at least 2 (gradient)", len(colours))
	}

	// Check corner pixels are background.
	corners := []struct{ x, y int }{
		{0, 0}, {1023, 0}, {0, 1023}, {1023, 1023},
	}
	for _, c := range corners {
		if col := img.At(c.x, c.y).(color.RGBA); col != bg {
			t.Errorf("corner (%d, %d) is not background", c.x, c.y)
		}
	}
}
