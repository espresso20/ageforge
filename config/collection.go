package config

// collection.go holds the coin collection's constants: the five coin types
// every age has, the number of levels a coin can be struck at, and the rule
// that turns the time spent in an age into a level.

// The coin types every age has. Keys are the strings the account stores.
const (
	CoinAge        = "age"
	CoinWonder     = "wonder"
	CoinTechs      = "techs"
	CoinMaximalist = "maximalist"
	CoinHarbinger  = "harbinger"
)

// CoinTypes is the coin types in their display order.
var CoinTypes = []string{CoinAge, CoinWonder, CoinTechs, CoinMaximalist, CoinHarbinger}

// IsCoinType reports whether key is one of the coin types.
func IsCoinType(key string) bool {
	for _, t := range CoinTypes {
		if t == key {
			return true
		}
	}
	return false
}

// CoinLevels is how many levels a coin has: 1 is the lowest, CoinLevels the finest.
const CoinLevels = 6

// coinMarks are the level marks as shares of an age's reference time, finest
// first: a coin struck at or under the share has at least that level. Every
// age uses the same shares. They are first figures, kept in one place to be
// tuned.
var coinMarks = [CoinLevels - 1]struct {
	Level int
	Share float64
}{
	{6, 0.60},
	{5, 0.80},
	{4, 1.00},
	{3, 1.30},
	{2, 1.80},
}

// CoinLevel is the level of a coin struck after ticksInAge ticks in age, by a
// run that plays the age at speed times the base rate. The reference time is
// the age's pacing target at 1x divided by speed (a speed of 0 or less counts
// as 1). An unknown age, or one with no target, gives level 1.
func CoinLevel(age string, ticksInAge, speed float64) int {
	if speed <= 0 {
		speed = 1
	}
	ref := AgeTargetTicks(age) / speed
	if ref <= 0 {
		return 1
	}
	for _, m := range coinMarks {
		if ticksInAge <= m.Share*ref {
			return m.Level
		}
	}
	return 1
}
