package mapmodel_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/mapmodel"
	"github.com/espresso20/ageforge/mapmodel/fixture"
)

var cat = mapmodel.NewCatalog()

func build(t *testing.T, st game.GameState, since *mapmodel.Visit) *mapmodel.Model {
	t.Helper()
	return mapmodel.NewBuilder(cat).Build(&st, since)
}

// TestDeterministic: the same snapshot builds the same model, from a fresh
// builder or a warm one; another seed builds a different town.
func TestDeterministic(t *testing.T) {
	for _, age := range []string{"primitive_age", "medieval_age", "cyberpunk_age", "transcendent_age"} {
		st := fixture.State(fixture.Options{Age: age, Seed: 7})
		a := build(t, st, nil).Fingerprint()
		b := mapmodel.NewBuilder(cat)
		b.Build(&st, nil)
		if got := b.Build(&st, nil).Fingerprint(); got != a {
			t.Errorf("%s: warm builder fingerprint %s, fresh %s", age, got, a)
		}
		st2 := fixture.State(fixture.Options{Age: age, Seed: 11})
		if build(t, st2, nil).Fingerprint() == a {
			t.Errorf("%s: seeds 7 and 11 give the same model", age)
		}
	}
}

type spot struct {
	key string
	n   int
}

// TestPlacementNeverMoves: growing a settlement (more of this age's types,
// new types) never moves a town tile or a skyline silhouette that was
// already there. Town tiles may change which of their lineage's types they
// show (an in-place re-skin), never where they are; the occupied lots only
// grow.
func TestPlacementNeverMoves(t *testing.T) {
	for _, age := range []string{"primitive_age", "bronze_age", "medieval_age", "industrial_age", "digital_age", "galactic_age"} {
		for _, seed := range []int64{3, 7} {
			st := fixture.State(fixture.Options{Age: age, Seed: seed, Scale: 0.5})
			b := mapmodel.NewBuilder(cat)
			prev := b.Build(&st, nil)
			for step := 1; step <= 6; step++ {
				st = fixture.Grow(st, step)
				cur := b.Build(&st, nil)
				checkTown(t, fmt.Sprintf("%s/%d step %d", age, seed, step), prev, cur)
				checkSkyline(t, fmt.Sprintf("%s/%d step %d", age, seed, step), prev, cur)
				prev = cur
			}
		}
	}
}

func checkTown(t *testing.T, name string, prev, cur *mapmodel.Model) {
	t.Helper()
	curAt := map[mapmodel.Pt]mapmodel.TownTile{}
	for _, tt := range cur.Town.Tiles {
		curAt[tt.Pt] = tt
	}
	for _, tt := range prev.Town.Tiles {
		c, ok := curAt[tt.Pt]
		if !ok {
			t.Fatalf("%s: town tile of %s at %v vanished", name, tt.Key, tt.Pt)
		}
		if c.Lineage != tt.Lineage || c.Ord != tt.Ord {
			t.Fatalf("%s: tile at %v was %s #%d, now %s #%d", name, tt.Pt, tt.Lineage, tt.Ord, c.Lineage, c.Ord)
		}
	}
	for _, w := range prev.Town.Wonders {
		found := false
		for _, c := range cur.Town.Wonders {
			if c.Key == w.Key && c.At == w.At {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s: wonder %s moved or vanished", name, w.Key)
		}
	}
}

func checkSkyline(t *testing.T, name string, prev, cur *mapmodel.Model) {
	t.Helper()
	type id struct {
		key  string
		copy int
	}
	at := map[id]mapmodel.Lot{}
	for _, l := range cur.Skyline.Lots {
		at[id{l.Key, l.Copy}] = l
	}
	for _, l := range prev.Skyline.Lots {
		c, ok := at[id{l.Key, l.Copy}]
		if !ok {
			t.Fatalf("%s: skyline lot %s#%d vanished", name, l.Key, l.Copy)
		}
		if c.X != l.X || c.Row != l.Row || c.Seed != l.Seed {
			t.Fatalf("%s: skyline lot %s#%d moved from %d/%d to %d/%d", name, l.Key, l.Copy, l.X, l.Row, c.X, c.Row)
		}
	}
	for i, d := range prev.Skyline.Districts {
		if cur.Skyline.Districts[i] != d {
			t.Fatalf("%s: district %d changed", name, i)
		}
	}
}

// TestUpgradeReskinsInPlace: upgrading copies from a legacy tier to the
// current one keeps the lineage's lots exactly where they were.
func TestUpgradeReskinsInPlace(t *testing.T) {
	st := fixture.State(fixture.Options{Age: "iron_age", Seed: 5})
	var legacy, current string
	for k, bs := range st.Buildings {
		d := cat.Defs[k]
		if d.Lineage != mapmodel.LinHousing {
			continue
		}
		if bs.IsLegacy && bs.Count > 2 && (legacy == "" || k < legacy) {
			legacy = k
		}
		if !bs.IsLegacy && bs.Count > 0 && (current == "" || k < current) {
			current = k
		}
	}
	if legacy == "" || current == "" {
		t.Skip("fixture has no legacy and current housing pair")
	}
	before := build(t, st, nil)
	l, c := st.Buildings[legacy], st.Buildings[current]
	l.Count -= 2
	c.Count += 2
	st.Buildings[legacy], st.Buildings[current] = l, c
	after := build(t, st, nil)
	pts := func(m *mapmodel.Model) map[mapmodel.Pt]bool {
		out := map[mapmodel.Pt]bool{}
		for _, tt := range m.Town.Tiles {
			if tt.Lineage == mapmodel.LinHousing {
				out[tt.Pt] = true
			}
		}
		return out
	}
	a, b := pts(before), pts(after)
	if len(a) != len(b) {
		t.Fatalf("housing tiles %d before the upgrade, %d after", len(a), len(b))
	}
	for p := range a {
		if !b[p] {
			t.Fatalf("housing lot %v vacated by an upgrade", p)
		}
	}
}

// TestRecap: the since-last-visit diff names what was built, who was met,
// the new age and the war, and marks the new tiles.
func TestRecap(t *testing.T) {
	old := fixture.State(fixture.Options{Age: "bronze_age", Seed: 7})
	oldM := build(t, old, nil)
	visit := mapmodel.VisitOf(oldM)

	st := fixture.State(fixture.Options{Age: "iron_age", Seed: 7, Wars: 1})
	st.Tick = old.Tick + 1800
	m := build(t, st, visit)
	r := m.Recap
	if !r.HasBaseline || r.Since != 1800 {
		t.Fatalf("recap baseline %v since %d", r.HasBaseline, r.Since)
	}
	if r.NewCount == 0 || len(r.Built) == 0 {
		t.Fatal("recap lists nothing built")
	}
	kinds := map[mapmodel.NewsKind]bool{}
	for _, it := range r.Items {
		kinds[it.Kind] = true
	}
	for _, k := range []mapmodel.NewsKind{mapmodel.NewsAge, mapmodel.NewsBuilt, mapmodel.NewsWar} {
		if !kinds[k] {
			t.Errorf("recap has no item of kind %d: %+v", k, r.Items)
		}
	}
	h := r.Headline(80)
	if h == "" || len([]rune(h)) > 80 {
		t.Errorf("headline %q", h)
	}
	fresh := 0
	for _, tt := range m.Town.Tiles {
		if tt.Fresh {
			fresh++
		}
	}
	if fresh == 0 {
		t.Error("no town tile is marked new")
	}
	newLots := 0
	for _, l := range m.Skyline.Lots {
		if l.New {
			newLots++
		}
	}
	if newLots == 0 {
		t.Error("no skyline lot is marked new")
	}
	// No baseline, no recap (a pending catastrophe still makes the news).
	if got := build(t, st, nil).Recap; got.HasBaseline || len(got.Items) != 0 {
		t.Errorf("recap without a baseline: %+v", got)
	}
}

// TestSessionMarkRecap: a baseline from the engine's load-time mark works
// the same as one taken from a model.
func TestSessionMarkRecap(t *testing.T) {
	st := fixture.State(fixture.Options{Age: "stone_age", Seed: 2})
	mark := &game.SessionMark{Tick: st.Tick - 900, Age: "stone_age", Buildings: map[string]int{}, Civs: map[string]string{}}
	for k, bs := range st.Buildings {
		if bs.Count > 1 {
			mark.Buildings[k] = bs.Count - 1
		}
	}
	m := build(t, st, mapmodel.VisitFromSession(mark))
	if m.Recap.NewCount == 0 || m.Recap.Since != 900 {
		t.Fatalf("session recap: %+v", m.Recap)
	}
	if mapmodel.VisitFromSession(nil) != nil {
		t.Fatal("a nil mark must give no baseline")
	}
}

// TestModelShape: every age builds a sane model quickly (once the seed's
// world and plan are cached; the budget is loose for -race).
func TestModelShape(t *testing.T) {
	b := mapmodel.NewBuilder(cat)
	warm := fixture.State(fixture.Options{Age: "primitive_age", Seed: 7})
	b.Build(&warm, nil)
	for _, age := range config.AgeOrder() {
		st := fixture.State(fixture.Options{Age: age, Seed: 7, Harbinger: true, Wars: 1})
		t0 := time.Now()
		m := b.Build(&st, nil)
		el := time.Since(t0)
		if el > 250*time.Millisecond {
			t.Errorf("%s: model took %v", age, el)
		}
		if len(m.Skyline.Districts) != m.AgeIdx+1 {
			t.Errorf("%s: %d districts", age, len(m.Skyline.Districts))
		}
		if len(m.Town.Tiles) == 0 || len(m.Skyline.Lots) == 0 {
			t.Errorf("%s: nothing placed", age)
		}
		seen := map[mapmodel.Pt]bool{}
		for _, tt := range m.Town.Tiles {
			if seen[tt.Pt] {
				t.Fatalf("%s: two tiles on %v", age, tt.Pt)
			}
			seen[tt.Pt] = true
			if m.Town.World.At(tt.X, tt.Y).Water() {
				t.Fatalf("%s: tile of %s in the water", age, tt.Key)
			}
		}
		for _, l := range m.Lineages {
			placed := 0
			for _, tt := range m.Town.Tiles {
				if tt.Lineage == l.Key {
					placed++
				}
			}
			if placed != l.Tiles {
				t.Errorf("%s: lineage %s wants %d tiles, placed %d", age, l.Key, l.Tiles, placed)
			}
		}
		if m.Harbinger == nil || m.Catastrophe.Pressure <= 0 {
			t.Errorf("%s: harbinger pressure missing", age)
		}
		if m.Activity.Traffic <= 0 || m.Activity.Traffic > 1 {
			t.Errorf("%s: traffic %v", age, m.Activity.Traffic)
		}
	}
}

// TestTrafficScalesWithRoutes: more trade routes, more traffic.
func TestTrafficScalesWithRoutes(t *testing.T) {
	few := build(t, fixture.State(fixture.Options{Age: "industrial_age", Seed: 7, Routes: -1}), nil)
	many := build(t, fixture.State(fixture.Options{Age: "industrial_age", Seed: 7, Routes: 5}), nil)
	if many.Activity.Routes <= few.Activity.Routes || many.Activity.Traffic <= few.Activity.Traffic {
		t.Fatalf("routes %d→%d traffic %v→%v", few.Activity.Routes, many.Activity.Routes, few.Activity.Traffic, many.Activity.Traffic)
	}
}

// TestFlows: the fixture's full and draining stores, idle workers and
// understaffed buildings all show up.
func TestFlows(t *testing.T) {
	m := build(t, fixture.State(fixture.Options{Age: "renaissance_age", Seed: 7, Idle: 12}), nil)
	f := m.Flows
	if len(f.Full) == 0 || f.IdleWorkers != 12 || len(f.Understaffed)+len(f.Idle) == 0 || f.Worst == "" {
		t.Fatalf("flows: full %d draining %d idle %d under %d worst %q", len(f.Full), len(f.Draining), f.IdleWorkers, len(f.Understaffed), f.Worst)
	}
	sum := 0.0
	for _, s := range f.Shares {
		sum += s.Share
	}
	if sum < 0.999 || sum > 1.001 {
		t.Fatalf("lineage shares sum to %v", sum)
	}
}

// TestLayoutKey: the clock moving keeps the layout key; a new building
// changes it.
func TestLayoutKey(t *testing.T) {
	st := fixture.State(fixture.Options{Age: "medieval_age", Seed: 7})
	a := build(t, st, nil)
	st.Tick += 450
	if b := build(t, st, nil); a.LayoutKey != b.LayoutKey {
		t.Error("a tick changed the layout key")
	}
	if c := build(t, fixture.Grow(st, 1), nil); c.LayoutKey == a.LayoutKey {
		t.Error("new buildings kept the layout key")
	}
}

// TestMaths: the deterministic helpers are accurate enough.
func TestMaths(t *testing.T) {
	for _, c := range []struct{ turns, want float64 }{{0, 0}, {0.25, 1}, {0.5, 0}, {0.75, -1}, {1.0 / 12, 0.5}} {
		if got := mapmodel.Sin(c.turns); got < c.want-1e-6 || got > c.want+1e-6 {
			t.Errorf("Sin(%v) = %v, want %v", c.turns, got, c.want)
		}
	}
	for _, c := range []struct{ x, want float64 }{{1, 0}, {2, 1}, {8, 3}, {6, 2.584962500721156}} {
		if got := mapmodel.Log2(c.x); got < c.want-1e-9 || got > c.want+1e-9 {
			t.Errorf("Log2(%v) = %v, want %v", c.x, got, c.want)
		}
	}
	c := mapmodel.ClockAt(0)
	if c.Hour != 9 || c.Phase != "day" {
		t.Errorf("tick 0 is %s (%s), want 09:00 day", c, c.Phase)
	}
	if n := mapmodel.ClockAtTOD(3, 0.0); n.Phase != "night" || n.Day != 3 {
		t.Errorf("midnight clock %+v", n)
	}
}
