package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// active_effects_test.go guards PRs C and D of the Active Effects work:
//   - Part D: recalculateRates now APPLIES negative per-resource / gather rate
//     pools (gate→floor), proven by the engine reducing a real production rate.
//   - Part C: the allied-faction trade bonus is emitted as a resolver Modifier so
//     it surfaces in the Active Multipliers panel, without changing apply math.

// addProductionBuilding injects a synthetic worker-free building that produces a
// flat `value × count` of `res` per tick (WorkerCapacity 0 → no fill scaling), so
// recalculateRates has a positive base rate to multiply. Mirrors addWonder's
// approach but for a production effect rather than a bonus effect.
func addProductionBuilding(ge *GameEngine, key, res string, value float64, count int) {
	ge.Buildings.injectDef(config.BuildingDef{
		Key:      key,
		Name:     key,
		Category: "production",
		Effects:  []config.Effect{{Type: "production", Target: res, Value: value}},
	})
	ge.Buildings.counts[key] = count
}

// TestRecalculateRates_NegativeResRateReducesRate is the Part D apply-side guard:
// a negative <res>_rate additive pool must now reduce the resource's rate. Before
// the gate→floor change the engine ran `if bonus > 0 { rate *= 1+bonus }`, so a
// negative pool was silently dropped and this assertion would have FAILED (rate
// unchanged). Morale is pinned neutral (×1.0) to isolate the per-resource factor.
func TestRecalculateRates_NegativeResRateReducesRate(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral // moraleMultiplier() == 1.0
	addProductionBuilding(ge, "test_food_farm", "food", 10.0, 1)

	// Baseline: no rate bonus → flat 10.0/tick.
	ge.recalculateRates()
	base := ge.Resources.GetRate("food")
	if math.Abs(base-10.0) > 1e-9 {
		t.Fatalf("baseline food rate = %.6f, want 10.0", base)
	}

	// Negative per-resource pool: -0.20 → factor max(0.10, 0.80) = 0.80.
	ge.permanentBonuses["food_rate"] = -0.20
	ge.recalculateRates()
	got := ge.Resources.GetRate("food")
	want := 10.0 * 0.80
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("negative food_rate: rate = %.6f, want %.6f (debuff must reduce the rate)", got, want)
	}
	if got >= base {
		t.Fatalf("negative food_rate did not reduce the rate: got %.6f, base %.6f", got, base)
	}
}

// TestRecalculateRates_NegativeGatherRateReducesRate is the Part D apply-side
// guard for the worker output pool (gather_rate). The pool is applied as a
// share of what workers add to the buildings they staff (80% of a building's
// listed rate at a full crew), added on top of the other bonuses, so a
// negative pool pulls a staffed building's output down by that share.
//
// This test used to assert only that the resolver summed the pool and that
// the floored factor came out under 1: the engine applied it to
// WorkerManager.GetProductionRates, which returned nothing, so no rate ever
// moved and the test could not tell. It measures the rate now.
func TestRecalculateRates_NegativeGatherRateReducesRate(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral
	const key = "gathering_camp"
	def := ge.Buildings.defs[key]
	ge.Buildings.counts[key] = 1
	pool := ge.Workers.domains["worker"]
	pool.count, pool.assignments[key] = def.WorkerCapacity, def.WorkerCapacity
	listed := def.Effects[0].Value

	ge.recalculateRates()
	base := ge.Resources.GetRate("food") + ge.Workers.FoodDrain()
	if math.Abs(base-listed) > 1e-9 {
		t.Fatalf("a fully staffed %s makes %.6f food/tick, want its listed %.6f", key, base, listed)
	}

	ge.permanentBonuses["gather_rate"] = -0.30
	if add := ge.buildResolver().AddTotal("gather_rate"); math.Abs(add-(-0.30)) > 1e-9 {
		t.Fatalf("gather_rate AddTotal = %.6f, want -0.30", add)
	}
	ge.recalculateRates()
	got := ge.Resources.GetRate("food") + ge.Workers.FoodDrain()
	want := listed * (1 - 0.30*staffedShare)
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("with -30%% worker output the camp makes %.6f food/tick, want %.6f (its crew's 80%% less 30%%)", got, want)
	}
}

// allyFaction marks a faction allied directly (bypassing the gold cost of
// SetStatus) so GetTradeBonus returns its specialty TradeBonus. Mirrors how the
// diplomacy manager stores state internally.
func allyFaction(ge *GameEngine, key string) {
	ge.Diplomacy.factions[key] = &FactionState{
		Discovered: true,
		Opinion:    50,
		Status:     "allied",
	}
}

// TestDiplomacyModifiers_EmitsOpMulForAlly is the Part C guard: an allied faction
// produces an OpMul Modifier on its specialty <res>_rate, and that Modifier shows
// up in buildResolver().All() (the slice the Active Multipliers panel renders).
func TestDiplomacyModifiers_EmitsOpMulForAlly(t *testing.T) {
	ge := NewGameEngine()
	// merchant_guild: Specialty "gold", TradeBonus 0.20.
	allyFaction(ge, "merchant_guild")

	if got := ge.Diplomacy.GetTradeBonus("gold"); math.Abs(got-0.20) > 1e-9 {
		t.Fatalf("GetTradeBonus(gold) = %.6f, want 0.20 (ally not set up)", got)
	}

	mods := ge.diplomacyModifiers()
	var found *Modifier
	for i := range mods {
		if mods[i].Target == "gold_rate" {
			found = &mods[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("diplomacyModifiers() emitted no gold_rate modifier: %+v", mods)
	}
	if found.Source != "diplomacy" || found.Op != OpMul || math.Abs(found.Value-1.20) > 1e-9 {
		t.Fatalf("gold_rate modifier = %+v, want {Source:diplomacy Op:OpMul Value:1.20}", *found)
	}

	// It must reach the panel via buildResolver().All().
	all := ge.buildResolver().All()
	var inAll bool
	for _, m := range all {
		if m.Source == "diplomacy" && m.Target == "gold_rate" && m.Op == OpMul {
			inAll = true
			break
		}
	}
	if !inAll {
		t.Fatalf("diplomacy gold_rate OpMul not present in buildResolver().All() — panel would not show it")
	}
}

// TestDiplomacyModifiers_FreshEngineEmpty confirms a fresh engine has no allies, so
// GetTradeBonus is 0 and diplomacyModifiers emits nothing — this is why the golden
// fixtures (default fresh engines) are unaffected by Part C.
func TestDiplomacyModifiers_FreshEngineEmpty(t *testing.T) {
	ge := NewGameEngine()
	if got := ge.Diplomacy.GetTradeBonus("gold"); got != 0 {
		t.Fatalf("fresh-engine GetTradeBonus(gold) = %.6f, want 0", got)
	}
	if mods := ge.diplomacyModifiers(); len(mods) != 0 {
		t.Fatalf("fresh-engine diplomacyModifiers() = %+v, want empty", mods)
	}
}

// TestDiplomacyModifiers_DoesNotChangeAppliedRate verifies the no-double-count
// invariant: the engine consumes <res>_rate via AddTotal (OpAdd only), so adding
// an OpMul diplomacy Modifier does NOT change the additive pool the engine
// applies. The panel gains a line; the math does not move.
func TestDiplomacyModifiers_DoesNotChangeAppliedRate(t *testing.T) {
	ge := NewGameEngine()
	allyFaction(ge, "merchant_guild") // gold_rate OpMul ×1.20

	// AddTotal ignores OpMul, so the additive gold_rate pool stays 0.
	if got := ge.buildResolver().AddTotal("gold_rate"); got != 0 {
		t.Fatalf("gold_rate AddTotal with diplomacy OpMul = %.6f, want 0 (OpMul must not enter additive pool)", got)
	}
}
