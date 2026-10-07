package config

import (
	"math"
	"testing"
)

// TestSoftCapApplied: a pool applies in full up to the knee and a quarter of
// every point past it. Below, at and above the knee, on numbers a reader can
// check by hand.
func TestSoftCapApplied(t *testing.T) {
	c := ProductionSoftCap()
	if c.Knee != 2 || c.PastKnee != 0.25 {
		t.Fatalf("the production soft cap is %+v, want a knee of +200%% and a quarter past it", c)
	}
	for _, row := range []struct{ earned, applied float64 }{
		{-0.50, -0.50}, // a penalty is no business of the knee
		{0, 0},
		{0.85, 0.85}, // what a first run holds leaving the Industrial Age
		{1.45, 1.45}, // and the Atomic
		{1.99, 1.99},
		{2.00, 2.00},   // at the knee: all of it
		{2.01, 2.0025}, // one point past: a quarter of a point
		{2.40, 2.10},
		{3.20, 2.30},   // the worked example
		{6.61, 3.1525}, // every milestone, wonder and monument in the last age
		{20.0, 6.50},   // a runaway stack still counts, at a quarter
	} {
		if got := c.Applied(row.earned); math.Abs(got-row.applied) > 1e-12 {
			t.Errorf("a pool that has earned %+.2f applies %+.6f, want %+.4f", row.earned, got, row.applied)
		}
	}
	// Under the knee the pool comes back bit for bit: runs that never reach
	// +200% must replay exactly as before the soft cap.
	for _, earned := range []float64{-3, -0.1, 0, 0.05, 0.15, 1.0 / 3, 0.85, 1.3, 1.45, 1.7, 1.85, 2} {
		if got := c.Applied(earned); got != earned {
			t.Errorf("a pool at %v under the knee applies %v, want it untouched", earned, got)
		}
	}
}

// TestSoftCapIsContinuousAndNeverDecreases: no jump at the knee, and more
// earned is never less applied.
func TestSoftCapIsContinuousAndNeverDecreases(t *testing.T) {
	c := ProductionSoftCap()
	const eps = 1e-9
	if below, at, above := c.Applied(c.Knee-eps), c.Applied(c.Knee), c.Applied(c.Knee+eps); at-below > 2*eps || above-at > 2*eps || above < at || at < below {
		t.Errorf("the soft cap jumps at the knee: %v, %v, %v", below, at, above)
	}
	const step = 0.0005
	prev := c.Applied(-2)
	for i := 1; i <= 20000; i++ {
		earned := -2 + float64(i)*step
		got := c.Applied(earned)
		if got < prev {
			t.Fatalf("applied falls from %v to %v as earned grows to %v", prev, got, earned)
		}
		if got-prev > step+1e-12 {
			t.Fatalf("applied jumps by %v at %v for a step of %v", got-prev, earned, step)
		}
		if got > earned+1e-12 {
			t.Fatalf("a pool at %v applies %v, more than it earned", earned, got)
		}
		prev = got
	}
}

// TestSoftCapCounts: what one bonus adds depends on where the pool stands:
// all of it under the knee, a quarter past it, part of each across it. It is
// never nothing and never negative for a bonus.
func TestSoftCapCounts(t *testing.T) {
	c := ProductionSoftCap()
	for _, row := range []struct{ pool, bonus, counts float64 }{
		{1.00, 0.10, 0.10},    // under the knee: in full
		{1.90, 0.10, 0.10},    // lands exactly on the knee: in full
		{1.95, 0.10, 0.0625},  // across: +5% in full, +5% at a quarter
		{2.00, 0.10, 0.025},   // past: a quarter
		{6.51, 0.10, 0.025},   // far past: still a quarter
		{2.40, 0.20, 0.05},    // a festival past the knee
		{2.40, 1.00, 0.25},    // a power surge past the knee
		{1.45, 1.20, 0.7125},  // a festival and a surge on what a first run holds at the Atomic Age
		{2.40, -0.10, -0.025}, // a penalty past the knee costs a quarter too
	} {
		got := c.Counts(row.pool, row.bonus)
		if math.Abs(got-row.counts) > 1e-12 {
			t.Errorf("a bonus of %+.2f on a pool at %+.2f counts %+.6f, want %+.4f", row.bonus, row.pool, got, row.counts)
		}
		if row.bonus > 0 && got <= 0 {
			t.Errorf("a bonus of %+.2f on a pool at %+.2f counts for nothing", row.bonus, row.pool)
		}
	}
}

// TestIncomeFactorFollowsTheSoftCap: the pacing model passes the pool a
// typical player holds through the same rule the engine applies. Ages that
// hold +200% or less are untouched, bit for bit, so nothing measured moves;
// from the Digital Age on the model holds more than +200% and a quarter of
// the rest counts.
func TestIncomeFactorFollowsTheSoftCap(t *testing.T) {
	order := AgeOrder()
	pos := AgePositions(order)
	layer := func(age string) float64 {
		l := 1.0
		for _, tech := range Technologies() {
			if j, ok := pos[tech.Age]; !ok || j > pos[age] {
				continue
			}
			for _, e := range tech.Effects {
				if e.Kind == EffectAllOutput {
					l += e.Value
				}
			}
		}
		return l
	}
	// "zz_none" is a resource no tech names: the factor is the pool's
	// multiplier times the techs' all-production layer.
	factor := func(age string) float64 { return IncomeFactor(Technologies(), pos, age, "zz_none") }
	past := 0
	for _, age := range order {
		held := ProductionAllHeld[age]
		want := 1 + held
		if held > 2 {
			past++
			want = 1 + 2 + (held-2)*0.25
		}
		if got := factor(age); math.Abs(got-want*layer(age)) > 1e-12 {
			t.Errorf("%s: the model holds %+.2f and multiplies income by %v, want %v", age, held, got, want*layer(age))
		}
		if held <= 2 {
			if got := factor(age); got != float64((1+held)*layer(age)) {
				t.Errorf("%s: under the knee the factor is %v, want exactly %v", age, got, (1+held)*layer(age))
			}
		}
	}
	if past == 0 {
		t.Error("no age holds more than +200%: the model never meets the soft cap")
	}
	if got, want := ProductionAllHeld["digital_age"], 2.2; got != want {
		t.Errorf("the Digital Age holds %v, want %v: the first age the model is past the knee", got, want)
	}
	for i := 1; i < len(order); i++ {
		if ProductionAllHeld[order[i]] < ProductionAllHeld[order[i-1]] {
			t.Errorf("%s holds less all production than %s", order[i], order[i-1])
		}
	}
}
