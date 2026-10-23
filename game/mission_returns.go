package game

import "github.com/espresso20/ageforge/config"

// The returns log: the last missions to come home, kept on the state and
// saved with the game, so the Expedition Report can list them.

// MissionReturn is one mission that came home.
type MissionReturn struct {
	// Label is the mission's tier name in the age it was sent ("Trade Trek");
	// Title and Place are the offer's.
	Label string `json:"label,omitempty"`
	Tier  string `json:"tier"`
	Title string `json:"title"`
	Place string `json:"place,omitempty"`
	// Success says how it ended. Haul is the loot as the log words it
	// ("1.2M wood, 408K stone"); a failure brings back its share, a mission
	// that brought nothing home has none.
	Success    bool   `json:"success"`
	Haul       string `json:"haul,omitempty"`
	Tick       int    `json:"tick"`
	RoadEvents int    `json:"road_events,omitempty"`
	// Party is the soldiers that set out and Fell how many did not come home
	// (on the road and on the way back).
	Party int `json:"party,omitempty"`
	Fell  int `json:"fell,omitempty"`
	// Experience is what the mission gave the army (negative: took from it).
	Experience float64 `json:"experience,omitempty"`
}

// wordsKind is the index of the offer's tier name in the age's words.
func (o MissionOffer) wordsKind() int {
	switch {
	case o.Tier == TierLong:
		return config.WordsLong
	case o.Military():
		return config.WordsGrandMilitary
	case o.Tier == TierGrand:
		return config.WordsGrandCivil
	}
	return config.WordsShort
}

// Label is the offer's tier name as its age words it ("Scout", "Trade
// Trek"); empty for an age with no words.
func (mm *MilitaryManager) Label(o MissionOffer) string {
	if mm.rules == nil {
		return ""
	}
	if words, ok := mm.rules.MissionWords(o.Age); ok {
		return words.Tiers[o.wordsKind()]
	}
	return ""
}

// recordReturn puts a resolved mission at the head of the returns log and
// counts it as new. At most config.MissionReturnsKept are kept.
func (mm *MilitaryManager) recordReturn(res ExpeditionResult, tick int) {
	entry := MissionReturn{
		Label: mm.Label(res.Offer), Tier: res.Offer.Tier, Title: res.Name, Place: res.Offer.Place,
		Success: res.Success, Haul: amountsText(res.Rewards), Tick: tick, RoadEvents: res.RoadEvents,
		Party: res.Party, Fell: res.Party - res.Marching + res.SoldiersFell, Experience: res.Experience,
	}
	mm.returns = append([]MissionReturn{entry}, mm.returns...)
	if len(mm.returns) > config.MissionReturnsKept {
		mm.returns = mm.returns[:config.MissionReturnsKept]
	}
	mm.returnsMade++
}

// MarkReturnsSeen counts every mission that has come home as seen: the
// Expedition Report was viewed.
func (mm *MilitaryManager) MarkReturnsSeen() { mm.returnsSeen = mm.returnsMade }

// MarkReturnsSeen is the Expedition Report being viewed: the missions that
// came home are no longer new.
func (ge *GameEngine) MarkReturnsSeen() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	ge.Military.MarkReturnsSeen()
}
