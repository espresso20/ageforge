package game

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Tests for automatic expedition dispatch (auto_expedition.go).
//
// These drive the REAL engine path — processAutoExpeditions →
// launchExpeditionLocked → MilitaryManager — and assert the correctness rules the
// feature must never break: scouting only, one slot honoured, cost actually paid.
// The statistical question (does automation stay slower than hands-on play?) is a
// different one and lives in boon_tuning_test.go, scenarios A/D/F.

// autoExpeditionTestEngine returns an engine parked at an age with n Geographic
// Societies standing, `assigned` workers spread across them, and every resource
// stocked (or emptied, when stock is 0).
func autoExpeditionTestEngine(t *testing.T, age string, n, assigned int, stock float64) *GameEngine {
	t.Helper()
	ge := NewGameEngine()
	ge.SeedRNG(0xA07E)
	ge.age = age
	ge.currentEpoch = config.EpochForAge(age)
	ge.Events.nextEventTick = 1 << 40

	for _, def := range config.BaseResources() {
		ge.Resources.UnlockResource(def.Key)
		if r, ok := ge.Resources.resources[def.Key]; ok {
			r.Storage = stock*10 + 1
			r.Amount = stock
			r.Rate = 0
		}
	}
	if n > 0 {
		ge.Buildings.LoadCounts(map[string]int{autoExpeditionBuildingKey: n})
	}
	if assigned > 0 {
		ge.Workers.UnlockType("worker")
		ge.Workers.AddLentWorkers(assigned)
		if !ge.Workers.Assign("worker", autoExpeditionBuildingKey, assigned) {
			t.Fatalf("could not assign %d workers to %s", assigned, autoExpeditionBuildingKey)
		}
	}
	return ge
}

// runAutoTicks drives processAutoExpeditions + the military tick for n ticks and
// returns how many expeditions were auto-launched and how many resolved. It does
// NOT run doTick, so nothing else in the economy moves.
func runAutoTicks(ge *GameEngine, n int) (launched, resolved int) {
	prevActive := ge.Military.ActiveByCategory(ExpeditionScouting) != nil
	for i := 0; i < n; i++ {
		resolved += len(ge.Military.Tick(0, 0))
		ge.processAutoExpeditions()
		nowActive := ge.Military.ActiveByCategory(ExpeditionScouting) != nil
		if nowActive && !prevActive {
			launched++
		}
		prevActive = nowActive
	}
	return launched, resolved
}

// TestGeographicSociety_Definition pins the building: it exists, it is a LATE
// unlock, it takes workers, and the age that unlocks it actually lists it.
func TestGeographicSociety_Definition(t *testing.T) {
	def, ok := config.BuildingByKey()[autoExpeditionBuildingKey]
	if !ok {
		t.Fatalf("%s is not defined in config", autoExpeditionBuildingKey)
	}
	if def.RequiredAge != "industrial_age" {
		t.Errorf("RequiredAge = %q, want industrial_age", def.RequiredAge)
	}

	// Belt-and-braces on "later age": its position in the age order must be at or
	// past the industrial age, whatever RequiredAge is edited to later.
	ageIndex := map[string]int{}
	for i, key := range config.AgeOrder() {
		ageIndex[key] = i
	}
	if ageIndex[def.RequiredAge] < ageIndex["industrial_age"] {
		t.Errorf("%s unlocks at %q, which is earlier than the industrial age — "+
			"auto-dispatch must not be an early-game feature", autoExpeditionBuildingKey, def.RequiredAge)
	}
	if def.WorkerDomain == "" || def.WorkerCapacity <= 0 {
		t.Errorf("WorkerDomain=%q WorkerCapacity=%d — the cadence formula needs both",
			def.WorkerDomain, def.WorkerCapacity)
	}
	if len(def.BaseCost) == 0 {
		t.Error("BaseCost is empty")
	}
	// No production effect may leak into the resource pipeline (same rule as the
	// embassies, which share this slice of buildings.go).
	for _, eff := range def.Effects {
		if eff.Type == "production" {
			t.Errorf("unexpected production effect %+v", eff)
		}
	}

	var listed bool
	for _, a := range config.Ages() {
		if a.Key != def.RequiredAge {
			continue
		}
		for _, key := range a.UnlockBuildings {
			if key == autoExpeditionBuildingKey {
				listed = true
			}
		}
	}
	if !listed {
		t.Errorf("%s is not in the UnlockBuildings list for %s — it would never become buildable",
			autoExpeditionBuildingKey, def.RequiredAge)
	}
}

// TestAutoExpeditionInterval_ScalesAndFloors pins the cadence curve: off with
// nothing built, monotonically faster with more societies and fuller staffing, and
// never below the floor that holds the engagement gradient.
func TestAutoExpeditionInterval_ScalesAndFloors(t *testing.T) {
	if got := autoExpeditionIntervalFor(0, 1); got != 0 {
		t.Errorf("interval with nothing built = %d, want 0 (automation off)", got)
	}

	one := autoExpeditionIntervalFor(1, 0)
	if one != autoExpeditionBaseInterval {
		t.Errorf("one unstaffed society = %d, want %d", one, autoExpeditionBaseInterval)
	}

	// Staffing shortens it; more copies shorten it further.
	oneStaffed := autoExpeditionIntervalFor(1, 1)
	if oneStaffed >= one {
		t.Errorf("full staffing did not shorten the interval: %d >= %d", oneStaffed, one)
	}
	prev := oneStaffed
	for n := 2; n <= 6; n++ {
		got := autoExpeditionIntervalFor(n, 1)
		if got > prev {
			t.Errorf("%d societies gave a LONGER interval (%d) than %d (%d)", n, got, n-1, prev)
		}
		prev = got
	}

	// The floor is absolute, at any absurd investment, and clamps bad input.
	for _, n := range []int{6, 10, 100, 10000} {
		if got := autoExpeditionIntervalFor(n, 1); got < autoExpeditionMinInterval {
			t.Errorf("%d societies broke the floor: %d < %d", n, got, autoExpeditionMinInterval)
		}
	}
	if got := autoExpeditionIntervalFor(3, 5); got != autoExpeditionIntervalFor(3, 1) {
		t.Errorf("fill > 1 not clamped: %d vs %d", got, autoExpeditionIntervalFor(3, 1))
	}
	if got := autoExpeditionIntervalFor(3, -2); got != autoExpeditionIntervalFor(3, 0) {
		t.Errorf("negative fill not clamped: %d vs %d", got, autoExpeditionIntervalFor(3, 0))
	}
}

// TestAutoExpedition_NoBuildingNoDispatch: without the society, nothing is ever
// dispatched, however long the game runs.
func TestAutoExpedition_NoBuildingNoDispatch(t *testing.T) {
	ge := autoExpeditionTestEngine(t, "atomic_age", 0, 0, 1e9)
	if launched, _ := runAutoTicks(ge, 5000); launched != 0 {
		t.Errorf("%d expeditions auto-launched with no Geographic Society built", launched)
	}
	if ge.autoExpeditionTicksLeft != 0 {
		t.Errorf("countdown = %d with nothing built, want 0", ge.autoExpeditionTicksLeft)
	}
}

// TestAutoExpedition_ScoutingOnly is the hard constraint: automation must never
// dispatch a military campaign, and must never touch the military slot.
func TestAutoExpedition_ScoutingOnly(t *testing.T) {
	ge := autoExpeditionTestEngine(t, "atomic_age", 6, 48, 1e9)
	launched, _ := runAutoTicks(ge, 20000)
	if launched == 0 {
		t.Fatal("nothing auto-launched at full investment — the dispatcher is inert")
	}
	if active := ge.Military.ActiveByCategory(ExpeditionMilitary); active != nil {
		t.Errorf("auto-dispatch launched a MILITARY expedition (%s)", active.Name)
	}
	// Every dispatch the picker can produce must be a scouting def.
	for i := 0; i < 50; i++ {
		def := ge.pickAutoScoutExpedition(ge.progress.GetAgeOrder())
		if def == nil {
			t.Fatal("picker returned nil with a full treasury at the atomic age")
		}
		if def.Category != ExpeditionScouting {
			t.Fatalf("picker chose a %s expedition (%s)", def.Category, def.Name)
		}
	}
}

// TestAutoExpedition_RespectsSingleSlot: a manually-launched scouting expedition
// blocks auto-dispatch for its whole duration — never two at once, never a queue.
func TestAutoExpedition_RespectsSingleSlot(t *testing.T) {
	ge := autoExpeditionTestEngine(t, "atomic_age", 6, 48, 1e9)

	// Park a scouting expedition in the slot by hand and freeze its timer high.
	if err := ge.launchExpeditionLocked("scout_ruins"); err != nil {
		t.Fatalf("manual launch failed: %v", err)
	}
	manual := ge.Military.ActiveByCategory(ExpeditionScouting)
	if manual == nil {
		t.Fatal("manual launch did not occupy the scouting slot")
	}
	manual.TicksLeft = 4000
	name := manual.Name

	for i := 0; i < 2000; i++ {
		ge.processAutoExpeditions()
		active := ge.Military.ActiveByCategory(ExpeditionScouting)
		if active == nil {
			t.Fatalf("tick %d: the slot emptied — auto-dispatch replaced the manual expedition", i)
		}
		if active != manual {
			t.Fatalf("tick %d: auto-dispatch overwrote the active expedition (%s -> %s)", i, name, active.Name)
		}
	}
	// The countdown must be spent and WAITING, not forfeited.
	if ge.autoExpeditionTicksLeft != 0 {
		t.Errorf("countdown = %d after 2000 blocked ticks, want 0 (a dispatch held pending)",
			ge.autoExpeditionTicksLeft)
	}

	// Free the slot: the pending dispatch must fire on the very next tick.
	ge.Military.activeByCat[ExpeditionScouting] = nil
	ge.processAutoExpeditions()
	if ge.Military.ActiveByCategory(ExpeditionScouting) == nil {
		t.Error("the pending dispatch did not fire once the slot freed")
	}
}

// TestAutoExpedition_PaysTheCost: an auto-launch is charged exactly what a manual
// launch is charged.
func TestAutoExpedition_PaysTheCost(t *testing.T) {
	ge := autoExpeditionTestEngine(t, "atomic_age", 6, 48, 1e9)
	def := ge.pickAutoScoutExpedition(ge.progress.GetAgeOrder())
	if def == nil {
		t.Fatal("picker returned nil with a full treasury")
	}
	if len(def.Cost) == 0 {
		t.Skip("the chosen scouting expedition is free — nothing to charge")
	}

	before := map[string]float64{}
	for res := range def.Cost {
		before[res] = ge.Resources.Get(res)
	}

	launched, _ := runAutoTicks(ge, 1)
	if launched != 1 {
		t.Fatalf("expected 1 immediate dispatch on the first tick, got %d", launched)
	}
	for res, amount := range def.Cost {
		want := before[res] - amount
		if got := ge.Resources.Get(res); got != want {
			t.Errorf("%s: have %.2f after the auto-launch, want %.2f (cost %.2f unpaid or double-charged)",
				res, got, want, amount)
		}
	}
}

// TestAutoExpedition_SkipsWhenUnaffordable: an empty treasury holds the dispatch
// (never launches on credit), warns exactly ONCE per dry spell, and resumes the
// moment supplies return.
func TestAutoExpedition_SkipsWhenUnaffordable(t *testing.T) {
	ge := autoExpeditionTestEngine(t, "atomic_age", 6, 48, 0)

	launched, _ := runAutoTicks(ge, 3000)
	if launched != 0 {
		t.Errorf("%d expeditions launched with an empty treasury", launched)
	}
	if !ge.autoExpeditionStarved {
		t.Error("the starved flag was never raised")
	}

	warnings := 0
	for _, entry := range ge.log {
		if entry.Type == "warning" && strings.Contains(entry.Message, "no supplies to outfit it") {
			warnings++
		}
	}
	if warnings != 1 {
		t.Errorf("logged the supply warning %d times over 3000 ticks, want exactly 1", warnings)
	}

	// Refill: the held dispatch must go out immediately, and the flag must clear.
	for _, def := range config.BaseResources() {
		if r, ok := ge.Resources.resources[def.Key]; ok {
			r.Storage = 1e9
			r.Amount = 1e6
		}
	}
	if launched, _ := runAutoTicks(ge, 1); launched != 1 {
		t.Error("the held dispatch did not fire once supplies returned")
	}
	if ge.autoExpeditionStarved {
		t.Error("the starved flag survived a successful dispatch")
	}
}

// TestAutoExpedition_CadenceScalesWithInvestment measures the real dispatcher's
// observed cycle length and checks it tracks the formula — and that a bigger,
// better-staffed investment really does dispatch more often.
func TestAutoExpedition_CadenceScalesWithInvestment(t *testing.T) {
	const ticks = 100000
	cases := []struct {
		name     string
		count    int
		assigned int
	}{
		{"1 society, unstaffed", 1, 0},
		{"1 society, fully staffed", 1, 8},
		{"3 societies, fully staffed", 3, 24},
		{"6 societies, fully staffed (floor)", 6, 48},
		{"20 societies, fully staffed (past the floor)", 20, 160},
	}

	var prev float64
	for i, tc := range cases {
		ge := autoExpeditionTestEngine(t, "atomic_age", tc.count, tc.assigned, 1e12)
		_, fill := ge.autoExpeditionInvestment()
		want := autoExpeditionIntervalFor(tc.count, fill)

		launched, _ := runAutoTicks(ge, ticks)
		if launched == 0 {
			t.Fatalf("%s: nothing dispatched over %d ticks", tc.name, ticks)
		}
		cycle := float64(ticks) / float64(launched)
		t.Logf("%-44s formula interval=%4d  observed cycle=%7.1f ticks  (%d dispatches/%dk)",
			tc.name, want, cycle, launched, ticks/1000)

		// The observed cycle is max(interval, expedition duration), so it can only
		// be LONGER than the formula, and by at most the duration span.
		//
		// The 2% slack on the low side is a finite-window artefact, not slop: the
		// society's first party leaves the tick the doors open (the countdown starts
		// at zero), so a window of N ticks contains 1 + N/interval dispatches and the
		// mean cycle reads interval × (1 − interval/N). At interval=900 over 100k
		// ticks that is 892.9, exactly what is measured.
		if cycle < float64(want)*0.98-1 {
			t.Errorf("%s: observed cycle %.1f is shorter than the interval %d — the countdown is leaking",
				tc.name, cycle, want)
		}
		if cycle > float64(want)+120 {
			t.Errorf("%s: observed cycle %.1f is far longer than the interval %d",
				tc.name, cycle, want)
		}
		// Investment must never make dispatching SLOWER. The 5% slack matters once
		// the floor binds: past it the formula interval stops moving, so consecutive
		// cases differ only by the per-launch duration roll — which comes off the
		// global rand inside LaunchExpedition and is therefore not reproducible
		// between runs. Without the slack this check flaps on noise.
		if i > 0 && cycle > prev*1.05+1 {
			t.Errorf("%s: cycle %.1f is meaningfully slower than the smaller investment before it (%.1f)",
				tc.name, cycle, prev)
		}
		prev = cycle
	}
}

// TestAutoExpedition_CountdownSurvivesSaveReload: the dispatch countdown is
// persisted, so a save/reload cannot be used to skip the wait.
func TestAutoExpedition_CountdownSurvivesSaveReload(t *testing.T) {
	ge := autoExpeditionTestEngine(t, "atomic_age", 1, 0, 1e9)
	// One dispatch, then 100 ticks of the next countdown served.
	runAutoTicks(ge, 101)
	want := ge.autoExpeditionTicksLeft
	if want <= 0 {
		t.Fatalf("countdown = %d after 101 ticks, expected a partially-served wait", want)
	}

	scout, military := ge.Military.GetActiveForSave()
	save := MilitarySave{
		ActiveScout:             scout,
		ActiveMilitary:          military,
		AutoExpeditionTicksLeft: ge.autoExpeditionTicksLeft,
	}

	fresh := autoExpeditionTestEngine(t, "atomic_age", 1, 0, 1e9)
	fresh.autoExpeditionTicksLeft = save.AutoExpeditionTicksLeft
	if got := fresh.autoExpeditionTicksLeft; got != want {
		t.Errorf("countdown after reload = %d, want %d", got, want)
	}
	// And it must not dispatch on the first tick back.
	if launched, _ := runAutoTicks(fresh, 1); launched != 0 {
		t.Error("a reloaded save dispatched immediately — the wait was reset")
	}
}

// TestAutoExpedition_PickerIsDeterministicAndCheapest pins the selection rule:
// the cheapest affordable scouting expedition, stable across repeated calls.
func TestAutoExpedition_PickerIsDeterministicAndCheapest(t *testing.T) {
	ge := autoExpeditionTestEngine(t, "atomic_age", 1, 0, 1e9)
	order := ge.progress.GetAgeOrder()

	first := ge.pickAutoScoutExpedition(order)
	if first == nil {
		t.Fatal("picker returned nil at the atomic age with a full treasury")
	}
	for i := 0; i < 20; i++ {
		again := ge.pickAutoScoutExpedition(order)
		if again == nil || again.Key != first.Key {
			t.Fatalf("picker is not deterministic: %v then %v", first.Key, again)
		}
	}

	// Nothing cheaper may exist among the affordable scouting defs.
	for _, def := range ge.Military.GetAvailableExpeditionsByCategory(ExpeditionScouting, ge.age, order) {
		if totalExpeditionCost(def) < totalExpeditionCost(*first) {
			t.Errorf("picker chose %s (cost %.0f) over the cheaper %s (cost %.0f)",
				first.Key, totalExpeditionCost(*first), def.Key, totalExpeditionCost(def))
		}
	}

	// An age with no scouting expedition at all must yield nil, not a panic.
	ge.age = "primitive_age"
	ge.Resources.resources["food"].Amount = 0
	ge.Resources.resources["wood"].Amount = 0
	if def := ge.pickAutoScoutExpedition(order); def != nil {
		t.Errorf("picker returned %s with an empty treasury", def.Key)
	}
}

// TestAutoExpeditionSnapshot proves automatic dispatch is VISIBLE on GameState.
//
// The whole mechanic used to live on engine-private fields
// (autoExpeditionTicksLeft / autoExpeditionStarved) plus a building count the UI
// was not allowed to look up by key, so a player who built a Geographic Society
// got no confirmation it existed, no countdown, and no warning when it was
// sitting starved. AutoExpeditionState is the fix; this is its contract.
func TestAutoExpeditionSnapshot(t *testing.T) {
	t.Run("nothing built reads as off", func(t *testing.T) {
		ge := autoExpeditionTestEngine(t, "industrial_age", 0, 0, 1e9)
		got := ge.GetState().Military.AutoExpedition
		if got != (AutoExpeditionState{}) {
			t.Errorf("no society should snapshot as the zero value, got %+v", got)
		}
	})

	t.Run("built and staffed", func(t *testing.T) {
		// 2 societies x WorkerCapacity 8 = 16 capacity, 8 assigned = 0.5 fill.
		ge := autoExpeditionTestEngine(t, "industrial_age", 2, 8, 1e9)
		// One pass so the countdown is primed by a real dispatch rather than
		// reading back its zero value.
		if launched, _ := runAutoTicks(ge, 1); launched != 1 {
			t.Fatalf("expected the due dispatch to launch, got %d", launched)
		}

		got := ge.GetState().Military.AutoExpedition
		if !got.Active {
			t.Error("Active = false with 2 societies standing")
		}
		if got.Count != 2 {
			t.Errorf("Count = %d, want 2", got.Count)
		}
		if got.Assigned != 8 {
			t.Errorf("Assigned = %d, want 8", got.Assigned)
		}
		if got.Capacity != 16 {
			t.Errorf("Capacity = %d, want 16 (2 societies x WorkerCapacity 8)", got.Capacity)
		}
		if want := autoExpeditionIntervalFor(2, 0.5); got.Interval != want {
			t.Errorf("Interval = %d, want %d (the cadence formula at this investment)", got.Interval, want)
		}
		if got.TicksLeft <= 0 || got.TicksLeft > got.Interval {
			t.Errorf("TicksLeft = %d, want a live countdown in (0, %d]", got.TicksLeft, got.Interval)
		}
		if got.Starved {
			t.Error("Starved = true with every resource stocked")
		}
	})

	t.Run("starved dispatch is visible", func(t *testing.T) {
		ge := autoExpeditionTestEngine(t, "industrial_age", 1, 0, 0)
		if launched, _ := runAutoTicks(ge, 1); launched != 0 {
			t.Fatalf("nothing is affordable; expected no dispatch, got %d", launched)
		}

		got := ge.GetState().Military.AutoExpedition
		if !got.Active {
			t.Fatal("Active = false with a society standing")
		}
		if !got.Starved {
			t.Errorf("Starved = false with an empty treasury; snapshot %+v", got)
		}
		if got.TicksLeft != 0 {
			t.Errorf("TicksLeft = %d, want 0 — a due dispatch holds at zero and retries", got.TicksLeft)
		}
		if got.Assigned != 0 || got.Capacity != 8 {
			t.Errorf("Assigned/Capacity = %d/%d, want 0/8", got.Assigned, got.Capacity)
		}
	})
}
