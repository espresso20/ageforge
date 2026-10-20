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
	learn(ge, "military_tactics") // the tech that opens campaigns
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
	// siege_campaign (ten minutes of soldier income: 3 soldiers here).
	// Spending must not touch the lifetime counter.
	setAge(ge, "iron_age")
	setSoldiers(ge, 10)
	setIncome(ge, "soldiers", 0.01)

	if err := ge.LaunchExpedition("siege_campaign"); err != nil {
		t.Fatalf("LaunchExpedition(siege_campaign) error: %v", err)
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
// CanLaunch flag accounts for the cost: scouting_party (stone and wood, no
// soldiers) is not launchable without the materials and launchable with them.
func TestCanLaunch_RespectsCost(t *testing.T) {
	find := func(ge *GameEngine) (ExpeditionInfo, bool) {
		for _, exp := range ge.GetState().Military.Expeditions {
			if exp.Key == "scouting_party" {
				return exp, true
			}
		}
		return ExpeditionInfo{}, false
	}

	// --- Not affordable: an income of stone and wood, none in store. ---
	ge := NewGameEngine()
	setAge(ge, "stone_age")
	setResource(ge, "stone", 0)
	setResource(ge, "wood", 0)
	setIncome(ge, "stone", 0.5)
	setIncome(ge, "wood", 0.5)
	scout, ok := find(ge)
	if !ok {
		t.Fatalf("scouting_party not present in available expeditions at stone_age")
	}
	if scout.CanLaunch {
		t.Errorf("scouting_party CanLaunch=true with no stone/wood in store, want false")
	}
	if scout.LaunchBlockReason == "" {
		t.Errorf("scouting_party blocked but LaunchBlockReason empty")
	}

	// --- Affordable: enough stone + wood. ---
	ge2 := NewGameEngine()
	setAge(ge2, "stone_age")
	setResource(ge2, "stone", 1000)
	setResource(ge2, "wood", 1000)
	setIncome(ge2, "stone", 0.5)
	setIncome(ge2, "wood", 0.5)
	exp, found := find(ge2)
	if !found {
		t.Fatalf("scouting_party not present in available expeditions with resources")
	}
	if !exp.CanLaunch {
		t.Errorf("scouting_party CanLaunch=false with 1000 stone + 1000 wood, want true (reason: %q)", exp.LaunchBlockReason)
	}
	if exp.LaunchBlockReason != "" {
		t.Errorf("scouting_party launchable but LaunchBlockReason=%q, want empty", exp.LaunchBlockReason)
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

// TestScoutingExpeditions_SoldierFree verifies every scout and expedition
// (the scouting category) costs zero soldiers whatever the town makes of
// them, and that there are fourteen: a scout and an expedition for each of
// the seven eras. Scouting must be playable before the soldiers resource
// exists.
func TestScoutingExpeditions_SoldierFree(t *testing.T) {
	scoutCount := 0
	for _, age := range config.Epochs() {
		ge := NewGameEngine()
		setAge(ge, age.Ages[0])
		setIncome(ge, "soldiers", 5)
		for _, def := range ge.Military.expeditions {
			if def.Category != ExpeditionScouting || def.Era != age.Key {
				continue
			}
			scoutCount++
			if def.Kind == config.MissionCampaign {
				t.Errorf("campaign %q is in the scouting category", def.Key)
			}
			if got := ge.MissionCost(def.Key)["soldiers"]; got != 0 {
				t.Errorf("scouting expedition %q costs %v soldiers, want 0", def.Key, got)
			}
		}
	}
	if scoutCount != 14 {
		t.Errorf("expected 14 scouting expeditions, found %d", scoutCount)
	}
}

// TestScoutRuins_LaunchableWithZeroSoldiers guards the Phase 1 bug fix: the
// first Iron Era scout, road_survey, is launchable with zero soldiers given
// the materials its cost asks for, in every age of its era.
func TestScoutRuins_LaunchableWithZeroSoldiers(t *testing.T) {
	probe := NewGameEngine()
	def := probe.Military.ExpeditionDefByKey("road_survey")
	if def == nil {
		t.Fatal("road_survey not found")
	}
	for _, age := range config.EpochByKey()[def.Era].Ages {
		ge := NewGameEngine()
		setAge(ge, age)
		for _, res := range []string{"iron", "stone", "wood", "food", "gold"} {
			setResource(ge, res, 10000)
			setIncome(ge, res, 1)
		}
		// Note: no soldiers resource at all.
		if ge.MissionCost("road_survey")["soldiers"] != 0 {
			t.Fatalf("road_survey costs soldiers in %s", age)
		}
		available := ge.Military.GetAvailableExpeditionsByCategory(ExpeditionScouting, age, fullAgeOrder())
		foundRuins := false
		for _, d := range available {
			if d.Key == "road_survey" {
				foundRuins = true
			}
		}
		if !foundRuins {
			t.Errorf("road_survey not present among available scouting expeditions in %s", age)
		}
		if err := ge.LaunchExpedition("road_survey"); err != nil {
			t.Errorf("road_survey not launchable with 0 soldiers + sufficient resources in %s: %v", age, err)
		}
	}
}

// TestMilitaryExpeditions_RequireSoldiers verifies every campaign costs at
// least one soldier, even when the town trains none (the classification
// didn't accidentally zero everyone out).
func TestMilitaryExpeditions_RequireSoldiers(t *testing.T) {
	found := 0
	for _, era := range config.Epochs() {
		ge := NewGameEngine()
		setAge(ge, era.Ages[0])
		for _, def := range ge.Military.expeditions {
			if def.Category != ExpeditionMilitary || def.Era != era.Key {
				continue
			}
			found++
			if got := ge.MissionCost(def.Key)["soldiers"]; got < 1 {
				t.Errorf("campaign %q costs %v soldiers with no soldier income, want at least 1", def.Key, got)
			}
		}
	}
	if found != 7 {
		t.Errorf("found %d campaigns, want 7 (one per era)", found)
	}
}

// === Scouting / Army split (Phase 2a): concurrent per-category actives ===

// TestConcurrentExpeditions_ScoutAndMilitary verifies a scouting expedition and
// a military expedition can be active simultaneously (one per category).
func TestConcurrentExpeditions_ScoutAndMilitary(t *testing.T) {
	ge := NewGameEngine()
	stoneEraReady(ge, "bronze_age")

	// scouting_party is scouting (food+wood, no soldiers); raid_the_rival_clan is military. Both available in bronze_age.
	if err := ge.LaunchExpedition("scouting_party"); err != nil {
		t.Fatalf("LaunchExpedition(scouting_party) error: %v", err)
	}
	if err := ge.LaunchExpedition("raid_the_rival_clan"); err != nil {
		t.Fatalf("LaunchExpedition(raid_the_rival_clan) error: %v", err)
	}

	if ge.Military.ActiveByCategory(ExpeditionScouting) == nil {
		t.Error("scouting slot empty after launching scouting_party")
	}
	if ge.Military.ActiveByCategory(ExpeditionMilitary) == nil {
		t.Error("military slot empty after launching raid_the_rival_clan")
	}

	// Snapshot should surface both as active.
	state := ge.GetState()
	if state.Military.ActiveScout == nil {
		t.Error("Snapshot ActiveScout nil, want scouting_party active")
	}
	if state.Military.ActiveMilitary == nil {
		t.Error("Snapshot ActiveMilitary nil, want raid_the_rival_clan active")
	}
}

// TestPerCategorySingleActive verifies a second expedition of the SAME category
// is rejected while one is running, but the OTHER category is still launchable.
func TestPerCategorySingleActive(t *testing.T) {
	ge := NewGameEngine()
	stoneEraReady(ge, "bronze_age")

	// Launch a scouting expedition. A second scouting one must be rejected, but a
	// military one is allowed.
	if err := ge.LaunchExpedition("scouting_party"); err != nil {
		t.Fatalf("LaunchExpedition(scouting_party) error: %v", err)
	}
	if err := ge.LaunchExpedition("scouting_party"); err == nil {
		t.Error("launching a second scouting expedition should be rejected")
	}
	if err := ge.LaunchExpedition("raid_the_rival_clan"); err != nil {
		t.Fatalf("military expedition should be launchable alongside scouting: %v", err)
	}
	// Now the military slot is occupied: a second military one must be rejected.
	if err := ge.LaunchExpedition("raid_the_rival_clan"); err == nil {
		t.Error("launching a second military expedition should be rejected")
	}
}

// TestExpeditionsTickIndependently verifies both active expeditions tick down and
// complete on their own schedules.
func TestExpeditionsTickIndependently(t *testing.T) {
	ge := NewGameEngine()
	stoneEraReady(ge, "bronze_age")

	// Expedition durations are now RANDOMIZED per launch (Bug RQPHYAHC), so we can no
	// longer key assertions off fixed def durations. Launch both, then pin their
	// remaining ticks to fixed, distinct values — the point of this test is that the
	// two categories tick down and resolve INDEPENDENTLY, the shorter one first.
	if err := ge.LaunchExpedition("scouting_party"); err != nil {
		t.Fatalf("LaunchExpedition(scouting_party) error: %v", err)
	}
	if err := ge.LaunchExpedition("raid_the_rival_clan"); err != nil {
		t.Fatalf("LaunchExpedition(raid_the_rival_clan) error: %v", err)
	}

	ge.mu.Lock()
	ge.Military.ActiveByCategory(ExpeditionMilitary).TicksLeft = 3
	ge.Military.ActiveByCategory(ExpeditionScouting).TicksLeft = 8
	ge.mu.Unlock()

	// 3 ticks: the military expedition (3 left) resolves; scouting (8 left) still runs.
	for i := 0; i < 3; i++ {
		ge.mu.Lock()
		ge.processExpeditions()
		ge.mu.Unlock()
	}
	if ge.Military.ActiveByCategory(ExpeditionMilitary) != nil {
		t.Error("military expedition (3 ticks left) should have resolved after 3 ticks")
	}
	if ge.Military.ActiveByCategory(ExpeditionScouting) == nil {
		t.Error("scouting expedition (8 ticks left) should still be running after 3 ticks")
	}

	// 5 more ticks (8 total): scouting resolves too.
	for i := 0; i < 5; i++ {
		ge.mu.Lock()
		ge.processExpeditions()
		ge.mu.Unlock()
	}
	if ge.Military.ActiveByCategory(ExpeditionScouting) != nil {
		t.Error("scouting expedition should have resolved after 8 ticks total")
	}
	if ge.Military.HasActive() {
		t.Error("no expedition should remain active after both resolved")
	}
}

// TestConcurrentExpeditions_SaveRoundTrip verifies both active expeditions
// survive GetActiveForSave + LoadState.
func TestConcurrentExpeditions_SaveRoundTrip(t *testing.T) {
	ge := NewGameEngine()
	stoneEraReady(ge, "bronze_age")

	if err := ge.LaunchExpedition("scouting_party"); err != nil {
		t.Fatalf("LaunchExpedition(scouting_party) error: %v", err)
	}
	if err := ge.LaunchExpedition("raid_the_rival_clan"); err != nil {
		t.Fatalf("LaunchExpedition(raid_the_rival_clan) error: %v", err)
	}

	scout, military := ge.Military.GetActiveForSave()
	if scout == nil || scout.Key != "scouting_party" {
		t.Fatalf("GetActiveForSave scout = %+v, want scouting_party", scout)
	}
	if military == nil || military.Key != "raid_the_rival_clan" {
		t.Fatalf("GetActiveForSave military = %+v, want raid_the_rival_clan", military)
	}

	// Restore into a fresh manager and confirm both slots come back.
	mm := NewMilitaryManager()
	mm.LoadState(scout, military, 0, nil)
	if got := mm.ActiveByCategory(ExpeditionScouting); got == nil || got.Key != "scouting_party" {
		t.Errorf("after LoadState scouting slot = %+v, want scouting_party", got)
	}
	if got := mm.ActiveByCategory(ExpeditionMilitary); got == nil || got.Key != "raid_the_rival_clan" {
		t.Errorf("after LoadState military slot = %+v, want raid_the_rival_clan", got)
	}
}

// TestLegacyActiveExpedition_MigratesByCategory verifies a pre-Phase-2a save with
// only the deprecated ActiveExpedition field is routed to the correct slot by the
// expedition's Category.
func TestLegacyActiveExpedition_MigratesByCategory(t *testing.T) {
	mm := NewMilitaryManager()

	// Legacy military key → military slot.
	milSave := MilitarySave{ActiveExpedition: &ActiveExpedition{Key: "raid_the_rival_clan", Name: "Raid Bandit Camp", Soldiers: 5, TicksLeft: 7}}
	scout, military := mm.migrateActives(milSave)
	if scout != nil {
		t.Errorf("legacy military expedition leaked into scouting slot: %+v", scout)
	}
	if military == nil || military.Key != "raid_the_rival_clan" {
		t.Errorf("legacy military expedition not routed to military slot: %+v", military)
	}

	// Legacy scouting key → scouting slot.
	scoutSave := MilitarySave{ActiveExpedition: &ActiveExpedition{Key: "scouting_party", Name: "Scout Party", Soldiers: 0, TicksLeft: 11}}
	scout, military = mm.migrateActives(scoutSave)
	if military != nil {
		t.Errorf("legacy scouting expedition leaked into military slot: %+v", military)
	}
	if scout == nil || scout.Key != "scouting_party" {
		t.Errorf("legacy scouting expedition not routed to scouting slot: %+v", scout)
	}

	// New fields present → legacy ignored.
	bothSave := MilitarySave{
		ActiveScout:      &ActiveExpedition{Key: "scouting_party", TicksLeft: 3},
		ActiveMilitary:   &ActiveExpedition{Key: "raid_the_rival_clan", TicksLeft: 4},
		ActiveExpedition: &ActiveExpedition{Key: "raid_the_rival_clan", TicksLeft: 99},
	}
	scout, military = mm.migrateActives(bothSave)
	if scout == nil || scout.Key != "scouting_party" || military == nil || military.Key != "raid_the_rival_clan" {
		t.Errorf("new fields should win over legacy: scout=%+v military=%+v", scout, military)
	}
}
