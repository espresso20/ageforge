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
// different one and lives in boon_tuning_test.go, scenarios B/D/F.

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
	// Exploration gates every expedition mission, but not the era's scout.
	learn(ge, "exploration")

	for _, def := range config.BaseResources() {
		ge.Resources.UnlockResource(def.Key)
		if r, ok := ge.Resources.resources[def.Key]; ok {
			r.Amount = stock
			r.Rate = 1 // an income, so a scout's cost is sized on something
		}
	}
	// Every mission takes a party: an army at home that covers any number.
	setSoldiers(ge, config.Trillion)
	if n > 0 {
		ge.Buildings.LoadCounts(map[string]int{autoExpeditionBuildingKey: n})
	}
	if assigned > 0 {
		ge.Workers.Unlock()
		ge.Workers.AddLentWorkers(assigned)
		if !ge.Workers.Assign(autoExpeditionBuildingKey, assigned) {
			t.Fatalf("could not assign %d workers to %s", assigned, autoExpeditionBuildingKey)
		}
	}
	return ge
}

// runAutoTicks drives processAutoExpeditions + the military tick for n ticks and
// returns how many expeditions were auto-launched and how many resolved. It does
// NOT run doTick, so nothing else in the economy moves.
func runAutoTicks(ge *GameEngine, n int) (launched, resolved int) {
	// A launch is a new active mission: a party that comes home and is sent
	// out again in the same tick still counts.
	seen := map[*ActiveExpedition]bool{}
	for _, a := range ge.Military.ActiveAll() {
		seen[a] = true
	}
	for i := 0; i < n; i++ {
		ge.tick++ // the board's refills and expiries run on the clock
		ge.ensureMissionBoardLocked()
		ge.Military.RefreshBoard(ge.gameRNG(), ge.tick)
		resolved += len(ge.Military.Tick())
		ge.processAutoExpeditions()
		for _, a := range ge.Military.ActiveAll() {
			if !seen[a] {
				launched++
				seen[a] = true
			}
		}
	}
	return launched, resolved
}

// TestGeographicSociety_Definition checks that the age that unlocks the building
// actually lists it.
func TestGeographicSociety_Definition(t *testing.T) {
	def, ok := config.BuildingByKey()[autoExpeditionBuildingKey]
	if !ok {
		t.Fatalf("%s is not defined in config", autoExpeditionBuildingKey)
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
	ge := autoExpeditionTestEngine(t, "atomic_age", 0, 0, config.Billion)
	if launched, _ := runAutoTicks(ge, 5000); launched != 0 {
		t.Errorf("%d expeditions auto-launched with no Geographic Society built", launched)
	}
	if ge.autoExpeditionTicksLeft != 0 {
		t.Errorf("countdown = %d with nothing built, want 0", ge.autoExpeditionTicksLeft)
	}
}

// TestAutoExpedition_NeverSendsWhatItCannotPay: the society sends only what
// the town can field. With no soldiers at home it sends nothing at all (every
// mission takes a party); with an army it sends, and never a grand mission,
// whatever its haul per minute.
func TestAutoExpedition_NeverSendsWhatItCannotPay(t *testing.T) {
	ge := autoExpeditionTestEngine(t, "atomic_age", 6, 48, config.Billion)
	learn(ge, "military_tactics")
	setSoldiers(ge, 0)
	if launched, _ := runAutoTicks(ge, 5000); launched != 0 {
		t.Fatalf("%d missions auto-launched with no soldiers at home", launched)
	}
	setSoldiers(ge, config.Trillion)
	launched, _ := runAutoTicks(ge, 20000)
	if launched == 0 {
		t.Fatal("nothing auto-launched at full investment with an army at home — the dispatcher is inert")
	}
	for _, a := range ge.Military.ActiveAll() {
		if a.Tier == TierGrand {
			t.Errorf("auto-dispatch launched a grand mission (%s)", a.Title)
		}
	}
	ge.EnsureMissionBoard()
	for i := 0; i < 50; i++ {
		n := ge.pickAutoOffer()
		if n == 0 {
			break
		}
		if o, _ := ge.Military.BoardOffer(n); o.Tier == TierGrand {
			t.Fatalf("picker chose %s", o.Title)
		}
	}
}

// TestAutoExpedition_RespectsTheArmy: the army is the only limit. Missions
// sent by hand do not block the society, but it sends only while the soldiers
// at home can make up the party, holds the dispatch when they cannot and
// fires it the tick they can.
func TestAutoExpedition_RespectsTheArmy(t *testing.T) {
	ge := autoExpeditionTestEngine(t, "atomic_age", 6, 48, config.Billion)
	// The soldiers at home: one short party and a little over.
	short := ge.partySizeLocked(MissionOffer{Tier: TierShort})
	setSoldiers(ge, float64(short)+1)
	if err := ge.launchExpeditionLocked(plantOffer(ge, TierShort, "")); err != nil {
		t.Fatalf("manual launch failed: %v", err)
	}
	manual := ge.Military.ActiveByTier(TierShort)
	manual.TicksLeft = 4000
	if got := ge.soldiersHome(); got != 1 {
		t.Fatalf("%d soldiers at home after a party of %d left, want 1", got, short)
	}

	for i := 0; i < 2000; i++ {
		ge.processAutoExpeditions()
		if n := len(ge.Military.ActiveAll()); n != 1 {
			t.Fatalf("tick %d: %d missions under way, want only the manual one (no soldiers to send another)", i, n)
		}
	}
	// The countdown must be spent and WAITING, not forfeited.
	if ge.autoExpeditionTicksLeft != 0 {
		t.Errorf("countdown = %d after 2000 held ticks, want 0 (a dispatch held pending)",
			ge.autoExpeditionTicksLeft)
	}

	// Soldiers come home: the pending dispatch fires on the very next tick.
	setSoldiers(ge, config.Trillion)
	ge.processAutoExpeditions()
	if n := len(ge.Military.ActiveAll()); n != 2 {
		t.Errorf("%d missions under way after the soldiers came home, want 2 (the manual one and the society's)", n)
	}
}

// TestAutoExpedition_PaysTheCost: an auto-launch is charged exactly what a manual
// launch is charged.
func TestAutoExpedition_PaysTheCost(t *testing.T) {
	ge := autoExpeditionTestEngine(t, "atomic_age", 6, 48, config.Billion)
	ge.EnsureMissionBoard()
	n := ge.pickAutoOffer()
	if n == 0 {
		t.Fatal("picker returned nothing with a full treasury")
	}
	offer, _ := ge.Military.BoardOffer(n)
	cost := ge.missionCostLocked(offer)
	if len(cost) == 0 {
		t.Skip("the chosen mission is free — nothing to charge")
	}

	// The party is the soldiers at home, not a resource.
	held := func(res string) float64 {
		if res == "soldiers" {
			return float64(ge.soldiersHome())
		}
		return ge.Resources.Get(res)
	}
	before := map[string]float64{}
	for res := range cost {
		before[res] = held(res)
	}

	launched, _ := runAutoTicks(ge, 1)
	if launched != 1 {
		t.Fatalf("expected 1 immediate dispatch on the first tick, got %d", launched)
	}
	for res, amount := range cost {
		want := before[res] - amount
		if got := held(res); got != want {
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
			r.Amount = config.Million
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
		ge := autoExpeditionTestEngine(t, "atomic_age", tc.count, tc.assigned, config.Trillion)
		_, fill := ge.autoExpeditionInvestment()
		want := autoExpeditionIntervalIn(ge.age, tc.count, fill)
		scoutLongest := minutesToTicks(config.MissionBoardLongestMinutes())

		launched, _ := runAutoTicks(ge, ticks)
		if launched == 0 {
			t.Fatalf("%s: nothing dispatched over %d ticks", tc.name, ticks)
		}
		cycle := float64(ticks) / float64(launched)
		t.Logf("%-44s formula interval=%4d  observed cycle=%7.1f ticks  (%d dispatches/%dk)",
			tc.name, want, cycle, launched, ticks/1000)

		// The observed cycle is max(interval, expedition duration), so it can only
		// be LONGER than the formula, and by at most the longest the era's scout
		// can be out (rolled at every launch, the age does not stretch it).
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
		if cycle > float64(want+scoutLongest) {
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
	ge := autoExpeditionTestEngine(t, "atomic_age", 1, 0, config.Billion)
	// One dispatch, then 100 ticks of the next countdown served.
	runAutoTicks(ge, 101)
	want := ge.autoExpeditionTicksLeft
	if want <= 0 {
		t.Fatalf("countdown = %d after 101 ticks, expected a partially-served wait", want)
	}

	save := MilitarySave{
		Active:                  ge.Military.ActiveForSave(),
		AutoExpeditionTicksLeft: ge.autoExpeditionTicksLeft,
	}

	fresh := autoExpeditionTestEngine(t, "atomic_age", 1, 0, config.Billion)
	fresh.autoExpeditionTicksLeft = save.AutoExpeditionTicksLeft
	if got := fresh.autoExpeditionTicksLeft; got != want {
		t.Errorf("countdown after reload = %d, want %d", got, want)
	}
	// And it must not dispatch on the first tick back.
	if launched, _ := runAutoTicks(fresh, 1); launched != 0 {
		t.Error("a reloaded save dispatched immediately — the wait was reset")
	}
}

// TestAutoExpedition_PickerIsDeterministicAndBest pins the selection rule:
// the offer with the best haul per minute of length that can be sent, stable
// across repeated calls.
func TestAutoExpedition_PickerIsDeterministicAndBest(t *testing.T) {
	ge := autoExpeditionTestEngine(t, "atomic_age", 1, 0, config.Billion)
	learn(ge, "military_tactics")
	ge.EnsureMissionBoard()

	first := ge.pickAutoOffer()
	if first == 0 {
		t.Fatal("picker returned nothing at the atomic age with a full treasury")
	}
	for i := 0; i < 20; i++ {
		if again := ge.pickAutoOffer(); again != first {
			t.Fatalf("picker is not deterministic: %d then %d", first, again)
		}
	}

	// Nothing on the board may beat it among the offers that can be sent
	// (short only: the Society sends the era's scout, never a long or grand offer).
	chosen, _ := ge.Military.BoardOffer(first)
	if chosen.Tier != TierShort {
		t.Fatalf("picker chose %s offer %d", chosen.Tier, first)
	}
	for i, slot := range ge.Military.board.Slots {
		if slot.Offer.Tier != TierShort {
			continue
		}
		if _, err := ge.offerRefusalLocked(*slot.Offer, ge.missionCostLocked(*slot.Offer)); err != nil {
			continue
		}
		if rate, best := slot.Offer.HaulMinutes/slot.Offer.LengthMinutes, chosen.HaulMinutes/chosen.LengthMinutes; rate > best {
			t.Errorf("picker chose %d (rate %.3f) over %d (rate %.3f)", first, best, i+1, rate)
		}
	}

	// An empty treasury yields nothing, not a panic.
	for _, def := range config.BaseResources() {
		ge.Resources.resources[def.Key].Amount = 0
	}
	if n := ge.pickAutoOffer(); n != 0 {
		t.Errorf("picker returned offer %d with an empty treasury", n)
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
		ge := autoExpeditionTestEngine(t, "industrial_age", 0, 0, config.Billion)
		got := ge.GetState().Military.AutoExpedition
		if got != (AutoExpeditionState{}) {
			t.Errorf("no society should snapshot as the zero value, got %+v", got)
		}
	})

	t.Run("built and staffed", func(t *testing.T) {
		// 2 societies x WorkerCapacity 8 = 16 capacity, 8 assigned = 0.5 fill.
		ge := autoExpeditionTestEngine(t, "industrial_age", 2, 8, config.Billion)
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
		if want := autoExpeditionIntervalIn("industrial_age", 2, 0.5); got.Interval != want {
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
