package config

// ResourceDef defines a resource type
type ResourceDef struct {
	Name        string
	Key         string
	BaseStorage float64
	Age         string // minimum age to unlock
	Description string
}

// BaseResources returns all resource definitions
// Base storage is intentionally low — players must build storage buildings to hold more
func BaseResources() []ResourceDef {
	return []ResourceDef{
		// Primitive Age
		{Name: "Food", Key: "food", BaseStorage: 50, Age: "primitive_age", Description: "Feeds your population"},
		{Name: "Wood", Key: "wood", BaseStorage: 50, Age: "primitive_age", Description: "Basic building material"},
		// Stone Age
		{Name: "Stone", Key: "stone", BaseStorage: 50, Age: "stone_age", Description: "Durable building material"},
		{Name: "Knowledge", Key: "knowledge", BaseStorage: 30, Age: "stone_age", Description: "Powers research"},
		// Bronze Age
		{Name: "Iron", Key: "iron", BaseStorage: 50, Age: "bronze_age", Description: "Metal for tools and weapons"},
		{Name: "Gold", Key: "gold", BaseStorage: 50, Age: "bronze_age", Description: "Currency and trade"},
		// Iron Age
		{Name: "Coal", Key: "coal", BaseStorage: 50, Age: "iron_age", Description: "Fuel for smelting and industry"},
		// Medieval Age
		{Name: "Steel", Key: "steel", BaseStorage: 30, Age: "medieval_age", Description: "Refined metal for advanced construction"},
		{Name: "Faith", Key: "faith", BaseStorage: 50, Age: "medieval_age", Description: "Spiritual influence"},
		// Renaissance Age
		{Name: "Culture", Key: "culture", BaseStorage: 50, Age: "renaissance_age", Description: "Art and cultural influence"},
		// Industrial Age
		{Name: "Oil", Key: "oil", BaseStorage: 50, Age: "industrial_age", Description: "Fuel for machines and industry"},
		// Victorian Age
		{Name: "Electricity", Key: "electricity", BaseStorage: 50, Age: "victorian_age", Description: "Powers modern infrastructure"},
		// Atomic Age
		{Name: "Uranium", Key: "uranium", BaseStorage: 30, Age: "atomic_age", Description: "Radioactive fuel for reactors"},
		// Modern Age
		{Name: "Data", Key: "data", BaseStorage: 50, Age: "modern_age", Description: "Digital information and analytics"},
		// Cyberpunk Age
		{Name: "Crypto", Key: "crypto", BaseStorage: 50, Age: "cyberpunk_age", Description: "Decentralized digital currency"},
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
