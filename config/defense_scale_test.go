package config

import (
	"math"
	"testing"
)

// TestMilitaryYardstick pins the military yardstick age by age: what the
// typical player's military power was when the raid threat and the missions
// were measured, what it is on today's tree, and the two scales that put the
// difference back. A change to a military tech moves Now and the scales with
// it; this table then says what moved.
func TestMilitaryYardstick(t *testing.T) {
	order := AgeOrder()
	if len(MilitaryCalibrated) != len(order) {
		t.Fatalf("MilitaryCalibrated lists %d ages, the game has %d", len(MilitaryCalibrated), len(order))
	}
	want := []struct {
		age       string
		then, now float64
	}{
		{"primitive_age", 0, 0}, {"stone_age", 0, 0}, {"bronze_age", 0.20, 0.15},
		{"iron_age", 0.65, 0.45}, {"classical_age", 1.05, 0.60}, {"medieval_age", 1.05, 0.70},
		{"renaissance_age", 1.55, 0.82}, {"colonial_age", 1.85, 0.94}, {"industrial_age", 2.35, 1.06},
		{"victorian_age", 2.35, 1.06}, {"electric_age", 2.35, 1.16}, {"atomic_age", 4.85, 1.36},
		{"modern_age", 4.85, 1.36}, {"information_age", 5.85, 1.46}, {"digital_age", 5.85, 1.56},
		{"cyberpunk_age", 6.85, 1.76}, {"fusion_age", 6.85, 1.86}, {"space_age", 6.85, 1.86},
		{"interstellar_age", 6.85, 1.96}, {"galactic_age", 6.85, 2.06}, {"quantum_age", 6.85, 2.16},
		{"transcendent_age", 6.85, 2.16},
	}
	if len(want) != len(order) {
		t.Fatalf("the table lists %d ages, the game has %d", len(want), len(order))
	}
	prev := 0.0
	for i, w := range want {
		if order[i] != w.age {
			t.Fatalf("age %d is %s, the table says %s", i, order[i], w.age)
		}
		s := MilitaryScaleAt(i)
		if math.Abs(s.Then-w.then) > 1e-9 || math.Abs(s.Now-w.now) > 1e-9 {
			t.Errorf("%s: the typical player's military power was %.2f and is %.2f, want %.2f and %.2f", w.age, s.Then, s.Now, w.then, w.now)
		}
		// The typical player's Defense Rating against the threat, and what
		// their power takes off a mission, are what they were.
		if got := (1 + s.Now) / s.Threat; math.Abs(got-(1+s.Then)) > 1e-9 {
			t.Errorf("%s: the threat's scale leaves the typical garrison at %.4f of a soldier's rating, want %.4f", w.age, got, 1+s.Then)
		}
		if got := MissionPower(s.Now, i); math.Abs(got-s.Then) > 1e-9 {
			t.Errorf("%s: the typical player's mission power is %.4f, want what it was, %.4f", w.age, got, s.Then)
		}
		// The threat still rises with every age.
		if th := AgeThreat(i); th <= prev {
			t.Errorf("%s: the threat is %.0f, not above the age before's %.0f", w.age, th, prev)
		} else {
			prev = th
		}
	}
	if s := MilitaryScaleAt(0); s.Threat != 1 || s.Mission != 1 || AgeThreat(0) != EarlyThreatPrimitive {
		t.Errorf("the Primitive Age's yardstick is %+v and its threat %v, want no scale and the early threat", s, AgeThreat(0))
	}
	// From the Iron Age on the threat is the measured curve, untouched by
	// the early threats: the base doubled once an age, on the age's scale.
	curve := DefenseThreatBase
	for i := range order {
		if i >= 3 {
			if got, want := AgeThreat(i), curve*MilitaryScaleAt(i).Threat; math.Abs(got-want) > 1e-6*want {
				t.Errorf("%s: the threat is %.0f, want the curve's %.0f", order[i], got, want)
			}
		}
		curve *= DefenseThreatGrowth
	}
	// Before it, the threat is sized to the garrison a moderate set of the
	// age's military buildings trains over the age's pacing target: that
	// garrison blunts about a fifth of a raid, and half of it about an
	// eighth.
	for i, age := range order[1:3] {
		garrison := FlowIncome("soldiers", age) * AgeTargetTicks(age)
		if garrison <= 0 {
			t.Fatalf("%s: a moderate town trains no soldiers", age)
		}
		rating := 2 * garrison * (1 + MilitaryScaleAt(i+1).Now)
		full := DefenseMitigation(rating, AgeThreat(i+1))
		half := DefenseMitigation(rating/2, AgeThreat(i+1))
		if full < 0.17 || full > 0.26 {
			t.Errorf("%s: a moderate garrison of %.0f soldiers blunts %.3f of a raid, want about a fifth", age, garrison, full)
		}
		if half < 0.10 || half > 0.17 {
			t.Errorf("%s: half a moderate garrison blunts %.3f of a raid, want about an eighth", age, half)
		}
	}
	// A mission's odds: the listed difficulty with no army, the floor with a
	// large one.
	if got := MissionDifficulty(0.6, 0); got != 0.6 {
		t.Errorf("a 60%% mission with no military power fails %.2f of the time, want 0.60", got)
	}
	if got := MissionDifficulty(0.6, 10); got != MissionDifficultyFloor {
		t.Errorf("a 60%% mission with an overwhelming army fails %.2f of the time, want the floor %.2f", got, MissionDifficultyFloor)
	}
}
