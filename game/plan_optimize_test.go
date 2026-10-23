package game

import (
	"reflect"
	"testing"
	"time"
)

// optimizeTown is a Primitive Age town that is already running: two huts, three
// wood camps and two gathering camps with their crews, stocks to build from.
func optimizeTown(t *testing.T) *GameEngine {
	t.Helper()
	ge := planTestEngine(t)
	ge.Buildings.counts["hut"] = 2
	ge.Buildings.counts["wood_camp"] = 3
	ge.Buildings.counts["gathering_camp"] = 2
	ge.Workers.Recruit("worker", 12, ge.popCapLocked())
	ge.staffByShares(false)
	ge.recalculateRates()
	setAmount(ge, "wood", 200)
	setAmount(ge, "food", 100)
	return ge
}

func planBuilds(t *testing.T, ge *GameEngine, plan ...any) {
	t.Helper()
	for i := 0; i < len(plan); i += 2 {
		if _, err := ge.PlanAddBuild(plan[i].(string), plan[i+1].(int)); err != nil {
			t.Fatalf("plan build %v: %v", plan[i], err)
		}
	}
}

// buildTotals is the copies of each building the plan's build items hold.
func buildTotals(plan []PlanItem) map[string]int {
	out := map[string]int{}
	for _, it := range plan {
		if it.Kind == PlanBuild {
			out[it.Key] += it.Count
		}
	}
	return out
}

// ticksToEmpty plays the engine until its plan is empty and says how many
// ticks that took (-1 if it never was).
func ticksToEmpty(ge *GameEngine, limit int) int {
	for i := 1; i <= limit; i++ {
		ge.StepTicks(1)
		if len(ge.plan) == 0 {
			return i
		}
	}
	return -1
}

func withinShare(got, want, share float64) bool {
	return got >= want*(1-share) && got <= want*(1+share)
}

// Huts, wood camps and gathering camps planned one after the other are mixed:
// the town grows faster when its homes and producers come together. The
// engine itself, playing both plans, agrees with the estimate.
func TestPlanOptimize_MixesHomesAndProducers(t *testing.T) {
	plain, tuned := optimizeTown(t), optimizeTown(t)
	for _, ge := range []*GameEngine{plain, tuned} {
		planBuilds(t, ge, "hut", 10, "wood_camp", 10, "gathering_camp", 10)
	}
	res, err := tuned.PlanOptimize()
	if err != nil {
		t.Fatal(err)
	}
	if !res.Changed || res.After >= res.Before {
		t.Fatalf("result = %+v, want a faster order", res)
	}
	if got := len(tuned.plan); got <= 3 {
		t.Fatalf("plan has %d items, want the three kinds mixed: %v", got, res.Order)
	}
	items := map[string]int{}
	for _, it := range tuned.plan {
		items[it.Key]++
	}
	mixed := false
	for _, n := range items {
		mixed = mixed || n > 1
	}
	if !mixed {
		t.Errorf("no building comes in more than one item: %v", res.Order)
	}
	want := map[string]int{"hut": 10, "wood_camp": 10, "gathering_camp": 10}
	if got := buildTotals(tuned.plan); !reflect.DeepEqual(got, want) {
		t.Errorf("totals = %v, want %v", got, want)
	}
	if planLoad(tuned.plan) > MaxPlanItems {
		t.Errorf("plan load %d is over the limit", planLoad(tuned.plan))
	}

	slow, fast := ticksToEmpty(plain, 5000), ticksToEmpty(tuned, 5000)
	if slow < 0 || fast < 0 || fast >= slow {
		t.Fatalf("engine took %d ticks in the new order, %d in the old; want fewer", fast, slow)
	}
	iv := BaseTickInterval
	if est := float64(res.Before / iv); !withinShare(est, float64(slow), 0.25) {
		t.Errorf("estimated %v ticks for the old order, engine took %d", est, slow)
	}
	if est := float64(res.After / iv); !withinShare(est, float64(fast), 0.25) {
		t.Errorf("estimated %v ticks for the new order, engine took %d", est, fast)
	}
}

// Homes come first when the producers have no one to work them.
func TestPlanOptimize_HomesBeforeCampsThatNeedPeople(t *testing.T) {
	ge := planTestEngine(t)
	ge.Buildings.counts["hut"] = 1
	ge.Buildings.counts["wood_camp"] = 3
	ge.Buildings.counts["gathering_camp"] = 2
	ge.Workers.Recruit("worker", 10, ge.popCapLocked())
	ge.staffByShares(false)
	ge.recalculateRates()
	setAmount(ge, "wood", 100)
	setAmount(ge, "food", 100)
	planBuilds(t, ge, "wood_camp", 3, "hut", 3)
	res, err := ge.PlanOptimize()
	if err != nil {
		t.Fatal(err)
	}
	if !res.Changed || ge.plan[0].Key != "hut" {
		t.Fatalf("plan = %v, want the huts first (homes are full, the camps would stand empty)", res.Order)
	}
}

// Research, deals and the advance keep their order; research stays where
// it is among the builds, counted in copies. Items the optimizer cannot value
// (storage) stay put the same way.
func TestPlanOptimize_KeepsTheOtherItemsInOrder(t *testing.T) {
	ge := optimizeTown(t)
	planBuilds(t, ge, "hut", 4)
	if err := ge.PlanAddResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	planBuilds(t, ge, "wood_camp", 5, "stash", 2, "gathering_camp", 5)
	if err := ge.PlanAddResearch("fire_mastery"); err != nil {
		t.Fatal(err)
	}
	planBuilds(t, ge, "hut", 3)
	if err := ge.PlanAddAdvance(); err != nil {
		t.Fatal(err)
	}
	before := append([]PlanItem(nil), ge.plan...)
	res, err := ge.PlanOptimize()
	if err != nil {
		t.Fatal(err)
	}
	if !res.Changed {
		t.Fatalf("plan not reordered: %v", res.Order)
	}
	kinds := func(plan []PlanItem, kind string) []string {
		var out []string
		for _, it := range plan {
			if it.Kind == kind {
				out = append(out, it.Key)
			}
		}
		return out
	}
	if got, want := kinds(ge.plan, PlanResearch), kinds(before, PlanResearch); !reflect.DeepEqual(got, want) {
		t.Errorf("research = %v, want %v", got, want)
	}
	if last := ge.plan[len(ge.plan)-1]; last.Kind != PlanAdvance {
		t.Errorf("the advance is not last: %v", res.Order)
	}
	wantTotals := map[string]int{"hut": 7, "wood_camp": 5, "stash": 2, "gathering_camp": 5}
	if got := buildTotals(ge.plan); !reflect.DeepEqual(got, wantTotals) {
		t.Errorf("totals = %v, want %v", got, wantTotals)
	}
	// The stash comes after the same 9 producers and homes as before: the
	// 4 huts, the wood camps, nothing else of the first nine.
	copies, stash := 0, -1
	for _, it := range ge.plan {
		if it.Kind == PlanBuild && it.Key == "stash" {
			stash = copies
		}
		if it.Kind == PlanBuild && it.Key != "stash" {
			copies += it.Count
		}
	}
	if stash != 9 {
		t.Errorf("the stash follows %d copies, want 9", stash)
	}
}

// Everything after an advance stays as it was: the next age is not the
// town the model knows.
func TestPlanOptimize_LeavesWhatFollowsAnAdvance(t *testing.T) {
	ge := optimizeTown(t)
	planBuilds(t, ge, "hut", 5, "wood_camp", 5)
	if err := ge.PlanAddAdvance(); err != nil {
		t.Fatal(err)
	}
	planBuilds(t, ge, "longhouse", 4, "woodcutter_camp", 2, "longhouse", 1)
	tail := append([]PlanItem(nil), ge.plan[2:]...)
	if _, err := ge.PlanOptimize(); err != nil {
		t.Fatal(err)
	}
	if got := ge.plan[len(ge.plan)-len(tail):]; !reflect.DeepEqual(got, tail) {
		t.Errorf("tail = %+v, want %+v", got, tail)
	}
}

// flatPrices makes the three Primitive Age basics cost the same for every copy,
// so a thousand copies stay within reach of the town's income.
func flatPrices(ge *GameEngine, scale float64) {
	for _, key := range []string{"hut", "wood_camp", "gathering_camp"} {
		def := ge.Buildings.defs[key]
		def.CostScale = scale
		ge.Buildings.injectDef(def)
	}
	ge.recalculateRates()
}

// A plan can hold 1,000 copies of a building. The optimizer plays the first
// 2,000 and keeps the rest in order after them, and the result still fits the
// plan.
func TestPlanOptimize_CapsTheSimulation(t *testing.T) {
	ge := optimizeTown(t)
	flatPrices(ge, 1)
	planBuilds(t, ge, "hut", 700, "wood_camp", 700, "gathering_camp", 700, "wood_camp", 900)
	started := time.Now()
	res, err := ge.PlanOptimize()
	if err != nil {
		t.Fatal(err)
	}
	if took := time.Since(started); took > 10*time.Second {
		t.Errorf("optimizing 3,000 copies took %v", took)
	}
	if !res.Changed || res.Beyond != 1000 {
		t.Errorf("result = %+v, want a reordered plan with 1000 copies beyond the simulation", res)
	}
	want := map[string]int{"hut": 700, "wood_camp": 1600, "gathering_camp": 700}
	if got := buildTotals(ge.plan); !reflect.DeepEqual(got, want) {
		t.Errorf("totals = %v, want %v", got, want)
	}
	if planLoad(ge.plan) > MaxPlanItems {
		t.Errorf("plan load %d is over the limit", planLoad(ge.plan))
	}
	tail := ge.plan[len(ge.plan)-2:]
	if tail[0].Key != "gathering_camp" || tail[0].Count != 100 || tail[1].Key != "wood_camp" || tail[1].Count != 900 {
		t.Errorf("last items = %+v, want the 100 gathering camps and the 900 wood camps past the cap, in order", tail)
	}
	for _, it := range ge.plan {
		if it.Count > maxPlanCount {
			t.Errorf("item %+v is over %d", it, maxPlanCount)
		}
	}
}

// A plan of 60 items that the optimizer would split into more than 60 is
// bought in bigger steps until it fits.
func TestPlanOptimize_FitsThePlan(t *testing.T) {
	ge := optimizeTown(t)
	flatPrices(ge, 1.02)
	keys := []string{"hut", "wood_camp", "gathering_camp"}
	for i := 0; i < MaxPlanItems; i++ {
		planBuilds(t, ge, keys[i%3], 30)
	}
	totals := buildTotals(ge.plan)
	res, err := ge.PlanOptimize()
	if err != nil {
		t.Fatal(err)
	}
	if res.After > res.Before || planLoad(ge.plan) > MaxPlanItems {
		t.Errorf("before %v, after %v, load %d", res.Before, res.After, planLoad(ge.plan))
	}
	if got := buildTotals(ge.plan); !reflect.DeepEqual(got, totals) {
		t.Errorf("totals = %v, want %v", got, totals)
	}
}

// A plan the optimizer cannot improve is left alone and says so, and so is
// one it has already improved.
func TestPlanOptimize_LeavesAnOptimalPlanAlone(t *testing.T) {
	ge := optimizeTown(t)
	planBuilds(t, ge, "wood_camp", 5)
	before := append([]PlanItem(nil), ge.plan...)
	res, err := ge.PlanOptimize()
	if err != nil {
		t.Fatal(err)
	}
	if res.Changed || res.Reason != PlanOptimizeBest || !reflect.DeepEqual(ge.plan, before) {
		t.Errorf("result = %+v, plan = %+v; want the plan left as it was", res, ge.plan)
	}

	ge = optimizeTown(t)
	planBuilds(t, ge, "hut", 10, "wood_camp", 10, "gathering_camp", 10)
	if res, _ = ge.PlanOptimize(); !res.Changed {
		t.Fatalf("first pass changed nothing: %v", res.Order)
	}
	once := append([]PlanItem(nil), ge.plan...)
	res, err = ge.PlanOptimize()
	if err != nil {
		t.Fatal(err)
	}
	if res.Changed || res.Reason != PlanOptimizeBest || !reflect.DeepEqual(ge.plan, once) {
		t.Errorf("second pass = %+v, plan changed to %+v", res, ge.plan)
	}
}

func TestPlanOptimize_NothingToDo(t *testing.T) {
	ge := optimizeTown(t)
	if _, err := ge.PlanOptimize(); err == nil {
		t.Error("an empty plan was reordered")
	}
	if err := ge.PlanAddResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	planBuilds(t, ge, "stash", 2)
	res, err := ge.PlanOptimize()
	if err != nil || res.Changed || res.Reason != PlanOptimizeNothing {
		t.Errorf("result = %+v, %v; want nothing to reorder", res, err)
	}
}

// The same town and plan give the same order, run after run.
func TestPlanOptimize_Deterministic(t *testing.T) {
	var first []PlanItem
	var firstRes PlanOptimizeResult
	for i := 0; i < 3; i++ {
		ge := optimizeTown(t)
		planBuilds(t, ge, "hut", 10, "wood_camp", 10, "gathering_camp", 10, "stash", 1)
		res, err := ge.PlanOptimize()
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first, firstRes = ge.plan, res
			continue
		}
		if !reflect.DeepEqual(ge.plan, first) || !reflect.DeepEqual(res, firstRes) {
			t.Fatalf("run %d: plan %v, result %+v; first plan %v, result %+v", i, ge.planLabels(ge.plan), res, ge.planLabels(first), firstRes)
		}
	}
}

// Every age has its own prices, crews and homes; the optimizer reads them
// from the game, so it plays a town of any age and never makes the plan slower.
func TestPlanOptimize_EveryAge(t *testing.T) {
	for _, age := range ageKeys() {
		ge := newSeededEngine(1)
		advanceTo(ge, age)
		ge.mu.Lock()
		var keys []string
		stock := map[string]float64{}
		for _, key := range ge.Buildings.order {
			def := ge.Buildings.defs[key]
			if def.RequiredAge != age || def.Category != "production" && def.Category != "housing" {
				continue
			}
			keys = append(keys, key)
			ge.Buildings.counts[key] = 2
			for res, c := range def.BaseCost {
				stock[res] = max(stock[res], 30*c)
			}
		}
		ge.Workers.Recruit("worker", 40, ge.popCapLocked())
		ge.staffByShares(false)
		ge.recalculateRates()
		ge.mu.Unlock()
		for res, v := range stock {
			setAmount(ge, res, v)
		}
		for _, key := range keys[:min(len(keys), 5)] {
			if _, err := ge.PlanAddBuild(key, 12); err != nil {
				t.Fatalf("%s: plan %s: %v", age, key, err)
			}
		}
		totals := buildTotals(ge.plan)
		res, err := ge.PlanOptimize()
		if err != nil {
			t.Fatalf("%s: %v", age, err)
		}
		if res.After > res.Before {
			t.Errorf("%s: %v before, %v after", age, res.Before, res.After)
		}
		if got := buildTotals(ge.plan); !reflect.DeepEqual(got, totals) {
			t.Errorf("%s: totals = %v, want %v", age, got, totals)
		}
		if planLoad(ge.plan) > MaxPlanItems {
			t.Errorf("%s: plan load %d", age, planLoad(ge.plan))
		}
	}
}

// The walk plays a plan by the plan's rules: an affordable plan starts at
// once, and a copy no income will pay for never does.
func TestPlanSim_Walk(t *testing.T) {
	ge := optimizeTown(t)
	m := ge.newPlanModel([]string{"hut", "wood_camp"})
	hut, camp := m.kidx["hut"], m.kidx["wood_camp"]
	if out := m.walk([]simItem{{key: hut, left: 2}}); out.bought != 2 || out.finish != 0 {
		t.Errorf("two huts on a full store: %+v, want both at tick 0", out)
	}
	setAmount(ge, "wood", 0)
	m = ge.newPlanModel([]string{"hut", "wood_camp"})
	out := m.walk([]simItem{{key: camp, left: 3}})
	if out.bought != 3 || out.finish <= 0 {
		t.Errorf("three camps on an empty store: %+v, want all three after some saving", out)
	}
	ge.Buildings.counts["wood_camp"] = 0
	ge.recalculateRates()
	setAmount(ge, "wood", 0)
	m = ge.newPlanModel([]string{"wood_camp"})
	if out := m.walk([]simItem{{key: 0, left: 1}}); out.bought != 0 {
		t.Errorf("with no wood and no income: %+v, want nothing to start", out)
	}
}
