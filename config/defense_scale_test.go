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
		// The threat still rises with every age, in the garrison it was sized
		// against: measured in soldiers at their posts it also follows the
		// reference town's army (ArmyThreatScale), which can dip where a gate
		// asks for more copies of a building with fewer posts each.
		if th := AgeThreat(i) / ArmyThreatScale(i); th <= prev {
			t.Errorf("%s: the threat is %.0f, not above the age before's %.0f", w.age, th, prev)
		} else {
			prev = th
		}
	}
	if s := MilitaryScaleAt(0); s.Threat != 1 || s.Mission != 1 || AgeThreat(0) != float64(EarlyThreatPrimitive*ArmyThreatScale(0)) {
		t.Errorf("the Primitive Age's yardstick is %+v and its threat %v, want no scale and the early threat", s, AgeThreat(0))
	}
	// From the Iron Age on the threat is the measured curve, untouched by
	// the early threats: the base doubled once an age, on the age's scale.
	curve := DefenseThreatBase
	for i := range order {
		if i >= 3 {
			if got, want := AgeThreat(i), curve*MilitaryScaleAt(i).Threat*ArmyThreatScale(i); math.Abs(got-want) > 1e-6*want {
				t.Errorf("%s: the threat is %.0f, want the curve's %.0f", order[i], got, want)
			}
		}
		curve *= DefenseThreatGrowth
	}
	// Before it, the threat is sized to the garrison a moderate town has at
	// its posts: the reference town's military buildings, each with its
	// soldiers. That garrison blunts about a fifth of a raid, and half of it
	// about an eighth.
	posted := RefSoldiers(BaseBuildings(), Ages())
	for i, age := range order[1:3] {
		garrison := posted[i+1]
		if garrison <= 0 {
			t.Fatalf("%s: a moderate town has no soldiers at its posts", age)
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
}

// TestReferenceArmyBluntsWhatTheModerateGarrisonDid: in every age the
// reference town's army at its posts blunts the share of a raid, and half of it
// the share, that the moderate garrison did when soldiers were a stockpile a
// town's buildings made (recorded at 6b3a4162).
func TestReferenceArmyBluntsWhatTheModerateGarrisonDid(t *testing.T) {
	was := [][2]float64{
		{0, 0}, {0.213158, 0.139655}, {0.238056, 0.161834}, {0.188970, 0.119596}, {0.254449, 0.177371},
		{0.283329, 0.206753}, {0.332867, 0.264118}, {0.354662, 0.292659}, {0.408047, 0.373250},
		{0.414129, 0.383555}, {0.422115, 0.397484}, {0.437068, 0.424859}, {0.438236, 0.427071},
		{0.441776, 0.433848}, {0.443254, 0.436707}, {0.444750, 0.439620}, {0.445326, 0.440748},
		{0.445781, 0.441640}, {0.446175, 0.442415}, {0.446361, 0.442781}, {0.446493, 0.443039},
		{0.446801, 0.443647},
	}
	army := RefSoldiers(BaseBuildings(), Ages())
	for i, age := range AgeOrder() {
		if i == 0 {
			continue // no military building yet
		}
		rating := 2 * army[i] * (1 + MilitaryScaleAt(i).Now)
		full := DefenseMitigation(rating, AgeThreat(i))
		half := DefenseMitigation(rating/2, AgeThreat(i))
		if math.Abs(full-was[i][0]) > 2e-6 || math.Abs(half-was[i][1]) > 2e-6 {
			t.Errorf("%s: the reference army of %.0f blunts %.6f and half of it %.6f of a raid, want %.6f and %.6f", age, army[i], full, half, was[i][0], was[i][1])
		}
	}
}
