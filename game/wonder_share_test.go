package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/rules"
)

func wonderOf(t *testing.T, set *rules.Set, age string) (string, map[string]float64) {
	t.Helper()
	key := set.Wonder(age)
	def, ok := set.Building(key)
	if !ok || len(def.BaseCost) == 0 {
		t.Fatalf("%s has no wonder with a price", age)
	}
	return key, def.BaseCost
}

func TestWonderBankShareEnds(t *testing.T) {
	set := rules.Core()
	for _, age := range set.AgeKeys() {
		if set.Wonder(age) == "" {
			continue
		}
		_, cost := wonderOf(t, set, age)
		if got := WonderBankShare(set, age, cost, nil); got != 0 {
			t.Errorf("%s: an empty bank is %v", age, got)
		}
		full := map[string]float64{}
		for r, c := range cost {
			full[r] = c
		}
		if got := WonderBankShare(set, age, cost, full); math.Abs(got-1) > 1e-12 {
			t.Errorf("%s: a full bank is %v", age, got)
		}
		// More than the price in every part is still one whole.
		over := map[string]float64{}
		for r, c := range cost {
			over[r] = c * 3
		}
		if got := WonderBankShare(set, age, cost, over); got > 1 {
			t.Errorf("%s: an overfull bank is %v", age, got)
		}
	}
	if got := WonderBankShare(set, "bronze_age", nil, nil); got != 0 {
		t.Errorf("no price is %v", got)
	}
}

// A part counts for what it is worth, not for its raw amount: in the Quantum
// Age two parts are named in the quintillions and one in the hundreds of
// quadrillions, yet the one is most of the price.
func TestWonderBankShareWeighsByPrice(t *testing.T) {
	set := rules.Core()
	age := "quantum_age"
	_, cost := wonderOf(t, set, age)
	bank := map[string]float64{"antimatter": cost["antimatter"], "dark_matter": cost["dark_matter"]}
	got := WonderBankShare(set, age, cost, bank)
	if got > 0.05 {
		t.Errorf("the two cheap parts banked read %.3f of the price; by raw amount they would read about 0.78", got)
	}
	bank = map[string]float64{"quantum_flux": cost["quantum_flux"]}
	if got := WonderBankShare(set, age, cost, bank); got < 0.9 {
		t.Errorf("the dear part banked reads %.3f, want most of the price", got)
	}
}

// A part cannot be paid with another's surplus, and more bank never reads as
// less.
func TestWonderBankShareCapsEachPartAndNeverFalls(t *testing.T) {
	set := rules.Core()
	age := "bronze_age"
	_, cost := wonderOf(t, set, age)
	only := map[string]float64{"stone": cost["stone"] * 50}
	whole := WonderBankShare(set, age, cost, only)
	if whole >= 1 {
		t.Fatalf("a surplus of one part paid the price: %v", whole)
	}
	capped := WonderBankShare(set, age, cost, map[string]float64{"stone": cost["stone"]})
	if whole != capped {
		t.Errorf("a surplus of one part read %v, the part banked exactly reads %v", whole, capped)
	}
	prev := 0.0
	for step := 0; step <= 20; step++ {
		bank := map[string]float64{}
		for r, c := range cost {
			bank[r] = c * float64(step) / 20
		}
		got := WonderBankShare(set, age, cost, bank)
		if got < prev {
			t.Fatalf("the share fell from %v to %v at step %d", prev, got, step)
		}
		prev = got
	}
}

// A part the age has no level for counts as the average of the others.
func TestWonderBankShareFallbackForAnUnpricedPart(t *testing.T) {
	set := rules.Core()
	age := "bronze_age"
	_, cost := wonderOf(t, set, age)
	with := map[string]float64{}
	for r, c := range cost {
		with[r] = c
	}
	with["moon_dust"] = 10
	if lvl := set.DealPriceLevel("moon_dust", age); lvl != 0 {
		t.Fatalf("moon_dust has a level %v", lvl)
	}
	empty := WonderBankShare(set, age, with, nil)
	if empty != 0 {
		t.Errorf("empty bank read %v", empty)
	}
	// The unpriced part alone: it weighs the mean of the others'.
	got := WonderBankShare(set, age, with, map[string]float64{"moon_dust": 10})
	sum := 0.0
	for r, c := range cost {
		sum += c / set.DealPriceLevel(r, age)
	}
	mean := sum / float64(len(cost))
	if want := mean / (sum + mean); math.Abs(got-want) > 1e-9 {
		t.Errorf("the unpriced part reads %v, want %v", got, want)
	}
	// A nil set reads the core rules.
	if a, b := WonderBankShare(nil, age, cost, map[string]float64{"stone": cost["stone"]}),
		WonderBankShare(set, age, cost, map[string]float64{"stone": cost["stone"]}); a != b {
		t.Errorf("a nil set read %v, the core rules %v", a, b)
	}
}
