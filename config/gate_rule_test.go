package config

import (
	"fmt"
	"sort"
	"testing"
)

// gateTownShortfalls lists every place a gate asks for less than the gate
// rule (gate_rule.go) gives: a building the rule requires that the gate does
// not ask for, or asks for in fewer copies.
func gateTownShortfalls(defs []BuildingDef, ages []AgeDef, techs []TechDef) []string {
	var out []string
	rule := GateTown(defs, ages, techs)
	for _, a := range ages {
		keys := make([]string, 0, len(rule[a.Key]))
		for k := range rule[a.Key] {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if got, want := a.BuildingReqs[k], rule[a.Key][k]; got < want {
				out = append(out, fmt.Sprintf("the gate into %s asks for %d %s, the gate rule asks for %d", a.Key, got, k, want))
			}
		}
	}
	return out
}

// TestGatesFollowTheTownRule holds the gate tables to the gate rule: the gate
// out of each age asks for the reference town's copies of every building the
// age introduced that makes something and that a run can build.
func TestGatesFollowTheTownRule(t *testing.T) {
	defs, ages, techs := BaseBuildings(), Ages(), Technologies()
	for _, s := range gateTownShortfalls(defs, ages, techs) {
		t.Error(s)
	}
	rule := GateTown(defs, ages, techs)
	if len(rule) != len(ages)-1 {
		t.Fatalf("the rule gives %d gates, there are %d", len(rule), len(ages)-1)
	}
}

// TestGateTownRuleAsksOnlyForWhatARunCanBuild: everything the rule asks for
// is buildable in the age the gate leaves, makes something, is not a wonder,
// monument, store or house, and waits for no tech but one every run has by
// the gate. Every building it passes over for a tech or for making nothing is
// one of those, so the rule is not hiding a building by accident.
func TestGateTownRuleAsksOnlyForWhatARunCanBuild(t *testing.T) {
	defs, ages, techs := BaseBuildings(), Ages(), Technologies()
	byKey := map[string]BuildingDef{}
	for _, d := range defs {
		byKey[d.Key] = d
	}
	idx := map[string]int{}
	for i, a := range ages {
		idx[a.Key] = i
	}
	held := GateTownHeldTechs(defs, ages, techs)
	for to, asked := range GateTown(defs, ages, techs) {
		from := idx[to] - 1
		for k, n := range asked {
			d := byKey[k]
			switch {
			case d.RequiredAge != ages[from].Key:
				t.Errorf("%s: the gate into %s asks for it, it is built in %s", k, to, d.RequiredAge)
			case IsKnowledgeBuilding(d):
				if want := KnowledgeGateCount(KnowledgeGateWas()[k]); n != want {
					t.Errorf("%s: %d copies, the knowledge rule asks %d", k, n, want)
				}
			case n != TownCopies:
				t.Errorf("%s: %d copies, the reference town holds %d", k, n, TownCopies)
			case !makesSomething(d):
				t.Errorf("%s makes nothing", k)
			case d.Category == "wonder" || d.Category == "monument" || d.Category == "storage" || d.Category == "housing":
				t.Errorf("%s is a %s", k, d.Category)
			case d.RequiredTech != "" && !held[from][d.RequiredTech]:
				t.Errorf("%s waits for %s, which a run may not have by the gate", k, d.RequiredTech)
			}
		}
		// Every building of the age that makes something is either asked for
		// or passed over for one named reason.
		for _, d := range defs {
			if d.RequiredAge != ages[from].Key {
				continue
			}
			if _, ok := asked[d.Key]; ok {
				continue
			}
			switch {
			case d.Category == "wonder" || d.Category == "monument" || d.Category == "storage" || d.Category == "housing":
			case !makesSomething(d):
			case d.RequiredTech != "" && !held[from][d.RequiredTech]:
			default:
				if _, listed := GateTownLeftOut()[d.Key]; !listed {
					t.Errorf("%s (%s) is passed over by the rule for no reason it names", d.Key, ages[from].Key)
				}
			}
		}
	}
}

// TestGateTownLeftOutIsTrue: each building the rule leaves out because a
// covenant stops it exists, would otherwise be asked for, and carries a
// reason.
func TestGateTownLeftOutIsTrue(t *testing.T) {
	defs, ages, techs := BaseBuildings(), Ages(), Technologies()
	byKey := map[string]BuildingDef{}
	for _, d := range defs {
		byKey[d.Key] = d
	}
	held := GateTownHeldTechs(defs, ages, techs)
	idx := map[string]int{}
	for i, a := range ages {
		idx[a.Key] = i
	}
	for k, why := range GateTownLeftOut() {
		d, ok := byKey[k]
		if !ok {
			t.Errorf("%s is left out of the gate rule and does not exist", k)
			continue
		}
		if why == "" {
			t.Errorf("%s is left out of the gate rule with no reason", k)
		}
		if !makesSomething(d) || (d.RequiredTech != "" && !held[idx[d.RequiredAge]][d.RequiredTech]) {
			t.Errorf("%s is left out of the gate rule, and the rule would not ask for it anyway", k)
		}
	}
}

// TestGateTownRuleCatchesAGateThatAsksForLess: a gate with a building taken
// out, or asked for in too few copies, is reported.
func TestGateTownRuleCatchesAGateThatAsksForLess(t *testing.T) {
	defs, techs := BaseBuildings(), Technologies()
	ages := Ages()
	if got := gateTownShortfalls(defs, ages, techs); len(got) != 0 {
		t.Fatalf("setup: the real gates fall short of the rule: %v", got)
	}
	delete(ages[1].BuildingReqs, "gathering_camp")
	ages[2].BuildingReqs["forager_post"] = TownCopies - 1
	got := gateTownShortfalls(defs, ages, techs)
	if len(got) != 2 {
		t.Fatalf("want the two broken gates reported, got %v", got)
	}
}
