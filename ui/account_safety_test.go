package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/theme"
)

// The account commands during play (achievements audit B1, B3 and B4), through the real
// command handler and dashboard. Every test runs in a temp data root, never data/.

// waitRunning waits until eng's tick loop is going.
func waitRunning(t *testing.T, eng *game.GameEngine) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !eng.Running() {
		if time.Now().After(deadline) {
			t.Fatal("the tick loop did not start")
		}
		time.Sleep(time.Millisecond)
	}
}

// slotSavePath is where save name lives in account id's slot; accounts is the data
// root's accounts/ folder.
func slotSavePath(accounts, id, name string) string {
	return filepath.Join(accounts, id, "saves", name+".json")
}

// TestAccountSwitchInGameGoesToMenu is B4 through the dashboard: `account switch` during a
// game used to leave the run going under the new account. It must save the run to the
// account it belongs to, stop it, switch, and take the player to the main menu.
func TestAccountSwitchInGameGoesToMenu(t *testing.T) {
	d, eng := mapTestDashboard(t, true)      // account "Map Tester", temp data root
	accounts := filepath.Dir(game.DataDir()) // DataDir is the active slot, <root>/accounts/<id>
	first := eng.Account()
	second, err := game.CreateAccount("Second Player")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := game.SwitchAccount(first.AccountID); err != nil { // creating made Second active
		t.Fatal(err)
	}
	if err := eng.StartNewNamedGame("Rome"); err != nil {
		t.Fatal(err)
	}
	go eng.Start()
	t.Cleanup(eng.Stop)
	waitRunning(t, eng)

	d.runForTest("account switch Second Player")

	if eng.Running() {
		t.Error("the game is still running after the switch")
	}
	if got := eng.Account(); got == nil || got.AccountID != second.AccountID {
		t.Fatalf("live account after the switch: %v", got)
	}
	if name, _ := d.pages.GetFrontPage(); name != accountNoticePage {
		t.Errorf("front page after the switch is %q, want the notice over the main menu", name)
	}
	if _, err := os.Stat(slotSavePath(accounts, first.AccountID, "Rome")); err != nil {
		t.Errorf("the game was not saved to the account it belongs to: %v", err)
	}
	if _, err := os.Stat(slotSavePath(accounts, second.AccountID, "Rome")); !os.IsNotExist(err) {
		t.Errorf("the game was saved into the new account's slot (stat err = %v)", err)
	}
}

// TestAccountSwitchReplies covers the command's other answers: an unknown name and the
// account already in use change nothing.
func TestAccountSwitchReplies(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	alice, err := game.CreateAccount("Alice")
	if err != nil {
		t.Fatal(err)
	}
	eng := game.NewGameEngine()
	eng.SetAccount(alice)

	if res := HandleCommand("account switch Nobody", eng); res.Type != "error" || res.ToMenu || !strings.Contains(res.Message, "No account named") {
		t.Errorf("unknown name: %+v", res)
	}
	if res := HandleCommand("account switch alice", eng); res.Type != "info" || res.ToMenu || !strings.Contains(res.Message, "already playing") {
		t.Errorf("the account in use: %+v", res)
	}
	if eng.Account() != alice {
		t.Error("a refused switch changed the live account")
	}
}

// TestAccountImportInGameDoesNotSwitch is B4 for import: `account import` used to switch
// to the imported account and leave the game running under it. It must restore the backup
// into its own slot and leave the game and the live account alone.
func TestAccountImportInGameDoesNotSwitch(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	bob, err := game.CreateAccount("Bob")
	if err != nil {
		t.Fatal(err)
	}
	blob, err := bob.ExportProgress()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "bob.json")
	if err := os.WriteFile(path, blob, 0644); err != nil {
		t.Fatal(err)
	}
	alice, err := game.CreateAccount("Alice")
	if err != nil {
		t.Fatal(err)
	}
	eng := game.NewGameEngine()
	eng.SetAccount(alice)
	if err := eng.StartNewNamedGame("Rome"); err != nil {
		t.Fatal(err)
	}
	go eng.Start()
	t.Cleanup(eng.Stop)
	waitRunning(t, eng)

	res := HandleCommand("account import "+path, eng)
	if res.Type != "info" || res.ToMenu {
		t.Fatalf("import reply: %+v", res)
	}
	if !strings.Contains(res.Message, "account switch Bob") {
		t.Errorf("the reply does not say how to switch: %q", res.Message)
	}
	if !eng.Running() || eng.Account() != alice {
		t.Errorf("the import disturbed the game: running=%v account=%v", eng.Running(), eng.Account())
	}
}

// TestAccountRecoverGuardCoversAllProgress is B1's guard: it asked for confirm only when
// the account had theme unlocks, so an account with achievements and stats but no themes
// was switched away without a word. Any progress now asks first, and says what the account
// holds; the player's own code restores nothing.
func TestAccountRecoverGuardCoversAllProgress(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	bob, err := game.CreateAccount("Bob")
	if err != nil {
		t.Fatal(err)
	}
	bobCode := bob.RecoveryCode()
	alice, err := game.CreateAccount("Alice")
	if err != nil {
		t.Fatal(err)
	}
	eng := game.NewGameEngine()
	eng.SetAccount(alice)
	alice.RecordPrestige() // an achievement and a lifetime prestige, no themes

	res := HandleCommand("account recover "+alice.RecoveryCode(), eng)
	if res.Type != "info" || !strings.Contains(res.Message, "nothing to restore") {
		t.Errorf("recovering your own code: %+v", res)
	}

	res = HandleCommand("account recover "+bobCode, eng)
	if res.Type != "warning" || !strings.Contains(res.Message, "confirm") {
		t.Fatalf("an account with achievements but no themes was not guarded: %+v", res)
	}
	for _, want := range []string{"1 achievement", "1 prestige", "account switch Alice"} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("the guard does not mention %q: %q", want, res.Message)
		}
	}
	if eng.Account() != alice {
		t.Fatal("the guard let the recovery through")
	}

	res = HandleCommand("account recover "+bobCode+" confirm", eng)
	if res.Type != "info" || res.ToMenu {
		t.Fatalf("confirmed recovery: %+v", res)
	}
	if got := eng.Account(); got == nil || got.AccountID != bob.AccountID {
		t.Fatalf("after recovery the live account is %v, want Bob", got)
	}
	// Alice kept everything, including the record she had not flushed.
	listed := false
	for _, s := range eng.ListAccounts() {
		if s.AccountID == alice.AccountID {
			listed = true
			if s.TotalPrestiges != 1 || s.Achievements != 1 {
				t.Errorf("Alice after the recovery: %+v", s)
			}
		}
	}
	if !listed {
		t.Error("Alice is gone from the account list after the recovery")
	}
}

// TestDevTouchedRunUnlocksNoThemes is B3's theme leak: with dev mode on, `/age
// galactic_age` completed the theme milestones on the next tick and the dashboard wrote
// all five flavor themes to the account for good.
func TestDevTouchedRunUnlocksNoThemes(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	prev := game.DevModeActive
	game.DevModeActive = true
	t.Cleanup(func() { game.DevModeActive = prev })

	if msg := game.DevExecCommand("/age galactic_age", eng); msg != "jumped to galactic_age" {
		t.Fatalf("/age: %q", msg)
	}
	eng.StepTicks(1)
	state := eng.GetState()
	gated := 0
	for _, key := range completedUnlockKeys(state.Milestones) {
		if _, ok := theme.UnlockedBy(key); ok {
			gated++
		}
	}
	if gated == 0 {
		t.Fatal("precondition: no theme-gating milestone completed, so the test proves nothing")
	}
	d.refresh()
	if got := eng.Account().UnlockedThemes(); len(got) != 0 {
		t.Errorf("a dev-touched run unlocked themes on the account: %v", got)
	}
}

// TestThemeUnlocksFollowAccountRecords: the dashboard grants a milestone's theme only when
// the run records to the account, and a skipped key is still evaluated for a later run.
func TestThemeUnlocksFollowAccountRecords(t *testing.T) {
	d, eng := mapTestDashboard(t, true)
	const key = "bronze_pioneer"
	themeKey, ok := theme.UnlockedBy(key)
	if !ok {
		t.Fatalf("%s no longer gates a theme; pick another key", key)
	}
	state := game.GameState{Milestones: game.MilestoneState{
		Milestones: map[string]game.MilestoneInfo{key: {Completed: true}},
	}}

	d.processThemeUnlocks(state) // AccountRecords false: a dev-touched or foreign run
	if eng.Account().HasTheme(themeKey) {
		t.Fatal("a run that does not record to the account unlocked a theme")
	}
	state.AccountRecords = true
	d.processThemeUnlocks(state)
	if !eng.Account().HasTheme(themeKey) {
		t.Error("a clean run's milestone did not unlock its theme")
	}

	// After a switch, the new account's own run earns the theme too, although the key
	// was processed for the first account earlier in this session.
	other, err := game.CreateAccount("Other Player")
	if err != nil {
		t.Fatal(err)
	}
	eng.SetAccount(other)
	d.processThemeUnlocks(state)
	if !other.HasTheme(themeKey) {
		t.Error("the second account's run did not unlock the theme after a switch")
	}
}

// TestAccountWipeWordingMatchesWipe: `account wipe` said game saves were not affected,
// but a wipe deletes the account's whole slot, saves included, after backing it up. The
// replies must say so, and the wipe must do what they say.
func TestAccountWipeWordingMatchesWipe(t *testing.T) {
	defer game.SetDataDirForTest(t.TempDir())()
	alice, err := game.CreateAccount("Alice")
	if err != nil {
		t.Fatal(err)
	}
	eng := game.NewGameEngine()
	eng.SetAccount(alice)
	if err := eng.StartNewNamedGame("Rome"); err != nil {
		t.Fatal(err)
	}
	save := slotSavePath(filepath.Dir(game.DataDir()), alice.AccountID, "Rome")
	if _, err := os.Stat(save); err != nil {
		t.Fatalf("no save in the slot to wipe: %v", err)
	}

	if msg := HandleCommand("account wipe", eng).Message; !strings.Contains(msg, "every save in its slot") || strings.Contains(msg, "not affected") {
		t.Errorf("account wipe = %q; it must say the account's saves are deleted", msg)
	}
	if msg := HandleCommand("account", eng).Message; !strings.Contains(msg, "stats and saves for good") {
		t.Errorf("account does not say a wipe deletes saves:\n%s", msg)
	}

	backup, err := game.WipeAccountByID(alice.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(save); !os.IsNotExist(err) {
		t.Errorf("the save survived the wipe (stat err = %v)", err)
	}
	if _, err := os.Stat(filepath.Join(backup, "saves", "Rome.json")); err != nil {
		t.Errorf("the backup taken before the wipe has no copy of the save: %v", err)
	}
}
