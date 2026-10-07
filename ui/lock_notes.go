package ui

import (
	"strings"

	"github.com/espresso20/ageforge/game"
)

// lockNotes is a line for each of the feature locks keys that holds its
// command shut, saying which tech opens it, in the words the game uses when
// it refuses the command: "Campaigns need Military Tactics first. Research
// it to send one." A lock is only mentioned once its tech's age is reached,
// so a panel never names a tech of an age the player has not seen. "" when
// every one is open.
func lockNotes(state game.GameState, keys ...string) string {
	set := state.Ruleset()
	here, _ := set.Index(state.Age)
	var sb strings.Builder
	for _, key := range keys {
		f, ok := state.Features[key]
		if !ok || f.Open {
			continue
		}
		def, _ := set.FeatureLock(key)
		tech, _ := set.Tech(f.Tech)
		if at, ok := set.Index(tech.Age); !ok || at > here {
			continue
		}
		sb.WriteString(" [yellow]" + def.Refusal(f.TechName) + "[-]\n")
	}
	return sb.String()
}
