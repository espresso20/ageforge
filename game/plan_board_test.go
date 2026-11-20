package game

import (
	"reflect"
	"testing"
)

// A held plan starts nothing and drops nothing, its views say what running
// again would do, and any clear runs it again.
func TestPlanPause_HoldsTheWalk(t *testing.T) {
	ge := planTestEngine(t)
	setAmount(ge, "wood", 500)
	planBuilds(t, ge, "hut", 1)
	ge.PlanSetPaused(true)
	if !logHas(ge, "Plan paused.") {
		t.Error("no log line for the pause")
	}
	ge.runPlanTick()
	if queued(ge, "hut") != 0 || len(ge.plan) != 1 {
		t.Fatalf("a held plan started %d huts, plan %v", queued(ge, "hut"), ge.plan)
	}
	if v := ge.planViews(); len(v) != 1 || v[0].Status != PlanStatusReady || v[0].Cause != "" {
		t.Errorf("held view = %+v, want ready as running would show", v)
	}
	if !ge.GetState().PlanPaused {
		t.Error("the snapshot does not say the plan is held")
	}
	ge.PlanSetPaused(false)
	ge.runPlanTick()
	if queued(ge, "hut") != 1 {
		t.Error("the plan did not start the hut once resumed")
	}
	ge.PlanSetPaused(true)
	planBuilds(t, ge, "hut", 1)
	ge.PlanClear()
	if ge.planPaused {
		t.Error("clearing the plan left it held")
	}
}

// A held plan is saved with the run.
func TestPlanPause_Saved(t *testing.T) {
	ge := planTestEngine(t)
	planBuilds(t, ge, "hut", 1)
	ge.PlanSetPaused(true)
	ge.mu.RLock()
	save := ge.buildSaveSnapshot()
	ge.mu.RUnlock()
	if !save.PlanPaused {
		t.Fatal("the save does not hold the pause")
	}
}

// PlanSetCount keeps PlanAdjust's limits.
func TestPlanSetCount(t *testing.T) {
	ge := planTestEngine(t)
	planBuilds(t, ge, "hut", 2)
	if n, err := ge.PlanSetCount(1, 9); err != nil || n != 9 {
		t.Fatalf("set 9 = %d, %v", n, err)
	}
	if n, err := ge.PlanSetCount(1, 3); err != nil || n != 3 {
		t.Fatalf("set 3 = %d, %v", n, err)
	}
	if n, err := ge.PlanSetCount(1, 1<<20); err != nil || n != maxPlanCount {
		t.Fatalf("set past the cap = %d, %v, want %d", n, err, maxPlanCount)
	}
	if n, err := ge.PlanSetCount(1, 0); err != nil || n != 1 {
		t.Fatalf("set 0 = %d, %v, want 1", n, err)
	}
	if _, err := ge.PlanSetCount(1, 0); err == nil {
		t.Error("lowered a count of 1")
	}
	if _, err := ge.PlanSetCount(4, 2); err == nil {
		t.Error("set the count of an item that is not there")
	}
	if err := ge.PlanAddAdvance(); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanSetCount(2, 2); err == nil {
		t.Error("gave an advance a count")
	}
}

// PlanMove reaches the top and the bottom with the deltas the panel sends.
func TestPlanMove_TopBottomAndDrop(t *testing.T) {
	ge := planTestEngine(t)
	planBuilds(t, ge, "hut", 1, "wood_camp", 1, "shrine", 1, "stash", 1)
	order := func() []string {
		var keys []string
		for _, it := range ge.plan {
			keys = append(keys, it.Key)
		}
		return keys
	}
	n := len(ge.plan)
	if at, err := ge.PlanMove(3, 1-3); err != nil || at != 1 {
		t.Fatalf("to top = %d, %v", at, err)
	}
	if at, err := ge.PlanMove(1, n-1); err != nil || at != n {
		t.Fatalf("to bottom = %d, %v", at, err)
	}
	if want := []string{"hut", "wood_camp", "stash", "shrine"}; !reflect.DeepEqual(order(), want) {
		t.Fatalf("order = %v, want %v", order(), want)
	}
	if at, err := ge.PlanMove(1, 2); err != nil || at != 3 {
		t.Fatalf("drop = %d, %v", at, err)
	}
}

// Cause says what each kind of wait is.
func TestPlanCause(t *testing.T) {
	ge := planTestEngine(t)
	setAmount(ge, "wood", 0)
	planBuilds(t, ge, "hut", 1)
	cause := func() string { return ge.planViews()[0].Cause }
	if got := cause(); got != PlanCauseFunds {
		t.Errorf("a price the income will cover: %q, want funds", got)
	}
	ge.Resources.resources["wood"].Rate = 0
	setAmount(ge, "wood", 0)
	if got := cause(); got != PlanCauseIncome {
		t.Errorf("a price no income covers: %q, want income", got)
	}
	setAmount(ge, "wood", 500)
	if got := cause(); got != "" {
		t.Errorf("a price the stores cover: %q, want ready", got)
	}
	ge.PlanClear()
	for _, k := range []string{"pottery", "tool_making"} {
		if err := ge.PlanAddResearch(k); err != nil {
			t.Fatal(err)
		}
	}
	ge.Research.currentTech = "fire_mastery"
	for _, v := range ge.planViews() {
		if v.Cause != PlanCauseResearch && v.Cause != PlanCauseAge {
			t.Errorf("%s waits on %q, want research or age", v.Key, v.Cause)
		}
	}
	ge.PlanClear()
	if err := ge.PlanAddAdvance(); err != nil {
		t.Fatal(err)
	}
	if got := cause(); got != PlanCauseHand {
		t.Errorf("an advance short of its pillars: %q, want hand", got)
	}
}

// Needs sums the price of every copy left, rising prices included, and says
// when the income covers it; an advance ends the scope.
func TestPlanEstimate_Needs(t *testing.T) {
	ge := planTestEngine(t)
	setAmount(ge, "wood", 10)
	planBuilds(t, ge, "hut", 3, "wood_camp", 1)
	want := 0.0
	for i := 0; i < 3; i++ {
		want += ge.Buildings.CopyCost("hut", i, nil)["wood"]
	}
	want += ge.Buildings.CopyCost("wood_camp", 0, nil)["wood"]
	est := ge.PlanEstimate()
	var wood *PlanNeed
	for i := range est.Needs {
		if est.Needs[i].Res == "wood" {
			wood = &est.Needs[i]
		}
	}
	if wood == nil {
		t.Fatalf("needs = %+v, want wood", est.Needs)
	}
	if !withinShare(wood.Need, want, 1e-9) || wood.Have != 10 || wood.Never || wood.Waiting != 2 {
		t.Errorf("wood need = %+v, want need %v have 10, 2 waiting", *wood, want)
	}
	rate := ge.Resources.GetRate("wood") / ge.tickIntervalLocked().Seconds()
	if !withinShare(wood.CoverIn.Seconds(), (want-10)/rate, 1e-6) {
		t.Errorf("covered in %v, want %vs", wood.CoverIn, (want-10)/rate)
	}
	// Past an advance nothing counts.
	if err := ge.PlanAddAdvance(); err != nil {
		t.Fatal(err)
	}
	planBuilds(t, ge, "hut", 50)
	if after := ge.PlanEstimate(); !withinShare(after.Needs[0].Need, want, 1e-9) {
		t.Errorf("need after an advance = %v, want %v", after.Needs[0].Need, want)
	}
	// No income sorts first.
	ge.Buildings.counts["wood_camp"] = 0
	ge.estimateCache = estimateCache{}
	if n := ge.PlanEstimate().Needs; len(n) == 0 || !n[0].Never {
		t.Errorf("needs with no income = %+v, want never first", n)
	}
}

// Copies lists the start of each copy of a modelled item of 2 to 12 copies,
// and a research item gets an end.
func TestPlanEstimate_CopiesAndResearchEnd(t *testing.T) {
	ge := planTestEngine(t)
	setAmount(ge, "wood", 0)
	planBuilds(t, ge, "hut", 4, "wood_camp", 20)
	if err := ge.PlanAddResearch("fire_mastery"); err != nil {
		t.Fatal(err)
	}
	est := ge.PlanEstimate()
	huts := est.Items[0]
	if len(huts.Copies) != 4 || huts.Copies[0] != huts.Start || huts.Copies[3] != huts.End {
		t.Errorf("hut copies %v, start %v end %v", huts.Copies, huts.Start, huts.End)
	}
	for i := 1; i < len(huts.Copies); i++ {
		if huts.Copies[i] < huts.Copies[i-1] {
			t.Errorf("copies run backwards: %v", huts.Copies)
		}
	}
	if len(est.Items[1].Copies) != 0 {
		t.Errorf("20 copies kept %d starts, want none", len(est.Items[1].Copies))
	}
	if r := est.Items[2]; r.End <= r.Start {
		t.Errorf("research start %v end %v, want a length", r.Start, r.End)
	}
}

// Rows say where each piece came from: two pieces of one item are a split.
func TestPlanOptimize_RowsFrom(t *testing.T) {
	ge := optimizeTown(t)
	planBuilds(t, ge, "hut", 10, "wood_camp", 10, "gathering_camp", 10)
	res, err := ge.PlanOptimizePreview()
	if err != nil || !res.Changed {
		t.Fatalf("preview = %+v, %v", res, err)
	}
	if len(res.Rows) != len(res.Order) {
		t.Fatalf("%d rows for %d items", len(res.Rows), len(res.Order))
	}
	from := map[int]int{}
	for i, r := range res.Rows {
		if r.From < 1 || r.From > 3 {
			t.Fatalf("row %d from %d, want a place of the old plan", i, r.From)
		}
		from[r.From] += r.Count
	}
	if from[1] != 10 || from[2] != 10 || from[3] != 10 {
		t.Errorf("copies by source = %v, want 10 each", from)
	}
	split := false
	seen := map[int]bool{}
	for _, r := range res.Rows {
		split = split || seen[r.From]
		seen[r.From] = true
	}
	if !split {
		t.Error("no item was split in a mixed order")
	}
	if last := res.Rows[len(res.Rows)-1]; last.End <= 0 {
		t.Errorf("last row has no end: %+v", last)
	}
}

// Two items of one name keep their own places: the first copies come from the first.
func TestPlanOptimize_RowsTwoItemsOfOneName(t *testing.T) {
	ge := optimizeTown(t)
	planBuilds(t, ge, "hut", 3, "wood_camp", 2)
	ge.mu.Lock()
	ge.plan = append(ge.plan, PlanItem{Kind: PlanBuild, Key: "hut", Count: 4})
	rows := ge.planOptimizeRows([]PlanItem{{Kind: PlanBuild, Key: "hut", Count: 5}, {Kind: PlanBuild, Key: "wood_camp", Count: 2}, {Kind: PlanBuild, Key: "hut", Count: 2}})
	ge.mu.Unlock()
	if rows[0].From != 1 || rows[1].From != 2 || rows[2].From != 3 {
		t.Errorf("sources = %d %d %d, want 1 2 3", rows[0].From, rows[1].From, rows[2].From)
	}
}
