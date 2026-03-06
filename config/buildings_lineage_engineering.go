package config

// buildingsLineageEngineering returns lineages 5-9:
// knowledge, faith, military, trade, engineering.
// Merged into newProductionBuildings() via init — see buildings_new_merge.go.
func buildingsLineageEngineering() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 9 — ENGINEERING (lineageKey: "engineering", domain: "engineering")
	// starts at bronze_age (tier 0)
	// CostScale: 1.35  Category: "production"
	// Output transitions: iron(0-3) → steel(4-7) → electricity(8-12) → plasma+electricity(13-14)
	//                     → dark_matter(15+)
	// =========================================================================

	// tier 0 — bronze_age  output=iron  rate≈0.10
	b = append(b, BuildingDef{
		Name: "Smithy", Key: "smithy", Category: "production",
		BaseCost:    map[string]float64{"wood": 900, "stone": 600, "iron": 200},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.10}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "A forge smelting iron tools. +0.10 iron/tick (4 workers).",
		LineageKey:  "engineering", LineageTier: 0,
		WorkerDomain: "engineering", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "iron",
	})
	// tier 1 — iron_age  output=iron  rate=0.20
	b = append(b, BuildingDef{
		Name: "Ironworks", Key: "ironworks", Category: "production",
		BaseCost:    map[string]float64{"stone": 6000, "iron": 3000, "coal": 1500},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.20}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "Organised iron working and tools. +0.20 iron/tick (5 workers).",
		LineageKey:  "engineering", LineageTier: 1,
		WorkerDomain: "engineering", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 2 — classical_age  output=iron  rate=0.40
	b = append(b, BuildingDef{
		Name: "Aqueduct", Key: "aqueduct", Category: "production",
		BaseCost:    map[string]float64{"stone": 35000, "gold": 10000, "iron": 8000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.40}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Roman engineering feats improve iron output. +0.40 iron/tick (5 workers).",
		LineageKey:  "engineering", LineageTier: 2,
		WorkerDomain: "engineering", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 3 — medieval_age  output=iron  rate=0.80
	b = append(b, BuildingDef{
		Name: "Workshop", Key: "workshop", Category: "production",
		BaseCost:    map[string]float64{"stone": 190000, "gold": 60000, "iron": 30000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.80}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "A skilled craftsman's workshop. +0.80 iron/tick (6 workers).",
		LineageKey:  "engineering", LineageTier: 3,
		WorkerDomain: "engineering", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 4 — renaissance_age  output=steel  rate=0.50
	b = append(b, BuildingDef{
		Name: "Mill", Key: "mill", Category: "production",
		BaseCost:    map[string]float64{"gold": 600000, "steel": 200000, "iron": 100000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 0.50}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "Water-powered mill begins steel production. +0.50 steel/tick (6 workers).",
		LineageKey:  "engineering", LineageTier: 4,
		WorkerDomain: "engineering", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 5 — colonial_age  output=steel  rate=1.0
	b = append(b, BuildingDef{
		Name: "Dockyard", Key: "dockyard", Category: "production",
		BaseCost:    map[string]float64{"gold": 3.8e6, "steel": 2e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 1.0}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "Naval dockyard producing steel ships. +1.0 steel/tick (7 workers).",
		LineageKey:  "engineering", LineageTier: 5,
		WorkerDomain: "engineering", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 6 — industrial_age  output=steel  rate=2.0
	b = append(b, BuildingDef{
		Name: "Iron Works Complex", Key: "iron_works_complex", Category: "production",
		BaseCost:    map[string]float64{"steel": 26e6, "coal": 12e6, "gold": 14e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 2.0}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Large-scale industrial iron and steel works. +2.0 steel/tick (8 workers).",
		LineageKey:  "engineering", LineageTier: 6,
		WorkerDomain: "engineering", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 7 — victorian_age  output=steel  rate=4.0  (+ small electricity bonus)
	b = append(b, BuildingDef{
		Name: "Steam Works", Key: "steam_works", Category: "production",
		BaseCost:  map[string]float64{"steel": 195e6, "coal": 100e6, "gold": 120e6},
		CostScale: 1.35,
		Effects: []Effect{
			{Type: "production", Target: "steel", Value: 4.0},
			{Type: "production", Target: "electricity", Value: 5.0},
		},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "Steam-powered steelworks with electricity generation. +4.0 steel, +5.0 electricity/tick (9 workers).",
		LineageKey:  "engineering", LineageTier: 7,
		WorkerDomain: "engineering", WorkerCapacity: 9,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 8 — electric_age  output=electricity  rate=50
	b = append(b, BuildingDef{
		Name: "Power Station", Key: "power_station", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.2e9, "electricity": 500e6, "coal": 400e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 50}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "Coal-fired power station. +50 electricity/tick (10 workers).",
		LineageKey:  "engineering", LineageTier: 8,
		WorkerDomain: "engineering", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 9 — atomic_age  output=electricity  rate=100
	b = append(b, BuildingDef{
		Name: "Nuclear Plant", Key: "nuclear_plant", Category: "production",
		BaseCost:    map[string]float64{"steel": 6e9, "electricity": 2.5e9, "uranium": 500e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 100}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Nuclear fission power generation. +100 electricity/tick (11 workers).",
		LineageKey:  "engineering", LineageTier: 9,
		WorkerDomain: "engineering", WorkerCapacity: 11,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 10 — modern_age  output=electricity  rate=200
	b = append(b, BuildingDef{
		Name: "Power Grid Hub", Key: "power_grid_hub", Category: "production",
		BaseCost:    map[string]float64{"steel": 35e9, "electricity": 14e9, "data": 1.5e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 200}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Smart power grid distribution hub. +200 electricity/tick (12 workers).",
		LineageKey:  "engineering", LineageTier: 10,
		WorkerDomain: "engineering", WorkerCapacity: 12,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 11 — information_age  output=electricity  rate=400
	b = append(b, BuildingDef{
		Name: "Smart Grid Node", Key: "smart_grid_node", Category: "production",
		BaseCost:    map[string]float64{"electricity": 100e9, "data": 10e9, "steel": 180e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 400}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "AI-managed smart grid node. +400 electricity/tick (13 workers).",
		LineageKey:  "engineering", LineageTier: 11,
		WorkerDomain: "engineering", WorkerCapacity: 13,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 12 — digital_age  output=electricity  rate=800
	b = append(b, BuildingDef{
		Name: "Neural Grid", Key: "neural_grid", Category: "production",
		BaseCost:    map[string]float64{"electricity": 480e9, "data": 60e9, "steel": 700e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 800}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Neural-network managed power grid. +800 electricity/tick (14 workers).",
		LineageKey:  "engineering", LineageTier: 12,
		WorkerDomain: "engineering", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 13 — cyberpunk_age  output=plasma+electricity  rate: plasma=5, electricity=500
	b = append(b, BuildingDef{
		Name: "Augmentation Foundry", Key: "augmentation_foundry", Category: "production",
		BaseCost:  map[string]float64{"data": 230e9, "crypto": 1.2e12, "electricity": 2.4e12},
		CostScale: 1.35,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 5},
			{Type: "production", Target: "electricity", Value: 500},
		},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Cyberpunk foundry producing plasma and power. +5 plasma, +500 electricity/tick (15 workers).",
		LineageKey:  "engineering", LineageTier: 13,
		WorkerDomain: "engineering", WorkerCapacity: 15,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 14 — fusion_age  output=plasma  rate=10
	b = append(b, BuildingDef{
		Name: "Fusion Reactor", Key: "fusion_reactor", Category: "production",
		BaseCost:  map[string]float64{"plasma": 5e12, "electricity": 16e12, "steel": 22e12},
		CostScale: 1.35,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 10},
			{Type: "production", Target: "electricity", Value: 1000},
		},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Fusion reactor generating plasma and electricity. +10 plasma, +1000 electricity/tick.",
		LineageKey:  "engineering", LineageTier: 14,
		WorkerDomain: "engineering", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 15 — space_age  output=plasma  rate=20
	b = append(b, BuildingDef{
		Name: "Launch Complex", Key: "launch_complex", Category: "production",
		BaseCost:    map[string]float64{"titanium": 85e12, "plasma": 42e12, "electricity": 105e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "plasma", Value: 20}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Space launch complex generating plasma thrust. +20 plasma/tick (20 workers).",
		LineageKey:  "engineering", LineageTier: 15,
		WorkerDomain: "engineering", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 16 — interstellar_age  output=dark_matter  rate=5
	b = append(b, BuildingDef{
		Name: "Warp Drive Plant", Key: "warp_drive_plant", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 100e12, "titanium": 800e12, "plasma": 500e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 5}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Warp drive manufacturing harnessing dark matter. +5 dark_matter/tick (22 workers).",
		LineageKey:  "engineering", LineageTier: 16,
		WorkerDomain: "engineering", WorkerCapacity: 22,
		EpochKey: "cosmic_era", OutputResource: "dark_matter",
	})
	// tier 17 — galactic_age  output=dark_matter  rate=10
	b = append(b, BuildingDef{
		Name: "Dyson Assembly", Key: "dyson_assembly", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 200e12, "dark_matter": 1e15, "titanium": 5e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 10}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Assembly of Dyson sphere panels harvesting dark matter. +10 dark_matter/tick (25 workers).",
		LineageKey:  "engineering", LineageTier: 17,
		WorkerDomain: "engineering", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "dark_matter",
	})
	// tier 18 — quantum_age  output=quantum_flux  rate=20
	b = append(b, BuildingDef{
		Name: "Reality Forge", Key: "reality_forge", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 220e12, "antimatter": 65e15, "dark_matter": 55e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 20}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Forges structures from the quantum fabric. +20 quantum_flux/tick (30 workers).",
		LineageKey:  "engineering", LineageTier: 18,
		WorkerDomain: "engineering", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	return b
}
