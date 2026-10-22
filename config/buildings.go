package config

import (
	"math"

	"github.com/espresso20/ageforge/detmath"
)

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
	Description  string
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
// Storage per copy is sized by the Storage Covenant (config.StorageHold): a
// full stack of an age's storage, with every earlier age's, holds 4.5 hours
// of the age's typical income from the Bronze Age on (1.5 hours before).
// Pacing v2's away-proofing raised only the ages that fell short, to two
// significant figures: Warehouse 11K -> 18K (Bronze held 3.0 hours),
// Classical Vault 110K -> 130K, Strongroom 410K -> 470K, Colonial Warehouse
// 33M -> 38M, Industrial Depot 170M -> 190M, Victorian Vault 1.1B -> 1.3B,
// Electric Warehouse 3.5B -> 3.9B, Info Vault 790B -> 930B and Cyber Vault
// 8T -> 8.6T. Prices were left alone, so the storage ladder (smoke/static.go)
// is unchanged.
func baseBuildingsRaw() []BuildingDef {
	return []BuildingDef{
		// ===== PRIMITIVE AGE (costs: 30-100) =====
		{
			Name: "Stash", Key: "stash", Category: "storage",
			// note: the first copy MUST stay affordable within wood's 50 base storage
			// cap — stash is the only building that raises that cap, so a first-copy
			// cost above 50 is an unbuildable deadlock. 35 leaves a comfortable margin.
			BaseCost:  map[string]float64{"wood": 35},
			CostScale: 1.75,

			// note: +500/copy (not 300) so 50 stashes reach 50+50*500=25,050 food cap,
			// clearing the 16,000 stone-age gate — 300 topped out at 15,050, a softlock.
			// ~32 stashes now cover 16k, so the late-copy price spike is felt far less.
			// See yQw8uK8S + storage_deadlock_test.go.
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 100}},
			BuildTicks:  9,
			RequiredAge: "primitive_age",
			Description: "A hidden pile of supplies.",
		},

		// ===== STONE AGE (costs: 200-1000) =====
		{
			Name: "Storage Pit", Key: "storage_pit", Category: "storage",
			// note: cap raised 500->600 so the first (most expensive) copy stays
			// affordable within the storage it provides; MaxCount caps the stack
			// before the flattened 1.13 curve outruns the cap. See storage_cost_curve_test.go.
			BaseCost:  map[string]float64{"wood": 520, "stone": 340},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 1200}},
			BuildTicks:  28,
			RequiredAge: "stone_age",
			Description: "A hole in the ground to stash things.",
		},

		// ===== BRONZE AGE (costs: 1500-5000) =====
		{
			Name: "Warehouse", Key: "warehouse", Category: "storage",
			BaseCost:  map[string]float64{"wood": 3400, "stone": 2600, "iron": 520},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 6800}},
			BuildTicks:  80,
			RequiredAge: "bronze_age",
			Description: "Proper storage building.",
		},

		// ===== IRON AGE (costs: 8k-25k) =====
		{
			Name: "Granary", Key: "granary", Category: "storage",
			BaseCost:  map[string]float64{"wood": 14000, "stone": 10000},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 25000}},
			BuildTicks:  120,
			RequiredAge: "iron_age",
			Description: "Organized supply storage.",
		},

		// ===== CLASSICAL AGE (costs: 40k-120k) =====
		{
			Name: "Classical Vault", Key: "classical_vault", Category: "storage",
			BaseCost:  map[string]float64{"stone": 86000, "iron": 21000, "gold": 17000},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 210000}},
			BuildTicks:  150,
			RequiredAge: "classical_age",
			Description: "Stone vault for valuables.",
		},

		// ===== MEDIEVAL AGE (costs: 200k-600k) =====
		{
			Name: "Strongroom", Key: "keep", Category: "storage",
			// note: cap was badly under-provisioned (60k vs a ~340k normalized stone
			// cost) — copy #1 cost ~5.7x the storage it gave. Raised to 400k so the
			// keep delivers storage worthy of its (unchanged) high price.
			BaseCost:  map[string]float64{"stone": 340000, "iron": 100000, "gold": 69000},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 1.5e06}},
			BuildTicks:  200,
			RequiredAge: "medieval_age",
			Description: "Fortified storehouse.",
		},

		// ===== RENAISSANCE AGE (costs: 1M-3M) =====
		{
			Name: "Renaissance Vault", Key: "renaissance_vault", Category: "storage",
			BaseCost:  map[string]float64{"stone": 430000, "gold": 260000, "iron": 100000},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 6e06}},
			BuildTicks:  250,
			RequiredAge: "renaissance_age",
			Description: "Ornate storage facility.",
		},

		// ===== COLONIAL AGE (costs: 5M-15M) =====
		{
			Name: "Colonial Warehouse", Key: "colonial_warehouse", Category: "storage",
			BaseCost:  map[string]float64{"stone": 7.7e+06, "gold": 1000000},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 2.4e07}},
			BuildTicks:  300,
			RequiredAge: "colonial_age",
			Description: "Trade goods warehouse.",
		},

		// ===== INDUSTRIAL AGE (costs: 25M-75M) =====
		{
			Name: "Industrial Depot", Key: "industrial_depot", Category: "storage",
			BaseCost:  map[string]float64{"steel": 2.6e07, "iron": 3.4e07, "coal": 1.7e07},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 8.8e07}},
			BuildTicks:  400,
			RequiredAge: "industrial_age",
			Description: "Industrial-scale storage.",
		},

		// ===== VICTORIAN AGE (costs: 125M-375M) =====
		{
			Name: "Victorian Vault", Key: "victorian_vault", Category: "storage",
			BaseCost:  map[string]float64{"steel": 2.1e08, "gold": 1.7e08, "iron": 1.3e08},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 1400000000}},
			BuildTicks:  500,
			RequiredAge: "victorian_age",
			Description: "Reinforced vault.",
		},

		// ===== ELECTRIC AGE (costs: 600M-2B) =====
		{
			Name: "Electric Warehouse", Key: "electric_warehouse", Category: "storage",
			BaseCost:  map[string]float64{"steel": 1.3e09, "electricity": 2.1e08, "iron": 8.6e08},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 1.6e10}},
			BuildTicks:  600,
			RequiredAge: "electric_age",
			Description: "Climate-controlled storage.",
		},

		// ===== ATOMIC AGE (costs: 3B-10B) =====
		{
			Name: "Atomic Vault", Key: "atomic_vault", Category: "storage",
			// note: 1.25 scale inflated the normalized stone cost to ~19B vs a 15B cap,
			// so copy #1 walled immediately. Cap raised to 20B (still cheaper per-copy
			// than its cost is high). MaxCount caps the stack under the flattened curve.
			BaseCost:  map[string]float64{"steel": 1.2e10, "stone": 1.9e10, "iron": 7.4e09},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 79000000000}},
			BuildTicks:  700,
			RequiredAge: "atomic_age",
			Description: "Radiation-shielded storage.",
		},

		// ===== MODERN AGE (costs: 15B-50B) =====
		{
			Name: "Modern Depot", Key: "modern_depot", Category: "storage",
			BaseCost:  map[string]float64{"steel": 8.7e10, "gold": 6.2e10, "electricity": 2e10},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 340000000000}},
			BuildTicks:  800,
			RequiredAge: "modern_age",
			Description: "Automated logistics center.",
		},

		// ===== INFORMATION AGE (costs: 75B-250B) =====
		{
			Name: "Info Vault", Key: "info_vault", Category: "storage",
			BaseCost:  map[string]float64{"steel": 2.5e11, "electricity": 9.9e10, "data": 1.6e09},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 1500000000000}},
			BuildTicks:  900,
			RequiredAge: "information_age",
			Description: "Digital-physical storage hybrid.",
		},

		// ===== DIGITAL AGE (costs: 400B-1.2T) =====
		{
			Name: "Digital Archive", Key: "digital_archive", Category: "storage",
			BaseCost:  map[string]float64{"steel": 1.2e12, "data": 2.5e10, "electricity": 3.7e11},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 3200000000000}},
			BuildTicks:  1000,
			RequiredAge: "digital_age",
			Description: "Quantum-encrypted storage.",
		},

		// ===== CYBERPUNK AGE (costs: 2T-6T) =====
		{
			Name: "Cyber Vault", Key: "cyber_vault", Category: "storage",
			BaseCost:  map[string]float64{"steel": 7.4e12, "data": 1.5e11, "crypto": 2.5e10},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 28000000000000}},
			BuildTicks:  1400,
			RequiredAge: "cyberpunk_age",
			Description: "Encrypted digital vault.",
		},

		// ===== FUSION AGE (costs: 10T-30T) =====
		{
			Name: "Fusion Vault", Key: "fusion_vault", Category: "storage",
			BaseCost:  map[string]float64{"steel": 2.5e13, "plasma": 1.2e12, "electricity": 1.2e13},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 87000000000000}},
			BuildTicks:  1950,
			RequiredAge: "fusion_age",
			Description: "Plasma-shielded storage.",
		},

		// ===== SPACE AGE (costs: 50T-150T) =====
		{
			Name: "Orbital Depot", Key: "orbital_depot", Category: "storage",
			BaseCost:  map[string]float64{"steel": 1.8e14, "plasma": 2.1e13, "electricity": 1.1e14},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 3e14}},
			BuildTicks:  2145,
			RequiredAge: "space_age",
			Description: "Zero-gravity storage facility.",
		},

		// ===== INTERSTELLAR AGE (costs: 250T-750T) =====
		{
			Name: "Stellar Vault", Key: "stellar_vault", Category: "storage",
			BaseCost:  map[string]float64{"titanium": 2.1e14, "plasma": 1.8e14, "electricity": 2.8e14},
			CostScale: 1.75,

			// note: 500T -> 2Q (galactic 2Q -> 20Q, quantum 10Q -> 200Q). Cosmic-era
			// storage had fallen out of band with its own prices (median first copy
			// 13-57% of max storage vs 3-8% everywhere else), so no age gate there
			// fit the Storage Covenant. See the Gate Covenant in smoke/static.go.
			Effects:     []Effect{{Type: "storage", Target: "all", Value: 8.6e15}},
			BuildTicks:  2340,
			RequiredAge: "interstellar_age",
			Description: "Pocket-dimension storage.",
		},

		// ===== GALACTIC AGE (costs: 1.25Q-3.75Q) =====
		{
			Name: "Galactic Vault", Key: "galactic_vault", Category: "storage",
			BaseCost:  map[string]float64{"dark_matter": 3.5e14, "titanium": 1.8e15, "plasma": 7.1e14},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 5e16}},
			BuildTicks:  2340,
			RequiredAge: "galactic_age",
			Description: "Galaxy-spanning storage network.",
		},

		// ===== QUANTUM AGE (costs: 6Q-20Q) =====
		{
			Name: "Quantum Vault", Key: "quantum_vault", Category: "storage",
			// note: the steepest raw scale (1.35) inflated the normalized dark_matter
			// cost to ~9.9Q vs a 5Q cap — copy #1 walled. Cap raised to 10Q to match.
			BaseCost:  map[string]float64{"antimatter": 5e15, "dark_matter": 9.9e15, "titanium": 2.5e15},
			CostScale: 1.75,

			Effects:     []Effect{{Type: "storage", Target: "all", Value: 2.3e18}},
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
				{Type: "storage", Target: "all", Value: 5000},
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
			BaseCost:  map[string]float64{"stone": 3.84e07, "gold": 3.84e07, "faith": 32000, "culture": 950000},
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
			BaseCost:  map[string]float64{"stone": 3.2e08, "gold": 1.92e08, "steel": 3.68e07},
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
			BaseCost:  map[string]float64{"steel": 7.2e08, "iron": 3.2e08, "iron_ore": 3.2e08, "gold": 4.96e08, "coal": 6.4e08},
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
			BaseCost:  map[string]float64{"steel": 7.52e09, "iron": 6.08e09, "gold": 8.16e09},
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
			BaseCost:  map[string]float64{"steel": 1.136e11, "electricity": 3.2e10},
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
			BaseCost:  map[string]float64{"steel": 1.86e11, "electricity": 2.6e11, "uranium": 2.98e10},
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
			BaseCost:  map[string]float64{"steel": 6.6e11, "titanium_ore": 6.6e11, "gold": 1.19e12, "electricity": 7.38e11},
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
			BaseCost:  map[string]float64{"steel": 3.84e12, "titanium_ore": 3.84e12, "data": 9.28e11, "electricity": 1.536e12, "gold": 3.84e12},
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
			BaseCost:  map[string]float64{"steel": 2.72e13, "titanium_ore": 2.72e13, "data": 1.6e12, "electricity": 3.2e13},
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
			BaseCost:  map[string]float64{"steel": 1.6e14, "electricity": 6.16e13, "dark_matter_crystals": 6.16e13, "crypto": 1.92e13, "data": 1.024e13},
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
			BaseCost:  map[string]float64{"steel": 6.88e14, "plasma": 5.44e14, "electricity": 3.68e14, "dark_matter_crystals": 3.68e14},
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
			BaseCost:  map[string]float64{"titanium": 1.456e15, "plasma": 1.056e15, "steel": 8.8e15},
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
			BaseCost:  map[string]float64{"titanium": 4.96e16, "dark_matter": 3.52e15, "plasma": 4.64e16},
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
			BaseCost:  map[string]float64{"dark_matter": 3.68e16, "antimatter": 2.72e16, "titanium": 1.04e17},
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
			BaseCost:  map[string]float64{"quantum_flux": 5.6e16, "antimatter": 9.11e16, "dark_matter": 1.03e17},
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
			BaseCost:  map[string]float64{"quantum_flux": 5.76e17, "antimatter": 5.92e17, "dark_matter": 5.92e17},
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
			BaseCost:     map[string]float64{"gold": 4.7e07, "iron": 1.15e07, "iron_ore": 1.15e07},
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
			BaseCost:     map[string]float64{"gold": 1.8e08, "steel": 8.8e07},
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
			BaseCost:     map[string]float64{"gold": 6.4e07, "steel": 3.4e07, "coal": 1.3e07},
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
	rederiveLatePriceStorage(result)
	// Normalize cost curves at the single chokepoint so every consumer
	// (BuildingByKey, the engine, the audit tool) inherits flattened values,
	// then derive production rates and build times from those prices and the
	// age targets (pacing.go). The mechanical half of every Description is
	// written last, from those final values (effect_text.go).
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
