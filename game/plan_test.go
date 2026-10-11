package game

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// planTestEngine is a Primitive Age game with room to store things (50
// stashes, and a big knowledge cap for research tests) and a trickle of wood
// and knowledge, so waiting items are saving up rather than stuck. Storage
// is set on the resource directly, so tests that call recalculateRates must
// set it again.
func planTestEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := newSeededEngine(1)
	ge.Buildings.counts["stash"] = 50
	ge.Buildings.counts["wood_camp"] = 1
	ge.Buildings.counts["story_circle"] = 1
	ge.recalculateRates()
	if ge.Resources.GetRate("wood") <= 0 || ge.Resources.GetRate("knowledge") <= 0 {
		t.Fatal("setup: no wood or knowledge income")
	}
	ge.Resources.resources["knowledge"].Storage = 1e6
	return ge
}

func setAmount(ge *GameEngine, res string, v float64) { ge.Resources.resources[res].Amount = v }

func queued(ge *GameEngine, key string) int { return ge.Buildings.GetQueueCount(key, ge.buildQueue) }

func logHas(ge *GameEngine, sub string) bool {
	return logHasAll(ge, sub)
}

// logHasAll reports whether one log line contains every sub.
// logHasLine reports whether some log line is exactly want.
func logHasLine(ge *GameEngine, want string) bool {
	for _, l := range ge.log {
		if l.Message == want {
			return true
		}
	}
	return false
}

func logHasAll(ge *GameEngine, subs ...string) bool {
	for _, l := range ge.log {
		all := true
		for _, sub := range subs {
			if !strings.Contains(l.Message, sub) {
				all = false
				break
			}
		}
		if all {
			return true
		}
	}
	return false
}

func TestPlan_PaysWhenItStartsNotWhenQueued(t *testing.T) {
	ge := planTestEngine(t)
	setAmount(ge, "wood", 0)
	if _, err := ge.PlanAddBuild("hut", 3); err != nil {
		t.Fatal(err)
	}
	ge.runPlanTick()
	if queued(ge, "hut") != 0 || ge.Resources.Get("wood") != 0 {
		t.Fatalf("started with no wood: queue %d, wood %v", queued(ge, "hut"), ge.Resources.Get("wood"))
	}
	two, _ := ge.Buildings.BuildBatchCost("hut", 2, ge.buildQueue)
	setAmount(ge, "wood", two["wood"])
	ge.runPlanTick()
	if queued(ge, "hut") != 2 {
		t.Fatalf("queued huts = %d, want 2 (the price of two copies was held)", queued(ge, "hut"))
	}
	if w := ge.Resources.Get("wood"); w > 1e-9 {
		t.Errorf("wood left = %v, want 0 (paid along the cost curve)", w)
	}
	if len(ge.plan) != 1 || ge.plan[0].Count != 1 || ge.plan[0].Started != 2 {
		t.Errorf("plan = %+v, want one item with 1 left and 2 started", ge.plan)
	}
	if !logHas(ge, "Plan: started building 2 Huts.") {
		t.Error("no summary log line for the two starts")
	}
}

// A waiting item reserves its price: a later item may only use what is left.
func TestPlan_LaterItemsCannotSpendAnEarlierItemsReservation(t *testing.T) {
	ge := planTestEngine(t)
	shrine := ge.Buildings.GetCost("shrine")["wood"]
	hut := ge.Buildings.GetCost("hut")["wood"]
	if _, err := ge.PlanAddBuild("shrine", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	setAmount(ge, "wood", shrine-1) // enough for the hut, not the shrine
	ge.runPlanTick()
	if queued(ge, "hut") != 0 {
		t.Fatal("the hut spent wood the shrine above it is saving")
	}
	v := ge.planViews()
	if v[0].Status != PlanStatusWaiting || v[1].Status != PlanStatusWaiting {
		t.Errorf("statuses = %s, %s; want waiting, waiting", v[0].Status, v[1].Status)
	}
	setAmount(ge, "wood", shrine+hut)
	ge.runPlanTick()
	if queued(ge, "shrine") != 1 || queued(ge, "hut") != 1 {
		t.Errorf("queued shrine %d, hut %d; want 1 and 1", queued(ge, "shrine"), queued(ge, "hut"))
	}
	if len(ge.plan) != 0 {
		t.Errorf("plan not empty after both started: %+v", ge.plan)
	}
}

// A waiting item doesn't hold back resources it doesn't need.
func TestPlan_SkipsPastAWaitingItemForOtherResources(t *testing.T) {
	ge := planTestEngine(t)
	if err := ge.PlanAddResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	setAmount(ge, "knowledge", 0)
	setAmount(ge, "wood", 100)
	ge.runPlanTick()
	if queued(ge, "hut") != 1 {
		t.Error("the hut waited behind research that needs no wood")
	}
	if ge.Research.currentTech != "" || len(ge.plan) != 1 {
		t.Errorf("research started without knowledge, or the item left: %+v", ge.plan)
	}
}

// An item priced over the cap can't be saved for, so it reserves nothing.
func TestPlan_OverCapItemReservesNothing(t *testing.T) {
	ge := newSeededEngine(1) // base storage: 50 wood
	if cost := ge.Buildings.GetCost("shrine")["wood"]; cost <= ge.Resources.GetStorage("wood") {
		t.Skipf("setup: shrine (%v wood) fits the base cap", cost)
	}
	if _, err := ge.PlanAddBuild("shrine", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	setAmount(ge, "wood", 40)
	ge.runPlanTick()
	if queued(ge, "hut") != 1 {
		t.Error("the hut waited behind a shrine that can't fit in storage")
	}
	v := ge.planViews()
	if len(v) != 1 || v[0].Status != PlanStatusBlocked || !strings.Contains(v[0].Note, "storage") {
		t.Errorf("shrine view = %+v, want blocked on storage", v)
	}
}

// An item short of something nothing makes can't start until the player
// acts, so it reserves nothing either.
func TestPlan_UnfundedItemReservesNothing(t *testing.T) {
	ge := planTestEngine(t)
	ge.Resources.resources["wood"].Rate = 0 // nothing makes wood now
	shrine := ge.Buildings.GetCost("shrine")["wood"]
	if _, err := ge.PlanAddBuild("shrine", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	setAmount(ge, "wood", shrine-1)
	ge.runPlanTick()
	if queued(ge, "hut") != 1 {
		t.Error("the hut waited behind a shrine nothing can fund")
	}
	if v := ge.planViews(); len(v) != 1 || v[0].Note != "too little wood coming in" {
		t.Errorf("shrine view = %+v", v)
	}
}

func TestPlan_ResearchQueueStartsInOrder(t *testing.T) {
	ge := planTestEngine(t)
	ge.age = "stone_age" // Pottery is a Stone Age tech, and it needs Fire Mastery
	// Planning Pottery plans Fire Mastery before it.
	if err := ge.PlanAddResearch("pottery"); err != nil {
		t.Fatalf("plan pottery: %v", err)
	}
	if len(ge.plan) != 2 || ge.plan[0].Key != "fire_mastery" || ge.plan[1].Key != "pottery" {
		t.Fatalf("plan = %+v, want Fire Mastery then Pottery", ge.plan)
	}
	setAmount(ge, "knowledge", 10000)
	ge.runPlanTick()
	if ge.Research.currentTech != "fire_mastery" {
		t.Fatalf("researching %q, want fire_mastery", ge.Research.currentTech)
	}
	if v := ge.planViews(); len(v) != 1 || v[0].Note != "research slot busy" {
		t.Fatalf("views = %+v, want pottery waiting on the slot", v)
	}
	for i := 0; i < 1000 && ge.Research.currentTech != ""; i++ {
		ge.processResearch()
	}
	ge.runPlanTick()
	if ge.Research.currentTech != "pottery" || len(ge.plan) != 0 {
		t.Errorf("after fire_mastery: researching %q, plan %+v", ge.Research.currentTech, ge.plan)
	}
}

func TestPlan_InvalidItemsDropOut(t *testing.T) {
	ge := planTestEngine(t)
	if _, err := ge.PlanAddBuild("hut", 2); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddResearch("fire_mastery"); err != nil {
		t.Fatal(err)
	}
	ge.Research.researched["tool_making"] = true
	ge.age = "stone_age" // the huts belong to the Primitive Age now
	ge.runPlanTick()
	if len(ge.plan) != 1 || ge.plan[0].Key != "fire_mastery" {
		t.Errorf("plan = %+v, want only fire_mastery", ge.plan)
	}
	if !logHas(ge, "Plan: dropped 2 Huts") || !logHas(ge, "Plan: dropped research Tool Making (already researched)") {
		t.Error("missing drop log lines")
	}
	// Removing a prerequisite from the plan does not drop what needed it:
	// that tech could still start, so it waits and says what it needs.
	ge2 := planTestEngine(t)
	ge2.age = "stone_age"
	_ = ge2.PlanAddResearch("fire_mastery")
	_ = ge2.PlanAddResearch("pottery")
	if _, err := ge2.PlanRemove(1); err != nil {
		t.Fatal(err)
	}
	setAmount(ge2, "knowledge", 10000)
	ge2.runPlanTick()
	if len(ge2.plan) != 1 || ge2.plan[0].Key != "pottery" || ge2.Research.currentTech != "" {
		t.Errorf("pottery should wait in the plan for Fire Mastery: plan %+v, researching %q", ge2.plan, ge2.Research.currentTech)
	}
	if v := ge2.planViews(); len(v) != 1 || v[0].Status != PlanStatusBlocked || v[0].Note != "needs Fire Mastery first" {
		t.Errorf("views = %+v, want pottery blocked on Fire Mastery", v)
	}
	if logHas(ge2, "Plan: dropped research Pottery") {
		t.Error("the plan dropped a tech that only waits for its prerequisite")
	}
}

func TestPlan_AddRespectsMaxCountAndMerges(t *testing.T) {
	ge := planTestEngine(t)
	// A building with a copy limit. Storage had one (50 Stashes) until the
	// storage wall; only wonders and monuments have one now, at 1, so the
	// test gives its own engine's Stash the old limit to count against.
	limited := ge.Buildings.defs["stash"]
	if limited.MaxCount != 0 {
		t.Fatalf("setup: the Stash has a copy limit of %d; storage has none since the storage wall", limited.MaxCount)
	}
	limited.MaxCount = 50
	ge.Buildings.defs["stash"] = limited
	ge.Buildings.counts["stash"] = 48
	n, err := ge.PlanAddBuild("stash", 5)
	if err != nil || n != 2 {
		t.Fatalf("PlanAddBuild(stash, 5) = %d, %v; want 2 (MaxCount 50)", n, err)
	}
	if _, err := ge.PlanAddBuild("stash", 1); err == nil {
		t.Error("planned a stash past MaxCount")
	}
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("hut", 4); err != nil {
		t.Fatal(err)
	}
	if len(ge.plan) != 2 || ge.plan[1].Count != 5 {
		t.Errorf("plan = %+v, want the huts merged into one item of 5", ge.plan)
	}
	if _, err := ge.PlanAddBuild("bronze_smithy_not_a_key", 1); err == nil {
		t.Error("planned an unknown building")
	}
}

func TestPlan_MoveRemoveClear(t *testing.T) {
	ge := planTestEngine(t)
	for _, k := range []string{"hut", "shrine", "wood_camp"} {
		if _, err := ge.PlanAddBuild(k, 1); err != nil {
			t.Fatal(err)
		}
	}
	if to, err := ge.PlanMove(3, -1); err != nil || to != 2 {
		t.Fatalf("PlanMove(3, -1) = %d, %v", to, err)
	}
	if to, _ := ge.PlanMove(1, -5); to != 1 {
		t.Errorf("moving past the top landed at %d", to)
	}
	keys := func() string {
		var s []string
		for _, it := range ge.plan {
			s = append(s, it.Key)
		}
		return strings.Join(s, ",")
	}
	if got := keys(); got != "hut,wood_camp,shrine" {
		t.Errorf("order = %s", got)
	}
	if _, err := ge.PlanRemove(4); err == nil {
		t.Error("removed item 4 of 3")
	}
	if _, err := ge.PlanRemove(1); err != nil || keys() != "wood_camp,shrine" {
		t.Errorf("after remove 1: %s, %v", keys(), err)
	}
	if ge.PlanClear() != 2 || len(ge.plan) != 0 {
		t.Error("clear left items behind")
	}
}

// Offline catch-up runs the plan as the time passes: with a 50-wood cap a
// lump sum could pay for two or three huts; stepping through the hour pays
// for every copy the hour's production covers.
func TestOffline_RunsThePlanAsResourcesComeIn(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["wood_camp"] = 6
	ge.recalculateRates()
	rate := ge.Resources.GetRate("wood")
	if rate <= 0 {
		t.Fatal("setup: no wood income")
	}
	if _, err := ge.PlanAddBuild("hut", 30); err != nil {
		t.Fatal(err)
	}
	ge.SimulateOffline(time.Hour)
	started := 0
	if len(ge.plan) > 0 {
		started = ge.plan[0].Started
	} else {
		started = 30
	}
	if started < 8 {
		t.Errorf("the plan started %d huts in an hour offline at %.2f wood/tick, want many more than one cap's worth", started, rate)
	}
	want := "While you were away, your plan started building " + BuildingCount(started, "hut") + "."
	if !logHasLine(ge, want) {
		t.Errorf("no offline plan summary %q", want)
	}
	if got := ge.Buildings.GetCount("hut"); got == 0 {
		t.Error("no hut finished construction offline")
	}
}

// With nothing to do offline, the stepped catch-up pays what the old lump
// sum did: rate x ticks x efficiency, capped.
func TestOffline_EmptyPlanPaysTheLumpSum(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["stash"] = 50
	ge.Buildings.counts["wood_camp"] = 2
	ge.recalculateRates()
	ge.wonderOverflowOff = true
	setAmount(ge, "wood", 0)
	rate := ge.Resources.GetRate("wood")
	store := ge.Resources.GetStorage("wood")
	tick0 := ge.tick
	ge.SimulateOffline(20 * time.Minute)
	ticks := ge.tick - tick0
	want := min(rate*float64(ticks)*OfflineEfficiency, store)
	if got := ge.Resources.Get("wood"); abs(got-want) > 1e-6*want {
		t.Errorf("wood after 20 min = %v, want %v (%d ticks at %v)", got, want, ticks, rate)
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// A full day offline with a long plan resolves quickly and deterministically.
func TestOffline_DayWithPlanIsFastAndDeterministic(t *testing.T) {
	run := func() (*GameEngine, time.Duration) {
		ge := newSeededEngine(7)
		ge.Buildings.counts["wood_camp"] = 10
		ge.Buildings.counts["stash"] = 20
		ge.recalculateRates()
		for _, k := range []string{"hut", "stash", "gathering_camp", "wood_camp", "shrine", "story_circle"} {
			if _, err := ge.PlanAddBuild(k, 25); err != nil {
				t.Fatal(err)
			}
		}
		_ = ge.PlanAddResearch("tool_making")
		_ = ge.PlanAddResearch("fire_mastery")
		start := time.Now()
		ge.SimulateOffline(MaxOfflineTime)
		return ge, time.Since(start)
	}
	a, took := run()
	b, _ := run()
	if took > 3*time.Second {
		t.Errorf("24h offline took %s", took)
	}
	t.Logf("24h offline with a plan: %s", took)
	sa, sb := a.buildSaveSnapshot(), b.buildSaveSnapshot()
	sa.Timestamp, sb.Timestamp = time.Time{}, time.Time{}
	sa.Stats.GameStarted, sb.Stats.GameStarted = time.Time{}, time.Time{} // wall clock
	va, vb := reflect.ValueOf(sa), reflect.ValueOf(sb)
	for i := 0; i < va.NumField(); i++ {
		if !reflect.DeepEqual(va.Field(i).Interface(), vb.Field(i).Interface()) {
			t.Errorf("two identical games came back from 24h offline with different %s", va.Type().Field(i).Name)
		}
	}
}

func TestPlanAndOverflow_SurviveSaveAndLoad(t *testing.T) {
	isolateAccountDir(t)
	ge := planTestEngine(t)
	if _, err := ge.PlanAddBuild("hut", 4); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddResearch("tool_making"); err != nil {
		t.Fatal(err)
	}
	ge.plan[0].Started = 3
	ge.SetWonderOverflow(false)
	if err := ge.SaveGame("plan-roundtrip"); err != nil {
		t.Fatal(err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("plan-roundtrip"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.plan, ge.plan) {
		t.Errorf("plan after load = %+v, want %+v", loaded.plan, ge.plan)
	}
	if loaded.WonderOverflow() {
		t.Error("overflow came back on")
	}
	if st := loaded.GetState(); len(st.Plan) != 2 || st.WonderOverflow {
		t.Errorf("state: plan %d items, overflow %v", len(st.Plan), st.WonderOverflow)
	}
}

func TestPlan_ClearedByPrestigeSuccumbAndReset(t *testing.T) {
	ge := planTestEngine(t)
	_, _ = ge.PlanAddBuild("hut", 2)
	ge.SetWonderOverflow(false)
	ge.Reset()
	if len(ge.plan) != 0 || !ge.WonderOverflow() {
		t.Errorf("after Reset: plan %+v, overflow %v", ge.plan, ge.WonderOverflow())
	}

	ge = planTestEngine(t)
	_, _ = ge.PlanAddBuild("hut", 2)
	ge.SetWonderOverflow(false)
	ge.mu.Lock()
	ge.completePrestige(prestigePlain)
	ge.mu.Unlock()
	if len(ge.plan) != 0 {
		t.Errorf("after prestige: plan %+v", ge.plan)
	}
	if ge.WonderOverflow() {
		t.Error("prestige reset the overflow preference")
	}

	ge = planTestEngine(t)
	_, _ = ge.PlanAddBuild("hut", 2)
	succumbIn(t, ge, "iron_age")
	if len(ge.plan) != 0 {
		t.Errorf("after Succumb: plan %+v", ge.plan)
	}
}

func TestOverflow_BanksWhatTheCapCutsOff(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["wood_camp"] = 5
	ge.recalculateRates()
	w := ge.progress.WonderForAge(ge.age)
	rate := ge.Resources.GetRate("wood")
	setAmount(ge, "wood", ge.Resources.GetStorage("wood"))
	ge.applyTickRates()
	if got := ge.Buildings.wonderBanks[w]["wood"]; abs(got-rate) > 1e-9 {
		t.Errorf("banked %v wood, want the tick's %v", got, rate)
	}
	if ge.Resources.Get("wood") != ge.Resources.GetStorage("wood") {
		t.Error("overflow took from the store itself")
	}
	ge.wonderOverflowOff = true
	ge.applyTickRates()
	if got := ge.Buildings.wonderBanks[w]["wood"]; abs(got-rate) > 1e-9 {
		t.Errorf("banked with overflow off: %v", got)
	}
	// Never past what the wonder needs, and one log line when a part fills.
	ge.wonderOverflowOff = false
	need := ge.Buildings.defs[w].BaseCost["wood"]
	ge.Buildings.wonderBanks[w]["wood"] = need - rate/2
	ge.applyTickRates()
	if got := ge.Buildings.wonderBanks[w]["wood"]; got != need {
		t.Errorf("bank = %v, want exactly %v", got, need)
	}
	if !logHas(ge, "Overflow finished banking wood") {
		t.Error("no log line when overflow filled the wood part")
	}
}

func TestOverflow_WorksOffline(t *testing.T) {
	ge := newSeededEngine(1)
	ge.Buildings.counts["wood_camp"] = 5
	ge.recalculateRates()
	w := ge.progress.WonderForAge(ge.age)
	ge.SimulateOffline(time.Hour)
	if ge.Buildings.wonderBanks[w]["wood"] <= 0 {
		t.Error("an hour offline at the cap banked nothing into the wonder")
	}
	if !logHas(ge, "Overflow banked into") {
		t.Error("no offline overflow summary")
	}
}
