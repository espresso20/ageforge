package game

import (
	"os"
	"strings"
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

// loadContentFixture loads a tree version 2 save and checks that the
// signature the old game wrote still verifies.
func loadContentFixture(t *testing.T, path, name string) *GameEngine {
	t.Helper()
	isolateAccountDir(t)
	ge := loadFixture(t, path, name)
	if ge.cheaterBadge {
		t.Fatalf("%s failed its signature check", path)
	}
	return ge
}

// TestContentOldSaveGetsAnAgeOfGrace: a game saved under tree version 2 that
// was using none of the commands the new techs open. The Bronze Age save
// has Stonehenge's bank half full and no Calendar. It loads with its age
// exempt: one log line says so, Stonehenge builds without its keystone, the
// Altar builds without Calendar and a trade route would start without The
// Wheel. The next age ends the grace: what the run never used now waits
// for its tech, and nothing researched, banked or built was touched.
func TestContentOldSaveGetsAnAgeOfGrace(t *testing.T) {
	ge := loadContentFixture(t, contentFixtureFresh, "tree_v2_bronze")
	if ge.age != "bronze_age" || ge.Buildings.graceAge != "bronze_age" {
		t.Fatalf("loaded in %s with grace age %q, want the Bronze Age and its grace", ge.age, ge.Buildings.graceAge)
	}
	for _, k := range contentFixtureTechs {
		if !ge.Research.IsResearched(k) {
			t.Errorf("%s is no longer researched", k)
		}
	}
	for _, k := range []string{"calendar", "the_wheel", "exploration", "envoys", "drama"} {
		if ge.Research.IsResearched(k) {
			t.Fatalf("%s is researched: the save does not test the grace", k)
		}
	}
	if got := grantedList(ge); got != "" {
		t.Errorf("granted on load: %q, want nothing: the save used none of these commands", got)
	}
	n, line := treeNotices(ge)
	if n != 1 || !strings.Contains(line, "the tech tree gained new techs") || !strings.Contains(line, "Bronze Age") ||
		!strings.Contains(line, "the locks start when you next advance") || strings.Contains(line, "smaller") {
		t.Errorf("%d research update lines, the last %q; want one about the new techs that names the Bronze Age", n, line)
	}
	// The bank is as the old game left it: half of every part.
	for res, need := range ge.Buildings.defs["stonehenge"].BaseCost {
		if got := ge.Buildings.wonderBanks["stonehenge"][res]; got != need/2 {
			t.Errorf("Stonehenge's bank holds %v %s, want half of %v", got, res, need)
		}
	}
	// In the grace age every lock is open.
	for key, f := range ge.GetState().Features {
		if !f.Open {
			t.Errorf("in the save's age %s is shut", key)
		}
	}
	stock(ge, 0)
	if bs := ge.GetState().Buildings["altar"]; bs.NeedsTech != "" && !bs.Unlocked {
		t.Errorf("the Altar in the grace age: %+v, want open", bs)
	}
	if err := ge.BuildBuilding("altar"); err != nil {
		t.Errorf("an Altar in the grace age, without Calendar: %v", err)
	}
	fillWonderBank(t, ge, "stonehenge")
	if err := ge.BuildBuilding("stonehenge"); err != nil {
		t.Fatalf("Stonehenge in the grace age, without Calendar: %v", err)
	}

	// The next age: the grace is over, and what was never used is locked.
	if err := ge.EnterAgeForTest("iron_age"); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.graceAge != "" {
		t.Fatalf("after the advance the grace age is still %q", ge.Buildings.graceAge)
	}
	if ge.Buildings.GetCount("stonehenge")+queued(ge, "stonehenge") != 1 || ge.Buildings.GetCount("altar")+queued(ge, "altar") != 1 {
		t.Error("what was built in the grace age did not stay")
	}
	stock(ge, 0)
	ge.mu.Lock()
	ge.Buildings.counts[ge.Trade.routeDefs["local_barter"].RequiredBld] = ge.Trade.routeDefs["local_barter"].MinCount
	ge.mu.Unlock()
	for what, c := range map[string]struct {
		err  error
		want string
	}{
		"a trade route, never started before": {ge.StartTradeRoute("local_barter"), "Trade routes need The Wheel first. Research it to start one."},
		"scouting past the Scout Party":       {ge.LaunchExpedition("scout_ruins"), "Expeditions past the Scout Party need Exploration first. Research it to send one."},
		"a festival":                          {ge.DoFestival(), "Festivals need Drama first. Research it to hold one."},
	} {
		if c.err == nil || c.err.Error() != c.want {
			t.Errorf("%s, after the grace age: %v, want %q", what, c.err, c.want)
		}
	}
	// A building of the new age waits for its tech like any other.
	if bs := ge.GetState().Buildings["smelter"]; bs.NeedsTech != "iron_smelting" {
		t.Errorf("the Smelter reads %+v, want it to wait for Iron Smelting", bs)
	}
	if err := ge.BuildBuilding("smelter"); err == nil || !strings.Contains(err.Error(), "Iron Smelting") {
		t.Errorf("a Smelter without Iron Smelting: %v", err)
	}
	learn(ge, "the_wheel")
	if err := ge.StartTradeRoute("local_barter"); err != nil {
		t.Errorf("a trade route with The Wheel researched: %v", err)
	}

	// A second load: nothing is said twice, and the save is on today's tree.
	again := saveAndLoad(t, ge, "tree_v2_bronze_again")
	if v, grace := savedTree(t, "tree_v2_bronze_again"); v != config.TechTreeVersion || grace != "" {
		t.Errorf("the re-saved game carries tree version %d and grace age %q, want %d and none", v, grace, config.TechTreeVersion)
	}
	if n, _ := treeNotices(again); n != 0 {
		t.Errorf("a second load said it again (%d lines)", n)
	}
}

// TestContentOldSaveKeepsWhatItUsed: a game saved under tree version 2 that
// was trading on a route and had an ally. Its save says so itself. It keeps
// both for the rest of its run, through the next age and a second load,
// without The Wheel or Envoys; what it never used waits for its tech; and
// its next run meets every lock.
func TestContentOldSaveKeepsWhatItUsed(t *testing.T) {
	ge := loadContentFixture(t, contentFixtureInUse, "tree_v2_iron")
	if ge.age != "iron_age" || ge.Buildings.graceAge != "iron_age" {
		t.Fatalf("loaded in %s with grace age %q, want the Iron Age and its grace", ge.age, ge.Buildings.graceAge)
	}
	if got, want := grantedList(ge), "diplomacy trade_routes"; got != want {
		t.Errorf("granted on load: %q, want %q", got, want)
	}
	if n, line := treeNotices(ge); n != 1 || !strings.Contains(line, "Iron Age") || !strings.Contains(line, "A command you were already using stays open") {
		t.Errorf("%d research update lines, the last %q; want one naming the Iron Age and what stays open", n, line)
	}
	if _, running := ge.Trade.activeRoutes["local_barter"]; !running {
		t.Fatal("the route is no longer running")
	}
	if f := ge.Diplomacy.factions["riverlands_tribes"]; f == nil || f.Status != "allied" {
		t.Fatalf("the alliance did not load: %+v", f)
	}

	if err := ge.EnterAgeForTest("classical_age"); err != nil {
		t.Fatal(err)
	}
	stock(ge, 0)
	// In use: the route stops and starts again, and the ally takes a gift.
	if err := ge.StopTradeRoute("local_barter"); err != nil {
		t.Fatalf("stopping the route: %v", err)
	}
	if err := ge.StartTradeRoute("local_barter"); err != nil {
		t.Errorf("starting the route again without The Wheel, in a run that was running it: %v", err)
	}
	if err := ge.SendGift("riverlands_tribes"); err != nil {
		t.Errorf("a gift without Envoys, in a run that had an ally: %v", err)
	}
	if f := ge.GetState().Features[config.FeatureTradeRoutes]; !f.Live || !f.Open || !f.Granted || f.TechName != "The Wheel" {
		t.Errorf("trade routes read %+v, want a live lock, open and granted", f)
	}
	// Never used: festivals and scouting past the Scout Party wait.
	if err := ge.DoFestival(); err == nil || err.Error() != "Festivals need Drama first. Research it to hold one." {
		t.Errorf("a festival, never held before: %v", err)
	}
	if err := ge.LaunchExpedition("scout_ruins"); err == nil || err.Error() != "Expeditions past the Scout Party need Exploration first. Research it to send one." {
		t.Errorf("scouting past the Scout Party, never sent before: %v", err)
	}

	again := saveAndLoad(t, ge, "tree_v2_iron_again")
	if got, want := strings.Join(savedFeatures(t, "tree_v2_iron_again"), " "), "diplomacy trade_routes"; got != want {
		t.Errorf("the re-saved game carries granted features %q, want %q", got, want)
	}
	if n, _ := treeNotices(again); n != 0 {
		t.Errorf("a second load said it again (%d lines)", n)
	}
	if err := again.SendGift("riverlands_tribes"); err != nil && !strings.Contains(err.Error(), "gold") && !strings.Contains(err.Error(), "recently") {
		t.Errorf("a gift after a second load: %v", err)
	}

	// The grant is the run's. A new run meets the locks.
	again.Reset()
	if got := grantedList(again); got != "" {
		t.Errorf("a new run kept the granted features %q", got)
	}
	for _, age := range []string{"stone_age", "bronze_age"} {
		if err := again.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	stock(again, 0)
	again.mu.Lock()
	again.Buildings.counts[again.Trade.routeDefs["local_barter"].RequiredBld] = again.Trade.routeDefs["local_barter"].MinCount
	again.mu.Unlock()
	if err := again.StartTradeRoute("local_barter"); err == nil || err.Error() != "Trade routes need The Wheel first. Research it to start one." {
		t.Errorf("a trade route on a new run without The Wheel: %v", err)
	}
}
