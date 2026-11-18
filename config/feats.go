package config

// feats.go holds the numbers of the legendary feats, one per age. Shares are of
// the age's reference time (CoinRefTicks), so they follow the run's speed.

// FeatMarks is each feat's number, by age.
var FeatMarks = map[string]float64{
	"primitive_age":    31,      // huts standing at once
	"stone_age":        0.5,     // share of the age's time to leave within
	"bronze_age":       0.6,     // share of the age's time to leave within, with no upgrade bought
	"iron_age":         650,     // ticks after entering to research the keystone, at 1x
	"classical_age":    8,       // festivals held in the age, at 1x
	"electric_age":     3,       // Nuclear Plants' worth of electricity in stock on leaving
	"atomic_age":       0.86,    // share of the age's time to leave within, with no pacts made
	"modern_age":       0.76,    // share of the summed times of every age through Modern, for the whole run
	"colonial_age":     9,       // trade routes running at once, each having finished a cycle
	"industrial_age":   1.0 / 3, // share of the staffed workers in the energy jobs
	"victorian_age":    1.0 / 6, // share of the age's time to raise its wonder within
	"fusion_age":       8,       // civilizations allied at once, with no war running
	"galactic_age":     2,       // Appease level bought against the Last Passage
	"quantum_age":      2,       // Brace level held with the Last Passage invited
	"transcendent_age": 2,       // Brace level held when the Last Passage is endured
	"interstellar_age": 0.5,     // share of the age's time the Warp Commerce route runs unbroken
}
