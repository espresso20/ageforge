package game

import (
	"fmt"

	"github.com/espresso20/ageforge/config"
)

// Deal items: `plan deal <civ> <n>` takes a civilization's trade deal (see
// deals.go) as soon as its price is there, like `diplomacy accept` queued.
// A deal is a one-off purchase with a fixed price, so it fits the plan the
// way a single build does:
//
//   - It has the priority of its place in the plan: while it waits for its
//     price it reserves it, and items below only see the rest.
//   - It is blocked, reserving nothing, while the goods would not fit in
//     their store or the civ won't trade (war, embargo, rivalry, hostility);
//     standing can recover within the round.
//   - It drops out once taken, or when the offer is gone: the civ's offers
//     rotated (an hour of live play, or an age advance) or it was taken by
//     hand. Offline catch-up does not rotate offers, so a deal planned
//     before leaving is still there to take while the player is away.
//
// The item names the offer by its ID, not its number, so a rotation can
// never make it take a different deal.

// PlanDeal is the deal item kind.
const PlanDeal = "deal"

// PlanAddDeal appends a deal item for offer n (1-based) of civ key.
func (ge *GameEngine) PlanAddDeal(key string, n int) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	def, ok := ge.Diplomacy.factionDefs[key]
	if !ok {
		return ge.Diplomacy.errUnknownCiv(key)
	}
	fs, ok := ge.Diplomacy.factions[key]
	if !ok || !fs.Discovered {
		return ge.Diplomacy.errUnknownCiv(key)
	}
	if err := ge.featureErr(config.FeatureDiplomacy); err != nil {
		return err
	}
	if fs.DealsFor != ge.age || n < 1 || n > len(fs.Deals) {
		return fmt.Errorf("There is no deal %d with the %s (they offer %d). Type diplomacy deals to see them.", n, def.Name, len(dealInfos(ge.rules, fs, ge.age, ge.Diplomacy.feeScale)))
	}
	d := fs.Deals[n-1]
	if d.Taken {
		return fmt.Errorf("Deal %d with the %s is already taken.", n, def.Name)
	}
	if why := dealBlocked(*fs); why != "" {
		return fmt.Errorf("The %s won't trade: they are %s.", def.Name, why)
	}
	for _, it := range ge.plan {
		if it.Kind == PlanDeal && it.Key == key && it.Deal == d.ID {
			return fmt.Errorf("The plan already takes that deal.")
		}
	}
	if len(ge.plan) >= MaxPlanItems {
		return errPlanFull()
	}
	ge.plan = append(ge.plan, PlanItem{Kind: PlanDeal, Key: key, Count: 1, Deal: d.ID})
	return nil
}

// planDealGone is why deal item it can never be taken ("" while it can).
func (ge *GameEngine) planDealGone(it PlanItem) string {
	fs, i := ge.findDeal(it.Key, it.Deal)
	switch {
	case fs == nil:
		return "unknown civilization"
	case i < 0 || fs.DealsFor != ge.age:
		return "the offer is gone"
	case fs.Deals[i].Taken:
		return "already taken"
	}
	return ""
}

// planDealCheck is the walk's verdict on deal item it (still there): its
// price and why it is blocked ("" when only the price is missing, if that).
func (ge *GameEngine) planDealCheck(it PlanItem) (d FactionDeal, blocked string) {
	fs, i := ge.findDeal(it.Key, it.Deal)
	d = fs.Deals[i]
	if lock, shut := ge.featureLock(config.FeatureDiplomacy); shut {
		tech, _ := ge.rules.Tech(lock.Tech)
		return d, "needs " + tech.Name + " researched"
	}
	blocked, _ = ge.dealProblem(fs, d)
	return d, blocked
}

// runPlanDeal takes deal item it if its price is free after the
// reservations above, or reserves the price. Reports whether it was taken.
func (ge *GameEngine) runPlanDeal(it PlanItem, reserved map[string]float64) bool {
	d, blocked := ge.planDealCheck(it)
	if blocked != "" {
		return false
	}
	cost := map[string]float64{d.Give: d.GiveAmt}
	if !ge.planCovers(cost, reserved) {
		reserved[d.Give] += d.GiveAmt
		return false
	}
	fs, i := ge.findDeal(it.Key, it.Deal)
	ge.takeDeal(ge.Diplomacy.factionDefs[it.Key], fs, i)
	return true
}

// planDealLabel is "deal with the Merchant Guild".
func (ge *GameEngine) planDealLabel(it PlanItem) string {
	name := it.Key
	if def, ok := ge.Diplomacy.factionDefs[it.Key]; ok {
		name = def.Name
	}
	return "deal with the " + name
}

// planDealView is deal item it for the UI; a waiting deal's price counts
// against reserved, as in the walk.
func (ge *GameEngine) planDealView(it PlanItem, reserved map[string]float64) PlanItemView {
	v := PlanItemView{Kind: it.Kind, Key: it.Key, Count: it.Count, Name: ge.planDealLabel(it)}
	if gone := ge.planDealGone(it); gone != "" {
		v.Status, v.Note = PlanStatusBlocked, gone
		return v
	}
	d, blocked := ge.planDealCheck(it)
	v.Name += " (" + dealLogTerms(d) + ")"
	v.Cost = map[string]float64{d.Give: d.GiveAmt}
	if blocked != "" {
		v.Status, v.Note = PlanStatusBlocked, blocked
		return v
	}
	free := ge.planFree(d.Give, reserved)
	reserved[d.Give] += d.GiveAmt
	if free >= d.GiveAmt {
		v.Status, v.Progress = PlanStatusReady, 1
		return v
	}
	v.Status, v.Short = PlanStatusWaiting, d.Give
	v.Progress = max(0, free/d.GiveAmt)
	return v
}
