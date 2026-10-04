package game

import (
	"time"

	"github.com/espresso20/ageforge/config"
)

// MeetFactionForTest discovers civ key at opinion and rolls its trade deals,
// as first contact on an expedition followed by a tick would. A test hook for
// other packages (the ui panel tests and theme sweep); not reachable from
// play.
func (ge *GameEngine) MeetFactionForTest(key string, opinion int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if _, ok := ge.Diplomacy.factionDefs[key]; !ok {
		return ge.Diplomacy.errUnknownCiv(key)
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

// SetGarrisonForTest moves the engine to age and gives it n soldiers (the
// soldiers resource unlocked, its storage raised to fit), so a garrison
// blunts raids. A test hook for other packages (the Army panel, catastrophe
// modal and theme sweep tests); not reachable from play.
func (ge *GameEngine) SetGarrisonForTest(age string, n float64) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.age = age
	ge.currentEpoch = config.EpochForAge(age)
	ge.Resources.UnlockResource("soldiers")
	if short := n - ge.Resources.GetStorage("soldiers"); short > 0 {
		ge.Resources.AddStorage("soldiers", short)
	}
	ge.Resources.Add("soldiers", n-ge.Resources.Get("soldiers"))
}

// GrantTechsForTest marks techs as researched, as finishing each would, and
// recalculates the rates: with no keys, every tech up to the current age.
// A test hook for other packages (the panels that list capped bonuses); not
// reachable from play.
func (ge *GameEngine) GrantTechsForTest(keys ...string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if len(keys) == 0 {
		at := ageOrders()[ge.age]
		for _, key := range ge.Research.order {
			if ageOrders()[ge.Research.defs[key].Age] <= at {
				keys = append(keys, key)
			}
		}
	}
	for _, key := range keys {
		if _, ok := ge.Research.defs[key]; ok {
			ge.Research.researched[key] = true
		}
	}
	ge.Research.rebuildBonuses()
	ge.recalculateRates()
	ge.recalculateTickSpeed()
}

// SetLegacyBonusForTest marks epochs as succumbed in, as a Succumb in each
// would: their legacy bonuses and Ancient Knowledge apply at once. A test
// hook for other packages (the panels that show research speed and legacy
// bonuses); not reachable from play.
func (ge *GameEngine) SetLegacyBonusForTest(epochs ...string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	for _, ep := range epochs {
		if _, ok := config.EpochByKey()[ep]; ok && !ge.legacyBonuses[ep] {
			ge.legacyBonuses[ep] = true
			for res, mult := range config.LegacyBonusForEpoch(ep) {
				ge.permanentBonuses[res+"_rate"] += mult
			}
		}
	}
	ge.recalculateRates()
}
