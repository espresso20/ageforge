package smoke

import (
	"sort"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
)

// TestMilestonesAreFeasible: every milestone, chain and title can be
// completed, proven from config alone (the Milestone Covenant, see
// static_milestones.go). It runs in plain `go test ./...`, so a balance
// change that strands a milestone (a max count, housing, the tech tree, the
// age a run prestiges at) fails CI with the math that broke.
func TestMilestonesAreFeasible(t *testing.T) {
	problems, reach := StaticMilestones()
	for _, p := range problems {
		t.Error(p.Why)
	}
	if len(reach) != len(config.Milestones()) {
		t.Errorf("verdicts = %d, want one per milestone (%d)", len(reach), len(config.Milestones()))
	}
	for _, r := range reach {
		if r.Earliest == "" && len(problems) == 0 {
			t.Errorf("%s has no earliest age but no problem either", r.Key)
		}
	}
}

// preCovenantMilestones is the milestone data master shipped before the
// feasibility check existed, patched onto today's config, as far as it can
// still break: the two that named an age too early for their count. The
// others (the last Stone Pit, Temple, Trading Post and Coal Plant costing more
// than their age could store, a billion people, 20,000 structures) were
// broken by the stores, which are gone: stock has no cap, so those counts
// can be built.
func preCovenantMilestones(t *testing.T) []config.MilestoneDef {
	t.Helper()
	ms := config.Milestones()
	patched := map[string]bool{}
	for i := range ms {
		d := &ms[i]
		switch d.Key {
		case "wonder_empire":
			d.MinAge = "modern_age"
		case "tech_ascendant":
			d.MinAge = "quantum_age"
		default:
			continue
		}
		patched[d.Key] = true
	}
	if len(patched) != 2 {
		t.Fatalf("patched %d milestones, want 2: a key was renamed?", len(patched))
	}
	return ms
}

// TestMilestoneFeasibilityCatchesBrokenMilestones keeps the guard honest:
// the numbers master shipped before this check that can still break must be
// flagged, for the right reason, along with the chains they block. Two named
// an age too early for their count (15 wonders, every tech).
func TestMilestoneFeasibilityCatchesBrokenMilestones(t *testing.T) {
	// Judged under the rules those numbers shipped with: prestige opened in
	// the Modern Age. The prestige test covers moving it.
	problems, _ := staticMilestones(preCovenantMilestones(t), config.MilestoneChains(), config.MilestoneTitles(), config.BuildingByKey(), "modern_age")
	got := map[string]string{}
	for _, p := range problems {
		got[p.Key] = p.Kind
		t.Log(p.Why)
	}
	want := map[string]string{"tech_ascendant": "techs", "wonder_empire": "wonders"}
	for _, c := range config.MilestoneChains() {
		for _, k := range c.MilestoneKeys {
			if _, broken := want[k]; broken {
				want[c.Key] = "chain"
			}
		}
	}
	for k, kind := range want {
		if got[k] != kind {
			t.Errorf("%s: flagged as %q, want %q", k, got[k], kind)
		}
	}
	var extra []string
	for k := range got {
		if _, ok := want[k]; !ok {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	if len(extra) > 0 {
		t.Errorf("flagged beyond the known breakage: %v", extra)
	}
}

// TestMilestoneFeasibilityBoundaries: each count sits exactly on its
// ceiling and one past it. A building count may equal the copy ceiling, a
// population or build count the share of its ceiling; one more fails with
// the milestone, the requirement and the numbers in the message.
func TestMilestoneFeasibilityBoundaries(t *testing.T) {
	defs := config.BuildingByKey()
	m := newMilestoneModel(defs, game.PrestigeMinAge)
	run := func(d config.MilestoneDef) []MilestoneProblem {
		_, p := m.check(d)
		return p
	}
	expect := func(label string, d config.MilestoneDef, fails bool, words ...string) {
		t.Helper()
		p := run(d)
		if !fails {
			for _, x := range p {
				t.Errorf("%s: unexpected problem: %s", label, x.Why)
			}
			return
		}
		if len(p) == 0 {
			t.Errorf("%s: not flagged", label)
			return
		}
		for _, w := range append([]string{d.Name, d.Key}, words...) {
			if !strings.Contains(p[0].Why, w) {
				t.Errorf("%s: %q should mention %q", label, p[0].Why, w)
			}
		}
	}

	// A building with no max count stands up to the model's ceiling: stock
	// has no cap, so nothing else stops it.
	c := m.ceiling("stone_pit")
	if c.n != copyLimit {
		t.Fatalf("stone_pit ceiling = %+v, want the model's ceiling %d (no max count, no store)", c, copyLimit)
	}
	pits := config.MilestoneDef{Name: "Pits", Key: "pits", MinBuildings: map[string]int{"stone_pit": c.n}}
	expect("stone pits at the ceiling", pits, false)
	pits.MinBuildings = map[string]int{"stone_pit": c.n + 1}
	expect("stone pits past the ceiling", pits, true, "Stone Pits", grouped(c.n))

	// A copy limit is the other ceiling: the monuments have one each.
	g := m.ceiling("cultural_obelisk")
	if g.n != 1 {
		t.Fatalf("cultural_obelisk ceiling = %+v, want its max count (1) to bind", g)
	}
	obelisks := config.MilestoneDef{Name: "Obelisks", Key: "obelisks", MinBuildings: map[string]int{"cultural_obelisk": g.n + 1}}
	expect("obelisks past the max count", obelisks, true, "max count")

	end := m.runEnd
	popLimit := int(MilestonePopShare * m.housing[end])
	pop := config.MilestoneDef{Name: "Crowd", Key: "crowd", MinPopulation: popLimit}
	expect("population at the share", pop, false)
	pop.MinPopulation = popLimit + 1
	expect("population past the share", pop, true, "population", game.AgeName(m.ages[end].Key), grouped(popLimit))

	buildLimit := int(MilestoneBuildShare * float64(m.builds[end]))
	builds := config.MilestoneDef{Name: "Busy", Key: "busy", MinTotalBuilt: buildLimit}
	expect("builds at the share", builds, false)
	builds.MinTotalBuilt = buildLimit + 1
	expect("builds past the share", builds, true, "structures", grouped(buildLimit))

	techs := config.MilestoneDef{Name: "Nerd", Key: "nerd", MinTechCount: m.techs[end]}
	expect("techs at the run's count", techs, false)
	techs.MinTechCount = m.techs[end] + 1
	expect("techs past the run's count", techs, true, "technologies")
	techs.MinAge, techs.MinTechCount = m.ages[len(m.ages)-1].Key, len(config.Technologies())+1
	expect("more techs than exist", techs, true, "whole game")

	// Past the run, a milestone is due in its own age. The probes below
	// pick their ages from config, so moving prestige doesn't break them.
	if after := end + 1; after < len(m.ages) {
		late := config.MilestoneDef{Name: "Late", Key: "late", MinAge: m.ages[after].Key, MinWonders: m.wonders[after]}
		expect("wonders in their own age", late, false)
		late.MinWonders++
		expect("wonders past their own age", late, true, "wonders", game.AgeName(m.ages[after].Key))
	}

	sum := config.MilestoneDef{Name: "Sum", Key: "sum", MinBuildingSum: config.BuildingSum{Keys: []string{"stone_pit", "cultural_obelisk"}, Count: c.n + g.n}}
	expect("building sum at the ceiling", sum, false)
	sum.MinBuildingSum.Count++
	expect("building sum past the ceiling", sum, true, "Stone Pits and Cultural Obelisks", "between them")

	// Stock has no cap: any amount of a resource an age supplies can be held.
	hoard := config.MilestoneDef{Name: "Hoard", Key: "hoard", MinResources: map[string]float64{"gold": 1e300}}
	expect("gold, any amount", hoard, false)
	for _, r := range config.BaseResources() {
		if m.resAge[r.Key] > end {
			expect("a resource from after the run", config.MilestoneDef{Name: "Later", Key: "later", MinResources: map[string]float64{r.Key: 1}}, true, game.ResourceName(r.Key))
			break
		}
	}

	soldiers := config.MilestoneDef{Name: "Army", Key: "army", MinSoldiersTrained: int(m.soldiers[end])}
	expect("soldiers at a moderate income", soldiers, false)
	soldiers.MinSoldiersTrained = int(m.soldiers[end]) + 1
	expect("soldiers past a moderate income", soldiers, true, "soldiers trained")

	staff := int(MilestonePopShare * min(m.staff[end], m.housing[end]))
	scholars := config.MilestoneDef{Name: "Scholars", Key: "scholars", MinKnowledgeWorkers: staff}
	expect("knowledge workers at the share", scholars, false)
	scholars.MinKnowledgeWorkers = staff + 1
	expect("knowledge workers past the share", scholars, true, "knowledge workers")

	for _, tc := range config.Technologies() {
		if a := m.techAt[tc.Key]; a > end && a < len(m.ages) {
			named := config.MilestoneDef{Name: "Named", Key: "named", RequiredTechs: []string{tc.Key}}
			expect("a tech from after the run", named, true, game.TechName(tc.Key), game.AgeName(m.ages[a].Key))
			named.MinAge = m.ages[a].Key
			expect("a tech in its own age", named, false)
			break
		}
	}

	unknown := config.MilestoneDef{Name: "Typo", Key: "typo", MinBuildings: map[string]int{"stone_pits": 1}}
	expect("unknown building", unknown, true, "stone_pits", "does not exist")
}

// TestMilestoneFeasibilityFollowsPrestige: a run ends where prestige opens
// (game.PrestigeMinAge), so the check re-validates every milestone when that
// moves. Opening prestige at the Industrial Age would cut the run off at the
// Colonial Age, too early for the technologies two milestones ask for and the
// chain that holds them; opening it at the last age only relaxes things.
func TestMilestoneFeasibilityFollowsPrestige(t *testing.T) {
	ms, chains, titles, defs := config.Milestones(), config.MilestoneChains(), config.MilestoneTitles(), config.BuildingByKey()
	if m := newMilestoneModel(defs, game.PrestigeMinAge); m.ages[m.runEnd+1].Key != game.PrestigeMinAge {
		t.Fatalf("the run ends before %s, want before %s", m.ages[m.runEnd+1].Key, game.PrestigeMinAge)
	}
	early, _ := staticMilestones(ms, chains, titles, defs, "industrial_age")
	flagged := map[string]string{}
	for _, p := range early {
		flagged[p.Key] = p.Kind
	}
	for k, kind := range map[string]string{"philosophes": "techs", "renaissance_mind": "techs", "scholar_chain": "chain"} {
		if flagged[k] != kind {
			t.Errorf("with prestige at the Industrial Age, %s: flagged as %q, want %q", k, flagged[k], kind)
		}
	}
	late, _ := staticMilestones(ms, chains, titles, defs, config.AgeOrder()[len(config.AgeOrder())-1])
	for _, p := range late {
		t.Errorf("with prestige at the last age: %s", p.Why)
	}
}
