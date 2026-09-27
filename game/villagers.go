package game

import (
	"github.com/espresso20/ageforge/config"
)

// domainRuntime holds runtime state for one worker domain
type domainRuntime struct {
	count       int
	assignments map[string]int // buildingKey → assigned worker count
}

// WorkerManager manages population and worker assignments.
//
// Architecture note: the game previously modelled separate domain pools (food,
// military, knowledge, etc.) but was simplified to a single "worker" pool so
// any worker can be assigned to any building. The domain parameter accepted by
// many methods is intentionally ignored to maintain API compatibility with
// callers that still pass domain strings (e.g. "military", "food").
//
// The domains map always contains exactly one entry keyed "worker". Legacy save
// files may contain multiple domain entries; LoadWorkers sums them all into the
// single pool so old saves still load correctly.
type WorkerManager struct {
	domains  map[string]*domainRuntime
	unlocked map[string]bool
	// ageKey is used for WorkerClassDef lookups (food cost per tick, displayed class name).
	// Update via SetAge on every age advance.
	ageKey string
	// buildingDefs is the building table, built once at construction so
	// GetDomainCount (hit every tick by checkMilestones and every UI refresh by
	// GetState) doesn't rebuild ~300 normalized defs per call. Read-only.
	buildingDefs map[string]config.BuildingDef

	// foodClass caches WorkerClassByDomainAndAge("food", ageKey), which rebuilds
	// and scans the full worker-class table. FoodDrain runs several times per
	// tick, so the lookup is resolved eagerly whenever ageKey changes (the
	// constructor and SetAge, both under the engine write lock) and only read
	// elsewhere — Snapshot runs under GetState's read lock, so no lazy writes.
	foodClass   config.WorkerClassDef
	foodClassOK bool
}

// foodClassFor returns the cached food-domain worker class for the current age.
func (vm *WorkerManager) foodClassFor() (config.WorkerClassDef, bool) {
	return vm.foodClass, vm.foodClassOK
}

// refreshFoodClass re-resolves the cached food worker class for vm.ageKey.
func (vm *WorkerManager) refreshFoodClass() {
	vm.foodClass, vm.foodClassOK = config.WorkerClassByDomainAndAge("food", vm.ageKey)
}

// NewWorkerManager creates a new single-pool worker manager.
// All workers belong to a single "worker" pool — domain restrictions are removed.
func NewWorkerManager() *WorkerManager {
	vm := &WorkerManager{
		domains:      make(map[string]*domainRuntime),
		unlocked:     make(map[string]bool),
		ageKey:       "primitive_age",
		buildingDefs: config.BuildingByKey(),
	}
	vm.domains["worker"] = &domainRuntime{assignments: make(map[string]int)}
	vm.refreshFoodClass()
	return vm
}

// SetAge updates the current age for WorkerClassDef lookups
func (vm *WorkerManager) SetAge(ageKey string) {
	vm.ageKey = ageKey
	vm.refreshFoodClass()
}

// UnlockType marks workers as recruitable. Always routes to "worker" pool.
func (vm *WorkerManager) UnlockType(_ string) {
	vm.unlocked["worker"] = true
}

// IsUnlocked returns whether workers are available.
func (vm *WorkerManager) IsUnlocked(_ string) bool {
	return vm.unlocked["worker"]
}

// Recruit adds workers to the single pool. Returns false if not unlocked, if
// count is not positive, or if it would take the population over popCap.
func (vm *WorkerManager) Recruit(_ string, count int, popCap int) bool {
	rt := vm.domains["worker"]
	if !vm.unlocked["worker"] || count <= 0 {
		return false
	}
	// Compare against the room left rather than TotalPop()+count: a count near
	// math.MaxInt wraps that sum negative and sails under the cap.
	if count > popCap-vm.TotalPop() {
		return false
	}
	rt.count += count
	return true
}

// Assign assigns workers from the single pool to a specific building.
// The domain argument is ignored — any worker can go to any building with WorkerCapacity > 0.
func (vm *WorkerManager) Assign(_, buildingKey string, count int) bool {
	rt := vm.domains["worker"]
	byKey := vm.buildingDefs
	def, ok := byKey[buildingKey]
	if !ok || def.WorkerCapacity == 0 || count <= 0 {
		return false
	}
	if vm.IdleCount("worker") < count {
		return false
	}
	rt.assignments[buildingKey] += count
	return true
}

// Unassign removes workers from a building assignment. A count that is not
// positive is refused: a negative one would assign workers instead.
func (vm *WorkerManager) Unassign(_, buildingKey string, count int) bool {
	rt := vm.domains["worker"]
	if count <= 0 || rt.assignments[buildingKey] < count {
		return false
	}
	rt.assignments[buildingKey] -= count
	return true
}

// Dismiss removes workers from a building assignment and from the pool entirely.
// Returns the actual number dismissed.
func (vm *WorkerManager) Dismiss(buildingKey string, count int) int {
	rt := vm.domains["worker"]
	assigned := rt.assignments[buildingKey]
	if assigned <= 0 {
		return 0
	}
	if count > assigned {
		count = assigned
	}
	rt.assignments[buildingKey] -= count
	rt.count -= count
	if rt.count < 0 {
		rt.count = 0
	}
	return count
}

// IdleCount returns how many workers are not assigned to any building.
// The key argument is ignored.
func (vm *WorkerManager) IdleCount(_ string) int {
	rt := vm.domains["worker"]
	assigned := 0
	for _, c := range rt.assignments {
		assigned += c
	}
	return rt.count - assigned
}

// TotalPop returns total population
func (vm *WorkerManager) TotalPop() int {
	return vm.domains["worker"].count
}

// FoodDrain returns total food consumed per tick for all workers.
// The per-worker cost is looked up by (domain="food", age) so it scales with
// age — early ages have very low drain (primitive: 0.08/tick) to ease onboarding.
// Falls back to 0.1/worker if no class definition is found for the current age.
func (vm *WorkerManager) FoodDrain() float64 {
	rt := vm.domains["worker"]
	if rt.count == 0 {
		return 0
	}
	cls, ok := vm.foodClassFor()
	if !ok {
		return 0.1 * float64(rt.count)
	}
	return cls.FoodCost * float64(rt.count)
}

// GetAssignedCount returns how many workers are assigned to buildingKey.
// The domain argument is ignored — there is only one pool.
func (vm *WorkerManager) GetAssignedCount(_, buildingKey string) int {
	return vm.domains["worker"].assignments[buildingKey]
}

// RenameAssignment transfers all worker assignments from oldBuildingKey to newBuildingKey.
// Called during the age-transition building transformation pass. Domain arg is ignored.
func (vm *WorkerManager) RenameAssignment(_, oldKey, newKey string) {
	rt := vm.domains["worker"]
	count := rt.assignments[oldKey]
	if count > 0 {
		rt.assignments[newKey] += count
		delete(rt.assignments, oldKey)
	}
}

// GetDomainCount returns the number of workers assigned to buildings of the given domain.
// This is used for e.g. soldierCount (domain "military") — workers assigned to military buildings.
func (vm *WorkerManager) GetDomainCount(domain string) int {
	rt := vm.domains["worker"]
	byKey := vm.buildingDefs
	total := 0
	for buildingKey, count := range rt.assignments {
		if def, ok := byKey[buildingKey]; ok && def.WorkerDomain == domain {
			total += count
		}
	}
	return total
}

// GetProductionRates returns empty — villager production is now handled via
// building fill ratios in BuildingManager.WorkerScaledProduction.
func (vm *WorkerManager) GetProductionRates() map[string]float64 {
	return make(map[string]float64)
}

// KillWorker removes workers from starvation. Assignments are scaled down
// proportionally so assigned count never exceeds remaining population.
func (vm *WorkerManager) KillWorker(count int) int {
	rt := vm.domains["worker"]
	if rt.count <= 0 {
		return 0
	}
	if count > rt.count {
		count = rt.count
	}
	rt.count -= count
	// Reconcile: assigned total must not exceed new count
	for {
		assigned := 0
		for _, c := range rt.assignments {
			assigned += c
		}
		if assigned <= rt.count {
			break
		}
		// Find the largest assignment and reduce it by 1. Ties break on the
		// smaller key, as in RemovePct, so map order never picks the loser.
		maxKey := ""
		maxVal := 0
		for k, c := range rt.assignments {
			if c > maxVal || (c == maxVal && c > 0 && k < maxKey) {
				maxVal = c
				maxKey = k
			}
		}
		if maxKey == "" {
			break
		}
		rt.assignments[maxKey]--
	}
	return count
}

// AddLentWorkers adds workers lent by an allied civilization directly to the
// pool, deliberately bypassing the population cap (lent workers are on loan and
// may temporarily exceed your normal capacity). They arrive idle/unassigned;
// returning them later is done via KillWorker, which reconciles assignments.
func (vm *WorkerManager) AddLentWorkers(count int) {
	if count <= 0 {
		return
	}
	vm.domains["worker"].count += count
}

// AddPctAll adds a percentage of current count to the pool (used by Population Surge event).
func (vm *WorkerManager) AddPctAll(pct float64) {
	rt := vm.domains["worker"]
	add := int(float64(rt.count) * pct)
	if add > 0 {
		rt.count += add
	}
}

// RemovePct removes a percentage of workers from the pool (e.g. Epidemic event).
// Assignments are proportionally scaled down so the total assigned count never
// exceeds the remaining population. Integer truncation can leave the total
// assigned count one or two above the new count; the reconciliation loop at
// the bottom corrects this by reducing the largest assignment first.
//
// There is a single worker pool, so this already spans every worker domain:
// food, knowledge, military and the rest all lose the same share of their
// assigned workers (Endure's 25% loss relies on this).
func (vm *WorkerManager) RemovePct(pct float64) {
	rt := vm.domains["worker"]
	remove := int(float64(rt.count) * pct)
	if remove <= 0 {
		return
	}
	rt.count -= remove
	if rt.count < 0 {
		rt.count = 0
	}
	// Scale down assignments proportionally
	for k, assigned := range rt.assignments {
		reduced := assigned - int(float64(assigned)*pct)
		if reduced < 0 {
			reduced = 0
		}
		rt.assignments[k] = reduced
	}
	// Reconciliation: integer truncation can leave total assigned > new count.
	// Reduce the largest assignments one-by-one until total <= count.
	totalAssigned := 0
	for _, v := range rt.assignments {
		totalAssigned += v
	}
	for totalAssigned > rt.count {
		// Find the building with the most assigned workers and reduce it by 1.
		// Ties break on the smaller key so the result never depends on map
		// iteration order (keeps seeded catastrophe runs reproducible).
		maxKey := ""
		maxVal := 0
		for k, v := range rt.assignments {
			if v > maxVal || (v == maxVal && v > 0 && k < maxKey) {
				maxVal = v
				maxKey = k
			}
		}
		if maxKey == "" {
			break
		}
		rt.assignments[maxKey]--
		if rt.assignments[maxKey] == 0 {
			delete(rt.assignments, maxKey)
		}
		totalAssigned--
	}
}

// GetAll returns serializable worker info for saving, keyed by "worker".
func (vm *WorkerManager) GetAll() map[string]WorkerInfo {
	rt := vm.domains["worker"]
	assign := make(map[string]int, len(rt.assignments))
	for k, v := range rt.assignments {
		if v > 0 {
			assign[k] = v
		}
	}
	cls, _ := vm.foodClassFor()
	return map[string]WorkerInfo{
		"worker": {
			Count:      rt.count,
			FoodCost:   cls.FoodCost,
			Assignment: assign,
		},
	}
}

// LoadWorkers restores state from save data. Handles forward and backward
// compatibility: new saves have a single "worker" key; saves from before Phase
// 19 may have domain-specific keys ("food", "faith", "military", etc.). All
// domain counts are summed into the single pool and all valid assignments are
// merged, so old saves load without data loss.
func (vm *WorkerManager) LoadWorkers(data map[string]WorkerInfo) {
	rt := vm.domains["worker"]
	// Reset before loading to prevent accumulation on repeated loads.
	rt.assignments = make(map[string]int)
	rt.count = 0
	byBuilding := vm.buildingDefs
	totalCount := 0
	for _, info := range data {
		totalCount += info.Count
		for bKey, cnt := range info.Assignment {
			if def, ok := byBuilding[bKey]; ok && def.WorkerCapacity > 0 && cnt > 0 {
				rt.assignments[bKey] += cnt
			}
		}
	}
	rt.count = totalCount
}

// Snapshot returns worker state for UI consumption.
// Types map has a single "worker" key.
func (vm *WorkerManager) Snapshot(popCap int) WorkerState {
	rt := vm.domains["worker"]
	if rt == nil {
		return WorkerState{Types: map[string]WorkerDomainState{}}
	}
	idle := vm.IdleCount("worker")
	assign := make(map[string]int, len(rt.assignments))
	for k, v := range rt.assignments {
		if v > 0 {
			assign[k] = v
		}
	}
	className := ""
	drain := 0.0
	if vm.ageKey != "" {
		if wc, ok := vm.foodClassFor(); ok {
			className = wc.ClassName
			drain = wc.FoodCost * float64(rt.count)
		}
	}
	if className == "" {
		className = "Worker"
	}
	return WorkerState{
		Types: map[string]WorkerDomainState{
			"worker": {
				Name:        className,
				Count:       rt.count,
				IdleCount:   idle,
				Assignments: assign,
				Unlocked:    vm.unlocked["worker"],
			},
		},
		TotalPop:  rt.count,
		MaxPop:    popCap,
		TotalIdle: idle,
		FoodDrain: drain,
	}
}
