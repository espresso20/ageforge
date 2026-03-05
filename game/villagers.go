package game

import "github.com/espresso20/ageforge/config"

// legacyAlias maps old villager type keys to canonical domain keys
var legacyAlias = map[string]string{
	"worker":    "food",
	"workers":   "food", // plural alias — common typo
	"shaman":    "faith",
	"scholar":   "knowledge",
	"soldier":   "military",
	"merchant":  "trade",
	"engineer":  "engineering",
	"hacker":    "hacker",
	"astronaut": "astronaut",
}

// domainToLegacy is the reverse — used to key VillagerState.Types by old names for UI compat
var domainToLegacy = map[string]string{
	"food":        "worker",
	"faith":       "shaman",
	"knowledge":   "scholar",
	"military":    "soldier",
	"trade":       "merchant",
	"engineering": "engineer",
	"hacker":      "hacker",
	"astronaut":   "astronaut",
}

// resolveDomain returns the canonical domain key, translating legacy type names
func resolveDomain(key string) string {
	if d, ok := legacyAlias[key]; ok {
		return d
	}
	return key
}

// domainRuntime holds runtime state for one worker domain
type domainRuntime struct {
	count       int
	assignments map[string]int // buildingKey → assigned worker count
}

// VillagerManager manages population and domain-based worker assignments.
// Workers are assigned to buildings rather than resources; buildings then produce
// at a rate scaled by their fill ratio (20% floor + 80% × assigned/capacity).
type VillagerManager struct {
	domains  map[string]*domainRuntime
	unlocked map[string]bool // domain key → unlocked
	ageKey   string          // current age for WorkerClassDef lookups (food cost, class name)
}

// VillagerTypeDef is kept for UI backward compatibility (villager_panel.go)
type VillagerTypeDef struct {
	Name       string
	Key        string
	FoodCost   float64
	CanGather  []string
	GatherRate float64
}

// DefaultVillagerTypes returns legacy-style definitions for UI display (villager_panel.go).
// The GatherRate values reflect old pre-Phase-6 gather rates shown for reference.
func DefaultVillagerTypes() []VillagerTypeDef {
	return []VillagerTypeDef{
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

// NewVillagerManager creates a new domain-based villager manager
func NewVillagerManager() *VillagerManager {
	vm := &VillagerManager{
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
func (vm *VillagerManager) SetAge(ageKey string) {
	vm.ageKey = ageKey
}

// UnlockType makes a villager type (or domain) recruitable
func (vm *VillagerManager) UnlockType(key string) {
	vm.unlocked[resolveDomain(key)] = true
}

// IsUnlocked returns whether a type or domain is available
func (vm *VillagerManager) IsUnlocked(key string) bool {
	return vm.unlocked[resolveDomain(key)]
}

// Recruit adds workers to a domain. Returns false if not unlocked or over pop cap.
func (vm *VillagerManager) Recruit(key string, count int, popCap int) bool {
	domain := resolveDomain(key)
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
// domain may be a legacy type key ("worker") or a domain key ("food").
// buildingKey must be a building whose WorkerDomain matches the resolved domain.
func (vm *VillagerManager) Assign(domain, buildingKey string, count int) bool {
	domain = resolveDomain(domain)
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
func (vm *VillagerManager) Unassign(domain, buildingKey string, count int) bool {
	domain = resolveDomain(domain)
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
func (vm *VillagerManager) IdleCount(key string) int {
	domain := resolveDomain(key)
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
func (vm *VillagerManager) TotalPop() int {
	total := 0
	for _, rt := range vm.domains {
		total += rt.count
	}
	return total
}

// FoodDrain returns total food consumption per tick using WorkerClassDef costs
func (vm *VillagerManager) FoodDrain() float64 {
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
func (vm *VillagerManager) GetAssignedCount(domain, buildingKey string) int {
	rt, ok := vm.domains[domain]
	if !ok {
		return 0
	}
	return rt.assignments[buildingKey]
}

// RenameAssignment transfers all worker assignments from oldBuildingKey to newBuildingKey
// within the given domain. Called during the age-transition building transformation pass.
func (vm *VillagerManager) RenameAssignment(domain, oldKey, newKey string) {
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
func (vm *VillagerManager) GetDomainCount(domain string) int {
	rt, ok := vm.domains[domain]
	if !ok {
		return 0
	}
	return rt.count
}

// GetProductionRates returns empty — villager production is now handled via
// building fill ratios in BuildingManager.WorkerScaledProduction.
func (vm *VillagerManager) GetProductionRates() map[string]float64 {
	return make(map[string]float64)
}

// RemoveSoldiers removes workers from the military domain (expedition losses)
func (vm *VillagerManager) RemoveSoldiers(count int) {
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
func (vm *VillagerManager) AddPctAll(pct float64) {
	for _, rt := range vm.domains {
		add := int(float64(rt.count) * pct)
		if add > 0 {
			rt.count += add
		}
	}
}

// RemovePct removes a percentage of workers from all domains (used by Epidemic event).
// Each domain loses floor(count * pct) workers; assignments are proportionally reduced.
func (vm *VillagerManager) RemovePct(pct float64) {
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
func (vm *VillagerManager) GetAll() map[string]VillagerInfo {
	out := make(map[string]VillagerInfo)
	for domain, rt := range vm.domains {
		assign := make(map[string]int, len(rt.assignments))
		for k, v := range rt.assignments {
			if v > 0 {
				assign[k] = v
			}
		}
		cls, _ := config.WorkerClassByDomainAndAge(domain, vm.ageKey)
		out[domain] = VillagerInfo{
			Count:      rt.count,
			FoodCost:   cls.FoodCost,
			Assignment: assign,
		}
	}
	return out
}

// LoadVillagers restores domain state from save data.
// Handles both new (domain-keyed, buildingKey assignments) and
// old (type-keyed, resource assignments) save formats.
func (vm *VillagerManager) LoadVillagers(data map[string]VillagerInfo) {
	byBuilding := config.BuildingByKey()
	for key, info := range data {
		domain := resolveDomain(key)
		rt, ok := vm.domains[domain]
		if !ok {
			continue
		}
		rt.count = info.Count
		if info.Assignment != nil {
			for bKey, cnt := range info.Assignment {
				def, ok := byBuilding[bKey]
				if ok && def.WorkerDomain == domain && cnt > 0 {
					rt.assignments[bKey] = cnt
				}
				// Legacy resource keys (e.g. "food", "wood") are silently discarded;
				// workers become idle and can be reassigned to buildings.
			}
		}
	}
}

// Snapshot returns villager state for UI consumption.
// Types are keyed by domain keys (e.g. "food", "knowledge", "faith").
func (vm *VillagerManager) Snapshot(popCap int) VillagerState {
	state := VillagerState{
		Types:     make(map[string]VillagerTypeState),
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
		state.Types[domain] = VillagerTypeState{
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
