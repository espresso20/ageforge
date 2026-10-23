package game

import (
	"maps"

	"github.com/espresso20/ageforge/config"
)

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

	// Words, Age, Grand, Rare and Loot are what Label, Title, Place and Haul
	// were drawn from: the offer's words, age, kind and Legendary mark, and
	// the loot by resource. A save holds them in place of those four texts,
	// so it reads the same in any script (VibeMode); draw writes the texts
	// again. An entry with no Words is written by hand and its texts stay.
	Words OfferWords         `json:"words,omitzero"`
	Age   string             `json:"age,omitempty"`
	Grand string             `json:"grand,omitempty"`
	Rare  bool               `json:"rare,omitempty"`
	Loot  map[string]float64 `json:"loot,omitempty"`
}

// draw writes the entry's Label, Title, Place and Haul from what it was drawn
// from, in the words of the script mm plays by.
func (r *MissionReturn) draw(mm *MilitaryManager) {
	if !r.Words.Made || mm.rules == nil {
		return
	}
	words, ok := mm.rules.MissionWords(r.Age)
	if !ok {
		return
	}
	offer := MissionOffer{Tier: r.Tier, Grand: r.Grand, Age: r.Age, Rare: r.Rare, Words: r.Words}
	offer.draw(words)
	r.Label, r.Title, r.Place, r.Haul = words.Tiers[offer.wordsKind()], offer.Title, offer.Place, amountsText(r.Loot)
}

// forSave is the entry as a save holds it: without the texts it can draw.
func (r MissionReturn) forSave() MissionReturn {
	if r.Words.Made {
		r.Label, r.Title, r.Place, r.Haul = "", "", "", ""
	}
	return r
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
		Words: res.Offer.Words, Age: res.Offer.Age, Grand: res.Offer.Grand, Rare: res.Offer.Rare, Loot: maps.Clone(res.Rewards),
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
