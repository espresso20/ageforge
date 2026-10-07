package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestHelpKeepsTwoColumns: at any width the Help panel's rows are a command
// or a key, then what it does, and a description too long for its line
// continues under itself, never under the command column. No line is left
// for the terminal to wrap.
func TestHelpKeepsTwoColumns(t *testing.T) {
	for _, screenW := range []int{80, 100, 120, 144, 200} {
		width := overlayTextWidth(screenW)
		lines := strings.Split(visible(safeTags(helpProvider(game.GameState{}, screenW))), "\n")
		wrapped := 0
		for i, line := range lines {
			if n := runeLen(line); n > width {
				t.Errorf("%d columns: line %d is %d cells in a panel %d wide: %q", screenW, i+1, n, width, line)
			}
			at := strings.Index(line, " - ")
			if !strings.HasPrefix(line, "  ") || at < 0 {
				continue
			}
			// A row: the lines after it that are deeper than a row's
			// two-space start are its description going on, and must start
			// where the description did.
			col := runeLen(line[:at]) + 3
			for _, next := range lines[i+1:] {
				indent := len(next) - len(strings.TrimLeft(next, " "))
				if indent <= 2 || strings.TrimSpace(next) == "" {
					break
				}
				wrapped++
				if indent != col && indent != 6 {
					t.Errorf("%d columns: a description continues at column %d, not under itself at %d:\n%s\n%s", screenW, indent, col, line, next)
				}
			}
		}
		if screenW == 100 && wrapped == 0 {
			t.Error("100 columns: no description wrapped, so the test checks nothing")
		}
	}
	// The dashboard's own keys are listed, the new one among them.
	help := visible(safeTags(helpProvider(game.GameState{}, 0)))
	for _, want := range []string{"The dashboard", "PgUp/PgDn - Scroll the Buildings list", resourcePageKeyName + "    - The next page of the Resources box"} {
		if !strings.Contains(help, want) {
			t.Errorf("the Help panel does not list %q", want)
		}
	}
}
