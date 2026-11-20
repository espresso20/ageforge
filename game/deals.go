package game

import (
	"cmp"
	"maps"
	"math"
	"math/rand"
	"slices"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rowid"
	"github.com/espresso20/ageforge/rules"
)

// Faction offers: each civilization you have met offers a small, rotating
// set of gifts, derived from its data the way factionProfile derives its
// boons. There are no per-civ tables; a civ's specialty, personality and
// strength and your standing with it decide everything.
//
// An offer is a cost, a lump of one of your goods (never the civ's own
// specialty), and one gift that goods cannot buy. Goods trade for goods on
// the trade routes (trade.go), not here.
//
// Rules (site/docs/factions.md, Offers):
//
//   - Gifts. A boost (DealBoost) adds dealBoostShare to the gross output of one
//     resource the town makes, for dealBoostTicks ticks of play (stretched for
//     the age); the resource is one of the top dealWantPool producers, the
//     civ's specialty preferred. A research gift (DealResearch) is instant
//     progress on the tech in progress, no more than it still needs. People
//     (DealPeople) join the idle pool, within the housing. Know-how
//     (DealKnowhow) is the civ's own perk, permanent: dealKnowhowShare more of
//     its specialty, once per civ per run.
//   - Prices, in price units of the age (a unit of res is
//     rules.Set.DealPriceLevel(res) of it). mult is 0.84-1.0 of parity by
//     standing and personality; a gift is worth mult of what is paid. A boost
//     costs dealBoostWorth of the units of output it adds over its run, over
//     mult. Research gives dealResearchWorth x the lot's units at research's
//     level, times mult. People give the lot's units times mult over one
//     worker's worth: the food it eats a tick for one rotation of offers
//     (headFood x refresh, at food's level). Know-how costs dealKnowhowUnits.
//   - Personality. Mercantile civs offer one more offer, in bigger lots.
//     Aggressive civs offer one fewer, at worse rates. Isolationists offer one
//     (two when allied).
//   - Standing. Neutral civs offer 2, friendly (friendly status or opinion
//     25+) 3, allied 4. Civs at war, under embargo, rival or with opinion at
//     or below -50 offer nothing and accept nothing. An accepted offer adds
//     dealTradeStanding opinion, up to dealStandingCap: offers can bring a
//     civ to the edge of an alliance, never past it, and never out of a war.
//     Know-how is offered at opinion dealStandingCap or allied, in the first
//     slot of a set only.
//   - Size. A lot is 1.5 price units of the age, times 0.9 + 0.1 per point of
//     strength, the standing and personality factors and a 0.75-1.25 roll,
//     capped so the price comes to 80% of a good stock of the resource
//     (dealEnv.size: minutes of your own income of it). Know-how is not
//     capped: it is saved up for.
//   - Refresh. Offers rotate every dealRefreshTicks ticks of play (stretched
//     for the age, dealRefreshIn) and on an age advance. The timer counts
//     ticks of live play only: offline catch-up does not advance it.
//   - Determinism. Rolls come off the engine's seeded rng in roster order,
//     three draws per slot whether or not the slot yields an offer. A blocked
//     civ draws nothing. Offers, taken flags and the timer are saved.

// Offer kinds, by gift.
const (
	DealBoost    = "boost"
	DealResearch = "research"
	DealPeople   = "people"
	DealKnowhow  = "knowhow"
)

// Offer tuning, grouped so balancing is a data edit.
const (
	// dealRefreshTicks is how many ticks of live play an offer set lasts
	// (an hour at 1x) on the base curve; dealRefreshIn stretches it.
	dealRefreshTicks = 1800

	// dealHostileOpinion: at or below this opinion a civ will not trade.
	dealHostileOpinion = -50
	// dealStandingCap is the opinion offers can raise a civ to, no further.
	dealStandingCap = 50
	// dealTradeStanding is what an accepted offer adds.
	dealTradeStanding = 1

	// Rates as a multiple of parity.
	dealRateNeutral       = 0.88
	dealRateFriendly      = 0.92
	dealRateAllied        = 0.96
	dealMercantileBonus   = 0.02
	dealAggressivePenalty = 0.04
	dealRateMin           = 0.84
	dealRateMax           = 1.0

	// Lot size in price units of the age.
	dealLotBase         = 1.5
	dealLotStrengthBase = 0.9
	dealLotPerStrength  = 0.1
	dealLotFriendly     = 1.25
	dealLotAllied       = 1.5
	dealLotMercantile   = 1.2
	dealLotJitter       = 0.5 // lots vary from 1 - J/2 to 1 + J/2
	// dealGiveShare caps a price at this share of a good stock of the
	// resource (dealEnv.size).
	dealGiveShare = 0.8
	// dealWantPool is how many of the resources you hold most of (in minutes
	// of your income of each) a price chooses from, and how many top
	// producers a boost chooses from.
	dealWantPool = 3

	// The gifts. A boost is dealBoostShare more output for dealBoostTicks
	// base ticks (20 minutes at 1x, stretched for the age); it costs
	// dealBoostWorth of the output it adds.
	dealBoostShare = 0.25
	dealBoostTicks = 600
	dealBoostWorth = 0.6
	// dealResearchWorth is the research a lot buys, as a multiple of its value.
	dealResearchWorth = 1.25
	// Know-how: dealKnowhowShare more of the civ's specialty, for
	// dealKnowhowUnits price units, in the first slot of a set once the civ
	// qualifies.
	dealKnowhowShare = 0.05
	dealKnowhowUnits = 10
	// dealPeopleShare caps a people offer at this share of the homes' capacity.
	dealPeopleShare = 0.1

	// dealIDStride spaces offer IDs by round: ID = round*stride + slot + 1.
	dealIDStride = 16
)

// FactionDeal is one offer, as saved. Give is what you pay. Res and Amt say
// the gift: a boost's resource and ticks, a research gift's work, the number
// of people, a know-how's resource.
type FactionDeal struct {
	ID      int     `json:"id"`
	Kind    string  `json:"kind"`
	Give    string  `json:"give"`
	GiveAmt float64 `json:"give_amt"`
	Res     string  `json:"res,omitempty"`
	Amt     float64 `json:"amt,omitempty"`
	Taken   bool    `json:"taken,omitempty"`
}

// DealInfo is one offer for the UI. Num is what `diplomacy accept` takes.
type DealInfo struct {
	Num     int
	ID      int
	Kind    string
	Give    string
	GiveAmt float64
	Res     string
	Amt     float64
	Taken   bool
}

// dealBlocked is why a civ will not trade with you ("" if it will).
func dealBlocked(fs FactionState) string {
	switch {
	case fs.AtWar:
		return rowid.Line("game.deals.dealBlocked.01", "at war with you")
	case fs.Status == "embargo":
		return rowid.Line("game.deals.dealBlocked.02", "under your embargo")
	case fs.Status == "rival":
		return rowid.Line("game.deals.dealBlocked.03", "your rival")
	case fs.Opinion <= dealHostileOpinion:
		return rowid.Sprintf("game.deals.dealBlocked.04", "hostile (opinion %d)", fs.Opinion)
	}
	return ""
}

// dealTier is your standing for offers: 0 neutral, 1 friendly, 2 allied.
func dealTier(fs FactionState) int {
	switch {
	case fs.Status == "allied":
		return 2
	case fs.Status == "friendly" || fs.Opinion >= 25:
		return 1
	}
	return 0
}

// dealSlots is how many offers a civ makes at a standing tier.
func dealSlots(personality string, tier int) int {
	if personality == "isolationist" {
		if tier == 2 {
			return 2
		}
		return 1
	}
	n := 2 + tier
	switch personality {
	case "mercantile":
		n++
	case "aggressive":
		n--
	}
	return max(n, 1)
}

// slots is dealSlots with what the techs add to every civilization's set.
func (env dealEnv) slots(personality string, tier int) int {
	return dealSlots(personality, tier) + max(env.extraSlots, 0)
}

// dealMult is a gift as a multiple of parity.
func dealMult(personality string, tier int) float64 {
	m := dealRateNeutral
	switch tier {
	case 1:
		m = dealRateFriendly
	case 2:
		m = dealRateAllied
	}
	switch personality {
	case "mercantile":
		m += dealMercantileBonus
	case "aggressive":
		m -= dealAggressivePenalty
	}
	return math.Min(dealRateMax, math.Max(dealRateMin, m))
}

// dealLot is a slot's size in price units of the age; r is its third draw.
func dealLot(def config.FactionDef, tier int, r float64) float64 {
	str := min(max(def.Strength, 1), 5)
	lot := dealLotBase * (dealLotStrengthBase + float64(dealLotPerStrength*float64(str)))
	switch tier {
	case 1:
		lot *= dealLotFriendly
	case 2:
		lot *= dealLotAllied
	}
	if def.Personality == "mercantile" {
		lot *= dealLotMercantile
	}
	return lot * (1 - dealLotJitter/2 + float64(dealLotJitter*r))
}

// roundDeal rounds a deal amount to 3 significant figures, up for what you
// pay and down for what you get, so rounding never tips a deal your way.
func roundDeal(v float64, up bool) float64 {
	if !(v > 0) || math.IsInf(v, 0) {
		return 0
	}
	mag := detmath.Pow(10, 2-math.Floor(detmath.Log10(v)))
	if up {
		return math.Ceil(v*mag) / mag
	}
	return math.Floor(v*mag) / mag
}

// dealEnv is what offer generation reads of your empire.
type dealEnv struct {
	// rules prices the offers: the engine's ruleset (nil reads the core set).
	rules *rules.Set
	age   string
	// extraSlots is what the techs add to the offers every civilization
	// makes in a set (config.MechanicDealSlots).
	extraSlots int
	unlocked   map[string]bool
	amount     map[string]float64
	// size is, per resource, what a good stock of it comes to for the town:
	// minutes of the town's own income of it in the current age
	// (rules.Set.StockOf). A price is capped at a share of it and the
	// resources held most are the ones whose amount is the most of it.
	size map[string]float64
	// income is what the town makes of each resource a tick.
	income map[string]float64
	// researchLeft is what the tech in progress still needs (0 with none).
	researchLeft float64
	// room is how many people the housing has room for; headFood is the food
	// one eats a tick.
	room     int
	cap      int // the homes' worker capacity: an offer is at most dealPeopleShare of it
	headFood float64
	// boostTicks is how long a boost runs; refresh how long a set of offers lasts.
	boostTicks, refresh int
}

// newDealEnv captures the engine's side of offer generation. Under the lock.
//
// What counts as unlocked is the age table's list up to the current age,
// the set a save records (getUnlockedState), rather than the live flags, so
// a loaded game rolls exactly what the saved one would have.
func (ge *GameEngine) newDealEnv() dealEnv {
	env := dealEnv{
		rules:      ge.rules,
		extraSlots: int(ge.Research.Mechanic(config.MechanicDealSlots)),
		age:        ge.age,
		unlocked:   map[string]bool{},
		amount:     map[string]float64{},
		size:       map[string]float64{},
		income:     map[string]float64{},
		headFood:   ge.Workers.FoodCostPerWorker(),
		boostTicks: ge.stretchTicks(dealBoostTicks),
		refresh:    dealRefreshIn(ge.rules, ge.age),
	}
	for _, k := range ge.getUnlockedState().Resources {
		env.unlocked[k] = true
	}
	for _, k := range ge.Resources.order {
		env.amount[k] = ge.Resources.Get(k)
		env.income[k] = ge.incomeOf(k)
		env.size[k] = ge.rules.StockOf(k, ge.age, 1, env.income[k])
	}
	if ge.Workers.IsUnlocked() {
		env.cap = ge.workerCapLocked()
		env.room = max(env.cap-ge.Workers.TotalPop(), 0)
	}
	env.researchLeft = ge.researchLeftLocked()
	return env
}

// researchLeftLocked is the work the tech in progress still needs (0 with none).
func (ge *GameEngine) researchLeftLocked() float64 {
	if ge.Research.currentTech == "" {
		return 0
	}
	return max(ge.Research.defs[ge.Research.currentTech].Cost-ge.Research.progress, 0)
}

// top is up to dealWantPool of res by descending score, ties by key.
func top(res []string, score func(string) float64) []string {
	out := slices.Clone(res)
	slices.SortFunc(out, func(a, b string) int {
		return cmp.Or(cmp.Compare(score(b), score(a)), cmp.Compare(a, b))
	})
	return out[:min(len(out), dealWantPool)]
}

// fullest is up to dealWantPool of res, the most held first (amount over a
// good stock of it, then key), which is what a price is paid from.
func (env dealEnv) fullest(res []string) []string {
	return top(res, func(k string) float64 {
		if s := env.size[k]; s > 0 {
			return env.amount[k] / s
		}
		return 0
	})
}

// dealPick is the element of s a draw in [0,1) selects.
func dealPick(s []string, r float64) string {
	i := int(r * float64(len(s)))
	return s[min(max(i, 0), len(s)-1)]
}

// payWith picks the good an offer is paid in: one of your fullest priced
// resources (amount over a good stock of it) other than the civ's specialty
// and skip, by draw r. lv is its price level; "" when there is none.
func (env dealEnv) payWith(def config.FactionDef, skip string, r float64) (give string, lv float64) {
	cands := env.payCands(def, skip)
	if len(cands) == 0 {
		return "", 0
	}
	give = dealPick(cands, r)
	return give, orCore(env.rules).DealPriceLevel(give, env.age)
}

// payCands is the goods payWith chooses from: your fullest priced resources
// other than the civ's specialty and skip.
func (env dealEnv) payCands(def config.FactionDef, skip string) []string {
	var cands []string
	for _, res := range orCore(env.rules).PricedResources(env.age) {
		if res != def.Specialty && res != skip && env.unlocked[res] {
			cands = append(cands, res)
		}
	}
	return env.fullest(cands)
}

// rollFactionDeals rolls a civ's offers for one round. Blocked civs get none
// and draw nothing; otherwise every slot draws exactly three values from rng.
func rollFactionDeals(def config.FactionDef, fs FactionState, env dealEnv, rng *rand.Rand, round int) []FactionDeal {
	if dealBlocked(fs) != "" {
		return nil
	}
	tier := dealTier(fs)
	var out []FactionDeal
	people := false // a civ offers people once a set, so they do not crowd out the rest
	for slot := 0; slot < env.slots(def.Personality, tier); slot++ {
		r1, r2, r3 := rng.Float64(), rng.Float64(), rng.Float64()
		d, ok := makeDeal(def, fs, tier, env, slot == 0, !people, r1, r2, r3)
		if !ok || hasSameDeal(out, d) {
			continue
		}
		people = people || d.Kind == DealPeople
		d.ID = round*dealIDStride + slot + 1
		out = append(out, d)
	}
	return out
}

// hasSameDeal reports whether ds already offers d's gift.
func hasSameDeal(ds []FactionDeal, d FactionDeal) bool {
	for _, x := range ds {
		if x.Kind == d.Kind && x.Give == d.Give && x.Res == d.Res {
			return true
		}
	}
	return false
}

// knowhowOpen reports whether the civ may offer its know-how: earned by
// standing (the cap, or allied), not yet had, with a specialty that has a price.
func knowhowOpen(def config.FactionDef, fs FactionState, env dealEnv) bool {
	return !fs.Knowhow && (fs.Status == "allied" || fs.Opinion >= dealStandingCap) &&
		env.unlocked[def.Specialty] && orCore(env.rules).DealPriceLevel(def.Specialty, env.age) > 0
}

// makeDeal builds one slot's offer. A civ that has earned it offers its
// know-how in the first slot; otherwise the first draw picks the gift (a boost
// half the time, research a quarter, people a quarter, people only when
// morePeople); a gift the town cannot use falls to the next.
func makeDeal(def config.FactionDef, fs FactionState, tier int, env dealEnv, first, morePeople bool, r1, r2, r3 float64) (FactionDeal, bool) {
	if first && knowhowOpen(def, fs, env) {
		if give, lv := env.payWith(def, "", r3); lv > 0 {
			return FactionDeal{Kind: DealKnowhow, Give: give, GiveAmt: roundDeal(float64(dealKnowhowUnits*lv), true), Res: def.Specialty}, true
		}
	}
	mult, lot := dealMult(def.Personality, tier), dealLot(def, tier, r3)
	makers := [...]func() (FactionDeal, bool){
		func() (FactionDeal, bool) { return boostDeal(def, env, mult, r2, r3) },
		func() (FactionDeal, bool) { return researchDeal(def, env, mult, lot, r3) },
		func() (FactionDeal, bool) { return peopleDeal(def, env, mult, lot, r3) },
	}
	if !morePeople {
		makers[2] = func() (FactionDeal, bool) { return FactionDeal{}, false }
	}
	start := 0
	if r1 >= 0.5 {
		start = 1 + int((r1-0.5)*4)
	}
	for i := range makers {
		if d, ok := makers[(start+i)%len(makers)](); ok {
			return d, true
		}
	}
	return FactionDeal{}, false
}

// boostDeal is a timed boost of one of your top producers (the civ's
// specialty when it is making any and the draw prefers it), paid at
// dealBoostWorth of the output it adds over its run, over mult.
func boostDeal(def config.FactionDef, env dealEnv, mult, r2, r3 float64) (FactionDeal, bool) {
	set := orCore(env.rules)
	var cands []string
	for _, res := range slices.Sorted(maps.Keys(env.income)) {
		if env.unlocked[res] && env.income[res] > 0 && set.DealPriceLevel(res, env.age) > 0 {
			cands = append(cands, res)
		}
	}
	tops := top(cands, func(k string) float64 { return env.income[k] / set.DealPriceLevel(k, env.age) })
	if len(tops) == 0 {
		return FactionDeal{}, false
	}
	res := dealPick(tops, r2)
	if r2 < 0.5 && slices.Contains(cands, def.Specialty) {
		res = def.Specialty
	}
	// Price units of output the boost adds.
	units := float64(float64(dealBoostShare*env.income[res])*float64(env.boostTicks)) / set.DealPriceLevel(res, env.age)
	// Pay in the drawn good, or the next fullest one the price fits in: a boost
	// of a big producer is dear and often too much for the drawn good's stock.
	pays := env.payCands(def, res)
	for i := range pays {
		give := pays[(int(r3*float64(len(pays)))+i)%len(pays)]
		d := FactionDeal{Kind: DealBoost, Give: give, GiveAmt: roundDeal(float64(units*dealBoostWorth/mult*set.DealPriceLevel(give, env.age)), true), Res: res, Amt: float64(env.boostTicks)}
		if d.GiveAmt >= 1 && d.GiveAmt <= float64(dealGiveShare*env.size[give]) {
			return d, true
		}
	}
	return FactionDeal{}, false
}

// researchDeal is progress on the tech in progress: dealResearchWorth of the
// lot's value, at research's level and mult, no more than the tech still needs.
func researchDeal(def config.FactionDef, env dealEnv, mult, lot, r3 float64) (FactionDeal, bool) {
	set := orCore(env.rules)
	lr := set.DealPriceLevel(config.ResearchTarget, env.age)
	give, lv := env.payWith(def, "", r3)
	if env.researchLeft <= 0 || lr <= 0 || lv <= 0 {
		return FactionDeal{}, false
	}
	per := float64(dealResearchWorth * mult * lr) // research per price unit paid
	units := min(lot, env.researchLeft/per, float64(dealGiveShare*env.size[give])/lv)
	d := FactionDeal{Kind: DealResearch, Give: give, GiveAmt: roundDeal(float64(units*lv), true), Res: config.ResearchTarget, Amt: roundDeal(float64(units*per), false)}
	return d, d.GiveAmt >= 1 && d.Amt > 0
}

// peopleDeal is workers for the idle pool: the lot's value times mult over
// the value of one worker (the food it eats a tick for one rotation of
// offers, at food's level), within the housing.
func peopleDeal(def config.FactionDef, env dealEnv, mult, lot, r3 float64) (FactionDeal, bool) {
	set := orCore(env.rules)
	lf := set.DealPriceLevel("food", env.age)
	give, lv := env.payWith(def, "", r3)
	if env.room < 1 || lf <= 0 || lv <= 0 {
		return FactionDeal{}, false
	}
	head := float64(env.headFood*float64(env.refresh)) / lf // price units of one worker
	n := min(math.Floor(float64(lot*mult)/head), math.Floor(float64(dealGiveShare*env.size[give])/lv*mult/head), float64(env.room), math.Max(1, math.Floor(float64(dealPeopleShare*float64(env.cap)))))
	d := FactionDeal{Kind: DealPeople, Give: give, GiveAmt: roundDeal(float64(n*head/mult*lv), true), Res: "workers", Amt: n}
	return d, n >= 1 && d.GiveAmt >= 1
}

// dealRefreshIn is how long an offer set lasts in age: dealRefreshTicks
// stretched like the age (set.StretchTicks), so an age sees as many
// rotations as it did before the one-week curve.
func dealRefreshIn(set *rules.Set, age string) int {
	return set.StretchTicks(age, dealRefreshTicks)
}

// ===== Engine side =====

// tickFactionDeals advances every met civ's offer timer by a tick of live
// play and re-rolls the offers that are due: never rolled, rolled for an
// earlier age, or dealRefreshIn(age) old. Roster order, so the rng draws are
// the same every run. Called from processDiplomacy under the write lock;
// offline catch-up never calls it (see the rules above).
func (ge *GameEngine) tickFactionDeals() {
	var env *dealEnv
	refresh := ge.Diplomacy.dealRefreshFor(ge.age)
	for _, def := range ge.Diplomacy.factionList {
		fs, ok := ge.Diplomacy.factions[def.Key]
		if !ok || !fs.Discovered {
			continue
		}
		fs.DealTicks++
		if fs.DealRound > 0 && fs.DealsFor == ge.age && fs.DealTicks < refresh {
			continue
		}
		if env == nil {
			e := ge.newDealEnv()
			env = &e
		}
		fs.DealRound++
		fs.DealTicks = 0
		fs.DealsFor = ge.age
		fs.Deals = rollFactionDeals(def, *fs, *env, ge.gameRNG(), fs.DealRound)
	}
}

// dealProblem is why deal d of civ fs can't be taken now, split into what
// waiting for resources would fix (short: the resource to wait for) and what
// it wouldn't (blocked). Both empty: it can be taken.
func (ge *GameEngine) dealProblem(fs *FactionState, d FactionDeal) (blocked, short string) {
	if why := dealBlocked(*fs); why != "" {
		return rowid.Line("game.deals.dealProblem.01", "they are ") + why, ""
	}
	switch {
	case d.Taken:
		return rowid.Line("game.deals.dealProblem.02", "already taken"), ""
	case fs.DealsFor != ge.age:
		return rowid.Line("game.deals.dealProblem.03", "the offer lapsed with the age"), ""
	case !ge.Resources.IsUnlocked(d.Give):
		return ResourceName(d.Give) + rowid.Line("game.deals.dealProblem.04", " is not unlocked yet"), ""
	case d.Kind == DealKnowhow && fs.Knowhow:
		return rowid.Line("game.deals.dealProblem.05", "you already have their know-how"), ""
	case d.Kind == DealKnowhow && fs.Status != "allied" && fs.Opinion < dealStandingCap:
		return rowid.Line("game.deals.dealProblem.08", "they do not trust you enough for their know-how"), ""
	case d.Kind == DealResearch && ge.Research.currentTech == "":
		return rowid.Line("game.deals.dealProblem.06", "nothing is being researched"), ""
	case d.Kind == DealBoost && !(ge.incomeOf(d.Res) > 0):
		return ResourceName(d.Res) + rowid.Line("game.deals.dealProblem.09", " is not being made"), ""
	case d.Kind == DealPeople && ge.workerCapLocked()-ge.Workers.TotalPop() < int(d.Amt):
		return rowid.Line("game.deals.dealProblem.07", "there is no housing for the people"), ""
	}
	if !(ge.Resources.Get(d.Give) >= d.GiveAmt) {
		return "", d.Give
	}
	return "", ""
}

// findDeal returns civ key's state and the index of its deal with the given
// ID (-1 when it is gone).
func (ge *GameEngine) findDeal(key string, id int) (*FactionState, int) {
	fs, ok := ge.Diplomacy.factions[key]
	if !ok || !fs.Discovered {
		return nil, -1
	}
	for i, d := range fs.Deals {
		if d.ID == id {
			return fs, i
		}
	}
	return fs, -1
}

// takeDeal applies deal i of civ def: pays, gives the gift, raises standing
// (to dealStandingCap at most) and marks it taken. The caller has checked
// dealProblem. Under the write lock.
func (ge *GameEngine) takeDeal(def config.FactionDef, fs *FactionState, i int) FactionDeal {
	d := &fs.Deals[i]
	ge.useFeature(config.FeatureDiplomacy)
	// A research gift that now lands for less than offered (the tech needs
	// less than it did) is paid for in proportion.
	pay, amt := d.GiveAmt, d.Amt
	if d.Kind == DealResearch {
		if amt = min(d.Amt, ge.researchLeftLocked()); amt < d.Amt {
			pay = min(d.GiveAmt, roundDeal(float64(d.GiveAmt*amt)/d.Amt, true))
		}
	}
	ge.Resources.Remove(d.Give, pay)
	switch d.Kind {
	case DealBoost:
		ge.addBoost(d.Res, int(d.Amt))
	case DealResearch:
		ge.Stats.ResearchWork += amt
		ge.Research.progress += amt
		if done := ge.Research.CompleteIfFull(); done != "" {
			ge.finishResearch(done)
		}
	case DealPeople:
		if ge.Workers.Recruit(int(d.Amt), ge.workerCapLocked()) {
			ge.Stats.RecordRecruit(int(d.Amt))
			ge.holdStaffing()
		}
	case DealKnowhow:
		fs.Knowhow = true
		ge.recalculateRates()
	}
	fs.TradeCount++
	if fs.Opinion < dealStandingCap {
		fs.Opinion = min(fs.Opinion+dealTradeStanding, dealStandingCap)
	}
	fs.warm()
	d.Taken = true
	ge.note(config.BadgeEvDeal, def.Key)
	ge.addLog(OfficeTrade, "success", rowid.Sprintf("game.deals.takeDeal.01", "Deal with the %s. %s.", def.Name, DealTerms(d.Kind, d.Give, pay, d.Res, amt, textfmt.Number, ge.durationLocked)), logSubj(def.Key), logPaid(LogAmt{d.Give, pay}))
	return *d
}

// DealKindLabel is an offer's kind from the player's side: the gift.
func DealKindLabel(kind string) string {
	switch kind {
	case DealResearch:
		return rowid.Line("game.deals.DealKindLabel.01", "Research")
	case DealPeople:
		return rowid.Line("game.deals.DealKindLabel.02", "People")
	case DealKnowhow:
		return rowid.Line("game.deals.DealKindLabel.03", "Know-how")
	}
	return rowid.Line("game.deals.DealKindLabel.04", "Boost")
}

// DealGift is what an offer gives you: "+25% stone 20m", "research +3.2K",
// "+4 workers", "know-how: +5% stone". num formats amounts, dur a number of
// ticks as a time.
func DealGift(kind, res string, amt float64, num func(float64) string, dur func(int) string) string {
	switch kind {
	case DealResearch:
		return rowid.Sprintf("game.deals.DealGift.01", "research +%s", num(amt))
	case DealPeople:
		return rowid.Sprintf("game.deals.DealGift.02", "+%s workers", num(amt))
	case DealKnowhow:
		return rowid.Sprintf("game.deals.DealGift.03", "know-how: +%s%% %s", num(float64(dealKnowhowShare*100)), ResourceName(res))
	}
	return rowid.Sprintf("game.deals.DealGift.04", "+%s%% %s %s", num(float64(dealBoostShare*100)), ResourceName(res), dur(int(amt)))
}

// DealTerms is how every surface words an offer, from the player's side:
// "Boost: give 876M iron ore → +25% stone 20m". The Factions panel,
// `diplomacy deals` and `diplomacy accept` use it, and the log.
func DealTerms(kind, give string, giveAmt float64, res string, amt float64, num func(float64) string, dur func(int) string) string {
	return rowid.Sprintf("game.deals.DealTerms.01", "%s: give %s %s → %s", DealKindLabel(kind), num(giveAmt), ResourceName(give), DealGift(kind, res, amt, num, dur))
}

// AcceptFactionDeal takes offer n (1-based, as the Factions panel and
// `diplomacy deals` number them) of the civ key.
func (ge *GameEngine) AcceptFactionDeal(key string, n int) (FactionDeal, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	def, fs, ok := ge.Diplomacy.metCiv(key)
	if !ok {
		return FactionDeal{}, ge.Diplomacy.errUnknownCiv(key)
	}
	if err := ge.featureErr(config.FeatureDiplomacy); err != nil {
		return FactionDeal{}, err
	}
	if why := dealBlocked(*fs); why != "" {
		return FactionDeal{}, rowid.Errorf("game.deals.AcceptFactionDeal.01", "The %s won't trade: they are %s.", def.Name, why)
	}
	if fs.DealsFor != ge.age || len(fs.Deals) == 0 {
		return FactionDeal{}, rowid.Errorf("game.deals.AcceptFactionDeal.02", "The %s have no offers right now. New offers arrive with each rotation.", def.Name)
	}
	if n < 1 || n > len(fs.Deals) {
		return FactionDeal{}, rowid.Errorf("game.deals.AcceptFactionDeal.03", "There is no deal %d with the %s (they offer %d). Type diplomacy deals to see them.", n, def.Name, len(fs.Deals))
	}
	d := fs.Deals[n-1]
	blocked, short := ge.dealProblem(fs, d)
	switch {
	case blocked != "":
		return FactionDeal{}, rowid.Errorf("game.deals.AcceptFactionDeal.04", "Can't take deal %d with the %s: %s.", n, def.Name, blocked)
	case short != "":
		return FactionDeal{}, rowid.Errorf("game.deals.AcceptFactionDeal.05", "Not enough %s for deal %d: need %s, have %s.", ResourceName(short), n, textfmt.Number(d.GiveAmt), textfmt.Number(ge.Resources.Get(short)))
	}
	return ge.takeDeal(def, fs, n-1), nil
}

// dealInfos is a civ's current offers for the UI, numbered from 1. Offers
// rolled for an earlier age are not shown (they re-roll on the next tick).
func dealInfos(fs *FactionState, age string) []DealInfo {
	if fs.DealsFor != age || dealBlocked(*fs) != "" {
		return nil
	}
	out := make([]DealInfo, 0, len(fs.Deals))
	for i, d := range fs.Deals {
		out = append(out, DealInfo{Num: i + 1, ID: d.ID, Kind: d.Kind, Give: d.Give, GiveAmt: d.GiveAmt, Res: d.Res, Amt: d.Amt, Taken: d.Taken})
	}
	return out
}
