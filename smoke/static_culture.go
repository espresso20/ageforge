package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// The culture strength check: the faith strength check's towns (static_faith.go),
// for culture. Towns that never spend their culture are walked through the
// whole run at the pacing targets, hoard carried from age to age, and read
// by the rule the epoch roll reads (game.CultureStrengthOf) on entering each
// age, a fifth in, halfway and on leaving it (the roll falls at an advance
// into a new era).
//
//   - A moderate town (config.FlowCopies copies of every culture building)
//     must stay on Minor events throughout. So must a town with no culture
//     buildings at all, whatever its wonders hand it.
//   - A town with twice the culture buildings must have Major events open.
//   - A town with CultureTopSets times the culture buildings must have the
//     Legendary tier open.
//
// The towns the check walks, in moderate sets of culture buildings.
const (
	CultureTwiceSets  = 2.0
	CultureThriceSets = 3.0
	// CultureTopSets is the devotion the Legendary tier must be open at.
	CultureTopSets = 3.5
)

// CultureStrengthRow is one age of the culture strength check.
type CultureStrengthRow = StrengthRow

// The rules a CultureStrengthProblem can break.
const (
	// CultureRuleModerate: a moderate town, or one with no culture
	// buildings, has more than Minor events open.
	CultureRuleModerate = "moderate"
	// CultureRuleTwice: a town with twice the culture buildings does not
	// have Major events open (or has the Legendary too).
	CultureRuleTwice = "twice"
	// CultureRuleTop: a town with CultureTopSets times the culture buildings
	// does not have the Legendary tier open.
	CultureRuleTop = "top"
)

// CultureStrengthProblem is one moment at which a town reads the wrong tier.
type CultureStrengthProblem struct{ StrengthProblem }

// cultureTierWord names a tier as the report and the wiki do.
func cultureTierWord(t game.CultureTier) string {
	switch t {
	case game.CultureTierMajor:
		return "Major"
	case game.CultureTierLegendary:
		return "Legendary"
	}
	return "Minor"
}

// String says what is wrong in plain words.
func (p CultureStrengthProblem) String() string {
	town := fmt.Sprintf("a town with %g times the moderate set of culture buildings", p.Sets)
	switch p.Sets {
	case 0:
		town = "a town with no culture buildings"
	case 1:
		town = "a moderate town"
	}
	want := "Minor events only"
	switch p.Rule {
	case CultureRuleTwice:
		want = "Major events open and no more"
	case CultureRuleTop:
		want = "the Legendary tier open"
	}
	wonder := "the age's wonder up as it ends"
	if p.WonderEarly {
		wonder = "the age's wonder up from its first tick"
	}
	return fmt.Sprintf("%s, %s (%s): %s that never spends its culture reads %s, which opens the %s tier; want %s", p.Age, p.Moment, wonder, town, sharePct(p.Strength), cultureTierWord(game.CultureTierAt(p.Strength)), want)
}

// cultureSpan prints a span of culture strength with the tier it opens.
func cultureSpan(s FaithSpan) string {
	lo, hi := cultureTierWord(game.CultureTierAt(s.Low)), cultureTierWord(game.CultureTierAt(s.High))
	if sharePct(s.Low) == sharePct(s.High) && lo == hi {
		return fmt.Sprintf("%s (%s)", sharePct(s.Low), lo)
	}
	return fmt.Sprintf("%s (%s) to %s (%s)", sharePct(s.Low), lo, sharePct(s.High), hi)
}

// StaticCultureStrength walks the check's towns through the core ruleset by
// the rule the epoch roll reads, and returns the problems with the table
// they come from. Ages before culture exists have no row.
func StaticCultureStrength() ([]CultureStrengthProblem, []CultureStrengthRow) {
	set := rules.Core()
	problems, rows := cultureStrengthProblems(flowIncomes(set, "culture"), func(held float64, m game.CultureSave) float64 {
		return game.CultureStrengthOf(held, m)
	})
	return problems, rows
}

// cultureStrengthProblems is StaticCultureStrength's walk over a model and a
// rule (the culture held and the town's run totals): the check's
// broken-number tests feed in other rules.
func cultureStrengthProblems(incomes []faithIncome, strengthOf func(held float64, m game.CultureSave) float64) ([]CultureStrengthProblem, []CultureStrengthRow) {
	in := func(tier game.CultureTier) func(float64) bool {
		return func(strength float64) bool { return game.CultureTierAt(strength) == tier }
	}
	towns := []strengthTown{
		{0, CultureRuleModerate, in(game.CultureTierMinor)},
		{1, CultureRuleModerate, in(game.CultureTierMinor)},
		{CultureTwiceSets, CultureRuleTwice, in(game.CultureTierMajor)},
		{CultureThriceSets, "", nil}, // reported, not held to a tier
		{CultureTopSets, CultureRuleTop, in(game.CultureTierLegendary)},
	}
	walked, rows := strengthWalk(incomes, towns, func(held float64, m game.FaithSave, _ faithIncome) float64 {
		return strengthOf(held, m)
	})
	var problems []CultureStrengthProblem
	for _, p := range walked {
		problems = append(problems, CultureStrengthProblem{p})
	}
	// Ages before culture have nothing to show.
	kept := rows[:0]
	for _, r := range rows {
		if r.ModerateSet > 0 {
			kept = append(kept, r)
		}
	}
	return problems, kept
}

// writeCultureStrength renders the check for the static scenario.
func writeCultureStrength(sb *strings.Builder, problems []CultureStrengthProblem, rows []CultureStrengthRow) {
	fmt.Fprintf(sb, "Culture strength is the culture a town holds that its own culture buildings made, against %g moderate sets' worth (config.FlowCopies copies of every culture building): over %s opens Major good epoch events, over %s the Legendary one. Towns that never spend are walked through the run at the pacing targets, as for faith. A moderate town, and one with no culture buildings, must stay on Minor events; twice the culture buildings must open Major events; %g times, the Legendary.\n\n", game.CultureFullSets, sharePct(game.CultureMajorAbove), sharePct(game.CultureLegendaryAbove), CultureTopSets)
	if len(problems) == 0 {
		sb.WriteString("No problems.\n\n")
	} else {
		for _, p := range problems {
			fmt.Fprintf(sb, "- %s\n", p)
		}
		sb.WriteString("\n")
	}
	fmt.Fprintf(sb, "| era | age | a moderate set makes, per tick | wonders and techs make | no culture buildings | moderate | twice | three times | %g times |\n|---|---|---|---|---|---|---|---|---|\n", CultureTopSets)
	for _, r := range rows {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n", r.Epoch, r.Age, faithAmount(r.ModerateSet), faithAmount(r.Given),
			cultureSpan(r.None), cultureSpan(r.Moderate), cultureSpan(r.Twice), cultureSpan(r.Thrice), cultureSpan(r.Top))
	}
}
