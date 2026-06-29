package game

import (
	"math/rand"
	"os"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Ancient Civilization Memory (Trello yn98pTQw).
//
// The trigger roll and tech pick read ge.memoryRand, an injectable *rand.Rand seam.
// rand.NewSource(2).Float64() == 0.167 (< ancientMemoryChance 0.40) -> the cache
// fires; seed 1 -> 0.605 (>= 0.40) -> the cache stays buried. These seeds are stable.
func seedFires() *rand.Rand  { return rand.New(rand.NewSource(2)) } // first Float64 < 0.40
func seedBuried() *rand.Rand { return rand.New(rand.NewSource(1)) } // first Float64 >= 0.40

// setPrestigeLevel bumps the prestige manager to the requested level.
func setPrestigeLevel(ge *GameEngine, level int) {
	for ge.Prestige.GetLevel() < level {
		ge.Prestige.Prestige(0)
	}
}

func TestAncientMemory_FiresEarlyOnQualifyingRun(t *testing.T) {
	ge := NewGameEngine()
	setPrestigeLevel(ge, 1)
	ge.age = "primitive_age"
	ge.memoryRand = seedFires()

	ge.maybeOfferAncientMemory()

	if !ge.ancientMemoryUsed {
		t.Fatal("expected ancientMemoryUsed to be set after a successful roll")
	}
	if ge.pendingMemoryTech == "" {
		t.Fatal("expected a pending memory tech to be offered")
	}
	if _, ok := config.TechByKey()[ge.pendingMemoryTech]; !ok {
		t.Fatalf("offered tech %q is not a real tech", ge.pendingMemoryTech)
	}
}

func TestAncientMemory_NoCacheOnFirstEverRun(t *testing.T) {
	ge := NewGameEngine()
	// Prestige level 0 — there is no previous civilization to remember.
	ge.age = "primitive_age"
	ge.memoryRand = seedFires() // would fire if the gate allowed it

	ge.maybeOfferAncientMemory()

	if ge.ancientMemoryUsed {
		t.Error("cache must not fire on the first-ever run (prestige level 0)")
	}
	if ge.pendingMemoryTech != "" {
		t.Error("no tech should be offered on the first-ever run")
	}
}

func TestAncientMemory_BuriedWhenRollFails(t *testing.T) {
	ge := NewGameEngine()
	setPrestigeLevel(ge, 3)
	ge.age = "stone_age"
	ge.memoryRand = seedBuried() // first Float64 >= 0.40

	ge.maybeOfferAncientMemory()

	if ge.ancientMemoryUsed || ge.pendingMemoryTech != "" {
		t.Error("cache should stay buried when the probability roll fails")
	}
}

func TestAncientMemory_OnlyFiresInEarlyAges(t *testing.T) {
	ge := NewGameEngine()
	setPrestigeLevel(ge, 5)
	ge.age = "iron_age" // not primitive/stone
	ge.memoryRand = seedFires()

	ge.maybeOfferAncientMemory()

	if ge.ancientMemoryUsed || ge.pendingMemoryTech != "" {
		t.Error("cache should only surface in primitive/stone age")
	}
}

// selectMemoryTech: tier of reachable techs scales with prestige level.
func TestAncientMemory_TechSelectionTierGating(t *testing.T) {
	ge := NewGameEngine()
	ageOrder := ge.progress.GetAgeOrder()
	// orders: primitive_age=0, stone_age=1, bronze_age=2, iron_age=3, classical_age=4

	// Low prestige at primitive_age: maxOrder = 0 + 1/2 = 0 -> only primitive techs.
	ge.memoryRand = rand.New(rand.NewSource(7))
	for i := 0; i < 20; i++ {
		key := ge.selectMemoryTech("primitive_age", ageOrder, 1)
		if key == "" {
			t.Fatal("expected a primitive-age candidate at prestige 1")
		}
		o := ageOrder[config.TechByKey()[key].Age]
		if o != 0 {
			t.Fatalf("prestige 1 should only offer primitive techs (order 0), got %q order %d", key, o)
		}
	}

	// High prestige at primitive_age: maxOrder = 0 + 8/2 = 4 -> may reach up to order 4.
	// Confirm the reachable window genuinely extends past the current age.
	ge.memoryRand = rand.New(rand.NewSource(7))
	sawHigher := false
	for i := 0; i < 200; i++ {
		key := ge.selectMemoryTech("primitive_age", ageOrder, 8)
		if key == "" {
			t.Fatal("expected candidates at prestige 8")
		}
		o := ageOrder[config.TechByKey()[key].Age]
		if o < 0 || o > 4 {
			t.Fatalf("prestige 8 from primitive should stay within orders [0,4], got order %d", o)
		}
		if o > 0 {
			sawHigher = true
		}
	}
	if !sawHigher {
		t.Error("high prestige should be able to reach a higher-age tech than the current age")
	}
}

// selectMemoryTech should never offer an already-researched tech.
func TestAncientMemory_SelectionSkipsResearched(t *testing.T) {
	ge := NewGameEngine()
	ageOrder := ge.progress.GetAgeOrder()
	// Mark every primitive-age tech researched, leaving only higher tiers.
	for _, td := range config.Technologies() {
		if ageOrder[td.Age] == 0 {
			ge.Research.researched[td.Key] = true
		}
	}
	ge.memoryRand = rand.New(rand.NewSource(3))
	// At prestige 4 (maxOrder = 0 + 2 = 2) only un-researched stone/bronze remain.
	for i := 0; i < 50; i++ {
		key := ge.selectMemoryTech("primitive_age", ageOrder, 4)
		if key == "" {
			t.Fatal("expected an unresearched candidate")
		}
		if ge.Research.IsResearched(key) {
			t.Fatalf("selected an already-researched tech: %q", key)
		}
	}
}

// Accepting begins a prereq-bypassing research at 50% speed (2x ticks).
func TestAncientMemory_AcceptBypassesPrereqsAndHalvesSpeed(t *testing.T) {
	ge := NewGameEngine()
	// Choose a tech with prerequisites the player does NOT have, and zero knowledge.
	const tech = "agriculture" // requires animal_husbandry -> fire_mastery; bronze_age (order 2)
	def := config.TechByKey()[tech]
	if len(def.Prerequisites) == 0 {
		t.Fatalf("test precondition: %q should have prerequisites", tech)
	}

	ge.pendingMemoryTech = tech
	ge.ancientMemoryUsed = true
	// No knowledge, no prereqs researched, current age primitive — all would normally block.
	if err := ge.AcceptAncientMemory(); err != nil {
		t.Fatalf("AcceptAncientMemory should bypass gates, got error: %v", err)
	}

	if ge.Research.currentTech != tech {
		t.Fatalf("expected %q to be in progress, got %q", tech, ge.Research.currentTech)
	}
	// 50% speed with no research_speed bonus == exactly double the base ticks.
	wantTicks := def.ResearchTicks * 2
	if ge.Research.totalTicks != wantTicks {
		t.Errorf("memory research ticks = %d, want %d (2x for half speed)", ge.Research.totalTicks, wantTicks)
	}
	if ge.Research.ticksLeft != wantTicks {
		t.Errorf("memory research ticksLeft = %d, want %d", ge.Research.ticksLeft, wantTicks)
	}
	// The offer is consumed.
	if ge.pendingMemoryTech != "" {
		t.Error("pendingMemoryTech should be cleared after accept")
	}
	// Prereqs were genuinely skipped (not researched as a side effect).
	for _, p := range def.Prerequisites {
		if ge.Research.IsResearched(p) {
			t.Errorf("prerequisite %q should NOT have been researched by accepting a memory", p)
		}
	}
}

func TestAncientMemory_Decline(t *testing.T) {
	ge := NewGameEngine()
	ge.pendingMemoryTech = "tool_making"
	ge.ancientMemoryUsed = true

	ge.DeclineAncientMemory()

	if ge.pendingMemoryTech != "" {
		t.Error("declining should clear the pending offer")
	}
	if ge.Research.currentTech != "" {
		t.Error("declining should not start any research")
	}
	// Run chance stays consumed (set on offer) — no re-roll.
	if !ge.ancientMemoryUsed {
		t.Error("declining must not refund the run's cache chance")
	}
}

// Only one memory per run: a second trigger is suppressed even with a firing roll.
func TestAncientMemory_OnlyOnePerRun(t *testing.T) {
	ge := NewGameEngine()
	setPrestigeLevel(ge, 2)
	ge.age = "primitive_age"
	ge.memoryRand = seedFires()

	ge.maybeOfferAncientMemory()
	if !ge.ancientMemoryUsed {
		t.Fatal("first trigger should fire")
	}
	first := ge.pendingMemoryTech
	// Simulate the player resolving it.
	ge.pendingMemoryTech = ""

	// Second attempt, still a firing seed — must be suppressed by ancientMemoryUsed.
	ge.memoryRand = seedFires()
	ge.maybeOfferAncientMemory()
	if ge.pendingMemoryTech != "" {
		t.Errorf("second trigger should be suppressed; got new offer %q (first was %q)", ge.pendingMemoryTech, first)
	}
}

// The flag persists across save/load and the in-progress doubled-tick research
// round-trips without a dedicated research-save field.
func TestAncientMemory_PersistsAcrossSaveLoad(t *testing.T) {
	ge := NewGameEngine()
	ge.ancientMemoryUsed = true
	ge.pendingMemoryTech = "pottery"
	// Accept to bake a doubled-tick research into the save.
	if err := ge.AcceptAncientMemory(); err != nil {
		t.Fatalf("accept failed: %v", err)
	}
	wantTicks := config.TechByKey()["pottery"].ResearchTicks * 2

	if err := ge.SaveGame("test_ancient_memory"); err != nil {
		t.Fatalf("SaveGame failed: %v", err)
	}
	defer os.Remove("data/saves/test_ancient_memory.json")

	ge2 := NewGameEngine()
	if err := ge2.LoadGame("test_ancient_memory"); err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	if !ge2.ancientMemoryUsed {
		t.Error("ancientMemoryUsed should persist across save/load (true)")
	}
	// Pending offer is intentionally NOT restored (no re-pop after reload).
	if ge2.pendingMemoryTech != "" {
		t.Errorf("pendingMemoryTech should not be restored, got %q", ge2.pendingMemoryTech)
	}
	if ge2.Research.currentTech != "pottery" {
		t.Errorf("in-progress memory research lost across save/load, got %q", ge2.Research.currentTech)
	}
	if ge2.Research.totalTicks != wantTicks {
		t.Errorf("doubled-tick research did not round-trip: got %d, want %d", ge2.Research.totalTicks, wantTicks)
	}
}

// The flag resets on a new prestige run so the next run can roll again.
func TestAncientMemory_ResetsOnNewRun(t *testing.T) {
	ge := NewGameEngine()
	ge.ancientMemoryUsed = true
	ge.pendingMemoryTech = "tool_making"

	// Succumb is a new-run boundary that preserves prestige — it must clear the flag
	// (and may re-offer). Drive it via the reset block by calling Reset(), the simplest
	// full new-run path, then assert the flag cleared.
	ge.Reset()

	if ge.ancientMemoryUsed {
		t.Error("Reset (new game) must clear ancientMemoryUsed so a future run can roll")
	}
	if ge.pendingMemoryTech != "" {
		t.Error("Reset must clear any pending offer")
	}
}
