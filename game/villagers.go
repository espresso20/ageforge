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
// All workers belong to a single "worker" pool that can be assigned to any building,
// regardless of what domain/resource that building produces.
type WorkerManager struct {
	domains  map[string]*domainRuntime
	unlocked map[string]bool
	ageKey   string // current age for WorkerClassDef lookups (food cost, class name)
}

// NewWorkerManager creates a new single-pool worker manager.
// All workers belong to a single "worker" pool — domain restrictions are removed.
func NewWorkerManager() *WorkerManager {
	vm := &WorkerManager{
		domains:  make(map[string]*domainRuntime),
		unlocked: make(map[string]bool),
		ageKey:   "primitive_age",
	}
	vm.domains["worker"] = &domainRuntime{assignments: make(map[string]int)}
	return vm
}

// SetAge updates the current age for WorkerClassDef lookups
func (vm *WorkerManager) SetAge(ageKey string) {
	vm.ageKey = ageKey
}

// UnlockType marks workers as recruitable. Always routes to "worker" pool.
func (vm *WorkerManager) UnlockType(_ string) {
	vm.unlocked["worker"] = true
}

// IsUnlocked returns whether workers are available.
func (vm *WorkerManager) IsUnlocked(_ string) bool {
	return vm.unlocked["worker"]
}

// Recruit adds workers to the single pool. Returns false if not unlocked or over pop cap.
func (vm *WorkerManager) Recruit(_ string, count int, popCap int) bool {
	rt := vm.domains["worker"]
	if !vm.unlocked["worker"] {
		return false
	}
	if vm.TotalPop()+count > popCap {
		return false
	}
	rt.count += count
	return true
}

// Assign assigns workers from the single pool to a specific building.
// The domain argument is ignored — any worker can go to any building with WorkerCapacity > 0.
func (vm *WorkerManager) Assign(_, buildingKey string, count int) bool {
	rt := vm.domains["worker"]
	byKey := config.BuildingByKey()
	def, ok := byKey[buildingKey]
	if !ok || def.WorkerCapacity == 0 {
		return false
	}
	if vm.IdleCount("worker") < count {
		return false
	}
	rt.assignments[buildingKey] += count
	return true
}

// Unassign removes workers from a building assignment
func (vm *WorkerManager) Unassign(_, buildingKey string, count int) bool {
	rt := vm.domains["worker"]
	if rt.assignments[buildingKey] < count {
		return false
	}
	rt.assignments[buildingKey] -= count
	return true
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

// FoodDrain returns total food consumption per tick using WorkerClassDef costs.
// Uses the "food" domain class progression for naming/cost lookups.
func (vm *WorkerManager) FoodDrain() float64 {
	rt := vm.domains["worker"]
	if rt.count == 0 {
		return 0
	}
	cls, ok := config.WorkerClassByDomainAndAge("food", vm.ageKey)
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

// GetDomainCount returns the total worker count. Domain arg is ignored.
func (vm *WorkerManager) GetDomainCount(_ string) int {
	return vm.domains["worker"].count
}

// GetProductionRates returns empty — villager production is now handled via
// building fill ratios in BuildingManager.WorkerScaledProduction.
func (vm *WorkerManager) GetProductionRates() map[string]float64 {
	return make(map[string]float64)
}

// RemoveSoldiers removes workers from the pool (expedition losses).
// Previously removed from "military" domain; now removes from the single pool.
func (vm *WorkerManager) RemoveSoldiers(count int) {
	rt := vm.domains["worker"]
	rt.count -= count
	if rt.count < 0 {
		rt.count = 0
	}
}

// AddPctAll adds a percentage of current count to the pool (used by Population Surge event).
func (vm *WorkerManager) AddPctAll(pct float64) {
	rt := vm.domains["worker"]
	add := int(float64(rt.count) * pct)
	if add > 0 {
		rt.count += add
	}
}

// RemovePct removes a percentage of workers from the pool (used by Epidemic event).
// Assignments are proportionally reduced.
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
	cls, _ := config.WorkerClassByDomainAndAge("food", vm.ageKey)
	return map[string]WorkerInfo{
		"worker": {
			Count:      rt.count,
			FoodCost:   cls.FoodCost,
			Assignment: assign,
		},
	}
}

// LoadWorkers restores state from save data.
// Handles both new single-pool saves (keyed "worker") and old domain-keyed saves
// (e.g. "food", "faith") by summing all domain counts into the single "worker" pool.
func (vm *WorkerManager) LoadWorkers(data map[string]WorkerInfo) {
	rt := vm.domains["worker"]
	byBuilding := config.BuildingByKey()
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
		if wc, ok := config.WorkerClassByDomainAndAge("food", vm.ageKey); ok {
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
