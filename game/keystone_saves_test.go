package game

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
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

// loadKeystoneFixture loads a pre-keystone save and checks that the
// signature the old game wrote still verifies: the tree's version and grace
// age are not written when empty, and no saved key was renamed.
func loadKeystoneFixture(t *testing.T, path, name string) *GameEngine {
	t.Helper()
	isolateAccountDir(t)
	ge := loadFixture(t, path, name)
	if ge.cheaterBadge {
		t.Fatalf("%s failed its signature check", path)
	}
	return ge
}

// treeNotices counts the log lines a save from before the lock gets on its
// first load.
func treeNotices(ge *GameEngine) (n int, last string) {
	for _, l := range ge.log {
		if strings.HasPrefix(l.Message, "Research update:") {
			n, last = n+1, l.Message
		}
	}
	return n, last
}

// fillWonderBank banks everything wonder key still lacks, through the same
// deposit a player makes, from stores raised to hold it.
func fillWonderBank(t *testing.T, ge *GameEngine, key string) {
	t.Helper()
	for _, res := range sortedKeys(ge.Buildings.defs[key].BaseCost) {
		need := ge.Buildings.defs[key].BaseCost[res] - ge.Buildings.wonderBanks[key][res]
		if need <= 0.001 {
			continue
		}
		ge.Resources.UnlockResource(res)
		ge.Resources.resources[res].Storage = math.Max(ge.Resources.resources[res].Storage, need)
		setAmount(ge, res, need)
		if _, err := ge.BankWonderMax(key, res); err != nil {
			t.Fatalf("banking %s for %s: %v", res, key, err)
		}
	}
	if !ge.Buildings.IsWonderBankFull(key) {
		t.Fatalf("the %s bank is not full after banking every part", key)
	}
}

// savedTree reads the tree's fields back from a save file.
func savedTree(t *testing.T, name string) (version int, grace string) {
	t.Helper()
	data, err := os.ReadFile(savePath(name))
	if err != nil {
		t.Fatal(err)
	}
	var gs GameSave
	if err := json.Unmarshal(data, &gs); err != nil {
		t.Fatal(err)
	}
	return gs.TreeVersion, gs.TreeGraceAge
}

// TestKeystoneOldSaveKeepsItsAge is the grace an old save gets, on a game
// with its wonder half banked. The Iron Age save has no Mathematics, which
// the Colosseum needs now. It loads with its age exempt: the bank it had is
// still there, takes the rest, and the Colosseum builds without the tech.
// One log line says so, once. The lock starts at the next advance: in the
// Classical Age the Parthenon waits for Philosophy, and a save and a load
// later it still does.
func TestKeystoneOldSaveKeepsItsAge(t *testing.T) {
	ge := loadKeystoneFixture(t, keystoneFixtureBanked, "pre_keystone_iron")
	if ge.age != "iron_age" || ge.Buildings.graceAge != "iron_age" {
		t.Fatalf("loaded in %s with grace age %q, want the Iron Age and its grace", ge.age, ge.Buildings.graceAge)
	}
	// A tech already researched stays researched; the one in progress
	// finished or is still going, never cancelled.
	for _, k := range keystoneFixtureTechs {
		if !ge.Research.IsResearched(k) {
			t.Errorf("%s is no longer researched", k)
		}
	}
	if ge.Research.IsResearched("mathematics") {
		t.Fatal("Mathematics is researched: the save does not test the grace")
	}
	if !ge.Research.IsResearched("road_building") && ge.Research.currentTech != "road_building" {
		t.Errorf("Road Building, 400 ticks from done in the save, is neither researched nor in progress (researching %q)", ge.Research.currentTech)
	}
	// One line, and it names the age.
	n, line := treeNotices(ge)
	if n != 1 || !strings.Contains(line, "Iron Age") || !strings.Contains(line, "the locks start when you next advance") {
		t.Errorf("%d research update lines, the last %q; want one that names the Iron Age and says when the lock starts", n, line)
	}
	// Half the bank was in the save, and it is still there.
	def := ge.Buildings.defs["colosseum"]
	for res, need := range def.BaseCost {
		if got := ge.Buildings.wonderBanks["colosseum"][res]; got < need/2 {
			t.Errorf("the Colosseum's bank holds %v %s, the save held %v", got, res, need/2)
		}
	}
	bs := ge.GetState().Buildings["colosseum"]
	if !bs.Unlocked || bs.NeedsTech != "" {
		t.Errorf("in its grace age the Colosseum is unlocked %v and needs %q, want open with no tech", bs.Unlocked, bs.NeedsTech)
	}
	fillWonderBank(t, ge, "colosseum")
	if !ge.GetState().Buildings["colosseum"].CanBuild {
		t.Error("with a full bank, in its grace age, the Colosseum cannot be built")
	}
	if err := ge.BuildBuilding("colosseum"); err != nil {
		t.Fatalf("building the Colosseum without Mathematics in the save's age: %v", err)
	}
	if queued(ge, "colosseum") != 1 {
		t.Fatalf("the Colosseum is not under construction")
	}

	// A second load: the version is written, the grace age with it, and
	// nothing is said twice.
	again := saveAndLoad(t, ge, "pre_keystone_iron_again")
	if v, grace := savedTree(t, "pre_keystone_iron_again"); v != config.TechTreeVersion || grace != "iron_age" {
		t.Errorf("the re-saved game carries tree version %d and grace age %q, want %d and iron_age", v, grace, config.TechTreeVersion)
	}
	if again.Buildings.graceAge != "iron_age" {
		t.Errorf("after a second load the grace age is %q, want the Iron Age still", again.Buildings.graceAge)
	}
	if n, _ := treeNotices(again); n != 0 {
		t.Errorf("a second load said it again (%d lines)", n)
	}
	if queued(again, "colosseum")+again.Buildings.GetCount("colosseum") != 1 {
		t.Errorf("the Colosseum under construction did not survive the second load")
	}

	// The next advance ends the grace: the Parthenon waits for Philosophy.
	again.mu.Lock()
	again.Buildings.counts["colosseum"] = 1
	again.buildQueue = nil
	again.advanceAge("classical_age")
	again.mu.Unlock()
	if again.Buildings.graceAge != "" {
		t.Fatalf("after the advance the grace age is still %q", again.Buildings.graceAge)
	}
	if bs := again.GetState().Buildings["parthenon"]; !bs.Unlocked || bs.NeedsTech != "philosophy" {
		t.Errorf("in the Classical Age the Parthenon is unlocked %v and needs %q, want open for banking and waiting for philosophy", bs.Unlocked, bs.NeedsTech)
	}
	fillWonderBank(t, again, "parthenon")
	err := again.BuildBuilding("parthenon")
	if err == nil || err.Error() != "Parthenon needs Philosophy first. Research it to build here." {
		t.Errorf("building the Parthenon without Philosophy: %v", err)
	}
	// Saved and loaded in the new age: still locked, no new grace.
	third := saveAndLoad(t, again, "pre_keystone_iron_classical")
	if v, grace := savedTree(t, "pre_keystone_iron_classical"); v != config.TechTreeVersion || grace != "" {
		t.Errorf("after the advance the save carries tree version %d and grace age %q, want %d and none", v, grace, config.TechTreeVersion)
	}
	if third.Buildings.graceAge != "" || !third.Buildings.TechLocked("parthenon") {
		t.Errorf("after a load in the Classical Age: grace age %q, Parthenon tech-locked %v; want none and locked", third.Buildings.graceAge, third.Buildings.TechLocked("parthenon"))
	}
	learn(third, "primitive_writing", "mathematics", "philosophy")
	if err := third.BuildBuilding("parthenon"); err != nil {
		t.Errorf("building the Parthenon with Philosophy researched: %v", err)
	}
}

// TestKeystoneOldSaveFinishesItsWonder: a wonder half built stays on its
// way. The Industrial Age save has the Crystal Palace paid for and half way
// through its construction, and no Industrialization. A load runs the time
// since the save was written, which finishes it in any run later than two
// hours after the fixture was made; a sooner run finds it still under
// construction, never cancelled, and plays out the day here. Built, it
// stays built, and the age asks for no keystone.
func TestKeystoneOldSaveFinishesItsWonder(t *testing.T) {
	ge := loadKeystoneFixture(t, keystoneFixtureBuilding, "pre_keystone_industrial")
	if ge.Research.IsResearched("industrialization") {
		t.Fatal("Industrialization is researched: the save does not test the grace")
	}
	if ge.Buildings.GetCount("crystal_palace") == 0 {
		if queued(ge, "crystal_palace") != 1 {
			t.Fatalf("the Crystal Palace is neither built nor under construction after loading")
		}
		ge.SimulateOffline(MaxOfflineTime)
	}
	if ge.Buildings.GetCount("crystal_palace") != 1 || queued(ge, "crystal_palace") != 0 {
		t.Fatalf("the Crystal Palace: %d built, %d under construction; want it built", ge.Buildings.GetCount("crystal_palace"), queued(ge, "crystal_palace"))
	}
	if n, _ := treeNotices(ge); n != 1 {
		t.Errorf("%d research update lines, want one", n)
	}
	st := ge.GetState()
	if st.CurrentAgeWonderKey != "" {
		t.Errorf("the age still asks for a wonder (%s) with the Crystal Palace built", st.CurrentAgeWonderKey)
	}
	// Every earlier wonder the save had stands, whatever its keystone.
	for _, w := range []string{"great_monolith", "stonehenge", "colosseum", "parthenon", "great_library", "sistine_chapel", "grand_lighthouse"} {
		if ge.Buildings.GetCount(w) != 1 {
			t.Errorf("%s is no longer built", w)
		}
	}
	for _, k := range keystoneFixtureTechs {
		if !ge.Research.IsResearched(k) {
			t.Errorf("%s is no longer researched", k)
		}
	}
	if got := ge.Research.ResearchedCount(); got != len(keystoneFixtureTechs) {
		t.Errorf("%d techs researched, the save held %d", got, len(keystoneFixtureTechs))
	}
	again := saveAndLoad(t, ge, "pre_keystone_industrial_again")
	if again.Buildings.GetCount("crystal_palace") != 1 {
		t.Error("a second load lost the Crystal Palace")
	}
	if n, _ := treeNotices(again); n != 0 {
		t.Errorf("a second load said it again (%d lines)", n)
	}
}

// TestKeystoneOldSaveKeepsItsWonders: a wonder already built stays built.
// The Fusion Age save has the Stellar Cradle and every wonder before it,
// with none of their keystones researched. Nothing is taken down and nothing
// is asked for in that age. The age after it is held to the lock: the Dyson
// Scaffold waits for Orbital Mechanics.
func TestKeystoneOldSaveKeepsItsWonders(t *testing.T) {
	ge := loadKeystoneFixture(t, keystoneFixtureBuilt, "pre_keystone_fusion")
	wonders, unkeyed := 0, 0
	for key, def := range ge.Buildings.defs {
		if def.Category != "wonder" {
			continue
		}
		at, _ := ge.rules.Index(def.RequiredAge)
		fusion, _ := ge.rules.Index("fusion_age")
		if at > fusion {
			continue
		}
		wonders++
		if ge.Buildings.GetCount(key) != 1 {
			t.Errorf("%s is no longer built", key)
		}
		if def.RequiredTech != "" && !ge.Research.IsResearched(def.RequiredTech) {
			unkeyed++
		}
	}
	// 17 wonders stand, and 15 of them without the tech they now ask for:
	// all but the Sacred Grove, which asks for none, and the Great
	// Monolith, whose Stoneworking the save holds.
	if wonders != 17 || unkeyed != 15 {
		t.Errorf("%d wonders up to the Fusion Age, %d of them built without their keystone; want 17 and 15", wonders, unkeyed)
	}
	if got := ge.Research.ResearchedCount(); got != len(keystoneFixtureTechs) {
		t.Errorf("%d techs researched, the save held %d", got, len(keystoneFixtureTechs))
	}
	if st := ge.GetState(); st.CurrentAgeWonderKey != "" || st.Age != "fusion_age" {
		t.Errorf("in %s the age asks for the wonder %q; want the Fusion Age with nothing to build", st.Age, st.CurrentAgeWonderKey)
	}
	if n, line := treeNotices(ge); n != 1 || !strings.Contains(line, "Fusion Age") {
		t.Errorf("%d research update lines, the last %q; want one naming the Fusion Age", n, line)
	}
	if err := ge.EnterAgeForTest("space_age"); err != nil {
		t.Fatal(err)
	}
	if ge.Buildings.graceAge != "" {
		t.Errorf("the grace age is still %q after the advance", ge.Buildings.graceAge)
	}
	if bs := ge.GetState().Buildings["dyson_scaffold"]; !bs.Unlocked || bs.NeedsTech != "orbital_mechanics" || bs.CanBuild {
		t.Errorf("in the Space Age the Dyson Scaffold is unlocked %v, needs %q, can be built %v; want open, waiting for orbital_mechanics", bs.Unlocked, bs.NeedsTech, bs.CanBuild)
	}
	if !logHas(ge, "Wonder available: Dyson Scaffold. Bank its cost and research Orbital Mechanics, its keystone, then build it.") {
		t.Error("the advance did not say the Dyson Scaffold needs Orbital Mechanics")
	}
}

// TestKeystoneGraceIsOneAgeOfOneRun: the grace belongs to the age the old
// save was in, and to that run. A new game has none and writes the tree's
// version; a grace age that is not the save's age is dropped; and a new run
// (a reset here; a prestige and a Succumb rebuild the buildings the same
// way) starts without it, so coming back to that age later finds the lock.
func TestKeystoneGraceIsOneAgeOfOneRun(t *testing.T) {
	isolateAccountDir(t)
	fresh := newSeededEngine(7)
	if fresh.Buildings.graceAge != "" {
		t.Fatalf("a new game has the grace age %q", fresh.Buildings.graceAge)
	}
	loaded := saveAndLoad(t, fresh, "keystone_fresh")
	if v, grace := savedTree(t, "keystone_fresh"); v != config.TechTreeVersion || grace != "" {
		t.Errorf("a new game saves tree version %d and grace age %q, want %d and none", v, grace, config.TechTreeVersion)
	}
	if n, _ := treeNotices(loaded); n != 0 || loaded.Buildings.graceAge != "" {
		t.Errorf("a save of a new game loads with %d research update lines and grace age %q", n, loaded.Buildings.graceAge)
	}

	// A save on the current version naming another age as its grace age.
	gs := GameSave{Age: "stone_age", TreeVersion: config.TechTreeVersion, TreeGraceAge: "iron_age"}
	if fresh.graceTreeLocked(&gs) || fresh.Buildings.graceAge != "" {
		t.Errorf("a grace age that is not the save's age was kept (%q)", fresh.Buildings.graceAge)
	}
	// An older version: the save's age, whatever it claims.
	gs = GameSave{Age: "stone_age", TreeGraceAge: "iron_age"}
	if !fresh.graceTreeLocked(&gs) || fresh.Buildings.graceAge != "stone_age" {
		t.Errorf("a save from before the lock got the grace age %q, want its own age", fresh.Buildings.graceAge)
	}

	old := loadKeystoneFixture(t, keystoneFixtureBanked, "pre_keystone_iron")
	if old.Buildings.graceAge != "iron_age" {
		t.Fatalf("setup: grace age %q", old.Buildings.graceAge)
	}
	old.Reset()
	if old.Buildings.graceAge != "" {
		t.Errorf("a new run kept the grace age %q", old.Buildings.graceAge)
	}
	for _, age := range []string{"stone_age", "bronze_age", "iron_age"} {
		if err := old.EnterAgeForTest(age); err != nil {
			t.Fatal(err)
		}
	}
	if !old.Buildings.TechLocked("colosseum") {
		t.Error("back in the Iron Age on a new run, the Colosseum does not wait for Mathematics")
	}
}
