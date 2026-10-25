package smoke

import (
	"math"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/config"
)

// The measured gate in the reference model (AGEFORGE_GATES=v2).
//
// The measures are read off the model, not written by hand. For each age
// that has a gate, the town the ordinary player has at the moment he meets
// today's gate (all the reference counts, the wonder paid) has some
// population and makes some income of the age's main materials; those two
// numbers are what the measured gate asks of any town (config.GateMeasure).
// The ordinary player's walk under the measured gate is then the same player
// on the same purchases, who starts paying the wonder as soon as the town
// reaches the measures and advances when it is paid.

// measured is what the measured gate reads of a town in age i: its
// population and its income of materials a tick. The population is what the
// game itself would show for the town (WorkerState.TotalPop): the workers,
// the reserve, the scholars and the soldiers, with the homes filled as the
// food allows (townState.population). Hand gathering in the first age is not
// the makers' income, so it is left out.
func (t *refTables) measured(st townState, i int, materials []string) (pop, income float64) {
	pop = st.population
	for _, res := range materials {
		income += st.income[res]
	}
	if i == 0 {
		for _, res := range materials {
			if res == "wood" {
				income -= RefHandGather
			}
		}
	}
	return pop, income
}

// refIncomes is the income of the reference town of each age, in the age's
// price units: what a walk counts doublings of income against.
func (t *refTables) refIncomes() []float64 {
	out := make([]float64, len(t.ages))
	for i := range t.ages {
		out[i] = units(t.state(t.refTown(i), i, 1).income, t.levels[i])
	}
	return out
}

// floorDigits rounds v down to n significant digits (v >= 0).
func floorDigits(v float64, n int) float64 {
	if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	s := strconv.FormatFloat(v, 'e', -1, 64)
	mant, expo, _ := strings.Cut(s, "e")
	digits := strings.Replace(mant, ".", "", 1)
	if len(digits) > n {
		digits = digits[:n]
	}
	out, _ := strconv.ParseFloat(digits[:1]+"."+digits[1:]+"e"+expo, 64)
	return out
}

// DeriveGateMeasures reads the measured gate off the reference model: for each
// age that has a gate, the population and the main materials' income the
// ordinary player's town has when he meets today's gate. The population is
// rounded down to a whole person and the income down to three significant
// digits, so that town itself meets its own measures.
func DeriveGateMeasures() []config.GateMeasure {
	t := newRefTables()
	walk := t.walk("ordinary", 1, t.refIncomes())
	all := config.BaseBuildings()
	techs := config.Technologies()
	levels := config.PriceLevelsByAge(all)
	var out []config.GateMeasure
	for i, a := range t.ages {
		if i+1 >= len(t.ages) {
			break
		}
		wonder := ""
		for _, key := range a.UnlockBuildings {
			for _, d := range all {
				if d.Key == key && d.Category == "wonder" && wonder == "" {
					wonder = key
				}
			}
		}
		mats := config.GateMainMaterials(a.Key, all, levels[a.Key])
		pop, income := t.measured(walk.Runs[i].end, i, mats)
		out = append(out, config.GateMeasure{
			Age:       a.Key,
			People:    int(math.Floor(pop)),
			Materials: mats,
			Income:    floorDigits(income, 3),
			KeyTech:   config.GateKeyTech(a.Key, wonder, all, techs),
		})
	}
	return out
}

// GatePacing is how long the ordinary player's stay in each age is under a
// gate rule, from the reference model.
type GatePacing struct {
	// Hours is the stay in each age, in order.
	Hours []float64
	// Names are the ages' names.
	Names []string
}

// ReferenceGatePacing walks the ordinary player through every age at speed 1
// under today's gate (measured false) or the measured gate (true, with the
// measures in config's table).
func ReferenceGatePacing(measured bool) GatePacing {
	t := newRefTables()
	if measured {
		t.measures = map[string]config.GateMeasure{}
		for _, m := range config.GateMeasures() {
			t.measures[m.Age] = m
		}
	}
	w := t.walk("ordinary", 1, t.refIncomes())
	p := GatePacing{Hours: w.Hours}
	for _, a := range t.ages {
		p.Names = append(p.Names, a.Name)
	}
	return p
}
