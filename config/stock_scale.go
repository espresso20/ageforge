package config

// The stock scale: how much of one resource an age's economy calls a good
// stock. Nothing stores resources any more (stock has no ceiling), but a few
// rules are written as a share of a good stock: the festival and Black Market
// prices, the size of a faction deal and the Brace ceiling. Each of them is
// that share expressed in minutes of income: the town's own income for the
// ones that follow the town, the age's typical income (TypicalIncome) for the
// one that is priced on the age (Brace).
//
// StockScale is the number they are shares of, per age. It is the size a
// store had when stores were sized, as measured on the reference town: the
// base of a resource and five copies of every storage building up to the age,
// each copy sized to hold the age's dearest ask. It is kept as measured, a
// table, so the prices that came out of it stay where they were. It is not
// derived from the buildings and moves with nothing: change it only to move
// those prices on purpose.
//
// The Transcendent Age has no storage building of its own and keeps the
// Quantum Age's scale.
const (
	// TicksPerMinute and TicksPerHour turn minutes and hours of income into
	// ticks.
	TicksPerMinute = 60.0 / TickSeconds
	TicksPerHour   = 3600.0 / TickSeconds
)

var stockScale = map[string]float64{
	"primitive_age":    800,
	"stone_age":        9300,
	"bronze_age":       57800,
	"iron_age":         232800,
	"classical_age":    1682800,
	"medieval_age":     12182800,
	"renaissance_age":  55182800,
	"colonial_age":     260182800,
	"industrial_age":   860182800,
	"victorian_age":    10.8601828 * Billion,
	"electric_age":     120.8601828 * Billion,
	"atomic_age":       720.8601828 * Billion,
	"modern_age":       3.120860183 * Trillion,
	"information_age":  19.62086018 * Trillion,
	"digital_age":      74.62086018 * Trillion,
	"cyberpunk_age":    674.6208602 * Trillion,
	"fusion_age":       3.32462086 * Quadrillion,
	"space_age":        16.32462086 * Quadrillion,
	"interstellar_age": 486.3246209 * Quadrillion,
	"galactic_age":     4.436324621 * Quintillion,
	"quantum_age":      344.4363246 * Quintillion,
	"transcendent_age": 344.4363246 * Quintillion,
}

// StockScale is the age's stock scale (0 for an unknown age).
func StockScale(age string) float64 { return stockScale[age] }

// StockMinutes is a share of the age's stock scale expressed in minutes of an
// income of typical per tick: the minutes of the age's typical income of a
// resource that make that share. 0 where typical is not above 0 (the age has
// no typical income of the resource). Pure.
func StockMinutes(age string, share, typical float64) float64 {
	if typical <= 0 {
		return 0
	}
	return float64(share*StockScale(age)) / float64(TicksPerMinute*typical)
}

// StakeMinutes is StockMinutes held to a ceiling of ceiling minutes: what a
// culture stake (the festival, a Black Market deal) asks, in minutes of the
// age's typical income. The scale is the size a store had, and the stores
// followed the price of what they held, which climbs far faster than any
// income: from the Classical Age to the Quantum Age the scale grows about
// 2e14 times and a moderate town's culture income about 6e5 times, so a share
// of the scale left alone asks hundreds of minutes of culture in the Classical
// Age and hundreds of billions in the Quantum Age. The ceiling is written
// where the stake is (game's festivalMaxMinutes and blackMarketMaxMinutes), a
// little over the dearest the prices were tuned at, so the ages they were
// tuned in keep their price and the later ones stop at the ceiling. 0 where
// typical is not above 0, as StockMinutes. Pure.
func StakeMinutes(age string, share, typical, ceiling float64) float64 {
	return min(StockMinutes(age, share, typical), ceiling)
}

// StockOf is a share of a good stock of a resource for a town that makes
// income of it per tick in the age: StockMinutes of the town's own income,
// held to the band around the age's typical income (SizedIncome), so a town
// that makes little of the resource is still sized on something and one that
// makes a great deal is not sized without bound. Where the age has no typical
// income of the resource there is nothing to measure the town against, and
// the share of the scale itself stands. Pure.
func StockOf(age string, share, typical, income float64) float64 {
	if typical <= 0 {
		return float64(share * StockScale(age))
	}
	return float64(float64(StockMinutes(age, share, typical)*TicksPerMinute) * SizedIncome(income, typical))
}

// SizedIncome is income held to the band an event gain is sized in: no less
// than EventIncomeFloor and no more than EventIncomeCeil times typical. Where
// typical is not above 0 there is no band and income stands. Pure.
func SizedIncome(income, typical float64) float64 {
	return sizedIncome(max(income, 0), typical, true)
}
