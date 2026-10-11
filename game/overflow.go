package game

import (
	"fmt"

	"github.com/espresso20/ageforge/config"
)

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
//
// A store is a wall (the storage wall): what is left after the wonder is
// lost, and a price larger than a store cannot be bought (overStore in
// engine.go, checkPlanItem in plan.go). The age's wonder is the one price
// allowed over a store, because it is paid into this bank in rounds.
// Overflow used to pay the plan too, banking toward the next copy of every
// queued build; it does not any more, and a wonder queued in the plan takes
// overflow on the same terms as one that is not: only while wonder overflow
// is on. The end of this file reads the plan banks an old save carries.

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
	if ge.runFacts.Counts[config.BadgeEvWonderOverflow] == 0 {
		// Once a run: overflow has fed a wonder.
		ge.note(config.BadgeEvWonderOverflow, "")
	}
	if need-bank[res] <= 0.001 {
		ge.addLog("info", fmt.Sprintf("Overflow finished banking %s for %s.", ResourceName(res), def.Name))
		if ge.Buildings.IsWonderBankFull(w) {
			ge.addLog("success", fmt.Sprintf("The %s bank is full. %s", def.Name, ge.wonderNextStep(w)))
			ge.report(Event{Kind: config.BadgeEvWonderBanked, Subject: w, Attrs: map[string]float64{"away": boolFact(ge.away)}})
		}
	}
	return dep
}

// applyTickRates applies one tick of production. What a full store cannot
// hold goes to the age's wonder, up to what its bank still lacks, when wonder
// overflow is on; the rest is lost. Nothing is banked toward the plan: a
// price larger than a store cannot be bought, and the store is the limit.
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

// ===== Plan banks from before the storage wall =====
//
// Overflow used to pay the plan: what a full store discarded was banked
// toward the plan's next copies, so a price larger than a store could still
// be bought. It no longer is. What is left here reads the banks an old save
// carries and gives them back to the stores.

// planBankEpsilon is the least a plan bank tracks: a part of a price within
// it counts as paid, so float dust never leaves a copy waiting on a
// thousandth of a resource (the wonder bank uses the same margin).
const planBankEpsilon = 0.001

// splitBank splits the price of a banked item's next copy into what the bank
// covers (used: at most the price, per resource) and what is still due from
// the stores (parts within planBankEpsilon count as covered). Both are nil
// when the bank holds nothing.
func splitBank(price, bank map[string]float64) (due, used map[string]float64) {
	if len(bank) == 0 {
		return price, nil
	}
	due = make(map[string]float64, len(price))
	used = make(map[string]float64, len(bank))
	for res, c := range price {
		u := min(bank[res], c)
		if u > 0 {
			used[res] = u
		}
		if d := c - u; d > planBankEpsilon {
			due[res] = d
		}
	}
	return due, used
}

// returnPlanBank puts what item it banked back into the stores, up to each
// cap (the rest was overflow a cap cut off, and still doesn't fit), adds what
// went back to back (when non-nil) and empties the bank. Reports whether the
// bank held anything.
func (ge *GameEngine) returnPlanBank(it *PlanItem, back map[string]float64) bool {
	if len(it.Banked) == 0 {
		it.Banked = nil
		return false
	}
	for _, res := range sortedKeys(it.Banked) {
		before := ge.Resources.Get(res)
		if got := ge.Resources.Add(res, it.Banked[res]) - before; got > 0 && back != nil {
			back[res] += got
		}
	}
	it.Banked = nil
	return true
}

// returnPlanBanks empties every plan item's bank into the stores, up to the
// caps, and logs what went back (why says when: "for the advance"). The
// advance calls it before trimming the stockpiles, so a bank never carries
// more into the next age than the trim lets the stores carry.
func (ge *GameEngine) returnPlanBanks(why string) {
	back := map[string]float64{}
	held := false
	for i := range ge.plan {
		if ge.returnPlanBank(&ge.plan[i], back) {
			held = true
		}
	}
	if held {
		ge.logBankReturn("the plan's banks", why, back)
	}
}

// logBankReturn logs a bank going back to the stores: "Plan: the bank of 3
// Huts went back to the stores (1.2K wood).", or that nothing fit.
func (ge *GameEngine) logBankReturn(whose, why string, back map[string]float64) {
	when := ""
	if why != "" {
		when = " " + why
	}
	if len(back) == 0 {
		ge.addLog("info", fmt.Sprintf("Plan: %s went back to the stores%s, but they were full.", whose, when))
		return
	}
	ge.addLog("info", fmt.Sprintf("Plan: %s went back to the stores%s (%s).", whose, when, Amounts(back)))
}
