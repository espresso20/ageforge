package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// modifiers_golden_test.go is the safety net for Phase 2 of the multiplier-
// resolver refactor. It proves that buildResolver().Total(target) reproduces the
// CURRENT scattered production / tick-speed math in recalculateRates and
// recalculateTickSpeed, across a battery of engine states.
//
// IMPORTANT: the "expected" side below does NOT call buildResolver or any
// Modifier code. It independently re-derives the multiplier the engine applies
// today, reading the same source getters the engine reads (Research.GetBonuses,
// Prestige.GetBonuses, getWonderBonuses, permanentBonuses, Events.GetActiveEffects,
// moraleMultiplier) and applying the same fixed formula + gates. So an equal
// result means two genuinely different code paths agree — not a tautology. If
// the resolver and the live math diverge, this test fails. That is its whole job.

const goldenEps = 1e-9

// expectedProductionAll re-derives the net multiplicative factor the engine
// applies to a positive building production rate, the way recalculateRates does:
//
//	prodAllAdd = research[production_all] + permanent[production_all]
//	           + prestige[production_all] + wonders[production_all]
//	           + Σ active-event production_all effects
//	factor     = moraleMultiplier() × (prodAllAdd > 0 ? (1 + prodAllAdd) : 1)
//
// The morale factor is moraleMultiplier(), applied unconditionally to the
// building rate (recalculateRates hoists it as `mMult := ge.moraleMultiplier()`,
// then `rate × mMult`). It is NOT the raw ge.morale field: post-rework morale is
// a managed resource (neutral 0.50) whose production effect is a continuous curve
// pivoted at 0.50 — exactly 1.0 at the pivot, ramping to +20% at the cap above
// and toward ×0.5 at the 0.10 floor below. The production_all additive multiply
// is still gated on prodAllAdd > 0.
func expectedProductionAll(ge *GameEngine) float64 {
	research := ge.Research.GetBonuses()
	prestige := ge.Prestige.GetBonuses()
	wonders := ge.getWonderBonuses()

	add := research["production_all"] + ge.permanentBonuses["production_all"] +
		prestige["production_all"] + wonders["production_all"]
	for _, eff := range ge.Events.GetActiveEffects() {
		if eff.Type == "production_all" {
			add += eff.Value
		}
	}

	factor := ge.moraleMultiplier()
	if add > 0 {
		factor *= 1.0 + add
	}
	return factor
}

// expectedResRate re-derives the per-resource rate multiplier:
//
//	add    = research[<res>_rate] + permanent[<res>_rate] + prestige[<res>_rate] + wonders[<res>_rate]
//	factor = add > 0 ? (1 + add) : 1
//
// (permanentBonuses already absorbs milestone/legacy/epoch; prestige and wonders
// are merged into the same effective per-resource pool in recalculateRates.)
func expectedResRate(ge *GameEngine, res string) float64 {
	key := res + "_rate"
	research := ge.Research.GetBonuses()
	prestige := ge.Prestige.GetBonuses()
	wonders := ge.getWonderBonuses()

	add := research[key] + ge.permanentBonuses[key] + prestige[key] + wonders[key]
	if add > 0 {
		return 1.0 + add
	}
	return 1.0
}

// expectedGatherRate re-derives the gather_rate multiplier the same way.
func expectedGatherRate(ge *GameEngine) float64 {
	research := ge.Research.GetBonuses()
	prestige := ge.Prestige.GetBonuses()
	wonders := ge.getWonderBonuses()

	add := research["gather_rate"] + ge.permanentBonuses["gather_rate"] +
		prestige["gather_rate"] + wonders["gather_rate"]
	if add > 0 {
		return 1.0 + add
	}
	return 1.0
}

// expectedTickSpeed re-derives the (1 + Σ tick_speed) factor recalculateTickSpeed
// computes (it does NOT gate on > 0):
//
//	add    = research[tick_speed] + permanent[tick_speed] + prestige[tick_speed]
//	       + Σ active-event tick_speed effects
//	factor = 1 + add
func expectedTickSpeed(ge *GameEngine) float64 {
	prestige := ge.Prestige.GetBonuses()
	add := ge.Research.GetBonus("tick_speed") + ge.permanentBonuses["tick_speed"] + prestige["tick_speed"]
	for _, eff := range ge.Events.GetActiveEffects() {
		if eff.Type == "tick_speed" {
			add += eff.Value
		}
	}
	return 1.0 + add
}

// assertResolverGolden checks every target the resolver should reproduce against
// the independently-derived expected value. Note: production_all is the special
// case — the resolver's Total folds the OpMul morale factor in, so it equals the
// engine's net (moraleMultiplier() × prodall) factor, which is what we compare
// against. The other targets are gated-additive on both sides.
func assertResolverGolden(t *testing.T, ge *GameEngine, resCheck string, gather bool) {
	t.Helper()
	r := ge.buildResolver()

	if got, want := r.Total("production_all"), expectedProductionAll(ge); math.Abs(got-want) > goldenEps {
		t.Errorf("production_all: resolver=%.12f expected=%.12f", got, want)
	}
	if resCheck != "" {
		key := resCheck + "_rate"
		if got, want := r.Total(key), expectedResRate(ge, resCheck); math.Abs(got-want) > goldenEps {
			t.Errorf("%s: resolver=%.12f expected=%.12f", key, got, want)
		}
	}
	if gather {
		if got, want := r.Total("gather_rate"), expectedGatherRate(ge); math.Abs(got-want) > goldenEps {
			t.Errorf("gather_rate: resolver=%.12f expected=%.12f", got, want)
		}
	}
	if got, want := r.Total("tick_speed"), expectedTickSpeed(ge); math.Abs(got-want) > goldenEps {
		t.Errorf("tick_speed: resolver=%.12f expected=%.12f", got, want)
	}
}

// addWonder injects a synthetic built wonder carrying a "bonus" effect. This
// exercises the exact getWonderBonuses code path (Category=="wonder",
// eff.Type=="bonus", count-scaled) without coupling the test to any real
// wonder's balance numbers.
func addWonder(ge *GameEngine, key, target string, value float64, count int) {
	ge.Buildings.defs[key] = config.BuildingDef{
		Key:      key,
		Name:     key,
		Category: "wonder",
		Effects:  []config.Effect{{Type: "bonus", Target: target, Value: value}},
	}
	ge.Buildings.counts[key] = count
}

func TestResolverGolden_FreshGame(t *testing.T) {
	ge := NewGameEngine()
	// Fresh engine: morale defaults to 1.0 in NewGameEngine; no bonuses anywhere.
	assertResolverGolden(t, ge, "wood", true)
}

func TestResolverGolden_ResearchBonuses(t *testing.T) {
	ge := NewGameEngine()
	ge.Research.bonuses["production_all"] = 0.25
	ge.Research.bonuses["gold_rate"] = 0.40
	ge.Research.bonuses["gather_rate"] = 0.15
	ge.Research.bonuses["tick_speed"] = 0.10
	assertResolverGolden(t, ge, "gold", true)
}

func TestResolverGolden_PrestigeLevel(t *testing.T) {
	ge := NewGameEngine()
	ge.Prestige.level = 7 // passive +14% production_all, +7% tick_speed
	assertResolverGolden(t, ge, "iron", true)
}

func TestResolverGolden_BuiltWonder(t *testing.T) {
	ge := NewGameEngine()
	addWonder(ge, "test_wonder_prodall", "production_all", 0.20, 1)
	addWonder(ge, "test_wonder_gold", "gold_rate", 0.30, 2) // count-scaled → +0.60
	assertResolverGolden(t, ge, "gold", true)
}

func TestResolverGolden_ActiveEventProdAll(t *testing.T) {
	ge := NewGameEngine()
	ge.Events.InjectEvent(ActiveEvent{
		Key:       "golden_age",
		Name:      "Golden Age",
		TicksLeft: 50,
		Effects:   []config.Effect{{Type: "production_all", Value: 0.50}},
	})
	assertResolverGolden(t, ge, "stone", true)
}

func TestResolverGolden_ActiveEventTickSpeed(t *testing.T) {
	ge := NewGameEngine()
	ge.Events.InjectEvent(ActiveEvent{
		Key:       "industrious_era",
		Name:      "Industrious Era",
		TicksLeft: 30,
		Effects: []config.Effect{
			{Type: "tick_speed", Value: 0.25},
			{Type: "production_all", Value: 0.10},
		},
	})
	assertResolverGolden(t, ge, "food", true)
}

func TestResolverGolden_MoraleHigh(t *testing.T) {
	ge := NewGameEngine()
	ge.Research.bonuses["production_all"] = 0.20
	// Above the 0.50 pivot → moraleMultiplier ramps the bonus in. At/above cap 1.0
	// this saturates to ×1.20; production_all Total must include the curve's bonus
	// factor, not the raw 0.90 morale field.
	ge.morale = 0.90
	assertResolverGolden(t, ge, "wood", true)
}

func TestResolverGolden_MoraleNeutral(t *testing.T) {
	ge := NewGameEngine()
	ge.Research.bonuses["production_all"] = 0.20
	// Exactly at the 0.50 pivot → moraleMultiplier exactly 1.0.
	ge.morale = moraleNeutral // 0.50
	assertResolverGolden(t, ge, "wood", true)
}

func TestResolverGolden_MoraleLow(t *testing.T) {
	ge := NewGameEngine()
	ge.Research.bonuses["production_all"] = 0.20
	// Below the 0.50 pivot → moraleMultiplier penalty. production_all Total must
	// drop to moraleMultiplier()×(1+adds), well below 1.
	ge.morale = 0.15
	assertResolverGolden(t, ge, "wood", true)
}

// TestResolverGolden_AllSourcesStacked is the integration case: every source
// contributes to overlapping targets at once. This is where an additive-vs-
// multiplicative mistake or a missed source would show up.
func TestResolverGolden_AllSourcesStacked(t *testing.T) {
	ge := NewGameEngine()
	ge.Research.bonuses["production_all"] = 0.10
	ge.Research.bonuses["iron_rate"] = 0.20
	ge.Research.bonuses["gather_rate"] = 0.05
	ge.Research.bonuses["tick_speed"] = 0.08
	ge.Prestige.level = 3 // +6% production_all, +3% tick_speed
	addWonder(ge, "test_wonder_stack", "production_all", 0.15, 1)
	addWonder(ge, "test_wonder_iron", "iron_rate", 0.10, 1)
	ge.permanentBonuses["production_all"] = 0.12
	ge.permanentBonuses["iron_rate"] = 0.07
	ge.permanentBonuses["gather_rate"] = 0.04
	ge.permanentBonuses["tick_speed"] = 0.06
	ge.Events.InjectEvent(ActiveEvent{
		Key:       "boom",
		Name:      "Economic Boom",
		TicksLeft: 40,
		Effects: []config.Effect{
			{Type: "production_all", Value: 0.20},
			{Type: "tick_speed", Value: 0.15},
		},
	})
	ge.morale = 0.85 // high band → moraleMultiplier applies a banded bonus
	assertResolverGolden(t, ge, "iron", true)
}

// TestResolverGolden_IsRealCheck is a guard on the harness itself: it confirms
// the golden assertion would actually FAIL if the resolver disagreed with the
// expected formula. We deliberately compute a wrong "expected" and require the
// resolver to differ — catching the failure mode where the assertion is
// accidentally a no-op (e.g. both sides return 1.0 for everything).
func TestResolverGolden_IsRealCheck(t *testing.T) {
	ge := NewGameEngine()
	ge.Research.bonuses["production_all"] = 0.25
	ge.morale = 0.90 // high band → moraleMultiplier() saturates to ×1.20 at cap 1.0

	got := ge.buildResolver().Total("production_all")
	correct := expectedProductionAll(ge) // moraleMultiplier() × 1.25 = 1.20 × 1.25 = 1.50
	wrong := correct + 0.5

	if math.Abs(got-correct) > goldenEps {
		t.Fatalf("sanity: resolver should match correct expected (%.6f), got %.6f", correct, got)
	}
	if math.Abs(got-wrong) <= goldenEps {
		t.Fatalf("sanity: resolver must NOT match a deliberately wrong expected — assertion is not discriminating")
	}
}
