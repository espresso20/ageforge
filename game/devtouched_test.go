package game

import (
	"strings"
	"testing"
)

// The dev console must not earn account records (achievements audit B3). Every test runs
// in a temp data root (isolateAccountDir), never data/.

// withDevMode turns the dev console on for the test and restores it (and god mode) after.
func withDevMode(t *testing.T) {
	t.Helper()
	prevDev, prevGod := DevModeActive, DevGodMode
	DevModeActive = true
	t.Cleanup(func() { DevModeActive, DevGodMode = prevDev, prevGod })
}

// devEngine is a fresh engine holding a fresh account, with the run owned by it.
func devEngine(t *testing.T, name string) (*GameEngine, *Account) {
	t.Helper()
	acct, err := CreateAccount(name)
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	ge := NewGameEngine()
	ge.SetAccount(acct)
	if err := ge.StartNewNamedGame("run"); err != nil {
		t.Fatalf("StartNewNamedGame: %v", err)
	}
	return ge, acct
}

// devLineCount counts the dev-touched notice in ge's log.
func devLineCount(ge *GameEngine) int {
	n := 0
	for _, e := range ge.GetLogs() {
		if e.Message == devTouchedLogLine {
			n++
		}
	}
	return n
}

// assertNoRecords fails if acct holds any lifetime stat or achievement, in memory or on disk.
func assertNoRecords(t *testing.T, acct *Account) {
	t.Helper()
	stats, ach := acct.LifetimeStats()
	if stats.TotalPrestiges != 0 || stats.HighestAge != "" || len(ach) != 0 {
		t.Errorf("a dev-touched run recorded to the account: stats=%+v achievements=%v", stats, ach)
	}
	if err := acct.FlushIfDirty(); err != nil {
		t.Fatal(err)
	}
	if disk := slotAccount(t, acct.AccountID); disk.Stats.TotalPrestiges != 0 || len(disk.Achievements) != 0 {
		t.Errorf("a dev-touched run's records reached the account file: %+v %v", disk.Stats, disk.Achievements)
	}
}

// TestDevTouchedRunGrantsNothing is B3: with dev mode on, `/age modern_age` then prestige
// granted first_prestige (and a lifetime prestige with no highest age). A run the dev
// console has changed records nothing, and says so once in the log.
func TestDevTouchedRunGrantsNothing(t *testing.T) {
	isolateAccountDir(t)
	withDevMode(t)
	ge, acct := devEngine(t, "Dev Tester")

	if msg := DevExecCommand("/age modern_age", ge); msg != "jumped to modern_age" {
		t.Fatalf("/age: %q", msg)
	}
	st := ge.GetState()
	if !st.DevTouched || st.AccountRecords {
		t.Errorf("after /age: DevTouched=%v AccountRecords=%v", st.DevTouched, st.AccountRecords)
	}
	if n := devLineCount(ge); n != 1 {
		t.Errorf("the dev-touched notice was logged %d times, want once", n)
	}
	// Later dev commands do not repeat the notice.
	DevExecCommand("/fill", ge)
	DevExecCommand("/give food 5", ge)
	if n := devLineCount(ge); n != 1 {
		t.Errorf("the notice repeated: %d lines", n)
	}

	if err := ge.DoPrestige(); err != nil {
		t.Fatalf("DoPrestige: %v", err)
	}
	assertNoRecords(t, acct)
}

// TestDevTouchedSurvivesSaveLoadAndPrestige: the flag is saved with the run, so reloading
// does not clear it, and prestige (which carries the run's points and upgrades forward)
// keeps it too.
func TestDevTouchedSurvivesSaveLoadAndPrestige(t *testing.T) {
	isolateAccountDir(t)
	withDevMode(t)
	ge, acct := devEngine(t, "Dev Tester")

	DevExecCommand("/give food 100", ge)
	if err := ge.SaveGame("dev"); err != nil {
		t.Fatal(err)
	}

	ge2 := NewGameEngine()
	ge2.SetAccount(acct)
	if err := ge2.LoadGame("dev"); err != nil {
		t.Fatal(err)
	}
	if !ge2.GetState().DevTouched {
		t.Fatal("reloading cleared the dev-touched flag")
	}
	if ge2.GetState().CheaterBadge {
		t.Error("a dev-touched save reads as modified; the flag is signed like any field")
	}
	prestigeNow(t, ge2)
	if !ge2.GetState().DevTouched {
		t.Error("prestige cleared the dev-touched flag")
	}
	assertNoRecords(t, acct)

	// Still set after another save and load.
	if err := ge2.SaveGame("dev"); err != nil {
		t.Fatal(err)
	}
	ge3 := NewGameEngine()
	ge3.SetAccount(acct)
	if err := ge3.LoadGame("dev"); err != nil {
		t.Fatal(err)
	}
	if !ge3.GetState().DevTouched {
		t.Error("the flag did not survive a second save and load")
	}
}

// TestDevModeWithoutCommandsStillRecords: the owner plays local builds with the console
// unlocked. With dev mode on but no dev command used, a run records as usual.
func TestDevModeWithoutCommandsStillRecords(t *testing.T) {
	isolateAccountDir(t)
	withDevMode(t)
	ge, acct := devEngine(t, "Owner")

	if st := ge.GetState(); st.DevTouched || !st.AccountRecords {
		t.Fatalf("dev mode alone marked the run: DevTouched=%v AccountRecords=%v", st.DevTouched, st.AccountRecords)
	}
	// Read-only /ages does not count as changing the run.
	if out := DevExecCommand("/ages", ge); !strings.Contains(out, "modern_age") {
		t.Fatalf("/ages = %q", out)
	}
	// Neither does a command that is refused.
	for _, cmd := range []string{"/age atlantis", "/give food lots", "/build no_such_building", "/prestige 99", "/speed -1", "/nonsense"} {
		DevExecCommand(cmd, ge)
	}
	if ge.GetState().DevTouched {
		t.Fatal("a read-only or refused dev command marked the run")
	}

	prestigeNow(t, ge)
	stats, ach := acct.LifetimeStats()
	if stats.TotalPrestiges != 1 || len(ach) != 1 || ach[0] != "first_prestige" {
		t.Errorf("a clean run with dev mode on did not record: stats=%+v achievements=%v", stats, ach)
	}
}

// TestEveryDevCommandMarksTheRun: each console command that succeeds marks the run.
func TestEveryDevCommandMarksTheRun(t *testing.T) {
	isolateAccountDir(t)
	withDevMode(t)

	for _, cmd := range []string{
		"/age bronze_age", "/fill", "/give food 5", "/techs", "/build hut",
		"/prestige 1", "/speed 2", "/god",
	} {
		ge := NewGameEngine()
		DevGodMode = false
		DevConsoleCommand(cmd, ge)
		if !ge.GetState().DevTouched {
			t.Errorf("%s did not mark the run", cmd)
		}
	}
	DevGodMode = false

	// The console-only commands, where they succeed.
	ge := catEngine(t, "iron_age", 1)
	if out := DevConsoleCommand("/catastrophe", ge); ge.pendingCatastrophe == "" {
		t.Fatalf("/catastrophe did not run: %q", out)
	}
	if !ge.GetState().DevTouched {
		t.Error("/catastrophe did not mark the run")
	}
	refused := catEngine(t, "bronze_age", 1)
	if out := DevConsoleCommand("/catastrophe", refused); !strings.Contains(out, "refused") {
		t.Fatalf("/catastrophe in the Stone Era: %q", out)
	}
	if refused.GetState().DevTouched {
		t.Error("a refused /catastrophe marked the run")
	}
}

// TestNewGameClearsDevTouched: a new game is a new run and starts clean, unless god mode
// is still on, which makes it free to build and so marks it from the start.
func TestNewGameClearsDevTouched(t *testing.T) {
	isolateAccountDir(t)
	withDevMode(t)
	ge, _ := devEngine(t, "Dev Tester")

	DevExecCommand("/fill", ge)
	if err := ge.StartNewNamedGame("fresh"); err != nil {
		t.Fatal(err)
	}
	if st := ge.GetState(); st.DevTouched || !st.AccountRecords {
		t.Errorf("a new game kept the flag: DevTouched=%v AccountRecords=%v", st.DevTouched, st.AccountRecords)
	}

	DevExecCommand("/god", ge) // on, and left on
	if err := ge.StartNewNamedGame("godly"); err != nil {
		t.Fatal(err)
	}
	if !ge.GetState().DevTouched {
		t.Error("a new game with god mode still on was not marked")
	}
	if n := devLineCount(ge); n != 1 {
		t.Errorf("the new god-mode run logged the notice %d times, want once", n)
	}

	// Loading a clean save with god mode on marks that run too.
	DevGodMode = false
	if err := ge.StartNewNamedGame("clean"); err != nil {
		t.Fatal(err)
	}
	DevGodMode = true
	ge2 := NewGameEngine()
	if err := ge2.LoadGame("clean"); err != nil {
		t.Fatal(err)
	}
	if !ge2.GetState().DevTouched {
		t.Error("loading a save with god mode on did not mark the run")
	}
}
