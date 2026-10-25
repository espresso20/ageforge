package smoke

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/detmath"
)

// The knowledge ladder (R4): tech prices are set against what the knowledge
// buildings make. For each age, the ordinary reference player's knowledge
// buildings (the copies the gates ask for, which is the cap, with their
// scholars all researching) make, during his stay, a quantity of work: the
// town's research income over the stay (the quantity property 10 reads); the
// age's techs are priced to total config.ResearchShare of it (0.8 in the
// first age, 1.2 from the Stone Age on: research simmers). Inside the age
// the prices climb by one ratio, KnowledgeLadderRatio, from the cheapest to
// the dearest, in the order the age's techs are priced in the table
// (cheapest first, a tie in table order), and every price is written to
// three significant figures, the table's usual. A tech's price is the work it
// takes; there is no store for it to fit in.
const KnowledgeLadderRatio = 1.3

// KnowledgeLadder is one age of the ladder.
type KnowledgeLadder struct {
	Age string
	// Techs is the age's techs, cheapest first in the order they are priced.
	Techs []string
	// Made is what the ordinary player's knowledge buildings make over his
	// stay (income per tick, times the stay's ticks); Total is the share of
	// it the techs are priced to.
	Made, Total float64
	Ratio       float64
	Prices      []float64
}

// roundSig3 rounds v to three significant figures.
func roundSig3(v float64) float64 {
	if v <= 0 {
		return 0
	}
	mag := detmath.Pow(10, math.Floor(detmath.Log10(v))-2)
	return math.Round(v/mag) * mag
}

// ladderPrices is n prices that climb by q from the first to the last and
// total total, each to three significant figures.
func ladderPrices(total float64, n int, q float64) []float64 {
	out := make([]float64, n)
	if n == 0 {
		return out
	}
	if n == 1 {
		out[0] = roundSig3(total)
		return out
	}
	first := float64(total*(q-1)) / (detmath.Pow(q, float64(n)) - 1)
	for j := range out {
		out[j] = roundSig3(float64(first * detmath.Pow(q, float64(j))))
	}
	return out
}

// KnowledgeMade is, for each age, what the ordinary reference player's
// knowledge buildings make during the age's stay: the quantity property 10
// divides the age's techs by.
func KnowledgeMade() []float64 {
	t := newRefTables()
	n := len(t.ages)
	refIncome := make([]float64, n)
	ref := make([]townState, n)
	for i := range t.ages {
		ref[i] = t.state(t.refTown(i), i, 1)
		refIncome[i] = units(ref[i].income, t.levels[i])
	}
	ordinary := t.walk("ordinary", 1, refIncome)
	out := make([]float64, n)
	for i := range t.ages {
		out[i] = float64(ref[i].income[config.ResearchTarget] * ordinary.Hours[i] * refTicksPerHour)
	}
	return out
}

// KnowledgeLadders works R4 out for every age, from the building tables, the
// gates and the tech order of the tables given. The ratio is the ladder's
// (the fit is reported by the caller through config.KnowledgeStoreRows).
func KnowledgeLadders(techs []config.TechDef, made []float64, ratio func(age string) float64) []KnowledgeLadder {
	ages := config.Ages()
	byAge := map[string][]int{}
	for i, tc := range techs {
		byAge[tc.Age] = append(byAge[tc.Age], i)
	}
	out := make([]KnowledgeLadder, len(ages))
	for i, a := range ages {
		idx := byAge[a.Key]
		sort.SliceStable(idx, func(x, y int) bool { return techs[idx[x]].Cost < techs[idx[y]].Cost })
		l := KnowledgeLadder{Age: a.Key, Made: made[i], Total: float64(config.ResearchShare(a.Key) * made[i]), Ratio: ratio(a.Key)}
		for _, k := range idx {
			l.Techs = append(l.Techs, techs[k].Key)
		}
		l.Prices = ladderPrices(l.Total, len(idx), l.Ratio)
		out[i] = l
	}
	return out
}
