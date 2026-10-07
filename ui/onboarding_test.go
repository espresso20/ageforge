package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestOnboardingCommandsRun runs every command the first-steps guide teaches,
// in order, on a fresh game and fails if any of them is refused. The guide
// once taught `recruit worker`, which the handler rejects; this keeps the
// guide honest. Stock is topped up so a refusal means bad syntax, not an
// empty larder, and the construction queue is finished between steps so the
// hut exists before the guide asks for recruits.
func TestOnboardingCommandsRun(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	ge := game.NewGameEngine()
	ge.SeedRNG(1)
	stock := map[string]float64{}
	for k := range ge.GetState().Resources {
		stock[k] = 1e6
	}
	ge.Resources.LoadAmounts(stock)

	cmds := onboardingCommands()
	if len(cmds) == 0 {
		t.Fatal("onboarding teaches no commands")
	}
	for _, c := range cmds {
		res := HandleCommand(c, ge)
		if res.Type == "error" {
			t.Errorf("onboarding command %q was refused: %s", c, res.Message)
		}
		finishConstruction(t, ge)
	}
}

// finishConstruction steps the engine until the build queue is empty.
func finishConstruction(t *testing.T, ge *game.GameEngine) {
	t.Helper()
	for i := 0; i < 500 && len(ge.GetState().BuildQueue) > 0; i++ {
		ge.StepTicks(10)
	}
	if n := len(ge.GetState().BuildQueue); n > 0 {
		t.Fatalf("build queue still has %d items after 5,000 ticks", n)
	}
}

// TestOnboardingRendersEveryCommand checks the panel text shows each command
// the test above runs, so the data and the rendering cannot drift apart.
func TestOnboardingRendersEveryCommand(t *testing.T) {
	out := renderOnboarding(200)
	for _, c := range onboardingCommands() {
		if !strings.Contains(out, "[cyan]"+c+"[-]") {
			t.Errorf("onboarding panel does not show %q:\n%s", c, out)
		}
	}
}
