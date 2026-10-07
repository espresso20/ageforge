package game

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Saves from before the tech tree, for TestTreeOldSave*. Each was written by
// the game as it stood one commit before the tree (a461a0e), with research
// in an order that game allowed and the tree's prerequisites do not:
//
//   - Alchemy needed Mathematics; it needs Philosophy now.
//   - Quantum Computing needed Clockwork Automation; it needs Quantum
//     Mechanics too now.
const (
	// A Medieval Age game that researched Alchemy without Philosophy.
	treeFixtureResearched = "testdata/pre_tree_medieval.json"
	// An Industrial Age game with Alchemy in progress, Philosophy not
	// researched.
	treeFixtureInProgress = "testdata/pre_tree_industrial.json"
	// A Galactic Age game whose plan holds Alchemy (Philosophy neither
	// researched nor planned) and the next age's Quantum Computing (Quantum
	// Mechanics neither).
	treeFixturePlanned = "testdata/pre_tree_galactic.json"
)

// treeFixtureTechs is what the three games had researched: the first four
// ages' techs but Philosophy and what stood on it, in the order the old
// prerequisites allowed.
var treeFixtureTechs = []string{
	"tool_making", "fire_mastery", "stoneworking", "animal_husbandry", "pottery", "primitive_writing",
	"bronze_working", "agriculture", "currency", "masonry", "military_tactics",
	"iron_smelting", "road_building", "mathematics", "siege_warfare",
}

// TestWriteTreeFixtures regenerates the three saves. Run it with
// UPDATE_FIXTURES=1 on a checkout of the game from before the tree: the
// point of the committed bytes is that the old game wrote and signed them.
// It sets research and the plan directly, because the commands of today's
// game refuse these orders, which is what the fixtures are for.
func TestWriteTreeFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite the pre-tree saves")
	}
	write := func(path, name, age string, setup func(ge *GameEngine)) {
		isolateAccountDir(t)
		ge := newSeededEngine(41)
		ge.StepTicks(30)
		ge.mu.Lock()
		ge.age = age
		ge.currentEpoch = config.EpochForAge(age)
		ge.Stats.AgesReached = nil
		for _, a := range ageKeys() {
			ge.Stats.AgesReached = append(ge.Stats.AgesReached, a)
			if a == age {
				break
			}
		}
		for _, k := range treeFixtureTechs {
			ge.Research.researched[k] = true
		}
		setup(ge)
		ge.Research.rebuildBonuses()
		ge.mu.Unlock()
		if err := ge.SaveGame(name); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(savePath(name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(treeFixtureResearched, "pre_tree_medieval", "medieval_age", func(ge *GameEngine) {
		ge.Research.researched["alchemy"] = true
	})
	write(treeFixtureInProgress, "pre_tree_industrial", "industrial_age", func(ge *GameEngine) {
		ge.Research.currentTech, ge.Research.ticksLeft, ge.Research.totalTicks = "alchemy", 1200, 2200
	})
	write(treeFixturePlanned, "pre_tree_galactic", "galactic_age", func(ge *GameEngine) {
		for _, k := range []string{"chronometry", "clockwork_automation"} {
			ge.Research.researched[k] = true
		}
		ge.plan = []PlanItem{
			{Kind: PlanResearch, Key: "alchemy", Count: 1},
			{Kind: PlanResearch, Key: "quantum_computing", Count: 1},
		}
	})
}

// loadTreeFixture loads a pre-tree save and checks that the signature the
// old game wrote still verifies: nothing the tree added to a save is written
// when it is empty, and no saved key was renamed.
func loadTreeFixture(t *testing.T, path, name string) *GameEngine {
	t.Helper()
	isolateAccountDir(t)
	ge := loadFixture(t, path, name)
	if ge.cheaterBadge {
		t.Fatalf("%s failed its signature check", path)
	}
	return ge
}

// saveAndLoad saves ge and loads the save into a new engine: a second load.
func saveAndLoad(t *testing.T, ge *GameEngine, name string) *GameEngine {
	t.Helper()
	if err := ge.SaveGame(name); err != nil {
		t.Fatal(err)
	}
	again := NewGameEngine()
	if err := again.LoadGame(name); err != nil {
		t.Fatal(err)
	}
	if again.cheaterBadge {
		t.Fatalf("the re-saved %s failed its signature check", name)
	}
	return again
}

// treeFixtureKnowledge is the knowledge output the fixtures' techs add once
// Alchemy is among them: its 15%, Mathematics' 20% and Primitive Writing's
// 10%, added in the manager's order (by key) as floats, not as a constant.
func treeFixtureKnowledge() float64 {
	sum := 0.0
	for _, v := range []float64{0.15, 0.2, 0.1} {
		sum += v
	}
	return sum
}

// planKeys is the keys of the plan's items, in order.
func planKeys(ge *GameEngine) []string {
	var keys []string
	for _, it := range ge.plan {
		keys = append(keys, it.Key)
	}
	return keys
}

// TestTreeOldSaveKeepsWhatItResearched: researched stays researched,
// whatever the new prerequisites say. The save holds Alchemy without
// Philosophy, which the tree no longer allows. It loads with Alchemy
// researched, its bonuses counted and the techs that need it open, and a
// second load changes nothing.
func TestTreeOldSaveKeepsWhatItResearched(t *testing.T) {
	ge := loadTreeFixture(t, treeFixtureResearched, "pre_tree_medieval")
	check := func(ge *GameEngine, when string) {
		t.Helper()
		held := append([]string{"alchemy"}, treeFixtureTechs...)
		for _, k := range held {
			if !ge.Research.IsResearched(k) {
				t.Errorf("%s: %s is no longer researched", when, k)
			}
		}
		if got := ge.Research.ResearchedCount(); got != len(held) {
			t.Errorf("%s: %d techs researched, the save held %d", when, got, len(held))
		}
		if ge.Research.IsResearched("philosophy") {
			t.Fatalf("%s: Philosophy is researched: the save is not out of order any more", when)
		}
		// Alchemy still does what it does: +15% knowledge on top of
		// Primitive Writing's 10% and Mathematics' 20%, and 0.1 gold a tick.
		if got, want := ge.Research.Bonus(config.EffectOutput, "knowledge"), treeFixtureKnowledge(); got != want {
			t.Errorf("%s: knowledge output bonus %v, want %v", when, got, want)
		}
		if got := ge.Research.Bonus(config.EffectFlatOutput, "gold"); got != 0.1 {
			t.Errorf("%s: gold a tick from techs %v, want Alchemy's 0.1", when, got)
		}
		// And it counts as a prerequisite: Gunpowder needs Alchemy and Siege
		// Warfare, and has both.
		techs := ge.GetState().Research.Techs
		if !techs["alchemy"].Researched || !techs["gunpowder"].PrereqsMet {
			t.Errorf("%s: Alchemy researched %v, Gunpowder's prerequisites met %v", when, techs["alchemy"].Researched, techs["gunpowder"].PrereqsMet)
		}
		// A new start is held to the new rules: Alchemy would be refused now.
		if def, _ := ge.rules.Tech("alchemy"); def.PrereqsMet(func(k string) bool { return k != "alchemy" && ge.Research.IsResearched(k) }) {
			t.Errorf("%s: Alchemy's new prerequisites are met: the save does not test the grandfathering", when)
		}
	}
	check(ge, "first load")
	check(saveAndLoad(t, ge, "pre_tree_medieval_again"), "second load")
}

// TestTreeOldSaveFinishesItsResearch: a research in progress finishes, even
// if its new prerequisites are not met. The save has Alchemy 1,000 ticks
// into its 2,200 and no Philosophy. A load first runs the time since the
// save was written, and that finishes it in any run later than 40 minutes
// after the fixture was made; a run sooner than that finds it still going,
// never cancelled, and plays out the rest of the day away here. (The test
// after this one is the same promise tick by tick.)
func TestTreeOldSaveFinishesItsResearch(t *testing.T) {
	ge := loadTreeFixture(t, treeFixtureInProgress, "pre_tree_industrial")
	if ge.Research.IsResearched("philosophy") {
		t.Fatal("Philosophy is researched: the save is not out of order any more")
	}
	if !ge.Research.IsResearched("alchemy") {
		if ge.Research.currentTech != "alchemy" || ge.Research.ticksLeft <= 0 || ge.Research.ticksLeft > 1200 {
			t.Fatalf("after loading: researching %q with %d ticks left, want alchemy with 1,200 or fewer", ge.Research.currentTech, ge.Research.ticksLeft)
		}
		ge.SimulateOffline(MaxOfflineTime)
	}
	if !ge.Research.IsResearched("alchemy") || ge.Research.currentTech != "" {
		t.Fatalf("Alchemy researched %v, still researching %q", ge.Research.IsResearched("alchemy"), ge.Research.currentTech)
	}
	if got, want := ge.Research.Bonus(config.EffectOutput, "knowledge"), treeFixtureKnowledge(); got != want {
		t.Errorf("knowledge output bonus %v after Alchemy, want %v", got, want)
	}
	again := saveAndLoad(t, ge, "pre_tree_industrial_again")
	if !again.Research.IsResearched("alchemy") || again.Research.ResearchedCount() != ge.Research.ResearchedCount() {
		t.Errorf("a second load changed what is researched: %v, was %v", again.Research.GetResearched(), ge.Research.GetResearched())
	}
}

// TestResearchInProgressOutOfOrderFinishes is the same promise without the
// clock: research restored mid-way, as a load restores it, with a
// prerequisite the tree now asks for missing, ticks to its end.
func TestResearchInProgressOutOfOrderFinishes(t *testing.T) {
	ge := NewGameEngine()
	ge.Research.LoadState(treeFixtureTechs, "alchemy", 1200, 2200)
	if def, _ := ge.rules.Tech("alchemy"); def.PrereqsMet(ge.Research.IsResearched) {
		t.Fatal("Alchemy's prerequisites are met: nothing is out of order")
	}
	for i := 0; i < 1199; i++ {
		ge.processResearch()
	}
	if ge.Research.currentTech != "alchemy" || ge.Research.ticksLeft != 1 {
		t.Fatalf("one tick from the end: researching %q, %d ticks left", ge.Research.currentTech, ge.Research.ticksLeft)
	}
	ge.processResearch()
	if !ge.Research.IsResearched("alchemy") || ge.Research.currentTech != "" {
		t.Errorf("Alchemy researched %v, still researching %q", ge.Research.IsResearched("alchemy"), ge.Research.currentTech)
	}
	if !logHas(ge, "Research complete: Alchemy.") {
		t.Error("no line in the log for the finished research")
	}
}

// TestTreeOldSavePlanWaits: plan items whose prerequisites changed simply
// wait. The save's plan holds Alchemy (Philosophy missing) and the next
// age's Quantum Computing (Quantum Mechanics missing). Through the load, the
// time away and a second load they stay in the plan, in order, saying what
// they need. Neither starts, though the knowledge is there, and each starts
// once what it needs is researched.
func TestTreeOldSavePlanWaits(t *testing.T) {
	ge := loadTreeFixture(t, treeFixturePlanned, "pre_tree_galactic")
	fund := func(ge *GameEngine) {
		ge.Resources.resources["knowledge"].Storage = 1e12
		setAmount(ge, "knowledge", 1e10)
	}
	waiting := func(ge *GameEngine, when string) {
		t.Helper()
		if got, want := planKeys(ge), []string{"alchemy", "quantum_computing"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: the plan holds %v, want %v", when, got, want)
		}
		if ge.Research.currentTech != "" || ge.Research.IsResearched("alchemy") || ge.Research.IsResearched("quantum_computing") {
			t.Errorf("%s: researching %q; Alchemy researched %v, Quantum Computing %v; want both waiting", when,
				ge.Research.currentTech, ge.Research.IsResearched("alchemy"), ge.Research.IsResearched("quantum_computing"))
		}
		v := ge.planViews()
		if len(v) != 2 || v[0].Status != PlanStatusBlocked || v[0].Note != "needs Philosophy first" ||
			v[1].Status != PlanStatusBlocked || v[1].Note != "needs Quantum Mechanics first" {
			t.Errorf("%s: views = %+v, want each item blocked on its new prerequisite", when, v)
		}
		for _, l := range ge.log {
			if strings.HasPrefix(l.Message, "Plan: dropped") {
				t.Errorf("%s: the plan dropped an item: %s", when, l.Message)
			}
		}
	}
	waiting(ge, "after loading")
	// A full day away, whatever the load itself ran: the plan walks the
	// whole time and still holds both.
	ge.SimulateOffline(MaxOfflineTime)
	waiting(ge, "after a day away")
	// With the knowledge to pay for both, the plan still starts neither.
	fund(ge)
	ge.runPlanTick()
	waiting(ge, "with knowledge to spare")

	ge = saveAndLoad(t, ge, "pre_tree_galactic_again")
	waiting(ge, "after a second load")

	// Philosophy arrives: Alchemy starts from the plan. Quantum Computing
	// goes on waiting.
	fund(ge)
	learn(ge, "philosophy")
	ge.runPlanTick()
	if ge.Research.currentTech != "alchemy" || !reflect.DeepEqual(planKeys(ge), []string{"quantum_computing"}) {
		t.Fatalf("with Philosophy: researching %q, plan %v; want alchemy started and Quantum Computing left", ge.Research.currentTech, planKeys(ge))
	}
	researchOut(ge)
	// The Quantum Age and Quantum Mechanics arrive: so does its turn.
	ge.age = "quantum_age"
	ge.runPlanTick()
	if v := ge.planViews(); ge.Research.currentTech != "" || len(v) != 1 || v[0].Note != "needs Quantum Mechanics first" {
		t.Fatalf("in the Quantum Age without Quantum Mechanics: researching %q, views %+v", ge.Research.currentTech, v)
	}
	learn(ge, "quantum_mechanics")
	fund(ge)
	ge.runPlanTick()
	if ge.Research.currentTech != "quantum_computing" || len(ge.plan) != 0 {
		t.Errorf("with Quantum Mechanics: researching %q, plan %v", ge.Research.currentTech, planKeys(ge))
	}
}
