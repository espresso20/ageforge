package config

// buildingsLineageEngineering returns lineages 5-9:
// knowledge, faith, military, trade, engineering.
// Merged into newProductionBuildings() via init — see buildings_new_merge.go.
func buildingsLineageEngineering() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 9 — ENGINEERING (lineageKey: "engineering", domain: "engineering")
	// starts at bronze_age (tier 0)
	// CostScale: ProducerRate (1.15)  Category: "production"  (both in the lineageSpec at the foot)
	// Output transitions: iron(0-3) → steel(4-7) → electricity(8-12) → plasma+electricity(13-14)
	//                     → dark_matter(15+)
	// =========================================================================

	// tier 0 — bronze_age  output=iron  rate≈0.10
	b = append(b, BuildingDef{
		Name: "Smithy", Key: "smithy",
		// note: no iron in the price. The smithy is the Bronze Age's only iron
		// source and the age grants 30 iron, so an iron price walled the age.
		BaseCost:       map[string]float64{"wood": 3800, "stone": 2500},
		Effects:        []Effect{{Type: "production", Target: "iron", Value: 2.41}},
		BuildTicks:     150,
		RequiredAge:    "bronze_age",
		Description:    "A forge smelting iron tools.",
		WorkerCapacity: 4,
	})
	// tier 1 — iron_age  output=iron  rate=0.20
	b = append(b, BuildingDef{
		Name: "Ironworks", Key: "ironworks",
		BaseCost:       map[string]float64{"stone": 12500, "marble": 12500, "iron": 6500, "iron_ore": 6500},
		Effects:        []Effect{{Type: "production", Target: "iron", Value: 17.8}},
		BuildTicks:     300,
		RequiredAge:    "iron_age",
		Description:    "Organized iron working and tools.",
		WorkerCapacity: 5,
	})
	// tier 2 — classical_age  output=iron  rate=0.40
	b = append(b, BuildingDef{
		Name: "Aqueduct", Key: "aqueduct",
		BaseCost:       map[string]float64{"stone": 75000, "marble": 75000, "gold": 42000, "iron": 17000, "iron_ore": 17000},
		Effects:        []Effect{{Type: "production", Target: "iron", Value: 43.2}},
		BuildTicks:     600,
		RequiredAge:    "classical_age",
		Description:    "Roman engineering feats improve iron output.",
		WorkerCapacity: 5,
	})
	// tier 3 — medieval_age  output=iron  rate=0.80
	b = append(b, BuildingDef{
		Name: "Workshop", Key: "workshop",
		BaseCost:       map[string]float64{"stone": 800000, "gold": 250000, "iron": 65000, "iron_ore": 65000},
		Effects:        []Effect{{Type: "production", Target: "iron", Value: 140}},
		BuildTicks:     1200,
		RequiredAge:    "medieval_age",
		Description:    "A skilled craftsman's workshop.",
		WorkerCapacity: 6,
	})
	// tier 4 — renaissance_age  output=steel  rate=1.60
	b = append(b, BuildingDef{
		Name: "Mill", Key: "mill",
		// note: no steel in the price. Every steel producer cost steel, with a
		// 0.1/tick tech the only way in, so the Renaissance walled at ~850K
		// steel per mill. The mill is the bootstrap; the foundry still costs steel.
		BaseCost:       map[string]float64{"gold": 2.5 * Million, "iron": 210000, "iron_ore": 210000},
		Effects:        []Effect{{Type: "production", Target: "steel", Value: 191}},
		BuildTicks:     2400,
		RequiredAge:    "renaissance_age",
		Description:    "Water-powered mill begins steel production.",
		WorkerCapacity: 6,
	})
	// tier 5 — colonial_age  output=steel  rate=3.20
	b = append(b, BuildingDef{
		Name: "Dockyard", Key: "dockyard",
		BaseCost:       map[string]float64{"gold": 16 * Million, "steel": 8.5 * Million},
		Effects:        []Effect{{Type: "production", Target: "steel", Value: 2590}},
		BuildTicks:     3600,
		RequiredAge:    "colonial_age",
		Description:    "Naval dockyard producing steel ships.",
		WorkerCapacity: 7,
	})
	// tier 6 — industrial_age  output=steel  rate=6.40
	b = append(b, BuildingDef{
		Name: "Integrated Steelworks", Key: "iron_works_complex",
		BaseCost:       map[string]float64{"steel": 110 * Million, "coal": 51 * Million, "gold": 59 * Million},
		Effects:        []Effect{{Type: "production", Target: "steel", Value: 38700}},
		BuildTicks:     3600,
		RequiredAge:    "industrial_age",
		Description:    "Large-scale industrial iron and steel works.",
		WorkerCapacity: 8,
	})
	// tier 7 — victorian_age  output=steel  rate=12.80  (+ electricity bonus)
	b = append(b, BuildingDef{
		Name: "Steam Works", Key: "steam_works",
		BaseCost: map[string]float64{"steel": 830 * Million, "coal": 420 * Million, "gold": 510 * Million},
		Effects: []Effect{
			{Type: "production", Target: "steel", Value: 148000},
			{Type: "production", Target: "electricity", Value: 5},
		},
		BuildTicks:     3600,
		RequiredAge:    "victorian_age",
		RequiredTech:   "electrification",
		Description:    "Steam-powered steelworks with electricity generation.",
		WorkerCapacity: 9,
	})
	// tier 8 — electric_age  output=electricity  rate=25.60
	b = append(b, BuildingDef{
		Name: "Power Station", Key: "power_station",
		BaseCost:       map[string]float64{"steel": 9.49 * Billion, "coal": 3.16 * Billion},
		Effects:        []Effect{{Type: "production", Target: "electricity", Value: 273000}},
		BuildTicks:     3600,
		RequiredAge:    "electric_age",
		Description:    "Coal-fired power station.",
		WorkerCapacity: 10,
	})
	// tier 9 — atomic_age  output=electricity  rate=51.20
	b = append(b, BuildingDef{
		Name: "Nuclear Plant", Key: "nuclear_plant",
		BaseCost:       map[string]float64{"steel": 25 * Billion, "electricity": 11 * Billion, "uranium": 2.1 * Billion},
		Effects:        []Effect{{Type: "production", Target: "electricity", Value: 1.26 * Million}},
		BuildTicks:     3600,
		RequiredAge:    "atomic_age",
		RequiredTech:   "civilian_reactors",
		Description:    "Nuclear fission power generation.",
		WorkerCapacity: 11,
	})
	// tier 10 — modern_age  output=electricity  rate=102.40
	b = append(b, BuildingDef{
		Name: "Power Grid Hub", Key: "power_grid_hub",
		BaseCost: map[string]float64{"steel": 75 * Billion, "titanium_ore": 75 * Billion, "electricity": 59 * Billion, "data": 6.4 * Billion},
		Effects: []Effect{
			{Type: "production", Target: "electricity", Value: 7.29 * Million},
		},
		BuildTicks:     3600,
		RequiredAge:    "modern_age",
		RequiredTech:   "electricity_tech",
		Description:    "Smart power grid distribution hub.",
		WorkerCapacity: 12,
	})
	// tier 11 — information_age  output=electricity  rate=204.80
	b = append(b, BuildingDef{
		Name: "Smart Grid Node", Key: "smart_grid_node",
		BaseCost:       map[string]float64{"electricity": 420 * Billion, "data": 42 * Billion, "steel": 380 * Billion, "titanium_ore": 380 * Billion},
		Effects:        []Effect{{Type: "production", Target: "electricity", Value: 86.8 * Million}},
		BuildTicks:     3600,
		RequiredAge:    "information_age",
		Description:    "AI-managed smart grid node.",
		WorkerCapacity: 13,
	})
	// tier 12 — digital_age  output=electricity  rate=409.60
	b = append(b, BuildingDef{
		Name: "Neural Grid", Key: "neural_grid",
		BaseCost:       map[string]float64{"electricity": 2 * Trillion, "data": 250 * Billion, "steel": 1.5 * Trillion, "titanium_ore": 1.5 * Trillion, "nanobots": 15000},
		Effects:        []Effect{{Type: "production", Target: "electricity", Value: 150 * Million}},
		BuildTicks:     3600,
		RequiredAge:    "digital_age",
		Description:    "Neural-network managed power grid.",
		WorkerCapacity: 14,
	})
	// tier 13 — cyberpunk_age  output=plasma+electricity  rate: plasma=819.20, electricity=500
	b = append(b, BuildingDef{
		Name: "Augmentation Foundry", Key: "augmentation_foundry",
		BaseCost: map[string]float64{"data": 970 * Billion, "crypto": 5.1 * Trillion, "electricity": 5 * Trillion, "dark_matter_crystals": 5 * Trillion, "nanobots": 160000},
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 819.2},
			{Type: "production", Target: "electricity", Value: 717 * Million},
		},
		BuildTicks:     3600,
		RequiredAge:    "cyberpunk_age",
		Description:    "Cyberpunk foundry producing plasma and power.",
		WorkerCapacity: 15,
	})
	// tier 14 — fusion_age  output=plasma  rate=1638.40
	b = append(b, BuildingDef{
		Name: "Fusion Reactor", Key: "fusion_reactor",
		BaseCost: map[string]float64{"electricity": 49.6 * Trillion, "dark_matter_crystals": 49.5 * Trillion, "steel": 136 * Trillion},
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 259 * Million},
			{Type: "production", Target: "electricity", Value: 776 * Million},
		},
		BuildTicks:     3600,
		RequiredAge:    "fusion_age",
		Description:    "Fusion reactor generating plasma and electricity.",
		WorkerCapacity: 18,
	})
	// tier 15 — space_age  output=plasma  rate=3276.80
	b = append(b, BuildingDef{
		Name: "Launch Complex", Key: "launch_complex",
		BaseCost:       map[string]float64{"titanium": 360 * Trillion, "plasma": 180 * Trillion, "electricity": 220 * Trillion, "dark_matter_crystals": 220 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "plasma", Value: 5.52 * Billion}},
		BuildTicks:     3600,
		RequiredAge:    "space_age",
		Description:    "Space launch complex generating plasma thrust.",
		WorkerCapacity: 20,
	})
	// tier 16 — interstellar_age  output=dark_matter  rate=6553.60
	b = append(b, BuildingDef{
		Name: "Warp Drive Plant", Key: "warp_drive_plant",
		BaseCost: map[string]float64{"titanium": 5.09 * Quadrillion, "plasma": 3.14 * Quadrillion},
		Effects: []Effect{
			{Type: "production", Target: "dark_matter", Value: 6.45 * Billion},
			{Type: "production", Target: "plasma", Value: 46.8 * Billion},
		},
		BuildTicks:     3600,
		RequiredAge:    "interstellar_age",
		Description:    "Warp drive manufacturing harnessing dark matter.",
		WorkerCapacity: 22,
	})
	// tier 17 — galactic_age  output=dark_matter  rate=13107.20
	b = append(b, BuildingDef{
		Name: "Dyson Assembly", Key: "dyson_assembly",
		BaseCost:       map[string]float64{"antimatter": 850 * Trillion, "dark_matter": 4.2 * Quadrillion, "titanium": 21 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "dark_matter", Value: 74.7 * Billion}},
		BuildTicks:     3600,
		RequiredAge:    "galactic_age",
		Description:    "Assembly of Dyson sphere panels harvesting dark matter.",
		WorkerCapacity: 25,
	})
	// tier 18 — quantum_age  output=quantum_flux  rate=26214.40
	b = append(b, BuildingDef{
		Name: "Reality Forge", Key: "reality_forge",
		BaseCost: map[string]float64{"antimatter": 444 * Quadrillion, "quantum_flux": 1.47 * Quadrillion},
		Effects: []Effect{
			{Type: "production", Target: "quantum_flux", Value: 28.5 * Billion},
			{Type: "production", Target: "dark_matter", Value: 5.01 * Trillion},
		},
		BuildTicks:     3600,
		RequiredAge:    "quantum_age",
		Description:    "Forges structures from the quantum fabric.",
		WorkerCapacity: 30,
	})
	// tier 19 — transcendent_age  output=quantum_flux  rate=52428.80
	b = append(b, BuildingDef{
		Name: "Singularity Engine", Key: "singularity_engine",
		BaseCost:       map[string]float64{"quantum_flux": 9.3 * Quadrillion, "antimatter": 2.8 * Quintillion, "dark_matter": 2.3 * Quintillion},
		Effects:        []Effect{{Type: "production", Target: "quantum_flux", Value: 594 * Billion}},
		BuildTicks:     3600,
		RequiredAge:    "transcendent_age",
		RequiredTech:   "singularity_engineering",
		Description:    "A singularity-powered engine generating quantum flux.",
		WorkerCapacity: 35,
	})

	return finishLineage(b, lineageSpec{
		Key: "engineering", Category: "production", Scale: ProducerRate, Domain: "engineering", Lines: true,
	})
}
