package game

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// Overflow does not pay the plan (overflow.go): what a full store cannot hold
// goes to the age's wonder when wonder overflow is on, and is otherwise
// lost. A plan item banks nothing. What is left here: a wonder queued in the
// plan, and the banks a save from before the storage wall still carries.

// overflowEngine is planTestEngine with wonder overflow off and wood storage
// set to cap.
func overflowEngine(t *testing.T, cap float64) *GameEngine {
	t.Helper()
	ge := planTestEngine(t)
	ge.wonderOverflowOff = true
	ge.Resources.resources["wood"].Storage = cap
	setAmount(ge, "wood", cap)
	return ge
}

// A wonder queued in the plan is paid the way a wonder is: its bank takes
// what the stores discard while wonder overflow is on, and the plan pays the
// rest and starts it once the stores cover what the bank lacks. With wonder
// overflow off nothing is banked, planned or not: the switch means what it
// says, and the item waits on its bank. No other plan item banks anything.
func TestPlanOverflow_QueuedWonderBanks(t *testing.T) {
	ge := planTestEngine(t)
	ge.wonderOverflowOff = true
	w := ge.progress.WonderForAge(ge.age)
	need := ge.Buildings.defs[w].BaseCost
	if need["wood"] <= 0 || need["food"] <= 0 {
		t.Fatalf("setup: %s costs %v, want wood and food", w, need)
	}
	for k := range ge.Resources.resources {
		ge.Resources.SetRate(k, 0)
	}
	woodCap := need["wood"] / 4 // the wood part is four full stores
	ge.Resources.resources["wood"].Storage = woodCap
	setAmount(ge, "wood", woodCap)
	rate := need["wood"] / 10
	ge.Resources.SetRate("wood", rate)
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild(w, 1); err != nil {
		t.Fatal(err)
	}

	// Wonder overflow off: the tick's surplus is lost. Neither the hut nor
	// the wonder banks it, and the store stays at its cap.
	ge.applyTickRates()
	if len(ge.Buildings.wonderBanks[w]) != 0 {
		t.Errorf("wonder overflow is off, but the planned wonder's bank holds %v", ge.Buildings.wonderBanks[w])
	}
	if ge.plan[0].Banked != nil || ge.plan[1].Banked != nil {
		t.Errorf("plan items banked overflow: hut %v, wonder %v", ge.plan[0].Banked, ge.plan[1].Banked)
	}
	if got := ge.Resources.Get("wood"); got != woodCap {
		t.Errorf("the store changed to %v: a full store stays full", got)
	}
	st := ge.GetState()
	if v := st.Plan[1]; v.Status != PlanStatusBlocked || v.Note != "bank not full" {
		t.Errorf("the wonder is %s (%s) while four stores of wood are still to bank, want blocked (bank not full)", v.Status, v.Note)
	}

	// Wonder overflow on: the whole surplus goes to the wonder, wherever it
	// sits in the plan, and still none to the hut.
	ge.wonderOverflowOff = false
	ge.applyTickRates()
	if got := ge.Buildings.wonderBanks[w]["wood"]; abs(got-rate) > 1e-9 || ge.plan[0].Banked != nil {
		t.Errorf("wonder overflow on: the wonder banked %v and the hut %v, want the whole %v in the wonder", got, ge.plan[0].Banked, rate)
	}

	// The plan walk runs with the ticks: the hut is paid from the store,
	// and the overflow goes on into the wonder until the stores cover what
	// its bank lacks. Then the plan pays the rest and starts the wonder.
	// Nobody deposited anything.
	ge.Resources.resources["food"].Storage = 2 * need["food"]
	setAmount(ge, "food", need["food"])
	for i := 0; i < 20 && queued(ge, w) == 0; i++ {
		ge.applyTickRates()
		ge.runPlanTick()
	}
	if queued(ge, w) != 1 {
		t.Fatalf("the queued wonder never started: bank %v of %v, plan %+v", ge.Buildings.wonderBanks[w], need, ge.plan)
	}
	if queued(ge, "hut") != 1 {
		t.Errorf("the hut above the wonder did not start")
	}
	if !logHas(ge, "Overflow finished banking") && ge.Buildings.wonderBanks[w]["wood"] < need["wood"]-woodCap-1e-6 {
		t.Errorf("the wonder started with %v wood banked of %v and a store of %v", ge.Buildings.wonderBanks[w]["wood"], need["wood"], woodCap)
	}
}

// Offline catch-up fills the age's wonder's bank the same way, starts a
// planned wonder when the stores cover the rest, and says what was banked.
// With wonder overflow off it banks nothing, planned or not.
func TestPlanOverflow_QueuedWonderBanksOffline(t *testing.T) {
	away := func(overflow, planned bool) (*GameEngine, string) {
		ge := planTestEngine(t)
		ge.wonderOverflowOff = !overflow
		ge.Buildings.counts["wood_camp"] = 20
		ge.Buildings.counts["gathering_camp"] = 20
		ge.Buildings.counts["stash"] = 0
		ge.recalculateRates()
		w := ge.progress.WonderForAge(ge.age)
		need := ge.Buildings.defs[w].BaseCost
		woodCap := ge.Resources.GetStorage("wood")
		if need["wood"] <= woodCap {
			t.Fatalf("setup: the wonder's %v wood fits the %v cap", need["wood"], woodCap)
		}
		if ge.Resources.GetRate("wood") <= 0 || ge.Resources.GetRate("food") <= 0 {
			t.Fatalf("setup: wood %v/tick, food %v/tick", ge.Resources.GetRate("wood"), ge.Resources.GetRate("food"))
		}
		if planned {
			ge.plan = []PlanItem{{Kind: PlanBuild, Key: w, Count: 1}}
		}
		setAmount(ge, "wood", woodCap)
		ge.SimulateOffline(8 * time.Hour)
		return ge, w
	}

	ge, w := away(true, true)
	if queued(ge, w) == 0 && ge.Buildings.GetCount(w) == 0 {
		t.Fatalf("eight hours away with the wonder planned: bank %v of %v, not started (plan %+v)", ge.Buildings.wonderBanks[w], ge.Buildings.defs[w].BaseCost, ge.plan)
	}
	if !logHas(ge, "Overflow banked into "+ge.Buildings.defs[w].Name) {
		t.Error("no offline summary of what the wonder banked")
	}
	if logHas(ge, "toward your plan") {
		t.Error("the offline summary still speaks of overflow banked toward the plan")
	}

	for _, planned := range []bool{true, false} {
		off, w := away(false, planned)
		if len(off.Buildings.wonderBanks[w]) != 0 || queued(off, w) != 0 || off.Buildings.GetCount(w) != 0 {
			t.Errorf("wonder overflow off (planned: %v): bank %v, queued %d, built %d; want nothing banked and nothing started", planned, off.Buildings.wonderBanks[w], queued(off, w), off.Buildings.GetCount(w))
		}
		if planned && (len(off.plan) != 1 || off.plan[0].Banked != nil) {
			t.Errorf("wonder overflow off: the planned wonder's item is %+v, want it waiting with no bank", off.plan)
		}
	}
}

// Removing or clearing puts a bank back in the stores, up to the cap, and
// says so.
func TestPlanOverflow_RemoveAndClearReturnTheBank(t *testing.T) {
	ge := overflowEngine(t, 1000)
	if _, err := ge.PlanAddBuild("shrine", 2); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	ge.plan[0].Banked = map[string]float64{"wood": 300}
	ge.plan[1].Banked = map[string]float64{"wood": 5}
	setAmount(ge, "wood", 900)
	if _, err := ge.PlanRemove(1); err != nil {
		t.Fatal(err)
	}
	if w := ge.Resources.Get("wood"); w != 1000 {
		t.Errorf("wood = %v after removing a 300 bank onto 900 of a 1000 cap, want 1000", w)
	}
	if !logHasAll(ge, "Plan: the bank of", "went back to the stores", "100 wood") {
		t.Error("no log line for the bank going back")
	}
	setAmount(ge, "wood", 0)
	ge.PlanClear()
	if w := ge.Resources.Get("wood"); w != 5 {
		t.Errorf("wood = %v after clearing a plan with 5 banked, want 5", w)
	}
}

// The advance puts the banks back before it trims the stockpiles, so a bank
// carries no more into the new age than the stores could.
func TestPlanOverflow_AdvanceReturnsBanksBeforeTheTrim(t *testing.T) {
	ge := overflowEngine(t, 1e6)
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	ge.plan[0].Banked = map[string]float64{"wood": 1e5}
	setAmount(ge, "wood", 0)
	next := ge.progress.GetNextAge(ge.age)
	trim := CarryoverStarterBuildings * ageEntryCost(next, "wood")
	ge.advanceAge(next)
	if ge.plan[0].Banked != nil {
		t.Errorf("bank after the advance = %v, want returned", ge.plan[0].Banked)
	}
	if w := ge.Resources.Get("wood"); trim > 0 && w > trim+1e-9 {
		t.Errorf("wood after the advance = %v, over the trim's %v: the bank dodged it", w, trim)
	}
	if !logHas(ge, "went back to the stores for the advance") {
		t.Error("no log line for the banks going back at the advance")
	}
}

// ageEntryCost is the advance trim's base for res in age: the cheapest price
// in it among the age's buildings.
func ageEntryCost(age, res string) float64 {
	best := 0.0
	for _, b := range NewGameEngine().Buildings.defs {
		if b.RequiredAge != age || b.Category == "wonder" {
			continue
		}
		if c := b.BaseCost[res]; c > 0 && (best == 0 || c < best) {
			best = c
		}
	}
	return best
}

// A saved bank is taken as the engine would write it: only on build items,
// positive and finite, known resources.
func TestPlanOverflow_LoadPlanCleansBanks(t *testing.T) {
	inf := 1.0
	for i := 0; i < 2000; i++ {
		inf *= 10
	}
	got := loadPlan([]PlanItem{
		{Kind: PlanBuild, Key: "hut", Count: 1, Banked: map[string]float64{"wood": 5, "stone": -1, "nonsense": 3, "food": inf}},
		{Kind: PlanResearch, Key: "tool_making", Count: 1, Banked: map[string]float64{"knowledge": 9}},
		{Kind: PlanBuild, Key: "hut", Count: 1, Banked: map[string]float64{"wood": 0}},
	})
	if !reflect.DeepEqual(got[0].Banked, map[string]float64{"wood": 5}) {
		t.Errorf("build bank = %v, want just the wood", got[0].Banked)
	}
	if got[1].Banked != nil || got[2].Banked != nil {
		t.Errorf("banks %v and %v, want none", got[1].Banked, got[2].Banked)
	}
}

// The plan holds 60 items and refuses the 61st.
func TestPlan_HoldsSixtyItems(t *testing.T) {
	if MaxPlanItems != 60 {
		t.Fatalf("MaxPlanItems = %d, want 60 (site/docs/plan.md says so)", MaxPlanItems)
	}
	ge := planTestEngine(t)
	keys := []string{"hut", "wood_camp"}
	for i := 0; i < MaxPlanItems; i++ {
		// Alternate so the items don't merge.
		if _, err := ge.PlanAddBuild(keys[i%2], 1); err != nil {
			t.Fatalf("item %d refused: %v", i+1, err)
		}
	}
	if len(ge.plan) != 60 {
		t.Fatalf("plan holds %d items, want 60", len(ge.plan))
	}
	// The 60th item is a wood camp: another merges into it, a hut doesn't fit.
	if _, err := ge.PlanAddBuild("wood_camp", 1); err != nil {
		t.Errorf("the same building as the last item should still merge: %v", err)
	}
	if _, err := ge.PlanAddBuild("hut", 1); err == nil || !strings.Contains(err.Error(), "60 items") {
		t.Errorf("the 61st item: err = %v, want the plan-full refusal", err)
	}
	saved := make([]PlanItem, 70)
	for i := range saved {
		saved[i] = PlanItem{Kind: PlanBuild, Key: keys[i%2], Count: 1}
	}
	if got := len(loadPlan(saved)); got != 60 {
		t.Errorf("a saved plan of 70 loads %d items, want 60", got)
	}
}

// NextCost is BuildBatchCost's price for one copy, to the bit, queue and
// cost multiplier included.
func TestNextCost_MatchesBuildBatchCost(t *testing.T) {
	ge := newLateGameEngine(t)
	ge.buildQueue = append(ge.buildQueue, BuildQueueItem{BuildingKey: "factory", TicksLeft: 5, TotalTicks: 5})
	ge.Buildings.costMult = 0.93
	for _, k := range sortedKeys(ge.Buildings.defs) {
		cost, _ := ge.Buildings.BuildBatchCost(k, 1, ge.buildQueue)
		for res, c := range cost {
			if got := ge.Buildings.NextCost(k, res, ge.buildQueue); got != c {
				t.Fatalf("NextCost(%s, %s) = %v, BuildBatchCost says %v", k, res, got, c)
			}
		}
		if got := ge.Buildings.NextCost(k, "no_such_resource", nil); got != 0 {
			t.Fatalf("NextCost of a resource %s doesn't cost = %v", k, got)
		}
	}
}
