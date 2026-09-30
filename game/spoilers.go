package game

import "github.com/espresso20/ageforge/config"

// spoilers.go is the no-spoiler rule for player text (playtest 2026-09-29):
// no screen, log line or refusal names an age the player cannot see yet, an
// era they have not reached, or a civilization they have not met.
//
//   - An age may be named once it is the next one (the Next Age goal names it
//     on purpose) or once the player has reached it: in this run, or on their
//     account in any run.
//   - An era may be named once the player has reached one of its ages.
//   - A civilization may be named once met (FactionInfo.Discovered).
//
// Refusals from the managers (research, trade routes, expeditions) only know
// the current age, so they name the next age at most (laterAgeRef).

// AgeSight is which ages and eras the player may see named.
type AgeSight struct {
	next    int // the furthest age order that may be named
	reached int // the furthest age order reached
}

// newAgeSight is the sight of a player in current who has reached the ages
// in reached this run and highest on their account ("" for none).
func newAgeSight(current string, reached []string, highest string) AgeSight {
	ages := config.AgeByKey()
	far := ages[current].Order
	for _, a := range append([]string{highest}, reached...) {
		if def, ok := ages[a]; ok && def.Order > far {
			far = def.Order
		}
	}
	next := far
	if o := ages[current].Order + 1; o > next {
		next = o
	}
	return AgeSight{next: next, reached: far}
}

// SightOf is the sight of the player in the snapshot st.
func SightOf(st *GameState) AgeSight {
	highest := ""
	if st.AccountStats != nil {
		highest = st.AccountStats.HighestAge
	}
	return newAgeSight(st.Age, st.Stats.AgesReached, highest)
}

// ageSightLocked is the engine's own sight. Takes the account's lock; call
// under ge.mu (read or write), never from a bus handler.
func (ge *GameEngine) ageSightLocked() AgeSight {
	highest := ""
	if ge.account != nil {
		s, _ := ge.account.LifetimeStats()
		highest = s.HighestAge
	}
	return newAgeSight(ge.age, ge.Stats.AgesReached, highest)
}

// Age reports whether the player may see age named. Unknown keys: no.
func (s AgeSight) Age(age string) bool {
	def, ok := config.AgeByKey()[age]
	return ok && def.Order <= s.next
}

// Era reports whether the player may see the era named: they have reached
// one of its ages.
func (s AgeSight) Era(epoch string) bool {
	def, ok := config.EpochByKey()[epoch]
	if !ok || len(def.Ages) == 0 {
		return false
	}
	return config.AgeByKey()[def.Ages[0]].Order <= s.reached
}

// AgeRef names age for player text: "the Iron Age" when the player may see
// it named, "a later age" when not.
func (s AgeSight) AgeRef(age string) string {
	if s.Age(age) {
		return "the " + AgeName(age)
	}
	return "a later age"
}

// laterAgeRef is AgeRef for a refusal that only knows the current age: the
// next age by name, anything past it as "a later age".
func laterAgeRef(currentAge, age string) string {
	return newAgeSight(currentAge, nil, "").AgeRef(age)
}
