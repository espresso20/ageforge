package config

import (
	"strings"
	"testing"
)

// The table holds the calendar the panel was designed with, one year per
// core age, in order and going up.
func TestAgeYearsCoverTheCoreAges(t *testing.T) {
	order := AgeOrder()
	years := AgeStartYears(order)
	want := map[string]int{
		"primitive_age": -100000000, "stone_age": -8000, "bronze_age": -3300, "iron_age": -1200,
		"classical_age": -500, "medieval_age": 500, "renaissance_age": 1400, "colonial_age": 1600,
		"industrial_age": 1760, "victorian_age": 1840, "electric_age": 1880, "atomic_age": 1945,
		"modern_age": 1960, "information_age": 1980, "digital_age": 2000, "cyberpunk_age": 2050,
		"fusion_age": 2100, "space_age": 2200, "interstellar_age": 2400, "galactic_age": 3000,
		"quantum_age": 5000, "transcendent_age": 9999,
	}
	if len(order) != len(want) {
		t.Fatalf("%d ages, the table holds %d", len(order), len(want))
	}
	for i, k := range order {
		if years[i] != want[k] {
			t.Errorf("%s starts in %d, want %d", k, years[i], want[k])
		}
		if i > 0 && years[i] <= years[i-1] {
			t.Errorf("%s (%d) does not come after %s (%d)", k, years[i], order[i-1], years[i-1])
		}
	}
	if AgeYearLabel("transcendent_age") == "" {
		t.Error("the last age has no label in place of its number")
	}
	if AgeYearLabel("bronze_age") != "" {
		t.Error("a middle age has a label in place of its number")
	}
}

// An age the table does not know takes a year between its neighbours, and
// the years still do not go down.
func TestAgeYearsInterpolateUnknownAges(t *testing.T) {
	keys := []string{"stone_age", "bronze_age", "gen_a", "gen_b", "gen_c", "medieval_age", "gen_end", "gen_end2"}
	years := AgeStartYears(keys)
	if years[0] != -8000 || years[1] != -3300 || years[5] != 500 {
		t.Fatalf("known ages moved: %v", years)
	}
	for i := 1; i < len(years); i++ {
		if years[i] < years[i-1] {
			t.Errorf("the calendar goes back at %s: %v", keys[i], years)
		}
	}
	// Three ages between -3300 and 500: a quarter, a half, three quarters.
	if years[2] != -3300+3800/4 || years[3] != -3300+3800/2 || years[4] != -3300+3800*3/4 {
		t.Errorf("interpolated years %v", years[2:5])
	}
	// Past the last known age: a year a step.
	if years[6] != 501 || years[7] != 502 {
		t.Errorf("years past the last known age %v", years[6:])
	}
	// Before the first known age.
	if got := AgeStartYears([]string{"gen_first", "stone_age"}); got[0] != -8001 {
		t.Errorf("year before the first known age: %v", got)
	}
	if got := AgeStartYears([]string{"gen_only"}); got[0] != 0 {
		t.Errorf("an order of unknown ages: %v", got)
	}
}

// The year of a town between two ages is the first at 0, the second at 1,
// and never goes down on the way.
func TestYearBetweenIsMonotonicAndHitsBothEnds(t *testing.T) {
	order := AgeOrder()
	years := AgeStartYears(order)
	for i := 0; i+1 < len(order); i++ {
		a, b := years[i], years[i+1]
		if got := YearBetween(a, b, 0); got != a {
			t.Errorf("%s at 0: %d, want %d", order[i], got, a)
		}
		if got := YearBetween(a, b, 1); got != b {
			t.Errorf("%s at 1: %d, want %d", order[i], got, b)
		}
		prev := a
		for p := 0.0; p <= 1.0001; p += 0.01 {
			y := YearBetween(a, b, p)
			if y < prev {
				t.Fatalf("%s: the year went back from %d to %d at %.2f", order[i], prev, y, p)
			}
			prev = y
		}
		// Out-of-range shares clamp.
		if YearBetween(a, b, -1) != a || YearBetween(a, b, 7) != b {
			t.Errorf("%s: an out-of-range share is not clamped", order[i])
		}
	}
}

func TestYearBetweenDeepTime(t *testing.T) {
	start, next := -100000000, -8000
	if got := YearBetween(start, next, 0); got != start {
		t.Errorf("at 0: %d", got)
	}
	if got := YearBetween(start, next, 1); got != next {
		t.Errorf("at 1: %d", got)
	}
	if got := YearBetween(start, next, 0.5); got < -900000 || got > -888000 {
		t.Errorf("midpoint %d, want near -894,000", got)
	}
	prev := start
	for p := 0; p <= 1000; p++ {
		y := YearBetween(start, next, float64(p)/1000)
		if y < prev {
			t.Fatalf("goes down at %d: %d after %d", p, y, prev)
		}
		prev = y
	}
	if got := YearBetween(-8000, -3300, 0.5); got != -5650 {
		t.Errorf("linear span changed: %d", got)
	}
}

func TestYearLabel(t *testing.T) {
	for year, want := range map[int]string{
		-100000000: "Late Cretaceous", -66000000: "Late Cretaceous", -65999999: "Paleocene",
		-56000000: "Paleocene", -55999999: "Eocene", -33900000: "Eocene", -33899999: "Oligocene",
		-23000000: "Oligocene", -22999999: "Miocene", -5300000: "Miocene", -5299999: "Pliocene",
		-2580000: "Pliocene", -2579999: "Pleistocene", -894000: "Pleistocene", -100000: "Pleistocene",
		-99999: "c. 100,000 BC", -99400: "c. 99,000 BC", -40400: "c. 40,000 BC", -10001: "c. 10,000 BC",
		-10000: "c. 10,000 BC", -8000: "c. 8,000 BC", -3300: "c. 3,300 BC", -1200: "c. 1,200 BC",
		-500: "c. 500 BC", -1: "c. 1 BC",
		0: "c. 1 AD", 1: "c. 1 AD", 500: "c. 500 AD", 1400: "c. 1400 AD", 1474: "c. 1450 AD",
		1475: "c. 1500 AD", 1600: "c. 1600 AD", 1699: "c. 1700 AD",
		1700: "1700s", 1760: "1760s", 1848: "1840s", 1945: "1940s", 2050: "2050s", 2099: "2090s",
		2100: "22nd century", 2200: "23rd century", 2400: "25th century", 3000: "31st century",
		5000: "51st century", 1000000 - 1: "10000th century",
		1000: "c. 1000 AD",
	} {
		if got := YearLabel(year); got != want {
			t.Errorf("YearLabel(%d) = %q, want %q", year, got, want)
		}
	}
	for n, want := range map[int]string{1: "1st", 2: "2nd", 3: "3rd", 4: "4th", 11: "11th", 12: "12th", 13: "13th",
		21: "21st", 22: "22nd", 23: "23rd", 101: "101st", 111: "111th", 112: "112th"} {
		if got := ordinal(n); got != want {
			t.Errorf("ordinal(%d) = %q, want %q", n, got, want)
		}
	}
	for _, y := range []int{-3300, 1948, 2100} {
		if strings.Contains(YearLabel(y), "  ") {
			t.Errorf("YearLabel(%d) has a double space", y)
		}
	}
}
