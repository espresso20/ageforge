package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// oldBraceCeilings is the most any material of an age's Brace asked while the
// ceiling was three fifths of a moderate builder's store, rounded down to two
// figures. Every material of an age had the same one. The ceiling is now the
// same share of the age's stock scale in hours of the moderate economy's
// income, and keeps these amounts.
var oldBraceCeilings = map[string]float64{
	"iron_age": 130000, "classical_age": 1000000, "medieval_age": 7300000, "renaissance_age": 33000000,
	"colonial_age": 150000000, "industrial_age": 510000000, "victorian_age": 6500000000, "electric_age": 7.2e+10,
	"atomic_age": 4.3e+11, "modern_age": 1.8e+12, "information_age": 1.1e+13, "digital_age": 4.4e+13,
	"cyberpunk_age": 4e+14, "fusion_age": 1.9e+15, "space_age": 9.7e+15, "interstellar_age": 2.9e+17,
	"galactic_age": 2.6e+18, "quantum_age": 2e+20, "transcendent_age": 2e+20,
}

// TestBraceCeilingKeepsTheAmountsItHad: in every age a thread can arrive in,
// every material's ceiling, on a doom's thread and on the Last Passage's, is
// the amount it was when the ceiling was a share of a store; and it is
// the same ceiling in hours of the age's typical income of the material.
func TestBraceCeilingKeepsTheAmountsItHad(t *testing.T) {
	set := rules.Core()
	checked := 0
	for _, row := range HarbingerPriceTable() {
		want, ok := oldBraceCeilings[row.Age]
		if !ok {
			t.Fatalf("no old ceiling pinned for %s", row.Age)
		}
		if len(row.BraceCeiling) == 0 {
			t.Errorf("%s (Last Passage %v): no ceiling", row.Age, row.LastPassage)
		}
		for res, got := range row.BraceCeiling {
			checked++
			if got != want {
				t.Errorf("%s (Last Passage %v): the ceiling of %s is %v, it was %v", row.Age, row.LastPassage, res, got, want)
			}
			hours := row.BraceCeilingHours[res]
			if hours <= 0 {
				t.Errorf("%s: no ceiling in hours for %s", row.Age, res)
				continue
			}
			// Hours of the typical income are the ceiling's measure: they
			// come to the ceiling (before it is rounded down) in units.
			unrounded := hours * config.TicksPerHour * set.TypicalIncome(res, row.Age)
			if unrounded < got || unrounded > 1.1*got {
				t.Errorf("%s: %s's ceiling %v is %.4g hours of its typical income, which comes to %v", row.Age, res, got, hours, unrounded)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no ceiling checked")
	}
}
