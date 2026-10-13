package config

import (
	"math"
	"sort"
)

// The knowledge rule: knowledge is not a physical thing, so it does not sit
// in the general store and does not climb like a producer. It has a store of
// its own that only knowledge buildings raise, and the buildings that make it
// climb at their own rate.
//
//  1. A knowledge building (IsKnowledgeBuilding) climbs at KnowledgeRate a
//     copy, not at ProducerRate.
//  2. A gate asks the same spend on a knowledge building as it did when the
//     building climbed at ProducerRate: the count whose total price at
//     KnowledgeRate is nearest the total KnowledgeGateWas copies cost at
//     ProducerRate, a tie going to the lower count (KnowledgeGateCount). The
//     tables are literals; a test holds them to this.
const KnowledgeRate = 1.5

// KnowledgeBase is the knowledge a town holds before it owns any knowledge
// building (the knowledge store's base).
const KnowledgeBase float64 = 240

// IsKnowledgeBuilding reports whether d is a knowledge building: not a
// wonder, and its output is knowledge.
func IsKnowledgeBuilding(d BuildingDef) bool {
	if d.Category == "wonder" {
		return false
	}
	for _, e := range d.Effects {
		if e.Type == "production" && e.Target == "knowledge" && e.Value > 0 {
			return true
		}
	}
	return false
}

// KnowledgeGateWas is the copies of each knowledge building the gate out of
// its age asked for while it climbed at ProducerRate, by building key.
func KnowledgeGateWas() map[string]int {
	return map[string]int{
		"story_circle": 8, "elders_hall": 8, "scriptorium": 8, "agora": 12,
		"library": 15, "monastery_library": 8, "university": 12,
		"natural_philosophy_hall": 8, "research_institute": 8, "academy": 12,
		"physics_laboratory": 15, "research_campus": 18, "think_tank": 20,
		"innovation_hub": 15, "ai_research_lab": 15, "neuro_research_center": 8,
		"theoretical_institute": 8, "deep_space_observatory": 8,
		"xenology_institute": 8, "cosmic_research_station": 8, "reality_academy": 20,
	}
}

// climbTotal is the price of the first n copies of something that costs 1 and
// climbs by rate a copy: 1 + rate + ... + rate^(n-1). Every part of a
// building's price climbs by the same factor, so this is the whole price in
// units of its first copy.
func climbTotal(rate float64, n int) float64 {
	sum, p := 0.0, 1.0
	for i := 0; i < n; i++ {
		sum += p
		p *= rate
	}
	return sum
}

// KnowledgeGateCount is the count at KnowledgeRate whose total price is
// nearest what was copies cost at ProducerRate (a tie goes to the lower).
func KnowledgeGateCount(was int) int {
	want := climbTotal(ProducerRate, was)
	best, bestGap := 1, math.Inf(1)
	for n := 1; n <= 4*was; n++ {
		if gap := math.Abs(climbTotal(KnowledgeRate, n) - want); gap < bestGap {
			best, bestGap = n, gap
		}
	}
	return best
}

// The knowledge store (R3). Knowledge has a store of its own: a base, which a
// town holds before it owns any knowledge building, plus what each copy of a
// knowledge building holds (BuildingDef.KnowledgeStore). Nothing else raises
// it, and the general store's size has nothing to do with it.
//
// What a copy holds is authored (a literal on the building); this is the rule
// the figures are held to. The town an age is priced for owns, of each
// knowledge building, the copies the gate out of its age asks for. That town,
// on top of what the earlier ages' required copies hold, holds the larger of
//
//	(a) the age's dearest tech plus a quarter of it, and
//	(b) KnowledgeNextMargin times the second-cheapest tech of the next age,
//
// so the dearest tech of the age can be saved for in full, and a town let
// through the gate can hold the first techs of the next age. Where an age has
// two knowledge buildings they hold equal amounts. A figure is rounded up to
// two significant figures. (a) and (b) are tech prices, and the tech prices
// are set against what the knowledge buildings make (smoke.KnowledgeLadders),
// so the two are solved together: the prices first, from the stay, then the
// figures from the prices; the figures change nothing the prices read.
//
// The base is the knowledge the first age's cheapest tech costs, rounded up
// to two significant figures: with no knowledge building a town can hold the
// first tech there is, and it is the least that lets research begin.
const (
	// KnowledgeDearestMargin is the share of an age's dearest tech the store
	// holds on top of it.
	KnowledgeDearestMargin = 1.25
	// KnowledgeNextMargin is the multiple of the next age's second-cheapest
	// tech the store holds.
	KnowledgeNextMargin = 1.2
)

// KnowledgeStoreRow is one age of the knowledge store.
type KnowledgeStoreRow struct {
	Age string
	// Need is what the rule asks the store to hold on leaving the age: the
	// larger of (a) and (b), and the two of them.
	Need, DearestPlusQuarter, NextSecond float64
	// Copies is the copies the gate out of the age asks of the age's
	// knowledge buildings, over every one of them.
	Copies int
	// Figure is what each of the age's knowledge buildings holds a copy
	// (0 where the age has none, or the earlier ages already hold the need).
	Figure float64
	// Before is what the earlier ages' required copies and the base hold;
	// Held is Before plus this age's.
	Before, Held float64
}

// techPricesOf returns the prices of an age's techs, cheapest first.
func techPricesOf(techs []TechDef, age string) []float64 {
	var out []float64
	for _, t := range techs {
		if t.Age == age && t.Cost > 0 {
			out = append(out, t.Cost)
		}
	}
	sort.Float64s(out)
	return out
}

// KnowledgeBaseFor is the base the rule gives: the first age's cheapest tech,
// rounded up to two significant figures.
func KnowledgeBaseFor(ages []AgeDef, techs []TechDef) float64 {
	if len(ages) == 0 {
		return 0
	}
	p := techPricesOf(techs, ages[0].Key)
	if len(p) == 0 {
		return 0
	}
	return roundUpSignificant(p[0], 2)
}

// KnowledgeStoreRows works the knowledge store rule out, an age at a time,
// for the given base. A row whose earlier ages already hold the need gets a
// figure of 0: the rule gives the age no figure to author (the tests name
// such an age rather than invent one).
func KnowledgeStoreRows(defs []BuildingDef, ages []AgeDef, techs []TechDef, base float64) []KnowledgeStoreRow {
	rows := make([]KnowledgeStoreRow, len(ages))
	held := base
	for i, a := range ages {
		r := KnowledgeStoreRow{Age: a.Key, Before: held}
		if p := techPricesOf(techs, a.Key); len(p) > 0 {
			r.DearestPlusQuarter = float64(KnowledgeDearestMargin * p[len(p)-1])
		}
		if i+1 < len(ages) {
			if p := techPricesOf(techs, ages[i+1].Key); len(p) > 1 {
				r.NextSecond = float64(KnowledgeNextMargin * p[1])
			}
		}
		r.Need = max(r.DearestPlusQuarter, r.NextSecond)
		for _, d := range defs {
			if d.RequiredAge == a.Key && IsKnowledgeBuilding(d) && i+1 < len(ages) {
				r.Copies += ages[i+1].BuildingReqs[d.Key]
			}
		}
		if r.Copies > 0 && r.Need > held {
			r.Figure = roundUpSignificant((r.Need-held)/float64(r.Copies), 2)
		}
		held += float64(float64(r.Copies) * r.Figure)
		r.Held = held
		rows[i] = r
	}
	return rows
}

// KnowledgeStoreHeld is the knowledge store of the town an age is priced
// for: KnowledgeBase and what the authored figures hold at the gates' counts,
// by age key, on leaving the age.
func KnowledgeStoreHeld(defs []BuildingDef, ages []AgeDef) map[string]float64 {
	out := make(map[string]float64, len(ages))
	held := KnowledgeBase
	for i, a := range ages {
		for _, d := range defs {
			if d.RequiredAge == a.Key && IsKnowledgeBuilding(d) && i+1 < len(ages) {
				held += float64(float64(ages[i+1].BuildingReqs[d.Key]) * d.KnowledgeStore)
			}
		}
		out[a.Key] = held
	}
	return out
}
