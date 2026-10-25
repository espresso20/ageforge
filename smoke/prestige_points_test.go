package smoke

import (
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// TestPrestigeFormulaMatchesDocs: the worked examples in
// site/docs/prestige.md. Depth points: 3^epoch per completed age, with no
// divisor, so the level and the run's milestones, techs and buildings
// change nothing.
func TestPrestigeFormulaMatchesDocs(t *testing.T) {
	for _, c := range []struct {
		age  string
		want int
	}{
		{"medieval_age", 9},     // 1+1+1 (Stone Era) + 3+3 (Iron, Classical)
		{"renaissance_age", 12}, // + 3 for the Medieval Age
		{"modern_age", 120},     // 3 + 9 + 27 + 81
		{"information_age", 201},
		{"cyberpunk_age", 363}, // a run through the Digital Age
		{"interstellar_age", 1092},
		{"transcendent_age", 3279},
	} {
		st := game.GameState{Age: c.age}
		st.Milestones.CompletedCount, st.Research.TotalResearched, st.Stats.TotalBuilt = 25, 31, 149
		st.Prestige.Level = 7
		if got := PrestigePoints(st); got != c.want {
			t.Errorf("%s: %d points, want %d", c.age, got, c.want)
		}
		if got := config.DepthPoints(c.age); got != c.want {
			t.Errorf("config.DepthPoints(%s) = %d, want %d", c.age, got, c.want)
		}
	}
}
