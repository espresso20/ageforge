package game

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
	"github.com/espresso20/ageforge/vibe"
)

// WorkerManager manages population and worker assignments.
//
// Every worker belongs to one pool, and any worker can be assigned to any
// building with WorkerCapacity > 0.
type WorkerManager struct {
	rules *rules.Set
	// count is the whole population and assignments is the workers at each
	// building (buildingKey → assigned worker count).
	count       int
	assignments map[string]int
	unlocked    bool
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
	vm.foodClass, vm.foodClassOK = vm.rules.WorkerClass("food", vm.ageKey)
}

// NewWorkerManagerWith creates a worker manager on set.
func NewWorkerManagerWith(set *rules.Set) *WorkerManager {
	vm := &WorkerManager{
		rules:        set,
		assignments:  make(map[string]int),
		ageKey:       "primitive_age",
		buildingDefs: set.BuildingMap(),
	}
	vm.refreshFoodClass()
	return vm
}

// Rebind moves the manager onto set: it takes set's building definitions
// and worker classes. The pool and its assignments stay.
func (vm *WorkerManager) Rebind(set *rules.Set) {
	vm.rules = set
	vm.buildingDefs = set.BuildingMap()
	vm.refreshFoodClass()
}

// SetAge updates the current age for WorkerClassDef lookups
func (vm *WorkerManager) SetAge(ageKey string) {
	vm.ageKey = ageKey
	vm.refreshFoodClass()
}

// Unlock marks workers as recruitable.
func (vm *WorkerManager) Unlock() {
	vm.unlocked = true
}

// IsUnlocked returns whether workers are available.
func (vm *WorkerManager) IsUnlocked() bool {
	return vm.unlocked
}

// Recruit adds workers to the pool. Returns false if not unlocked, if count
// is not positive, or if it would take the population over popCap.
func (vm *WorkerManager) Recruit(count int, popCap int) bool {
	if !vm.unlocked || count <= 0 {
		return false
	}
	// Compare against the room left rather than TotalPop()+count: a count near
	// math.MaxInt wraps that sum negative and sails under the cap.
	if count > popCap-vm.TotalPop() {
		return false
	}
	vm.count += count
	return true
}

// Assign assigns idle workers to a building that takes workers.
func (vm *WorkerManager) Assign(buildingKey string, count int) bool {
	def, ok := vm.buildingDefs[buildingKey]
	if !ok || def.WorkerCapacity == 0 || count <= 0 {
		return false
	}
	if vm.IdleCount() < count {
		return false
	}
	vm.assignments[buildingKey] += count
	return true
}

// Unassign removes workers from a building assignment. A count that is not
// positive is refused: a negative one would assign workers instead.
func (vm *WorkerManager) Unassign(buildingKey string, count int) bool {
	if count <= 0 || vm.assignments[buildingKey] < count {
		return false
	}
	vm.assignments[buildingKey] -= count
	return true
}

// Dismiss removes workers from a building assignment and from the pool entirely.
// Returns the actual number dismissed.
func (vm *WorkerManager) Dismiss(buildingKey string, count int) int {
	assigned := vm.assignments[buildingKey]
	if assigned <= 0 {
		return 0
	}
	if count > assigned {
		count = assigned
	}
	vm.assignments[buildingKey] -= count
	vm.count -= count
	if vm.count < 0 {
		vm.count = 0
	}
	return count
}

// IdleCount returns how many workers are not assigned to any building.
func (vm *WorkerManager) IdleCount() int {
	return vm.count - vm.assignedTotal()
}

// assignedTotal is how many workers are assigned to buildings.
func (vm *WorkerManager) assignedTotal() int {
	total := 0
	for _, count := range vm.assignments {
		total += count
	}
	return total
}

// TakeIdle removes up to n workers with no job from the pool (they leave it
// for a post as soldiers) and returns how many went.
func (vm *WorkerManager) TakeIdle(n int) int {
	n = min(n, vm.IdleCount())
	if n <= 0 {
		return 0
	}
	vm.count -= n
	return n
}

// TotalPop returns total population
func (vm *WorkerManager) TotalPop() int {
	return vm.count
}

// FoodDrain returns total food consumed per tick for all workers.
// The per-worker cost is looked up by (domain="food", age) so it scales with
// age — early ages have very low drain (primitive: 0.08/tick) to ease onboarding.
// Falls back to 0.1/worker if no class definition is found for the current age.
func (vm *WorkerManager) FoodDrain() float64 {
	if vm.count == 0 {
		return 0
	}
	return vm.FoodCostPerWorker() * float64(vm.count)
}

// FoodCostPerWorker is the food each worker eats a tick in the current age:
// every worker pays the food class's cost (0.1 when no class matches).
func (vm *WorkerManager) FoodCostPerWorker() float64 {
	if cls, ok := vm.foodClassFor(); ok {
		return cls.FoodCost
	}
	return 0.1
}

// GetAssignedCount returns how many workers are assigned to buildingKey.
func (vm *WorkerManager) GetAssignedCount(buildingKey string) int {
	return vm.assignments[buildingKey]
}

// GetDomainCount returns the number of workers assigned to buildings of the given domain.
// This is used for e.g. soldierCount (domain "military") — workers assigned to military buildings.
func (vm *WorkerManager) GetDomainCount(domain string) int {
	total := 0
	for buildingKey, count := range vm.assignments {
		if def, ok := vm.buildingDefs[buildingKey]; ok && def.WorkerDomain == domain {
			total += count
		}
	}
	return total
}

// KillWorker removes workers from starvation. Assignments are scaled down
// proportionally so assigned count never exceeds remaining population.
func (vm *WorkerManager) KillWorker(count int) int {
	if vm.count <= 0 {
		return 0
	}
	count = min(count, vm.count)
	vm.count -= count
	vm.trimAssignments()
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
	vm.count += count
}

// AddPctAll adds a percentage of current count to the pool (used by Population Surge event).
func (vm *WorkerManager) AddPctAll(pct float64) {
	if add := int(float64(vm.count) * pct); add > 0 {
		vm.count += add
	}
}

// RemovePct removes a percentage of workers from the pool (e.g. Epidemic event).
// Every building loses the same share of its assigned workers, then
// trimAssignments takes off what integer truncation left over. There is a
// single worker pool, so food, knowledge, military and the rest all lose the
// same share (Endure's 25% loss relies on this).
func (vm *WorkerManager) RemovePct(pct float64) {
	remove := int(float64(vm.count) * pct)
	if remove <= 0 {
		return
	}
	vm.count = max(vm.count-remove, 0)
	for key, assigned := range vm.assignments {
		vm.assignments[key] = max(assigned-int(float64(assigned)*pct), 0)
	}
	vm.trimAssignments()
}

// trimAssignments takes workers off the largest assignments, one at a time,
// until no more are assigned than the pool holds. Ties break on the smaller
// key so the result never depends on map iteration order (seeded runs stay
// reproducible). An assignment trimmed to nothing is dropped.
func (vm *WorkerManager) trimAssignments() {
	for excess := vm.assignedTotal() - vm.count; excess > 0; excess-- {
		largestKey := ""
		largest := 0
		for key, assigned := range vm.assignments {
			if assigned > largest || (assigned == largest && assigned > 0 && key < largestKey) {
				largest = assigned
				largestKey = key
			}
		}
		if largestKey == "" {
			return
		}
		vm.assignments[largestKey]--
		if vm.assignments[largestKey] == 0 {
			delete(vm.assignments, largestKey)
		}
	}
}

// GetAll returns serializable worker info for saving, keyed by "worker".
func (vm *WorkerManager) GetAll() map[string]WorkerInfo {
	return map[string]WorkerInfo{
		"worker": {
			Count:      vm.count,
			FoodCost:   vm.foodClass.FoodCost,
			Assignment: vm.positiveAssignments(),
		},
	}
}

// positiveAssignments is a copy of the assignments that hold workers.
func (vm *WorkerManager) positiveAssignments() map[string]int {
	assign := make(map[string]int, len(vm.assignments))
	for key, count := range vm.assignments {
		if count > 0 {
			assign[key] = count
		}
	}
	return assign
}

// LoadWorkers restores state from save data: the single "worker" pool, with
// the assignments that name a building that takes workers.
func (vm *WorkerManager) LoadWorkers(data map[string]WorkerInfo) {
	info := data["worker"]
	vm.count = info.Count
	// Reset before loading to prevent accumulation on repeated loads.
	vm.assignments = make(map[string]int)
	for key, count := range info.Assignment {
		if def, ok := vm.buildingDefs[key]; ok && def.WorkerCapacity > 0 && count > 0 {
			vm.assignments[key] = count
		}
	}
}

// Snapshot returns worker state for UI consumption.
// Types map has a single "worker" key.
func (vm *WorkerManager) Snapshot(popCap int) WorkerState {
	idle := vm.IdleCount()
	className := ""
	drain := 0.0
	if vm.ageKey != "" {
		if wc, ok := vm.foodClassFor(); ok {
			className = wc.ClassName
			drain = wc.FoodCost * float64(vm.count)
		}
	}
	if className == "" {
		className = vibe.Line("game.villagers.Snapshot.01", "Worker")
	}
	return WorkerState{
		Types: map[string]WorkerDomainState{
			"worker": {
				Name:        className,
				Count:       vm.count,
				IdleCount:   idle,
				Assignments: vm.positiveAssignments(),
				Unlocked:    vm.unlocked,
			},
		},
		TotalPop:  vm.count,
		MaxPop:    popCap,
		TotalIdle: idle,
		FoodDrain: drain,
	}
}
