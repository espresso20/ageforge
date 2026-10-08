package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/espresso20/ageforge/game"
)

// main_test.go: what a signal that ends the program does, through the same
// handler main installs (exitOnSignal) and the same engine main builds.
// Every test has a data root of its own; none touches the player's.

// dataTree reads every file under root: its bytes and its modification time.
func dataTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		rel, _ := filepath.Rel(root, path)
		out[rel] = hex.EncodeToString(sum[:]) + " " + fi.ModTime().Format(time.RFC3339Nano)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// stage makes an account in a new data root with a game saved under each
// name, and returns the root.
func stage(t *testing.T, games ...string) string {
	t.Helper()
	root := t.TempDir()
	t.Cleanup(game.SetDataDirForTest(root))
	acct, err := game.CreateNamedAccount("Exit Tests")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range games {
		mk := game.NewGameEngine()
		mk.SetAccount(acct)
		if err := mk.StartNewNamedGame(name); err != nil {
			t.Fatal(err)
		}
		// Something in it, so an empty game written over it would show.
		mk.StepTicks(30)
		if err := mk.SaveGame(name); err != nil {
			t.Fatal(err)
		}
	}
	if err := acct.Save(); err != nil {
		t.Fatal(err)
	}
	return root
}

// launch is the engine as main makes it: new, holding the account on disk.
func launch(t *testing.T) *game.GameEngine {
	t.Helper()
	engine := game.NewGameEngine()
	acct, found, err := game.LoadAccount()
	if err != nil || !found || !acct.Established() {
		t.Fatalf("LoadAccount: found=%v err=%v", found, err)
	}
	engine.SetAccount(acct)
	t.Cleanup(engine.Stop)
	return engine
}

// signalled sends the handler a terminate signal and waits for it to close
// the UI.
func signalled(t *testing.T, engine *game.GameEngine) {
	t.Helper()
	sigs := make(chan os.Signal, 1)
	closed := make(chan struct{})
	go exitOnSignal(sigs, engine, func() { close(closed) })
	sigs <- syscall.SIGTERM
	select {
	case <-closed:
	case <-time.After(10 * time.Second):
		t.Fatal("the handler did not close the UI")
	}
}

func playing(t *testing.T, engine *game.GameEngine) {
	t.Helper()
	go engine.Start()
	deadline := time.Now().Add(5 * time.Second)
	for !engine.Running() {
		if time.Now().After(deadline) {
			t.Fatal("the game did not start")
		}
		time.Sleep(time.Millisecond)
	}
}

// TestSignalAtTheMenuWritesNothing: a signal while the player is at the
// menu leaves every file under the data root byte for byte as it was, the
// account's autosave included. (The handler used to write the engine's
// empty game to "autosave".)
func TestSignalAtTheMenuWritesNothing(t *testing.T) {
	root := stage(t, game.AutosaveName, "rome")
	engine := launch(t)
	before := dataTree(t, root)
	if len(before) < 3 {
		t.Fatalf("staging left only %d files", len(before))
	}

	signalled(t, engine)
	leave(engine) // and main's own way out, after Run returns

	after := dataTree(t, root)
	if !reflect.DeepEqual(before, after) {
		for k, v := range after {
			if before[k] != v {
				t.Errorf("a signal at the menu changed %s", k)
			}
		}
		for k := range before {
			if _, ok := after[k]; !ok {
				t.Errorf("a signal at the menu removed %s", k)
			}
		}
	}
	st, err := game.ViewSave(game.AutosaveName)
	if err != nil || st.Tick != 30 {
		t.Errorf("the autosave is at tick %d (%v), it was saved at 30", st.Tick, err)
	}
}

// TestSignalInANamedGameSavesItUnderItsName: a signal during a game saves
// that game to its own save, last moment included, writes no autosave
// beside it, and leaves Continue pointing at it.
func TestSignalInANamedGameSavesItUnderItsName(t *testing.T) {
	stage(t, "carthage", "rome")
	engine := launch(t)
	if err := engine.LoadGame("carthage"); err != nil {
		t.Fatal(err)
	}
	if err := engine.LoadGame("rome"); err != nil {
		t.Fatal(err)
	}
	// Nothing on record, as on an account from an older version: it is the
	// exit's save that must name the game.
	if err := engine.Account().ForgetGame(""); err != nil {
		t.Fatal(err)
	}
	playing(t, engine)
	saved, err := game.ViewSave("rome")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.GatherResource("food", 3); err != nil {
		t.Fatal(err)
	}
	want := engine.GetState().Resources["food"].Amount

	signalled(t, engine)

	if engine.Running() {
		t.Error("the game is still running after the signal")
	}
	now, err := game.ViewSave("rome")
	if err != nil {
		t.Fatal(err)
	}
	if got := now.Resources["food"].Amount; got != want || got == saved.Resources["food"].Amount {
		t.Errorf("rome's save holds %v food: %v before the last moments, and the game ended on %v", got, saved.Resources["food"].Amount, want)
	}
	if game.SaveExists(game.AutosaveName) {
		t.Error("the signal wrote an autosave beside the named game")
	}
	cur, ok := engine.CurrentGame()
	if !ok || cur.Save.Name != "rome" || cur.Why != game.CurrentLast {
		t.Errorf("Continue opens %q (%s), want rome, the game last played", cur.Save.Name, cur.Why)
	}

	// Run returns after the handler closed the UI, and main leaves again:
	// nothing more is written.
	root := filepath.Dir(filepath.Dir(game.DataDir()))
	before := dataTree(t, root)
	leave(engine)
	if !reflect.DeepEqual(before, dataTree(t, root)) {
		t.Error("leaving a second time wrote a file")
	}
}

// TestSignalInAnUnnamedGameSavesTheAutosave: a game that was never given a
// name is still saved to the autosave.
func TestSignalInAnUnnamedGameSavesTheAutosave(t *testing.T) {
	stage(t)
	engine := launch(t)
	playing(t, engine)
	if _, err := engine.GatherResource("food", 3); err != nil {
		t.Fatal(err)
	}
	want := engine.GetState().Resources["food"].Amount

	signalled(t, engine)

	st, err := game.ViewSave(game.AutosaveName)
	if err != nil {
		t.Fatalf("no autosave after the signal: %v", err)
	}
	if got := st.Resources["food"].Amount; got != want {
		t.Errorf("the autosave holds %v food, the game ended on %v", got, want)
	}
}

// TestLeaveWithoutASignal: Ctrl+C in the terminal is a key, not a signal,
// and the menu's Quit and the quit command end Run the same way. main's
// way out after Run saves a game in play, and nothing at the menu.
func TestLeaveWithoutASignal(t *testing.T) {
	root := stage(t, "rome")
	engine := launch(t)
	before := dataTree(t, root)
	leave(engine)
	if !reflect.DeepEqual(before, dataTree(t, root)) {
		t.Error("leaving from the menu wrote a file")
	}

	if err := engine.LoadGame("rome"); err != nil {
		t.Fatal(err)
	}
	playing(t, engine)
	if _, err := engine.GatherResource("food", 3); err != nil {
		t.Fatal(err)
	}
	want := engine.GetState().Resources["food"].Amount
	leave(engine)
	st, err := game.ViewSave("rome")
	if err != nil {
		t.Fatal(err)
	}
	if got := st.Resources["food"].Amount; got != want || engine.Running() {
		t.Errorf("after leaving a game: rome holds %v food (the game ended on %v), running %v", got, want, engine.Running())
	}
	if game.SaveExists(game.AutosaveName) {
		t.Error("leaving a named game wrote an autosave")
	}
}
