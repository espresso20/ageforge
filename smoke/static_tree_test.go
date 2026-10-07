package smoke

import (
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// TestTechReachableTakesEitherBranch: the reachability check behind the gate
// covenant reads an either-or group as the game does. Every prerequisite
// must open in time; of an either-or group, one key is enough.
func TestTechReachableTakesEitherBranch(t *testing.T) {
	ages := []string{"a1", "a2", "a3"}
	techs := map[string]config.TechDef{}
	for _, d := range []config.TechDef{
		{Key: "base", Age: "a1"},
		{Key: "early", Age: "a1"},
		{Key: "late", Age: "a3"},
		{Key: "later", Age: "a3"},
		{Key: "join", Age: "a2", Prerequisites: []string{"base"}, AnyOf: []string{"late", "early"}},
		{Key: "both_late", Age: "a2", AnyOf: []string{"late", "later"}},
		{Key: "needs_late", Age: "a2", Prerequisites: []string{"late"}, AnyOf: []string{"early", "base"}},
		{Key: "on_join", Age: "a2", Prerequisites: []string{"join"}},
		{Key: "typo", Age: "a1", AnyOf: []string{"no_such_tech", "early"}},
		{Key: "all_typos", Age: "a1", AnyOf: []string{"no_such_tech", "nor_this"}},
		// A loop, which the config tests refuse; the check must still end.
		{Key: "loop_a", Age: "a1", Prerequisites: []string{"loop_b"}},
		{Key: "loop_b", Age: "a1", AnyOf: []string{"loop_a", "no_such_tech"}},
	} {
		techs[d.Key] = d
	}
	for _, c := range []struct {
		tech, age string
		want      bool
		why       string
	}{
		{"join", "a2", true, "one branch opens in time"},
		{"join", "a1", false, "the tech itself opens later"},
		{"on_join", "a2", true, "it stands on a tech one branch reaches"},
		{"both_late", "a2", false, "neither branch opens in time"},
		{"both_late", "a3", true, "both branches are open by then"},
		{"needs_late", "a2", false, "a prerequisite is late, whatever the either-or group offers"},
		{"needs_late", "a3", true, "everything is open by then"},
		{"typo", "a1", true, "one key of the group is real and in time"},
		{"all_typos", "a1", false, "no key of the group exists"},
		{"loop_a", "a3", false, "a loop can never be finished"},
		{"no_such_tech", "a3", false, "it is not a tech"},
		{"base", "no_such_age", false, "it is not an age"},
	} {
		if got := techReachableIn(techs, ages, c.tech, c.age); got != c.want {
			t.Errorf("%s by the end of %s: reachable %v, want %v (%s)", c.tech, c.age, got, c.want, c.why)
		}
	}
	// The real tables, through the entry point the gate check uses.
	if !techReachable("civilian_reactors", "atomic_age") || techReachable("civilian_reactors", "electric_age") {
		t.Error("Civilian Reactors should be reachable in the Atomic Age and not before")
	}
}

// TestMilestoneTechAgeTakesTheEarlierBranch: the milestone model's earliest
// age for a tech waits for every prerequisite but only for the earliest key
// of an either-or group.
func TestMilestoneTechAgeTakesTheEarlierBranch(t *testing.T) {
	m := newMilestoneModel(config.BuildingByKey(), game.PrestigeMinAge)
	age := func(i int) string { return m.ages[i].Key }
	byKey := map[string]config.TechDef{}
	for _, d := range []config.TechDef{
		{Key: "zz_early", Age: age(1)},
		{Key: "zz_late", Age: age(4)},
		{Key: "zz_either", Age: age(0), AnyOf: []string{"zz_late", "zz_early"}},
		{Key: "zz_both", Age: age(0), Prerequisites: []string{"zz_early", "zz_late"}},
		{Key: "zz_mixed", Age: age(2), Prerequisites: []string{"zz_early"}, AnyOf: []string{"zz_late", "zz_either"}},
		{Key: "zz_stuck", Age: age(0), AnyOf: []string{"zz_missing", "zz_gone"}},
	} {
		byKey[d.Key] = d
	}
	for key, want := range map[string]int{
		"zz_either": 1,           // the earlier of its two branches
		"zz_both":   4,           // the later of its two prerequisites
		"zz_mixed":  2,           // its own age: its prerequisite and one branch come sooner
		"zz_stuck":  len(m.ages), // no branch exists: never
	} {
		if got := m.techAge(key, byKey, 0); got != want {
			t.Errorf("%s can first be researched in age %d, want %d", key, got, want)
		}
	}
}
