package game

import (
	"strings"
	"testing"
	"time"
)

// bronzeEngine is a Bronze Age game with a market, big stores and no
// catastrophe pending, for trade and advance items.
func bronzeEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	ge.advanceAge("stone_age")
	ge.pendingCatastrophe = ""
	ge.advanceAge("bronze_age")
	ge.pendingCatastrophe = ""
	ge.Buildings.counts["market"] = 1
	ge.recalculateRates()
	for _, r := range []string{"wood", "stone", "gold", "food", "iron", "knowledge"} {
		ge.Resources.resources[r].Storage = 1e6
	}
	return ge
}

func TestPlanTrade_SellsOnceTheMarketRecoversAndStopsAtItsAmount(t *testing.T) {
	ge := bronzeEngine(t)
	setAmount(ge, "wood", 50_000)
	setAmount(ge, "stone", 0)
	if err := ge.PlanAddTrade("wood", "stone", 1000); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddTrade("wood", "stone", 5); err == nil {
		t.Error("planned the same pair twice")
	}
	ge.runPlanTick()
	got := ge.Resources.Get("stone")
	if got < 1000-1e-6 || got > 1000+1e-6 {
		t.Fatalf("stone after the trade = %v, want exactly the 1000 asked for", got)
	}
	if len(ge.plan) != 0 {
		t.Errorf("a finished trade stayed in the plan: %+v", ge.plan)
	}
	if !logHas(ge, "traded") {
		t.Error("no summary line for the trade")
	}

	// An open-ended trade waits for the market to recover between sales.
	if err := ge.PlanAddTrade("wood", "stone", 0); err != nil {
		t.Fatal(err)
	}
	ge.Trade.supplyPressure["wood:stone"] = 0.5
	before := ge.Resources.Get("wood")
	ge.runPlanTick()
	if ge.Resources.Get("wood") != before {
		t.Error("sold into a market that hasn't recovered")
	}
	if v := ge.planViews(); len(v) != 1 || !strings.Contains(v[0].Note, "recovering") {
		t.Errorf("view = %+v", v)
	}
	ge.Trade.DecayPressure(1000)
	ge.runPlanTick()
	if ge.Resources.Get("wood") >= before {
		t.Error("didn't sell once the market recovered")
	}
}

// A trade item holds back what it will sell from the items below it.
func TestPlanTrade_ReservesForItself(t *testing.T) {
	ge := bronzeEngine(t)
	setAmount(ge, "wood", 20_000)
	ge.Trade.supplyPressure["wood:stone"] = 0.5 // not due yet
	if err := ge.PlanAddTrade("wood", "stone", 0); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("market", 1); err != nil {
		t.Fatal(err)
	}
	ge.runPlanTick()
	if queued(ge, "market") != 0 {
		t.Error("a build below the trade spent the wood the trade holds")
	}
}

func TestPlanTrade_InvalidPairsAndMissingMarket(t *testing.T) {
	ge := newSeededEngine(1)
	if err := ge.PlanAddTrade("wood", "wood", 0); err == nil {
		t.Error("planned a trade of a resource for itself")
	}
	if err := ge.PlanAddTrade("wood", "iron", 0); err == nil {
		t.Error("planned a trade for a locked resource")
	}
	be := bronzeEngine(t)
	be.Buildings.counts["market"] = 0
	if err := be.PlanAddTrade("wood", "stone", 0); err != nil {
		t.Fatal(err)
	}
	if v := be.planViews(); len(v) != 1 || !strings.Contains(v[0].Note, "Market") {
		t.Errorf("view without a market = %+v", v)
	}
}

func TestPlanAdvance_AdvancesWhenReadyAndDropsTheOldAge(t *testing.T) {
	ge := newSeededEngine(1)
	if err := ge.PlanAddAdvance(); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddAdvance(); err == nil {
		t.Error("planned two advances")
	}
	if _, err := ge.PlanAddBuild("hut", 50); err != nil {
		t.Fatal(err)
	}
	// Next age's building, planned ahead: waits for the advance.
	if _, err := ge.PlanAddBuild("longhouse", 1); err != nil {
		t.Fatalf("planning the next age's building: %v", err)
	}
	ge.runPlanTick()
	if ge.age != "primitive_age" {
		t.Fatal("advanced before the requirements were met")
	}
	// Meet the Stone Age gate by hand.
	reqs, blds := ge.progress.GetRequirementsForNext(ge.age)
	for r, v := range reqs {
		ge.Resources.resources[r].Storage = v * 2
		setAmount(ge, r, v)
	}
	for k, n := range blds {
		ge.Buildings.counts[k] = n
	}
	w := ge.progress.WonderForAge(ge.age)
	ge.Buildings.counts[w] = 1
	ge.runPlanTick()
	if ge.age != "stone_age" {
		t.Fatalf("age = %s, want stone_age", ge.age)
	}
	if !logHas(ge, "Plan: advancing to the Stone Age") {
		t.Error("no advance log line")
	}
	for _, it := range ge.plan {
		if it.Kind == PlanAdvance {
			t.Error("the advance item stayed in the plan")
		}
	}
	ge.runPlanTick()
	keys := ""
	for _, it := range ge.plan {
		keys += it.Key + ","
	}
	if strings.Contains(keys, "hut") || !strings.Contains(keys, "longhouse") && queued(ge, "longhouse") == 0 {
		t.Errorf("after the advance: plan %s, longhouses queued %d; want the huts gone and the longhouse kept", keys, queued(ge, "longhouse"))
	}
}

func TestPlan_StaffsWhatItBuilds(t *testing.T) {
	ge := planTestEngine(t)
	ge.Buildings.counts["hut"] = 5
	ge.recalculateRates()
	if err := ge.RecruitWorker("worker", 3); err != nil {
		t.Fatal(err)
	}
	setAmount(ge, "wood", 1000)
	if _, err := ge.PlanAddBuild("gathering_camp", 1); err != nil {
		t.Fatal(err)
	}
	ge.runPlanTick()
	for i := 0; i < 100 && queued(ge, "gathering_camp") > 0; i++ {
		ge.processBuildQueue()
	}
	if got := ge.Workers.GetAssignedCount("worker", "gathering_camp"); got == 0 {
		t.Error("the plan's gathering camp was left unstaffed with idle workers around")
	}
}

// A `plan advance` that fires while the player is away is followed by the
// next age's producers, but the old age's copies took every idle worker. The
// plan staffs the new ones with workers from the buildings the advance
// superseded, leaves the food producers alone and recruits nobody.
func TestOffline_PlanStaffsTheNewAgeAfterAnAdvance(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["stash"] = 50
	ge.Buildings.counts["hut"] = 10
	ge.Buildings.counts["wood_camp"] = 10
	ge.Buildings.counts["gathering_camp"] = 10
	ge.recalculateRates()
	if err := ge.RecruitWorker("worker", 60); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"wood_camp", "gathering_camp"} {
		if err := ge.AssignWorker(k, 30); err != nil {
			t.Fatal(err)
		}
	}
	if err := ge.PlanAddAdvance(); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("stone_camp", 3); err != nil {
		t.Fatal(err)
	}
	// Meet the Stone Age gate by hand.
	reqs, blds := ge.progress.GetRequirementsForNext(ge.age)
	for r, v := range reqs {
		ge.Resources.resources[r].Storage = v * 2
		setAmount(ge, r, v)
	}
	for k, n := range blds {
		ge.Buildings.counts[k] = max(ge.Buildings.counts[k], n)
	}
	ge.Buildings.counts[ge.progress.WonderForAge(ge.age)] = 1
	ge.pendingCatastrophe = ""

	ge.SimulateOffline(time.Hour)

	if ge.age != "stone_age" {
		t.Fatalf("age = %s, want the plan to have advanced to stone_age", ge.age)
	}
	camps := ge.Buildings.GetCount("stone_camp")
	if camps != 3 {
		t.Fatalf("stone camps built = %d, want 3", camps)
	}
	if got, want := ge.Workers.GetAssignedCount("worker", "stone_camp"), 3*camps; got != want {
		t.Errorf("stone camps staffed with %d workers, want %d", got, want)
	}
	if got := ge.Workers.GetAssignedCount("worker", "gathering_camp"); got != 30 {
		t.Errorf("gathering camps have %d workers, want the 30 they had (the plan must not move food workers)", got)
	}
	if got := ge.Workers.GetAssignedCount("worker", "wood_camp"); got != 30-3*camps {
		t.Errorf("wood camps have %d workers, want %d (the stone camps' workers came from them)", got, 30-3*camps)
	}
	if pop := ge.Workers.TotalPop(); pop != 60 {
		t.Errorf("population = %d, want 60 (the plan never recruits)", pop)
	}
}

// Workers in buildings of the current age are the player's arrangement: a
// plan copy with no idle hands to fill it waits for them.
func TestPlan_StaffingLeavesCurrentBuildingsAlone(t *testing.T) {
	ge := planTestEngine(t)
	ge.Buildings.counts["hut"] = 5
	ge.Buildings.counts["wood_camp"] = 3
	ge.recalculateRates()
	if err := ge.RecruitWorker("worker", 9); err != nil {
		t.Fatal(err)
	}
	if err := ge.AssignWorker("wood_camp", 9); err != nil {
		t.Fatal(err)
	}
	setAmount(ge, "wood", 1000)
	if _, err := ge.PlanAddBuild("gathering_camp", 1); err != nil {
		t.Fatal(err)
	}
	ge.runPlanTick()
	for i := 0; i < 100 && queued(ge, "gathering_camp") > 0; i++ {
		ge.processBuildQueue()
	}
	if got := ge.Workers.GetAssignedCount("worker", "gathering_camp"); got != 0 {
		t.Errorf("the plan moved %d workers out of this age's wood camps", got)
	}
}

// Offline trades meet a market that recovers as the time passes.
func TestOffline_TradesAsTimePasses(t *testing.T) {
	ge := bronzeEngine(t)
	ge.Buildings.counts["lumber_mill"] = 20
	ge.recalculateRates()
	for _, r := range []string{"wood", "stone"} {
		ge.Resources.resources[r].Storage = 1e6
	}
	setAmount(ge, "stone", 0)
	if err := ge.PlanAddTrade("wood", "stone", 0); err != nil {
		t.Fatal(err)
	}
	ge.SimulateOffline(2 * time.Hour)
	if ge.Resources.Get("stone") <= 0 {
		t.Error("two hours offline with a trade planned bought no stone")
	}
	if !logHasAll(ge, "While you were away", "traded ") {
		t.Error("no offline trade summary")
	}
}
