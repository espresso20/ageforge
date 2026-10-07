package game

import (
	"os"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// A save written under tree version 3, for TestContent2OldSave: the tree as
// it stood before the Steel and Electric Eras' new techs (2881192), when no
// building of those eras but the Nuclear Plant waited for a tech.
//
// An Industrial Age game with two Coal Plants built, a Steam Mine half way
// up and a Research Institute in the plan, and none of the techs those
// buildings now wait for.
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
