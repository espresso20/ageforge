package mapgen

import (
	"reflect"
	"testing"
)

func cosmosAt(age string, seed int64, anim int) *Overview {
	return (&Generator{}).Draw(model(state(age, seed, noon)), anim).OverviewCosmos(200, 55)
}

func nebulaBg(g *Grid) int {
	n := 0
	for _, c := range g.Cells {
		if c.Bg.Hue == HNebula {
			n++
		}
	}
	return n
}

// Same seed and frame, same picture; the cosmic view has no purple blotches.
func TestCosmosIsStillAndHasNoBlotches(t *testing.T) {
	a, b := cosmosAt("interstellar_age", 7, 900), cosmosAt("interstellar_age", 7, 900)
	if !reflect.DeepEqual(a.Grid.Cells, b.Grid.Cells) {
		t.Error("the cosmic view differs for one seed and frame")
	}
	if n := nebulaBg(a.Grid); n != 0 {
		t.Errorf("%d cells still have a nebula ground at the cosmic view", n)
	}
	sc := (&Generator{}).Draw(model(state("interstellar_age", 7, noon)), 900)
	if n := nebulaBg(sc.Overview(200, 55).Grid); n == 0 {
		t.Error("the overview lost its nebula; only the cosmic view should")
	}
}

// the band's glow and the Oort specks exist at the cosmic view and not at the
// others
func TestCosmosHasItsLayers(t *testing.T) {
	sc := (&Generator{}).Draw(model(state("interstellar_age", 7, noon)), 900)
	glow := func(g *Grid) (n int) {
		for _, c := range g.Cells {
			if c.Bg.Hue == HCool {
				n++
			}
		}
		return
	}
	if n := glow(sc.OverviewCosmos(200, 55).Grid); n < 200 {
		t.Errorf("only %d cells of Milky Way glow at the cosmic view", n)
	}
	for name, g := range map[string]*Grid{"overview": sc.Overview(200, 55).Grid, "inner": sc.OverviewInner(200, 55).Grid} {
		if n := glow(g); n > 10 {
			t.Errorf("%s: %d cells of Milky Way glow", name, n)
		}
	}
	// the specks of the Oort cloud and the belt are pixels; count the cells with a
	// quadrant glyph at the cosmic view against the inner view
	dots := func(g *Grid) (n int) {
		for _, c := range g.Cells {
			if c.R != ' ' && c.R != 0 {
				n++
			}
		}
		return
	}
	if cosmos, inner := dots(sc.OverviewCosmos(200, 55).Grid), dots(sc.OverviewInner(200, 55).Grid); cosmos < 3*inner/2 {
		t.Errorf("%d glyph cells at the cosmic view, %d inner", cosmos, inner)
	}
}

// Relay beacons stand from the Interstellar age; their lights blink slowly.
func TestCosmosBeacons(t *testing.T) {
	sc := (&Generator{}).Draw(model(state("interstellar_age", 7, noon)), 0)
	o := &orrery{s: sc, pix: NewPixels(200, 55), seed: 7, maxRY: 10}
	defer o.pix.Release()
	for anim := 0; anim < 100; anim++ {
		sc.Anim = anim
		o.beacons()
	}
	if got := countSet(o.pix); got < 3 || got > 5 {
		t.Errorf("%d beacons, want 3 to 5", got)
	}
	for k := 0; k < 5; k++ {
		changes, prev := 0, blinkWith(7, 0, 20+k, beaconPeriod, beaconLit)
		for anim := 1; anim <= 100; anim++ {
			if on := blinkWith(7, anim, 20+k, beaconPeriod, beaconLit); on != prev {
				changes++
				prev = on
			}
		}
		if changes > 4 {
			t.Errorf("beacon %d changes %d times in 100 frames", k, changes)
		}
	}
	space := (&Generator{}).Draw(model(state("space_age", 7, noon)), 0).OverviewCosmos(200, 55)
	for _, h := range space.Grid.Hit {
		if th := space.Grid.Things[h]; th.Key == "sun" {
			t.Error("the Space Age shows the Sun at its cosmic view")
		}
	}
}

func countSet(p *Pixels) (n int) {
	for _, c := range p.px {
		if c.Hue == HCyan || c.Hue == HSteel {
			n++
		}
	}
	return
}

// Nothing new is drawn over a planet: the Sun and the planets keep their cells.
func TestCosmosKeepsTheBodies(t *testing.T) {
	sc := (&Generator{}).Draw(model(state("interstellar_age", 7, noon)), 900)
	cells := func(g *Grid) (n int) {
		for _, h := range g.Hit {
			if th := g.Things[h]; th.Kind == KindTerrain && (th.Key == "sun" && th.Ref == 0 || th.Key == "planet" || th.Key == "home") {
				n++
			}
		}
		return
	}
	if n := cells(sc.OverviewCosmos(200, 55).Grid); n < 9 {
		t.Errorf("only %d cells of the Sun and the planets at the cosmic view", n)
	}
}
