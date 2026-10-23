package game

import "fmt"

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

// returnPlanBank puts what item it banked back into the stores, adds what
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

// returnPlanBanks empties every plan item's bank into the stores and logs what went back (why says when: "for the advance"). The
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
