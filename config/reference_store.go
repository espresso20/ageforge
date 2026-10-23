package config

import (
	"math"

	"github.com/espresso20/ageforge/detmath"
)

// The reference store: the size a store would have, were it still sized, as a
// pure function of the age. Storage buildings no longer raise any store (they
// boost production, storage_boost.go), but a few rules written in terms of a
// store still need a measure, and they read this one until they are redone:
// the festival and Black Market prices (a share of the culture store), deals,
// the Brace ceiling (TypicalStorages), the Maximalist badges and the event
// stock shares. It is the store master had, so those rules behave as they did.
// S3 of the no-caps work finds them by searching for ReferenceStore.
//
// It is the old storage rule, kept as arithmetic and nothing else. The
// reference town holds TownCopies copies of every building that is not
// storage, or the gate's count where the gate asks for more; of the age's
// storage building it holds StorageRuleCopies. For each age, in order:
//
//  1. The dearest ask is the largest single price, in any one resource, the
//     reference town pays in the age: the last copy of each building other
//     than storage, at its reference count. A wonder is not counted.
//  2. StorageRuleCopies copies of the age's storage building, on top of what
//     the earlier ages hold, hold the dearest ask and referenceRoom more.
//  3. Each copy up to the last fits under the store the copies before it give,
//     with referenceMargin to spare.
//  4. A town arrives in the next age holding the first copy of that age's
//     storage building and of its house, with referenceMargin to spare. If the
//     next age has no storage building of its own, this age's store holds what
//     that age asks instead.
//  5. The size is the smallest that satisfies 2 to 4, never under
//     referenceFloor of what rule 2 asks in all, rounded up to two figures.
const (
	referenceRoom   = 1.25
	referenceMargin = 1.05
	referenceFloor  = 0.05
)

// storageRuleCount is how many copies of d, a building that is not storage,
// the reference town holds.
func storageRuleCount(d BuildingDef, gate map[string]int) int {
	if IsKnowledgeBuilding(d) {
		// The town holds the copies the gate asks, at the spend the
		// knowledge rule keeps (knowledge_rule.go), not TownCopies.
		return gate[d.Key]
	}
	return max(TownCopies, gate[d.Key])
}

func maxPrice(d BuildingDef, owned int) float64 {
	f := detmath.Pow(d.CostScale, float64(owned))
	top := 0.0
	for r, v := range d.BaseCost {
		if r == "knowledge" {
			continue // knowledge has a store of its own (knowledge_rule.go)
		}
		top = math.Max(top, math.Max(1, math.Floor(v*f)))
	}
	return top
}

func roundUpSignificant(v float64, sig int) float64 {
	if v <= 0 {
		return 0
	}
	mag := detmath.Pow(10, math.Floor(detmath.Log10(v))-float64(sig-1))
	return math.Ceil(v/mag-1e-9) * mag
}

// ReferenceStoreSizes works the reference store out: the size of each age's
// storage building (what one copy held), by building key, and the reference
// store on leaving each age, by age key. baseStore is what a resource holds
// before any building.
func ReferenceStoreSizes(defs []BuildingDef, ages []AgeDef, baseStore float64) (sizes map[string]float64, store map[string]float64) {
	n := len(ages)
	idx := make(map[string]int, n)
	for i, a := range ages {
		idx[a.Key] = i
	}
	byAge := make([][]BuildingDef, n)
	for _, d := range defs {
		if i, ok := idx[d.RequiredAge]; ok && d.Category != "wonder" && d.Category != "monument" {
			byAge[i] = append(byAge[i], d)
		}
	}
	gate := func(i int) map[string]int {
		if i+1 < n {
			return ages[i+1].BuildingReqs
		}
		return nil
	}
	find := func(i int, category string) (BuildingDef, bool) {
		for _, d := range byAge[i] {
			if d.Category == category {
				return d, true
			}
		}
		return BuildingDef{}, false
	}
	dearest := func(i int) float64 {
		top := 0.0
		for _, d := range byAge[i] {
			if d.Category != "storage" {
				top = math.Max(top, maxPrice(d, storageRuleCount(d, gate(i))-1))
			}
		}
		return top
	}
	sizes = map[string]float64{}
	store = map[string]float64{}
	prev := baseStore
	for i := range ages {
		b, ok := find(i, "storage")
		if !ok {
			store[ages[i].Key] = prev
			continue
		}
		asked := referenceRoom * dearest(i)
		need := asked
		if i+1 < n {
			if nb, has := find(i+1, "storage"); has {
				need = math.Max(need, referenceMargin*maxPrice(nb, 0))
				if h, hasHouse := find(i+1, "housing"); hasHouse {
					need = math.Max(need, referenceMargin*maxPrice(h, 0))
				}
			} else {
				need = math.Max(need, referenceRoom*dearest(i+1))
				for _, d := range byAge[i+1] {
					need = math.Max(need, referenceMargin*maxPrice(d, 0))
				}
			}
		}
		size := math.Max((need-prev)/StorageRuleCopies, referenceFloor*asked/StorageRuleCopies)
		for k := 2; k <= StorageRuleCopies; k++ {
			size = math.Max(size, (float64(referenceMargin*maxPrice(b, k-1))-prev)/float64(k-1))
		}
		size = roundUpSignificant(size, 2)
		sizes[b.Key] = size
		prev += float64(StorageRuleCopies * size)
		store[ages[i].Key] = prev
	}
	return sizes, store
}

// ReferenceBaseStore is the base store the reference store is worked out on:
// wood's, the one every store was sized by (a resource with a lower base
// holds up to 20 less before any building).
func ReferenceBaseStore() float64 {
	for _, r := range BaseResources() {
		if r.Key == "wood" {
			return r.BaseStorage
		}
	}
	return 0
}
