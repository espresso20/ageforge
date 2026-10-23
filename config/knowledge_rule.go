package config

import (
	"math"
)

// The knowledge rule: knowledge is not a thing a town holds. A knowledge
// building comes with its scholars (BuildingDef.Scholars), who work on one
// tech at a time, and the buildings that employ them climb at their own rate.
//
//  1. A knowledge building (IsKnowledgeBuilding) climbs at KnowledgeRate a
//     copy, not at ProducerRate.
//  2. A gate asks the same spend on a knowledge building as it did when the
//     building climbed at ProducerRate: the count whose total price at
//     KnowledgeRate is nearest the total KnowledgeGateWas copies cost at
//     ProducerRate, a tie going to the lower count (KnowledgeGateCount). The
//     tables are literals; a test holds them to this.
//  3. A knowledge building researches at full rate up to the count its gate
//     asks for (KnowledgeCap): an age's research has a soft ceiling, and the
//     tech prices are set against it. Copies past it still build, at the
//     normal climb in price, and each researches at KnowledgePastCapShare
//     of a copy's rate (BuildingDef.ResearchCopies).
const KnowledgeRate = 1.5

// KnowledgePastCapShare is the share of a copy's research that each copy of a
// knowledge building past its cap (BuildingDef.FullRateCopies) makes: half.
const KnowledgePastCapShare = 0.5

// Research simmers. It is an idle game: a player should be able to get every
// tech, but it takes some time. The work an age's techs ask is a share of the
// work the ordinary reference player's scholars do during that age's stay
// (the knowledge ladder, smoke.KnowledgeLadders, and property 10 hold it).
// ResearchBaseShare is that share for the first age. From the Stone Age on the
// share is ResearchBaseShare times ResearchSimmer, so an age's research takes
// one and a half stays and the unfinished part carries into the next age.
// ResearchSimmerExemptAge stays at the base share, so a new player sees a tech
// finish early. The keystone (the tech an age's wonder needs) must still
// finish inside its age, because the wonder needs it to leave: the Research
// Covenant holds that, and it is never relaxed to fit the multiplier.
const (
	ResearchBaseShare       = 0.8
	ResearchSimmer          = 1.5
	ResearchSimmerExemptAge = "primitive_age"
)

// ResearchShare is the share of the ordinary player's stay in age that the
// age's techs total in work: ResearchBaseShare in the exempt age,
// ResearchBaseShare times ResearchSimmer in every other (a constant product,
// so it is exactly 1.2 and carries no run-time rounding).
func ResearchShare(age string) float64 {
	if age == ResearchSimmerExemptAge {
		return ResearchBaseShare
	}
	return ResearchBaseShare * ResearchSimmer
}

// ResearchTarget is the name an effect gives the work scholars do: the
// target of a tech's bonus to research (EffectOutput), of an event's gain or
// rate and of a milestone's reward. It is not a resource: nothing holds it,
// and a gain of it lands as progress on the tech in progress.
const ResearchTarget = "research"

// producedBy is what effect e makes each tick: a "production" effect makes
// its Target, a "research" effect (a knowledge building's or a wonder's)
// makes ResearchTarget.
func producedBy(e Effect) (res string, ok bool) {
	switch e.Type {
	case "production":
		return e.Target, true
	case "research":
		return ResearchTarget, true
	}
	return "", false
}

// IsKnowledgeBuilding reports whether d is a knowledge building: not a
// wonder, and it comes with scholars, whose work it lists as a "research"
// effect.
func IsKnowledgeBuilding(d BuildingDef) bool {
	if d.Category == "wonder" {
		return false
	}
	for _, e := range d.Effects {
		if e.Type == "research" && e.Value > 0 {
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

// KnowledgeCap is the most copies of the knowledge building key that research
// at full rate: the count the gate out of its age asks for
// (KnowledgeGateCount). Each copy past it researches at half rate. 0 for a
// key that is not one of the knowledge buildings.
func KnowledgeCap(key string) int {
	was, ok := KnowledgeGateWas()[key]
	if !ok {
		return 0
	}
	return KnowledgeGateCount(was)
}

// ResearchCopies is the copies of d that count toward its research when a
// town holds count of it: all of them up to d.FullRateCopies, and a share
// (KnowledgePastCapShare) of each copy past it. A building with no such cap
// counts every copy. Under the cap the figure is the count itself, bit for
// bit.
func (d BuildingDef) ResearchCopies(count int) float64 {
	if d.FullRateCopies <= 0 || count <= d.FullRateCopies {
		return float64(count)
	}
	return float64(d.FullRateCopies) + float64(KnowledgePastCapShare*float64(count-d.FullRateCopies))
}

// ModelCopyLimit is the most copies of d the reference models let a town
// hold: its MaxCount, or for a knowledge building the copies that research at
// full rate (the models build to the cap and no further). 0 where there is no
// limit.
func (d BuildingDef) ModelCopyLimit() int {
	if d.MaxCount > 0 {
		return d.MaxCount
	}
	return d.FullRateCopies
}

// The teaching rule. A scholar who is not researching teaches, and teaching
// raises what workers make. With every scholar teaching the bonus is
// TeachingBonus of the town's age, in proportion to the share of the
// scholars that teach. It joins the worker-output bonus pool (gather_rate)
// as its own source.
//
// The sizes are one table, by the age's place in the order of ages: ages 1 to
// 3 give 10 percent, 4 to 6 give 8, 7 to 9 give 6, 10 to 15 give 5, and 16
// and later give 4.
type TeachingStep struct {
	// From is the first age (1-based place) the step holds for.
	From int
	// Bonus is the worker-output bonus with every scholar teaching, a fraction.
	Bonus float64
}

// TeachingTable is the teaching ladder, in order of ages.
func TeachingTable() []TeachingStep {
	return []TeachingStep{
		{From: 1, Bonus: 0.10},
		{From: 4, Bonus: 0.08},
		{From: 7, Bonus: 0.06},
		{From: 10, Bonus: 0.05},
		{From: 16, Bonus: 0.04},
	}
}

// TeachingBonus is the worker-output bonus with every scholar teaching, in
// the age at 0-based position pos in the order of ages.
func TeachingBonus(pos int) float64 {
	place := pos + 1
	bonus := 0.0
	for _, s := range TeachingTable() {
		if place >= s.From {
			bonus = s.Bonus
		}
	}
	return bonus
}
