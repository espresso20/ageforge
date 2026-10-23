package game

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// production_pool_test.go guards the production pools. The pools
// (production_all and every "<res>_rate") have no ceiling: a pool applies in
// full, however large it grows, and is never under the floor:
// max(productionFloor, 1 + Σ). (They were once clamped at x3.0, and then bent
// at +200%; a boon soak measured x20.3 on knowledge_rate from stacked timed
// events before either, and the cap on concurrent boons is what bounds that.)

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

// TestRecalculateRates_ResRatePoolAppliesInFull: 40 stacked +0.50
// knowledge_rate events are a x21 multiplier, all of it.
func TestRecalculateRates_ResRatePoolAppliesInFull(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral // moraleMultiplier() == 1.0, isolates the pool
	addProductionBuilding(ge, "test_library", "knowledge", 10.0, 1)

	ge.recalculateRates()
	base := resRate(ge, "knowledge")
	if math.Abs(base-10.0) > 1e-9 {
		t.Fatalf("baseline knowledge rate = %.6f, want 10.0", base)
	}

	stackRateEvents(ge, "knowledge_rate", 0.50, 40) // Σ = +20.0
	ge.recalculateRates()

	got := resRate(ge, "knowledge")
	if want := base * 21; math.Abs(got-want) > 1e-6 {
		t.Fatalf("stacked knowledge_rate gave rate %.6f (x%.2f), want %.6f (x21.00): all of the +2000%%",
			got, got/base, want)
	}

	// The pool reads the same earned and applied.
	if sum := ge.buildResolver().AddTotal("knowledge_rate"); sum < 19.0 {
		t.Fatalf("resolver AddTotal(knowledge_rate) = %.4f, want the raw Σ (~20) for the panel", sum)
	}
}

// TestRecalculateRates_ProductionAllPoolAppliesInFull: the same rule on
// the broad pool. Σ = +12.0 applies x13.
func TestRecalculateRates_ProductionAllPoolAppliesInFull(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral
	addProductionBuilding(ge, "test_farm", "food", 8.0, 1)

	ge.recalculateRates()
	base := ge.Resources.GetRate("food")
	if math.Abs(base-8.0) > 1e-9 {
		t.Fatalf("baseline food rate = %.6f, want 8.0", base)
	}

	stackRateEvents(ge, "production_all", 0.40, 30) // Σ = +12.0
	ge.recalculateRates()

	got := ge.Resources.GetRate("food")
	if want := base * 13; math.Abs(got-want) > 1e-6 {
		t.Fatalf("stacked production_all gave rate %.6f (x%.2f), want %.6f (x13.00)", got, got/base, want)
	}
}

// TestRecalculateRates_FloorStillHolds: a pathological negative stack still
// saturates at productionFloor, and the rate never flips negative.
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

// TestRecalculateRates_ModerateBonusesAreUnaffected: a realistic +60% pool
// passes through untouched.
func TestRecalculateRates_ModerateBonusesAreUnaffected(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral
	addProductionBuilding(ge, "test_farm", "food", 8.0, 1)
	ge.permanentBonuses["food_rate"] = 0.60

	ge.recalculateRates()
	got := ge.Resources.GetRate("food")
	want := 8.0 * 1.60
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("moderate food_rate bonus gave %.6f, want %.6f ", got, want)
	}
}

// TestEveryProductionPoolAppliesInFull: one rule for all production and for
// every resource's own pool, at any size. Each row is a pool total and the
// multiplier a rate must show for it: 1 + the total.
func TestEveryProductionPoolAppliesInFull(t *testing.T) {
	rows := []struct{ earned, factor float64 }{
		{0.50, 1.50},  // modest
		{1.99, 2.99},  // just under +200%
		{2.00, 3.00},  // at +200%
		{2.04, 3.04},  // just over: still all of it
		{3.20, 4.20},  // +320%
		{6.61, 7.61},  // everything a run can hold in the last age
		{99.0, 100.0}, // far past anything a run holds
	}
	for _, pool := range []string{"production_all", "knowledge_rate", "gold_rate", "food_rate"} {
		res := "food"
		if r, ok := strings.CutSuffix(pool, "_rate"); ok {
			res = r
		}
		for _, row := range rows {
			ge := NewGameEngine()
			ge.morale = moraleNeutral
			ge.Resources.UnlockResource(res)
			addProductionBuilding(ge, "test_maker", res, 10.0, 1)
			ge.permanentBonuses[pool] = row.earned
			ge.recalculateRates()
			if got, want := resRate(ge, res), 10.0*row.factor; math.Abs(got-want) > 1e-9 {
				t.Errorf("%s at %+.0f%%: %s rate %.6f (x%.4f), want x%.4f", pool, row.earned*100, res, got, got/10, row.factor)
			}
			p := ge.bonusPoolLocked(ge.buildResolver(), pool)
			if want := row.factor - 1; math.Abs(p.Applied-want) > 1e-9 || math.Abs(p.Earned-row.earned) > 1e-9 {
				t.Errorf("%s at %+.0f%%: the pool reads earned %v, applied %v, want applied %v", pool, row.earned*100, p.Earned, p.Applied, want)
			}
			if p.Limited || CapNote(p, row.earned, true) != "" || PoolNote(p) != "" || PoolLine(p) != "" {
				t.Errorf("%s at %+.0f%%: the pool says something is held back: limited %v, note %q, pool note %q", pool, row.earned*100, p.Limited, CapNote(p, row.earned, true), PoolNote(p))
			}
		}
	}
}

// TestPoolFactorGrowsOneForOne: the engine's factor for a growing pool never
// drops and never bends: past the floor, every point earned is a point of
// multiplier.
func TestPoolFactorGrowsOneForOne(t *testing.T) {
	const step = 0.001
	prev := poolFactor("production_all", -1.5)
	for i := 1; i <= 9000; i++ {
		earned := -1.5 + float64(i)*step
		f := poolFactor("production_all", earned)
		if f < prev {
			t.Fatalf("the factor falls from %v to %v as the pool grows to %+.1f%%", prev, f, earned*100)
		}
		if earned > -0.89 && math.Abs(f-prev-step) > 1e-9 {
			t.Fatalf("the factor moves by %v at %+.1f%% for a step of %v", f-prev, earned*100, step)
		}
		prev = f
	}
}

// TestTimedBonusesCountInFull: a festival and an Age of Plenty (the era
// event that doubles production) join the all-production pool like any
// bonus, and count in full whatever the pool holds.
func TestTimedBonusesCountInFull(t *testing.T) {
	festival := config.Effect{Type: "production_all", Target: "production_all", Value: festivalBuffPercent}
	plenty := config.Effect{Type: "production_all", Target: "production_all", Value: 1.0}
	for _, c := range []struct {
		name    string
		held    float64
		effects []config.Effect
		factor  float64
	}{
		{"a festival on a small pool", 1.00, []config.Effect{festival}, 2.20},
		{"a festival on a pool near +200%", 1.90, []config.Effect{festival}, 3.10},
		{"a festival on a large pool", 2.40, []config.Effect{festival}, 3.60},
		{"a festival and an Age of Plenty", 1.45, []config.Effect{festival, plenty}, 3.65},
		{"a festival and an Age of Plenty on a large pool", 2.40, []config.Effect{festival, plenty}, 4.60},
	} {
		ge := NewGameEngine()
		ge.morale = moraleNeutral
		addProductionBuilding(ge, "test_farm", "food", 10.0, 1)
		ge.permanentBonuses["production_all"] = c.held
		for i, e := range c.effects {
			ge.Events.InjectEvent(ActiveEvent{Key: "timed_" + string(rune('a'+i)), Name: "Timed", TicksLeft: 100, Effects: []config.Effect{e}})
		}
		ge.recalculateRates()
		if got, want := ge.Resources.GetRate("food"), 10.0*c.factor; math.Abs(got-want) > 1e-9 {
			t.Errorf("%s (pool at %+.0f%% before): food rate x%.4f, want x%.4f", c.name, c.held*100, got/10, c.factor)
		}
	}
}

// TestMilestoneOnALargePoolRaisesProductionInFull: a milestone earned with
// the all-production pool already past +200% counts in full. Wonder Builder's
// +5% all production, granted through the game's own milestone check, moves
// every rate by the whole of it, and the log says nothing is held back.
func TestMilestoneOnALargePoolRaisesProductionInFull(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.morale = moraleNeutral
	addProductionBuilding(ge, "test_farm", "food", 10.0, 1)
	ge.permanentBonuses["production_all"] = 3.0 // +300%
	ge.checkMilestones()                        // whatever a new game completes at once
	ge.recalculateRates()
	before := ge.Resources.GetRate("food")
	pool := ge.bonusPoolLocked(ge.buildResolver(), "production_all")
	if pool.Earned <= 2 || pool.Limited {
		t.Fatalf("the all-production pool is not large and unlimited: %+v", pool)
	}

	var wonder string
	for _, key := range ge.Buildings.order {
		if ge.Buildings.defs[key].Category == "wonder" {
			wonder = key
			break
		}
	}
	ge.Buildings.counts[wonder] = 1
	wonderAll := 0.0
	for _, e := range ge.Buildings.defs[wonder].Effects {
		if e.Type == "bonus" && e.Target == "production_all" {
			wonderAll += e.Value
		}
	}
	logged := len(ge.log)
	ge.checkMilestones()
	if !ge.Milestones.IsCompleted("wonder_builder") {
		t.Fatal("building a wonder did not complete Wonder Builder")
	}
	ge.recalculateRates()

	after := ge.bonusPoolLocked(ge.buildResolver(), "production_all")
	reward := after.Earned - pool.Earned - wonderAll
	if math.Abs(reward-0.05) > 1e-9 {
		t.Fatalf("Wonder Builder added %+.4f to the all-production pool, want +0.05", reward)
	}
	gained := ge.Resources.GetRate("food") - before
	if want := 10.0 * (0.05 + wonderAll); math.Abs(gained-want) > 1e-9 {
		t.Errorf("the milestone moved the food rate by %+.6f, want %+.6f: all of what was earned", gained, want)
	}
	for _, e := range ge.log[logged:] {
		if strings.Contains(e.Message, "counts a quarter") || strings.Contains(e.Message, "capped") {
			t.Errorf("the log holds a reward back that counts in full: %q", e.Message)
		}
	}
}
