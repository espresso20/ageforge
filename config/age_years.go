package config

import (
	"math"
	"strconv"

	"github.com/espresso20/ageforge/detmath"
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
	{"primitive_age", -100000000}, {"stone_age", -8000}, {"bronze_age", -3300}, {"iron_age", -1200},
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
// start to the year next. The share is taken to the thousandth, so every
// machine agrees on it. It never goes down as the share goes up (when next is
// not before start), and it is start at 0 and next at 1. Across deep time
// (start at or before 100,000 BC, next still BC) the count of years ago is
// interpolated geometrically, so the first stretch of a long span does not
// swallow the rest; every other span is linear.
func YearBetween(start, next int, share float64) int {
	if share < 0 || share != share {
		share = 0
	}
	if share > 1 {
		share = 1
	}
	thousandths := int(share * 1000)
	if start <= -deepTime && next < 0 {
		if thousandths == 0 {
			return start
		}
		if thousandths == 1000 {
			return next
		}
		startAgo, nextAgo := float64(-start), float64(-next)
		ago := startAgo * detmath.Pow(nextAgo/startAgo, float64(thousandths)/1000)
		return -int(math.Round(ago))
	}
	return start + (next-start)*thousandths/1000
}

// deepTime is how many years ago the calendar turns to geologic epochs.
const deepTime = 100000

// epochName is the geologic epoch that ago years ago (100,000 or more) falls
// in. The first case catches everything older.
func epochName(ago int) string {
	switch {
	case ago >= 66000000:
		return vibe.Line("config.age_years.epoch.late_cretaceous", "Late Cretaceous")
	case ago >= 56000000:
		return vibe.Line("config.age_years.epoch.paleocene", "Paleocene")
	case ago >= 33900000:
		return vibe.Line("config.age_years.epoch.eocene", "Eocene")
	case ago >= 23000000:
		return vibe.Line("config.age_years.epoch.oligocene", "Oligocene")
	case ago >= 5300000:
		return vibe.Line("config.age_years.epoch.miocene", "Miocene")
	case ago >= 2580000:
		return vibe.Line("config.age_years.epoch.pliocene", "Pliocene")
	}
	return vibe.Line("config.age_years.epoch.pleistocene", "Pleistocene")
}

// YearLabel writes a year as the panel shows it, rounded by how far off it
// is: a geologic epoch for deep time, "c. 40,000 BC", "c. 3,300 BC",
// "c. 1450 AD", a decade ("1840s") or a century ("22nd century"). There is
// no year 0; it reads "c. 1 AD".
func YearLabel(year int) string {
	if year < 0 {
		ago := -year
		if ago >= deepTime {
			return epochName(ago)
		}
		step := 100
		if ago > 10000 {
			step = 1000
		}
		if r := roundTo(ago, step); r > 0 {
			ago = r
		}
		return vibe.Sprintf("config.age_years.YearLabel.bc", "c. %s BC", groupThousands(ago))
	}
	switch {
	case year < 1700:
		return vibe.Sprintf("config.age_years.YearLabel.ad", "c. %s AD", strconv.Itoa(max(roundTo(year, 50), 1)))
	case year < 2100:
		return vibe.Sprintf("config.age_years.YearLabel.decade", "%ds", year/10*10)
	}
	return vibe.Sprintf("config.age_years.YearLabel.century", "%s century", ordinal(year/100+1))
}

// roundTo rounds n to the nearest multiple of step, halves up.
func roundTo(n, step int) int { return (n + step/2) / step * step }

// ordinal writes n as "22nd", "11th", "21st".
func ordinal(n int) string {
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			return vibe.Sprintf("config.age_years.ordinal.st", "%dst", n)
		case 2:
			return vibe.Sprintf("config.age_years.ordinal.nd", "%dnd", n)
		case 3:
			return vibe.Sprintf("config.age_years.ordinal.rd", "%drd", n)
		}
	}
	return vibe.Sprintf("config.age_years.ordinal.th", "%dth", n)
}

// groupThousands writes n with a comma between each three digits.
func groupThousands(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
