package game

import (
	"fmt"
	"strings"
	"testing"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// measuredSet is the core ruleset with the measured gate on, as
// AGEFORGE_GATES=v2 makes it at start-up.
func measuredSet() *rules.Set {
	src := rules.FromConfig()
	src.GatesV2 = true
	return rules.Compile(src)
}

// measuredTownState is a town as the measured gate reads it.
type measuredTownState struct {
	pop     int
	income  float64 // the main materials' income a tick, in all
	keyTech bool
	wonder  bool
}

// setMeasuredTown puts the engine's town in state s for the gate out of m's
// age: the people, the rates of its materials (half on each of two, so the
// gate sums them), the key tech and the wonder.
func setMeasuredTown(ge *GameEngine, m config.GateMeasure, s measuredTownState) {
	wonder := ge.progress.WonderForAge(m.Age)
	counts := map[string]int{}
	if wonder != "" {
		counts[wonder] = 0
		if s.wonder {
			counts[wonder] = 1
		}
	}
	ge.SetTownForTest(counts, 0)
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.Workers.domains["worker"].count = s.pop
	for _, res := range m.Materials {
		ge.Resources.UnlockResource(res)
		ge.Resources.SetRate(res, s.income/float64(len(m.Materials)))
	}
	if m.KeyTech != "" {
		if s.keyTech {
			ge.Research.researched[m.KeyTech] = true
		} else {
			delete(ge.Research.researched, m.KeyTech)
		}
	}
	// The keystone is not a measure: a built wonder comes with it.
	if wonder != "" && s.wonder {
		if ks := ge.rules.Keystone(m.Age); ks != "" {
			ge.Research.researched[ks] = true
		}
	}
}

// Under the measured gate the pillars and CheckAdvancement agree in every
// age: with the measures met exactly, over, and each one short in turn.
func TestMeasuredPillarsAgreeWithCheckAdvancement(t *testing.T) {
	restore := SetDataDirForTest(t.TempDir())
	defer restore()
	set := measuredSet()
	ge := NewGameEngineWith(set)
	ge.SeedRNG(1)
	measures := config.GateMeasures()
	ages := set.Ages()
	if len(measures) != len(ages)-1 {
		t.Fatalf("%d measures for %d ages", len(measures), len(ages))
	}
	for i, m := range measures {
		from := ages[i]
		if m.Age != from.Key {
			t.Fatalf("measure %d is for %s, want %s", i, m.Age, from.Key)
		}
		_ = ge.ForceQuietFateForTest(from.EpochKey)

		whole := measuredTownState{pop: m.People, income: m.Income, keyTech: true, wonder: true}
		notReady := 0
		check := func(label string, s measuredTownState, want bool) {
			t.Helper()
			setMeasuredTown(ge, m, s)
			engineReady := ge.progress.CheckAdvancement(ge.age, ge.Buildings) != ""
			v := ge.GetState().AdvanceView()
			if engineReady != want {
				t.Fatalf("%s, %s: CheckAdvancement ready = %v, want %v", from.Key, label, engineReady, want)
			}
			if !engineReady {
				notReady++
			}
			if v.AllStand != engineReady {
				var rows []string
				for _, p := range v.Pillars {
					for _, r := range p.Rows {
						rows = append(rows, fmt.Sprintf("%s/%s %s met=%v", p.Key, r.Key, r.Label, r.Met))
					}
				}
				t.Errorf("%s, %s: pillars all stand = %v, CheckAdvancement ready = %v\n%s", from.Key, label, v.AllStand, engineReady, strings.Join(rows, "\n"))
			}
			if v.Ready != (engineReady && v.Blockers == nil) {
				t.Errorf("%s, %s: Ready = %v with all standing %v and blockers %v", from.Key, label, v.Ready, v.AllStand, v.Blockers)
			}
		}

		check("nothing", measuredTownState{}, false)
		check("the measures met exactly", whole, true)
		over := whole
		over.pop, over.income = m.People*2, m.Income*3
		check("over the measures", over, true)
		shortPop := whole
		shortPop.pop = m.People - 1
		check("one person short", shortPop, false)
		shortIncome := whole
		shortIncome.income = m.Income * 0.99
		check("the income short", shortIncome, false)
		noTech := whole
		noTech.keyTech = false
		check("the key tech missing", noTech, false)
		noWonder := whole
		noWonder.wonder = false
		check("the wonder missing", noWonder, false)
		if notReady < 5 {
			t.Errorf("%s: only %d of the checks left the gate shut", from.Key, notReady)
		}
		noErr(t, ge.EnterAgeForTest(set.Next(from.Key)))
	}
}

// The measured pillars carry the measures: a row of population, a row of the
// main materials' income, the keystone and the key tech.
func TestMeasuredPillarRows(t *testing.T) {
	restore := SetDataDirForTest(t.TempDir())
	defer restore()
	set := measuredSet()
	ge := NewGameEngineWith(set)
	ge.SeedRNG(1)
	for i, m := range config.GateMeasures() {
		_ = ge.ForceQuietFateForTest(set.Ages()[i].EpochKey)
		setMeasuredTown(ge, m, measuredTownState{pop: m.People / 2, income: m.Income / 4})
		v := ge.GetState().AdvanceView()
		if len(v.Pillars) != 4 {
			t.Fatalf("%s: %d pillars", m.Age, len(v.Pillars))
		}
		kn, people, industry := v.Pillars[0], v.Pillars[1], v.Pillars[2]
		if len(people.Rows) != 1 || people.Rows[0].Kind != GateRowAmount || people.Rows[0].Label != "Population" ||
			people.Rows[0].Target != float64(m.People) || people.Rows[0].Value != float64(m.People/2) || people.Met {
			t.Errorf("%s: People rows %+v", m.Age, people.Rows)
		}
		var names []string
		for _, res := range m.Materials {
			names = append(names, set.Name(rules.KindResource, res))
		}
		if len(industry.Rows) != 1 || industry.Rows[0].Kind != GateRowAmount ||
			industry.Rows[0].Label != strings.Join(names, " + ")+" income" || industry.Rows[0].Unit != "a tick" ||
			industry.Rows[0].Target != m.Income || industry.Met {
			t.Errorf("%s: Industry rows %+v", m.Age, industry.Rows)
		}
		want := 1
		if set.Keystone(m.Age) != "" {
			want = 2
		}
		if len(kn.Rows) != want || kn.Rows[want-1].Key != m.KeyTech || kn.Rows[want-1].Label != "Key tech: "+set.Name(rules.KindTech, m.KeyTech) {
			t.Errorf("%s: Knowledge rows %+v", m.Age, kn.Rows)
		}
		if want == 2 && kn.Rows[0].Key != set.Keystone(m.Age) {
			t.Errorf("%s: Knowledge does not lead with its keystone: %+v", m.Age, kn.Rows)
		}
		if _, ok := set.Tech(m.KeyTech); !ok {
			t.Errorf("%s: key tech %s is not a tech", m.Age, m.KeyTech)
		}
		noErr(t, ge.EnterAgeForTest(set.Next(m.Age)))
	}
}

// Under the measured gate no building count is asked, and a town of the
// whole building list with none of the measures is not ready; a town of the
// measures with no buildings but the wonder is.
func TestMeasuredGateAsksForNoBuildingCounts(t *testing.T) {
	restore := SetDataDirForTest(t.TempDir())
	defer restore()
	set := measuredSet()
	ge := NewGameEngineWith(set)
	ge.SeedRNG(1)
	_ = ge.ForceQuietFateForTest(set.Ages()[0].EpochKey)
	m, _ := set.GateMeasure("primitive_age")

	if got := ge.progress.GetRequirementsForNext("primitive_age"); len(got) != 0 {
		t.Errorf("the measured gate lists building counts: %v", got)
	}
	if got := ge.GetState().NextAgeBldReqs; len(got) != 0 {
		t.Errorf("the snapshot lists building counts: %v", got)
	}

	full := map[string]int{ge.progress.WonderForAge("primitive_age"): 1}
	for k, n := range set.Ages()[1].BuildingReqs {
		full[k] = n
	}
	ge.SetTownForTest(full, 0)
	if next := ge.progress.CheckAdvancement(ge.age, ge.Buildings); next != "" {
		t.Errorf("the building list opened the measured gate: %q", next)
	}

	setMeasuredTown(ge, m, measuredTownState{pop: m.People, income: m.Income, keyTech: true, wonder: true})
	if next := ge.progress.CheckAdvancement(ge.age, ge.Buildings); next != "stone_age" {
		t.Errorf("the measures and the wonder: gate leads to %q, want stone_age", next)
	}
}

// The plan's advance item, the advance command and the age-ready notice
// follow the measured gate.
func TestMeasuredGateReachesThePlanAndTheAdvance(t *testing.T) {
	restore := SetDataDirForTest(t.TempDir())
	defer restore()
	set := measuredSet()
	ge := NewGameEngineWith(set)
	ge.SeedRNG(1)
	_ = ge.ForceQuietFateForTest(set.Ages()[0].EpochKey)
	m, _ := set.GateMeasure("primitive_age")
	noErr(t, ge.PlanAddAdvance())

	setMeasuredTown(ge, m, measuredTownState{pop: m.People - 1, income: m.Income, keyTech: true, wonder: true})
	if b := ge.planAdvanceBlocker(); b == "" {
		t.Error("the plan's advance item is ready a person short")
	}
	if err := ge.AdvanceAge(); err == nil {
		t.Error("the advance command went through a person short")
	}
	if ge.age != "primitive_age" {
		t.Fatalf("advanced to %s a person short", ge.age)
	}

	setMeasuredTown(ge, m, measuredTownState{pop: m.People, income: m.Income, keyTech: true, wonder: true})
	if b := ge.planAdvanceBlocker(); b != "" {
		t.Errorf("the plan's advance item is blocked with the measures met: %q", b)
	}
	if err := ge.AdvanceAge(); err != nil {
		t.Fatalf("the advance command refused a town that meets the measures: %v", err)
	}
	if ge.age != "stone_age" {
		t.Errorf("age = %s after the advance, want stone_age", ge.age)
	}
}

// Without the switch the gate is today's: the measures are never read, even
// when the town meets them, and a town of the buildings passes.
func TestTodaysGateIgnoresTheMeasures(t *testing.T) {
	restore := SetDataDirForTest(t.TempDir())
	defer restore()
	src := rules.FromConfig()
	src.GatesV2 = false
	set := rules.Compile(src)
	ge := NewGameEngineWith(set)
	ge.SeedRNG(1)
	m, _ := measuredSet().GateMeasure("primitive_age")

	setMeasuredTown(ge, m, measuredTownState{pop: m.People, income: m.Income, keyTech: true, wonder: true})
	if next := ge.progress.CheckAdvancement(ge.age, ge.Buildings); next != "" {
		t.Errorf("the measures opened today's gate: %q", next)
	}
	full := map[string]int{ge.progress.WonderForAge("primitive_age"): 1}
	for k, n := range set.Ages()[1].BuildingReqs {
		full[k] = n
	}
	ge.SetTownForTest(full, 0)
	if next := ge.progress.CheckAdvancement(ge.age, ge.Buildings); next != "stone_age" {
		t.Errorf("today's gate with the buildings and the wonder leads to %q, want stone_age", next)
	}
	if got := ge.progress.GetRequirementsForNext("primitive_age"); len(got) == 0 {
		t.Error("today's gate lists no building counts")
	}
}

// The tick opens the measured gate for a real town: homes, wood camps and the
// key tech in, the engine says ready and the notice is logged; with the key
// tech out it says not.
func TestMeasuredGateOpensThroughTheTick(t *testing.T) {
	restore := SetDataDirForTest(t.TempDir())
	defer restore()
	ge := NewGameEngineWith(measuredSet())
	ge.SeedRNG(1)
	_ = ge.ForceQuietFateForTest("stone_era")
	m, _ := ge.rules.GateMeasure("primitive_age")

	// A town of homes and wood camps with the wonder built, and no key tech.
	ge.SetTownForTest(map[string]int{"hut": 30, "wood_camp": 20, "gathering_camp": 20, "sacred_grove": 1, "stash": 5, "story_circle": 5}, 200)
	ge.StepTicks(50)
	st := ge.GetState()
	if st.Workers.TotalPop < m.People || st.Resources["wood"].Rate < m.Income {
		t.Fatalf("staging: the town has %d people and %.2f wood a tick, the measures are %d and %.2f",
			st.Workers.TotalPop, st.Resources["wood"].Rate, m.People, m.Income)
	}
	if st.AgeReady || st.AdvanceView().AllStand {
		t.Fatal("the gate is open without its key tech")
	}

	ge.GrantTechsForTest(m.KeyTech)
	ge.StepTicks(1)
	st = ge.GetState()
	if !st.AgeReady || !st.AdvanceView().AllStand {
		t.Fatalf("the gate is shut with the measures met: ready %v, pillars %v", st.AgeReady, st.AdvanceView().AllStand)
	}
	if !logHas(ge, "Ready to advance") {
		t.Error("no notice that the age is ready")
	}
	if err := ge.AdvanceAge(); err != nil {
		t.Fatalf("advance: %v", err)
	}
	if got := ge.GetState().Age; got != "stone_age" {
		t.Errorf("age = %s, want stone_age", got)
	}
}
