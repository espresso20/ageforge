package game

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestPacingNoticeOnceForOlderSaves: a save written before the one-week curve
// (no pacing_week field) still passes its signature check, says once on load
// that the ages ahead are longer, and the save it writes next carries the
// marker, so loading that is silent. A current save never gets the notice.
func TestPacingNoticeOnceForOlderSaves(t *testing.T) {
	isolateAccountDir(t)
	notices := func(ge *GameEngine) int {
		n := 0
		for _, l := range ge.log {
			if l.Message == pacingNotice {
				n++
			}
		}
		return n
	}

	ge := newSeededEngine(1)
	ge.StepTicks(20)
	if err := ge.SaveGame("current-curve"); err != nil {
		t.Fatal(err)
	}
	if !readSaveFromDisk(t, "current-curve").PacingWeek {
		t.Fatal("a new save does not carry pacing_week")
	}
	fresh := NewGameEngine()
	if err := fresh.LoadGame("current-curve"); err != nil {
		t.Fatal(err)
	}
	if n := notices(fresh); n != 0 || fresh.cheaterBadge {
		t.Errorf("a current save: %d notices, tamper flag %v", n, fresh.cheaterBadge)
	}

	// An older save: the same state as the old build wrote and signed it.
	ge.mu.RLock()
	old := ge.buildSaveSnapshot()
	ge.mu.RUnlock()
	old.PacingWeek = false
	old.Signature = signSave(old, saveHMACKey)
	data, err := json.MarshalIndent(old, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "pacing_week") {
		t.Fatal("an unmarked save still writes pacing_week: the field must be omitempty")
	}
	if err := os.MkdirAll(saveDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(savePath("old-curve"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	loaded := NewGameEngine()
	if err := loaded.LoadGame("old-curve"); err != nil {
		t.Fatal(err)
	}
	if loaded.cheaterBadge {
		t.Error("an older save failed its signature check")
	}
	if n := notices(loaded); n != 1 {
		t.Errorf("an older save logged the pacing notice %d times, want once", n)
	}
	if err := loaded.SaveGame("old-curve"); err != nil {
		t.Fatal(err)
	}
	if !readSaveFromDisk(t, "old-curve").PacingWeek {
		t.Error("the older save's next save does not carry pacing_week")
	}
	again := NewGameEngine()
	if err := again.LoadGame("old-curve"); err != nil {
		t.Fatal(err)
	}
	if n := notices(again); n != 0 {
		t.Errorf("the notice came back on the second load (%d)", n)
	}
}
