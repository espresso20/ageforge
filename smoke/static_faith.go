package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/game"
	"github.com/espresso20/ageforge/rules"
)

// The faith strength check: towns that never spend their faith are walked
// through the whole run at the pacing targets, hoard carried from age to
// age, and read by the rule the rolls read (game.FaithStrengthOf) at the
// moments a roll can fall: on entering each age, at the end of a doom's
// shortest warning, halfway, and on leaving it (a doom's strike, the Last
// Passage's roll, the epoch roll at an advance).
//
//   - A moderate town (config.FlowCopies fully staffed copies of every faith
//     building, the economy Appease is priced on) must sit in the bottom
//     band throughout: those are the odds the game was tuned and measured
//     at. So must a town with no faith buildings at all, whatever its
//     wonders hand it.
//   - A town with twice the faith buildings must sit in the middle band.
//   - A town with FaithTopSets times the faith buildings must sit in the top
//     band: the top is reachable for a little under three and a half times
//     the moderate set.
//
// Every town has the same wonders and techs, and each age is walked twice:
// with its own wonder standing from the first tick, and with it going up
// only as the age ends. Stonehenge makes 0.6 faith a tick where a moderate
// Bronze Age set makes 0.07, and every player must build it to advance, so
// a measure that let a wonder's faith count would hand its builder a band.
//
// The check fails in either direction: a moderate town that leaves the
// bottom band (the first rule this replaced read faith against one age's
// income, and a hoard carried in from earlier ages put every saver in the
// top band), or a devoted one that cannot reach its band (the rule before
// that read the fill of a store faith could never fill).

// The towns the check walks, in moderate sets of faith buildings.
const (
	FaithTwiceSets  = 2.0
	FaithThriceSets = 3.0
	// FaithTopSets is the devotion the top band must be reachable at.
	FaithTopSets = 3.5
)

// faithMoments are the points of an age the check reads a town at, as a
// share of its pacing target: entering it, the end of a doom's shortest
// warning (a fifth of the age), halfway, and leaving it.
var faithMoments = []struct {
	name string
	at   float64
}{{"on entering", 0}, {"a fifth in", 0.2}, {"halfway", 0.5}, {"on leaving", 1}}

// FaithSpan is the lowest and highest faith strength a town read in an age,
// over the check's moments and both timings of the age's wonder.
type FaithSpan struct {
	Low  float64 `json:"low"`
	High float64 `json:"high"`
}

func (s *FaithSpan) add(v float64, first bool) {
	if first || v < s.Low {
		s.Low = v
	}
	if first || v > s.High {
		s.High = v
	}
}

// String is the span as the report prints it: "22% (bottom)", or both ends
// when they differ.
func (s FaithSpan) String() string {
	lo, hi := faithBandWord(game.FaithBandAt(s.Low)), faithBandWord(game.FaithBandAt(s.High))
	if sharePct(s.Low) == sharePct(s.High) && lo == hi {
		return fmt.Sprintf("%s (%s)", sharePct(s.Low), lo)
	}
	return fmt.Sprintf("%s (%s) to %s (%s)", sharePct(s.Low), lo, sharePct(s.High), hi)
}

// faithBandWord names a band as the report and the wiki do.
func faithBandWord(b game.FaithBand) string {
	switch b {
	case game.FaithBandLow:
		return "bottom"
	case game.FaithBandHigh:
		return "top"
	}
	return "middle"
}

// FaithStrengthRow is one age of the faith strength check.
type FaithStrengthRow struct {
	Epoch string `json:"epoch"`
	Age   string `json:"age"`
	// ModerateSet is what the moderate set of faith buildings makes per tick
	// in the age, and Given what the wonders and techs every town has make
	// (with the age's own wonder standing), both at the model's bonuses.
	ModerateSet float64 `json:"moderate_set_per_tick"`
	Given       float64 `json:"given_per_tick"`
	// The towns: no faith buildings, the moderate set, twice, three times
	// and FaithTopSets times it.
	None     FaithSpan `json:"no_faith_buildings"`
	Moderate FaithSpan `json:"moderate"`
	Twice    FaithSpan `json:"twice"`
	Thrice   FaithSpan `json:"three_times"`
	Top      FaithSpan `json:"top_sets"`
}

// The rules a FaithStrengthProblem can break.
const (
	// FaithRuleModerate: a moderate town, or one with no faith buildings,
	// reads above the bottom band.
	FaithRuleModerate = "moderate"
	// FaithRuleTwice: a town with twice the faith buildings is not in the
	// middle band.
	FaithRuleTwice = "twice"
	// FaithRuleTop: a town with FaithTopSets times the faith buildings is not
	// in the top band.
	FaithRuleTop = "top"
)

// FaithStrengthProblem is one moment at which a town reads the wrong band.
type FaithStrengthProblem struct {
	Epoch string `json:"epoch"`
	Age   string `json:"age"`
	Rule  string `json:"rule"`
	// Sets is the town's faith buildings in moderate sets, Moment where in
	// the age it was read, WonderEarly whether the age's wonder stood from
	// its first tick.
	Sets        float64 `json:"sets"`
	Moment      string  `json:"moment"`
	WonderEarly bool    `json:"wonder_early"`
	Strength    float64 `json:"strength"`
}

// String says what is wrong in plain words.
func (p FaithStrengthProblem) String() string {
	town := fmt.Sprintf("a town with %g times the moderate set of faith buildings", p.Sets)
	switch p.Sets {
	case 0:
		town = "a town with no faith buildings"
	case 1:
		town = "a moderate town"
	}
	want := "the bottom band"
	switch p.Rule {
	case FaithRuleTwice:
		want = "the middle band"
	case FaithRuleTop:
		want = "the top band"
	}
	wonder := "the age's wonder up as it ends"
	if p.WonderEarly {
		wonder = "the age's wonder up from its first tick"
	}
	return fmt.Sprintf("%s, %s (%s): %s that never spends its faith reads %s, the %s band; want %s", p.Age, p.Moment, wonder, town, sharePct(p.Strength), faithBandWord(game.FaithBandAt(p.Strength)), want)
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

// StaticFaithStrength walks the check's towns through the core ruleset by
// the rule the rolls read, and returns the problems with the table they come
// from.
func StaticFaithStrength() ([]FaithStrengthProblem, []FaithStrengthRow) {
	return faithStrengthProblems(faithIncomes(rules.Core()), func(held float64, m game.FaithSave, _ faithIncome) float64 {
		return game.FaithStrengthOf(held, m)
	})
}

// faithIncome is one age of the model the check walks: what a moderate set
// of faith buildings makes per tick, what the wonders and techs every town
// has by then make (the wonders of earlier ages and every tech up to the
// age), and what the age's own wonder adds once it stands, all at the
// model's production bonus (config.Incomes), over the age's pacing target.
type faithIncome struct {
	epoch, age              string
	ticks                   float64
	set, given, givenWonder float64
}

// faithIncomes reads the model from set: the moderate set is
// FlowBuildingOutput at the model's bonus, and the rest of FlowIncome is
// what every town is given.
func faithIncomes(set *rules.Set) []faithIncome {
	wonder := map[string]float64{} // age -> faith of the wonder built there, before bonuses
	for _, d := range set.Buildings() {
		if d.Category != "wonder" {
			continue
		}
		for _, e := range d.Effects {
			if e.Type == "production" && e.Target == "faith" && e.Value > 0 {
				wonder[d.RequiredAge] += e.Value
			}
		}
	}
	var out []faithIncome
	for _, ep := range set.Eras() {
		for _, a := range ep.Ages {
			bonus := faithModelBonus(set, a)
			in := faithIncome{epoch: ep.Key, age: a, ticks: set.TargetTicks(a),
				set: float64(set.FlowBuildingOutput("faith", a) * bonus), givenWonder: float64(wonder[a] * bonus)}
			if in.given = set.FlowIncome("faith", a) - in.set; in.given < 0 {
				in.given = 0
			}
			out = append(out, in)
		}
	}
	return out
}

// faithModelBonus is the production bonus config.Incomes applies in age:
// every tech's all-production up to the age and every earlier age's
// wonder's, capped as the engine caps it.
func faithModelBonus(set *rules.Set, age string) float64 {
	idx := set.Indexes()
	bonus := 0.0
	for _, d := range set.Buildings() {
		if j, ok := idx[d.RequiredAge]; !ok || j >= idx[age] || d.Category != "wonder" {
			continue
		}
		for _, e := range d.Effects {
			if e.Type == "bonus" && e.Target == "production_all" {
				bonus += e.Value
			}
		}
	}
	for _, t := range set.Techs() {
		if j, ok := idx[t.Age]; !ok || j > idx[age] {
			continue
		}
		for _, e := range t.Effects {
			if e.Kind == config.EffectAllOutput {
				bonus += e.Value
			}
		}
	}
	if m := 1 + bonus; m < config.ProductionAllCap {
		return m
	}
	return config.ProductionAllCap
}

// faithStrengthProblems is StaticFaithStrength over a model and a rule (the
// faith held, the town's run totals and the age it is read in): the check's
// broken-number tests feed in the measures this one replaced.
func faithStrengthProblems(incomes []faithIncome, strengthOf func(held float64, m game.FaithSave, in faithIncome) float64) ([]FaithStrengthProblem, []FaithStrengthRow) {
	towns := []struct {
		sets float64
		rule string
		want game.FaithBand
	}{
		{0, FaithRuleModerate, game.FaithBandLow},
		{1, FaithRuleModerate, game.FaithBandLow},
		{FaithTwiceSets, FaithRuleTwice, game.FaithBandMid},
		{FaithThriceSets, "", ""}, // reported, not held to a band
		{FaithTopSets, FaithRuleTop, game.FaithBandHigh},
	}
	var problems []FaithStrengthProblem
	rows := make([]FaithStrengthRow, len(incomes))
	for i, in := range incomes {
		rows[i] = FaithStrengthRow{Epoch: in.epoch, Age: in.age, ModerateSet: in.set, Given: in.given + in.givenWonder}
	}
	for ti, town := range towns {
		// The town's totals on entering each age: every earlier age in full.
		carried := game.FaithSave{}
		for i, in := range incomes {
			spans := []*FaithSpan{&rows[i].None, &rows[i].Moderate, &rows[i].Twice, &rows[i].Thrice, &rows[i].Top}
			first := true
			for _, wonderEarly := range []bool{true, false} {
				for _, mom := range faithMoments {
					span := float64(in.ticks * mom.at)
					m := carried
					m.Moderate += float64(in.set * span)
					m.Own += float64(float64(town.sets*in.set) * span)
					m.Other += float64(in.given * span)
					if wonderEarly {
						m.Other += float64(in.givenWonder * span)
					}
					if m.Moderate <= 0 {
						continue // the run's first tick: nothing to read yet
					}
					// It has spent nothing: it holds all its income made.
					strength := strengthOf(m.Own+m.Other, m, in)
					spans[ti].add(strength, first)
					first = false
					if town.rule != "" && game.FaithBandAt(strength) != town.want {
						problems = append(problems, FaithStrengthProblem{Epoch: in.epoch, Age: in.age, Rule: town.rule,
							Sets: town.sets, Moment: mom.name, WonderEarly: wonderEarly, Strength: strength})
					}
				}
			}
			carried.Moderate += float64(in.set * in.ticks)
			carried.Own += float64(float64(town.sets*in.set) * in.ticks)
			// The age's wonder went up during it: carried as standing for half.
			carried.Other += float64(in.given*in.ticks) + float64(float64(in.givenWonder*in.ticks)*0.5)
		}
	}
	return problems, rows
}

// writeFaithStrength renders the check for the static scenario.
func writeFaithStrength(sb *strings.Builder, problems []FaithStrengthProblem, rows []FaithStrengthRow) {
	fmt.Fprintf(sb, "Faith strength is the faith a town holds that its own faith buildings made, against %g moderate sets' worth (config.FlowCopies fully staffed copies of every faith building): under %s is the bottom band, over %s the top. Towns that never spend are walked through the run at the pacing targets, hoards carried from age to age, and read on entering each age, a fifth in, halfway and on leaving, with the age's wonder up from its first tick and with it up only at the end. A moderate town, and one with no faith buildings, must stay in the bottom band; twice the faith buildings must read the middle band; %g times, the top band.\n\n", game.FaithFullSets, sharePct(game.FaithMidAt), sharePct(game.FaithHighAbove), FaithTopSets)
	if len(problems) == 0 {
		sb.WriteString("No problems.\n\n")
	} else {
		for _, p := range problems {
			fmt.Fprintf(sb, "- %s\n", p)
		}
		sb.WriteString("\n")
	}
	fmt.Fprintf(sb, "| era | age | a moderate set makes, per tick | wonders and techs make | no faith buildings | moderate | twice | three times | %g times |\n|---|---|---|---|---|---|---|---|---|\n", FaithTopSets)
	for _, r := range rows {
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n", r.Epoch, r.Age, faithAmount(r.ModerateSet), faithAmount(r.Given),
			r.None, r.Moderate, r.Twice, r.Thrice, r.Top)
	}
}

// faithAmount prints a faith amount: num, with decimals kept for the first
// ages, where a tick makes a fraction of a faith.
func faithAmount(v float64) string {
	if v > 0 && v < 100 {
		return fmt.Sprintf("%.3g", v)
	}
	return num(v)
}
