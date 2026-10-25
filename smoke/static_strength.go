package smoke

import "github.com/espresso20/ageforge/game"

// The walker the faith and culture strength checks share (static_faith.go,
// static_culture.go). Each check is a family: a flow resource, a rule that
// reads a strength from what a town holds and its run totals, and the band
// or tier each town must read. The walk, the rows and the problems are the
// same for both; what differs (the rule, the words) lives with the family.

// StrengthRow is one age of a strength check.
type StrengthRow struct {
	Epoch string `json:"epoch"`
	Age   string `json:"age"`
	// ModerateSet is what the moderate set of the family's buildings makes
	// per tick in the age, and Given what the wonders and techs every town
	// has make (with the age's own wonder standing), both at the model's
	// bonuses.
	ModerateSet float64 `json:"moderate_set_per_tick"`
	Given       float64 `json:"given_per_tick"`
	// The towns: no buildings of the family, the moderate set, twice, three
	// times and the top devotion times it.
	None     FaithSpan `json:"no_buildings"`
	Moderate FaithSpan `json:"moderate"`
	Twice    FaithSpan `json:"twice"`
	Thrice   FaithSpan `json:"three_times"`
	Top      FaithSpan `json:"top_sets"`
}

// spans is the row's five spans, in the order of the towns the walk reads.
func (r *StrengthRow) spans() []*FaithSpan {
	return []*FaithSpan{&r.None, &r.Moderate, &r.Twice, &r.Thrice, &r.Top}
}

// StrengthProblem is one moment at which a town reads the wrong band or tier.
type StrengthProblem struct {
	Epoch string `json:"epoch"`
	Age   string `json:"age"`
	Rule  string `json:"rule"`
	// Sets is the town's buildings of the family in moderate sets, Moment
	// where in the age it was read, WonderEarly whether the age's wonder
	// stood from its first tick.
	Sets        float64 `json:"sets"`
	Moment      string  `json:"moment"`
	WonderEarly bool    `json:"wonder_early"`
	Strength    float64 `json:"strength"`
}

// strengthTown is one of the towns a check walks: its buildings of the family
// in moderate sets, the rule it is held to ("" for a town that is only
// reported), and whether a strength is in the band or tier that rule wants.
type strengthTown struct {
	sets float64
	rule string
	ok   func(strength float64) bool
}

// strengthWalk walks every town through the model, in the ages' order: a town
// that never spends holds all its income made, and carries its totals from age
// to age. It reads the strength at each of faithMoments, with the age's own
// wonder standing from the first tick and with it going up only as the age
// ends, and returns the problems with the table they come from. A moment
// before the family's first building has made anything (the model's moderate
// set has made nothing yet) is skipped.
func strengthWalk(incomes []faithIncome, towns []strengthTown, strengthOf func(held float64, m game.FaithSave, in faithIncome) float64) ([]StrengthProblem, []StrengthRow) {
	var problems []StrengthProblem
	rows := make([]StrengthRow, len(incomes))
	for i, in := range incomes {
		rows[i] = StrengthRow{Epoch: in.epoch, Age: in.age, ModerateSet: in.set, Given: in.given + in.givenWonder}
	}
	for ti, town := range towns {
		// The town's totals on entering each age: every earlier age in full.
		carried := game.FaithSave{}
		for i, in := range incomes {
			spans := rows[i].spans()
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
						continue // the run's first tick, or before the family exists: nothing to read yet
					}
					// It has spent nothing: it holds all its income made.
					strength := strengthOf(m.Own+m.Other, m, in)
					spans[ti].add(strength, first)
					first = false
					if town.rule != "" && !town.ok(strength) {
						problems = append(problems, StrengthProblem{Epoch: in.epoch, Age: in.age, Rule: town.rule,
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
