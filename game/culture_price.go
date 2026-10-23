package game

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// What a festival and a Black Market deal cost, before any tech's cut: minutes
// of the town's culture income, and a minimum for a town that makes none.
//
// The minutes are a share of the age's stock scale worked out at the age's
// typical culture income (config.StockMinutes), held to a ceiling
// (config.StakeMinutes). The scale is the size a store had and follows
// prices, which climb far faster than culture income, so the share alone
// asked a few hundred minutes in the Classical Age and hundreds of billions
// in the Quantum Age. The ceiling is a little over the dearest minutes the
// prices were tuned at (the Colonial Age's: 5,970 for a festival, 11,941 for
// a deal), so every age before the Victorian keeps its price and every age
// from it asks the ceiling: 100 hours of your culture income for a festival,
// 200 for a deal.

// festivalPriceOf is the culture a festival costs a town making income of it
// per tick in age (set is the ruleset the typical income is read from).
// Pure.
func festivalPriceOf(set *rules.Set, age string, income float64) float64 {
	return max(festivalMinCost, cultureStakeOf(set, age, income, festivalCostShare, festivalMaxMinutes))
}

// blackMarketPriceOf is the culture one Black Market deal costs, likewise.
// Pure.
func blackMarketPriceOf(set *rules.Set, age string, income float64) float64 {
	return max(blackMarketMinCost, cultureStakeOf(set, age, income, blackMarketCostShare, blackMarketMaxMinutes))
}

// cultureStakeOf is a share of the age's stock scale, held to ceiling
// minutes, as minutes of a culture income of income per tick: the minutes are
// what the share comes to at the age's typical culture income
// (rules.Set.StakeMinutes), and the price is that many minutes of what the
// town really makes. A town making what the age's typical town does pays the
// minutes' worth of that income; one making half as much pays half. A town
// making none pays nothing here, which is why the prices keep a minimum.
func cultureStakeOf(set *rules.Set, age string, income, share, ceiling float64) float64 {
	minutes := set.StakeMinutes("culture", age, share, ceiling)
	return float64(float64(minutes*config.TicksPerMinute) * income)
}

// CulturePrices is what a festival and one Black Market deal cost, before the
// techs' cut, in age for a town making income of culture per tick: the
// engine's own prices, for the checks that read them against another measure
// of the town. Pure.
func CulturePrices(set *rules.Set, age string, income float64) (festival, blackMarket float64) {
	return festivalPriceOf(set, age, income), blackMarketPriceOf(set, age, income)
}
