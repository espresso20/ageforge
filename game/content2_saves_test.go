package game

import (
	"os"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// A save written under tree version 3, for TestContent2OldSave: the tree as
// it stood before the Steel and Electric Eras' new techs (2881192), when no
// building of those eras but the Nuclear Plant waited for a tech.
//
// An Industrial Age game with two Coal Plants built, a Steam Mine half way
// up and Research Institutes in the plan, and no Steam Power, which the
// Coal Plant now waits for.
const content2Fixture = "testdata/tree_v3_industrial.json"

// content2FixtureTechs is what the game had researched: early techs only.
var content2FixtureTechs = []string{"language", "tool_making", "fire_mastery", "stoneworking", "primitive_writing"}

// TestWriteContent2Fixture regenerates the save. Run it with
// UPDATE_FIXTURES=1 on a checkout of the game at tree version 3: the point
// of the committed bytes is that the old game wrote and signed them. It
// places the game in its age directly, as TestWriteKeystoneFixtures does.
func TestWriteContent2Fixture(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite the tree version 3 save")
	}
	if config.TechTreeVersion != 3 {
		t.Fatalf("this checkout writes tree version %d saves: the fixture is version 3", config.TechTreeVersion)
	}
	isolateAccountDir(t)
	ge := newSeededEngine(59)
	ge.StepTicks(30)
	ge.mu.Lock()
	ge.Stats.AgesReached = nil
	const age = "industrial_age"
	for _, a := range ageKeys() {
		ge.Stats.AgesReached = append(ge.Stats.AgesReached, a)
		ge.applyAgeUnlocks(a)
		if a == age {
			break
		}
		if w := ge.progress.WonderForAge(a); w != "" {
			ge.Buildings.counts[w] = 1
		}
	}
	ge.age = age
	ge.currentEpoch = ge.rules.EraOf(age)
	ge.Workers.SetAge(age)
	for _, k := range content2FixtureTechs {
		ge.Research.researched[k] = true
	}
	ge.Buildings.counts["coal_plant"] = 2
	def := ge.Buildings.defs["steam_mine"]
	ge.buildQueue = []BuildQueueItem{{BuildingKey: "steam_mine", TicksLeft: def.BuildTicks / 2, TotalTicks: def.BuildTicks}}
	ge.plan = []PlanItem{{Kind: PlanBuild, Key: "research_institute", Count: 2}}
	ge.Research.rebuildBonuses()
	ge.recalculateRates()
	ge.mu.Unlock()
	if err := ge.SaveGame("tree_v3_industrial"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(savePath("tree_v3_industrial"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(content2Fixture, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestContent2OldSave: a game saved under tree version 3 in the Industrial
// Age, with two Coal Plants standing, a Steam Mine half built, Research
// Institutes in the plan and no Steam Power. The Coal Plant waits for Steam
// Power now, and the Grand Embassy and the Geographic Society for techs of
// their own. The save loads with its age exempt: one log line says so, what
// stands, what is half built and what is planned are as they were, and a
// third Coal Plant, a Grand Embassy and a Geographic Society build without
// their techs. The next age ends the grace: the Steam Works waits for
// Electrification, as it does for anyone, and nothing built is lost. A
// second load says nothing.
func TestContent2OldSave(t *testing.T) {
	ge := loadContentFixture(t, content2Fixture, "tree_v3_industrial")
	if ge.age != "industrial_age" || ge.Buildings.graceAge != "industrial_age" {
		t.Fatalf("loaded in %s with grace age %q, want the Industrial Age and its grace", ge.age, ge.Buildings.graceAge)
	}
	for _, k := range content2FixtureTechs {
		if !ge.Research.IsResearched(k) {
			t.Errorf("%s is no longer researched", k)
		}
	}
	for _, k := range []string{"steam_power", "concert_of_nations", "geographic_societies", "electrification"} {
		if ge.Research.IsResearched(k) {
			t.Fatalf("%s is researched: the save does not test the grace", k)
		}
	}
	n, line := treeNotices(ge)
	if n != 1 || !strings.Contains(line, "new techs from the Renaissance to the Atomic Age") || !strings.Contains(line, "Industrial Age") ||
		!strings.Contains(line, "the locks start when you next advance") || !strings.Contains(line, "Everything you have researched or built stays") ||
		strings.Contains(line, "command") {
		t.Errorf("%d research update lines, the last %q; want one about the new techs and the buildings that wait for one, naming the Industrial Age", n, line)
	}
	if got := grantedList(ge); got != "" {
		t.Errorf("granted on load: %q, want nothing: no command is newly locked", got)
	}
	// What stood, what was going up and what was planned are as they were.
	if got := ge.Buildings.GetCount("coal_plant"); got != 2 {
		t.Errorf("%d Coal Plants after the load, want the 2 that stood", got)
	}
	// (The time away goes on building it, so how far along it is depends
	// on the day the test runs.)
	if got := ge.Buildings.GetCount("steam_mine") + queued(ge, "steam_mine"); got != 1 {
		t.Errorf("%d Steam Mines standing or going up, want the one the save was building (queue %+v)", got, ge.buildQueue)
	}
	if len(ge.plan) != 1 || ge.plan[0].Key != "research_institute" || ge.plan[0].Count != 2 {
		t.Errorf("the plan is %+v, want two Research Institutes", ge.plan)
	}
	// In the grace age no building of the age waits for a tech.
	for _, key := range []string{"coal_plant", "grand_embassy", "geographic_society"} {
		stock(ge, 0)
		if bs := ge.GetState().Buildings[key]; bs.NeedsTech != "" {
			t.Errorf("in the grace age %s reads %+v, want it open", key, bs)
		}
		before := ge.Buildings.GetCount(key) + queued(ge, key)
		if err := ge.BuildBuilding(key); err != nil {
			t.Errorf("%s in the grace age, without its tech: %v", key, err)
		}
		if got := ge.Buildings.GetCount(key) + queued(ge, key); got != before+1 {
			t.Errorf("%s: %d standing or going up, want %d", key, got, before+1)
		}
	}

	// The next age: the grace is over. Its own buildings wait for their
	// techs like anyone's, and nothing built is lost.
	if err := ge.EnterAgeForTest("victorian_age"); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.graceAge != "" {
		t.Fatalf("after the advance the grace age is still %q", ge.Buildings.graceAge)
	}
	if got := ge.Buildings.GetCount("coal_plant") + queued(ge, "coal_plant"); got != 3 {
		t.Errorf("%d Coal Plants after the advance, want the 3 the run built", got)
	}
	stock(ge, 0)
	if bs := ge.GetState().Buildings["steam_works"]; bs.NeedsTech != "electrification" {
		t.Errorf("the Steam Works reads %+v, want it to wait for Electrification", bs)
	}
	if err := ge.BuildBuilding("steam_works"); err == nil || !strings.Contains(err.Error(), "Electrification") {
		t.Errorf("a Steam Works without Electrification: %v", err)
	}
	learn(ge, "electrification")
	stock(ge, 0)
	if err := ge.BuildBuilding("steam_works"); err != nil {
		t.Errorf("a Steam Works with Electrification researched: %v", err)
	}

	// A second load: nothing is said twice, and the save is on today's tree.
	again := saveAndLoad(t, ge, "tree_v3_industrial_again")
	if v, grace := savedTree(t, "tree_v3_industrial_again"); v != config.TechTreeVersion || grace != "" {
		t.Errorf("the re-saved game carries tree version %d and grace age %q, want %d and none", v, grace, config.TechTreeVersion)
	}
	if n, _ := treeNotices(again); n != 0 {
		t.Errorf("a second load said it again (%d lines)", n)
	}
	if got := again.Buildings.GetCount("coal_plant") + queued(again, "coal_plant"); got != 3 {
		t.Errorf("%d Coal Plants after the second load, want 3", got)
	}
}
