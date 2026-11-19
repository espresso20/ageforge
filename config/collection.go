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

// CoinFeat is the legendary coin of an age: struck once for the age's one hard
// feat. It is not a well, so it is not in CoinTypes, but the account keeps and
// merges its copies like the others. It has no level ladder: a copy is held at
// CoinLegendary.
const CoinFeat = "feat"

// CoinTypes is the coin types in their display order.
var CoinTypes = []string{CoinAge, CoinWonder, CoinTechs, CoinMaximalist, CoinHarbinger}

// IsCoinType reports whether key is one of the coin types, the feat coin included.
func IsCoinType(key string) bool {
	if key == CoinFeat || key == CoinTally {
		return true
	}
	for _, t := range CoinTypes {
		if t == key {
			return true
		}
	}
	return false
}

// CoinLevels is how many levels a coin has: 1 is the lowest, CoinLevels the finest.
const CoinLevels = 6

// CoinLegendary is the one level a feat coin is held at, above the six.
const CoinLegendary = CoinLevels + 1

// CoinLevelOK reports whether level is one a coin of typ can be held at.
func CoinLevelOK(typ string, level int) bool {
	if typ == CoinFeat {
		return level == CoinLegendary
	}
	return level >= 1 && level <= CoinLevels
}

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

// CoinLevelAt is the level of a coin struck after ticks ticks in an age whose
// reference time is ref ticks: the finest level whose share of ref the time
// stays under, and 1 when ref is not positive or none holds.
func CoinLevelAt(ref, ticks float64) int {
	if ref <= 0 {
		return 1
	}
	for _, m := range coinMarks {
		if ticks <= m.Share*ref {
			return m.Level
		}
	}
	return 1
}
