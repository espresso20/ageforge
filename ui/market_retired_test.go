package ui

import (
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// The market's swap is retired: a town makes what it needs, and the market
// no longer turns one resource into another. `trade <give> <get> <amount>`,
// `trade list` and `plan trade` are gone outright, as `speed` is: from the
// registry (and with it help and completion) and from HandleCommand, where
// they answer as any unknown command does. Bare `trade` still opens the
// panel, which now shows trade routes only; `trade route ...` and
// `trade black ...` are untouched.

// TestMarketSwapCommandsRetired: a player can neither find nor run the swap.
func TestMarketSwapCommandsRetired(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	eng := game.NewGameEngine()

	for _, info := range Commands() {
		switch info.Names[0] {
		case "trade":
			if slices.Contains(info.Subs, "list") || info.Accepts["list"] {
				t.Errorf("the registry still lists trade list: %v", info.Subs)
			}
		case "plan":
			if slices.Contains(info.Subs, "trade") || info.Accepts["trade"] {
				t.Errorf("the registry still lists plan trade: %v", info.Subs)
			}
		}
	}
	help := plainText(helpProvider(game.GameState{}, 0))
	for _, gone := range []string{"trade <give>", "trade list", "plan trade", "market rate"} {
		if strings.Contains(help, gone) {
			t.Errorf("help still mentions %q", gone)
		}
	}

	comp := NewAutoCompleter(eng)
	if got := comp("trade "); !slices.Equal(got, []string{"trade black", "trade route"}) {
		t.Errorf("completing `trade ` = %v, want black and route only", got)
	}
	if got := comp("plan "); slices.Contains(got, "plan trade") {
		t.Errorf("completing `plan ` offers plan trade: %v", got)
	}

	for line, name := range map[string]string{
		"trade list":             "trade list",
		"t list":                 "trade list",
		"trade food wood 10":     "trade food",
		"trade gold knowledge":   "trade gold",
		"plan trade food wood":   "plan trade",
		"plan trade food wood 5": "plan trade",
	} {
		res := HandleCommand(line, eng)
		if res.Type != "error" || !strings.Contains(res.Message, "Unknown command '"+name+"'") {
			t.Errorf("%q = %+v, want the unknown command reply for %q", line, res, name)
		}
	}
	if res := HandleCommand("trade", eng); res.OverlayName != "trade" || res.Type == "error" {
		t.Errorf("bare trade = %+v, want the Trade routes panel", res)
	}
}
