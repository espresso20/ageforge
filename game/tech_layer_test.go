package game

import (
	"math"
	"math/rand"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// TestBonusTruth measures every tech's every promise, in every age, on a
// typical player. The tests here pin the layer's rules one at a time, on
// numbers a reader can check by hand.

// rateOf recalculates and returns what res makes per tick.
func rateOf(ge *GameEngine, res string) float64 {
	ge.recalculateRates()
	return ge.Resources.resources[res].Rate
}

// TestTechLayerCountsAfterTheCap: a tech's output bonus multiplies what is
// left after the x3 clamp, so it counts in full on a pool that is far past
// its cap. Bonuses on the same thing add up inside the layer, a bonus on
// all production adds to the one on the resource, and the breakdown's
// research line is exactly what the layer added.
func TestTechLayerCountsAfterTheCap(t *testing.T) {
	ge := newTruthEngine("industrial_age", truthClean)
	// Milestone rewards worth +500% on all production and on stone: both
	// pools are at their caps, x3 each.
	ge.permanentBonuses["production_all"] = 5
	ge.permanentBonuses["stone_rate"] = 5
	base := rateOf(ge, "stone")
	b := ge.Resources.resources["stone"].Breakdown
	if got := (b.BuildingRate + b.BonusRate) / b.BuildingRate; math.Abs(got-9) > 1e-9 {
		t.Fatalf("setup: stone runs at x%v of what its buildings make, want both pools at their cap (x9)", got)
	}
	if b.ResearchRate != 0 {
		t.Fatalf("setup: research adds %v to stone with no tech researched", b.ResearchRate)
	}
	near := func(got, want float64) bool { return math.Abs(got-want) <= 1e-9*math.Abs(want) }

	// Stoneworking: +10% stone, on top of the capped pools.
	learn(ge, "stoneworking")
	if got := rateOf(ge, "stone"); !near(got, base*1.10) {
		t.Errorf("with +10%% stone the rate is %v, want %v (x1.10 of the capped rate)", got, base*1.10)
	}
	// Bronze Working: another +10%. They add up: +20%, not x1.21.
	learn(ge, "bronze_working")
	if got := rateOf(ge, "stone"); !near(got, base*1.20) {
		t.Errorf("with two +10%% stone techs the rate is %v, want %v (x1.20)", got, base*1.20)
	}
	// Industrialization: +5% all production. It adds to stone's own +20%.
	learn(ge, "industrialization")
	got := rateOf(ge, "stone")
	if !near(got, base*1.25) {
		t.Errorf("with +20%% stone and +5%% all production the rate is %v, want %v (x1.25)", got, base*1.25)
	}
	b = ge.Resources.resources["stone"].Breakdown
	if !near(b.ResearchRate, base*0.25) {
		t.Errorf("the breakdown's research line reads %v, want the layer's share %v", b.ResearchRate, base*0.25)
	}
	if sum := b.BuildingRate + b.WorkerRate + b.ResearchRate + b.EventRate + b.TradeRate + b.FoodDrain + b.BonusRate + b.LegacyRate + b.MasteryRate; !near(sum, got) {
		t.Errorf("the breakdown adds up to %v, the rate is %v", sum, got)
	}
	// A resource no tech names gets the all-production bonus alone.
	ge.Research.researched = map[string]bool{}
	ge.Research.rebuildBonuses()
	woodBefore := rateOf(ge, "wood")
	learn(ge, "industrialization")
	if got := rateOf(ge, "wood"); !near(got, woodBefore*1.05) {
		t.Errorf("with +5%% all production wood makes %v, want %v", got, woodBefore*1.05)
	}
	// And the pools themselves never moved: no tech joined one.
	r := ge.buildResolver()
	if got := r.AddTotal("production_all"); got != 5 {
		t.Errorf("the all-production pool holds %v, want the 5 the milestones gave it", got)
	}
	if got := r.AddTotal("stone_rate"); got != 5 {
		t.Errorf("stone's pool holds %v, want the 5 the milestones gave it", got)
	}
}

// TestTechLayerLeavesFlatAmountsAlone: a first source's flat output and an
// event's are promised as amounts. The layer multiplies what buildings and
// crews make, not those.
func TestTechLayerLeavesFlatAmountsAlone(t *testing.T) {
	ge := NewGameEngine()
	ge.Resources.UnlockResource("steel")
	if got := rateOf(ge, "steel"); got != 0 {
		t.Fatalf("setup: a new game makes %v steel", got)
	}
	// Steel Forging: first steel, 0.25 a tick. Steam Power: +6% steel.
	// Industrialization: +5% all production.
	learn(ge, "steel_forging", "steam_power", "industrialization")
	if got := rateOf(ge, "steel"); got != 0.25 {
		t.Errorf("Steel Forging's first steel is %v a tick with +11%% on steel researched, want 0.25 exactly", got)
	}
}

// TestTechStorageAndHousing: the techs' storage bonus is a share of every
// store and their housing bonus a share of housing, rounded up to a whole
// person. Both add up across techs.
func TestTechStorageAndHousing(t *testing.T) {
	ge := newTruthEngine("bronze_age", truthClean)
	ge.recalculateRates()
	before := map[string]float64{}
	for _, key := range ge.Resources.order {
		before[key] = ge.Resources.resources[key].Storage
	}
	learn(ge, "pottery", "masonry") // +10% storage each
	ge.recalculateRates()
	for _, key := range ge.Resources.order {
		if got, want := ge.Resources.resources[key].Storage, before[key]*1.20; math.Abs(got-want) > 1e-9*want {
			t.Errorf("%s storage is %v with two +10%% techs, want %v (x1.20)", key, got, want)
		}
	}

	for _, c := range []struct {
		base  int
		bonus float64
		want  int
	}{
		{0, 0.05, 0}, {10, 0, 10}, {10, 0.05, 11}, {19, 0.05, 20}, {20, 0.05, 21}, {100, 0.05, 105}, {100, 0.13, 113}, {1500, 0.18, 1770},
	} {
		if got := TechHousing(c.base, c.bonus); got != c.want {
			t.Errorf("TechHousing(%d, %v) = %d, want %d", c.base, c.bonus, got, c.want)
		}
	}
	plain := ge.popCapLocked()
	learn(ge, "fire_mastery") // +5% housing
	if got, want := ge.popCapLocked(), TechHousing(plain, 0.05); got != want || got <= plain {
		t.Errorf("housing is %d with Fire Mastery, want %d (from %d)", got, want, plain)
	}
}

// TestTechCutsMultiplyAndHoldAtTheirFloor: the cuts of a price or a time
// multiply (two 3% cuts leave 0.97 x 0.97), they multiply what the
// milestones' pool leaves, and however many stack, a floor holds.
func TestTechCutsMultiplyAndHoldAtTheirFloor(t *testing.T) {
	ge := NewGameEngine()
	cost := func() float64 {
		ge.recalculateRates()
		return truthBuildCost(ge)
	}
	near := func(got, want float64) bool { return math.Abs(got-want) <= 1e-9 }
	if got := cost(); got != 1 {
		t.Fatalf("setup: a new game pays %v of the listed price", got)
	}
	learn(ge, "civil_engineering", "nanofabrication") // 3% each
	if got, want := cost(), 0.97*0.97; !near(got, want) {
		t.Errorf("two 3%% cuts leave %v of the price, want %v", got, want)
	}
	ge.permanentBonuses["build_cost"] = -0.19 // the milestones' pool
	if got, want := cost(), 0.81*0.97*0.97; !near(got, want) {
		t.Errorf("with the milestones' -19%% the price is %v of its listing, want %v", got, want)
	}

	// Construction time: Mass Production's 8%, then Self-Replication's 5%.
	hall := config.BuildingDef{BuildTicks: 100_000}
	learn(ge, "mass_production")
	if got := ge.buildTicksLocked(hall); got != 92_000 {
		t.Errorf("a 100,000 tick build takes %d with Mass Production, want 92,000", got)
	}
	learn(ge, "self_replication")
	if got := ge.buildTicksLocked(hall); got != 87_400 {
		t.Errorf("a 100,000 tick build takes %d with both, want 87,400", got)
	}
	// Research time: Printing Press's 3% multiplies what the speed pool left.
	learn(ge, "printing_press")
	if got := ResearchTicks(100_000, 0.20, ge.Research.TimeFactor(), 0, 1); got != 77_600 {
		t.Errorf("a 100,000 tick tech with +20%% research speed and a 3%% cut takes %d, want 77,600", got)
	}

	// The floors: a made-up tree whose every tech halves all three.
	src := rules.FromConfig()
	for _, key := range []string{"zz_half_a", "zz_half_b", "zz_half_c", "zz_half_d"} {
		src.Techs = append(src.Techs, config.TechDef{Key: key, Name: key, Age: "primitive_age", Lane: config.LaneCraft, Cost: 1, ResearchTicks: 1,
			Effects: []config.TechEffect{
				{Kind: config.EffectBuildCost, Value: -0.5},
				{Kind: config.EffectBuildTime, Value: -0.5},
				{Kind: config.EffectResearchTime, Value: -0.5},
			}})
	}
	deep := NewGameEngineWith(rules.Compile(src))
	learn(deep, "zz_half_a", "zz_half_b", "zz_half_c", "zz_half_d")
	deep.recalculateRates()
	if got := truthBuildCost(deep); !near(got, config.BuildCostFloor) {
		t.Errorf("four halvings leave %v of the price, want the floor %v", got, config.BuildCostFloor)
	}
	if got, want := deep.buildTicksLocked(hall), int(100_000*config.BuildTimeFloor); got != want {
		t.Errorf("four halvings leave a 100,000 tick build at %d, want the floor %d", got, want)
	}
	if got := deep.Research.TimeFactor(); got != config.ResearchTimeFloor {
		t.Errorf("four halvings leave research at %v of its time, want the floor %v", got, config.ResearchTimeFloor)
	}
	if raw := deep.Research.raw[config.TechEffectKey{Kind: config.EffectBuildCost}]; raw != 0.0625 {
		t.Errorf("the raw build cost term is %v, want 0.0625 (what the floor held back)", raw)
	}
}

// TestTechGatherAndRaids runs the two mechanic numbers the truth guard reads
// only as a term: a real gather, a real war raid and a real raid event.
func TestTechGatherAndRaids(t *testing.T) {
	ge := NewGameEngine()
	setResource(ge, "food", 0)
	gather := func() float64 {
		before := ge.Resources.Get("food")
		if _, err := ge.GatherResource("food", 3); err != nil {
			t.Fatal(err)
		}
		return ge.Resources.Get("food") - before
	}
	if got := gather(); got != 3 {
		t.Fatalf("a gather of 3 brings %v", got)
	}
	learn(ge, "tool_making")
	if got := gather(); got != 5 {
		t.Errorf("a gather of 3 brings %v with Tool Making, want 5", got)
	}

	// A war raid of 100 gold, no garrison: all of it, then 10% less with
	// Imperial Legions, then 19% less with Nuclear Deterrence too.
	raid := func(ge *GameEngine) float64 {
		setResource(ge, "gold", 1000)
		ge.mu.Lock()
		ge.Diplomacy.pendingRaids = append(ge.Diplomacy.pendingRaids, RaidRequest{FactionKey: "riverlands_tribes", Resource: "gold", Amount: 100})
		ge.applyWarRaids()
		ge.mu.Unlock()
		return 1000 - ge.Resources.Get("gold")
	}
	war := NewGameEngine()
	if got := raid(war); got != 100 {
		t.Fatalf("a raid of 100 takes %v with no tech", got)
	}
	learn(war, "imperial_legions")
	if got := raid(war); math.Abs(got-90) > 1e-9 {
		t.Errorf("a raid of 100 takes %v with Imperial Legions, want 90", got)
	}
	learn(war, "nuclear_deterrence")
	if got := raid(war); math.Abs(got-81) > 1e-9 {
		t.Errorf("a raid of 100 takes %v with both raid techs, want 81", got)
	}
	// A raid too big to land still takes nothing: the cut never turns a
	// miss into a hit.
	setResource(war, "gold", 95)
	war.mu.Lock()
	war.Diplomacy.pendingRaids = append(war.Diplomacy.pendingRaids, RaidRequest{FactionKey: "riverlands_tribes", Resource: "gold", Amount: 100})
	war.applyWarRaids()
	war.mu.Unlock()
	if got := war.Resources.Get("gold"); got != 95 {
		t.Errorf("a raid of 100 on a stock of 95 left %v, want it to miss as before", got)
	}

	// A raid event takes its cut too; a loss that is no raid does not.
	ev := NewGameEngine()
	learn(ev, "imperial_legions")
	setResource(ev, "gold", 1000)
	ev.mu.Lock()
	ev.applyEventEffects(config.EventDef{Key: "zz_raid", Name: "Raid", Raid: true, Effects: []config.Effect{{Type: "steal_resource", Target: "gold", Value: 100}}})
	ev.mu.Unlock()
	if got := 1000 - ev.Resources.Get("gold"); math.Abs(got-90) > 1e-9 {
		t.Errorf("a raid event of 100 takes %v with Imperial Legions, want 90", got)
	}
	setResource(ev, "gold", 1000)
	ev.mu.Lock()
	ev.applyEventEffects(config.EventDef{Key: "zz_tax", Name: "Tax", Effects: []config.Effect{{Type: "steal_resource", Target: "gold", Value: 100}}})
	ev.mu.Unlock()
	if got := 1000 - ev.Resources.Get("gold"); got != 100 {
		t.Errorf("a loss of 100 that is no raid takes %v, want 100", got)
	}
}

// TestTechMechanicsReachTheirCommands: each mechanic number a tech moves is
// read where the game uses it: a market trade, a route's time, a scouting
// expedition's, a gift's price, a festival's price and wait, a deal set's
// life.
func TestTechMechanicsReachTheirCommands(t *testing.T) {
	ge := newTruthEngine("modern_age", truthClean)
	read := func() map[string]float64 {
		ge.recalculateRates()
		return truthMechanics(ge)
	}
	before := read()
	// Currency, Banking: the fee falls 3 points each. Road Building,
	// Railroads: routes 15% faster each. Cartography: expeditions 10%
	// faster. Telecommunications: deals 30% sooner, gifts 25% cheaper.
	// Radio: festivals 20% sooner. Social Media (the next age's): 20%
	// cheaper.
	learn(ge, "currency", "banking", "road_building", "railroads", "cartography", "telecommunications", "radio", "social_media")
	after := read()
	near := func(got, want, tol float64) bool { return math.Abs(got-want) <= tol }
	if got, want := after[config.MechanicMarketFee], config.ExchangeFee-0.06; !near(got, want, 1e-9) || !near(before[config.MechanicMarketFee], config.ExchangeFee, 1e-9) {
		t.Errorf("the market's fee is %v (%v before), want %v (%v before)", got, before[config.MechanicMarketFee], want, config.ExchangeFee)
	}
	for key, want := range map[string]float64{
		config.MechanicRouteTicks:            0.85 * 0.85,
		config.MechanicDealRefreshTicks:      0.70,
		config.MechanicGiftCost:              0.75,
		config.MechanicFestivalCooldownTicks: 0.80,
		config.MechanicFestivalCost:          0.80,
		config.MechanicExpeditionTicks:       0.90,
	} {
		// Whole ticks round down: good to a tick.
		if got := after[key] / before[key]; !near(got, want, 2/before[key]+1e-9) {
			t.Errorf("%s is %v of what it was (%v from %v), want %v", key, got, after[key], before[key], want)
		}
	}
	if got := ge.Diplomacy.GiftPrice(); got != 150 {
		t.Errorf("a gift costs %v with Telecommunications, want 150", got)
	}

	// The market pays what the lower fee leaves, on a real trade.
	plain := newTruthEngine("modern_age", truthClean)
	priced := plain.rules.PricedResources("modern_age")
	from, to := priced[0], priced[1]
	trade := func(e *GameEngine) float64 {
		e.recalculateRates()
		setResource(e, from, 1e9)
		setResource(e, to, 0)
		e.Resources.resources[to].Storage = 1e18
		got, err := e.ExchangeResources(from, to, 1000)
		if err != nil {
			t.Fatalf("trading %s for %s: %v", from, to, err)
		}
		return got
	}
	full, cut := trade(plain), trade(ge)
	if want := full * (1 - config.ExchangeFee + 0.06) / (1 - config.ExchangeFee); !near(cut, want, 1e-6*want) {
		t.Errorf("1,000 %s buys %v %s at a 14 point fee and %v at 20, want %v", from, cut, to, full, want)
	}

	// A route started now takes the shorter run.
	def := ge.Trade.routeDefs["local_barter"]
	ge.Buildings.counts[def.RequiredBld] = def.MinCount
	if err := ge.StartTradeRoute("local_barter"); err != nil {
		t.Fatal(err)
	}
	if got, want := ge.Trade.activeRoutes["local_barter"].TicksLeft, int(float64(def.TicksPerRun)*0.85*0.85); got != want {
		t.Errorf("Local Barter runs every %d ticks with both route techs, want %d (listed %d)", got, want, def.TicksPerRun)
	}

	// A campaign keeps its time: the cut is for scouting expeditions.
	march := func(e *GameEngine) int {
		delete(e.Military.activeByCat, ExpeditionMilitary)
		if err := e.Military.LaunchExpedition(rand.New(rand.NewSource(3)), "raid_bandits", e.age, ageOrders()); err != nil {
			t.Fatal(err)
		}
		return e.Military.activeByCat[ExpeditionMilitary].TicksLeft
	}
	if a, b := march(plain), march(ge); a != b {
		t.Errorf("a campaign takes %d ticks with Cartography and %d without, want the same", b, a)
	}
}
