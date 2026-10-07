package smoke

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/game"
)

// TestTreeResetCheck: the prestige check that the tech tree belongs to the
// run passes on a real prestige, with or without research planned, and
// fails on a state that still holds a tech or one of its bonuses.
func TestTreeResetCheck(t *testing.T) {
	t.Cleanup(game.SetDataDirForTest(t.TempDir()))
	ge := game.NewGameEngine()
	ge.SeedRNG(3)
	for _, a := range ge.Rules().AgeKeys()[1:] {
		if err := ge.EnterAgeForTest(a); err != nil {
			t.Fatal(err)
		}
		if a == game.PrestigeRunAge {
			break
		}
	}
	ge.GrantTechsForTest() // every tech up to the age
	held := ge.GetState()
	probs := treeResetProblems(held)
	var checks []string
	for _, p := range probs {
		checks = append(checks, p.check)
	}
	if got := strings.Join(checks, ","); got != "prestige_kept_techs,prestige_kept_tech_bonus" {
		t.Fatalf("a state full of techs fails %q, want the techs and their bonuses\n%+v", got, probs)
	}
	if held.Research.TotalResearched < 30 {
		t.Fatalf("setup: only %d techs researched", held.Research.TotalResearched)
	}

	_ = ge.SummonHarbingerForTest(game.PrestigeRunAge)
	if err := ge.DoPrestige(); err != nil {
		t.Fatal(err)
	}
	after := ge.GetState()
	if probs := treeResetProblems(after); len(probs) != 0 {
		t.Errorf("after a prestige the tree is not wiped: %+v", probs)
	}
	if probs := prestigeCarryProblems(held, after, "plain"); len(probs) != 0 {
		t.Errorf("a plain prestige fails its carry checks: %+v", probs)
	}

	// A research in progress is caught too.
	if err := ge.StartResearch(firstFreeTech(ge)); err != nil {
		t.Fatal(err)
	}
	if probs := treeResetProblems(ge.GetState()); len(probs) != 1 || probs[0].check != "prestige_kept_research" {
		t.Errorf("a research in progress: %+v", probs)
	}
}

// firstFreeTech is a tech of the first age that needs nothing, stocked for.
func firstFreeTech(ge *game.GameEngine) string {
	set := ge.Rules()
	for _, def := range set.Techs() {
		if def.Age == set.AgeKeys()[0] && len(def.Prerequisites) == 0 && len(def.AnyOf) == 0 {
			ge.SetStockForTest("knowledge", def.Cost)
			return def.Key
		}
	}
	return ""
}
