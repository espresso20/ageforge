package config

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

	// DefenseThreatBase is the raid threat of the Primitive Age (age order 0),
	// in Defense Rating points: 1.28M in the Iron Age, 2.56M in the
	// Classical, doubling on. Measured, not guessed. Every player carries a
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

// AgeThreat is the raid threat for the age with the given order (0 =
// Primitive Age): DefenseThreatBase × DefenseThreatGrowth^order. The power is
// built by repeated exact multiplication (no math.Pow) so every machine gets
// the same bits. Negative orders count as 0.
func AgeThreat(order int) float64 {
	t := DefenseThreatBase
	for i := 0; i < order; i++ {
		t *= DefenseThreatGrowth
	}
	return t
}

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
