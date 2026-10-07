package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Engine-level account switch, import and recovery during play (achievements audit B1
// and B4). Every test runs in a temp data root (isolateAccountDir), never data/.

// saveAccountID returns the account_id stamped in the save file at path.
func saveAccountID(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read save %s: %v", path, err)
	}
	var head struct {
		AccountID string `json:"account_id"`
	}
	if err := json.Unmarshal(b, &head); err != nil {
		t.Fatalf("parse save %s: %v", path, err)
	}
	return head.AccountID
}

// slotSave is the path of save name in account id's slot.
func slotSave(id, name string) string {
	return filepath.Join(accountDir(id), "saves", name+".json")
}

// prestigeNow puts ge in the Modern Age and prestiges, as a player at the prestige gate.
func prestigeNow(t *testing.T, ge *GameEngine) {
	t.Helper()
	ge.mu.Lock()
	ge.age = "modern_age"
	ge.currentEpoch = config.EpochForAge("modern_age")
	ge.mu.Unlock()
	if err := ge.DoPrestige(); err != nil {
		t.Fatalf("DoPrestige: %v", err)
	}
}

// TestEngineSwitchFlushesOldAccount is B4: switching accounts swapped ge.account without
// flushing the old one, so a prestige or age-up since the last autosave was lost.
func TestEngineSwitchFlushesOldAccount(t *testing.T) {
	isolateAccountDir(t)

	dave, err := CreateAccount("Dave")
	if err != nil {
		t.Fatal(err)
	}
	carol, err := CreateAccount("Carol") // active
	if err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	ge.SetAccount(carol)
	recordPrestige(carol, "modern_age") // not flushed yet

	ended, err := ge.SwitchAccount(dave.AccountID)
	if err != nil {
		t.Fatalf("SwitchAccount: %v", err)
	}
	if ended {
		t.Error("no run was live, but the switch says it ended one")
	}
	if got := slotAccount(t, carol.AccountID).Stats.TotalPrestiges; got != 1 {
		t.Errorf("Carol's unflushed prestige was lost in the switch: %d prestiges on disk", got)
	}
	if ge.Account() == nil || ge.Account().AccountID != dave.AccountID {
		t.Errorf("live account after the switch: %v", ge.Account())
	}
	if getActiveAccountID() != dave.AccountID {
		t.Errorf("active account %s, want Dave", getActiveAccountID())
	}
	if ptr, _ := readActivePointer(); ptr != dave.AccountID {
		t.Errorf("active pointer %s, want Dave", ptr)
	}
	if d := slotAccount(t, dave.AccountID); d.Stats.TotalPrestiges != 0 || len(d.Badges) != 0 {
		t.Errorf("Carol's records reached Dave: %+v %v", d.Stats, d.Badges)
	}
}

// TestEngineSwitchToCurrentKeepsLiveObject: "switching" to the account in use must keep the
// live object, not swap it for a copy read from disk that lacks its unflushed records.
func TestEngineSwitchToCurrentKeepsLiveObject(t *testing.T) {
	isolateAccountDir(t)

	carol, err := CreateAccount("Carol")
	if err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	ge.SetAccount(carol)
	recordPrestige(carol, "modern_age")

	if _, err := ge.SwitchAccount(carol.AccountID); err != nil {
		t.Fatalf("SwitchAccount to self: %v", err)
	}
	if ge.Account() != carol {
		t.Error("switching to the account in use replaced the live object")
	}
	if carol.Stats.TotalPrestiges != 1 {
		t.Errorf("live record lost: %d prestiges", carol.Stats.TotalPrestiges)
	}
}

// TestEngineSwitchToMissingAccountChangesNothing: a switch to an empty slot fails before it
// touches the run, the live account or the pointer.
func TestEngineSwitchToMissingAccountChangesNothing(t *testing.T) {
	isolateAccountDir(t)

	carol, err := CreateAccount("Carol")
	if err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	ge.SetAccount(carol)
	if err := ge.StartNewNamedGame("Rome"); err != nil {
		t.Fatal(err)
	}
	startedLoop(t, ge)
	t.Cleanup(ge.Stop)

	ended, err := ge.SwitchAccount("0123456789abcdef0123456789abcdef")
	if err == nil {
		t.Fatal("switching to an account that does not exist succeeded")
	}
	if ended {
		t.Error("a failed switch ended the run")
	}
	ge.mu.RLock()
	running := ge.running
	ge.mu.RUnlock()
	if !running || ge.Account() != carol || getActiveAccountID() != carol.AccountID {
		t.Errorf("a failed switch changed state: running=%v account=%v active=%s", running, ge.Account(), getActiveAccountID())
	}
}

// TestSwitchMidGameNeverCrossWrites is B4 end to end. With a live run, `account switch`
// used to leave the game running under the new account: autosave wrote the run into the
// new account's saves and later prestiges were credited to it. Now the run is saved into
// its own account's slot and stopped, and nothing from it can reach the new account.
func TestSwitchMidGameNeverCrossWrites(t *testing.T) {
	isolateAccountDir(t)

	bob, err := CreateAccount("Bob")
	if err != nil {
		t.Fatal(err)
	}
	alice, err := CreateAccount("Alice") // active
	if err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	ge.SetAccount(alice)
	if err := ge.StartNewNamedGame("Rome"); err != nil {
		t.Fatal(err)
	}
	done := startedLoop(t, ge)
	t.Cleanup(ge.Stop)
	recordPrestige(alice, "modern_age") // earned in the run since the last autosave
	bobBefore := slotFile(t, bob.AccountID)

	ended, err := ge.SwitchAccount(bob.AccountID)
	if err != nil {
		t.Fatalf("SwitchAccount: %v", err)
	}
	if !ended {
		t.Error("the live run was not ended by the switch")
	}
	select {
	case <-done:
	default:
		t.Fatal("the tick loop is still running after the switch")
	}

	// The run is saved in Alice's slot, stamped as hers; Bob's slot has no copy.
	if got := saveAccountID(t, slotSave(alice.AccountID, "Rome")); got != alice.AccountID {
		t.Errorf("Rome is stamped %s, want Alice", got)
	}
	if _, err := os.Stat(slotSave(bob.AccountID, "Rome")); !os.IsNotExist(err) {
		t.Errorf("Rome was written into Bob's saves (stat err = %v)", err)
	}
	if got := slotAccount(t, alice.AccountID).Stats.TotalPrestiges; got != 1 {
		t.Errorf("Alice's record from the run was lost: %d prestiges on disk", got)
	}

	// A late save of the run in memory (the exit handler saves "autosave") still goes to Alice.
	if err := ge.SaveGame("autosave"); err != nil {
		t.Fatalf("late SaveGame: %v", err)
	}
	if _, err := os.Stat(slotSave(bob.AccountID, "autosave")); !os.IsNotExist(err) {
		t.Errorf("a late save of Alice's run landed in Bob's slot (stat err = %v)", err)
	}
	if got := saveAccountID(t, slotSave(alice.AccountID, "autosave")); got != alice.AccountID {
		t.Errorf("the late save is stamped %s, want Alice", got)
	}

	// And the run in memory records nothing to Bob, even if it goes on (headless play).
	if ge.GetState().AccountRecords {
		t.Error("Alice's run still records to the account after switching to Bob")
	}
	prestigeNow(t, ge)
	if err := bob.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	if b := ge.Account(); b.Stats.TotalPrestiges != 0 || len(b.Badges) != 0 {
		t.Errorf("Alice's run credited Bob: %+v %v", b.Stats, b.Badges)
	}
	if after := slotFile(t, bob.AccountID); string(after) != string(bobBefore) {
		t.Errorf("Bob's account file changed:\n%s", after)
	}

	// Loading one of Bob's games makes the run his: it records to him again.
	if err := ge.StartNewNamedGame("Carthage"); err != nil {
		t.Fatal(err)
	}
	if !ge.GetState().AccountRecords {
		t.Error("a new game under Bob does not record to Bob")
	}
	if _, err := os.Stat(slotSave(bob.AccountID, "Carthage")); err != nil {
		t.Errorf("Bob's new game was not saved in his slot: %v", err)
	}
}

// TestImportMidGameNeverCrossWrites is B4 for import: during a live run, importing another
// account's backup must land in that account's slot and leave the live account, the active
// account and the run alone.
func TestImportMidGameNeverCrossWrites(t *testing.T) {
	isolateAccountDir(t)

	bob, err := CreateAccount("Bob")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bob.UnlockTheme("amber_crt"); err != nil {
		t.Fatal(err)
	}
	blob, err := bob.ExportProgress()
	if err != nil {
		t.Fatal(err)
	}
	alice, err := CreateAccount("Alice") // active
	if err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	ge.SetAccount(alice)
	if err := ge.StartNewNamedGame("Rome"); err != nil {
		t.Fatal(err)
	}
	startedLoop(t, ge)
	t.Cleanup(ge.Stop)
	aliceBefore := slotFile(t, alice.AccountID)

	imported, err := ge.ImportAccountExport(blob, false)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if imported.AccountID != bob.AccountID {
		t.Errorf("import landed in %s, want Bob", imported.AccountID)
	}
	ge.mu.RLock()
	running := ge.running
	ge.mu.RUnlock()
	if !running || ge.Account() != alice || getActiveAccountID() != alice.AccountID {
		t.Errorf("the import disturbed the game: running=%v account=%v active=%s", running, ge.Account(), getActiveAccountID())
	}
	if after := slotFile(t, alice.AccountID); string(after) != string(aliceBefore) {
		t.Errorf("importing Bob's backup changed Alice's account file")
	}
	if !slotAccount(t, bob.AccountID).HasTheme("amber_crt") {
		t.Error("Bob's slot does not hold the imported data")
	}
	if _, err := os.Stat(slotSave(bob.AccountID, "Rome")); !os.IsNotExist(err) {
		t.Errorf("Alice's run was saved into Bob's slot (stat err = %v)", err)
	}
}

// TestImportOwnBackupKeepsLiveRecords is B4's last case: importing a backup of the account
// in use replaced the live object with a copy read from disk, dropping records it had not
// flushed. The backup must fold into the live account, which keeps them.
func TestImportOwnBackupKeepsLiveRecords(t *testing.T) {
	isolateAccountDir(t)

	alice, err := CreateAccount("Alice")
	if err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	ge.SetAccount(alice)
	recordPrestige(alice, "modern_age") // live, not flushed

	// A backup of Alice holding a theme she does not have in memory.
	exp := progressExport{
		Version:     accountSchemaVersion,
		AccountID:   alice.AccountID,
		DisplayName: "Alice",
		Unlocks:     AccountUnlocks{Themes: []string{"amber_crt"}},
	}
	exp.Signature = signProgressExport(&exp)
	blob, err := json.Marshal(&exp)
	if err != nil {
		t.Fatal(err)
	}

	got, err := ge.ImportAccountExport(blob, true)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if got != alice || ge.Account() != alice {
		t.Error("importing the live account's backup did not use the live account")
	}
	if !alice.HasTheme("amber_crt") || alice.Stats.TotalPrestiges != 1 {
		t.Errorf("live account after import: themes=%v stats=%+v", alice.UnlockedThemes(), alice.Stats)
	}
	disk := slotAccount(t, alice.AccountID)
	if !disk.HasTheme("amber_crt") || disk.Stats.TotalPrestiges != 1 {
		t.Errorf("disk after import: themes=%v stats=%+v", disk.Unlocks.Themes, disk.Stats)
	}
}

// TestRecoverAccountSwitchesAndKeepsOriginal is B1 at the engine level: recovery restores
// the code's account in its own slot and switches to it, and the account that was in use
// keeps everything, including records it had not flushed.
func TestRecoverAccountSwitchesAndKeepsOriginal(t *testing.T) {
	isolateAccountDir(t)

	bob, err := CreateAccount("Bob")
	if err != nil {
		t.Fatal(err)
	}
	bobCode := bob.RecoveryCode()
	if _, err := WipeAccountByID(bob.AccountID); err != nil {
		t.Fatal(err)
	}
	alice, err := CreateAccount("Alice")
	if err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	ge.SetAccount(alice)
	if err := ge.StartNewNamedGame("Rome"); err != nil {
		t.Fatal(err)
	}
	startedLoop(t, ge)
	t.Cleanup(ge.Stop)
	recordPrestige(alice, "modern_age")

	restored, ended, err := ge.RecoverAccount(bobCode)
	if err != nil {
		t.Fatalf("RecoverAccount: %v", err)
	}
	if !ended {
		t.Error("recovering during a live run did not end it")
	}
	if restored.AccountID != bob.AccountID || ge.Account() != restored || getActiveAccountID() != bob.AccountID {
		t.Errorf("after recovery: restored=%s live=%v active=%s", restored.AccountID, ge.Account(), getActiveAccountID())
	}
	a := slotAccount(t, alice.AccountID)
	if a.AccountID != alice.AccountID || a.DisplayName != "Alice" || a.Stats.TotalPrestiges != 1 {
		t.Errorf("Alice after the recovery: id=%s name=%q stats=%+v", a.AccountID, a.DisplayName, a.Stats)
	}
	if _, err := os.Stat(slotSave(alice.AccountID, "Rome")); err != nil {
		t.Errorf("Alice's run was not saved to her slot: %v", err)
	}

	// Recovering the code of the account in use changes nothing.
	again, ended, err := ge.RecoverAccount(bobCode)
	if err != nil || ended || again != restored {
		t.Errorf("recovering the live account's own code: acct=%v ended=%v err=%v", again, ended, err)
	}
	// A typo'd code is refused with nothing changed.
	if _, _, err := ge.RecoverAccount("AGEF-0000-0000"); err == nil {
		t.Error("a bad code was accepted")
	}
	if ge.Account() != restored {
		t.Error("a refused code changed the live account")
	}
}

// TestWipeOfRunOwnerOrphansRun: wiping the account that owns the run in memory must not let
// a later save of that run recreate the wiped slot (the exit handler saves the game).
func TestWipeOfRunOwnerOrphansRun(t *testing.T) {
	isolateAccountDir(t)

	alice, err := CreateAccount("Alice")
	if err != nil {
		t.Fatal(err)
	}
	ge := NewGameEngine()
	ge.SetAccount(alice)
	if err := ge.StartNewNamedGame("Rome"); err != nil {
		t.Fatal(err)
	}
	recordPrestige(alice, "modern_age") // dirty: detaching must not flush it back into the wiped slot

	if _, err := ge.WipeAccountByID(alice.AccountID); err != nil {
		t.Fatalf("wipe: %v", err)
	}
	if ge.Account() != nil {
		t.Error("the wiped account is still live")
	}
	if err := ge.SaveGame("autosave"); err == nil {
		t.Error("a save of the wiped account's run was written")
	}
	if _, err := os.Stat(accountDir(alice.AccountID)); !os.IsNotExist(err) {
		t.Errorf("the wiped slot came back (stat err = %v)", err)
	}

	// A new game has an owner again and saves normally.
	bob, err := CreateAccount("Bob")
	if err != nil {
		t.Fatal(err)
	}
	ge.SetAccount(bob)
	if err := ge.StartNewNamedGame("Carthage"); err != nil {
		t.Fatalf("new game after the wipe: %v", err)
	}
}
