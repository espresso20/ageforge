package mapgen

import (
	"reflect"
	"slices"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

// aroundAges are an age of each scale.
var aroundAges = []string{"medieval_age", "modern_age", "space_age", "interstellar_age", "galactic_age"}

// busyWorld is a model with a civ at war, routes, a harbinger, idle workers and
// both kinds of party out.
func busyWorld(age string, seed int64) *mapmodel.Model {
	st := fixture.State(fixture.Options{Age: age, Seed: seed, Tick: noon, Wars: 1, Harbinger: true, Idle: 14, Routes: 4})
	st.Military.ActiveScout = &game.ExpeditionSnapshot{Name: "Scout the hills", TicksLeft: 40}
	st.Military.ActiveMilitary = &game.ExpeditionSnapshot{Name: "March on the raiders", TicksLeft: 25}
	return model(st)
}

// quietWorld is the same town with none of it: no civ met, no route, nobody
// out, no harbinger and no idle hands.
func quietWorld(m *mapmodel.Model) *mapmodel.Model {
	q := *m
	q.Factions = slices.Clone(m.Factions)
	for i := range q.Factions {
		q.Factions[i].Discovered = false
	}
	q.Routes, q.Expeditions, q.Harbinger = nil, mapmodel.Expeditions{}, nil
	q.Workers.Idle = 0
	return &q
}

// kindsOf counts the cells of each kind of the world round the town.
func kindsOf(g *Grid) map[Kind]int {
	out := map[Kind]int{}
	for _, h := range g.Hit {
		switch k := g.Things[h].Kind; k {
		case KindCiv, KindRoute, KindMission, KindHarbinger, KindIdle:
			out[k]++
		}
	}
	return out
}

// TestWorldRoundTheTownDrawsAtEveryScale: on the land, in the city, in orbit,
// in the solar system and on the star map, every part of it is drawn.
func TestWorldRoundTheTownDrawsAtEveryScale(t *testing.T) {
	for _, age := range aroundAges {
		for _, seed := range []int64{4, 6} {
			m := busyWorld(age, seed)
			g := (&Generator{}).DrawAt(m, 9, -1).Grid
			kinds := kindsOf(g)
			for _, k := range []Kind{KindCiv, KindRoute, KindMission, KindHarbinger, KindIdle} {
				if kinds[k] == 0 {
					t.Errorf("%s seed %d: nothing of kind %d is drawn", age, seed, k)
				}
			}
			civs := map[string]bool{}
			for _, h := range g.Hit {
				if th := g.Things[h]; th.Kind == KindCiv {
					civs[th.Key] = true
				}
			}
			if met := m.DiscoveredCount(); len(civs) != met {
				t.Errorf("%s seed %d: %d civs are marked, %d are met", age, seed, len(civs), met)
			}
		}
	}
}

// TestNothingIsDrawnForACivNotMet: with no civ met and nothing out, none of the
// world round the town shows.
func TestNothingIsDrawnForACivNotMet(t *testing.T) {
	for _, age := range aroundAges {
		g := (&Generator{}).DrawAt(quietWorld(busyWorld(age, 4)), 9, -1).Grid
		if kinds := kindsOf(g); len(kinds) > 0 {
			t.Errorf("%s: a quiet town draws %v", age, kinds)
		}
	}
}

// TestCivsAreColouredByRelation: a site wears the colour of how the civ stands
// toward you, and sits on the border of the picture (a node, on the star map).
func TestCivsAreColouredByRelation(t *testing.T) {
	for _, age := range aroundAges {
		var gen Generator
		m := busyWorld(age, 4)
		sc := gen.DrawAt(m, 0, -1)
		f := gen.frame(m, 0)
		f.draw()
		for _, s := range f.sites {
			var cell Cell
			for dx := -1; dx <= 1; dx++ {
				if th := sc.Grid.ThingAt(s.at.X+dx, s.at.Y); th.Kind == KindCiv && th.Key == s.civ.Key {
					cell = sc.Grid.At(s.at.X+dx, s.at.Y)
				}
			}
			if want := relationHue[s.civ.Relation]; cell.Fg.Hue != want {
				t.Errorf("%s: %s (%v) is drawn in hue %d, want %d", age, s.civ.Key, s.civ.Relation, cell.Fg.Hue, want)
			}
			if f.st.Scale == Stars {
				if !slices.Contains(f.plan.Nodes, s.at) {
					t.Errorf("%s: %s stands at %v, not at a system", age, s.civ.Key, s.at)
				}
				continue
			}
			e := f.edge()
			if s.at.X < e.X || s.at.X >= e.X+e.W || s.at.Y < e.Y || s.at.Y >= e.Y+e.H ||
				s.at.X > e.X+1 && s.at.X < e.X+e.W-2 && s.at.Y != e.Y && s.at.Y != e.Y+e.H-1 {
				t.Errorf("%s: %s stands at %v, off the border %v", age, s.civ.Key, s.at, e)
			}
		}
		if len(f.sites) == 0 {
			t.Fatalf("%s: no civ is met", age)
		}
		wantBounds := f.edge()
		if f.st.Scale != Stars && sc.Bounds != wantBounds {
			t.Errorf("%s: the scene reaches %v, not the edge %v", age, sc.Bounds, wantBounds)
		}
		for _, s := range f.sites {
			if !slices.Contains(sc.Targets, s.at) {
				t.Errorf("%s: the cursor cannot visit %s", age, s.civ.Key)
			}
		}
	}
}

// TestRoutesRunToTheirPartners: each route has a line of dots toward its
// partner and caravans on it, a disrupted one a stalled caravan.
func TestRoutesRunToTheirPartners(t *testing.T) {
	for _, age := range aroundAges {
		var gen Generator
		m := busyWorld(age, 4)
		g := gen.DrawAt(m, 17, -1).Grid
		for _, rt := range m.Routes {
			if rt.Civ == "" {
				continue
			}
			dots, vans := 0, 0
			for i, h := range g.Hit {
				if th := g.Things[h]; th.Kind == KindRoute && th.Key == rt.Civ {
					if g.Cells[i].R == trailGlyph[gen.frame(m, 0).st.Scale] {
						dots++
					} else {
						vans++
					}
				}
			}
			if scale := gen.frame(m, 0).st.Scale; vans == 0 || dots < 3 && scale != Stars {
				t.Errorf("%s: the route to %s has %d dots and %d caravans", age, rt.Civ, dots, vans)
			}
		}
		if len(m.Routes) == 0 {
			t.Fatalf("%s: no routes to draw", age)
		}
	}
}

// TestDisruptedRouteStallsItsCaravan: a disrupted route carries one caravan
// that does not move.
func TestDisruptedRouteStallsItsCaravan(t *testing.T) {
	m := busyWorld("modern_age", 4)
	m.Routes = slices.Clone(m.Routes)
	for i := range m.Routes {
		m.Routes[i].Disrupted = true
	}
	var gen Generator
	at := func(anim int) map[Pt]bool {
		out := map[Pt]bool{}
		g := gen.DrawAt(m, anim, -1).Grid
		for i, h := range g.Hit {
			if th := g.Things[h]; th.Kind == KindRoute && g.Cells[i].R == '&' || th.Kind == KindRoute && g.Cells[i].R == '■' || th.Kind == KindRoute && g.Cells[i].R == '◆' {
				out[Pt{i % g.W, i / g.W}] = true
			}
		}
		return out
	}
	a, b := at(0), at(80)
	if len(a) == 0 || !reflect.DeepEqual(a, b) {
		t.Errorf("stalled caravans moved or are missing: %v then %v", a, b)
	}
}

// TestWorldRoundTheTownMovesAndIsDeterministic: caravans and parties are on
// different cells later, and a frame is the same from a fresh generator.
func TestWorldRoundTheTownMovesAndIsDeterministic(t *testing.T) {
	for _, age := range aroundAges {
		m := busyWorld(age, 4)
		var warm Generator
		for frame := 0; frame < 30; frame++ {
			warm.DrawAt(m, frame, -1)
		}
		a, b := warm.DrawAt(m, 30, -1).Grid, (&Generator{}).DrawAt(m, 30, -1).Grid
		if !reflect.DeepEqual(a.Cells, b.Cells) || !reflect.DeepEqual(a.Hit, b.Hit) {
			t.Fatalf("%s: frame 30 depends on the frames before it", age)
		}
		later := (&Generator{}).DrawAt(m, 60, -1).Grid
		moved := false
		for i := range a.Hit {
			if k := a.Things[a.Hit[i]].Kind; (k == KindRoute || k == KindMission) && a.Cells[i] != later.Cells[i] {
				moved = true
			}
		}
		if !moved {
			t.Errorf("%s: nothing on the routes moved in thirty frames", age)
		}
	}
}

// TestWorldRoundTheTownNeverCoversABuilding: every cell of a building, a wonder
// or an ordinary house is what it is in the same town with none of the world
// round it.
func TestWorldRoundTheTownNeverCoversABuilding(t *testing.T) {
	for _, age := range aroundAges {
		for _, tick := range []int{noon, midnight} {
			m := busyWorld(age, 6)
			m.Clock = mapmodel.ClockAt(tick)
			var gen Generator
			base, with := gen.DrawAt(quietWorld(m), 12, -1).Grid, gen.DrawAt(m, 12, -1).Grid
			for i := range base.Cells {
				th := base.Things[base.Hit[i]]
				if (th.Kind == KindBuilding || th.Kind == KindWonder || th.Kind == KindAmbient) &&
					(base.Cells[i] != with.Cells[i] || with.Things[with.Hit[i]] != th) {
					t.Fatalf("%s: the world covers the %v at (%d,%d)", age, th.Kind, i%base.W, i/base.W)
				}
			}
		}
	}
}

// TestIdleWorkersLoiterInProportion: one for six idle, nine at most, none with
// nobody idle, and the harbinger stands apart from them.
func TestIdleWorkersLoiterInProportion(t *testing.T) {
	for _, idle := range []int{0, 1, 6, 12, 30, 200} {
		st := fixture.State(fixture.Options{Age: "modern_age", Seed: 4, Tick: noon, Idle: idle, Harbinger: true})
		m := model(st)
		g := (&Generator{}).DrawAt(m, 3, -1).Grid
		figures, harbingers := 0, 0
		for _, h := range g.Hit {
			switch g.Things[h].Kind {
			case KindIdle:
				figures++
			case KindHarbinger:
				harbingers++
			}
		}
		want := 0
		if m.Workers.Idle > 0 {
			want = min(9, 1+m.Workers.Idle/6)
		}
		if figures != want || harbingers != 1 {
			t.Errorf("%d idle: %d figures and %d harbingers, want %d and 1", m.Workers.Idle, figures, harbingers, want)
		}
	}
}

// TestPartiesGoOutAndBack: a scouting party leaves the square along a way and
// comes back to where it started.
func TestPartiesGoOutAndBack(t *testing.T) {
	for _, age := range aroundAges {
		var gen Generator
		m := busyWorld(age, 4)
		f := gen.frame(m, 0)
		f.draw()
		to, node := f.frontier("Scout the hills")
		cells := f.journey(to, node)
		if len(cells) < 4 {
			t.Fatalf("%s: the way out is %d cells", age, len(cells))
		}
		first, _, _ := pace(cells, 0, 0, 3)
		far, _, _ := pace(cells, 0, 3*(len(cells)-1), 3)
		back, _, _ := pace(cells, 0, 3*2*(len(cells)-1), 3)
		if first != cells[0] || far != cells[len(cells)-1] || back != first {
			t.Errorf("%s: the party is at %v, then %v, then %v on a way from %v to %v", age, first, far, back, cells[0], cells[len(cells)-1])
		}
	}
}
