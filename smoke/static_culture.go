package smoke

import (
	"fmt"
	"strings"

	"github.com/espresso20/ageforge/config"
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
//     must stay on Minor events throughout: that is where every town stood
//     when the tier read the fill of culture's store. So must a town with no
//     culture buildings at all, whatever its wonders hand it.
//   - A town with twice the culture buildings must have Major events open.
//   - A town with CultureTopSets times the culture buildings must have the
//     Legendary tier open.
//
// Beside it, the rule this replaced, as it stood: culture is kept in the
// general store, so the old tier read a town's culture against that. OldFill
// is a moderate town's whole run of culture, never spent, on leaving the
// age, over the least general store the gate it has just met leaves it
// holding, and OldFillTop the same for CultureTopSets times the buildings:
// neither ever reached the 40% Major events asked. OldGoldTicks is how many ticks of a moderate economy's gold
// bought enough culture at the market to fill a typical store that far.

// The towns the check walks, in moderate sets of culture buildings.
const (
	CultureTwiceSets  = 2.0
	CultureThriceSets = 3.0
	// CultureTopSets is the devotion the Legendary tier must be open at.
	CultureTopSets = 3.5
)

// CultureStrengthRow is one age of the culture strength check.
type CultureStrengthRow struct {
	Epoch string `json:"epoch"`
	Age   string `json:"age"`
	// ModerateSet is what the moderate set of culture buildings makes per
	// tick in the age, and Given what the wonders and techs every town has
	// make, both at the model's bonuses.
	ModerateSet float64 `json:"moderate_set_per_tick"`
	Given       float64 `json:"given_per_tick"`
	// The towns: no culture buildings, the moderate set, twice, three times
	// and CultureTopSets times it.
	None     FaithSpan `json:"no_culture_buildings"`
	Moderate FaithSpan `json:"moderate"`
	Twice    FaithSpan `json:"twice"`
	Thrice   FaithSpan `json:"three_times"`
	Top      FaithSpan `json:"top_sets"`
	// The old rule (see the file comment): the fill it read on leaving the
	// age, and the gold that bought its way to 40% of a typical store.
	OldFill      float64 `json:"old_fill_moderate"`
	OldFillTop   float64 `json:"old_fill_top_sets"`
	OldGoldTicks float64 `json:"old_gold_ticks_to_major,omitempty"`
}

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
type CultureStrengthProblem struct {
	Epoch       string  `json:"epoch"`
	Age         string  `json:"age"`
	Rule        string  `json:"rule"`
	Sets        float64 `json:"sets"`
	Moment      string  `json:"moment"`
	WonderEarly bool    `json:"wonder_early"`
	Strength    float64 `json:"strength"`
}

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
	// The old rule's reading, on leaving each age.
	least, typical := leastGeneralStore(), typicalGeneralStore()
	hoard := map[string]float64{}
	sum := 0.0
	for _, a := range set.AgeKeys() {
		sum += float64(set.FlowIncome("culture", a) * set.TargetTicks(a))
		hoard[a] = sum
	}
	set1 := map[string]float64{} // a moderate set's own part of the hoard
	sum = 0
	for _, in := range flowIncomes(set, "culture") {
		sum += float64(in.set * in.ticks)
		set1[in.age] = sum
	}
	keys := set.AgeKeys()
	next := map[string]string{}
	for i, a := range keys {
		next[a] = a
		if i+1 < len(keys) {
			next[a] = keys[i+1]
		}
	}
	for i := range rows {
		a := rows[i].Age
		// On leaving the age the gate into the next one is met: the store
		// is at least what that gate forces.
		store := least[next[a]]
		rows[i].OldFill = hoard[a] / store
		rows[i].OldFillTop = (hoard[a] + float64((CultureTopSets-1)*set1[a])) / store
		if rate, ok := set.MarketOffers("gold", "culture", a); ok && rate > 0 && set.TypicalIncome("gold", a) > 0 {
			rows[i].OldGoldTicks = float64(game.CultureMajorAbove*typical[a]) / rate / set.TypicalIncome("gold", a)
		}
	}
	return problems, rows
}

// cultureStrengthProblems is StaticCultureStrength's walk over a model and a
// rule (the culture held and the town's run totals): the check's
// broken-number tests feed in other rules.
func cultureStrengthProblems(incomes []faithIncome, strengthOf func(held float64, m game.CultureSave) float64) ([]CultureStrengthProblem, []CultureStrengthRow) {
	towns := []struct {
		sets float64
		rule string
		want game.CultureTier
	}{
		{0, CultureRuleModerate, game.CultureTierMinor},
		{1, CultureRuleModerate, game.CultureTierMinor},
		{CultureTwiceSets, CultureRuleTwice, game.CultureTierMajor},
		{CultureThriceSets, "", ""}, // reported, not held to a tier
		{CultureTopSets, CultureRuleTop, game.CultureTierLegendary},
	}
	var problems []CultureStrengthProblem
	rows := make([]CultureStrengthRow, len(incomes))
	for i, in := range incomes {
		rows[i] = CultureStrengthRow{Epoch: in.epoch, Age: in.age, ModerateSet: in.set, Given: in.given + in.givenWonder}
	}
	for ti, town := range towns {
		carried := game.CultureSave{}
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
						continue // before culture: nothing to read yet
					}
					strength := strengthOf(m.Own+m.Other, m)
					spans[ti].add(strength, first)
					first = false
					if town.rule != "" && game.CultureTierAt(strength) != town.want {
						problems = append(problems, CultureStrengthProblem{Epoch: in.epoch, Age: in.age, Rule: town.rule,
							Sets: town.sets, Moment: mom.name, WonderEarly: wonderEarly, Strength: strength})
					}
				}
			}
			carried.Moderate += float64(in.set * in.ticks)
			carried.Own += float64(float64(town.sets*in.set) * in.ticks)
			carried.Other += float64(in.given*in.ticks) + float64(float64(in.givenWonder*in.ticks)*0.5)
		}
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

// leastGeneralStore is, per age, the least general storage a player can be
// holding on entering it: the base store plus the most any age gate so far
// has forced them to hold at once (ladderForced; storage is never lost).
// Faith and culture are kept in the general store.
func leastGeneralStore() map[string]float64 {
	ages, defs := config.Ages(), config.BuildingByKey()
	base := 0.0
	for _, r := range config.BaseResources() {
		if r.Key == "faith" {
			base = r.BaseStorage
		}
	}
	out := map[string]float64{}
	least := base
	for i, a := range ages {
		if i > 0 {
			if forced, _ := ladderForced(ages[i-1], a, defs); forced+base > least {
				least = forced + base
			}
		}
		out[a.Key] = least
	}
	return out
}

// typicalGeneralStore is, per age, the general storage of a moderate builder:
// the base store and config.FlowCopies copies of every storage building up
// to that age.
func typicalGeneralStore() map[string]float64 {
	idx := map[string]int{}
	for i, a := range config.AgeOrder() {
		idx[a] = i
	}
	out := map[string]float64{}
	for i, a := range config.AgeOrder() {
		total := 0.0
		for _, r := range config.BaseResources() {
			if r.Key == "faith" {
				total = r.BaseStorage
			}
		}
		for _, d := range config.BaseBuildings() {
			if j, ok := idx[d.RequiredAge]; !ok || j > i || d.Category != "storage" {
				continue
			}
			for _, e := range d.Effects {
				if e.Type == "storage" && (e.Target == "all" || e.Target == "faith") {
					total += float64(e.Value * config.FlowCopies)
				}
			}
		}
		out[a] = total
	}
	return out
}

// writeCultureStrength renders the check for the static scenario.
func writeCultureStrength(sb *strings.Builder, problems []CultureStrengthProblem, rows []CultureStrengthRow) {
	fmt.Fprintf(sb, "Culture strength is the culture a town holds that its own culture buildings made, against %g moderate sets' worth (config.FlowCopies copies of every culture building): over %s opens Major good epoch events, over %s the Legendary one. Towns that never spend are walked through the run at the pacing targets, as for faith. A moderate town, and one with no culture buildings, must stay on Minor events; twice the culture buildings must open Major events; %g times, the Legendary. The last three columns are the rule this replaced, which read the fill of culture's store (the general one): a never-spending town's whole run of culture over the least store its gates leave it holding, for a moderate town and for %g times the buildings, and how many ticks of a moderate economy's gold bought the 40%% it asked for at the market.\n\n", game.CultureFullSets, sharePct(game.CultureMajorAbove), sharePct(game.CultureLegendaryAbove), CultureTopSets, CultureTopSets)
	if len(problems) == 0 {
		sb.WriteString("No problems.\n\n")
	} else {
		for _, p := range problems {
			fmt.Fprintf(sb, "- %s\n", p)
		}
		sb.WriteString("\n")
	}
	fmt.Fprintf(sb, "| era | age | a moderate set makes, per tick | wonders and techs make | no culture buildings | moderate | twice | three times | %g times | old rule: moderate hoard, of the least store | old rule: %g times | old rule: gold to 40%%, ticks |\n|---|---|---|---|---|---|---|---|---|---|---|---|\n", CultureTopSets, CultureTopSets)
	for _, r := range rows {
		gold := "-"
		if r.OldGoldTicks > 0 {
			gold = fmt.Sprintf("%.0f", r.OldGoldTicks)
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n", r.Epoch, r.Age, faithAmount(r.ModerateSet), faithAmount(r.Given),
			cultureSpan(r.None), cultureSpan(r.Moderate), cultureSpan(r.Twice), cultureSpan(r.Thrice), cultureSpan(r.Top),
			sharePct(r.OldFill), sharePct(r.OldFillTop), gold)
	}
}
