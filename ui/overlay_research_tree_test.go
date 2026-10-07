package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// TestResearchPanelNamesAnEitherOrGroup: a tech's "needs:" line lists its
// prerequisites, then its either-or group as "A or B", for a locked tech and
// for one that can start.
func TestResearchPanelNamesAnEitherOrGroup(t *testing.T) {
	src := rules.FromConfig()
	first := src.Ages[0].Key
	src.Techs = append(src.Techs,
		config.TechDef{Key: "zz_base", Name: "Zz Base", Age: first, Lane: config.LaneCraft},
		config.TechDef{Key: "zz_left", Name: "Zz Left", Age: first, Lane: config.LaneCraft},
		config.TechDef{Key: "zz_right", Name: "Zz Right", Age: first, Lane: config.LaneCraft},
		config.TechDef{Key: "zz_join", Name: "Zz Join", Age: first, Lane: config.LaneCraft,
			Prerequisites: []string{"zz_base"}, AnyOf: []string{"zz_left", "zz_right"}},
		config.TechDef{Key: "zz_pick", Name: "Zz Pick", Age: first, Lane: config.LaneCraft, AnyOf: []string{"zz_left", "zz_right"}},
	)
	ge := game.NewGameEngineWith(rules.Compile(src))
	ge.SeedRNG(1)

	st := ge.GetState()
	tree := section(t, researchProvider(st, 120), "Tech tree", "")
	for _, want := range []string{"needs: Zz Base, Zz Left or Zz Right[-]", "needs: Zz Left or Zz Right[-]"} {
		if !strings.Contains(tree, want) {
			t.Errorf("the tech tree is missing %q:\n%s", want, tree)
		}
	}
	if got := techNeeds(st.Research.Techs["zz_join"], st.Ruleset().TechMap()); len(got) != 2 || got[0] != "Zz Base" || got[1] != "Zz Left or Zz Right" {
		t.Errorf("Zz Join needs %q, want its prerequisite and then its either-or group", got)
	}
	if got := techNeeds(st.Research.Techs["tool_making"], st.Ruleset().TechMap()); len(got) != 0 {
		t.Errorf("a tech that needs nothing lists %q", got)
	}
}
