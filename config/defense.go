package config

import "sync"

// Army defense: how a garrison softens raids and catastrophes.
//
// The Defense Rating (soldiers × 2 × (1 + military_power), see
// game.MilitaryManager.CalculateDefense) is measured against the THREAT of the
// age being played, and the share of a raid's losses the garrison blunts is
//
//	mitigation = DefenseMitigationCap × defense / (defense + threat)
//
// a smooth, saturating curve: 0 with no soldiers, a quarter of the cap at a
// third of the threat, half the cap when defense equals the threat, and it
// only approaches the cap. The threat doubles every age, the same rate the
// military lineage's soldier caps double (10 × 2^tier) and the rate a
// player's soldier stock was measured to grow, so a garrison has to keep up
// with the age: an Iron Age army that blunts a fifth of an Iron Age raid
// blunts a rounding error of a Quantum Age one.
//
// What mitigation touches (and nothing else):
//   - random events flagged Raid (EventDef.Raid): their steal_resource and
//     worker_loss effects
//   - diplomacy war raids: the resource each raid takes
//   - Endure: the share of buildings destroyed and of stock lost, after the
//     Harbinger's Brace, with the two together capped at EndureReductionCap
//
// A player with no soldiers takes exactly the losses they took before the
// army mattered.

const (
	// DefenseMitigationCap is the most a garrison can ever blunt of one
	// raid's losses, however large it is. The curve approaches it and never
	// reaches it.
	DefenseMitigationCap = 0.45

	// DefenseThreatBase is where the threat curve starts (age order 0), in
	// Defense Rating points: 1.28M in the Iron Age, 2.56M in the
	// Classical, doubling on. The curve is the threat from the Iron Age on;
	// the three ages before it have threats of their own (EarlyThreat*).
	// Measured, not guessed. Every player carries a
	// garrison they never chose: the age gates ask for military buildings (15
	// Hunting Lodges, 15 Military Academies, 15 Bunker Complexes, ...), the
	// lineage carries them into every later age, and they train soldiers
	// unstaffed. The greedy smoke bot, which does nothing else for its army,
	// holds about 60K × 2^order defense, so this base leaves that incidental
	// garrison blunting 8-19% of a raid (0.2-0.8x the threat). A deliberate
	// army is what reaches for the cap: the -army=on bot's four extra
	// buildings per age reach 25-30% from the Industrial Age and 40% by the
	// Atomic. (At 160, the first guess, the incidental garrison sat at 200-700x
	// the threat and every player was pinned at the 45% cap for free.)
	DefenseThreatBase = 160000.0

	// DefenseThreatGrowth multiplies the threat once per age. 2.0 matches the
	// military lineage, whose per-building soldier cap doubles every tier.
	DefenseThreatGrowth = 2.0

	// EndureReductionCap bounds how much Brace and the garrison TOGETHER can
	// soften an Endure, as a share of the unbraced loss: at most 60% fewer
	// buildings destroyed and 60% less stock lost. Brace is applied first,
	// then the garrison blunts part of what is left, and the result is held
	// to this cap. Brace level 2 alone already takes half the building loss,
	// so a maxed Brace leaves the garrison a little more to add, not a free
	// catastrophe.
	EndureReductionCap = 0.60
)

// AgeThreatAt is the raid threat for the age with the given order (0 =
// Primitive Age) on a military yardstick whose threat scale there is scale
// (MilitaryScaleDef.Threat): DefenseThreatBase × DefenseThreatGrowth^order ×
// scale from the Iron Age on, and EarlyThreat × scale before it. The power
// is built by repeated exact multiplication (no math.Pow) so every machine
// gets the same bits. Negative orders count as 0. Pure: an engine reads its
// ruleset's threat (rules.Set.AgeThreat), which calls this with its own
// tree's scale.
func AgeThreatAt(order int, scale float64) float64 {
	if t, early := earlyThreat(order); early {
		return float64(t * scale)
	}
	t := DefenseThreatBase
	for i := 0; i < order; i++ {
		t *= DefenseThreatGrowth
	}
	return float64(t * scale)
}

// The threat before the Iron Age.
//
// The doubling curve was measured from the Iron Age on, where every town
// carries a garrison the age gates ask for and general storage holds
// soldiers by the hundred thousand. Before it nothing asks for a military
// building: a garrison is a War Camp (Stone Age) or a Barracks (Bronze Age)
// someone chose to build and staff, training 0.2 and 0.4 soldiers a tick.
// Against the curve's 320K and 613K such a garrison blunted nothing, so the
// building did nothing for two ages.
//
// The two early threats are sized to that garrison instead: a moderate set
// of the age's military buildings, fully staffed (FlowIncome of soldiers: 1
// a tick in the Stone Age, 3 in the Bronze), trains over the age's pacing
// target a garrison that blunts about a fifth of a raid, and half of that
// garrison about an eighth. A larger garrison blunts more, toward the same
// cap as ever. The Primitive Age has no military building, so no garrison
// ever meets its threat: it is set under the Stone Age's so the threat
// still rises with every age.
const (
	// EarlyThreatPrimitive, EarlyThreatStone and EarlyThreatBronze are the
	// raid threats of the three ages before the Iron Age, in Defense Rating
	// points.
	EarlyThreatPrimitive = 1000.0
	EarlyThreatStone     = 3000.0
	EarlyThreatBronze    = 45000.0
)

// earlyThreat is the threat of an age before the Iron Age (orders 0 to 2),
// and whether order is such an age.
func earlyThreat(order int) (threat float64, early bool) {
	switch {
	case order <= 0:
		return EarlyThreatPrimitive, true
	case order == 1:
		return EarlyThreatStone, true
	case order == 2:
		return EarlyThreatBronze, true
	}
	return 0, false
}

// AgeThreat is AgeThreatAt on the yardstick of the tree this package
// defines (MilitaryScaleAt). For tests and tools, like the other lookups
// that describe this package's own tables.
func AgeThreat(order int) float64 {
	return AgeThreatAt(order, MilitaryScaleAt(order).Threat)
}

// The military yardstick.
//
// Two things are measured against military power: the raid threat (through
// the Defense Rating, soldiers × 2 × (1 + power)) and a mission's difficulty
// (its base less MissionPowerFactor × power). Both were sized while techs
// gave far more of it than they do: +20% in the Bronze Age, +220% by the
// Industrial, +470% in the Atomic and +670% from the Cyberpunk Age on. The
// finished tree gives +15% to +201% over the same ages, in steps of 10 to
// 15 points, so the same garrison stood at a third of the Defense Rating it
// was measured with and a campaign the old tree made a near certainty was
// back to even odds.
//
// MilitaryScale puts both back for the typical player, age by age: one who
// holds the military techs of every age up to their own and the two
// milestones any garrison earns in passing (MilitaryIncidental). For that
// player the garrison blunts the share of a raid it did when the threat was
// measured, and a mission succeeds as often. More military power than that
// still helps, and less still costs, in the same proportion as before.
//
// MilitaryCalibrated is what the techs gave when the threat and the missions
// were measured, by age order: frozen, the record of a tree that no longer
// exists. What they give now is read off the tree, so a change to a military
// tech moves the scale with it and the typical player's odds stay put.
var MilitaryCalibrated = []float64{
	0, 0, // Primitive, Stone
	0.2,      // Bronze: Military Tactics +20%
	0.5,      // Iron: Siege Warfare +30%
	0.9, 0.9, // Classical: Imperial Legions +40%; Medieval
	1.4,           // Renaissance: Gunpowder +50%
	1.7,           // Colonial: Colonialism +30%
	2.2, 2.2, 2.2, // Industrial: Rifling +50%; Victorian; Electric
	4.7, 4.7, // Atomic: Rocketry +100%, Nuclear Deterrence +150%; Modern
	5.7, 5.7, // Information: Cybersecurity +100%; Digital
	6.7, 6.7, 6.7, 6.7, 6.7, 6.7, 6.7, // Cyberpunk: Cybernetics +100%; and on
}

const (
	// MilitaryIncidental is the military power a typical player holds from
	// milestones, from the Iron Age on: First Soldiers (5 soldiers trained,
	// +5%) and War Machine (250 trained, +10%), which the buildings the age
	// gates ask for earn unstaffed. The other military milestones ask for an
	// army someone chose to build.
	MilitaryIncidental = 0.15
	// militaryIncidentalFrom is the age order MilitaryIncidental counts
	// from: the Iron Age, where soldiers begin.
	militaryIncidentalFrom = 3

	// MissionPowerFactor is how much of a mission's difficulty one whole
	// point (+100%) of scaled military power takes off.
	MissionPowerFactor = 0.3
)

// MilitaryScaleDef is one age's military yardstick.
type MilitaryScaleDef struct {
	// Then and Now are the typical player's military power in the age when
	// the threat and the missions were measured and on today's tree: the
	// techs up to the age and MilitaryIncidental.
	Then, Now float64
	// Threat multiplies the age's raid threat: (1 + Now) / (1 + Then), so
	// the typical player's Defense Rating stands where it stood against it.
	Threat float64
	// Mission multiplies military power where a mission's difficulty reads
	// it: Then / Now, so the typical player's power takes off what it took.
	Mission float64
}

// MilitaryScales is the military yardstick of every age in order, read off
// techs: Now adds up their military power age by age. Pure: a ruleset works
// its own out with it (rules.Set.MilitaryScale).
func MilitaryScales(techs []TechDef, order []string) []MilitaryScaleDef {
	pos := AgePositions(order)
	now := make([]float64, len(order))
	for _, t := range techs {
		i, ok := pos[t.Age]
		if !ok {
			continue
		}
		for _, e := range t.Effects {
			if e.Kind == EffectMilitaryPower {
				now[i] += e.Value
			}
		}
	}
	out := make([]MilitaryScaleDef, len(order))
	sum := 0.0
	for i := range order {
		sum += now[i]
		d := MilitaryScaleDef{Then: MilitaryCalibrated[min(i, len(MilitaryCalibrated)-1)], Now: sum, Threat: 1, Mission: 1}
		if i >= militaryIncidentalFrom {
			d.Then += MilitaryIncidental
			d.Now += MilitaryIncidental
		}
		d.Threat = (1 + d.Now) / (1 + d.Then)
		if d.Now > 0 {
			d.Mission = d.Then / d.Now
		}
		out[i] = d
	}
	return out
}

// MilitaryScaleIn is the yardstick of the age with the given order in
// scales (MilitaryScales). An order before the first age reads as the
// first, one past the last as the last; no scales at all read as none.
func MilitaryScaleIn(scales []MilitaryScaleDef, order int) MilitaryScaleDef {
	if len(scales) == 0 {
		return MilitaryScaleDef{Threat: 1, Mission: 1}
	}
	return scales[min(max(order, 0), len(scales)-1)]
}

// militaryScales is MilitaryScales for the tree this package defines, built
// once.
var militaryScales = sync.OnceValue(func() []MilitaryScaleDef {
	return MilitaryScales(Technologies(), AgeOrder())
})

// MilitaryScaleAt is the military yardstick of the age with the given order
// on this package's own tree. For tests and tools; an engine reads its
// ruleset's (rules.Set.MilitaryScale).
func MilitaryScaleAt(order int) MilitaryScaleDef {
	return MilitaryScaleIn(militaryScales(), order)
}

// MissionPower is military power as a mission's difficulty reads it in the
// age with the given order on this package's own tree: power × the age's
// MilitaryScale.Mission. For tests and tools (rules.Set.MissionPower).
func MissionPower(power float64, order int) float64 {
	return float64(power * MilitaryScaleAt(order).Mission)
}

// MissionDifficulty is a mission's chance of failure for a player whose
// MissionPower is missionPower: base less MissionPowerFactor × missionPower,
// and never under MissionDifficultyFloor.
func MissionDifficulty(base, missionPower float64) float64 {
	d := base - float64(missionPower*MissionPowerFactor)
	return max(d, MissionDifficultyFloor)
}

// MissionDifficultyFloor is the least chance of failure any mission keeps,
// whatever the army.
const MissionDifficultyFloor = 0.05

// DefenseMitigation is the share (0..DefenseMitigationCap) of a raid's losses
// a garrison with the given Defense Rating blunts against the given threat:
// DefenseMitigationCap × defense / (defense + threat). 0 when there is no
// defense; a non-positive threat counts as no threat to defend against (0).
func DefenseMitigation(defense, threat float64) float64 {
	if defense <= 0 || threat <= 0 {
		return 0
	}
	m := DefenseMitigationCap * (defense / (defense + threat))
	if m > DefenseMitigationCap {
		m = DefenseMitigationCap
	}
	if m < 0 {
		m = 0
	}
	return m
}
