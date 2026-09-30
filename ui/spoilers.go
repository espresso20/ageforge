package ui

import "github.com/espresso20/ageforge/game"

// spoilers.go holds the no-spoiler rule for player text: a screen never names
// an age or an era the player has not reached. The one exception is the Next
// Age goal, which names the next age on purpose.

// currentEraName is the name of the era the player is in ("Stone Era"), or
// "this era" on a state that carries none.
func currentEraName(state game.GameState) string {
	if state.EpochName != "" {
		return state.EpochName
	}
	return "this era"
}
