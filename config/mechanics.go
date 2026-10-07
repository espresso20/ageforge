package config

import (
	"fmt"
	"sync"
)

// A mechanic number is one constant of a game mechanic that a tech can move:
// the market's fee, how long a trade route takes, what a gift costs. The
// engine reads such a number as "the constant, with the tech term": the
// constant times the term for a mechanic that multiplies, the constant plus
// the term for one that adds. With no tech researched the term is 1 or 0 and
// the number is the constant.
//
// The table lists the numbers a tech moves today. A number joins it in the
// same change that gives a tech an effect on it and teaches the engine to
// read it through the term (GameEngine.mechanic).

// The keys of the mechanic numbers.
const (
	// MechanicGatherAmount is added to what one `gather` brings in.
	MechanicGatherAmount = "gather_amount"
	// MechanicMarketFee is added to the market's fee (ExchangeFee, a
	// fraction of parity): -0.03 is 3 points lower.
	MechanicMarketFee = "market_fee"
	// MechanicRouteTicks multiplies the time a trade route takes per run.
	MechanicRouteTicks = "route_ticks"
	// MechanicExpeditionTicks multiplies the time a scouting expedition
	// takes. Campaigns keep theirs.
	MechanicExpeditionTicks = "expedition_ticks"
	// MechanicRaidLoss multiplies what a raid takes, after the garrison's
	// own share.
	MechanicRaidLoss = "raid_loss"
	// MechanicDealRefreshTicks multiplies how long a civilization's offers
	// last before the next set.
	MechanicDealRefreshTicks = "deal_refresh_ticks"
	// MechanicGiftCost multiplies what a gift to a civilization costs.
	MechanicGiftCost = "gift_cost"
	// MechanicFestivalCooldownTicks multiplies the wait between festivals.
	MechanicFestivalCooldownTicks = "festival_cooldown_ticks"
	// MechanicFestivalCost multiplies what a festival costs.
	MechanicFestivalCost = "festival_cost"
	// MechanicRouteIncome multiplies what a trade route brings in per run.
	MechanicRouteIncome = "route_income"
	// MechanicMoraleCap is added to the highest morale can rise (1.0, plus
	// what the wonders add): 0.05 is 5 points higher.
	MechanicMoraleCap = "morale_cap"
	// MechanicWonderBuildTicks multiplies the time a wonder takes to build,
	// beside the techs' cut of all construction.
	MechanicWonderBuildTicks = "wonder_build_ticks"
	// MechanicFestivalTicks multiplies how long a festival's bonus lasts.
	MechanicFestivalTicks = "festival_ticks"
	// MechanicGiftOpinion multiplies the opinion a gift earns, rounded down
	// to a whole point.
	MechanicGiftOpinion = "gift_opinion"
	// MechanicAllianceBonus multiplies what an allied civilization adds to
	// its specialty.
	MechanicAllianceBonus = "alliance_bonus"
	// MechanicUpgradeCost multiplies what upgrading a building to the next
	// of its line costs.
	MechanicUpgradeCost = "upgrade_cost"
	// MechanicCampaignTicks multiplies the time a military campaign takes.
	// Scouting expeditions have MechanicExpeditionTicks.
	MechanicCampaignTicks = "campaign_ticks"
	// MechanicSoldierStorage multiplies how many soldiers can be held.
	MechanicSoldierStorage = "soldier_storage"
	// MechanicCampaignReward multiplies what a military campaign brings
	// back, won or lost.
	MechanicCampaignReward = "campaign_reward"
)

// MechanicUnit is how the size of a step on a mechanic number is printed.
type MechanicUnit int

const (
	// UnitPercent prints a fraction as a percentage: -0.15 is "15%".
	UnitPercent MechanicUnit = iota
	// UnitPoints prints a fraction as percentage points: -0.03 is "3 points".
	UnitPoints
	// UnitAmount prints the number as it is: 2 is "2".
	UnitAmount
)

// MechanicDef is one mechanic number a tech can move.
type MechanicDef struct {
	Key string
	// Name is the number in running text: "market fee".
	Name string
	// Multiplies says how effects on the number stack: each multiplies it
	// (a Value of -0.15 leaves 85% of it), or, when false, each is added to
	// it.
	Multiplies bool
	// Min and Max hold the tech term: the factor of a number that
	// multiplies, the sum of one that adds. However many techs move the
	// number, the term stays inside them.
	Min, Max float64
	// Text is one effect in player words, with one %s for the size of the
	// step: "trade routes take %s less time".
	Text string
	Unit MechanicUnit
}

// Mechanics lists every mechanic number, in a fixed order.
func Mechanics() []MechanicDef {
	return []MechanicDef{
		{Key: MechanicGatherAmount, Name: "hand gathering", Min: 0, Max: 25,
			Text: "gathering by hand brings %s more", Unit: UnitAmount},
		{Key: MechanicMarketFee, Name: "market fee", Min: -0.15, Max: 0,
			Text: "market fee %s lower", Unit: UnitPoints},
		{Key: MechanicRouteTicks, Name: "trade route time", Multiplies: true, Min: 0.40, Max: 1,
			Text: "trade routes take %s less time", Unit: UnitPercent},
		{Key: MechanicExpeditionTicks, Name: "expedition time", Multiplies: true, Min: 0.40, Max: 1,
			Text: "expeditions take %s less time", Unit: UnitPercent},
		{Key: MechanicRaidLoss, Name: "raid losses", Multiplies: true, Min: 0.40, Max: 1,
			Text: "raids on you take %s less", Unit: UnitPercent},
		{Key: MechanicDealRefreshTicks, Name: "deal refresh time", Multiplies: true, Min: 0.40, Max: 1,
			Text: "deals refresh %s sooner", Unit: UnitPercent},
		{Key: MechanicGiftCost, Name: "gift cost", Multiplies: true, Min: 0.40, Max: 1,
			Text: "gifts cost %s less", Unit: UnitPercent},
		{Key: MechanicFestivalCooldownTicks, Name: "festival cooldown", Multiplies: true, Min: 0.40, Max: 1,
			Text: "festivals come back %s sooner", Unit: UnitPercent},
		{Key: MechanicFestivalCost, Name: "festival cost", Multiplies: true, Min: 0.40, Max: 1,
			Text: "festivals cost %s less", Unit: UnitPercent},
		{Key: MechanicRouteIncome, Name: "trade route income", Multiplies: true, Min: 1, Max: 2,
			Text: "trade routes bring in %s more", Unit: UnitPercent},
		{Key: MechanicMoraleCap, Name: "morale ceiling", Min: 0, Max: 0.30,
			Text: "morale can rise %s higher", Unit: UnitPoints},
		{Key: MechanicWonderBuildTicks, Name: "wonder construction time", Multiplies: true, Min: 0.40, Max: 1,
			Text: "wonders take %s less time to build", Unit: UnitPercent},
		{Key: MechanicFestivalTicks, Name: "festival length", Multiplies: true, Min: 1, Max: 2,
			Text: "festivals last %s longer", Unit: UnitPercent},
		{Key: MechanicGiftOpinion, Name: "opinion from gifts", Multiplies: true, Min: 1, Max: 3,
			Text: "gifts raise opinion %s more", Unit: UnitPercent},
		{Key: MechanicAllianceBonus, Name: "alliance bonuses", Multiplies: true, Min: 1, Max: 2,
			Text: "alliances give %s more", Unit: UnitPercent},
		{Key: MechanicUpgradeCost, Name: "upgrade cost", Multiplies: true, Min: 0.40, Max: 1,
			Text: "upgrades cost %s less", Unit: UnitPercent},
		{Key: MechanicCampaignTicks, Name: "campaign time", Multiplies: true, Min: 0.40, Max: 1,
			Text: "campaigns take %s less time", Unit: UnitPercent},
		{Key: MechanicSoldierStorage, Name: "soldier storage", Multiplies: true, Min: 1, Max: 2,
			Text: "room for %s more soldiers", Unit: UnitPercent},
		{Key: MechanicCampaignReward, Name: "campaign loot", Multiplies: true, Min: 1, Max: 2,
			Text: "campaigns bring back %s more", Unit: UnitPercent},
	}
}

// mechanicIndex is Mechanics by key, built once: the table never changes,
// and the tech layer looks a mechanic up for every effect it folds.
var mechanicIndex = sync.OnceValue(func() map[string]MechanicDef {
	m := make(map[string]MechanicDef)
	for _, d := range Mechanics() {
		m[d.Key] = d
	}
	return m
})

// MechanicByKey is Mechanics by key. The map is shared: read it, never
// change it.
func MechanicByKey() map[string]MechanicDef { return mechanicIndex() }

// EffectText is one effect of size v on the number, in player words:
// "trade routes take 15% less time", "market fee 3 points lower".
func (m MechanicDef) EffectText(v float64) string {
	var size string
	switch m.Unit {
	case UnitPoints:
		size = FormatRateValue(absFloat(v)*100) + " points"
	case UnitAmount:
		size = FormatAmount(v)
	default:
		size = FormatPercent(v)
	}
	return fmt.Sprintf(m.Text, size)
}

// Apply is the number with the tech term: base x term for a mechanic that
// multiplies, base + term for one that adds.
func (m MechanicDef) Apply(base, term float64) float64 {
	if m.Multiplies {
		return float64(base * term)
	}
	return base + term
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
