package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// The faith standing check: in every age, a moderate faith economy
// (config.FlowIncome) that starts saving when a doom's harbinger comes and
// saves through the shortest warning must reach the middle faith band by the
// rule the rolls read (game.FaithBandAt against game.FaithStandingFullIn),
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

// FaithStandingRow is one age of the faith standing check.
type FaithStandingRow struct {
	Epoch string `json:"epoch"`
	Age   string `json:"age"`
	// Full is the faith that reads as full standing in the age.
	Full float64 `json:"full_standing"`
	// WarningTicks is the shortest warning a doom gives there, in ticks at 1x.
	WarningTicks float64 `json:"shortest_warning_ticks"`
	// Moderate is what a moderate faith economy makes in that warning;
	// ModerateStanding and DevotedStanding are it, and FaithDevotedFactor
	// times it, as a share of Full (not capped at 1).
	Moderate         float64 `json:"moderate_faith"`
	ModerateStanding float64 `json:"moderate_standing"`
	ModerateBand     string  `json:"moderate_band"`
	DevotedStanding  float64 `json:"devoted_standing"`
	DevotedBand      string  `json:"devoted_band"`
}

// The rules a FaithStandingProblem can break.
const (
	// FaithRuleMeasure: the age has no measure of faith (nothing makes any).
	FaithRuleMeasure = "measure"
	// FaithRuleMiddle: a moderate economy saving through the shortest warning
	// stays in the bottom band.
	FaithRuleMiddle = "middle"
	// FaithRuleTop: a devoted one does not reach the top band.
	FaithRuleTop = "top"
)

// FaithStandingProblem is one age in which a faith band cannot be reached.
type FaithStandingProblem struct {
	Epoch    string  `json:"epoch"`
	Age      string  `json:"age"`
	Rule     string  `json:"rule"`
	Standing float64 `json:"standing"`
}

// String says what is wrong in plain words.
func (p FaithStandingProblem) String() string {
	switch p.Rule {
	case FaithRuleMeasure:
		return fmt.Sprintf("%s: nothing measures faith standing (no faith is made by then)", p.Age)
	case FaithRuleTop:
		return fmt.Sprintf("%s: a devoted faith economy (%gx the moderate one) saving through a doom's shortest warning reaches %s of full standing, not the top band (over %s)", p.Age, FaithDevotedFactor, sharePct(p.Standing), sharePct(game.FaithHighAbove))
	}
	return fmt.Sprintf("%s: a moderate faith economy saving through a doom's shortest warning reaches %s of full standing, under the middle band (%s)", p.Age, sharePct(p.Standing), sharePct(game.FaithMidAt))
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

// StaticFaithStanding checks every age's faith bands on the core ruleset and
// returns the problems with the table they come from.
func StaticFaithStanding() ([]FaithStandingProblem, []FaithStandingRow) {
	return faithStandingProblems(game.FaithMeasuresIn(rules.Core()), rules.Core().FlowIncome, FaithDevotedFactor)
}

// faithStandingProblems is StaticFaithStanding over measures, income (the
// per-tick income of a resource in an age at a moderate economy) and the
// devoted economy's factor: the check's broken-number tests feed in others.
func faithStandingProblems(measures []game.FaithMeasure, income func(res, age string) float64, devotedFactor float64) ([]FaithStandingProblem, []FaithStandingRow) {
	var problems []FaithStandingProblem
	rows := make([]FaithStandingRow, 0, len(measures))
	for _, m := range measures {
		row := FaithStandingRow{Epoch: m.Epoch, Age: m.Age, Full: m.Full, WarningTicks: m.WarningTicks,
			Moderate: float64(income("faith", m.Age) * m.WarningTicks)}
		measured := m.Full > 0
		if measured {
			row.ModerateStanding = row.Moderate / m.Full
			row.DevotedStanding = float64(row.Moderate*devotedFactor) / m.Full
		}
		moderate, devoted := game.FaithBandAt(row.ModerateStanding, measured), game.FaithBandAt(row.DevotedStanding, measured)
		row.ModerateBand, row.DevotedBand = string(moderate), string(devoted)
		rows = append(rows, row)
		add := func(rule string, standing float64) {
			problems = append(problems, FaithStandingProblem{Epoch: m.Epoch, Age: m.Age, Rule: rule, Standing: standing})
		}
		switch {
		case !measured:
			add(FaithRuleMeasure, 0)
		case moderate == game.FaithBandLow:
			add(FaithRuleMiddle, row.ModerateStanding)
		case devoted != game.FaithBandHigh:
			add(FaithRuleTop, row.DevotedStanding)
		}
	}
	return problems, rows
}

// writeFaithStanding renders the check for the static scenario.
func writeFaithStanding(sb *strings.Builder, problems []FaithStandingProblem, rows []FaithStandingRow) {
	fmt.Fprintf(sb, "Faith standing is the faith held as a share of what a moderate faith economy (config.FlowIncome) makes in three fifths of the age: under %s is the bottom band, over %s the top. In every age a moderate economy that saves through a doom's shortest warning must reach the middle band, and a devoted one (%gx the moderate) the top.\n\n", sharePct(game.FaithMidAt), sharePct(game.FaithHighAbove), FaithDevotedFactor)
	if len(problems) == 0 {
		sb.WriteString("No problems.\n\n")
	} else {
		for _, p := range problems {
			fmt.Fprintf(sb, "- %s\n", p)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("| era | age | full standing | moderate, shortest warning | standing | devoted, shortest warning | standing |\n|---|---|---|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s (%s) | %s | %s (%s) |\n", r.Epoch, r.Age, faithAmount(r.Full), faithAmount(r.Moderate),
			sharePct(r.ModerateStanding), r.ModerateBand, faithAmount(float64(r.Moderate*FaithDevotedFactor)), sharePct(r.DevotedStanding), r.DevotedBand)
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
