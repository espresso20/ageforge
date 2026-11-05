package mapgen

import "testing"

// No moving ship is drawn in a space scene at any zoom: not at the orbit or
// the system, not in the overviews, the close-up or the port.
func TestSpaceHasNoShips(t *testing.T) {
	for _, age := range []string{"space_age", "interstellar_age"} {
		for _, anim := range []int{0, 300, 900, 1500} {
			sc := (&Generator{}).Draw(model(state(age, 7, noon)), anim)
			grids := map[string]*Grid{"scene": sc.Grid, "overview": sc.Overview(120, 40).Grid, "cosmos": sc.OverviewCosmos(120, 40).Grid, "inner": sc.OverviewInner(120, 40).Grid,
				"close": closeUp(age, 7, anim).Grid, "port": portAt(1, true).Grid}
			for name, g := range grids {
				for _, h := range g.Hit {
					if g.Things[h].Kind == KindAgent {
						t.Fatalf("%s %s anim %d: a ship %q", age, name, anim, g.Things[h].Key)
					}
				}
			}
		}
	}
}

// A light is on in some frames and off in others, and two lights of one seed
// do not blink together.
func TestLightsBlink(t *testing.T) {
	on, apart := 0, false
	for anim := 0; anim < 10*blinkPeriod; anim++ {
		if blinkOn(7, anim, 0) {
			on++
		}
		apart = apart || blinkOn(7, anim, 0) != blinkOn(7, anim, 1) || blinkOn(7, anim, 2) != blinkOn(7, anim, 3)
	}
	if on != 10 || !apart {
		t.Errorf("lit %d frames in %d, apart=%v", on, 10*blinkPeriod, apart)
	}
	p := NewPixels(4, 4)
	defer p.Release()
	for anim := 0; anim < blinkPeriod; anim++ {
		p.light(7, anim, 0, 1, 1, HRed, 0)
		if (p.px[p.W+1].Hue == HRed) != (anim == (blinkPeriod-roll(blinkPeriod, int64(7), 5400, 0))%blinkPeriod) {
			t.Fatalf("anim %d: light on=%v", anim, p.px[p.W+1].Hue == HRed)
		}
		p.px[p.W+1] = Paint{}
	}
}

// The same seed and frame draw the same close-up.
func TestStationLightsSameSeedSameFrame(t *testing.T) {
	a, b := closeUp("space_age", 7, 123).Grid, closeUp("space_age", 7, 123).Grid
	for i := range a.Cells {
		if a.Cells[i] != b.Cells[i] {
			t.Fatalf("cell %d differs between two draws of one frame", i)
		}
	}
}

// The station by the home world shows at the overview from the Space Age on
// and not before.
func TestOverviewStation(t *testing.T) {
	has := func(age string, view func(*Scene) *Grid) bool {
		g := view((&Generator{}).Draw(model(state(age, 7, noon)), 100))
		for _, h := range g.Hit {
			if g.Things[h].Key == "space_station" {
				return true
			}
		}
		return false
	}
	over := func(s *Scene) *Grid { return s.Overview(120, 40).Grid }
	inner := func(s *Scene) *Grid { return s.OverviewInner(120, 40).Grid }
	if has("modern_age", over) {
		t.Error("modern age: a station at the overview")
	}
	for _, age := range []string{"space_age", "interstellar_age"} {
		if !has(age, over) {
			t.Errorf("%s: no station at the overview", age)
		}
	}
	if !has("interstellar_age", inner) {
		t.Error("no station at the inner view")
	}
}
