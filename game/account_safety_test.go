package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

// Reproductions for the account data-safety bugs (achievements audit B1, B4 and B5).
// Every test runs in a temp data root (isolateAccountDir), never the repo's data/.

// slotAccount reads the account in slot id straight from disk, failing the test if the
// slot holds none.
func slotAccount(t *testing.T, id string) *Account {
	t.Helper()
	acct, found, err := loadAccountFromSlot(id)
	if err != nil || !found || acct == nil {
		t.Fatalf("slot %s holds no account (found=%v, err=%v)", id, found, err)
	}
	return acct
}

// slotFile returns the raw bytes of the account.json in slot id.
func slotFile(t *testing.T, id string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(accountDir(id), accountFileName))
	if err != nil {
		t.Fatalf("read slot %s: %v", id, err)
	}
	return b
}

// TestRecoverKeepsActiveAccountIntact is B1: recovering another account's code while an
// account with progress was active overwrote the active account's file with an empty
// account carrying the recovered ID, losing its achievements and stats. The active
// account must stay byte for byte as it was, and the recovered identity must land in its
// own slot.
func TestRecoverKeepsActiveAccountIntact(t *testing.T) {
	isolateAccountDir(t)

	bob, err := CreateAccount("Bob")
	if err != nil {
		t.Fatalf("CreateAccount Bob: %v", err)
	}
	bobCode := bob.RecoveryCode()
	// Bob's slot goes, as on a machine that never had it: the code is all that is left.
	if _, err := WipeAccountByID(bob.AccountID); err != nil {
		t.Fatalf("wipe Bob: %v", err)
	}

	alice, err := CreateAccount("Alice")
	if err != nil {
		t.Fatalf("CreateAccount Alice: %v", err)
	}
	alice.RecordPrestige() // first_prestige and one lifetime prestige, no themes
	if err := alice.FlushIfDirty(); err != nil {
		t.Fatalf("flush Alice: %v", err)
	}
	before := slotFile(t, alice.AccountID)

	restored, err := ImportRecoveryCode(bobCode)
	if err != nil {
		t.Fatalf("ImportRecoveryCode: %v", err)
	}
	if restored.AccountID != bob.AccountID {
		t.Fatalf("restored ID %s, want Bob's %s", restored.AccountID, bob.AccountID)
	}

	if after := slotFile(t, alice.AccountID); string(after) != string(before) {
		t.Errorf("recovering Bob's code rewrote Alice's account file:\nbefore %s\nafter  %s", before, after)
	}
	got := slotAccount(t, alice.AccountID)
	if got.AccountID != alice.AccountID || got.DisplayName != "Alice" ||
		got.Stats.TotalPrestiges != 1 || len(got.Achievements) != 1 || got.Achievements[0] != "first_prestige" {
		t.Errorf("Alice's slot after the recovery: id=%s name=%q stats=%+v achievements=%v",
			got.AccountID, got.DisplayName, got.Stats, got.Achievements)
	}
	bobNow := slotAccount(t, bob.AccountID)
	if bobNow.AccountID != bob.AccountID || !verifyAccount(bobNow) || bobNow.Tampered {
		t.Errorf("Bob's restored slot: id=%s verified=%v tampered=%v", bobNow.AccountID, verifyAccount(bobNow), bobNow.Tampered)
	}
	if len(bobNow.Achievements) != 0 || bobNow.Stats.TotalPrestiges != 0 || len(bobNow.Unlocks.Themes) != 0 {
		t.Errorf("a recovery code restores identity only, but Bob's slot holds data: %+v", bobNow)
	}
}

// TestRecoverOwnCodeKeepsProgress is the self-inflicted form of B1: recovering the code
// of an account already on this machine (your own) reset it to an empty account. It must
// open the account as it is.
func TestRecoverOwnCodeKeepsProgress(t *testing.T) {
	isolateAccountDir(t)

	alice, err := CreateAccount("Alice")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if _, err := alice.UnlockTheme("amber_crt"); err != nil {
		t.Fatalf("UnlockTheme: %v", err)
	}
	alice.RecordPrestige()
	if err := alice.FlushIfDirty(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	before := slotFile(t, alice.AccountID)

	restored, err := ImportRecoveryCode(alice.RecoveryCode())
	if err != nil {
		t.Fatalf("ImportRecoveryCode: %v", err)
	}
	if !restored.HasTheme("amber_crt") || restored.Stats.TotalPrestiges != 1 || restored.DisplayName != "Alice" {
		t.Errorf("recovering your own code returned a reset account: name=%q themes=%v stats=%+v",
			restored.DisplayName, restored.Unlocks.Themes, restored.Stats)
	}
	if after := slotFile(t, alice.AccountID); string(after) != string(before) {
		t.Errorf("recovering your own code rewrote your account file")
	}
}

// TestRecoverSetsAsideUnreadableSlot: when the code's slot holds an account.json the game
// cannot read, recovery moves it aside rather than overwriting it.
func TestRecoverSetsAsideUnreadableSlot(t *testing.T) {
	isolateAccountDir(t)

	bob, err := CreateAccount("Bob")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	path := filepath.Join(accountDir(bob.AccountID), accountFileName)
	if err := os.WriteFile(path, []byte("not json {"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportRecoveryCode(bob.RecoveryCode()); err != nil {
		t.Fatalf("ImportRecoveryCode: %v", err)
	}
	if b, err := os.ReadFile(path + ".corrupt"); err != nil || string(b) != "not json {" {
		t.Errorf("the unreadable account.json was not kept as .corrupt (err %v, got %q)", err, b)
	}
	if got := slotAccount(t, bob.AccountID); got.AccountID != bob.AccountID {
		t.Errorf("restored slot holds %s", got.AccountID)
	}
}

// TestStaleAccountSaveStaysInItsOwnSlot is B4's root cause: Account.Save wrote to
// whichever slot was active, so an account object saved after a switch (a late autosave
// flush, a theme change on an old reference) wrote its data into the other account's file.
func TestStaleAccountSaveStaysInItsOwnSlot(t *testing.T) {
	isolateAccountDir(t)

	carol, err := CreateAccount("Carol")
	if err != nil {
		t.Fatalf("CreateAccount Carol: %v", err)
	}
	dave, err := CreateAccount("Dave") // Dave is active now; carol is a stale reference
	if err != nil {
		t.Fatalf("CreateAccount Dave: %v", err)
	}
	daveBefore := slotFile(t, dave.AccountID)

	carol.RecordPrestige()
	if err := carol.FlushIfDirty(); err != nil {
		t.Fatalf("flush Carol: %v", err)
	}
	if err := carol.SetActiveTheme("amber_crt"); err != nil {
		t.Fatalf("SetActiveTheme Carol: %v", err)
	}

	if after := slotFile(t, dave.AccountID); string(after) != string(daveBefore) {
		t.Errorf("Carol's saves wrote into Dave's slot:\n%s", after)
	}
	c := slotAccount(t, carol.AccountID)
	if c.Stats.TotalPrestiges != 1 || c.Prefs.ActiveTheme != "amber_crt" {
		t.Errorf("Carol's saves did not reach her own slot: stats=%+v prefs=%+v", c.Stats, c.Prefs)
	}
	if getActiveAccountID() != dave.AccountID {
		t.Errorf("saving Carol changed the active account to %s", getActiveAccountID())
	}
}

// TestImportNeverMovesActiveAccount is B4: ImportAccountExport pointed the process-wide
// active account at the imported slot for the length of its write, so an autosave on the
// tick goroutine in that window wrote the live game and account into the imported slot.
// The active account must not move at any moment of an import.
func TestImportNeverMovesActiveAccount(t *testing.T) {
	isolateAccountDir(t)

	bea, err := CreateAccount("Bea")
	if err != nil {
		t.Fatalf("CreateAccount Bea: %v", err)
	}
	blob, err := bea.ExportProgress()
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	abe, err := CreateAccount("Abe") // active
	if err != nil {
		t.Fatalf("CreateAccount Abe: %v", err)
	}

	var moved atomic.Bool
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
				if getActiveAccountID() != abe.AccountID {
					moved.Store(true)
				}
			}
		}
	}()
	for i := 0; i < 50; i++ {
		if _, err := ImportAccountExport(blob, i%2 == 0); err != nil {
			t.Fatalf("import %d: %v", i, err)
		}
	}
	close(stop)
	<-done
	if moved.Load() {
		t.Error("an import pointed the active account away from Abe while it wrote")
	}
}

// TestImportRejectsAccountIDThatIsAPath: the ID inside a backup names the slot it lands
// in, so a signed backup whose ID is a path must be refused before anything is written.
func TestImportRejectsAccountIDThatIsAPath(t *testing.T) {
	root := isolateAccountDir(t)

	exp := progressExport{Version: accountSchemaVersion, AccountID: "../../escaped", DisplayName: "Mallory"}
	exp.Signature = signProgressExport(&exp)
	blob, err := json.Marshal(&exp)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ImportAccountExport(blob, true); err == nil {
		t.Error("a backup whose account ID is a path was imported")
	}
	if _, err := os.Stat(filepath.Join(root, "..", "escaped")); !os.IsNotExist(err) {
		t.Errorf("the import wrote outside the data folder (stat err = %v)", err)
	}
}

// TestTamperFlagSurvivesResave is B5: a hand-edited account.json loaded flagged, but the
// next save (a theme change, an unlock, the autosave flush) re-signed the file without the
// flag, laundering the edit. Once set, the flag must survive every re-sign, come back if
// the key is deleted by hand, and ride along in an export.
func TestTamperFlagSurvivesResave(t *testing.T) {
	isolateAccountDir(t)

	acct, err := CreateAccount("Mallory")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	id := acct.AccountID
	path := filepath.Join(accountDir(id), accountFileName)
	editFile := func(edit func(map[string]any)) {
		t.Helper()
		var raw map[string]any
		if err := json.Unmarshal(slotFile(t, id), &raw); err != nil {
			t.Fatal(err)
		}
		edit(raw)
		b, err := json.MarshalIndent(raw, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, b, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// A fake achievement, added by hand without re-signing.
	editFile(func(raw map[string]any) { raw["achievements"] = []string{"prestige_x10"} })
	loaded := slotAccount(t, id)
	if !loaded.Tampered {
		t.Fatal("precondition: a hand-edited account.json did not load flagged")
	}

	// Any save re-signs the file. The flag must be in what it signs.
	if err := loaded.SetActiveTheme("amber_crt"); err != nil {
		t.Fatalf("SetActiveTheme: %v", err)
	}
	again := slotAccount(t, id)
	if !again.Tampered {
		t.Fatal("the next save cleared the tamper flag (the edit was laundered)")
	}
	if !verifyAccount(again) {
		t.Error("the re-signed file does not verify")
	}
	for _, s := range ListAccounts() {
		if s.AccountID == id && !s.Tampered {
			t.Error("the account list lost the modified marker after a save")
		}
	}

	// Deleting the flag by hand breaks the signature, so it comes straight back.
	editFile(func(raw map[string]any) { delete(raw, "tampered") })
	if !slotAccount(t, id).Tampered {
		t.Error("deleting the tampered key by hand cleared the flag")
	}

	// An export carries it, so wiping and importing does not launder the account either.
	blob, err := again.ExportProgress()
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if _, err := WipeAccountByID(id); err != nil {
		t.Fatalf("wipe: %v", err)
	}
	imported, err := ImportAccountExport(blob, true)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if !imported.Tampered || !slotAccount(t, id).Tampered {
		t.Error("export and import laundered the tamper flag")
	}
}

// Written by the previous version (before the tamper flag was persisted): an account.json
// and its export, byte for byte.
const (
	fixtureAccountID   = "f16d05ec6b29248d2c61adb1e9263f78"
	fixtureAccountJSON = `{
  "version": 1,
  "account_id": "f16d05ec6b29248d2c61adb1e9263f78",
  "display_name": "Fixture",
  "created": "2026-09-01T12:00:00Z",
  "last_seen": "2026-09-01T12:00:00Z",
  "unlocks": {
    "themes": [
      "amber_crt"
    ]
  },
  "stats": {
    "total_prestiges": 2,
    "highest_age": "modern_age"
  },
  "achievements": [
    "first_prestige",
    "reached_iron",
    "reached_modern"
  ],
  "prefs": {
    "active_theme": "amber_crt"
  },
  "_sig": "e0348b8c8c0395284dd2944fe5ef09a3308ab5966ec841a962565cdd4b9c86eb"
}`
	fixtureExportJSON = `{
  "version": 1,
  "account_id": "f16d05ec6b29248d2c61adb1e9263f78",
  "display_name": "Fixture",
  "unlocks": {
    "themes": [
      "amber_crt"
    ]
  },
  "stats": {
    "total_prestiges": 2,
    "highest_age": "modern_age"
  },
  "achievements": [
    "first_prestige",
    "reached_iron",
    "reached_modern"
  ],
  "prefs": {
    "active_theme": "amber_crt"
  },
  "_sig": "4cd3effa2b83aef8cc9be0783da9f83336fd0eb9553ed419dbd851b40e9aa9d4"
}`
)

// TestOlderFilesStillVerify: persisting the tamper flag (omitempty) must not change what
// an untampered account signs, so account files and exports written before it still
// verify, and a clean account saved now carries no tampered key.
func TestOlderFilesStillVerify(t *testing.T) {
	isolateAccountDir(t)

	if err := os.MkdirAll(accountDir(fixtureAccountID), 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(accountDir(fixtureAccountID), accountFileName)
	if err := os.WriteFile(path, []byte(fixtureAccountJSON), 0644); err != nil {
		t.Fatal(err)
	}
	old := slotAccount(t, fixtureAccountID)
	if old.Tampered || !verifyAccount(old) {
		t.Errorf("an account.json from the previous version reads as modified (tampered=%v)", old.Tampered)
	}

	// Its export verifies too (import merges it into the same account).
	if _, err := ImportAccountExport([]byte(fixtureExportJSON), true); err != nil {
		t.Errorf("an export from the previous version was refused: %v", err)
	}

	clean, err := CreateAccount("Clean")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(slotFile(t, clean.AccountID), &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["tampered"]; ok {
		t.Error("a clean account.json carries a tampered key")
	}
}
