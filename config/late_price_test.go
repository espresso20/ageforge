package config

import (
	"math"
	"testing"
)

func TestLatePriceMultiplier(t *testing.T) {
	for _, tc := range []struct {
		age string
		want float64
	}{
		{"atomic_age", 1},
		{"modern_age", 1},
		{"information_age", 1.2},
		{"digital_age", 1.44},
		{"transcendent_age", math.Pow(1.2, 9)},
		{"not_an_age", 1},
	} {
		if got := LatePriceMultiplier(tc.age); got != tc.want {
			t.Errorf("%s multiplier = %g, want %g", tc.age, got, tc.want)
		}
	}
}

func TestLatePriceClimbScalesEveryCostPart(t *testing.T) {
	defs := []BuildingDef{{RequiredAge: "information_age", BaseCost: map[string]float64{"steel": 100, "titanium_ore": 20}, Effects: []Effect{{Type: "production", Target: "steel", Value: 7}}}}
	applyLatePriceClimb(defs)
	if got, want := defs[0].BaseCost["steel"], 120.0; got != want {
		t.Errorf("steel cost = %g, want %g", got, want)
	}
	if got, want := defs[0].BaseCost["titanium_ore"], 24.0; got != want {
		t.Errorf("raw-material cost = %g, want %g", got, want)
	}
	if got, want := defs[0].Effects[0].Value, 7.0; got != want {
		t.Errorf("output = %g, want %g", got, want)
	}
}

func TestLatePricePaybackAndDealsFollowTheCurve(t *testing.T) {
	if got, want := MakerPaybackTicks["digital_age"], 40929*LatePriceMultiplier("digital_age"); got != want {
		t.Errorf("digital maker payback = %g, want %g", got, want)
	}
	if got, want := DealPaybackTicks["information_age"], 14935.330237167702*LatePriceMultiplier("information_age"); got != want {
		t.Errorf("information deal payback = %g, want %g", got, want)
	}
}
