package game

import (
	"os"
	"strings"
	"testing"
)

// writeSignedSave writes save, signed, under name in the save folder.
func writeSignedSave(t *testing.T, name string, save GameSave) {
	t.Helper()
	save.Signature = signSave(save)
	noErr(t, os.MkdirAll(saveDirectory(), 0o755))
	writeRawSave(t, name, save)
}

// A save the game writes carries the current format, verifies, and loads
// with no tamper flag.
func TestSaveRoundTripsAndVerifies(t *testing.T) {
	isolateAccountDir(t)
	ge := newSeededEngine(3)
	ge.StepTicks(50)
	noErr(t, ge.SaveGame("fresh"))

	saved := readSaveFromDisk(t, "fresh")
	if saved.Format != saveFormat {
		t.Errorf("the save carries format %d, want %d", saved.Format, saveFormat)
	}
	if sigValid, _ := verifySave(&saved); !sigValid || saved.Signature == "" {
		t.Error("a save the game wrote does not verify")
	}

	loaded := NewGameEngine()
	noErr(t, loaded.LoadGame("fresh"))
	if loaded.cheaterBadge {
		t.Error("loading the game's own save raised the tamper flag")
	}
	if loaded.tick != ge.tick || loaded.age != ge.age {
		t.Errorf("loaded tick %d in %s, saved tick %d in %s", loaded.tick, loaded.age, ge.tick, ge.age)
	}
}

// A save written before 4.0 has no format. It is refused with one plain
// line, not loaded and not crashed on, and the game in play is left alone.
// The load list still shows it, so the player can see what the file is.
func TestASaveFromBeforeFourIsRefused(t *testing.T) {
	isolateAccountDir(t)
	src := newSeededEngine(4)
	src.StepTicks(20)
	src.mu.RLock()
	old := src.buildSaveSnapshot()
	src.mu.RUnlock()
	old.Format = 0
	writeSignedSave(t, "old", old)

	ge := newSeededEngine(5)
	before := ge.age
	err := ge.LoadGame("old")
	if err == nil || !strings.Contains(err.Error(), "saved before 4.0") {
		t.Fatalf("loading a save from before 4.0: %v, want the refusal", err)
	}
	if ge.age != before || ge.tick != 0 || ge.activeSaveName != "" {
		t.Errorf("a refused load changed the game: age %s, tick %d, save %q", ge.age, ge.tick, ge.activeSaveName)
	}
	if _, err := ViewSave("old"); err == nil {
		t.Error("the main menu's look at a save from before 4.0 was not refused")
	}
	info := findDetail(t, "old")
	if info.Corrupt {
		t.Error("a save from before 4.0 is listed as corrupt")
	}
}

// A file that holds no save at all (empty JSON) has no format either: it is
// refused like a save from before 4.0, not crashed on.
func TestAnEmptySaveIsRefusedNotCrashedOn(t *testing.T) {
	isolateAccountDir(t)
	noErr(t, os.MkdirAll(saveDirectory(), 0o755))
	noErr(t, os.WriteFile(savePath("empty"), []byte("{}"), 0o644))
	if err := NewGameEngine().LoadGame("empty"); err == nil {
		t.Error("an empty save loaded")
	}
}

// A load unlocks every resource the loaded age and the ages before it
// unlock, whatever the save's own list says: a content patch that adds a
// resource to an age reaches saves already past it.
func TestLoadUnlocksWhatTheAgeAlreadyUnlocked(t *testing.T) {
	isolateAccountDir(t)
	src := newSeededEngine(6)
	noErr(t, src.EnterAgeForTest("stone_age"))
	noErr(t, src.EnterAgeForTest("bronze_age"))
	src.mu.RLock()
	save := src.buildSaveSnapshot()
	src.mu.RUnlock()

	var unlocks []string
	for _, def := range src.progress.ages {
		if src.progress.ageIndex[def.Key] <= src.progress.ageIndex["bronze_age"] {
			unlocks = append(unlocks, def.UnlockResources...)
		}
	}
	if len(unlocks) == 0 {
		t.Fatal("setup: no resource is unlocked by the Bronze Age or before")
	}
	save.Unlocked.Resources = nil
	writeSignedSave(t, "patched", save)

	loaded := NewGameEngine()
	noErr(t, loaded.LoadGame("patched"))
	for _, key := range unlocks {
		if !loaded.Resources.IsUnlocked(key) {
			t.Errorf("%s is locked after a load in the Bronze Age", key)
		}
	}
}
