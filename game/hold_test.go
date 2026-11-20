package game

import (
	"testing"
	"time"
)

// A held engine runs no tick; the time held is not time played.
func TestHoldStopsTheLoopAndPlayTime(t *testing.T) {
	ge := NewGameEngine()
	ge.Stats.GameStarted = time.Now().Add(-time.Hour)
	ge.Hold()
	go ge.Start()
	defer ge.Stop()
	time.Sleep(BaseTickInterval + 400*time.Millisecond)
	ge.mu.RLock()
	ticks := ge.tick
	ge.mu.RUnlock()
	if ticks != 0 {
		t.Fatalf("%d ticks ran while held", ticks)
	}
	ge.Release()
	if ge.Held() {
		t.Fatal("still held after Release")
	}
	played := ge.GetState().Stats.PlayTime
	if played > time.Hour+500*time.Millisecond {
		t.Errorf("played %v, the %v held counted", played, BaseTickInterval)
	}
}

// A save written while held is stamped when it is written, so the time held
// is never offline time on the next load.
func TestHoldIsNotOfflineTime(t *testing.T) {
	defer SetDataDirForTest(t.TempDir())()
	ge := NewGameEngine()
	ge.Hold()
	time.Sleep(150 * time.Millisecond)
	before := time.Now()
	if err := ge.SaveGame("held"); err != nil {
		t.Fatal(err)
	}
	if ge.SavedAt().Before(before) {
		t.Errorf("saved at %v, before the save began %v", ge.SavedAt(), before)
	}
	other := NewGameEngine()
	if err := other.LoadGame("held"); err != nil {
		t.Fatal(err)
	}
	if other.SavedAt().Before(before) {
		t.Errorf("the save is stamped %v, before the held time ended", other.SavedAt())
	}
}
