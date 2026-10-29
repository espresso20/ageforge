package game

import (
	"strings"
	"testing"
	"time"
)

// Each resource an item's next start takes gets a row with the need, the
// stock, what the items above hold back, the income and the time until the
// rest is free; the item starts when its slowest row is ready.
func TestPlanDetail_RowsAndStartsIn(t *testing.T) {
	ge := planTestEngine(t)
	setAmount(ge, "wood", 0)
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("wood_camp", 1); err != nil {
		t.Fatal(err)
	}
	ge.mu.RLock()
	views := ge.planViews()
	iv := ge.tickIntervalLocked()
	ge.mu.RUnlock()
	hut, camp := views[0], views[1]
	if len(hut.Rows) == 0 || hut.Rows[0].Res != "wood" {
		t.Fatalf("hut rows = %+v, want a wood row", hut.Rows)
	}
	row := hut.Rows[0]
	if row.Need != hut.Cost["wood"] || row.Have != 0 || row.Held != 0 || row.Rate <= 0 {
		t.Errorf("hut wood row = %+v, want need %v, have 0, held 0, income > 0", row, hut.Cost["wood"])
	}
	want := time.Duration(row.Need / row.Rate * float64(iv))
	if row.ReadyIn != want || hut.StartsIn != want || hut.When != PlanWhenTimed || hut.Bottleneck != "wood" {
		t.Errorf("hut ready in %v, starts in %v (when %q, bottleneck %q), want %v from wood", row.ReadyIn, hut.StartsIn, hut.When, hut.Bottleneck, want)
	}
	if hut.HeldBack["wood"] != hut.Cost["wood"] {
		t.Errorf("hut holds back %v, want its wood price %v", hut.HeldBack, hut.Cost["wood"])
	}
	// The camp below sees the hut's reservation as held, and waits for both.
	var campWood *PlanResourceRow
	for i := range camp.Rows {
		if camp.Rows[i].Res == "wood" {
			campWood = &camp.Rows[i]
		}
	}
	if campWood == nil {
		t.Skip("the wood camp does not cost wood here")
	}
	if campWood.Held != hut.Cost["wood"] || camp.StartsIn <= hut.StartsIn {
		t.Errorf("camp wood row = %+v starts in %v, want held %v and later than the hut (%v)", *campWood, camp.StartsIn, hut.Cost["wood"], hut.StartsIn)
	}
}

// A resource with no income that the item is short of is "never", not a time.
func TestPlanDetail_NoIncomeIsNever(t *testing.T) {
	ge := planTestEngine(t)
	ge.mu.Lock()
	defer ge.mu.Unlock()
	var v PlanItemView
	ge.planDetail(&v, map[string]float64{"no_such_thing": 5, "wood": 0}, map[string]float64{})
	if v.When != PlanWhenNever || v.Bottleneck != "no_such_thing" || len(v.Rows) != 1 || !v.Rows[0].Never {
		t.Errorf("detail = %+v, want never, held up by no_such_thing", v)
	}
}

// The estimate is the same every time for the same town, and its Before is
// the figure plan optimize reports as its own.
func TestPlanEstimate_MatchesOptimizeBefore(t *testing.T) {
	ge := planTestEngine(t)
	setAmount(ge, "wood", 0)
	for _, add := range []struct {
		key string
		n   int
	}{{"hut", 4}, {"wood_camp", 2}, {"hut", 3}} {
		if _, err := ge.PlanAddBuild(add.key, add.n); err != nil {
			t.Fatal(err)
		}
	}
	est := ge.PlanEstimate()
	if est.Before <= 0 || est.Finish < est.Before {
		t.Fatalf("estimate = %+v, want a positive Before and a Finish after it", est)
	}
	ge.mu.Lock()
	again := ge.planEstimateLocked()
	ge.mu.Unlock()
	if again.Before != est.Before || again.Finish != est.Finish || len(again.Items) != len(est.Items) {
		t.Errorf("estimate changed between runs: %+v then %+v", est, again)
	}
	prev, err := ge.PlanOptimizePreview()
	if err != nil {
		t.Fatal(err)
	}
	if prev.Before != est.Before {
		t.Errorf("optimize Before = %v, estimate Before = %v, want the same", prev.Before, est.Before)
	}
	for i := 1; i < len(est.Items); i++ {
		if est.Items[i].Modeled && est.Items[i-1].Modeled && est.Items[i].Start < est.Items[i-1].Start {
			t.Errorf("item %d starts at %v before item %d at %v", i+1, est.Items[i].Start, i, est.Items[i-1].Start)
		}
	}
	// Asking does not touch the plan.
	if len(ge.plan) != 3 {
		t.Errorf("plan has %d items after asking, want 3", len(ge.plan))
	}
}

// A preview shows the reorder without making it; PlanOptimize makes it.
func TestPlanOptimizePreview_DoesNotApply(t *testing.T) {
	ge := planTestEngine(t)
	setAmount(ge, "wood", 0)
	for _, key := range []string{"hut", "wood_camp", "hut", "wood_camp"} {
		if _, err := ge.PlanAddBuild(key, 3); err != nil {
			t.Fatal(err)
		}
	}
	before := append([]PlanItem(nil), ge.plan...)
	prev, err := ge.PlanOptimizePreview()
	if err != nil {
		t.Fatal(err)
	}
	if !planItemsEqual(before, ge.plan) {
		t.Fatalf("preview changed the plan: %+v", ge.plan)
	}
	if len(prev.Previous) != len(before) {
		t.Errorf("Previous has %d labels, want %d", len(prev.Previous), len(before))
	}
	res, err := ge.PlanOptimize()
	if err != nil {
		t.Fatal(err)
	}
	if res.Changed != prev.Changed || strings.Join(res.Order, "|") != strings.Join(prev.Order, "|") {
		t.Errorf("preview said %v %v, optimize did %v %v", prev.Changed, prev.Order, res.Changed, res.Order)
	}
	if res.Changed && len(prev.Moved) != len(prev.Order) {
		t.Errorf("Moved has %d marks for %d items", len(prev.Moved), len(prev.Order))
	}
}

// The advance item lists the charter's pillars with what each lacks.
func TestPlanAdvance_ListsPillars(t *testing.T) {
	ge := planTestEngine(t)
	if err := ge.PlanAddAdvance(); err != nil {
		t.Fatal(err)
	}
	st := ge.GetState()
	adv := st.Plan[0]
	if len(adv.Pillars) != 4 {
		t.Fatalf("advance pillars = %+v, want the four of the charter", adv.Pillars)
	}
	view := st.AdvanceView()
	if !strings.Contains(adv.Note, "pillars") && view.Blockers == nil {
		t.Errorf("advance note = %q, want the pillar count", adv.Note)
	}
	unmet := ""
	for i, p := range adv.Pillars {
		if p.Name != view.Pillars[i].Name || p.Done != view.Pillars[i].Met {
			t.Errorf("pillar %d = %+v, charter says %+v", i, p, view.Pillars[i])
		}
		if !p.Done && p.Blocker == "" {
			t.Errorf("pillar %q is unmet but names no blocker", p.Name)
		}
		if !p.Done && unmet == "" && p.BlockerKey != "" {
			unmet = p.BlockerKey
		}
	}
	// Plan the blocker above the advance: the pillar then names that item.
	if unmet == "" {
		t.Skip("no pillar blocker with a key in this age")
	}
	if _, ok := ge.rules.Building(unmet); !ok {
		t.Skipf("blocker %q is not a building", unmet)
	}
	ge.mu.Lock()
	ge.plan = append([]PlanItem{{Kind: PlanBuild, Key: unmet, Count: 1}}, ge.plan...)
	ge.mu.Unlock()
	st = ge.GetState()
	linked := false
	for _, p := range st.Plan[1].Pillars {
		linked = linked || p.BlockerKey == unmet && p.Item == 1
	}
	if !linked {
		t.Errorf("no pillar links %q to plan item 1: %+v", unmet, st.Plan[1].Pillars)
	}
}

// A wonder item says it needs a deposit by hand and shows its bank.
func TestPlanWonder_NeedsADeposit(t *testing.T) {
	ge := planTestEngine(t)
	w := ge.progress.WonderForAge(ge.age)
	if _, err := ge.PlanAddBuild(w, 1); err != nil {
		t.Fatal(err)
	}
	v := ge.GetState().Plan[0]
	if v.When != PlanWhenHand || v.Note == "" || v.BankShare < 0 || v.BankShare >= 1 {
		t.Errorf("wonder view = %+v, want a hand deposit with the bank share below 1", v)
	}
	if len(v.Rows) == 0 {
		t.Error("wonder has no rows for what its bank lacks")
	}
}

// expeditionFixture is a Stone Age town with a short offer on the board, a
// party at home, and income to price it from. It returns the offer's title.
func expeditionFixture(t *testing.T) (*GameEngine, string) {
	t.Helper()
	ge := planTestEngine(t)
	setAge(ge, "stone_age")
	n := plantOffer(ge, TierShort, "")
	setSoldiers(ge, 100)
	for _, res := range []string{"stone", "wood"} {
		setResource(ge, res, 0)
		setIncome(ge, res, 0.5)
	}
	ge.mu.RLock()
	title := ge.Military.board.Slots[n-1].Offer.Title
	ge.mu.RUnlock()
	return ge, title
}

// An expedition item waits for its price (holding it back), then launches
// the mission through the command's path.
func TestPlanExpedition_WaitsThenLaunches(t *testing.T) {
	ge, title := expeditionFixture(t)
	if err := ge.PlanAddExpedition(title); err != nil {
		t.Fatal(err)
	}
	if err := ge.PlanAddExpedition(title); err == nil {
		t.Error("the same expedition was planned twice")
	}
	v := ge.GetState().Plan[0]
	if v.Kind != PlanExpedition || v.Status != PlanStatusWaiting || len(v.Cost) == 0 || v.HeldBack == nil {
		t.Fatalf("view = %+v, want a waiting expedition that holds its price", v)
	}
	ge.runPlanTick()
	if ge.Military.HasActive() || len(ge.plan) != 1 {
		t.Fatal("the expedition left without its price")
	}
	ge.mu.Lock()
	n := ge.planOfferNumber(title)
	cost := ge.missionCostLocked(*ge.Military.board.Slots[n-1].Offer)
	ge.mu.Unlock()
	for res, c := range cost {
		if res != "soldiers" {
			setAmount(ge, res, c)
		}
	}
	ge.runPlanTick()
	if !ge.Military.HasActive() || len(ge.plan) != 0 {
		t.Fatalf("active %v, plan %d items, want the expedition sent and the item used up", ge.Military.HasActive(), len(ge.plan))
	}
	if !logHas(ge, "Plan: sent "+title) {
		t.Errorf("no summary line naming %q", title)
	}
}

// With no party at home the item is blocked and holds nothing back; it goes
// once the soldiers are there.
func TestPlanExpedition_BlockedWithoutAParty(t *testing.T) {
	ge, title := expeditionFixture(t)
	if err := ge.PlanAddExpedition(title); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	setSoldiers(ge, 0)
	v := ge.GetState().Plan[0]
	if v.Status != PlanStatusBlocked || !strings.Contains(v.Note, "soldiers") || v.HeldBack != nil || v.When != PlanWhenUnknown {
		t.Errorf("view = %+v, want blocked on the party, holding nothing", v)
	}
	if got := ge.GetState().Plan[1]; got.Rows[0].Held != 0 {
		t.Errorf("hut sees %v held by a blocked expedition, want none", got.Rows[0].Held)
	}
	for _, res := range []string{"stone", "wood"} {
		setAmount(ge, res, 100000)
	}
	ge.runPlanTick()
	if ge.Military.HasActive() {
		t.Fatal("it left with nobody to send")
	}
	setSoldiers(ge, 100)
	ge.runPlanTick()
	if !ge.Military.HasActive() {
		t.Fatal("it stayed once the party was home")
	}
}

// An item whose offer is off the board drops out with a log line.
func TestPlanExpedition_DropsWhenTheOfferIsGone(t *testing.T) {
	ge, title := expeditionFixture(t)
	if err := ge.PlanAddExpedition(title); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	for i := range ge.Military.board.Slots {
		if o := ge.Military.board.Slots[i].Offer; o != nil && o.Title == title {
			ge.Military.board.Slots[i].Offer = nil
		}
	}
	ge.mu.Unlock()
	ge.runPlanTick()
	if len(ge.plan) != 0 || !logHas(ge, "Plan: dropped expedition "+title) {
		t.Errorf("plan %+v, want the item dropped with a log line", ge.plan)
	}
}

// Expedition items survive a save and load; an empty title does not.
func TestPlanExpedition_SurvivesSaveAndLoad(t *testing.T) {
	ge, title := expeditionFixture(t)
	if err := ge.PlanAddExpedition(title); err != nil {
		t.Fatal(err)
	}
	got := loadPlanIn(ge.rules, append(append([]PlanItem(nil), ge.plan...), PlanItem{Kind: PlanExpedition, Count: 1}))
	if len(got) != 1 || got[0].Kind != PlanExpedition || got[0].Key != title || got[0].Count != 1 {
		t.Errorf("loaded %+v, want the one expedition item", got)
	}
}

// Offline catch-up sends the same mission live play would.
func TestPlanExpedition_LaunchesWhileAway(t *testing.T) {
	ge, title := expeditionFixture(t)
	if err := ge.PlanAddExpedition(title); err != nil {
		t.Fatal(err)
	}
	var starts planStarts
	ge.mu.Lock()
	n := ge.planOfferNumber(title)
	cost := ge.missionCostLocked(*ge.Military.board.Slots[n-1].Offer)
	for res, c := range cost {
		if res != "soldiers" {
			ge.Resources.resources[res].Amount = c
		}
	}
	started := ge.runPlan(&starts)
	ge.mu.Unlock()
	if !started || len(starts.expeditions) != 1 || !ge.Military.HasActive() {
		t.Errorf("started %v with %+v, want the expedition sent", started, starts)
	}
}
