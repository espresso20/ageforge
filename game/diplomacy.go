package game

import (
	"fmt"

	"github.com/user/ageforge/config"
)

// DiplomacyManager handles NPC factions and diplomatic relations
type DiplomacyManager struct {
	factions map[string]*FactionState
}

// FactionState tracks the relationship with an NPC faction
type FactionState struct {
	Discovered bool
	Opinion    int    // -100 to 100
	Status     string // "neutral", "friendly", "allied", "rival", "embargo"
	TradeCount int
}

// NewDiplomacyManager creates a new diplomacy manager
func NewDiplomacyManager() *DiplomacyManager {
	return &DiplomacyManager{
		factions: make(map[string]*FactionState),
	}
}

// DiscoverFactions auto-discovers factions when reaching their MinAge
func (dm *DiplomacyManager) DiscoverFactions(age string, ageOrder map[string]int) []string {
	var discovered []string
	for _, def := range config.BaseFactions() {
		if _, exists := dm.factions[def.Key]; exists {
			continue
		}
		if ageOrder[age] >= ageOrder[def.MinAge] {
			dm.factions[def.Key] = &FactionState{
				Discovered: true,
				Opinion:    0,
				Status:     "neutral",
			}
			discovered = append(discovered, def.Key)
		}
	}
	return discovered
}

// SetStatus changes diplomatic status with a faction
func (dm *DiplomacyManager) SetStatus(factionKey, status string, gold float64) (float64, error) {
	defs := config.FactionByKey()
	def, ok := defs[factionKey]
	if !ok {
		return 0, fmt.Errorf("unknown faction: %s", factionKey)
	}

	fs, ok := dm.factions[factionKey]
	if !ok || !fs.Discovered {
		return 0, fmt.Errorf("%s has not been discovered yet", def.Name)
	}

	var cost float64
	switch status {
	case "allied":
		if fs.Opinion < 50 {
			return 0, fmt.Errorf("need opinion >= 50 to ally with %s (current: %d)", def.Name, fs.Opinion)
		}
		cost = 500
	case "rival":
		cost = 0
	case "embargo":
		cost = 0
	case "neutral":
		cost = 0
	default:
		return 0, fmt.Errorf("invalid diplomatic status: %s (valid: allied, rival, embargo, neutral)", status)
	}

	if gold < cost {
		return 0, fmt.Errorf("not enough gold (have: %.0f, need: %.0f)", gold, cost)
	}

	fs.Status = status
	return cost, nil
}

// SendGift sends a gift to a faction, increasing opinion
func (dm *DiplomacyManager) SendGift(factionKey string, gold float64) (float64, error) {
	defs := config.FactionByKey()
	def, ok := defs[factionKey]
	if !ok {
		return 0, fmt.Errorf("unknown faction: %s", factionKey)
	}

	fs, ok := dm.factions[factionKey]
	if !ok || !fs.Discovered {
		return 0, fmt.Errorf("%s has not been discovered yet", def.Name)
	}

	cost := 200.0
	if gold < cost {
		return 0, fmt.Errorf("not enough gold to send gift (have: %.0f, need: %.0f)", gold, cost)
	}

	fs.Opinion += 15
	if fs.Opinion > 100 {
		fs.Opinion = 100
	}

	// Auto-upgrade to friendly if opinion hits 25+
	if fs.Status == "neutral" && fs.Opinion >= 25 {
		fs.Status = "friendly"
	}

	return cost, nil
}

// GetTradeBonus returns the sum of bonuses from allied factions for a resource
func (dm *DiplomacyManager) GetTradeBonus(resourceKey string) float64 {
	defs := config.FactionByKey()
	bonus := 0.0
	for key, fs := range dm.factions {
		if fs.Status != "allied" {
			continue
		}
		def, ok := defs[key]
		if !ok {
			continue
		}
		if def.Specialty == resourceKey {
			bonus += def.TradeBonus
		}
	}
	return bonus
}

// Tick processes diplomacy each game tick
func (dm *DiplomacyManager) Tick(age string, ageOrder map[string]int, tick int) []string {
	var messages []string

	// Discover new factions
	discovered := dm.DiscoverFactions(age, ageOrder)
	defs := config.FactionByKey()
	for _, key := range discovered {
		def := defs[key]
		messages = append(messages, fmt.Sprintf("Discovered faction: %s — %s", def.Name, def.Description))
	}

	// Opinion drift
	for _, fs := range dm.factions {
		if !fs.Discovered {
			continue
		}

		// Rival/embargo: -5 per 50 ticks
		if tick%50 == 0 {
			if fs.Status == "rival" || fs.Status == "embargo" {
				fs.Opinion -= 5
				if fs.Opinion < -100 {
					fs.Opinion = -100
				}
			}
		}

		// Natural drift toward 0 every 100 ticks
		if tick%100 == 0 {
			if fs.Opinion > 0 {
				fs.Opinion--
			} else if fs.Opinion < 0 {
				fs.Opinion++
			}
		}
	}

	return messages
}

// RecordTrade records a trade cycle completion for faction opinion
func (dm *DiplomacyManager) RecordTrade() {
	// Each active trade gives +1 opinion to all discovered factions
	for _, fs := range dm.factions {
		if !fs.Discovered {
			continue
		}
		fs.TradeCount++
		fs.Opinion++
		if fs.Opinion > 100 {
			fs.Opinion = 100
		}
	}
}

// Snapshot returns diplomacy state for UI
func (dm *DiplomacyManager) Snapshot(age string, ageOrder map[string]int) DiplomacyState {
	defs := config.FactionByKey()
	factions := make(map[string]FactionInfo)

	for _, def := range config.BaseFactions() {
		fs, exists := dm.factions[def.Key]
		info := FactionInfo{
			Name:       def.Name,
			Specialty:  def.Specialty,
			TradeBonus: def.TradeBonus,
		}
		if exists && fs.Discovered {
			info.Discovered = true
			info.Opinion = fs.Opinion
			info.Status = fs.Status
			info.TradeCount = fs.TradeCount
		} else if ageOrder[age] >= ageOrder[def.MinAge] {
			// Should be discovered but isn't yet (will be next tick)
			info.Discovered = false
		}
		_ = defs // used above via BaseFactions
		factions[def.Key] = info
	}

	return DiplomacyState{
		Factions: factions,
	}
}

// LoadState restores diplomacy state from save
func (dm *DiplomacyManager) LoadState(factions map[string]FactionStateSave) {
	if factions == nil {
		return
	}
	for k, v := range factions {
		dm.factions[k] = &FactionState{
			Discovered: v.Discovered,
			Opinion:    v.Opinion,
			Status:     v.Status,
			TradeCount: v.TradeCount,
		}
	}
}

// FactionStateSave is the serializable form of FactionState
type FactionStateSave struct {
	Discovered bool   `json:"discovered"`
	Opinion    int    `json:"opinion"`
	Status     string `json:"status"`
	TradeCount int    `json:"trade_count"`
}

// GetFactionsForSave returns faction states for serialization
func (dm *DiplomacyManager) GetFactionsForSave() map[string]FactionStateSave {
	out := make(map[string]FactionStateSave, len(dm.factions))
	for k, fs := range dm.factions {
		out[k] = FactionStateSave{
			Discovered: fs.Discovered,
			Opinion:    fs.Opinion,
			Status:     fs.Status,
			TradeCount: fs.TradeCount,
		}
	}
	return out
}
