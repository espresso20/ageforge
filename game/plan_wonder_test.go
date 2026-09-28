package game

import "testing"

// A planned wonder pays what its bank still lacks from what is held, once
// that covers all of it, and holds nothing back while it waits.
func TestPlan_WonderPaysItsBankFromStock(t *testing.T) {
	ge := planTestEngine(t)
	ge.wonderOverflowOff = true
	w := ge.progress.WonderForAge(ge.age)
	need := ge.Buildings.defs[w].BaseCost
	for res := range need {
		ge.Resources.resources[res].Storage = 1e6
	}
	// Half of each part is already banked.
	ge.Buildings.wonderBanks[w] = map[string]float64{}
	for res, c := range need {
		ge.Buildings.wonderBanks[w][res] = c / 2
	}
	if _, err := ge.PlanAddBuild(w, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	hut := ge.Buildings.GetCost("hut")
	// Short of one part the hut doesn't cost: nothing is banked, and the hut
	// below still starts from what the wonder would have taken.
	short := ""
	for _, res := range sortedKeys(need) {
		if hut[res] == 0 {
			short = res
			break
		}
	}
	if short == "" {
		t.Fatalf("setup: the hut costs every part of %s", w)
	}
	for res, c := range need {
		setAmount(ge, res, c/2)
	}
	for res, c := range hut {
		setAmount(ge, res, ge.Resources.Get(res)+c)
	}
	setAmount(ge, short, need[short]/2-1)
	if v := ge.planViews()[0]; v.Status != PlanStatusWaiting || v.Cost[short] != need[short]/2 {
		t.Errorf("wonder view = %+v, want waiting on the rest of its bank", v)
	}
	ge.runPlanTick()
	if queued(ge, w) != 0 {
		t.Fatal("the wonder started without its bank covered")
	}
	for res, c := range need {
		if got := ge.Buildings.wonderBanks[w][res]; got != c/2 {
			t.Errorf("banked %v %s while waiting, want %v (untouched)", got, res, c/2)
		}
	}
	if queued(ge, "hut") != 1 {
		t.Errorf("the waiting wonder held back the hut (queued %d)", queued(ge, "hut"))
	}
	// Covered: the rest is banked from stock and the wonder starts.
	for res, c := range need {
		setAmount(ge, res, c/2)
	}
	ge.runPlanTick()
	if queued(ge, w) != 1 {
		t.Fatalf("the wonder didn't start with its bank covered (plan %+v)", ge.plan)
	}
	for res := range need {
		if got := ge.Resources.Get(res); got > 1e-6 {
			t.Errorf("%s left = %v, want 0 (the rest was banked)", res, got)
		}
	}
}

// A part bigger than a full store can't be paid in one go: the wonder waits
// for deposits and overflow, blocked.
func TestPlan_WonderPartOverStorageIsBlocked(t *testing.T) {
	ge := planTestEngine(t)
	ge.wonderOverflowOff = true
	w := ge.progress.WonderForAge(ge.age)
	res := sortedKeys(ge.Buildings.defs[w].BaseCost)[0]
	ge.Resources.resources[res].Storage = ge.Buildings.defs[w].BaseCost[res] / 2
	setAmount(ge, res, ge.Resources.resources[res].Storage)
	if _, err := ge.PlanAddBuild(w, 1); err != nil {
		t.Fatal(err)
	}
	if v := ge.planViews()[0]; v.Status != PlanStatusBlocked || v.Note != "bank not full" {
		t.Errorf("view = %+v, want blocked: bank not full", v)
	}
	ge.runPlanTick()
	if queued(ge, w) != 0 || ge.Buildings.wonderBanks[w][res] != 0 {
		t.Error("a part over storage was banked or the wonder started")
	}
}
