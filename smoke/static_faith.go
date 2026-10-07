package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// The faith strength check: in every age, a moderate faith economy
// (config.FlowIncome) that starts saving when a doom's harbinger comes and
// saves through the shortest warning must reach the middle faith band by the
// rule the rolls read (game.FaithBandAt against game.FaithFullIn),
// and a devoted one, FaithDevotedFactor times the moderate, must reach the
// top band.
//
// The bands used to read faith as a share of its storage. Faith has no store
// of its own: it is kept in the general one, which is sized to construction
// resources, so against it no economy left the bottom band in any era and
// every doom and every Last Passage rolled at its worst odds
// (TestFaithNeverFilledTheGeneralStore keeps the arithmetic). This is the
// check that fails if a later change makes a band unreachable again.

// FaithDevotedFactor is how much more faith a devoted economy makes than the
// moderate one: three times, fifteen staffed copies of every faith building
// where the moderate economy keeps five (config.FlowCopies).
const FaithDevotedFactor = 3.0

// FaithStrengthRow is one age of the faith strength check.
type FaithStrengthRow struct {
	Epoch string `json:"epoch"`
	Age   string `json:"age"`
	// Full is the faith that reads as full strength in the age.
	Full float64 `json:"full_strength"`
	// WarningTicks is the shortest warning a doom gives there, in ticks at 1x.
	WarningTicks float64 `json:"shortest_warning_ticks"`
	// Moderate is what a moderate faith economy makes in that warning;
	// ModerateStrength and DevotedStrength are it, and FaithDevotedFactor
	// times it, as a share of Full (not capped at 1).
	Moderate         float64 `json:"moderate_faith"`
	ModerateStrength float64 `json:"moderate_strength"`
	ModerateBand     string  `json:"moderate_band"`
	DevotedStrength  float64 `json:"devoted_strength"`
	DevotedBand      string  `json:"devoted_band"`
}

// The rules a FaithStrengthProblem can break.
const (
	// FaithRuleMeasure: the age has no measure of faith (nothing makes any).
	FaithRuleMeasure = "measure"
	// FaithRuleMiddle: a moderate economy saving through the shortest warning
	// stays in the bottom band.
	FaithRuleMiddle = "middle"
	// FaithRuleTop: a devoted one does not reach the top band.
	FaithRuleTop = "top"
)

// FaithStrengthProblem is one age in which a faith band cannot be reached.
type FaithStrengthProblem struct {
	Epoch    string  `json:"epoch"`
	Age      string  `json:"age"`
	Rule     string  `json:"rule"`
	Strength float64 `json:"strength"`
}

// String says what is wrong in plain words.
func (p FaithStrengthProblem) String() string {
	switch p.Rule {
	case FaithRuleMeasure:
		return fmt.Sprintf("%s: nothing measures faith strength (no faith is made by then)", p.Age)
	case FaithRuleTop:
		return fmt.Sprintf("%s: a devoted faith economy (%gx the moderate one) saving through a doom's shortest warning reaches %s of full strength, not the top band (over %s)", p.Age, FaithDevotedFactor, sharePct(p.Strength), sharePct(game.FaithHighAbove))
	}
	return fmt.Sprintf("%s: a moderate faith economy saving through a doom's shortest warning reaches %s of full strength, under the middle band (%s)", p.Age, sharePct(p.Strength), sharePct(game.FaithMidAt))
}

// sharePct prints a share as a percentage: whole from 10% up, with enough
// figures to tell a tiny one from nothing below that.
func sharePct(share float64) string {
	v := share * 100
	switch {
	case v >= 10 || v == 0:
		return fmt.Sprintf("%.0f%%", v)
	case v >= 0.1:
		return fmt.Sprintf("%.1f%%", v)
	}
	return fmt.Sprintf("%.2g%%", v)
}

// StaticFaithStrength checks every age's faith bands on the core ruleset and
// returns the problems with the table they come from.
func StaticFaithStrength() ([]FaithStrengthProblem, []FaithStrengthRow) {
	return faithStrengthProblems(game.FaithMeasuresIn(rules.Core()), rules.Core().FlowIncome, FaithDevotedFactor)
}

// faithStrengthProblems is StaticFaithStrength over measures, income (the
// per-tick income of a resource in an age at a moderate economy) and the
// devoted economy's factor: the check's broken-number tests feed in others.
func faithStrengthProblems(measures []game.FaithMeasure, income func(res, age string) float64, devotedFactor float64) ([]FaithStrengthProblem, []FaithStrengthRow) {
	var problems []FaithStrengthProblem
	rows := make([]FaithStrengthRow, 0, len(measures))
	for _, m := range measures {
		row := FaithStrengthRow{Epoch: m.Epoch, Age: m.Age, Full: m.Full, WarningTicks: m.WarningTicks,
			Moderate: float64(income("faith", m.Age) * m.WarningTicks)}
		measured := m.Full > 0
		if measured {
			row.ModerateStrength = row.Moderate / m.Full
			row.DevotedStrength = float64(row.Moderate*devotedFactor) / m.Full
		}
		moderate, devoted := game.FaithBandAt(row.ModerateStrength, measured), game.FaithBandAt(row.DevotedStrength, measured)
		row.ModerateBand, row.DevotedBand = string(moderate), string(devoted)
		rows = append(rows, row)
		add := func(rule string, strength float64) {
			problems = append(problems, FaithStrengthProblem{Epoch: m.Epoch, Age: m.Age, Rule: rule, Strength: strength})
		}
		switch {
		case !measured:
			add(FaithRuleMeasure, 0)
		case moderate == game.FaithBandLow:
			add(FaithRuleMiddle, row.ModerateStrength)
		case devoted != game.FaithBandHigh:
			add(FaithRuleTop, row.DevotedStrength)
		}
	}
	return problems, rows
}

// writeFaithStrength renders the check for the static scenario.
func writeFaithStrength(sb *strings.Builder, problems []FaithStrengthProblem, rows []FaithStrengthRow) {
	fmt.Fprintf(sb, "Faith strength is the faith held as a share of what a moderate faith economy (config.FlowIncome) makes in three fifths of the age: under %s is the bottom band, over %s the top. In every age a moderate economy that saves through a doom's shortest warning must reach the middle band, and a devoted one (%gx the moderate) the top.\n\n", sharePct(game.FaithMidAt), sharePct(game.FaithHighAbove), FaithDevotedFactor)
	if len(problems) == 0 {
		sb.WriteString("No problems.\n\n")
	} else {
		for _, p := range problems {
			fmt.Fprintf(sb, "- %s\n", p)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("| era | age | full strength | moderate, shortest warning | strength | devoted, shortest warning | strength |\n|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s (%s) | %s | %s (%s) |\n", r.Epoch, r.Age, faithAmount(r.Full), faithAmount(r.Moderate),
			sharePct(r.ModerateStrength), r.ModerateBand, faithAmount(float64(r.Moderate*FaithDevotedFactor)), sharePct(r.DevotedStrength), r.DevotedBand)
	}
}

// faithAmount prints a faith amount: num, with a decimal kept for the first
// ages, where a warning makes less than ten faith.
func faithAmount(v float64) string {
	if v > 0 && v < 100 {
		return fmt.Sprintf("%.3g", v)
	}
	return num(v)
}
