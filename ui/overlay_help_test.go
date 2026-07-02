package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestHelpProviderDevSectionGated verifies the Developer Console section (and the
// dev commands) appear in the Help overlay only while dev mode is active, and are
// hidden entirely during normal play.
func TestHelpProviderDevSectionGated(t *testing.T) {
	devCmds := []string{"/god", "/fill", "/give", "/build", "/techs", "/age", "/ages", "/prestige", "/speed"}

	// Dev mode OFF — no section, no commands.
	game.DevModeActive = false
	off := helpProvider(game.GameState{}, 0)
	if strings.Contains(off, "Developer Console") {
		t.Fatal("help must not show the Developer Console section when dev mode is off")
	}
	for _, c := range devCmds {
		if strings.Contains(off, c) {
			t.Fatalf("help must not list dev command %q when dev mode is off", c)
		}
	}

	// Dev mode ON — section + every command present.
	game.DevModeActive = true
	defer func() { game.DevModeActive = false }()
	on := helpProvider(game.GameState{}, 0)
	if !strings.Contains(on, "Developer Console") {
		t.Fatal("help must show the Developer Console section when dev mode is on")
	}
	for _, c := range devCmds {
		if !strings.Contains(on, c) {
			t.Fatalf("dev help missing command %q", c)
		}
	}
}
