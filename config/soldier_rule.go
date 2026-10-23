package config

import (
	"math"
	"sync"
)

// The soldier rule: a military building comes with its soldiers. Soldiers are
// people, not a stock a building trains: a copy of a military building
// (BuildingDef.Soldiers a copy) brings its soldiers with it the moment it
// stands, they live in the homes and eat like anyone (game/population.go), and
// the army is the soldiers at their posts.
//
// How many. The army an ordinary town holds when it opens the gate out of an
// age is SoldierSharePercent of the homes that town has. For the military building of
// age A:
//
//	posts = ceil(SoldierSharePercent/100 x RefHousing(A) / GateMilitaryCopies(A))
//
// RefHousing(A) is the housing of the reference town at the end of age A, and
// GateMilitaryCopies(A) the copies of military buildings of age A that town
// holds: the gate's count of each, or TownCopies where the gate asks for
// fewer or none (refCopiesOf). A military building the gate does not ask for
// (the Barracks, say) is in the town all the same, so it shares the army.
// Every military building of one age has the same number of posts.
const SoldierSharePercent = 14

// SoldiersKey is the key of the soldiers' lifetime badge ladder
// (produced.soldiers): the soldiers who take their posts, across all runs.
// Soldiers are not a resource; the army is the soldiers at their posts.
const SoldiersKey = "soldiers"

// MilitaryDomain is the badge domain of the soldiers: the staffing ladder
// "military" counts the soldiers at their posts, as the other domains count
// their workers at work. (The military work area itself belongs to the
// Geographic Society, the one building left with military worker slots.)
const MilitaryDomain = "military"

// IsMilitaryBuilding reports whether d is of the military lineage.
func IsMilitaryBuilding(d BuildingDef) bool { return d.LineageKey == "military" }

// refCopiesOf is the copies of d the reference town holds: the same count
// smoke's reference players build (smoke.refTables.refCount). A knowledge
// building is the gate's count alone; storage is the storage rule's copies and
// everything else TownCopies, or the gate's count where the gate asks for
// more. gate is the buildings the gate out of d's age asks for.
func refCopiesOf(d BuildingDef, gate map[string]int) int {
	if IsKnowledgeBuilding(d) {
		return gate[d.Key]
	}
	n := TownCopies
	if d.Category == "storage" {
		n = StorageRuleCopies
	}
	return max(n, gate[d.Key])
}

// gateOutOf is the buildings the gate that leaves ages[i] asks for (nil for
// the last age).
func gateOutOf(ages []AgeDef, i int) map[string]int {
	if i+1 >= len(ages) {
		return nil
	}
	return ages[i+1].BuildingReqs
}

// RefHousing is the housing of the reference town at the end of each age, by
// the age's place in ages: the sum of the population capacity of every
// building the town holds, each at its reference count, over this age and all
// before it (wonders and monuments aside, as in the reference players).
// smoke.refTables.state works out the same sum.
func RefHousing(defs []BuildingDef, ages []AgeDef) []float64 {
	idx := make(map[string]int, len(ages))
	for i, a := range ages {
		idx[a.Key] = i
	}
	perAge := make([]float64, len(ages))
	for _, d := range defs {
		i, ok := idx[d.RequiredAge]
		if !ok || d.Category == "wonder" || d.Category == "monument" {
			continue
		}
		n := float64(refCopiesOf(d, gateOutOf(ages, i)))
		for _, e := range d.Effects {
			if e.Type == "capacity" && e.Target == "population" {
				perAge[i] += float64(n * e.Value)
			}
		}
	}
	out := make([]float64, len(ages))
	sum := 0.0
	for i := range ages {
		sum += perAge[i]
		out[i] = sum
	}
	return out
}

// GateMilitaryCopies is the copies of military buildings of each age the
// reference town holds, by the age's place in ages: the reference count
// (refCopiesOf) of every military building of that age, summed. It is the
// copies the gate out of the age asks for, with TownCopies standing in for a
// building the gate leaves out.
func GateMilitaryCopies(defs []BuildingDef, ages []AgeDef) []int {
	idx := make(map[string]int, len(ages))
	for i, a := range ages {
		idx[a.Key] = i
	}
	out := make([]int, len(ages))
	for _, d := range defs {
		if !IsMilitaryBuilding(d) {
			continue
		}
		if i, ok := idx[d.RequiredAge]; ok {
			out[i] += refCopiesOf(d, gateOutOf(ages, i))
		}
	}
	return out
}

// RefSoldiers is the soldiers the reference town has at its posts at the end
// of each age, by the age's place in ages: every military building of the age
// and before at its reference count, with its soldiers a copy. (smoke's
// reference town posts these as far as its food margin allows.)
func RefSoldiers(defs []BuildingDef, ages []AgeDef) []float64 {
	idx := make(map[string]int, len(ages))
	for i, a := range ages {
		idx[a.Key] = i
	}
	perAge := make([]float64, len(ages))
	for _, d := range defs {
		if !IsMilitaryBuilding(d) {
			continue
		}
		if i, ok := idx[d.RequiredAge]; ok {
			perAge[i] += float64(d.Soldiers * refCopiesOf(d, gateOutOf(ages, i)))
		}
	}
	out := make([]float64, len(ages))
	sum := 0.0
	for i := range ages {
		sum += perAge[i]
		out[i] = sum
	}
	return out
}

// SoldierPostsFor is the posts a copy of a military building comes with, from
// the reference housing of its age and the copies the reference town holds.
func SoldierPostsFor(housing float64, gateCopies int) int {
	// Whole numbers in the numerator and the divisor, so a share that comes out
	// exact is not rounded up by the float form of 0.14.
	return int(math.Ceil(float64(SoldierSharePercent*housing) / float64(100*max(gateCopies, 1))))
}

// applySoldierPosts gives every military building its soldiers a copy, and
// the table nothing else: a military building makes and stores no soldiers.
func applySoldierPosts(defs []BuildingDef) {
	ages := Ages()
	housing := RefHousing(defs, ages)
	gate := GateMilitaryCopies(defs, ages)
	idx := make(map[string]int, len(ages))
	for i, a := range ages {
		idx[a.Key] = i
	}
	for i := range defs {
		if !IsMilitaryBuilding(defs[i]) {
			continue
		}
		if at, ok := idx[defs[i].RequiredAge]; ok {
			defs[i].Soldiers = SoldierPostsFor(housing[at], gate[at])
		}
	}
}

// The party a mission takes. Every expedition marches with soldiers: a share
// of the age's reference army (RefSoldiers), so a party grows with the army
// an ordinary town holds in that age. PartyShares is the one table of those
// shares, in percent of the army, by kind of mission.
const (
	PartyShort    = "short"
	PartyLong     = "long"
	PartyMilitary = "military"
	PartyCivil    = "civil"
)

// PartyShares is the percent of the age's reference army each kind of mission
// takes: short 4, long 10, grand military 30, grand civil 12.
var PartyShares = map[string]int{
	PartyShort:    4,
	PartyLong:     10,
	PartyMilitary: 30,
	PartyCivil:    12,
}

var (
	refArmyOnce sync.Once
	refArmy     []float64
)

// RefArmy is the reference town's army at the end of the age with the given
// order (RefSoldiers), worked out once.
func RefArmy(order int) float64 {
	refArmyOnce.Do(func() { refArmy = RefSoldiers(BaseBuildings(), Ages()) })
	if order < 0 || len(refArmy) == 0 {
		return 0
	}
	return refArmy[min(order, len(refArmy)-1)]
}

// PartySizeOf is the party of a mission of the given kind (PartyShort, ...)
// out of an army: its share of it, rounded up, at least 1. An age with no army
// (the Primitive Age, before any military building) sends no soldiers: 0.
func PartySizeOf(kind string, army float64) int {
	n := int(math.Round(army))
	if n <= 0 {
		return 0
	}
	return max(1, (PartyShares[kind]*n+99)/100)
}

// PartySize is the party of a mission of the given kind in the age with the
// given order, from the reference army.
func PartySize(kind string, order int) int { return PartySizeOf(kind, RefArmy(order)) }
