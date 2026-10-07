package game

import (
	"path/filepath"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// secondSetTech is a tech only the second ruleset of these tests defines.
const secondSetTech = "second_set_tech"

// secondSet compiles today's tables plus one tech of the first age, free and
// quick, that raises all production: a ruleset that plays differently from
// the core set from the first tick it is researched.
func secondSet() *rules.Set {
	src := rules.FromConfig()
	src.Techs = append(src.Techs, config.TechDef{
		Key: secondSetTech, Name: "Second Set Tech", Age: src.Ages[0].Key, ResearchTicks: 3,
		Effects: []config.TechEffect{{Kind: config.EffectAllOutput, Value: 0.5}},
	})
	return rules.Compile(src)
}

// managerSets is the ruleset each of an engine's managers holds, by name.
func managerSets(ge *GameEngine) map[string]*rules.Set {
	return map[string]*rules.Set{
		"Resources": ge.Resources.rules, "Buildings": ge.Buildings.rules, "Workers": ge.Workers.rules,
		"Research": ge.Research.rules, "Military": ge.Military.rules, "Events": ge.Events.rules,
		"Milestones": ge.Milestones.rules, "Prestige": ge.Prestige.rules, "Trade": ge.Trade.rules,
		"Diplomacy": ge.Diplomacy.rules, "progress": ge.progress.rules,
	}
}

func wantManagersOn(t *testing.T, ge *GameEngine, set *rules.Set, when string) {
	t.Helper()
	if ge.rules != set {
		t.Errorf("%s: the engine is not on its ruleset", when)
	}
	for name, got := range managerSets(ge) {
		if got != set {
			t.Errorf("%s: %s is not on the engine's ruleset", when, name)
		}
	}
}

// TestEnginesKeepTheirOwnRules: two engines in one process, one on the core
// set and one on a set of its own. Each knows only its own definitions, and
// the core engine plays exactly as it does alone.
func TestEnginesKeepTheirOwnRules(t *testing.T) {
	other := secondSet()
	home, away := NewGameEngine(), NewGameEngineWith(other)
	home.SeedRNG(7)
	away.SeedRNG(7)
	wantManagersOn(t, home, rules.Core(), "a new engine")
	wantManagersOn(t, away, other, "a new engine on its own set")
	if st := home.GetState(); st.Rules != rules.Core() || st.Ruleset() != rules.Core() {
		t.Error("the core engine's snapshot does not carry the core set")
	}
	if st := away.GetState(); st.Rules != other || st.Ruleset() != other {
		t.Error("the second engine's snapshot does not carry its set")
	}
	if (&GameState{}).Ruleset() != rules.Core() {
		t.Error("a snapshot written by hand does not read the core set")
	}

	if err := home.StartResearch(secondSetTech); err == nil {
		t.Errorf("the core engine started researching %s, which only the second set defines", secondSetTech)
	}
	if err := away.StartResearch(secondSetTech); err != nil {
		t.Fatalf("the second engine could not research its own tech: %v", err)
	}
	if _, known := home.GetState().Research.Techs[secondSetTech]; known {
		t.Error("the core engine's snapshot lists the second set's tech")
	}

	// A twin of the core engine that never meets the second one.
	alone := NewGameEngine()
	alone.SeedRNG(7)
	for i := 0; i < 40; i++ {
		home.StepTicks(10)
		away.StepTicks(10)
		alone.StepTicks(10)
	}
	if !away.GetState().Research.Techs[secondSetTech].Researched {
		t.Error("the second engine did not finish its own tech")
	}
	if home.StateDigest() != alone.StateDigest() {
		t.Error("the core engine played differently beside an engine on another ruleset")
	}
	if home.StateDigest() == away.StateDigest() {
		t.Error("the second engine's tech changed nothing: its ruleset was not the one it played by")
	}
}

// TestRunEndsKeepTheRuleset: every path that rebuilds the managers (a new
// game, a prestige, a Succumb, a load) builds them on the engine's own set,
// not on the core one.
func TestRunEndsKeepTheRuleset(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	other := secondSet()
	ge := NewGameEngineWith(other)
	ge.SeedRNG(11)

	if err := ge.EnterAgeForTest(PrestigeMinAge); err != nil {
		t.Fatal(err)
	}
	if err := ge.DoPrestige(); err != nil {
		t.Fatalf("prestige: %v", err)
	}
	wantManagersOn(t, ge, other, "after a prestige")

	if err := ge.EnterAgeForTest("iron_age"); err != nil {
		t.Fatal(err)
	}
	if err := ge.ForceCatastropheForTest(); err != nil {
		t.Fatalf("forcing a catastrophe: %v", err)
	}
	if err := ge.Succumb(); err != nil {
		t.Fatalf("succumb: %v", err)
	}
	wantManagersOn(t, ge, other, "after a Succumb")

	if err := ge.StartResearch(secondSetTech); err != nil {
		t.Fatalf("researching the set's own tech after a Succumb: %v", err)
	}
	ge.StepTicks(5)
	if err := ge.SaveGame("ruleset"); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded := NewGameEngineWith(other)
	if err := loaded.LoadGame("ruleset"); err != nil {
		t.Fatalf("load: %v", err)
	}
	wantManagersOn(t, loaded, other, "after a load")
	if !loaded.GetState().Research.Techs[secondSetTech].Researched {
		t.Error("the loaded game lost the tech its ruleset defines")
	}

	ge.Reset()
	wantManagersOn(t, ge, other, "after a new game")
}

// TestRebind: an engine moved onto the set it already plays by carries on
// exactly as before, and one moved onto another set plays by that one from
// then on, while a snapshot taken earlier keeps the set it was made from.
func TestRebind(t *testing.T) {
	play := func() *GameEngine {
		ge := newLateGameEngine(t)
		ge.GrantTechsForTest()
		ge.StepTicks(200)
		return ge
	}
	moved, control := play(), play()
	if moved.StateDigest() != control.StateDigest() {
		t.Fatal("two engines with one seed and one script are in different states")
	}
	before := moved.GetState()
	moved.Rebind(moved.Rules())
	wantManagersOn(t, moved, rules.Core(), "after a rebind to the same set")
	if moved.StateDigest() != control.StateDigest() {
		t.Error("a rebind to the same ruleset changed the run's state")
	}
	moved.StepTicks(300)
	control.StepTicks(300)
	if moved.StateDigest() != control.StateDigest() {
		t.Error("a run rebound to the same ruleset played on differently")
	}

	other := secondSet()
	moved.Rebind(other)
	wantManagersOn(t, moved, other, "after a rebind to another set")
	if before.Rules != rules.Core() {
		t.Error("a snapshot taken before the rebind no longer holds the set it was made from")
	}
	if after := moved.GetState(); after.Rules != other {
		t.Error("a snapshot taken after the rebind does not carry the new set")
	}
	if err := moved.StartResearch(secondSetTech); err != nil {
		t.Errorf("the rebound engine cannot research the new set's tech: %v", err)
	}
	if err := control.StartResearch(secondSetTech); err == nil {
		t.Error("the engine that was not rebound knows the new set's tech")
	}
}

// TestPlayLeavesTheRulesetUnchanged: the core set is shared by every engine
// in the process, so nothing an engine does may write into it. This walks an
// engine through every age, with every tech, a doom endured, a doom
// succumbed to and a prestige, snapshotting as it goes, and checks the set
// against its digest from before.
func TestPlayLeavesTheRulesetUnchanged(t *testing.T) {
	defer SetDataDirForTest(filepath.Join(t.TempDir(), "data"))()
	set := rules.Core()
	before := set.Digest()

	ge := newLateGameEngine(t)
	for _, age := range set.AgeKeys() {
		if err := ge.EnterAgeForTest(age); err != nil {
			t.Fatalf("entering %s: %v", age, err)
		}
		ge.GrantTechsForTest()
		for key := range ge.Resources.resources {
			setRes(ge, key, 1e9)
		}
		for _, b := range set.Buildings() {
			if b.RequiredAge == age && b.Category != "wonder" {
				_, _ = ge.BuildMultiple(b.Key, 2)
			}
		}
		ge.StepTicks(120)
		st := ge.GetState()
		ShareRows(st)
		RecruitStatus(st)
		ge.GetAvailableUpgrades()
		if ge.ForceCatastropheForTest() == nil {
			if err := ge.Endure(); err != nil {
				t.Fatalf("enduring in %s: %v", age, err)
			}
		}
	}
	if err := ge.SaveGame("ruleset-digest"); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := NewGameEngine().LoadGame("ruleset-digest"); err != nil {
		t.Fatalf("load: %v", err)
	}
	ge.SimulateOffline(MaxOfflineTime)
	if err := ge.EnterAgeForTest("iron_age"); err != nil {
		t.Fatal(err)
	}
	// The day away may have brought the era's own doom: one pending is as
	// good as one forced.
	if err := ge.ForceCatastropheForTest(); err != nil && ge.GetState().PendingCatastrophe == "" {
		t.Fatalf("forcing a catastrophe: %v", err)
	}
	if err := ge.Succumb(); err != nil {
		t.Fatalf("succumb: %v", err)
	}
	if err := ge.EnterAgeForTest(PrestigeRunAge); err != nil {
		t.Fatal(err)
	}
	if err := ge.DoPrestige(); err != nil {
		t.Fatalf("prestige: %v", err)
	}
	ge.StepTicks(120)
	ge.GetState()
	HarbingerPriceTable()

	if got := set.Digest(); got != before {
		t.Errorf("the core ruleset changed while an engine played: digest %s, was %s. Something wrote into a definition the set handed out", got, before)
	}
}
