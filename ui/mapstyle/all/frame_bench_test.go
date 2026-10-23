package all

import (
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/ui/mapstyle"
	"github.com/espresso20/ageforge/ui/mapstyle/capture"
	"github.com/espresso20/ageforge/vibe"
)

// benchAges are a ground age of the city, the Earth arc's last, and two ages
// of the sky: one frame of each style at each is the map's typical cost.
var benchAges = []string{"medieval_age", "modern_age", "fusion_age", "space_age", "galactic_age"}

// BenchmarkMapFrame is one 200x60 frame of each map style at a few ages, with
// VibeMode off and on: the cost a player pays every redraw of the map, and
// what the switch adds to it (an atomic load for each text the frame writes,
// and with it on, a lookup of the script for each).
func BenchmarkMapFrame(b *testing.B) {
	for _, style := range Registry().Names() {
		for _, age := range benchAges {
			for _, on := range []bool{false, true} {
				name := "off"
				if on {
					name = "on"
				}
				b.Run(style+"/"+age+"/vibemode_"+name, func(b *testing.B) {
					was := vibe.On()
					b.Cleanup(func() { vibe.Set(was) })
					vibe.Set(false)
					m := arcModel(age)
					v, _ := Registry().New(style)
					scr := capture.NewScreen(200, 60)
					defer scr.Fini()
					r := mapstyle.Rect{W: 200, H: 60}
					v.Draw(scr, r, mapstyle.Frame{Model: m})
					vibe.Set(on)
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						v.Draw(scr, r, mapstyle.Frame{Model: m, Anim: i, Tier: mapmodel.TierUnicode})
					}
				})
			}
		}
	}
}
