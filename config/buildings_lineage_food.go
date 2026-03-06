package config

// buildingsLineageFood returns all production, housing, research, and military buildings
// for the 13 lineage chains introduced in Phase 10 of the economy redesign.
// Storage buildings and wonders remain in baseBuildingsRaw().
// All fields are set inline; no separate buildingMeta() entries needed for these buildings.
func buildingsLineageFood() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 2 — FOOD (lineageKey: "food", domain: "food", output: "food")
	// rate = 0.05 * 2^tier  CostScale: 1.18 (tier 0), 1.30 (tiers 1+)  Category: "production"
	// =========================================================================

	// tier 0 — primitive_age  rate=0.50
	// Rate raised from 0.05 → 0.50 so gathering camps can sustain early workers.
	// BuildTicks lowered from 40 → 12 so food production comes online before starvation.
	b = append(b, BuildingDef{
		Name: "Gathering Camp", Key: "gathering_camp", Category: "production",
		BaseCost:    map[string]float64{"wood": 20},
		CostScale:   1.18,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 0.50}},
		BuildTicks:  12,
		RequiredAge: "primitive_age",
		Description: "Foragers gather berries and roots. +0.50 food/tick (3 workers).",
		LineageKey:  "food", LineageTier: 0,
		WorkerDomain: "food", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "food",
	})
	// tier 1 — stone_age  rate=1.00
	// Rate raised from 0.10 → 1.00 to maintain ×2 per tier vs gathering_camp (0.50).
	// BuildTicks lowered from 100 → 30 proportional to gathering_camp reduction.
	b = append(b, BuildingDef{
		Name: "Forager Post", Key: "forager_post", Category: "production",
		BaseCost:    map[string]float64{"wood": 150, "stone": 80},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 1.00}},
		BuildTicks:  30,
		RequiredAge: "stone_age",
		Description: "Organised foraging post. +1.00 food/tick (4 workers).",
		LineageKey:  "food", LineageTier: 1,
		WorkerDomain: "food", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "food",
	})
	// tier 2 — bronze_age  rate=2.00
	b = append(b, BuildingDef{
		Name: "Farm", Key: "farm", Category: "production",
		BaseCost:    map[string]float64{"wood": 800, "stone": 500},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 2.00}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "Cultivated fields produce steady food. +2.00 food/tick (5 workers).",
		LineageKey:  "food", LineageTier: 2,
		WorkerDomain: "food", WorkerCapacity: 5,
		EpochKey: "stone_era", OutputResource: "food",
	})
	// tier 3 — iron_age  rate=4.00
	b = append(b, BuildingDef{
		Name: "Field Works", Key: "field_works", Category: "production",
		BaseCost:    map[string]float64{"stone": 5000, "iron": 2000, "wood": 3000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 4.00}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "Iron-tool farming with irrigation. +4.00 food/tick (5 workers).",
		LineageKey:  "food", LineageTier: 3,
		WorkerDomain: "food", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "food",
	})
	// tier 4 — classical_age  rate=8.00
	b = append(b, BuildingDef{
		Name: "Estate Farm", Key: "estate_farm", Category: "production",
		BaseCost:    map[string]float64{"stone": 35000, "gold": 12000, "iron": 10000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 8.00}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "A large estate with managed farmlands. +8.00 food/tick (6 workers).",
		LineageKey:  "food", LineageTier: 4,
		WorkerDomain: "food", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "food",
	})
	// tier 5 — medieval_age  rate=16.00
	b = append(b, BuildingDef{
		Name: "Demesne", Key: "demesne", Category: "production",
		BaseCost:    map[string]float64{"stone": 180000, "gold": 60000, "knowledge": 20000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 16.00}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "A lord's demesne with serfs and crop rotation. +16.00 food/tick (6 workers).",
		LineageKey:  "food", LineageTier: 5,
		WorkerDomain: "food", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "food",
	})
	// tier 6 — renaissance_age  rate=3.20
	b = append(b, BuildingDef{
		Name: "Market Garden", Key: "market_garden", Category: "production",
		BaseCost:    map[string]float64{"gold": 600000, "steel": 200000, "knowledge": 100000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 3.20}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "Scientific farming and market gardens. +3.20 food/tick (7 workers).",
		LineageKey:  "food", LineageTier: 6,
		WorkerDomain: "food", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "food",
	})
	// tier 7 — colonial_age  rate=6.40
	b = append(b, BuildingDef{
		Name: "Plantation", Key: "plantation", Category: "production",
		BaseCost:    map[string]float64{"gold": 3.5e6, "steel": 1.5e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 6.40}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "Large-scale colonial plantation. +6.40 food/tick (8 workers).",
		LineageKey:  "food", LineageTier: 7,
		WorkerDomain: "food", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "food",
	})
	// tier 8 — industrial_age  rate=12.80
	b = append(b, BuildingDef{
		Name: "Agricultural Works", Key: "agricultural_works", Category: "production",
		BaseCost:    map[string]float64{"steel": 25e6, "coal": 10e6, "gold": 15e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 12.80}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Industrial-scale agricultural works. +12.80 food/tick (10 workers).",
		LineageKey:  "food", LineageTier: 8,
		WorkerDomain: "food", WorkerCapacity: 10,
		EpochKey: "steel_era", OutputResource: "food",
	})
	// tier 9 — victorian_age  rate=25.60
	b = append(b, BuildingDef{
		Name: "Mechanized Farm", Key: "mechanized_farm", Category: "production",
		BaseCost:    map[string]float64{"steel": 180e6, "oil": 80e6, "gold": 100e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 25.60}},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "Steam and oil-powered mechanized farming. +25.60 food/tick (10 workers).",
		LineageKey:  "food", LineageTier: 9,
		WorkerDomain: "food", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "food",
	})
	// tier 10 — electric_age  rate=51.20
	b = append(b, BuildingDef{
		Name: "Industrial Farm", Key: "industrial_farm", Category: "production",
		BaseCost:    map[string]float64{"steel": 1e9, "electricity": 400e6, "oil": 300e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 51.20}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "Electrified industrial farming complex. +51.20 food/tick (12 workers).",
		LineageKey:  "food", LineageTier: 10,
		WorkerDomain: "food", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "food",
	})
	// tier 11 — atomic_age  rate=102.40
	b = append(b, BuildingDef{
		Name: "Agricultural Complex", Key: "agricultural_complex", Category: "production",
		BaseCost:    map[string]float64{"steel": 5e9, "electricity": 2e9, "uranium": 500e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 102.40}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Atomic-age agricultural mega-complex. +102.40 food/tick (12 workers).",
		LineageKey:  "food", LineageTier: 11,
		WorkerDomain: "food", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "food",
	})
	// tier 12 — modern_age  rate=204.80
	b = append(b, BuildingDef{
		Name: "Agri Complex", Key: "agri_complex", Category: "production",
		BaseCost:    map[string]float64{"steel": 30e9, "electricity": 12e9, "data": 1e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 204.80}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "AI-optimised modern agriculture. +204.80 food/tick (14 workers).",
		LineageKey:  "food", LineageTier: 12,
		WorkerDomain: "food", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "food",
	})
	// tier 13 — information_age  rate=409.60
	b = append(b, BuildingDef{
		Name: "Smart Farm", Key: "smart_farm", Category: "production",
		BaseCost:    map[string]float64{"electricity": 80e9, "data": 8e9, "steel": 150e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 409.60}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "Sensor-driven smart farming. +409.60 food/tick (15 workers).",
		LineageKey:  "food", LineageTier: 13,
		WorkerDomain: "food", WorkerCapacity: 15,
		EpochKey: "digital_era", OutputResource: "food",
	})
	// tier 14 — digital_age  rate=819.20
	b = append(b, BuildingDef{
		Name: "Nano Farm", Key: "nano_farm", Category: "production",
		BaseCost:    map[string]float64{"electricity": 400e9, "data": 50e9, "steel": 600e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 819.20}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Nanotechnology-based food synthesis. +819.20 food/tick (16 workers).",
		LineageKey:  "food", LineageTier: 14,
		WorkerDomain: "food", WorkerCapacity: 16,
		EpochKey: "digital_era", OutputResource: "food",
	})
	// tier 15 — cyberpunk_age  rate=1638.40
	b = append(b, BuildingDef{
		Name: "Vat Farm", Key: "vat_farm", Category: "production",
		BaseCost:    map[string]float64{"data": 200e9, "crypto": 1e12, "electricity": 2e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 1638.40}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Vat-grown protein synthesis at industrial scale. +1638.40 food/tick (18 workers).",
		LineageKey:  "food", LineageTier: 15,
		WorkerDomain: "food", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "food",
	})
	// tier 16 — fusion_age  rate=3276.80
	b = append(b, BuildingDef{
		Name: "Bio Reactor Farm", Key: "bio_reactor_farm", Category: "production",
		BaseCost:    map[string]float64{"plasma": 5e12, "electricity": 15e12, "steel": 20e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 3276.80}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Plasma-powered bio reactor food production. +3276.80 food/tick (20 workers).",
		LineageKey:  "food", LineageTier: 16,
		WorkerDomain: "food", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "food",
	})
	// tier 17 — space_age  rate=6553.60
	b = append(b, BuildingDef{
		Name: "Hydroponic Bay", Key: "hydroponic_bay", Category: "production",
		BaseCost:    map[string]float64{"titanium": 80e12, "plasma": 40e12, "electricity": 100e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 6553.60}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Zero-gravity hydroponic growing bays. +6553.60 food/tick (20 workers).",
		LineageKey:  "food", LineageTier: 17,
		WorkerDomain: "food", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "food",
	})
	// tier 18 — interstellar_age  rate=13107.20
	b = append(b, BuildingDef{
		Name: "Protein Synthesizer", Key: "protein_synthesizer", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 100e12, "titanium": 800e12, "plasma": 500e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 13107.20}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Matter-to-protein synthesizer. +13107.20 food/tick (25 workers).",
		LineageKey:  "food", LineageTier: 18,
		WorkerDomain: "food", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "food",
	})
	// tier 19 — galactic_age  rate=26214.40
	b = append(b, BuildingDef{
		Name: "Matter Converter", Key: "matter_converter", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 200e12, "dark_matter": 1e15, "titanium": 5e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 26214.40}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Converts raw matter into any food type. +26214.40 food/tick (25 workers).",
		LineageKey:  "food", LineageTier: 19,
		WorkerDomain: "food", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "food",
	})
	// tier 20 — quantum_age  rate=52428.80
	b = append(b, BuildingDef{
		Name: "Quantum Cultivator", Key: "quantum_cultivator", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 200e12, "antimatter": 60e15, "dark_matter": 50e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "food", Value: 52428.80}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Quantum probability manipulation to grow food. +52428.80 food/tick (30 workers).",
		LineageKey:  "food", LineageTier: 20,
		WorkerDomain: "food", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "food",
	})

	return b
}
