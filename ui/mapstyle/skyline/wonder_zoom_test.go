package skyline

import (
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// TestWonderFormAtEveryZoom: a wonder, built or still going up in its
// scaffolding, is the same landmark at the close level as at mid: every cell
// of it that the close window holds becomes a 2x2 block of the landmark, so
// it is twice the size both ways and never a thin single-column mast with a
// mark on it. The same camera and the same world frame at both levels, so
// the cells correspond one to one.
func TestWonderFormAtEveryZoom(t *testing.T) {
	const W, H = 160, 45
	S := H - 3 // the scene rows of the screen
	checked := 0
	for _, age := range []string{"stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age", "renaissance_age", "colonial_age", "industrial_age", "victorian_age", "electric_age", "atomic_age", "modern_age", "information_age", "digital_age"} {
		opts := fixture.Options{Age: age, Seed: 3, Routes: 4, Tick: tickAt(0.45)}
		key := fixture.State(opts).CurrentAgeWonderKey
		if key == "" {
			continue
		}
		for how, m := range map[string]*mapmodel.Model{"going up": build(opts, 1), "built": builtWonder(age, key)} {
			if m.Sky() != mapmodel.SkyGround {
				continue
			}
			x := wonderX(m.Skyline.Districts[m.AgeIdx], key)
			// the close window is half as wide as mid's: it starts at the camera
			// the close level would take (the wonder in its middle), and mid
			// starts at or before it, so every cell the window holds is also in
			// mid's frame, off columns to its right
			camClose := clampInt(x-W/4, 0, max(0, m.Skyline.Width-W/2))
			camMid := min(camClose, max(0, m.Skyline.Width-W))
			off := camClose - camMid
			var frames [2]*fb
			for i, z := range []zoom{zMid, zClose} {
				v := newView()
				v.zoom, v.sw = z, W
				v.follow = false
				v.cam = []int{camMid, camClose}[i]
				draw(v, m, W, H, 40, mapmodel.TierUnicode, false)
				frames[i] = &fb{w: v.fb.w, h: v.fb.h, c: append([]cell(nil), v.fb.c...)}
			}
			mid, closeUp := frames[0], frames[1]
			cells, inWindow := 0, 0
		cellsLoop:
			for y := 1; y <= S; y++ {
				for mx := off; mx < off+W/2; mx++ {
					if mid.at(mx, y).d != dWonder {
						continue
					}
					cells++
					r := 2*(y-1) - S // the screen scene row of the world row's upper cell
					if r < 0 {
						continue // above the window
					}
					inWindow++
					for dy := 0; dy < 2; dy++ {
						for dx := 0; dx < 2; dx++ {
							// (something nearer, a train's steam or a walker drawn
							// at the new frame's own place, may stand in front)
							if c := closeUp.at(2*(mx-off)+dx, 1+r+dy); c == nil || c.d > dWonder {
								if c == nil {
									c = &cell{}
								}
								t.Errorf("%s %s: %s's mid cell (%d,%d) is not a 2x2 block of it at close (%q, depth %d, at (%d,%d))", age, how, key, mx, y, c.ch, c.d, 2*(mx-off)+dx, 1+r+dy)
								break cellsLoop
							}
						}
					}
				}
			}
			if cells < 20 || inWindow < 20 {
				t.Errorf("%s %s: only %d cells of %s at mid, %d in the close window", age, how, cells, key, inWindow)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("no wonder was checked")
	}
}
