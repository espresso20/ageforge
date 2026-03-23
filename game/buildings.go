package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/espresso20/ageforge/config"
)

// BuildingManager manages all buildings, including production buildings, wonders,
// storage structures, wonder construction banks, legacy flags, and ruins.
//
// Key invariants:
//   - counts[key] is the number of fully-constructed instances of building key.
//   - ruins[key] are separate from counts; they produce at 50% base rate with no
//     worker scaling and cannot be rebuilt or demolished.
//   - legacyBuildings marks old-tier buildings that are still functional but
//     must not be queued for construction (CanBuild = false in snapshot).
//   - wonderBanks hold incremental resource deposits; a wonder is only completed
//     when the bank meets the full BaseCost (via BankResource + IsWonderBankFull).
type BuildingManager struct {
	counts          map[string]int
	defs            map[string]config.BuildingDef
	unlocked        map[string]bool
	wonderBanks     map[string]map[string]float64 // wonderKey -> resource -> banked amount
	legacyBuildings map[string]bool               // buildings superseded by lineage progression; functional but unbuildable
	ruins           map[string]int                // ruins from Succumb — produce at 50% base rate, no worker scaling
	pendingUpgrades map[string]string             // oldKey -> newKey: player-driven upgrade awaiting payment
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
		pendingUpgrades: make(map[string]string),
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

// GetQueueCount returns how many instances of a building key are currently in
// the build queue. This is used by cost helpers so that queued-but-not-yet-built
// instances are factored into the cost curve.
// NOTE: The queue lives on GameEngine, so the caller must pass it in.
func (bm *BuildingManager) GetQueueCount(key string, queue []BuildQueueItem) int {
	n := 0
	for _, item := range queue {
		if item.BuildingKey == key {
			n++
		}
	}
	return n
}

// BuildBatchCost returns the total cost to build n more of a building,
// accounting for already-built and already-queued instances in the cost curve.
// cost_i = floor(baseCost × scale^(built+queued+i))  for i in [0, n).
// Returns (nil, false) if the key is unknown.
func (bm *BuildingManager) BuildBatchCost(key string, n int, queue []BuildQueueItem) (map[string]float64, bool) {
	def, ok := bm.defs[key]
	if !ok {
		return nil, false
	}
	built := bm.counts[key]
	queued := bm.GetQueueCount(key, queue)

	total := make(map[string]float64)
	for i := 0; i < n; i++ {
		exp := float64(built + queued + i)
		for resource, base := range def.BaseCost {
			total[resource] += math.Floor(base * math.Pow(def.CostScale, exp))
		}
	}
	return total, true
}

// SellCost returns the 50% refund for selling n copies of a building,
// assuming current copies are currently built.
// Sells from the top (most expensive first): copy current, current-1, … current-n+1.
// Returns (refund map, true) or (nil, false) if key unknown or n <= 0.
func (bm *BuildingManager) SellCost(key string, n int) (map[string]float64, bool) {
	def, ok := bm.defs[key]
	if !ok || n <= 0 {
		return nil, false
	}
	current := bm.counts[key]
	if n > current {
		n = current
	}
	refund := make(map[string]float64)
	for i := 0; i < n; i++ {
		// copy number being removed: current-i (1-indexed), so exponent = current-1-i
		exp := float64(current - 1 - i)
		for res, base := range def.BaseCost {
			refund[res] += math.Floor(base*math.Pow(def.CostScale, exp)) * 0.5
		}
	}
	return refund, true
}

// RemoveBuilding decrements the count of a built building by n (min 0).
// Returns the actual number removed.
func (bm *BuildingManager) RemoveBuilding(key string, n int) int {
	if n <= 0 {
		return 0
	}
	current := bm.counts[key]
	if n > current {
		n = current
	}
	bm.counts[key] -= n
	return n
}

// GetCost calculates the cost for the next instance of a building type.
// Cost scales exponentially: base × CostScale^count, floored to the nearest
// integer. This makes later copies progressively more expensive.
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

// GetEffects returns the aggregate non-production effects from all built buildings
// (e.g. capacity, storage). Production effects are intentionally excluded here
// because they depend on the worker fill ratio; use WorkerScaledProduction for those.
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
			if def.WorkerCapacity > 0 && getAssigned != nil {
				assigned := getAssigned("worker", key)
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
	if remaining <= 0.001 {
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

// === Player-driven upgrade system ===

// SetPendingUpgrade marks oldKey as having a player-driven upgrade available to newKey.
// The buildings are not transformed immediately; the player must issue `upgrade <oldKey>`.
func (bm *BuildingManager) SetPendingUpgrade(oldKey, newKey string) {
	bm.pendingUpgrades[oldKey] = newKey
}

// GetPendingUpgrade returns the upgrade target for oldKey, or ("", false) if none.
func (bm *BuildingManager) GetPendingUpgrade(oldKey string) (string, bool) {
	v, ok := bm.pendingUpgrades[oldKey]
	return v, ok
}

// ClearPendingUpgrade removes the pending upgrade marker for oldKey.
func (bm *BuildingManager) ClearPendingUpgrade(oldKey string) {
	delete(bm.pendingUpgrades, oldKey)
}

// GetAllPendingUpgrades returns a copy of the pending upgrades map (oldKey -> newKey).
func (bm *BuildingManager) GetAllPendingUpgrades() map[string]string {
	out := make(map[string]string, len(bm.pendingUpgrades))
	for k, v := range bm.pendingUpgrades {
		out[k] = v
	}
	return out
}

// LoadPendingUpgrades restores pending upgrade markers from a save.
func (bm *BuildingManager) LoadPendingUpgrades(upgrades map[string]string) {
	for k, v := range upgrades {
		bm.pendingUpgrades[k] = v
	}
}

// UpgradeCost computes the total cost delta to upgrade upgradeCount copies of oldKey to newKey.
// Cost per copy = max(0, new_copy_cost[res] - old_copy_sell_value[res]) per resource.
// Old sell value = floor(old_copy_cost * 0.5). New copy cost is at the current new count + i.
func (bm *BuildingManager) UpgradeCost(oldKey, newKey string, upgradeCount int) (map[string]float64, bool) {
	oldDef, ok1 := bm.defs[oldKey]
	newDef, ok2 := bm.defs[newKey]
	if !ok1 || !ok2 || upgradeCount <= 0 {
		return nil, false
	}
	oldCount := bm.counts[oldKey]
	newCount := bm.counts[newKey]
	total := make(map[string]float64)
	for i := 0; i < upgradeCount; i++ {
		// Old copy being traded in: remove most expensive first (oldCount-1-i)
		oldExp := float64(oldCount - 1 - i)
		if oldExp < 0 {
			oldExp = 0
		}
		// New copy being created: the (newCount+i)th copy
		newExp := float64(newCount + i)
		for res, base := range newDef.BaseCost {
			newCopyCost := math.Floor(base * math.Pow(newDef.CostScale, newExp))
			// Old sell value for this resource (0 if old building doesn't cost this resource)
			oldBase := oldDef.BaseCost[res]
			oldCopyCost := math.Floor(oldBase * math.Pow(oldDef.CostScale, oldExp))
			oldSellValue := math.Floor(oldCopyCost * 0.5)
			delta := newCopyCost - oldSellValue
			if delta < 0 {
				delta = 0
			}
			total[res] += delta
		}
	}
	return total, true
}

// PartialTransform moves count copies of oldKey to newKey, returning actual moved.
// If oldKey count reaches 0, the pending upgrade is cleared automatically.
// If all copies are moved, worker assignments are transferred via renameWorker callback;
// partial upgrades leave remaining workers on the old key.
func (bm *BuildingManager) PartialTransform(oldKey, newKey string, count int, renameWorker func(domain, oldKey, newKey string)) int {
	have := bm.counts[oldKey]
	if count > have {
		count = have
	}
	if count <= 0 {
		return 0
	}
	bm.counts[oldKey] -= count
	bm.counts[newKey] += count
	bm.unlocked[newKey] = true
	if renameWorker != nil {
		newDef := bm.defs[newKey]
		if newDef.WorkerDomain != "" && bm.counts[oldKey] == 0 {
			// All copies moved — transfer worker assignments
			renameWorker(newDef.WorkerDomain, oldKey, newKey)
		}
		// Partial upgrade: workers stay on old key; player reassigns manually
	}
	if bm.counts[oldKey] == 0 {
		bm.ClearPendingUpgrade(oldKey)
	}
	return count
}

// === Phase 9: Ruins ===

// GenerateRuins converts up to n randomly selected non-wonder building instances
// into ruins on a Succumb catastrophe. Ruins are removed from counts (so they
// no longer receive worker assignments) but continue to produce at 50% base
// rate via the ruins map in WorkerScaledProduction.
// Returns a map of building key → number of ruins created (nil if none).
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
		bm.counts[key] -= count
		if bm.counts[key] < 0 {
			bm.counts[key] = 0
		}
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

// DestroyRandom destroys up to count individual building instances chosen
// uniformly at random. Wonders are excluded because they represent one-off
// civilisation milestones. Unlike GenerateRuins, destroyed buildings are
// removed entirely (they do not become ruins).
// Returns human-readable descriptions for the log (e.g. "3 Lumber Mill").
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

// TransformBuilding is called during age transition to upgrade all instances of
// a lower-tier building into the higher-tier lineage equivalent. It moves the
// count atomically, unlocks the new key, and delegates worker assignment
// renaming to the WorkerManager so that no assigned workers become stranded.
// NOTE: oldKey count is zeroed — do not hold a reference to it afterward.
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
// queue is the engine's current build queue, used to compute queue-aware next costs.
// getWorkerCount(workerDomain, buildingKey) returns the number of workers assigned to that building;
// pass nil to skip worker field population.
func (bm *BuildingManager) Snapshot(resources *ResourceManager, queue []BuildQueueItem, getWorkerCount func(domain, key string) int) map[string]BuildingState {
	out := make(map[string]BuildingState)
	for key, def := range bm.defs {
		// NextCost uses BuildBatchCost so it reflects queue depth — the displayed
		// price is the actual cost the player will pay for the next build.
		cost, _ := bm.BuildBatchCost(key, 1, queue)
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
		// Player-driven upgrade: populate pending upgrade target if set and building has copies
		if upgradeTarget, hasPending := bm.pendingUpgrades[key]; hasPending && count > 0 {
			state.PendingUpgrade = upgradeTarget
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
