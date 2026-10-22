package game

import (
	"math"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// warCampPosts is the soldiers one War Camp comes with, from the tables.
func warCampPosts(t *testing.T) int {
	t.Helper()
	def, ok := config.BuildingByKey()["war_camp"]
	if !ok || def.Soldiers <= 0 {
		t.Fatalf("the War Camp has %d posts", def.Soldiers)
	}
	return def.Soldiers
}

// No building makes soldiers or stores them any more: a military building
// comes with its soldiers (config.BuildingDef.Soldiers) and nothing else
// writes the resource.
func TestNoBuildingMakesOrStoresSoldiers(t *testing.T) {
	military := 0
	for _, d := range config.BaseBuildings() {
		for _, e := range d.Effects {
			if e.Target == "soldiers" {
				t.Errorf("%s has a %s effect on soldiers", d.Key, e.Type)
			}
		}
		if config.IsMilitaryBuilding(d) {
			military++
			if d.Soldiers <= 0 {
				t.Errorf("%s comes with %d soldiers", d.Key, d.Soldiers)
			}
		} else if d.Soldiers != 0 {
			t.Errorf("%s is not a military building and comes with %d soldiers", d.Key, d.Soldiers)
		}
	}
	if military != 22 {
		t.Errorf("%d military buildings, want 22", military)
	}
}

// A new military building brings its soldiers at once, with the housing full:
// they move in, the reserve stops growing and recruiting is refused.
func TestSoldiersMoveInWithHousingFull(t *testing.T) {
	posts := warCampPosts(t)
	ge := peopleEngine(t)
	before := ge.GetState()
	if before.Workers.TotalPop != before.Workers.MaxPop || before.Workers.TotalPop == 0 {
		t.Fatalf("setup: population %d of %d housing, want full", before.Workers.TotalPop, before.Workers.MaxPop)
	}
	ge.mu.Lock()
	ge.Buildings.counts["war_camp"] = 2
	ge.recalculateRates()
	ge.mu.Unlock()
	st := ge.GetState()
	if st.Workers.Soldiers != 2*posts {
		t.Fatalf("soldiers = %d, want %d", st.Workers.Soldiers, 2*posts)
	}
	if st.Workers.TotalPop != before.Workers.TotalPop+2*posts {
		t.Errorf("population %d, want %d: the soldiers move in at once", st.Workers.TotalPop, before.Workers.TotalPop+2*posts)
	}
	if st.Workers.TotalPop <= st.Workers.MaxPop {
		t.Errorf("population %d should be past the housing %d", st.Workers.TotalPop, st.Workers.MaxPop)
	}
	if got := st.Workers.Working() + st.Workers.Reserve() + st.Workers.Scholars + st.Workers.Soldiers; got != st.Workers.TotalPop {
		t.Errorf("working + reserve + scholars + soldiers = %d, population %d", got, st.Workers.TotalPop)
	}
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
}

// Soldiers eat like any person: the same food a tick as a worker.
func TestSoldiersEat(t *testing.T) {
	posts := warCampPosts(t)
	ge := peopleEngine(t)
	ge.mu.Lock()
	base := ge.foodDrainLocked()
	ge.Buildings.counts["war_camp"] = 1
	ge.recalculateRates()
	drain := ge.foodDrainLocked()
	per := ge.Workers.FoodCostPerWorker()
	ge.mu.Unlock()
	if per <= 0 {
		t.Fatalf("a person eats %v", per)
	}
	if want := base + float64(posts)*per; math.Abs(drain-want) > 1e-9 {
		t.Errorf("food drain %v, want %v + %d soldiers x %v = %v", drain, base, posts, per, want)
	}
	st := ge.GetState()
	if math.Abs(st.Workers.FoodDrain-drain) > 1e-9 {
		t.Errorf("the state says %v, the engine %v", st.Workers.FoodDrain, drain)
	}
	if got := st.Resources["food"].Breakdown.FoodDrain; math.Abs(got+drain) > 1e-9 {
		t.Errorf("the food rate's breakdown has drain %v, want %v", got, -drain)
	}
}

// The soldiers resource is the posted army, every tick: it follows the copies
// that stand, a sold copy takes its soldiers with it, and nothing else moves it.
func TestSoldiersResourceIsThePostedArmyEachTick(t *testing.T) {
	posts := warCampPosts(t)
	ge := peopleEngine(t)
	army := func() (resource float64, posted int, rate float64) {
		ge.mu.RLock()
		defer ge.mu.RUnlock()
		return ge.Resources.Get("soldiers"), ge.soldierTotal(), ge.Resources.GetRate("soldiers")
	}
	if r, p, _ := army(); r != 0 || p != 0 {
		t.Fatalf("no military building, yet %v soldiers (%d posted)", r, p)
	}
	ge.mu.Lock()
	ge.Buildings.counts["war_camp"] = 3
	ge.mu.Unlock()
	for i := 0; i < 5; i++ {
		ge.StepTicks(1)
		r, p, rate := army()
		if p != 3*posts || r != float64(p) || rate != 0 {
			t.Fatalf("tick %d: %v soldiers, %d posted, rate %v; want %d, %d and no rate", i, r, p, rate, 3*posts, 3*posts)
		}
	}
	st := ge.GetState()
	if st.Military.SoldierCount != 3*posts || st.Workers.Soldiers != 3*posts {
		t.Errorf("the state says %d soldiers in the army and %d in the town, want %d", st.Military.SoldierCount, st.Workers.Soldiers, 3*posts)
	}
	if !st.Resources["soldiers"].Unlocked {
		t.Error("a standing War Camp did not unlock the soldiers")
	}
	ge.mu.Lock()
	ge.Buildings.counts["war_camp"] = 1
	ge.mu.Unlock()
	ge.StepTicks(1)
	if r, p, _ := army(); p != posts || r != float64(posts) {
		t.Errorf("after two copies went: %v soldiers, %d posted, want %d", r, p, posts)
	}
	// Defense reads the posted army.
	ge.mu.RLock()
	rating := ge.defenseRating()
	ge.mu.RUnlock()
	if want := ge.Military.CalculateDefense(posts, ge.militaryPower()); rating != want {
		t.Errorf("defense rating %v, want %v for %d soldiers", rating, want, posts)
	}
}

// The state lists the posts building by building: the army at home over the
// posts the copies have.
func TestSoldierPostsInTheState(t *testing.T) {
	posts := warCampPosts(t)
	ge := peopleEngine(t)
	ge.mu.Lock()
	ge.Buildings.counts["war_camp"] = 2
	ge.recalculateRates()
	ge.mu.Unlock()
	st := ge.GetState()
	if len(st.Military.Posts) != 1 {
		t.Fatalf("posts %+v, want one row", st.Military.Posts)
	}
	row := st.Military.Posts[0]
	if row.Key != "war_camp" || row.Count != 2 || row.PerCopy != posts || row.Posted != 2*posts || row.Posts != 2*posts {
		t.Errorf("War Camp row %+v, want 2 copies of %d posts, all posted", row, posts)
	}
	if st.Buildings["war_camp"].Soldiers != posts {
		t.Errorf("the building says %d soldiers a copy, want %d", st.Buildings["war_camp"].Soldiers, posts)
	}
	if st.Military.SoldierCap != 2*posts {
		t.Errorf("the army's strength is %d, want %d", st.Military.SoldierCap, 2*posts)
	}
}

// A military building says, when it stands, that its soldiers moved in, and
// the soldiers count as trained once (a sold copy does not take them back).
func TestSoldiersArrivalIsLoggedAndCounted(t *testing.T) {
	posts := warCampPosts(t)
	ge := peopleEngine(t)
	for key := range ge.Resources.resources {
		if key != "soldiers" {
			setAmount(ge, key, 1e6)
		}
	}
	ge.Buildings.UnlockBuilding("war_camp")
	if err := ge.BuildBuilding("war_camp"); err != nil {
		t.Fatalf("build: %v", err)
	}
	for i := 0; i < 1000 && len(ge.buildQueue) > 0; i++ {
		ge.StepTicks(1)
	}
	if ge.Buildings.GetCount("war_camp") != 1 {
		t.Fatal("the War Camp did not finish")
	}
	ge.StepTicks(1)
	if !logHas(ge, "moved in with the War Camp.") {
		t.Errorf("no line for the soldiers arriving: %v", ge.log)
	}
	if got := ge.Stats.SoldiersTrained; got != float64(posts) {
		t.Errorf("trained %v soldiers, want %d", got, posts)
	}
	ge.mu.Lock()
	ge.Buildings.counts["war_camp"] = 0
	ge.mu.Unlock()
	ge.StepTicks(1)
	if got := ge.Stats.SoldiersTrained; got != float64(posts) {
		t.Errorf("trained %v soldiers after the copy went, want the %d it brought", got, posts)
	}
}
