package config

import "github.com/espresso20/ageforge/vibe"

// Mission kinds describe the three mission tiers and what each is worth.
// A short mission is MissionScout, a long one MissionExpedition and a grand
// one MissionCampaign; the board (game/mission_board.go) offers all three.
const (
	MissionScout      = "scout"
	MissionExpedition = "expedition"
	MissionCampaign   = "campaign"
)

// MissionTuning is what a kind of mission is: how long it is out, how likely
// it is to come back with a success, how many minutes of the town's income
// its haul is worth for each minute out, and what it costs in minutes of
// income (of soldier income, for a military grand mission).
type MissionTuning struct {
	MinMinutes  float64
	MaxMinutes  float64
	BaseSuccess float64
	Yield       float64
	CostMinutes float64
}

// missionKinds is the tuning of each kind, in the order the board lists
// them. The military grand mission keeps the campaign's numbers.
var missionKinds = map[string]MissionTuning{
	MissionScout:      {5, 30, .85, 2, 5},
	MissionExpedition: {30, 60, .75, 3, 15},
	MissionCampaign:   {60, 120, .65, 4, 10},
}

// The civil grand mission: no soldiers and no army term, paid in minutes of
// income in the age's two most-charged materials.
const (
	CivilGrandSuccess     = .60
	CivilGrandCostMinutes = 45
)

// MissionKinds lists the kinds in board order: short, long, grand.
func MissionKinds() []string {
	return []string{MissionScout, MissionExpedition, MissionCampaign}
}

// MissionKindTuning is the tuning of a kind; false for a kind that does
// not exist.
func MissionKindTuning(kind string) (MissionTuning, bool) {
	t, ok := missionKinds[kind]
	return t, ok
}

// The board's rules (see game/mission_board.go).
const (
	// BoardOffersPerTier is how many offers of each tier the board holds.
	BoardOffersPerTier = 3
	// BoardRefillMinutes is the wait before a taken offer's slot is filled
	// with a new offer of its tier.
	BoardRefillMinutes = 5
	// BoardExpireMinutes is how long an offer stays up before it expires and
	// is replaced.
	BoardExpireMinutes = 60
	// MissionReturnsKept is how many returned missions the returns log
	// keeps (the Expedition Report lists them).
	MissionReturnsKept = 10
	// RareOfferChance is the chance a new offer is rare (Legendary).
	RareOfferChance = .05
	// RareOfferGap is how many ordinary offers must be made after a rare one
	// before another may be rare.
	RareOfferGap = 10
	// RareOfferFloor is how many ordinary offers in a row force the next to
	// be rare.
	RareOfferFloor = 40
	// A rare offer is longer, harder and pays more.
	RareLengthFactor  = 1.5
	RareSuccessCut    = .15
	RareSuccessFloor  = .10
	RareHaulFactor    = 4
	MissionSuccessMin = .10
	MissionSuccessMax = .95
)

// RareTitleWord is the word a rare offer's title begins with.
func RareTitleWord() string { return vibe.Line("config.missions.RareTitleWord.01", "Legendary") }

// MissionBoardLongestMinutes is the longest an ordinary offer can be out: the
// grand tier's top length. A Legendary offer is longer still.
func MissionBoardLongestMinutes() float64 {
	longest := 0.0
	for _, t := range missionKinds {
		longest = max(longest, t.MaxMinutes)
	}
	return longest
}
