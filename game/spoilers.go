package game

import "github.com/espresso20/ageforge/rules"

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
	set     *rules.Set // the ruleset whose ages these are; nil reads the core set
	next    int        // the furthest age order that may be named
	reached int        // the furthest age order reached
}

// ageSightIn is the sight, on set's ages, of a player in current who has
// reached the ages in reached this run and highest on their account ("" for
// none).
func ageSightIn(set *rules.Set, current string, reached []string, highest string) AgeSight {
	here, _ := set.Index(current)
	far := here
	if o, ok := set.Index(highest); ok && o > far {
		far = o
	}
	for _, a := range reached {
		if o, ok := set.Index(a); ok && o > far {
			far = o
		}
	}
	return AgeSight{set: set, next: max(far, here+1), reached: far}
}

// SightOf is the sight of the player in the snapshot st.
func SightOf(st *GameState) AgeSight {
	highest := ""
	if st.AccountStats != nil {
		highest = st.AccountStats.HighestAge
	}
	return ageSightIn(st.Ruleset(), st.Age, st.Stats.AgesReached, highest)
}

// ageSightLocked is the engine's own sight. Takes the account's lock; call
// under ge.mu (read or write), never from a bus handler.
func (ge *GameEngine) ageSightLocked() AgeSight {
	highest := ""
	if ge.account != nil {
		s, _ := ge.account.LifetimeStats()
		highest = s.HighestAge
	}
	return ageSightIn(ge.rules, ge.age, ge.Stats.AgesReached, highest)
}

// Age reports whether the player may see age named. Unknown keys: no.
func (s AgeSight) Age(age string) bool {
	o, ok := orCore(s.set).Index(age)
	return ok && o <= s.next
}

// Reached reports whether the player has reached age: in this run, or on
// their account in any run. Unknown keys: no.
func (s AgeSight) Reached(age string) bool {
	o, ok := orCore(s.set).Index(age)
	return ok && o <= s.reached
}

// SeenNext reports whether age has ever been the player's next age, or
// reached: the age after the furthest they have been. The Next Age goal
// named it then, so it stays known in a later run that starts over from the
// first age. (Age is the stricter rule for text about the run in play.)
func (s AgeSight) SeenNext(age string) bool {
	o, ok := orCore(s.set).Index(age)
	return ok && o <= s.reached+1
}

// ReachedLast reports whether the player has reached the last age, so
// nothing about the ages is left to spoil.
func (s AgeSight) ReachedLast() bool {
	return s.reached >= orCore(s.set).NumAges()-1
}

// Era reports whether the player may see the era named: they have reached
// one of its ages.
func (s AgeSight) Era(epoch string) bool {
	o, ok := orCore(s.set).EraFirstAge(epoch)
	return ok && o <= s.reached
}

// AgeRef names age for player text: "the Iron Age" when the player may see
// it named, "a later age" when not.
func (s AgeSight) AgeRef(age string) string {
	if s.Age(age) {
		return "the " + orCore(s.set).Name(rules.KindAge, age)
	}
	return "a later age"
}

// laterAgeRef is AgeRef for a refusal that only knows the current age: the
// next age by name, anything past it as "a later age". set is the ruleset
// of the manager that refuses.
func laterAgeRef(set *rules.Set, currentAge, age string) string {
	return ageSightIn(set, currentAge, nil, "").AgeRef(age)
}

// orCore is set, or the core ruleset for a value built without one: a
// snapshot or a sight written by hand, as tests and fixtures do.
func orCore(set *rules.Set) *rules.Set {
	if set == nil {
		return rules.Core()
	}
	return set
}
