package theme

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestShade(t *testing.T) {
	c := tcell.NewRGBColor(100, 200, 40)
	if got := Shade(c, 1); got != c {
		t.Errorf("Shade by 1 = %v, want %v", got, c)
	}
	r, g, b := Shade(c, 0.5).RGB()
	if r != 50 || g != 100 || b != 20 {
		t.Errorf("Shade by 0.5 = %d,%d,%d, want 50,100,20", r, g, b)
	}
	r, g, b = Shade(c, 2).RGB()
	if r != 200 || g != 255 || b != 80 {
		t.Errorf("Shade by 2 = %d,%d,%d, want 200,255,80 (clamped)", r, g, b)
	}
	if got := Shade(tcell.ColorDefault, 0.5); got != tcell.ColorDefault {
		t.Errorf("Shade of the default colour changed it: %v", got)
	}
}

func TestTint(t *testing.T) {
	c := tcell.NewRGBColor(200, 100, 50)
	if got := Tint(c, tcell.NewRGBColor(255, 255, 255)); got != c {
		t.Errorf("Tint by white = %v, want %v", got, c)
	}
	r, g, b := Tint(c, tcell.NewRGBColor(0, 255, 51)).RGB()
	if r != 0 || g != 100 || b != 10 {
		t.Errorf("Tint = %d,%d,%d, want 0,100,10", r, g, b)
	}
}

func TestSkyTables(t *testing.T) {
	for f := 0; f < SkyFamilies; f++ {
		for v := 0; v < SkyVariants; v++ {
			m := SkyMaterialFor(f, v)
			for _, c := range []tcell.Color{m.Wall, m.Roof, m.Trim, m.Metal, m.Glass, m.GlassHi, m.Neon1, m.Neon2, m.Neon3, m.Glow} {
				if !c.Valid() {
					t.Fatalf("material %d/%d has an invalid colour", f, v)
				}
			}
		}
	}
	if SkyMaterialFor(-3, 99) != SkyMaterialFor(0, SkyVariants-1) {
		t.Error("SkyMaterialFor does not clamp its indexes")
	}
	for e := 0; e < 7; e++ {
		k := SkyKeysFor(e)
		for _, g := range []SkyGradient{k.Day, k.Dusk, k.Dawn, k.Night} {
			for _, c := range g {
				if !c.Valid() {
					t.Fatalf("epoch %d sky has an invalid colour", e)
				}
			}
		}
	}
	for h := SkyHue(0); h < numSkyHues; h++ {
		if skyHues[h] == 0 && h != SkyInk {
			t.Errorf("sky hue %d has no colour", h)
		}
	}
}
