package game

import (
	"os"
	"path/filepath"
	"testing"
)

// Phase A — account-scoping tests: the active-account pointer, and where an account and
// its saves land. They use the in-package isolateAccountDir seam (account_test.go), which
// points the ROOT at a temp dir and resets the process-global active id between tests.

// TestActivePointerRoundTrip covers the pointer I/O contract directly: write an id, read it
// back; a missing pointer reads as ("", nil); and the write is atomic (no leftover .tmp).
func TestActivePointerRoundTrip(t *testing.T) {
	root := isolateAccountDir(t)

	// Missing pointer reads empty, no error.
	if got, err := readActivePointer(); err != nil || got != "" {
		t.Fatalf("readActivePointer (absent) = (%q, %v), want (\"\", nil)", got, err)
	}

	const id = "fedcba9876543210fedcba9876543210"
	if err := writeActivePointer(id); err != nil {
		t.Fatalf("writeActivePointer: %v", err)
	}
	if got, err := readActivePointer(); err != nil {
		t.Fatalf("readActivePointer (present): %v", err)
	} else if got != id {
		t.Errorf("pointer round-trip = %q, want %q", got, id)
	}

	// Atomic write must leave no temp file behind.
	if _, err := os.Stat(activeAccountPointerPath() + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("writeActivePointer left a .tmp file (stat err = %v)", err)
	}
	// And the pointer lives at the ROOT, not in a slot.
	if _, err := os.Stat(filepath.Join(root, activePointerFileName)); err != nil {
		t.Errorf("active pointer not at root: %v", err)
	}
}

// TestCreateAccountScopesAccountAndSaves confirms the primary production write path
// lands BOTH the account.json and a subsequently-written save inside the account's slot
// (<root>/accounts/<id>/...), not at the shared root — the core "saves are scoped to the
// active account" guarantee of Phase A.
func TestCreateAccountScopesAccountAndSaves(t *testing.T) {
	root := isolateAccountDir(t)

	acct, err := CreateAccount("Scoped Empire")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	id := acct.AccountID

	// account.json must be in the slot, the pointer must name it, and nothing flat at root.
	if _, err := os.Stat(filepath.Join(root, accountsDirName, id, accountFileName)); err != nil {
		t.Errorf("account.json not in slot after CreateAccount: %v", err)
	}
	if got, err := readActivePointer(); err != nil || got != id {
		t.Errorf("active pointer = (%q, %v), want (%q, nil)", got, err, id)
	}
	if _, err := os.Stat(filepath.Join(root, accountFileName)); !os.IsNotExist(err) {
		t.Errorf("CreateAccount wrote a flat root account.json (stat err = %v)", err)
	}

	// A save written now (active account = the named one) must land in the slot's saves dir.
	ge := NewGameEngine()
	ge.SetAccount(acct)
	if err := ge.SaveGame("scoped_run"); err != nil {
		t.Fatalf("SaveGame: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(savePath("scoped_run")) })

	if _, err := os.Stat(filepath.Join(root, accountsDirName, id, "saves", "scoped_run.json")); err != nil {
		t.Errorf("save not written into the account slot's saves dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "saves", "scoped_run.json")); !os.IsNotExist(err) {
		t.Errorf("save leaked into the shared root saves dir (stat err = %v)", err)
	}
}
