package game

import (
	"fmt"
	"time"
)

// MeetFactionForTest discovers civ key at opinion and rolls its trade deals,
// as first contact on an expedition followed by a tick would. A test hook for
// other packages (the ui panel tests and theme sweep); not reachable from
// play.
func (ge *GameEngine) MeetFactionForTest(key string, opinion int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if _, ok := ge.Diplomacy.factionDefs[key]; !ok {
		return fmt.Errorf("unknown civilization: %s", key)
	}
	ge.Diplomacy.DiscoverFaction(key)
	ge.Diplomacy.factions[key].Opinion = opinion
	ge.tickFactionDeals()
	return nil
}

// StepTicks runs n game ticks synchronously on the calling goroutine, as fast
// as the CPU allows, and returns the wall-clock time those ticks would have
// taken in the real timer-driven loop (the sum of each tick's interval at the
// current speed, tick_speed bonuses included).
//
// It exists for simulation and tests (the smoke autoplayer in package smoke),
// not for the UI: the game itself runs on Start's timer. Do not call it while
// Start is running on the same engine. Unlike Start it does not autosave and
// does not recover panics, so a crash inside a tick reaches the caller.
func (ge *GameEngine) StepTicks(n int) time.Duration {
	var elapsed time.Duration
	for i := 0; i < n; i++ {
		elapsed += ge.getTickInterval()
		ge.doTick()
	}
	return elapsed
}

// SimulateOffline applies the catch-up a save gets when it is loaded elapsed
// after it was written: the same applyOfflineProgress LoadGame runs (capped
// at MaxOfflineTime, at OfflineEfficiency), as if the game had been closed
// that long. For simulation and tests (the smoke suite's offline scenario);
// the game itself reaches it only through LoadGame. Takes the write lock.
func (ge *GameEngine) SimulateOffline(elapsed time.Duration) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.applyOfflineProgress(elapsed)
}

// ForceCatastropheForTest makes the current epoch's catastrophe pending now,
// as the dev console's /catastrophe does, with the same gates (the Iron
// epoch, nothing already pending). A test hook for other packages (the smoke
// suite's prestige scenario); not reachable from play.
func (ge *GameEngine) ForceCatastropheForTest() error {
	return ge.forceCatastrophe()
}
