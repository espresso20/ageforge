package game

import (
	"os"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// Saves written under tree version 4, for TestContent3OldSave*: the tree as
// it stood before the Digital, Neon and Cosmic Eras' new techs (0925106).
// The Warp Commerce route waited for no tech then (its lock named one the
// tree did not hold yet), and no building of those eras but the Smart Farm,
// the Smart Complex, the Holographic Theater and the Energy Exchange waited
// for one.
const (
	// An Interstellar Age game that never ran a Warp Commerce route: two
	// Pulsar Taps standing, a Galactic Trade Hub half way up and Galactic
	// Network Nodes in the plan.
	content3FixtureFresh = "testdata/tree_v4_interstellar.json"
	// The same game with the Warp Commerce route running, started through
	// the game's own command, so the save carries what that game recorded
	// as in use.
	content3FixtureInUse = "testdata/tree_v4_interstellar_route.json"
)

// content3FixtureTechs is what both games had researched: early techs, The
// Wheel among them (trade routes wait for it), and none a building of the
// Cosmic Era waits for.
var content3FixtureTechs = []string{"language", "tool_making", "fire_mastery", "stoneworking", "primitive_writing", "woodworking", "the_wheel"}

// TestWriteContent3Fixtures regenerates the two saves. Run it with
// UPDATE_FIXTURES=1 on a checkout of the game at tree version 4: the point
// of the committed bytes is that the old game wrote and signed them. It
// places each game in its age directly, as TestWriteKeystoneFixtures does.
func TestWriteContent3Fixtures(t *testing.T) {
	if os.Getenv("UPDATE_FIXTURES") == "" {
		t.Skip("set UPDATE_FIXTURES=1 to rewrite the tree version 4 saves")
	}
	if config.TechTreeVersion != 4 {
		t.Fatalf("this checkout writes tree version %d saves: the fixtures are version 4", config.TechTreeVersion)
	}
	write := func(path, name string, play func(ge *GameEngine)) {
		isolateAccountDir(t)
		ge := newSeededEngine(61)
		ge.StepTicks(30)
		ge.mu.Lock()
		ge.Stats.AgesReached = nil
		const age = "interstellar_age"
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
		for _, k := range content3FixtureTechs {
			ge.Research.researched[k] = true
		}
		ge.Buildings.counts["pulsar_tap"] = 2
		ge.Buildings.counts["warp_drive_plant"] = 2
		def := ge.Buildings.defs["galactic_trade_hub"]
		ge.buildQueue = []BuildQueueItem{{BuildingKey: "galactic_trade_hub", TicksLeft: def.BuildTicks / 2, TotalTicks: def.BuildTicks}}
		ge.plan = []PlanItem{{Kind: PlanBuild, Key: "galactic_network_node", Count: 2}}
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
	write(content3FixtureFresh, "tree_v4_interstellar", nil)
	write(content3FixtureInUse, "tree_v4_interstellar_route", func(ge *GameEngine) {
		// The game's own command, so its save says what it used.
		setResource(ge, "gold", 1e9)
		if err := ge.StartTradeRoute("warp_commerce"); err != nil {
			t.Fatalf("starting Warp Commerce: %v", err)
		}
	})
}
