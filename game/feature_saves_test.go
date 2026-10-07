package game

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
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

// loadFeatureFixture loads a pre-feature-lock save and checks that the
// signature the old game wrote still verifies: the granted features are not
// written when empty, and no saved key was renamed.
func loadFeatureFixture(t *testing.T, path, name string) *GameEngine {
	t.Helper()
	isolateAccountDir(t)
	ge := loadFixture(t, path, name)
	if ge.cheaterBadge {
		t.Fatalf("%s failed its signature check", path)
	}
	return ge
}

// savedFeatures reads the granted features back from a save file.
func savedFeatures(t *testing.T, name string) []string {
	t.Helper()
	data, err := os.ReadFile(savePath(name))
	if err != nil {
		t.Fatal(err)
	}
	var gs GameSave
	if err := json.Unmarshal(data, &gs); err != nil {
		t.Fatal(err)
	}
	return gs.GrantedFeatures
}

// stock gives the game plenty of everything a command might ask for.
func stock(ge *GameEngine, soldiers float64) {
	for _, res := range []string{"food", "wood", "stone", "iron", "gold", "culture", "steel", "coal"} {
		setResource(ge, res, 1e9)
	}
	setSoldiers(ge, soldiers)
}

// clearArmy calls every expedition home, so the next can be sent.
func clearArmy(ge *GameEngine) {
	ge.mu.Lock()
	ge.Military.activeByCat = map[string]*ActiveExpedition{}
	ge.mu.Unlock()
}

// grantedList is the run's granted features, sorted and joined.
func grantedList(ge *GameEngine) string {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return strings.Join(ge.grantedFeatureKeys(), " ")
}

// TestFeatureOldSaveKeepsWhatItUses is the grace an old save gets, on a game
// in the middle of using commands that now wait for a tech. The Bronze Age
// save has a campaign under way and no Military Tactics. It loads with its
// age exempt and with every command it was using granted for the run: one
// log line says so, once. The lock never reaches this run's campaigns, in
// the next age either, through a save and a load. A new run meets it.
func TestFeatureOldSaveKeepsWhatItUses(t *testing.T) {
	ge := loadFeatureFixture(t, featureFixtureInUse, "pre_features_bronze")
	if ge.age != "bronze_age" || ge.Buildings.graceAge != "bronze_age" {
		t.Fatalf("loaded in %s with grace age %q, want the Bronze Age and its grace", ge.age, ge.Buildings.graceAge)
	}
	for _, k := range featureFixtureTechs {
		if !ge.Research.IsResearched(k) {
			t.Errorf("%s is no longer researched", k)
		}
	}
	if ge.Research.IsResearched("military_tactics") {
		t.Fatal("Military Tactics is researched: the save does not test the grant")
	}
	// What the save was using: a campaign, a route, a festival, an ally.
	if got, want := grantedList(ge), "campaigns diplomacy festivals trade_routes"; got != want {
		t.Errorf("granted on load: %q, want %q", got, want)
	}
	// One line, in the words for a save that already knew the keystones.
	n, line := treeNotices(ge)
	if n != 1 || !strings.Contains(line, "Bronze Age") || !strings.Contains(line, "the locks start when you next advance") ||
		!strings.Contains(line, "some commands now wait for a tech") || strings.Contains(line, "keystone") {
		t.Errorf("%d research update lines, the last %q; want one about commands that names the Bronze Age", n, line)
	}

	// In the save's age a campaign goes out without the tech.
	stock(ge, 50)
	clearArmy(ge)
	if err := ge.LaunchExpedition("raid_bandits"); err != nil {
		t.Fatalf("a campaign in the save's age: %v", err)
	}

	// The next age ends the grace, and the campaign the run was already
	// sending stays open.
	if err := ge.EnterAgeForTest("iron_age"); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.graceAge != "" {
		t.Fatalf("after the advance the grace age is still %q", ge.Buildings.graceAge)
	}
	stock(ge, 50)
	clearArmy(ge)
	if err := ge.LaunchExpedition("conquer_territory"); err != nil {
		t.Errorf("a campaign in the next age, in a run that was already sending them: %v", err)
	}
	if f := ge.GetState().Features[config.FeatureCampaigns]; !f.Live || !f.Open || !f.Granted || f.TechName != "Military Tactics" {
		t.Errorf("campaigns read %+v, want a live lock, open and granted", f)
	}

	// A second load: the grants are in the save, nothing is said twice.
	again := saveAndLoad(t, ge, "pre_features_bronze_again")
	if got, want := strings.Join(savedFeatures(t, "pre_features_bronze_again"), " "), "campaigns diplomacy festivals trade_routes"; got != want {
		t.Errorf("the re-saved game carries granted features %q, want %q", got, want)
	}
	if v, grace := savedTree(t, "pre_features_bronze_again"); v != config.TechTreeVersion || grace != "" {
		t.Errorf("the re-saved game carries tree version %d and grace age %q, want %d and none", v, grace, config.TechTreeVersion)
	}
	if n, _ := treeNotices(again); n != 0 {
		t.Errorf("a second load said it again (%d lines)", n)
	}
	stock(again, 50)
	clearArmy(again)
	if err := again.LaunchExpedition("conquer_territory"); err != nil {
		t.Errorf("a campaign after a second load: %v", err)
	}

	// The grant is the run's. A new run meets the lock in the Bronze Age.
	again.Reset()
	if got := grantedList(again); got != "" {
		t.Errorf("a new run kept the granted features %q", got)
	}
	for _, age := range []string{"stone_age", "bronze_age"} {
		if err := again.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	stock(again, 50)
	err := again.LaunchExpedition("raid_bandits")
	if err == nil || err.Error() != "Campaigns need Military Tactics first. Research it to send one." {
		t.Errorf("a campaign on a new run without Military Tactics: %v", err)
	}
	learn(again, "military_tactics")
	if err := again.LaunchExpedition("raid_bandits"); err != nil {
		t.Errorf("a campaign with Military Tactics researched: %v", err)
	}
}

// TestFeatureOldSaveLocksWhatItNeverUsed: the grant covers what the save was
// using and nothing else. The Industrial Age save runs Rail Freight and has
// dealt with the smugglers, and has sent nothing from its Army panel. In its
// own age every command is open. From the next age on it keeps its route
// and its black market without Railroads or Mercantilism, and the Naval
// Expedition, which it never sent, waits for Navigation.
func TestFeatureOldSaveLocksWhatItNeverUsed(t *testing.T) {
	ge := loadFeatureFixture(t, featureFixtureMixed, "pre_features_industrial")
	if ge.age != "industrial_age" || ge.Buildings.graceAge != "industrial_age" {
		t.Fatalf("loaded in %s with grace age %q, want the Industrial Age and its grace", ge.age, ge.Buildings.graceAge)
	}
	for _, k := range []string{"railroads", "mercantilism", "navigation", "military_tactics"} {
		if ge.Research.IsResearched(k) {
			t.Fatalf("%s is researched: the save does not test the locks", k)
		}
	}
	if got, want := grantedList(ge), "black_market route_rail_freight trade_routes"; got != want {
		t.Errorf("granted on load: %q, want %q", got, want)
	}
	if n, line := treeNotices(ge); n != 1 || !strings.Contains(line, "Industrial Age") {
		t.Errorf("%d research update lines, the last %q; want one naming the Industrial Age", n, line)
	}
	// In the grace age every lock is open, used or not.
	for key, f := range ge.GetState().Features {
		if !f.Open {
			t.Errorf("in the save's age %s is shut", key)
		}
	}

	if err := ge.EnterAgeForTest("victorian_age"); err != nil {
		t.Fatal(err)
	}
	stock(ge, 50)
	clearArmy(ge)
	// Never sent: the Naval Expedition and campaigns wait for their techs.
	err := ge.LaunchExpedition("naval_expedition")
	if err == nil || err.Error() != "The Naval Expedition needs Navigation first. Research it to send it." {
		t.Errorf("the Naval Expedition in the next age, never sent before: %v", err)
	}
	err = ge.LaunchExpedition("colonial_campaign")
	if err == nil || err.Error() != "Campaigns need Military Tactics first. Research it to send one." {
		t.Errorf("a campaign in the next age, never sent before: %v", err)
	}
	// In use: the route stops and starts again, and the smugglers deal.
	if err := ge.StopTradeRoute("rail_freight"); err != nil {
		t.Fatalf("stopping Rail Freight: %v", err)
	}
	if err := ge.StartTradeRoute("rail_freight"); err != nil {
		t.Errorf("starting Rail Freight again without Railroads, in a run that was running it: %v", err)
	}
	ge.mu.Lock()
	ge.blackMarketReadyTick = 0
	ge.mu.Unlock()
	if _, _, err := ge.DoBlackMarket("gold"); err != nil {
		t.Errorf("the black market without Mercantilism, in a run that had used it: %v", err)
	}
	// Researching the tech opens the rest.
	learn(ge, "navigation")
	if err := ge.LaunchExpedition("naval_expedition"); err != nil {
		t.Errorf("the Naval Expedition with Navigation researched: %v", err)
	}

	// The naval voyage also used the lock on expeditions past the Scout
	// Party, which waits for a tech not in the tree yet: remembered too.
	again := saveAndLoad(t, ge, "pre_features_industrial_again")
	if got, want := grantedList(again), "black_market expeditions route_rail_freight trade_routes"; got != want {
		t.Errorf("after a second load: granted %q, want %q", got, want)
	}
	if n, _ := treeNotices(again); n != 0 {
		t.Errorf("a second load said it again (%d lines)", n)
	}
}

// TestFeatureGraceCoversASaveFromBeforeTheTree: a save from before the tree
// had any lock (the keystone fixtures) gets the same age of grace for the
// commands as for its wonder, and one line that says both.
func TestFeatureGraceCoversASaveFromBeforeTheTree(t *testing.T) {
	ge := loadKeystoneFixture(t, keystoneFixtureBanked, "pre_keystone_iron")
	n, line := treeNotices(ge)
	if n != 1 || !strings.Contains(line, "keystone") || !strings.Contains(line, "some commands wait for a tech") || !strings.Contains(line, "Iron Age") {
		t.Errorf("%d research update lines, the last %q; want one about the keystone and the commands", n, line)
	}
	stock(ge, 50)
	clearArmy(ge)
	if err := ge.LaunchExpedition("raid_bandits"); err != nil {
		t.Errorf("a campaign in the grace age of a save from before the tree: %v", err)
	}
	// Sent in the grace age, so the run keeps it.
	if err := ge.EnterAgeForTest("classical_age"); err != nil {
		t.Fatal(err)
	}
	clearArmy(ge)
	stock(ge, 50)
	if err := ge.LaunchExpedition("raid_bandits"); err != nil {
		t.Errorf("a campaign after the grace age, in a run that sent one during it: %v", err)
	}
}

// TestFeatureLocksRefuseAndOpen: each live lock refuses its command with the
// name of its tech and opens when the tech is researched; the Scout Party
// and the locks whose tech is not in the tree yet refuse nothing; and a new
// game saves no granted feature until it uses a command without its tech.
func TestFeatureLocksRefuseAndOpen(t *testing.T) {
	isolateAccountDir(t)
	at := func(age string) *GameEngine {
		ge := newSeededEngine(5)
		for _, a := range ageKeys() {
			if a == "primitive_age" {
				continue
			}
			if err := ge.EnterAgeForTest(a); err != nil {
				t.Fatal(err)
			}
			if a == age {
				break
			}
		}
		stock(ge, 100)
		return ge
	}

	// Campaigns, from the Bronze Age. The Scout Party walks out regardless.
	ge := at("bronze_age")
	if err := ge.LaunchExpedition("scout_party"); err != nil {
		t.Errorf("the Scout Party: %v", err)
	}
	if err := ge.LaunchExpedition("raid_bandits"); err == nil || err.Error() != "Campaigns need Military Tactics first. Research it to send one." {
		t.Errorf("a campaign without Military Tactics: %v", err)
	}
	learn(ge, "military_tactics")
	if err := ge.LaunchExpedition("raid_bandits"); err != nil {
		t.Errorf("a campaign with Military Tactics: %v", err)
	}
	// Researched, so nothing is granted: the save stays as it was.
	if got := grantedList(ge); got != "" {
		t.Errorf("a campaign sent with its tech researched granted %q", got)
	}
	// A lock whose tech is not in the tree is open, and using its command
	// is remembered for the day the tech arrives.
	if err := ge.StartTradeRoute("local_barter"); err == nil || strings.Contains(err.Error(), "first. Research it") {
		t.Errorf("Local Barter with no market: %v, want the route's own refusal", err)
	}
	ge.mu.Lock()
	ge.Buildings.counts[ge.Trade.routeDefs["local_barter"].RequiredBld] = ge.Trade.routeDefs["local_barter"].MinCount
	ge.mu.Unlock()
	if err := ge.StartTradeRoute("local_barter"); err != nil {
		t.Errorf("Local Barter, whose lock waits for The Wheel: %v", err)
	}
	if got := grantedList(ge); got != "trade_routes" {
		t.Errorf("a route started under an inert lock granted %q, want trade_routes", got)
	}
	if f := ge.GetState().Features[config.FeatureTradeRoutes]; f.Live || !f.Open || !f.Granted {
		t.Errorf("trade routes read %+v, want an inert lock, open, granted", f)
	}
	saveAndLoad(t, ge, "feature_lock_fresh")
	if got := strings.Join(savedFeatures(t, "feature_lock_fresh"), " "); got != "trade_routes" {
		t.Errorf("the save carries granted features %q, want trade_routes", got)
	}

	// The Naval Expedition, from the Renaissance Age.
	ge = at("renaissance_age")
	if err := ge.LaunchExpedition("naval_expedition"); err == nil || err.Error() != "The Naval Expedition needs Navigation first. Research it to send it." {
		t.Errorf("the Naval Expedition without Navigation: %v", err)
	}
	// The other scouting expeditions wait only for Exploration, which is
	// not in the tree yet.
	if err := ge.LaunchExpedition("scout_ruins"); err != nil {
		t.Errorf("Scout Nearby Ruins: %v", err)
	}
	clearArmy(ge)
	learn(ge, "navigation")
	if err := ge.LaunchExpedition("naval_expedition"); err != nil {
		t.Errorf("the Naval Expedition with Navigation: %v", err)
	}

	// The black market: its age first, then its tech.
	if _, _, err := ge.DoBlackMarket("gold"); err == nil || !strings.HasPrefix(err.Error(), "The black market opens in") {
		t.Errorf("the black market in the Renaissance Age: %v, want its age named", err)
	}
	ge = at("colonial_age")
	if _, _, err := ge.DoBlackMarket("gold"); err == nil || err.Error() != "The black market needs Mercantilism first. Research it to deal there." {
		t.Errorf("the black market without Mercantilism: %v", err)
	}
	learn(ge, "mercantilism")
	if _, _, err := ge.DoBlackMarket("gold"); err != nil {
		t.Errorf("the black market with Mercantilism: %v", err)
	}

	// Rail Freight, from the Industrial Age: the route's own lock.
	ge = at("industrial_age")
	ge.mu.Lock()
	ge.Buildings.counts["iron_works_complex"] = 1
	ge.mu.Unlock()
	if err := ge.StartTradeRoute("rail_freight"); err == nil || err.Error() != "The Rail Freight route needs Railroads first. Research it to start it." {
		t.Errorf("Rail Freight without Railroads: %v", err)
	}
	learn(ge, "railroads")
	if err := ge.StartTradeRoute("rail_freight"); err != nil {
		t.Errorf("Rail Freight with Railroads: %v", err)
	}
	// A festival and a gift wait for techs that are not in the tree yet.
	if err := ge.DoFestival(); err != nil {
		t.Errorf("a festival, whose lock waits for Drama: %v", err)
	}
}

// TestFeatureLockSwitchesOnWithItsTech: a lock that waits for its tech
// switches on when a ruleset adds a tech with that key, with no other
// change. A new game on that ruleset meets the lock. A game that was
// already using the command, saved before the tech existed, keeps it for
// the rest of its run.
func TestFeatureLockSwitchesOnWithItsTech(t *testing.T) {
	isolateAccountDir(t)
	src := rules.FromConfig()
	src.Techs = append(src.Techs, config.TechDef{
		Key: "the_wheel", Name: "The Wheel", Age: "bronze_age", Lane: config.LaneCraft, Cost: 10, ResearchTicks: 4,
	})
	withWheel := rules.Compile(src)
	market := func(ge *GameEngine) {
		ge.mu.Lock()
		def := ge.Trade.routeDefs["local_barter"]
		ge.Buildings.counts[def.RequiredBld] = def.MinCount
		ge.age = def.MinAge
		ge.mu.Unlock()
	}

	// Today's tree: the route starts, and the save remembers it.
	old := newSeededEngine(6)
	market(old)
	if err := old.StartTradeRoute("local_barter"); err != nil {
		t.Fatalf("Local Barter on today's tree: %v", err)
	}
	if err := old.SaveGame("before_the_wheel"); err != nil {
		t.Fatal(err)
	}

	// The tree with The Wheel: a new game is refused until it researches it.
	fresh := NewGameEngineWith(withWheel)
	market(fresh)
	if err := fresh.StartTradeRoute("local_barter"); err == nil || err.Error() != "Trade routes need The Wheel first. Research it to start one." {
		t.Errorf("Local Barter on a tree with The Wheel, not researched: %v", err)
	}
	learn(fresh, "the_wheel")
	if err := fresh.StartTradeRoute("local_barter"); err != nil {
		t.Errorf("Local Barter with The Wheel researched: %v", err)
	}

	// The game saved before The Wheel existed, loaded on the new tree: its
	// route is running, and it can stop it and start it again.
	loaded := NewGameEngineWith(withWheel)
	if err := loaded.LoadGame("before_the_wheel"); err != nil {
		t.Fatal(err)
	}
	if loaded.Research.IsResearched("the_wheel") {
		t.Fatal("The Wheel is researched: the save does not test the grant")
	}
	if err := loaded.StopTradeRoute("local_barter"); err != nil {
		t.Fatalf("stopping Local Barter after the load: %v", err)
	}
	if err := loaded.StartTradeRoute("local_barter"); err != nil {
		t.Errorf("Local Barter in a game that ran it before The Wheel existed: %v", err)
	}
}

// TestResearchSaysWhatItOpens: finishing a tech that opens a command logs
// it, in the game's words.
func TestResearchSaysWhatItOpens(t *testing.T) {
	ge := newSeededEngine(8)
	if err := ge.EnterAgeForTest("stone_age"); err != nil {
		t.Fatal(err)
	}
	if err := ge.EnterAgeForTest("bronze_age"); err != nil {
		t.Fatal(err)
	}
	ge.mu.Lock()
	ge.Research.currentTech, ge.Research.ticksLeft, ge.Research.totalTicks = "military_tactics", 1, 1
	ge.processResearch()
	ge.mu.Unlock()
	if !ge.Research.IsResearched("military_tactics") || !logHas(ge, "Military Tactics opens campaigns.") {
		t.Errorf("finishing Military Tactics did not say it opens campaigns; the log ends %q", lastLog(ge))
	}
}
