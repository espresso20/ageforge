package game

import "sort"

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
