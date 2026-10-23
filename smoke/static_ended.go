package smoke

import (
	"fmt"
	"sort"

	"github.com/espresso20/ageforge/config"
)

// The Ended-Resources check: a building is never priced in a resource nobody
// can make any more. A resource R is available in age A when some building
// with RequiredAge <= A makes R (a production effect, main or side output,
// wonders included) and A is at most one age after the latest such maker's
// age. Food, faith, culture, knowledge and soldiers are always available. Written out
// again here from the tables, not calling config.EndedParts.

// EndedProblem is one price part charging a resource that has ended.
type EndedProblem struct {
	Building, Age, Resource string
	Amount                  float64
}

func (p EndedProblem) String() string {
	return fmt.Sprintf("%s (%s) is priced in %s (%s), which nothing in the age makes", p.Building, p.Age, p.Resource, num(p.Amount))
}

// StaticEnded lists every price part whose resource is not available in the
// building's age.
func StaticEnded() []EndedProblem {
	defs := config.BaseBuildings()
	pos := map[string]int{}
	for i, a := range config.Ages() {
		pos[a.Key] = i
	}
	makes := map[string]map[int]bool{}
	for _, d := range defs {
		p, ok := pos[d.RequiredAge]
		if !ok {
			continue
		}
		for _, e := range d.Effects {
			if e.Type != "production" || e.Value <= 0 {
				continue
			}
			if makes[e.Target] == nil {
				makes[e.Target] = map[int]bool{}
			}
			makes[e.Target][p] = true
		}
	}
	always := map[string]bool{"food": true, "faith": true, "culture": true, "knowledge": true}
	var out []EndedProblem
	for _, d := range defs {
		p, ok := pos[d.RequiredAge]
		if !ok {
			continue
		}
		keys := make([]string, 0, len(d.BaseCost))
		for res := range d.BaseCost {
			keys = append(keys, res)
		}
		sort.Strings(keys)
		for _, res := range keys {
			if d.BaseCost[res] <= 0 || always[res] {
				continue
			}
			ok := false
			for q := p; q >= 0 && q >= p-1; q-- {
				if makes[res][q] {
					ok = true
					break
				}
			}
			if !ok {
				out = append(out, EndedProblem{d.Key, d.RequiredAge, res, d.BaseCost[res]})
			}
		}
	}
	return out
}
