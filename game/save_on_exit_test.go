package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

// save_on_exit_test.go: what leaving the program does to the saves
// (SaveOnExit). Nothing at the menu, the game's own save in a game.

// waitLoop waits until ge's tick loop is going.
func waitLoop(t *testing.T, ge *GameEngine) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !ge.Running() {
		if time.Now().After(deadline) {
			t.Fatal("the tick loop did not start")
		}
		time.Sleep(time.Millisecond)
	}
}

// savedFood is the food a save file holds.
func savedFood(t *testing.T, name string) float64 {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(saveDirectory(), name+".json"))
	if err != nil {
		t.Fatalf("the save %q: %v", name, err)
	}
	var gs GameSave
	if err := json.Unmarshal(data, &gs); err != nil {
		t.Fatal(err)
	}
	return gs.Resources["food"]
}

// launch is the engine as main makes it: a new engine holding the account
// read from disk.
func launch(t *testing.T) (*GameEngine, *Account) {
	t.Helper()
	acct, found, err := LoadAccount()
	if err != nil || !found {
		t.Fatalf("LoadAccount: found=%v err=%v", found, err)
	}
	ge := NewGameEngine()
	ge.SetAccount(acct)
	return ge, acct
}

// TestSaveOnExitAtTheMenuWritesNothing: with no game in play, leaving the
// program changes no file, not a byte and not a time. The account's real
// autosave, which the old exit handler overwrote with an empty game, is as
// it was. The same holds back at the menu after a game.
func TestSaveOnExitAtTheMenuWritesNothing(t *testing.T) {
	mk, acct := menuAccount(t)
	root := rootDataDir()
	// An autosave with a game in it, and a named game.
	startGame(t, mk, AutosaveName)
	startGame(t, mk, "ashford")
	if err := acct.Save(); err != nil {
		t.Fatal(err)
	}
	autosaveFood := savedFood(t, AutosaveName)

	ge, _ := launch(t)
	before := treeOf(t, root)
	name, err := ge.SaveOnExit()
	ge.Stop()
	if name != "" || err != nil {
		t.Errorf("at the menu SaveOnExit saved %q (%v)", name, err)
	}
	if after := treeOf(t, root); !reflect.DeepEqual(before, after) {
		for k, v := range after {
			if before[k] != v {
				t.Errorf("leaving from the menu changed %s", k)
			}
		}
	}
	if got := savedFood(t, AutosaveName); got != autosaveFood {
		t.Errorf("the autosave holds %v food, it held %v", got, autosaveFood)
	}

	// A game is played and left with Esc (which saves and stops), and the
	// program is left from the menu after it: the exit adds nothing.
	if err := ge.LoadGame("ashford"); err != nil {
		t.Fatal(err)
	}
	go ge.Start()
	waitLoop(t, ge)
	if err := ge.SaveGame(ge.ActiveSaveName()); err != nil {
		t.Fatal(err)
	}
	ge.Stop()
	before = treeOf(t, root)
	if name, err := ge.SaveOnExit(); name != "" || err != nil {
		t.Errorf("back at the menu SaveOnExit saved %q (%v)", name, err)
	}
	ge.Stop()
	if after := treeOf(t, root); !reflect.DeepEqual(before, after) {
		t.Error("leaving from the menu after a game changed a file")
	}
}

// TestSaveOnExitSavesTheGameInPlayUnderItsName: in a named game the exit
// saves that game, to its own save, with its last moment in it; it does
// not write an autosave beside it; and the account records it as the game
// last played, so Continue opens it.
func TestSaveOnExitSavesTheGameInPlayUnderItsName(t *testing.T) {
	mk, acct := menuAccount(t)
	startGame(t, mk, "carthage")
	startGame(t, mk, "rome")
	if err := acct.Save(); err != nil {
		t.Fatal(err)
	}

	ge, acct := launch(t)
	if err := ge.LoadGame("rome"); err != nil {
		t.Fatal(err)
	}
	// An account with nothing on record (as one from before the records
	// is), so that it is the exit's save that names the game.
	if err := acct.ForgetGame(""); err != nil {
		t.Fatal(err)
	}
	go ge.Start()
	waitLoop(t, ge)
	had := savedFood(t, "rome")
	if _, err := ge.GatherResource("food", 3); err != nil {
		t.Fatal(err)
	}
	want := ge.GetState().Resources["food"].Amount

	name, err := ge.SaveOnExit()
	if name != "rome" || err != nil {
		t.Fatalf("SaveOnExit saved %q (%v), want rome", name, err)
	}
	if ge.Running() {
		t.Error("the game is still running after the exit's save")
	}
	if got := savedFood(t, "rome"); got != want || got == had {
		t.Errorf("rome's save holds %v food: it held %v before the last moments and the game ended on %v", got, had, want)
	}
	if SaveExists(AutosaveName) {
		t.Error("the exit wrote an autosave beside the named game")
	}
	if _, last := acct.GameRecord(); last != "rome" {
		t.Errorf("the game last played is on record as %q", last)
	}
	if cur, ok := ge.CurrentGame(); !ok || cur.Save.Name != "rome" || cur.Why != CurrentLast {
		t.Errorf("Continue opens %q (%s)", cur.Save.Name, cur.Why)
	}

	// Leaving twice, as a signal and then the end of main do, saves once.
	stamp, _ := os.Stat(filepath.Join(saveDirectory(), "rome.json"))
	if name, err := ge.SaveOnExit(); name != "" || err != nil {
		t.Errorf("a second SaveOnExit saved %q (%v)", name, err)
	}
	if again, _ := os.Stat(filepath.Join(saveDirectory(), "rome.json")); !again.ModTime().Equal(stamp.ModTime()) {
		t.Error("a second SaveOnExit wrote the save again")
	}
}

// TestSaveOnExitUnnamedGameSavesTheAutosave: a game in play that was never
// given a name is saved where it always was, the autosave.
func TestSaveOnExitUnnamedGameSavesTheAutosave(t *testing.T) {
	ge, _ := menuAccount(t)
	go ge.Start()
	waitLoop(t, ge)
	if _, err := ge.GatherResource("food", 3); err != nil {
		t.Fatal(err)
	}
	want := ge.GetState().Resources["food"].Amount
	name, err := ge.SaveOnExit()
	if name != AutosaveName || err != nil {
		t.Fatalf("SaveOnExit saved %q (%v), want the autosave", name, err)
	}
	if got := savedFood(t, AutosaveName); got != want {
		t.Errorf("the autosave holds %v food, the game ended on %v", got, want)
	}
}

// TestSaveOnExitFromTwoGoroutines: a signal's handler and the end of main
// may leave at the same moment. One saves, the other finds nothing to do,
// and the save is whole.
func TestSaveOnExitFromTwoGoroutines(t *testing.T) {
	ge, _ := menuAccount(t)
	if err := ge.StartNewNamedGame("rome"); err != nil {
		t.Fatal(err)
	}
	go ge.Start()
	waitLoop(t, ge)
	var wg sync.WaitGroup
	names := make([]string, 2)
	for i := range names {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name, err := ge.SaveOnExit()
			if err != nil {
				t.Errorf("SaveOnExit: %v", err)
			}
			names[i] = name
		}(i)
	}
	wg.Wait()
	if !(names[0] == "rome" && names[1] == "" || names[0] == "" && names[1] == "rome") {
		t.Errorf("two exits saved %q", names)
	}
	if _, err := ViewSave("rome"); err != nil {
		t.Errorf("the save does not read after two exits: %v", err)
	}
	if _, err := os.Stat(filepath.Join(saveDirectory(), "rome.json.tmp")); !os.IsNotExist(err) {
		t.Error("a half-written save was left behind")
	}
}
