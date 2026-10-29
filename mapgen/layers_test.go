package mapgen

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// groundFrame draws a frame's ground and town only, before detail, traffic and
// light, so each cell still shows what its own layer drew.
func groundFrame(age string, tick int) *frame {
	var gen Generator
	f := gen.frame(model(state(age, 7, tick)), 0)
	f.paintGround()
	f.paintTown()
	return f
}

func TestJoinPicksThePieceForItsNeighbours(t *testing.T) {
	for _, c := range []struct {
		mask int
		want rune
	}{{0b0101, '│'}, {0b1010, '─'}, {0b0011, '└'}, {0b1100, '┐'}, {0b0111, '├'}, {0b1111, '┼'}} {
		if got := joinRune(joinsLight, c.mask, 0); got != c.want {
			t.Errorf("mask %04b: got %c, want %c", c.mask, got, c.want)
		}
	}
}

// TestRoadsJoinUp: every laid road cell wears the piece that matches the road
// cells round it, and every river cell the piece that matches its water.
func TestRoadsJoinUp(t *testing.T) {
	for _, age := range []string{"bronze_age", "modern_age"} {
		f := groundFrame(age, noon)
		for y := 0; y < worldH; y++ {
			for x := 0; x < worldW; x++ {
				if !f.laid(x, y) || f.plan.Bridge[Pt{x, y}] {
					continue
				}
				joins := f.st.Roads[f.roadAt[at(x, y)]-1].Joins
				if utf8.RuneCountInString(joins) != 16 {
					continue
				}
				got := strings.IndexRune(joins, f.g.At(x, y).R)
				if got < 0 || neighbourMask(x, y, f.laid) != utf8.RuneCountInString(joins[:got]) {
					t.Fatalf("%s: road cell (%d,%d) shows %c for neighbours %04b", age, x, y, f.g.At(x, y).R, neighbourMask(x, y, f.laid))
				}
			}
		}
		wet := func(x, y int) bool { return !inWorld(x, y) || f.fields.Biome[at(x, y)].Water() }
		for y := 0; y < worldH; y++ {
			for x := 0; x < worldW; x++ {
				if f.fields.Biome[at(x, y)] != River || f.laid(x, y) {
					continue
				}
				want := joinRune(joinsRounded, neighbourMask(x, y, wet), 0)
				if f.g.At(x, y).R != want {
					t.Fatalf("%s: river cell (%d,%d) shows %c, want %c", age, x, y, f.g.At(x, y).R, want)
				}
			}
		}
	}
}

// TestScatterKeepsApart: detail lands only on open ground, and at most one
// piece to a small block of the world.
func TestScatterKeepsApart(t *testing.T) {
	for _, age := range []string{"bronze_age", "modern_age"} {
		f := groundFrame(age, noon)
		before := append([]Cell(nil), f.g.Cells...)
		f.paintDetail()
		perBlock := map[Pt]int{}
		for i, c := range f.g.Cells {
			if c == before[i] {
				continue
			}
			if f.busy[i] {
				t.Fatalf("%s: detail drawn on built ground at (%d,%d)", age, i%worldW, i/worldW)
			}
			perBlock[Pt{i % worldW / spotW, i / worldW / spotH}]++
		}
		if len(perBlock) < 50 {
			t.Errorf("%s: only %d blocks have any detail", age, len(perBlock))
		}
		for b, n := range perBlock {
			if n > 1 {
				t.Fatalf("%s: %d pieces of detail in block %v", age, n, b)
			}
		}
	}
}

// TestTrafficStaysOnRoads: every agent is on a laid road, and a frame draws
// the same whether or not the ones before it were drawn.
func TestTrafficStaysOnRoads(t *testing.T) {
	for _, age := range testAges {
		var gen Generator
		m := model(state(age, 7, noon))
		f := gen.frame(m, 100)
		f.draw()
		agents := 0
		for i := range f.g.Cells {
			if f.g.Things[f.g.Hit[i]].Kind == KindAgent {
				agents++
				if !f.laid(i%worldW, i/worldW) {
					t.Fatalf("%s: an agent at (%d,%d) is off the roads", age, i%worldW, i/worldW)
				}
			}
		}
		if agents < 3 {
			t.Errorf("%s: only %d agents", age, agents)
		}
		var warm Generator
		for frame := 0; frame < 100; frame++ {
			warm.Render(m, 120, 36, frame)
		}
		a, b := warm.Render(m, 120, 36, 100), (&Generator{}).Render(m, 120, 36, 100)
		for i := range a.Cells {
			if a.Cells[i] != b.Cells[i] {
				t.Fatalf("%s: frame 100 depends on the frames before it", age)
			}
		}
	}
}

// TestNightLightsUp: at midnight there are lit windows or lamps in the style's
// light, and at noon none.
func TestNightLightsUp(t *testing.T) {
	for _, age := range testAges {
		for _, c := range []struct {
			tick int
			lit  bool
		}{{noon, false}, {midnight, true}} {
			var gen Generator
			f := gen.frame(model(state(age, 7, c.tick)), 0)
			f.draw()
			lit := 0
			for _, cell := range f.g.Cells {
				if cell.Fg.Hue == f.st.Night.Light && cell.Dim == 0 {
					lit++
				}
			}
			if (lit > 0) != c.lit {
				t.Errorf("%s tick %d: %d lit cells", age, c.tick, lit)
			}
		}
	}
}

// TestRuinsLeaveRubble: a lost building leaves rubble on its parcel, and
// the inspect index says so.
func TestRuinsLeaveRubble(t *testing.T) {
	st := fixture.State(fixture.Options{Age: "modern_age", Seed: 7, Tick: noon, Ruins: true})
	g := (&Generator{}).Render(model(st), worldW, worldH, 0)
	rubble := 0
	for i := range g.Cells {
		if th := g.Things[g.Hit[i]]; th.Kind == KindBuilding && th.Ruin {
			rubble++
		}
	}
	if rubble == 0 {
		t.Fatal("no rubble in a town with ruins")
	}
}

// TestSpaceScalesDrawTheirBodies: each space age shows what its scale is made
// of (a system a civilization lives at still counts as a star).
func TestSpaceScalesDrawTheirBodies(t *testing.T) {
	for _, c := range []struct {
		age  string
		want map[string]int
	}{
		{"space_age", map[string]int{"home": 60, "space": 3000}},
		{"interstellar_age", map[string]int{"sun": 8, "planet": 6, "space": 3000}},
		{"galactic_age", map[string]int{"star": 20, "space": 3000}},
	} {
		g := (&Generator{}).Render(model(state(c.age, 4, noon)), worldW, worldH, 0)
		seen := map[string]int{}
		for _, h := range g.Hit {
			if th := g.Things[h]; th.Kind == KindTerrain {
				seen[th.Key]++
			} else if th.Kind == KindCiv {
				seen["star"]++
			}
		}
		for key, n := range c.want {
			if seen[key] < n {
				t.Errorf("%s: %d cells of %s, want %d: %v", c.age, seen[key], key, n, seen)
			}
		}
	}
}
