package config

// buildingsLineageFood returns all production, housing, research, and military buildings
// for the 13 lineage chains introduced in Phase 10 of the economy redesign.
// Storage buildings and wonders remain in baseBuildingsRaw().
// A row sets its own fields; the ones every row shares come from the lineageSpec at the foot,
// so no separate buildingMeta() entries are needed for these buildings.
func buildingsLineageFood() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 2 — FOOD (lineageKey: "food", domain: "food", output: "food")
	// rate = 0.05 * 2^tier  CostScale: ProducerRate (1.15)  Category: "production"  (both in the lineageSpec at the foot)
	// =========================================================================

	// tier 0 — primitive_age  rate=1.0 (doubled for the opening; food is the Primitive bottleneck)
	// Rate raised from 0.05 → 0.50 so gathering camps can sustain early workers.
	// BuildTicks lowered from 40 → 12 so food production comes online before starvation.
	b = append(b, BuildingDef{
		Name: "Gathering Camp", Key: "gathering_camp",
		BaseCost:       map[string]float64{"wood": 16},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 1}},
		BuildTicks:     12,
		RequiredAge:    "primitive_age",
		Description:    "Foragers gather berries and roots.",
		WorkerCapacity: 3,
	})
	// tier 1 — stone_age  rate=1.5 (raised with tier 0)
	// Rate raised from 0.10 → 1.00 to maintain ×2 per tier vs gathering_camp (0.50).
	// BuildTicks lowered from 100 → 30 proportional to gathering_camp reduction.
	b = append(b, BuildingDef{
		Name: "Forager Post", Key: "forager_post",
		BaseCost:       map[string]float64{"wood": 450, "stone": 240},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 1.5}},
		BuildTicks:     30,
		RequiredAge:    "stone_age",
		Description:    "Organized foraging post.",
		WorkerCapacity: 4,
	})
	// tier 2 — bronze_age  rate=2.00
	b = append(b, BuildingDef{
		Name: "Farm", Key: "farm",
		BaseCost:       map[string]float64{"wood": 2400, "stone": 1500},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 2}},
		BuildTicks:     150,
		RequiredAge:    "bronze_age",
		Description:    "Cultivated fields produce steady food.",
		WorkerCapacity: 5,
	})
	// tier 3 — iron_age  rate=4.00
	b = append(b, BuildingDef{
		Name: "Field Works", Key: "field_works",
		BaseCost:       map[string]float64{"stone": 7500, "marble": 7500, "iron": 3000, "iron_ore": 3000, "wood": 9000},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 4}},
		BuildTicks:     300,
		RequiredAge:    "iron_age",
		Description:    "Iron-tool farming with irrigation.",
		WorkerCapacity: 5,
	})
	// tier 4 — classical_age  rate=8.00
	b = append(b, BuildingDef{
		Name: "Terrace Farm", Key: "estate_farm",
		BaseCost:       map[string]float64{"stone": 55000, "marble": 55000, "gold": 36000, "iron": 15000, "iron_ore": 15000},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 8}},
		BuildTicks:     600,
		RequiredAge:    "classical_age",
		Description:    "A large estate with managed farmlands.",
		WorkerCapacity: 6,
	})
	// tier 5 — medieval_age  rate=16.00
	b = append(b, BuildingDef{
		Name: "Demesne", Key: "demesne",
		BaseCost:       map[string]float64{"stone": 900000, "gold": 300000},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 16}},
		BuildTicks:     1200,
		RequiredAge:    "medieval_age",
		Description:    "A lord's demesne with serfs and crop rotation.",
		WorkerCapacity: 6,
	})
	// tier 6 — renaissance_age  rate=32.00
	b = append(b, BuildingDef{
		Name: "Market Garden", Key: "market_garden",
		BaseCost:       map[string]float64{"gold": 2.93 * Million, "steel": 975000},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 32}},
		BuildTicks:     2400,
		RequiredAge:    "renaissance_age",
		Description:    "Scientific farming and market gardens.",
		WorkerCapacity: 7,
	})
	// tier 7 — colonial_age  rate=64.00
	b = append(b, BuildingDef{
		Name: "Plantation", Key: "plantation",
		BaseCost:       map[string]float64{"gold": 11 * Million, "steel": 4.5 * Million},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 64}},
		BuildTicks:     3600,
		RequiredAge:    "colonial_age",
		Description:    "Large-scale colonial plantation.",
		WorkerCapacity: 8,
	})
	// tier 8 — industrial_age  rate=128.00
	b = append(b, BuildingDef{
		Name: "Agricultural Works", Key: "agricultural_works",
		BaseCost:       map[string]float64{"steel": 75 * Million, "coal": 30 * Million, "gold": 45 * Million},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 128}},
		BuildTicks:     3600,
		RequiredAge:    "industrial_age",
		Description:    "Industrial-scale agricultural works.",
		WorkerCapacity: 10,
	})
	// tier 9 — victorian_age  rate=256.00
	b = append(b, BuildingDef{
		Name: "Mechanized Farm", Key: "mechanized_farm",
		BaseCost:       map[string]float64{"steel": 540 * Million, "oil": 240 * Million, "gold": 300 * Million},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 256}},
		BuildTicks:     3600,
		RequiredAge:    "victorian_age",
		Description:    "Steam and oil-powered mechanized farming.",
		WorkerCapacity: 10,
	})
	// tier 10 — electric_age  rate=512.00
	b = append(b, BuildingDef{
		Name: "Industrial Farm", Key: "industrial_farm",
		BaseCost:       map[string]float64{"steel": 3 * Billion, "electricity": 1.2 * Billion, "oil": 900 * Million},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 512}},
		BuildTicks:     3600,
		RequiredAge:    "electric_age",
		Description:    "Electrified industrial farming complex.",
		WorkerCapacity: 12,
	})
	// tier 11 — atomic_age  rate=1024.00
	b = append(b, BuildingDef{
		Name: "Agricultural Complex", Key: "agricultural_complex",
		BaseCost:       map[string]float64{"steel": 15 * Billion, "electricity": 6 * Billion, "uranium": 1.5 * Billion},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 1024}},
		BuildTicks:     3600,
		RequiredAge:    "atomic_age",
		Description:    "Atomic-age agricultural mega-complex.",
		WorkerCapacity: 12,
	})
	// tier 12 — modern_age  rate=2048.00
	b = append(b, BuildingDef{
		Name: "Agritech Campus", Key: "agri_complex",
		BaseCost:       map[string]float64{"steel": 45 * Billion, "titanium_ore": 45 * Billion, "electricity": 36 * Billion, "data": 3 * Billion},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 2048}},
		BuildTicks:     3600,
		RequiredAge:    "modern_age",
		Description:    "AI-optimized modern agriculture.",
		WorkerCapacity: 14,
	})
	// tier 13 — information_age  rate=4096.00
	b = append(b, BuildingDef{
		Name: "Smart Farm", Key: "smart_farm",
		BaseCost:       map[string]float64{"electricity": 240 * Billion, "data": 24 * Billion, "steel": 225 * Billion, "titanium_ore": 225 * Billion},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 4096}},
		BuildTicks:     3600,
		RequiredAge:    "information_age",
		RequiredTech:   "internet_of_things",
		Description:    "Sensor-driven smart farming.",
		WorkerCapacity: 15,
	})
	// tier 14 — digital_age  rate=8192.00
	b = append(b, BuildingDef{
		Name: "Nano Farm", Key: "nano_farm",
		BaseCost:       map[string]float64{"electricity": 1.2 * Trillion, "data": 150 * Billion, "steel": 900 * Billion, "titanium_ore": 900 * Billion},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 8192}},
		BuildTicks:     3600,
		RequiredAge:    "digital_age",
		Description:    "Nanotechnology-based food synthesis.",
		WorkerCapacity: 16,
	})
	// tier 15 — cyberpunk_age  rate=16384.00
	b = append(b, BuildingDef{
		Name: "Vat Farm", Key: "vat_farm",
		BaseCost:       map[string]float64{"data": 600 * Billion, "crypto": 3 * Trillion, "electricity": 3 * Trillion, "dark_matter_crystals": 3 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 16384}},
		BuildTicks:     3600,
		RequiredAge:    "cyberpunk_age",
		Description:    "Vat-grown protein synthesis at industrial scale.",
		WorkerCapacity: 18,
	})
	// tier 16 — fusion_age  rate=32768.00
	b = append(b, BuildingDef{
		Name: "Bio Reactor Farm", Key: "bio_reactor_farm",
		BaseCost:       map[string]float64{"plasma": 15 * Trillion, "electricity": 22.5 * Trillion, "dark_matter_crystals": 22.5 * Trillion, "steel": 60 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 32768}},
		BuildTicks:     3600,
		RequiredAge:    "fusion_age",
		Description:    "Plasma-powered bio reactor food production.",
		WorkerCapacity: 20,
	})
	// tier 17 — space_age  rate=65536.00
	b = append(b, BuildingDef{
		Name: "Hydroponic Bay", Key: "hydroponic_bay",
		BaseCost:       map[string]float64{"titanium": 240 * Trillion, "plasma": 120 * Trillion, "electricity": 150 * Trillion, "dark_matter_crystals": 150 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 65536}},
		BuildTicks:     3600,
		RequiredAge:    "space_age",
		Description:    "Zero-gravity hydroponic growing bays.",
		WorkerCapacity: 20,
	})
	// tier 18 — interstellar_age  rate=131072.00
	b = append(b, BuildingDef{
		Name: "Protein Synthesizer", Key: "protein_synthesizer",
		BaseCost:       map[string]float64{"dark_matter": 300 * Trillion, "titanium": 2.4 * Quadrillion, "plasma": 1.5 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 131072}},
		BuildTicks:     3600,
		RequiredAge:    "interstellar_age",
		Description:    "Matter-to-protein synthesizer.",
		WorkerCapacity: 25,
	})
	// tier 19 — galactic_age  rate=262144.00
	b = append(b, BuildingDef{
		Name: "Matter Converter", Key: "matter_converter",
		BaseCost:       map[string]float64{"antimatter": 600 * Trillion, "dark_matter": 3 * Quadrillion, "titanium": 15 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 262144}},
		BuildTicks:     3600,
		RequiredAge:    "galactic_age",
		Description:    "Converts raw matter into any food type.",
		WorkerCapacity: 25,
	})
	// tier 20 — quantum_age  rate=524288.00
	b = append(b, BuildingDef{
		Name: "Quantum Cultivator", Key: "quantum_cultivator",
		BaseCost:       map[string]float64{"quantum_flux": 600 * Trillion, "antimatter": 180 * Quadrillion, "dark_matter": 150 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "food", Value: 524288}},
		BuildTicks:     3600,
		RequiredAge:    "quantum_age",
		Description:    "Quantum probability manipulation to grow food.",
		WorkerCapacity: 30,
	})

	return finishLineage(b, lineageSpec{
		Key: "food", Category: "production", Scale: ProducerRate, Domain: "food",
	})
}
