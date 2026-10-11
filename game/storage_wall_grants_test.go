package game

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/boon"
	"github.com/espresso20/ageforge/config"
)

// Paying into a full store. Production is not the only way stock arrives: a
// market sale, a trade route's run, a deal, loot, a boon, an event, a refund
// all put stock into a store that may be full. A store is a wall (the storage
// wall), so in every one of them the surplus is lost, nothing goes negative
// or past the cap, nothing crashes, and the player is told with the line the
// path already had ("Storage was nearly full: only ... fit." or its own).
// A sale or a deal never charges for what the store could not take.

const nearlyFull = "Storage was nearly full: only "

// wallHeld fails when any store is negative, not a number, or above its cap
// (stock a save kept above a new cap is graced and may be).
func wallHeld(t *testing.T, ge *GameEngine) {
	t.Helper()
	for key, r := range ge.Resources.resources {
		if math.IsNaN(r.Amount) || r.Amount < 0 {
			t.Errorf("%s holds %v", key, r.Amount)
		}
		if r.Amount > r.Storage+1e-9 && !ge.Resources.Graced(key) {
			t.Errorf("%s holds %v, over its cap of %v", key, r.Amount, r.Storage)
		}
	}
}

// fillStore sets res's cap and stock: the store holds have of cap.
func fillStore(ge *GameEngine, res string, have, cap float64) {
	ge.Resources.resources[res].Storage = cap
	ge.Resources.resources[res].Amount = have
}

// A market sale into a store with room for less than the sale pays takes a
// smaller sale and charges for only that; into a full store it is refused.
// The plan's trade item meets the same wall.
func TestWallGrants_MarketSaleIsNotChargedForWhatCannotBeHeld(t *testing.T) {
	ge := bronzeEngine(t)
	setAmount(ge, "wood", 50_000)
	fillStore(ge, "stone", 90, 100)
	rate := ge.Trade.RateIn("wood", "stone", ge.age)
	if rate <= 0 {
		t.Fatal("setup: the market does not trade wood for stone")
	}

	got, err := ge.ExchangeResources("wood", "stone", 10_000)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-10) > 1e-9 || ge.Resources.Get("stone") != 100 {
		t.Errorf("sold into 10 stone of room: received %v, stone = %v, want 10 and 100", got, ge.Resources.Get("stone"))
	}
	if gave := 50_000 - ge.Resources.Get("wood"); math.Abs(gave-10/rate) > 1e-6 {
		t.Errorf("charged %v wood for 10 stone, want %v: only what the store could hold", gave, 10/rate)
	}
	if !logHas(ge, nearlyFull+"10 stone") {
		t.Error("no line saying the store took less than the sale asked for")
	}
	wallHeld(t, ge)

	// Full: refused, nothing charged, the market's pressure untouched.
	wood, pressure := ge.Resources.Get("wood"), ge.Trade.Pressure("wood", "stone")
	if _, err := ge.ExchangeResources("wood", "stone", 100); err == nil || !strings.Contains(err.Error(), "storage is full") {
		t.Errorf("sale into a full store = %v, want a refusal", err)
	}
	if ge.Resources.Get("wood") != wood || ge.Trade.Pressure("wood", "stone") != pressure {
		t.Error("a refused sale still cost wood or moved the market")
	}

	// The plan: a trade item sells only what fits and stops charging.
	pe := bronzeEngine(t)
	setAmount(pe, "wood", 50_000)
	fillStore(pe, "stone", 97, 100)
	if err := pe.PlanAddTrade("wood", "stone", 0); err != nil {
		t.Fatal(err)
	}
	pe.runPlanTick()
	if pe.Resources.Get("stone") > 100 {
		t.Errorf("the plan's trade filled stone to %v", pe.Resources.Get("stone"))
	}
	wood = pe.Resources.Get("wood")
	pe.Trade.DecayPressure(1000)
	pe.runPlanTick()
	if pe.Resources.Get("wood") != wood {
		t.Error("the plan kept selling wood into a full stone store")
	}
	wallHeld(t, pe)
}

// A deal whose goods do not fit is refused before it charges anything, by
// hand and from the plan.
func TestWallGrants_DealIntoAFullStoreIsRefusedNotCharged(t *testing.T) {
	ge := dealEngine(t, "colonial_age")
	meet(ge, "merchant_guild", 30)
	ge.mu.Lock()
	idx := -1
	for i, d := range ge.Diplomacy.factions["merchant_guild"].Deals {
		if d.Get != "" && !d.Taken {
			idx = i
			break
		}
	}
	if idx < 0 {
		ge.mu.Unlock()
		t.Fatal("setup: the guild offers no deal that pays goods")
	}
	d := ge.Diplomacy.factions["merchant_guild"].Deals[idx]
	ge.Resources.resources[d.Give].Amount = d.GiveAmt * 3
	fillStore(ge, d.Get, ge.Resources.GetStorage(d.Get), ge.Resources.GetStorage(d.Get)) // full
	give := ge.Resources.Get(d.Give)
	ge.mu.Unlock()

	if _, err := ge.AcceptFactionDeal("merchant_guild", idx+1); err == nil || !strings.Contains(err.Error(), "storage") {
		t.Errorf("deal into a full store = %v, want a refusal about storage", err)
	}
	if err := ge.PlanAddDeal("merchant_guild", idx+1); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	ge.runPlanTick()
	taken := ge.Diplomacy.factions["merchant_guild"].Deals[idx].Taken
	left := len(ge.plan)
	paid := ge.Resources.Get(d.Give)
	ge.mu.Unlock()
	if taken || left != 1 {
		t.Errorf("the plan took a deal into a full store (taken %v, %d items left)", taken, left)
	}
	if paid != give {
		t.Errorf("%s %v → %v: a refused deal charged the price", d.Give, give, paid)
	}
	wallHeld(t, ge)
}

// A trade route's run into a full store loses the surplus and says what
// fit. (It still takes its export: see the report, that is a decision.)
func TestWallGrants_RouteRunSaysWhatFit(t *testing.T) {
	ge := newSeededEngine(9)
	for _, a := range []string{"stone_age", "bronze_age"} {
		if err := ge.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
	}
	ge.mu.Lock()
	defer ge.mu.Unlock()
	def := ge.Trade.routeDefs["local_barter"]
	ge.Buildings.counts[def.RequiredBld] = def.MinCount
	ge.recalculateRates()
	var imp string
	for imp = range def.Import {
	}
	for res := range def.Export {
		fillStore(ge, res, 1e6, 1e9)
	}
	ge.Resources.UnlockResource(imp)
	fillStore(ge, imp, 100, 101) // room for 1 of the run's several
	ge.Trade.activeRoutes["local_barter"] = &ActiveRoute{Key: "local_barter", TicksLeft: 1}
	msgs := ge.Trade.Tick(ge.Resources, ge.Buildings, nil, 0)
	if ge.Resources.Get(imp) != 101 {
		t.Errorf("%s = %v after a run into 1 of room, want the cap", imp, ge.Resources.Get(imp))
	}
	if !strings.Contains(strings.Join(msgs, "\n"), nearlyFull+"1 "+ResourceName(imp)) {
		t.Errorf("the run said %q, want the line saying what fit", msgs)
	}
	if got := ge.Trade.totalImported[imp]; math.Abs(got-1) > 1e-9 {
		t.Errorf("the route's tally counts %v %s imported, want the 1 that fit", got, imp)
	}
	wallHeld(t, ge)
}

// Expedition and campaign loot: the message states the full loot, and a line
// under it says what fit.
func TestWallGrants_ExpeditionLootSaysWhatFit(t *testing.T) {
	ge := newSeededEngine(1)
	fillStore(ge, "food", 40, 50)
	fillStore(ge, "wood", 50, 50)
	ge.mu.Lock()
	ge.applyExpeditionRewards(map[string]float64{"food": 500, "wood": 20, "stone": 0})
	ge.mu.Unlock()
	if ge.Resources.Get("food") != 50 || ge.Resources.Get("wood") != 50 {
		t.Errorf("food %v, wood %v after loot, want both at their cap of 50", ge.Resources.Get("food"), ge.Resources.Get("wood"))
	}
	if !logHas(ge, nearlyFull+"10 food (of 500)") || !logHas(ge, "0 wood (of 20)") {
		t.Error("no line saying what the full stores took")
	}
	lines := len(ge.log)
	ge.mu.Lock()
	ge.applyExpeditionRewards(map[string]float64{"stone": 0})
	ge.mu.Unlock()
	if len(ge.log) != lines {
		t.Error("loot that fit logged a line")
	}
	wallHeld(t, ge)
}

// A faction boon's lump says what fit, in its own line.
func TestWallGrants_BoonLumpSaysWhatFit(t *testing.T) {
	ge := newSeededEngine(1)
	fillStore(ge, "food", 45, 50)
	ge.mu.Lock()
	line := ge.applyRolledFactionBoon(config.FactionDef{Name: "Test Folk", Key: "test_folk"},
		boon.Boon{Kind: boon.InstantResource, Polarity: boon.Positive, Resource: "food", InstantAmount: 500})
	ge.mu.Unlock()
	if !strings.Contains(line, "storage was nearly full: only 5 food fit") {
		t.Errorf("boon line = %q, want it to say what fit", line)
	}
	if ge.Resources.Get("food") != 50 {
		t.Errorf("food = %v after the boon, want the cap", ge.Resources.Get("food"))
	}
	wallHeld(t, ge)
}

// Event gains, milestone rewards and the Ancient Cache: each states its full
// grant and a line under it says what a full store took.
func TestWallGrants_EventsMilestonesAndTheCacheSayWhatFit(t *testing.T) {
	fullStores := func() *GameEngine {
		ge := newTruthEngine("classical_age", truthClean)
		for _, key := range ge.Resources.order {
			if r := ge.Resources.resources[key]; ge.Resources.IsUnlocked(key) {
				r.Amount = r.Storage
			}
		}
		return ge
	}

	ge := fullStores()
	ge.fireEvent(config.EventByKey()["bountiful_harvest"])
	if !logHas(ge, nearlyFull+"0 food (of ") {
		t.Error("an event gain into a full store did not say what fit")
	}
	wallHeld(t, ge)

	ge = fullStores()
	ge.applyMilestoneRewards([]config.Effect{{Type: "instant_resource", Target: "food", Value: 250}})
	if !logHas(ge, nearlyFull+"0 food (of 250) fit.") {
		t.Error("a milestone reward into a full store did not say what fit")
	}
	wallHeld(t, ge)

	ge = fullStores()
	ge.applyGoodEpochEvent(config.EpochEventByKey()["ancient_cache"])
	if !logHas(ge, nearlyFull) {
		t.Error("the Ancient Cache into full stores did not say what fit")
	}
	wallHeld(t, ge)
}

// Hand-gathering and the black market state what fit in their own lines.
func TestWallGrants_GatheringAndSmugglingSayWhatFit(t *testing.T) {
	ge := newSeededEngine(1)
	fillStore(ge, "food", 50, 50)
	got, err := ge.GatherResource("food", 10)
	if err != nil {
		t.Fatal(err)
	}
	if got != 50 || !logHas(ge, "Food storage is full. Gathered nothing.") {
		t.Errorf("gathering into a full store: stock %v, log %v", got, ge.log)
	}
	fillStore(ge, "wood", 48, 50)
	if _, err := ge.GatherResource("wood", 10); err != nil {
		t.Fatal(err)
	}
	if ge.Resources.Get("wood") != 50 || !logHas(ge, "wood storage is now full") {
		t.Errorf("gathering into 2 of room: wood = %v", ge.Resources.Get("wood"))
	}
	wallHeld(t, ge)

	bm := bmEngine(50000)
	bm.blackMarketRand = blackMarketSeedWin()
	fillStore(bm, "gold", 100, 100)
	if won, _, err := bm.DoBlackMarket("gold"); err != nil || !won {
		t.Fatalf("setup: won %v, err %v", won, err)
	}
	if bm.Resources.Get("gold") != 100 || !logHas(bm, nearlyFull+"0 gold (of ") {
		t.Errorf("a winning run into a full store: gold = %v", bm.Resources.Get("gold"))
	}
	wallHeld(t, bm)
}

// Selling a building refunds half its price: into a full store the part that
// does not fit is lost and the line says so. An upgrade pays nothing out (it
// charges the difference), so it has no surplus to lose.
func TestWallGrants_SaleRefundIsLostAndSaid(t *testing.T) {
	ge := newSeededEngine(1)
	ge.mu.Lock()
	ge.advanceAge("stone_age") // buildings are sold from the Stone Age on
	ge.mu.Unlock()
	ge.Buildings.counts["hut"] = 3
	refund, _ := ge.Buildings.SellCost("hut", 3)
	if len(refund) == 0 {
		t.Fatal("setup: selling huts refunds nothing")
	}
	for res := range refund {
		fillStore(ge, res, ge.Resources.GetStorage(res), ge.Resources.GetStorage(res))
	}
	if err := ge.SellBuilding("hut", 3); err != nil {
		t.Fatal(err)
	}
	if !logHas(ge, "Storage was full, so part of the refund was lost.") {
		t.Error("a refund into full stores did not say part was lost")
	}
	wallHeld(t, ge)

	up := wallUpgradeEngine(t)
	cost, _ := up.Buildings.UpgradeCost("wood_camp", "woodcutter_camp", 1)
	for res, c := range cost {
		if c < 0 {
			t.Errorf("an upgrade pays out %v %s", c, res)
		}
	}
}

// A plan bank an old save carried goes back to the stores up to each cap; what
// does not fit is lost, not carried.
func TestWallGrants_ReturnedPlanBankFillsToTheCapAndNoFurther(t *testing.T) {
	ge := newSeededEngine(1)
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	fillStore(ge, "wood", 45, 50)
	ge.plan[0].Banked = map[string]float64{"wood": 500}
	ge.mu.Lock()
	ge.returnPlanBanks("")
	ge.mu.Unlock()
	if ge.Resources.Get("wood") != 50 || ge.plan[0].Banked != nil {
		t.Errorf("wood = %v, bank = %v: want the cap and an empty bank", ge.Resources.Get("wood"), ge.plan[0].Banked)
	}
	if !logHas(ge, "went back to the stores") {
		t.Error("no line for the bank going back")
	}

	ge.plan[0].Banked = map[string]float64{"wood": 500}
	ge.mu.Lock()
	ge.returnPlanBanks("")
	ge.mu.Unlock()
	if !logHas(ge, "but they were full") || ge.Resources.Get("wood") != 50 {
		t.Errorf("a bank returned to a full store: wood = %v", ge.Resources.Get("wood"))
	}
	wallHeld(t, ge)
}
