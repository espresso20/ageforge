package game

import (
	"fmt"
	"math"
	"strconv"

	"github.com/espresso20/ageforge/config"
)

// Trade items: `plan trade <from> <to> [amount]` sells from for to at the
// market as from comes in, until amount of to has been bought (no amount:
// until removed). It is how a plan gets the resources an age buys rather
// than makes (stone after the Bronze Age, the Industrial Age's iron) while
// the player is away, instead of one trade per visit.
//
// Rules:
//   - Like any item, it has the priority of its place in the plan: it holds
//     back the from it will sell (what the items above leave free, up to
//     what it still wants to buy), and items below it only see the rest.
//   - It sells once the market has recovered from its last sale (the pair's
//     supply pressure under planTradeRecovered), not every tick: every trade
//     raises the pressure by the same step whatever its size, so selling in
//     one lump a minute or so apart keeps the rate within a percent of the
//     market's, where selling every tick would cost up to 30%.
//   - It never buys more than to's store has room for (the rest would be
//     lost), at the market's current rate (parity less the fee and the
//     pressure), and only with a trade building standing, like `trade`.
//   - It drops out when it has bought its amount, or when the pair no longer
//     trades (an age advance) or a resource is locked.

// planTradeRecovered is the supply pressure under which a trade item sells
// again (a rate within 0.6% of the market's).
const planTradeRecovered = 0.02

// PlanAddTrade appends a trade item: sell from for to as from comes in, until
// amount of to is bought (0: until removed). The pair must trade at the
// market in this age.
func (ge *GameEngine) PlanAddTrade(from, to string, amount float64) error {
	if !(amount >= 0) || math.IsInf(amount, 0) {
		return fmt.Errorf("the amount must be a positive number")
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	it := PlanItem{Kind: PlanTrade, Key: from, To: to, Count: 1, Amount: amount}
	if reason := ge.planTradeInvalid(it); reason != "" {
		return fmt.Errorf("can't plan a trade of %s for %s: %s", from, to, reason)
	}
	for _, p := range ge.plan {
		if p.Kind == PlanTrade && p.Key == from && p.To == to {
			return fmt.Errorf("the plan already trades %s for %s — remove that item first", from, to)
		}
	}
	if len(ge.plan) >= MaxPlanItems {
		return fmt.Errorf("the plan is full (%d items) — remove one first", MaxPlanItems)
	}
	ge.plan = append(ge.plan, it)
	return nil
}

// planTradeInvalid is why trade item it can never run in this age ("" if it
// can). A missing trade building only blocks it (planTradeView says so).
func (ge *GameEngine) planTradeInvalid(it PlanItem) string {
	defs := config.ResourceByKey()
	for _, r := range []string{it.Key, it.To} {
		if _, ok := defs[r]; !ok {
			return "unknown resource " + r
		}
		if !ge.Resources.IsUnlocked(r) {
			return r + " is not unlocked"
		}
	}
	if it.Key == it.To {
		return "it is the same resource"
	}
	if _, ok := config.MarketRate(it.Key, it.To, ge.age); !ok {
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
		return 0, false, "needs a market (or a later trade building)"
	}
	rate := ge.planTradeRate(it)
	if rate <= 0 {
		return 0, false, "the market doesn't trade that pair in this age"
	}
	room := ge.Resources.GetStorage(it.To) - ge.Resources.Get(it.To)
	if room < 1 {
		return 0, false, it.To + " storage is full"
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
		Name: it.Key + " → " + it.To}
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

// formatPlanAmount prints an amount for plan log lines: 950, 12.5K, 3.1M.
func formatPlanAmount(v float64) string {
	for _, s := range []struct {
		at  float64
		sfx string
	}{{1e15, "Q"}, {1e12, "T"}, {1e9, "B"}, {1e6, "M"}, {1e3, "K"}} {
		if v >= s.at {
			return strconv.FormatFloat(v/s.at, 'f', 1, 64) + s.sfx
		}
	}
	return strconv.FormatFloat(v, 'f', 0, 64)
}
