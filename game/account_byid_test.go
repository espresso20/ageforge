package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Phase D adds the by-id account helpers the Accounts panel needs to act on a SELECTED account
// that may not be the live one: ExportAccountByID, RecoveryCodeForID, and WipeAccountByID. The
// load-bearing invariants are (a) reading/exporting a slot must NOT change the active account,
// and (b) wiping a slot must spare every OTHER slot — and clear the active pointer only when the
// wiped slot WAS active. Every case isolates the data ROOT via isolateAccountDir(t), which also
// resets the process-global active id between cases.

// TestExportAccountByIDRoundTripsNonActiveSlot covers the headline export contract: a backup of a
// NON-active slot round-trips its identity + data and leaves the active account untouched.
func TestExportAccountByIDRoundTripsNonActiveSlot(t *testing.T) {
	isolateAccountDir(t)

	// Create account A (becomes active), give it progress, then create B so B is the active one.
	a, err := CreateAccount("Ada Lovelace")
	if err != nil {
		t.Fatalf("CreateAccount A: %v", err)
	}
	if _, err := a.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("UnlockTheme: %v", err)
	}
	a.Stats.TotalPrestiges = 3
	if err := a.Save(); err != nil {
		t.Fatalf("Save A: %v", err)
	}
	aID := a.AccountID

	b, err := CreateAccount("Grace Hopper")
	if err != nil {
		t.Fatalf("CreateAccount B: %v", err)
	}
	if getActiveAccountID() != b.AccountID {
		t.Fatalf("precondition: active should be B (%s), got %s", b.AccountID, getActiveAccountID())
	}

	// Export A (the NON-active slot).
	blob, err := ExportAccountByID(aID)
	if err != nil {
		t.Fatalf("ExportAccountByID(A): %v", err)
	}

	// Active account must NOT have changed (export only reads the slot).
	if got := getActiveAccountID(); got != b.AccountID {
		t.Fatalf("active changed by export: got %s, want %s (B)", got, b.AccountID)
	}

	// The blob round-trips A's identity + data.
	var exp progressExport
	if err := json.Unmarshal(blob, &exp); err != nil {
		t.Fatalf("unmarshal export: %v", err)
	}
	if exp.AccountID != aID {
		t.Fatalf("export AccountID = %q, want %q", exp.AccountID, aID)
	}
	if exp.DisplayName != "Ada Lovelace" {
		t.Fatalf("export DisplayName = %q, want %q", exp.DisplayName, "Ada Lovelace")
	}
	if exp.Stats.TotalPrestiges != 3 {
		t.Fatalf("export prestiges = %d, want 3", exp.Stats.TotalPrestiges)
	}
}

// TestExportAccountByIDMissingSlotErrors covers the absent-slot guard: exporting an id with no
// account.json errors rather than emitting an empty blob.
func TestExportAccountByIDMissingSlotErrors(t *testing.T) {
	isolateAccountDir(t)
	if _, err := ExportAccountByID("deadbeefdeadbeefdeadbeefdeadbeef"); err == nil {
		t.Fatal("ExportAccountByID on a missing slot should error")
	}
}

// TestRecoveryCodeForIDMatchesAccount covers that the by-id recovery code equals the account's
// own RecoveryCode() and that a missing slot errors. It also confirms the lookup leaves the
// active account untouched.
func TestRecoveryCodeForIDMatchesAccount(t *testing.T) {
	isolateAccountDir(t)

	a, err := CreateAccount("Katherine Johnson")
	if err != nil {
		t.Fatalf("CreateAccount A: %v", err)
	}
	aID := a.AccountID
	want := a.RecoveryCode()

	// Switch active to a second account so A is non-active for the lookup.
	b, err := CreateAccount("Margaret Hamilton")
	if err != nil {
		t.Fatalf("CreateAccount B: %v", err)
	}

	got, err := RecoveryCodeForID(aID)
	if err != nil {
		t.Fatalf("RecoveryCodeForID(A): %v", err)
	}
	if got != want {
		t.Fatalf("RecoveryCodeForID = %q, want %q", got, want)
	}
	if active := getActiveAccountID(); active != b.AccountID {
		t.Fatalf("active changed by recovery lookup: got %s, want %s (B)", active, b.AccountID)
	}

	if _, err := RecoveryCodeForID("00000000000000000000000000000000"); err == nil {
		t.Fatal("RecoveryCodeForID on a missing slot should error")
	}
}

// TestWipeAccountByIDRemovesNonActiveSlotOnly covers the spare-others contract: wiping a NON-active
// slot deletes only that slot and leaves the active account + its slot + every other slot intact,
// and does NOT clear the active pointer.
func TestWipeAccountByIDRemovesNonActiveSlotOnly(t *testing.T) {
	root := isolateAccountDir(t)

	a, err := CreateAccount("Account A")
	if err != nil {
		t.Fatalf("CreateAccount A: %v", err)
	}
	aID := a.AccountID

	b, err := CreateAccount("Account B") // B is now active
	if err != nil {
		t.Fatalf("CreateAccount B: %v", err)
	}
	bID := b.AccountID

	if getActiveAccountID() != bID {
		t.Fatalf("precondition: active should be B (%s), got %s", bID, getActiveAccountID())
	}

	// Wipe the NON-active slot A.
	if err := WipeAccountByID(aID); err != nil {
		t.Fatalf("WipeAccountByID(A): %v", err)
	}

	// A's slot is gone.
	if _, statErr := os.Stat(filepath.Join(root, accountsDirName, aID)); !os.IsNotExist(statErr) {
		t.Fatalf("A's slot should be removed, stat err = %v", statErr)
	}
	// B's slot survives and is still loadable.
	if _, found, _ := loadAccountFromSlot(bID); !found {
		t.Fatal("B's slot should survive a wipe of A")
	}
	// Active account + pointer unchanged.
	if active := getActiveAccountID(); active != bID {
		t.Fatalf("active changed by non-active wipe: got %s, want %s (B)", active, bID)
	}
	pointer, _ := readActivePointer()
	if pointer != bID {
		t.Fatalf("active pointer = %q, want %q (B) — non-active wipe must not clear it", pointer, bID)
	}
}

// TestWipeAccountByIDActiveClearsPointer covers wiping the ACTIVE slot: the slot is removed AND the
// active-account pointer + in-memory id are cleared, so the next boot starts clean.
func TestWipeAccountByIDActiveClearsPointer(t *testing.T) {
	root := isolateAccountDir(t)

	keep, err := CreateAccount("Keep Me")
	if err != nil {
		t.Fatalf("CreateAccount keep: %v", err)
	}
	keepID := keep.AccountID

	active, err := CreateAccount("Active One") // now the active slot
	if err != nil {
		t.Fatalf("CreateAccount active: %v", err)
	}
	activeID := active.AccountID

	if err := WipeAccountByID(activeID); err != nil {
		t.Fatalf("WipeAccountByID(active): %v", err)
	}

	// Active slot removed.
	if _, statErr := os.Stat(filepath.Join(root, accountsDirName, activeID)); !os.IsNotExist(statErr) {
		t.Fatalf("active slot should be removed, stat err = %v", statErr)
	}
	// In-memory active id cleared.
	if got := getActiveAccountID(); got != "" {
		t.Fatalf("active id after wiping active slot = %q, want empty", got)
	}
	// On-disk pointer cleared.
	pointer, _ := readActivePointer()
	if pointer != "" {
		t.Fatalf("active pointer after wiping active slot = %q, want empty", pointer)
	}
	// The untouched slot survives.
	if _, found, _ := loadAccountFromSlot(keepID); !found {
		t.Fatal("the non-active 'Keep Me' slot should survive wiping the active slot")
	}
}

// TestWipeAccountByIDEmptyIDIsRefused covers the empty-id guard: it must error (never remove the
// shared accounts root) and must not disturb existing slots.
func TestWipeAccountByIDEmptyIDIsRefused(t *testing.T) {
	root := isolateAccountDir(t)

	a, err := CreateAccount("Survivor")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}

	if err := WipeAccountByID(""); err == nil {
		t.Fatal("WipeAccountByID(\"\") should error, not silently nuke the accounts root")
	}

	// The accounts root and the existing slot must still be there.
	if _, statErr := os.Stat(filepath.Join(root, accountsDirName)); statErr != nil {
		t.Fatalf("accounts root should survive an empty-id wipe: %v", statErr)
	}
	if _, found, _ := loadAccountFromSlot(a.AccountID); !found {
		t.Fatal("existing slot should survive an empty-id wipe")
	}
}

// TestAccountExportPathInsideSlot covers that the default export path for a slot lives INSIDE that
// account's own slot dir (never the active DataDir), so a by-id export of a non-active selection
// doesn't collide with or land beside the wrong account.
func TestAccountExportPathInsideSlot(t *testing.T) {
	root := isolateAccountDir(t)

	id := "abcdef0123456789abcdef0123456789"
	got := AccountExportPath(id)
	wantDir := filepath.Join(root, accountsDirName, id)
	if filepath.Dir(got) != wantDir {
		t.Fatalf("AccountExportPath dir = %q, want %q", filepath.Dir(got), wantDir)
	}
	if filepath.Base(got) != "account-abcdef01-export.json" {
		t.Fatalf("AccountExportPath base = %q, want account-abcdef01-export.json", filepath.Base(got))
	}
}
