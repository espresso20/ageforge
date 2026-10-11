package smoke

import (
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
)

// scaledTechs is the real techs with every price multiplied by f.
func scaledTechs(f float64) []config.TechDef {
	techs := config.Technologies()
	for i := range techs {
		techs[i].Cost = float64(techs[i].Cost * f)
	}
	return techs
}

// researchProblems runs the Research Covenant over techs and returns what
// breaks it, by age.
func researchProblems(techs []config.TechDef, defs map[string]config.BuildingDef) map[string][]string {
	out := map[string][]string{}
	for _, r := range staticResearch(config.Ages(), defs, techs, referenceKnowledge) {
		if p := r.Problems(); len(p) > 0 {
			out[r.Age] = p
		}
	}
	return out
}

// TestResearchCovenant: with the tree's lock on, every age stays passable
// from config alone. The first tech of an age fits the knowledge storage a
// player enters with, or does within a few copies of the age's storage
// building; every tech a wonder waits for fits the age's knowledge storage;
// and the keystone is affordable well inside the age. The rules are at the
// top of static_research.go.
func TestResearchCovenant(t *testing.T) {
	rows := StaticResearch()
	if len(rows) != len(config.Ages()) {
		t.Fatalf("%d rows, want one per age (%d)", len(rows), len(config.Ages()))
	}
	keystones, most := 0, 0
	worst := ResearchRow{}
	for _, r := range rows {
		for _, p := range r.Problems() {
			t.Error(p)
		}
		if r.First == "" {
			t.Errorf("%s has no tech to be its first", r.Age)
		}
		most = max(most, r.StorageCopies)
		if r.Keystone == "" {
			continue
		}
		keystones++
		if len(r.Chain) == 0 || r.Chain[len(r.Chain)-1] != r.Keystone {
			t.Errorf("%s: the chain %v does not end on the keystone %s", r.Age, r.Chain, r.Keystone)
		}
		if r.TimeShare() > worst.TimeShare() {
			worst = r
		}
	}
	if keystones != 21 {
		t.Errorf("%d ages have a keystone, want 21: every age but the Primitive", keystones)
	}
	// The allowance is what the worst age needs, no more: when techs are
	// added and prices fall, this says to bring TechEntryStorageCopies down.
	if most != TechEntryStorageCopies {
		t.Errorf("the age that needs the most storage before its first tech fits needs %d copies; TechEntryStorageCopies is %d: make it match", most, TechEntryStorageCopies)
	}
	// Even the tightest age leaves a fifth of itself to spare.
	if _, known := ResearchTimeKnown[worst.Age]; !known && worst.TimeShare() > 0.8 {
		t.Errorf("%s: the keystone, its research and the wonder take %.0f%% of the age, over 80%%", worst.Age, 100*worst.TimeShare())
	}
	// What a keystones-only run carries in is reported: the Warp Nexus's
	// keystone stands on the Space Age's spine and, through it, the Fusion
	// Age's.
	for _, r := range rows {
		if r.Age != "interstellar_age" {
			continue
		}
		if got, want := strings.Join(r.Chain, ", "), "warp_drive"; got != want {
			t.Errorf("the Warp Nexus waits for %s of its own age, want %s", got, want)
		}
		if got, want := strings.Join(r.Carried, ", "), "space_mining, superconductors, zero_g_manufacturing"; got != want {
			t.Errorf("a keystones-only run carries %s into the Interstellar Age, want %s", got, want)
		}
		if r.CarriedShare <= 0 {
			t.Errorf("the carried techs cost %.0f%% of what the Interstellar Age makes", 100*r.CarriedShare)
		}
	}
}

func hasProblem(problems []string, sub string) bool {
	for _, p := range problems {
		if strings.Contains(p, sub) {
			return true
		}
	}
	return false
}

// TestEveryKeystoneIsReachableInItsAge: the gate check follows the wonder's
// new requirement. On the real tables every wonder's keystone, and all it
// stands on, opens by the wonder's own age, and not an age sooner (a
// keystone belongs to its wonder's age). A wonder whose keystone opens an
// age late is a gate nobody can pass, and the covenant says so.
func TestEveryKeystoneIsReachableInItsAge(t *testing.T) {
	ages := config.Ages()
	defs := config.BuildingByKey()
	checked := 0
	for i, a := range ages {
		w := ageWonder(a, defs)
		key := defs[w].RequiredTech
		if key == "" {
			continue
		}
		checked++
		if !techReachable(key, a.Key) {
			t.Errorf("%s needs %s, which cannot be finished by the end of the %s", w, key, a.Name)
		}
		if i > 0 && techReachable(key, ages[i-1].Key) {
			t.Errorf("%s's keystone %s can be finished in the %s already: it does not belong to its wonder's age", w, key, ages[i-1].Name)
		}
	}
	if checked != 21 {
		t.Errorf("%d wonders have a keystone, want 21", checked)
	}

	// The Colosseum behind Philosophy, a Classical Age tech: the Iron Age
	// could never be left.
	late := config.BuildingByKey()
	colosseum := late["colosseum"]
	colosseum.RequiredTech = "philosophy"
	late["colosseum"] = colosseum
	problems, _ := staticGates(ages, late)
	found := false
	for _, g := range problems {
		if g.Kind == "tech_locked" && g.Key == "colosseum" && g.From == "iron_age" {
			found = true
			continue
		}
		t.Errorf("unexpected problem %+v", g)
	}
	if !found {
		t.Error("a wonder whose keystone opens an age late was not flagged")
	}
}
