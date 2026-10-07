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
	for _, r := range staticResearch(config.Ages(), defs, techs, config.AgeKnowledge) {
		if p := r.Problems(); len(p) > 0 {
			out[r.Age] = p
		}
	}
	return out
}

// TestResearchCovenant: with the tree's lock on, every age stays passable
// from config alone. The first tech of an age fits the knowledge storage a
// player enters with, or does after one copy of the age's storage building;
// every tech a wonder waits for fits the age's knowledge storage; and what
// the wonder waits for is affordable well inside the age, even for a run
// that researched nothing an earlier wonder did not ask for. The rules are
// at the top of static_research.go.
func TestResearchCovenant(t *testing.T) {
	rows := StaticResearch()
	if len(rows) != len(config.Ages()) {
		t.Fatalf("%d rows, want one per age (%d)", len(rows), len(config.Ages()))
	}
	keystones := 0
	for _, r := range rows {
		for _, p := range r.Problems() {
			t.Error(p)
		}
		if r.First == "" {
			t.Errorf("%s has no tech to be its first", r.Age)
		}
		if r.Keystone == "" {
			continue
		}
		keystones++
		if len(r.Chain) == 0 || r.Chain[len(r.Chain)-1] != r.Keystone {
			t.Errorf("%s: the chain %v does not end on the keystone %s", r.Age, r.Chain, r.Keystone)
		}
	}
	if keystones != 20 {
		t.Errorf("%d ages have a keystone, want 20: every age but the Primitive and the Bronze", keystones)
	}
	// The chain is the worst case: a run that researched only what each
	// wonder asked for arrives in the Interstellar Age owing the Space
	// Age's spine too.
	for _, r := range rows {
		if r.Age == "interstellar_age" {
			if want := "space_mining, superconductors, zero_g_manufacturing, warp_drive"; strings.Join(r.Chain, ", ") != want {
				t.Errorf("the Warp Nexus waits for %v, want %s", r.Chain, want)
			}
		}
	}
}

// TestResearchCovenantCatchesBrokenNumbers keeps the guard honest. The tree
// was drawn for a budget share of 0.9 and nine techs an age; with today's 77
// techs that is every price three times today's, and the covenant must say
// why it will not do: the first tech of the Classical to Industrial Ages
// would not fit the knowledge storage a player enters with, even after a
// storage building (but for the Medieval and Industrial Ages, where one is
// enough), and a keystones-only run could not afford the Warp Nexus's chain
// inside the Interstellar Age. A keystone priced over the age's storage and
// an age left with no storage building are caught too.
func TestResearchCovenantCatchesBrokenNumbers(t *testing.T) {
	defs := config.BuildingByKey()
	if got := researchProblems(config.Technologies(), defs); len(got) != 0 {
		t.Fatalf("the real tables break the covenant: %v", got)
	}
	// A share of 0.9 is today's prices × 0.9 ÷ ResearchBudgetShare; the two
	// ages with a share of their own do not scale, and are not looked at.
	got := researchProblems(scaledTechs(0.9/config.ResearchBudgetShare), defs)
	for _, age := range []string{"classical_age", "renaissance_age", "colonial_age"} {
		if !hasProblem(got[age], "its first tech") {
			t.Errorf("at a share of 0.9 the %s's first tech should not fit the storage a player enters with: %v", age, got[age])
		}
	}
	for _, age := range []string{"stone_age", "iron_age", "medieval_age", "industrial_age", "victorian_age", "atomic_age"} {
		if hasProblem(got[age], "its first tech") {
			t.Errorf("at a share of 0.9 the %s's first tech still fits: %v", age, got[age])
		}
	}
	if !hasProblem(got["interstellar_age"], "over 100%") {
		t.Errorf("at a share of 0.9 the Warp Nexus's chain should not fit the Interstellar Age: %v", got["interstellar_age"])
	}

	// One keystone priced over anything its age can store.
	techs := config.Technologies()
	for i := range techs {
		if techs[i].Key == "mathematics" {
			techs[i].Cost = 2 * maxStorageIn(defs, "iron_age", "knowledge")
		}
	}
	if got := researchProblems(techs, defs); !hasProblem(got["iron_age"], "knowledge storage buildable in the age") {
		t.Errorf("a keystone over the Iron Age's storage was not flagged: %v", got["iron_age"])
	}

	// The Renaissance needs its first vault before Navigation fits. With a
	// vault that holds a tenth as much, one is not enough.
	small := config.BuildingByKey()
	vault := small["renaissance_vault"]
	vault.Effects = append([]config.Effect(nil), vault.Effects...)
	for i, e := range vault.Effects {
		if e.Type == "storage" {
			vault.Effects[i].Value = e.Value / 10
		}
	}
	small["renaissance_vault"] = vault
	if got := researchProblems(config.Technologies(), small); !hasProblem(got["renaissance_age"], "its first tech") {
		t.Errorf("a Renaissance vault a tenth the size was not flagged: %v", got["renaissance_age"])
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
	if checked != 20 {
		t.Errorf("%d wonders have a keystone, want 20", checked)
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
