package game

import (
	"encoding/json"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/pkg/textfmt"
	"github.com/espresso20/ageforge/rules"
)

// testDealEnv is a dealEnv for age with every resource the ages up to it
// unlock, roomy stores, and each store a different amount full (so the
// fullest-first asks have a fixed order).
func testDealEnv(age string) dealEnv {
	order := config.AgeOrder()
	ages := config.AgeByKey()
	env := dealEnv{age: age, unlocked: map[string]bool{}, amount: map[string]float64{}, size: map[string]float64{}}
	for _, a := range order {
		for _, r := range ages[a].UnlockResources {
			env.unlocked[r] = true
		}
		if a == age {
			break
		}
	}
	env.income = map[string]float64{}
	for i, r := range config.BaseResources() {
		env.size[r.Key] = config.Sextillion
		env.amount[r.Key] = float64(i+1) * config.Quintillion
		env.income[r.Key] = float64(i+1) * 1000
	}
	env.researchLeft, env.room, env.headFood, env.boostTicks, env.refresh = 1e12, 50, 0.1, 600, 1800
	return env
}

func factionDef(t *testing.T, key string) config.FactionDef {
	t.Helper()
	def, ok := config.FactionByKey()[key]
	if !ok {
		t.Fatalf("no civ %q", key)
	}
	return def
}

// dealAge is the first age at which civ def's specialty can be traded: its
// MinAge, or later when the specialty is locked there.
func dealAge(def config.FactionDef) string {
	order := config.AgeOrder()
	for i, a := range order {
		if a != def.MinAge {
			continue
		}
		for _, b := range order[i:] {
			if testDealEnv(b).unlocked[def.Specialty] {
				return b
			}
		}
	}
	return def.MinAge
}

var (
	neutralState = FactionState{Discovered: true, Status: "neutral"}
	alliedState  = FactionState{Discovered: true, Status: "allied", Opinion: 60}
)

// dealEngine is a seeded engine in age with every earlier unlock and roomy,
// half-full stores.
func dealEngine(t *testing.T, age string) *GameEngine {
	t.Helper()
	ge := newSeededEngine(11)
	for _, a := range config.AgeOrder() {
		ge.applyAgeUnlocks(a)
		if a == age {
			break
		}
	}
	ge.advanceAge(age)
	ge.pendingCatastrophe = ""
	learn(ge, "envoys") // the tech that opens deals, gifts and alliances
	// The deals below are rolled from a fresh stream, so they do not move
	// when the transition's own draws (its epoch event, the era's fate) do.
	ge.SeedRNG(11)
	for _, r := range ge.Resources.resources {
		r.Amount = 500 * config.Trillion
	}
	// A town that makes wood and has room for people, so boosts and people
	// are both on offer.
	ge.mu.Lock()
	ge.Buildings.counts["wood_camp"] = 20
	ge.permanentBonuses["population"] += 100
	ge.recalculateRates()
	ge.mu.Unlock()
	return ge
}

// meet discovers civ key at opinion and rolls its offers.
func meet(ge *GameEngine, key string, opinion int) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.Diplomacy.DiscoverFaction(key)
	ge.Diplomacy.factions[key].Opinion = opinion
	ge.tickFactionDeals()
}

// takeOffer takes offer n of civ key, as the player would.
func (ge *GameEngine) takeOffer(key string, n int) error {
	_, err := ge.AcceptFactionDeal(key, n)
	return err
}

// TestDeals_RollsEveryCiv: each civ's offers are gifts (never goods for
// goods), paid in goods other than its specialty, within its slots.
func TestDeals_RollsEveryCiv(t *testing.T) {
	gifts := map[string]bool{DealBoost: true, DealResearch: true, DealPeople: true, DealKnowhow: true}
	for _, def := range config.BaseFactions() {
		env := testDealEnv(dealAge(def))
		for _, fs := range []FactionState{neutralState, alliedState} {
			ds := rollFactionDeals(def, fs, env, testRNG(), 1)
			if n := env.slots(def.Personality, dealTier(fs)); len(ds) == 0 || len(ds) > n {
				t.Errorf("%s %s: %d offers, want 1..%d", def.Key, fs.Status, len(ds), n)
			}
			for _, d := range ds {
				if !gifts[d.Kind] || d.Give == def.Specialty || d.Give == d.Res || !(d.GiveAmt >= 1) || !env.unlocked[d.Give] {
					t.Errorf("%s: bad offer %+v", def.Key, d)
				}
				if d.Kind == DealKnowhow && (d.Res != def.Specialty || fs.Status != "allied") {
					t.Errorf("%s: know-how %+v at %s", def.Key, d, fs.Status)
				}
			}
		}
	}
	for _, pers := range []string{"peaceful", "mercantile", "aggressive", "isolationist"} {
		prev := 0.0
		for tier := 0; tier <= 2; tier++ {
			if m := dealMult(pers, tier); m < dealRateMin || m > 1 || m < prev {
				t.Errorf("dealMult(%s, %d) = %v (previous tier %v)", pers, tier, m, prev)
			} else {
				prev = m
			}
		}
	}
	merchant := factionDef(t, "merchant_guild")
	if !(dealLot(merchant, 2, 0.5) > dealLot(merchant, 1, 0.5) && dealLot(merchant, 1, 0.5) > dealLot(merchant, 0, 0.5)) {
		t.Error("lots should grow with standing")
	}
}

// TestDeals_NoGoodsForGoods: no offer kind other than the four gifts exists,
// and the offer carries no goods to receive.
func TestDeals_NoGoodsForGoods(t *testing.T) {
	for _, f := range []string{"Get", "GetAmt"} {
		if _, ok := reflect.TypeOf(FactionDeal{}).FieldByName(f); ok {
			t.Errorf("FactionDeal still has %s", f)
		}
	}
	for _, def := range config.BaseFactions() {
		env := testDealEnv(dealAge(def))
		for seed := int64(0); seed < 30; seed++ {
			for _, d := range rollFactionDeals(def, alliedState, env, rand.New(rand.NewSource(seed)), 1) {
				switch d.Kind {
				case DealBoost, DealResearch, DealPeople, DealKnowhow:
				default:
					t.Fatalf("%s rolled a %q offer", def.Key, d.Kind)
				}
			}
		}
	}
}

// TestDeals_GiftMath pins each gift's formula against the price levels.
func TestDeals_GiftMath(t *testing.T) {
	def := factionDef(t, "riverlands_tribes")
	env := testDealEnv("bronze_age")
	set := rules.Core()
	lv := func(r string) float64 { return set.DealPriceLevel(r, env.age) }
	const mult = 0.9
	near := func(what string, got, want float64) {
		t.Helper()
		if math.Abs(got-want) > 0.01*want {
			t.Errorf("%s = %v, want %v", what, got, want)
		}
	}

	// Boost: 0.6 of the output it adds, over mult, in the paid good's units.
	d, ok := boostDeal(def, env, mult, 0.9, 0.1)
	if !ok || d.Kind != DealBoost || d.Amt != 600 {
		t.Fatalf("boost = %+v, %v", d, ok)
	}
	added := dealBoostShare * env.income[d.Res] * 600 / lv(d.Res)
	near("boost price", d.GiveAmt, added*dealBoostWorth/mult*lv(d.Give))
	if d2, ok := boostDeal(def, env, mult, 0.1, 0.1); !ok || d2.Res != def.Specialty {
		t.Errorf("boost on %q, want the specialty %s", d2.Res, def.Specialty)
	}

	// Research: 1.25 x lot x mult at research's level, for lot units paid.
	d, ok = researchDeal(def, env, mult, 2, 0.1)
	if !ok {
		t.Fatal("no research offer")
	}
	near("research gift", d.Amt, 2*dealResearchWorth*mult*lv(config.ResearchTarget))
	near("research price", d.GiveAmt, 2*lv(d.Give))
	small := env
	small.researchLeft = d.Amt / 4
	if c, ok := researchDeal(def, small, mult, 2, 0.1); !ok || c.Amt > small.researchLeft || c.GiveAmt > d.GiveAmt/3.9 {
		t.Errorf("capped research = %+v (left %v, uncapped %+v)", c, small.researchLeft, d)
	}
	small.researchLeft = 0
	if _, ok := researchDeal(def, small, mult, 2, 0.1); ok {
		t.Error("research offered with nothing being researched")
	}

	// People: lot x mult over a worker's worth (food eaten for one rotation).
	head := env.headFood * float64(env.refresh) / lv("food")
	d, ok = peopleDeal(def, env, mult, 2, 0.1)
	n := math.Floor(2 * mult / head)
	if n > float64(env.room) {
		n = float64(env.room)
	}
	if !ok || d.Amt != n || d.Amt < 1 {
		t.Fatalf("people = %+v (want %v workers)", d, n)
	}
	near("people price", d.GiveAmt, n*head/mult*lv(d.Give))
	tight := env
	tight.room = 1
	if d, _ := peopleDeal(def, tight, mult, 2, 0.1); d.Amt != 1 {
		t.Errorf("people with room for 1 = %v", d.Amt)
	}
	tight.room = 0
	if _, ok := peopleDeal(def, tight, mult, 2, 0.1); ok {
		t.Error("people offered with no room")
	}
}

// TestDeals_NoneWhenHostile: blocked civs offer nothing, draw nothing and
// refuse to trade.
func TestDeals_NoneWhenHostile(t *testing.T) {
	def := factionDef(t, "riverlands_tribes")
	env := testDealEnv("bronze_age")
	for name, fs := range map[string]FactionState{
		"war":     {Discovered: true, Status: "neutral", AtWar: true},
		"embargo": {Discovered: true, Status: "embargo"},
		"rival":   {Discovered: true, Status: "rival"},
		"hostile": {Discovered: true, Status: "neutral", Opinion: dealHostileOpinion},
	} {
		rng := testRNG()
		if dealBlocked(fs) == "" {
			t.Errorf("%s: not blocked", name)
		}
		if ds := rollFactionDeals(def, fs, env, rng, 1); ds != nil {
			t.Errorf("%s: offered %v", name, ds)
		}
		if rng.Int63() != testRNG().Int63() {
			t.Errorf("%s: a blocked civ drew from the rng", name)
		}
	}
}

// giftEngine is a colonial-age engine with a wood camp running, one civ met
// and its offers replaced by deals.
func giftEngine(t *testing.T, deals ...FactionDeal) *GameEngine {
	t.Helper()
	ge := dealEngine(t, "colonial_age")
	meet(ge, "merchant_guild", 0)
	ge.mu.Lock()
	ge.Buildings.counts["wood_camp"] = 20
	ge.Diplomacy.factions["merchant_guild"].Deals = deals
	ge.recalculateRates()
	ge.mu.Unlock()
	return ge
}

func woodRate(ge *GameEngine) float64 {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.Resources.resources["wood"].Rate
}

// TestGifts_BoostRatesAndExpiry: an accepted boost adds a quarter of the
// resource's gross rate as its own breakdown line, extends instead of
// stacking, and ends on its tick.
func TestGifts_BoostRatesAndExpiry(t *testing.T) {
	ge := giftEngine(t,
		FactionDeal{ID: 1, Kind: DealBoost, Give: "stone", GiveAmt: 10, Res: "wood", Amt: 100},
		FactionDeal{ID: 2, Kind: DealBoost, Give: "stone", GiveAmt: 10, Res: "wood", Amt: 50})
	base := woodRate(ge)
	if base <= 0 {
		t.Fatal("no wood production")
	}
	noErr(t, ge.takeOffer("merchant_guild", 1))
	if got := woodRate(ge); math.Abs(got-1.25*base) > 1e-9*base {
		t.Fatalf("boosted wood rate %v, want %v", got, 1.25*base)
	}
	ge.mu.RLock()
	gift := ge.Resources.resources["wood"].Breakdown.GiftRate
	end := ge.boosts[0].End
	ge.mu.RUnlock()
	if math.Abs(gift-0.25*base) > 1e-9*base {
		t.Errorf("breakdown gift line %v, want %v", gift, 0.25*base)
	}
	noErr(t, ge.takeOffer("merchant_guild", 2))
	ge.mu.RLock()
	n, end2 := len(ge.boosts), ge.boosts[0].End
	ge.mu.RUnlock()
	if n != 1 || end2 != end+50 {
		t.Errorf("a second boost made %d boosts ending %d, want one ending %d", n, end2, end+50)
	}
	if got := woodRate(ge); math.Abs(got-1.25*base) > 1e-9*base {
		t.Errorf("two boosts stacked: rate %v", got)
	}
	ge.mu.Lock()
	ge.tick = end2 - 2
	ge.mu.Unlock()
	ge.doTick() // tick end2-1: still running
	if woodRate(ge) < 1.2*base {
		t.Fatal("the boost ended early")
	}
	ge.doTick() // tick end2: over
	ge.mu.RLock()
	left := len(ge.boosts)
	ge.mu.RUnlock()
	if left != 0 || math.Abs(woodRate(ge)-base) > 1e-3*base {
		t.Errorf("after its end: %d boosts, rate %v, want 0 and %v", left, woodRate(ge), base)
	}
}

// TestGifts_BoostOffline: time away counts a boost only while it runs.
func TestGifts_BoostOffline(t *testing.T) {
	ge := giftEngine(t, FactionDeal{ID: 1, Kind: DealBoost, Give: "stone", GiveAmt: 10, Res: "wood", Amt: 100})
	base := woodRate(ge)
	noErr(t, ge.takeOffer("merchant_guild", 1))
	ge.mu.Lock()
	ge.Resources.resources["wood"].Amount = 0
	away := 400 * ge.tickIntervalLocked()
	ge.mu.Unlock()
	ge.SimulateOffline(away)
	ge.mu.RLock()
	got, boosts := ge.Resources.resources["wood"].Amount, len(ge.boosts)
	ge.mu.RUnlock()
	want := base * OfflineEfficiency * (400 + 0.25*100)
	if boosts != 0 || math.Abs(got-want) > 0.05*want {
		t.Errorf("wood after 400 ticks away: %v (boosts left %d), want about %v", got, boosts, want)
	}
}

// TestGifts_ResearchCapped: research lands whole, never past what the tech
// needs, completes it when it fills, and is refused with none in progress.
func TestGifts_ResearchCapped(t *testing.T) {
	ge := giftEngine(t, FactionDeal{ID: 1, Kind: DealResearch, Give: "stone", GiveAmt: 10, Res: config.ResearchTarget, Amt: 1e18})
	if err := ge.takeOffer("merchant_guild", 1); err == nil || !strings.Contains(err.Error(), "nothing is being researched") {
		t.Errorf("research with nothing in progress: %v", err)
	}
	ge.mu.Lock()
	keys := ge.Research.StartableTechs(ge.age, ge.progress.GetAgeOrder())
	ge.mu.Unlock()
	if len(keys) == 0 {
		t.Fatal("nothing to research")
	}
	noErr(t, ge.StartResearch(keys[0]))
	ge.mu.Lock()
	ge.Research.progress = 0
	cost := ge.Research.defs[keys[0]].Cost
	ge.Diplomacy.factions["merchant_guild"].Deals[0].Amt = cost / 4
	ge.mu.Unlock()
	noErr(t, ge.takeOffer("merchant_guild", 1))
	ge.mu.RLock()
	got := ge.Research.progress
	ge.mu.RUnlock()
	if math.Abs(got-cost/4) > 1e-9*cost {
		t.Errorf("progress %v, want a quarter of %v", got, cost)
	}
	ge.mu.Lock()
	ge.Diplomacy.factions["merchant_guild"].Deals = []FactionDeal{{ID: 9, Kind: DealResearch, Give: "stone", GiveAmt: 10, Res: config.ResearchTarget, Amt: 1e18}}
	ge.mu.Unlock()
	noErr(t, ge.takeOffer("merchant_guild", 1))
	ge.mu.RLock()
	done, cur := ge.Research.researched[keys[0]], ge.Research.currentTech
	ge.mu.RUnlock()
	if !done || cur != "" {
		t.Errorf("a gift past the cost left the tech unfinished (done %v, current %q)", done, cur)
	}
}

// TestGifts_PeopleRespectTheCap: people join up to the housing; an offer for
// more than fits is refused.
func TestGifts_PeopleRespectTheCap(t *testing.T) {
	ge := giftEngine(t,
		FactionDeal{ID: 1, Kind: DealPeople, Give: "stone", GiveAmt: 10, Res: "workers", Amt: 3},
		FactionDeal{ID: 2, Kind: DealPeople, Give: "stone", GiveAmt: 10, Res: "workers", Amt: 1e6})
	ge.mu.RLock()
	before := ge.Workers.TotalPop()
	ge.mu.RUnlock()
	noErr(t, ge.takeOffer("merchant_guild", 1))
	ge.mu.RLock()
	after := ge.Workers.TotalPop()
	ge.mu.RUnlock()
	if after != before+3 {
		t.Errorf("population %d -> %d, want +3", before, after)
	}
	if err := ge.takeOffer("merchant_guild", 2); err == nil || !strings.Contains(err.Error(), "housing") {
		t.Errorf("people past the housing: %v", err)
	}
}

// TestGifts_KnowhowOncePerCiv: know-how is offered only at the cap or allied,
// once; it adds its share of the specialty and survives a save.
func TestGifts_KnowhowOncePerCiv(t *testing.T) {
	isolateAccountDir(t)
	def := factionDef(t, "merchant_guild")
	env := testDealEnv("colonial_age")
	low := FactionState{Discovered: true, Status: "friendly", Opinion: 30}
	had := alliedState
	had.Knowhow = true
	found := 0
	for seed := int64(0); seed < 40; seed++ {
		for _, d := range rollFactionDeals(def, low, env, rand.New(rand.NewSource(seed)), 1) {
			if d.Kind == DealKnowhow {
				t.Fatalf("know-how offered at opinion %d", low.Opinion)
			}
		}
		n := 0
		for _, d := range rollFactionDeals(def, alliedState, env, rand.New(rand.NewSource(seed)), 1) {
			if d.Kind == DealKnowhow {
				n++
			}
		}
		if n > 1 {
			t.Fatalf("%d know-hows in one rotation", n)
		}
		found += n
		for _, d := range rollFactionDeals(def, had, env, rand.New(rand.NewSource(seed)), 1) {
			if d.Kind == DealKnowhow {
				t.Fatal("know-how offered again after it was earned")
			}
		}
	}
	if found == 0 {
		t.Fatal("an allied civ never offered know-how")
	}

	ge := giftEngine(t, FactionDeal{ID: 1, Kind: DealKnowhow, Give: "stone", GiveAmt: 10, Res: def.Specialty})
	ge.mu.Lock()
	ge.Diplomacy.factions["merchant_guild"].Opinion = dealStandingCap
	ge.Buildings.counts["wood_camp"] = 20
	base := ge.Resources.resources[def.Specialty].Rate
	ge.mu.Unlock()
	noErr(t, ge.takeOffer("merchant_guild", 1))
	ge.mu.RLock()
	got := ge.Resources.resources[def.Specialty].Rate
	ge.mu.RUnlock()
	if base > 0 && !(got > base) {
		t.Errorf("%s rate %v -> %v after know-how", def.Specialty, base, got)
	}
	b := saveAndLoad(t, ge, "knowhow")
	b.mu.RLock()
	earned := b.Diplomacy.factions["merchant_guild"].Knowhow
	b.mu.RUnlock()
	if !earned {
		t.Error("earned know-how lost through save and load")
	}
	ge.mu.Lock()
	ge.Diplomacy.factions["merchant_guild"].Deals = []FactionDeal{{ID: 5, Kind: DealKnowhow, Give: "stone", GiveAmt: 10, Res: def.Specialty}}
	ge.mu.Unlock()
	if err := ge.takeOffer("merchant_guild", 1); err == nil {
		t.Error("took the same know-how twice")
	}
}

// TestGifts_BoostsSaveAndLoad: running boosts survive a save and load.
func TestGifts_BoostsSaveAndLoad(t *testing.T) {
	isolateAccountDir(t)
	ge := giftEngine(t, FactionDeal{ID: 1, Kind: DealBoost, Give: "stone", GiveAmt: 10, Res: "wood", Amt: 500})
	noErr(t, ge.takeOffer("merchant_guild", 1))
	b := saveAndLoad(t, ge, "boosts")
	b.mu.RLock()
	got, want := b.boosts, ge.boosts
	b.mu.RUnlock()
	if !reflect.DeepEqual(got, want) || len(want) != 1 {
		t.Errorf("boosts after load %v, want %v", got, want)
	}
}

// TestDeals_PlanDealIsGone: a plan holds builds, techs and the rest, not deals.
func TestDeals_PlanDealIsGone(t *testing.T) {
	if _, ok := reflect.TypeOf(PlanItem{}).FieldByName("Deal"); ok {
		t.Error("PlanItem still carries a deal")
	}
}

// TestDeals_AcceptOnce: taking an offer pays once, adds standing and a trade,
// and a second accept is refused.
func TestDeals_AcceptOnce(t *testing.T) {
	ge := giftEngine(t, FactionDeal{ID: 1, Kind: DealBoost, Give: "stone", GiveAmt: 10, Res: "wood", Amt: 100})
	ge.mu.RLock()
	stone := ge.Resources.Get("stone")
	ge.mu.RUnlock()
	noErr(t, ge.takeOffer("merchant_guild", 1))
	ge.mu.RLock()
	fs := ge.Diplomacy.factions["merchant_guild"]
	paid, op, trades := stone-ge.Resources.Get("stone"), fs.Opinion, fs.TradeCount
	ge.mu.RUnlock()
	if paid != 10 || op != dealTradeStanding || trades != 1 {
		t.Errorf("paid %v, opinion %d, trades %d", paid, op, trades)
	}
	if err := ge.takeOffer("merchant_guild", 1); err == nil || !strings.Contains(err.Error(), "already taken") {
		t.Errorf("second accept: %v", err)
	}
	for _, bad := range []int{0, 99} {
		if ge.takeOffer("merchant_guild", bad) == nil {
			t.Errorf("accept %d: no error", bad)
		}
	}
	ge.mu.Lock()
	ge.Diplomacy.factions["merchant_guild"].Deals[0].Taken = false
	ge.Resources.resources["stone"].Amount = 1
	ge.mu.Unlock()
	if err := ge.takeOffer("merchant_guild", 1); err == nil || !strings.Contains(err.Error(), "Not enough") {
		t.Errorf("unaffordable accept: %v", err)
	}
}

// TestDeals_StandingCap: offers raise opinion to dealStandingCap and no
// further, and never lower one already past it.
func TestDeals_StandingCap(t *testing.T) {
	ge := giftEngine(t)
	take := func(id, opinion int) int {
		ge.mu.Lock()
		fs := ge.Diplomacy.factions["merchant_guild"]
		fs.Opinion = opinion
		fs.Deals = []FactionDeal{{ID: id, Kind: DealBoost, Give: "stone", GiveAmt: 10, Res: "wood", Amt: 10}}
		ge.mu.Unlock()
		noErr(t, ge.takeOffer("merchant_guild", 1))
		ge.mu.RLock()
		defer ge.mu.RUnlock()
		return ge.Diplomacy.factions["merchant_guild"].Opinion
	}
	if got := take(100, 10); got != 10+dealTradeStanding {
		t.Errorf("from 10: opinion %d", got)
	}
	if got := take(101, dealStandingCap); got != dealStandingCap {
		t.Errorf("at the cap: opinion %d", got)
	}
	if got := take(102, 70); got != 70 {
		t.Errorf("above the cap: opinion %d, want 70 untouched", got)
	}
}

// TestDeals_Refresh: offers roll on first contact, hold for a rotation of
// live play, re-roll then and on an age advance, and do not move offline.
func TestDeals_Refresh(t *testing.T) {
	ge := dealEngine(t, "colonial_age")
	refresh := dealRefreshFor("colonial_age")
	meet(ge, "merchant_guild", 20)
	state := func() FactionState {
		ge.mu.RLock()
		defer ge.mu.RUnlock()
		fs := *ge.Diplomacy.factions["merchant_guild"]
		fs.Deals = append([]FactionDeal(nil), fs.Deals...)
		return fs
	}
	first := state()
	if first.DealRound != 1 || first.DealsFor != "colonial_age" || len(first.Deals) == 0 {
		t.Fatalf("after first contact: round %d for %q, %d deals", first.DealRound, first.DealsFor, len(first.Deals))
	}
	tick := func(n int) {
		ge.mu.Lock()
		for i := 0; i < n; i++ {
			ge.tickFactionDeals()
		}
		ge.mu.Unlock()
	}
	tick(refresh - 1)
	if s := state(); s.DealRound != 1 || !reflect.DeepEqual(s.Deals, first.Deals) {
		t.Fatalf("offers rotated early: round %d", s.DealRound)
	}
	ge.SimulateOffline(20 * time.Hour)
	if s := state(); s.DealRound != 1 || s.DealTicks != refresh-1 {
		t.Fatalf("offline moved the offers: round %d, ticks %d", s.DealRound, s.DealTicks)
	}
	tick(1)
	if s := state(); s.DealRound != 2 || s.DealTicks != 0 {
		t.Fatalf("after %d ticks: round %d, ticks %d", refresh, s.DealRound, s.DealTicks)
	}
	ge.mu.Lock()
	ge.advanceAge("industrial_age")
	ge.pendingCatastrophe = ""
	ge.mu.Unlock()
	if err := ge.takeOffer("merchant_guild", 1); err == nil {
		t.Error("took a colonial offer in the Industrial Age")
	}
	tick(1)
	if s := state(); s.DealRound != 3 || s.DealsFor != "industrial_age" {
		t.Errorf("after the advance: round %d for %q", s.DealRound, s.DealsFor)
	}
}

// TestDeals_SaveLoadAndDeterminism: offers, taken flags, the round and the
// timer survive a save and load; the same seed rolls the same offers.
func TestDeals_SaveLoadAndDeterminism(t *testing.T) {
	isolateAccountDir(t)
	ge := dealEngine(t, "colonial_age")
	meet(ge, "merchant_guild", 30)
	meet(ge, "riverlands_tribes", 10)
	_ = ge.takeOffer("merchant_guild", 1)
	ge.mu.Lock()
	for i := 0; i < 123; i++ {
		ge.tickFactionDeals()
	}
	want := ge.Diplomacy.GetFactionsForSave()
	ge.mu.Unlock()
	b := saveAndLoad(t, ge, "deals-roundtrip")
	b.mu.RLock()
	got := b.Diplomacy.GetFactionsForSave()
	b.mu.RUnlock()
	if !reflect.DeepEqual(got, want) {
		gj, _ := json.Marshal(got)
		wj, _ := json.Marshal(want)
		t.Fatalf("round trip:\n got %s\nwant %s", gj, wj)
	}
	roll := func(seed int64) string {
		e := dealEngine(t, "colonial_age")
		e.SeedRNG(seed)
		for _, key := range []string{"riverlands_tribes", "ironhold_clans", "merchant_guild"} {
			meet(e, key, 30)
		}
		e.mu.RLock()
		defer e.mu.RUnlock()
		j, _ := json.Marshal(e.Diplomacy.GetFactionsForSave())
		return string(j)
	}
	if roll(3) != roll(3) || roll(3) == roll(4) {
		t.Error("offers are not a function of the seed")
	}
}

// TestDeals_Wording pins how an offer reads.
func TestDeals_Wording(t *testing.T) {
	num := textfmt.Number
	dur := func(int) string { return "20m" }
	for _, c := range []struct {
		kind, res string
		amt       float64
		want      string
	}{
		{DealBoost, "stone", 600, "Boost: give 876M coal → +25% stone 20m"},
		{DealResearch, config.ResearchTarget, 3200, "Research: give 876M coal → research +3.2K"},
		{DealPeople, "workers", 4, "People: give 876M coal → +4 workers"},
		{DealKnowhow, "stone", 0, "Know-how: give 876M coal → know-how: +5% stone"},
	} {
		if got := DealTerms(c.kind, "coal", 876*config.Million, c.res, c.amt, num, dur); got != c.want {
			t.Errorf("DealTerms(%s) = %q, want %q", c.kind, got, c.want)
		}
	}
}

// TestGifts_KnowhowRefusedWhenStandingFell: the opinion gate holds at accept,
// not only when the offer is rolled.
func TestGifts_KnowhowRefusedWhenStandingFell(t *testing.T) {
	def := factionDef(t, "merchant_guild")
	ge := giftEngine(t, FactionDeal{ID: 1, Kind: DealKnowhow, Give: "stone", GiveAmt: 10, Res: def.Specialty})
	ge.mu.Lock()
	ge.Diplomacy.factions["merchant_guild"].Opinion = dealStandingCap - 1
	ge.mu.Unlock()
	stone := ge.Resources.Get("stone")
	if err := ge.takeOffer("merchant_guild", 1); err == nil || !strings.Contains(err.Error(), "trust") {
		t.Fatalf("know-how below the opinion gate: %v", err)
	}
	if ge.Resources.Get("stone") != stone {
		t.Error("a refused know-how was paid for")
	}
	ge.mu.Lock()
	ge.Diplomacy.factions["merchant_guild"].Status = "allied"
	ge.mu.Unlock()
	noErr(t, ge.takeOffer("merchant_guild", 1))
}

// TestGifts_ResearchPaidInProportion: a research gift that lands for less than
// offered costs less, and an unaffordable one is refused with nothing spent.
func TestGifts_ResearchPaidInProportion(t *testing.T) {
	ge := giftEngine(t, FactionDeal{ID: 1, Kind: DealResearch, Give: "stone", GiveAmt: 100, Res: config.ResearchTarget, Amt: 1e18})
	ge.mu.Lock()
	keys := ge.Research.StartableTechs(ge.age, ge.progress.GetAgeOrder())
	ge.mu.Unlock()
	noErr(t, ge.StartResearch(keys[0]))
	ge.mu.Lock()
	cost := ge.Research.defs[keys[0]].Cost
	ge.Research.progress = cost - cost/10 // a tenth left
	ge.Diplomacy.factions["merchant_guild"].Deals[0].Amt = cost / 2
	stone := ge.Resources.Get("stone")
	ge.mu.Unlock()
	noErr(t, ge.takeOffer("merchant_guild", 1))
	ge.mu.RLock()
	paid := stone - ge.Resources.Get("stone")
	ge.mu.RUnlock()
	if math.Abs(paid-20) > 0.5 { // a tenth of the tech of a half-tech lot: 100 * 0.1/0.5
		t.Errorf("paid %v for a fifth of the lot, want about 20", paid)
	}
}

// TestGifts_BoostLastsItsTicksLive: a boost of 10 ticks adds 10 ticks of a
// quarter of the rate, as time away does, no more and no less.
func TestGifts_BoostLastsItsTicksLive(t *testing.T) {
	run := func(take bool) float64 {
		ge := giftEngine(t, FactionDeal{ID: 1, Kind: DealBoost, Give: "stone", GiveAmt: 10, Res: "wood", Amt: 10})
		if take {
			noErr(t, ge.takeOffer("merchant_guild", 1))
		}
		ge.mu.Lock()
		ge.Resources.resources["wood"].Amount = 0
		ge.mu.Unlock()
		for range 14 {
			ge.doTick()
		}
		ge.mu.RLock()
		defer ge.mu.RUnlock()
		return ge.Resources.resources["wood"].Amount
	}
	base := woodRate(giftEngine(t))
	if extra, want := run(true)-run(false), float64(0.25*base*10); math.Abs(extra-want) > 0.01*want {
		t.Errorf("a 10-tick boost added %v wood, want %v", extra, want)
	}
}

// TestGifts_BoostRefusedWhenNothingMade: a boost on a good the town no longer
// makes is refused with nothing spent.
func TestGifts_BoostRefusedWhenNothingMade(t *testing.T) {
	ge := giftEngine(t, FactionDeal{ID: 1, Kind: DealBoost, Give: "stone", GiveAmt: 10, Res: "steel", Amt: 100})
	ge.mu.Lock()
	made := ge.incomeOf("steel")
	stone := ge.Resources.Get("stone")
	ge.mu.Unlock()
	if made != 0 {
		t.Fatalf("test town makes steel (%v a tick); pick another good", made)
	}
	if err := ge.takeOffer("merchant_guild", 1); err == nil || !strings.Contains(err.Error(), "not being made") {
		t.Fatalf("boost on a good not made: %v", err)
	}
	if ge.Resources.Get("stone") != stone {
		t.Error("a refused boost was paid for")
	}
}
