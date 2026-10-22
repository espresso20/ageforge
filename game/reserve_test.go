package game

import (
	"testing"
	"time"
)

// A pass brings in 1% of the population, rounded up, and at least one.
func TestReserve_GrowthStepIsAShareOfThePopulation(t *testing.T) {
	for _, c := range []struct{ pop, want int }{{0, 1}, {1, 1}, {100, 1}, {101, 2}, {619, 7}, {1000, 10}, {1113, 12}} {
		if got := popGrowthStep(c.pop); got != c.want {
			t.Errorf("popGrowthStep(%d) = %d, want %d", c.pop, got, c.want)
		}
	}
}

// Past the free slots, one pass adds the step and no more, in a town with
// housing and food to spare (the 130 camps' 390 slots are full).
func TestReserve_OnePassAddsOnePercent(t *testing.T) {
	ge := sharesEngine(t, 100, map[string]int{"gathering_camp": 130})
	ge.Workers.UnlockType("worker")
	ge.Workers.Recruit("worker", 400, ge.popCapLocked())
	ge.recalculateRates()
	setAmount(ge, "food", 1e6)
	for _, pop := range []int{400, 404, 409} {
		if got := ge.Workers.TotalPop(); got != pop {
			t.Fatalf("population %d, want %d", got, pop)
		}
		ge.keepSharesLive()
	}
	if got := ge.Workers.TotalPop(); got != 414 {
		t.Errorf("population %d after three passes from 400, want 414 (4, 5 and 5)", got)
	}
}

// Population grows past the job count, up to the housing, while food allows;
// the extra villagers are idle: the reserve.
func TestReserve_GrowsPastTheJobsUpToHousing(t *testing.T) {
	ge := sharesEngine(t, 4, map[string]int{"gathering_camp": 12})
	slots := slotsOf(ge, "gathering_camp")
	growFully(ge)
	pop := ge.Workers.TotalPop()
	if pop != 40 || pop <= slots {
		t.Fatalf("population %d with %d slots, want the housing's 40", pop, slots)
	}
	if got, want := ge.Workers.IdleCount("worker"), pop-slots; got != want {
		t.Errorf("reserve %d, want %d", got, want)
	}
	st := ge.GetState()
	if st.Workers.TotalIdle != pop-slots {
		t.Errorf("state reserve %d, want %d", st.Workers.TotalIdle, pop-slots)
	}
	if st.Workers.HeldBy != HeldHousing {
		t.Errorf("held by %q, want %q", st.Workers.HeldBy, HeldHousing)
	}
	if got := RecruitStatus(st); got != RecruitHousing {
		t.Errorf("RecruitStatus = %q, want %q", got, RecruitHousing)
	}
}

// Growth stops at the food margin, and the state says how much food is left
// and how many villagers it feeds.
func TestReserve_StopsAtTheFoodMargin(t *testing.T) {
	ge := sharesEngine(t, 10, map[string]int{"gathering_camp": 1, "wood_camp": 10, "shrine": 10})
	growFully(ge)
	pop := ge.Workers.TotalPop()
	if pop == 0 || pop >= 100 {
		t.Fatalf("population %d, want it held short of the housing's 100 by food", pop)
	}
	st := ge.GetState()
	if st.Workers.HeldBy != HeldFood {
		t.Fatalf("held by %q, want %q", st.Workers.HeldBy, HeldFood)
	}
	food := st.Resources["food"]
	per := ge.Workers.FoodCostPerWorker()
	net := foodBeforeMastery(food.Rate, food.Breakdown) - st.Workers.FoodDrain
	if net-per >= recruitFoodMargin(net, st.Workers.FoodDrain, per) {
		t.Errorf("held by food, but one more villager still passes the margin (net %v, per %v)", net, per)
	}
	if st.Workers.FoodNet != net {
		t.Errorf("FoodNet = %v, want %v", st.Workers.FoodNet, net)
	}
	if want := max(0, int(net/per)); st.Workers.FoodFeeds != want {
		t.Errorf("FoodFeeds = %d, want %d (%v food over %v each)", st.Workers.FoodFeeds, want, net, per)
	}
	// More food lets it grow again.
	ge.Buildings.counts["gathering_camp"] = 10
	ge.recalculateRates()
	growFully(ge)
	if got := ge.Workers.TotalPop(); got <= pop {
		t.Errorf("population stayed at %d after more food buildings", got)
	}
}

// Housing full reads as held by housing, ahead of food.
func TestReserve_HousingFullIsHeldByHousing(t *testing.T) {
	ge := sharesEngine(t, 2, map[string]int{"gathering_camp": 20})
	growFully(ge)
	if got := ge.GetState().Workers.HeldBy; got != HeldHousing {
		t.Errorf("held by %q, want %q", got, HeldHousing)
	}
	ge.Buildings.counts["hut"] = 3
	ge.recalculateRates()
	if got := ge.GetState().Workers.HeldBy; got != HeldNone {
		t.Errorf("held by %q after a hut, want nothing", got)
	}
}

// Manual recruiting takes villagers with no free slot, within housing, into
// the reserve.
func TestReserve_ManualRecruitWithNoFreeSlot(t *testing.T) {
	ge := sharesEngine(t, 2, nil)
	ge.Workers.UnlockType("worker")
	ge.SetAutoRecruit(false)
	if err := ge.RecruitWorker("worker", 7); err != nil {
		t.Fatal(err)
	}
	if got := ge.Workers.IdleCount("worker"); got != 7 {
		t.Errorf("reserve %d, want 7", got)
	}
	if err := ge.RecruitWorker("worker", 14); err == nil {
		t.Error("recruited past the housing's 20")
	}
}

// A fresh town staffs at once: the first pass fills every free slot, and only
// the growth past them is a share of the population.
func TestReserve_JobsFillAtOnce(t *testing.T) {
	ge := sharesEngine(t, 20, map[string]int{"gathering_camp": 20, "wood_camp": 20})
	slots := slotsOf(ge, "gathering_camp") + slotsOf(ge, "wood_camp")
	ge.keepSharesLive()
	if got := ge.Workers.GetAssignedCount("worker", "gathering_camp") + ge.Workers.GetAssignedCount("worker", "wood_camp"); got != slots {
		t.Errorf("%d of %d slots filled by the first pass", got, slots)
	}
	if got, want := ge.Workers.IdleCount("worker"), popGrowthStep(0); got != want {
		t.Errorf("reserve %d after the first pass, want the step's %d", got, want)
	}
}

// An hour away grows the town as an hour at the screen does: the offline
// catch-up runs one pass for every staffEveryTicks ticks of its steps.
func TestReserve_OfflineGrowsLikeLive(t *testing.T) {
	const ticks = 5 * OfflineStepTicks
	town := map[string]int{"gathering_camp": 200}
	live := sharesEngine(t, 100, town)
	away := sharesEngine(t, 100, town)
	growFullyFor := func(ge *GameEngine) { ge.Workers.UnlockType("worker") }
	growFullyFor(live)
	growFullyFor(away)
	live.StepTicks(ticks)
	away.SimulateOffline(time.Duration(ticks) * BaseTickInterval)
	lp, ap := live.Workers.TotalPop(), away.Workers.TotalPop()
	if lp <= slotsOf(live, "gathering_camp") {
		t.Fatalf("live population %d never passed the %d slots: the test shows nothing", lp, slotsOf(live, "gathering_camp"))
	}
	if lp != ap {
		t.Errorf("after %d ticks: %d live, %d offline", ticks, lp, ap)
	}
}
