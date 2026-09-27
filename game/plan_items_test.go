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
	if v := be.planViews(); len(v) != 1 || !strings.Contains(v[0].Note, "market") {
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
	if !logHas(ge, "While you were away your plan started: traded") {
		t.Error("no offline trade summary")
	}
}
