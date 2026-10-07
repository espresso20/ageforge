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
		BaseCost:    map[string]float64{"wood": 900, "stone": 600},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.10}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "A forge smelting iron tools.",
		LineageKey:  "engineering", LineageTier: 0,
		WorkerDomain: "engineering", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "iron",
	})
	// tier 1 — iron_age  output=iron  rate=0.20
	b = append(b, BuildingDef{
		Name: "Ironworks", Key: "ironworks", Category: "production",
		BaseCost:    map[string]float64{"stone": 6000, "iron": 3000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.20}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "Organized iron working and tools.",
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
		Description: "Roman engineering feats improve iron output.",
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
		Description: "A skilled craftsman's workshop.",
		LineageKey:  "engineering", LineageTier: 3,
		WorkerDomain: "engineering", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 4 — renaissance_age  output=steel  rate=1.60
	b = append(b, BuildingDef{
		Name: "Mill", Key: "mill", Category: "production",
		// note: no steel in the price. Every steel producer cost steel, with a
		// 0.1/tick tech the only way in, so the Renaissance walled at ~850K
		// steel per mill. The mill is the bootstrap; the foundry still costs steel.
		BaseCost:    map[string]float64{"gold": 600000, "iron": 100000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 1.60}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "Water-powered mill begins steel production.",
		LineageKey:  "engineering", LineageTier: 4,
		WorkerDomain: "engineering", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 5 — colonial_age  output=steel  rate=3.20
	b = append(b, BuildingDef{
		Name: "Dockyard", Key: "dockyard", Category: "production",
		BaseCost:    map[string]float64{"gold": 3.8e6, "steel": 2e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 3.20}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "Naval dockyard producing steel ships.",
		LineageKey:  "engineering", LineageTier: 5,
		WorkerDomain: "engineering", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 6 — industrial_age  output=steel  rate=6.40
	b = append(b, BuildingDef{
		Name: "Integrated Steelworks", Key: "iron_works_complex", Category: "production",
		BaseCost:    map[string]float64{"steel": 26e6, "coal": 12e6, "gold": 14e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 6.40}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "Large-scale industrial iron and steel works.",
		LineageKey:  "engineering", LineageTier: 6,
		WorkerDomain: "engineering", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 7 — victorian_age  output=steel  rate=12.80  (+ electricity bonus)
	b = append(b, BuildingDef{
		Name: "Steam Works", Key: "steam_works", Category: "production",
		BaseCost:  map[string]float64{"steel": 195e6, "coal": 100e6, "gold": 120e6},
		CostScale: 1.35,
		Effects: []Effect{
			{Type: "production", Target: "steel", Value: 12.80},
			{Type: "production", Target: "electricity", Value: 5.0},
		},
		BuildTicks:   3600,
		RequiredAge:  "victorian_age",
		RequiredTech: "electrification",
		Description:  "Steam-powered steelworks with electricity generation.",
		LineageKey:   "engineering", LineageTier: 7,
		WorkerDomain: "engineering", WorkerCapacity: 9,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 8 — electric_age  output=electricity  rate=25.60
	b = append(b, BuildingDef{
		Name: "Power Station", Key: "power_station", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.2e9, "electricity": 500e6, "coal": 400e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 25.60}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "Coal-fired power station.",
		LineageKey:  "engineering", LineageTier: 8,
		WorkerDomain: "engineering", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 9 — atomic_age  output=electricity  rate=51.20
	b = append(b, BuildingDef{
		Name: "Nuclear Plant", Key: "nuclear_plant", Category: "production",
		BaseCost:     map[string]float64{"steel": 6e9, "electricity": 2.5e9, "uranium": 500e6},
		CostScale:    1.35,
		Effects:      []Effect{{Type: "production", Target: "electricity", Value: 51.20}},
		BuildTicks:   3600,
		RequiredAge:  "atomic_age",
		RequiredTech: "civilian_reactors",
		Description:  "Nuclear fission power generation.",
		LineageKey:   "engineering", LineageTier: 9,
		WorkerDomain: "engineering", WorkerCapacity: 11,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 10 — modern_age  output=electricity  rate=102.40
	b = append(b, BuildingDef{
		Name: "Power Grid Hub", Key: "power_grid_hub", Category: "production",
		BaseCost:     map[string]float64{"steel": 35e9, "electricity": 14e9, "data": 1.5e9},
		CostScale:    1.35,
		Effects:      []Effect{{Type: "production", Target: "electricity", Value: 102.40}},
		BuildTicks:   3600,
		RequiredAge:  "modern_age",
		RequiredTech: "electricity_tech",
		Description:  "Smart power grid distribution hub.",
		LineageKey:   "engineering", LineageTier: 10,
		WorkerDomain: "engineering", WorkerCapacity: 12,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 11 — information_age  output=electricity  rate=204.80
	b = append(b, BuildingDef{
		Name: "Smart Grid Node", Key: "smart_grid_node", Category: "production",
		BaseCost:    map[string]float64{"electricity": 100e9, "data": 10e9, "steel": 180e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 204.80}},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "AI-managed smart grid node.",
		LineageKey:  "engineering", LineageTier: 11,
		WorkerDomain: "engineering", WorkerCapacity: 13,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 12 — digital_age  output=electricity  rate=409.60
	b = append(b, BuildingDef{
		Name: "Neural Grid", Key: "neural_grid", Category: "production",
		BaseCost:    map[string]float64{"electricity": 480e9, "data": 60e9, "steel": 700e9, "nanobots": 3500},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 409.60}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Neural-network managed power grid.",
		LineageKey:  "engineering", LineageTier: 12,
		WorkerDomain: "engineering", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 13 — cyberpunk_age  output=plasma+electricity  rate: plasma=819.20, electricity=500
	b = append(b, BuildingDef{
		Name: "Augmentation Foundry", Key: "augmentation_foundry", Category: "production",
		BaseCost:  map[string]float64{"data": 230e9, "crypto": 1.2e12, "electricity": 2.4e12, "nanobots": 38000},
		CostScale: 1.35,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 819.20},
			{Type: "production", Target: "electricity", Value: 500},
		},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "Cyberpunk foundry producing plasma and power.",
		LineageKey:  "engineering", LineageTier: 13,
		WorkerDomain: "engineering", WorkerCapacity: 15,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 14 — fusion_age  output=plasma  rate=1638.40
	b = append(b, BuildingDef{
		Name: "Fusion Reactor", Key: "fusion_reactor", Category: "production",
		BaseCost:  map[string]float64{"plasma": 5e12, "electricity": 16e12, "steel": 22e12},
		CostScale: 1.35,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 1638.40},
			{Type: "production", Target: "electricity", Value: 1000},
		},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "Fusion reactor generating plasma and electricity.",
		LineageKey:  "engineering", LineageTier: 14,
		WorkerDomain: "engineering", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 15 — space_age  output=plasma  rate=3276.80
	b = append(b, BuildingDef{
		Name: "Launch Complex", Key: "launch_complex", Category: "production",
		BaseCost:    map[string]float64{"titanium": 85e12, "plasma": 42e12, "electricity": 105e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "plasma", Value: 3276.80}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "Space launch complex generating plasma thrust.",
		LineageKey:  "engineering", LineageTier: 15,
		WorkerDomain: "engineering", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 16 — interstellar_age  output=dark_matter  rate=6553.60
	b = append(b, BuildingDef{
		Name: "Warp Drive Plant", Key: "warp_drive_plant", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 100e12, "titanium": 800e12, "plasma": 500e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 6553.60}},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "Warp drive manufacturing harnessing dark matter.",
		LineageKey:  "engineering", LineageTier: 16,
		WorkerDomain: "engineering", WorkerCapacity: 22,
		EpochKey: "cosmic_era", OutputResource: "dark_matter",
	})
	// tier 17 — galactic_age  output=dark_matter  rate=13107.20
	b = append(b, BuildingDef{
		Name: "Dyson Assembly", Key: "dyson_assembly", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 200e12, "dark_matter": 1e15, "titanium": 5e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 13107.20}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Assembly of Dyson sphere panels harvesting dark matter.",
		LineageKey:  "engineering", LineageTier: 17,
		WorkerDomain: "engineering", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "dark_matter",
	})
	// tier 18 — quantum_age  output=quantum_flux  rate=26214.40
	b = append(b, BuildingDef{
		Name: "Reality Forge", Key: "reality_forge", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 220e12, "antimatter": 65e15, "dark_matter": 55e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 26214.40}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Forges structures from the quantum fabric.",
		LineageKey:  "engineering", LineageTier: 18,
		WorkerDomain: "engineering", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})
	// tier 19 — transcendent_age  output=quantum_flux  rate=52428.80
	b = append(b, BuildingDef{
		Name: "Singularity Engine", Key: "singularity_engine", Category: "production",
		BaseCost:     map[string]float64{"quantum_flux": 2.2e15, "antimatter": 650e15, "dark_matter": 550e15},
		CostScale:    1.35,
		Effects:      []Effect{{Type: "production", Target: "quantum_flux", Value: 52428.80}},
		BuildTicks:   3600,
		RequiredAge:  "transcendent_age",
		RequiredTech: "singularity_engineering",
		Description:  "A singularity-powered engine generating quantum flux.",
		LineageKey:   "engineering", LineageTier: 19,
		WorkerDomain: "engineering", WorkerCapacity: 35,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	return b
}
