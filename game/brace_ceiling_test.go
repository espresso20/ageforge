package game

import (
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// TestBraceCeilingIsHoursOfTypicalIncome: in every age a thread can arrive in,
// every material's ceiling, on a doom's thread and on the Last Passage's, is
// the same amount for every material of the thread, and it is the same ceiling
// in hours of the age's typical income of the material.
func TestBraceCeilingIsHoursOfTypicalIncome(t *testing.T) {
	set := rules.Core()
	checked := 0
	for _, row := range HarbingerPriceTable() {
		if len(row.BraceCeiling) == 0 {
			t.Errorf("%s (Last Passage %v): no ceiling", row.Age, row.LastPassage)
		}
		same := -1.0
		for res, got := range row.BraceCeiling {
			checked++
			if same < 0 {
				same = got
			} else if got != same {
				t.Errorf("%s (Last Passage %v): the ceiling of %s is %s, but another material's is %s", row.Age, row.LastPassage, res, config.GoLiteral(got), config.GoLiteral(same))
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
				t.Errorf("%s: %s's ceiling %s is %.4g hours of its typical income, which comes to %s", row.Age, res, config.GoLiteral(got), hours, config.GoLiteral(unrounded))
			}
		}
	}
	if checked == 0 {
		t.Fatal("no ceiling checked")
	}
}
