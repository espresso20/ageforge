package game

import (
	"sort"

	"github.com/espresso20/ageforge/rules"
)

// wonder_share.go is the one answer to "how much of this wonder is banked":
// the Next Age strip, the Wonders panel, the Map's scaffolding and the
// Advance panel all read it.

// WonderBankShare is the share, from 0 to 1, of a wonder's price that is
// banked. It is weighted by price: each resource counts for what it is worth
// in age's price levels (the amount over the level the age prices that
// resource at), so a resource the price names in the trillions does not
// outweigh one it names in the thousands. A surplus in one resource cannot
// pay another part of the price. A resource the age has no level for counts
// as much as the average of the rest. cost is the wonder's price, bank what
// is in the bank; a nil set reads the core rules.
//
// Everything is summed in key order and with the products rounded on their
// own, so the share is the same on every machine (CONTRIBUTING.md, "Float
// rules for simulation code").
func WonderBankShare(set *rules.Set, age string, cost, bank map[string]float64) float64 {
	set = orCore(set)
	keys := make([]string, 0, len(cost))
	for res, need := range cost {
		if need > 0 {
			keys = append(keys, res)
		}
	}
	if len(keys) == 0 {
		return 0
	}
	sort.Strings(keys)
	weights := make([]float64, len(keys))
	known, sum := 0, 0.0
	for i, res := range keys {
		if level := set.DealPriceLevel(res, age); level > 0 {
			weights[i] = cost[res] / level
			sum += weights[i]
			known++
		}
	}
	fallback := 1.0
	if known > 0 {
		fallback = sum / float64(known)
	}
	total, done := 0.0, 0.0
	for i, res := range keys {
		if weights[i] == 0 {
			weights[i] = fallback
		}
		frac := min(bank[res], cost[res]) / cost[res]
		if frac < 0 {
			frac = 0
		}
		total += weights[i]
		done += float64(weights[i] * frac)
	}
	if total <= 0 {
		return 0
	}
	return min(1, done/total)
}
