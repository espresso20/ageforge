package game

import (
	"fmt"
	"math"
)

// Trade items: `plan trade <give> <get> [amount]` sells give for get at the
// market as give comes in, until amount of get has been bought (no amount:
// until removed). Note the amount counts the get side here, while
// `trade <give> <get> <amount>` counts the give side. Internally the item
// stores give as Key and get as To. It is how a plan gets the resources an
// age buys rather than makes (stone after the Bronze Age, the Industrial
// Age's iron) while the player is away, instead of one trade per visit.
//
// Rules:
//   - Like any item, it has the priority of its place in the plan: it holds
//     back the give resource it will sell (what the items above leave free, up to
//     what it still wants to buy), and items below it only see the rest.
//   - It sells once the market has recovered from its last sale (the pair's
//     supply pressure under planTradeRecovered), not every tick: every trade
//     raises the pressure by the same step whatever its size, so selling in
//     one lump a minute or so apart keeps the rate within a percent of the
//     market's, where selling every tick would cost up to 30%.
//   - It never buys more than the get resource's store has room for (the rest would be
//     lost), at the market's current rate (parity less the fee and the
//     pressure), and only with a trade building standing, like `trade`.
//   - It drops out when it has bought its amount, or when the pair no longer
//     trades (an age advance) or a resource is locked.

// planTradeRecovered is the supply pressure under which a trade item sells
// again (a rate within 0.6% of the market's).
const planTradeRecovered = 0.02

// PlanAddTrade appends a trade item: sell give for get as give comes in,
// until amount of get is bought (0: until removed). The pair must trade at
// the market in this age.
func (ge *GameEngine) PlanAddTrade(give, get string, amount float64) error {
	from, to := give, get
	if !(amount >= 0) || math.IsInf(amount, 0) {
		return fmt.Errorf("The amount (how much %s to buy) must be a positive number.", ResourceName(to))
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	it := PlanItem{Kind: PlanTrade, Key: from, To: to, Count: 1, Amount: amount}
	if reason := ge.planTradeInvalid(it); reason != "" {
		return fmt.Errorf("Can't plan a trade of %s for %s: %s.", ResourceName(from), ResourceName(to), reason)
	}
	for _, p := range ge.plan {
		if p.Kind == PlanTrade && p.Key == from && p.To == to {
			return fmt.Errorf("The plan already trades %s for %s. Remove that item first.", ResourceName(from), ResourceName(to))
		}
	}
	if len(ge.plan) >= MaxPlanItems {
		return errPlanFull()
	}
	ge.plan = append(ge.plan, it)
	ge.logPlanAddLocked(it, 1)
	return nil
}

// PlanTradeText words a trade item from the player's side, so the amount
// reads as what is bought: "buy 500 gold with wood as wood comes in", or
// with no amount "sell wood for gold as wood comes in, until gold storage is
// full". The UI's confirmation uses it after "Plan: ".
func PlanTradeText(give, get string, amount float64) string {
	g, w := ResourceName(give), ResourceName(get)
	if amount > 0 {
		return fmt.Sprintf("buy %s with %s as %s comes in", Amount(amount, get), g, g)
	}
	return fmt.Sprintf("sell %s for %s as %s comes in, until %s storage is full", g, w, g, w)
}

// planTradeInvalid is why trade item it can never run in this age ("" if it
// can). A missing trade building only blocks it (planTradeView says so).
func (ge *GameEngine) planTradeInvalid(it PlanItem) string {
	for _, r := range []string{it.Key, it.To} {
		if _, ok := ge.rules.Resource(r); !ok {
			return "there is no resource called '" + r + "'"
		}
		if !ge.Resources.IsUnlocked(r) {
			return ResourceName(r) + " is not unlocked yet"
		}
	}
	if it.Key == it.To {
		return "it is the same resource"
	}
	if _, ok := ge.rules.MarketRate(it.Key, it.To, ge.age); !ok {
		return "the market doesn't trade that pair in this age"
	}
	return ""
}

// planTradeRate is the rate a trade of it would get now, as Exchange
// computes it.
func (ge *GameEngine) planTradeRate(it PlanItem) float64 {
	return ge.Trade.RateIn(it.Key, it.To, ge.age)
}

// planTradeBatch is how much of it.Key trade item it would sell (what the
// reservations above leave free, capped by the room in to's store and by
// what it still wants), whether the market has recovered enough to sell it
// now, and a note when it can't trade for a reason money won't fix.
func (ge *GameEngine) planTradeBatch(it PlanItem, reserved map[string]float64) (n float64, due bool, note string) {
	if ge.Buildings.TradeBuildingCount() < 1 {
		return 0, false, "you need a Market to trade"
	}
	rate := ge.planTradeRate(it)
	if rate <= 0 {
		return 0, false, "the market doesn't trade that pair in this age"
	}
	room := ge.Resources.GetStorage(it.To) - ge.Resources.Get(it.To)
	if room < 1 {
		return 0, false, ResourceName(it.To) + " storage is full"
	}
	n = math.Max(0, math.Min(ge.planFree(it.Key, reserved), room/rate))
	if it.Amount > 0 {
		n = math.Min(n, it.Amount/rate)
	}
	return n, n >= 1 && ge.Trade.Pressure(it.Key, it.To) < planTradeRecovered, ""
}

// runPlanTrade sells trade item it's batch if one is due, or holds it back
// from the items below, and returns what it sold and bought. It updates it
// (Got, Amount, Count).
func (ge *GameEngine) runPlanTrade(it *PlanItem, reserved map[string]float64) (float64, float64) {
	n, due, _ := ge.planTradeBatch(*it, reserved)
	if !due {
		reserved[it.Key] += n
		return 0, 0
	}
	ge.Trade.SetAge(ge.age)
	got, err := ge.Trade.Exchange(it.Key, it.To, n, ge.Resources, ge.Buildings, ge.tick)
	if err != nil {
		ge.addLog("debug", fmt.Sprintf("Plan: %s refused: %v", ge.planItemLabel(*it), err))
		reserved[it.Key] += n
		return 0, 0
	}
	it.Got += got
	if it.Amount > 0 {
		it.Amount -= got
		if it.Amount <= 1e-9*math.Max(1, it.Got) {
			it.Amount, it.Count = 0, 0
		}
	}
	return n, got
}

// planTradeView is trade item it for the UI; what it would sell counts
// against reserved, as in the walk.
func (ge *GameEngine) planTradeView(it PlanItem, reserved map[string]float64) PlanItemView {
	v := PlanItemView{Kind: it.Kind, Key: it.Key, To: it.To, Amount: it.Amount, Got: it.Got, Count: it.Count,
		Name: ResourceName(it.Key) + " → " + ResourceName(it.To)}
	if reason := ge.planTradeInvalid(it); reason != "" {
		v.Status, v.Note = PlanStatusBlocked, reason
		return v
	}
	n, due, note := ge.planTradeBatch(it, reserved)
	reserved[it.Key] += n
	switch {
	case note != "":
		v.Status, v.Note = PlanStatusBlocked, note
	case due:
		v.Status, v.Progress = PlanStatusReady, 1
		v.Cost = map[string]float64{it.Key: n}
	case n >= 1:
		v.Status, v.Note = PlanStatusBlocked, "market recovering from the last sale"
	default:
		v.Status, v.Short = PlanStatusWaiting, it.Key
	}
	return v
}
