package smoke

import (
	"math"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// goodStockOf is a good stock of res for the town st describes: the age's
// stock scale in minutes of the town's own income of res (rules.Set.StockOf).
// Stock has no ceiling, so reports and the bot weigh a pile of it against this
// where they once weighed it against a store.
func goodStockOf(set *rules.Set, st game.GameState, res string) float64 {
	return set.StockOf(res, st.Age, 1, math.Max(st.Resources[res].Rate, 0))
}
