package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Phase A — account-scoping migration tests. These cover the non-destructive relocation of
// the legacy FLAT layout (<root>/account.json + <root>/saves/) into the account-scoped slot
// (<root>/accounts/<id>/...), the active-account pointer, and idempotency. They use the
// in-package isolateAccountDir seam (account_test.go), which points the ROOT at a temp dir
// and resets the process-global active id between tests.

// seedLegacyFlatAccount writes a signed legacy account.json directly under the ROOT (the
// pre-Phase-A flat location) and returns its account ID. It mints the account in-memory,
// signs it via the production signer (so verifyAccount passes post-migration), and writes
// it WITHOUT going through Save() — Save() would resolve the scoped path, but the whole
// point of the fixture is the flat root location. The active id is intentionally left
// unset, exactly as it is on a cold boot before migration runs.
func seedLegacyFlatAccount(t *testing.T, root string) string {
	t.Helper()
	now := time.Now()
	acct := &Account{
		Version:     accountSchemaVersion,
		AccountID:   "0123456789abcdef0123456789abcdef", // fixed 32-hex id for deterministic slot
		DisplayName: "Legacy Empire",
		Created:     now,
		LastSeen:    now,
	}
	acct.Signature = signAccount(acct)
	data, err := json.MarshalIndent(acct, "", "  ")
	if err != nil {
		t.Fatalf("marshal legacy account: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, accountFileName), data, 0644); err != nil {
		t.Fatalf("write legacy account.json: %v", err)
	}
	return acct.AccountID
}

// seedLegacySave writes a minimal raw (unsigned) save into the legacy flat saves dir
// (<root>/saves/) so migration has something to move. ListSaveDetails reads JSON only, so
// an unsigned hand-built save is a valid fixture.
func seedLegacySave(t *testing.T, root, name string) {
	t.Helper()
	dir := filepath.Join(root, "saves")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir legacy saves: %v", err)
	}
	save := GameSave{Timestamp: time.Now(), Age: "stone_age", Tick: 42}
	data, err := json.Marshal(save)
	if err != nil {
		t.Fatalf("marshal legacy save: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".json"), data, 0644); err != nil {
		t.Fatalf("write legacy save: %v", err)
	}
}

// TestMigrateLegacyMovesAccountAndSaves covers the headline Phase-A promise: a legacy flat
// account.json + saves/ are relocated into the account slot, the active pointer is written,
// the account loads with the SAME id, and a pre-existing save survives the move.
func TestMigrateLegacyMovesAccountAndSaves(t *testing.T) {
	root := isolateAccountDir(t)

	id := seedLegacyFlatAccount(t, root)
	seedLegacySave(t, root, "old_run")

	// LoadOrCreate runs migrateLegacyAccountIfNeeded at the top, then resolves the pointer.
	acct, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate (post-seed): %v", err)
	}
	if acct.AccountID != id {
		t.Errorf("loaded account id = %q, want migrated id %q", acct.AccountID, id)
	}
	if acct.DisplayName != "Legacy Empire" {
		t.Errorf("DisplayName = %q, want Legacy Empire (DATA preserved across migration)", acct.DisplayName)
	}
	if acct.Tampered {
		t.Error("migrated account flagged Tampered (signature should survive the move)")
	}

	// The account.json now lives in the slot, and NOT at the root.
	slotAccount := filepath.Join(root, accountsDirName, id, accountFileName)
	if _, err := os.Stat(slotAccount); err != nil {
		t.Errorf("account.json not present in slot after migration: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, accountFileName)); !os.IsNotExist(err) {
		t.Errorf("legacy account.json still present at root after migration (stat err = %v)", err)
	}

	// The save moved into the slot's saves dir, and NOT at the root saves dir.
	slotSave := filepath.Join(root, accountsDirName, id, "saves", "old_run.json")
	if _, err := os.Stat(slotSave); err != nil {
		t.Errorf("legacy save not migrated into slot saves dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "saves")); !os.IsNotExist(err) {
		t.Errorf("legacy saves/ dir still present at root after migration (stat err = %v)", err)
	}

	// The active pointer points at the migrated id, and the scoped listing finds the save.
	if got, err := readActivePointer(); err != nil {
		t.Fatalf("readActivePointer: %v", err)
	} else if got != id {
		t.Errorf("active pointer = %q, want migrated id %q", got, id)
	}
	if !SaveExists("old_run") {
		t.Error("SaveExists(old_run) = false after migration; scoped save not discoverable")
	}
}

// findPreMigrationSnapshot returns the single pre-migration-* dir under <root>/backups/,
// failing the test if zero or more than one exists. The dir name carries a runtime timestamp,
// so the test cannot predict it — it scans by the stable "pre-migration-" prefix instead.
func findPreMigrationSnapshot(t *testing.T, root string) string {
	t.Helper()
	backupsRoot := filepath.Join(root, backupsDirName)
	entries, err := os.ReadDir(backupsRoot)
	if err != nil {
		t.Fatalf("read backups root for pre-migration snapshot: %v", err)
	}
	var found []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "pre-migration-") {
			found = append(found, filepath.Join(backupsRoot, e.Name()))
		}
	}
	if len(found) != 1 {
		t.Fatalf("expected exactly one pre-migration snapshot dir, found %d: %v", len(found), found)
	}
	return found[0]
}

// TestMigrateLegacyTakesPreMigrationSnapshot covers the one-time safety net: the legacy flat
// data (account.json + saves/ + a stray account-export.json) is COPIED into
// <root>/backups/pre-migration-<ts>/ BEFORE the migration relocates anything, AND the migration
// still completes (slot account.json present, flat one gone, active pointer written). The
// snapshot must NOT contain the reserved accounts/ or backups/ trees.
func TestMigrateLegacyTakesPreMigrationSnapshot(t *testing.T) {
	root := isolateAccountDir(t)

	id := seedLegacyFlatAccount(t, root)
	seedLegacySave(t, root, "old_run")

	// A stray top-level file (a user-made export) must be captured by the snapshot too — it is
	// part of the flat data/ that is about to be relocated.
	const exportContents = "stray-export-bytes"
	if err := os.WriteFile(filepath.Join(root, "account-export.json"), []byte(exportContents), 0644); err != nil {
		t.Fatalf("seed stray account-export.json: %v", err)
	}

	if err := migrateLegacyAccountIfNeeded(); err != nil {
		t.Fatalf("migrateLegacyAccountIfNeeded: %v", err)
	}

	// --- The pre-migration snapshot exists and captured the pre-migration state. ---
	snap := findPreMigrationSnapshot(t, root)

	if _, err := os.Stat(filepath.Join(snap, accountFileName)); err != nil {
		t.Errorf("snapshot missing account.json: %v", err)
	}
	if _, err := os.Stat(filepath.Join(snap, "saves", "old_run.json")); err != nil {
		t.Errorf("snapshot missing saves/old_run.json: %v", err)
	}
	gotExport, err := os.ReadFile(filepath.Join(snap, "account-export.json"))
	if err != nil {
		t.Errorf("snapshot missing stray account-export.json: %v", err)
	} else if string(gotExport) != exportContents {
		t.Errorf("snapshot account-export.json = %q, want %q", gotExport, exportContents)
	}
	// The snapshot must not have recursed into backups/ or pulled in the (absent) accounts/.
	if _, err := os.Stat(filepath.Join(snap, backupsDirName)); !os.IsNotExist(err) {
		t.Errorf("snapshot recursed into backups/ (stat err = %v)", err)
	}
	if _, err := os.Stat(filepath.Join(snap, accountsDirName)); !os.IsNotExist(err) {
		t.Errorf("snapshot captured an accounts/ dir (stat err = %v)", err)
	}

	// --- The migration still completed despite (alongside) the snapshot. ---
	if _, err := os.Stat(filepath.Join(root, accountsDirName, id, accountFileName)); err != nil {
		t.Errorf("account.json not in slot after migration: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, accountFileName)); !os.IsNotExist(err) {
		t.Errorf("flat account.json still present after migration (stat err = %v)", err)
	}
	if got, err := readActivePointer(); err != nil {
		t.Fatalf("readActivePointer: %v", err)
	} else if got != id {
		t.Errorf("active pointer = %q, want migrated id %q", got, id)
	}
}

// TestMigrateLegacyIsIdempotent covers the trigger guard: once migrated (accounts/ exists),
// a second migration call is a clean no-op that does not error and does not disturb the slot.
func TestMigrateLegacyIsIdempotent(t *testing.T) {
	root := isolateAccountDir(t)

	id := seedLegacyFlatAccount(t, root)
	seedLegacySave(t, root, "old_run")

	// First migration via the function directly.
	if err := migrateLegacyAccountIfNeeded(); err != nil {
		t.Fatalf("migrateLegacyAccountIfNeeded (first): %v", err)
	}
	slotAccount := filepath.Join(root, accountsDirName, id, accountFileName)
	firstInfo, err := os.Stat(slotAccount)
	if err != nil {
		t.Fatalf("slot account.json missing after first migration: %v", err)
	}

	// Second call must be a no-op: no error, slot file untouched (same mod time), and no
	// spurious recreation of the flat root files.
	if err := migrateLegacyAccountIfNeeded(); err != nil {
		t.Fatalf("migrateLegacyAccountIfNeeded (second, idempotent): %v", err)
	}
	secondInfo, err := os.Stat(slotAccount)
	if err != nil {
		t.Fatalf("slot account.json missing after second migration: %v", err)
	}
	if !firstInfo.ModTime().Equal(secondInfo.ModTime()) {
		t.Errorf("idempotent migration rewrote the slot account.json (mod time changed %v -> %v)",
			firstInfo.ModTime(), secondInfo.ModTime())
	}
	if _, err := os.Stat(filepath.Join(root, accountFileName)); !os.IsNotExist(err) {
		t.Error("idempotent migration recreated the flat root account.json")
	}
}

// TestMigrateNoLegacyIsNoOp confirms migration does nothing when there is no flat
// account.json: no slot, no pointer, no accounts/ dir is created.
func TestMigrateNoLegacyIsNoOp(t *testing.T) {
	root := isolateAccountDir(t)

	if err := migrateLegacyAccountIfNeeded(); err != nil {
		t.Fatalf("migrateLegacyAccountIfNeeded (no legacy): %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, accountsDirName)); !os.IsNotExist(err) {
		t.Error("migration created accounts/ dir with no legacy account to migrate")
	}
	if _, err := os.Stat(activeAccountPointerPath()); !os.IsNotExist(err) {
		t.Error("migration wrote an active pointer with no legacy account to migrate")
	}
}

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

// TestCreateNamedAccountScopesAccountAndSaves confirms the primary production write path
// lands BOTH the account.json and a subsequently-written save inside the account's slot
// (<root>/accounts/<id>/...), not at the shared root — the core "saves are scoped to the
// active account" guarantee of Phase A.
func TestCreateNamedAccountScopesAccountAndSaves(t *testing.T) {
	root := isolateAccountDir(t)

	acct, err := CreateNamedAccount("Scoped Empire")
	if err != nil {
		t.Fatalf("CreateNamedAccount: %v", err)
	}
	id := acct.AccountID

	// account.json must be in the slot, the pointer must name it, and nothing flat at root.
	if _, err := os.Stat(filepath.Join(root, accountsDirName, id, accountFileName)); err != nil {
		t.Errorf("account.json not in slot after CreateNamedAccount: %v", err)
	}
	if got, err := readActivePointer(); err != nil || got != id {
		t.Errorf("active pointer = (%q, %v), want (%q, nil)", got, err, id)
	}
	if _, err := os.Stat(filepath.Join(root, accountFileName)); !os.IsNotExist(err) {
		t.Errorf("CreateNamedAccount wrote a flat root account.json (stat err = %v)", err)
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
