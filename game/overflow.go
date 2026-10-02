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
//
// Overflow pays the plan (Pacing v2, away-proofing): what is left after the
// wonder, which would be lost, goes into the banks of the build plan's items
// instead, in plan order, each up to what its next copy still lacks (see
// bankPlanOverflow, and plan.go for how a bank pays). It is always on: like
// the wonder's, it only takes what a cap was about to discard, and only for
// copies the player queued. It needs no switch of its own: a player who
// wants nothing banked plans nothing. Live play logs nothing per tick (the
// Plan panel shows each bank); the welcome-back summary totals it.

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
// off into the wonder when overflow is on, and what is left into the plan.
func (ge *GameEngine) applyTickRates() {
	w := ge.overflowWonder()
	if w == "" && len(ge.plan) == 0 {
		ge.Resources.ApplyRates()
		return
	}
	losses := ge.overflowScratch[:0]
	ge.Resources.ApplyRatesCapped(func(res string, lost float64) {
		if lost -= ge.bankOverflow(w, res, lost); lost > 0 {
			losses = append(losses, overflowLoss{res: res, amount: lost})
		}
	})
	ge.overflowScratch = losses
	ge.bankPlanOverflow(losses, nil)
}

// ===== Overflow pays the plan =====

// planBankEpsilon is the least a plan bank tracks: a part of a price within
// it counts as paid, so float dust never leaves a copy waiting on a
// thousandth of a resource (the wonder bank uses the same margin).
const planBankEpsilon = 0.001

// overflowLoss is what a cap cut off of one resource in one tick or offline
// step, after the wonder took its share.
type overflowLoss struct {
	res    string
	amount float64
}

// bankPlanOverflow puts what the caps cut off (after the wonder) into the
// banks of the plan's build items, in plan order: each item takes what the
// price of its next copy still lacks of each resource, the first item first,
// and what no item needs is lost as before. An item banks only if it could
// start in this age: a build of this age (not the next age's, which waits for
// the advance, so nothing is carried past the advance's stockpile trim), not
// a wonder (it has its own bank, which wonder overflow fills first) and not
// at its MaxCount. It adds what it banked to into, by resource, when into is
// non-nil. Cheap when nothing overflowed or nothing is planned, so it runs
// every tick. Must be called with the write lock held.
func (ge *GameEngine) bankPlanOverflow(losses []overflowLoss, into map[string]float64) {
	if len(losses) == 0 || len(ge.plan) == 0 {
		return
	}
	open := len(losses)      // losses with something left to place
	var above map[string]int // capped buildings: copies the items above will take
	for i := range ge.plan {
		it := &ge.plan[i]
		if it.Kind != PlanBuild || it.Count <= 0 {
			continue
		}
		def, ok := ge.Buildings.defs[it.Key]
		if !ok || def.Category == "wonder" || (def.RequiredAge != "" && def.RequiredAge != ge.age) || !ge.Buildings.IsUnlocked(it.Key) {
			continue
		}
		if def.MaxCount > 0 {
			planned := above[it.Key]
			if above == nil {
				above = map[string]int{}
			}
			above[it.Key] += it.Count
			if ge.Buildings.GetCount(it.Key)+ge.Buildings.GetQueueCount(it.Key, ge.buildQueue)+planned >= def.MaxCount {
				continue
			}
		}
		needs := false
		for _, l := range losses {
			if l.amount > 0 && def.BaseCost[l.res] > 0 {
				needs = true
				break
			}
		}
		if !needs {
			continue
		}
		cost, _ := ge.Buildings.BuildBatchCost(it.Key, 1, ge.buildQueue)
		for j := range losses {
			l := &losses[j]
			if l.amount <= 0 {
				continue
			}
			want := cost[l.res] - it.Banked[l.res]
			if want <= planBankEpsilon {
				continue
			}
			dep := min(l.amount, want)
			if it.Banked == nil {
				it.Banked = make(map[string]float64, len(cost))
			}
			it.Banked[l.res] += dep
			if into != nil {
				into[l.res] += dep
			}
			if l.amount -= dep; l.amount <= 0 {
				if open--; open == 0 {
					return
				}
			}
		}
	}
}

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
