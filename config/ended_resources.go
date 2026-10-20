package config

import (
	"fmt"
	"sort"
)

// Ended resources stay ended. When the game stops needing a resource it simply
// ends: a building is never priced in a resource nobody can make any more.
//
// The rule. A resource R is available in age A when some building with
// RequiredAge <= A makes R, and A is at most one age after the latest such
// maker's age. A building "makes" R when it has a production effect for R,
// main or side output (wonders count). So R is priced up to one age past its
// last maker (the cold-start bridge into the next age), never after. Food,
// faith, culture, knowledge and soldiers are always available and are left
// out of the check.

// EndedExempt are the resources the ended-resource check leaves alone.
var EndedExempt = map[string]bool{"food": true, "faith": true, "culture": true, "knowledge": true, "soldiers": true}

// EndedPart is one price part that charges a resource which is not available
// in the building's age.
type EndedPart struct {
	Building, Age, Resource string
	Amount                  float64
}

func (p EndedPart) String() string {
	return fmt.Sprintf("%s (%s): %s %g", p.Building, p.Age, p.Resource, p.Amount)
}

// EndedAvailability returns, per resource, a map of age key to whether the
// resource is available in that age (see the rule above).
func EndedAvailability(defs []BuildingDef, order []string) map[string]map[string]bool {
	pos := AgePositions(order)
	makes := map[string]map[int]bool{} // resource -> ages that have a maker
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
	out := map[string]map[string]bool{}
	for res, ages := range makes {
		out[res] = map[string]bool{}
		latest := -1 // latest maker age at or before i
		for i, age := range order {
			if ages[i] {
				latest = i
			}
			out[res][age] = latest >= 0 && i <= latest+1
		}
	}
	return out
}

// EndedParts lists every price part (wonders included) charging a resource
// that is not available in the building's age, in building-definition order
// then resource key order.
func EndedParts(defs []BuildingDef, ages []AgeDef) []EndedPart {
	order := ageKeys(ages)
	avail := EndedAvailability(defs, order)
	pos := AgePositions(order)
	var out []EndedPart
	for _, d := range defs {
		if _, ok := pos[d.RequiredAge]; !ok {
			continue
		}
		keys := make([]string, 0, len(d.BaseCost))
		for res := range d.BaseCost {
			keys = append(keys, res)
		}
		sort.Strings(keys)
		for _, res := range keys {
			amt := d.BaseCost[res]
			if amt <= 0 || EndedExempt[res] || avail[res][d.RequiredAge] {
				continue
			}
			out = append(out, EndedPart{Building: d.Key, Age: d.RequiredAge, Resource: res, Amount: amt})
		}
	}
	return out
}
