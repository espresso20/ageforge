package game

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// current_game_test.go covers the two things the main menu asks of the game:
// which save is the current one (current_game.go), and a look at a save that
// changes nothing (ViewSave).

// menuAccount makes a named account in a temp data root and an engine that
// holds it.
func menuAccount(t *testing.T) (*GameEngine, *Account) {
	t.Helper()
	isolateAccountDir(t)
	acct, err := CreateNamedAccount("Menu Tests")
	if err != nil {
		t.Fatalf("CreateNamedAccount: %v", err)
	}
	ge := NewGameEngine()
	ge.SetAccount(acct)
	return ge, acct
}

// backdate rewrites a save as if it had been written ago before now, signed
// as the game signs it, so it loads as an untouched save that has been away.
func backdate(t *testing.T, name string, ago time.Duration) {
	t.Helper()
	path := savePath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var gs GameSave
	if err := json.Unmarshal(data, &gs); err != nil {
		t.Fatal(err)
	}
	_, elite := verifySave(&gs)
	gs.Timestamp = time.Now().Add(-ago)
	gs.Signature, gs.Proof = "", ""
	sig := signSave(gs, saveHMACKey)
	gs.Signature = sig
	if elite {
		mac := hmac.New(sha256.New, []byte(forgeMasterKey))
		mac.Write([]byte(sig))
		gs.Proof = hex.EncodeToString(mac.Sum(nil))
	}
	out, err := json.MarshalIndent(gs, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		t.Fatal(err)
	}
	if valid, _ := verifySave(&gs); !valid {
		t.Fatalf("the backdated save %q does not verify", name)
	}
}

// treeOf reads every file under root: its bytes and its modification time.
func treeOf(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if d.IsDir() {
			out[rel+"/"] = "dir"
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		out[rel] = hex.EncodeToString(sum[:]) + " " + fi.ModTime().Format(time.RFC3339Nano)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// startGame starts a named game on ge and gives it something to show.
func startGame(t *testing.T, ge *GameEngine, name string) {
	t.Helper()
	if err := ge.StartNewNamedGame(name); err != nil {
		t.Fatalf("StartNewNamedGame(%q): %v", name, err)
	}
	ge.mu.Lock()
	ge.tick = 4321
	ge.Buildings.counts["hut"] = 3
	ge.Buildings.unlocked["hut"] = true
	ge.Resources.Add("wood", 40)
	ge.mu.Unlock()
	if err := ge.SaveGame(name); err != nil {
		t.Fatalf("SaveGame(%q): %v", name, err)
	}
}

// TestViewSaveLooksAndChangesNothing: a look at a save returns the state as
// it was written, with no time away applied, and leaves every file and the
// engine in play exactly as they were. A real load of the same save, for
// contrast, does move on.
func TestViewSaveLooksAndChangesNothing(t *testing.T) {
	ge, acct := menuAccount(t)
	root := rootDataDir()
	startGame(t, ge, "ashford")
	backdate(t, "ashford", 3*time.Hour)

	// The engine in play: a fresh one, as the menu holds after a launch.
	live := NewGameEngine()
	live.SetAccount(acct)
	liveBefore := live.GetState()
	daysBefore, _ := acct.LifetimeStats()
	_, sumBefore := live.Badges()
	before := treeOf(t, root)

	st, err := ViewSave("ashford")
	if err != nil {
		t.Fatalf("ViewSave: %v", err)
	}
	if _, ok := live.CurrentGame(); !ok {
		t.Fatal("CurrentGame found no save")
	}

	if st.Tick != 4321 {
		t.Errorf("the view is at tick %d, the save at 4321: time away was applied", st.Tick)
	}
	if got := st.Buildings["hut"].Count; got != 3 {
		t.Errorf("the view has %d huts, the save 3", got)
	}
	if st.Seed == 0 {
		t.Error("the view carries no seed, so its map would not be the save's")
	}
	if st.AccountStats != nil {
		t.Error("the view's engine holds an account")
	}
	if after := treeOf(t, root); !reflect.DeepEqual(before, after) {
		for k, v := range after {
			if before[k] != v {
				t.Errorf("a look at the save changed %s", k)
			}
		}
		for k := range before {
			if _, ok := after[k]; !ok {
				t.Errorf("a look at the save removed %s", k)
			}
		}
	}
	liveAfter := live.GetState()
	if liveAfter.Tick != liveBefore.Tick || liveAfter.Age != liveBefore.Age || live.ActiveSaveName() != AutosaveName ||
		len(liveAfter.Buildings) != len(liveBefore.Buildings) || liveAfter.Buildings["hut"].Count != liveBefore.Buildings["hut"].Count {
		t.Errorf("the engine in play changed: tick %d to %d, active save %q", liveBefore.Tick, liveAfter.Tick, live.ActiveSaveName())
	}
	if daysAfter, _ := acct.LifetimeStats(); !reflect.DeepEqual(daysBefore, daysAfter) {
		t.Errorf("the account's lifetime stats changed: %+v to %+v", daysBefore, daysAfter)
	}
	if _, sumAfter := live.Badges(); sumAfter.Earned != sumBefore.Earned || sumAfter.Points != sumBefore.Points {
		t.Errorf("a look at the save earned something: %+v to %+v", sumBefore, sumAfter)
	}

	// The contrast: loading it plays the three hours away.
	scratch := NewGameEngine()
	if err := scratch.LoadGame("ashford"); err != nil {
		t.Fatalf("LoadGame: %v", err)
	}
	if scratch.GetState().Tick <= 4321 {
		t.Error("the control load did not apply the time away, so this test proves nothing")
	}

	// A save that is not there, and one that cannot be read.
	if _, err := ViewSave("no_such_game"); err == nil {
		t.Error("ViewSave of a missing save returned no error")
	}
	if err := os.WriteFile(filepath.Join(saveDirectory(), "torn.json"), []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ViewSave("torn"); err == nil {
		t.Error("ViewSave of a damaged save returned no error")
	}
}

// TestCurrentGameRule: none, one, many, a marked game, a marked game that
// was deleted, and a record that names a save that is gone.
func TestCurrentGameRule(t *testing.T) {
	ge, acct := menuAccount(t)

	if cur, ok := ge.CurrentGame(); ok {
		t.Fatalf("an account with no saves has a current game: %+v", cur)
	}

	// One save: it is the current game, as the game last played.
	startGame(t, ge, "ashford")
	cur, ok := ge.CurrentGame()
	if !ok || cur.Save.Name != "ashford" || cur.Why != CurrentLast {
		t.Fatalf("one save: got %q (%s, ok=%v), want ashford as the game last played", cur.Save.Name, cur.Why, ok)
	}

	// Many saves: the one played last wins, whatever the files' ages.
	startGame(t, ge, "brackwater")
	startGame(t, ge, "corvale")
	backdate(t, "corvale", 48*time.Hour)
	if cur, _ := ge.CurrentGame(); cur.Save.Name != "corvale" || cur.Why != CurrentLast {
		t.Errorf("many saves: got %q (%s), want corvale, the game last played", cur.Save.Name, cur.Why)
	}
	// Loading another makes that one the game last played.
	if err := ge.LoadGame("ashford"); err != nil {
		t.Fatal(err)
	}
	if cur, _ := ge.CurrentGame(); cur.Save.Name != "ashford" || cur.Why != CurrentLast {
		t.Errorf("after a load: got %q (%s), want ashford", cur.Save.Name, cur.Why)
	}
	// A copy saved under another name (the exit handler's autosave) is not
	// the game in play and names nothing.
	if err := ge.SaveGame(AutosaveName); err != nil {
		t.Fatal(err)
	}
	if _, last := acct.GameRecord(); last != "ashford" {
		t.Errorf("a save under another name changed the game last played to %q", last)
	}

	// A marked game wins over the game last played.
	if err := acct.SetMainGame("brackwater"); err != nil {
		t.Fatal(err)
	}
	if cur, _ := ge.CurrentGame(); cur.Save.Name != "brackwater" || cur.Why != CurrentMain {
		t.Errorf("a marked game: got %q (%s), want brackwater as the main game", cur.Save.Name, cur.Why)
	}
	if err := ge.LoadGame("corvale"); err != nil {
		t.Fatal(err)
	}
	if cur, _ := ge.CurrentGame(); cur.Save.Name != "brackwater" {
		t.Errorf("playing another game unseated the marked one: got %q", cur.Save.Name)
	}

	// The marked game is deleted: the game last played is current again,
	// with the mark still on record and ignored.
	if err := DeleteSave("brackwater"); err != nil {
		t.Fatal(err)
	}
	if cur, _ := ge.CurrentGame(); cur.Save.Name != "corvale" || cur.Why != CurrentLast {
		t.Errorf("a deleted marked game: got %q (%s), want corvale, the game last played", cur.Save.Name, cur.Why)
	}
	// Forgetting it drops the mark, so a new save of that name is not it.
	if err := acct.ForgetGame("brackwater"); err != nil {
		t.Fatal(err)
	}
	if mainGame, _ := acct.GameRecord(); mainGame != "" {
		t.Errorf("the mark outlived ForgetGame: %q", mainGame)
	}

	// Unmarking: SetMainGame("") after a mark.
	_ = acct.SetMainGame("ashford")
	_ = acct.SetMainGame("")
	if cur, _ := ge.CurrentGame(); cur.Save.Name != "corvale" {
		t.Errorf("after unmarking: got %q, want corvale", cur.Save.Name)
	}

	// The record names a save that is gone, and nothing is marked: the save
	// written most recently.
	if err := DeleteSave("corvale"); err != nil {
		t.Fatal(err)
	}
	backdate(t, AutosaveName, 72*time.Hour)
	if cur, _ := ge.CurrentGame(); cur.Save.Name != "ashford" || cur.Why != CurrentRecent {
		t.Errorf("a record with no save behind it: got %q (%s), want ashford, the newest", cur.Save.Name, cur.Why)
	}

	// A rename carries the records.
	_ = acct.SetMainGame("ashford")
	_ = acct.NoteGamePlayed("ashford")
	if err := RenameSave("ashford", "ashford_reborn"); err != nil {
		t.Fatal(err)
	}
	if err := acct.RenameGame("ashford", "ashford_reborn"); err != nil {
		t.Fatal(err)
	}
	if mainGame, last := acct.GameRecord(); mainGame != "ashford_reborn" || last != "ashford_reborn" {
		t.Errorf("after a rename the records are %q and %q", mainGame, last)
	}

	// A save that cannot be read is never the current game, even when it is
	// the newest file and the one on record.
	if err := os.WriteFile(filepath.Join(saveDirectory(), "torn.json"), []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	_ = acct.SetMainGame("torn")
	if cur, _ := ge.CurrentGame(); cur.Save.Name == "torn" {
		t.Error("a damaged save is the current game")
	}

	// Every save gone: no current game, whatever is on record.
	if err := WipeAllSaves(); err != nil {
		t.Fatal(err)
	}
	if cur, ok := ge.CurrentGame(); ok {
		t.Errorf("no saves, but the current game is %q", cur.Save.Name)
	}
	_ = acct.ForgetGame("")
	if mainGame, last := acct.GameRecord(); mainGame != "" || last != "" {
		t.Errorf("ForgetGame(\"\") left %q and %q on record", mainGame, last)
	}
}

// TestPickCurrentGameNewest: with nothing on record the newest save wins by
// the time it carries; the file's time breaks a tie, then the name.
func TestPickCurrentGameNewest(t *testing.T) {
	t0 := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	saves := []SaveInfo{
		{Name: "old", Timestamp: t0, ModTime: t0.Add(9 * time.Hour)},
		{Name: "new", Timestamp: t0.Add(time.Hour), ModTime: t0},
		{Name: "damaged", Timestamp: t0.Add(5 * time.Hour), Corrupt: true},
	}
	if cur, ok := PickCurrentGame(saves, "", ""); !ok || cur.Save.Name != "new" || cur.Why != CurrentRecent {
		t.Errorf("got %q (%s), want new", cur.Save.Name, cur.Why)
	}
	twins := []SaveInfo{
		{Name: "game", Timestamp: t0, ModTime: t0},
		{Name: "game-copy", Timestamp: t0, ModTime: t0.Add(time.Minute)},
	}
	if cur, _ := PickCurrentGame(twins, "", ""); cur.Save.Name != "game-copy" {
		t.Errorf("a tie on the save's time: got %q, want the file written last", cur.Save.Name)
	}
	twins[1].ModTime = t0
	if cur, _ := PickCurrentGame(twins, "", ""); cur.Save.Name != "game" {
		t.Errorf("a full tie: got %q, want the first by name", cur.Save.Name)
	}
	if _, ok := PickCurrentGame([]SaveInfo{{Name: "damaged", Corrupt: true}}, "damaged", "damaged"); ok {
		t.Error("a damaged save alone is a current game")
	}
	if _, ok := PickCurrentGame(nil, "x", "y"); ok {
		t.Error("no saves, but a current game")
	}
}

// TestGameRecordLivesInSettings: the records go to settings.json and never
// to account.json, whose bytes (and so its signature) stay as they were; a
// settings file from before the records reads as none; and the record
// survives a reload of the account.
func TestGameRecordLivesInSettings(t *testing.T) {
	ge, acct := menuAccount(t)
	accountFile := accountFilePath()
	settingsFile := filepath.Join(accountDir(acct.AccountID), settingsFileName)

	// A settings file as a build before the records wrote it.
	if err := acct.SetMotion(false); err != nil {
		t.Fatal(err)
	}
	if mainGame, last := acct.GameRecord(); mainGame != "" || last != "" {
		t.Fatalf("a settings file without the records reads %q and %q", mainGame, last)
	}

	startGame(t, ge, "ashford")
	// Flush whatever starting a game wrote to the account, then hold its bytes.
	if err := acct.Save(); err != nil {
		t.Fatal(err)
	}
	accountBytes, err := os.ReadFile(accountFile)
	if err != nil {
		t.Fatal(err)
	}

	if err := acct.SetMainGame("ashford"); err != nil {
		t.Fatal(err)
	}
	if err := acct.NoteGamePlayed("brackwater"); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(accountFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(accountBytes, after) {
		t.Error("recording the current game rewrote account.json")
	}
	raw, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatal(err)
	}
	var onDisk map[string]any
	if err := json.Unmarshal(raw, &onDisk); err != nil {
		t.Fatal(err)
	}
	if onDisk["main_game"] != "ashford" || onDisk["last_played"] != "brackwater" || onDisk["motion"] != "off" {
		t.Errorf("settings.json holds %v", onDisk)
	}

	// An unchanged record does not rewrite the file.
	fi1, _ := os.Stat(settingsFile)
	time.Sleep(15 * time.Millisecond)
	_ = acct.NoteGamePlayed("brackwater")
	_ = acct.SetMainGame("ashford")
	if fi2, _ := os.Stat(settingsFile); !fi2.ModTime().Equal(fi1.ModTime()) {
		t.Error("an unchanged record rewrote settings.json")
	}

	// The account read again from disk still verifies and has the records.
	loaded, found, err := LoadAccount()
	if err != nil || !found {
		t.Fatalf("LoadAccount: found=%v err=%v", found, err)
	}
	if loaded.Tampered {
		t.Error("the account reads as edited after the records were written")
	}
	if mainGame, last := loaded.GameRecord(); mainGame != "ashford" || last != "brackwater" {
		t.Errorf("after a reload the records are %q and %q", mainGame, last)
	}
}
