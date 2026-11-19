package rules

import "github.com/espresso20/ageforge/config"

// coins.go is the coin collection's timing: the reference time of an age
// and the level a strike earns against it, read from the set's pacing targets.

// CoinRefTicks is the reference time of age for a run that plays it at speed
// times the base rate: the age's pacing target divided by speed (a speed of 0
// or less counts as 1). Every time mark on a coin or a feat is a share of it.
func (s *Set) CoinRefTicks(age string, speed float64) float64 {
	if speed <= 0 {
		speed = 1
	}
	return s.TargetTicks(age) / speed
}

// CoinLevel is the level of a coin struck after ticksInAge ticks in age, by a
// run that plays the age at speed times the base rate. An age with no target
// gives level 1.
func (s *Set) CoinLevel(age string, ticksInAge, speed float64) int {
	return config.CoinLevelAt(s.CoinRefTicks(age, speed), ticksInAge)
}

// FeatMark is the number age's legendary feat asks for.
func (s *Set) FeatMark(age string) float64 { return s.featMarks[age] }
