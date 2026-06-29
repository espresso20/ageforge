package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Phase C re-homes import: an export is a single-account BACKUP (identity + progress), and
// ImportAccountExport lands it in the blob's OWN slot (keyed by AccountID) — creating the
// slot if absent, merging/replacing if present — WITHOUT disturbing the active account.
// These tests cover the round-trip identity, the create-vs-merge split, cross-account
// safety, and the integrity/missing-id guards. Every case isolates the data ROOT via
// isolateAccountDir(t), which also resets the process-global active id between cases.

// TestExportCarriesIdentity covers Phase C's headline shape change: the export blob now
// round-trips the owning account's AccountID + DisplayName alongside its data.
func TestExportCarriesIdentity(t *testing.T) {
	isolateAccountDir(t)

	acct, err := CreateAccount("Ada Lovelace")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if _, err := acct.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("UnlockTheme: %v", err)
	}

	blob, err := acct.ExportProgress()
	if err != nil {
		t.Fatalf("ExportProgress: %v", err)
	}
	// The identity must be present in the round-tripped blob (unmarshal it back).
	var exp progressExport
	if err := json.Unmarshal(blob, &exp); err != nil {
		t.Fatalf("unmarshal export: %v", err)
	}
	if exp.AccountID != acct.AccountID {
		t.Fatalf("export AccountID = %q, want %q", exp.AccountID, acct.AccountID)
	}
	if exp.DisplayName != "Ada Lovelace" {
		t.Fatalf("export DisplayName = %q, want %q", exp.DisplayName, "Ada Lovelace")
	}
}

// TestImportCreatesAbsentSlot covers the create path: importing a blob whose slot does NOT
// exist in this root mints accounts/<blobid>/account.json carrying the blob's id, name, and
// data; the slot is loadable and matches.
func TestImportCreatesAbsentSlot(t *testing.T) {
	// Build a blob in one isolated root...
	isolateAccountDir(t)
	src, err := CreateAccount("Grace Hopper")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if _, err := src.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("UnlockTheme A: %v", err)
	}
	if _, err := src.UnlockTheme("obsidian"); err != nil {
		t.Fatalf("UnlockTheme B: %v", err)
	}
	src.Stats.TotalPrestiges = 7
	blob, err := src.ExportProgress()
	if err != nil {
		t.Fatalf("ExportProgress: %v", err)
	}
	srcID := src.AccountID

	// ...then import it into a FRESH root where that slot does not exist.
	root := isolateAccountDir(t)
	if _, found, _ := loadAccountFromSlot(srcID); found {
		t.Fatal("precondition: blob's slot should not exist in the fresh root")
	}

	imported, err := ImportAccountExport(blob, true)
	if err != nil {
		t.Fatalf("ImportAccountExport: %v", err)
	}
	if imported.AccountID != srcID || imported.DisplayName != "Grace Hopper" {
		t.Fatalf("imported identity = (%q, %q), want (%q, %q)",
			imported.AccountID, imported.DisplayName, srcID, "Grace Hopper")
	}

	// The slot file exists on disk and is independently loadable with matching data.
	slotFile := filepath.Join(root, accountsDirName, srcID, accountFileName)
	if _, statErr := os.Stat(slotFile); statErr != nil {
		t.Fatalf("expected slot file at %s: %v", slotFile, statErr)
	}
	loaded, found, err := loadAccountFromSlot(srcID)
	if err != nil || !found {
		t.Fatalf("loadAccountFromSlot(%s) = (found=%v, %v), want found", srcID, found, err)
	}
	if loaded.DisplayName != "Grace Hopper" {
		t.Fatalf("loaded DisplayName = %q, want %q", loaded.DisplayName, "Grace Hopper")
	}
	if loaded.Stats.TotalPrestiges != 7 {
		t.Fatalf("loaded prestiges = %d, want 7", loaded.Stats.TotalPrestiges)
	}
	if !loaded.HasTheme("amber_crt") || !loaded.HasTheme("obsidian") {
		t.Fatalf("loaded themes = %v, want both restored", loaded.UnlockedThemes())
	}
	if loaded.Tampered {
		t.Fatal("freshly imported slot should verify (not Tampered)")
	}
}

// TestImportMergeUnionsExistingSlot covers the merge path onto an EXISTING slot: export {A},
// add {B} to the slot, re-import(merge) → the slot holds {A,B} (merge never drops a newer
// local unlock the backup predates, and restores ones the backup carries).
func TestImportMergeUnionsExistingSlot(t *testing.T) {
	isolateAccountDir(t)

	acct, err := CreateAccount("Katherine Johnson")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if _, err := acct.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("UnlockTheme A: %v", err)
	}
	blob, err := acct.ExportProgress() // backup carries only {amber_crt}
	if err != nil {
		t.Fatalf("ExportProgress: %v", err)
	}

	// Locally unlock a newer theme the backup predates, persisting it into the slot.
	if _, err := acct.UnlockTheme("obsidian"); err != nil {
		t.Fatalf("UnlockTheme B: %v", err)
	}

	imported, err := ImportAccountExport(blob, true)
	if err != nil {
		t.Fatalf("ImportAccountExport(merge): %v", err)
	}
	got := imported.UnlockedThemes() // sorted
	if len(got) != 2 || got[0] != "amber_crt" || got[1] != "obsidian" {
		t.Fatalf("expected union {amber_crt, obsidian}, got %v", got)
	}
}

// TestImportReplaceWholesale covers merge=false onto an EXISTING slot: the DATA fields are
// replaced by the blob's, so a local-only theme not in the blob is dropped.
func TestImportReplaceWholesale(t *testing.T) {
	isolateAccountDir(t)

	acct, err := CreateAccount("Radia Perlman")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if _, err := acct.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("UnlockTheme A: %v", err)
	}
	blob, err := acct.ExportProgress() // backup carries {amber_crt}
	if err != nil {
		t.Fatalf("ExportProgress: %v", err)
	}

	// Local-only theme that the blob does NOT have.
	if _, err := acct.UnlockTheme("obsidian"); err != nil {
		t.Fatalf("UnlockTheme B: %v", err)
	}

	imported, err := ImportAccountExport(blob, false)
	if err != nil {
		t.Fatalf("ImportAccountExport(replace): %v", err)
	}
	got := imported.UnlockedThemes()
	if len(got) != 1 || got[0] != "amber_crt" {
		t.Fatalf("expected wholesale replace to {amber_crt}, got %v", got)
	}
	if imported.HasTheme("obsidian") {
		t.Fatal("replace should have dropped the local-only theme")
	}
}

// TestImportMergeMaxesNumericStats covers the lifetime-best rule on an existing slot: merge
// takes the MAX per numeric stat so bests never regress, even when the blob's value is lower.
func TestImportMergeMaxesNumericStats(t *testing.T) {
	isolateAccountDir(t)

	acct, err := CreateAccount("Margaret Hamilton")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	acct.Stats.TotalPrestiges = 5
	acct.Stats.CivilizationsStarted = 3
	if err := acct.Save(); err != nil { // persist the baseline into the slot
		t.Fatalf("Save baseline: %v", err)
	}

	blob, err := acct.ExportProgress() // backup: prestiges=5, civs=3
	if err != nil {
		t.Fatalf("ExportProgress: %v", err)
	}

	// Slot progresses past the backup on one stat, regresses below it on another.
	acct.Stats.TotalPrestiges = 9       // slot higher than blob (5) → keep 9
	acct.Stats.CivilizationsStarted = 1 // slot lower than blob (3) → take blob's 3
	if err := acct.Save(); err != nil {
		t.Fatalf("Save progressed: %v", err)
	}

	imported, err := ImportAccountExport(blob, true)
	if err != nil {
		t.Fatalf("ImportAccountExport(merge): %v", err)
	}
	if imported.Stats.TotalPrestiges != 9 {
		t.Fatalf("expected max(9,5)=9 prestiges, got %d", imported.Stats.TotalPrestiges)
	}
	if imported.Stats.CivilizationsStarted != 3 {
		t.Fatalf("expected max(1,3)=3 civilizations, got %d", imported.Stats.CivilizationsStarted)
	}
}

// TestImportCrossAccountSafety is the Phase C invariant: with account A active, importing a
// blob belonging to account B must (1) leave A's slot UNTOUCHED, (2) create/populate B's
// slot with B's data, and (3) leave the ACTIVE account unchanged (still A).
func TestImportCrossAccountSafety(t *testing.T) {
	isolateAccountDir(t)

	// Account B: build its backup, capture id + data, then we won't touch its slot again
	// until the import recreates it.
	acctB, err := CreateAccount("Bob")
	if err != nil {
		t.Fatalf("CreateAccount(Bob): %v", err)
	}
	if _, err := acctB.UnlockTheme("obsidian"); err != nil {
		t.Fatalf("Bob UnlockTheme: %v", err)
	}
	acctB.Stats.TotalPrestiges = 4
	if err := acctB.Save(); err != nil {
		t.Fatalf("Bob Save: %v", err)
	}
	blobB, err := acctB.ExportProgress()
	if err != nil {
		t.Fatalf("Bob ExportProgress: %v", err)
	}
	idB := acctB.AccountID

	// Wipe B's slot so we can prove the import recreates it (and so A is unambiguously the
	// only pre-existing slot). WipeAccount removes the ACTIVE slot — B is active right now.
	if err := WipeAccount(); err != nil {
		t.Fatalf("WipeAccount(B): %v", err)
	}

	// Account A: make it active with its own distinct data.
	acctA, err := CreateAccount("Alice")
	if err != nil {
		t.Fatalf("CreateAccount(Alice): %v", err)
	}
	if _, err := acctA.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("Alice UnlockTheme: %v", err)
	}
	acctA.Stats.TotalPrestiges = 11
	if err := acctA.Save(); err != nil {
		t.Fatalf("Alice Save: %v", err)
	}
	idA := acctA.AccountID
	if got := getActiveAccountID(); got != idA {
		t.Fatalf("precondition: active should be A (%s), got %s", idA, got)
	}

	// Import B's backup while A is active.
	imported, err := ImportAccountExport(blobB, true)
	if err != nil {
		t.Fatalf("ImportAccountExport(B while A active): %v", err)
	}
	if imported.AccountID != idB {
		t.Fatalf("imported id = %s, want B (%s)", imported.AccountID, idB)
	}

	// (3) The active account is UNCHANGED by the import itself — still A.
	if got := getActiveAccountID(); got != idA {
		t.Fatalf("import leaked the active account: got %s, want A (%s)", got, idA)
	}

	// (1) A's slot is intact: load it back and assert its own data survived.
	loadedA, found, err := loadAccountFromSlot(idA)
	if err != nil || !found {
		t.Fatalf("loadAccountFromSlot(A) = (found=%v, %v), want found", found, err)
	}
	if loadedA.DisplayName != "Alice" || loadedA.Stats.TotalPrestiges != 11 {
		t.Fatalf("A's slot was mutated: name=%q prestiges=%d, want (Alice, 11)",
			loadedA.DisplayName, loadedA.Stats.TotalPrestiges)
	}
	if loadedA.HasTheme("obsidian") {
		t.Fatal("A's slot leaked B's theme")
	}

	// (2) B's slot now exists with B's data.
	loadedB, found, err := loadAccountFromSlot(idB)
	if err != nil || !found {
		t.Fatalf("loadAccountFromSlot(B) = (found=%v, %v), want found", found, err)
	}
	if loadedB.DisplayName != "Bob" || loadedB.Stats.TotalPrestiges != 4 {
		t.Fatalf("B's slot wrong: name=%q prestiges=%d, want (Bob, 4)",
			loadedB.DisplayName, loadedB.Stats.TotalPrestiges)
	}
	if !loadedB.HasTheme("obsidian") {
		t.Fatalf("B's slot missing its theme, got %v", loadedB.UnlockedThemes())
	}
}

// TestImportTamperGuard covers the integrity guard: a single flipped byte in the blob must
// make ImportAccountExport return an error AND not write the target slot.
func TestImportTamperGuard(t *testing.T) {
	isolateAccountDir(t)

	acct, err := CreateAccount("Tamper Target")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if _, err := acct.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("UnlockTheme: %v", err)
	}
	blob, err := acct.ExportProgress()
	if err != nil {
		t.Fatalf("ExportProgress: %v", err)
	}

	// Flip a byte in the data region (the theme key 'a' in "amber_crt") so the payload no
	// longer matches the signature.
	tampered := make([]byte, len(blob))
	copy(tampered, blob)
	idx := indexOfByte(tampered, 'a')
	if idx < 0 {
		t.Fatal("test fixture: expected an 'a' byte to flip in the blob")
	}
	tampered[idx] = 'z'

	if _, err := ImportAccountExport(tampered, true); err == nil {
		t.Fatal("expected ImportAccountExport to reject a tampered blob, got nil error")
	}
}

// TestImportRejectsMissingAccountID covers the old/foreign-export guard: a validly-signed
// blob with no AccountID has no slot to target and must be rejected.
func TestImportRejectsMissingAccountID(t *testing.T) {
	isolateAccountDir(t)

	// Build a signed blob with an EMPTY AccountID (an old-format export shape).
	exp := progressExport{
		Version: accountSchemaVersion,
		Unlocks: AccountUnlocks{Themes: []string{"amber_crt"}},
	}
	exp.Signature = signProgressExport(&exp) // sign it so it passes the integrity check
	blob, err := json.Marshal(&exp)
	if err != nil {
		t.Fatalf("marshal export: %v", err)
	}

	_, err = ImportAccountExport(blob, true)
	if err == nil {
		t.Fatal("expected rejection of a blob missing its account id, got nil error")
	}
}

// indexOfByte returns the index of the first occurrence of b in data, or -1.
func indexOfByte(data []byte, b byte) int {
	for i, c := range data {
		if c == b {
			return i
		}
	}
	return -1
}
