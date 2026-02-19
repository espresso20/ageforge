package game

// VillagerTypeDef defines a villager type's properties
type VillagerTypeDef struct {
	Name     string
	Key      string
	FoodCost float64
	// What resources this type can be assigned to gather
	CanGather []string
	// Production rate per villager per tick when assigned
	GatherRate float64
}

// VillagerManager manages population and assignments
type VillagerManager struct {
	types      map[string]*villagerRuntime
	unlocked   map[string]bool
	definitions map[string]VillagerTypeDef
}

type villagerRuntime struct {
	count      int
	assignment map[string]int // resource key -> assigned count
}

// DefaultVillagerTypes returns the base villager type definitions
func DefaultVillagerTypes() []VillagerTypeDef {
	return []VillagerTypeDef{
		{
			Name: "Worker", Key: "worker", FoodCost: 0.15,
			CanGather:  []string{"food", "wood", "stone", "iron", "gold", "coal", "oil", "electricity", "uranium", "titanium"},
			GatherRate: 0.3,
		},
		{
			Name: "Shaman", Key: "shaman", FoodCost: 0.2,
			CanGather:  []string{"knowledge"},
			GatherRate: 0.4,
		},
		{
			Name: "Scholar", Key: "scholar", FoodCost: 0.2,
			CanGather:  []string{"knowledge", "culture", "data"},
			GatherRate: 0.5,
		},
		{
			Name: "Soldier", Key: "soldier", FoodCost: 0.25,
			CanGather:  []string{}, // soldiers don't gather, used for military
			GatherRate: 0,
		},
		{
			Name: "Merchant", Key: "merchant", FoodCost: 0.2,
			CanGather:  []string{"gold", "crypto"},
			GatherRate: 0.6,
		},
		{
			Name: "Engineer", Key: "engineer", FoodCost: 0.25,
			CanGather:  []string{"oil", "electricity", "data"},
			GatherRate: 0.7,
		},
		{
			Name: "Hacker", Key: "hacker", FoodCost: 0.3,
			CanGather:  []string{"data", "crypto"},
			GatherRate: 0.8,
		},
		{
			Name: "Astronaut", Key: "astronaut", FoodCost: 0.4,
			CanGather:  []string{"titanium", "dark_matter", "plasma"},
			GatherRate: 1.0,
		},
	}
}

// NewVillagerManager creates a new villager manager
func NewVillagerManager() *VillagerManager {
	vm := &VillagerManager{
		types:       make(map[string]*villagerRuntime),
		unlocked:    make(map[string]bool),
		definitions: make(map[string]VillagerTypeDef),
	}
	for _, def := range DefaultVillagerTypes() {
		vm.definitions[def.Key] = def
		vm.types[def.Key] = &villagerRuntime{
			count:      0,
			assignment: make(map[string]int),
		}
	}
	return vm
}

// UnlockType makes a villager type recruitable
func (vm *VillagerManager) UnlockType(key string) {
	vm.unlocked[key] = true
}

// IsUnlocked returns whether a villager type is available
func (vm *VillagerManager) IsUnlocked(key string) bool {
	return vm.unlocked[key]
}

// Recruit adds villagers of a type. Returns false if not unlocked or over pop cap.
func (vm *VillagerManager) Recruit(key string, count int, popCap int) bool {
	if !vm.unlocked[key] {
		return false
	}
	rt, ok := vm.types[key]
	if !ok {
		return false
	}
	if vm.TotalPop()+count > popCap {
		return false
	}
	rt.count += count
	return true
}

// Assign assigns villagers to gather a resource
func (vm *VillagerManager) Assign(vType, resource string, count int) bool {
	rt, ok := vm.types[vType]
	if !ok {
		return false
	}
	def := vm.definitions[vType]
	// Check this type can gather this resource
	canGather := false
	for _, r := range def.CanGather {
		if r == resource {
			canGather = true
			break
		}
	}
	if !canGather {
		return false
	}
	idle := vm.IdleCount(vType)
	if idle < count {
		return false
	}
	rt.assignment[resource] += count
	return true
}

// Unassign removes villagers from a resource assignment
func (vm *VillagerManager) Unassign(vType, resource string, count int) bool {
	rt, ok := vm.types[vType]
	if !ok {
		return false
	}
	assigned := rt.assignment[resource]
	if assigned < count {
		return false
	}
	rt.assignment[resource] -= count
	return true
}

// IdleCount returns how many of a type are not assigned
func (vm *VillagerManager) IdleCount(vType string) int {
	rt, ok := vm.types[vType]
	if !ok {
		return 0
	}
	assigned := 0
	for _, c := range rt.assignment {
		assigned += c
	}
	return rt.count - assigned
}

// TotalPop returns total population across all types
func (vm *VillagerManager) TotalPop() int {
	total := 0
	for _, rt := range vm.types {
		total += rt.count
	}
	return total
}

// FoodDrain returns total food consumption per tick
func (vm *VillagerManager) FoodDrain() float64 {
	drain := 0.0
	for key, rt := range vm.types {
		def := vm.definitions[key]
		drain += def.FoodCost * float64(rt.count)
	}
	return drain
}

// GetProductionRates returns resource production from assigned villagers
func (vm *VillagerManager) GetProductionRates() map[string]float64 {
	rates := make(map[string]float64)
	for key, rt := range vm.types {
		def := vm.definitions[key]
		for resource, count := range rt.assignment {
			rates[resource] += def.GatherRate * float64(count)
		}
	}
	return rates
}

// GetAll returns serializable villager info (for save)
func (vm *VillagerManager) GetAll() map[string]VillagerInfo {
	out := make(map[string]VillagerInfo)
	for key, rt := range vm.types {
		def := vm.definitions[key]
		assign := make(map[string]int)
		for k, v := range rt.assignment {
			assign[k] = v
		}
		out[key] = VillagerInfo{
			Count:      rt.count,
			FoodCost:   def.FoodCost,
			Assignment: assign,
		}
	}
	return out
}

// LoadVillagers restores villager state from save data
func (vm *VillagerManager) LoadVillagers(data map[string]VillagerInfo) {
	for key, info := range data {
		if rt, ok := vm.types[key]; ok {
			rt.count = info.Count
			rt.assignment = info.Assignment
		}
	}
}

// RemoveSoldiers removes soldiers (from expedition losses)
func (vm *VillagerManager) RemoveSoldiers(count int) {
	rt, ok := vm.types["soldier"]
	if !ok {
		return
	}
	rt.count -= count
	if rt.count < 0 {
		rt.count = 0
	}
}

// Snapshot returns villager state for UI
func (vm *VillagerManager) Snapshot(popCap int) VillagerState {
	state := VillagerState{
		Types:     make(map[string]VillagerTypeState),
		MaxPop:    popCap,
		TotalPop:  vm.TotalPop(),
		FoodDrain: vm.FoodDrain(),
	}
	for key, rt := range vm.types {
		def := vm.definitions[key]
		idle := vm.IdleCount(key)
		assign := make(map[string]int)
		for k, v := range rt.assignment {
			assign[k] = v
		}
		state.Types[key] = VillagerTypeState{
			Name:        def.Name,
			Count:       rt.count,
			IdleCount:   idle,
			Assignments: assign,
			Unlocked:    vm.unlocked[key],
		}
		state.TotalIdle += idle
	}
	return state
}
