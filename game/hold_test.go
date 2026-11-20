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

// A save written while held is stamped when it is written and carries the run
// start with the time held already taken out: the time held is neither offline
// time nor play time on the next load.
func TestHoldIsNeitherOfflineNorPlayTimeInASave(t *testing.T) {
	defer SetDataDirForTest(t.TempDir())()
	ge := NewGameEngine()
	start := time.Now().Add(-time.Hour)
	ge.Stats.GameStarted = start
	ge.Hold()
	time.Sleep(150 * time.Millisecond)
	before := time.Now()
	if err := ge.SaveGame("held"); err != nil {
		t.Fatal(err)
	}
	other := NewGameEngine()
	if err := other.LoadGame("held"); err != nil {
		t.Fatal(err)
	}
	if other.SavedAt().Before(before) {
		t.Errorf("the save is stamped %v, before the held time ended", other.SavedAt())
	}
	if d := other.Stats.GameStarted.Sub(start); d < 100*time.Millisecond {
		t.Errorf("saved run start moved %v, the held time was left in play time", d)
	}
}
