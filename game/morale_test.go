package game

import (
	"math"
	"testing"
)

const moraleEps = 1e-9

// ---------------------------------------------------------------------------
// moraleMultiplier — banded production curve
// ---------------------------------------------------------------------------

func TestMorale_MultiplierNeutralBandIsOne(t *testing.T) {
	ge := NewGameEngine()
	// Across the whole neutral band the multiplier must be exactly 1.0 so the
	// historic economy baseline is preserved.
	for _, m := range []float64{moraleNeutralLow, 0.30, moraleNeutral, 0.60, moraleNeutralHigh} {
		ge.morale = m
		if got := ge.moraleMultiplier(); math.Abs(got-1.0) > moraleEps {
			t.Errorf("moraleMultiplier(%.3f) = %.6f, want 1.0", m, got)
		}
	}
}

func TestMorale_MultiplierBaselineAtNeutral(t *testing.T) {
	// The load-bearing invariant: at moraleNeutral (0.50) production is ×1.0.
	ge := NewGameEngine()
	ge.morale = moraleNeutral
	if got := ge.moraleMultiplier(); math.Abs(got-1.0) > moraleEps {
		t.Fatalf("at neutral morale multiplier = %.6f, want exactly 1.0 (baseline must be preserved)", got)
	}
}

func TestMorale_MultiplierHighBandRampsToMaxBonus(t *testing.T) {
	ge := NewGameEngine()
	// No wonders → cap is 1.0, which equals the high-band ceiling. At cap the
	// multiplier is the full bonus.
	cap := ge.moraleCap()
	if math.Abs(cap-1.0) > moraleEps {
		t.Fatalf("precondition: fresh engine moraleCap = %.3f, want 1.0", cap)
	}
	ge.morale = cap
	want := 1.0 + moraleMaxBonus
	if got := ge.moraleMultiplier(); math.Abs(got-want) > moraleEps {
		t.Errorf("moraleMultiplier(cap=%.3f) = %.6f, want %.6f", cap, got, want)
	}

	// Midway between neutral-high and cap → half the bonus. Give the engine a
	// wonder so the high band has width to ramp across.
	ge.Buildings.counts["sacred_grove"] = 1 // a wonder → cap = 1.05
	cap = ge.moraleCap()
	if cap <= moraleNeutralHigh {
		t.Fatalf("precondition: cap %.3f must exceed neutralHigh %.3f", cap, moraleNeutralHigh)
	}
	mid := moraleNeutralHigh + (cap-moraleNeutralHigh)/2
	ge.morale = mid
	want = 1.0 + 0.5*moraleMaxBonus
	if got := ge.moraleMultiplier(); math.Abs(got-want) > 1e-6 {
		t.Errorf("moraleMultiplier(mid=%.4f, cap=%.3f) = %.6f, want %.6f", mid, cap, got, want)
	}
}

func TestMorale_MultiplierLowBandRampsToMinMult(t *testing.T) {
	ge := NewGameEngine()
	// At the 0.10 floor the multiplier is moraleMinMult.
	ge.morale = 0.10
	if got := ge.moraleMultiplier(); math.Abs(got-moraleMinMult) > moraleEps {
		t.Errorf("moraleMultiplier(0.10) = %.6f, want %.6f", got, moraleMinMult)
	}

	// Halfway between the floor (0.10) and neutral-low (0.25) → half the penalty.
	mid := 0.10 + (moraleNeutralLow-0.10)/2
	ge.morale = mid
	want := 1.0 - 0.5*(1.0-moraleMinMult)
	if got := ge.moraleMultiplier(); math.Abs(got-want) > 1e-6 {
		t.Errorf("moraleMultiplier(%.4f) = %.6f, want %.6f", mid, got, want)
	}
}

// ---------------------------------------------------------------------------
// Starting morale
// ---------------------------------------------------------------------------

func TestMorale_NewGameStartsNeutral(t *testing.T) {
	ge := NewGameEngine()
	if math.Abs(ge.morale-moraleNeutral) > moraleEps {
		t.Errorf("new game morale = %.4f, want moraleNeutral %.4f", ge.morale, moraleNeutral)
	}
	state := ge.GetState()
	if math.Abs(state.Morale-moraleNeutral) > moraleEps {
		t.Errorf("new game state.Morale = %.4f, want %.4f", state.Morale, moraleNeutral)
	}
}

// ---------------------------------------------------------------------------
// Drift toward neutral
// ---------------------------------------------------------------------------

// tickMoraleN runs the morale update N times under the lock with no economic
// pressures (no food deficit, no military, no idle workers, no morale buildings).
func tickMoraleN(ge *GameEngine, n int) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	for i := 0; i < n; i++ {
		ge.updateMoraleTick()
	}
}

func TestMorale_DriftStaysAtNeutral(t *testing.T) {
	ge := NewGameEngine() // starts at neutral, no buildings
	ge.morale = moraleNeutral
	tickMoraleN(ge, 500)
	if math.Abs(ge.morale-moraleNeutral) > 1e-6 {
		t.Errorf("with no pressures morale drifted from neutral to %.6f, want ~%.4f", ge.morale, moraleNeutral)
	}
}

func TestMorale_DriftDownFromHigh(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = 0.70
	before := ge.morale
	tickMoraleN(ge, 10)
	if ge.morale >= before {
		t.Errorf("morale did not drift down from 0.70: before=%.4f after=%.4f", before, ge.morale)
	}
	// And over many ticks it should settle at neutral, not overshoot below it.
	tickMoraleN(ge, 1000)
	if math.Abs(ge.morale-moraleNeutral) > 1e-6 {
		t.Errorf("morale from 0.70 settled at %.6f, want ~%.4f (no overshoot)", ge.morale, moraleNeutral)
	}
}

func TestMorale_DriftUpFromLow(t *testing.T) {
	ge := NewGameEngine()
	ge.morale = 0.30
	before := ge.morale
	tickMoraleN(ge, 10)
	if ge.morale <= before {
		t.Errorf("morale did not drift up from 0.30: before=%.4f after=%.4f", before, ge.morale)
	}
	tickMoraleN(ge, 1000)
	if math.Abs(ge.morale-moraleNeutral) > 1e-6 {
		t.Errorf("morale from 0.30 settled at %.6f, want ~%.4f (no overshoot)", ge.morale, moraleNeutral)
	}
}

// ---------------------------------------------------------------------------
// Morale-building effect summation
// ---------------------------------------------------------------------------

func TestMorale_BuildingsRaiseMorale(t *testing.T) {
	ge := NewGameEngine()
	// Shrines carry a {Type:"morale", Value:0.0006} effect. Several of them must
	// out-pull the per-tick drift (0.0008) so morale climbs above neutral.
	ge.mu.Lock()
	ge.Buildings.counts["shrine"] = 10 // 10 × 0.0006 = 0.006/tick of lift
	ge.mu.Unlock()

	start := ge.morale
	tickMoraleN(ge, 1)
	if ge.morale <= start {
		t.Fatalf("one tick with 10 shrines: morale %.6f did not rise above start %.6f", ge.morale, start)
	}

	// Over time morale should climb well above neutral (lift > drift).
	tickMoraleN(ge, 100)
	if ge.morale <= moraleNeutral+0.05 {
		t.Errorf("with 10 shrines morale settled at %.4f, want comfortably above neutral", ge.morale)
	}
}

func TestMorale_NoMoraleBuildingsNoLift(t *testing.T) {
	// A building WITHOUT a morale effect must not lift morale. After Stage 2A every
	// faith and culture_arts building carries one, so pick a pure resource producer:
	// gathering_camp (food lineage) has no morale effect.
	ge := NewGameEngine()
	ge.mu.Lock()
	ge.Buildings.counts["gathering_camp"] = 10
	ge.mu.Unlock()
	ge.morale = moraleNeutral
	tickMoraleN(ge, 50)
	if ge.morale > moraleNeutral+1e-6 {
		t.Errorf("gathering_camps (no morale effect) raised morale to %.6f, want ≤ neutral", ge.morale)
	}
}

// ---------------------------------------------------------------------------
// Production baseline preserved at neutral morale
// ---------------------------------------------------------------------------

func TestMorale_ProductionBaselinePreservedAtNeutral(t *testing.T) {
	ge := NewGameEngine()

	ge.mu.Lock()
	// Huts supply population capacity so the workers below can be recruited.
	ge.Buildings.counts["hut"] = 10
	// Build and fully staff a food producer so there is a non-zero building rate.
	ge.Buildings.counts["gathering_camp"] = 3
	ge.Buildings.unlocked["gathering_camp"] = true
	ge.mu.Unlock()

	ge.RecruitWorker("food", 9)
	if _, err := ge.AssignAll("gathering_camp"); err != nil {
		t.Fatalf("AssignAll: %v", err)
	}

	ge.mu.Lock()
	defer ge.mu.Unlock()

	// Raw, morale-free building production straight from the source. The morale
	// multiplier scales exactly this quantity, which the engine records in
	// RateBreakdown.BuildingRate (isolated from worker-upkeep consumption and
	// research/prestige modifiers, so it's the clean thing to assert on).
	raw := ge.Buildings.WorkerScaledProduction(ge.Workers.GetAssignedCount)
	rawFood := raw["food"]
	if rawFood <= 0 {
		t.Fatalf("precondition: expected positive raw food production, got %.6f", rawFood)
	}

	buildingRate := func() float64 {
		r := ge.Resources.resources["food"]
		if r == nil {
			t.Fatal("no food resource")
		}
		return r.Breakdown.BuildingRate
	}

	// At neutral morale the multiplier is 1.0 → building production is the raw
	// baseline, unchanged from before the morale rework.
	ge.morale = moraleNeutral
	if mm := ge.moraleMultiplier(); math.Abs(mm-1.0) > moraleEps {
		t.Fatalf("neutral multiplier = %.6f, want 1.0", mm)
	}
	ge.recalculateRates()
	if got := buildingRate(); math.Abs(got-rawFood) > 1e-9 {
		t.Errorf("neutral-morale food BuildingRate = %.9f, want baseline %.9f (×1.0 must be preserved)", got, rawFood)
	}

	// Low morale (0.10 floor) scales building production down by moraleMinMult.
	ge.morale = 0.10
	ge.recalculateRates()
	if got, want := buildingRate(), rawFood*moraleMinMult; math.Abs(got-want) > 1e-9 {
		t.Errorf("floor-morale food BuildingRate = %.9f, want %.9f (×moraleMinMult)", got, want)
	}
}
