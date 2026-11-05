package mapgen

import (
	"reflect"
	"testing"
)

// spacePlans are the plans of the three space scales, with the ages whose
// wonders each holds a plot for.
var spacePlans = []struct {
	name string
	plan func(seed int64) *Plan
	ages []int
}{
	{"orbit", func(s int64) *Plan { return NewOrbitPlan(s, testCatalog) }, []int{17}},
	{"system", func(s int64) *Plan { return NewSystemPlan(s, testCatalog) }, []int{18}},
	{"stars", func(s int64) *Plan { return NewStarPlan(s, testCatalog) }, nil},
}

func TestSpacePlansDeterministic(t *testing.T) {
	for _, sp := range spacePlans {
		a, b := sp.plan(7), sp.plan(7)
		if sp.name == "stars" {
			if !reflect.DeepEqual(a.Sectors, b.Sectors) || !reflect.DeepEqual(a.Rivals, b.Rivals) || !reflect.DeepEqual(a.Rift, b.Rift) {
				t.Errorf("%s: same seed, different territory", sp.name)
			}
			continue
		}
		if len(a.Lots) != len(b.Lots) || len(a.Roads) != len(b.Roads) || a.Lots[0] != b.Lots[0] || a.Lots[len(a.Lots)-1] != b.Lots[len(b.Lots)-1] {
			t.Errorf("%s: same seed, different plan", sp.name)
		}
	}
}

// TestSpacePlansHoldTogether: every road cell is in the world and joins the
// hub, every parcel's door is a cell of its road, no parcel covers another or
// a road that is not its own ring's, and each scale has room for the wonders
// of its ages.
func TestSpacePlansHoldTogether(t *testing.T) {
	for _, sp := range spacePlans {
		for _, seed := range []int64{4, 7, 23} {
			p := sp.plan(seed)
			if len(p.Lots) < 60 && sp.name != "stars" {
				t.Errorf("%s/%d: only %d parcels", sp.name, seed, len(p.Lots))
			}
			inBounds := func(x, y int) bool {
				if sp.name == "system" {
					return x >= 0 && y >= 0 && x < SystemW && y < SystemH
				}
				return inWorld(x, y)
			}
			road := map[Pt]int{}
			for i, r := range p.Roads {
				for _, c := range r.Cells {
					if !inBounds(c.X, c.Y) {
						t.Fatalf("%s/%d: road %d leaves the world at %v", sp.name, seed, i, c)
					}
					road[c] = i
				}
			}
			taken := map[Pt]bool{}
			for _, l := range append(append([]Lot{}, p.Lots...), p.Plots...) {
				if l.Rect.W == 0 {
					continue
				}
				if p.Roads[l.Road].Cells[l.At] != l.Door {
					t.Fatalf("%s/%d: door %v is not cell %d of road %d", sp.name, seed, l.Door, l.At, l.Road)
				}
				l.Rect.cells(func(x, y int) {
					if taken[Pt{x, y}] {
						t.Fatalf("%s/%d: two parcels cover (%d,%d)", sp.name, seed, x, y)
					}
					taken[Pt{x, y}] = true
				})
			}
			reached := map[Pt]bool{p.Roads[0].Cells[0]: true}
			queue := []Pt{p.Roads[0].Cells[0]}
			for head := 0; head < len(queue); head++ {
				for _, d := range neighbours {
					n := Pt{queue[head].X + d[0], queue[head].Y + d[1]}
					if _, onRoad := road[n]; onRoad && !reached[n] {
						reached[n] = true
						queue = append(queue, n)
					}
				}
			}
			if len(reached) != len(road) {
				t.Errorf("%s/%d: %d road cells, %d reached from the hub", sp.name, seed, len(road), len(reached))
			}
			for _, age := range sp.ages {
				if p.Plots[age].Rect.W == 0 {
					t.Errorf("%s/%d: no plot for the wonder of age %d", sp.name, seed, age)
				}
			}
		}
	}
}

// TestRingParcelsStandOnTheirOwnRing: on the ring scales a parcel covers no
// road but its own ring.
func TestRingParcelsStandOnTheirOwnRing(t *testing.T) {
	for _, sp := range spacePlans[:2] {
		p := sp.plan(7)
		road := map[Pt]int{}
		for i, r := range p.Roads {
			for _, c := range r.Cells {
				road[c] = i
			}
		}
		for _, l := range p.Lots {
			l.Rect.cells(func(x, y int) {
				if i, onRoad := road[Pt{x, y}]; onRoad && (i < 2 || ringOf(i) != ringOf(l.Road)) {
					t.Fatalf("%s: a parcel on ring %d covers road %d at (%d,%d)", sp.name, ringOf(l.Road), i, x, y)
				}
			})
		}
	}
}

func TestHalfRingsAreClosedCurves(t *testing.T) {
	for _, ry := range []int{4, 7, 16} {
		east, west := halfRing(60, 40, ry, 1), halfRing(60, 40, ry, -1)
		for _, half := range [][]Pt{east, west} {
			for i := 1; i < len(half); i++ {
				if abs(half[i].X-half[i-1].X)+abs(half[i].Y-half[i-1].Y) != 1 {
					t.Fatalf("ry %d: cells %v and %v do not touch", ry, half[i-1], half[i])
				}
			}
			if half[0].Y != 40-ry || half[len(half)-1].Y != 40+ry {
				t.Errorf("ry %d: a half ring runs from row %d to %d, want %d to %d", ry, half[0].Y, half[len(half)-1].Y, 40-ry, 40+ry)
			}
		}
	}
}

func TestHomePlanetHighResolution(t *testing.T) {
	if discRX != 50 || discRY != 20 {
		t.Fatalf("discRX = %d, discRY = %d, want 50, 20", discRX, discRY)
	}

	st := state("space_age", 7, noon)
	m := model(st)
	g := (&Generator{}).Render(m, worldW, worldH, 0)

	minX, maxX := worldW, 0
	minY, maxY := worldH, 0
	quadrantCount := 0
	rimCyanCount := 0
	coastlineCount := 0
	polarCount := 0
	nightLightCount := 0

	quadrantSet := map[rune]bool{}
	for _, r := range quadrantRunes {
		if r != ' ' {
			quadrantSet[r] = true
		}
	}

	for y := 0; y < worldH; y++ {
		for x := 0; x < worldW; x++ {
			th := g.ThingAt(x, y)
			if th.Kind == KindTerrain && th.Key == "home" {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}

				cell := g.At(x, y)
				if quadrantSet[cell.R] {
					quadrantCount++
					if cell.Fg.Hue == HCyan {
						rimCyanCount++
					}
					if cell.Fg.Hue == HLand && cell.Bg.Hue == HSea {
						coastlineCount++
					}
				}
				if cell.R == '*' || cell.R == '░' {
					polarCount++
				}
				if x > worldW/2 && cell.Fg.Hue == HGold {
					nightLightCount++
				}
			}
		}
	}

	// Verify aspect ratio (width / height) ~ 2.5 (50:20 = 101:41)
	w := maxX - minX + 1
	h := maxY - minY + 1
	if w < 98 || w > 103 {
		t.Errorf("planet width = %d, want ~101", w)
	}
	if h < 39 || h > 43 {
		t.Errorf("planet height = %d, want ~41", h)
	}

	// Verify subpixel anti-aliasing features
	if rimCyanCount < 10 {
		t.Errorf("cyan anti-aliased rim count = %d, want >= 10", rimCyanCount)
	}
	if coastlineCount < 10 {
		t.Errorf("subpixel coastline count = %d, want >= 10", coastlineCount)
	}
	if polarCount < 10 {
		t.Errorf("polar ice cap cell count = %d, want >= 10", polarCount)
	}
	if nightLightCount < 2 {
		t.Errorf("night-side city lights count = %d, want >= 2", nightLightCount)
	}
}

func TestEarthSpinsWithAnim(t *testing.T) {
	st := state("space_age", 7, noon)
	m := model(st)
	gen := &Generator{}
	g0 := gen.Render(m, worldW, worldH, 0)
	g25 := gen.Render(m, worldW, worldH, 25)

	c := Pt{worldW / 2, worldH / 2}
	diffCount := 0
	interiorCount := 0

	for y := c.Y - discRY; y <= c.Y+discRY; y++ {
		for x := c.X - discRX; x <= c.X+discRX; x++ {
			th0 := g0.ThingAt(x, y)
			th25 := g25.ThingAt(x, y)
			if th0.Kind == KindTerrain && th0.Key == "home" && th25.Kind == KindTerrain && th25.Key == "home" {
				c0 := g0.At(x, y)
				c25 := g25.At(x, y)
				// Skip outer atmospheric rim cells (which have HCyan Fg)
				if c0.Fg.Hue == HCyan && c25.Fg.Hue == HCyan {
					continue
				}
				interiorCount++
				if c0.R != c25.R || c0.Fg != c25.Fg || c0.Bg != c25.Bg {
					diffCount++
				}
			}
		}
	}

	if interiorCount == 0 {
		t.Fatalf("no interior planet cells found")
	}
	if diffCount < 50 {
		t.Errorf("Earth did not rotate: only %d / %d cells changed between anim=0 and anim=25, want >= 50", diffCount, interiorCount)
	}
}

func TestSpaceCellExposedAndContinuous(t *testing.T) {
	gen := &Generator{}
	for _, tc := range []struct {
		age       string
		wantSpace bool
	}{
		{"stone_age", false},
		{"medieval_age", false},
		{"space_age", true},
		{"interstellar_age", true},
		{"galactic_age", true},
	} {
		m := model(state(tc.age, 7, noon))
		sc := gen.Draw(m, 0)
		if tc.wantSpace && sc.SpaceCell == nil {
			t.Errorf("%s: want SpaceCell != nil", tc.age)
		}
		if !tc.wantSpace && sc.SpaceCell != nil {
			t.Errorf("%s: want SpaceCell == nil", tc.age)
		}
		if sc.SpaceCell != nil {
			coords := [][2]int{
				{-100, -50},
				{-1, 0},
				{0, -1},
				{worldW, 50},
				{worldW + 100, worldH + 50},
			}
			for _, pt := range coords {
				cell := sc.SpaceCell(pt[0], pt[1])
				if cell.Bg.Hue != HVoid && cell.Bg.Hue != HNebula {
					t.Errorf("%s: cell at %v has unexpected hue %v", tc.age, pt, cell.Bg.Hue)
				}
			}
			if tc.age == "space_age" {
				for y := 0; y < 10; y++ {
					for x := 0; x < 10; x++ {
						if sc.Grid.ThingAt(x, y).Key == "space" {
							want := sc.Grid.At(x, y)
							got := sc.SpaceCell(x, y)
							if want.Bg != got.Bg || want.R != got.R || want.Fg != got.Fg {
								t.Errorf("at (%d,%d): want %+v, got %+v", x, y, want, got)
							}
						}
					}
				}
			}
		}
	}
}
