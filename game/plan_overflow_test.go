package game

import (
	"strings"
	"testing"
	"time"

	"github.com/espresso20/ageforge/config"
)

// Stock has no ceiling and nothing overflows: a plan item banks nothing. What
// is left here: a wonder queued in the plan, and a saved bank, which a load
// drops (the save from before the storage wall that carried one has it given
// back to the stores: storage_wall_load_test.go).

// A wonder queued in the plan is paid the way a wonder is: the plan pays the
// rest of its bank from the stores and starts it once they cover all of it.
// A store smaller than a part of the price is no obstacle, and while the
// wonder waits nothing is banked. No plan item banks anything.
func TestPlanOverflow_QueuedWonderStartsFromStock(t *testing.T) {
	ge := planTestEngine(t)
	w := ge.progress.WonderForAge(ge.age)
	need := ge.Buildings.defs[w].BaseCost
	if need["wood"] <= 0 || need["food"] <= 0 {
		t.Fatalf("setup: %s costs %v, want wood and food", w, need)
	}
	for k := range ge.Resources.resources {
		ge.Resources.SetRate(k, 0)
	}
	woodCap := need["wood"] / 4 // the town holds a quarter of the wood part
	setAmount(ge, "wood", woodCap)
	setAmount(ge, "food", need["food"])
	if _, err := ge.PlanAddBuild(w, 1); err != nil {
		t.Fatal(err)
	}

	// Short of wood: it waits, and a tick of surplus banks nothing.
	ge.Resources.SetRate("wood", need["wood"]/1000)
	ge.Resources.ApplyRates()
	ge.runPlanTick()
	if len(ge.Buildings.wonderBanks[w]) != 0 {
		t.Errorf("a waiting planned wonder banked: bank %v", ge.Buildings.wonderBanks[w])
	}
	if queued(ge, w) != 0 {
		t.Fatalf("the wonder started with %v wood of the %v it needs", ge.Resources.Get("wood"), need["wood"])
	}
	if v := ge.GetState().Plan[0]; v.Note == "bank not full" {
		t.Errorf("the wonder is blocked (%s) on a store smaller than its price; nothing limits stock", v.Note)
	}

	// Stock past the old store: the plan pays the bank and starts it.
	setAmount(ge, "wood", need["wood"])
	ge.runPlanTick()
	if queued(ge, w) != 1 {
		t.Fatalf("the queued wonder never started: bank %v of %v, plan %+v", ge.Buildings.wonderBanks[w], need, ge.plan)
	}
}

// Offline catch-up is not clipped either: it starts a planned wonder when the
// income away covers the whole price, with no store large enough for it.
func TestPlanOverflow_QueuedWonderStartsOffline(t *testing.T) {
	ge := planTestEngine(t)
	ge.Buildings.counts["wood_camp"] = 20
	ge.Buildings.counts["gathering_camp"] = 20
	ge.Buildings.counts["stash"] = 0
	ge.recalculateRates()
	w := ge.progress.WonderForAge(ge.age)
	need := ge.Buildings.defs[w].BaseCost
	woodHeld := config.WoodBaseStorage // a small stock: what a new town's wood once filled
	if need["wood"] <= woodHeld {
		t.Fatalf("setup: the wonder's %v wood is within the %v held", need["wood"], woodHeld)
	}
	if ge.Resources.GetRate("wood") <= 0 || ge.Resources.GetRate("food") <= 0 {
		t.Fatalf("setup: wood %v/tick, food %v/tick", ge.Resources.GetRate("wood"), ge.Resources.GetRate("food"))
	}
	ge.plan = []PlanItem{{Kind: PlanBuild, Key: w, Count: 1}}
	setAmount(ge, "wood", woodHeld)
	ge.SimulateOffline(8 * time.Hour)
	if queued(ge, w) == 0 && ge.Buildings.GetCount(w) == 0 {
		t.Fatalf("eight hours away with the wonder planned: bank %v of %v, not started (plan %+v)", ge.Buildings.wonderBanks[w], need, ge.plan)
	}
	if logHas(ge, "Overflow banked") || logHas(ge, "toward your plan") {
		t.Error("the offline summary speaks of overflow")
	}
}

// A bank in a saved plan is dropped on load, whatever the item, and the rest
// of the item loads as it was.
func TestPlanOverflow_LoadPlanDropsBanks(t *testing.T) {
	got := loadPlan([]PlanItem{
		{Kind: PlanBuild, Key: "hut", Count: 2, Started: 1, Banked: map[string]float64{"wood": 5}},
		{Kind: PlanResearch, Key: "tool_making", Count: 1, Banked: map[string]float64{"knowledge": 9}},
		{Kind: PlanBuild, Key: "wood_camp", Count: 1},
	})
	if len(got) != 3 {
		t.Fatalf("%d items loaded, want 3", len(got))
	}
	for i, it := range got {
		if it.Banked != nil {
			t.Errorf("item %d loaded with the bank %v, want none", i, it.Banked)
		}
	}
	if got[0].Key != "hut" || got[0].Count != 2 || got[0].Started != 1 {
		t.Errorf("the banked hut loaded as %+v", got[0])
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
