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
	"victorian_age":    1.08601828e10,
	"electric_age":     1.208601828e11,
	"atomic_age":       7.208601828e11,
	"modern_age":       3.120860183e12,
	"information_age":  1.962086018e13,
	"digital_age":      7.462086018e13,
	"cyberpunk_age":    6.746208602e14,
	"fusion_age":       3.32462086e15,
	"space_age":        1.632462086e16,
	"interstellar_age": 4.863246209e17,
	"galactic_age":     4.436324621e18,
	"quantum_age":      3.444363246e20,
	"transcendent_age": 3.444363246e20,
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
