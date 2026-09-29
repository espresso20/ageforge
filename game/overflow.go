package game

import "fmt"

// Wonder overflow: production a storage cap would cut off goes into the
// current age's wonder bank instead, when the wonder still needs that
// resource, up to what it still needs. It is on by default and the player can
// turn it off (`wonder overflow off`). It runs every tick and during offline
// catch-up. Overflow never touches what the player holds: only what the cap
// was about to discard.
//
// Logging is sparing and follows only from state, so a loaded game logs what
// the saved one would have: one line when overflow finishes a resource's part
// of the bank, the usual "bank is full" line when the bank fills, and a total
// in the welcome-back summary.

// SetWonderOverflow turns wonder overflow on or off.
func (ge *GameEngine) SetWonderOverflow(on bool) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.wonderOverflowOff = !on
}

// WonderOverflow reports whether wonder overflow is on.
func (ge *GameEngine) WonderOverflow() bool {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return !ge.wonderOverflowOff
}

// overflowWonder is the wonder overflow banks into right now ("" for none):
// the current age's wonder, while it is unbuilt and not under construction.
func (ge *GameEngine) overflowWonder() string {
	if ge.wonderOverflowOff {
		return ""
	}
	w := ge.progress.WonderForAge(ge.age)
	if w == "" || !ge.Buildings.IsUnlocked(w) || ge.Buildings.GetCount(w) > 0 {
		return ""
	}
	for _, q := range ge.buildQueue {
		if q.BuildingKey == w {
			return ""
		}
	}
	return w
}

// bankOverflow deposits lost (what res's cap just cut off) into wonder w's
// bank, up to what w still needs of res, and returns the amount banked.
// Must be called with the write lock held.
func (ge *GameEngine) bankOverflow(w, res string, lost float64) float64 {
	if w == "" || !(lost > 0) {
		return 0
	}
	def := ge.Buildings.defs[w]
	need, ok := def.BaseCost[res]
	if !ok {
		return 0
	}
	bank := ge.Buildings.wonderBanks[w]
	left := need - bank[res]
	if left <= 0.001 {
		return 0
	}
	dep := min(lost, left)
	if bank == nil {
		bank = make(map[string]float64)
		ge.Buildings.wonderBanks[w] = bank
	}
	bank[res] += dep
	if need-bank[res] <= 0.001 {
		ge.addLog("info", fmt.Sprintf("Overflow finished banking %s for %s.", ResourceName(res), def.Name))
		if ge.Buildings.IsWonderBankFull(w) {
			ge.addLog("success", fmt.Sprintf("The %s bank is full. Type 'build %s' to start construction.", def.Name, w))
		}
	}
	return dep
}

// applyTickRates applies one tick of production, banking what the caps cut
// off into the wonder when overflow is on.
func (ge *GameEngine) applyTickRates() {
	w := ge.overflowWonder()
	if w == "" {
		ge.Resources.ApplyRates()
		return
	}
	ge.Resources.ApplyRatesCapped(func(res string, lost float64) {
		ge.bankOverflow(w, res, lost)
	})
}
