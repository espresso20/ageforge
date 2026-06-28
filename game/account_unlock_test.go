package game

import (
	"reflect"
	"testing"
)

// TestUnlockThemeNewlyAndHasTheme covers the core unlock contract: the first
// UnlockTheme returns newly=true, a repeat returns false, and HasTheme reflects the
// unlocked set throughout (accounts.md §8).
func TestUnlockThemeNewlyAndHasTheme(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}

	if acct.HasTheme("amber_crt") {
		t.Fatal("HasTheme(amber_crt) true before any unlock")
	}

	newly, err := acct.UnlockTheme("amber_crt")
	if err != nil {
		t.Fatalf("UnlockTheme (first): %v", err)
	}
	if !newly {
		t.Error("first UnlockTheme returned newly=false, want true")
	}
	if !acct.HasTheme("amber_crt") {
		t.Error("HasTheme(amber_crt) false after unlock")
	}

	newly, err = acct.UnlockTheme("amber_crt")
	if err != nil {
		t.Fatalf("UnlockTheme (second): %v", err)
	}
	if newly {
		t.Error("second UnlockTheme returned newly=true, want false")
	}

	// Re-unlocking must not have duplicated the key.
	if got := acct.UnlockedThemes(); len(got) != 1 || got[0] != "amber_crt" {
		t.Errorf("UnlockedThemes after re-unlock = %v, want [amber_crt]", got)
	}
}

// TestUnlockThemePersistsAcrossRestart confirms an unlock survives a fresh
// LoadOrCreate (the on-disk store), with the signature still valid and no tamper
// flag — i.e. UnlockTheme persisted via a properly-signed Save.
func TestUnlockThemePersistsAcrossRestart(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (first): %v", err)
	}
	if _, err := acct.UnlockTheme("obsidian"); err != nil {
		t.Fatalf("UnlockTheme: %v", err)
	}

	reloaded, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (reload): %v", err)
	}
	if !reloaded.HasTheme("obsidian") {
		t.Error("HasTheme(obsidian) false after restart")
	}
	if reloaded.Tampered {
		t.Error("reloaded account flagged Tampered after UnlockTheme persisted")
	}
	if !verifyAccount(reloaded) {
		t.Error("reloaded account fails signature verification after UnlockTheme")
	}
}

// TestUnlockThemeSurvivesReset asserts unlocks live at the account layer, not in a
// save: an account-level unlock is unaffected by a new-game flow. Reset() preserves
// the *Account (engine-level coverage is TestResetPreservesAccount); here we assert
// the account store itself keeps the unlock when reloaded fresh, mimicking the
// account a new game would carry forward.
func TestUnlockThemeSurvivesReset(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	if _, err := acct.UnlockTheme("cyberpunk"); err != nil {
		t.Fatalf("UnlockTheme: %v", err)
	}

	// A new game reloads the persisted account — the unlock must still be present.
	afterNewGame, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (new game): %v", err)
	}
	if !afterNewGame.HasTheme("cyberpunk") {
		t.Error("HasTheme(cyberpunk) false after new-game reload; unlocks must be account-wide")
	}
}

// TestActiveThemePersistsAcrossRestart covers the active-theme pref: it is "" before
// any set, persists via SetActiveTheme, and reloads across a fresh LoadOrCreate.
func TestActiveThemePersistsAcrossRestart(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (first): %v", err)
	}
	if got := acct.ActiveTheme(); got != "" {
		t.Errorf("ActiveTheme() = %q before any set, want \"\"", got)
	}

	if err := acct.SetActiveTheme("obsidian"); err != nil {
		t.Fatalf("SetActiveTheme: %v", err)
	}
	if got := acct.ActiveTheme(); got != "obsidian" {
		t.Errorf("ActiveTheme() = %q after set, want obsidian", got)
	}

	reloaded, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (reload): %v", err)
	}
	if got := reloaded.ActiveTheme(); got != "obsidian" {
		t.Errorf("ActiveTheme() = %q after restart, want obsidian", got)
	}
	if reloaded.Tampered || !verifyAccount(reloaded) {
		t.Error("reloaded account tampered or fails verification after SetActiveTheme")
	}
}

// TestSetActiveThemeIsKeyAgnostic confirms accounts does not validate the key: a key
// that was never unlocked is still persisted. Theming owns validity/always-unlocked.
func TestSetActiveThemeIsKeyAgnostic(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	if err := acct.SetActiveTheme("never_unlocked"); err != nil {
		t.Fatalf("SetActiveTheme rejected an unvalidated key: %v", err)
	}
	if got := acct.ActiveTheme(); got != "never_unlocked" {
		t.Errorf("ActiveTheme() = %q, want never_unlocked (accounts is key-agnostic)", got)
	}
}

// TestUnlockedThemesSorted confirms UnlockedThemes returns a deterministic, sorted
// list regardless of unlock insertion order.
func TestUnlockedThemesSorted(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	for _, k := range []string{"obsidian", "amber_crt", "parchment", "cyberpunk"} {
		if _, err := acct.UnlockTheme(k); err != nil {
			t.Fatalf("UnlockTheme(%q): %v", k, err)
		}
	}

	want := []string{"amber_crt", "cyberpunk", "obsidian", "parchment"}
	if got := acct.UnlockedThemes(); !reflect.DeepEqual(got, want) {
		t.Errorf("UnlockedThemes() = %v, want %v (sorted)", got, want)
	}
}
