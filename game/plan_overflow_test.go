package game

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Overflow pays the plan (overflow.go, bankPlanOverflow): what a cap cuts off,
// after the wonder, goes into the banks of the plan's build items in plan
// order, each up to its next copy's price, and the copy pays from its bank
// first.

// overflowEngine is planTestEngine with wonder overflow off, so the plan gets
// everything a cap cuts off, and wood storage set to cap.
func overflowEngine(t *testing.T, cap float64) *GameEngine {
	t.Helper()
	ge := planTestEngine(t)
	ge.wonderOverflowOff = true
	ge.Resources.resources["wood"].Storage = cap
	setAmount(ge, "wood", cap)
	return ge
}

func woodLoss(v float64) []overflowLoss { return []overflowLoss{{res: "wood", amount: v}} }

func TestPlanOverflow_PaysQueuedItemsInOrder(t *testing.T) {
	ge := overflowEngine(t, 1000)
	if _, err := ge.PlanAddBuild("hut", 2); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("wood_camp", 1); err != nil {
		t.Fatal(err)
	}
	hut := ge.Buildings.GetCost("hut")["wood"]
	camp := ge.Buildings.GetCost("wood_camp")["wood"]
	if hut <= 0 || camp <= 0 {
		t.Fatalf("setup: hut %v wood, wood camp %v wood", hut, camp)
	}
	got := map[string]float64{}

	// The first item takes it first, up to its next copy's price.
	ge.bankPlanOverflow(woodLoss(hut/2), got)
	if b := ge.plan[0].Banked["wood"]; abs(b-hut/2) > 1e-9 || ge.plan[1].Banked != nil {
		t.Fatalf("after half a hut: banks %v and %v, want %v and nothing", ge.plan[0].Banked, ge.plan[1].Banked, hut/2)
	}
	// Then the next item gets what the first doesn't need.
	ge.bankPlanOverflow(woodLoss(hut), got)
	if b := ge.plan[0].Banked["wood"]; abs(b-hut) > 1e-9 {
		t.Errorf("hut bank = %v, want its next copy's %v and no more", b, hut)
	}
	if b := ge.plan[1].Banked["wood"]; abs(b-hut/2) > 1e-9 {
		t.Errorf("wood camp bank = %v, want the %v the huts left", b, hut/2)
	}
	// Past every item's next copy, the rest is lost as before.
	ge.bankPlanOverflow(woodLoss(10*camp), got)
	if b := ge.plan[1].Banked["wood"]; abs(b-camp) > 1e-9 {
		t.Errorf("wood camp bank = %v, want its price %v", b, camp)
	}
	if want := hut + camp; abs(got["wood"]-want) > 1e-9 {
		t.Errorf("banked %v wood in all, want %v", got["wood"], want)
	}
	if ge.Resources.Get("wood") != 1000 {
		t.Errorf("the store changed to %v: overflow only takes what the cap cut off", ge.Resources.Get("wood"))
	}

	// With nothing in the stores, the banks pay: one hut and the camp start,
	// and the store stays empty.
	setAmount(ge, "wood", 0)
	ge.runPlanTick()
	if queued(ge, "hut") != 1 || queued(ge, "wood_camp") != 1 {
		t.Fatalf("queued %d huts and %d wood camps, want 1 and 1 paid from their banks", queued(ge, "hut"), queued(ge, "wood_camp"))
	}
	if w := ge.Resources.Get("wood"); w != 0 {
		t.Errorf("wood = %v, want 0 (the banks paid)", w)
	}
	if len(ge.plan) != 1 || ge.plan[0].Key != "hut" || ge.plan[0].Count != 1 || ge.plan[0].Banked != nil {
		t.Errorf("plan = %+v, want one hut left with an empty bank", ge.plan)
	}
	if !ge.buildQueue[0].FromPlan {
		t.Error("a copy the bank paid isn't marked as the plan's (it would go unstaffed)")
	}
}

// A banked item holds back only what its bank doesn't cover, so the items
// below it get the rest.
func TestPlanOverflow_BankedItemReservesOnlyTheRest(t *testing.T) {
	ge := overflowEngine(t, 1e5)
	ge.advanceAge("stone_age")
	ge.Resources.resources["stone"].Storage = 1e5
	setAmount(ge, "stone", 0) // the longhouse waits on stone, which is coming in
	ge.Resources.SetRate("stone", 1)
	if _, err := ge.PlanAddBuild("longhouse", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("stone_camp", 1); err != nil {
		t.Fatal(err)
	}
	lh := ge.Buildings.GetCost("longhouse")
	camp := ge.Buildings.GetCost("stone_camp")["wood"]
	if lh["stone"] <= 0 || lh["wood"] <= 1 || camp <= 0 {
		t.Fatalf("setup: longhouse %v, stone camp %v wood", lh, camp)
	}
	// Its bank covers all but 1 of its wood, so it holds back 1 wood, and
	// the camp below starts from the rest. Held back in full (110 wood), the
	// camp would wait.
	ge.plan[0].Banked = map[string]float64{"wood": lh["wood"] - 1}
	setAmount(ge, "wood", camp+1)
	ge.runPlanTick()
	if queued(ge, "stone_camp") != 1 || queued(ge, "longhouse") != 0 {
		t.Fatalf("queued %d stone camps and %d longhouses, want the camp only (plan %+v)", queued(ge, "stone_camp"), queued(ge, "longhouse"), ge.plan)
	}
	v := ge.planViews()
	if len(v) != 1 || v[0].Key != "longhouse" || v[0].Banked["wood"] != lh["wood"]-1 || v[0].Cost["wood"] != lh["wood"] {
		t.Errorf("longhouse view = %+v, want its whole price and its bank", v)
	}
	if v[0].Status != PlanStatusWaiting || v[0].Short != "stone" {
		t.Errorf("longhouse view = %+v, want waiting on stone", v[0])
	}
}

// Nothing is ever built that wasn't queued: a plan of three copies under
// constant overflow builds exactly three, then banks nothing more.
func TestPlanOverflow_BuildsNothingUnqueued(t *testing.T) {
	ge := overflowEngine(t, 500)
	ge.Resources.SetRate("wood", 40)
	before := ge.Buildings.GetCount("hut")
	if _, err := ge.PlanAddBuild("hut", 3); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 500; i++ {
		ge.applyTickRates()
		ge.runPlanTick()
	}
	if got := ge.Buildings.GetCount("hut") + queued(ge, "hut") - before; got != 3 {
		t.Errorf("built or queued %d huts from a plan of 3", got)
	}
	if len(ge.plan) != 0 {
		t.Errorf("plan = %+v, want empty", ge.plan)
	}
	got := map[string]float64{}
	ge.bankPlanOverflow(woodLoss(1e6), got)
	if len(got) != 0 {
		t.Errorf("an empty plan banked %v", got)
	}
}

// A copy priced over the cap, which nothing could ever pay before, starts
// once overflow has banked the part over the cap; until then it is blocked
// and holds nothing back.
func TestPlanOverflow_OverCapItemBecomesBuyable(t *testing.T) {
	ge := newSeededEngine(1) // base storage: 50 wood
	ge.wonderOverflowOff = true
	price := ge.Buildings.GetCost("shrine")
	cap := ge.Resources.GetStorage("wood")
	if price["wood"] <= cap {
		t.Skipf("setup: the shrine (%v wood) fits the %v cap", price["wood"], cap)
	}
	for res := range price {
		if res != "wood" {
			ge.Resources.resources[res].Storage = 1e6
			setAmount(ge, res, price[res])
		}
	}
	if _, err := ge.PlanAddBuild("shrine", 1); err != nil {
		t.Fatal(err)
	}
	setAmount(ge, "wood", cap)
	for k := range ge.Resources.resources {
		ge.Resources.SetRate(k, 0)
	}
	ge.Resources.SetRate("wood", 3)
	v := ge.planViews()[0]
	if v.Status != PlanStatusBlocked || !strings.Contains(v.Note, "wood storage") {
		t.Fatalf("view = %+v, want blocked on wood storage", v)
	}
	ticks := 0
	for ; ticks < 1000 && queued(ge, "shrine") == 0; ticks++ {
		ge.applyTickRates()
		if ticks == 2 {
			if v := ge.planViews()[0]; v.Status != PlanStatusBlocked || v.Banked["wood"] <= 0 {
				t.Errorf("view while banking = %+v, want blocked with wood banked", v)
			}
		}
		ge.runPlanTick()
	}
	if queued(ge, "shrine") != 1 {
		t.Fatalf("the shrine never started (plan %+v)", ge.plan)
	}
	// It needed (price - cap) of overflow at 3 a tick.
	if want := int((price["wood"]-cap)/3) + 1; ticks > want+1 {
		t.Errorf("took %d ticks, want about %d", ticks, want)
	}
	// The full store paid the rest: what is left is under the last tick's
	// overflow, by which the bank may have passed the part over the cap.
	if w := ge.Resources.Get("wood"); w < 0 || w >= 3 {
		t.Errorf("wood left = %v: the full store should have paid all but the bank", w)
	}
	if len(ge.plan) != 0 {
		t.Errorf("plan = %+v, want empty", ge.plan)
	}
}

// The wonder keeps first call on overflow; the plan gets what it doesn't need.
func TestPlanOverflow_WonderFirst(t *testing.T) {
	ge := planTestEngine(t)
	w := ge.progress.WonderForAge(ge.age)
	need := ge.Buildings.defs[w].BaseCost["wood"]
	if need <= 0 {
		t.Skipf("setup: %s needs no wood", w)
	}
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	hut := ge.Buildings.GetCost("hut")["wood"]
	for k := range ge.Resources.resources {
		ge.Resources.SetRate(k, 0)
	}
	rate := hut / 2
	ge.Resources.SetRate("wood", rate)
	ge.Resources.resources["wood"].Storage = 1e6
	// Production only (no plan walk), so the banks show where the tick's
	// overflow went.
	setAmount(ge, "wood", 1e6)
	ge.applyTickRates()
	if got := ge.Buildings.wonderBanks[w]["wood"]; abs(got-rate) > 1e-9 {
		t.Errorf("wonder banked %v, want the whole %v", got, rate)
	}
	if ge.plan[0].Banked != nil {
		t.Errorf("the plan banked %v while the wonder still needed wood", ge.plan[0].Banked)
	}
	// The wonder short of a quarter of the tick: it takes that, the plan the rest.
	ge.Buildings.wonderBanks[w]["wood"] = need - rate/4
	ge.applyTickRates()
	if got := ge.Buildings.wonderBanks[w]["wood"]; got != need {
		t.Errorf("wonder bank = %v, want exactly %v", got, need)
	}
	if got := ge.plan[0].Banked["wood"]; abs(got-3*rate/4) > 1e-9 {
		t.Errorf("hut bank = %v, want the %v the wonder left", got, 3*rate/4)
	}
	// With wonder overflow off, the plan still banks.
	ge2 := planTestEngine(t)
	ge2.wonderOverflowOff = true
	if _, err := ge2.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	for k := range ge2.Resources.resources {
		ge2.Resources.SetRate(k, 0)
	}
	ge2.Resources.SetRate("wood", rate)
	setAmount(ge2, "wood", ge2.Resources.GetStorage("wood"))
	ge2.applyTickRates()
	if ge2.Buildings.wonderBanks[w]["wood"] != 0 || abs(ge2.plan[0].Banked["wood"]-rate) > 1e-9 {
		t.Errorf("overflow off: wonder %v, hut bank %v; want 0 and %v", ge2.Buildings.wonderBanks[w]["wood"], ge2.plan[0].Banked, rate)
	}
}

// Only builds that could start in this age bank: not the next age's
// buildings (the bank would dodge the advance's trim), not one at its
// MaxCount. A queued wonder of this age banks into its own bank, not an
// item bank (TestPlanOverflow_QueuedWonderBanks).
func TestPlanOverflow_OnlyThisAgesBuildsBank(t *testing.T) {
	ge := overflowEngine(t, 1000)
	next := ge.progress.GetNextAge(ge.age)
	nextKey := ""
	for _, k := range sortedKeys(ge.Buildings.defs) {
		d := ge.Buildings.defs[k]
		if d.RequiredAge == next && d.Category != "wonder" && d.BaseCost["wood"] > 0 {
			nextKey = k
			break
		}
	}
	if nextKey == "" {
		t.Fatal("setup: no next-age building costs wood")
	}
	w := ge.progress.WonderForAge(ge.age)
	if _, err := ge.PlanAddBuild(nextKey, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild(w, 1); err != nil {
		t.Fatal(err)
	}
	// Stash at its MaxCount (planTestEngine has 50).
	ge.plan = append(ge.plan, PlanItem{Kind: PlanBuild, Key: "stash", Count: 1})
	// The next age's wonder waits for the advance like its other buildings.
	nextWonder := ge.progress.WonderForAge(next)
	ge.plan = append(ge.plan, PlanItem{Kind: PlanBuild, Key: nextWonder, Count: 1})
	got := map[string]float64{}
	ge.bankPlanOverflow(woodLoss(1e6), got)
	for _, it := range ge.plan {
		if it.Banked != nil {
			t.Errorf("%s banked %v", it.Key, it.Banked)
		}
	}
	need := ge.Buildings.defs[w].BaseCost["wood"]
	if len(got) != 1 || got["wood"] != need || ge.Buildings.wonderBanks[w]["wood"] != need {
		t.Errorf("banked %v (the age's wonder holds %v); want only the %v wood the queued wonder lacks, in its own bank", got, ge.Buildings.wonderBanks[w], need)
	}
	if len(ge.Buildings.wonderBanks[nextWonder]) != 0 {
		t.Errorf("the next age's wonder banked %v before the advance", ge.Buildings.wonderBanks[nextWonder])
	}
}

// A wonder queued in the plan takes the plan's overflow into its own bank,
// in plan order, with wonder overflow off: a player who checks in a few
// times a day no longer has to be there for the bank to fill. A part of the
// price bigger than a full store is banked this way, and once the stores
// cover what is left the plan pays it and starts the wonder.
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
	// Not queued: with wonder overflow off nothing reaches the wonder.
	woodCap := need["wood"] / 4 // the wood part is four full stores
	ge.Resources.resources["wood"].Storage = woodCap
	setAmount(ge, "wood", woodCap)
	rate := need["wood"] / 10
	ge.Resources.SetRate("wood", rate)
	ge.applyTickRates()
	if len(ge.Buildings.wonderBanks[w]) != 0 {
		t.Fatalf("wonder overflow is off and the wonder is not planned, but its bank holds %v", ge.Buildings.wonderBanks[w])
	}

	// Queued behind a hut: the hut's next copy takes the overflow first, the
	// wonder what is left, tick by tick.
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild(w, 1); err != nil {
		t.Fatal(err)
	}
	hut := ge.Buildings.GetCost("hut")["wood"]
	if hut >= rate {
		t.Fatalf("setup: a hut costs %v wood, the tick overflows %v", hut, rate)
	}
	ge.applyTickRates()
	if got := ge.plan[0].Banked["wood"]; abs(got-hut) > 1e-9 {
		t.Errorf("the hut above the wonder banked %v, want its price %v first", got, hut)
	}
	if got := ge.Buildings.wonderBanks[w]["wood"]; abs(got-(rate-hut)) > 1e-9 {
		t.Errorf("the queued wonder banked %v, want the %v the hut left", got, rate-hut)
	}
	if ge.plan[1].Banked != nil {
		t.Errorf("the wonder item holds an item bank %v: a wonder banks into its own", ge.plan[1].Banked)
	}
	if got := ge.Resources.Get("wood"); got != woodCap {
		t.Errorf("the store changed to %v: overflow only takes what the cap cut off", got)
	}
	st := ge.GetState()
	if v := st.Plan[1]; v.Status != PlanStatusBlocked || v.Note != "bank not full" {
		t.Errorf("the wonder is %s (%s) while four stores of wood are still to bank, want blocked (bank not full)", v.Status, v.Note)
	}

	// The plan walk runs with the ticks: the hut starts from its bank, and
	// the overflow goes on into the wonder until the stores cover what its
	// bank lacks. Then the plan pays the rest and starts the wonder. Nobody
	// deposited anything.
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
		t.Errorf("the hut above the wonder did not start from its bank")
	}
	if !logHas(ge, "Overflow finished banking") && ge.Buildings.wonderBanks[w]["wood"] < need["wood"]-woodCap-1e-6 {
		t.Errorf("the wonder started with %v wood banked of %v and a store of %v", ge.Buildings.wonderBanks[w]["wood"], need["wood"], woodCap)
	}

	// With wonder overflow on, the age's wonder has first call wherever it
	// sits in the plan (TestPlanOverflow_WonderFirst): nothing changes.
	on := planTestEngine(t)
	w = on.progress.WonderForAge(on.age)
	if _, err := on.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := on.PlanAddBuild(w, 1); err != nil {
		t.Fatal(err)
	}
	for k := range on.Resources.resources {
		on.Resources.SetRate(k, 0)
	}
	on.Resources.SetRate("wood", rate)
	setAmount(on, "wood", on.Resources.GetStorage("wood"))
	on.applyTickRates()
	if got := on.Buildings.wonderBanks[w]["wood"]; abs(got-rate) > 1e-9 || on.plan[0].Banked != nil {
		t.Errorf("wonder overflow on: the wonder banked %v and the hut %v, want the whole %v in the wonder", got, on.plan[0].Banked, rate)
	}
}

// Offline catch-up fills a queued wonder's bank the same way, starts the
// wonder when the stores cover the rest, and says what was banked.
func TestPlanOverflow_QueuedWonderBanksOffline(t *testing.T) {
	ge := planTestEngine(t)
	ge.wonderOverflowOff = true
	ge.Buildings.counts["wood_camp"] = 20
	ge.Buildings.counts["gathering_camp"] = 20
	ge.Buildings.counts["stash"] = 0
	ge.recalculateRates()
	w := ge.progress.WonderForAge(ge.age)
	need := ge.Buildings.defs[w].BaseCost
	woodCap := ge.Resources.GetStorage("wood")
	if need["wood"] <= woodCap {
		t.Skipf("setup: the wonder's %v wood fits the %v cap", need["wood"], woodCap)
	}
	if ge.Resources.GetRate("wood") <= 0 || ge.Resources.GetRate("food") <= 0 {
		t.Fatalf("setup: wood %v/tick, food %v/tick", ge.Resources.GetRate("wood"), ge.Resources.GetRate("food"))
	}
	ge.plan = []PlanItem{{Kind: PlanBuild, Key: w, Count: 1}}
	setAmount(ge, "wood", woodCap)
	ge.SimulateOffline(8 * time.Hour)
	if queued(ge, w) == 0 && ge.Buildings.GetCount(w) == 0 {
		t.Fatalf("eight hours away with the wonder planned: bank %v of %v, not started (plan %+v)", ge.Buildings.wonderBanks[w], need, ge.plan)
	}
	if !logHas(ge, "Overflow banked toward your plan") {
		t.Error("no offline summary of what the plan banked")
	}

	// Not planned, with wonder overflow off: nothing is banked, as before.
	idle := planTestEngine(t)
	idle.wonderOverflowOff = true
	idle.Buildings.counts["wood_camp"] = 20
	idle.Buildings.counts["stash"] = 0
	idle.recalculateRates()
	setAmount(idle, "wood", idle.Resources.GetStorage("wood"))
	idle.SimulateOffline(8 * time.Hour)
	if len(idle.Buildings.wonderBanks[w]) != 0 {
		t.Errorf("wonder overflow off and nothing planned, but the bank holds %v", idle.Buildings.wonderBanks[w])
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

// Offline catch-up banks into the plan as the time passes and totals it in
// the welcome back.
func TestPlanOverflow_WorksOffline(t *testing.T) {
	ge := planTestEngine(t)
	ge.wonderOverflowOff = true
	ge.Buildings.counts["wood_camp"] = 5
	ge.recalculateRates()
	// A copy over the cap: it banks the whole time and can't start.
	ge.plan = []PlanItem{{Kind: PlanBuild, Key: "shrine", Count: 1}}
	price := ge.Buildings.GetCost("shrine")["wood"]
	ge.Buildings.counts["stash"] = 0
	ge.recalculateRates()
	if price <= ge.Resources.GetStorage("wood") {
		t.Skipf("setup: shrine %v fits the %v cap", price, ge.Resources.GetStorage("wood"))
	}
	setAmount(ge, "wood", ge.Resources.GetStorage("wood"))
	ge.SimulateOffline(10 * time.Minute)
	if len(ge.plan) > 0 && ge.plan[0].Banked["wood"] <= 0 && queued(ge, "shrine") == 0 {
		t.Errorf("ten minutes offline at the cap banked nothing into the plan (%+v)", ge.plan)
	}
	if !logHas(ge, "Overflow banked toward your plan") {
		t.Error("no offline summary of what the plan banked")
	}
}

// Banks are saved: a round trip keeps them, an item without one writes no
// field (so a save from before the bank keeps its bytes and its signature).
func TestPlanOverflow_BanksSurviveSaveAndLoad(t *testing.T) {
	isolateAccountDir(t)
	ge := overflowEngine(t, 1000)
	if _, err := ge.PlanAddBuild("shrine", 2); err != nil {
		t.Fatal(err)
	}
	if _, err := ge.PlanAddBuild("hut", 1); err != nil {
		t.Fatal(err)
	}
	ge.plan[0].Banked = map[string]float64{"wood": 123.5}
	if err := ge.SaveGame("plan-bank"); err != nil {
		t.Fatal(err)
	}
	loaded := NewGameEngine()
	if err := loaded.LoadGame("plan-bank"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.plan, ge.plan) {
		t.Errorf("plan after load = %+v, want %+v", loaded.plan, ge.plan)
	}
	if loaded.cheaterBadge {
		t.Error("a save with a bank failed its signature")
	}
	b, err := json.Marshal(PlanItem{Kind: PlanBuild, Key: "hut", Count: 1})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "banked") {
		t.Errorf("an item with no bank writes %s", b)
	}
	// Snapshots share no bank with the live plan.
	snap := clonePlan(ge.plan)
	snap[0].Banked["wood"] = 1
	if ge.plan[0].Banked["wood"] != 123.5 {
		t.Error("a cloned plan shares its bank with the live one")
	}
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

// BenchmarkBankPlanOverflow is the most bankPlanOverflow adds to a tick:
// a full plan of 60 build items this age can start, every store overflowing,
// and every bank already holding its next copy's price, so each item prices
// its next copy and finds nothing to take, against the 250 microsecond tick
// budget (smoke.TickBudget).
func BenchmarkBankPlanOverflow(b *testing.B) {
	ge, losses := fullPlanBanks(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range losses {
			losses[j].amount = 1
		}
		ge.bankPlanOverflow(losses, nil)
	}
}

// fullPlanBanks is newLateGameEngine with a 60-item plan of this age's
// buildings, each bank full, and a loss for every resource.
func fullPlanBanks(tb testing.TB) (*GameEngine, []overflowLoss) {
	tb.Helper()
	ge := newLateGameEngine(tb)
	ge.doTick()
	var keys []string
	for _, k := range sortedKeys(ge.Buildings.defs) {
		d := ge.Buildings.defs[k]
		if d.RequiredAge == ge.age && d.Category != "wonder" && d.MaxCount == 0 && len(d.BaseCost) > 0 {
			keys = append(keys, k)
		}
	}
	if len(keys) < 2 {
		tb.Fatal("setup: too few buildings for a full plan")
	}
	for i := 0; len(ge.plan) < MaxPlanItems; i++ {
		k := keys[i%len(keys)]
		cost, _ := ge.Buildings.BuildBatchCost(k, 1, ge.buildQueue)
		ge.plan = append(ge.plan, PlanItem{Kind: PlanBuild, Key: k, Count: maxPlanCount, Banked: cost})
	}
	var losses []overflowLoss
	for _, res := range ge.Resources.order {
		losses = append(losses, overflowLoss{res: res, amount: 1})
	}
	return ge, losses
}

// TestBankPlanOverflow_FullBanksTakeNothing: the benchmark's worst case
// banks nothing, since every bank already holds its next copy's price.
// Timing is the benchmark's job, not a test's (CI runs tests under -race).
func TestBankPlanOverflow_FullBanksTakeNothing(t *testing.T) {
	ge, losses := fullPlanBanks(t)
	got := map[string]float64{}
	ge.bankPlanOverflow(losses, got)
	if len(got) != 0 {
		t.Fatalf("full banks took %v", got)
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
