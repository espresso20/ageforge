package rules

import (
	"slices"

	"github.com/espresso20/ageforge/config"
)

// calendar.go is the year each age begins in, for the Advance panel's march
// of ages. A set works it out once, at compile, from package config's table
// (config.AgeStartYears), so an age a ruleset adds that the table does not
// know still has a year, interpolated between its neighbours.

func (s *Set) deriveCalendar() {
	s.years = config.AgeStartYears(s.ageKeys)
	s.yearLabels = map[string]string{}
	for _, k := range s.ageKeys {
		if w := config.AgeYearLabel(k); w != "" {
			s.yearLabels[k] = w
		}
	}
}

// AgeYears returns the calendar year each age begins in, in age order
// (negative for BC). The years do not go down along the order.
func (s *Set) AgeYears() []int { return slices.Clone(s.years) }

// AgeYear is the year age begins in, and false for an unknown age.
func (s *Set) AgeYear(age string) (int, bool) {
	i, ok := s.agePos[age]
	if !ok {
		return 0, false
	}
	return s.years[i], true
}

// AgeYearLabel is the word shown in place of age's year, or "" when it is a
// number.
func (s *Set) AgeYearLabel(age string) string { return s.yearLabels[age] }
