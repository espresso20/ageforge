package game

import (
	"strings"

	"github.com/espresso20/ageforge/config"
)

// domainRuntime holds runtime state for one worker domain
type domainRuntime struct {
	count       int
	assignments map[string]int // buildingKey → assigned worker count
}

// WorkerManager manages population and domain-based worker assignments.
// Workers are assigned to buildings rather than resources; buildings then produce
// at a rate scaled by their fill ratio (20% floor + 80% × assigned/capacity).
type WorkerManager struct {
	domains  map[string]*domainRuntime
	unlocked map[string]bool // domain key → unlocked
	ageKey   string          // current age for WorkerClassDef lookups (food cost, class name)
}

// WorkerTypeDef is kept for UI backward compatibility (villager_panel.go)
type WorkerTypeDef struct {
	Name       string
	Key        string
	FoodCost   float64
	CanGather  []string
	GatherRate float64
}

// DefaultWorkerTypes returns legacy-style definitions for UI display (villager_panel.go).
// The GatherRate values reflect old pre-Phase-6 gather rates shown for reference.
func DefaultWorkerTypes() []WorkerTypeDef {
	return []WorkerTypeDef{
		{Name: "Worker", Key: "worker", FoodCost: 0.10, CanGather: []string{"food"}, GatherRate: 0.35},
		{Name: "Shaman", Key: "shaman", FoodCost: 0.5, CanGather: []string{"faith"}, GatherRate: 0.01},
		{Name: "Scholar", Key: "scholar", FoodCost: 0.8, CanGather: []string{"knowledge"}, GatherRate: 0.05},
		{Name: "Soldier", Key: "soldier", FoodCost: 0.25, CanGather: []string{}, GatherRate: 0},
		{Name: "Merchant", Key: "merchant", FoodCost: 0.2, CanGather: []string{"gold"}, GatherRate: 0.3},
		{Name: "Engineer", Key: "engineer", FoodCost: 0.25, CanGather: []string{"electricity"}, GatherRate: 0.35},
		{Name: "Hacker", Key: "hacker", FoodCost: 0.3, CanGather: []string{"data"}, GatherRate: 0.4},
		{Name: "Astronaut", Key: "astronaut", FoodCost: 0.4, CanGather: []string{"titanium"}, GatherRate: 0.5},
	}
}

// NewWorkerManager creates a new domain-based villager manager
func NewWorkerManager() *WorkerManager {
	vm := &WorkerManager{
		domains:  make(map[string]*domainRuntime),
		unlocked: make(map[string]bool),
		ageKey:   "primitive_age",
	}
	for _, d := range config.WorkerDomains() {
		vm.domains[d] = &domainRuntime{assignments: make(map[string]int)}
	}
	return vm
}

// SetAge updates the current age for WorkerClassDef lookups
func (vm *WorkerManager) SetAge(ageKey string) {
	vm.ageKey = ageKey
}

// UnlockType makes a domain recruitable
func (vm *WorkerManager) UnlockType(key string) {
	vm.unlocked[key] = true
}

// IsUnlocked returns whether a domain is available
func (vm *WorkerManager) IsUnlocked(key string) bool {
	return vm.unlocked[key]
}

// Recruit adds workers to a domain. Returns false if not unlocked or over pop cap.
func (vm *WorkerManager) Recruit(key string, count int, popCap int) bool {
	domain := key
	rt, ok := vm.domains[domain]
	if !ok || !vm.unlocked[domain] {
		return false
	}
	if vm.TotalPop()+count > popCap {
		return false
	}
	rt.count += count
	return true
}

// Assign assigns workers from a domain to a specific building.
// domain must be a domain key ("food", "faith", etc.).
// buildingKey must be a building whose WorkerDomain matches the domain.
func (vm *WorkerManager) Assign(domain, buildingKey string, count int) bool {
	rt, ok := vm.domains[domain]
	if !ok {
		return false
	}
	byKey := config.BuildingByKey()
	def, ok := byKey[buildingKey]
	if !ok || def.WorkerDomain != domain {
		return false
	}
	if vm.IdleCount(domain) < count {
		return false
	}
	rt.assignments[buildingKey] += count
	return true
}

// Unassign removes workers of a domain from a building assignment
func (vm *WorkerManager) Unassign(domain, buildingKey string, count int) bool {
	rt, ok := vm.domains[domain]
	if !ok {
		return false
	}
	if rt.assignments[buildingKey] < count {
		return false
	}
	rt.assignments[buildingKey] -= count
	return true
}

// IdleCount returns how many workers in a domain are not assigned to any building
func (vm *WorkerManager) IdleCount(key string) int {
	domain := key
	rt, ok := vm.domains[domain]
	if !ok {
		return 0
	}
	assigned := 0
	for _, c := range rt.assignments {
		assigned += c
	}
	return rt.count - assigned
}

// TotalPop returns total population across all domains
func (vm *WorkerManager) TotalPop() int {
	total := 0
	for _, rt := range vm.domains {
		total += rt.count
	}
	return total
}

// FoodDrain returns total food consumption per tick using WorkerClassDef costs
func (vm *WorkerManager) FoodDrain() float64 {
	drain := 0.0
	for domain, rt := range vm.domains {
		if rt.count == 0 {
			continue
		}
		cls, ok := config.WorkerClassByDomainAndAge(domain, vm.ageKey)
		if !ok {
			drain += 0.1 * float64(rt.count)
			continue
		}
		drain += cls.FoodCost * float64(rt.count)
	}
	return drain
}

// GetAssignedCount returns how many workers from a domain are assigned to buildingKey.
// Used by BuildingManager.WorkerScaledProduction for the fill-ratio formula.
func (vm *WorkerManager) GetAssignedCount(domain, buildingKey string) int {
	rt, ok := vm.domains[domain]
	if !ok {
		return 0
	}
	return rt.assignments[buildingKey]
}

// RenameAssignment transfers all worker assignments from oldBuildingKey to newBuildingKey
// within the given domain. Called during the age-transition building transformation pass.
func (vm *WorkerManager) RenameAssignment(domain, oldKey, newKey string) {
	rt, ok := vm.domains[domain]
	if !ok {
		return
	}
	count := rt.assignments[oldKey]
	if count > 0 {
		rt.assignments[newKey] += count
		delete(rt.assignments, oldKey)
	}
}

// GetDomainCount returns the total worker count for a domain
func (vm *WorkerManager) GetDomainCount(domain string) int {
	rt, ok := vm.domains[domain]
	if !ok {
		return 0
	}
	return rt.count
}

// GetProductionRates returns empty — villager production is now handled via
// building fill ratios in BuildingManager.WorkerScaledProduction.
func (vm *WorkerManager) GetProductionRates() map[string]float64 {
	return make(map[string]float64)
}

// RemoveSoldiers removes workers from the military domain (expedition losses)
func (vm *WorkerManager) RemoveSoldiers(count int) {
	rt, ok := vm.domains["military"]
	if !ok {
		return
	}
	rt.count -= count
	if rt.count < 0 {
		rt.count = 0
	}
}

// AddPctAll adds a percentage of current count to all domains (used by Population Surge event).
func (vm *WorkerManager) AddPctAll(pct float64) {
	for _, rt := range vm.domains {
		add := int(float64(rt.count) * pct)
		if add > 0 {
			rt.count += add
		}
	}
}

// RemovePct removes a percentage of workers from all domains (used by Epidemic event).
// Each domain loses floor(count * pct) workers; assignments are proportionally reduced.
func (vm *WorkerManager) RemovePct(pct float64) {
	for _, rt := range vm.domains {
		remove := int(float64(rt.count) * pct)
		if remove <= 0 {
			continue
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
}

// GetAll returns serializable domain-keyed villager info for saving
func (vm *WorkerManager) GetAll() map[string]WorkerInfo {
	out := make(map[string]WorkerInfo)
	for domain, rt := range vm.domains {
		assign := make(map[string]int, len(rt.assignments))
		for k, v := range rt.assignments {
			if v > 0 {
				assign[k] = v
			}
		}
		cls, _ := config.WorkerClassByDomainAndAge(domain, vm.ageKey)
		out[domain] = WorkerInfo{
			Count:      rt.count,
			FoodCost:   cls.FoodCost,
			Assignment: assign,
		}
	}
	return out
}

// LoadWorkers restores domain state from save data.
// Expects domain-keyed data (e.g. "food", "faith"); legacy type-keyed saves
// (e.g. "worker", "shaman") will simply be skipped (workers become idle).
func (vm *WorkerManager) LoadWorkers(data map[string]WorkerInfo) {
	byBuilding := config.BuildingByKey()
	for key, info := range data {
		rt, ok := vm.domains[key]
		if !ok {
			continue
		}
		rt.count = info.Count
		if info.Assignment != nil {
			for bKey, cnt := range info.Assignment {
				def, ok := byBuilding[bKey]
				if ok && def.WorkerDomain == key && cnt > 0 {
					rt.assignments[bKey] = cnt
				}
				// Legacy resource keys (e.g. "food", "wood") are silently discarded;
				// workers become idle and can be reassigned to buildings.
			}
		}
	}
}

// ResolveClassToDomain maps a class name or legacy type name to a domain key.
// Accepts: domain keys ("food"), current class names ("forager"), legacy type names ("worker").
func (vm *WorkerManager) ResolveClassToDomain(nameOrDomain string) string {
	lower := strings.ToLower(nameOrDomain)

	// 1. Already a domain key?
	if _, ok := vm.domains[lower]; ok {
		return lower
	}

	// 2. Match current class name for any domain
	for domain := range vm.domains {
		if vm.ageKey != "" {
			wc, ok := config.WorkerClassByDomainAndAge(domain, vm.ageKey)
			if ok && strings.ToLower(wc.ClassName) == lower {
				return domain
			}
		}
	}

	// 3. Legacy save/input alias fallback
	legacyFallback := map[string]string{
		"worker": "food", "workers": "food",
		"shaman":   "faith",
		"scholar":  "knowledge",
		"soldier":  "military",
		"merchant": "trade",
		"engineer": "engineering",
	}
	if canon, ok := legacyFallback[lower]; ok {
		return canon
	}

	return nameOrDomain // return as-is; will fail gracefully downstream
}

// Snapshot returns villager state for UI consumption.
// Types are keyed by domain keys (e.g. "food", "knowledge", "faith").
func (vm *WorkerManager) Snapshot(popCap int) WorkerState {
	state := WorkerState{
		Types:     make(map[string]WorkerDomainState),
		MaxPop:    popCap,
		TotalPop:  vm.TotalPop(),
		FoodDrain: vm.FoodDrain(),
	}
	for domain, rt := range vm.domains {
		idle := vm.IdleCount(domain)
		assign := make(map[string]int)
		for k, v := range rt.assignments {
			if v > 0 {
				assign[k] = v
			}
		}
		cls, _ := config.WorkerClassByDomainAndAge(domain, vm.ageKey)
		name := cls.ClassName
		if name == "" {
			name = domain
		}
		state.Types[domain] = WorkerDomainState{
			Name:        name,
			Count:       rt.count,
			IdleCount:   idle,
			Assignments: assign,
			Unlocked:    vm.unlocked[domain],
		}
		state.TotalIdle += idle
	}
	return state
}
