package game

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/boon"
	"github.com/espresso20/ageforge/config"
)

// faction_malus_test.go covers the capacity mechanic and the malus triggers
// wired into the encounter path: what happens on a FAILED expedition, on an
// encounter with a civ you are AT WAR with, and when your court is already
// holding the maximum number of foreign favours.
//
// Log-line prefixes are the observable contract for which branch fired:
//
//	"[gold]✦"  a positive boon landed
//	"[red]✖"   a setback landed
//	"[gray]✦"  at capacity, nothing happened
//	"[gray]✖"  at war, contact made, neither side came away with anything
const (
	boonLinePrefix     = "[gold]✦"
	malusLinePrefix    = "[red]✖"
	capacityLinePrefix = "[gray]✦"
	standoffLinePrefix = "[gray]✖"
)

// discoverAllFactions marks every roster faction discovered with the given
// status/war flag, so encounters always re-encounter a known civ. Caller must
// hold the write lock.
func discoverAllFactions(ge *GameEngine, status string, atWar bool) {
	for _, def := range config.BaseFactions() {
		ge.Diplomacy.factions[def.Key] = &FactionState{
			Discovered: true,
			Opinion:    0,
			Status:     status,
			AtWar:      atWar,
		}
	}
}

// encounterHarness returns an engine parked in the last age with every faction
// discovered, random events suppressed, and the RNG seeded. Returns it LOCKED —
// the caller unlocks.
func encounterHarness(t *testing.T, seed int64, status string, atWar bool) *GameEngine {
	t.Helper()
	ge := NewGameEngine()
	ge.mu.Lock()
	ge.age = lastAgeKey()
	ge.currentEpoch = config.EpochForAge(ge.age)
	ge.Events.nextEventTick = 1 << 40
	ge.SeedRNG(seed)
	discoverAllFactions(ge, status, atWar)
	return ge
}

// driveEncounters runs n encounter attempts and tallies the outcome lines. The
// at-war standoff line shares the "nothing happened" tally with the at-capacity
// line: both mean the encounter fired and the player got neither gift nor grief.
func driveEncounters(ge *GameEngine, category string, success bool, n int) (boons, maluses, capacity int) {
	for i := 0; i < n; i++ {
		for _, msg := range ge.rollExpeditionEncounter(category, success) {
			switch {
			case strings.HasPrefix(msg, boonLinePrefix):
				boons++
			case strings.HasPrefix(msg, malusLinePrefix):
				maluses++
			case strings.HasPrefix(msg, capacityLinePrefix), strings.HasPrefix(msg, standoffLinePrefix):
				capacity++
			}
		}
	}
	return
}

// --- B4: failed expeditions bring back trouble ------------------------------

func TestEncounter_FailedExpeditionRollsMalus(t *testing.T) {
	ge := encounterHarness(t, 0xFA11ED, "neutral", false)
	defer ge.mu.Unlock()

	boons, maluses, _ := driveEncounters(ge, ExpeditionScouting, false, 3000)

	if maluses == 0 {
		t.Fatal("3000 failed-expedition encounters produced no setbacks — the failure trigger is not wired")
	}
	if boons != 0 {
		t.Fatalf("failed expeditions granted %d positive boons; a botched run must not be rewarded", boons)
	}
}

// --- B4: war bites ----------------------------------------------------------

func TestEncounter_AtWarFactionRollsMalus(t *testing.T) {
	ge := encounterHarness(t, 0x0A7_4A11, "rival", true)
	defer ge.mu.Unlock()

	boons, maluses, _ := driveEncounters(ge, ExpeditionScouting, true, 3000)

	if maluses == 0 {
		t.Fatal("3000 at-war encounters produced no setbacks — war is still a silent skip")
	}
	if boons != 0 {
		t.Fatalf("an at-war faction granted %d positive boons; enemies do not send gifts", boons)
	}
}

// TestEncounter_SuccessUnderCapacityStillGrantsBoons is the control: with none of
// the three sour conditions met, the old happy path is unchanged.
func TestEncounter_SuccessUnderCapacityStillGrantsBoons(t *testing.T) {
	ge := encounterHarness(t, 0x600D, "friendly", false)
	defer ge.mu.Unlock()

	boons, _, _ := driveEncounters(ge, ExpeditionScouting, true, 200)
	if boons == 0 {
		t.Fatal("successful encounters under capacity granted nothing — the happy path regressed")
	}
}

// --- B3: capacity -----------------------------------------------------------

// fillBoonCapacity injects n long-lived faction boon events, occupying capacity
// slots without going through the encounter path.
func fillBoonCapacity(ge *GameEngine, n int) {
	defs := config.BaseFactions()
	for i := 0; i < n; i++ {
		def := defs[i%len(defs)]
		ge.Events.InjectEvent(ActiveEvent{
			Key:       factionBuffKey(def.Key),
			Name:      def.Name + ": Test Favour",
			TicksLeft: 1 << 20,
			Effects:   []config.Effect{{Type: "food_rate", Target: "food", Value: 0.10}},
		})
	}
}

// TestEncounter_AtCapacityRefusesOnlyTimedBoons: once the court is full, a TIMED
// favour is turned away — empty-handed, or occasionally worse. A gift that
// occupies no slot (an instant lump of goods, a work-gang) still lands, because
// capacity is a limit on what you HOLD, not a blanket refusal of everything an
// envoy might carry.
//
// The live-boon count is the load-bearing assertion: the occupying boons never
// expire here and no timed grant may join them, so any drift means a slotted boon
// slipped past the gate.
func TestEncounter_AtCapacityRefusesOnlyTimedBoons(t *testing.T) {
	ge := encounterHarness(t, 0xCA9, "allied", false)
	defer ge.mu.Unlock()

	fillBoonCapacity(ge, MaxConcurrentFactionBoons)
	if got := ge.activeFactionBoonCount(); got != MaxConcurrentFactionBoons {
		t.Fatalf("harness set up %d boons, want %d", got, MaxConcurrentFactionBoons)
	}

	boons, maluses, capacity := driveEncounters(ge, ExpeditionScouting, true, 2000)

	if boons == 0 {
		t.Fatal("at capacity, not one slot-free gift landed over 2000 encounters — " +
			"instant grants and worker loans are being refused along with the timed boons")
	}
	if capacity == 0 {
		t.Fatal("at capacity, no empty-handed line was ever logged — the refusal branch is unreachable")
	}
	if maluses == 0 {
		t.Fatalf("at capacity, atCapacityMalusChance=%.2f never fired over 2000 encounters", atCapacityMalusChance)
	}
	// The boons occupying capacity are long-lived and nothing that landed may take
	// a slot, so the count must not have moved by even one.
	if got := ge.activeFactionBoonCount(); got != MaxConcurrentFactionBoons {
		t.Fatalf("live boon count drifted to %d while at capacity %d — a TIMED boon was granted at capacity",
			got, MaxConcurrentFactionBoons)
	}
}

// TestEncounter_ConcurrentBoonsNeverExceedCapacity drives the real path from
// empty across a long run of encounters (with no expiry ticking at all, the
// harshest case) and asserts the gate holds every single time.
func TestEncounter_ConcurrentBoonsNeverExceedCapacity(t *testing.T) {
	for _, status := range []string{"neutral", "friendly", "allied"} {
		ge := encounterHarness(t, 0xB0BB1E, status, false)

		for i := 0; i < 4000; i++ {
			ge.rollExpeditionEncounter(ExpeditionScouting, true)
			if n := ge.activeFactionBoonCount(); n > MaxConcurrentFactionBoons {
				ge.mu.Unlock()
				t.Fatalf("status %s: encounter %d pushed live boons to %d, cap %d",
					status, i, n, MaxConcurrentFactionBoons)
			}
			if n := ge.activeFactionMalusCount(); n > MaxConcurrentFactionMaluses {
				ge.mu.Unlock()
				t.Fatalf("status %s: encounter %d pushed live setbacks to %d, cap %d",
					status, i, n, MaxConcurrentFactionMaluses)
			}
		}
		ge.mu.Unlock()
	}
}

// TestFactionMalusKeysAreSeparateNamespace: a setback must not consume a boon
// slot, or a run of bad luck would lock you out of rewards.
func TestFactionMalusKeysAreSeparateNamespace(t *testing.T) {
	ge := encounterHarness(t, 0x5E9A_9A7E, "rival", true)
	defer ge.mu.Unlock()

	driveEncounters(ge, ExpeditionScouting, true, 2000)

	if got := ge.activeFactionBoonCount(); got != 0 {
		t.Fatalf("at-war encounters filled %d BOON slots; setbacks belong to the malus namespace", got)
	}
	if got := ge.activeFactionMalusCount(); got == 0 {
		t.Fatal("at-war encounters produced no live setbacks at all")
	}
}

// --- Applier floors ---------------------------------------------------------

// TestDrainResourceFloorsAtZero: a drain can empty a store but never invert it,
// even at a pathological fraction.
func TestDrainResourceFloorsAtZero(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	defer ge.mu.Unlock()

	def := config.BaseFactions()[0]
	a := boonApplier{ge: ge, name: def.Name, key: def.Key, malus: true}

	// Derive the expectation from the ACTUAL stock: food starts seeded and Add
	// clamps at the storage cap, so the post-Add amount is not simply what we
	// added. The invariant under test is "drain removes exactly that fraction of
	// current", which holds whatever the store happens to hold.
	ge.Resources.Add("food", 40)
	start := ge.Resources.Get("food")
	want := start * 0.75
	a.DrainResource("food", 0.25)
	if got := ge.Resources.Get("food"); got < want-0.01 || got > want+0.01 {
		t.Fatalf("25%% drain of %.4f food left %.4f, want %.4f", start, got, want)
	}

	a.DrainResource("food", 1.0)
	if got := ge.Resources.Get("food"); got < 0 || got > 1e-9 {
		t.Fatalf("full drain left %.6f food, want exactly 0", got)
	}

	// Over-fraction, empty store, unknown resource: all no-ops, never negative.
	a.DrainResource("food", 5.0)
	a.DrainResource("food", 0.5)
	a.DrainResource("not_a_resource", 0.5)
	a.DrainResource("", 0.5)
	if got := ge.Resources.Get("food"); got < 0 {
		t.Fatalf("food went negative: %.6f", got)
	}

	for _, rdef := range config.BaseResources() {
		if got := ge.Resources.Get(rdef.Key); got < 0 {
			t.Fatalf("resource %q went negative after drains: %.6f", rdef.Key, got)
		}
	}
}

// TestLoseWorkersFloorsAtZero: worker loss clamps to the live population.
func TestLoseWorkersFloorsAtZero(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	defer ge.mu.Unlock()

	def := config.BaseFactions()[0]
	a := boonApplier{ge: ge, name: def.Name, key: def.Key, malus: true}

	start := ge.Workers.TotalPop()
	a.LoseWorkers(2)
	afterSmall := ge.Workers.TotalPop()
	if afterSmall > start {
		t.Fatalf("LoseWorkers increased population: %d → %d", start, afterSmall)
	}

	a.LoseWorkers(10_000) // far more than exist
	if got := ge.Workers.TotalPop(); got < 0 {
		t.Fatalf("worker pool went negative: %d", got)
	}
	if got := ge.Workers.TotalPop(); got != 0 {
		t.Fatalf("after losing 10000 workers population is %d, want 0", got)
	}

	// Non-positive counts are no-ops.
	a.LoseWorkers(0)
	a.LoseWorkers(-5)
	if got := ge.Workers.TotalPop(); got != 0 {
		t.Fatalf("no-op LoseWorkers changed population to %d", got)
	}
}

// --- Malus profile ----------------------------------------------------------

// TestFactionMalusProfile_SeverityInvertsStanding: standing scales boons UP and
// setbacks DOWN. An enemy's ill will must outweigh an ally's.
func TestFactionMalusProfile_SeverityInvertsStanding(t *testing.T) {
	def := config.BaseFactions()[0]
	const age = "atomic_age"

	sev := func(status string, atWar bool) float64 {
		return factionMalusProfile(def, FactionState{Discovered: true, Status: status, AtWar: atWar}, age).MagnitudeScale
	}

	allied := sev("allied", false)
	friendly := sev("friendly", false)
	neutral := sev("neutral", false)
	rival := sev("rival", false)
	war := sev("rival", true)

	if !(allied < friendly && friendly < neutral && neutral < rival && rival < war) {
		t.Fatalf("malus severity is not monotonic in hostility: allied=%.3f friendly=%.3f neutral=%.3f rival=%.3f war=%.3f",
			allied, friendly, neutral, rival, war)
	}
	if p := factionMalusProfile(def, FactionState{Status: "neutral"}, age); p.Polarity != boon.Negative {
		t.Fatalf("malus profile Polarity = %v, want Negative", p.Polarity)
	}
}

// TestFactionMalusProfile_StrengthScalesSeverity: a mightier civ bites harder.
func TestFactionMalusProfile_StrengthScalesSeverity(t *testing.T) {
	const age = "atomic_age"
	state := FactionState{Discovered: true, Status: "neutral"}

	weak := factionMalusProfile(config.FactionDef{Key: "w", Strength: 1}, state, age).MagnitudeScale
	strong := factionMalusProfile(config.FactionDef{Key: "s", Strength: 5}, state, age).MagnitudeScale
	if strong <= weak {
		t.Fatalf("str-5 severity %.3f is not greater than str-1 %.3f", strong, weak)
	}

	// Out-of-range strengths clamp rather than running away.
	low := factionMalusProfile(config.FactionDef{Key: "l", Strength: -3}, state, age).MagnitudeScale
	high := factionMalusProfile(config.FactionDef{Key: "h", Strength: 99}, state, age).MagnitudeScale
	if low != weak || high != strong {
		t.Fatalf("strength clamp broken: low=%.3f (want %.3f) high=%.3f (want %.3f)", low, weak, high, strong)
	}
}
