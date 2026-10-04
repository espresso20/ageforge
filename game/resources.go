package game

import (
	"math"

	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/rules"
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
	rules     *rules.Set
	resources map[string]*Resource
	defs      map[string]config.ResourceDef
	unlocked  map[string]bool
	// order is every resource key, sorted, fixed at construction, so the
	// per-tick walks (and the overflow they report) never follow map order.
	order []string
	// grace is the resources whose stock may sit above the cap (the grace
	// rule, mastery.go): Add never cuts them down to it, and never adds
	// while they are over. recalculateRates drops a resource from it once
	// its stock is at or under the cap. nil when none.
	grace map[string]bool
}

// NewResourceManager creates a resource manager on the core ruleset.
func NewResourceManager() *ResourceManager { return NewResourceManagerWith(rules.Core()) }

// NewResourceManagerWith creates a resource manager with set's resources.
func NewResourceManagerWith(set *rules.Set) *ResourceManager {
	rm := &ResourceManager{
		resources: make(map[string]*Resource),
		unlocked:  make(map[string]bool),
	}
	rm.Rebind(set)
	return rm
}

// Rebind moves the manager onto set: it takes set's resource definitions
// and gives every resource it did not hold yet an empty store. What it
// already holds stays, a resource set no longer defines included.
func (rm *ResourceManager) Rebind(set *rules.Set) {
	rm.rules = set
	rm.defs = set.ResourceMap()
	for _, def := range set.Resources() {
		if _, held := rm.resources[def.Key]; held {
			continue
		}
		rm.resources[def.Key] = &Resource{
			Amount:  0,
			Rate:    0,
			Storage: def.BaseStorage,
		}
	}
	rm.order = sortedKeys(rm.resources)
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
	old := r.Amount
	r.Amount += float64(amount) // callers pass products: round them, no FMA
	if r.Amount > r.Storage {
		if rm.grace[key] && old > r.Storage {
			// Graced stock above the cap: spending lowers it, nothing raises it.
			r.Amount = math.Min(old, r.Amount)
		} else {
			r.Amount = r.Storage
		}
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
	r.Amount -= float64(amount) // as in Add
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
		amount := float64(r.Rate * scale)
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

// markGraceAll marks every resource holding stock for the grace rule; the
// storage pass that follows keeps the mark only where stock is over the cap.
func (rm *ResourceManager) markGraceAll() {
	for _, key := range rm.order {
		if rm.resources[key].Amount > 0 {
			if rm.grace == nil {
				rm.grace = make(map[string]bool)
			}
			rm.grace[key] = true
		}
	}
}

// Graced reports whether key's stock is under the grace rule.
func (rm *ResourceManager) Graced(key string) bool { return rm.grace[key] }

// graceSave is the graced set for a save (nil when empty).
func (rm *ResourceManager) graceSave() map[string]bool {
	if len(rm.grace) == 0 {
		return nil
	}
	out := make(map[string]bool, len(rm.grace))
	for k := range rm.grace {
		out[k] = true
	}
	return out
}

// loadGrace restores the graced set from a save (known resources only).
func (rm *ResourceManager) loadGrace(g map[string]bool) {
	rm.grace = nil
	for k, on := range g {
		if _, ok := rm.resources[k]; ok && on {
			if rm.grace == nil {
				rm.grace = make(map[string]bool)
			}
			rm.grace[k] = true
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

			OverCapGrace: rm.grace[key],
		}
	}
	return out
}
