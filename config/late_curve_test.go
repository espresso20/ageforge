package config

import (
	"testing"
)

// TestNanobotsAreNeverPriced: nanobots are a build material sized by hand to
// their producers' listed rates, not a construction resource. No age prices
// them, so their producers keep what they are listed at, and a building that
// asks for them is not valued by them. Priced, 150K of them stood for a third
// of a Cyberpunk Age building's worth: the payback rule cut a Nanobot Vat
// from 6,554 a tick to 20.
func TestNanobotsAreNeverPriced(t *testing.T) {
	if !IsFlowResource("nanobots") {
		t.Fatal("nanobots are priced like a construction resource")
	}
	asked := 0
	for _, a := range AgeOrder() {
		if lv := PriceLevels(a)["nanobots"]; lv != 0 {
			t.Errorf("%s prices nanobots at %v", a, lv)
		}
	}
	by := BuildingByKey()
	for key, want := range map[string]float64{"nano_foundry": 80, "bio_fabrication_lab": 3276.8, "nanobot_vat": 6553.6, "molecular_synthesizer": 13107.2} {
		if got, ok := prodEffectValue(by[key].Effects, "nanobots"); !ok || got != want {
			t.Errorf("%s makes %v nanobots a tick, want its listed %v", key, got, want)
		}
	}
	// What a building asks of them is small beside what one producer makes:
	// under an hour of a single lab or vat of its own age.
	makes := map[string]float64{"digital_age": 3276.8, "cyberpunk_age": 6553.6}
	for _, d := range BaseBuildings() {
		n, ok := d.BaseCost["nanobots"]
		if !ok {
			continue
		}
		asked++
		rate := makes[d.RequiredAge]
		if rate == 0 || n/rate > 1800 {
			t.Errorf("%s (%s) asks %v nanobots: %v ticks of one producer of its age, want under an hour", d.Key, d.RequiredAge, n, n/rate)
		}
	}
	if asked == 0 {
		t.Error("no building asks for nanobots: the material has no use")
	}
}
