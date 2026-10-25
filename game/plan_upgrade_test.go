package game

import (
	"testing"
)

func upgradePlanEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := planTestEngine(t)
	ge.Buildings.counts["gathering_camp"] = 5
	ge.advanceAge("stone_age")
	return ge
}

func TestPlanAddUpgrade(t *testing.T) {
	ge := upgradePlanEngine(t)
	if _, err := ge.PlanAddUpgrade("hut", 1); err == nil {
		t.Fatal("accepted building with no upgrade")
	}
	n, err := ge.PlanAddUpgrade("gathering_camp", 2)
	if err != nil || n != 2 {
		t.Fatalf("first PlanAddUpgrade = %d, %v; want 2", n, err)
	}
	n, err = ge.PlanAddUpgrade("gathering_camp", 4)
	if err != nil || n != 3 {
		t.Fatalf("capped merge = %d, %v; want 3", n, err)
	}
	if len(ge.plan) != 1 || ge.plan[0].Kind != PlanUpgrade || ge.plan[0].Count != 5 {
		t.Fatalf("plan = %+v", ge.plan)
	}
	t.Cleanup(SetDataDirForTest(t.TempDir()))
	noErr(t, ge.SaveGame("plan_upgrade_roundtrip"))
	loaded := NewGameEngine()
	noErr(t, loaded.LoadGame("plan_upgrade_roundtrip"))
	st := loaded.GetState()
	if len(st.Plan) != 1 || st.Plan[0].Kind != PlanUpgrade || st.Plan[0].Count != 5 {
		t.Fatalf("loaded plan = %+v", st.Plan)
	}
}

func TestPlanUpgradeRunsOneAtATimeAndWaits(t *testing.T) {
	ge := upgradePlanEngine(t)
	if _, err := ge.PlanAddUpgrade("gathering_camp", 2); err != nil {
		t.Fatal(err)
	}
	target, _ := ge.Buildings.UpgradeTarget("gathering_camp")
	to := target.Key
	cost := ge.Buildings.UpgradeCopyCost("gathering_camp", 0, nil)
	for res := range cost {
		setAmount(ge, res, 0)
	}
	ge.runPlanTick()
	if ge.Buildings.GetCount("gathering_camp") != 5 || ge.plan[0].Count != 2 {
		t.Fatalf("started while unaffordable: plan=%+v", ge.plan)
	}
	for res, amount := range cost {
		setAmount(ge, res, amount)
	}
	ge.runPlanTick()
	if ge.Buildings.GetCount("gathering_camp") != 4 || ge.Buildings.GetCount(to) != 1 || ge.plan[0].Count != 1 {
		t.Fatalf("after first: plan=%+v old=%d new=%d", ge.plan, ge.Buildings.GetCount("gathering_camp"), ge.Buildings.GetCount(to))
	}
	cost = ge.Buildings.UpgradeCopyCost("gathering_camp", 0, nil)
	for res, amount := range cost {
		setAmount(ge, res, amount)
	}
	ge.runPlanTick()
	if len(ge.plan) != 0 || ge.Buildings.GetCount(to) != 2 {
		t.Fatalf("item did not finish: plan=%+v upgraded=%d", ge.plan, ge.Buildings.GetCount(to))
	}

	stale := upgradePlanEngine(t)
	if _, err := stale.PlanAddUpgrade("gathering_camp", 1); err != nil {
		t.Fatal(err)
	}
	stale.Buildings.counts["gathering_camp"] = 0 // the copies are gone (sold, or lost to a catastrophe)
	stale.runPlanTick()
	if len(stale.plan) != 0 || !logHas(stale, "Plan: dropped") {
		t.Fatalf("stale upgrade was not dropped: plan=%+v", stale.plan)
	}
}
