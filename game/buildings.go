package game

import (
	"fmt"
	"math"

	"github.com/user/ageforge/config"
)

// BuildingManager manages all buildings
type BuildingManager struct {
	counts      map[string]int
	defs        map[string]config.BuildingDef
	unlocked    map[string]bool
	wonderBanks map[string]map[string]float64 // wonderKey -> resource -> banked amount
}

// NewBuildingManager creates a building manager
func NewBuildingManager() *BuildingManager {
	return &BuildingManager{
		counts:      make(map[string]int),
		defs:        config.BuildingByKey(),
		unlocked:    make(map[string]bool),
		wonderBanks: make(map[string]map[string]float64),
	}
}

// UnlockBuilding makes a building available
func (bm *BuildingManager) UnlockBuilding(key string) {
	bm.unlocked[key] = true
}

// IsUnlocked returns whether a building type is available
func (bm *BuildingManager) IsUnlocked(key string) bool {
	return bm.unlocked[key]
}

// SuggestKey returns the closest building key to the input, or "" if none is close
func (bm *BuildingManager) SuggestKey(input string) string {
	best := ""
	bestDist := 3 // max edit distance to suggest
	for key := range bm.defs {
		d := editDistance(input, key)
		if d < bestDist {
			bestDist = d
			best = key
		}
	}
	return best
}

// editDistance computes Levenshtein distance between two strings
func editDistance(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr := make([]int, lb+1)
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			min := ins
			if del < min {
				min = del
			}
			if sub < min {
				min = sub
			}
			curr[j] = min
		}
		prev = curr
	}
	return prev[lb]
}

// GetCount returns how many of a building exist
func (bm *BuildingManager) GetCount(key string) int {
	return bm.counts[key]
}

// GetCost calculates the cost for the next building of this type (with scaling)
func (bm *BuildingManager) GetCost(key string) map[string]float64 {
	def, ok := bm.defs[key]
	if !ok {
		return nil
	}
	count := bm.counts[key]
	cost := make(map[string]float64)
	for res, base := range def.BaseCost {
		cost[res] = math.Floor(base * math.Pow(def.CostScale, float64(count)))
	}
	return cost
}

// Build constructs a building. Returns false if can't afford or not unlocked.
func (bm *BuildingManager) Build(key string, resources *ResourceManager) bool {
	if !bm.unlocked[key] {
		return false
	}
	def, ok := bm.defs[key]
	if !ok {
		return false
	}
	if def.MaxCount > 0 && bm.counts[key] >= def.MaxCount {
		return false
	}
	cost := bm.GetCost(key)
	if !resources.Pay(cost) {
		return false
	}
	bm.counts[key]++
	return true
}

// GetEffects returns the total effects from all built buildings
func (bm *BuildingManager) GetEffects() []config.Effect {
	var effects []config.Effect
	for key, count := range bm.counts {
		if count == 0 {
			continue
		}
		def := bm.defs[key]
		for _, eff := range def.Effects {
			scaled := config.Effect{
				Type:   eff.Type,
				Target: eff.Target,
				Value:  eff.Value * float64(count),
			}
			effects = append(effects, scaled)
		}
	}
	return effects
}

// GetPopCapacity returns total population capacity from housing buildings
func (bm *BuildingManager) GetPopCapacity() int {
	cap := 0
	for key, count := range bm.counts {
		def := bm.defs[key]
		for _, eff := range def.Effects {
			if eff.Type == "capacity" && eff.Target == "population" {
				cap += int(eff.Value) * count
			}
		}
	}
	return cap
}

// GetStorageBonuses returns per-resource storage bonuses from buildings
// "all" key means it applies to every resource
func (bm *BuildingManager) GetStorageBonuses() map[string]float64 {
	bonuses := make(map[string]float64)
	for key, count := range bm.counts {
		def := bm.defs[key]
		for _, eff := range def.Effects {
			if eff.Type == "storage" {
				bonuses[eff.Target] += eff.Value * float64(count)
			}
		}
	}
	return bonuses
}

// GetAll returns building counts (for save)
func (bm *BuildingManager) GetAll() map[string]int {
	out := make(map[string]int)
	for key, count := range bm.counts {
		out[key] = count
	}
	return out
}

// LoadCounts restores building counts from save data
func (bm *BuildingManager) LoadCounts(counts map[string]int) {
	for key, count := range counts {
		bm.counts[key] = count
	}
}

// BankResource deposits amount of resource into a wonder's bank from player storage.
// Returns the actual amount deposited (capped at remaining need), or an error.
func (bm *BuildingManager) BankResource(wonderKey, resource string, amount float64, rm *ResourceManager) (float64, error) {
	def, ok := bm.defs[wonderKey]
	if !ok {
		return 0, fmt.Errorf("unknown building: %s", wonderKey)
	}
	if def.Category != "wonder" {
		return 0, fmt.Errorf("%s is not a wonder", def.Name)
	}
	if !bm.unlocked[wonderKey] {
		return 0, fmt.Errorf("%s is not yet unlocked", def.Name)
	}
	if bm.counts[wonderKey] > 0 {
		return 0, fmt.Errorf("%s is already built", def.Name)
	}
	required, exists := def.BaseCost[resource]
	if !exists {
		return 0, fmt.Errorf("%s does not require %s", def.Name, resource)
	}
	banked := bm.wonderBanks[wonderKey][resource]
	remaining := required - banked
	if remaining <= 0 {
		return 0, fmt.Errorf("%s bank for %s is already full", resource, def.Name)
	}
	if amount > remaining {
		amount = remaining
	}
	if !rm.Pay(map[string]float64{resource: amount}) {
		return 0, fmt.Errorf("not enough %s", resource)
	}
	if bm.wonderBanks[wonderKey] == nil {
		bm.wonderBanks[wonderKey] = make(map[string]float64)
	}
	bm.wonderBanks[wonderKey][resource] += amount
	return amount, nil
}

// GetWonderBank returns a copy of the banked resources for a wonder.
func (bm *BuildingManager) GetWonderBank(wonderKey string) map[string]float64 {
	bank := bm.wonderBanks[wonderKey]
	out := make(map[string]float64, len(bank))
	for k, v := range bank {
		out[k] = v
	}
	return out
}

// IsWonderBankFull returns true when all required resources have been banked.
func (bm *BuildingManager) IsWonderBankFull(wonderKey string) bool {
	def, ok := bm.defs[wonderKey]
	if !ok || def.Category != "wonder" {
		return false
	}
	if len(def.BaseCost) == 0 {
		return false
	}
	bank := bm.wonderBanks[wonderKey]
	for res, need := range def.BaseCost {
		if bank[res] < need {
			return false
		}
	}
	return true
}

// GetWonderBanks returns a deep copy of all wonder banks for saving.
func (bm *BuildingManager) GetWonderBanks() map[string]map[string]float64 {
	out := make(map[string]map[string]float64, len(bm.wonderBanks))
	for k, v := range bm.wonderBanks {
		inner := make(map[string]float64, len(v))
		for rk, rv := range v {
			inner[rk] = rv
		}
		out[k] = inner
	}
	return out
}

// LoadWonderBanks restores wonder bank state from a save.
func (bm *BuildingManager) LoadWonderBanks(banks map[string]map[string]float64) {
	for k, v := range banks {
		bm.wonderBanks[k] = v
	}
}

// Snapshot returns building states for UI
func (bm *BuildingManager) Snapshot(resources *ResourceManager) map[string]BuildingState {
	out := make(map[string]BuildingState)
	for key, def := range bm.defs {
		cost := bm.GetCost(key)
		state := BuildingState{
			Count:       bm.counts[key],
			Name:        def.Name,
			Category:    def.Category,
			Description: def.Description,
			Unlocked:    bm.unlocked[key],
			NextCost:    cost,
		}
		if def.Category == "wonder" {
			state.WonderBank = bm.GetWonderBank(key)
			state.WonderBankFull = bm.IsWonderBankFull(key)
			state.CanBuild = bm.unlocked[key] && bm.counts[key] == 0 && state.WonderBankFull
		} else {
			state.CanBuild = bm.unlocked[key] && resources.CanAfford(cost)
		}
		out[key] = state
	}
	return out
}
