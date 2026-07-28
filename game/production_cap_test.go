package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// production_cap_test.go guards the productionCap ceiling (A1). Before it landed,
// recalculateRates applied both the production_all pool and every "<res>_rate"
// pool as max(productionFloor, 1+Σ) — floored below, UNBOUNDED above. A boon soak
// measured a x20.3 multiplier on knowledge_rate purely from stacked timed events.
// Both pools are now clamp(1+Σ, productionFloor, productionCap).

// stackRateEvents injects n active events each carrying `value` on the given
// effect type, the same shape faction boons and catastrophes use.
func stackRateEvents(ge *GameEngine, effectType string, value float64, n int) {
	for i := 0; i < n; i++ {
		ge.Events.InjectEvent(ActiveEvent{
			Key:       "stack_test",
			Name:      "Stack Test",
			TicksLeft: 1 << 20,
			Effects:   []config.Effect{{Type: effectType, Target: effectType, Value: value}},
		})
	}
}

// TestRecalculateRates_ResRatePoolClampsAtCap: 40 stacked +0.50 knowledge_rate
// events would be a x21 multiplier unclamped. It must saturate at productionCap.
func TestRecalculateRates_ResRatePoolClampsAtCap(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral // moraleMultiplier() == 1.0, isolates the pool
	addProductionBuilding(ge, "test_library", "knowledge", 10.0, 1)

	ge.recalculateRates()
	base := ge.Resources.GetRate("knowledge")
	if math.Abs(base-10.0) > 1e-9 {
		t.Fatalf("baseline knowledge rate = %.6f, want 10.0", base)
	}

	stackRateEvents(ge, "knowledge_rate", 0.50, 40) // Σ = +20.0 → x21 unclamped
	ge.recalculateRates()

	got := ge.Resources.GetRate("knowledge")
	want := base * productionCap
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("stacked knowledge_rate gave rate %.6f (x%.2f), want %.6f (x%.2f) — "+
			"the productionCap ceiling is not binding",
			got, got/base, want, productionCap)
	}

	// The raw pool is still reported honestly to the breakdown panel; only the
	// APPLIED multiplier is capped.
	if sum := ge.buildResolver().AddTotal("knowledge_rate"); sum < 19.0 {
		t.Fatalf("resolver AddTotal(knowledge_rate) = %.4f, want the uncapped Σ (~20) for the panel", sum)
	}
}

// TestRecalculateRates_ProductionAllPoolClampsAtCap: the same ceiling on the
// broad pool.
func TestRecalculateRates_ProductionAllPoolClampsAtCap(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral
	addProductionBuilding(ge, "test_farm", "food", 8.0, 1)

	ge.recalculateRates()
	base := ge.Resources.GetRate("food")
	if math.Abs(base-8.0) > 1e-9 {
		t.Fatalf("baseline food rate = %.6f, want 8.0", base)
	}

	stackRateEvents(ge, "production_all", 0.40, 30) // Σ = +12.0 → x13 unclamped
	ge.recalculateRates()

	got := ge.Resources.GetRate("food")
	want := base * productionCap
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("stacked production_all gave rate %.6f (x%.2f), want %.6f (x%.2f)",
			got, got/base, want, productionCap)
	}
}

// TestRecalculateRates_FloorStillHolds: adding a ceiling must not have cost us
// the floor. A pathological negative stack saturates at productionFloor, and the
// rate never flips negative.
func TestRecalculateRates_FloorStillHolds(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral
	addProductionBuilding(ge, "test_farm", "food", 8.0, 1)

	ge.recalculateRates()
	base := ge.Resources.GetRate("food")

	stackRateEvents(ge, "food_rate", -0.50, 20) // Σ = -10.0 → x-9 unfloored
	ge.recalculateRates()

	got := ge.Resources.GetRate("food")
	want := base * productionFloor
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("stacked negative food_rate gave rate %.6f, want %.6f (productionFloor)", got, want)
	}
	if got <= 0 {
		t.Fatalf("floored rate %.6f is not positive — a debuff stack flipped production negative", got)
	}
}

// TestRecalculateRates_ModerateBonusesAreUnaffected: the clamp must only bind at
// the extremes. A realistic +60% pool passes through untouched.
func TestRecalculateRates_ModerateBonusesAreUnaffected(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral
	addProductionBuilding(ge, "test_farm", "food", 8.0, 1)
	ge.permanentBonuses["food_rate"] = 0.60

	ge.recalculateRates()
	got := ge.Resources.GetRate("food")
	want := 8.0 * 1.60
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("moderate food_rate bonus gave %.6f, want %.6f (clamp must not bind here)", got, want)
	}
}
