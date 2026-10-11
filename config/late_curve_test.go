package config

import (
	"testing"
)

// TestNanobotsAreNeverPriced: nanobots are a build material sized by hand to
// their producers' listed rates, not a construction resource. No age prices
// them, so their producers keep what they are listed at, a building that
// asks for them is not valued by them, and the market does not trade them.
// Priced, 150K of them stood for a third of a Cyberpunk Age building's
// worth: the payback rule cut a Nanobot Vat from 6,554 a tick to 20 and the
// market paid 66M electricity for one.
func TestNanobotsAreNeverPriced(t *testing.T) {
	if !IsFlowResource("nanobots") {
		t.Fatal("nanobots are priced like a construction resource")
	}
	asked := 0
	for _, a := range AgeOrder() {
		if lv := PriceLevels(a)["nanobots"]; lv != 0 {
			t.Errorf("%s prices nanobots at %v", a, lv)
		}
		for _, p := range MarketPairs(a) {
			if p.From == "nanobots" || p.To == "nanobots" {
				t.Errorf("%s: the market trades %s for %s", a, p.From, p.To)
			}
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

// TestListedPairsClosedWhereTheAgePricesWhatTheyBuy: a listed pair's typed
// rate is for goods the age does not build with. Where the age prices what
// the pair buys and not what it takes, the pair is closed: gold, which no
// age builds with after the Information Age, bought the Digital and
// Cyberpunk Ages their data and crypto at 0.15 and 0.04 a coin.
func TestListedPairsClosedWhereTheAgePricesWhatTheyBuy(t *testing.T) {
	for _, c := range []struct {
		from, to, age string
		open          bool
		rate          float64 // the typed rate where it applies (0: parity, not checked)
	}{
		// Gold for data: parity while the age builds with both, closed while
		// it builds with data alone, the typed rate once it builds with
		// neither (the Space Age's gate still asks for data).
		{"gold", "data", "modern_age", true, 0},
		{"gold", "data", "information_age", true, 0},
		{"gold", "data", "digital_age", false, 0},
		{"gold", "data", "cyberpunk_age", false, 0},
		{"gold", "data", "fusion_age", true, 0.15},
		{"gold", "crypto", "cyberpunk_age", false, 0},
		{"gold", "crypto", "fusion_age", true, 0.04},
		// What an age does not price it still sells at the typed rate.
		{"gold", "knowledge", "digital_age", true, 5},
		{"gold", "culture", "cyberpunk_age", true, 3},
		{"gold", "stone", "electric_age", true, 30},
		{"data", "gold", "digital_age", true, 5},
		// A flow resource is never priced, and still buys.
		{"food", "wood", "bronze_age", true, 1},
		{"food", "stone", "bronze_age", true, 0.8},
		// Old stock no longer turns into what the age builds with.
		{"coal", "gold", "colonial_age", false, 0},
		{"iron", "gold", "modern_age", false, 0},
		{"oil", "gold", "information_age", false, 0},
	} {
		rate, ok := MarketOffers(c.from, c.to, c.age)
		if ok != c.open {
			t.Errorf("%s for %s in %s: open %v (rate %v), want %v", c.from, c.to, c.age, ok, rate, c.open)
			continue
		}
		if c.open && c.rate != 0 && rate != c.rate {
			t.Errorf("%s for %s in %s: rate %v, want the typed %v", c.from, c.to, c.age, rate, c.rate)
		}
		listed := false
		for _, p := range MarketPairs(c.age) {
			if p.From == c.from && p.To == c.to {
				listed = true
			}
		}
		if listed != c.open {
			t.Errorf("%s for %s in %s: listed %v, want %v", c.from, c.to, c.age, listed, c.open)
		}
	}
	// Whatever an age builds with it still buys with its other construction
	// resources: every priced resource has a parity pair into it.
	for _, a := range AgeOrder() {
		lv := PriceLevels(a)
		if len(lv) < 2 {
			continue
		}
		for to := range lv {
			found := false
			for _, p := range MarketPairs(a) {
				if p.To == to && lv[p.From] > 0 {
					found = true
				}
			}
			if !found {
				t.Errorf("%s: nothing the age builds with buys %s", a, to)
			}
		}
	}
}
