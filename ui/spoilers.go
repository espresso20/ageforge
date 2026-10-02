package ui

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/theme"
)

// spoilers.go holds the UI side of the no-spoiler rule (game/spoilers.go): a
// screen never names an age the player cannot see yet (past the next one,
// unless reached before), an era they have not reached, or a civilization
// they have not met. The Next Age goal names the next age on purpose.
// spoiler_guard_test.go holds every panel of a fresh game to it.

// ageRef names age for player text in state: "the Iron Age", or "a later
// age" when the player cannot see it named yet.
func ageRef(state game.GameState, age string) string {
	return game.SightOf(&state).AgeRef(age)
}

// ageRefCap is ageRef for the start of a sentence: "The Iron Age", "A later
// age".
func ageRefCap(state game.GameState, age string) string {
	return textfmt.Capitalize(ageRef(state, age))
}

// civRef names civilization key for player text in state: its name once
// the player has met it, "a civilization you have not met" until then.
func civRef(state game.GameState, key string) string {
	if f, ok := state.Diplomacy.Factions[key]; ok && f.Discovered {
		if f.Name != "" {
			return f.Name
		}
		return game.CivName(key)
	}
	return "a civilization you have not met"
}

// buildingAge is the age that unlocks building key, from the snapshot, or
// fallback when the snapshot does not list it.
func buildingAge(state game.GameState, key, fallback string) string {
	if bs, ok := state.Buildings[key]; ok && bs.AgeKey != "" {
		return bs.AgeKey
	}
	return fallback
}

// currentEraName is the name of the era the player is in ("Stone Era"), or
// "this era" on a state that carries none.
func currentEraName(state game.GameState) string {
	if state.EpochName != "" {
		return state.EpochName
	}
	return "this era"
}

// themeUnlockHint is a locked theme's unlock condition in words the player
// may read: a theme earned in an age they cannot see yet says so without
// naming it.
func themeUnlockHint(t theme.Theme, state game.GameState) string {
	if m, ok := config.MilestoneByKey()[t.UnlockMilestone]; ok && m.MinAge != "" && !game.SightOf(&state).Age(m.MinAge) {
		return "Reach a later age"
	}
	return t.UnlockHint
}
