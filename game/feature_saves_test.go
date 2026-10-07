package game

import (
	"os"
	"testing"
)

// Saves from before commands waited for a tech, for TestFeatureOldSave*. Each
// was written by the game as it stood one commit before the locks (c41c468,
// tree version 1), with commands in use that now ask for a tech the save
// does not hold.
const (
	// A Bronze Age game with a campaign under way, a trade route running, a
	// festival's cooldown ticking and an ally, and no Military Tactics.
	featureFixtureInUse = "testdata/pre_features_bronze.json"
	// An Industrial Age game with the Rail Freight route running and the
	// smugglers lying low after a deal, and none of Railroads, Mercantilism,
	// Navigation or Military Tactics. Nothing has been sent from its Army
	// panel.
	featureFixtureMixed = "testdata/pre_features_industrial.json"
)

// featureFixtureTechs is what both games had researched: a handful of early
// techs, none of them a tech that opens a command.
var featureFixtureTechs = []string{"tool_making", "fire_mastery", "stoneworking", "pottery"}

// TestWriteFeatureFixtures regenerates the two saves. Run it with
// UPDATE_FIXTURES=1 on a checkout of the game from before the locks: the
// point of the committed bytes is that the old game wrote and signed them.
// It places each game in its age directly, as TestWriteKeystoneFixtures
// does.
func TestWriteFeatureFixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite the pre-feature-lock saves")
	}
	write := func(path, name, age string, setup func(ge *GameEngine)) {
		isolateAccountDir(t)
		ge := newSeededEngine(47)
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
		for _, k := range featureFixtureTechs {
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
	write(featureFixtureInUse, "pre_features_bronze", "bronze_age", func(ge *GameEngine) {
		ge.Military.activeByCat[ExpeditionMilitary] = &ActiveExpedition{Key: "raid_bandits", Name: "Raid Bandit Camp", Soldiers: 5, TicksLeft: 200}
		ge.Trade.activeRoutes["local_barter"] = &ActiveRoute{Key: "local_barter", TicksLeft: 10}
		ge.festivalReadyTick = ge.tick + 250
		ge.Diplomacy.factions["riverlands_tribes"] = &FactionState{Discovered: true, Opinion: 60, Status: "allied"}
	})
	write(featureFixtureMixed, "pre_features_industrial", "industrial_age", func(ge *GameEngine) {
		ge.Buildings.counts["iron_works_complex"] = 1
		ge.Trade.activeRoutes["rail_freight"] = &ActiveRoute{Key: "rail_freight", TicksLeft: 10}
		ge.blackMarketReadyTick = ge.tick + 400
	})
}
