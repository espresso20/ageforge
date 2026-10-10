package config

// Event sizes.
//
// An event's gains and losses follow the town it happens to. A fixed amount
// (+250 food, lost up to 30 wood) is outgrown within an age: it means
// nothing to a town with ten thousand in store. So an event's effects are
// written as sizes, and the engine turns each into an amount when the event
// fires (EventSize), from what the town holds and makes:
//
//	EventGain  Value is minutes of the town's own income of Target, given
//	           at once.
//	EventLoss  Value is the share of what the town holds of Target that is
//	           taken.
//	EventRate  Value is a share of the town's own income of Target, added to
//	           its rate each tick while the event lasts (negative: taken).
//
// Each amount is held to a band that depends on the age, so an event is
// felt in a big town and does not ruin a small one. The band is counted in
// the age's typical income of the resource (TypicalIncome): what a town that
// invests moderately makes of it there.
//
//   - A gain or a boost is sized on the town's income, but on no less than
//     EventIncomeFloor and no more than EventIncomeCeil times the typical
//     income. A town that makes little of the resource still gets something,
//     and a town that makes a great deal does not get a windfall out of
//     scale with its age.
//   - A setback to a rate has no floor: a town that makes none of the
//     resource loses none. It has the same ceiling.
//   - A loss is at least EventLossFloorMinutes and at most
//     EventLossCeilMinutes of the typical income, and never more than
//     EventLossMostShare of what the town holds, floor or not. The floor is
//     a token: a town that has just spent its stock loses a few seconds of
//     income, not a larger share of what little it holds.
//
// The log line states the amounts that were applied (the engine writes them
// into it), so an event's text carries no numbers of its own.
const (
	// EventGain, EventLoss and EventRate are the sized effect types an event
	// definition uses.
	EventGain = "gain"
	EventLoss = "loss"
	EventRate = "rate"

	// EventTicksPerMinute converts an EventGain's minutes to ticks.
	EventTicksPerMinute = 60.0 / TickSeconds

	// EventIncomeFloor and EventIncomeCeil bound the income a gain or a
	// boost is sized on, as multiples of the age's typical income.
	EventIncomeFloor = 0.25
	EventIncomeCeil  = 4.0

	// EventLossFloorMinutes and EventLossCeilMinutes bound a loss, in
	// minutes of the age's typical income.
	EventLossFloorMinutes = 0.1
	EventLossCeilMinutes  = 30.0

	// EventLossMostShare is the most of a stock any one loss takes.
	EventLossMostShare = 0.25
)

// EventTown is what an event's size is read from: what the town holds of the
// resource, what it makes of it each tick, and what a moderate town of its
// age makes (TypicalIncome, at the town's own speed). Typical is 0 for a
// resource the age has no typical income of: the size then has no band.
type EventTown struct {
	Stock, Income, Typical float64
}

// IsSizedEffect reports whether typ is one of the sized effect types.
func IsSizedEffect(typ string) bool {
	return typ == EventGain || typ == EventLoss || typ == EventRate
}

// EventSize is the amount a sized effect comes to in town: units gained
// (EventGain), units lost (EventLoss), or units per tick, signed
// (EventRate). 0 for any other effect type. Pure.
func EventSize(eff Effect, town EventTown) float64 {
	income := max(town.Income, 0)
	switch eff.Type {
	case EventGain:
		if eff.Value <= 0 {
			return 0
		}
		return float64(float64(eff.Value*EventTicksPerMinute) * sizedIncome(income, town.Typical, true))
	case EventRate:
		// A boost has the gain's floor; a setback takes a share of what
		// the town really makes.
		return float64(eff.Value * sizedIncome(income, town.Typical, eff.Value > 0))
	case EventLoss:
		stock := max(town.Stock, 0)
		if eff.Value <= 0 || stock <= 0 {
			return 0
		}
		loss := float64(eff.Value * stock)
		if floor, ceil := EventBand(eff, town.Typical); ceil > 0 {
			loss = min(max(loss, floor), ceil)
		}
		return min(loss, float64(EventLossMostShare*stock))
	}
	return 0
}

// sizedIncome is the income a gain or a rate is sized on: the town's own,
// held to the band around the typical income (no floor for a setback, and
// no band at all where the age has no typical income).
func sizedIncome(income, typical float64, floored bool) float64 {
	if typical <= 0 {
		return income
	}
	income = min(income, float64(EventIncomeCeil*typical))
	if floored {
		income = max(income, float64(EventIncomeFloor*typical))
	}
	return income
}

// EventBand is the least and the most a sized effect can come to in an age
// whose typical income of its resource is typical, as magnitudes (a rate's
// are per tick). A loss can come to less than its floor in a town that
// holds little (EventLossMostShare). Both are 0 where typical is 0: the
// size has no band there. Pure.
func EventBand(eff Effect, typical float64) (floor, ceil float64) {
	if typical <= 0 {
		return 0, 0
	}
	v := eff.Value
	if v < 0 {
		v = -v
	}
	switch eff.Type {
	case EventGain:
		ticks := float64(v * EventTicksPerMinute)
		return float64(float64(ticks*EventIncomeFloor) * typical), float64(float64(ticks*EventIncomeCeil) * typical)
	case EventRate:
		floor = float64(float64(v*EventIncomeFloor) * typical)
		if eff.Value < 0 {
			floor = 0
		}
		return floor, float64(float64(v*EventIncomeCeil) * typical)
	case EventLoss:
		return float64(float64(EventLossFloorMinutes*EventTicksPerMinute) * typical), float64(float64(EventLossCeilMinutes*EventTicksPerMinute) * typical)
	}
	return 0, 0
}
