package smoke

import "github.com/espresso20/ageforge/game"

// PrestigePoints is the prestige formula as site/docs/prestige.md documents
// it: depth points, the sum of 3^epoch over every age before the one
// prestiged from (1 per Stone Era age, 3 Iron, 9 Steel, 27 Electric, 81
// Digital, 243 Neon, 729 Cosmic), with no divisor. It is computed here from
// the epoch table on its own, so a change to config.DepthPoints that leaves
// the docs behind fails the check.
func PrestigePoints(st game.GameState) int {
	total := 0
	for _, ep := range st.Ruleset().Eras() {
		w := 1
		for i := 0; i < ep.Order; i++ {
			w *= 3
		}
		for _, a := range ep.Ages {
			if a == st.Age {
				return total
			}
			total += w
		}
	}
	return 0
}
