package game

import (
	"encoding/json"
	"os"
	"testing"
)

// Phase 2 (accounts.md §6) — startup wiring + lazy save attribution.
//
// These tests exercise the migration invariant from §3.5/§6: a save gains its
// account_id ONLY by being re-written through SaveGame (which re-signs _sig over
// the new payload), never by an in-place JSON edit. They use isolateAccountDir
// (see account_test.go) so both the account file and the saves directory resolve
// inside a temp dir — no real ./data is read or clobbered.

// TestLegacySaveWithoutAccountIDLoadsClean covers the backward-compat path: a save
// written with NO account set must load fine, verify clean (sig valid), and pick up
// no spurious CheaterBadge. This is the "legacy save loads unchanged" guarantee.
func TestLegacySaveWithoutAccountIDLoadsClean(t *testing.T) {
	isolateAccountDir(t)

	name := "phase2_legacy_no_account"
	ge := NewGameEngine() // no SetAccount → AccountID() == ""
	if got := ge.AccountID(); got != "" {
		t.Fatalf("AccountID() = %q on an engine with no account, want \"\"", got)
	}
	if err := ge.SaveGame(name); err != nil {
		t.Fatalf("SaveGame(%q): %v", name, err)
	}

	// Read the bytes back: omitempty must keep account_id out of the file entirely.
	data, err := os.ReadFile(savePath(name))
	if err != nil {
		t.Fatalf("read save: %v", err)
	}
	var onDisk map[string]any
	if err := json.Unmarshal(data, &onDisk); err != nil {
		t.Fatalf("unmarshal save: %v", err)
	}
	if _, present := onDisk["account_id"]; present {
		t.Errorf("account_id present in a no-account save; omitempty should drop it")
	}

	// Parse + verify as the loader would: unsigned/valid sig, no cheater badge.
	var save GameSave
	if err := json.Unmarshal(data, &save); err != nil {
		t.Fatalf("parse save: %v", err)
	}
	if save.AccountID != "" {
		t.Errorf("legacy save AccountID = %q, want empty", save.AccountID)
	}
	sigValid, _ := verifySave(&save)
	if !sigValid {
		t.Error("legacy no-account save failed signature verification")
	}

	// And the full load path must not flag a cheater badge.
	loaded := NewGameEngine()
	if err := loaded.LoadGame(name); err != nil {
		t.Fatalf("LoadGame(%q): %v", name, err)
	}
	if loaded.GetState().CheaterBadge {
		t.Error("legacy no-account save wrongly flagged CheaterBadge on load")
	}
}

// TestSaveStampsAccountIDAndStaysSigned covers the core Phase 2 behavior: with an
// account set on the engine, SaveGame stamps account_id into the save AND the file
// still verifies — proving _sig was recomputed over the new field (the §6 invariant
// that stamping rides SaveGame, not an in-place edit).
func TestSaveStampsAccountIDAndStaysSigned(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	if acct.AccountID == "" {
		t.Fatal("created account has empty AccountID")
	}

	name := "phase2_stamped"
	ge := NewGameEngine()
	ge.SetAccount(acct)
	if got := ge.AccountID(); got != acct.AccountID {
		t.Fatalf("AccountID() = %q, want %q", got, acct.AccountID)
	}
	if err := ge.SaveGame(name); err != nil {
		t.Fatalf("SaveGame(%q): %v", name, err)
	}

	// Reload the file from disk: account_id must be present and equal the account's ID.
	data, err := os.ReadFile(savePath(name))
	if err != nil {
		t.Fatalf("read save: %v", err)
	}
	var save GameSave
	if err := json.Unmarshal(data, &save); err != nil {
		t.Fatalf("parse save: %v", err)
	}
	if save.AccountID != acct.AccountID {
		t.Errorf("stamped save AccountID = %q, want %q", save.AccountID, acct.AccountID)
	}

	// The signature must still verify — it covers account_id because the stamp went
	// through SaveGame's sign step, not an after-the-fact JSON splice.
	sigValid, _ := verifySave(&save)
	if !sigValid {
		t.Error("stamped save failed signature verification — _sig did not cover account_id")
	}
	if save.CheaterBadge {
		t.Error("stamped save carries a CheaterBadge — stale signature?")
	}

	// Full load path: no cheater badge stamped on a legitimately attributed save.
	loaded := NewGameEngine()
	if err := loaded.LoadGame(name); err != nil {
		t.Fatalf("LoadGame(%q): %v", name, err)
	}
	if loaded.GetState().CheaterBadge {
		t.Error("attributed save wrongly flagged CheaterBadge on load")
	}
	// Loading a save must NOT switch the active account (it is set once at boot).
	if loaded.Account() != nil {
		t.Error("LoadGame populated the engine account from the save; load must be informational only")
	}
}

// TestResetPreservesAccount asserts the account is player-level and survives a new
// game / Reset, per accounts.md §6 ("the account is player-level and must survive
// new-game/reset"). Reset reinitializes managers but must leave ge.account intact.
func TestResetPreservesAccount(t *testing.T) {
	isolateAccountDir(t)

	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate: %v", err)
	}
	ge := NewGameEngine()
	ge.SetAccount(acct)

	ge.Reset()

	got := ge.Account()
	if got == nil {
		t.Fatal("Reset cleared the account; it must survive a new game")
	}
	if got.AccountID != acct.AccountID {
		t.Errorf("account changed across Reset: %q != %q", got.AccountID, acct.AccountID)
	}
	if got != acct {
		t.Error("Reset swapped the account pointer; it should be the same instance")
	}
}

// TestFreshlyCreatedFlagOnlyOnFirstRun confirms the first-run signal: LoadOrCreate
// sets FreshlyCreated only when it mints a brand-new account (no file present), and
// a subsequent load of the existing file leaves it false. Boot code keys the
// one-time welcome notice off this (accounts.md §6).
func TestFreshlyCreatedFlagOnlyOnFirstRun(t *testing.T) {
	isolateAccountDir(t)

	first, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (first): %v", err)
	}
	if !first.FreshlyCreated {
		t.Error("first LoadOrCreate did not set FreshlyCreated on a genuine first run")
	}

	second, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (second): %v", err)
	}
	if second.FreshlyCreated {
		t.Error("second LoadOrCreate set FreshlyCreated on an existing-file load")
	}
}
