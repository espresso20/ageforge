package game

import (
	"os"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// stageSoldierProduction builds & fully staffs `numBarracks` barracks so the
// engine produces the soldiers resource, then returns the per-tick training rate
// and the soldiers storage cap. Caller must NOT hold the engine lock. Food is
// topped up and morale pinned by the caller's tick helper.
func stageSoldierProduction(t *testing.T, ge *GameEngine, numBarracks int) {
	t.Helper()
	const barracks = "barracks"
	def := config.BuildingByKey()[barracks]

	ge.mu.Lock()
	ge.Buildings.counts[barracks] = numBarracks
	ge.Buildings.unlocked[barracks] = true
	ge.Resources.UnlockResource("soldiers")
	ge.Resources.UnlockResource("food")
	ge.Resources.AddStorage("food", 1e9)
	ge.Workers.UnlockType("worker")
	totalSlots := def.WorkerCapacity * numBarracks
	if !ge.Workers.Recruit("worker", totalSlots, 100000) {
		ge.mu.Unlock()
		t.Fatalf("failed to recruit %d workers", totalSlots)
	}
	if !ge.Workers.Assign("worker", barracks, totalSlots) {
		ge.mu.Unlock()
		t.Fatalf("failed to assign %d workers to barracks", totalSlots)
	}
	ge.mu.Unlock()
}

// tickWithFood advances one engine tick with food topped up and morale pinned to
// 1.0 (so soldiers production is isolated from food-starvation / military-ratio
// morale penalties — same approach as TestMilitaryBuildings_ProduceAndStoreSoldiers).
func tickWithFood(ge *GameEngine) {
	ge.mu.Lock()
	ge.Resources.Add("food", 1e6)
	ge.morale = 1.0
	ge.mu.Unlock()
	ge.doTick()
	ge.mu.Lock()
	ge.morale = 1.0
	ge.mu.Unlock()
}

// TestSoldiersTrained_AccruesAcrossTicks verifies the cumulative lifetime
// soldiers-trained counter increments by the soldiers actually added each tick.
func TestSoldiersTrained_AccruesAcrossTicks(t *testing.T) {
	ge := NewGameEngine()
	stageSoldierProduction(t, ge, 2)

	if got := ge.Stats.SoldiersTrained; got != 0 {
		t.Fatalf("SoldiersTrained before any tick = %v, want 0", got)
	}

	for i := 0; i < 10; i++ {
		tickWithFood(ge)
	}

	trained := ge.Stats.SoldiersTrained
	if trained <= 0 {
		t.Fatalf("SoldiersTrained after 10 ticks = %v, want > 0", trained)
	}

	// The cumulative counter should track the soldiers resource amount while the
	// resource is still below its storage cap and nothing has been spent.
	if soldiers := ge.Resources.Get("soldiers"); trained < soldiers {
		t.Errorf("SoldiersTrained=%v should be >= current soldiers resource=%v", trained, soldiers)
	}
}

// TestSoldiersTrained_NotDecreasedBySpending verifies that spending soldiers on
// an expedition does NOT reduce the cumulative lifetime trained counter.
func TestSoldiersTrained_NotDecreasedBySpending(t *testing.T) {
	ge := NewGameEngine()
	stageSoldierProduction(t, ge, 2)

	// Accrue some trained soldiers, then jump the resource amount high enough to
	// launch an expedition that costs soldiers.
	for i := 0; i < 20; i++ {
		tickWithFood(ge)
	}
	trainedBefore := ge.Stats.SoldiersTrained
	if trainedBefore <= 0 {
		t.Fatalf("expected SoldiersTrained > 0 after warmup, got %v", trainedBefore)
	}

	// Put us in iron_age with a healthy soldiers stockpile and launch
	// trade_escort (needs 3 soldiers, no Cost). Spending must not touch the
	// lifetime counter.
	setAge(ge, "iron_age")
	setSoldiers(ge, 10)

	if err := ge.LaunchExpedition("trade_escort"); err != nil {
		t.Fatalf("LaunchExpedition(trade_escort) error: %v", err)
	}
	if got := ge.Resources.Get("soldiers"); got != 7 {
		t.Fatalf("soldiers after launch = %v, want 7 (10 - 3)", got)
	}

	if got := ge.Stats.SoldiersTrained; got != trainedBefore {
		t.Errorf("SoldiersTrained changed by spending soldiers: before=%v after=%v (must be unchanged)", trainedBefore, got)
	}
}

// TestFirstSoldiersMilestone_TriggersOnCumulativeTrained verifies the
// first_soldiers milestone (5) fires off the cumulative trained count, not the
// live military-worker count.
func TestFirstSoldiersMilestone_TriggersOnCumulativeTrained(t *testing.T) {
	mm := NewMilestoneManager()
	ageOrder := fullAgeOrder()
	rm := NewResourceManager()
	bm := NewBuildingManager()

	check := func(soldiersTrained int) []config.MilestoneDef {
		return mm.CheckMilestones(
			100, "iron_age", ageOrder,
			rm, bm,
			0,   // population
			0,   // techCount
			0,   // totalBuilt
			nil, // researchedTechs
			soldiersTrained,
			0, // wonderCount
			0, // knowledgeCount
		)
	}

	// 4 trained: not enough.
	if got := check(4); len(got) != 0 {
		for _, d := range got {
			if d.Key == "first_soldiers" {
				t.Fatalf("first_soldiers should NOT trigger at 4 trained")
			}
		}
	}
	if mm.completed["first_soldiers"] {
		t.Fatalf("first_soldiers marked complete at 4 trained")
	}

	// 5 trained: triggers.
	got := check(5)
	found := false
	for _, d := range got {
		if d.Key == "first_soldiers" {
			found = true
		}
	}
	if !found {
		t.Fatalf("first_soldiers should trigger at 5 cumulative trained")
	}
	if !mm.completed["first_soldiers"] {
		t.Errorf("first_soldiers should be marked complete after triggering")
	}
}

// TestSoldiersTrained_SaveLoadRoundTrip verifies the cumulative counter survives
// a save/load round trip. (An old save without the field unmarshals to 0 — the
// json default — which is the accepted "progress re-accrues" behavior.)
func TestSoldiersTrained_SaveLoadRoundTrip(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	ge.Stats.SoldiersTrained = 137
	ge.mu.Unlock()

	if err := ge.SaveGame("test_soldiers_trained_roundtrip"); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}
	defer os.Remove("data/saves/test_soldiers_trained_roundtrip.json")

	ge2 := NewGameEngine()
	if err := ge2.LoadGame("test_soldiers_trained_roundtrip"); err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}
	if got := ge2.Stats.SoldiersTrained; got != 137 {
		t.Errorf("SoldiersTrained after load = %v, want 137", got)
	}
}

// TestCanLaunch_RespectsCost verifies the available-expedition snapshot's
// CanLaunch flag accounts for Cost affordability: scout_party (food 30 + wood 30,
// no soldiers) is not launchable without the resources and launchable with them.
func TestCanLaunch_RespectsCost(t *testing.T) {
	// --- Not affordable: no food/wood. ---
	ge := NewGameEngine()
	setAge(ge, "stone_age")
	state := ge.GetState()

	var scout ExpeditionInfo
	ok := false
	for _, exp := range state.Military.Expeditions {
		if exp.Key == "scout_party" {
			scout = exp
			ok = true
		}
	}
	if !ok {
		t.Fatalf("scout_party not present in available expeditions at stone_age")
	}
	if scout.CanLaunch {
		t.Errorf("scout_party CanLaunch=true with no food/wood, want false")
	}
	if scout.LaunchBlockReason == "" {
		t.Errorf("scout_party blocked but LaunchBlockReason empty")
	}

	// --- Affordable: enough food + wood. ---
	ge2 := NewGameEngine()
	setAge(ge2, "stone_age")
	setResource(ge2, "food", 100)
	setResource(ge2, "wood", 100)
	state2 := ge2.GetState()

	found := false
	for _, exp := range state2.Military.Expeditions {
		if exp.Key == "scout_party" {
			found = true
			if !exp.CanLaunch {
				t.Errorf("scout_party CanLaunch=false with 100 food + 100 wood, want true (reason: %q)", exp.LaunchBlockReason)
			}
			if exp.LaunchBlockReason != "" {
				t.Errorf("scout_party launchable but LaunchBlockReason=%q, want empty", exp.LaunchBlockReason)
			}
		}
	}
	if !found {
		t.Fatalf("scout_party not present in available expeditions with resources")
	}
}

// === Scouting / Army split (Phase 1) ===

// TestExpeditionCategory_AllValid verifies every expedition is classified as one
// of the two known categories — no blanks or typos that would drop it from both
// grouped subsections.
func TestExpeditionCategory_AllValid(t *testing.T) {
	mm := NewMilitaryManager()
	for _, def := range mm.expeditions {
		switch def.Category {
		case ExpeditionScouting, ExpeditionMilitary:
			// ok
		default:
			t.Errorf("expedition %q has invalid Category %q (want %q or %q)",
				def.Key, def.Category, ExpeditionScouting, ExpeditionMilitary)
		}
	}
}

// TestScoutingExpeditions_SoldierFree verifies all scouting expeditions cost zero
// soldiers — scouting must be playable before the soldiers resource exists
// (pre-iron_age).
func TestScoutingExpeditions_SoldierFree(t *testing.T) {
	mm := NewMilitaryManager()
	scoutCount := 0
	for _, def := range mm.expeditions {
		if def.Category != ExpeditionScouting {
			continue
		}
		scoutCount++
		if def.SoldiersNeeded != 0 {
			t.Errorf("scouting expedition %q has SoldiersNeeded=%d, want 0",
				def.Key, def.SoldiersNeeded)
		}
	}
	if scoutCount != 3 {
		t.Errorf("expected 3 scouting expeditions, found %d", scoutCount)
	}
}

// TestScoutRuins_LaunchableWithZeroSoldiers guards the Phase 1 bug fix:
// scout_ruins previously needed 2 (nonexistent, pre-iron_age) soldiers and so was
// never launchable. It must now be launchable with zero soldiers given enough
// resources for its Cost.
func TestScoutRuins_LaunchableWithZeroSoldiers(t *testing.T) {
	mm := NewMilitaryManager()
	def := mm.ExpeditionDefByKey("scout_ruins")
	if def == nil {
		t.Fatal("scout_ruins not found")
	}
	if def.SoldiersNeeded != 0 {
		t.Fatalf("scout_ruins SoldiersNeeded=%d, want 0", def.SoldiersNeeded)
	}

	// Resources covering its Cost, zero soldiers.
	resources := map[string]float64{"food": 100, "wood": 100}
	ok, reason := mm.launchability(*def, 0, resources)
	if !ok {
		t.Errorf("scout_ruins not launchable with 0 soldiers + sufficient resources: %q", reason)
	}

	// It should also be available in bronze_age (its MinAge).
	available := mm.GetAvailableExpeditionsByCategory(ExpeditionScouting, "bronze_age", fullAgeOrder())
	foundRuins := false
	for _, d := range available {
		if d.Key == "scout_ruins" {
			foundRuins = true
		}
	}
	if !foundRuins {
		t.Error("scout_ruins not present among available scouting expeditions in bronze_age")
	}
}

// TestMilitaryExpeditions_RequireSoldiers verifies at least one military
// expedition still costs soldiers (the classification didn't accidentally zero
// everyone out).
func TestMilitaryExpeditions_RequireSoldiers(t *testing.T) {
	mm := NewMilitaryManager()
	found := false
	for _, def := range mm.expeditions {
		if def.Category == ExpeditionMilitary && def.SoldiersNeeded > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("no military expedition requires soldiers — classification likely broken")
	}
}
