package config

import (
	"strconv"

	"github.com/espresso20/ageforge/vibe"
)

// The calendar: the year each age begins, for the Advance panel's march of
// ages. It is a picture of the town's history, not a length of play: the year
// shown while a town closes on the next age is its own age's start plus the
// share of the way to the next start that the gate is done (see
// YearBetween).
//
// The last age has no number: it shows ageYearLabels' word instead.

// ageStartYears is each age's first year, negative for BC. The table is in
// age order and holds every core age.
var ageStartYears = []struct {
	Key  string
	Year int
}{
	{"primitive_age", -10000}, {"stone_age", -8000}, {"bronze_age", -3300}, {"iron_age", -1200},
	{"classical_age", -500}, {"medieval_age", 500}, {"renaissance_age", 1400}, {"colonial_age", 1600},
	{"industrial_age", 1760}, {"victorian_age", 1840}, {"electric_age", 1880}, {"atomic_age", 1945},
	{"modern_age", 1960}, {"information_age", 1980}, {"digital_age", 2000}, {"cyberpunk_age", 2050},
	{"fusion_age", 2100}, {"space_age", 2200}, {"interstellar_age", 2400}, {"galactic_age", 3000},
	{"quantum_age", 5000}, {"transcendent_age", 9999},
}

// ageYearLabels are the ages whose start is shown as a word, not a number.
var ageYearLabels = map[string]string{"transcendent_age": "Beyond time"}

// AgeStartYears gives the start year of each age in keys (an age order), in
// order. An age the table does not know takes a year interpolated between
// its nearest known neighbours by position (one year a step past either end),
// so a ruleset with ages of its own still has a calendar. The years never
// go down along an order whose known ages do not.
func AgeStartYears(keys []string) []int {
	known := map[string]int{}
	for _, a := range ageStartYears {
		known[a.Key] = a.Year
	}
	years := make([]int, len(keys))
	have := make([]bool, len(keys))
	for i, k := range keys {
		if y, ok := known[k]; ok {
			years[i], have[i] = y, true
		}
	}
	for i := range keys {
		if have[i] {
			continue
		}
		lo, hi := i-1, i+1
		for lo >= 0 && !have[lo] {
			lo--
		}
		for hi < len(keys) && !have[hi] {
			hi++
		}
		switch {
		case lo >= 0 && hi < len(keys):
			years[i] = years[lo] + (years[hi]-years[lo])*(i-lo)/(hi-lo)
		case lo >= 0:
			years[i] = years[lo] + (i - lo)
		case hi < len(keys):
			years[i] = years[hi] - (hi - i)
		default:
			years[i] = i
		}
	}
	return years
}

// AgeYearLabel is the word shown in place of age's start year, or "" when
// the year is a number.
func AgeYearLabel(age string) string { return ageYearLabels[age] }

// YearBetween is the year that share of the way (0 to 1) from the year
// start to the year next. It is whole-number arithmetic after the share is
// taken to the thousandth, so every machine agrees, it never goes down as
// the share goes up (when next is not before start), and it is start at 0 and
// next at 1.
func YearBetween(start, next int, share float64) int {
	if share < 0 || share != share {
		share = 0
	}
	if share > 1 {
		share = 1
	}
	thousandths := int(share * 1000)
	return start + (next-start)*thousandths/1000
}

// YearLabel writes a year as the panel shows it: "3,300 BC", "1,948 AD".
// There is no year 0; it reads "1 AD".
func YearLabel(year int) string {
	suffix := vibe.Line("config.age_years.YearLabel.01", " AD")
	if year < 0 {
		year, suffix = -year, " BC"
	}
	if year == 0 {
		year = 1
	}
	return groupThousands(year) + suffix
}

// groupThousands writes n with a comma between each three digits.
func groupThousands(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
