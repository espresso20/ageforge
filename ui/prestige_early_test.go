package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// `prestige confirm` before the Modern Age says plainly what the early
// prestige pays and that a deeper run pays far more; from the Modern Age on
// it says nothing of the kind.
func TestPrestigeConfirmSaysAnEarlyPrestigePaysLittle(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	engine := game.NewGameEngine()
	engine.SeedRNG(1)
	for _, age := range []string{"stone_age", "bronze_age", "iron_age", "classical_age", "medieval_age"} {
		if err := engine.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	confirm := untag(HandleCommand("prestige confirm", engine).Message)
	for _, want := range []string{
		"You will earn 9 prestige points.",
		"This is an early taste: a prestige from the Medieval Age pays 9 prestige points, for the 5 ages this run completed. " +
			"Going deeper pays far more: each era's ages are worth 3 times the era before, and a full run, 7 ages further on, pays 120 prestige points.",
	} {
		if !strings.Contains(confirm, want) {
			t.Errorf("prestige confirm in the Medieval Age does not say %q:\n%s", want, confirm)
		}
	}
	if strings.Contains(confirm, "Modern Age") {
		t.Errorf("prestige confirm names the Modern Age before the player has seen it:\n%s", confirm)
	}

	for _, age := range []string{"renaissance_age", "colonial_age", "industrial_age", "victorian_age", "electric_age", "atomic_age", "modern_age"} {
		if err := engine.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	confirm = untag(HandleCommand("prestige confirm", engine).Message)
	if strings.Contains(confirm, "early taste") || !strings.Contains(confirm, "You will earn 120 prestige points.") {
		t.Errorf("prestige confirm in the Modern Age is a full run's:\n%s", confirm)
	}
}
