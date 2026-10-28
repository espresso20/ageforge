package mapstyle

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/espresso20/ageforge/mapgen"
	"github.com/espresso20/ageforge/mapmodel/fixture"
	"github.com/espresso20/ageforge/theme"
)

// onRamp is the distance of c from the nearest colour on the ramp from the
// theme's background to its text.
func onRamp(c tcell.Color) float64 {
	bg, text := theme.Color(theme.RoleBackground), theme.Color(theme.RoleText)
	best := 99.0
	for i := 0; i <= 100; i++ {
		best = min(best, theme.Distance(c, theme.Mix(bg, text, float64(i)/100)))
	}
	return best
}

// TestDuotoneThemesDrawOnTheirRamp: on Monochrome and Parchment, one ink on
// one paper, every colour the map draws in lies on the ramp from the theme's
// background to its text; on a theme with colour, some colours do not.
func TestDuotoneThemesDrawOnTheirRamp(t *testing.T) {
	for th := range fixture.Themes(t) {
		off := 0
		p := cellPaletteFor(nil, &mapgen.Grid{Wall: mapgen.HStone, Tint: mapgen.HNone})
		for _, c := range append(p.hue[:], p.night) {
			if onRamp(c) > 0.01 {
				off++
			}
		}
		switch {
		case th.Duotone && off > 0:
			t.Errorf("%s is one ink on one paper, and %d of the map's colours are off its ramp", th.Key, off)
		case !th.Duotone && off == 0:
			t.Errorf("%s has colour, and every colour the map draws in is on one ramp", th.Key)
		}
	}
}
