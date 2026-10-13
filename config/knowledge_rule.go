package config

import "math"

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
