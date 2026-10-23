package game

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// stageSoldierProduction stands and fully staffs `numBarracks` barracks, whose
// soldiers move in. Caller must NOT hold the engine lock. Food is topped up and
// morale pinned by the caller's tick helper.
func stageSoldierProduction(t *testing.T, ge *GameEngine, numBarracks int) {
	t.Helper()
	const barracks = "barracks"

	ge.mu.Lock()
	ge.Buildings.counts[barracks] = numBarracks
	ge.Buildings.unlocked[barracks] = true
	// Food enough that the food margin lets the soldiers take their posts.
	ge.Buildings.counts["gathering_camp"] = 60
	ge.Resources.UnlockResource("food")
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
// soldiers-trained counter increments by the soldiers that arrive with a new
// military building.
func TestSoldiersTrained_AccruesAcrossTicks(t *testing.T) {
	ge := NewGameEngine()
	stageSoldierProduction(t, ge, 2)

	if got := ge.Stats.SoldiersTrained; got != 0 {
		t.Fatalf("SoldiersTrained before any tick = %v, want 0", got)
	}

	for i := 0; i < 2*staffEveryTicks; i++ {
		tickWithFood(ge)
	}

	trained := ge.Stats.SoldiersTrained
	if trained <= 0 {
		t.Fatalf("SoldiersTrained after two staffing passes = %v, want > 0", trained)
	}

	// The cumulative counter should track the soldiers at home while nothing
	// has been spent.
	if soldiers := float64(ge.soldiersHome()); trained < soldiers {
		t.Errorf("SoldiersTrained=%v should be >= soldiers at home=%v", trained, soldiers)
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
	for i := 0; i < 2*staffEveryTicks; i++ {
		tickWithFood(ge)
	}
	trainedBefore := ge.Stats.SoldiersTrained
	if trainedBefore <= 0 {
		t.Fatalf("expected SoldiersTrained > 0 after warmup, got %v", trainedBefore)
	}

	// Put us in iron_age with an army at home and launch a military grand
	// mission (a party of 129 here). Sending a party must not touch the
	// lifetime counter.
	setAge(ge, "iron_age")
	setSoldiers(ge, 1000)

	if err := ge.LaunchExpedition(plantOffer(ge, TierGrand, GrandMilitary)); err != nil {
		t.Fatalf("LaunchExpedition(military grand) error: %v", err)
	}
	if got := float64(ge.soldiersHome()); got != 871 {
		t.Fatalf("soldiers at home after launch = %v, want 871 (1000 less the party of 129)", got)
	}

	if got := ge.Stats.SoldiersTrained; got != trainedBefore {
		t.Errorf("SoldiersTrained changed by sending a party: before=%v after=%v (must be unchanged)", trainedBefore, got)
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

// TestCanLaunch_RespectsCost verifies the board snapshot's CanLaunch flag
// accounts for the cost: a short mission (stone and wood, no soldiers) is not
// launchable without the materials and launchable with them.
func TestCanLaunch_RespectsCost(t *testing.T) {
	find := func(ge *GameEngine, n int) (ExpeditionInfo, bool) {
		for _, exp := range ge.GetState().Military.Expeditions {
			if exp.Number == n {
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
	n := plantOffer(ge, TierShort, "")
	scout, ok := find(ge, n)
	if !ok {
		t.Fatalf("offer %d not present on the board at stone_age", n)
	}
	if scout.CanLaunch {
		t.Errorf("short offer CanLaunch=true with no stone/wood in store, want false")
	}
	if scout.LaunchBlockReason == "" {
		t.Errorf("short offer blocked but LaunchBlockReason empty")
	}

	// --- Affordable: enough stone + wood. ---
	ge2 := NewGameEngine()
	setAge(ge2, "stone_age")
	setSoldiers(ge2, 100)
	setResource(ge2, "stone", 1000)
	setResource(ge2, "wood", 1000)
	setIncome(ge2, "stone", 0.5)
	setIncome(ge2, "wood", 0.5)
	n2 := plantOffer(ge2, TierShort, "")
	exp, found := find(ge2, n2)
	if !found {
		t.Fatalf("offer %d not present on the board with resources", n2)
	}
	if !exp.CanLaunch {
		t.Errorf("short offer CanLaunch=false with 1000 stone + 1000 wood, want true (reason: %q)", exp.LaunchBlockReason)
	}
	if exp.LaunchBlockReason != "" {
		t.Errorf("short offer launchable but LaunchBlockReason=%q, want empty", exp.LaunchBlockReason)
	}
}

// === Scouting / Army split (Phase 1) ===

// TestExpeditionCategory_AllValid verifies every offer is classified as one
// of the two known categories, and that only a military grand offer is the
// military one.
func TestExpeditionCategory_AllValid(t *testing.T) {
	ge := NewGameEngine()
	ge.EnsureMissionBoard()
	for i, slot := range ge.Military.board.Slots {
		o := slot.Offer
		want := ExpeditionScouting
		if slotTier(i) == TierGrand && o.Grand == GrandMilitary {
			want = ExpeditionMilitary
		}
		if got := o.Category(); got != want {
			t.Errorf("offer %d (%s %s) has category %q, want %q", i+1, o.Tier, o.Grand, got, want)
		}
	}
}

// TestEveryOfferCostsItsParty verifies every offer on the board, in every
// age, costs a party of soldiers: the tier's share of the reference army,
// whatever the town makes of them.
func TestEveryOfferCostsItsParty(t *testing.T) {
	shares := map[string]int{"short": 4, "long": 10, "military": 30, "civil": 12}
	army := config.RefSoldiers(config.BaseBuildings(), config.Ages())
	offers := 0
	for order, age := range config.Ages() {
		ge := NewGameEngine()
		setAge(ge, age.Key)
		ge.EnsureMissionBoard()
		for i, slot := range ge.Military.board.Slots {
			kind := slot.Offer.Tier
			if slot.Offer.Military() {
				kind = "military"
			} else if kind == TierGrand {
				kind = "civil"
			}
			want := (shares[kind]*int(army[order]) + 99) / 100 // none where there is no army
			offers++
			if got := ge.MissionCost(i + 1)["soldiers"]; got != float64(want) {
				t.Errorf("%s: %s costs %v soldiers, want %d", age.Key, slot.Offer.Title, got, want)
			}
		}
	}
	if offers == 0 {
		t.Error("no offers were checked")
	}
}

// TestScoutRuins_LaunchableWithItsParty guards the Phase 1 bug fix: a short
// mission is launchable given the materials its cost asks for and its party at
// home, in every age of the Iron Era.
func TestScoutRuins_LaunchableWithItsParty(t *testing.T) {
	for _, age := range config.EpochByKey()["iron_era"].Ages {
		ge := NewGameEngine()
		setAge(ge, age)
		setSoldiers(ge, 1e7)
		for _, res := range []string{"iron", "stone", "wood", "food", "gold"} {
			setResource(ge, res, 10000)
			setIncome(ge, res, 1)
		}
		n := plantOffer(ge, TierShort, "")
		if ge.MissionCost(n)["soldiers"] < 1 {
			t.Fatalf("a short mission costs no party in %s", age)
		}
		if err := ge.LaunchExpedition(n); err != nil {
			t.Errorf("a short mission not launchable with its party at home + sufficient resources in %s: %v", age, err)
		}
	}
}

// TestMilitaryExpeditions_RequireSoldiers verifies every military grand
// mission costs at least one soldier, even when the town trains none (the
// classification didn't accidentally zero everyone out).
func TestMilitaryExpeditions_RequireSoldiers(t *testing.T) {
	for _, age := range config.Ages() {
		ge := NewGameEngine()
		setAge(ge, age.Key)
		n := plantOffer(ge, TierGrand, GrandMilitary)
		if got := ge.MissionCost(n)["soldiers"]; got < 1 && age.Key != "primitive_age" {
			t.Errorf("%s: a military grand mission costs %v soldiers with no soldier income, want at least 1", age.Key, got)
		}
	}
}

// === Scouting / Army split (Phase 2a): concurrent per-category actives ===

// TestConcurrentExpeditions_ScoutAndMilitary verifies a scouting expedition and
// a military expedition can be active simultaneously (one per category).
func TestConcurrentExpeditions_ScoutAndMilitary(t *testing.T) {
	ge := NewGameEngine()
	stoneEraReady(ge, "bronze_age")

	// scouting_party is scouting (food+wood, no soldiers); raid_the_rival_clan is military. Both available in bronze_age.
	if err := ge.LaunchExpedition(keyOffer(ge, "scouting_party")); err != nil {
		t.Fatalf("LaunchExpedition(scouting_party) error: %v", err)
	}
	if err := ge.LaunchExpedition(keyOffer(ge, "raid_the_rival_clan")); err != nil {
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

// TestSameTierMissionsRunTogether verifies a second mission of the same tier
// goes while the first is out, as long as the soldiers at home cover its
// party; the army is the only limit.
func TestSameTierMissionsRunTogether(t *testing.T) {
	ge := NewGameEngine()
	stoneEraReady(ge, "bronze_age")

	for i := 0; i < 2; i++ {
		if err := ge.LaunchExpedition(keyOffer(ge, "scouting_party")); err != nil {
			t.Fatalf("short mission %d: %v", i+1, err)
		}
		if err := ge.LaunchExpedition(keyOffer(ge, "raid_the_rival_clan")); err != nil {
			t.Fatalf("military expedition %d: %v", i+1, err)
		}
	}
	if got := len(ge.Military.ActiveOf(TierShort)) + len(ge.Military.ActiveOf(TierGrand)); got != 4 {
		t.Errorf("%d missions out, want 4 (two short, two grand)", got)
	}
	// Two short parties of 8 and two campaigns of 60 left 1000 - 136 home.
	if got := float64(ge.soldiersHome()); got != 864 {
		t.Errorf("%v soldiers at home, want 864", got)
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
	if err := ge.LaunchExpedition(keyOffer(ge, "scouting_party")); err != nil {
		t.Fatalf("LaunchExpedition(scouting_party) error: %v", err)
	}
	if err := ge.LaunchExpedition(keyOffer(ge, "raid_the_rival_clan")); err != nil {
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
// survive ActiveForSave + LoadState.
func TestConcurrentExpeditions_SaveRoundTrip(t *testing.T) {
	ge := NewGameEngine()
	stoneEraReady(ge, "bronze_age")

	if err := ge.LaunchExpedition(keyOffer(ge, "scouting_party")); err != nil {
		t.Fatalf("LaunchExpedition(short) error: %v", err)
	}
	if err := ge.LaunchExpedition(keyOffer(ge, "raid_the_rival_clan")); err != nil {
		t.Fatalf("LaunchExpedition(military grand) error: %v", err)
	}

	saved := ge.Military.ActiveForSave()
	if len(saved) != 2 || saved[0].Slot != TierShort || saved[1].Slot != TierGrand {
		t.Fatalf("ActiveForSave = %+v, want the short and grand slots in tier order", saved)
	}

	// Restore into a fresh manager and confirm both slots come back.
	mm := NewMilitaryManager()
	mm.LoadState(MilitarySave{Active: saved})
	if got := mm.ActiveByTier(TierShort); got == nil || got.Title != saved[0].Mission.Title {
		t.Errorf("after LoadState short slot = %+v, want %s", got, saved[0].Mission.Title)
	}
	if got := mm.ActiveByCategory(ExpeditionMilitary); got == nil || got.Title != saved[1].Mission.Title || !got.Military() {
		t.Errorf("after LoadState military slot = %+v, want %s", got, saved[1].Mission.Title)
	}
}

// TestOldActiveFieldsAreIgnoredOnLoad verifies a save from before the board,
// with its active_scout and active_military fields, loads without them: beta
// saves are disposable, so those missions are simply gone.
func TestOldActiveFieldsAreIgnoredOnLoad(t *testing.T) {
	var save MilitarySave
	old := `{"active_scout":{"Key":"scouting_party","Name":"Scout","TicksLeft":7},` +
		`"active_military":{"Key":"raid_the_rival_clan","Soldiers":5,"TicksLeft":9},"completed_count":4}`
	if err := json.Unmarshal([]byte(old), &save); err != nil {
		t.Fatal(err)
	}
	mm := NewMilitaryManager()
	mm.LoadState(save)
	if mm.HasActive() {
		t.Errorf("an old save's missions were loaded: %+v", mm.active)
	}
	if mm.completedCount != 4 {
		t.Errorf("completed count = %d, want 4", mm.completedCount)
	}
}
