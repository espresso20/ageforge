package ui

import (
	"slices"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// The player `speed` command is retired: game speed is fixed at 1x so the
// calendar paces the game. The registry has no hidden commands, so it is gone
// outright, from the registry (and with it help, completion, the docsync
// check and the fuzz corpus) and from HandleCommand. The dev console keeps
// /speed, behind dev mode.

// TestSpeedCommandRetired: a player can neither find nor run `speed`.
func TestSpeedCommandRetired(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	prev := game.DevModeActive
	game.DevModeActive = false
	t.Cleanup(func() { game.DevModeActive = prev })
	eng := game.NewGameEngine()

	for _, info := range Commands() {
		if slices.Contains(info.Names, "speed") {
			t.Errorf("the registry still lists speed (as %v)", info.Names)
		}
	}

	help := plainText(helpProvider(game.GameState{}, 0))
	for _, line := range strings.Split(help, "\n") {
		if l := strings.TrimSpace(line); l == "speed" || strings.HasPrefix(l, "speed ") {
			t.Errorf("help still lists the speed command: %q", l)
		}
	}
	if strings.Contains(strings.ToLower(help), "speed cap") {
		t.Error("help still mentions the speed cap")
	}

	comp := NewAutoCompleter(eng)
	for _, typed := range []string{"s", "sp", "spe", "speed", "speed "} {
		for _, got := range comp(typed) {
			if strings.HasPrefix(got, "speed") {
				t.Errorf("completing %q offers %q", typed, got)
			}
		}
	}

	for _, line := range []string{"speed", "speed 2", "speed 1.5", "/speed 4"} {
		res := HandleCommand(line, eng)
		word := strings.Fields(line)[0]
		if res.Type != "error" || !strings.Contains(res.Message, "Unknown command '"+word+"'") {
			t.Errorf("%q = %+v, want the unknown command reply", line, res)
		}
	}
	if got := eng.GetState().TickIntervalMs; got != int(game.BaseTickInterval.Milliseconds()) {
		t.Errorf("tick interval %dms after the retired commands, want the 1x %dms", got, game.BaseTickInterval.Milliseconds())
	}
}

// TestDevSpeedCompletionNeedsDevMode: the dev console's /speed completes only
// while dev mode is on.
func TestDevSpeedCompletionNeedsDevMode(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	prev := game.DevModeActive
	t.Cleanup(func() { game.DevModeActive = prev })
	comp := NewAutoCompleter(game.NewGameEngine())

	game.DevModeActive = false
	if got := comp("/spe"); len(got) != 0 {
		t.Errorf("completing /spe without dev mode = %v, want nothing", got)
	}
	game.DevModeActive = true
	if got := comp("/spe"); !slices.Contains(got, "/speed") {
		t.Errorf("completing /spe with dev mode = %v, want /speed", got)
	}
}
