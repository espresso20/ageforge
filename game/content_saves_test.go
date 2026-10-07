package game

import (
	"os"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Saves written under tree version 2, for TestContentOldSave*: the tree as it
// stood before the Stone and Iron Eras' new techs (310f3a5). Stonehenge had
// no keystone then, and trade routes, expeditions, diplomacy and festivals
// waited for no tech: their locks named techs the tree did not hold yet.
const (
	// A Bronze Age game with Stonehenge's bank half full, no trade route
	// ever started, nobody met and no festival held.
	contentFixtureFresh = "testdata/tree_v2_bronze.json"
	// An Iron Age game with a trade route running and an ally, both begun
	// through the game's own commands, so the save carries what that game
	// recorded as in use. It has sent no expedition and held no festival.
	contentFixtureInUse = "testdata/tree_v2_iron.json"
)

// contentFixtureTechs is what both games had researched: early techs, none
// of them one the new locks ask for.
var contentFixtureTechs = []string{"tool_making", "fire_mastery", "stoneworking", "pottery", "primitive_writing"}

// TestWriteContentFixtures regenerates the two saves. Run it with
// UPDATE_FIXTURES=1 on a checkout of the game at tree version 2: the point
// of the committed bytes is that the old game wrote and signed them. It
// places each game in its age directly, as TestWriteKeystoneFixtures does.
func TestWriteContentFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite the tree version 2 saves")
	}
	if config.TechTreeVersion != 2 {
		t.Fatalf("this checkout writes tree version %d saves: the fixtures are version 2", config.TechTreeVersion)
	}
	write := func(path, name, age string, setup func(ge *GameEngine), play func(ge *GameEngine)) {
		isolateAccountDir(t)
		ge := newSeededEngine(53)
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
		for _, k := range contentFixtureTechs {
			ge.Research.researched[k] = true
		}
		setup(ge)
		ge.Research.rebuildBonuses()
		ge.recalculateRates()
		ge.mu.Unlock()
		if play != nil {
			play(ge)
		}
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
	write(contentFixtureFresh, "tree_v2_bronze", "bronze_age", func(ge *GameEngine) {
		bank := map[string]float64{}
		for res, need := range ge.Buildings.defs["stonehenge"].BaseCost {
			bank[res] = need / 2
		}
		ge.Buildings.wonderBanks["stonehenge"] = bank
	}, nil)
	write(contentFixtureInUse, "tree_v2_iron", "iron_age", func(ge *GameEngine) {
		ge.Buildings.counts["market"] = 3
		ge.Diplomacy.factions["riverlands_tribes"] = &FactionState{Discovered: true, Opinion: 60, Status: "neutral"}
	}, func(ge *GameEngine) {
		// The game's own commands, so its save says what it used.
		for _, res := range []string{"food", "wood", "stone", "gold"} {
			setResource(ge, res, 1e6)
		}
		if err := ge.StartTradeRoute("local_barter"); err != nil {
			t.Fatalf("starting a route: %v", err)
		}
		if err := ge.SetDiplomaticStatus("riverlands_tribes", "allied"); err != nil {
			t.Fatalf("allying: %v", err)
		}
	})
}
