package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// current_game.go says which of an account's saves is "the current game":
// the one the main menu draws behind itself and Continue opens.
//
// Three things can name it, in this order:
//
//  1. The main game: a save the player marked in the Load Game browser. It
//     stays the current game until it is unmarked or deleted.
//  2. The game last played: the save last loaded or saved on the account.
//  3. With neither on record (an account from before the records, or one
//     whose record names a save that is gone): the save written most
//     recently, by the time the save itself carries; the file's time breaks
//     a tie.
//
// The two records live in the account's settings file (account_settings.go),
// not in account.json, whose signature must stay what older builds verify.
// A save that cannot be read is never the current game: it cannot be opened.

// Why a save is the current game.
const (
	CurrentMain   = "main"   // the player marked it as the main game
	CurrentLast   = "last"   // it is the game last played
	CurrentRecent = "recent" // nothing on record names a save: the newest one
)

// CurrentGame is the save Continue opens, and why it is that one.
type CurrentGame struct {
	Save SaveInfo
	Why  string
}

// PickCurrentGame chooses the current game among saves by the rule above.
// ok is false when no save can be opened.
func PickCurrentGame(saves []SaveInfo, mainGame, lastPlayed string) (cur CurrentGame, ok bool) {
	var usable []SaveInfo
	for _, s := range saves {
		if !s.Corrupt {
			usable = append(usable, s)
		}
	}
	if len(usable) == 0 {
		return CurrentGame{}, false
	}
	named := func(name string) (SaveInfo, bool) {
		for _, s := range usable {
			if name != "" && s.Name == name {
				return s, true
			}
		}
		return SaveInfo{}, false
	}
	if s, found := named(mainGame); found {
		return CurrentGame{Save: s, Why: CurrentMain}, true
	}
	if s, found := named(lastPlayed); found {
		return CurrentGame{Save: s, Why: CurrentLast}, true
	}
	sort.SliceStable(usable, func(i, j int) bool {
		a, b := usable[i], usable[j]
		if !a.Timestamp.Equal(b.Timestamp) {
			return a.Timestamp.After(b.Timestamp)
		}
		if !a.ModTime.Equal(b.ModTime) {
			return a.ModTime.After(b.ModTime)
		}
		return a.Name < b.Name
	})
	return CurrentGame{Save: usable[0], Why: CurrentRecent}, true
}

// GameRecord returns the account's two records: the save marked as the main
// game and the save last played, "" for either that is not set. A record
// may name a save that is gone; PickCurrentGame ignores it.
func (a *Account) GameRecord() (mainGame, lastPlayed string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.settingsLocked()
	return s.MainGame, s.LastPlayed
}

// NoteGamePlayed records name as the game last played. It writes the
// settings file only when the record changes, so an autosave costs nothing.
func (a *Account) NoteGamePlayed(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.settingsLocked()
	if name == "" || s.LastPlayed == name {
		return nil
	}
	s.LastPlayed = name
	return a.saveSettingsLocked()
}

// SetMainGame marks name as the main game, or with "" unmarks it.
func (a *Account) SetMainGame(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.settingsLocked()
	if s.MainGame == name {
		return nil
	}
	s.MainGame = name
	return a.saveSettingsLocked()
}

// RenameGame carries the records over a rename: a record that named the
// save by its old name names it by the new one.
func (a *Account) RenameGame(oldName, newName string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.settingsLocked()
	if oldName == "" || s.MainGame != oldName && s.LastPlayed != oldName {
		return nil
	}
	if s.MainGame == oldName {
		s.MainGame = newName
	}
	if s.LastPlayed == oldName {
		s.LastPlayed = newName
	}
	return a.saveSettingsLocked()
}

// ForgetGame drops the records that name a deleted save, so a later save
// of the same name is not taken for it. With "" it drops both records
// (every save was deleted).
func (a *Account) ForgetGame(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.settingsLocked()
	changed := false
	if s.MainGame != "" && (name == "" || s.MainGame == name) {
		s.MainGame, changed = "", true
	}
	if s.LastPlayed != "" && (name == "" || s.LastPlayed == name) {
		s.LastPlayed, changed = "", true
	}
	if !changed {
		return nil
	}
	return a.saveSettingsLocked()
}

// CurrentGame is the held account's current game among the saves in its
// slot: the one Continue opens. Without an account nothing is on record,
// and it is the newest save. ok is false when there is no save to open.
// It reads; it never writes.
func (ge *GameEngine) CurrentGame() (cur CurrentGame, ok bool) {
	saves, err := ListSaveDetails()
	if err != nil {
		return CurrentGame{}, false
	}
	mainGame, lastPlayed := "", ""
	if acct := ge.Account(); acct != nil {
		mainGame, lastPlayed = acct.GameRecord()
	}
	return PickCurrentGame(saves, mainGame, lastPlayed)
}

// WipeSavesByID deletes every save in the slot of the account with this ID
// and returns how many it deleted. The account itself is kept: its themes,
// lifetime stats and badges. It is the Accounts panel's "delete all saves",
// which acts on the selected account, current or not.
//
// The account's records of its current game go with the saves, so a later
// save that happens to take a deleted one's name is not mistaken for it.
// When the account is the one held, the run in memory was one of its saves:
// the engine is reset to a fresh game, as it would be at a launch.
func (ge *GameEngine) WipeSavesByID(id string) (int, error) {
	if !validAccountID(id) {
		return 0, fmt.Errorf("there is no account with the ID %q", id)
	}
	dir := filepath.Join(accountDir(id), "saves")
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return 0, fmt.Errorf("could not read the save folder: %w", err)
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
			return n, fmt.Errorf("could not delete the save file: %w", err)
		}
		n++
	}
	if held := ge.Account(); held != nil && held.AccountID == id {
		_ = held.ForgetGame("")
		ge.Reset()
		return n, nil
	}
	forgetGamesInSlot(id)
	return n, nil
}

// forgetGamesInSlot drops the current-game records from the settings file
// of an account that is not held (the held one goes through ForgetGame, so
// what it has in memory stays true). Best effort: a record with no save
// behind it is ignored anyway.
func forgetGamesInSlot(id string) {
	path := filepath.Join(accountDir(id), settingsFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var s accountSettings
	if json.Unmarshal(data, &s) != nil || s.MainGame == "" && s.LastPlayed == "" {
		return
	}
	s.MainGame, s.LastPlayed, s.Version = "", "", settingsFileVersion
	out, err := json.MarshalIndent(&s, "", "  ")
	if err != nil {
		return
	}
	tmp := path + ".tmp"
	if os.WriteFile(tmp, out, 0644) == nil {
		_ = os.Rename(tmp, path)
	}
}
