package config

// newProductionBuildings returns all production, housing, research, and military buildings
// for the 13 lineage chains introduced in Phase 10 of the economy redesign.
// Storage buildings and wonders remain in baseBuildingsRaw().
// All fields are set inline; no separate buildingMeta() entries needed for these buildings.
func newProductionBuildings() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 1 — HOUSING (lineageKey: "housing", no workers, no output)
	// CostScale: 1.12  Category: "housing"
	// Effect: capacity / population
	// =========================================================================

	// tier 0 — primitive_age
	b = append(b, BuildingDef{
		Name: "Hut", Key: "hut", Category: "housing",
		BaseCost:       map[string]float64{"wood": 15},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 10}},
		BuildTicks:     80,
		RequiredAge:    "primitive_age",
		Description:    "A crude shelter of sticks and leaves. +10 pop cap.",
		LineageKey:     "housing", LineageTier: 0,
		EpochKey: "stone_era",
	})
	// tier 1 — stone_age
	b = append(b, BuildingDef{
		Name: "Longhouse", Key: "longhouse", Category: "housing",
		BaseCost:       map[string]float64{"wood": 120, "stone": 80},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 25}},
		BuildTicks:     200,
		RequiredAge:    "stone_age",
		Description:    "A communal longhouse. +25 pop cap.",
		LineageKey:     "housing", LineageTier: 1,
		EpochKey: "stone_era",
	})
	// tier 2 — bronze_age
	b = append(b, BuildingDef{
		Name: "House", Key: "house", Category: "housing",
		BaseCost:       map[string]float64{"wood": 600, "stone": 400},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 50}},
		BuildTicks:     500,
		RequiredAge:    "bronze_age",
		Description:    "A sturdy family dwelling. +50 pop cap.",
		LineageKey:     "housing", LineageTier: 2,
		EpochKey: "stone_era",
	})
	// tier 3 — iron_age
	b = append(b, BuildingDef{
		Name: "Townhouse", Key: "townhouse", Category: "housing",
		BaseCost:       map[string]float64{"stone": 6000, "iron": 2000},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 80}},
		BuildTicks:     1000,
		RequiredAge:    "iron_age",
		Description:    "A multi-floor townhouse. +80 pop cap.",
		LineageKey:     "housing", LineageTier: 3,
		EpochKey: "iron_era",
	})
	// tier 4 — classical_age
	b = append(b, BuildingDef{
		Name: "Villa", Key: "villa", Category: "housing",
		BaseCost:       map[string]float64{"stone": 40000, "gold": 10000},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 120}},
		BuildTicks:     3000,
		RequiredAge:    "classical_age",
		Description:    "An elegant classical villa. +120 pop cap.",
		LineageKey:     "housing", LineageTier: 4,
		EpochKey: "iron_era",
	})
	// tier 5 — medieval_age
	b = append(b, BuildingDef{
		Name: "Manor", Key: "manor", Category: "housing",
		BaseCost:       map[string]float64{"stone": 200000, "gold": 50000},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 200}},
		BuildTicks:     6000,
		RequiredAge:    "medieval_age",
		Description:    "A lord's country manor. +200 pop cap.",
		LineageKey:     "housing", LineageTier: 5,
		EpochKey: "iron_era",
	})
	// tier 6 — renaissance_age
	b = append(b, BuildingDef{
		Name: "Estate", Key: "estate", Category: "housing",
		BaseCost:       map[string]float64{"stone": 600000, "gold": 250000},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 350}},
		BuildTicks:     12000,
		RequiredAge:    "renaissance_age",
		Description:    "A grand estate with grounds. +350 pop cap.",
		LineageKey:     "housing", LineageTier: 6,
		EpochKey: "steel_era",
	})
	// tier 7 — colonial_age
	b = append(b, BuildingDef{
		Name: "Settlement Block", Key: "settlement_block", Category: "housing",
		BaseCost:       map[string]float64{"wood": 2e6, "stone": 1.5e6},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 600}},
		BuildTicks:     18000,
		RequiredAge:    "colonial_age",
		Description:    "A colonial settlement block. +600 pop cap.",
		LineageKey:     "housing", LineageTier: 7,
		EpochKey: "steel_era",
	})
	// tier 8 — industrial_age
	b = append(b, BuildingDef{
		Name: "Tenement", Key: "tenement", Category: "housing",
		BaseCost:       map[string]float64{"stone": 12e6, "iron": 8e6},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 1000}},
		BuildTicks:     25000,
		RequiredAge:    "industrial_age",
		Description:    "Dense worker housing. +1000 pop cap.",
		LineageKey:     "housing", LineageTier: 8,
		EpochKey: "steel_era",
	})
	// tier 9 — victorian_age
	b = append(b, BuildingDef{
		Name: "Row House", Key: "row_house", Category: "housing",
		BaseCost:       map[string]float64{"steel": 80e6, "stone": 60e6},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 1800}},
		BuildTicks:     50000,
		RequiredAge:    "victorian_age",
		Description:    "A Victorian terrace of row houses. +1800 pop cap.",
		LineageKey:     "housing", LineageTier: 9,
		EpochKey: "electric_era",
	})
	// tier 10 — electric_age
	b = append(b, BuildingDef{
		Name: "Apartment Block", Key: "apartment_block", Category: "housing",
		BaseCost:       map[string]float64{"steel": 500e6, "electricity": 200e6},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 3200}},
		BuildTicks:     75000,
		RequiredAge:    "electric_age",
		Description:    "A modern apartment block. +3200 pop cap.",
		LineageKey:     "housing", LineageTier: 10,
		EpochKey: "electric_era",
	})
	// tier 11 — atomic_age
	b = append(b, BuildingDef{
		Name: "Housing Project", Key: "housing_project", Category: "housing",
		BaseCost:       map[string]float64{"steel": 3e9, "electricity": 1e9},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 5500}},
		BuildTicks:     100000,
		RequiredAge:    "atomic_age",
		Description:    "Government-built housing towers. +5500 pop cap.",
		LineageKey:     "housing", LineageTier: 11,
		EpochKey: "electric_era",
	})
	// tier 12 — modern_age
	b = append(b, BuildingDef{
		Name: "Tower Block", Key: "tower_block", Category: "housing",
		BaseCost:       map[string]float64{"steel": 20e9, "electricity": 8e9},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 10000}},
		BuildTicks:     150000,
		RequiredAge:    "modern_age",
		Description:    "A soaring residential tower. +10000 pop cap.",
		LineageKey:     "housing", LineageTier: 12,
		EpochKey: "digital_era",
	})
	// tier 13 — information_age
	b = append(b, BuildingDef{
		Name: "Smart Complex", Key: "smart_complex", Category: "housing",
		BaseCost:       map[string]float64{"steel": 120e9, "electricity": 50e9, "data": 5e9},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 18000}},
		BuildTicks:     300000,
		RequiredAge:    "information_age",
		Description:    "AI-managed smart living complex. +18000 pop cap.",
		LineageKey:     "housing", LineageTier: 13,
		EpochKey: "digital_era",
	})
	// tier 14 — digital_age
	b = append(b, BuildingDef{
		Name: "Megaplex", Key: "megaplex", Category: "housing",
		BaseCost:       map[string]float64{"steel": 700e9, "electricity": 300e9, "data": 30e9},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 32000}},
		BuildTicks:     500000,
		RequiredAge:    "digital_age",
		Description:    "A self-contained urban megaplex. +32000 pop cap.",
		LineageKey:     "housing", LineageTier: 14,
		EpochKey: "digital_era",
	})
	// tier 15 — cyberpunk_age
	b = append(b, BuildingDef{
		Name: "Arcology Pod", Key: "arcology_pod", Category: "housing",
		BaseCost:       map[string]float64{"steel": 5e12, "data": 150e9, "crypto": 500e9},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 55000}},
		BuildTicks:     1000000,
		RequiredAge:    "cyberpunk_age",
		Description:    "A self-sustaining arcology pod. +55000 pop cap.",
		LineageKey:     "housing", LineageTier: 15,
		EpochKey: "neon_era",
	})
	// tier 16 — fusion_age
	b = append(b, BuildingDef{
		Name: "Habitat Ring", Key: "habitat_ring", Category: "housing",
		BaseCost:       map[string]float64{"steel": 35e12, "plasma": 5e12, "electricity": 15e12},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 100000}},
		BuildTicks:     1500000,
		RequiredAge:    "fusion_age",
		Description:    "A rotating habitat ring in orbit. +100000 pop cap.",
		LineageKey:     "housing", LineageTier: 16,
		EpochKey: "neon_era",
	})
	// tier 17 — space_age
	b = append(b, BuildingDef{
		Name: "Orbital Habitat", Key: "orbital_habitat", Category: "housing",
		BaseCost:       map[string]float64{"titanium": 50e12, "plasma": 30e12, "steel": 200e12},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 180000}},
		BuildTicks:     2000000,
		RequiredAge:    "space_age",
		Description:    "A vast orbital habitat complex. +180000 pop cap.",
		LineageKey:     "housing", LineageTier: 17,
		EpochKey: "neon_era",
	})
	// tier 18 — interstellar_age
	b = append(b, BuildingDef{
		Name: "Generation Ship", Key: "generation_ship", Category: "housing",
		BaseCost:       map[string]float64{"titanium": 500e12, "dark_matter": 50e12, "plasma": 300e12},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 320000}},
		BuildTicks:     2500000,
		RequiredAge:    "interstellar_age",
		Description:    "A ship that houses entire generations. +320000 pop cap.",
		LineageKey:     "housing", LineageTier: 18,
		EpochKey: "cosmic_era",
	})
	// tier 19 — galactic_age
	b = append(b, BuildingDef{
		Name: "Dyson Sphere Habitat", Key: "dyson_sphere_habitat", Category: "housing",
		BaseCost:       map[string]float64{"dark_matter": 500e12, "titanium": 5e15, "antimatter": 100e12},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 600000}},
		BuildTicks:     3000000,
		RequiredAge:    "galactic_age",
		Description:    "Living quarters within a Dyson sphere. +600000 pop cap.",
		LineageKey:     "housing", LineageTier: 19,
		EpochKey: "cosmic_era",
	})
	// tier 20 — quantum_age
	b = append(b, BuildingDef{
		Name: "Reality Fold", Key: "reality_fold", Category: "housing",
		BaseCost:       map[string]float64{"quantum_flux": 100e12, "antimatter": 50e15, "dark_matter": 40e15},
		CostScale:      1.12,
		Effects:        []Effect{{Type: "capacity", Target: "population", Value: 1000000}},
		BuildTicks:     5000000,
		RequiredAge:    "quantum_age",
		Description:    "A folded-reality habitation zone. +1000000 pop cap.",
		LineageKey:     "housing", LineageTier: 20,
		EpochKey: "cosmic_era",
	})

	// =========================================================================
	// LINEAGE 2 — FOOD (lineageKey: "food", domain: "food", output: "food")
	// rate = 0.05 * 2^tier  CostScale: 1.30  Category: "production"
	// =========================================================================

	// tier 0 — primitive_age  rate=0.05
	b = append(b, BuildingDef{
		Name: "Gathering Camp", Key: "gathering_camp", Category: "production",
		BaseCost:       map[string]float64{"wood": 20},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 0.05}},
		BuildTicks:     80,
		RequiredAge:    "primitive_age",
		Description:    "Foragers gather berries and roots. +0.05 food/tick (3 workers).",
		LineageKey:     "food", LineageTier: 0,
		WorkerDomain: "food", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "food",
	})
	// tier 1 — stone_age  rate=0.10
	b = append(b, BuildingDef{
		Name: "Forager Post", Key: "forager_post", Category: "production",
		BaseCost:       map[string]float64{"wood": 150, "stone": 80},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 0.10}},
		BuildTicks:     200,
		RequiredAge:    "stone_age",
		Description:    "Organised foraging post. +0.10 food/tick (4 workers).",
		LineageKey:     "food", LineageTier: 1,
		WorkerDomain: "food", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "food",
	})
	// tier 2 — bronze_age  rate=0.20
	b = append(b, BuildingDef{
		Name: "Farm", Key: "farm", Category: "production",
		BaseCost:       map[string]float64{"wood": 800, "stone": 500},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 0.20}},
		BuildTicks:     500,
		RequiredAge:    "bronze_age",
		Description:    "Cultivated fields produce steady food. +0.20 food/tick (5 workers).",
		LineageKey:     "food", LineageTier: 2,
		WorkerDomain: "food", WorkerCapacity: 5,
		EpochKey: "stone_era", OutputResource: "food",
	})
	// tier 3 — iron_age  rate=0.40
	b = append(b, BuildingDef{
		Name: "Field Works", Key: "field_works", Category: "production",
		BaseCost:       map[string]float64{"stone": 5000, "iron": 2000, "wood": 3000},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 0.40}},
		BuildTicks:     1000,
		RequiredAge:    "iron_age",
		Description:    "Iron-tool farming with irrigation. +0.40 food/tick (5 workers).",
		LineageKey:     "food", LineageTier: 3,
		WorkerDomain: "food", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "food",
	})
	// tier 4 — classical_age  rate=0.80
	b = append(b, BuildingDef{
		Name: "Estate Farm", Key: "estate_farm", Category: "production",
		BaseCost:       map[string]float64{"stone": 35000, "gold": 12000, "iron": 10000},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 0.80}},
		BuildTicks:     3000,
		RequiredAge:    "classical_age",
		Description:    "A large estate with managed farmlands. +0.80 food/tick (6 workers).",
		LineageKey:     "food", LineageTier: 4,
		WorkerDomain: "food", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "food",
	})
	// tier 5 — medieval_age  rate=1.60
	b = append(b, BuildingDef{
		Name: "Demesne", Key: "demesne", Category: "production",
		BaseCost:       map[string]float64{"stone": 180000, "gold": 60000, "knowledge": 20000},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 1.60}},
		BuildTicks:     6000,
		RequiredAge:    "medieval_age",
		Description:    "A lord's demesne with serfs and crop rotation. +1.60 food/tick (6 workers).",
		LineageKey:     "food", LineageTier: 5,
		WorkerDomain: "food", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "food",
	})
	// tier 6 — renaissance_age  rate=3.20
	b = append(b, BuildingDef{
		Name: "Market Garden", Key: "market_garden", Category: "production",
		BaseCost:       map[string]float64{"gold": 600000, "steel": 200000, "knowledge": 100000},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 3.20}},
		BuildTicks:     12000,
		RequiredAge:    "renaissance_age",
		Description:    "Scientific farming and market gardens. +3.20 food/tick (7 workers).",
		LineageKey:     "food", LineageTier: 6,
		WorkerDomain: "food", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "food",
	})
	// tier 7 — colonial_age  rate=6.40
	b = append(b, BuildingDef{
		Name: "Plantation", Key: "plantation", Category: "production",
		BaseCost:       map[string]float64{"gold": 3.5e6, "steel": 1.5e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 6.40}},
		BuildTicks:     18000,
		RequiredAge:    "colonial_age",
		Description:    "Large-scale colonial plantation. +6.40 food/tick (8 workers).",
		LineageKey:     "food", LineageTier: 7,
		WorkerDomain: "food", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "food",
	})
	// tier 8 — industrial_age  rate=12.80
	b = append(b, BuildingDef{
		Name: "Agricultural Works", Key: "agricultural_works", Category: "production",
		BaseCost:       map[string]float64{"steel": 25e6, "coal": 10e6, "gold": 15e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 12.80}},
		BuildTicks:     25000,
		RequiredAge:    "industrial_age",
		Description:    "Industrial-scale agricultural works. +12.80 food/tick (10 workers).",
		LineageKey:     "food", LineageTier: 8,
		WorkerDomain: "food", WorkerCapacity: 10,
		EpochKey: "steel_era", OutputResource: "food",
	})
	// tier 9 — victorian_age  rate=25.60
	b = append(b, BuildingDef{
		Name: "Mechanized Farm", Key: "mechanized_farm", Category: "production",
		BaseCost:       map[string]float64{"steel": 180e6, "oil": 80e6, "gold": 100e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 25.60}},
		BuildTicks:     50000,
		RequiredAge:    "victorian_age",
		Description:    "Steam and oil-powered mechanized farming. +25.60 food/tick (10 workers).",
		LineageKey:     "food", LineageTier: 9,
		WorkerDomain: "food", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "food",
	})
	// tier 10 — electric_age  rate=51.20
	b = append(b, BuildingDef{
		Name: "Industrial Farm", Key: "industrial_farm", Category: "production",
		BaseCost:       map[string]float64{"steel": 1e9, "electricity": 400e6, "oil": 300e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 51.20}},
		BuildTicks:     75000,
		RequiredAge:    "electric_age",
		Description:    "Electrified industrial farming complex. +51.20 food/tick (12 workers).",
		LineageKey:     "food", LineageTier: 10,
		WorkerDomain: "food", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "food",
	})
	// tier 11 — atomic_age  rate=102.40
	b = append(b, BuildingDef{
		Name: "Agricultural Complex", Key: "agricultural_complex", Category: "production",
		BaseCost:       map[string]float64{"steel": 5e9, "electricity": 2e9, "uranium": 500e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 102.40}},
		BuildTicks:     100000,
		RequiredAge:    "atomic_age",
		Description:    "Atomic-age agricultural mega-complex. +102.40 food/tick (12 workers).",
		LineageKey:     "food", LineageTier: 11,
		WorkerDomain: "food", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "food",
	})
	// tier 12 — modern_age  rate=204.80
	b = append(b, BuildingDef{
		Name: "Agri Complex", Key: "agri_complex", Category: "production",
		BaseCost:       map[string]float64{"steel": 30e9, "electricity": 12e9, "data": 1e9},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 204.80}},
		BuildTicks:     150000,
		RequiredAge:    "modern_age",
		Description:    "AI-optimised modern agriculture. +204.80 food/tick (14 workers).",
		LineageKey:     "food", LineageTier: 12,
		WorkerDomain: "food", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "food",
	})
	// tier 13 — information_age  rate=409.60
	b = append(b, BuildingDef{
		Name: "Smart Farm", Key: "smart_farm", Category: "production",
		BaseCost:       map[string]float64{"electricity": 80e9, "data": 8e9, "steel": 150e9},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 409.60}},
		BuildTicks:     300000,
		RequiredAge:    "information_age",
		Description:    "Sensor-driven smart farming. +409.60 food/tick (15 workers).",
		LineageKey:     "food", LineageTier: 13,
		WorkerDomain: "food", WorkerCapacity: 15,
		EpochKey: "digital_era", OutputResource: "food",
	})
	// tier 14 — digital_age  rate=819.20
	b = append(b, BuildingDef{
		Name: "Nano Farm", Key: "nano_farm", Category: "production",
		BaseCost:       map[string]float64{"electricity": 400e9, "data": 50e9, "steel": 600e9},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 819.20}},
		BuildTicks:     500000,
		RequiredAge:    "digital_age",
		Description:    "Nanotechnology-based food synthesis. +819.20 food/tick (16 workers).",
		LineageKey:     "food", LineageTier: 14,
		WorkerDomain: "food", WorkerCapacity: 16,
		EpochKey: "digital_era", OutputResource: "food",
	})
	// tier 15 — cyberpunk_age  rate=1638.40
	b = append(b, BuildingDef{
		Name: "Vat Farm", Key: "vat_farm", Category: "production",
		BaseCost:       map[string]float64{"data": 200e9, "crypto": 1e12, "electricity": 2e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 1638.40}},
		BuildTicks:     1000000,
		RequiredAge:    "cyberpunk_age",
		Description:    "Vat-grown protein synthesis at industrial scale. +1638.40 food/tick (18 workers).",
		LineageKey:     "food", LineageTier: 15,
		WorkerDomain: "food", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "food",
	})
	// tier 16 — fusion_age  rate=3276.80
	b = append(b, BuildingDef{
		Name: "Bio Reactor Farm", Key: "bio_reactor_farm", Category: "production",
		BaseCost:       map[string]float64{"plasma": 5e12, "electricity": 15e12, "steel": 20e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 3276.80}},
		BuildTicks:     1500000,
		RequiredAge:    "fusion_age",
		Description:    "Plasma-powered bio reactor food production. +3276.80 food/tick (20 workers).",
		LineageKey:     "food", LineageTier: 16,
		WorkerDomain: "food", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "food",
	})
	// tier 17 — space_age  rate=6553.60
	b = append(b, BuildingDef{
		Name: "Hydroponic Bay", Key: "hydroponic_bay", Category: "production",
		BaseCost:       map[string]float64{"titanium": 80e12, "plasma": 40e12, "electricity": 100e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 6553.60}},
		BuildTicks:     2000000,
		RequiredAge:    "space_age",
		Description:    "Zero-gravity hydroponic growing bays. +6553.60 food/tick (20 workers).",
		LineageKey:     "food", LineageTier: 17,
		WorkerDomain: "food", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "food",
	})
	// tier 18 — interstellar_age  rate=13107.20
	b = append(b, BuildingDef{
		Name: "Protein Synthesizer", Key: "protein_synthesizer", Category: "production",
		BaseCost:       map[string]float64{"dark_matter": 100e12, "titanium": 800e12, "plasma": 500e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 13107.20}},
		BuildTicks:     2500000,
		RequiredAge:    "interstellar_age",
		Description:    "Matter-to-protein synthesizer. +13107.20 food/tick (25 workers).",
		LineageKey:     "food", LineageTier: 18,
		WorkerDomain: "food", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "food",
	})
	// tier 19 — galactic_age  rate=26214.40
	b = append(b, BuildingDef{
		Name: "Matter Converter", Key: "matter_converter", Category: "production",
		BaseCost:       map[string]float64{"antimatter": 200e12, "dark_matter": 1e15, "titanium": 5e15},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 26214.40}},
		BuildTicks:     3000000,
		RequiredAge:    "galactic_age",
		Description:    "Converts raw matter into any food type. +26214.40 food/tick (25 workers).",
		LineageKey:     "food", LineageTier: 19,
		WorkerDomain: "food", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "food",
	})
	// tier 20 — quantum_age  rate=52428.80
	b = append(b, BuildingDef{
		Name: "Quantum Cultivator", Key: "quantum_cultivator", Category: "production",
		BaseCost:       map[string]float64{"quantum_flux": 200e12, "antimatter": 60e15, "dark_matter": 50e15},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "food", Value: 52428.80}},
		BuildTicks:     5000000,
		RequiredAge:    "quantum_age",
		Description:    "Quantum probability manipulation to grow food. +52428.80 food/tick (30 workers).",
		LineageKey:     "food", LineageTier: 20,
		WorkerDomain: "food", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "food",
	})

	// =========================================================================
	// LINEAGE 3 — ORGANIC EXTRACTION (lineageKey: "organic_extraction", domain: "lumber")
	// rate = 0.04 * 2^tier  CostScale: 1.30  Category: "production"
	// Output transitions: wood(0-5) → coal(6-8) → oil(9-13) → nanobots(14-16) → quantum_flux(17-20)
	// =========================================================================

	// tier 0 — primitive_age  rate=0.04  output=wood
	b = append(b, BuildingDef{
		Name: "Wood Camp", Key: "wood_camp", Category: "production",
		BaseCost:       map[string]float64{"wood": 15},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 0.04}},
		BuildTicks:     80,
		RequiredAge:    "primitive_age",
		Description:    "A basic camp for collecting wood. +0.04 wood/tick (3 workers).",
		LineageKey:     "organic_extraction", LineageTier: 0,
		WorkerDomain: "lumber", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "wood",
	})
	// tier 1 — stone_age  rate=0.08  output=wood
	b = append(b, BuildingDef{
		Name: "Woodcutter Camp", Key: "woodcutter_camp", Category: "production",
		BaseCost:       map[string]float64{"wood": 120, "stone": 60},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 0.08}},
		BuildTicks:     200,
		RequiredAge:    "stone_age",
		Description:    "Choppers fell trees with stone axes. +0.08 wood/tick (4 workers).",
		LineageKey:     "organic_extraction", LineageTier: 1,
		WorkerDomain: "lumber", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "wood",
	})
	// tier 2 — bronze_age  rate=0.16  output=wood
	b = append(b, BuildingDef{
		Name: "Lumber Mill", Key: "lumber_mill", Category: "production",
		BaseCost:       map[string]float64{"wood": 700, "stone": 400, "iron": 150},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 0.16}},
		BuildTicks:     500,
		RequiredAge:    "bronze_age",
		Description:    "Bronze-saw lumber processing. +0.16 wood/tick (5 workers).",
		LineageKey:     "organic_extraction", LineageTier: 2,
		WorkerDomain: "lumber", WorkerCapacity: 5,
		EpochKey: "stone_era", OutputResource: "wood",
	})
	// tier 3 — iron_age  rate=0.32  output=wood
	b = append(b, BuildingDef{
		Name: "Timber Yard", Key: "timber_yard", Category: "production",
		BaseCost:       map[string]float64{"stone": 4000, "iron": 2500, "wood": 3000},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 0.32}},
		BuildTicks:     1000,
		RequiredAge:    "iron_age",
		Description:    "Iron-saw timber processing yard. +0.32 wood/tick (5 workers).",
		LineageKey:     "organic_extraction", LineageTier: 3,
		WorkerDomain: "lumber", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "wood",
	})
	// tier 4 — classical_age  rate=0.64  output=wood
	b = append(b, BuildingDef{
		Name: "Wood Workshop", Key: "wood_workshop", Category: "production",
		BaseCost:       map[string]float64{"stone": 30000, "gold": 10000, "iron": 8000},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 0.64}},
		BuildTicks:     3000,
		RequiredAge:    "classical_age",
		Description:    "Skilled carpenters maximise wood output. +0.64 wood/tick (6 workers).",
		LineageKey:     "organic_extraction", LineageTier: 4,
		WorkerDomain: "lumber", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "wood",
	})
	// tier 5 — medieval_age  rate=1.28  output=wood
	b = append(b, BuildingDef{
		Name: "Sawmill", Key: "sawmill", Category: "production",
		BaseCost:       map[string]float64{"stone": 160000, "gold": 55000, "knowledge": 15000},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 1.28}},
		BuildTicks:     6000,
		RequiredAge:    "medieval_age",
		Description:    "Water-wheel-powered sawmill. +1.28 wood/tick (6 workers).",
		LineageKey:     "organic_extraction", LineageTier: 5,
		WorkerDomain: "lumber", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "wood",
	})
	// tier 6 — renaissance_age  rate=2.56  output=coal
	b = append(b, BuildingDef{
		Name: "Coal Mine", Key: "coal_mine", Category: "production",
		BaseCost:       map[string]float64{"gold": 550000, "steel": 200000, "knowledge": 80000},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "coal", Value: 2.56}},
		BuildTicks:     12000,
		RequiredAge:    "renaissance_age",
		Description:    "Early coal extraction for industry. +2.56 coal/tick (7 workers).",
		LineageKey:     "organic_extraction", LineageTier: 6,
		WorkerDomain: "lumber", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 7 — colonial_age  rate=5.12  output=coal
	b = append(b, BuildingDef{
		Name: "Coal Works", Key: "coal_works", Category: "production",
		BaseCost:       map[string]float64{"gold": 3e6, "steel": 1.5e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "coal", Value: 5.12}},
		BuildTicks:     18000,
		RequiredAge:    "colonial_age",
		Description:    "Organised coal extraction and processing. +5.12 coal/tick (8 workers).",
		LineageKey:     "organic_extraction", LineageTier: 7,
		WorkerDomain: "lumber", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 8 — industrial_age  rate=10.24  output=coal
	b = append(b, BuildingDef{
		Name: "Steam Coal Plant", Key: "steam_coal_plant", Category: "production",
		BaseCost:       map[string]float64{"steel": 22e6, "coal": 8e6, "gold": 12e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "coal", Value: 10.24}},
		BuildTicks:     25000,
		RequiredAge:    "industrial_age",
		Description:    "Steam-powered coal extraction plant. +10.24 coal/tick (10 workers).",
		LineageKey:     "organic_extraction", LineageTier: 8,
		WorkerDomain: "lumber", WorkerCapacity: 10,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 9 — victorian_age  rate=20.48  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Derrick", Key: "oil_derrick", Category: "production",
		BaseCost:       map[string]float64{"steel": 170e6, "iron": 80e6, "gold": 100e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "oil", Value: 20.48}},
		BuildTicks:     50000,
		RequiredAge:    "victorian_age",
		Description:    "Early oil extraction derrick. +20.48 oil/tick (10 workers).",
		LineageKey:     "organic_extraction", LineageTier: 9,
		WorkerDomain: "lumber", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "oil",
	})
	// tier 10 — electric_age  rate=40.96  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Field", Key: "oil_field", Category: "production",
		BaseCost:       map[string]float64{"steel": 900e6, "electricity": 350e6, "oil": 100e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "oil", Value: 40.96}},
		BuildTicks:     75000,
		RequiredAge:    "electric_age",
		Description:    "Electrified oil field operations. +40.96 oil/tick (12 workers).",
		LineageKey:     "organic_extraction", LineageTier: 10,
		WorkerDomain: "lumber", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "oil",
	})
	// tier 11 — atomic_age  rate=81.92  output=oil
	b = append(b, BuildingDef{
		Name: "Petroleum Refinery", Key: "petroleum_refinery", Category: "production",
		BaseCost:       map[string]float64{"steel": 5e9, "electricity": 2e9, "uranium": 300e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "oil", Value: 81.92}},
		BuildTicks:     100000,
		RequiredAge:    "atomic_age",
		Description:    "Advanced petroleum refinery. +81.92 oil/tick (12 workers).",
		LineageKey:     "organic_extraction", LineageTier: 11,
		WorkerDomain: "lumber", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "oil",
	})
	// tier 12 — modern_age  rate=163.84  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Platform", Key: "oil_platform", Category: "production",
		BaseCost:       map[string]float64{"steel": 28e9, "electricity": 10e9, "data": 800e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "oil", Value: 163.84}},
		BuildTicks:     150000,
		RequiredAge:    "modern_age",
		Description:    "Offshore AI-monitored oil platform. +163.84 oil/tick (14 workers).",
		LineageKey:     "organic_extraction", LineageTier: 12,
		WorkerDomain: "lumber", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "oil",
	})
	// tier 13 — information_age  rate=327.68  output=oil
	b = append(b, BuildingDef{
		Name: "Smart Refinery", Key: "smart_refinery", Category: "production",
		BaseCost:       map[string]float64{"electricity": 75e9, "data": 7e9, "steel": 140e9},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "oil", Value: 327.68}},
		BuildTicks:     300000,
		RequiredAge:    "information_age",
		Description:    "AI-optimised smart petroleum refinery. +327.68 oil/tick (15 workers).",
		LineageKey:     "organic_extraction", LineageTier: 13,
		WorkerDomain: "lumber", WorkerCapacity: 15,
		EpochKey: "digital_era", OutputResource: "oil",
	})
	// tier 14 — digital_age  rate=655.36  output=nanobots
	b = append(b, BuildingDef{
		Name: "Bio Fabrication Lab", Key: "bio_fabrication_lab", Category: "production",
		BaseCost:       map[string]float64{"electricity": 380e9, "data": 45e9, "steel": 550e9},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "nanobots", Value: 655.36}},
		BuildTicks:     500000,
		RequiredAge:    "digital_age",
		Description:    "Digital-biological nanofabrication. +655.36 nanobots/tick (16 workers).",
		LineageKey:     "organic_extraction", LineageTier: 14,
		WorkerDomain: "lumber", WorkerCapacity: 16,
		EpochKey: "digital_era", OutputResource: "nanobots",
	})
	// tier 15 — cyberpunk_age  rate=1310.72  output=nanobots
	b = append(b, BuildingDef{
		Name: "Nanobot Vat", Key: "nanobot_vat", Category: "production",
		BaseCost:       map[string]float64{"data": 180e9, "crypto": 1.2e12, "electricity": 2.5e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "nanobots", Value: 1310.72}},
		BuildTicks:     1000000,
		RequiredAge:    "cyberpunk_age",
		Description:    "Vat-grown nanobot manufacturing. +1310.72 nanobots/tick (18 workers).",
		LineageKey:     "organic_extraction", LineageTier: 15,
		WorkerDomain: "lumber", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "nanobots",
	})
	// tier 16 — fusion_age  rate=2621.44  output=nanobots
	b = append(b, BuildingDef{
		Name: "Molecular Synthesizer", Key: "molecular_synthesizer", Category: "production",
		BaseCost:       map[string]float64{"plasma": 4e12, "electricity": 14e12, "steel": 18e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "nanobots", Value: 2621.44}},
		BuildTicks:     1500000,
		RequiredAge:    "fusion_age",
		Description:    "Plasma-powered molecular synthesis. +2621.44 nanobots/tick (20 workers).",
		LineageKey:     "organic_extraction", LineageTier: 16,
		WorkerDomain: "lumber", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "nanobots",
	})
	// tier 17 — space_age  rate=5242.88  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Quantum Organic Extractor", Key: "quantum_organic_extractor", Category: "production",
		BaseCost:       map[string]float64{"titanium": 75e12, "plasma": 35e12, "electricity": 90e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "quantum_flux", Value: 5242.88}},
		BuildTicks:     2000000,
		RequiredAge:    "space_age",
		Description:    "Quantum-state organic matter extraction. +5242.88 quantum_flux/tick (20 workers).",
		LineageKey:     "organic_extraction", LineageTier: 17,
		WorkerDomain: "lumber", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "quantum_flux",
	})
	// tier 18 — interstellar_age  rate=10485.76  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Reality Matter Weaver", Key: "reality_matter_weaver", Category: "production",
		BaseCost:       map[string]float64{"dark_matter": 90e12, "titanium": 750e12, "plasma": 450e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "quantum_flux", Value: 10485.76}},
		BuildTicks:     2500000,
		RequiredAge:    "interstellar_age",
		Description:    "Weaves reality matter into quantum flux. +10485.76 quantum_flux/tick (25 workers).",
		LineageKey:     "organic_extraction", LineageTier: 18,
		WorkerDomain: "lumber", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})
	// tier 19 — galactic_age  rate=20971.52  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Cosmic Organic Works", Key: "cosmic_organic_works", Category: "production",
		BaseCost:       map[string]float64{"antimatter": 180e12, "dark_matter": 900e12, "titanium": 4.5e15},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "quantum_flux", Value: 20971.52}},
		BuildTicks:     3000000,
		RequiredAge:    "galactic_age",
		Description:    "Galactic-scale cosmic organic works. +20971.52 quantum_flux/tick (25 workers).",
		LineageKey:     "organic_extraction", LineageTier: 19,
		WorkerDomain: "lumber", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})
	// tier 20 — quantum_age  rate=41943.04  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Reality Harvester", Key: "reality_harvester", Category: "production",
		BaseCost:       map[string]float64{"quantum_flux": 180e12, "antimatter": 55e15, "dark_matter": 45e15},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "quantum_flux", Value: 41943.04}},
		BuildTicks:     5000000,
		RequiredAge:    "quantum_age",
		Description:    "Harvests raw quantum flux from reality itself. +41943.04 quantum_flux/tick (30 workers).",
		LineageKey:     "organic_extraction", LineageTier: 20,
		WorkerDomain: "lumber", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	// =========================================================================
	// LINEAGE 4 — GEOLOGICAL EXTRACTION (lineageKey: "geological_extraction", domain: "masonry")
	// rate = 0.03 * 2^tier  CostScale: 1.30  Category: "production"
	// Output: stone(0-2) → dual marble+iron_ore(3-4) → iron_ore(5-8) → uranium(9-11)
	//         → titanium_ore(12-14) → dark_matter_crystals(15-17) → antimatter(18-20)
	// Dual-output tiers (3-4): two Effects each at half rate
	// =========================================================================

	// tier 0 — primitive_age  rate=0.03  output=stone
	b = append(b, BuildingDef{
		Name: "Stone Camp", Key: "stone_camp", Category: "production",
		BaseCost:       map[string]float64{"wood": 15},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "stone", Value: 0.03}},
		BuildTicks:     80,
		RequiredAge:    "primitive_age",
		Description:    "A basic camp for gathering stones. +0.03 stone/tick (3 workers).",
		LineageKey:     "geological_extraction", LineageTier: 0,
		WorkerDomain: "masonry", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "stone",
	})
	// tier 1 — stone_age  rate=0.06  output=stone
	b = append(b, BuildingDef{
		Name: "Stone Pit", Key: "stone_pit", Category: "production",
		BaseCost:       map[string]float64{"wood": 100, "stone": 60},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "stone", Value: 0.06}},
		BuildTicks:     200,
		RequiredAge:    "stone_age",
		Description:    "A shallow pit dug for stone. +0.06 stone/tick (4 workers).",
		LineageKey:     "geological_extraction", LineageTier: 1,
		WorkerDomain: "masonry", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "stone",
	})
	// tier 2 — bronze_age  rate=0.12  output=stone
	b = append(b, BuildingDef{
		Name: "Quarry", Key: "quarry", Category: "production",
		BaseCost:       map[string]float64{"wood": 600, "stone": 400, "iron": 100},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "stone", Value: 0.12}},
		BuildTicks:     500,
		RequiredAge:    "bronze_age",
		Description:    "Organised stone quarrying. +0.12 stone/tick (5 workers).",
		LineageKey:     "geological_extraction", LineageTier: 2,
		WorkerDomain: "masonry", WorkerCapacity: 5,
		EpochKey: "stone_era", OutputResource: "stone",
	})
	// tier 3 — iron_age  rate=0.24  dual: marble(0.12) + iron_ore(0.12)
	b = append(b, BuildingDef{
		Name: "Marble Quarry", Key: "marble_quarry", Category: "production",
		BaseCost:       map[string]float64{"stone": 4500, "iron": 2000},
		CostScale:      1.30,
		Effects: []Effect{
			{Type: "production", Target: "marble", Value: 0.12},
			{Type: "production", Target: "iron_ore", Value: 0.12},
		},
		BuildTicks:  1000,
		RequiredAge: "iron_age",
		Description: "Marble and iron ore dual extraction. +0.12 marble, +0.12 iron_ore/tick (5 workers).",
		LineageKey:  "geological_extraction", LineageTier: 3,
		WorkerDomain: "masonry", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "stone",
	})
	// tier 4 — classical_age  rate=0.48  dual: marble(0.24) + iron_ore(0.24)
	b = append(b, BuildingDef{
		Name: "Marble Works", Key: "marble_works", Category: "production",
		BaseCost:       map[string]float64{"stone": 28000, "gold": 10000, "iron": 8000},
		CostScale:      1.30,
		Effects: []Effect{
			{Type: "production", Target: "marble", Value: 0.24},
			{Type: "production", Target: "iron_ore", Value: 0.24},
		},
		BuildTicks:  3000,
		RequiredAge: "classical_age",
		Description: "Classical marble and ore works. +0.24 marble, +0.24 iron_ore/tick (6 workers).",
		LineageKey:  "geological_extraction", LineageTier: 4,
		WorkerDomain: "masonry", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "stone",
	})
	// tier 5 — medieval_age  rate=0.96  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Stonemason's Guild", Key: "stonemasons_guild", Category: "production",
		BaseCost:       map[string]float64{"stone": 160000, "gold": 50000, "knowledge": 15000},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "iron_ore", Value: 0.96}},
		BuildTicks:     6000,
		RequiredAge:    "medieval_age",
		Description:    "Guild of stonemasons extracting iron ore. +0.96 iron_ore/tick (6 workers).",
		LineageKey:     "geological_extraction", LineageTier: 5,
		WorkerDomain: "masonry", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "iron_ore",
	})
	// tier 6 — renaissance_age  rate=1.92  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Iron Mine", Key: "iron_mine", Category: "production",
		BaseCost:       map[string]float64{"gold": 500000, "steel": 180000, "knowledge": 75000},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "iron_ore", Value: 1.92}},
		BuildTicks:     12000,
		RequiredAge:    "renaissance_age",
		Description:    "Deep iron ore mining. +1.92 iron_ore/tick (7 workers).",
		LineageKey:     "geological_extraction", LineageTier: 6,
		WorkerDomain: "masonry", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 7 — colonial_age  rate=3.84  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Deep Iron Mine", Key: "deep_iron_mine", Category: "production",
		BaseCost:       map[string]float64{"gold": 3e6, "steel": 1.8e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "iron_ore", Value: 3.84}},
		BuildTicks:     18000,
		RequiredAge:    "colonial_age",
		Description:    "Colonial deep iron ore extraction. +3.84 iron_ore/tick (8 workers).",
		LineageKey:     "geological_extraction", LineageTier: 7,
		WorkerDomain: "masonry", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 8 — industrial_age  rate=7.68  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Steam Mine", Key: "steam_mine", Category: "production",
		BaseCost:       map[string]float64{"steel": 22e6, "coal": 8e6, "gold": 12e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "iron_ore", Value: 7.68}},
		BuildTicks:     25000,
		RequiredAge:    "industrial_age",
		Description:    "Steam-powered industrial iron ore mine. +7.68 iron_ore/tick (10 workers).",
		LineageKey:     "geological_extraction", LineageTier: 8,
		WorkerDomain: "masonry", WorkerCapacity: 10,
		EpochKey: "steel_era", OutputResource: "iron_ore",
	})
	// tier 9 — victorian_age  rate=15.36  output=uranium
	b = append(b, BuildingDef{
		Name: "Uranium Mine", Key: "uranium_mine", Category: "production",
		BaseCost:       map[string]float64{"steel": 175e6, "iron": 80e6, "gold": 100e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "uranium", Value: 15.36}},
		BuildTicks:     50000,
		RequiredAge:    "victorian_age",
		Description:    "Early uranium ore extraction. +15.36 uranium/tick (10 workers).",
		LineageKey:     "geological_extraction", LineageTier: 9,
		WorkerDomain: "masonry", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 10 — electric_age  rate=30.72  output=uranium
	b = append(b, BuildingDef{
		Name: "Nuclear Extraction Plant", Key: "nuclear_extraction_plant", Category: "production",
		BaseCost:       map[string]float64{"steel": 950e6, "electricity": 380e6, "uranium": 50e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "uranium", Value: 30.72}},
		BuildTicks:     75000,
		RequiredAge:    "electric_age",
		Description:    "High-tech nuclear material extraction. +30.72 uranium/tick (12 workers).",
		LineageKey:     "geological_extraction", LineageTier: 10,
		WorkerDomain: "masonry", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 11 — atomic_age  rate=61.44  output=uranium
	b = append(b, BuildingDef{
		Name: "Uranium Processing Works", Key: "uranium_processing_works", Category: "production",
		BaseCost:       map[string]float64{"steel": 5e9, "electricity": 2e9, "uranium": 400e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "uranium", Value: 61.44}},
		BuildTicks:     100000,
		RequiredAge:    "atomic_age",
		Description:    "Industrial uranium processing. +61.44 uranium/tick (12 workers).",
		LineageKey:     "geological_extraction", LineageTier: 11,
		WorkerDomain: "masonry", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "uranium",
	})
	// tier 12 — modern_age  rate=122.88  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Titanium Mine", Key: "titanium_mine", Category: "production",
		BaseCost:       map[string]float64{"steel": 28e9, "electricity": 10e9, "data": 800e6},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "titanium_ore", Value: 122.88}},
		BuildTicks:     150000,
		RequiredAge:    "modern_age",
		Description:    "High-precision titanium ore extraction. +122.88 titanium_ore/tick (14 workers).",
		LineageKey:     "geological_extraction", LineageTier: 12,
		WorkerDomain: "masonry", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 13 — information_age  rate=245.76  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Precision Mine", Key: "precision_mine", Category: "production",
		BaseCost:       map[string]float64{"electricity": 70e9, "data": 7e9, "steel": 130e9},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "titanium_ore", Value: 245.76}},
		BuildTicks:     300000,
		RequiredAge:    "information_age",
		Description:    "AI-guided precision titanium mining. +245.76 titanium_ore/tick (15 workers).",
		LineageKey:     "geological_extraction", LineageTier: 13,
		WorkerDomain: "masonry", WorkerCapacity: 15,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 14 — digital_age  rate=491.52  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Nano Drill Complex", Key: "nano_drill_complex", Category: "production",
		BaseCost:       map[string]float64{"electricity": 360e9, "data": 40e9, "steel": 520e9},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "titanium_ore", Value: 491.52}},
		BuildTicks:     500000,
		RequiredAge:    "digital_age",
		Description:    "Nanoscale drilling for titanium ore. +491.52 titanium_ore/tick (16 workers).",
		LineageKey:     "geological_extraction", LineageTier: 14,
		WorkerDomain: "masonry", WorkerCapacity: 16,
		EpochKey: "digital_era", OutputResource: "titanium_ore",
	})
	// tier 15 — cyberpunk_age  rate=983.04  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Dark Crystal Mine", Key: "dark_crystal_mine", Category: "production",
		BaseCost:       map[string]float64{"data": 170e9, "crypto": 1.1e12, "electricity": 2.2e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 983.04}},
		BuildTicks:     1000000,
		RequiredAge:    "cyberpunk_age",
		Description:    "Extraction of dark matter crystals. +983.04 dark_matter_crystals/tick (18 workers).",
		LineageKey:     "geological_extraction", LineageTier: 15,
		WorkerDomain: "masonry", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 16 — fusion_age  rate=1966.08  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Exotic Mineral Extractor", Key: "exotic_mineral_extractor", Category: "production",
		BaseCost:       map[string]float64{"plasma": 4.5e12, "electricity": 13e12, "steel": 17e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 1966.08}},
		BuildTicks:     1500000,
		RequiredAge:    "fusion_age",
		Description:    "Plasma-assisted exotic mineral extraction. +1966.08 dark_matter_crystals/tick (20 workers).",
		LineageKey:     "geological_extraction", LineageTier: 16,
		WorkerDomain: "masonry", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 17 — space_age  rate=3932.16  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Asteroid Crystal Mine", Key: "asteroid_crystal_mine", Category: "production",
		BaseCost:       map[string]float64{"titanium": 70e12, "plasma": 32e12, "electricity": 85e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 3932.16}},
		BuildTicks:     2000000,
		RequiredAge:    "space_age",
		Description:    "Asteroid belt dark matter crystal mining. +3932.16 dark_matter_crystals/tick (20 workers).",
		LineageKey:     "geological_extraction", LineageTier: 17,
		WorkerDomain: "masonry", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "dark_matter_crystals",
	})
	// tier 18 — interstellar_age  rate=7864.32  output=antimatter
	b = append(b, BuildingDef{
		Name: "Stellar Core Drill", Key: "stellar_core_drill", Category: "production",
		BaseCost:       map[string]float64{"dark_matter": 85e12, "titanium": 700e12, "plasma": 420e12},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "antimatter", Value: 7864.32}},
		BuildTicks:     2500000,
		RequiredAge:    "interstellar_age",
		Description:    "Drills into stellar cores for antimatter. +7864.32 antimatter/tick (25 workers).",
		LineageKey:     "geological_extraction", LineageTier: 18,
		WorkerDomain: "masonry", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 19 — galactic_age  rate=15728.64  output=antimatter
	b = append(b, BuildingDef{
		Name: "Neutron Star Mine", Key: "neutron_star_mine", Category: "production",
		BaseCost:       map[string]float64{"antimatter": 170e12, "dark_matter": 850e12, "titanium": 4.2e15},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "antimatter", Value: 15728.64}},
		BuildTicks:     3000000,
		RequiredAge:    "galactic_age",
		Description:    "Mining neutron stars for antimatter. +15728.64 antimatter/tick (25 workers).",
		LineageKey:     "geological_extraction", LineageTier: 19,
		WorkerDomain: "masonry", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 20 — quantum_age  rate=31457.28  output=antimatter
	b = append(b, BuildingDef{
		Name: "Reality Excavator", Key: "reality_excavator", Category: "production",
		BaseCost:       map[string]float64{"quantum_flux": 160e12, "antimatter": 48e15, "dark_matter": 38e15},
		CostScale:      1.30,
		Effects:        []Effect{{Type: "production", Target: "antimatter", Value: 31457.28}},
		BuildTicks:     5000000,
		RequiredAge:    "quantum_age",
		Description:    "Excavates antimatter from the quantum foam. +31457.28 antimatter/tick (30 workers).",
		LineageKey:     "geological_extraction", LineageTier: 20,
		WorkerDomain: "masonry", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})

	return b
}
