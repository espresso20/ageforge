package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/theme"
)

// A gated theme is the reward of a badge, and the engine unlocks it as it
// grants the badge (the game package tests that). These are the UI's part:
// what a locked theme says, and the line that announces an unlock.

// TestThemeUnlockToastFormat checks the toast text carries the cue and theme name.
func TestThemeUnlockToastFormat(t *testing.T) {
	restoreForge(t) // pin the active theme so the accent tag is deterministic
	msg := themeUnlockToast("Cyberpunk")
	if !strings.Contains(msg, "Cyberpunk") {
		t.Errorf("toast missing theme name: %q", msg)
	}
	if !strings.Contains(msg, "theme") {
		t.Errorf("toast should tell the player to use `theme`: %q", msg)
	}
	if !strings.Contains(msg, "🎨") {
		t.Errorf("toast missing the 🎨 cue: %q", msg)
	}
}

// TestThemeDetailShowsLockCondition: a LOCKED flavor theme's detail pane shows the
// 🔒 Locked line with its unlock hint, and still renders swatches as a preview.
func TestThemeDetailShowsLockCondition(t *testing.T) {
	cp, ok := theme.ByKey("cyberpunk")
	if !ok {
		t.Fatal("precondition: cyberpunk theme not registered")
	}
	// available=false → locked detail.
	d := themeDetailText(cp, false, nil)
	if !strings.Contains(d, "🔒") || !strings.Contains(d, "Locked") {
		t.Errorf("locked detail should show a lock marker\ngot: %s", d)
	}
	if !strings.Contains(d, cp.UnlockHint) {
		t.Errorf("locked detail should show the unlock hint %q\ngot: %s", cp.UnlockHint, d)
	}
	// Swatch preview still present (the role label "Accent" appears in the swatch rows).
	if !strings.Contains(d, "Accent") {
		t.Errorf("locked detail should still show swatch preview\ngot: %s", d)
	}
	// available=true → no lock line.
	if d := themeDetailText(cp, true, nil); strings.Contains(d, "Locked") {
		t.Errorf("available detail must not show a Locked line\ngot: %s", d)
	}
}

// TestCmdThemeListShowsLockHints: `theme list` for an account without the flavor
// themes shows the 🔒 hint on each locked theme and the accessible note on the
// always-available accessible ones.
func TestCmdThemeListShowsLockHints(t *testing.T) {
	restoreForge(t)
	acct := newIsolatedAccount(t) // owns only the default-unlock set
	res := cmdThemeList(acct, nil)
	if res.Type == "error" {
		t.Fatalf("theme list errored: %q", res.Message)
	}
	// Cyberpunk is a gated flavor theme this account hasn't unlocked → its hint shows.
	cp, _ := theme.ByKey("cyberpunk")
	if !strings.Contains(res.Message, cp.UnlockHint) {
		t.Errorf("theme list should show locked %q's hint %q\ngot: %s", cp.Key, cp.UnlockHint, res.Message)
	}
	if !strings.Contains(res.Message, "🔒") {
		t.Errorf("theme list should show a 🔒 marker for locked themes\ngot: %s", res.Message)
	}
}

// --- small slice helpers (test-local) ---

func sliceContains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

func countOccurrences(s []string, v string) int {
	n := 0
	for _, x := range s {
		if x == v {
			n++
		}
	}
	return n
}
