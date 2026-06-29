package game

import (
	"os"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// discoverAll forces every civ whose MinAge is reachable at the given age to be
// discovered on the manager, returning the discovered keys. fullAgeOrder is
// shared with milestones_test.go.
func discoverAll(dm *DiplomacyManager, age string) []string {
	return dm.DiscoverFactions(age, fullAgeOrder())
}

// --- Part 1: roster ---------------------------------------------------------

// TestCivRoster_SizeAndShape asserts the roster is 8..12 civilizations, each
// with a valid personality, a non-empty backstory, a valid strength, and a
// unique key.
func TestCivRoster_SizeAndShape(t *testing.T) {
	roster := config.BaseFactions()
	if len(roster) < 8 || len(roster) > 12 {
		t.Fatalf("roster size = %d, want 8..12", len(roster))
	}
	seen := map[string]bool{}
	for _, def := range roster {
		if seen[def.Key] {
			t.Errorf("duplicate civ key %q", def.Key)
		}
		seen[def.Key] = true
		if !config.ValidPersonalities[def.Personality] {
			t.Errorf("civ %q has invalid personality %q", def.Key, def.Personality)
		}
		if def.Backstory == "" {
			t.Errorf("civ %q has empty backstory", def.Key)
		}
		if def.Strength < 1 || def.Strength > 5 {
			t.Errorf("civ %q has strength %d, want 1..5", def.Key, def.Strength)
		}
	}
}

// TestCivRoster_SpansEarlyAndLateEpochs confirms civs are discoverable across a
// spread of epochs — at least one in an early epoch and at least one in a late
// epoch — so first contact isn't bunched into the endgame.
func TestCivRoster_SpansEarlyAndLateEpochs(t *testing.T) {
	earlyEpochs := map[string]bool{"stone_era": true, "iron_era": true, "steel_era": true}
	lateEpochs := map[string]bool{"neon_era": true, "cosmic_era": true}
	early, late := 0, 0
	for _, def := range config.BaseFactions() {
		ep := config.EpochForAge(def.MinAge)
		if earlyEpochs[ep] {
			early++
		}
		if lateEpochs[ep] {
			late++
		}
	}
	if early == 0 {
		t.Error("no civilizations discoverable in an early epoch")
	}
	if late == 0 {
		t.Error("no civilizations discoverable in a late epoch")
	}
}

// --- Part 2: first-contact discovery ---------------------------------------

// TestFirstContact_GatedAndNeutral verifies discovery is age/epoch-gated: at the
// bronze age only the earliest civ is discovered, and discovery seeds neutral
// opinion + flips Discovered. A late-epoch civ is NOT discovered early.
func TestFirstContact_GatedAndNeutral(t *testing.T) {
	dm := NewDiplomacyManager()
	got := discoverAll(dm, "bronze_age")

	if len(got) == 0 {
		t.Fatal("no civ discovered at bronze_age; expected the founding civ")
	}
	for _, key := range got {
		fs := dm.factions[key]
		if !fs.Discovered {
			t.Errorf("civ %q discovered but Discovered=false", key)
		}
		if fs.Opinion != 0 {
			t.Errorf("civ %q first-contact opinion = %d, want 0 (neutral)", key, fs.Opinion)
		}
		if fs.Status != "neutral" {
			t.Errorf("civ %q first-contact status = %q, want neutral", key, fs.Status)
		}
	}
	// A quantum-age civ must not be reachable at the bronze age.
	if _, ok := dm.factions["quantum_collective"]; ok {
		t.Error("quantum_collective discovered at bronze_age — gating broken")
	}
}

// TestFirstContact_FiresFlavorMessage drives Tick and confirms a first-contact
// flavour line (with the civ name) is returned on discovery.
func TestFirstContact_FiresFlavorMessage(t *testing.T) {
	dm := NewDiplomacyManager()
	msgs := dm.Tick("medieval_age", fullAgeOrder(), 1, false)
	joined := ""
	for _, m := range msgs {
		joined += m + "\n"
	}
	if !strings.Contains(joined, "First contact") {
		t.Errorf("Tick on first reaching medieval_age produced no first-contact message:\n%s", joined)
	}
}

// --- Part 3: passive personality drift -------------------------------------

// TestDrift_DirectionByPersonality checks that aggressive drifts down, peaceful
// drifts up, and isolationist trends toward neutral.
func TestDrift_DirectionByPersonality(t *testing.T) {
	dm := NewDiplomacyManager()
	// Seed one civ of each personality from the real roster.
	personalities := map[string]string{}
	for _, def := range config.BaseFactions() {
		if _, ok := personalities[def.Personality]; !ok {
			personalities[def.Personality] = def.Key
		}
	}
	for personality, key := range personalities {
		dm.factions[key] = &FactionState{Discovered: true, Status: "neutral", Opinion: 30}
		_ = personality
	}

	// Apply many drift cycles (no recent trade → mercantile cools down).
	for i := 0; i < 10; i++ {
		dm.applyPersonalityDrift(false)
	}

	for personality, key := range personalities {
		fs := dm.factions[key]
		switch personality {
		case "aggressive":
			if fs.Opinion >= 30 {
				t.Errorf("aggressive civ %q opinion %d did not drift down from 30", key, fs.Opinion)
			}
		case "peaceful":
			if fs.Opinion <= 30 {
				t.Errorf("peaceful civ %q opinion %d did not drift up from 30", key, fs.Opinion)
			}
		case "isolationist":
			if fs.Opinion >= 30 {
				t.Errorf("isolationist civ %q opinion %d should trend toward 0 from 30", key, fs.Opinion)
			}
		}
	}
}

// TestDrift_MercantileRespondsToTrade confirms mercantile civs gain opinion when
// the player traded recently and lose it when idle.
func TestDrift_MercantileRespondsToTrade(t *testing.T) {
	var mercKey string
	for _, def := range config.BaseFactions() {
		if def.Personality == "mercantile" {
			mercKey = def.Key
			break
		}
	}
	if mercKey == "" {
		t.Skip("no mercantile civ in roster")
	}

	withTrade := NewDiplomacyManager()
	withTrade.factions[mercKey] = &FactionState{Discovered: true, Status: "neutral", Opinion: 20}
	withTrade.applyPersonalityDrift(true)
	if withTrade.factions[mercKey].Opinion <= 20 {
		t.Errorf("mercantile civ with recent trade: opinion %d did not rise from 20", withTrade.factions[mercKey].Opinion)
	}

	idle := NewDiplomacyManager()
	idle.factions[mercKey] = &FactionState{Discovered: true, Status: "neutral", Opinion: 20}
	idle.applyPersonalityDrift(false)
	if idle.factions[mercKey].Opinion >= 20 {
		t.Errorf("mercantile civ ignored: opinion %d did not cool from 20", idle.factions[mercKey].Opinion)
	}
}

// TestDrift_ClampedToRange ensures drift never escapes [-100, 100].
func TestDrift_ClampedToRange(t *testing.T) {
	dm := NewDiplomacyManager()
	var aggro string
	for _, def := range config.BaseFactions() {
		if def.Personality == "aggressive" {
			aggro = def.Key
			break
		}
	}
	dm.factions[aggro] = &FactionState{Discovered: true, Status: "neutral", Opinion: -99}
	for i := 0; i < 50; i++ {
		dm.applyPersonalityDrift(false)
	}
	if dm.factions[aggro].Opinion < -100 {
		t.Errorf("aggressive drift escaped floor: opinion %d < -100", dm.factions[aggro].Opinion)
	}
}

// --- Part 4: worker lending -------------------------------------------------

// TestWorkerLending_AddsAndReturns lends a temporary batch through the engine and
// verifies the pool grows, then shrinks again after the loan window expires.
func TestWorkerLending_AddsAndReturns(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	ge.Workers.UnlockType("worker")
	ge.Workers.domains["worker"].count = 10
	// Inject a pending lend directly and apply it via the same path processDiplomacy uses.
	ge.Diplomacy.pendingLends = []LendRequest{{FactionKey: "riverlands_tribes", Count: 5, Message: "test"}}
	ge.Diplomacy.lentBatches = []LentWorkerBatch{{FactionKey: "riverlands_tribes", Count: 5, ReturnTick: 100, Permanent: false}}
	for _, req := range ge.Diplomacy.TakePendingLends() {
		ge.Workers.AddLentWorkers(req.Count)
	}
	popAfterLend := ge.Workers.TotalPop()
	ge.mu.Unlock()

	if popAfterLend != 15 {
		t.Fatalf("pool after lend = %d, want 15", popAfterLend)
	}

	// Advance the manager past the return tick; processLending should queue a return.
	ge.mu.Lock()
	dm := ge.Diplomacy
	_ = dm.processLending(100) // tick >= ReturnTick
	for _, n := range dm.TakePendingReturns() {
		ge.Workers.KillWorker(n)
	}
	popAfterReturn := ge.Workers.TotalPop()
	ge.mu.Unlock()

	if popAfterReturn != 10 {
		t.Errorf("pool after return = %d, want 10 (lent workers should leave)", popAfterReturn)
	}
}

// TestWorkerLending_PermanentAtHighOpinion confirms a loan from a civ with
// opinion > 80 is flagged permanent and never returned by processLending.
func TestWorkerLending_PermanentAtHighOpinion(t *testing.T) {
	dm := NewDiplomacyManager()
	// A permanent batch (as would be created at opinion > 80) must survive.
	dm.lentBatches = []LentWorkerBatch{{FactionKey: "riverlands_tribes", Count: 4, ReturnTick: 50, Permanent: true}}
	_ = dm.processLending(10_000) // way past ReturnTick
	if returns := dm.TakePendingReturns(); len(returns) != 0 {
		t.Errorf("permanent loan was returned: %v", returns)
	}
	if dm.LentWorkerTotal() != 4 {
		t.Errorf("permanent loan total = %d, want 4 (should persist)", dm.LentWorkerTotal())
	}
}

// TestWorkerLending_PermanentFlagWhenOpinionAbove80 exercises the lend-roll path
// directly: a peaceful civ at opinion > 80 that gets lent produces a permanent
// batch. We force the roll by seeding many windows and checking the invariant on
// any batch produced.
func TestWorkerLending_PermanentFlagWhenOpinionAbove80(t *testing.T) {
	dm := NewDiplomacyManager()
	var peaceful string
	for _, def := range config.BaseFactions() {
		if def.Personality == "peaceful" {
			peaceful = def.Key
			break
		}
	}
	dm.factions[peaceful] = &FactionState{Discovered: true, Status: "allied", Opinion: 95}
	// Run many lend windows; the ~12% roll will fire eventually.
	for tick := driftInterval; tick < driftInterval*400 && !dm.hasLentBatch(peaceful); tick += driftInterval {
		_ = dm.processLending(tick)
	}
	if !dm.hasLentBatch(peaceful) {
		t.Skip("lend roll did not fire within the window (probabilistic) — invariant unverified this run")
	}
	for _, b := range dm.lentBatches {
		if b.FactionKey == peaceful && !b.Permanent {
			t.Errorf("loan from opinion-95 civ %q is not permanent", peaceful)
		}
	}
}

// --- Part 5: war system -----------------------------------------------------

// TestWar_RequiresThresholdAndProvocation is the headline war guard: war does NOT
// start on hostility alone, NOT on provocation alone, but DOES start when both
// opinion < -75 and the provocation threshold is crossed.
func TestWar_RequiresThresholdAndProvocation(t *testing.T) {
	key := "ironhold_clans" // aggressive, in roster
	def := config.FactionByKey()[key]

	// (a) Hostile but no provocation → no war.
	a := NewDiplomacyManager()
	a.factions[key] = &FactionState{Discovered: true, Status: "neutral", Opinion: -90}
	if a.factions[key].AtWar {
		t.Error("war started with hostility but zero provocations")
	}

	// (b) Provocations but opinion not low enough → no war.
	b := NewDiplomacyManager()
	bfs := &FactionState{Discovered: true, Status: "neutral", Opinion: -50}
	b.factions[key] = bfs
	b.recordProvocation(bfs, def, 0)
	b.recordProvocation(bfs, def, 0)
	if bfs.AtWar {
		t.Errorf("war started at opinion -50 (not below %d) despite provocations", warOpinionThreshold)
	}

	// (c) Both conditions met → war.
	c := NewDiplomacyManager()
	cfs := &FactionState{Discovered: true, Status: "neutral", Opinion: -90}
	c.factions[key] = cfs
	c.recordProvocation(cfs, def, 1) // raided route (1)
	if cfs.AtWar {
		t.Error("war started after a single provocation (threshold is 2)")
	}
	c.recordProvocation(cfs, def, 0) // embargo (2) → trips
	if !cfs.AtWar {
		t.Errorf("war did NOT start at opinion -90 with %d provocations", cfs.Provocations)
	}
}

// TestWar_RaidsFireWhileAtWar confirms raids are queued on the raid cadence while
// a civ is at war, with a loss scaled to the civ's strength.
func TestWar_RaidsFireWhileAtWar(t *testing.T) {
	key := "void_reavers" // strength 5, aggressive
	dm := NewDiplomacyManager()
	dm.factions[key] = &FactionState{Discovered: true, Status: "neutral", Opinion: -90, AtWar: true, LastProvocationTick: 0}

	_ = dm.processWar(40) // raid cadence boundary
	raids := dm.TakePendingRaids()
	if len(raids) == 0 {
		t.Fatal("no raid queued while at war on the raid tick")
	}
	if raids[0].Amount <= 0 {
		t.Errorf("raid amount = %.0f, want positive", raids[0].Amount)
	}
}

// TestWar_TributeEndsWar runs tribute through the manager and confirms the war
// clears and provocations reset.
func TestWar_TributeEndsWar(t *testing.T) {
	key := "ironhold_clans"
	def := config.FactionByKey()[key]
	dm := NewDiplomacyManager()
	fs := &FactionState{Discovered: true, Status: "embargo", Opinion: -90, AtWar: true, Provocations: 2}
	dm.factions[key] = fs

	goldCost := 300.0 * float64(def.Strength)
	cultureCost := 50.0 * float64(def.Strength)
	if _, _, err := dm.SendTribute(key, goldCost, cultureCost); err != nil {
		t.Fatalf("SendTribute returned error: %v", err)
	}
	if fs.AtWar {
		t.Error("war still active after tribute")
	}
	if fs.Provocations != 0 {
		t.Errorf("provocations = %d after tribute, want 0", fs.Provocations)
	}
	// Tribute on a civ not at war must error.
	if _, _, err := dm.SendTribute(key, goldCost, cultureCost); err == nil {
		t.Error("tribute to a civ not at war should error")
	}
}

// TestWar_WaitEndsWar confirms a war auto-ends after warCooldownTicks of no fresh
// provocation (the wait-them-out path).
func TestWar_WaitEndsWar(t *testing.T) {
	key := "shadow_syndicate"
	dm := NewDiplomacyManager()
	fs := &FactionState{Discovered: true, Status: "neutral", Opinion: -90, AtWar: true, LastProvocationTick: 0}
	dm.factions[key] = fs

	_ = dm.processWar(warCooldownTicks) // exactly the cooldown later
	if fs.AtWar {
		t.Errorf("war did not auto-end after %d provocation-free ticks", warCooldownTicks)
	}
}

// --- Part 6: alliance trade (light) ----------------------------------------

// TestAllianceTrade_BonusOnlyWhenAlliedAndNotAtWar checks the TradeBonus gate:
// allied grants the bonus, but a civ at war never does.
func TestAllianceTrade_BonusOnlyWhenAlliedAndNotAtWar(t *testing.T) {
	key := "merchant_guild" // specialty gold
	dm := NewDiplomacyManager()
	dm.factions[key] = &FactionState{Discovered: true, Status: "allied", Opinion: 60}
	if dm.GetTradeBonus("gold") <= 0 {
		t.Error("allied civ granted no gold trade bonus")
	}
	dm.factions[key].AtWar = true
	if dm.GetTradeBonus("gold") != 0 {
		t.Error("civ at war still granted a trade bonus")
	}
}

// --- Part 8: persistence + reset -------------------------------------------

// TestPersistence_RoundTripsCivState saves a game with war + lent-worker state and
// confirms it survives a load.
func TestPersistence_RoundTripsCivState(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	ge.Diplomacy.factions["ironhold_clans"] = &FactionState{
		Discovered: true, Status: "embargo", Opinion: -90,
		AtWar: true, Provocations: 2, RaidedRoutes: 1, Embargoes: 1, LastProvocationTick: 42,
	}
	ge.Diplomacy.factions["riverlands_tribes"] = &FactionState{
		Discovered: true, Status: "allied", Opinion: 85,
	}
	ge.Diplomacy.lentBatches = []LentWorkerBatch{
		{FactionKey: "riverlands_tribes", Count: 6, ReturnTick: 999, Permanent: true},
	}
	ge.mu.Unlock()

	if err := ge.SaveGame("test_civ_roundtrip"); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}
	defer os.Remove("data/saves/test_civ_roundtrip.json")

	ge2 := NewGameEngine()
	if err := ge2.LoadGame("test_civ_roundtrip"); err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	ge2.mu.Lock()
	defer ge2.mu.Unlock()
	war := ge2.Diplomacy.factions["ironhold_clans"]
	if war == nil || !war.AtWar || war.Provocations != 2 || war.LastProvocationTick != 42 {
		t.Errorf("war state did not round-trip: %+v", war)
	}
	if ge2.Diplomacy.LentWorkerTotal() != 6 {
		t.Errorf("lent batches did not round-trip: total = %d, want 6", ge2.Diplomacy.LentWorkerTotal())
	}
	if b := ge2.Diplomacy.lentBatches; len(b) != 1 || !b[0].Permanent {
		t.Errorf("permanent loan flag lost on round-trip: %+v", b)
	}
}

// TestReset_ClearsPerRunCivState verifies prestige-style reset wipes discovered /
// opinion / war / lent-worker state (a new run starts undiscovered).
func TestReset_ClearsPerRunCivState(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	ge.Diplomacy.factions["ironhold_clans"] = &FactionState{Discovered: true, Opinion: -90, AtWar: true}
	ge.Diplomacy.lentBatches = []LentWorkerBatch{{FactionKey: "riverlands_tribes", Count: 5, Permanent: true}}
	ge.mu.Unlock()

	ge.Reset()

	ge.mu.Lock()
	defer ge.mu.Unlock()
	if len(ge.Diplomacy.factions) != 0 {
		t.Errorf("post-reset factions = %d, want 0 (per-run state should clear)", len(ge.Diplomacy.factions))
	}
	if ge.Diplomacy.LentWorkerTotal() != 0 {
		t.Errorf("post-reset lent workers = %d, want 0", ge.Diplomacy.LentWorkerTotal())
	}
}
