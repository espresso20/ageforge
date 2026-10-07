package game

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestSplashSeesNamedSaves is B14: the main menu looked only at the save
// named autosave, but a game is saved under the name it was started with.
// So the elite line never showed for a named game, and the menu preselected
// New game although there was a game to load. The menu now asks about every
// save in the account's slot. Runs in a temp data root.
func TestSplashSeesNamedSaves(t *testing.T) {
	isolateAccountDir(t)
	if exists, elite := SavesOnSplash(); exists || elite {
		t.Fatalf("an empty slot reads as exists=%v elite=%v", exists, elite)
	}

	ge, acct := devEngine(t, "Ada") // starts and saves a game named "run"
	if err := ge.SaveGame("run"); err != nil {
		t.Fatal(err)
	}
	// What the menu used to ask: nothing is there under the autosave name.
	if SaveExists(AutosaveName) {
		t.Fatal("precondition: a named game wrote an autosave file")
	}
	if _, elite := PeekSaveBadges(AutosaveName); elite {
		t.Fatal("precondition: the autosave name reads as elite")
	}
	if exists, elite := SavesOnSplash(); !exists || elite {
		t.Errorf("a slot with one plain named save reads as exists=%v elite=%v, want true and false", exists, elite)
	}

	// Sign the save with the forge master's key, as the easter egg does.
	path := filepath.Join(accountDir(acct.AccountID), "saves", "run.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	var sig string
	if err := json.Unmarshal(raw["_sig"], &sig); err != nil || sig == "" {
		t.Fatalf("the save has no signature: %v", err)
	}
	mac := hmac.New(sha256.New, []byte(forgeMasterKey))
	mac.Write([]byte(sig))
	raw["_proof"], _ = json.Marshal(hex.EncodeToString(mac.Sum(nil)))
	proved, _ := json.Marshal(raw)
	if err := os.WriteFile(path, proved, 0644); err != nil {
		t.Fatal(err)
	}

	if _, elite := PeekSaveBadges("run"); !elite {
		t.Fatal("precondition: the proved save does not read as elite")
	}
	if exists, elite := SavesOnSplash(); !exists || !elite {
		t.Errorf("a slot whose named save carries the proof reads as exists=%v elite=%v", exists, elite)
	}
}
