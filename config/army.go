package config

import "fmt"

// Army experience. Missions that come home give the army experience; the
// rank it reaches raises the odds and hauls of the missions it is sent on
// and the chance of Legendary offers, and from Veteran the board offers the
// next tier of mission an era early. Experience belongs to the run: a new run
// starts the army Green again.

// armyExperience is what a mission of a kind gives the army when it comes
// home a success.
var armyExperience = map[string]float64{
	MissionScout:      1,
	MissionExpedition: 3,
	MissionCampaign:   8,
}

// ArmyExperienceFor is the experience a mission of kind brings: its value on a
// success, minus half of it on a failure. An unknown kind brings none.
func ArmyExperienceFor(kind string, success bool) float64 {
	v := armyExperience[kind]
	if success {
		return v
	}
	return -v / 2
}

// armyRankFloors is the experience each rank starts at: Green, Seasoned,
// Veteran, Elite and Legendary.
var armyRankFloors = [...]int{0, 10, 30, 80, 200}

// ArmyEarlyRank is the rank from which the board's next tier opens one era
// early: Veteran.
const ArmyEarlyRank = 2

// What each rank index gives: points of odds, a share of the haul and a
// share of the Legendary offer chance, each for every rank above Green.
const (
	ArmyOddsPerRank   = .02
	ArmyHaulPerRank   = .05
	ArmyRarePerRank   = .25
	ArmyRankCount     = len(armyRankFloors)
	armyRankKeyPrefix = "army_rank_"
)

// ArmyRankOf is the rank index (0 to 4) experience reaches.
func ArmyRankOf(experience float64) int {
	rank := 0
	for i, floor := range armyRankFloors {
		if experience >= float64(floor) {
			rank = i
		}
	}
	return rank
}

// ArmyRankFloor is the experience rank starts at; for a rank past the top it
// is the top's.
func ArmyRankFloor(rank int) int {
	return armyRankFloors[max(0, min(rank, len(armyRankFloors)-1))]
}

// ArmyRankKey is the lexicon key of a rank's name ("army_rank_2"). The plain
// names (Green, Seasoned, Veteran, Elite, Legendary) are the lexicon's
// fallback and placeholders for the owner's copy.
func ArmyRankKey(rank int) string {
	return fmt.Sprintf("%s%d", armyRankKeyPrefix, max(0, min(rank, len(armyRankFloors)-1)))
}

// ArmyOddsBonus is the odds a rank adds to every mission: two points a rank.
func ArmyOddsBonus(rank int) float64 { return float64(ArmyOddsPerRank * float64(rank)) }

// ArmyHaulFactor is the factor a rank puts on every haul.
func ArmyHaulFactor(rank int) float64 { return 1 + float64(ArmyHaulPerRank*float64(rank)) }

// ArmyRareFactor is the factor a rank puts on the chance of a Legendary
// offer.
func ArmyRareFactor(rank int) float64 { return 1 + float64(ArmyRarePerRank*float64(rank)) }

// MissionTiersOpen is how many tiers of the board are open (1 short, 2 plus
// long, 3 plus grand) in the era of that order for an army of that rank: the
// Stone Era (order 0) offers short missions, the Iron Era long ones too, the
// Steel Era and later grand ones too. From ArmyEarlyRank the era counts as
// the next one.
func MissionTiersOpen(eraOrder, rank int) int {
	if rank >= ArmyEarlyRank {
		eraOrder++
	}
	return max(1, min(len(missionKinds), eraOrder+1))
}
