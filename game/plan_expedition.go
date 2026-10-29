package game

import (
	"github.com/espresso20/ageforge/vibe"
)

// Expedition items: `plan expedition <mission>` sends a mission off the
// mission board as soon as it can go, like the `expedition` command queued.
// It runs through the same launch path (launchExpeditionLocked), so the
// party, the price and the board's slot are the command's own.
//
//   - The item names an offer by its title. It sends the first offer on the
//     board with that title.
//   - While the town lacks the price it waits and reserves it, like a build.
//     A price the income won't bring in within a day reserves nothing.
//   - It is blocked, reserving nothing, while the party of soldiers is not at
//     home: that is the army's doing, not a price. Note says so.
//   - It drops out once sent, or when no offer of that title is on the
//     board any more (it expired, was sent by hand, or the age moved on).
//     Offline catch-up runs the same walk, tick by tick with the board.
//
// Count is 1.

// PlanAddExpedition appends an expedition item for the offer on the board
// called title (as the board spells it).
func (ge *GameEngine) PlanAddExpedition(title string) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.ensureMissionBoardLocked()
	if ge.planOfferNumber(title) == 0 {
		return vibe.Errorf("game.plan_expedition.PlanAddExpedition.01", "No offer called %q is on the board.", title)
	}
	if plannedCount(ge.plan, PlanExpedition, title) > 0 {
		return vibe.Errorf("game.plan_expedition.PlanAddExpedition.02", "The plan already sends that expedition.")
	}
	if planLoad(ge.plan) >= MaxPlanItems {
		return errPlanFull()
	}
	ge.plan = append(ge.plan, PlanItem{Kind: PlanExpedition, Key: title, Count: 1})
	ge.notePlanSize()
	return nil
}

// planOfferNumber is the board number (1-based) of the first offer called
// title, 0 when there is none.
func (ge *GameEngine) planOfferNumber(title string) int {
	for i, slot := range ge.Military.board.Slots {
		if slot.Offer != nil && slot.Offer.Title == title {
			return i + 1
		}
	}
	return 0
}

// planExpeditionGone is why expedition item it can never be sent ("" while
// it can).
func (ge *GameEngine) planExpeditionGone(it PlanItem) string {
	if ge.planOfferNumber(it.Key) == 0 {
		return vibe.Line("game.plan_expedition.planExpeditionGone.01", "the offer is gone")
	}
	return ""
}

// planExpeditionCheck is the walk's verdict on expedition item it (still
// there): its price in materials, and why it is blocked.
func (ge *GameEngine) planExpeditionCheck(it PlanItem) planCheck {
	n := ge.planOfferNumber(it.Key)
	if n == 0 {
		return planCheck{blocked: ge.planExpeditionGone(it)}
	}
	offer := *ge.Military.board.Slots[n-1].Offer
	cost := ge.missionCostLocked(offer)
	block := ge.offerBlockLocked(cost)
	if block.party > 0 {
		return planCheck{blocked: block.reason()}
	}
	delete(cost, "soldiers")
	chk := planCheck{cost: cost}
	if res := ge.planUnfunded(cost); res != "" {
		chk.blocked = ge.incomeShortfall(res)
		return chk
	}
	chk.reserve = true
	return chk
}

// launchPlanExpedition sends expedition item it through the command's path.
func (ge *GameEngine) launchPlanExpedition(it PlanItem) error {
	n := ge.planOfferNumber(it.Key)
	if n == 0 {
		return vibe.Errorf("game.plan_expedition.launchPlanExpedition.01", "No offer called %q is on the board.", it.Key)
	}
	return ge.launchExpeditionLocked(n)
}

// PlanAdjust changes the count of build or upgrade item n (1-based) by delta
// and returns the new count. Raising it stops at what the building's limit or
// the copies held leave; lowering it stops at 1 (remove the item instead).
func (ge *GameEngine) PlanAdjust(n, delta int) (int, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	if n < 1 || n > len(ge.plan) {
		return 0, ge.planIndexErr(n)
	}
	it := &ge.plan[n-1]
	if it.Kind != PlanBuild && it.Kind != PlanUpgrade {
		return 0, vibe.Errorf("game.plan_expedition.PlanAdjust.01", "Only builds and upgrades have a count.")
	}
	if delta < 0 {
		next := max(it.Count+delta, 1)
		if next == it.Count {
			return 0, vibe.Errorf("game.plan_expedition.PlanAdjust.02", "The count is already 1. Remove the item to drop it.")
		}
		it.Count = next
		return next, nil
	}
	planned := plannedCount(ge.plan, it.Kind, it.Key)
	room := maxPlanCount - it.Count
	if it.Kind == PlanBuild {
		def, ok := ge.Buildings.defs[it.Key]
		if !ok {
			return 0, ge.unknownBuildingErr(it.Key)
		}
		room = min(room, ge.planRoom(def, planned))
	} else {
		newDef, pending := ge.Buildings.UpgradeTarget(it.Key)
		if !pending {
			return 0, vibe.Errorf("game.plan_expedition.PlanAdjust.03", "That upgrade is no longer available.")
		}
		room = min(room, ge.Buildings.GetCount(it.Key)-planned, ge.planRoom(newDef, planned))
	}
	add := min(delta, room)
	if add <= 0 {
		return 0, vibe.Errorf("game.plan_expedition.PlanAdjust.04", "That is the most the plan can take.")
	}
	it.Count += add
	ge.logPlanAddLocked(*it, add)
	return it.Count, nil
}
