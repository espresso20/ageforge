package game

import (
	"os"
	"strings"
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

// TestContent3OldSaveGetsAnAgeOfGrace: a game saved under tree version 4 in
// the Interstellar Age that never ran a Warp Commerce route, with two Pulsar
// Taps standing, a Galactic Trade Hub half built, Galactic Network Nodes in
// the plan and neither Stellar Engineering nor Interstellar Trade. It loads
// with its age exempt: one log line says so, what stands, what is going up
// and what is planned are as they were, a third Pulsar Tap builds without
// its tech and the Warp Commerce route starts without Interstellar Trade.
// The next age ends the grace: the Quasar Tap waits for Antimatter
// Synthesis, as it does for anyone. A second load says nothing.
func TestContent3OldSaveGetsAnAgeOfGrace(t *testing.T) {
	ge := loadContentFixture(t, content3FixtureFresh, "tree_v4_interstellar")
	if ge.age != "interstellar_age" || ge.Buildings.graceAge != "interstellar_age" {
		t.Fatalf("loaded in %s with grace age %q, want the Interstellar Age and its grace", ge.age, ge.Buildings.graceAge)
	}
	for _, k := range content3FixtureTechs {
		if !ge.Research.IsResearched(k) {
			t.Errorf("%s is no longer researched", k)
		}
	}
	for _, k := range []string{"stellar_engineering", "interstellar_trade", "antimatter_synthesis"} {
		if ge.Research.IsResearched(k) {
			t.Fatalf("%s is researched: the save does not test the grace", k)
		}
	}
	if got := grantedList(ge); got != "" {
		t.Errorf("granted on load: %q, want nothing: the save never ran the route", got)
	}
	n, line := treeNotices(ge)
	if n != 1 || !strings.Contains(line, "the tech tree is complete") || !strings.Contains(line, "from the Modern Age to the Transcendent") ||
		!strings.Contains(line, "Warp Commerce") || !strings.Contains(line, "Interstellar Age") ||
		!strings.Contains(line, "the locks start when you next advance") {
		t.Errorf("%d research update lines, the last %q; want one about the finished tree that names the Interstellar Age", n, line)
	}
	// What stood, what was going up and what was planned are as they were.
	// (The time away goes on building, so how far along the hub is depends
	// on the day the test runs.)
	if got := ge.Buildings.GetCount("pulsar_tap"); got != 2 {
		t.Errorf("%d Pulsar Taps after the load, want the 2 that stood", got)
	}
	if got := ge.Buildings.GetCount("galactic_trade_hub") + queued(ge, "galactic_trade_hub"); got != 1 {
		t.Errorf("%d Galactic Trade Hubs standing or going up, want the one the save was building", got)
	}
	if len(ge.plan) != 1 || ge.plan[0].Key != "galactic_network_node" || ge.plan[0].Count != 2 {
		t.Errorf("the plan is %+v, want two Galactic Network Nodes", ge.plan)
	}
	// In the grace age every lock is open and no building waits for a tech.
	for key, f := range ge.GetState().Features {
		if !f.Open {
			t.Errorf("in the save's age %s is shut", key)
		}
	}
	rich := func() {
		t.Helper()
		ge.mu.Lock()
		for res, r := range ge.Resources.resources {
			ge.Resources.UnlockResource(res)
			r.Storage, r.Amount = 1e30, 1e29
		}
		ge.mu.Unlock()
	}
	rich()
	if bs := ge.GetState().Buildings["pulsar_tap"]; bs.NeedsTech != "" {
		t.Errorf("in the grace age the Pulsar Tap reads %+v, want it open", bs)
	}
	if err := ge.BuildBuilding("pulsar_tap"); err != nil {
		t.Errorf("a Pulsar Tap in the grace age, without Stellar Engineering: %v", err)
	}
	if got := ge.Buildings.GetCount("pulsar_tap") + queued(ge, "pulsar_tap"); got != 3 {
		t.Errorf("%d Pulsar Taps standing or going up, want 3", got)
	}
	if err := ge.StartTradeRoute("warp_commerce"); err != nil {
		t.Errorf("Warp Commerce in the grace age, without Interstellar Trade: %v", err)
	}
	if err := ge.StopTradeRoute("warp_commerce"); err != nil {
		t.Fatal(err)
	}

	// The next age: the grace is over. The run used the route in its grace
	// age, so it keeps it; a building of the new age waits for its tech.
	if err := ge.EnterAgeForTest("galactic_age"); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.graceAge != "" {
		t.Fatalf("after the advance the grace age is still %q", ge.Buildings.graceAge)
	}
	if got := ge.Buildings.GetCount("pulsar_tap") + queued(ge, "pulsar_tap"); got != 3 {
		t.Errorf("%d Pulsar Taps after the advance, want the 3 the run built", got)
	}
	rich()
	if bs := ge.GetState().Buildings["quasar_tap"]; bs.NeedsTech != "antimatter_synthesis" {
		t.Errorf("the Quasar Tap reads %+v, want it to wait for Antimatter Synthesis", bs)
	}
	if err := ge.BuildBuilding("quasar_tap"); err == nil || !strings.Contains(err.Error(), "Antimatter Synthesis") {
		t.Errorf("a Quasar Tap without Antimatter Synthesis: %v", err)
	}
	if err := ge.StartTradeRoute("warp_commerce"); err != nil {
		t.Errorf("Warp Commerce after the grace age, in a run that ran it there: %v", err)
	}

	// A second load: nothing is said twice, and the save is on today's tree.
	again := saveAndLoad(t, ge, "tree_v4_interstellar_again")
	if v, grace := savedTree(t, "tree_v4_interstellar_again"); v != config.TechTreeVersion || grace != "" {
		t.Errorf("the re-saved game carries tree version %d and grace age %q, want %d and none", v, grace, config.TechTreeVersion)
	}
	if n, _ := treeNotices(again); n != 0 {
		t.Errorf("a second load said it again (%d lines)", n)
	}
}

// TestContent3OldSaveNeverUsedTheRoute: the same save, taken into the next
// age without ever starting Warp Commerce. The lock is live there: the route
// is refused and names Interstellar Trade, and opens once it is researched.
func TestContent3OldSaveNeverUsedTheRoute(t *testing.T) {
	ge := loadContentFixture(t, content3FixtureFresh, "tree_v4_interstellar")
	if err := ge.EnterAgeForTest("galactic_age"); err != nil {
		t.Fatal(err)
	}
	stock(ge, 0)
	if err := ge.StartTradeRoute("warp_commerce"); err == nil || err.Error() != "The Warp Commerce route needs Interstellar Trade first. Research it to start it." {
		t.Errorf("Warp Commerce after the grace age, never used: %v", err)
	}
	learn(ge, "interstellar_trade")
	if err := ge.StartTradeRoute("warp_commerce"); err != nil {
		t.Errorf("Warp Commerce with Interstellar Trade researched: %v", err)
	}
}

// TestContent3OldSaveKeepsItsRoute: the save with the Warp Commerce route
// running. It keeps the route for the rest of its run, through the next age
// and a second load, without Interstellar Trade: it can stop it and start
// it again. A new run meets the lock.
func TestContent3OldSaveKeepsItsRoute(t *testing.T) {
	ge := loadContentFixture(t, content3FixtureInUse, "tree_v4_interstellar_route")
	// A running route is trade routes in use as well as its own lock.
	if got := grantedList(ge); got != "route_warp_commerce trade_routes" {
		t.Errorf("granted on load: %q, want the route's lock and trade routes", got)
	}
	if ge.Research.IsResearched("interstellar_trade") {
		t.Fatal("Interstellar Trade is researched: the save does not test the grant")
	}
	running := func(ge *GameEngine) bool {
		ge.mu.RLock()
		defer ge.mu.RUnlock()
		return ge.Trade.activeRoutes["warp_commerce"] != nil
	}
	if !running(ge) {
		t.Fatal("the route the save was running is not running")
	}
	if err := ge.EnterAgeForTest("galactic_age"); err != nil {
		t.Fatal(err)
	}
	stock(ge, 0)
	if err := ge.StopTradeRoute("warp_commerce"); err != nil {
		t.Fatalf("stopping Warp Commerce after the advance: %v", err)
	}
	if err := ge.StartTradeRoute("warp_commerce"); err != nil {
		t.Errorf("Warp Commerce after the grace age, in a run that was using it: %v", err)
	}
	again := saveAndLoad(t, ge, "tree_v4_interstellar_route_again")
	if got := grantedList(again); got != "route_warp_commerce trade_routes" {
		t.Errorf("granted after a second load: %q, want the route's lock and trade routes", got)
	}
	if !running(again) {
		t.Error("the route is not running after the second load")
	}
	if n, _ := treeNotices(again); n != 0 {
		t.Errorf("a second load said it again (%d lines)", n)
	}

	// A new run meets the lock.
	fresh := newSeededEngine(62)
	fresh.mu.Lock()
	def := fresh.Trade.routeDefs["warp_commerce"]
	fresh.Buildings.counts[def.RequiredBld] = def.MinCount
	fresh.age = def.MinAge
	fresh.mu.Unlock()
	learn(fresh, "the_wheel")
	stock(fresh, 0)
	if err := fresh.StartTradeRoute("warp_commerce"); err == nil || !strings.Contains(err.Error(), "Interstellar Trade") {
		t.Errorf("Warp Commerce in a new run, without Interstellar Trade: %v", err)
	}
}
