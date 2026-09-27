package game

import (
	"strings"
	"testing"
)

// TestBuildBuilding_QueuesSeveralStorageCopies: a capped building that isn't
// unique (storage, MaxCount 50 for the stash) queues copies one `build` at a
// time up to its cap, as `build stash N` always could. It used to refuse the
// second copy while the first was under construction, so a player checking
// in a few times a day got one storage copy per visit.
func TestBuildBuilding_QueuesSeveralStorageCopies(t *testing.T) {
	ge := NewGameEngine()
	ge.mu.Lock()
	for _, r := range []string{"wood", "food", "stone"} {
		ge.Resources.AddStorage(r, 1e6)
		ge.Resources.Add(r, 5e5)
	}
	ge.mu.Unlock()

	for i := 0; i < 3; i++ {
		if err := ge.BuildBuilding("stash"); err != nil {
			t.Fatalf("stash copy %d: %v", i+1, err)
		}
	}
	queued := 0
	for _, q := range ge.GetState().BuildQueue {
		if strings.EqualFold(q.Name, "stash") {
			queued++
		}
	}
	if queued != 3 {
		t.Errorf("queued stashes = %d, want 3", queued)
	}
}
