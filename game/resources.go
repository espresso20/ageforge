package game

import (
	"math"

	"github.com/espresso20/ageforge/config"
)

// Resource holds the runtime state of a single resource
type Resource struct {
	Amount    float64
	Rate      float64
	Storage   float64
	Breakdown RateBreakdown
}

// ResourceManager manages all resources
type ResourceManager struct {
	resources map[string]*Resource
	defs      map[string]config.ResourceDef
	unlocked  map[string]bool
	// order is every resource key, sorted, fixed at construction, so the
	// per-tick walks (and the overflow they report) never follow map order.
	order []string
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
	rm.order = sortedKeys(rm.resources)
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

// GetRate returns the current per-tick rate (production − consumption) for a
// resource, or 0 if it doesn't exist.
func (rm *ResourceManager) GetRate(key string) float64 {
	if r, ok := rm.resources[key]; ok {
		return r.Rate
	}
	return 0
}

// Add adds an amount to a resource, respecting storage limits. A NaN amount is
// ignored: it would otherwise stick to the resource forever, since the clamps
// below compare false against NaN.
func (rm *ResourceManager) Add(key string, amount float64) float64 {
	r, ok := rm.resources[key]
	if !ok {
		return 0
	}
	if math.IsNaN(amount) {
		return r.Amount
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

// Remove subtracts from a resource. Returns false if insufficient (a NaN
// amount is never sufficient: written as !(have >= need) so NaN fails it).
func (rm *ResourceManager) Remove(key string, amount float64) bool {
	r, ok := rm.resources[key]
	if !ok || !(r.Amount >= amount) {
		return false
	}
	r.Amount -= amount
	return true
}

// CanAfford checks if all costs can be paid
func (rm *ResourceManager) CanAfford(costs map[string]float64) bool {
	for key, amount := range costs {
		if !(rm.Get(key) >= amount) { // NaN costs are unaffordable
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
	rm.ApplyRatesCapped(nil)
}

// ApplyRatesCapped applies one tick of rates like ApplyRates, in key order,
// and calls lost (when non-nil) with what the storage cap cut off each
// resource that hit it.
func (rm *ResourceManager) ApplyRatesCapped(lost func(key string, amount float64)) {
	for _, key := range rm.order {
		r := rm.resources[key]
		if !rm.unlocked[key] || r.Rate == 0 {
			continue
		}
		want := r.Amount + r.Rate
		rm.Add(key, r.Rate)
		if lost != nil && r.Rate > 0 && want > r.Amount {
			lost(key, want-r.Amount)
		}
	}
}

// AddProduced credits scale ticks of every positive rate (offline catch-up),
// in key order, never past a cap: a resource at or over its cap gains
// nothing. gained and lost (either may be nil) receive what each resource
// took in and what its cap cut off.
func (rm *ResourceManager) AddProduced(scale float64, gained, lost func(key string, amount float64)) {
	for _, key := range rm.order {
		r := rm.resources[key]
		if !rm.unlocked[key] || r.Rate <= 0 {
			continue
		}
		amount := r.Rate * scale
		g := math.Min(amount, r.Storage-r.Amount)
		if g > 0 {
			r.Amount += g
			if gained != nil {
				gained(key, g)
			}
		} else {
			g = 0
		}
		if lost != nil && amount-g > 0 {
			lost(key, amount-g)
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
			Amount:    r.Amount,
			Rate:      r.Rate,
			Storage:   r.Storage,
			Name:      def.Name,
			Unlocked:  rm.unlocked[key],
			Breakdown: r.Breakdown,
		}
	}
	return out
}
