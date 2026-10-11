package config

import (
	"math"
	"sort"

	"github.com/espresso20/ageforge/detmath"
)

// The storage rule: how large each age's storage building is, and what a
// gate asks to hold. Both are written in the tables (the storage buildings'
// own effects, the ages' ResourceReqs); this file is the derivation, and a
// test recomputes the tables from it and fails when an entry drifts.
//
// The reference town holds StorageRuleCopies copies of every building, or
// the gate's count where the gate asks for more. For each age, in order:
//
//  1. The dearest ask is the largest single price, in any one resource, the
//     reference town pays in the age: the last copy of each building other
//     than storage, at its reference count. A wonder is not counted: it is
//     paid into its bank and may be larger than a store.
//  2. Five copies of the age's storage building, on top of what the stores of
//     earlier ages hold, must hold the dearest ask and StorageRuleRoom more.
//  3. Each copy of the storage building up to the fifth must fit under the
//     store the copies before it give, with StorageRuleMargin to spare.
//  4. A town must arrive in the next age able to pay the first copy of that
//     age's storage building and of its house, with StorageRuleMargin to
//     spare. If the next age has no storage building of its own, this age's
//     store must hold what that age asks instead (rule 2 for it, and the
//     first copy of each of its buildings).
//  5. The size is the smallest that satisfies 2 to 4, never under
//     StorageRuleFloor of what rule 2 asks in all, rounded up to two figures.
//
// A gate then asks, in every construction resource of the age (knowledge
// aside: no gate asks for knowledge), for GateStoreShare of the store that
// resource needs (GateRuleAmounts), rounded to two figures.
const (
	StorageRuleCopies = 5
	StorageRuleRoom   = 1.25
	StorageRuleMargin = 1.05
	StorageRuleFloor  = 0.05
	GateStoreShare    = 0.25
	// StorageRate is the rate a storage building's price climbs at for each
	// copy owned: Kittens Game's barn.
	StorageRate = 1.75
)

// storageRuleCount is how many copies of d the reference town holds.
func storageRuleCount(d BuildingDef, gate map[string]int) int {
	return max(StorageRuleCopies, gate[d.Key])
}

func maxPrice(d BuildingDef, owned int) float64 {
	f := detmath.Pow(d.CostScale, float64(owned))
	top := 0.0
	for _, v := range d.BaseCost {
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

// StorageRuleSizes works the storage rule out: the size of each age's
// storage building ("all" storage per copy), by building key, and the
// reference store on leaving each age, by age key. baseStore is what a
// resource holds before any building.
func StorageRuleSizes(defs []BuildingDef, ages []AgeDef, baseStore float64) (sizes map[string]float64, store map[string]float64) {
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
		asked := StorageRuleRoom * dearest(i)
		need := asked
		if i+1 < n {
			if nb, has := find(i+1, "storage"); has {
				need = math.Max(need, StorageRuleMargin*maxPrice(nb, 0))
				if h, hasHouse := find(i+1, "housing"); hasHouse {
					need = math.Max(need, StorageRuleMargin*maxPrice(h, 0))
				}
			} else {
				need = math.Max(need, StorageRuleRoom*dearest(i+1))
				for _, d := range byAge[i+1] {
					need = math.Max(need, StorageRuleMargin*maxPrice(d, 0))
				}
			}
		}
		size := math.Max((need-prev)/StorageRuleCopies, StorageRuleFloor*asked/StorageRuleCopies)
		for k := 2; k <= StorageRuleCopies; k++ {
			size = math.Max(size, (StorageRuleMargin*maxPrice(b, k-1)-prev)/float64(k-1))
		}
		size = roundUpSignificant(size, 2)
		sizes[b.Key] = size
		prev += StorageRuleCopies * size
		store[ages[i].Key] = prev
	}
	return sizes, store
}

// GateRuleAmounts works out what the gate out of each age asks to hold, by
// the age the gate leads into. For every construction resource of the age
// left (knowledge aside), the store that resource needs is the dearest price
// the reference town pays in it, plus StorageRuleRoom; the gate asks for
// GateStoreShare of that.
//
// It is not a share of the one store every resource has. That store is sized
// by the age's dearest resource, and a quarter of it in a resource the age
// prices low is worth more than the whole age: 80 price units of uranium to
// leave the Atomic Age, 3,000 of quantum flux to leave the Quantum Age, where
// a wonder costs 40.
func GateRuleAmounts(defs []BuildingDef, ages []AgeDef) map[string]map[string]float64 {
	levels := priceLevels(defs)
	out := map[string]map[string]float64{}
	for i := 0; i+1 < len(ages); i++ {
		gate := ages[i+1].BuildingReqs
		dearest := map[string]float64{}
		for _, d := range defs {
			if d.RequiredAge != ages[i].Key || d.Category == "wonder" || d.Category == "monument" || d.Category == "storage" {
				continue
			}
			f := detmath.Pow(d.CostScale, float64(storageRuleCount(d, gate)-1))
			for r, v := range d.BaseCost {
				dearest[r] = math.Max(dearest[r], math.Max(1, math.Floor(v*f)))
			}
		}
		res := make([]string, 0, len(levels[ages[i].Key]))
		for r := range levels[ages[i].Key] {
			res = append(res, r)
		}
		sort.Strings(res)
		asks := map[string]float64{}
		for _, r := range res {
			if r == "knowledge" || dearest[r] <= 0 {
				continue
			}
			asks[r] = roundSignificant(GateStoreShare*StorageRuleRoom*dearest[r], 2)
		}
		out[ages[i+1].Key] = asks
	}
	return out
}
