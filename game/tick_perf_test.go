package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// newLateGameEngine builds a deterministic mid/late-game engine for tick
// benchmarks: industrial age, every building up to that age owned several
// times over (including the harbour lineage), resources topped up, and every
// age-eligible trade route running. The point is to exercise the per-tick
// paths that scale with owned buildings and active routes.
func newLateGameEngine(tb testing.TB) *GameEngine {
	tb.Helper()
	ge := NewGameEngine()
	ge.SeedRNG(42)
	ge.age = "industrial_age"
	ageOrder := ge.progress.GetAgeOrder()

	for key, def := range ge.Buildings.defs {
		if ageOrder[def.RequiredAge] > ageOrder[ge.age] || def.Category == "wonder" {
			continue
		}
		ge.Buildings.UnlockBuilding(key)
		ge.Buildings.counts[key] = 5
	}
	// Harbour lineage explicitly, so harborRouteBonus has real work.
	ge.Buildings.counts["harbor"] = 4
	ge.Buildings.counts["harbor_authority"] = 3

	// Era Mastery in play: a full mastery map and a record deep enough that
	// the industrial age runs at catch-up speed (k ≠ 1), so the per-tick
	// scaling and the snapshot's mastery view are in the measurement.
	for _, a := range config.AgeOrder() {
		ge.Prestige.SetMastery(a, 3)
	}
	ge.Prestige.SetRecord("space_age")
	ge.Prestige.NoteAgeEntered(ge.age)

	for key := range ge.Resources.resources {
		setRes(ge, key, 1e6)
	}

	started := 0
	for key := range config.TradeRouteByKey() {
		if err := ge.Trade.StartRoute(key, ge.Buildings, ge.age, ageOrder); err == nil {
			started++
		}
	}
	if started == 0 {
		tb.Fatal("no trade routes started; benchmark would not exercise trade")
	}
	return ge
}

// BenchmarkTick measures one full engine tick on a late-game state. Guards
// against config tables (notably the ~300-def building table) being rebuilt on
// the per-tick path.
func BenchmarkTick(b *testing.B) {
	ge := newLateGameEngine(b)
	// Warm up once so one-off lazy work doesn't skew the first iteration.
	ge.doTick()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ge.doTick()
	}
}

// BenchmarkHarborRouteBonus isolates the harbour route-income sum that runs
// once per tick from processTrade.
func BenchmarkHarborRouteBonus(b *testing.B) {
	ge := newLateGameEngine(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ge.harborRouteBonus()
	}
}

// BenchmarkGetState measures the UI snapshot the dashboard pulls every refresh
// (500ms ticker plus input handlers). It holds the engine read lock, so config
// rebuilds here stall the tick goroutine too.
func BenchmarkGetState(b *testing.B) {
	ge := newLateGameEngine(b)
	ge.doTick()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ge.GetState()
	}
}

// TestTickDoesNotRebuildBuildingTable is the regression guard for the per-tick
// config rebuilds (card kVrX2nnT). config.BuildingByKey() reconstructs and
// re-normalizes every building def on each call; one such call on the tick
// path costs over a thousand allocations, while a whole late-game tick needs
// well under a hundred. The ceiling is calibrated against a live rebuild rather
// than hard-coded, so adding buildings to config doesn't make the test drift:
// if any per-tick path starts rebuilding the table again, the tick's
// allocations jump past half a rebuild and this fails.
func TestTickDoesNotRebuildBuildingTable(t *testing.T) {
	rebuild := testing.AllocsPerRun(5, func() { _ = config.BuildingByKey() })
	if rebuild < 200 {
		t.Fatalf("config.BuildingByKey() allocates only %.0f times per call; this guard assumes a rebuild is expensive, recalibrate it", rebuild)
	}

	ge := newLateGameEngine(t)
	ge.doTick() // warm-up: first-tick one-offs aren't the steady state
	perTick := testing.AllocsPerRun(50, ge.doTick)
	if ceiling := rebuild / 2; perTick > ceiling {
		t.Errorf("doTick allocates %.0f times per tick, over the ceiling of %.0f (half of one building-table rebuild, %.0f); a per-tick path is rebuilding config tables again, read the manager's held defs instead",
			perTick, ceiling, rebuild)
	}

	if n := testing.AllocsPerRun(50, func() { _ = ge.harborRouteBonus() }); n != 0 {
		t.Errorf("harborRouteBonus allocates %.0f times per call, want 0 (it should only read held defs)", n)
	}
}
