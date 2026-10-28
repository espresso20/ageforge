package mapmodel_test

import (
	"strings"
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

// TestRecap: the since-last-visit diff names what was built, who was met,
// the new age and the war, and marks the new buildings.
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
	for _, b := range m.Buildings {
		if b.Delta > 0 {
			fresh++
		}
	}
	if fresh == 0 {
		t.Error("no building is marked new")
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

// TestModelShape: every age builds a sane model quickly (the budget is loose
// for -race).
func TestModelShape(t *testing.T) {
	b := mapmodel.NewBuilder(cat)
	for _, age := range config.AgeOrder() {
		st := fixture.State(fixture.Options{Age: age, Seed: 7, Harbinger: true, Wars: 1})
		t0 := time.Now()
		m := b.Build(&st, nil)
		el := time.Since(t0)
		if el > 250*time.Millisecond {
			t.Errorf("%s: model took %v", age, el)
		}
		if m.TownName == "" || len(m.Buildings) == 0 {
			t.Errorf("%s: no town name or no buildings", age)
		}
		if m.Harbinger == nil || m.Catastrophe.Pressure <= 0 {
			t.Errorf("%s: harbinger pressure missing", age)
		}
	}
}

// TestFlows: the fixture's draining stocks, idle workers and understaffed
// buildings all show up, and the verdict never speaks of a full store: stock
// has no ceiling.
func TestFlows(t *testing.T) {
	m := build(t, fixture.State(fixture.Options{Age: "renaissance_age", Seed: 7, Idle: 12}), nil)
	f := m.Flows
	if strings.Contains(f.Worst, "full") || f.IdleWorkers != 12 || len(f.Understaffed)+len(f.Idle) == 0 || f.Worst == "" {
		t.Fatalf("flows: draining %d idle %d under %d worst %q", len(f.Draining), f.IdleWorkers, len(f.Understaffed), f.Worst)
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
	c := mapmodel.ClockAt(0)
	if c.Hour != 9 || c.Phase != "day" {
		t.Errorf("tick 0 is %s (%s), want 09:00 day", c, c.Phase)
	}
}
