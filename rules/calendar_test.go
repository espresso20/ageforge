package rules

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// A set carries the calendar of its ages: one year each, in order.
func TestSetCalendar(t *testing.T) {
	s := Core()
	years := s.AgeYears()
	if len(years) != s.NumAges() {
		t.Fatalf("%d years for %d ages", len(years), s.NumAges())
	}
	for i, k := range s.AgeKeys() {
		got, ok := s.AgeYear(k)
		if !ok || got != years[i] {
			t.Errorf("AgeYear(%s) = %d, %v; AgeYears[%d] = %d", k, got, ok, i, years[i])
		}
		if i > 0 && years[i] <= years[i-1] {
			t.Errorf("%s (%d) does not come after the age before (%d)", k, years[i], years[i-1])
		}
	}
	if y, _ := s.AgeYear("bronze_age"); y != -3300 {
		t.Errorf("the Bronze Age begins in %d", y)
	}
	if _, ok := s.AgeYear("not_an_age"); ok {
		t.Error("an unknown age has a year")
	}
	if s.AgeYearLabel("transcendent_age") == "" || s.AgeYearLabel("bronze_age") != "" {
		t.Error("only the last age shows a word for its year")
	}
	// The years handed out are copies.
	years[0] = 12345
	if again := s.AgeYears(); again[0] == 12345 {
		t.Error("AgeYears handed out the set's own slice")
	}
}

// A set with an age of its own, one the table does not know, still has a year
// for it, between its neighbours.
func TestSetCalendarInterpolatesAGeneratedAge(t *testing.T) {
	src := FromConfig()
	var ages []config.AgeDef
	for _, a := range src.Ages {
		ages = append(ages, a)
		if a.Key == "bronze_age" {
			gen := a
			gen.Key, gen.Name = "generated_age", "Generated Age"
			gen.BuildingReqs = nil
			ages = append(ages, gen)
		}
	}
	for i := range ages {
		ages[i].Order = i
	}
	src.Ages = ages
	s := Compile(src)
	gen, ok := s.AgeYear("generated_age")
	if !ok {
		t.Fatal("the generated age has no year")
	}
	bronze, _ := s.AgeYear("bronze_age")
	iron, _ := s.AgeYear("iron_age")
	if gen <= bronze || gen >= iron {
		t.Errorf("the generated age begins in %d, not between %d and %d", gen, bronze, iron)
	}
}
