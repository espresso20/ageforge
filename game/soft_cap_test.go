package game

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// soft_cap_test.go guards the soft cap on the production pools. The pools
// (production_all and every "<res>_rate") were once unbounded above: a boon
// soak measured a x20.3 multiplier on knowledge_rate purely from stacked
// timed events. Then they were clamped at x3.0, and a player with every
// milestone and wonder got nothing for two thirds of them. Now a pool applies
// in full up to +200% and a quarter of every point past it, never under the
// floor: max(productionFloor, 1 + softCap(Σ)).

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

// TestRecalculateRates_ResRatePoolFollowsTheSoftCap: 40 stacked +0.50
// knowledge_rate events would be a x21 multiplier with no cap, and x3 under
// the old clamp. The soft cap makes it 1 + 2 + 18/4 = x7.5.
func TestRecalculateRates_ResRatePoolFollowsTheSoftCap(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral // moraleMultiplier() == 1.0, isolates the pool
	addProductionBuilding(ge, "test_library", "knowledge", 10.0, 1)

	ge.recalculateRates()
	base := ge.Resources.GetRate("knowledge")
	if math.Abs(base-10.0) > 1e-9 {
		t.Fatalf("baseline knowledge rate = %.6f, want 10.0", base)
	}

	stackRateEvents(ge, "knowledge_rate", 0.50, 40) // Σ = +20.0
	ge.recalculateRates()

	got := ge.Resources.GetRate("knowledge")
	if want := base * 7.5; math.Abs(got-want) > 1e-6 {
		t.Fatalf("stacked knowledge_rate gave rate %.6f (x%.2f), want %.6f (x7.50): +200%% in full and a quarter of the other +1800%%",
			got, got/base, want)
	}

	// The raw pool is still reported honestly to the panels; only the
	// APPLIED multiplier goes through the soft cap.
	if sum := ge.buildResolver().AddTotal("knowledge_rate"); sum < 19.0 {
		t.Fatalf("resolver AddTotal(knowledge_rate) = %.4f, want the raw Σ (~20) for the panel", sum)
	}
}

// TestRecalculateRates_ProductionAllPoolFollowsTheSoftCap: the same rule on
// the broad pool. Σ = +12.0 applies 1 + 2 + 10/4 = x5.5.
func TestRecalculateRates_ProductionAllPoolFollowsTheSoftCap(t *testing.T) {
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
	if want := base * 5.5; math.Abs(got-want) > 1e-6 {
		t.Fatalf("stacked production_all gave rate %.6f (x%.2f), want %.6f (x5.50)", got, got/base, want)
	}
}

// TestRecalculateRates_FloorStillHolds: the soft cap is about the top. A
// pathological negative stack still saturates at productionFloor, and the
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

// TestRecalculateRates_ModerateBonusesAreUnaffected: the soft cap must only
// bend the top. A realistic +60% pool passes through untouched.
func TestRecalculateRates_ModerateBonusesAreUnaffected(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = moraleNeutral
	addProductionBuilding(ge, "test_farm", "food", 8.0, 1)
	ge.permanentBonuses["food_rate"] = 0.60

	ge.recalculateRates()
	got := ge.Resources.GetRate("food")
	want := 8.0 * 1.60
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("moderate food_rate bonus gave %.6f, want %.6f (the soft cap must not bend here)", got, want)
	}
}

// TestSoftCapEveryProductionPool: one rule for all production and for every
// resource's own pool, read below, at and above the knee. Each row is a pool
// total and the multiplier a rate must show for it.
func TestSoftCapEveryProductionPool(t *testing.T) {
	rows := []struct{ earned, factor float64 }{
		{0.50, 1.50},   // well under
		{1.99, 2.99},   // just under
		{2.00, 3.00},   // at the knee: all of it
		{2.04, 3.01},   // just over: a quarter of the +4%
		{3.20, 3.30},   // the worked example: +320% applies +230%
		{6.61, 4.1525}, // everything a run can hold in the last age
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
			if got, want := ge.Resources.GetRate(res), 10.0*row.factor; math.Abs(got-want) > 1e-9 {
				t.Errorf("%s at %+.0f%%: %s rate %.6f (x%.4f), want x%.4f", pool, row.earned*100, res, got, got/10, row.factor)
			}
			p := ge.bonusPoolLocked(ge.buildResolver(), pool)
			if want := row.factor - 1; math.Abs(p.Applied-want) > 1e-9 || math.Abs(p.Earned-row.earned) > 1e-9 {
				t.Errorf("%s at %+.0f%%: the pool reads earned %v, applied %v, want applied %v", pool, row.earned*100, p.Earned, p.Applied, want)
			}
			if past := row.earned > 2; p.Limited != past || p.Soft != past {
				t.Errorf("%s at %+.0f%%: limited %v, soft %v, want both %v", pool, row.earned*100, p.Limited, p.Soft, past)
			}
		}
	}
}

// TestSoftCapIsContinuousAndNeverFalls: the engine's factor for a growing
// pool never drops and never jumps, across the knee included.
func TestSoftCapIsContinuousAndNeverFalls(t *testing.T) {
	soft := NewGameEngine().rules.SoftCap()
	const step = 0.001
	prev := poolFactor("production_all", -1.5, soft)
	for i := 1; i <= 9000; i++ {
		earned := -1.5 + float64(i)*step
		f := poolFactor("production_all", earned, soft)
		if f < prev {
			t.Fatalf("the factor falls from %v to %v as the pool grows to %+.1f%%", prev, f, earned*100)
		}
		if f-prev > step+1e-12 {
			t.Fatalf("the factor jumps by %v at %+.1f%% for a step of %v", f-prev, earned*100, step)
		}
		prev = f
	}
}

// TestTimedBonusesFollowTheSoftCap: a festival and an Age of Plenty (the
// era event that doubles production) join the all-production pool like any
// bonus. Under the knee they count in full;
// past it they add a quarter, and they add nothing less.
func TestTimedBonusesFollowTheSoftCap(t *testing.T) {
	festival := config.Effect{Type: "production_all", Target: "production_all", Value: festivalBuffPercent}
	plenty := config.Effect{Type: "production_all", Target: "production_all", Value: 1.0}
	for _, c := range []struct {
		name    string
		held    float64
		effects []config.Effect
		factor  float64
	}{
		{"a festival under the knee", 1.00, []config.Effect{festival}, 2.20},
		{"a festival across the knee", 1.90, []config.Effect{festival}, 3.025},
		{"a festival past the knee", 2.40, []config.Effect{festival}, 3.15},
		{"a festival and an Age of Plenty across the knee", 1.45, []config.Effect{festival, plenty}, 3.1625},
		{"a festival and an Age of Plenty past the knee", 2.40, []config.Effect{festival, plenty}, 3.40},
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

// TestMilestonePastTheKneeRaisesProduction: a milestone earned with the
// all-production pool already past +200% is not wasted. Wonder Builder's
// +5% all production, granted through the game's own milestone check, moves
// every rate by a quarter of it, and the log says so.
func TestMilestonePastTheKneeRaisesProduction(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.morale = moraleNeutral
	addProductionBuilding(ge, "test_farm", "food", 10.0, 1)
	ge.permanentBonuses["production_all"] = 3.0 // +300%: past the knee
	ge.checkMilestones()                        // whatever a new game completes at once
	ge.recalculateRates()
	before := ge.Resources.GetRate("food")
	pool := ge.bonusPoolLocked(ge.buildResolver(), "production_all")
	if !pool.Soft {
		t.Fatalf("the all-production pool is not past its knee: %+v", pool)
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
	if want := 10.0 * (0.05 + wonderAll) * 0.25; math.Abs(gained-want) > 1e-9 {
		t.Errorf("the milestone moved the food rate by %+.6f, want %+.6f: a quarter of what was earned past the knee", gained, want)
	}
	if gained <= 0 {
		t.Errorf("a milestone earned past the knee did not raise production (%+.6f)", gained)
	}
	var said bool
	for _, e := range ge.log[logged:] {
		if strings.Contains(e.Message, "+5% all production counts a quarter past +200%: +1.25% now.") {
			said = true
		}
	}
	if !said {
		var lines []string
		for _, e := range ge.log[logged:] {
			lines = append(lines, e.Message)
		}
		t.Errorf("the log does not say the reward counts a quarter:\n%s", strings.Join(lines, "\n"))
	}
}
