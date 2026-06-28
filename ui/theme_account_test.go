package ui

import (
	"testing"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// Phase-2 bridge tests (theming.md §6): boot-apply, account-wide persistence, and
// the unlock gate. The active theme is process-global state, so every case that
// switches it registers restoreForge (theme_picker_test.go) to reset to the default
// in cleanup, keeping the suite order-independent.

// newIsolatedAccount mints a named account whose data root is an isolated temp dir,
// so SetActiveTheme writes a throwaway account.json instead of clobbering a real
// ./data/account.json. The temp dir + override are torn down via t.Cleanup. Returns
// a real, established account wired through the engine in the persist tests.
func newIsolatedAccount(t *testing.T) *game.Account {
	t.Helper()
	restore := game.SetDataDirForTest(t.TempDir())
	t.Cleanup(restore)
	acct, err := game.CreateNamedAccount("Theme Test Empire")
	if err != nil {
		t.Fatalf("CreateNamedAccount: %v", err)
	}
	return acct
}

// TestCmdThemePersistsToAccount covers the §6 CLI-persist path: `theme <key>`
// switches the active theme AND records it on the account, so a CLI switch survives
// saves. Uses an isolated-data-dir account wired through a real engine.
func TestCmdThemePersistsToAccount(t *testing.T) {
	restoreForge(t)
	const target = "high_contrast"
	if _, ok := theme.ByKey(target); !ok {
		t.Fatalf("precondition: theme %q not registered", target)
	}

	acct := newIsolatedAccount(t)
	engine := game.NewGameEngine()
	engine.SetAccount(acct)

	res := cmdTheme([]string{target}, engine)
	if res.Type != "success" {
		t.Fatalf("theme %s: got %q %q, want success", target, res.Type, res.Message)
	}
	if got := theme.Active().Key; got != target {
		t.Errorf("active theme = %q after switch, want %q", got, target)
	}
	if got := engine.Account().ActiveTheme(); got != target {
		t.Errorf("account ActiveTheme = %q after `theme %s`, want %q (CLI switch must persist)", got, target, target)
	}
}

// TestCmdThemeAccountlessDoesNotPanic confirms the persist path nil-guards: with no
// account wired, `theme <key>` still switches and reports success, it just doesn't
// persist. The game must run accountless.
func TestCmdThemeAccountlessDoesNotPanic(t *testing.T) {
	restoreForge(t)
	engine := game.NewGameEngine() // no SetAccount → Account() == nil

	res := cmdTheme([]string{"deuteranopia"}, engine)
	if res.Type != "success" {
		t.Errorf("accountless theme switch: got %q %q, want success", res.Type, res.Message)
	}
	if got := theme.Active().Key; got != "deuteranopia" {
		t.Errorf("active theme = %q, want deuteranopia", got)
	}
}

// TestApplyAccountThemeAppliesStored covers the §6 boot path: a stored active theme
// is read off the account and applied before the first Draw. Set the account's
// active theme to a non-default, call applyAccountTheme, and assert the process-
// global active theme matches.
func TestApplyAccountThemeAppliesStored(t *testing.T) {
	restoreForge(t)
	const stored = "high_contrast"

	acct := newIsolatedAccount(t)
	if err := acct.SetActiveTheme(stored); err != nil {
		t.Fatalf("SetActiveTheme: %v", err)
	}
	engine := game.NewGameEngine()
	engine.SetAccount(acct)

	applyAccountTheme(engine)

	if got := theme.Active().Key; got != stored {
		t.Errorf("applyAccountTheme: active = %q, want stored %q", got, stored)
	}
}

// TestApplyAccountThemeFallsBackToForge covers the §6 defensive fallbacks: an empty
// stored key (a brand-new account) and an unknown stored key both resolve to the
// default (Forge), never crashing. nil engine / nil account also fall back.
func TestApplyAccountThemeFallsBackToForge(t *testing.T) {
	restoreForge(t)

	// Start from a non-default theme so a real fallback to Forge is observable
	// (rather than already sitting on the default).
	if err := theme.SetActive("deuteranopia"); err != nil {
		t.Fatalf("setup SetActive(deuteranopia): %v", err)
	}

	// Empty stored key (fresh account) → Forge.
	acct := newIsolatedAccount(t) // CreateNamedAccount leaves ActiveTheme == ""
	engine := game.NewGameEngine()
	engine.SetAccount(acct)
	applyAccountTheme(engine)
	if got := theme.Active().Key; got != theme.DefaultKey {
		t.Errorf("empty stored key: active = %q, want default %q", got, theme.DefaultKey)
	}

	// Unknown stored key → Forge (and no crash).
	if err := theme.SetActive("deuteranopia"); err != nil {
		t.Fatalf("re-setup SetActive(deuteranopia): %v", err)
	}
	if err := acct.SetActiveTheme("no_such_theme_zzz"); err != nil {
		t.Fatalf("SetActiveTheme(unknown): %v", err)
	}
	applyAccountTheme(engine)
	if got := theme.Active().Key; got != theme.DefaultKey {
		t.Errorf("unknown stored key: active = %q, want default %q", got, theme.DefaultKey)
	}

	// nil engine → Forge, no panic.
	if err := theme.SetActive("deuteranopia"); err != nil {
		t.Fatalf("re-setup SetActive(deuteranopia): %v", err)
	}
	applyAccountTheme(nil)
	if got := theme.Active().Key; got != theme.DefaultKey {
		t.Errorf("nil engine: active = %q, want default %q", got, theme.DefaultKey)
	}
}

// TestThemeAvailableGate covers the unlock-gating predicate (theming.md §4/§5):
//   - Accessible themes: always available (never gated), even accountless.
//   - The default theme (Forge): always available.
//   - A flavor (non-accessible, non-default) theme: available only if the account
//     has unlocked it; locked otherwise, and locked when accountless.
//
// All shipped themes are Accessible/Forge today, so a hypothetical locked flavor key
// is simulated via a synthetic Theme literal (the registry is unexported; this keeps
// the test from depending on a not-yet-shipped Phase-3 theme).
func TestThemeAvailableGate(t *testing.T) {
	forge, ok := theme.ByKey("forge")
	if !ok {
		t.Fatal("precondition: forge not registered")
	}
	deut, ok := theme.ByKey("deuteranopia")
	if !ok {
		t.Fatal("precondition: deuteranopia not registered")
	}
	// Synthetic locked flavor theme: not accessible, not the default key.
	fakeLocked := theme.Theme{Key: "fake_locked_flavor", Name: "Fake", Accessible: false}

	// Accessible + Forge are available regardless of account (incl. nil).
	if !themeAvailable(nil, forge) {
		t.Error("forge should be available accountless (it is the default)")
	}
	if !themeAvailable(nil, deut) {
		t.Error("accessible theme should be available accountless")
	}

	acct := newIsolatedAccount(t)
	if !themeAvailable(acct, forge) || !themeAvailable(acct, deut) {
		t.Error("forge + accessible themes must always be available with an account")
	}

	// Flavor theme: locked until the account unlocks it; locked when accountless.
	if themeAvailable(nil, fakeLocked) {
		t.Error("a flavor theme should be locked accountless")
	}
	if themeAvailable(acct, fakeLocked) {
		t.Error("a flavor theme should be locked before the account unlocks it")
	}
	if _, err := acct.UnlockTheme(fakeLocked.Key); err != nil {
		t.Fatalf("UnlockTheme: %v", err)
	}
	if !themeAvailable(acct, fakeLocked) {
		t.Error("a flavor theme should be available once the account unlocks it")
	}
}

// TestThemeAvailableGate_DevBypass verifies the developer-console affordance:
// while game.DevModeActive is set, every theme is available (for preview),
// including a locked flavor theme on an account that has not unlocked it; and
// the bypass disappears the moment dev mode is off.
func TestThemeAvailableGate_DevBypass(t *testing.T) {
	prev := game.DevModeActive
	t.Cleanup(func() { game.DevModeActive = prev })

	locked := theme.Theme{Key: "fake_locked_flavor", Name: "Fake", Accessible: false}

	game.DevModeActive = false
	if themeAvailable(nil, locked) {
		t.Fatal("precondition: locked theme must be gated when dev mode is off")
	}

	game.DevModeActive = true
	if !themeAvailable(nil, locked) {
		t.Error("dev mode should unlock every theme for preview, even accountless")
	}

	game.DevModeActive = false
	if themeAvailable(nil, locked) {
		t.Error("the dev bypass must vanish once dev mode is off")
	}
}
