package config

// buildingsLineageGeologicalExtraction returns all production, housing, research, and military buildings
// for the 13 lineage chains introduced in Phase 10 of the economy redesign.
// Storage buildings and wonders remain in baseBuildingsRaw().
// All fields are set inline; no separate buildingMeta() entries needed for these buildings.
func buildingsLineageGeologicalExtraction() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 4 — GEOLOGICAL EXTRACTION (lineageKey: "geological_extraction", domain: "masonry")
	// rate = 0.03 * 2^tier  CostScale: ProducerRate (1.15)  Category: "production"  (both in the lineageSpec at the foot)
	// Output: stone(0-2) → dual marble+iron_ore(3-4) → iron_ore(5-8) → uranium(9-11)
	//         → titanium_ore(12-14) → dark_matter_crystals(15-17) → antimatter(18-20)
	// Dual-output tiers (3-4): two Effects each at half rate
	// =========================================================================

	// tier 0 — stone_age  rate=0.08  output=stone
	b = append(b, BuildingDef{
		Name: "Stone Camp", Key: "stone_camp",
		BaseCost:       map[string]float64{"wood": 45},
		Effects:        []Effect{{Type: "production", Target: "stone", Value: 0.274}},
		BuildTicks:     80,
		RequiredAge:    "stone_age",
		Description:    "A basic camp for gathering stones.",
		WorkerCapacity: 3,
	})
	// tier 1 — stone_age  rate=0.16  output=stone
	b = append(b, BuildingDef{
		Name: "Stone Pit", Key: "stone_pit",
		BaseCost:       map[string]float64{"wood": 300, "stone": 180},
		Effects:        []Effect{{Type: "production", Target: "stone", Value: 3.48}},
		BuildTicks:     200,
		RequiredAge:    "stone_age",
		Description:    "A shallow pit dug for stone.",
		WorkerCapacity: 4,
	})
	// tier 2 — bronze_age  rate=0.32  output=stone
	b = append(b, BuildingDef{
		Name: "Quarry", Key: "quarry",
		BaseCost:       map[string]float64{"wood": 1800, "stone": 1200, "iron": 300},
		Effects:        []Effect{{Type: "production", Target: "stone", Value: 4}},
		BuildTicks:     150,
		RequiredAge:    "bronze_age",
		Description:    "Organized stone quarrying.",
		WorkerCapacity: 5,
	})
	// tier 3 — iron_age  rate=1.28  dual: marble(0.64) + iron_ore(0.64)
	b = append(b, BuildingDef{
		Name: "Marble Quarry", Key: "marble_quarry",
		BaseCost: map[string]float64{"stone": 14000, "iron": 6000},
		Effects: []Effect{
			{Type: "production", Target: "marble", Value: 4.51},
			{Type: "production", Target: "iron_ore", Value: 1.89},
			{Type: "production", Target: "stone", Value: 6.29},
			{Type: "production", Target: "iron", Value: 2.26},
		},
		BuildTicks:     300,
		RequiredAge:    "iron_age",
		Description:    "Marble, iron ore, stone and iron extraction.",
		WorkerCapacity: 5,
		// stays stone, though the first production target is marble; LineKey derives from it
		OutputResource: "stone",
	})
	// tier 4 — classical_age  rate=2.56  dual: marble(1.28) + iron_ore(1.28)
	b = append(b, BuildingDef{
		Name: "Marble Works", Key: "marble_works",
		BaseCost: map[string]float64{"stone": 84000, "gold": 30000, "iron": 24000},
		Effects: []Effect{
			{Type: "production", Target: "marble", Value: 28.4},
			{Type: "production", Target: "iron_ore", Value: 6.37},
			{Type: "production", Target: "stone", Value: 35.3},
			{Type: "production", Target: "iron", Value: 8.17},
		},
		BuildTicks:     600,
		RequiredAge:    "classical_age",
		Description:    "Classical marble, ore, stone and iron works.",
		WorkerCapacity: 6,
		// stays stone, though the first production target is marble; LineKey derives from it
		OutputResource: "stone",
	})
	// tier 5 — medieval_age  rate=2.56  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Stonemason's Guild", Key: "stonemasons_guild",
		BaseCost: map[string]float64{"stone": 759000, "gold": 237000},
		Effects: []Effect{
			{Type: "production", Target: "iron_ore", Value: 10.1},
			{Type: "production", Target: "iron", Value: 14.9},
			{Type: "production", Target: "stone", Value: 339},
		},
		BuildTicks:     1200,
		RequiredAge:    "medieval_age",
		Description:    "Guild of stonemasons extracting iron ore, iron and stone.",
		WorkerCapacity: 6,
	})
	// tier 6 — renaissance_age  rate=5.12  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Iron Mine", Key: "iron_mine",
		BaseCost: map[string]float64{"gold": 2.33 * Million, "steel": 839000},
		Effects: []Effect{
			{Type: "production", Target: "iron_ore", Value: 5.27},
			{Type: "production", Target: "iron", Value: 9.02},
			{Type: "production", Target: "stone", Value: 207},
		},
		BuildTicks:     2400,
		RequiredAge:    "renaissance_age",
		Description:    "Deep iron ore, iron and stone mining.",
		WorkerCapacity: 7,
	})
	// tier 7 — colonial_age  rate=10.24  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Deep Iron Mine", Key: "deep_iron_mine",
		BaseCost: map[string]float64{"gold": 9 * Million, "steel": 5.4 * Million},
		Effects: []Effect{
			{Type: "production", Target: "iron_ore", Value: 112},
			{Type: "production", Target: "iron", Value: 112},
			{Type: "production", Target: "stone", Value: 2090},
		},
		BuildTicks:     3600,
		RequiredAge:    "colonial_age",
		Description:    "Colonial deep iron ore, iron and stone extraction.",
		WorkerCapacity: 8,
	})
	// tier 8 — industrial_age  rate=20.48  output=iron_ore
	b = append(b, BuildingDef{
		Name: "Steam Mine", Key: "steam_mine",
		BaseCost: map[string]float64{"steel": 66 * Million, "coal": 24 * Million, "gold": 36 * Million},
		Effects: []Effect{
			{Type: "production", Target: "iron_ore", Value: 161},
			{Type: "production", Target: "iron", Value: 2970},
			{Type: "production", Target: "stone", Value: 4450},
		},
		BuildTicks:     3600,
		RequiredAge:    "industrial_age",
		Description:    "Steam-powered industrial iron ore, iron and stone mine.",
		WorkerCapacity: 10,
	})
	// tier 9 — victorian_age  rate=40.96  output=uranium
	b = append(b, BuildingDef{
		Name: "Uranium Mine", Key: "uranium_mine",
		BaseCost: map[string]float64{"steel": 530 * Million, "iron": 240 * Million, "gold": 300 * Million},
		Effects: []Effect{
			{Type: "production", Target: "uranium", Value: 40.96},
			{Type: "production", Target: "stone", Value: 46600},
		},
		BuildTicks:     3600,
		RequiredAge:    "victorian_age",
		Description:    "Early uranium ore and stone extraction.",
		WorkerCapacity: 10,
	})
	// tier 10 — electric_age  rate=81.92  output=uranium
	b = append(b, BuildingDef{
		Name: "Nuclear Extraction Plant", Key: "nuclear_extraction_plant",
		BaseCost:       map[string]float64{"steel": 2.9 * Billion, "electricity": 1.1 * Billion},
		Effects:        []Effect{{Type: "production", Target: "uranium", Value: 81.92}},
		BuildTicks:     3600,
		RequiredAge:    "electric_age",
		Description:    "High-tech nuclear material extraction.",
		WorkerCapacity: 12,
	})
	// tier 11 — atomic_age  rate=163.84  output=uranium
	b = append(b, BuildingDef{
		Name: "Uranium Processing Works", Key: "uranium_processing_works",
		BaseCost: map[string]float64{"steel": 24.6 * Billion, "electricity": 9.85 * Billion},
		Effects: []Effect{
			{Type: "production", Target: "uranium", Value: 110000},
			{Type: "production", Target: "stone", Value: 1.04 * Million},
		},
		BuildTicks:     3600,
		RequiredAge:    "atomic_age",
		Description:    "Industrial uranium and stone processing.",
		WorkerCapacity: 12,
	})
	// tier 12 — modern_age  rate=327.68  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Titanium Mine", Key: "titanium_mine",
		BaseCost:       map[string]float64{"steel": 84 * Billion, "electricity": 30 * Billion, "data": 2.4 * Billion},
		Effects:        []Effect{{Type: "production", Target: "titanium_ore", Value: 9.17 * Million}},
		BuildTicks:     3600,
		RequiredAge:    "modern_age",
		Description:    "High-precision titanium ore extraction.",
		WorkerCapacity: 14,
	})
	// tier 13 — information_age  rate=655.36  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Precision Mine", Key: "precision_mine",
		BaseCost:       map[string]float64{"electricity": 210 * Billion, "data": 21 * Billion, "steel": 390 * Billion},
		Effects:        []Effect{{Type: "production", Target: "titanium_ore", Value: 72.9 * Million}},
		BuildTicks:     3600,
		RequiredAge:    "information_age",
		Description:    "AI-guided precision titanium mining.",
		WorkerCapacity: 15,
	})
	// tier 14 — digital_age  rate=1310.72  output=titanium_ore
	b = append(b, BuildingDef{
		Name: "Nano Drill Complex", Key: "nano_drill_complex",
		BaseCost:       map[string]float64{"electricity": 1.1 * Trillion, "data": 120 * Billion, "steel": 1.6 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "titanium_ore", Value: 111 * Million}},
		BuildTicks:     3600,
		RequiredAge:    "digital_age",
		Description:    "Nanoscale drilling for titanium ore.",
		WorkerCapacity: 16,
	})
	// tier 15 — cyberpunk_age  rate=2621.44  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Dark Crystal Mine", Key: "dark_crystal_mine",
		BaseCost:       map[string]float64{"data": 510 * Billion, "crypto": 3.3 * Trillion, "electricity": 6.6 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 435 * Million}},
		BuildTicks:     3600,
		RequiredAge:    "cyberpunk_age",
		Description:    "Extraction of dark matter crystals.",
		WorkerCapacity: 18,
	})
	// tier 16 — fusion_age  rate=5242.88  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Exotic Mineral Extractor", Key: "exotic_mineral_extractor",
		BaseCost:       map[string]float64{"plasma": 14 * Trillion, "electricity": 39 * Trillion, "steel": 51 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 922 * Million}},
		BuildTicks:     3600,
		RequiredAge:    "fusion_age",
		Description:    "Plasma-assisted exotic mineral extraction.",
		WorkerCapacity: 20,
	})
	// tier 17 — space_age  rate=10485.76  output=dark_matter_crystals
	b = append(b, BuildingDef{
		Name: "Asteroid Crystal Mine", Key: "asteroid_crystal_mine",
		BaseCost:       map[string]float64{"titanium": 210 * Trillion, "plasma": 96 * Trillion, "electricity": 260 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "dark_matter_crystals", Value: 7.84 * Billion}},
		BuildTicks:     3600,
		RequiredAge:    "space_age",
		Description:    "Asteroid belt dark matter crystal mining.",
		WorkerCapacity: 20,
	})
	// tier 18 — interstellar_age  rate=20971.52  output=antimatter
	b = append(b, BuildingDef{
		Name: "Stellar Core Drill", Key: "stellar_core_drill",
		BaseCost:       map[string]float64{"dark_matter": 260 * Trillion, "titanium": 2.1 * Quadrillion, "plasma": 1.3 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "antimatter", Value: 20971.52}},
		BuildTicks:     3600,
		RequiredAge:    "interstellar_age",
		RequiredTech:   "stellar_core_mining",
		Description:    "Drills into stellar cores for antimatter.",
		WorkerCapacity: 25,
	})
	// tier 19 — galactic_age  rate=41943.04  output=antimatter
	b = append(b, BuildingDef{
		Name: "Neutron Star Mine", Key: "neutron_star_mine",
		BaseCost:       map[string]float64{"antimatter": 510 * Trillion, "dark_matter": 2.6 * Quadrillion, "titanium": 13 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "antimatter", Value: 9.15 * Billion}},
		BuildTicks:     3600,
		RequiredAge:    "galactic_age",
		RequiredTech:   "neutron_mining",
		Description:    "Mining neutron stars for antimatter.",
		WorkerCapacity: 25,
	})
	// tier 20 — quantum_age  rate=83886.08  output=antimatter
	b = append(b, BuildingDef{
		Name: "Reality Excavator", Key: "reality_excavator",
		BaseCost: map[string]float64{"titanium": 5.71 * Quadrillion},
		Effects: []Effect{
			{Type: "production", Target: "antimatter", Value: 7.04 * Trillion},
		},
		BuildTicks:     3600,
		RequiredAge:    "quantum_age",
		Description:    "Excavates antimatter from the quantum foam.",
		WorkerCapacity: 30,
	})

	return finishLineage(b, lineageSpec{
		Key: "geological_extraction", Category: "production", Scale: ProducerRate, Domain: "masonry", Lines: true,
	})
}
