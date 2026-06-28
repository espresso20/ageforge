package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// Milestone-gated theme unlock tests (theming.md §5). These exercise the pure-ish
// evaluateThemeUnlock helper (theme_unlock.go) rather than the dashboard method, so
// no tview app is needed; the dashboard just loops this helper over completed keys.

// cyberpunkUnlockKey is the milestone key that grants the Cyberpunk theme. Pulled
// from the registry so the test tracks the mapping rather than hard-coding it twice.
func cyberpunkUnlockKey(t *testing.T) (milestoneKey, themeKey string) {
	t.Helper()
	cp, ok := theme.ByKey("cyberpunk")
	if !ok {
		t.Fatal("precondition: cyberpunk theme not registered")
	}
	if cp.UnlockMilestone == "" {
		t.Fatal("precondition: cyberpunk theme has no UnlockMilestone")
	}
	return cp.UnlockMilestone, cp.Key
}

// TestEvaluateThemeUnlockNewThenReplay is the core idempotency case: a milestone the
// account doesn't yet own unlocks the mapped theme (newly → toast); replaying the
// same key is a no-op (already owned → no toast, no double-unlock).
func TestEvaluateThemeUnlockNewThenReplay(t *testing.T) {
	msKey, themeKey := cyberpunkUnlockKey(t)
	acct := newIsolatedAccount(t)

	if acct.HasTheme(themeKey) {
		t.Fatalf("precondition: account already has %q", themeKey)
	}

	// Live play (firstSync=false): a genuinely-new unlock toasts.
	res := evaluateThemeUnlock(acct, msKey, false)
	if !res.Toast {
		t.Errorf("first unlock of %q should toast", themeKey)
	}
	if !acct.HasTheme(themeKey) {
		t.Errorf("after evaluateThemeUnlock, account should have %q unlocked", themeKey)
	}
	if !strings.Contains(res.ThemeName, "Cyberpunk") {
		t.Errorf("toast theme name = %q, want it to contain the theme's display name", res.ThemeName)
	}

	// Replay the same completed key: UnlockTheme is idempotent (newly==false), so no
	// re-toast and the theme stays unlocked exactly once.
	res2 := evaluateThemeUnlock(acct, msKey, false)
	if res2.Toast {
		t.Errorf("replaying %q must not re-toast an already-owned theme", msKey)
	}
	if got := acct.UnlockedThemes(); countOccurrences(got, themeKey) != 1 {
		t.Errorf("theme %q should be unlocked exactly once, got set %v", themeKey, got)
	}
}

// TestEvaluateThemeUnlockFirstSyncSilent covers retroactive grants: on the first sync
// after load, an already-completed milestone unlocks its theme SILENTLY (no toast),
// so a player who reached the age before this update gets the theme without a launch-
// time toast spam. The unlock still happens — only the toast is suppressed.
func TestEvaluateThemeUnlockFirstSyncSilent(t *testing.T) {
	msKey, themeKey := cyberpunkUnlockKey(t)
	acct := newIsolatedAccount(t)

	res := evaluateThemeUnlock(acct, msKey, true) // firstSync=true
	if res.Toast {
		t.Errorf("first-sync (retroactive) unlock of %q must NOT toast", themeKey)
	}
	if !acct.HasTheme(themeKey) {
		t.Errorf("first-sync should still UNLOCK %q (silent grant), but account lacks it", themeKey)
	}
}

// TestEvaluateThemeUnlockUnmappedKey: a completed milestone that grants no theme is a
// no-op and never toasts.
func TestEvaluateThemeUnlockUnmappedKey(t *testing.T) {
	acct := newIsolatedAccount(t)
	res := evaluateThemeUnlock(acct, "first_shelter", false)
	if res.Toast {
		t.Errorf("an unmapped milestone (first_shelter) must not toast")
	}
	if got := acct.UnlockedThemes(); len(got) != 0 {
		t.Errorf("an unmapped milestone must not unlock anything, got %v", got)
	}
}

// TestEvaluateThemeUnlockNilAccount: accountless play has no store, so a mapped key
// unlocks nothing and never toasts (must not panic).
func TestEvaluateThemeUnlockNilAccount(t *testing.T) {
	msKey, _ := cyberpunkUnlockKey(t)
	res := evaluateThemeUnlock(nil, msKey, false)
	if res.Toast {
		t.Errorf("accountless unlock must not toast")
	}
}

// TestCompletedUnlockKeysGathersBoth verifies the snapshot reader collects completed
// milestone keys AND completed chain keys, and skips incomplete ones.
func TestCompletedUnlockKeysGathersBoth(t *testing.T) {
	ms := game.MilestoneState{
		Milestones: map[string]game.MilestoneInfo{
			"done_ms":    {Completed: true},
			"pending_ms": {Completed: false},
		},
		Chains: []game.ChainInfo{
			{Key: "done_chain", Complete: true},
			{Key: "pending_chain", Complete: false},
		},
	}
	got := completedUnlockKeys(ms)
	if !sliceContains(got, "done_ms") {
		t.Errorf("completed milestone done_ms missing from %v", got)
	}
	if !sliceContains(got, "done_chain") {
		t.Errorf("completed chain done_chain missing from %v", got)
	}
	if sliceContains(got, "pending_ms") || sliceContains(got, "pending_chain") {
		t.Errorf("incomplete keys must be excluded, got %v", got)
	}
	if len(got) != 2 {
		t.Errorf("expected exactly 2 completed keys, got %v", got)
	}
}

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
	d := themeDetailText(cp, false)
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
	if d := themeDetailText(cp, true); strings.Contains(d, "Locked") {
		t.Errorf("available detail must not show a Locked line\ngot: %s", d)
	}
}

// TestCmdThemeListShowsLockHints: `theme list` for an account without the flavor
// themes shows the 🔒 hint on each locked theme and the accessible note on the
// always-available accessible ones.
func TestCmdThemeListShowsLockHints(t *testing.T) {
	restoreForge(t)
	acct := newIsolatedAccount(t) // owns only the default-unlock set
	res := cmdThemeList(acct)
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
