package config

import (
	"math"

	"github.com/espresso20/ageforge/detmath"
)

// Effect represents a single game effect applied by a building, tech, or milestone.
// The semantics of Value depend on Type:
//   - "production":     +Value of Target resource per tick
//   - "capacity":       +Value to Target (e.g. "population") cap
//   - "bonus":          multiplier bonus — Value is a fraction (0.1 = +10%)
//   - "unlock":         unlocks the Target building/resource key
//   - "instant_resource": immediately add Value of Target resource (milestone reward only)
//   - "permanent_bonus": persistent multiplier on Target rate (milestone reward only)
//   - "morale":          +Value to the civilization's morale per tick while the
//     building stands (FLAT, not worker-scaled; no Target)
//   - "research":        +Value of work per tick on the tech in progress, at a full
//     crew of scholars (a knowledge building) or with no crew (a wonder); no Target
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
	// FullRateCopies is, for a knowledge building, the copies that research
	// at full rate: each copy past it researches at half rate
	// (KnowledgePastCapShare). It is a soft cap, not a limit: copies past it
	// still build. 0 for every other building.
	FullRateCopies int
	Description    string
	// Flavor is optional, purely-cosmetic personality text rendered as a dim line
	// beneath the functional Description. Never holds load-bearing info (costs,
	// rates, worker counts) — it is additive humor only and is safe to be empty.
	Flavor string
	// Lineage / economy metadata (added in Phase 5+)
	LineageKey     string // which of the 14 production lineages (e.g. "food", "metallurgy", "wonder")
	LineageTier    int    // 0-indexed position within the lineage; higher tier = later age
	LineKey        string // one-resource line, set where a lineage changes resource (e.g. "organic_extraction_wood"); "" = the lineage is the line; read via Line()
	WorkerDomain   string // worker domain key for assignment (e.g. "food", "knowledge"); "" = no workers
	WorkerCapacity int    // max workers assignable per individual building instance; 0 = not applicable
	EpochKey       string // epoch this building belongs to (e.g. "stone_era", "iron_era")
	OutputResource string // primary resource produced; used by the engine to remap lineage output on age transition
	// Scholars is the scholars one copy of a knowledge building comes with
	// (scholars.go in the rules): the crew size the building had when it was
	// staffed from the population. They need no housing or food, and work from
	// the moment the copy stands. 0 for every other building.
	Scholars int
	// Soldiers is the soldiers one copy of a military building comes with
	// (soldier_rule.go): people who live in the homes and eat, at their posts
	// from the moment the copy stands. Derived when the table is built; 0 for
	// every other building.
	Soldiers int
	// levelCost is the price before capColdStartBridge held a part of it, for
	// the price levels (levelPrice). nil when no part was held.
	levelCost map[string]float64
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
//
// A storage building no longer stores anything: stock is uncapped, so each
// copy adds StorageBoostPercent percent to all production instead (the
// "bonus" effect on production_all, summed by the engine as the "storage"
// source of that pool).
func baseBuildingsRaw() []BuildingDef {
	return []BuildingDef{
		// ===== PRIMITIVE AGE (costs: 30-100) =====
		{
			Name: "Stash", Key: "stash", Category: "storage",
			BaseCost:  map[string]float64{"wood": 35},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  9,
			RequiredAge: "primitive_age",
			Description: "A hidden pile of supplies.",
		},

		// ===== STONE AGE (costs: 200-1000) =====
		{
			Name: "Storage Pit", Key: "storage_pit", Category: "storage",
			BaseCost:  map[string]float64{"wood": 520, "stone": 340},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  28,
			RequiredAge: "stone_age",
			Description: "A hole in the ground to stash things.",
		},

		// ===== BRONZE AGE (costs: 1500-5000) =====
		{
			Name: "Warehouse", Key: "warehouse", Category: "storage",
			BaseCost:  map[string]float64{"wood": 3400, "stone": 2600, "iron": 520},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  80,
			RequiredAge: "bronze_age",
			Description: "Proper storage building.",
		},

		// ===== IRON AGE (costs: 8k-25k) =====
		{
			Name: "Granary", Key: "granary", Category: "storage",
			BaseCost:  map[string]float64{"wood": 14000, "stone": 10000},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  120,
			RequiredAge: "iron_age",
			Description: "Organized supply storage.",
		},

		// ===== CLASSICAL AGE (costs: 40k-120k) =====
		{
			Name: "Classical Vault", Key: "classical_vault", Category: "storage",
			BaseCost:  map[string]float64{"stone": 86000, "iron": 21000, "gold": 17000},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  150,
			RequiredAge: "classical_age",
			Description: "Stone vault for valuables.",
		},

		// ===== MEDIEVAL AGE (costs: 200k-600k) =====
		{
			Name: "Strongroom", Key: "keep", Category: "storage",
			BaseCost:  map[string]float64{"stone": 340000, "iron": 100000, "gold": 69000},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  200,
			RequiredAge: "medieval_age",
			Description: "Fortified storehouse.",
		},

		// ===== RENAISSANCE AGE (costs: 1M-3M) =====
		{
			Name: "Renaissance Vault", Key: "renaissance_vault", Category: "storage",
			BaseCost:  map[string]float64{"stone": 430000, "gold": 260000, "iron": 100000},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  250,
			RequiredAge: "renaissance_age",
			Description: "Ornate storage facility.",
		},

		// ===== COLONIAL AGE (costs: 5M-15M) =====
		{
			Name: "Colonial Warehouse", Key: "colonial_warehouse", Category: "storage",
			BaseCost:  map[string]float64{"stone": 7.7 * Million, "gold": 1000000},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  300,
			RequiredAge: "colonial_age",
			Description: "Trade goods warehouse.",
		},

		// ===== INDUSTRIAL AGE (costs: 25M-75M) =====
		{
			Name: "Industrial Depot", Key: "industrial_depot", Category: "storage",
			BaseCost:  map[string]float64{"steel": 26 * Million, "iron": 34 * Million, "coal": 17 * Million},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  400,
			RequiredAge: "industrial_age",
			Description: "Industrial-scale storage.",
		},

		// ===== VICTORIAN AGE (costs: 125M-375M) =====
		{
			Name: "Victorian Vault", Key: "victorian_vault", Category: "storage",
			BaseCost:  map[string]float64{"steel": 210 * Million, "gold": 170 * Million, "iron": 130 * Million},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  500,
			RequiredAge: "victorian_age",
			Description: "Reinforced vault.",
		},

		// ===== ELECTRIC AGE (costs: 600M-2B) =====
		{
			Name: "Electric Warehouse", Key: "electric_warehouse", Category: "storage",
			BaseCost:  map[string]float64{"steel": 1.3 * Billion, "electricity": 210 * Million, "iron": 860 * Million},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  600,
			RequiredAge: "electric_age",
			Description: "Climate-controlled storage.",
		},

		// ===== ATOMIC AGE (costs: 3B-10B) =====
		{
			Name: "Atomic Vault", Key: "atomic_vault", Category: "storage",
			BaseCost:  map[string]float64{"steel": 12 * Billion, "stone": 19 * Billion, "iron": 7.4 * Billion},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  700,
			RequiredAge: "atomic_age",
			Description: "Radiation-shielded storage.",
		},

		// ===== MODERN AGE (costs: 15B-50B) =====
		{
			Name: "Modern Depot", Key: "modern_depot", Category: "storage",
			BaseCost:  map[string]float64{"steel": 87 * Billion, "gold": 62 * Billion, "electricity": 20 * Billion},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  800,
			RequiredAge: "modern_age",
			Description: "Automated logistics center.",
		},

		// ===== INFORMATION AGE (costs: 75B-250B) =====
		{
			Name: "Info Vault", Key: "info_vault", Category: "storage",
			BaseCost:  map[string]float64{"steel": 250 * Billion, "electricity": 99 * Billion, "data": 1.6 * Billion},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  900,
			RequiredAge: "information_age",
			Description: "Digital-physical storage hybrid.",
		},

		// ===== DIGITAL AGE (costs: 400B-1.2T) =====
		{
			Name: "Digital Archive", Key: "digital_archive", Category: "storage",
			BaseCost:  map[string]float64{"steel": 1.2 * Trillion, "data": 25 * Billion, "electricity": 370 * Billion},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  1000,
			RequiredAge: "digital_age",
			Description: "Quantum-encrypted storage.",
		},

		// ===== CYBERPUNK AGE (costs: 2T-6T) =====
		{
			Name: "Cyber Vault", Key: "cyber_vault", Category: "storage",
			BaseCost:  map[string]float64{"steel": 7.4 * Trillion, "data": 150 * Billion, "crypto": 25 * Billion},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  1400,
			RequiredAge: "cyberpunk_age",
			Description: "Encrypted digital vault.",
		},

		// ===== FUSION AGE (costs: 10T-30T) =====
		{
			Name: "Fusion Vault", Key: "fusion_vault", Category: "storage",
			BaseCost:  map[string]float64{"steel": 25 * Trillion, "plasma": 1.2 * Trillion, "electricity": 12 * Trillion},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  1950,
			RequiredAge: "fusion_age",
			Description: "Plasma-shielded storage.",
		},

		// ===== SPACE AGE (costs: 50T-150T) =====
		{
			Name: "Orbital Depot", Key: "orbital_depot", Category: "storage",
			BaseCost:  map[string]float64{"steel": 180 * Trillion, "plasma": 21 * Trillion, "electricity": 110 * Trillion},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  2145,
			RequiredAge: "space_age",
			Description: "Zero-gravity storage facility.",
		},

		// ===== INTERSTELLAR AGE (costs: 250T-750T) =====
		{
			Name: "Stellar Vault", Key: "stellar_vault", Category: "storage",
			BaseCost:  map[string]float64{"titanium": 210 * Trillion, "plasma": 180 * Trillion, "electricity": 280 * Trillion},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  2340,
			RequiredAge: "interstellar_age",
			Description: "Pocket-dimension storage.",
		},

		// ===== GALACTIC AGE (costs: 1.25Q-3.75Q) =====
		{
			Name: "Galactic Vault", Key: "galactic_vault", Category: "storage",
			BaseCost:  map[string]float64{"dark_matter": 350 * Trillion, "titanium": 1.8 * Quadrillion, "plasma": 710 * Trillion},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  2340,
			RequiredAge: "galactic_age",
			Description: "Galaxy-spanning storage network.",
		},

		// ===== QUANTUM AGE (costs: 6Q-20Q) =====
		{
			Name: "Quantum Vault", Key: "quantum_vault", Category: "storage",
			BaseCost:  map[string]float64{"antimatter": 5 * Quadrillion, "dark_matter": 9.9 * Quadrillion, "titanium": 2.5 * Quadrillion},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "bonus", Target: "production_all", Value: StorageBoostPercent / 100.0}},
			BuildTicks:  2340,
			RequiredAge: "quantum_age",
			Description: "Stores matter in quantum superposition.",
		},

		// ===== TRANSCENDENT AGE =====
		// (singularity_core is a wonder, listed below)

		// ===== WONDERS =====
		// There is exactly one wonder per age (22 wonders total). Building a wonder
		// unlocks +0.5x game speed — the primary long-term speed upgrade mechanic.
		// Wonder costs use WonderBank: resources must be "banked" via 'wonder collect'
		// before 'build <key>' queues construction. CostScale is always 1.0.
		// Build ticks are extremely long to make each wonder a meaningful milestone.
		//
		// RequiredTech is the wonder's keystone: a tech of its own age that must
		// be researched before 'build <key>' (the bank fills without it). The
		// next age needs the wonder, so the keystone is the one tech an age
		// asks for. The Sacred Grove has none (nothing blocks the first age),
		// and Stonehenge has none yet (see there).

		// Primitive Age — normal costs: 30-300
		{
			Name: "Sacred Grove", Key: "sacred_grove", Category: "wonder",
			BaseCost:  map[string]float64{"wood": 1600, "food": 800},
			CostScale: 1,
			Effects: []Effect{
				{Type: "research", Value: 0.02},
				{Type: "production", Target: "food", Value: 0.05},
			},
			RequiredAge: "primitive_age",
			MaxCount:    1,
			BuildTicks:  75,
			Description: "An ancient clearing where nature's power flows.",
		},
		// Stone Age — normal costs: 200-1000
		{
			Name: "Great Monolith", Key: "great_monolith", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 10080, "wood": 8000, "food": 2400},
			CostScale: 1,
			Effects: []Effect{
				{Type: "research", Value: 0.05},
			},
			RequiredAge:  "stone_age",
			RequiredTech: "stoneworking",
			MaxCount:     1,
			BuildTicks:   225,
			Description:  "A towering stone pillar visible for miles.",
		},
		// Bronze Age — normal costs: 1500-2500
		{
			Name: "Stonehenge", Key: "stonehenge", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 54400, "wood": 30400, "iron": 5440},
			CostScale: 1,
			Effects: []Effect{
				{Type: "research", Value: 0.8},
				{Type: "production", Target: "faith", Value: 0.6},
			},
			RequiredAge:  "bronze_age",
			RequiredTech: "calendar",
			MaxCount:     1,
			BuildTicks:   1170,
			Description:  "Massive stone circle aligned to the cosmos.",
		},
		// Iron Age — normal costs: 8k-12k
		{
			Name: "Colosseum", Key: "colosseum", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 256000, "marble": 256000, "iron": 56800, "iron_ore": 56800, "gold": 100800},
			CostScale: 1,
			Effects: []Effect{
				{Type: "capacity", Target: "population", Value: 100},
				{Type: "production", Target: "culture", Value: 2},
			},
			RequiredAge:  "iron_age",
			RequiredTech: "mathematics",
			MaxCount:     1,
			BuildTicks:   1950,
			Description:  "Grand arena of blood and glory.",
		},
		// Classical Age — normal costs: 40k-80k
		{
			Name: "Parthenon", Key: "parthenon", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 800000, "marble": 800000, "gold": 704000, "iron": 352000, "iron_ore": 352000},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 2},
				{Type: "research", Value: 1.2},
			},
			RequiredAge:  "classical_age",
			RequiredTech: "philosophy",
			MaxCount:     1,
			BuildTicks:   2500,
			Description:  "Perfect temple of marble and wisdom.",
		},
		// Medieval Age — normal costs: 180k-360k
		{
			Name: "Great Library", Key: "great_library", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 10500000, "gold": 8070000},
			CostScale: 1,
			Effects: []Effect{
				{Type: "research", Value: 2},
				{Type: "bonus", Target: "knowledge_rate", Value: 0.3},
			},
			RequiredAge:  "medieval_age",
			RequiredTech: "theology",
			MaxCount:     1,
			BuildTicks:   3510,
			Description:  "Repository of all knowledge.",
		},
		// Renaissance Age — normal costs: 400k-600k
		{
			Name: "Sistine Chapel", Key: "sistine_chapel", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 38.4 * Million, "gold": 38.4 * Million, "faith": 32000, "culture": 960000},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 3.5},
				{Type: "production", Target: "faith", Value: 1.8},
			},
			RequiredAge:  "renaissance_age",
			RequiredTech: "patronage",
			MaxCount:     1,
			BuildTicks:   4680,
			Description:  "Ceiling painted by divine hands.",
		},
		// Colonial Age — normal costs: 1.2M-2M
		{
			Name: "Grand Lighthouse", Key: "grand_lighthouse", Category: "wonder",
			BaseCost:  map[string]float64{"stone": 320 * Million, "gold": 192 * Million, "steel": 36.8 * Million},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 5},
				{Type: "bonus", Target: "expedition_reward", Value: 0.8},
			},
			RequiredAge:  "colonial_age",
			RequiredTech: "cartography",
			MaxCount:     1,
			BuildTicks:   5460,
			Description:  "Beacon visible across oceans.",
		},
		// Industrial Age — normal costs: 12M-25M
		{
			Name: "Crystal Palace", Key: "crystal_palace", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 720 * Million, "iron": 320 * Million, "iron_ore": 320 * Million, "gold": 496 * Million, "coal": 640 * Million},
			CostScale: 1,
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.15},
				{Type: "production", Target: "gold", Value: 8},
			},
			RequiredAge:  "industrial_age",
			RequiredTech: "industrialization",
			MaxCount:     1,
			BuildTicks:   6240,
			Description:  "Glass cathedral of industry.",
		},
		// Victorian Age — normal costs: 90M-150M
		{
			Name: "Eiffel Tower", Key: "eiffel_tower", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 7.52 * Billion, "iron": 6.08 * Billion, "gold": 8.16 * Billion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 5},
				{Type: "research", Value: 2},
			},
			RequiredAge:  "victorian_age",
			RequiredTech: "mass_production",
			MaxCount:     1,
			BuildTicks:   7020,
			Description:  "Iron monument piercing the sky.",
		},
		// Electric Age — normal costs: 500M-1B
		{
			Name: "Hoover Dam", Key: "hoover_dam", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 113.6 * Billion, "electricity": 32 * Billion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 10},
				{Type: "bonus", Target: "production_all", Value: 0.2},
			},
			RequiredAge:  "electric_age",
			RequiredTech: "power_distribution",
			MaxCount:     1,
			BuildTicks:   7800,
			Description:  "Taming a river to power a nation.",
		},
		// Atomic Age — normal costs: 3B-10B
		{
			Name: "Particle Accelerator", Key: "particle_accelerator", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 186 * Billion, "electricity": 260 * Billion, "uranium": 29.8 * Billion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "research", Value: 10},
				{Type: "production", Target: "uranium", Value: 1.5},
			},
			RequiredAge:  "atomic_age",
			RequiredTech: "nuclear_fission",
			MaxCount:     1,
			BuildTicks:   9360,
			Description:  "Smashes atoms for science.",
		},
		// Modern Age — normal costs: 15B-40B
		{
			Name: "Space Program", Key: "space_program", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 660 * Billion, "titanium_ore": 660 * Billion, "gold": 1.19 * Trillion, "electricity": 738 * Billion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "research", Value: 6},
				{Type: "production", Target: "culture", Value: 8},
			},
			RequiredAge:  "modern_age",
			RequiredTech: "satellite_tech",
			MaxCount:     1,
			BuildTicks:   9360,
			Description:  "Reaching for the stars.",
		},
		// Information Age — normal costs: 75B-125B
		{
			Name: "Global Network", Key: "global_network", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 3.84 * Trillion, "titanium_ore": 3.84 * Trillion, "data": 928 * Billion, "electricity": 1.536 * Trillion, "gold": 3.84 * Trillion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "data", Value: 30},
				{Type: "bonus", Target: "knowledge_rate", Value: 0.3},
			},
			RequiredAge:  "information_age",
			RequiredTech: "internet",
			MaxCount:     1,
			BuildTicks:   10920,
			Description:  "Every mind connected.",
		},
		// Digital Age — normal costs: 400B-750B
		{
			Name: "World Simulation", Key: "world_simulation", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 27.2 * Trillion, "titanium_ore": 27.2 * Trillion, "data": 1.6 * Trillion, "electricity": 32 * Trillion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "data", Value: 60},
				{Type: "research", Value: 15},
			},
			RequiredAge:  "digital_age",
			RequiredTech: "machine_learning",
			MaxCount:     1,
			BuildTicks:   12480,
			Description:  "A digital twin of reality itself.",
		},
		// Cyberpunk Age — normal costs: 2T-4T
		{
			Name: "Neon Citadel", Key: "neon_citadel", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 160 * Trillion, "electricity": 61.6 * Trillion, "dark_matter_crystals": 61.6 * Trillion, "crypto": 19.2 * Trillion, "data": 10.24 * Trillion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "crypto", Value: 10},
				{Type: "capacity", Target: "population", Value: 500},
			},
			RequiredAge:  "cyberpunk_age",
			RequiredTech: "cybernetics",
			MaxCount:     1,
			BuildTicks:   14040,
			Description:  "A city within a city, lit by eternal neon.",
		},
		// Fusion Age — normal costs: 10T-15T
		{
			Name: "Stellar Cradle", Key: "stellar_cradle", Category: "wonder",
			BaseCost:  map[string]float64{"steel": 688 * Trillion, "plasma": 544 * Trillion, "electricity": 368 * Trillion, "dark_matter_crystals": 368 * Trillion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "plasma", Value: 15},
				{Type: "production", Target: "electricity", Value: 200},
			},
			RequiredAge:  "fusion_age",
			RequiredTech: "fusion_power",
			MaxCount:     1,
			BuildTicks:   15600,
			Description:  "A miniature star harnessed for power.",
		},
		// Space Age — normal costs: 50T-80T
		{
			Name: "Dyson Scaffold", Key: "dyson_scaffold", Category: "wonder",
			BaseCost:  map[string]float64{"titanium": 1.456 * Quadrillion, "plasma": 1.056 * Quadrillion, "steel": 8.8 * Quadrillion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 200},
				{Type: "production", Target: "plasma", Value: 30},
			},
			RequiredAge:  "space_age",
			RequiredTech: "orbital_mechanics",
			MaxCount:     1,
			BuildTicks:   17160,
			Description:  "Framework for a Dyson sphere.",
		},
		// Interstellar Age — normal costs: 250T-500T
		{
			Name: "Warp Nexus", Key: "warp_nexus", Category: "wonder",
			BaseCost:  map[string]float64{"titanium": 49.6 * Quadrillion, "dark_matter": 3.52 * Quadrillion, "plasma": 46.4 * Quadrillion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "dark_matter", Value: 8},
				{Type: "bonus", Target: "production_all", Value: 0.8},
			},
			RequiredAge:  "interstellar_age",
			RequiredTech: "warp_drive",
			MaxCount:     1,
			BuildTicks:   18720,
			Description:  "Hub of faster-than-light corridors.",
		},
		// Galactic Age — normal costs: 1Q-1.5Q
		{
			Name: "Cosmic Beacon", Key: "cosmic_beacon", Category: "wonder",
			BaseCost:  map[string]float64{"dark_matter": 36.8 * Quadrillion, "antimatter": 27.2 * Quadrillion, "titanium": 104 * Quadrillion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "antimatter", Value: 10},
				{Type: "bonus", Target: "production_all", Value: 0.5},
			},
			RequiredAge:  "galactic_age",
			RequiredTech: "galactic_navigation",
			MaxCount:     1,
			BuildTicks:   18720,
			Description:  "A signal fire across the galaxy.",
		},
		// Quantum Age — normal costs: 2.5Q-5Q
		{
			Name: "Reality Anchor", Key: "reality_anchor", Category: "wonder",
			BaseCost:  map[string]float64{"quantum_flux": 56 * Quadrillion, "antimatter": 91.1 * Quadrillion, "dark_matter": 103 * Quadrillion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "production", Target: "quantum_flux", Value: 15},
				{Type: "bonus", Target: "production_all", Value: 0.5},
			},
			RequiredAge:  "quantum_age",
			RequiredTech: "quantum_mechanics",
			MaxCount:     1,
			BuildTicks:   18720,
			Description:  "Stabilizes reality across dimensions.",
		},
		// Transcendent Age
		{
			Name: "Singularity Core", Key: "singularity_core", Category: "wonder",
			BaseCost:  map[string]float64{"quantum_flux": 576 * Quadrillion, "antimatter": 592 * Quadrillion, "dark_matter": 592 * Quadrillion},
			CostScale: 1,
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 2},
				{Type: "production", Target: "quantum_flux", Value: 20},
			},
			RequiredAge:  "transcendent_age",
			RequiredTech: "transcendence",
			MaxCount:     1,
			BuildTicks:   18720,
			Description:  "The final wonder.",
		},
		// Diplomacy / foreign affairs — these act on the world rather than on a
		// resource, so they live here in the admin/storage slice rather than a
		// lineage file (the lineage files are strict one-per-age tier ladders with
		// a per-tier output formula; a one-off administrative building does not fit
		// one). NO Type:"production" effect on any of them — that would pollute the
		// resource-rate map in WorkerScaledProduction.
		//
		// NOTE for anyone adding to this slice: BaseBuildings() only admits
		// Category "storage", "wonder" and "diplomacy" out of baseBuildingsRaw().
		// Any other Category here is silently dropped.
		{
			Name: "Embassy", Key: "embassy", Category: "diplomacy",
			BaseCost:     map[string]float64{"gold": 47 * Million, "iron": 11.5 * Million, "iron_ore": 11.5 * Million},
			CostScale:    1.15,
			Effects:      []Effect{{Type: "opinion", Target: "opinion", Value: 0.05}},
			BuildTicks:   3600,
			RequiredAge:  "colonial_age",
			RequiredTech: "embassies",
			Description:  "A diplomatic embassy.",
			WorkerDomain: "trade", WorkerCapacity: 5,
			EpochKey: "steel_era",
		},
		{
			Name: "Grand Embassy", Key: "grand_embassy", Category: "diplomacy",
			BaseCost:     map[string]float64{"gold": 180 * Million, "steel": 88 * Million},
			CostScale:    1.15,
			Effects:      []Effect{{Type: "opinion", Target: "opinion", Value: 0.1}},
			BuildTicks:   3600,
			RequiredAge:  "industrial_age",
			RequiredTech: "concert_of_nations",
			Description:  "A grand diplomatic complex.",
			WorkerDomain: "trade", WorkerCapacity: 8,
			EpochKey: "steel_era",
		},
		// The Geographic Society is the exploration arm of the foreign service: it
		// keeps standing survey parties on the books and sends them out on its own
		// initiative, which is what makes an idle empire keep meeting the world (see
		// game/auto_expedition.go). Deliberately an INDUSTRIAL-age unlock — the real
		// geographic societies are a 19th-century institution, the age already
		// carries the Grand Embassy and the Research Institute, and by then the
		// player has had eight ages of dispatching parties by hand and earned the
		// right to delegate it.
		//
		// It carries NO Effects entry on purpose. Its behaviour is not a modifier on
		// anything — the engine keys off the building KEY, exactly as the trade
		// system keys off "market"/"port" (see game/trade.go). An invented effect
		// type here would be inert in every consumer and would only have to be
		// whitelisted in the effect-target validator.
		{
			Name: "Geographic Society", Key: "geographic_society", Category: "diplomacy",
			BaseCost:     map[string]float64{"gold": 64 * Million, "steel": 34 * Million, "coal": 13 * Million},
			CostScale:    1.15,
			BuildTicks:   3600,
			RequiredAge:  "industrial_age",
			RequiredTech: "geographic_societies",
			Description:  "A chartered society of surveyors and cartographers. Sends out scouting expeditions on its own. More societies and more workers shorten the wait (8 workers).",
			WorkerDomain: "military", WorkerCapacity: 8,
			EpochKey: "steel_era",
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
// Worker domains: food, lumber, masonry, faith, military, trade, engineering,
//
//	metallurgy, energy, hacker, astronaut  (culture_arts has no domain)
//
// Epoch keys: stone_era, iron_era, steel_era, electric_era, digital_era, neon_era, cosmic_era
func buildingMeta() map[string]buildingMetaEntry {
	return map[string]buildingMetaEntry{
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
	d := math.Ceil(detmath.Log10(v))
	power := float64(sig) - d
	var rounded float64
	if power >= 0 {
		mag := detmath.Pow(10, power)
		rounded = math.Round(v*mag) / mag
	} else {
		// Multiply by an exact power of ten rather than divide by an inexact
		// fraction: 10/1e-5 came out as 999999.9999999999, and a wonder bank
		// filled in whole units could then never reach its price.
		mag := detmath.Pow(10, -power)
		rounded = math.Round(v/mag) * mag
	}
	if rounded < 1 {
		return 1
	}
	return rounded
}

func BaseBuildings() []BuildingDef {
	defs, _ := buildBase()
	return defs
}

// ColdStartCaps lists the price parts the cold-start rule held on the loaded
// table (capColdStartBridge), in age order.
func ColdStartCaps() []ColdStartCap {
	_, caps := buildBase()
	return caps
}

func buildBase() ([]BuildingDef, []ColdStartCap) {
	// Start with the new 13-lineage production buildings (all metadata inline).
	result := NewProductionBuildings()

	// Append storage and wonder buildings from legacy definitions, enriched with meta.
	meta := buildingMeta()
	for _, b := range baseBuildingsRaw() {
		if b.Category != "storage" && b.Category != "wonder" && b.Category != "diplomacy" {
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
	// Attach cosmetic Flavor text at the single chokepoint so every consumer
	// (BuildingByKey, the engine, the UI overlays) sees it without each lineage
	// file having to carry the strings inline. Flavor is purely additive humor;
	// see building_flavor.go. Functional Description is never modified here.
	applyBuildingFlavor(result)
	// Apply the late-age price curve after every authored definition has joined
	// the runtime table, including storage, wonders and raw-material parts.
	applyLatePriceClimb(result)
	applyNewCopyPrice(result)
	caps := capColdStartBridge(result)
	applyWonderBuildFactor(result)
	// Normalize cost curves at the single chokepoint so every consumer
	// (BuildingByKey, the engine, the audit tool) inherits flattened values,
	// then derive production rates and build times from those prices and the
	// age targets (pacing.go). The mechanical half of every Description is
	// written last, from those final values (effect_text.go).
	applySoldierPosts(result)
	return appendEffectText(result), caps
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

// Line is the one-resource line a building belongs to: its LineKey, or its
// lineage when the lineage never changes resource. A line never switches
// resource, so a building only upgrades to the next rung of its own line.
func (d BuildingDef) Line() string {
	if d.LineKey != "" {
		return d.LineKey
	}
	return d.LineageKey
}

// NextRung is the building one rung up the line from tier: the one with the
// lowest tier above it in the same line. A line that pauses (titanium after
// Modern, electricity while oil runs) has its next rung in a later age, which
// is still the next rung. ok is false when tier is the last of its line.
func NextRung(defs []BuildingDef, line string, tier int) (next BuildingDef, ok bool) {
	if line == "" {
		return BuildingDef{}, false
	}
	for _, b := range defs {
		if b.Line() != line || b.LineageTier <= tier {
			continue
		}
		if !ok || b.LineageTier < next.LineageTier {
			next, ok = b, true
		}
	}
	return next, ok
}
