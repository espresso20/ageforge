package game

import (
	"os"
	"testing"
)

// Saves from before wonders needed a keystone tech, for TestKeystoneOldSave*.
// Each was written by the game as it stood one commit before the lock
// (597602a), in an age whose wonder now asks for a tech the save does not
// hold.
const (
	// An Iron Age game with the Colosseum's bank half full and no
	// Mathematics, and Road Building 400 ticks from done.
	keystoneFixtureBanked = "testdata/pre_keystone_iron.json"
	// An Industrial Age game with the Crystal Palace paid for and half way
	// through its construction, and no Industrialization.
	keystoneFixtureBuilding = "testdata/pre_keystone_industrial.json"
	// A Fusion Age game with the Stellar Cradle built, and no Fusion Power.
	keystoneFixtureBuilt = "testdata/pre_keystone_fusion.json"
)

// keystoneFixtureTechs is what each of the three games had researched: a
// handful of early techs, none of them the tech its age's wonder now needs.
var keystoneFixtureTechs = []string{"tool_making", "fire_mastery", "stoneworking", "pottery", "bronze_working", "masonry"}

// TestWriteKeystoneFixtures regenerates the three saves. Run it with
// UPDATE_FIXTURES=1 on a checkout of the game from before the lock: the
// point of the committed bytes is that the old game wrote and signed them.
// It places each game in its age directly (the unlocks of every age up to
// it, the earlier ages' wonders built), because playing there takes days.
func TestWriteKeystoneFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite the pre-keystone saves")
	}
	write := func(path, name, age string, setup func(ge *GameEngine)) {
		isolateAccountDir(t)
		ge := newSeededEngine(43)
		ge.StepTicks(30)
		ge.mu.Lock()
		ge.Stats.AgesReached = nil
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
		for _, k := range keystoneFixtureTechs {
			ge.Research.researched[k] = true
		}
		setup(ge)
		ge.Research.rebuildBonuses()
		ge.recalculateRates()
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
	write(keystoneFixtureBanked, "pre_keystone_iron", "iron_age", func(ge *GameEngine) {
		bank := map[string]float64{}
		for res, need := range ge.Buildings.defs["colosseum"].BaseCost {
			bank[res] = need / 2
		}
		ge.Buildings.wonderBanks["colosseum"] = bank
		ge.Research.currentTech, ge.Research.ticksLeft, ge.Research.totalTicks = "road_building", 400, 950
	})
	write(keystoneFixtureBuilding, "pre_keystone_industrial", "industrial_age", func(ge *GameEngine) {
		def := ge.Buildings.defs["crystal_palace"]
		bank := map[string]float64{}
		for res, need := range def.BaseCost {
			bank[res] = need
		}
		ge.Buildings.wonderBanks["crystal_palace"] = bank
		ge.buildQueue = []BuildQueueItem{{BuildingKey: "crystal_palace", TicksLeft: def.BuildTicks / 2, TotalTicks: def.BuildTicks}}
	})
	write(keystoneFixtureBuilt, "pre_keystone_fusion", "fusion_age", func(ge *GameEngine) {
		ge.Buildings.counts["stellar_cradle"] = 1
	})
}
