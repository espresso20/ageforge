package game

import (
	"math"
	"testing"
)

// peopleEngine is a Stone Age town with a few huts, camps and food, grown to
// its housing, with no knowledge building yet.
func peopleEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := sharesEngine(t, 2, map[string]int{"gathering_camp": 10, "wood_camp": 10})
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.age = "stone_age"
	ge.Workers.SetAge("stone_age")
	ge.Workers.UnlockType("worker")
	ge.recalculateRates()
	growFully(ge)
	return ge
}

// The population is the workers, the reserve and the scholars.
func TestPopulationCountsScholars(t *testing.T) {
	ge := scholarEngine(t)
	st := ge.GetState()
	if st.Workers.Scholars != 6 || st.Workers.TotalPop != 6 {
		t.Fatalf("scholars %d, population %d: want 6 and 6 (the scholars are the whole town)", st.Workers.Scholars, st.Workers.TotalPop)
	}
	if st.Workers.Working() != 0 || st.Workers.Reserve() != 0 {
		t.Errorf("working %d, reserve %d: want none", st.Workers.Working(), st.Workers.Reserve())
	}
	ge.mu.Lock()
	ge.Buildings.counts["hut"] = 3
	ge.recalculateRates()
	ge.Workers.Recruit("worker", 3, ge.workerCapLocked())
	ge.mu.Unlock()
	st = ge.GetState()
	if st.Workers.TotalPop != 9 || st.Workers.Reserve() != 3 {
		t.Errorf("population %d reserve %d: want 9 and 3", st.Workers.TotalPop, st.Workers.Reserve())
	}
	if got := st.Workers.Working() + st.Workers.Reserve() + st.Workers.Scholars; got != st.Workers.TotalPop {
		t.Errorf("working + reserve + scholars = %d, population %d", got, st.Workers.TotalPop)
	}
}

// A new knowledge building brings its scholars at once, with the housing
// full: they move in, the reserve stops growing, and recruiting is refused.
func TestScholarsMoveInWithHousingFull(t *testing.T) {
	ge := peopleEngine(t)
	before := ge.GetState()
	if before.Workers.TotalPop != before.Workers.MaxPop || before.Workers.TotalPop == 0 {
		t.Fatalf("setup: population %d of %d housing, want full", before.Workers.TotalPop, before.Workers.MaxPop)
	}
	ge.mu.Lock()
	ge.Buildings.counts["story_circle"] = 2
	ge.recalculateRates()
	ge.mu.Unlock()
	st := ge.GetState()
	if st.Workers.Scholars != 4 {
		t.Fatalf("scholars = %d, want 4", st.Workers.Scholars)
	}
	if st.Workers.TotalPop != before.Workers.TotalPop+4 {
		t.Errorf("population %d, want %d: the scholars move in at once", st.Workers.TotalPop, before.Workers.TotalPop+4)
	}
	if st.Workers.TotalPop <= st.Workers.MaxPop {
		t.Errorf("population %d should be past the housing %d", st.Workers.TotalPop, st.Workers.MaxPop)
	}
	// Growth waits: the reserve does not grow and a recruit is refused.
	ge.StepTicks(staffEveryTicks * 4)
	if after := ge.GetState(); after.Workers.TotalPop != st.Workers.TotalPop || after.Workers.TotalIdle != st.Workers.TotalIdle {
		t.Errorf("the town grew from %d (%d idle) to %d (%d idle) with no room", st.Workers.TotalPop, st.Workers.TotalIdle, after.Workers.TotalPop, after.Workers.TotalIdle)
	}
	if err := ge.RecruitWorker("worker", 1); err == nil {
		t.Error("recruited into a full town")
	}
	if st.Workers.HeldBy != HeldHousing {
		t.Errorf("held by %q, want housing", st.Workers.HeldBy)
	}
	// Room again: more housing lets the reserve grow past the scholars.
	ge.mu.Lock()
	ge.Buildings.counts["hut"] += 20
	ge.recalculateRates()
	ge.mu.Unlock()
	cap := ge.GetState().Workers.MaxPop
	growFully(ge)
	if g := ge.GetState(); g.Workers.TotalPop <= st.Workers.TotalPop || g.Workers.TotalPop > cap {
		t.Errorf("after new huts population is %d of %d (was %d)", g.Workers.TotalPop, cap, st.Workers.TotalPop)
	}
}

// A knowledge building taken away (the game never sells one, but a lost copy
// is the same count going down) takes its scholars with it.
func TestScholarsLeaveWithTheirBuilding(t *testing.T) {
	ge := scholarEngine(t)
	ge.mu.Lock()
	ge.Buildings.counts["story_circle"] = 1
	ge.recalculateRates()
	ge.mu.Unlock()
	st := ge.GetState()
	if st.Workers.Scholars != 2+2 || st.Workers.TotalPop != 4 {
		t.Fatalf("scholars %d, population %d: want 4 and 4", st.Workers.Scholars, st.Workers.TotalPop)
	}
}

// Scholars eat like any person: the same food a tick as a worker.
func TestScholarsEat(t *testing.T) {
	ge := scholarEngine(t)
	ge.mu.Lock()
	per := ge.Workers.FoodCostPerWorker()
	drain := ge.foodDrainLocked()
	ge.mu.Unlock()
	if per <= 0 {
		t.Fatalf("a person eats %v", per)
	}
	if want := 6 * per; math.Abs(drain-want) > 1e-9 {
		t.Errorf("food drain %v, want 6 scholars x %v = %v", drain, per, want)
	}
	st := ge.GetState()
	if math.Abs(st.Workers.FoodDrain-drain) > 1e-9 {
		t.Errorf("the state says %v, the engine %v", st.Workers.FoodDrain, drain)
	}
	if got := st.Resources["food"].Breakdown.FoodDrain; math.Abs(got+drain) > 1e-9 {
		t.Errorf("the food rate's breakdown has drain %v, want %v", got, -drain)
	}
	// And the same town without them eats less.
	ge.mu.Lock()
	ge.Buildings.counts["story_circle"] = 0
	ge.Buildings.counts["elders_hall"] = 0
	ge.recalculateRates()
	none := ge.foodDrainLocked()
	ge.mu.Unlock()
	if none != 0 {
		t.Errorf("no people, food drain %v", none)
	}
}

// The state names the kinds of person in the age's words.
func TestStateTermsFollowTheAge(t *testing.T) {
	ge := scholarEngine(t) // a Stone Age town
	st := ge.GetState()
	if got := st.Term("scholars").Many; got != "Storytellers" {
		t.Errorf("stone scholars = %q", got)
	}
	st.Age = "bronze_age"
	if got := st.Term("scholars").Many; got != "Scribes" {
		t.Errorf("bronze scholars = %q", got)
	}
	if got := st.Term("workers").One; got != "Craftsperson" {
		t.Errorf("bronze worker = %q", got)
	}
}
