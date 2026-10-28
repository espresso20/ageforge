package mapgen

import (
	"reflect"
	"testing"

	"github.com/espresso20/ageforge/mapmodel"
)

// visitAges are one age of each scale a visit can come to: the land, the
// city, the orbit and the solar system. (The star map's ages schedule none:
// from the Galactic Age the aliens are traffic.)
var visitAges = []string{"bronze_age", "modern_age", "space_age", "interstellar_age"}

// visitSeeds are the worlds the visitor tests walk.
var visitSeeds = []int64{1, 3, 7}

// visitOf is a model's first visit at or after frame from.
func visitOf(m *mapmodel.Model, from int) mapmodel.Sighting {
	late := m.AgeIdx >= m.Catalog.SpaceAge()
	return mapmodel.NextSighting(m.Seed, late, from)
}

// isAlien says whether a visit is an alien's.
func isAlien(sg mapmodel.Sighting) bool { return sg.Roll>>16%alienOneIn == 0 }

// visitsOf lists the first n visits a model gets, in order.
func visitsOf(m *mapmodel.Model, n int) []mapmodel.Sighting {
	var out []mapmodel.Sighting
	for from := 0; len(out) < n; {
		sg := visitOf(m, from)
		out = append(out, sg)
		from = sg.Start + sg.Frames
	}
	return out
}

// visitorCells lists the cells a grid shows the visitor in.
func visitorCells(g *Grid) []Pt {
	var out []Pt
	for i := range g.Cells {
		if g.Things[g.Hit[i]].Kind == KindVisitor {
			out = append(out, Pt{i % g.W, i / g.W})
		}
	}
	return out
}

// firstOf finds, in a model's visits, the first of each form.
func firstOf(t *testing.T, m *mapmodel.Model) (visitor, alien mapmodel.Sighting) {
	t.Helper()
	var haveV, haveA bool
	for _, sg := range visitsOf(m, 40) {
		switch {
		case isAlien(sg) && !haveA:
			alien, haveA = sg, true
		case !isAlien(sg) && !haveV:
			visitor, haveV = sg, true
		}
	}
	if !haveV || !haveA {
		t.Fatalf("seed %d: forty visits held no visitor (%v) or no alien (%v)", m.Seed, haveV, haveA)
	}
	return visitor, alien
}

// TestVisitorDrawsAtEveryScale: through a visit, on the land, in the city, in
// orbit and in the solar system, the visitor is on the grid at some frame and
// the same grid has none before the visit starts or after it is over.
func TestVisitorDrawsAtEveryScale(t *testing.T) {
	for _, age := range visitAges {
		for _, seed := range visitSeeds {
			var gen Generator
			m := model(state(age, seed, noon))
			visitor, alien := firstOf(t, m)
			for form, sg := range map[string]mapmodel.Sighting{"visitor": visitor, "alien": alien} {
				seen := 0
				for pct := 0; pct < 100; pct += 5 {
					seen = max(seen, len(visitorCells(gen.DrawAt(m, 0, sg.Start+sg.Frames*pct/100).Grid)))
				}
				if seen == 0 {
					t.Errorf("%s seed %d: the %s is never on the grid", age, seed, form)
				}
				for _, outside := range []int{sg.Start - 1, sg.Start + sg.Frames, -1} {
					if cells := visitorCells(gen.DrawAt(m, 0, outside).Grid); len(cells) > 0 {
						t.Errorf("%s seed %d: the %s is out at frame %d, outside its visit", age, seed, form, outside)
					}
				}
			}
		}
	}
}

// TestNoVisitorFromTheGalacticAge: the star map's ages schedule none.
func TestNoVisitorFromTheGalacticAge(t *testing.T) {
	var gen Generator
	m := model(state("galactic_age", 3, noon))
	for _, sg := range visitsOf(model(state("space_age", 3, noon)), 5) {
		if cells := visitorCells(gen.DrawAt(m, 0, sg.Start+sg.Frames/2).Grid); len(cells) > 0 {
			t.Fatalf("a visitor in the Galactic Age at frame %d", sg.Start)
		}
	}
}

// TestVisitorIsDeterministic: a frame of a visit is the same from a fresh
// generator, from one that has drawn other frames, and the same twice.
func TestVisitorIsDeterministic(t *testing.T) {
	for _, age := range visitAges {
		m := model(state(age, 3, noon))
		visitor, alien := firstOf(t, m)
		var warm Generator
		for _, sg := range []mapmodel.Sighting{visitor, alien} {
			warm.DrawAt(m, 0, sg.Start+3)
			frame := sg.Start + sg.Frames/2
			a, b := warm.DrawAt(m, 7, frame).Grid, (&Generator{}).DrawAt(m, 7, frame).Grid
			if !reflect.DeepEqual(a.Cells, b.Cells) || !reflect.DeepEqual(a.Hit, b.Hit) {
				t.Fatalf("%s: frame %d of a visit depends on the frames before it", age, frame)
			}
		}
	}
}

// TestVisitorNeverCoversABuilding: through a visit, every cell of a building,
// a wonder or an ordinary house is what it is with no visitor out.
func TestVisitorNeverCoversABuilding(t *testing.T) {
	for _, age := range visitAges {
		var gen Generator
		m := model(state(age, 7, midnight))
		visitor, alien := firstOf(t, m)
		for _, sg := range []mapmodel.Sighting{visitor, alien} {
			for pct := 0; pct < 100; pct += 4 {
				base, with := gen.DrawAt(m, 0, -1).Grid, gen.DrawAt(m, 0, sg.Start+sg.Frames*pct/100).Grid
				for i := range base.Cells {
					th := base.Things[base.Hit[i]]
					if (th.Kind == KindBuilding || th.Kind == KindWonder || th.Kind == KindAmbient) &&
						(base.Cells[i] != with.Cells[i] || with.Things[with.Hit[i]] != th) {
						t.Fatalf("%s: the visitor covers the %v at (%d,%d) at %d%% of its visit", age, th.Kind, i%base.W, i/base.W, pct)
					}
				}
			}
		}
	}
}

// TestVisitorShinesAtNight: the visitor is lit whatever the hour, so the dark
// never dims it.
func TestVisitorShinesAtNight(t *testing.T) {
	var gen Generator
	m := model(state("modern_age", 3, midnight))
	_, alien := firstOf(t, m)
	g := gen.DrawAt(m, 0, alien.Start+alien.Frames/2).Grid
	cells := visitorCells(g)
	if len(cells) == 0 {
		t.Fatal("no visitor to look at")
	}
	for _, p := range cells {
		if g.At(p.X, p.Y).Dim != 0 {
			t.Errorf("the visitor at %v is dimmed by the night", p)
		}
	}
}

// TestVisitorArrivesAlongARoad: on the land a visitor starts at the far end
// of a trunk road, walks its cells in to stopAt from the square, and stays
// there; in space a ship starts on an edge of the map and flies in.
func TestVisitorArrivesAlongARoad(t *testing.T) {
	for _, age := range visitAges {
		var gen Generator
		m := model(state(age, 3, noon))
		visitor, _ := firstOf(t, m)
		first := visitorCells(gen.DrawAt(m, 0, visitor.Start).Grid)
		last := visitorCells(gen.DrawAt(m, 0, visitor.Start+visitor.Frames-1).Grid)
		if len(first) != 1 || len(last) != 1 {
			t.Fatalf("%s: a visitor is %d cells at the start and %d at the end of its visit", age, len(first), len(last))
		}
		f := gen.frame(m, 0)
		if f.st.Scale > City {
			if x := first[0].X; x != 0 && x != worldW-1 {
				t.Errorf("%s: the ship starts at %v, not on an edge of the map", age, first[0])
			}
			continue
		}
		onTrunk := func(p Pt, at int) bool {
			for i, road := range f.plan.Roads {
				if i > 0 && road.Parent == 0 && road.Tier == 0 && len(road.Cells) > stopAt+4 {
					if at < 0 && road.Cells[len(road.Cells)-1] == p || at >= 0 && road.Cells[at] == p {
						return true
					}
				}
			}
			return false
		}
		if !onTrunk(first[0], -1) {
			t.Errorf("%s: the traveller starts at %v, not at the far end of a trunk road", age, first[0])
		}
		if !onTrunk(last[0], stopAt) {
			t.Errorf("%s: the traveller ends at %v, not %d cells from the square", age, last[0], stopAt)
		}
		for pct := 0; pct < 100; pct += 3 {
			for _, p := range visitorCells(gen.DrawAt(m, 0, visitor.Start+visitor.Frames*pct/100).Grid) {
				if !f.plan.onRoad(p) {
					t.Fatalf("%s: the traveller is at %v at %d%%, off the trunk roads", age, p, pct)
				}
			}
		}
	}
}

// onRoad says whether a cell is on a road of the plan.
func (p *Plan) onRoad(c Pt) bool {
	for _, road := range p.Roads {
		for _, cell := range road.Cells {
			if cell == c {
				return true
			}
		}
	}
	return false
}

// TestAliensAreRarerStill: of all the visits, a minority are an alien's.
func TestAliensAreRarerStill(t *testing.T) {
	m := model(state("space_age", 5, noon))
	visits := visitsOf(m, 300)
	aliens := 0
	for _, sg := range visits {
		if isAlien(sg) {
			aliens++
		}
	}
	if aliens < len(visits)/5 || aliens > len(visits)*2/5 {
		t.Errorf("%d of %d visits are an alien's, want about one in %d", aliens, len(visits), alienOneIn)
	}
}

// TestSaucerHoversOverOpenGround: a hovering saucer stops on ground nothing is
// built or laid on, and a crossing one passes above the square.
func TestSaucerHoversOverOpenGround(t *testing.T) {
	var gen Generator
	m := model(state("modern_age", 3, noon))
	hovered := false
	for _, sg := range visitsOf(m, 60) {
		if !isAlien(sg) || sg.Kind == mapmodel.SightFlyby {
			continue
		}
		f := gen.frame(m, 0)
		f.draw()
		mid := f
		mid.visit = sg.Start + sg.Frames/2
		cells := mid.saucerCells(sg)
		for _, p := range cells {
			if mid.busy[at(p.X, p.Y)] {
				t.Fatalf("a saucer hangs over built ground at %v", p)
			}
		}
		hovered = true
		break
	}
	if !hovered {
		t.Skip("no hovering saucer among the visits of this seed")
	}
}
