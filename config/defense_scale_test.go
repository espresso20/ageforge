package config

import (
	"math"
	"testing"
)

// TestMilitaryYardstick holds the military yardstick age by age: the scales
// put the typical player's Defense Rating against the threat, and what their
// power takes off a mission, back to what they were when the raid threat and
// the missions were measured. A change to a military tech moves Now and the
// scales with it, and the typical player's odds stay put.
func TestMilitaryYardstick(t *testing.T) {
	order := AgeOrder()
	if len(MilitaryCalibrated) != len(order) {
		t.Fatalf("MilitaryCalibrated lists %d ages, the game has %d", len(MilitaryCalibrated), len(order))
	}
	prev := 0.0
	for i, age := range order {
		s := MilitaryScaleAt(i)
		// The typical player's Defense Rating against the threat, and what
		// their power takes off a mission, are what they were.
		if got := (1 + s.Now) / s.Threat; math.Abs(got-(1+s.Then)) > 1e-9 {
			t.Errorf("%s: the threat's scale leaves the typical garrison at %.4f of a soldier's rating, want %.4f", age, got, 1+s.Then)
		}
		if got := MissionPower(s.Now, i); math.Abs(got-s.Then) > 1e-9 {
			t.Errorf("%s: the typical player's mission power is %.4f, want what it was, %.4f", age, got, s.Then)
		}
		// The threat still rises with every age, in the garrison it was sized
		// against: measured in soldiers at their posts it also follows the
		// reference town's army (ArmyThreatScale), which can dip where a gate
		// asks for more copies of a building with fewer posts each.
		if th := AgeThreat(i) / ArmyThreatScale(i); th <= prev {
			t.Errorf("%s: the threat is %.0f, not above the age before's %.0f", age, th, prev)
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

// MissionPower is military power as a mission's difficulty reads it in the
// age with the given order on this package's own tree: power × the age's
// MilitaryScale.Mission. For tests and tools (rules.Set.MissionPower).
func MissionPower(power float64, order int) float64 {
	return float64(power * MilitaryScaleAt(order).Mission)
}
