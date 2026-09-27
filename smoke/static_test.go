package smoke

import (
	"testing"

	"github.com/espresso20/ageforge/config"
)

// TestGateCovenant is the regression guard for the whole "requirement you
// can never store" class of soft-lock (the stash deadlock, the 50-longhouse
// Stone Age wall, the Atomic 30-bunker wall): every age advance must pass the
// Gate Covenant from config alone. It runs in plain `go test ./...`, so CI
// catches a balance change that breaks a gate before any bot has to play
// into it. See design-and-architecture/economy.md, "Gate Covenant".
func TestGateCovenant(t *testing.T) {
	problems, slack := StaticGates()
	for _, g := range problems {
		switch g.Kind {
		case "unbuildable":
			t.Errorf("%s -> %s: requires %d %s, which can only be built in %s; name the building its lineage has in %s",
				g.From, g.To, g.Count, g.Key, g.BuiltIn, g.From)
		case "unsourced":
			t.Errorf("%s -> %s: needs %s (for %s) but nothing supplies it by %s without first spending some",
				g.From, g.To, g.Resource, g.Key, g.From)
		case "dead_building":
			t.Errorf("%s (%s) costs %s, which has no source in that age", g.Key, g.From, g.Resource)
		case "building":
			t.Errorf("%s -> %s: copy #%d of %s costs %s %s, over 1/%g of the %s storage buildable in %s",
				g.From, g.To, g.Count, g.Key, num(g.Need), g.Resource, g.Margin, num(g.MaxStorage), g.From)
		default:
			t.Errorf("%s -> %s: needs %s %s, more than 1/%g of the %s storage buildable in %s",
				g.From, g.To, num(g.Need), g.Resource, g.Margin, num(g.MaxStorage), g.From)
		}
	}
	if want := len(config.Ages()) - 1; len(slack) != want {
		t.Errorf("slack rows = %d, want one per advance (%d)", len(slack), want)
	}
}

// TestGateCovenantCatchesBrokenGates keeps the guard honest: the pre-fix
// numbers must each be flagged. 50 longhouses was the Stone Age wall, 30
// barracks for Medieval named a Bronze Age building the age lock forbids
// building later, 40K food cannot fit 1.25x under Stone Age storage, and
// iron does not exist before the Bronze Age; and a Bronze Age smithy priced
// in coal (which unlocks in the Renaissance) could never be built.
func TestGateCovenantCatchesBrokenGates(t *testing.T) {
	ages := config.Ages()
	for i := range ages {
		switch ages[i].Key {
		case "bronze_age":
			ages[i].BuildingReqs = map[string]int{"longhouse": 50}
			ages[i].ResourceReqs = map[string]float64{"food": 40000, "iron": 10}
		case "medieval_age":
			ages[i].BuildingReqs = map[string]int{"barracks": 30}
		}
	}
	defs := config.BuildingByKey()
	smithy := defs["smithy"]
	smithy.BaseCost = map[string]float64{"wood": 900, "coal": 100}
	defs["smithy"] = smithy
	problems, _ := staticGates(ages, defs)
	want := map[string]bool{"building/longhouse": false, "resource/food": false, "unbuildable/barracks": false,
		"unsourced/bronze_age requirement": false, "dead_building/smithy": false}
	for _, g := range problems {
		k := g.Kind + "/" + g.Key
		if _, ok := want[k]; ok {
			want[k] = true
			continue
		}
		t.Errorf("unexpected problem %+v", g)
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("guard missed %s", k)
		}
	}
}
