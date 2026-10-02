package game

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// atomicEngine is an engine in the Atomic Age with its buildings unlocked
// and plenty of every resource.
func atomicEngine(t *testing.T) *GameEngine {
	t.Helper()
	ge := NewGameEngine()
	ge.age = "atomic_age"
	ge.applyAgeUnlocks("atomic_age")
	for _, r := range config.BaseResources() {
		ge.Resources.UnlockResource(r.Key)
		ge.Resources.AddStorage(r.Key, 1e15)
		ge.Resources.Add(r.Key, 1e14)
	}
	return ge
}

// TestTechGatedBuilding: a building with a RequiredTech stays locked in its
// own age until the tech is researched, says why, and can be planned ahead.
func TestTechGatedBuilding(t *testing.T) {
	ge := atomicEngine(t)
	const bld, tech = "nuclear_plant", "civilian_reactors"
	if got := config.BuildingByKey()[bld].RequiredTech; got != tech {
		t.Fatalf("%s needs %q, want %q", bld, got, tech)
	}

	if ge.Buildings.IsUnlocked(bld) {
		t.Fatalf("%s is unlocked before %s", bld, tech)
	}
	err := ge.BuildBuilding(bld)
	if err == nil || !strings.Contains(err.Error(), "Civilian Reactors") {
		t.Fatalf("building %s before %s: %v, want an error naming the tech", bld, tech, err)
	}
	st := ge.GetState().Buildings[bld]
	if st.Unlocked || st.CanBuild || st.NeedsTech != tech {
		t.Errorf("state before the tech: unlocked %v, can build %v, needs %q", st.Unlocked, st.CanBuild, st.NeedsTech)
	}
	// Other Atomic buildings are not held back.
	if !ge.Buildings.IsUnlocked("nuclear_reactor") {
		t.Error("nuclear_reactor is locked; only tech-gated buildings should be")
	}

	// The plan takes it and waits for the tech.
	if _, err := ge.PlanAddBuild(bld, 1); err != nil {
		t.Fatalf("planning %s before %s: %v", bld, tech, err)
	}
	if c := ge.checkPlanItem(PlanItem{Kind: PlanBuild, Key: bld}, false); !strings.Contains(c.blocked, "Civilian Reactors") {
		t.Errorf("plan item blocked %q, want it to wait for Civilian Reactors", c.blocked)
	}

	ge.Research.researched[tech] = true
	if !ge.Buildings.IsUnlocked(bld) {
		t.Fatalf("%s still locked after %s", bld, tech)
	}
	if err := ge.BuildBuilding(bld); err != nil {
		t.Errorf("building %s after %s: %v", bld, tech, err)
	}
	if st := ge.GetState().Buildings[bld]; !st.Unlocked || st.NeedsTech != "" {
		t.Errorf("state after the tech: unlocked %v, needs %q", st.Unlocked, st.NeedsTech)
	}
}

// TestTechGateSurvivesReset: a reset builds a new BuildingManager, which must
// still read the engine's research (whichever ResearchManager it holds).
func TestTechGateSurvivesReset(t *testing.T) {
	ge := atomicEngine(t)
	ge.Research.researched["civilian_reactors"] = true
	ge.Reset()
	ge.age = "atomic_age"
	ge.applyAgeUnlocks("atomic_age")
	if ge.Buildings.IsUnlocked("nuclear_plant") {
		t.Fatal("nuclear_plant unlocked after a reset wiped research")
	}
	ge.Research.researched["civilian_reactors"] = true
	if !ge.Buildings.IsUnlocked("nuclear_plant") {
		t.Fatal("nuclear_plant locked after researching its tech on the new run")
	}
}

// TestMidAgeTechs: each mid-age tech opens in the age of the buildings it
// gates, and every gated building's tech is in the building's own age.
func TestMidAgeTechs(t *testing.T) {
	techs := config.TechByKey()
	want := map[string]string{
		"civilian_reactors":  "atomic_age",
		"internet_of_things": "information_age",
		"holography":         "cyberpunk_age",
		"maglev_transit":     "fusion_age",
	}
	gates := map[string]int{}
	for _, b := range config.BaseBuildings() {
		if b.RequiredTech == "" {
			continue
		}
		gates[b.RequiredTech]++
		if techs[b.RequiredTech].Age != b.RequiredAge {
			t.Errorf("%s (%s) needs %s from %s", b.Key, b.RequiredAge, b.RequiredTech, techs[b.RequiredTech].Age)
		}
	}
	for k, age := range want {
		td, ok := techs[k]
		if !ok || td.Age != age {
			t.Errorf("tech %s: found %v, age %q, want %q", k, ok, td.Age, age)
			continue
		}
		if gates[k] == 0 {
			t.Errorf("tech %s opens no building", k)
		}
	}
}
