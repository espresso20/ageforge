package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
)

// lockNotes is a line for each of the feature locks keys that holds its
// command shut, saying which tech opens it, in the words the game uses when
// it refuses the command: "Campaigns need Military Tactics first. Research
// it to send one." A lock is only mentioned once its tech's age is reached,
// so a panel never names a tech of an age the player has not seen. The
// lines end with a blank one; "" when every lock is open.
func lockNotes(state game.GameState, keys ...string) string {
	set := state.Ruleset()
	here, _ := set.Index(state.Age)
	return lockNotesWhen(state, func(age string) bool {
		at, ok := set.Index(age)
		return ok && at <= here
	}, keys...)
}

// lockNotesInSight is lockNotes for a panel that already lists what the
// lock shuts: the Expeditions panel shows Scout Nearby Ruins an age before
// Exploration can be researched. It mentions a lock as soon as its tech may
// be seen named (the next age's too), so the list never reads as open when
// it is not.
func lockNotesInSight(state game.GameState, keys ...string) string {
	return lockNotesWhen(state, game.SightOf(&state).Age, keys...)
}

// lockNotesWhen writes the notes for the locks whose tech's age shown
// accepts.
func lockNotesWhen(state game.GameState, shown func(age string) bool, keys ...string) string {
	set := state.Ruleset()
	var sb strings.Builder
	for _, key := range keys {
		f, ok := state.Features[key]
		if !ok || f.Open {
			continue
		}
		def, _ := set.FeatureLock(key)
		tech, _ := set.Tech(f.Tech)
		if !shown(tech.Age) {
			continue
		}
		sb.WriteString(" [yellow]" + def.Refusal(f.TechName) + "[-]\n")
	}
	if sb.Len() > 0 {
		sb.WriteString("\n") // a blank line sets the notes off from the list under them
	}
	return sb.String()
}
