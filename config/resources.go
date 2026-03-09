package config

// ResourceDef defines a single game resource. There are 25 resources in total,
// unlocking progressively through the 22 ages.
//
// BaseStorage is the starting storage cap before any storage buildings are built.
// It is intentionally low — storage buildings are the primary progression gate
// (you can't accumulate enough resources to advance without building them first).
type ResourceDef struct {
	Name        string
	Key         string
	BaseStorage float64 // starting storage cap before storage buildings; units match production rates (per-tick)
	Age         string  // minimum age key at which this resource becomes unlocked
	Description string
}

// BaseResources returns all resource definitions
// Base storage is intentionally low — players must build storage buildings to hold more
func BaseResources() []ResourceDef {
	return []ResourceDef{
		// Primitive Age (Stone Era)
		{Name: "Food", Key: "food", BaseStorage: 50, Age: "primitive_age", Description: "Feeds your population"},
		{Name: "Wood", Key: "wood", BaseStorage: 50, Age: "primitive_age", Description: "Basic building material"},
		{Name: "Knowledge", Key: "knowledge", BaseStorage: 30, Age: "primitive_age", Description: "Powers research"},
		{Name: "Faith", Key: "faith", BaseStorage: 50, Age: "primitive_age", Description: "Spiritual influence — drains over time; neglect has consequences"},
		// Stone Age (Stone Era)
		{Name: "Stone", Key: "stone", BaseStorage: 50, Age: "stone_age", Description: "Durable building material"},
		// Bronze Age (Stone Era)
		{Name: "Iron", Key: "iron", BaseStorage: 50, Age: "bronze_age", Description: "Metal for tools and weapons"},
		{Name: "Gold", Key: "gold", BaseStorage: 50, Age: "bronze_age", Description: "Currency and trade"},
		// Iron Age (Iron Era)
		{Name: "Coal", Key: "coal", BaseStorage: 50, Age: "iron_age", Description: "Fuel for smelting and industry"},
		// Classical Age (Iron Era) — intermediate ore for Geological Extraction lineage
		{Name: "Marble", Key: "marble", BaseStorage: 30, Age: "classical_age", Description: "Refined stone for monumental construction"},
		{Name: "Iron Ore", Key: "iron_ore", BaseStorage: 30, Age: "classical_age", Description: "Raw iron ore before smelting — feeds Metallurgy lineage"},
		// Medieval Age (Iron Era)
		{Name: "Steel", Key: "steel", BaseStorage: 30, Age: "medieval_age", Description: "Refined metal for advanced construction"},
		// Classical Age (Iron Era)
		{Name: "Culture", Key: "culture", BaseStorage: 50, Age: "classical_age", Description: "Art and cultural influence — accumulates permanently, gates prestige bonuses"},
		// Industrial Age
		{Name: "Oil", Key: "oil", BaseStorage: 50, Age: "industrial_age", Description: "Fuel for machines and industry"},
		// Victorian Age
		{Name: "Electricity", Key: "electricity", BaseStorage: 50, Age: "victorian_age", Description: "Powers modern infrastructure"},
		// Atomic Age
		{Name: "Uranium", Key: "uranium", BaseStorage: 30, Age: "atomic_age", Description: "Radioactive fuel for reactors"},
		// Modern Age (Digital Era)
		{Name: "Data", Key: "data", BaseStorage: 50, Age: "modern_age", Description: "Digital information and analytics"},
		{Name: "Nanobots", Key: "nanobots", BaseStorage: 20, Age: "modern_age", Description: "Microscopic machines — Organic Extraction output in the Digital Era"},
		{Name: "Titanium Ore", Key: "titanium_ore", BaseStorage: 20, Age: "modern_age", Description: "Raw titanium ore — feeds Metallurgy lineage; smelts into titanium"},
		// Cyberpunk Age (Neon Era)
		{Name: "Crypto", Key: "crypto", BaseStorage: 50, Age: "cyberpunk_age", Description: "Decentralized digital currency"},
		{Name: "Dark Matter Crystals", Key: "dark_matter_crystals", BaseStorage: 10, Age: "cyberpunk_age", Description: "Crystallised dark matter — raw form; refined into dark matter by Metallurgy lineage"},
		// Fusion Age
		{Name: "Plasma", Key: "plasma", BaseStorage: 30, Age: "fusion_age", Description: "Superheated ionized gas for energy"},
		// Space Age
		{Name: "Titanium", Key: "titanium", BaseStorage: 30, Age: "space_age", Description: "Lightweight metal for space construction"},
		// Interstellar Age
		{Name: "Dark Matter", Key: "dark_matter", BaseStorage: 20, Age: "interstellar_age", Description: "Exotic matter for warp technology"},
		// Galactic Age
		{Name: "Antimatter", Key: "antimatter", BaseStorage: 20, Age: "galactic_age", Description: "Annihilation fuel for megastructures"},
		// Quantum Age
		{Name: "Quantum Flux", Key: "quantum_flux", BaseStorage: 10, Age: "quantum_age", Description: "Unstable quantum energy for reality manipulation"},
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
