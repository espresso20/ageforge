package config

import (
	"math"
	"testing"
)

func TestLatePriceMultiplier(t *testing.T) {
	for _, tc := range []struct {
		age  string
		want float64
	}{
		{"atomic_age", 1},
		{"modern_age", 1},
		{"information_age", 1.4},
		{"digital_age", 1.96},
		{"transcendent_age", math.Pow(1.4, 9)},
		{"not_an_age", 1},
	} {
		if got := LatePriceMultiplier(tc.age); math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("%s multiplier = %g, want %g", tc.age, got, tc.want)
		}
	}
}

func TestLatePriceClimbScalesEveryCostPart(t *testing.T) {
	defs := []BuildingDef{{RequiredAge: "information_age", BaseCost: map[string]float64{"steel": 100, "titanium_ore": 20}, Effects: []Effect{{Type: "production", Target: "steel", Value: 7}}}}
	applyLatePriceClimb(defs)
	if got, want := defs[0].BaseCost["steel"], 140.0; got != want {
		t.Errorf("steel cost = %g, want %g", got, want)
	}
	if got, want := defs[0].BaseCost["titanium_ore"], 28.0; got != want {
		t.Errorf("raw-material cost = %g, want %g", got, want)
	}
	if got, want := defs[0].Effects[0].Value, 7.0; got != want {
		t.Errorf("output = %g, want %g", got, want)
	}
}

func TestLatePricePaybackAndDealsFollowTheCurve(t *testing.T) {
	if got, want := MakerPaybackTicks["digital_age"], 40929*LatePriceMultiplier("digital_age")*NewCopyPriceFactor; math.Abs(got-want) > 1e-9 {
		t.Errorf("digital maker payback = %g, want %g", got, want)
	}
	if got, want := DealPaybackTicks["information_age"], 14935.330237167702*LatePriceMultiplier("information_age"); math.Abs(got-want) > 1e-9 {
		t.Errorf("information deal payback = %g, want %g", got, want)
	}
}

func TestNewCopyPriceScalesEveryBuildingButWondersAndMonuments(t *testing.T) {
	authored := map[string]float64{"wood": 100, "stone": 40}
	defs := []BuildingDef{
		{Key: "camp", Category: "production", RequiredAge: "stone_age", BaseCost: authored},
		{Key: "pit", Category: "storage", RequiredAge: "stone_age", BaseCost: map[string]float64{"wood": 10}},
		{Key: "grove", Category: "wonder", RequiredAge: "stone_age", BaseCost: map[string]float64{"wood": 1000}},
		{Key: "obelisk", Category: "monument", RequiredAge: "stone_age", BaseCost: map[string]float64{"culture": 500}},
	}
	applyNewCopyPrice(defs)
	if got, want := defs[0].BaseCost["wood"], 100*NewCopyPriceFactor; got != want {
		t.Errorf("production wood = %g, want %g", got, want)
	}
	if got, want := defs[0].BaseCost["stone"], 40*NewCopyPriceFactor; got != want {
		t.Errorf("production stone = %g, want %g", got, want)
	}
	if got, want := defs[1].BaseCost["wood"], 10*NewCopyPriceFactor; got != want {
		t.Errorf("storage wood = %g, want %g", got, want)
	}
	if got := defs[2].BaseCost["wood"]; got != 1000 {
		t.Errorf("wonder wood = %g, want 1000 (a wonder is not scaled)", got)
	}
	if got := defs[3].BaseCost["culture"]; got != 500 {
		t.Errorf("monument culture = %g, want 500 (a monument is not scaled)", got)
	}
	// The authored table keeps its map: a second pass from it must not compound.
	if authored["wood"] != 100 || authored["stone"] != 40 {
		t.Errorf("the authored map was changed to %v", authored)
	}
}

// TestLiveBuildingsTookTheNewCopyFactor: the loaded table carries the factor,
// so every reader sees one price; a wonder is as authored.
func TestLiveBuildingsTookTheNewCopyFactor(t *testing.T) {
	live := BuildingByKey()
	if got, want := live["gathering_camp"].BaseCost["wood"], 16*NewCopyPriceFactor; got != want {
		t.Errorf("a Gathering Camp costs %g wood, want %g (16 as first written, times the factor)", got, want)
	}
	if got, want := live["sacred_grove"].BaseCost["wood"], 1600.0; got != want {
		t.Errorf("the Sacred Grove costs %g wood, want %g (a wonder is not scaled)", got, want)
	}
	if again := BuildingByKey(); again["gathering_camp"].BaseCost["wood"] != live["gathering_camp"].BaseCost["wood"] {
		t.Error("a second load of the table priced a camp differently: the factor compounded")
	}
}
