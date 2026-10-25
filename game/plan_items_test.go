package game

import (
	"strings"
	"testing"
	"time"
)

func TestPlanAdvance_AdvancesWhenReadyAndDropsTheOldAge(t *testing.T) {
	ge := newSeededEngine(1)
	noErr(t, ge.PlanAddAdvance())
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
	for k, n := range ge.progress.GetRequirementsForNext(ge.age) {
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
	noErr(t, ge.RecruitWorker("worker", 3))
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
// next age's producers, but the old age's copies took every idle worker.
// planAdvanceAway sets that up: 60 workers in the Primitive Age's camps, a
// plan to advance and build three stone camps, then an hour away.
func planAdvanceAway(t *testing.T, autoRecruit bool) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	ge.Buildings.counts["stash"] = 50
	ge.Buildings.counts["hut"] = 10
	ge.Buildings.counts["wood_camp"] = 10
	ge.Buildings.counts["gathering_camp"] = 10
	ge.recalculateRates()
	if !autoRecruit {
		ge.SetAutoRecruit(false)
	}
	noErr(t, ge.RecruitWorker("worker", 60))
	for _, k := range []string{"wood_camp", "gathering_camp"} {
		noErr(t, ge.AssignWorker(k, 30))
	}
	noErr(t, ge.PlanAddAdvance())
	if _, err := ge.PlanAddBuild("stone_camp", 3); err != nil {
		t.Fatal(err)
	}
	// Meet the Stone Age gate by hand.
	for k, n := range ge.progress.GetRequirementsForNext(ge.age) {
		ge.Buildings.counts[k] = max(ge.Buildings.counts[k], n)
	}
	ge.Buildings.counts[ge.progress.WonderForAge(ge.age)] = 1
	ge.pendingCatastrophe = ""

	ge.SimulateOffline(time.Hour)

	if ge.age != "stone_age" {
		t.Fatalf("age = %s, want the plan to have advanced to stone_age", ge.age)
	}
	if camps := ge.Buildings.GetCount("stone_camp"); camps != 3 {
		t.Fatalf("stone camps built = %d, want 3", camps)
	}
	return ge
}

// With auto-recruit off, the plan staffs the new age's copies with workers
// from the buildings the advance superseded, leaves the food producers alone
// and recruits nobody.
func TestOffline_PlanStaffsTheNewAgeAfterAnAdvance(t *testing.T) {
	ge := planAdvanceAway(t, false)
	camps := ge.Buildings.GetCount("stone_camp")
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
		t.Errorf("population = %d, want 60 (auto-recruit is off)", pop)
	}
}

// With auto-recruit on (the default), the new age's copies are staffed and
// the shares routine recruits into the slots the move left behind, as far as
// the housing goes, so nothing waits for the next visit.
func TestOffline_SharesStaffTheNewAgeAfterAnAdvance(t *testing.T) {
	ge := planAdvanceAway(t, true)
	for _, k := range []string{"stone_camp", "wood_camp", "gathering_camp"} {
		if got, want := ge.Workers.GetAssignedCount("worker", k), ge.Buildings.defs[k].WorkerCapacity*ge.Buildings.GetCount(k); got != want {
			t.Errorf("%s has %d of its %d workers after the hour away", k, got, want)
		}
	}
	pop, popCap := ge.Workers.TotalPop(), ge.popCapLocked()
	if pop <= 60 || pop > popCap {
		t.Errorf("population = %d, want recruits past the 60, within the housing's %d", pop, popCap)
	}
	if r := ge.Resources.GetRate("food"); r <= 0 {
		t.Errorf("food rate %v after the offline recruits", r)
	}
}

// Workers in buildings of the current age are the player's arrangement: a
// plan copy with no idle hands to fill it waits for them.
func TestPlan_StaffingLeavesCurrentBuildingsAlone(t *testing.T) {
	ge := planTestEngine(t)
	ge.Buildings.counts["hut"] = 5
	ge.Buildings.counts["wood_camp"] = 3
	ge.recalculateRates()
	noErr(t, ge.RecruitWorker("worker", 9))
	noErr(t, ge.AssignWorker("wood_camp", 9))
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
