package config

import "math"

// Effect represents a single game effect applied by a building, tech, or milestone.
// The semantics of Value depend on Type:
//   - "production":     +Value of Target resource per tick
//   - "storage":        +Value to the global storage cap for Target resource (or "all")
//   - "capacity":       +Value to Target (e.g. "population") cap
//   - "bonus":          multiplier bonus — Value is a fraction (0.1 = +10%)
//   - "unlock":         unlocks the Target building/resource key
//   - "instant_resource": immediately add Value of Target resource (milestone reward only)
//   - "permanent_bonus": persistent multiplier on Target rate (milestone reward only)
//   - "morale":          +Value to the civilization's morale per tick while the
//     building stands (FLAT, not worker-scaled; no Target)
type Effect struct {
	Type   string  // see above for valid types
	Target string  // resource key, building key, or special target (e.g. "all", "production_all")
	Value  float64 // meaning depends on Type; see above
}

// BuildingDef defines a single building type. All 284 buildings are one of:
// production (13 lineages), housing (pops), storage, or wonder (max 1 each).
//
// Cost scaling: actual cost of the N-th instance = BaseCost × CostScale^(N-1).
// For wonders CostScale is always 1.0 (no scaling, only 1 can be built).
type BuildingDef struct {
	Name         string
	Key          string
	Category     string             // "production", "housing", "research", "military", "storage", "wonder"
	BaseCost     map[string]float64 // resource costs for the first instance
	CostScale    float64            // exponential scale per additional instance; 1.0 = flat cost
	Effects      []Effect
	BuildTicks   int    // construction time in game ticks; 0 = instant (legacy only)
	RequiredAge  string // minimum age key the player must be in to build
	RequiredTech string // tech key that must be researched first; "" = no requirement
	MaxCount     int    // maximum instances allowed; 0 = unlimited
	Description  string
	// Lineage / economy metadata (added in Phase 5+)
	LineageKey     string // which of the 13 production lineages (e.g. "food", "metallurgy", "wonder")
	LineageTier    int    // 0-indexed position within the lineage; higher tier = later age
	WorkerDomain   string // worker domain key for assignment (e.g. "food", "knowledge"); "" = no workers
	WorkerCapacity int    // max workers assignable per individual building instance; 0 = not applicable
	EpochKey       string // epoch this building belongs to (e.g. "stone_era", "iron_era")
	OutputResource string // primary resource produced; used by the engine to remap lineage output on age transition
}

// baseBuildingsRaw returns only the storage and wonder BuildingDef entries,
// without economy-redesign metadata (lineage, worker domain, epoch key).
// These definitions are merged with buildingMeta() in BaseBuildings().
// Production/housing/military buildings are now defined in the 13 lineage files
// accessed via NewProductionBuildings() and should not be added here.
//
// Cost progression guideline per age:
//
//	Primitive 30–100 → Stone 200–1k → Bronze 1.5k–5k → Iron 8k–25k →
//	Classical 40k–120k → Medieval 200k–600k → Renaissance 1M–3M →
//	Colonial 5M–15M → Industrial 25M–75M → Victorian 125M–375M → ...
func baseBuildingsRaw() []BuildingDef {
	return []BuildingDef{
		// ===== PRIMITIVE AGE (costs: 30-100) =====
		{
			Name: "Stash", Key: "stash", Category: "storage",
			// note: the first copy MUST stay affordable within wood's 50 base storage
			// cap — stash is the only building that raises that cap, so a first-copy
			// cost above 50 is an unbuildable deadlock. normalizeCostCurves multiplies
			// this base by ~1.17 (storage 1.15->1.13, pivot@10), so keep the raw base
			// well under ~42. 30 -> ~35 normalized, a comfortable margin under 50.
			BaseCost:    map[string]float64{"wood": 30},
			CostScale:   1.15,
			MaxCount:    50,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 300}},
			BuildTicks:  10,
			RequiredAge: "primitive_age",
			Description: "A hidden pile of supplies. +300 storage.",
		},

		// ===== STONE AGE (costs: 200-1000) =====
		{
			Name: "Storage Pit", Key: "storage_pit", Category: "storage",
			BaseCost:    map[string]float64{"wood": 300, "stone": 200},
			CostScale:   1.2,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 500}},
			BuildTicks:  60,
			RequiredAge: "stone_age",
			Description: "A hole in the ground to stash things. +500 storage.",
		},

		// ===== BRONZE AGE (costs: 1500-5000) =====
		{
			Name: "Warehouse", Key: "warehouse", Category: "storage",
			BaseCost:    map[string]float64{"wood": 2000, "stone": 1500, "iron": 300},
			CostScale:   1.2,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 3000}},
			BuildTicks:  80,
			RequiredAge: "bronze_age",
			Description: "Proper storage building. +3000 storage.",
		},

		// ===== IRON AGE (costs: 8k-25k) =====
		{
			Name: "Granary", Key: "granary", Category: "storage",
			BaseCost:    map[string]float64{"wood": 8000, "stone": 6000},
			CostScale:   1.2,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 12000}},
			BuildTicks:  120,
			RequiredAge: "iron_age",
			Description: "Organized supply storage. +12000 storage.",
		},

		// ===== CLASSICAL AGE (costs: 40k-120k) =====
		{
			Name: "Classical Vault", Key: "classical_vault", Category: "storage",
			BaseCost:    map[string]float64{"stone": 50000, "iron": 12000, "gold": 10000},
			CostScale:   1.2,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 25000}},
			BuildTicks:  150,
			RequiredAge: "classical_age",
			Description: "Stone vault for valuables. +25000 storage.",
		},

		// ===== MEDIEVAL AGE (costs: 200k-600k) =====
		{
			Name: "Keep", Key: "keep", Category: "storage",
			BaseCost:    map[string]float64{"stone": 200000, "iron": 60000, "gold": 40000},
			CostScale:   1.2,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 60000}},
			BuildTicks:  200,
			RequiredAge: "medieval_age",
			Description: "Fortified storehouse. +60000 storage.",
		},

		// ===== RENAISSANCE AGE (costs: 1M-3M) =====
		{
			Name: "Renaissance Vault", Key: "renaissance_vault", Category: "storage",
			BaseCost:    map[string]float64{"stone": 250000, "gold": 150000, "iron": 60000},
			CostScale:   1.2,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 500000}},
			BuildTicks:  250,
			RequiredAge: "renaissance_age",
			Description: "Ornate storage facility. +500000 storage.",
		},

		// ===== COLONIAL AGE (costs: 5M-15M) =====
		{
			Name: "Colonial Warehouse", Key: "colonial_warehouse", Category: "storage",
			BaseCost:    map[string]float64{"wood": 1.5e6, "stone": 1e6, "gold": 600000},
			CostScale:   1.2,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 10e6}},
			BuildTicks:  300,
			RequiredAge: "colonial_age",
			Description: "Trade goods warehouse. +10M storage.",
		},

		// ===== INDUSTRIAL AGE (costs: 25M-75M) =====
		{
			Name: "Industrial Depot", Key: "industrial_depot", Category: "storage",
			BaseCost:    map[string]float64{"steel": 15e6, "iron": 20e6, "coal": 10e6},
			CostScale:   1.2,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 50e6}},
			BuildTicks:  400,
			RequiredAge: "industrial_age",
			Description: "Industrial-scale storage. +50M storage.",
		},

		// ===== VICTORIAN AGE (costs: 125M-375M) =====
		{
			Name: "Victorian Vault", Key: "victorian_vault", Category: "storage",
			BaseCost:    map[string]float64{"steel": 125e6, "gold": 100e6, "iron": 75e6},
			CostScale:   1.2,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 350e6}},
			BuildTicks:  500,
			RequiredAge: "victorian_age",
			Description: "Reinforced vault. +350M storage.",
		},

		// ===== ELECTRIC AGE (costs: 600M-2B) =====
		{
			Name: "Electric Warehouse", Key: "electric_warehouse", Category: "storage",
			BaseCost:    map[string]float64{"steel": 750e6, "electricity": 125e6, "iron": 500e6},
			CostScale:   1.2,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 3.5e9}},
			BuildTicks:  600,
			RequiredAge: "electric_age",
			Description: "Climate-controlled storage. +3.5B storage.",
		},

		// ===== ATOMIC AGE (costs: 3B-10B) =====
		{
			Name: "Atomic Vault", Key: "atomic_vault", Category: "storage",
			BaseCost:    map[string]float64{"steel": 5e9, "stone": 7.5e9, "iron": 3e9},
			CostScale:   1.25,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 15e9}},
			BuildTicks:  700,
			RequiredAge: "atomic_age",
			Description: "Radiation-shielded storage. +15B storage.",
		},

		// ===== MODERN AGE (costs: 15B-50B) =====
		{
			Name: "Modern Depot", Key: "modern_depot", Category: "storage",
			BaseCost:    map[string]float64{"steel": 35e9, "gold": 25e9, "electricity": 8e9},
			CostScale:   1.25,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 45e9}},
			BuildTicks:  800,
			RequiredAge: "modern_age",
			Description: "Automated logistics center. +45B storage.",
		},

		// ===== INFORMATION AGE (costs: 75B-250B) =====
		{
			Name: "Info Vault", Key: "info_vault", Category: "storage",
			BaseCost:    map[string]float64{"steel": 100e9, "electricity": 40e9, "data": 625e6},
			CostScale:   1.25,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 250e9}},
			BuildTicks:  900,
			RequiredAge: "information_age",
			Description: "Digital-physical storage hybrid. +250B storage.",
		},

		// ===== DIGITAL AGE (costs: 400B-1.2T) =====
		{
			Name: "Digital Archive", Key: "digital_archive", Category: "storage",
			BaseCost:    map[string]float64{"steel": 500e9, "data": 10e9, "electricity": 150e9},
			CostScale:   1.25,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 1.5e12}},
			BuildTicks:  1000,
			RequiredAge: "digital_age",
			Description: "Quantum-encrypted storage. +1.5T storage.",
		},

		// ===== CYBERPUNK AGE (costs: 2T-6T) =====
		{
			Name: "Cyber Vault", Key: "cyber_vault", Category: "storage",
			BaseCost:    map[string]float64{"steel": 3e12, "data": 60e9, "crypto": 10e9},
			CostScale:   1.25,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 5e12}},
			BuildTicks:  1400,
			RequiredAge: "cyberpunk_age",
			Description: "Encrypted digital vault. +5T storage.",
		},

		// ===== FUSION AGE (costs: 10T-30T) =====
		{
			Name: "Fusion Vault", Key: "fusion_vault", Category: "storage",
			BaseCost:    map[string]float64{"steel": 10e12, "plasma": 500e9, "electricity": 5e12},
			CostScale:   1.25,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 30e12}},
			BuildTicks:  2000,
			RequiredAge: "fusion_age",
			Description: "Plasma-shielded storage. +30T storage.",
		},

		// ===== SPACE AGE (costs: 50T-150T) =====
		{
			Name: "Orbital Depot", Key: "orbital_depot", Category: "storage",
			BaseCost:    map[string]float64{"steel": 50e12, "plasma": 6e12, "electricity": 30e12},
			CostScale:   1.3,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 200e12}},
			BuildTicks:  4000,
			RequiredAge: "space_age",
			Description: "Zero-gravity storage facility. +200T storage.",
		},

		// ===== INTERSTELLAR AGE (costs: 250T-750T) =====
		{
			Name: "Stellar Vault", Key: "stellar_vault", Category: "storage",
			BaseCost:    map[string]float64{"titanium": 60e12, "plasma": 50e12, "electricity": 80e12},
			CostScale:   1.3,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 500e12}},
			BuildTicks:  6000,
			RequiredAge: "interstellar_age",
			Description: "Pocket-dimension storage. +500T storage.",
		},

		// ===== GALACTIC AGE (costs: 1.25Q-3.75Q) =====
		{
			Name: "Galactic Vault", Key: "galactic_vault", Category: "storage",
			BaseCost:    map[string]float64{"dark_matter": 100e12, "titanium": 500e12, "plasma": 200e12},
			CostScale:   1.3,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 2e15}},
			BuildTicks:  10000,
			RequiredAge: "galactic_age",
			Description: "Galaxy-spanning storage network. +2Q storage.",
		},

		// ===== QUANTUM AGE (costs: 6Q-20Q) =====
		{
			Name: "Quantum Vault", Key: "quantum_vault", Category: "storage",
			BaseCost:    map[string]float64{"antimatter": 1e15, "dark_matter": 2e15, "titanium": 500e12},
			CostScale:   1.35,
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 5e15}},
			BuildTicks:  12000,
			RequiredAge: "quantum_age",
			Description: "Stores matter in quantum superposition. +5Q storage.",
		},

		// ===== TRANSCENDENT AGE =====
		// (singularity_core is a wonder, listed below)

		// ===== WONDERS =====
		// There is exactly one wonder per age (22 wonders total). Building a wonder
		// unlocks +0.5x game speed — the primary long-term speed upgrade mechanic.
		// Wonder costs use WonderBank: resources must be "banked" via 'wonder collect'
		// before 'build <key>' queues construction. CostScale is always 1.0.
		// Build ticks are extremely long to make each wonder a meaningful milestone.

		// Primitive Age — normal costs: 30-300
		{
			Name: "Sacred Grove", Key: "sacred_grove", Category: "wonder",
			BaseCost:  map[string]float64{"wood": 8000, "food": 5000},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 0.02},
				{Type: "production", Target: "food", Value: 0.05},
			},
			RequiredAge: "primitive_age",
			MaxCount:    1,
			BuildTicks:  300,
			Description: "An ancient clearing where nature's power flows. +0.02 knowledge, +0.05 food/tick. Unlocks +0.5x speed.",
		},
		// Stone Age — normal costs: 200-1000
		{
			Name: "Great Monolith", Key: "great_monolith", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 25000, "wood": 20000, "food": 10000},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 0.05},
				{Type: "storage", Target: "all", Value: 5000},
			},
			RequiredAge: "stone_age",
			MaxCount:    1,
			BuildTicks:  800,
			Description: "A towering stone pillar visible for miles. +0.05 knowledge/tick, +5000 storage. Unlocks +0.5x speed.",
		},
		// Bronze Age — normal costs: 1500-2500
		{
			Name: "Stonehenge", Key: "stonehenge", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 80000, "wood": 45000, "iron": 8000},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 0.8},
				{Type: "production", Target: "faith", Value: 0.6},
			},
			RequiredAge: "bronze_age",
			MaxCount:    1,
			BuildTicks:  1200,
			Description: "Massive stone circle aligned to the cosmos. +0.8 knowledge, +0.6 faith/tick. Unlocks +0.5x speed.",
		},
		// Iron Age — normal costs: 8k-12k
		{
			Name: "Colosseum", Key: "colosseum", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 400000, "iron": 90000, "gold": 80000},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "capacity", Target: "population", Value: 100},
				{Type: "production", Target: "culture", Value: 2.0},
			},
			RequiredAge: "iron_age",
			MaxCount:    1,
			BuildTicks:  2000,
			Description: "Grand arena of blood and glory. +100 pop cap, +2.0 culture/tick. Unlocks +0.5x speed.",
		},
		// Classical Age — normal costs: 40k-80k
		{
			Name: "Parthenon", Key: "parthenon", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 1800000, "gold": 800000, "iron": 800000},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 2.0},
				{Type: "production", Target: "knowledge", Value: 1.2},
			},
			RequiredAge: "classical_age",
			MaxCount:    1,
			BuildTicks:  2500,
			Description: "Perfect temple of marble and wisdom. +2.0 culture, +1.2 knowledge/tick. Unlocks +0.5x speed.",
		},
		// Medieval Age — normal costs: 180k-360k
		{
			Name: "Great Library", Key: "great_library", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 8000000, "gold": 6000000, "knowledge": 1900000},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 2.0},
				{Type: "bonus", Target: "knowledge_rate", Value: 0.3},
			},
			RequiredAge: "medieval_age",
			MaxCount:    1,
			BuildTicks:  3600,
			Description: "Repository of all knowledge. +2.0 knowledge/tick, +30% knowledge rate. Unlocks +0.5x speed.",
		},
		// Renaissance Age — normal costs: 400k-600k
		{
			Name: "Sistine Chapel", Key: "sistine_chapel", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 9900000, "gold": 7000000, "faith": 6000000, "culture": 8000000},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 3.5},
				{Type: "production", Target: "faith", Value: 1.8},
			},
			RequiredAge: "renaissance_age",
			MaxCount:    1,
			BuildTicks:  5200,
			Description: "Ceiling painted by divine hands. +3.5 culture, +1.8 faith/tick. Unlocks +0.5x speed.",
		},
		// Colonial Age — normal costs: 1.2M-2M
		{
			Name: "Grand Lighthouse", Key: "grand_lighthouse", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 40e6, "gold": 30e6, "steel": 6e6},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 5.0},
				{Type: "bonus", Target: "expedition_reward", Value: 0.8},
			},
			RequiredAge: "colonial_age",
			MaxCount:    1,
			BuildTicks:  6200,
			Description: "Beacon visible across oceans. +5.0 gold/tick, +80% expedition rewards. Unlocks +0.5x speed.",
		},
		// Industrial Age — normal costs: 12M-25M
		{
			Name: "Crystal Palace", Key: "crystal_palace", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 800e6, "iron": 700e6, "gold": 550e6, "coal": 700e6},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.15},
				{Type: "production", Target: "gold", Value: 8.0},
			},
			RequiredAge: "industrial_age",
			MaxCount:    1,
			BuildTicks:  9000,
			Description: "Glass cathedral of industry. +15% all production, +8.0 gold/tick. Unlocks +0.5x speed.",
		},
		// Victorian Age — normal costs: 90M-150M
		{
			Name: "Eiffel Tower", Key: "eiffel_tower", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 5.5e9, "iron": 4.5e9, "gold": 6e9},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 5.0},
				{Type: "production", Target: "knowledge", Value: 2.0},
			},
			RequiredAge: "victorian_age",
			MaxCount:    1,
			BuildTicks:  14000,
			Description: "Iron monument piercing the sky. +5.0 culture, +2.0 knowledge/tick. Unlocks +0.5x speed.",
		},
		// Electric Age — normal costs: 500M-1B
		{
			Name: "Hoover Dam", Key: "hoover_dam", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 25e9, "stone": 50e9, "electricity": 7e9},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 10.0},
				{Type: "bonus", Target: "production_all", Value: 0.2},
			},
			RequiredAge: "electric_age",
			MaxCount:    1,
			BuildTicks:  24000,
			Description: "Taming a river to power a nation. +10.0 electricity/tick, +20% all production. Unlocks +0.5x speed.",
		},
		// Atomic Age — normal costs: 3B-10B
		{
			Name: "Particle Accelerator", Key: "particle_accelerator", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 450e9, "electricity": 600e9, "uranium": 70e9},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 10.0},
				{Type: "production", Target: "uranium", Value: 1.5},
			},
			RequiredAge: "atomic_age",
			MaxCount:    1,
			BuildTicks:  32000,
			Description: "Smashes atoms for science. +10 knowledge, +1.5 uranium/tick. Unlocks +0.5x speed.",
		},
		// Modern Age — normal costs: 15B-40B
		{
			Name: "Space Program", Key: "space_program", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 900e9, "gold": 800e9, "electricity": 500e9, "knowledge": 600e9},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 6.0},
				{Type: "production", Target: "culture", Value: 8.0},
			},
			RequiredAge: "modern_age",
			MaxCount:    1,
			BuildTicks:  46000,
			Description: "Reaching for the stars. +6 knowledge, +8 culture/tick. Unlocks +0.5x speed.",
		},
		// Information Age — normal costs: 75B-125B
		{
			Name: "Global Network", Key: "global_network", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 5e12, "data": 600e9, "electricity": 990e9, "gold": 2.5e12},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "data", Value: 30.0},
				{Type: "bonus", Target: "knowledge_rate", Value: 0.3},
			},
			RequiredAge: "information_age",
			MaxCount:    1,
			BuildTicks:  63000,
			Description: "Every mind connected. +30.0 data/tick, +30% knowledge rate. Unlocks +0.5x speed.",
		},
		// Digital Age — normal costs: 400B-750B
		{
			Name: "World Simulation", Key: "world_simulation", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 50e12, "data": 900e9, "electricity": 6e12},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "data", Value: 60.0},
				{Type: "production", Target: "knowledge", Value: 15.0},
			},
			RequiredAge: "digital_age",
			MaxCount:    1,
			BuildTicks:  120000,
			Description: "A digital twin of reality itself. +60 data, +15 knowledge/tick. Unlocks +0.5x speed.",
		},
		// Cyberpunk Age — normal costs: 2T-4T
		{
			Name: "Neon Citadel", Key: "neon_citadel", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 80e12, "electricity": 60e12, "crypto": 9e12, "data": 5e12},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "crypto", Value: 10.0},
				{Type: "capacity", Target: "population", Value: 500},
			},
			RequiredAge: "cyberpunk_age",
			MaxCount:    1,
			BuildTicks:  360000,
			Description: "A city within a city, lit by eternal neon. +10 crypto/tick, +500 pop cap. Unlocks +0.5x speed.",
		},
		// Fusion Age — normal costs: 10T-15T
		{
			Name: "Stellar Cradle", Key: "stellar_cradle", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 750e12, "plasma": 600e12, "electricity": 800e12, "uranium": 940e12},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "plasma", Value: 15.0},
				{Type: "production", Target: "electricity", Value: 200.0},
			},
			RequiredAge: "fusion_age",
			MaxCount:    1,
			BuildTicks:  450000,
			Description: "A miniature star harnessed for power. +15 plasma, +200 electricity/tick. Unlocks +0.5x speed.",
		},
		// Space Age — normal costs: 50T-80T
		{
			Name: "Dyson Scaffold", Key: "dyson_scaffold", Category: "wonder",
			BaseCost:  map[string]float64{"titanium": 900e12, "plasma": 650e12, "steel": 5500e12},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 200.0},
				{Type: "production", Target: "plasma", Value: 30.0},
			},
			RequiredAge: "space_age",
			MaxCount:    1,
			BuildTicks:  640000,
			Description: "Framework for a Dyson sphere. +200 electricity, +30 plasma/tick. Unlocks +0.5x speed.",
		},
		// Interstellar Age — normal costs: 250T-500T
		{
			Name: "Warp Nexus", Key: "warp_nexus", Category: "wonder",
			BaseCost:  map[string]float64{"titanium": 7e15, "dark_matter": 500e12, "plasma": 6.5e15},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "dark_matter", Value: 8.0},
				{Type: "bonus", Target: "production_all", Value: 0.8},
			},
			RequiredAge: "interstellar_age",
			MaxCount:    1,
			BuildTicks:  860000,
			Description: "Hub of faster-than-light corridors. +8 dark matter/tick, +80% all production. Unlocks +0.5x speed.",
		},
		// Galactic Age — normal costs: 1Q-1.5Q
		{
			Name: "Cosmic Beacon", Key: "cosmic_beacon", Category: "wonder",
			BaseCost:  map[string]float64{"dark_matter": 8e15, "antimatter": 6e15, "titanium": 23e15},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "antimatter", Value: 10.0},
				{Type: "bonus", Target: "production_all", Value: 0.5},
			},
			RequiredAge: "galactic_age",
			MaxCount:    1,
			BuildTicks:  1860000,
			Description: "A signal fire across the galaxy. +10 antimatter/tick, +50% all production. Unlocks +0.5x speed.",
		},
		// Quantum Age — normal costs: 2.5Q-5Q
		{
			Name: "Reality Anchor", Key: "reality_anchor", Category: "wonder",
			BaseCost:  map[string]float64{"quantum_flux": 50e15, "antimatter": 80e15, "dark_matter": 90e15},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "production", Target: "quantum_flux", Value: 15.0},
				{Type: "bonus", Target: "production_all", Value: 0.5},
			},
			RequiredAge: "quantum_age",
			MaxCount:    1,
			BuildTicks:  2500000,
			Description: "Stabilizes reality across dimensions. +15 quantum flux/tick, +50% all production. Unlocks +0.5x speed.",
		},
		// Transcendent Age
		{
			Name: "Singularity Core", Key: "singularity_core", Category: "wonder",
			BaseCost:  map[string]float64{"quantum_flux": 890e15, "antimatter": 900e15, "dark_matter": 900e15},
			CostScale: 1.0,
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 2.0},
				{Type: "production", Target: "quantum_flux", Value: 20.0},
			},
			RequiredAge: "transcendent_age",
			MaxCount:    1,
			BuildTicks:  8675309,
			Description: "The final wonder. +200% all production, +20 quantum flux/tick. Unlocks +0.5x speed.",
		},
	}
}

// buildingMetaEntry holds the economy-redesign metadata for a single building.
// It supplements the legacy BuildingDef entries in baseBuildingsRaw() which
// were defined before lineage/epoch metadata existed.
type buildingMetaEntry struct {
	LineageKey     string
	LineageTier    int
	WorkerDomain   string
	WorkerCapacity int
	EpochKey       string
	OutputResource string
}

// buildingMeta returns the economy-redesign metadata for all buildings, keyed by building key.
// Lineages: housing, storage, food, organic_extraction, geological_extraction, knowledge,
//
//	faith, culture_arts, trade, military, engineering, metallurgy, energy, hacker,
//	astronaut, wonder
//
// Worker domains: food, lumber, masonry, knowledge, faith, military, trade, engineering,
//
//	metallurgy, energy, hacker, astronaut  (culture_arts has no domain)
//
// Epoch keys: stone_era, iron_era, steel_era, electric_era, digital_era, neon_era, cosmic_era
func buildingMeta() map[string]buildingMetaEntry {
	return map[string]buildingMetaEntry{
		// ── HOUSING lineage (no workers, no output resource) ──────────────────────
		"hut":             {LineageKey: "housing", LineageTier: 0, EpochKey: "stone_era"},
		"house":           {LineageKey: "housing", LineageTier: 1, EpochKey: "stone_era"},
		"manor":           {LineageKey: "housing", LineageTier: 2, EpochKey: "iron_era"},
		"apartment":       {LineageKey: "housing", LineageTier: 3, EpochKey: "steel_era"},
		"skyscraper":      {LineageKey: "housing", LineageTier: 4, EpochKey: "digital_era"},
		"neon_tower":      {LineageKey: "housing", LineageTier: 5, EpochKey: "neon_era"},
		"orbital_habitat": {LineageKey: "housing", LineageTier: 6, EpochKey: "neon_era"},

		// ── STORAGE lineage (no workers, no output resource) ─────────────────────
		"stash":              {LineageKey: "storage", LineageTier: 0, EpochKey: "stone_era"},
		"storage_pit":        {LineageKey: "storage", LineageTier: 1, EpochKey: "stone_era"},
		"warehouse":          {LineageKey: "storage", LineageTier: 2, EpochKey: "stone_era"},
		"granary":            {LineageKey: "storage", LineageTier: 3, EpochKey: "iron_era"},
		"keep":               {LineageKey: "storage", LineageTier: 4, EpochKey: "iron_era"},
		"classical_vault":    {LineageKey: "storage", LineageTier: 5, EpochKey: "iron_era"},
		"renaissance_vault":  {LineageKey: "storage", LineageTier: 6, EpochKey: "steel_era"},
		"colonial_warehouse": {LineageKey: "storage", LineageTier: 7, EpochKey: "steel_era"},
		"industrial_depot":   {LineageKey: "storage", LineageTier: 8, EpochKey: "steel_era"},
		"victorian_vault":    {LineageKey: "storage", LineageTier: 9, EpochKey: "electric_era"},
		"electric_warehouse": {LineageKey: "storage", LineageTier: 10, EpochKey: "electric_era"},
		"atomic_vault":       {LineageKey: "storage", LineageTier: 11, EpochKey: "electric_era"},
		"modern_depot":       {LineageKey: "storage", LineageTier: 12, EpochKey: "digital_era"},
		"info_vault":         {LineageKey: "storage", LineageTier: 13, EpochKey: "digital_era"},
		"digital_archive":    {LineageKey: "storage", LineageTier: 14, EpochKey: "digital_era"},
		"cyber_vault":        {LineageKey: "storage", LineageTier: 15, EpochKey: "neon_era"},
		"fusion_vault":       {LineageKey: "storage", LineageTier: 16, EpochKey: "neon_era"},
		"orbital_depot":      {LineageKey: "storage", LineageTier: 17, EpochKey: "neon_era"},
		"stellar_vault":      {LineageKey: "storage", LineageTier: 18, EpochKey: "cosmic_era"},
		"galactic_vault":     {LineageKey: "storage", LineageTier: 19, EpochKey: "cosmic_era"},
		"quantum_vault":      {LineageKey: "storage", LineageTier: 20, EpochKey: "cosmic_era"},

		// ── FOOD lineage (domain: food) ──────────────────────────────────────────
		"gathering_camp": {LineageKey: "food", LineageTier: 0, WorkerDomain: "food", WorkerCapacity: 5, EpochKey: "stone_era", OutputResource: "food"},
		"farm":           {LineageKey: "food", LineageTier: 1, WorkerDomain: "food", WorkerCapacity: 8, EpochKey: "stone_era", OutputResource: "food"},
		"aqueduct":       {LineageKey: "food", LineageTier: 2, WorkerDomain: "food", WorkerCapacity: 8, EpochKey: "iron_era", OutputResource: "food"},
		"colony":         {LineageKey: "food", LineageTier: 3, WorkerDomain: "food", WorkerCapacity: 12, EpochKey: "steel_era", OutputResource: "food"},
		"plantation":     {LineageKey: "food", LineageTier: 4, WorkerDomain: "food", WorkerCapacity: 10, EpochKey: "steel_era", OutputResource: "food"},
		"colony_ship":    {LineageKey: "food", LineageTier: 5, WorkerDomain: "food", WorkerCapacity: 20, EpochKey: "cosmic_era", OutputResource: "food"},

		// ── ORGANIC EXTRACTION lineage (domain: lumber) ──────────────────────────
		// Output resource transitions per epoch: wood (stone) → oil (steel/electric) → nanobots (digital+)
		"woodcutter_camp": {LineageKey: "organic_extraction", LineageTier: 0, WorkerDomain: "lumber", WorkerCapacity: 5, EpochKey: "stone_era", OutputResource: "wood"},
		"lumber_mill":     {LineageKey: "organic_extraction", LineageTier: 1, WorkerDomain: "lumber", WorkerCapacity: 8, EpochKey: "stone_era", OutputResource: "wood"},
		"oil_well":        {LineageKey: "organic_extraction", LineageTier: 2, WorkerDomain: "lumber", WorkerCapacity: 10, EpochKey: "steel_era", OutputResource: "oil"},

		// ── GEOLOGICAL EXTRACTION lineage (domain: masonry) ──────────────────────
		// Output resource transitions per epoch: stone (stone) → iron_ore/marble (iron) → coal (steel) → uranium (electric) → titanium_ore (digital) → dark_matter_crystals (neon) → antimatter (cosmic)
		"stone_pit": {LineageKey: "geological_extraction", LineageTier: 0, WorkerDomain: "masonry", WorkerCapacity: 5, EpochKey: "stone_era", OutputResource: "stone"},
		"quarry":    {LineageKey: "geological_extraction", LineageTier: 1, WorkerDomain: "masonry", WorkerCapacity: 8, EpochKey: "stone_era", OutputResource: "stone"},
		"mine":      {LineageKey: "geological_extraction", LineageTier: 2, WorkerDomain: "masonry", WorkerCapacity: 8, EpochKey: "stone_era", OutputResource: "iron"},
		"coal_mine": {LineageKey: "geological_extraction", LineageTier: 3, WorkerDomain: "masonry", WorkerCapacity: 10, EpochKey: "iron_era", OutputResource: "coal"},

		// ── KNOWLEDGE lineage (domain: knowledge) ────────────────────────────────
		"altar":              {LineageKey: "knowledge", LineageTier: 0, WorkerDomain: "knowledge", WorkerCapacity: 3, EpochKey: "stone_era", OutputResource: "knowledge"},
		"firepit":            {LineageKey: "knowledge", LineageTier: 1, WorkerDomain: "knowledge", WorkerCapacity: 3, EpochKey: "stone_era", OutputResource: "knowledge"},
		"library":            {LineageKey: "knowledge", LineageTier: 2, WorkerDomain: "knowledge", WorkerCapacity: 5, EpochKey: "stone_era", OutputResource: "knowledge"},
		"university":         {LineageKey: "knowledge", LineageTier: 3, WorkerDomain: "knowledge", WorkerCapacity: 8, EpochKey: "iron_era", OutputResource: "knowledge"},
		"observatory":        {LineageKey: "knowledge", LineageTier: 4, WorkerDomain: "knowledge", WorkerCapacity: 8, EpochKey: "steel_era", OutputResource: "knowledge"},
		"telegraph":          {LineageKey: "knowledge", LineageTier: 5, WorkerDomain: "knowledge", WorkerCapacity: 10, EpochKey: "electric_era", OutputResource: "knowledge"},
		"telephone_exchange": {LineageKey: "knowledge", LineageTier: 6, WorkerDomain: "knowledge", WorkerCapacity: 12, EpochKey: "electric_era", OutputResource: "knowledge"},
		"research_lab":       {LineageKey: "knowledge", LineageTier: 7, WorkerDomain: "knowledge", WorkerCapacity: 15, EpochKey: "digital_era", OutputResource: "knowledge"},
		"space_station":      {LineageKey: "knowledge", LineageTier: 8, WorkerDomain: "knowledge", WorkerCapacity: 20, EpochKey: "neon_era", OutputResource: "knowledge"},
		"quantum_computer":   {LineageKey: "knowledge", LineageTier: 9, WorkerDomain: "knowledge", WorkerCapacity: 25, EpochKey: "cosmic_era", OutputResource: "knowledge"},

		// ── FAITH lineage (domain: faith) ────────────────────────────────────────
		"cathedral": {LineageKey: "faith", LineageTier: 0, WorkerDomain: "faith", WorkerCapacity: 8, EpochKey: "iron_era", OutputResource: "faith"},

		// ── CULTURE/ARTS lineage (no domain — auto-produces passively) ────────────
		"amphitheater": {LineageKey: "culture_arts", LineageTier: 0, EpochKey: "iron_era", OutputResource: "culture"},
		"art_studio":   {LineageKey: "culture_arts", LineageTier: 1, EpochKey: "steel_era", OutputResource: "culture"},
		"media_center": {LineageKey: "culture_arts", LineageTier: 2, EpochKey: "digital_era", OutputResource: "culture"},

		// ── TRADE lineage (domain: trade) ────────────────────────────────────────
		"market":        {LineageKey: "trade", LineageTier: 0, WorkerDomain: "trade", WorkerCapacity: 5, EpochKey: "stone_era", OutputResource: "gold"},
		"forum":         {LineageKey: "trade", LineageTier: 1, WorkerDomain: "trade", WorkerCapacity: 8, EpochKey: "iron_era", OutputResource: "gold"},
		"bank":          {LineageKey: "trade", LineageTier: 2, WorkerDomain: "trade", WorkerCapacity: 10, EpochKey: "steel_era", OutputResource: "gold"},
		"port":          {LineageKey: "trade", LineageTier: 3, WorkerDomain: "trade", WorkerCapacity: 12, EpochKey: "steel_era", OutputResource: "gold"},
		"train_station": {LineageKey: "trade", LineageTier: 4, WorkerDomain: "trade", WorkerCapacity: 15, EpochKey: "electric_era", OutputResource: "gold"},
		"black_market":  {LineageKey: "trade", LineageTier: 5, WorkerDomain: "trade", WorkerCapacity: 15, EpochKey: "neon_era", OutputResource: "crypto"},
		"galactic_hub":  {LineageKey: "trade", LineageTier: 6, WorkerDomain: "trade", WorkerCapacity: 25, EpochKey: "cosmic_era", OutputResource: "gold"},

		// ── MILITARY lineage (domain: military) ──────────────────────────────────
		"barracks":     {LineageKey: "military", LineageTier: 0, WorkerDomain: "military", WorkerCapacity: 10, EpochKey: "iron_era"},
		"castle":       {LineageKey: "military", LineageTier: 1, WorkerDomain: "military", WorkerCapacity: 15, EpochKey: "iron_era"},
		"bunker":       {LineageKey: "military", LineageTier: 2, WorkerDomain: "military", WorkerCapacity: 20, EpochKey: "electric_era"},
		"missile_silo": {LineageKey: "military", LineageTier: 3, WorkerDomain: "military", WorkerCapacity: 25, EpochKey: "electric_era"},

		// ── ENGINEERING lineage (domain: engineering) ────────────────────────────
		"clocktower":     {LineageKey: "engineering", LineageTier: 0, WorkerDomain: "engineering", WorkerCapacity: 8, EpochKey: "electric_era"},
		"maglev_station": {LineageKey: "engineering", LineageTier: 1, WorkerDomain: "engineering", WorkerCapacity: 15, EpochKey: "neon_era", OutputResource: "gold"},
		"megastructure":  {LineageKey: "engineering", LineageTier: 2, WorkerDomain: "engineering", WorkerCapacity: 20, EpochKey: "cosmic_era"},
		"reality_engine": {LineageKey: "engineering", LineageTier: 3, WorkerDomain: "engineering", WorkerCapacity: 25, EpochKey: "cosmic_era", OutputResource: "quantum_flux"},

		// ── METALLURGY lineage (domain: metallurgy) ──────────────────────────────
		// Consumes geological ores, produces refined metals
		"smithy":        {LineageKey: "metallurgy", LineageTier: 0, WorkerDomain: "metallurgy", WorkerCapacity: 8, EpochKey: "iron_era", OutputResource: "steel"},
		"factory":       {LineageKey: "metallurgy", LineageTier: 1, WorkerDomain: "metallurgy", WorkerCapacity: 12, EpochKey: "steel_era", OutputResource: "steel"},
		"electric_mill": {LineageKey: "metallurgy", LineageTier: 2, WorkerDomain: "metallurgy", WorkerCapacity: 15, EpochKey: "electric_era", OutputResource: "steel"},
		"plasma_forge":  {LineageKey: "metallurgy", LineageTier: 3, WorkerDomain: "metallurgy", WorkerCapacity: 20, EpochKey: "neon_era", OutputResource: "steel"},
		"star_forge":    {LineageKey: "metallurgy", LineageTier: 4, WorkerDomain: "metallurgy", WorkerCapacity: 25, EpochKey: "cosmic_era", OutputResource: "steel"},

		// ── ENERGY lineage (domain: energy) ──────────────────────────────────────
		"power_grid":       {LineageKey: "energy", LineageTier: 0, WorkerDomain: "energy", WorkerCapacity: 10, EpochKey: "electric_era", OutputResource: "electricity"},
		"reactor":          {LineageKey: "energy", LineageTier: 1, WorkerDomain: "energy", WorkerCapacity: 15, EpochKey: "electric_era", OutputResource: "electricity"},
		"power_plant":      {LineageKey: "energy", LineageTier: 2, WorkerDomain: "energy", WorkerCapacity: 20, EpochKey: "digital_era", OutputResource: "electricity"},
		"fusion_reactor":   {LineageKey: "energy", LineageTier: 3, WorkerDomain: "energy", WorkerCapacity: 25, EpochKey: "neon_era", OutputResource: "electricity"},
		"smart_grid":       {LineageKey: "energy", LineageTier: 4, WorkerDomain: "energy", WorkerCapacity: 20, EpochKey: "digital_era", OutputResource: "electricity"},
		"antimatter_plant": {LineageKey: "energy", LineageTier: 5, WorkerDomain: "energy", WorkerCapacity: 25, EpochKey: "cosmic_era", OutputResource: "antimatter"},

		// ── HACKER / DIGITAL lineage (domain: hacker) ────────────────────────────
		"server_farm":         {LineageKey: "hacker", LineageTier: 0, WorkerDomain: "hacker", WorkerCapacity: 15, EpochKey: "digital_era", OutputResource: "data"},
		"fiber_hub":           {LineageKey: "hacker", LineageTier: 1, WorkerDomain: "hacker", WorkerCapacity: 20, EpochKey: "digital_era", OutputResource: "data"},
		"data_center":         {LineageKey: "hacker", LineageTier: 2, WorkerDomain: "hacker", WorkerCapacity: 25, EpochKey: "digital_era", OutputResource: "data"},
		"ai_lab":              {LineageKey: "hacker", LineageTier: 3, WorkerDomain: "hacker", WorkerCapacity: 25, EpochKey: "digital_era", OutputResource: "knowledge"},
		"augmentation_clinic": {LineageKey: "hacker", LineageTier: 4, WorkerDomain: "hacker", WorkerCapacity: 25, EpochKey: "neon_era", OutputResource: "crypto"},

		// ── ASTRONAUT lineage (domain: astronaut) ────────────────────────────────
		"launch_pad":           {LineageKey: "astronaut", LineageTier: 0, WorkerDomain: "astronaut", WorkerCapacity: 20, EpochKey: "neon_era", OutputResource: "titanium"},
		"warp_gate":            {LineageKey: "astronaut", LineageTier: 1, WorkerDomain: "astronaut", WorkerCapacity: 25, EpochKey: "cosmic_era", OutputResource: "dark_matter"},
		"transcendence_beacon": {LineageKey: "knowledge", LineageTier: 10, WorkerDomain: "knowledge", WorkerCapacity: 25, EpochKey: "cosmic_era", OutputResource: "quantum_flux"},

		// ── WONDER lineage (no workers, max 1 each) ───────────────────────────────
		"sacred_grove":         {LineageKey: "wonder", LineageTier: 0, EpochKey: "stone_era", OutputResource: "knowledge"},
		"great_monolith":       {LineageKey: "wonder", LineageTier: 1, EpochKey: "stone_era", OutputResource: "knowledge"},
		"stonehenge":           {LineageKey: "wonder", LineageTier: 2, EpochKey: "stone_era", OutputResource: "knowledge"},
		"colosseum":            {LineageKey: "wonder", LineageTier: 3, EpochKey: "iron_era", OutputResource: "culture"},
		"parthenon":            {LineageKey: "wonder", LineageTier: 4, EpochKey: "iron_era", OutputResource: "culture"},
		"great_library":        {LineageKey: "wonder", LineageTier: 5, EpochKey: "iron_era", OutputResource: "knowledge"},
		"sistine_chapel":       {LineageKey: "wonder", LineageTier: 6, EpochKey: "steel_era", OutputResource: "culture"},
		"grand_lighthouse":     {LineageKey: "wonder", LineageTier: 7, EpochKey: "steel_era", OutputResource: "gold"},
		"crystal_palace":       {LineageKey: "wonder", LineageTier: 8, EpochKey: "steel_era"},
		"eiffel_tower":         {LineageKey: "wonder", LineageTier: 9, EpochKey: "electric_era", OutputResource: "culture"},
		"hoover_dam":           {LineageKey: "wonder", LineageTier: 10, EpochKey: "electric_era", OutputResource: "electricity"},
		"particle_accelerator": {LineageKey: "wonder", LineageTier: 11, EpochKey: "electric_era", OutputResource: "knowledge"},
		"space_program":        {LineageKey: "wonder", LineageTier: 12, EpochKey: "digital_era", OutputResource: "knowledge"},
		"global_network":       {LineageKey: "wonder", LineageTier: 13, EpochKey: "digital_era", OutputResource: "data"},
		"world_simulation":     {LineageKey: "wonder", LineageTier: 14, EpochKey: "digital_era", OutputResource: "data"},
		"neon_citadel":         {LineageKey: "wonder", LineageTier: 15, EpochKey: "neon_era", OutputResource: "crypto"},
		"stellar_cradle":       {LineageKey: "wonder", LineageTier: 16, EpochKey: "neon_era", OutputResource: "plasma"},
		"dyson_scaffold":       {LineageKey: "wonder", LineageTier: 17, EpochKey: "neon_era", OutputResource: "electricity"},
		"warp_nexus":           {LineageKey: "wonder", LineageTier: 18, EpochKey: "cosmic_era", OutputResource: "dark_matter"},
		"cosmic_beacon":        {LineageKey: "wonder", LineageTier: 19, EpochKey: "cosmic_era", OutputResource: "antimatter"},
		"reality_anchor":       {LineageKey: "wonder", LineageTier: 20, EpochKey: "cosmic_era", OutputResource: "quantum_flux"},
		"singularity_core":     {LineageKey: "wonder", LineageTier: 21, EpochKey: "cosmic_era", OutputResource: "quantum_flux"},
	}
}

// BaseBuildings returns all building definitions.
// Production/housing/military/research buildings come from NewProductionBuildings() (Phase 10
// lineage redesign). Storage buildings and wonders come from baseBuildingsRaw(), enriched
// with lineage metadata from buildingMeta().
// roundSignificant rounds v to `sig` significant figures for readability
// (e.g. 46.3→46, 644→640, 6420→6400, 90150→90000). A positive value is never
// rounded down to 0 — it floors at 1 so a real cost never becomes free.
func roundSignificant(v float64, sig int) float64 {
	if v <= 0 {
		return v
	}
	d := math.Ceil(math.Log10(v))
	power := float64(sig) - d
	mag := math.Pow(10, power)
	rounded := math.Round(v*mag) / mag
	if rounded < 1 {
		return 1
	}
	return rounded
}

// normalizeCostCurves rewrites every building's cost curve as part of the
// economy rebalance (sub-ticket 1: flatten cost curves).
//
// Why: the old curves trivialized copy #1 (some bases were tiny relative to the
// steep CostScale) and then exploded in the late copies, making mid/late builds
// either free or impossibly expensive. We pivot around the 10th copy so the
// mid-game cost of a building is preserved while de-trivializing copy #1 and
// removing the late explosion.
//
// The 10th copy of a building costs base*scale^9. Holding base*scale^9 constant
// while changing scale to a flatter value requires multiplying base by
// (oldScale/newScale)^9. We apply that, then round the new base to 2 significant
// figures for readability.
//
// Wonders (CostScale 1.0, MaxCount 1) are intentionally left untouched.
//
// This operates on the freshly-constructed defs handed to it each call, reading
// each def's literal CostScale as the "old" value, so it is idempotent with
// respect to BaseBuildings() (which rebuilds from literals every call). The
// BaseCost maps are fresh per-call literals from the lineage constructors, so
// in-place mutation here cannot alias another copy.
func normalizeCostCurves(defs []BuildingDef) []BuildingDef {
	// --- Policy constants (re-tune here) ---
	const (
		pivotCopy    = 10   // copy whose cost we hold constant across the rewrite
		scaleDefault = 1.15 // new CostScale for production/research/military/etc.
		scaleInfra   = 1.13 // new CostScale for storage + housing (infrastructure)
	)
	const exponent = pivotCopy - 1 // base*scale^exponent is the pivot cost

	for i := range defs {
		d := &defs[i]
		if d.Category == "wonder" {
			continue // flat-cost, single-instance — leave completely unchanged
		}

		newScale := scaleDefault
		if d.Category == "storage" || d.Category == "housing" {
			newScale = scaleInfra
		}

		// Multiplier that preserves the pivot (10th) copy cost.
		m := math.Pow(d.CostScale/newScale, float64(exponent))
		for res := range d.BaseCost {
			d.BaseCost[res] = roundSignificant(d.BaseCost[res]*m, 2)
		}
		d.CostScale = newScale
	}
	return defs
}

func BaseBuildings() []BuildingDef {
	// Start with the new 13-lineage production buildings (all metadata inline).
	result := NewProductionBuildings()

	// Append storage and wonder buildings from legacy definitions, enriched with meta.
	meta := buildingMeta()
	for _, b := range baseBuildingsRaw() {
		if b.Category != "storage" && b.Category != "wonder" {
			continue // production buildings are now in NewProductionBuildings()
		}
		if m, ok := meta[b.Key]; ok {
			b.LineageKey = m.LineageKey
			b.LineageTier = m.LineageTier
			b.WorkerDomain = m.WorkerDomain
			b.WorkerCapacity = m.WorkerCapacity
			b.EpochKey = m.EpochKey
			b.OutputResource = m.OutputResource
		}
		result = append(result, b)
	}
	// Normalize cost curves at the single chokepoint so every consumer
	// (BuildingByKey, the engine, the audit tool) inherits flattened values.
	return normalizeCostCurves(result)
}

// BuildingByKey returns a map of building key → BuildingDef, sourced from BaseBuildings().
// If two entries share a key (e.g. a lineage file and baseBuildingsRaw() both define it),
// the last one in iteration order wins — in practice lineage files are prepended so
// storage/wonder entries always shadow any phantom duplicates.
func BuildingByKey() map[string]BuildingDef {
	m := make(map[string]BuildingDef)
	for _, b := range BaseBuildings() {
		m[b.Key] = b
	}
	return m
}

// AgeEntryCosts returns, for each resource, the cheapest base cost of that
// resource among the buildings that become available in the given age
// (RequiredAge == ageKey), excluding wonders. Used to scale age-transition
// carryover to "a handful of starter buildings." Pure function — no locks.
func AgeEntryCosts(ageKey string) map[string]float64 {
	out := make(map[string]float64)
	for _, b := range BaseBuildings() {
		if b.RequiredAge != ageKey || b.Category == "wonder" {
			continue
		}
		for res, amt := range b.BaseCost {
			if amt <= 0 {
				continue
			}
			if cur, ok := out[res]; !ok || amt < cur {
				out[res] = amt
			}
		}
	}
	return out
}

// BuildingNextTierForAge returns the next-tier BuildingDef in a lineage for the given new age.
// Returns nil if no building at tier+1 with RequiredAge == newAgeKey exists in that lineage.
// Used by the age-transition transformation pass to find what each building evolves into.
func BuildingNextTierForAge(lineageKey string, currentTier int, newAgeKey string) *BuildingDef {
	for _, b := range BaseBuildings() {
		if b.LineageKey == lineageKey && b.LineageTier == currentTier+1 && b.RequiredAge == newAgeKey {
			result := b
			return &result
		}
	}
	return nil
}
