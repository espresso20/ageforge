package game

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
)

// gate_v2.go is the measured gate (experimental, AGEFORGE_GATES=v2). It is a
// second rule for what stands between a town and the next age, picked once
// when the game starts (rules.Set.GatesV2). Today's rule asks for building
// counts and the age's wonder. The measured rule asks for the wonder and:
//
//   - People: the town's population reaching the reference town's;
//   - Industry: the income a tick of the age's main materials together
//     reaching the reference town's;
//   - Knowledge: the age's keystone tech (which the wonder cannot be built
//     without) and its key tech researched.
//
// The numbers are config.GateMeasure's, read off the reference model. An age
// with no measure keeps the building counts.
//
// ProgressManager.CheckAdvancement and the Advance panel's pillars
// (AdvanceView) read the same measuredTown, so "all four pillars stand" and
// "CheckAdvancement says ready" are one fact under this rule too.

// gateTown is what the measured gate reads of a town beyond its buildings.
// The engine fills it in; a nil function reads as nothing (no people, no
// income, no tech).
type gateTown struct {
	population func() int
	rate       func(resource string) float64
	researched func(tech string) bool
}

// measuredTown is how a town stands against one age's measures.
type measuredTown struct {
	Measure    config.GateMeasure
	Population int
	// Income is the income a tick of the measure's materials together.
	Income  float64
	KeyTech bool
}

// measureTown reads a town against m: its population, the rate of each of m's
// materials (summed in m's order) and whether m's key tech is researched.
func measureTown(m config.GateMeasure, population int, rate func(string) float64, researched func(string) bool) measuredTown {
	t := measuredTown{Measure: m, Population: population, KeyTech: m.KeyTech == ""}
	if rate != nil {
		for _, res := range m.Materials {
			t.Income += rate(res)
		}
	}
	if researched != nil && m.KeyTech != "" {
		t.KeyTech = researched(m.KeyTech)
	}
	return t
}

// measure reads this town against m.
func (g gateTown) measure(m config.GateMeasure) measuredTown {
	pop := 0
	if g.population != nil {
		pop = g.population()
	}
	return measureTown(m, pop, g.rate, g.researched)
}

// PeopleMet, IndustryMet and KeyTechMet are the pillars' measures; the
// keystone and the wonder are the wonder's own and CheckAdvancement asks for
// the wonder built, which the keystone comes with.
func (t measuredTown) PeopleMet() bool   { return t.Population >= t.Measure.People }
func (t measuredTown) IndustryMet() bool { return t.Income >= t.Measure.Income }
func (t measuredTown) KeyTechMet() bool  { return t.KeyTech }

// Met is true when the town meets every measure.
func (t measuredTown) Met() bool { return t.PeopleMet() && t.IndustryMet() && t.KeyTechMet() }

// measureFor is the measure of the gate out of age when the ruleset's gates
// are measured and the age has one.
func (pm *ProgressManager) measureFor(age string) (config.GateMeasure, bool) {
	if pm.rules == nil || !pm.rules.GatesV2() {
		return config.GateMeasure{}, false
	}
	return pm.rules.GateMeasure(age)
}

// checkMeasured is CheckAdvancement under the measured gate: the next age
// when the town meets m and the current age's wonder is built.
func (pm *ProgressManager) checkMeasured(currentKey, nextKey string, m config.GateMeasure, buildings *BuildingManager) string {
	if !pm.town.measure(m).Met() {
		return ""
	}
	// Require current age's wonder to be built before advancing
	if wonderKey := pm.ageWonders[currentKey]; wonderKey != "" {
		if buildings.GetCount(wonderKey) < 1 {
			return ""
		}
	}
	return nextKey
}

// keyTechOf returns the key tech of the gate out of age under the measured
// rule, "" if the gate is not measured or the age has none.
func keyTechOf(set *rules.Set, age string) string {
	if !set.GatesV2() {
		return ""
	}
	m, ok := set.GateMeasure(age)
	if !ok {
		return ""
	}
	return m.KeyTech
}

// gateTown is the engine's town as the measured gate reads it. It reads the
// engine's fields as the tick does and takes no lock: callers hold the
// engine's lock already, as CheckAdvancement's callers do.
func (ge *GameEngine) gateTown() gateTown {
	return gateTown{
		population: func() int { return ge.popLocked() },
		rate:       func(res string) float64 { return ge.Resources.GetRate(res) },
		researched: func(tech string) bool { return ge.Research != nil && ge.Research.IsResearched(tech) },
	}
}

// intOf is x as a whole number for a row that counts, held where an int holds it.
func intOf(x float64) int {
	switch {
	case x != x || x < 0:
		return 0
	case x > 1<<62:
		return 1 << 62
	}
	return int(x)
}
