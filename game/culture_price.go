package game

import (
	"sort"

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

// blackMarketDeals is the resources the smugglers deal in: gold and nine
// others, and no others.
var blackMarketDeals = map[string]bool{
	"gold": true, "iron": true, "coal": true, "steel": true, "oil": true, "electricity": true,
	"data": true, "crypto": true, "dark_matter": true, "quantum_flux": true,
}

// A win on the Black Market pays the same worth whichever resource it is
// taken in: the stake's minutes of culture income, times blackMarketWinMult,
// as minutes of the town's income of the resource. It was the stake as gold
// at a fixed rate for each resource, which paid from a few minutes of income
// to tens of thousands depending on the item (gold against coal), for the
// same stake. Each income is held to the band around the
// age's typical income (config.SizedIncome), as an event's gain is, so a town
// that makes none of the resource is still paid and one that makes a great
// deal is not paid without bound; the minutes are the stake over the culture
// income so held, so a town paying the price in its own income gets exactly
// blackMarketWinMult of them.

// blackMarketPayoutOf is the payout for a win on resource with a stake of
// stake culture, for a town whose culture income and income of resource are
// culture and have (each with the age's typical income, at the town's speed),
// held to its band by set. 0 for a resource the smugglers do not deal in or
// the age makes none of. Pure.
func blackMarketPayoutOf(set *rules.Set, stake float64, culture config.EventTown, resource string, have config.EventTown) float64 {
	if !blackMarketDeals[resource] || culture.Typical <= 0 || have.Typical <= 0 {
		return 0
	}
	paid := set.SizedIncome(culture.Income, culture.Typical)
	made := set.SizedIncome(have.Income, have.Typical)
	return float64(float64(blackMarketWinMult*stake)*made) / paid
}

// BlackMarketPayout is what a win on resource pays in age, for a town making
// cultureIncome of culture and income of resource per tick that paid stake
// culture: the engine's own figure, for the checks that read it against
// another measure of the town. Pure.
func BlackMarketPayout(set *rules.Set, age, resource string, stake, cultureIncome, income float64) float64 {
	return blackMarketPayoutOf(set, stake,
		config.EventTown{Income: cultureIncome, Typical: set.TypicalIncome("culture", age)},
		resource, config.EventTown{Income: income, Typical: set.TypicalIncome(resource, age)})
}

// BlackMarketResources lists the resources the smugglers deal in, sorted: the
// ones a win can be taken in (if the age makes any of them).
func BlackMarketResources() []string {
	out := make([]string, 0, len(blackMarketDeals))
	for r := range blackMarketDeals {
		out = append(out, r)
	}
	sort.Strings(out)
	return out
}
