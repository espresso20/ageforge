package game

import (
	"sync"

	"github.com/espresso20/ageforge/config"
)

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

// ageOrders maps each age key to its order, and eraFirstAges each epoch key
// to the order of its first age. Built once: config rebuilds its tables on
// every call, and these are asked per milestone on every GetState.
var (
	ageOrders = sync.OnceValue(func() map[string]int {
		m := map[string]int{}
		for _, a := range config.Ages() {
			m[a.Key] = a.Order
		}
		return m
	})
	eraFirstAges = sync.OnceValue(func() map[string]int {
		ages := ageOrders()
		m := map[string]int{}
		for _, e := range config.Epochs() {
			if len(e.Ages) > 0 {
				m[e.Key] = ages[e.Ages[0]]
			}
		}
		return m
	})
)

// AgeSight is which ages and eras the player may see named.
type AgeSight struct {
	next    int // the furthest age order that may be named
	reached int // the furthest age order reached
}

// newAgeSight is the sight of a player in current who has reached the ages
// in reached this run and highest on their account ("" for none).
func newAgeSight(current string, reached []string, highest string) AgeSight {
	ages := ageOrders()
	far := ages[current]
	if o, ok := ages[highest]; ok && o > far {
		far = o
	}
	for _, a := range reached {
		if o, ok := ages[a]; ok && o > far {
			far = o
		}
	}
	return AgeSight{next: max(far, ages[current]+1), reached: far}
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
	o, ok := ageOrders()[age]
	return ok && o <= s.next
}

// Era reports whether the player may see the era named: they have reached
// one of its ages.
func (s AgeSight) Era(epoch string) bool {
	o, ok := eraFirstAges()[epoch]
	return ok && o <= s.reached
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
