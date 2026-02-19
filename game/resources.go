package game

import "github.com/user/ageforge/config"

// Resource holds the runtime state of a single resource
type Resource struct {
	Amount  float64
	Rate    float64
	Storage float64
}

// ResourceManager manages all resources
type ResourceManager struct {
	resources map[string]*Resource
	defs      map[string]config.ResourceDef
	unlocked  map[string]bool
}

// NewResourceManager creates a resource manager with base definitions
func NewResourceManager() *ResourceManager {
	rm := &ResourceManager{
		resources: make(map[string]*Resource),
		defs:      config.ResourceByKey(),
		unlocked:  make(map[string]bool),
	}
	// Initialize all resources
	for _, def := range config.BaseResources() {
		rm.resources[def.Key] = &Resource{
			Amount:  0,
			Rate:    0,
			Storage: def.BaseStorage,
		}
	}
	return rm
}

// UnlockResource makes a resource visible/usable
func (rm *ResourceManager) UnlockResource(key string) {
	rm.unlocked[key] = true
}

// IsUnlocked returns whether a resource is unlocked
func (rm *ResourceManager) IsUnlocked(key string) bool {
	return rm.unlocked[key]
}

// Get returns the current amount of a resource
func (rm *ResourceManager) Get(key string) float64 {
	if r, ok := rm.resources[key]; ok {
		return r.Amount
	}
	return 0
}

// GetStorage returns the storage cap for a resource
func (rm *ResourceManager) GetStorage(key string) float64 {
	if r, ok := rm.resources[key]; ok {
		return r.Storage
	}
	return 0
}

// Add adds an amount to a resource, respecting storage limits
func (rm *ResourceManager) Add(key string, amount float64) float64 {
	r, ok := rm.resources[key]
	if !ok {
		return 0
	}
	r.Amount += amount
	if r.Amount > r.Storage {
		r.Amount = r.Storage
	}
	if r.Amount < 0 {
		r.Amount = 0
	}
	return r.Amount
}

// Remove subtracts from a resource. Returns false if insufficient.
func (rm *ResourceManager) Remove(key string, amount float64) bool {
	r, ok := rm.resources[key]
	if !ok || r.Amount < amount {
		return false
	}
	r.Amount -= amount
	return true
}

// CanAfford checks if all costs can be paid
func (rm *ResourceManager) CanAfford(costs map[string]float64) bool {
	for key, amount := range costs {
		if rm.Get(key) < amount {
			return false
		}
	}
	return true
}

// Pay deducts all costs. Returns false if can't afford (no partial deduction).
func (rm *ResourceManager) Pay(costs map[string]float64) bool {
	if !rm.CanAfford(costs) {
		return false
	}
	for key, amount := range costs {
		rm.Remove(key, amount)
	}
	return true
}

// SetRate sets the production rate for a resource
func (rm *ResourceManager) SetRate(key string, rate float64) {
	if r, ok := rm.resources[key]; ok {
		r.Rate = rate
	}
}

// AddStorage increases storage cap for a resource
func (rm *ResourceManager) AddStorage(key string, amount float64) {
	if r, ok := rm.resources[key]; ok {
		r.Storage += amount
	}
}

// ApplyRates applies per-tick production rates
func (rm *ResourceManager) ApplyRates() {
	for key, r := range rm.resources {
		if rm.unlocked[key] && r.Rate != 0 {
			rm.Add(key, r.Rate)
		}
	}
}

// GetAll returns all resource amounts (for save)
func (rm *ResourceManager) GetAll() map[string]float64 {
	out := make(map[string]float64)
	for key, r := range rm.resources {
		out[key] = r.Amount
	}
	return out
}

// GetAllStorage returns all storage caps (for save)
func (rm *ResourceManager) GetAllStorage() map[string]float64 {
	out := make(map[string]float64)
	for key, r := range rm.resources {
		out[key] = r.Storage
	}
	return out
}

// LoadAmounts restores resource amounts from save data
func (rm *ResourceManager) LoadAmounts(amounts map[string]float64) {
	for key, amount := range amounts {
		if r, ok := rm.resources[key]; ok {
			r.Amount = amount
		}
	}
}

// LoadStorage restores storage caps from save data
func (rm *ResourceManager) LoadStorage(storage map[string]float64) {
	for key, amount := range storage {
		if r, ok := rm.resources[key]; ok {
			r.Storage = amount
		}
	}
}

// Snapshot returns resource states for UI
func (rm *ResourceManager) Snapshot() map[string]ResourceState {
	out := make(map[string]ResourceState)
	for key, r := range rm.resources {
		def := rm.defs[key]
		out[key] = ResourceState{
			Amount:   r.Amount,
			Rate:     r.Rate,
			Storage:  r.Storage,
			Name:     def.Name,
			Unlocked: rm.unlocked[key],
		}
	}
	return out
}
