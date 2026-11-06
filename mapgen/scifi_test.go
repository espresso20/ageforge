package mapgen

import (
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/mapmodel/fixture"
)

func cometCells(g *Grid) int {
	return portCells(g, func(th Thing) bool { return th.Kind == KindTerrain && th.Key == "comet" })
}

// The comet crosses the cosmic view and no other.
func TestCometOnlyAtCosmos(t *testing.T) {
	anim := -1
	for a := 0; a < cometCycle && anim < 0; a++ {
		if _, _, ok := cometPath(7, a, 240, 80); ok {
			anim = a + 600
		}
	}
	sc := (&Generator{}).Draw(model(state("interstellar_age", 7, noon)), anim)
	if n := cometCells(sc.OverviewCosmos(120, 40).Grid); n == 0 {
		t.Error("no comet at the cosmos zoom")
	}
	for name, g := range map[string]*Grid{"overview": sc.Overview(120, 40).Grid, "inner": sc.OverviewInner(120, 40).Grid} {
		if n := cometCells(g); n != 0 {
			t.Errorf("%s: %d comet cells", name, n)
		}
	}
	again := (&Generator{}).Draw(model(state("interstellar_age", 7, noon)), anim)
	if !reflect.DeepEqual(sc.OverviewCosmos(120, 40).Grid.Cells, again.OverviewCosmos(120, 40).Grid.Cells) {
		t.Error("the cosmos frame is not the same for one seed and frame")
	}
}

// The head never moves more than a pixel a frame, keeps moving, and crosses
// in minutes with a quiet gap between crossings.
func TestCometDriftsAndRests(t *testing.T) {
	const w, h = 240, 80
	var px, py int
	prev, run, gap, longest, shortestCross := false, 0, 0, 0, 1<<30
	moved := false
	for a := 0; a < 4*cometCycle; a++ {
		x, y, ok := cometPath(7, a, w, h)
		if ok && prev && (abs(x/pxUnit-px/pxUnit) > 1 || abs(y/pxUnit-py/pxUnit) > 1) {
			t.Fatalf("frame %d: the head jumped (%d,%d) -> (%d,%d)", a, px, py, x, y)
		}
		if ok && prev && a%100 == 99 && (x != px || y != py) {
			moved = true
		}
		switch {
		case ok:
			run++
			if gap > longest {
				longest = gap
			}
			gap = 0
		default:
			if run > 0 {
				shortestCross = min(shortestCross, run)
			}
			run = 0
			gap++
		}
		prev, px, py = ok, x, y
	}
	if !moved {
		t.Error("the comet never moved")
	}
	if longest < 8*60 || shortestCross < 3*60*8 {
		t.Errorf("gap %d frames, shortest crossing %d frames", longest, shortestCross)
	}
}

func portParts(anim int) (*frame, *port, *Pixels) {
	m := model(fixture.State(fixture.Options{Age: "interstellar_age", Seed: 7, Tick: noon, Scale: 1}))
	f := (&Generator{}).frame(m, anim)
	f.close = true
	return f, f.port(), NewPixels(f.g.W, f.g.H)
}

func litSet(p *Pixels, skip func(Paint) bool) []int {
	var out []int
	for i, c := range p.px {
		if c.Hue != HNone && !skip(c) {
			out = append(out, i)
		}
	}
	return out
}

// Earth shows city lights on its dark half, on the disc.
func TestPortCityLights(t *testing.T) {
	f, pp, pix := portParts(0)
	shade := lighting(7)
	f.cityLights(pix, pp, shade)
	n := 0
	for i, c := range pix.px {
		if c.Hue == HNone {
			continue
		}
		n++
		x, y := i%pix.W-pp.cx, i/pix.W-pp.cy
		if x*x+y*y > pp.R*pp.R {
			t.Fatalf("a city light off the disc at (%d,%d)", x, y)
		}
		if shade(P(HPanel, 255), x*256/pp.R, y*256/pp.R).Weight > 130 {
			t.Fatalf("a city light in daylight at (%d,%d)", x, y)
		}
	}
	if n < 8 {
		t.Errorf("%d city lights", n)
	}
}

// Two or three elevators run from the planet to the ring.
func TestPortTethersJoinEarthAndRing(t *testing.T) {
	f, pp, pix := portParts(0)
	ts := pp.tethers(7)
	if len(ts) < 2 || len(ts) > 3 {
		t.Fatalf("%d tethers", len(ts))
	}
	f.paintTethers(pix, pp)
	for _, tt := range ts {
		if d := (tt.x0-pp.cx)*(tt.x0-pp.cx) + (tt.y0-pp.cy)*(tt.y0-pp.cy); d > pp.R*pp.R {
			t.Errorf("a tether starts off the disc: %d", d)
		}
		onRing := false
		for i := 0; i < 400; i++ {
			if x, y, _ := pp.at(i*turn/400, 100); abs(x-tt.x1) <= 1 && abs(y-tt.y1) <= 1 {
				onRing = true
			}
		}
		if !onRing {
			t.Errorf("a tether ends off the ring at (%d,%d)", tt.x1, tt.y1)
		}
		if pix.px[tt.y0*pix.W+tt.x0].Hue == HNone || pix.px[tt.y1*pix.W+tt.x1].Hue == HNone {
			t.Error("a tether is not drawn end to end")
		}
	}
}

// stable says whether the pixels the draw puts down hold still over 50 frames:
// only a beacon's pixel (a lamp hue) may come and go.
func stable(t *testing.T, draw func(f *frame, pp *port, pix *Pixels)) (count int, moved bool) {
	seen := map[int]int{}
	lamp := map[int]bool{}
	var glow []int
	for anim := 0; anim < 50; anim++ {
		f, pp, pix := portParts(anim)
		draw(f, pp, pix)
		g := 0
		for i, c := range pix.px {
			if c.Hue == HNone {
				continue
			}
			seen[i]++
			lamp[i] = lamp[i] || c.Hue == HRed || c.Hue == HLife
			if c.Hue == HPlasma {
				g += int(c.Weight)
			}
		}
		glow = append(glow, g)
		pix.Release()
	}
	for i, n := range seen {
		if n != 50 && !lamp[i] {
			t.Errorf("pixel %d is drawn in %d of 50 frames", i, n)
		}
	}
	for _, g := range glow {
		moved = moved || g != glow[0]
	}
	return len(seen), moved
}

// The docked starship is long, never moves and its drive glow pulses.
func TestPortStarshipStaysDocked(t *testing.T) {
	n, pulse := stable(t, func(f *frame, pp *port, pix *Pixels) { f.starship(pix, pp) })
	if n < 60 {
		t.Errorf("a ship of %d pixels", n)
	}
	if !pulse {
		t.Error("the drive glow does not pulse")
	}
}

// The big stations are still: only their beacons change.
func TestPortMegaStationsStandStill(t *testing.T) {
	n, _ := stable(t, func(f *frame, pp *port, pix *Pixels) { f.megaStations(pix, pp, lighting(7)) })
	if n < 120 {
		t.Errorf("only %d station pixels", n)
	}
}

// The swarm drifts a pixel at a time, fades softly and never flickers.
func TestPortSwarmIsCalm(t *testing.T) {
	_, pp, _ := portParts(0)
	for i := 0; i < swarmSize; i++ {
		px, py, pw := 0, 0, 0
		var changes []int
		for a := 0; a < 600; a++ {
			x, y, w := pp.sat(7, a, i)
			if w > 130 {
				t.Fatalf("satellite %d is bright: %d", i, w)
			}
			if a > 0 {
				if abs(x-px) > 1 || abs(y-py) > 1 {
					t.Fatalf("satellite %d jumped at frame %d", i, a)
				}
				if (w > 0) != (pw > 0) {
					changes = append(changes, a)
				}
			}
			px, py, pw = x, y, w
		}
		for k := 1; k < len(changes); k++ {
			if changes[k]-changes[k-1] < 16 {
				t.Fatalf("satellite %d changed state at frames %d and %d", i, changes[k-1], changes[k])
			}
		}
	}
}
