package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/espresso20/ageforge/config"
)

// BuildingManager manages all buildings
type BuildingManager struct {
	counts          map[string]int
	defs            map[string]config.BuildingDef
	unlocked        map[string]bool
	wonderBanks     map[string]map[string]float64 // wonderKey -> resource -> banked amount
	legacyBuildings map[string]bool               // Phase 7: buildings superseded by lineage progression
	ruins           map[string]int                // Phase 9: ruins from Succumb — produce at 50%
}

// NewBuildingManager creates a building manager
func NewBuildingManager() *BuildingManager {
	return &BuildingManager{
		counts:          make(map[string]int),
		defs:            config.BuildingByKey(),
		unlocked:        make(map[string]bool),
		wonderBanks:     make(map[string]map[string]float64),
		legacyBuildings: make(map[string]bool),
		ruins:           make(map[string]int),
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

// GetEffects returns the total effects from all built buildings (non-production effects only).
// For production rates, use WorkerScaledProduction instead.
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

// WorkerScaledProduction computes production rates per resource, applying worker fill ratios.
// getAssigned(workerDomain, buildingKey) returns the number of assigned workers for that building type.
// Buildings with WorkerDomain set use: rate = base × count × (0.20 + 0.80 × assigned/totalCap).
// Buildings without workers use: rate = base × count (unchanged behaviour).
func (bm *BuildingManager) WorkerScaledProduction(getAssigned func(domain, key string) int) map[string]float64 {
	rates := make(map[string]float64)
	for key, count := range bm.counts {
		if count == 0 {
			continue
		}
		def := bm.defs[key]
		for _, eff := range def.Effects {
			if eff.Type != "production" {
				continue
			}
			var rate float64
			if def.WorkerDomain != "" && def.WorkerCapacity > 0 && getAssigned != nil {
				assigned := getAssigned(def.WorkerDomain, key)
				totalCap := float64(count * def.WorkerCapacity)
				fillRatio := float64(assigned) / totalCap
				if fillRatio > 1.0 {
					fillRatio = 1.0
				}
				rate = eff.Value * float64(count) * (0.20 + 0.80*fillRatio)
			} else {
				rate = eff.Value * float64(count)
			}
			rates[eff.Target] += rate
		}
	}
	// Ruins produce at 50% base rate; no worker scaling
	for key, count := range bm.ruins {
		if count == 0 {
			continue
		}
		def := bm.defs[key]
		for _, eff := range def.Effects {
			if eff.Type == "production" {
				rates[eff.Target] += eff.Value * float64(count) * 0.50
			}
		}
	}
	return rates
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

// MarkLegacy flags a building type as legacy — functional but superseded, unbuildable.
func (bm *BuildingManager) MarkLegacy(key string) {
	bm.legacyBuildings[key] = true
}

// IsLegacy returns whether a building is flagged as legacy.
func (bm *BuildingManager) IsLegacy(key string) bool {
	return bm.legacyBuildings[key]
}

// GetLegacyBuildings returns all legacy building keys for save serialization.
func (bm *BuildingManager) GetLegacyBuildings() []string {
	var keys []string
	for key := range bm.legacyBuildings {
		keys = append(keys, key)
	}
	return keys
}

// LoadLegacyBuildings restores legacy building flags from a save.
func (bm *BuildingManager) LoadLegacyBuildings(keys []string) {
	for _, key := range keys {
		bm.legacyBuildings[key] = true
	}
}

// === Phase 9: Ruins ===

// GenerateRuins picks up to n random non-wonder building instances from current counts
// and marks them as ruins. Returns a map of key → ruin count created.
// The buildings are removed from counts and added to ruins.
func (bm *BuildingManager) GenerateRuins(n int) map[string]int {
	type inst struct{ key string }
	var pool []inst
	for key, c := range bm.counts {
		if c == 0 {
			continue
		}
		if def, ok := bm.defs[key]; !ok || def.Category == "wonder" {
			continue
		}
		for i := 0; i < c; i++ {
			pool = append(pool, inst{key})
		}
	}
	if len(pool) == 0 {
		return nil
	}
	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	if n > len(pool) {
		n = len(pool)
	}
	newRuins := make(map[string]int)
	for _, inst := range pool[:n] {
		newRuins[inst.key]++
	}
	for key, count := range newRuins {
		bm.ruins[key] += count
	}
	return newRuins
}

// GetAllRuins returns a copy of the ruins map.
func (bm *BuildingManager) GetAllRuins() map[string]int {
	out := make(map[string]int, len(bm.ruins))
	for k, v := range bm.ruins {
		if v > 0 {
			out[k] = v
		}
	}
	return out
}

// LoadRuins restores ruins from a saved state.
func (bm *BuildingManager) LoadRuins(ruins map[string]int) {
	for k, v := range ruins {
		if v > 0 {
			bm.ruins[k] += v
		}
	}
}

// DestroyRandom destroys up to count individual building instances chosen at random.
// Wonders are excluded. Returns a list of human-readable descriptions of what was destroyed.
func (bm *BuildingManager) DestroyRandom(count int) []string {
	// Build a pool of destroyable instances (key repeated by count)
	type inst struct{ key string }
	var pool []inst
	for key, c := range bm.counts {
		if c == 0 {
			continue
		}
		if def, ok := bm.defs[key]; !ok || def.Category == "wonder" {
			continue
		}
		for i := 0; i < c; i++ {
			pool = append(pool, inst{key})
		}
	}
	if len(pool) == 0 {
		return nil
	}
	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	if count > len(pool) {
		count = len(pool)
	}
	destroyed := make(map[string]int)
	for _, inst := range pool[:count] {
		destroyed[inst.key]++
	}
	var names []string
	for key, n := range destroyed {
		bm.counts[key] -= n
		if bm.counts[key] < 0 {
			bm.counts[key] = 0
		}
		if def, ok := bm.defs[key]; ok {
			names = append(names, fmt.Sprintf("%d %s", n, def.Name))
		}
	}
	return names
}

// TransformBuilding transfers count from oldKey to newKey (age-transition lineage upgrade).
// Unlocks the new building. Calls renameWorker to transfer worker assignments if provided.
func (bm *BuildingManager) TransformBuilding(oldKey, newKey string, renameWorker func(domain, oldKey, newKey string)) {
	count := bm.counts[oldKey]
	if count == 0 {
		return
	}
	bm.counts[newKey] += count
	bm.counts[oldKey] = 0
	bm.unlocked[newKey] = true
	if renameWorker != nil {
		newDef := bm.defs[newKey]
		if newDef.WorkerDomain != "" {
			renameWorker(newDef.WorkerDomain, oldKey, newKey)
		}
	}
}

// Snapshot returns building states for UI.
// getWorkerCount(workerDomain, buildingKey) returns the number of workers assigned to that building;
// pass nil to skip worker field population.
func (bm *BuildingManager) Snapshot(resources *ResourceManager, getWorkerCount func(domain, key string) int) map[string]BuildingState {
	out := make(map[string]BuildingState)
	for key, def := range bm.defs {
		cost := bm.GetCost(key)
		count := bm.counts[key]
		state := BuildingState{
			Count:       count,
			Name:        def.Name,
			Category:    def.Category,
			Description: def.Description,
			Unlocked:    bm.unlocked[key],
			AgeKey:      def.RequiredAge,
			NextCost:    cost,
		}
		if def.MaxCount > 0 && count >= def.MaxCount {
			state.AtMaxCount = true
		}
		if def.Category == "wonder" {
			state.WonderBank = bm.GetWonderBank(key)
			state.WonderBankFull = bm.IsWonderBankFull(key)
			state.CanBuild = bm.unlocked[key] && count == 0 && state.WonderBankFull
		} else {
			state.CanBuild = bm.unlocked[key] && !state.AtMaxCount && resources.CanAfford(cost)
		}
		// Phase 6: worker fields
		if def.WorkerDomain != "" {
			state.WorkerDomain = def.WorkerDomain
			state.WorkerCapacity = def.WorkerCapacity
			if getWorkerCount != nil {
				state.WorkersAssigned = getWorkerCount(def.WorkerDomain, key)
			}
		}
		// Phase 7: legacy flag — functional but unbuildable
		if bm.legacyBuildings[key] {
			state.IsLegacy = true
			state.CanBuild = false
		}
		out[key] = state
	}
	// Phase 9: annotate ruins into existing BuildingState entries (or create new entries)
	for key, count := range bm.ruins {
		if count == 0 {
			continue
		}
		bs, exists := out[key]
		if !exists {
			// Building not otherwise visible — create minimal entry for display
			def, ok := bm.defs[key]
			if !ok {
				continue
			}
			bs = BuildingState{
				Name:        def.Name,
				Category:    def.Category,
				Description: def.Description,
				Unlocked:    true,
				AgeKey:      def.RequiredAge,
			}
		}
		bs.RuinCount = count
		bs.CanBuild = false
		out[key] = bs
	}
	return out
}
