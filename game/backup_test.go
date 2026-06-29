package game

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The backup mechanism takes a FULL on-disk snapshot of an account's slot (account.json +
// saves/) into <root>/backups/<name>-<id8>-<timestamp>/. The load-bearing invariants are
// (a) the snapshot copies account.json AND the seeded saves with matching contents, (b) it
// returns the dir path and NEVER changes the active account, (c) a missing slot errors, and
// (d) pruning keeps only the newest N per account without touching other accounts' backups.
// Every case isolates the data ROOT via isolateAccountDir(t), which also resets the
// process-global active id between cases.

// seedSave writes a save file into the slot's saves/ subtree so BackupAccount has something to
// copy recursively. It returns the relative path under saves/ for later assertion.
func seedSave(t *testing.T, id, rel, contents string) string {
	t.Helper()
	dst := filepath.Join(accountDir(id), "saves", rel)
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		t.Fatalf("mkdir saves: %v", err)
	}
	if err := os.WriteFile(dst, []byte(contents), 0644); err != nil {
		t.Fatalf("write seed save: %v", err)
	}
	return rel
}

// TestBackupAccountCopiesAccountAndSaves covers the headline contract: a backup of a slot
// reproduces account.json + a seeded save with matching contents, returns the dir path, and
// leaves the active account untouched.
func TestBackupAccountCopiesAccountAndSaves(t *testing.T) {
	root := isolateAccountDir(t)

	acct, err := CreateAccount("Ada Lovelace")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	id := acct.AccountID
	seedSave(t, id, filepath.Join("run1", "slot.json"), "save-bytes-12345")

	activeBefore := getActiveAccountID()

	dir, err := BackupAccount(id)
	if err != nil {
		t.Fatalf("BackupAccount: %v", err)
	}

	// Returned path is absolute-ish and lives under <root>/backups/.
	if !strings.HasPrefix(dir, filepath.Join(root, backupsDirName)) {
		t.Fatalf("backup dir %q not under %q", dir, filepath.Join(root, backupsDirName))
	}
	// Dir name carries the id8 infix and the slugged display name.
	base := filepath.Base(dir)
	if !strings.Contains(base, "-"+id[:8]+"-") {
		t.Fatalf("backup dir name %q missing id8 infix -%s-", base, id[:8])
	}
	if !strings.HasPrefix(base, "Ada-Lovelace-") {
		t.Fatalf("backup dir name %q missing slugged name prefix", base)
	}

	// account.json copied with matching contents.
	srcAcct, _ := os.ReadFile(filepath.Join(accountDir(id), accountFileName))
	dstAcct, err := os.ReadFile(filepath.Join(dir, accountFileName))
	if err != nil {
		t.Fatalf("backup account.json missing: %v", err)
	}
	if string(srcAcct) != string(dstAcct) {
		t.Fatalf("backup account.json contents differ from source")
	}

	// Seeded save copied with matching contents at the same relative path.
	dstSave, err := os.ReadFile(filepath.Join(dir, "saves", "run1", "slot.json"))
	if err != nil {
		t.Fatalf("backup save missing: %v", err)
	}
	if string(dstSave) != "save-bytes-12345" {
		t.Fatalf("backup save contents = %q, want %q", dstSave, "save-bytes-12345")
	}

	// Active account must NOT have changed (backup only reads the slot).
	if got := getActiveAccountID(); got != activeBefore {
		t.Fatalf("active changed by backup: got %s, want %s", got, activeBefore)
	}
}

// TestBackupAccountMissingSlotErrors covers the absent-slot guard: backing up an id with no
// account.json errors rather than creating an empty backup dir.
func TestBackupAccountMissingSlotErrors(t *testing.T) {
	isolateAccountDir(t)
	if _, err := BackupAccount("deadbeefdeadbeefdeadbeefdeadbeef"); err == nil {
		t.Fatal("BackupAccount on a missing slot should error")
	}
}

// TestBackupAccountNonActiveSlotKeepsActive covers the no-mutate invariant for a NON-active
// slot: backing up A while B is active must not repoint the active account at A.
func TestBackupAccountNonActiveSlotKeepsActive(t *testing.T) {
	isolateAccountDir(t)

	a, err := CreateAccount("Account A")
	if err != nil {
		t.Fatalf("CreateAccount A: %v", err)
	}
	aID := a.AccountID
	b, err := CreateAccount("Account B") // B is now active
	if err != nil {
		t.Fatalf("CreateAccount B: %v", err)
	}
	if getActiveAccountID() != b.AccountID {
		t.Fatalf("precondition: active should be B, got %s", getActiveAccountID())
	}

	if _, err := BackupAccount(aID); err != nil {
		t.Fatalf("BackupAccount(A): %v", err)
	}
	if got := getActiveAccountID(); got != b.AccountID {
		t.Fatalf("active changed by backing up a non-active slot: got %s, want %s (B)", got, b.AccountID)
	}
}

// TestBackupAccountPrunesToRetention covers pruning: with more than backupRetention backups for
// one account, only the newest backupRetention survive — and a DIFFERENT account's backups are
// never touched. We seed older dirs directly (deterministic, no sleeping), then a real
// BackupAccount call triggers the prune.
func TestBackupAccountPrunesToRetention(t *testing.T) {
	root := isolateAccountDir(t)

	acct, err := CreateAccount("Pruner")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	id := acct.AccountID
	id8 := id[:8]

	backupsRoot := filepath.Join(root, backupsDirName)
	if err := os.MkdirAll(backupsRoot, 0755); err != nil {
		t.Fatalf("mkdir backups root: %v", err)
	}

	// Seed 11 OLDER backups for this account with strictly increasing, sortable timestamps so
	// the newest are unambiguous. Names mirror production: <name>-<id8>-<timestamp>.
	for i := 1; i <= 11; i++ {
		// e.g. 20200101-000001 .. 20200101-000011 — all sort before "today" stamps.
		name := "Pruner-" + id8 + "-20200101-0000" + pad2(i)
		if err := os.MkdirAll(filepath.Join(backupsRoot, name), 0755); err != nil {
			t.Fatalf("seed backup %d: %v", i, err)
		}
	}

	// Seed a DIFFERENT account's backups that must survive untouched (distinct id8 infix).
	otherID8 := "ffffffff"
	otherNames := []string{
		"Other-" + otherID8 + "-20200101-000001",
		"Other-" + otherID8 + "-20200101-000002",
	}
	for _, name := range otherNames {
		if err := os.MkdirAll(filepath.Join(backupsRoot, name), 0755); err != nil {
			t.Fatalf("seed other backup: %v", err)
		}
	}

	// A real backup → 12th for this account → triggers the prune down to backupRetention (10).
	newDir, err := BackupAccount(id)
	if err != nil {
		t.Fatalf("BackupAccount: %v", err)
	}

	// Count surviving backups for THIS account.
	entries, err := os.ReadDir(backupsRoot)
	if err != nil {
		t.Fatalf("read backups root: %v", err)
	}
	mineInfix := "-" + id8 + "-"
	var mine, others int
	for _, e := range entries {
		switch {
		case strings.Contains(e.Name(), mineInfix):
			mine++
		case strings.Contains(e.Name(), "-"+otherID8+"-"):
			others++
		}
	}
	if mine != backupRetention {
		t.Fatalf("after prune this account has %d backups, want %d", mine, backupRetention)
	}
	// The just-created backup (newest) must be one of the survivors.
	if _, statErr := os.Stat(newDir); statErr != nil {
		t.Fatalf("newest backup was pruned: %v", statErr)
	}
	// The OLDEST seeded backup must be gone (12 total, keep 10 → 2 oldest pruned).
	if _, statErr := os.Stat(filepath.Join(backupsRoot, "Pruner-"+id8+"-20200101-000001")); !os.IsNotExist(statErr) {
		t.Fatalf("oldest backup should have been pruned, stat err = %v", statErr)
	}
	// The other account's backups are untouched.
	if others != len(otherNames) {
		t.Fatalf("other account's backups changed: have %d, want %d", others, len(otherNames))
	}
}

// TestWipeAccountByIDBacksUpBeforeRemoval covers the wipe-time snapshot: after wiping, the slot
// is gone but a backup dir carrying that id8 exists and contains the account.json + saves.
func TestWipeAccountByIDBacksUpBeforeRemoval(t *testing.T) {
	root := isolateAccountDir(t)

	acct, err := CreateAccount("Doomed")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	id := acct.AccountID
	seedSave(t, id, "autosave.json", "doomed-save-contents")

	backupPath, err := WipeAccountByID(id)
	if err != nil {
		t.Fatalf("WipeAccountByID: %v", err)
	}

	// The slot itself is gone.
	if _, statErr := os.Stat(accountDir(id)); !os.IsNotExist(statErr) {
		t.Fatalf("slot should be removed after wipe, stat err = %v", statErr)
	}

	// A backup path was returned and exists.
	if backupPath == "" {
		t.Fatal("WipeAccountByID returned empty backup path — expected a pre-wipe snapshot")
	}
	if _, statErr := os.Stat(backupPath); statErr != nil {
		t.Fatalf("returned backup dir missing: %v", statErr)
	}
	// It lives under <root>/backups/ and carries the id8 infix.
	if !strings.HasPrefix(backupPath, filepath.Join(root, backupsDirName)) {
		t.Fatalf("backup %q not under %q", backupPath, filepath.Join(root, backupsDirName))
	}
	if !strings.Contains(filepath.Base(backupPath), "-"+id[:8]+"-") {
		t.Fatalf("backup dir name %q missing id8 infix", filepath.Base(backupPath))
	}
	// The snapshot contains account.json + the seeded save.
	if _, statErr := os.Stat(filepath.Join(backupPath, accountFileName)); statErr != nil {
		t.Fatalf("backup missing account.json: %v", statErr)
	}
	save, statErr := os.ReadFile(filepath.Join(backupPath, "saves", "autosave.json"))
	if statErr != nil {
		t.Fatalf("backup missing seeded save: %v", statErr)
	}
	if string(save) != "doomed-save-contents" {
		t.Fatalf("backup save contents = %q, want %q", save, "doomed-save-contents")
	}
}

// pad2 zero-pads a 1..99 int to two digits for the deterministic timestamp seeds above.
func pad2(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
