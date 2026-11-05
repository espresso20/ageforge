package mapgen

import "testing"

// far-apart bodies on a diagonal and a cross, in 16ths of a pixel
var testBodies = []shipBody{
	{20 * pxUnit, 20 * pxUnit, 3 * pxUnit}, {150 * pxUnit, 90 * pxUnit, 3 * pxUnit},
	{20 * pxUnit, 90 * pxUnit, 3 * pxUnit}, {150 * pxUnit, 20 * pxUnit, 3 * pxUnit},
}

func TestShipsMoveAPixelAtMost(t *testing.T) {
	diagonal := false
	for seed := int64(1); seed <= 8; seed++ {
		for i := 0; i < shipCount(seed); i++ {
			var px, py int
			for anim := 0; anim < 4000; anim++ {
				_, x, y, hx, hy, ok := shipPose(seed, anim, i, testBodies)
				if !ok {
					break
				}
				x, y = floorDiv(x, pxUnit), floorDiv(y, pxUnit)
				if anim > 0 && (abs(x-px) > 1 || abs(y-py) > 1) {
					t.Fatalf("seed %d ship %d anim %d: jumped (%d,%d) to (%d,%d)", seed, i, anim, px, py, x, y)
				}
				if anim > 0 && x != px && y != py {
					diagonal = true
				}
				if hx == 0 && hy == 0 {
					t.Fatalf("seed %d ship %d: no heading", seed, i)
				}
				px, py = x, y
			}
		}
	}
	if !diagonal {
		t.Error("no ship ever stepped in x and y together")
	}
}

func TestShipsNeverOverlayBodies(t *testing.T) {
	for seed := int64(1); seed <= 6; seed++ {
		for anim := 0; anim < 3000; anim += 37 {
			p := NewPixels(90, 50)
			bodies := []shipBody{{30 * pxUnit, 30 * pxUnit, 40 * pxUnit}, {120 * pxUnit, 60 * pxUnit, 30 * pxUnit}, {40 * pxUnit, 80 * pxUnit, 20 * pxUnit}}
			sun := []shipBody{{90 * pxUnit, 50 * pxUnit, 35 * pxUnit}}
			p.Ships(seed, anim, bodies, sun, nil)
			for y := 0; y < p.H; y++ {
				for x := 0; x < p.W; x++ {
					if p.px[y*p.W+x].Hue != HNone && (coveredBy(x, y, bodies) || coveredBy(x, y, sun)) {
						t.Fatalf("seed %d anim %d: ship pixel (%d,%d) over a body", seed, anim, x, y)
					}
				}
			}
			p.Release()
		}
	}
}

func TestShipsSameSeedSameFrame(t *testing.T) {
	draw := func(seed int64) []Paint {
		p := NewPixels(90, 50)
		p.Ships(seed, 500, testBodies, nil, nil)
		out := append([]Paint(nil), p.px...)
		p.Release()
		return out
	}
	a, b, other := draw(3), draw(3), draw(4)
	same, differs, drawn := true, false, 0
	for i := range a {
		same = same && a[i] == b[i]
		differs = differs || a[i] != other[i]
		if a[i].Hue != HNone {
			drawn++
		}
	}
	if !same || !differs || drawn == 0 || drawn > 4*5 {
		t.Errorf("same=%v differs=%v drawn=%d", same, differs, drawn)
	}
}

// A ship in the Sol system is found from its pixels, as one of the traffic
// the cursor already has words for.
func TestOverviewShipsAreNamed(t *testing.T) {
	gen := &Generator{}
	named := map[string]bool{"shuttle": true, "cargo_hauler": true, "generation_ship": true}
	found := 0
	for _, anim := range []int{0, 300, 900, 1500} {
		sc := gen.Draw(model(state("interstellar_age", 7, noon)), anim)
		ov := sc.Overview(120, 40)
		for y := 0; y < ov.Grid.H; y++ {
			for x := 0; x < ov.Grid.W; x++ {
				if th := ov.Grid.ThingAt(x, y); th.Kind == KindAgent {
					if !named[th.Key] {
						t.Fatalf("ship named %q", th.Key)
					}
					found++
				}
			}
		}
	}
	if found == 0 {
		t.Error("no ship found in the overview")
	}
}

// The Space Age overview flies pixel ships, and the Sol-system warp gate
// stands only once the Warp Nexus is built.
func TestOrbitShipsAndGate(t *testing.T) {
	m := model(state("space_age", 7, noon))
	sc := (&Generator{}).Draw(m, 100)
	ships := 0
	for _, h := range sc.Grid.Hit {
		if sc.Grid.Things[h].Kind == KindAgent {
			ships++
		}
	}
	if ships == 0 {
		t.Error("space age: no pixel ships")
	}
	for _, built := range []bool{false, true} {
		sm := model(state("interstellar_age", 7, noon))
		for i := range sm.Wonders {
			if sm.Wonders[i].Key == gateBuilding {
				sm.Wonders[i].Built = built
			}
		}
		ov := (&Generator{}).Draw(sm, 40).Overview(120, 40)
		gate := false
		for _, h := range ov.Grid.Hit {
			gate = gate || ov.Grid.Things[h].Key == "warp_gate"
		}
		if gate != built {
			t.Errorf("warp nexus built=%v: gate drawn=%v", built, gate)
		}
	}
}
