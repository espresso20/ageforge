package config

import (
	"math"
	"testing"
)

// The market's exchange is going, so a resource must have a maker in the town
// that needs it. These tests hold the tables to the rules that gave each
// resource its first maker. Each figure below was worked out from the tables
// as they stood before the change, and is written here so the rule can be
// checked again.

// TestFirstMakersKeepTheirPriceUnits: a first maker is paid for with what a
// town already makes, and its price in the age's price units does not move.
// The Server Farm (the first maker of data) lost its data part, and the Black
// Market Hub (the first maker of crypto) lost its crypto part; what remained
// was scaled by one factor each. Five more first makers lost the part of the
// price in the material they make, for the reason below. Units are counted against the age's levels as
// they now stand, which the change moved a little (data, crypto), so the
// bound is one percent.
func TestFirstMakersKeepTheirPriceUnits(t *testing.T) {
	defs := BuildingByKey()
	for _, c := range []struct {
		key, gone string
		was       float64 // price units before the change
	}{
		{"server_farm", "data", 4.7259},
		{"black_market", "crypto", 5.1457},
		// The cold start counts a carried supply toward a first copy only if
		// the copies the gates required make the price in one target of the
		// age. These first makers of an age's material were priced in the
		// material itself, which the carried copies make far too slowly
		// (the Uranium Processing Works asked 1.2 billion uranium, the
		// carried mines make 55 million in the Atomic Age's target), so each
		// lost that part, the rest scaled by one factor.
		{"uranium_processing_works", "uranium", 2.7524},
		{"power_station", "electricity", 4.5833},
		{"fusion_reactor", "plasma", 4.4611},
		{"warp_drive_plant", "dark_matter", 4.2167},
		{"stellar_metallurgy", "antimatter", 4.4167},
	} {
		d := defs[c.key]
		if _, has := d.BaseCost[c.gone]; has {
			t.Errorf("%s is still priced in %s, the resource it is the first maker of", c.key, c.gone)
		}
		got := priceUnits(d.BaseCost, PriceLevels(d.RequiredAge))
		if math.Abs(got/c.was-1) > 0.01 {
			t.Errorf("%s costs %.4f price units of %s, want %.4f as before", c.key, got, d.RequiredAge, c.was)
		}
	}
}

// TestBlackMarketMakesCryptoAtTheStandard: the lineage entry names crypto as
// the Black Market Hub's output, so it makes crypto, and it keeps the gold it
// always made beside it (the Cyberpunk Age prices no gold, so that gold is left
// as it was). Its crypto is on the maker rule's standard like any maker's: its
// first-copy price in units over the age's payback, turned into crypto at
// crypto's own level (a Rule 6 fill, held by TestEveryHoleIsFilledByRule). It
// was set to the median output of the age's other makers, which is 20% off the
// standard once the price is counted.
func TestBlackMarketMakesCryptoAtTheStandard(t *testing.T) {
	d := BuildingByKey()["black_market"]
	if d.OutputResource != "crypto" {
		t.Errorf("black_market output resource = %q, want crypto (its lineage entry names it)", d.OutputResource)
	}
	crypto, gold := 0.0, 0.0
	for _, e := range d.Effects {
		if e.Type != "production" {
			continue
		}
		switch e.Target {
		case "crypto":
			crypto += e.Value
		case "gold":
			gold += e.Value
		default:
			t.Errorf("black_market makes %s, want only crypto and the gold it always made", e.Target)
		}
	}
	if gold != 409.6 {
		t.Errorf("black_market makes %.4g gold a tick, want the 409.6 it always made", gold)
	}
	lv := PriceLevels(d.RequiredAge)
	if lv["crypto"] <= 0 {
		t.Fatalf("%s has no price level for crypto", d.RequiredAge)
	}
	price := priceUnits(d.BaseCost, lv)
	want := float64(price/MakerPaybackTicks[d.RequiredAge]) * lv["crypto"]
	if math.Abs(crypto/want-1) > 0.005 { // the literal is written to three figures
		t.Errorf("black_market makes %s crypto a tick, the standard is %s", GoLiteral(roundSignificant(crypto, 4)), GoLiteral(roundSignificant(want, 4)))
	}
}

// TestWonderPartsFitTheirAge: a flow part of a wonder's price (food, faith,
// culture) is covered by the age's income within its target, and the Space
// Program asks nothing its age does not price. It asked 600B knowledge, which
// the Modern Age does not price and no town made from buildings; the Sistine
// Chapel asked 12.8M culture where the Renaissance Age's income makes 967K in
// its target, and now asks 960K, the largest figure of two significant digits
// that income covers (it was 950K against 952K before the storage techs became
// production techs, which lifted that income).
func TestWonderPartsFitTheirAge(t *testing.T) {
	for _, d := range BaseBuildings() {
		if d.Category != "wonder" {
			continue
		}
		for res, c := range d.BaseCost {
			if !IsFlowResource(res) || c <= 0 {
				continue
			}
			if made := FlowIncome(res, d.RequiredAge) * LatePriceTargetTicks(d.RequiredAge); c > made {
				t.Errorf("%s asks %s %s, more than the %s the age's income makes in its target", d.Key, GoLiteral(roundSignificant(c, 4)), res, GoLiteral(roundSignificant(made, 4)))
			}
		}
	}
	space := BuildingByKey()["space_program"]
	levels := PriceLevels(space.RequiredAge)
	for res := range space.BaseCost {
		if levels[res] <= 0 {
			t.Errorf("space_program asks %s, which %s does not price", res, space.RequiredAge)
		}
	}
	// The Hoover Dam asked 80B stone, a material no Electric Age building is
	// priced in: it sat outside the wonder's units, as the Space Program's
	// knowledge did, and no building of the age made it. It asks none now.
	if _, has := BuildingByKey()["hoover_dam"].BaseCost["stone"]; has {
		t.Errorf("hoover_dam is priced in stone, which no Electric Age building is priced in or makes")
	}
	sistine := BuildingByKey()["sistine_chapel"]
	made := FlowIncome("culture", sistine.RequiredAge) * LatePriceTargetTicks(sistine.RequiredAge)
	if got := sistine.BaseCost["culture"]; got != 960000 || made < 960000 || made >= 970000 {
		t.Errorf("sistine_chapel culture = %s with %s made in the target: want 960_000, the largest two-figure number the income covers", GoLiteral(roundSignificant(got, 4)), GoLiteral(roundSignificant(made, 4)))
	}
}
