package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
	"github.com/espresso20/ageforge/pkg/textfmt"
)

// Faction trade deals: each civilization you have met offers a small,
// rotating set of trades, derived from its data the way factionProfile
// derives its boons. There are no per-civ tables; a civ's specialty,
// personality and strength and your standing with it decide everything.
//
// Rules (site/docs/trade.md, Trade deals):
//
//   - Kinds (keys are the civ's side; the player sees Buy, Sell, Favor and
//     Rare, see DealTerms). A "sell" deal is the civ's specialty for one of your
//     construction resources. A "want" deal is the civ asking for one of the
//     resources you hold most of (fullest stores first) and paying in its
//     specialty, a little better than a sell. A "favor" deal pays for the
//     same kind of ask in standing instead of goods. A "rare" deal is an
//     isolationist's: a construction resource of the NEXT age that this
//     age's market does not price, at a steep 0.6 of its next-age parity.
//   - Rates. Where the market trades the pair, a deal beats the market by
//     5-25% (its rate is the market's scaled by mult / (1 - ExchangeFee),
//     mult 0.84-1.0 of parity). Where it doesn't, the rate is the ratio of
//     the two resources' price levels (config.DealPriceLevel, which also
//     values flow resources) times the same mult. A deal never beats parity.
//   - Personality. Mercantile civs offer one more deal, at better rates and
//     in bigger lots, and ask more than they sell. Aggressive civs offer one
//     fewer at worse rates and ask for favours more. Isolationists offer one
//     deal (two when allied), a rare one when the next age has a good for
//     them to sell, their specialty otherwise.
//   - Standing. Neutral civs offer 2 deals, friendly (friendly status or
//     opinion 25+) 3, allied 4; rates and lot sizes rise with it. Civs at war,
//     under embargo, rival or with opinion at or below -50 offer nothing and
//     accept nothing. Standing from deals stops at opinion 50: they can bring
//     a civ to the edge of an alliance, never past it, and never out of a war.
//   - Size. A lot is 1.5 price units of the age, times 0.9 + 0.1 per point of
//     strength, the standing and personality factors and a 0.75-1.25 roll,
//     capped so the goods fit in half your store and the price in 80% of it.
//   - Refresh. Offers rotate every dealRefreshTicks ticks of play and on an
//     age advance. The timer counts ticks of live play only: offline catch-up
//     does not advance it, so a player who checks in finds the offers they
//     left (and a plan can take one while they are away).
//   - Determinism. Rolls come off the engine's seeded rng in roster order,
//     three draws per slot whether or not the slot yields a deal. A blocked
//     civ draws nothing. Offers, taken flags and the timer are saved.

// Deal kinds.
const (
	DealSell  = "sell"
	DealWant  = "want"
	DealFavor = "favor"
	DealRare  = "rare"
)

// Deal tuning, grouped so balancing is a data edit.
const (
	// dealRefreshTicks is how many ticks of live play an offer set lasts
	// (an hour at 1x).
	dealRefreshTicks = 1800

	// dealHostileOpinion: at or below this opinion a civ will not trade.
	dealHostileOpinion = -50
	// dealStandingCap is the opinion deals can raise a civ to, no further.
	dealStandingCap = 50
	// dealFavorStanding is what a favor deal pays; dealTradeStanding what
	// any other accepted deal adds.
	dealFavorStanding = 5
	dealTradeStanding = 1

	// Rates as a multiple of parity (the market pays 1 - ExchangeFee = 0.8).
	dealRateNeutral       = 0.88
	dealRateFriendly      = 0.92
	dealRateAllied        = 0.96
	dealWantBonus         = 0.03
	dealMercantileBonus   = 0.02
	dealAggressivePenalty = 0.04
	dealRateMin           = 0.84
	dealRateMax           = 1.0
	dealRareRate          = 0.6

	// Lot size in price units of the age.
	dealLotBase         = 1.5
	dealLotStrengthBase = 0.9
	dealLotPerStrength  = 0.1
	dealLotFriendly     = 1.25
	dealLotAllied       = 1.5
	dealLotMercantile   = 1.2
	dealLotRare         = 2.0
	dealLotJitter       = 0.5 // lots vary from 1 - J/2 to 1 + J/2
	// dealGetShare / dealGiveShare cap a lot at these shares of your store.
	dealGetShare  = 0.5
	dealGiveShare = 0.8
	// dealWantPool is how many of your fullest stores a want or favor deal
	// chooses from.
	dealWantPool = 3

	// dealIDStride spaces deal IDs by round: ID = round*stride + slot + 1.
	dealIDStride = 16
)

// FactionDeal is one offer, as saved. Give is what you pay; Get what you
// receive ("" for a favor deal, which pays Standing instead).
type FactionDeal struct {
	ID       int     `json:"id"`
	Kind     string  `json:"kind"`
	Give     string  `json:"give"`
	GiveAmt  float64 `json:"give_amt"`
	Get      string  `json:"get,omitempty"`
	GetAmt   float64 `json:"get_amt,omitempty"`
	Standing int     `json:"standing,omitempty"`
	Taken    bool    `json:"taken,omitempty"`
}

// DealInfo is one offer for the UI. Num is what `diplomacy accept` takes.
type DealInfo struct {
	Num      int
	ID       int
	Kind     string
	Give     string
	GiveAmt  float64
	Get      string
	GetAmt   float64
	Standing int
	Taken    bool
	// Edge is how much better than the market the deal pays (0.15 = 15%);
	// 0 when the market does not trade the pair.
	Edge float64
}

// dealBlocked is why a civ will not trade with you ("" if it will).
func dealBlocked(fs FactionState) string {
	switch {
	case fs.AtWar:
		return "at war with you"
	case fs.Status == "embargo":
		return "under your embargo"
	case fs.Status == "rival":
		return "your rival"
	case fs.Opinion <= dealHostileOpinion:
		return fmt.Sprintf("hostile (opinion %d)", fs.Opinion)
	}
	return ""
}

// dealTier is your standing for deals: 0 neutral, 1 friendly, 2 allied.
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

// dealKind picks a slot's kind from its first draw.
func dealKind(personality string, r float64) string {
	var sell, want float64
	switch personality {
	case "isolationist":
		return DealRare
	case "mercantile":
		sell, want = 0.5, 0.9
	case "aggressive":
		sell, want = 0.45, 0.65
	default: // peaceful
		sell, want = 0.55, 0.8
	}
	switch {
	case r < sell:
		return DealSell
	case r < want:
		return DealWant
	}
	return DealFavor
}

// dealMult is a deal's rate as a multiple of parity.
func dealMult(personality string, tier int, kind string) float64 {
	m := dealRateNeutral
	switch tier {
	case 1:
		m = dealRateFriendly
	case 2:
		m = dealRateAllied
	}
	if kind == DealWant {
		m += dealWantBonus
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
func dealLot(def config.FactionDef, tier int, kind string, r float64) float64 {
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
	if kind == DealRare {
		lot *= dealLotRare
	}
	return lot * (1 - dealLotJitter/2 + float64(dealLotJitter*r))
}

// dealRate is what a deal pays per unit of from, given its parity multiple:
// the market's rate scaled by mult / (1 - ExchangeFee) where the market
// trades the pair, the ratio of the deal price levels times mult where it
// doesn't. ok is false when neither applies.
func dealRate(from, to, age string, mult float64) (float64, bool) {
	if m, ok := config.MarketOffers(from, to, age); ok && m > 0 {
		return m * mult / (1 - config.ExchangeFee), true
	}
	lf, lt := config.DealPriceLevel(from, age), config.DealPriceLevel(to, age)
	if lf <= 0 || lt <= 0 {
		return 0, false
	}
	return lt / lf * mult, true
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

// dealEnv is what deal generation reads of your empire.
type dealEnv struct {
	age, next string
	unlocked  map[string]bool
	amount    map[string]float64
	storage   map[string]float64
}

// newDealEnv captures the engine's side of deal generation. Under the lock.
//
// What counts as unlocked is the age table's list up to the current age,
// the set a save records (getUnlockedState), rather than the live flags, so
// a loaded game rolls exactly what the saved one would have.
func (ge *GameEngine) newDealEnv() dealEnv {
	env := dealEnv{
		age:      ge.age,
		next:     ge.progress.GetNextAge(ge.age),
		unlocked: map[string]bool{},
		amount:   map[string]float64{},
		storage:  map[string]float64{},
	}
	for _, k := range ge.getUnlockedState().Resources {
		env.unlocked[k] = true
	}
	for _, k := range ge.Resources.order {
		env.amount[k] = ge.Resources.Get(k)
		env.storage[k] = ge.Resources.GetStorage(k)
	}
	return env
}

// fullest is up to dealWantPool of res, fullest store first (amount over
// storage, then key), which is what a want or favor deal asks for.
func (env dealEnv) fullest(res []string) []string {
	out := append([]string(nil), res...)
	fill := func(r string) float64 {
		if s := env.storage[r]; s > 0 {
			return env.amount[r] / s
		}
		return 0
	}
	// Insertion sort: short lists, stable, no map order.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0; j-- {
			a, b := out[j-1], out[j]
			if fill(b) > fill(a) || (fill(b) == fill(a) && b < a) {
				out[j-1], out[j] = b, a
				continue
			}
			break
		}
	}
	if len(out) > dealWantPool {
		out = out[:dealWantPool]
	}
	return out
}

// dealPick is the element of s a draw in [0,1) selects.
func dealPick(s []string, r float64) string {
	i := int(r * float64(len(s)))
	return s[min(max(i, 0), len(s)-1)]
}

// fit sizes a goods deal: give and get scaled down together until get fits
// dealGetShare of its store and give dealGiveShare of its store, then
// rounded against you. ok is false when a side rounds below 1.
func (env dealEnv) fit(d *FactionDeal, rate float64) bool {
	if d.Get != "" {
		if cap := dealGetShare * env.storage[d.Get]; d.GetAmt > cap {
			d.GetAmt = cap
			d.GiveAmt = d.GetAmt / rate
		}
	}
	if cap := dealGiveShare * env.storage[d.Give]; d.GiveAmt > cap {
		d.GiveAmt = cap
		if d.Get != "" {
			d.GetAmt = d.GiveAmt * rate
		}
	}
	d.GiveAmt = roundDeal(d.GiveAmt, true)
	if d.Get != "" {
		d.GetAmt = roundDeal(d.GetAmt, false)
		if d.GetAmt < 1 {
			return false
		}
	}
	return d.GiveAmt >= 1 && d.GiveAmt <= env.storage[d.Give]
}

// rollFactionDeals rolls a civ's offers for one round. Blocked civs get none
// and draw nothing; otherwise every slot draws exactly three values from rng.
func rollFactionDeals(def config.FactionDef, fs FactionState, env dealEnv, rng *rand.Rand, round int) []FactionDeal {
	if dealBlocked(fs) != "" {
		return nil
	}
	tier := dealTier(fs)
	var out []FactionDeal
	for slot := 0; slot < dealSlots(def.Personality, tier); slot++ {
		r1, r2, r3 := rng.Float64(), rng.Float64(), rng.Float64()
		d, ok := makeDeal(def, fs, tier, env, r1, r2, r3)
		if !ok || hasSameDeal(out, d) {
			continue
		}
		d.ID = round*dealIDStride + slot + 1
		out = append(out, d)
	}
	return out
}

// hasSameDeal reports whether ds already offers d's trade.
func hasSameDeal(ds []FactionDeal, d FactionDeal) bool {
	for _, x := range ds {
		if x.Kind == d.Kind && x.Give == d.Give && x.Get == d.Get {
			return true
		}
	}
	return false
}

// makeDeal builds one slot's offer. A kind that has nothing to offer falls
// back to a sell (for a rare), then to a favor.
func makeDeal(def config.FactionDef, fs FactionState, tier int, env dealEnv, r1, r2, r3 float64) (FactionDeal, bool) {
	kind := dealKind(def.Personality, r1)
	canFavor := fs.Opinion < dealStandingCap
	if kind == DealFavor && !canFavor {
		kind = DealWant
	}
	lot := dealLot(def, tier, kind, r3)
	switch kind {
	case DealRare:
		if d, ok := rareDeal(def, env, lot, r2); ok {
			return d, true
		}
		if d, ok := goodsDeal(def, tier, env, DealSell, dealLot(def, tier, DealSell, r3), r2); ok {
			return d, true
		}
	case DealSell, DealWant:
		if d, ok := goodsDeal(def, tier, env, kind, lot, r2); ok {
			return d, true
		}
	}
	if canFavor {
		return favorDeal(def, env, dealLot(def, tier, DealFavor, r3), r2)
	}
	return FactionDeal{}, false
}

// goodsDeal is a sell or want deal: the civ's specialty for one of your
// construction resources. A sell takes any of them (the draw picks); a want
// takes one of your fullest.
func goodsDeal(def config.FactionDef, tier int, env dealEnv, kind string, lot, r float64) (FactionDeal, bool) {
	get := def.Specialty
	if get == "" || !env.unlocked[get] {
		return FactionDeal{}, false
	}
	mult := dealMult(def.Personality, tier, kind)
	var cands []string
	for _, res := range config.PricedResources(env.age) {
		if res == get || !env.unlocked[res] {
			continue
		}
		if _, ok := dealRate(res, get, env.age, mult); ok {
			cands = append(cands, res)
		}
	}
	if kind == DealWant {
		cands = env.fullest(cands)
	}
	if len(cands) == 0 {
		return FactionDeal{}, false
	}
	give := dealPick(cands, r)
	rate, _ := dealRate(give, get, env.age, mult)
	d := FactionDeal{Kind: kind, Give: give, Get: get}
	if kind == DealSell {
		// A sell is sized by the goods, a want by what it asks for.
		if lv := config.DealPriceLevel(get, env.age); lv > 0 {
			d.GetAmt = lot * lv
			d.GiveAmt = d.GetAmt / rate
		}
	}
	if d.GiveAmt == 0 {
		d.GiveAmt = lot * config.DealPriceLevel(give, env.age)
		d.GetAmt = d.GiveAmt * rate
	}
	if !env.fit(&d, rate) {
		return FactionDeal{}, false
	}
	return d, true
}

// favorDeal is an ask paid in standing: one of your fullest construction
// resources (not the civ's own specialty) for dealFavorStanding opinion.
func favorDeal(def config.FactionDef, env dealEnv, lot, r float64) (FactionDeal, bool) {
	var cands []string
	for _, res := range config.PricedResources(env.age) {
		if res != def.Specialty && env.unlocked[res] {
			cands = append(cands, res)
		}
	}
	cands = env.fullest(cands)
	if len(cands) == 0 {
		return FactionDeal{}, false
	}
	give := dealPick(cands, r)
	d := FactionDeal{Kind: DealFavor, Give: give, GiveAmt: lot * config.DealPriceLevel(give, env.age), Standing: dealFavorStanding}
	if !env.fit(&d, 0) {
		return FactionDeal{}, false
	}
	return d, true
}

// rareDeal is an isolationist's offer: a construction resource of the next
// age that this age's market does not price and that you can already hold,
// for one of yours that both ages price, at dealRareRate of next-age parity.
func rareDeal(def config.FactionDef, env dealEnv, lot, r float64) (FactionDeal, bool) {
	if env.next == "" {
		return FactionDeal{}, false
	}
	var gives []string
	for _, res := range config.PricedResources(env.age) {
		if env.unlocked[res] && config.DealPriceLevel(res, env.next) > 0 {
			gives = append(gives, res)
		}
	}
	var gets []string
	for _, res := range config.PricedResources(env.next) {
		if !env.unlocked[res] || config.DealPriceLevel(res, env.age) > 0 || marketSellsFor(gives, res, env.age) {
			continue
		}
		gets = append(gets, res)
	}
	if len(gets) == 0 || len(gives) == 0 {
		return FactionDeal{}, false
	}
	// One draw picks both: the goods first, the price from what is left of it.
	i := min(int(r*float64(len(gets))), len(gets)-1)
	get := gets[i]
	rest := float64(r*float64(len(gets))) - float64(i)
	give := dealPick(gives, rest)
	rate := config.DealPriceLevel(get, env.next) / config.DealPriceLevel(give, env.next) * dealRareRate
	d := FactionDeal{Kind: DealRare, Give: give, Get: get, GiveAmt: lot * config.DealPriceLevel(give, env.age)}
	d.GetAmt = d.GiveAmt * rate
	if !env.fit(&d, rate) {
		return FactionDeal{}, false
	}
	return d, true
}

// marketSellsFor reports whether the market sells res in age for any of
// gives (then it isn't rare).
func marketSellsFor(gives []string, res, age string) bool {
	for _, g := range gives {
		if _, ok := config.MarketOffers(g, res, age); ok {
			return true
		}
	}
	return false
}

// ===== Engine side =====

// tickFactionDeals advances every met civ's offer timer by a tick of live
// play and re-rolls the offers that are due: never rolled, rolled for an
// earlier age, or dealRefreshTicks old. Roster order, so the rng draws are
// the same every run. Called from processDiplomacy under the write lock;
// offline catch-up never calls it (see the rules above).
func (ge *GameEngine) tickFactionDeals() {
	var env *dealEnv
	for _, def := range ge.Diplomacy.factionList {
		fs, ok := ge.Diplomacy.factions[def.Key]
		if !ok || !fs.Discovered {
			continue
		}
		fs.DealTicks++
		if fs.DealRound > 0 && fs.DealsFor == ge.age && fs.DealTicks < dealRefreshTicks {
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
		return "they are " + why, ""
	}
	if d.Taken {
		return "already taken", ""
	}
	if fs.DealsFor != ge.age {
		return "the offer lapsed with the age", ""
	}
	for _, r := range []string{d.Give, d.Get} {
		if r != "" && !ge.Resources.IsUnlocked(r) {
			return r + " is not unlocked", ""
		}
	}
	if d.Get != "" && ge.Resources.GetStorage(d.Get)-ge.Resources.Get(d.Get) < d.GetAmt {
		return fmt.Sprintf("not enough room for %s %s", textfmt.Number(d.GetAmt), d.Get), ""
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

// takeDeal applies deal i of civ def: pays, receives, raises standing (to
// dealStandingCap at most) and marks it taken. The caller has checked
// dealProblem. Under the write lock.
func (ge *GameEngine) takeDeal(def config.FactionDef, fs *FactionState, i int) FactionDeal {
	d := &fs.Deals[i]
	ge.Resources.Remove(d.Give, d.GiveAmt)
	gain := d.Standing
	if d.Get != "" {
		ge.Resources.Add(d.Get, d.GetAmt)
		fs.TradeCount++
		gain = dealTradeStanding
	}
	if fs.Opinion < dealStandingCap {
		fs.Opinion = min(fs.Opinion+gain, dealStandingCap)
	}
	if fs.Status == "neutral" && fs.Opinion >= 25 {
		fs.Status = "friendly"
	}
	d.Taken = true
	ge.addLog("success", fmt.Sprintf("Deal with the %s (%s).", def.Name, dealTerms(*d)))
	return *d
}

// DealKindLabel is a deal's kind from the player's side: the civ selling
// its specialty is your Buy, the civ wanting your goods is your Sell.
func DealKindLabel(kind string) string {
	switch kind {
	case DealWant:
		return "Sell"
	case DealFavor:
		return "Favor"
	case DealRare:
		return "Rare"
	}
	return "Buy"
}

// DealGets is what a deal pays you: "966K food" or "+5 standing".
func DealGets(get string, getAmt float64, standing int, num func(float64) string) string {
	if get == "" {
		return fmt.Sprintf("+%d standing", standing)
	}
	return num(getAmt) + " " + get
}

// DealTerms is how every surface words a deal, from the player's side:
// "Buy: give 876M coal → get 966K food", "Favor: give 899M steel → get +5
// standing". The Factions panel, `diplomacy deals`, `diplomacy accept`, the
// plan and the log all use it; num formats the amounts.
func DealTerms(kind, give string, giveAmt float64, get string, getAmt float64, standing int, num func(float64) string) string {
	return fmt.Sprintf("%s: give %s %s → get %s", DealKindLabel(kind), num(giveAmt), give, DealGets(get, getAmt, standing, num))
}

// dealTerms is DealTerms for a saved deal, with the log's number format.
func dealTerms(d FactionDeal) string {
	return DealTerms(d.Kind, d.Give, d.GiveAmt, d.Get, d.GetAmt, d.Standing, textfmt.Number)
}

// AcceptFactionDeal takes offer n (1-based, as the Factions panel and
// `diplomacy deals` number them) of the civ key.
func (ge *GameEngine) AcceptFactionDeal(key string, n int) (FactionDeal, error) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	def, ok := ge.Diplomacy.factionDefs[key]
	if !ok {
		return FactionDeal{}, fmt.Errorf("unknown civilization: %s", key)
	}
	fs, ok := ge.Diplomacy.factions[key]
	if !ok || !fs.Discovered {
		return FactionDeal{}, fmt.Errorf("%s has not been discovered yet", def.Name)
	}
	if why := dealBlocked(*fs); why != "" {
		return FactionDeal{}, fmt.Errorf("the %s won't trade: they are %s", def.Name, why)
	}
	if fs.DealsFor != ge.age || len(fs.Deals) == 0 {
		return FactionDeal{}, fmt.Errorf("the %s have no offers right now", def.Name)
	}
	if n < 1 || n > len(fs.Deals) {
		return FactionDeal{}, fmt.Errorf("no deal %d with the %s (they offer %d)", n, def.Name, len(fs.Deals))
	}
	d := fs.Deals[n-1]
	blocked, short := ge.dealProblem(fs, d)
	switch {
	case blocked != "":
		return FactionDeal{}, fmt.Errorf("can't take deal %d with the %s: %s", n, def.Name, blocked)
	case short != "":
		return FactionDeal{}, fmt.Errorf("not enough %s (have %s, need %s)", short, textfmt.Number(ge.Resources.Get(short)), textfmt.Number(d.GiveAmt))
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
		info := DealInfo{Num: i + 1, ID: d.ID, Kind: d.Kind, Give: d.Give, GiveAmt: d.GiveAmt,
			Get: d.Get, GetAmt: d.GetAmt, Standing: d.Standing, Taken: d.Taken}
		if d.Get != "" && d.GiveAmt > 0 {
			if m, ok := config.MarketOffers(d.Give, d.Get, age); ok && m > 0 {
				info.Edge = d.GetAmt/d.GiveAmt/m - 1
			}
		}
		out = append(out, info)
	}
	return out
}
