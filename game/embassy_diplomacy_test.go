package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestEmbassyBuildings_ExistAgeGatedAndCost verifies the two diplomacy buildings
// are defined, gated at the right ages, and priced as the card specifies.
func TestEmbassyBuildings_ExistAgeGatedAndCost(t *testing.T) {
	byKey := config.BuildingByKey()

	embassy, ok := byKey["embassy"]
	if !ok {
		t.Fatal("embassy building not defined in config")
	}
	if embassy.RequiredAge != "colonial_age" {
		t.Errorf("embassy RequiredAge = %q, want colonial_age", embassy.RequiredAge)
	}
	if embassy.WorkerDomain != "trade" {
		t.Errorf("embassy WorkerDomain = %q, want trade", embassy.WorkerDomain)
	}
	if embassy.WorkerCapacity <= 0 {
		t.Errorf("embassy WorkerCapacity = %d, want > 0", embassy.WorkerCapacity)
	}
	if embassy.BaseCost["gold"] <= 0 || embassy.BaseCost["iron"] <= 0 {
		t.Errorf("embassy must cost gold+iron, got %v", embassy.BaseCost)
	}
	if embassyRate := opinionRate(embassy); embassyRate <= 0 {
		t.Errorf("embassy must have a positive opinion effect, got %v", embassy.Effects)
	}

	grand, ok := byKey["grand_embassy"]
	if !ok {
		t.Fatal("grand_embassy building not defined in config")
	}
	if grand.RequiredAge != "industrial_age" {
		t.Errorf("grand_embassy RequiredAge = %q, want industrial_age", grand.RequiredAge)
	}
	if grand.WorkerDomain != "trade" {
		t.Errorf("grand_embassy WorkerDomain = %q, want trade", grand.WorkerDomain)
	}
	if grand.BaseCost["gold"] <= 0 || grand.BaseCost["steel"] <= 0 {
		t.Errorf("grand_embassy must cost gold+steel, got %v", grand.BaseCost)
	}

	// grand_embassy must be 2× the embassy per-worker opinion rate.
	if got, want := opinionRate(grand), opinionRate(embassy)*2; got != want {
		t.Errorf("grand_embassy opinion rate = %v, want 2× embassy (%v)", got, want)
	}

	// Embassies are not resource producers — no "production" effect should leak
	// into the resource pipeline.
	for _, eff := range embassy.Effects {
		if eff.Type == "production" {
			t.Errorf("embassy should not have a production effect, got %+v", eff)
		}
	}
}

// opinionRate pulls the per-worker opinion value from a building's effects.
func opinionRate(def config.BuildingDef) float64 {
	for _, eff := range def.Effects {
		if eff.Type == "opinion" {
			return eff.Value
		}
	}
	return 0
}

// TestAddPassiveOpinion_AccumulatesSpreadsAndCaps verifies the distribution rule:
// fractional opinion accumulates and spills into the integer Opinion, only
// non-hostile discovered factions receive it, and Opinion is capped at +100.
func TestAddPassiveOpinion_AccumulatesSpreadsAndCaps(t *testing.T) {
	dm := NewDiplomacyManager()
	dm.factions["a"] = &FactionState{Discovered: true, Status: "neutral", Opinion: 0}
	dm.factions["b"] = &FactionState{Discovered: true, Status: "allied", Opinion: 60}
	dm.factions["c"] = &FactionState{Discovered: true, Status: "rival", Opinion: -10}
	dm.factions["d"] = &FactionState{Discovered: true, Status: "embargo", Opinion: -20}
	dm.factions["e"] = &FactionState{Discovered: false, Status: "neutral", Opinion: 0}

	// 2 eligible factions (a, b). 1.0 total → 0.5 each: sub-1.0, so no integer
	// change yet, but the accumulator should carry.
	got := dm.AddPassiveOpinion(1.0)
	if got != 2 {
		t.Fatalf("AddPassiveOpinion returned %d eligible, want 2 (neutral+allied only)", got)
	}
	if dm.factions["a"].Opinion != 0 {
		t.Errorf("after 0.5 accrued, faction a Opinion = %d, want 0 (still sub-1.0)", dm.factions["a"].Opinion)
	}

	// Second call: accumulator reaches 1.0 each → +1 integer opinion.
	dm.AddPassiveOpinion(1.0)
	if dm.factions["a"].Opinion != 1 {
		t.Errorf("after 1.0 accrued, faction a Opinion = %d, want 1", dm.factions["a"].Opinion)
	}
	if dm.factions["b"].Opinion != 61 {
		t.Errorf("allied faction b Opinion = %d, want 61", dm.factions["b"].Opinion)
	}

	// Hostile factions never receive passive opinion.
	if dm.factions["c"].Opinion != -10 {
		t.Errorf("rival faction c Opinion changed to %d, want -10 (no passive gain)", dm.factions["c"].Opinion)
	}
	if dm.factions["d"].Opinion != -20 {
		t.Errorf("embargo faction d Opinion changed to %d, want -20 (no passive gain)", dm.factions["d"].Opinion)
	}
	// Undiscovered factions never receive passive opinion.
	if dm.factions["e"].Opinion != 0 {
		t.Errorf("undiscovered faction e Opinion changed to %d, want 0", dm.factions["e"].Opinion)
	}

	// Cap at +100: dump a huge amount and confirm no overflow.
	dm.factions["b"].Opinion = 99
	dm.AddPassiveOpinion(1000)
	if dm.factions["b"].Opinion != 100 {
		t.Errorf("faction b Opinion = %d after large injection, want capped at 100", dm.factions["b"].Opinion)
	}

	// No eligible factions → returns 0 (e.g. everyone hostile/undiscovered).
	empty := NewDiplomacyManager()
	empty.factions["x"] = &FactionState{Discovered: true, Status: "rival", Opinion: 0}
	if n := empty.AddPassiveOpinion(5); n != 0 {
		t.Errorf("AddPassiveOpinion with no eligible factions returned %d, want 0", n)
	}
}

// TestEmbassyOpinionGain_GrandIsDouble drives the engine's processDiplomacy step
// with a staffed embassy vs a staffed grand_embassy and confirms (a) opinion
// rises over ticks, and (b) the grand_embassy accrues at roughly 2× the rate.
func TestEmbassyOpinionGain_GrandIsDouble(t *testing.T) {
	run := func(buildingKey string) int {
		ge := NewGameEngine()
		setAge(ge, "industrial_age") // industrial: both embassy & grand_embassy buildable, merchant_guild+artisan_league discovered

		ge.mu.Lock()
		// Discover factions for the current age so opinion has somewhere to go.
		ge.Diplomacy.DiscoverFactions(ge.age, ge.progress.GetAgeOrder())
		// Seed one fully-staffed building of the requested type.
		def := config.BuildingByKey()[buildingKey]
		ge.Buildings.counts[buildingKey] = 1
		ge.Workers.domains["worker"].assignments[buildingKey] = def.WorkerCapacity // 100% fill
		ge.mu.Unlock()

		// Advance enough ticks for the fractional accumulator to spill into ints.
		// ge.tick is advanced so the natural opinion drift (which fires on 50/100
		// tick boundaries) isn't applied on every single call — otherwise a tick
		// stuck at 0 would decay opinion back down each iteration.
		for i := 0; i < 200; i++ {
			ge.mu.Lock()
			ge.tick++
			ge.processDiplomacy()
			ge.mu.Unlock()
		}

		// Read total integer opinion across discovered factions.
		ge.mu.Lock()
		total := 0
		for _, fs := range ge.Diplomacy.factions {
			if fs.Discovered {
				total += fs.Opinion
			}
		}
		ge.mu.Unlock()
		return total
	}

	embassyGain := run("embassy")
	grandGain := run("grand_embassy")

	if embassyGain <= 0 {
		t.Fatalf("staffed embassy produced no opinion over 200 ticks (got %d)", embassyGain)
	}
	if grandGain <= embassyGain {
		t.Fatalf("grand_embassy gain (%d) should exceed embassy gain (%d)", grandGain, embassyGain)
	}
	// grand_embassy is 2× the per-worker rate but also has a larger worker
	// capacity (8 vs 5), so total throughput is well above 2×. Just assert it is
	// at least ~1.8× to allow for integer-spill rounding at the tick boundary.
	if float64(grandGain) < float64(embassyGain)*1.8 {
		t.Errorf("grand_embassy gain (%d) should be >= 1.8× embassy gain (%d)", grandGain, embassyGain)
	}
}
