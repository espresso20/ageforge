package game

import (
	"math"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// warCampPosts is the soldiers one War Camp comes with, from the tables.
func warCampPosts(t *testing.T) int {
	t.Helper()
	def, ok := rules.Core().Building("war_camp")
	if !ok || def.Soldiers <= 0 {
		t.Fatalf("the War Camp has %d posts", def.Soldiers)
	}
	return def.Soldiers
}

// No building makes soldiers or stores them: a military building comes with
// its soldiers (config.BuildingDef.Soldiers). There is no soldiers resource.
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

// standWarCamps stands n War Camps and runs the staffing pass for the posts,
// as a tick does, so the soldiers the food allows take their posts.
func standWarCamps(ge *GameEngine, n int) {
	ge.mu.Lock()
	ge.Buildings.counts["war_camp"] = n
	ge.fillSoldierPosts()
	ge.recalculateRates()
	ge.mu.Unlock()
}

// With the homes full and no reserve, a new military building's soldiers
// arrive anyway, as far as the food margin allows: they move in past the
// homes, the reserve stops growing and recruiting is refused.
func TestSoldiersArriveWithHousingFull(t *testing.T) {
	posts := warCampPosts(t)
	ge := peopleEngine(t)
	before := ge.GetState()
	if before.Workers.TotalPop != before.Workers.MaxPop || before.Workers.TotalIdle != 0 || before.Workers.TotalPop == 0 {
		t.Fatalf("setup: population %d of %d housing with %d in reserve, want full and no reserve", before.Workers.TotalPop, before.Workers.MaxPop, before.Workers.TotalIdle)
	}
	standWarCamps(ge, 2)
	st := ge.GetState()
	if st.Workers.Soldiers != 2*posts {
		t.Fatalf("soldiers = %d, want %d", st.Workers.Soldiers, 2*posts)
	}
	if st.Workers.TotalPop != before.Workers.TotalPop+2*posts {
		t.Errorf("population %d, want %d: the soldiers arrive on top of the town", st.Workers.TotalPop, before.Workers.TotalPop+2*posts)
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
	// Arriving never starved the town: the food margin held.
	ge.mu.Lock()
	net, per, margin := ge.recruitFood()
	ge.mu.Unlock()
	if net < per && net < margin {
		t.Errorf("net food %v after the soldiers arrived, below one person's food %v and the margin %v", net, per, margin)
	}
}

// Posts fill from the reserve first: a person with no job takes a post and
// the population does not change; only the rest are new arrivals.
func TestSoldiersTakePostsFromTheReserveFirst(t *testing.T) {
	posts := warCampPosts(t)
	ge := peopleEngine(t)
	ge.mu.Lock()
	ge.Workers.Recruit(5, 1000) // five more than the jobs: the reserve
	ge.mu.Unlock()
	before := ge.GetState()
	idle := before.Workers.TotalIdle
	if idle <= 0 || idle >= posts {
		t.Fatalf("setup: %d in the reserve, want some but fewer than the %d posts", idle, posts)
	}
	standWarCamps(ge, 1)
	st := ge.GetState()
	if st.Workers.Soldiers != posts || st.Workers.TotalIdle != 0 {
		t.Fatalf("soldiers %d, reserve %d; want %d soldiers and the reserve used up", st.Workers.Soldiers, st.Workers.TotalIdle, posts)
	}
	if want := before.Workers.TotalPop + posts - idle; st.Workers.TotalPop != want {
		t.Errorf("population %d, want %d: %d came out of the reserve, %d arrived", st.Workers.TotalPop, want, idle, posts-idle)
	}
}

// With the food margin failing, the posts wait: nobody arrives, nobody starves,
// the log says so once, and the posts fill when the food is back.
func TestSoldierPostsWaitOnFood(t *testing.T) {
	posts := warCampPosts(t)
	ge := peopleEngine(t)
	ge.mu.Lock()
	camps := ge.Buildings.counts["gathering_camp"]
	ge.Buildings.counts["gathering_camp"] = 0
	ge.recalculateRates()
	ge.mu.Unlock()
	popBefore := ge.GetState().Workers.TotalPop
	standWarCamps(ge, 1)
	for i := 0; i < 6; i++ {
		ge.StepTicks(staffEveryTicks)
	}
	st := ge.GetState()
	if st.Workers.Soldiers != 0 || st.Workers.TotalPop > popBefore {
		t.Fatalf("with no food margin %d soldiers arrived (population %d, was %d)", st.Workers.Soldiers, st.Workers.TotalPop, popBefore)
	}
	if st.Military.SoldierCap != posts || st.Military.Posts[0].Posted != 0 || st.Military.Posts[0].Posts != posts {
		t.Errorf("posts %+v, cap %d: want 0 of %d posted", st.Military.Posts, st.Military.SoldierCap, posts)
	}
	lines := 0
	for _, l := range ge.log {
		if strings.Contains(l.Message, "posts waiting on food") {
			lines++
		}
	}
	if lines != 1 {
		t.Errorf("the log says the posts wait %d times, want once", lines)
	}
	ge.mu.Lock()
	ge.Buildings.counts["gathering_camp"] = camps
	ge.recalculateRates()
	ge.mu.Unlock()
	ge.StepTicks(staffEveryTicks * 4)
	if got := ge.GetState().Workers.Soldiers; got != posts {
		t.Errorf("with the food back %d soldiers are posted, want %d", got, posts)
	}
}

// Soldiers eat like any person: the same food a tick as a worker.
func TestSoldiersEat(t *testing.T) {
	posts := warCampPosts(t)
	ge := peopleEngine(t)
	ge.mu.Lock()
	base := ge.foodDrainLocked()
	ge.mu.Unlock()
	standWarCamps(ge, 1)
	ge.mu.Lock()
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

// The soldiers at home are the posted army, every tick: they follow the posts
// that fill, and a lost copy takes its soldiers with it.
func TestSoldiersAtHomeAreThePostedArmyEachTick(t *testing.T) {
	posts := warCampPosts(t)
	ge := peopleEngine(t)
	army := func() (home float64, posted int) {
		ge.mu.RLock()
		defer ge.mu.RUnlock()
		return float64(ge.soldiersHome()), ge.soldierTotal()
	}
	if r, p := army(); r != 0 || p != 0 {
		t.Fatalf("no military building, yet %v soldiers (%d posted)", r, p)
	}
	ge.mu.Lock()
	ge.Buildings.counts["war_camp"] = 3
	ge.mu.Unlock()
	ge.StepTicks(staffEveryTicks * 2)
	for i := 0; i < 5; i++ {
		ge.StepTicks(1)
		r, p := army()
		if p != 3*posts || r != float64(p) {
			t.Fatalf("tick %d: %v soldiers, %d posted; want %d and %d", i, r, p, 3*posts, 3*posts)
		}
	}
	st := ge.GetState()
	if st.Military.SoldierCount != 3*posts || st.Workers.Soldiers != 3*posts {
		t.Errorf("the state says %d soldiers in the army and %d in the town, want %d", st.Military.SoldierCount, st.Workers.Soldiers, 3*posts)
	}
	ge.mu.Lock()
	ge.Buildings.counts["war_camp"] = 1
	ge.mu.Unlock()
	ge.StepTicks(1)
	if r, p := army(); p != posts || r != float64(posts) {
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
	standWarCamps(ge, 2)
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
		setAmount(ge, key, config.Million)
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

// The posted soldiers are saved: a load brings the same army back at its
// posts, and the army that stands is not counted as trained again.
func TestPostedSoldiersSurviveASaveAndLoad(t *testing.T) {
	ge := peopleEngine(t)
	standWarCamps(ge, 2)
	isolateAccountDir(t)
	trained := ge.Stats.SoldiersTrained
	loaded := saveAndLoad(t, ge, "posted")
	st := loaded.GetState()
	if want := 2 * warCampPosts(t); st.Workers.Soldiers != want || st.Military.SoldierCount != want {
		t.Errorf("after a load %d soldiers in the town and %d in the army, want %d", st.Workers.Soldiers, st.Military.SoldierCount, want)
	}
	if loaded.Stats.SoldiersTrained != trained {
		t.Errorf("trained %v after a load, was %v", loaded.Stats.SoldiersTrained, trained)
	}
}
