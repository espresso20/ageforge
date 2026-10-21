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
		// note: no iron in the price. The smithy is the Bronze Age's only iron
		// source and the age grants 30 iron, so an iron price walled the age.
		BaseCost:    map[string]float64{"wood": 3800, "stone": 2500},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 2.41}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "A forge smelting iron tools.",
		LineageKey:  "engineering", LineageTier: 0, LineKey: "engineering_iron",
		WorkerDomain: "engineering", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "iron",
	})
	// tier 1 — iron_age  output=iron  rate=0.20
	b = append(b, BuildingDef{
		Name: "Ironworks", Key: "ironworks", Category: "production",
		BaseCost:    map[string]float64{"stone": 12500, "marble": 12500, "iron": 6500, "iron_ore": 6500},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 17.8}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "Organized iron working and tools.",
		LineageKey:  "engineering", LineageTier: 1, LineKey: "engineering_iron",
		WorkerDomain: "engineering", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 2 — classical_age  output=iron  rate=0.40
	b = append(b, BuildingDef{
		Name: "Aqueduct", Key: "aqueduct", Category: "production",
		BaseCost:    map[string]float64{"stone": 75000, "marble": 75000, "gold": 42000, "iron": 17000, "iron_ore": 17000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 43.2}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Roman engineering feats improve iron output.",
		LineageKey:  "engineering", LineageTier: 2, LineKey: "engineering_iron",
		WorkerDomain: "engineering", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 3 — medieval_age  output=iron  rate=0.80
	b = append(b, BuildingDef{
		Name: "Workshop", Key: "workshop", Category: "production",
		BaseCost:    map[string]float64{"stone": 800000, "gold": 250000, "iron": 65000, "iron_ore": 65000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 140}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "A skilled craftsman's workshop.",
		LineageKey:  "engineering", LineageTier: 3, LineKey: "engineering_iron",
		WorkerDomain: "engineering", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 4 — renaissance_age  output=steel  rate=1.60
	b = append(b, BuildingDef{
		Name: "Mill", Key: "mill", Category: "production",
		// note: no steel in the price. Every steel producer cost steel, with a
		// 0.1/tick tech the only way in, so the Renaissance walled at ~850K
		// steel per mill. The mill is the bootstrap; the foundry still costs steel.
		BaseCost:    map[string]float64{"gold": 2500000, "iron": 210000, "iron_ore": 210000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 191}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "Water-powered mill begins steel production.",
		LineageKey:  "engineering", LineageTier: 4, LineKey: "engineering_steel",
		WorkerDomain: "engineering", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 5 — colonial_age  output=steel  rate=3.20
	b = append(b, BuildingDef{
		Name: "Dockyard", Key: "dockyard", Category: "production",
		BaseCost:    map[string]float64{"gold": 1.6e07, "steel": 8500000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 2590}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "Naval dockyard producing steel ships.",
		LineageKey:  "engineering", LineageTier: 5, LineKey: "engineering_steel",
		WorkerDomain: "engineering", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 6 — industrial_age  output=steel  rate=6.40
	b = append(b, BuildingDef{
		Name: "Integrated Steelworks", Key: "iron_works_complex", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.1e08, "coal": 5.1e07, "gold": 5.9e07},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 38700}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "Large-scale industrial iron and steel works.",
		LineageKey:  "engineering", LineageTier: 6, LineKey: "engineering_steel",
		WorkerDomain: "engineering", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 7 — victorian_age  output=steel  rate=12.80  (+ electricity bonus)
	b = append(b, BuildingDef{
		Name: "Steam Works", Key: "steam_works", Category: "production",
		BaseCost:  map[string]float64{"steel": 8.3e08, "coal": 4.2e08, "gold": 5.1e08},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "steel", Value: 148000},
			{Type: "production", Target: "electricity", Value: 5},
		},
		BuildTicks:   3600,
		RequiredAge:  "victorian_age",
		RequiredTech: "electrification",
		Description:  "Steam-powered steelworks with electricity generation.",
		LineageKey:   "engineering", LineageTier: 7, LineKey: "engineering_steel",
		WorkerDomain: "engineering", WorkerCapacity: 9,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 8 — electric_age  output=electricity  rate=25.60
	b = append(b, BuildingDef{
		Name: "Power Station", Key: "power_station", Category: "production",
		BaseCost:    map[string]float64{"steel": 9.49e9, "coal": 3.16e9},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 273000}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "Coal-fired power station.",
		LineageKey:  "engineering", LineageTier: 8, LineKey: "engineering_electricity",
		WorkerDomain: "engineering", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 9 — atomic_age  output=electricity  rate=51.20
	b = append(b, BuildingDef{
		Name: "Nuclear Plant", Key: "nuclear_plant", Category: "production",
		BaseCost:     map[string]float64{"steel": 2.5e10, "electricity": 1.1e10, "uranium": 2.1e09},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "electricity", Value: 1260000}},
		BuildTicks:   3600,
		RequiredAge:  "atomic_age",
		RequiredTech: "civilian_reactors",
		Description:  "Nuclear fission power generation.",
		LineageKey:   "engineering", LineageTier: 9, LineKey: "engineering_electricity",
		WorkerDomain: "engineering", WorkerCapacity: 11,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 10 — modern_age  output=electricity  rate=102.40
	b = append(b, BuildingDef{
		Name: "Power Grid Hub", Key: "power_grid_hub", Category: "production",
		BaseCost:  map[string]float64{"steel": 7.5e10, "titanium_ore": 7.5e10, "electricity": 5.9e10, "data": 6.4e09},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "electricity", Value: 7290000},
		},
		BuildTicks:   3600,
		RequiredAge:  "modern_age",
		RequiredTech: "electricity_tech",
		Description:  "Smart power grid distribution hub.",
		LineageKey:   "engineering", LineageTier: 10, LineKey: "engineering_electricity",
		WorkerDomain: "engineering", WorkerCapacity: 12,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 11 — information_age  output=electricity  rate=204.80
	b = append(b, BuildingDef{
		Name: "Smart Grid Node", Key: "smart_grid_node", Category: "production",
		BaseCost:    map[string]float64{"electricity": 4.2e11, "data": 4.2e10, "steel": 3.8e11, "titanium_ore": 3.8e11},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 8.68e07}},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "AI-managed smart grid node.",
		LineageKey:  "engineering", LineageTier: 11, LineKey: "engineering_electricity",
		WorkerDomain: "engineering", WorkerCapacity: 13,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 12 — digital_age  output=electricity  rate=409.60
	b = append(b, BuildingDef{
		Name: "Neural Grid", Key: "neural_grid", Category: "production",
		BaseCost:    map[string]float64{"electricity": 2e12, "data": 2.5e11, "steel": 1.5e12, "titanium_ore": 1.5e12, "nanobots": 15000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 1.5e08}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Neural-network managed power grid.",
		LineageKey:  "engineering", LineageTier: 12, LineKey: "engineering_electricity",
		WorkerDomain: "engineering", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 13 — cyberpunk_age  output=plasma+electricity  rate: plasma=819.20, electricity=500
	b = append(b, BuildingDef{
		Name: "Augmentation Foundry", Key: "augmentation_foundry", Category: "production",
		BaseCost:  map[string]float64{"data": 9.7e11, "crypto": 5.1e12, "electricity": 5e12, "dark_matter_crystals": 5e12, "nanobots": 160000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 819.2},
			{Type: "production", Target: "electricity", Value: 7.17e08},
		},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "Cyberpunk foundry producing plasma and power.",
		LineageKey:  "engineering", LineageTier: 13, LineKey: "engineering_plasma",
		WorkerDomain: "engineering", WorkerCapacity: 15,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 14 — fusion_age  output=plasma  rate=1638.40
	b = append(b, BuildingDef{
		Name: "Fusion Reactor", Key: "fusion_reactor", Category: "production",
		BaseCost:  map[string]float64{"electricity": 4.96e13, "dark_matter_crystals": 4.95e13, "steel": 1.36e14},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 2.59e08},
			{Type: "production", Target: "electricity", Value: 7.76e08},
		},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "Fusion reactor generating plasma and electricity.",
		LineageKey:  "engineering", LineageTier: 14, LineKey: "engineering_plasma",
		WorkerDomain: "engineering", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 15 — space_age  output=plasma  rate=3276.80
	b = append(b, BuildingDef{
		Name: "Launch Complex", Key: "launch_complex", Category: "production",
		BaseCost:    map[string]float64{"titanium": 3.6e14, "plasma": 1.8e14, "electricity": 2.2e14, "dark_matter_crystals": 2.2e14},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "plasma", Value: 5.52e09}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "Space launch complex generating plasma thrust.",
		LineageKey:  "engineering", LineageTier: 15, LineKey: "engineering_plasma",
		WorkerDomain: "engineering", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 16 — interstellar_age  output=dark_matter  rate=6553.60
	b = append(b, BuildingDef{
		Name: "Warp Drive Plant", Key: "warp_drive_plant", Category: "production",
		BaseCost:  map[string]float64{"titanium": 5.09e15, "plasma": 3.14e15},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "dark_matter", Value: 6.45e9},
			{Type: "production", Target: "plasma", Value: 4.68e10},
		},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "Warp drive manufacturing harnessing dark matter.",
		LineageKey:  "engineering", LineageTier: 16, LineKey: "engineering_dark_matter",
		WorkerDomain: "engineering", WorkerCapacity: 22,
		EpochKey: "cosmic_era", OutputResource: "dark_matter",
	})
	// tier 17 — galactic_age  output=dark_matter  rate=13107.20
	b = append(b, BuildingDef{
		Name: "Dyson Assembly", Key: "dyson_assembly", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 8.5e14, "dark_matter": 4.2e15, "titanium": 2.1e16},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 7.47e10}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Assembly of Dyson sphere panels harvesting dark matter.",
		LineageKey:  "engineering", LineageTier: 17, LineKey: "engineering_dark_matter",
		WorkerDomain: "engineering", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "dark_matter",
	})
	// tier 18 — quantum_age  output=quantum_flux  rate=26214.40
	b = append(b, BuildingDef{
		Name: "Reality Forge", Key: "reality_forge", Category: "production",
		BaseCost:  map[string]float64{"antimatter": 4.44e17, "quantum_flux": 1.47e15},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "quantum_flux", Value: 2.85e10},
			{Type: "production", Target: "dark_matter", Value: 5.01e12},
		},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Forges structures from the quantum fabric.",
		LineageKey:  "engineering", LineageTier: 18, LineKey: "engineering_quantum_flux",
		WorkerDomain: "engineering", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})
	// tier 19 — transcendent_age  output=quantum_flux  rate=52428.80
	b = append(b, BuildingDef{
		Name: "Singularity Engine", Key: "singularity_engine", Category: "production",
		BaseCost:     map[string]float64{"quantum_flux": 9.3e15, "antimatter": 2.8e18, "dark_matter": 2.3e18},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "quantum_flux", Value: 5.94e11}},
		BuildTicks:   3600,
		RequiredAge:  "transcendent_age",
		RequiredTech: "singularity_engineering",
		Description:  "A singularity-powered engine generating quantum flux.",
		LineageKey:   "engineering", LineageTier: 19, LineKey: "engineering_quantum_flux",
		WorkerDomain: "engineering", WorkerCapacity: 35,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	return b
}
