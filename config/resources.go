package config

import "math"

// ResourceDef defines a single game resource. There are 24 resources in total,
// unlocking progressively through the 22 ages.
//
// BaseStorage is the starting storage cap before any storage buildings are built.
// It is intentionally low — storage buildings are the primary progression gate
// (you can't accumulate enough resources to advance without building them first).
type ResourceDef struct {
	Name        string
	Key         string
	BaseStorage float64 // starting storage cap before storage buildings; units match production rates (per-tick)
	Age         string  // the age whose UnlockResources list unlocks it (TestResourceAgesMatchAgeUnlocks)
	// BuiltUnlocks: the first standing building that makes the resource
	// unlocks it, whatever the age. Soldiers are the one: a War Camp can be
	// built two ages before the age that brings soldiers to every town, and
	// it trains them from the day it stands. Without one, the resource
	// waits for its Age.
	BuiltUnlocks bool
}

// WoodBaseStorage is the wood a new game stores: the 50 it was authored at times
// NewCopyPriceFactor, rounded up to a multiple of five. The first Stash is the
// only way to raise a store, so the first copy has to fit under it with room to
// spare (TestStash_FirstCopyAffordableWithinBaseWoodCap), and it costs the
// factor more than it did.
var WoodBaseStorage = math.Ceil(50*NewCopyPriceFactor/5) * 5

// BaseResources returns all resource definitions
// Base storage is intentionally low — players must build storage buildings to hold more
// The slice order is load-bearing (boon pools and a few UI lists walk it), so a
// resource whose unlock age moved keeps its place and just carries the new Age.
func BaseResources() []ResourceDef {
	return []ResourceDef{
		// Primitive Age (Stone Era)
		{Name: "Food", Key: "food", BaseStorage: 50, Age: "primitive_age"},
		{Name: "Wood", Key: "wood", BaseStorage: WoodBaseStorage, Age: "primitive_age"},
		{Name: "Faith", Key: "faith", BaseStorage: 50, Age: "primitive_age"},
		// Stone Age (Stone Era)
		{Name: "Stone", Key: "stone", BaseStorage: 50, Age: "stone_age"},
		// Bronze Age (Stone Era)
		{Name: "Iron", Key: "iron", BaseStorage: 50, Age: "bronze_age"},
		{Name: "Gold", Key: "gold", BaseStorage: 50, Age: "bronze_age"},
		// Iron Age (Iron Era); coal unlocks in the Renaissance Age (Steel Era)
		{Name: "Coal", Key: "coal", BaseStorage: 50, Age: "renaissance_age"},
		// Iron Age (Iron Era): intermediate ores for Geological Extraction and Metallurgy
		{Name: "Marble", Key: "marble", BaseStorage: 30, Age: "iron_age"},
		{Name: "Iron Ore", Key: "iron_ore", BaseStorage: 30, Age: "iron_age"},
		// Medieval Age (Iron Era)
		{Name: "Steel", Key: "steel", BaseStorage: 30, Age: "medieval_age"},
		// Classical Age (Iron Era)
		{Name: "Culture", Key: "culture", BaseStorage: 50, Age: "classical_age"},
		// Industrial Age
		{Name: "Oil", Key: "oil", BaseStorage: 50, Age: "industrial_age"},
		// Victorian Age
		{Name: "Electricity", Key: "electricity", BaseStorage: 50, Age: "victorian_age"},
		// Atomic Age
		{Name: "Uranium", Key: "uranium", BaseStorage: 30, Age: "atomic_age"},
		// Modern Age (Digital Era); titanium ore unlocks in the Space Age
		{Name: "Data", Key: "data", BaseStorage: 50, Age: "modern_age"},
		{Name: "Nanobots", Key: "nanobots", BaseStorage: 20, Age: "modern_age"},
		{Name: "Titanium Ore", Key: "titanium_ore", BaseStorage: 20, Age: "modern_age"},
		// Cyberpunk Age (Neon Era)
		{Name: "Crypto", Key: "crypto", BaseStorage: 50, Age: "cyberpunk_age"},
		{Name: "Dark Matter Crystals", Key: "dark_matter_crystals", BaseStorage: 10, Age: "cyberpunk_age"},
		// Fusion Age
		{Name: "Plasma", Key: "plasma", BaseStorage: 30, Age: "fusion_age"},
		// Space Age
		{Name: "Titanium", Key: "titanium", BaseStorage: 30, Age: "space_age"},
		// Interstellar Age
		{Name: "Dark Matter", Key: "dark_matter", BaseStorage: 20, Age: "interstellar_age"},
		// Galactic Age
		{Name: "Antimatter", Key: "antimatter", BaseStorage: 20, Age: "galactic_age"},
		// Quantum Age
		{Name: "Quantum Flux", Key: "quantum_flux", BaseStorage: 10, Age: "quantum_age"},
	}
}

// ResourceByKey returns a map of key -> ResourceDef
func ResourceByKey() map[string]ResourceDef {
	m := make(map[string]ResourceDef)
	for _, r := range BaseResources() {
		m[r.Key] = r
	}
	return m
}
