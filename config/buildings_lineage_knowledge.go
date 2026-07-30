package config

// newProductionBuildings2 returns lineages 5-9:
// knowledge, faith, military, trade, engineering.
// Merged into newProductionBuildings() via init — see buildings_new_merge.go.
func buildingsLineageKnowledge() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 5 — KNOWLEDGE (lineageKey: "knowledge", domain: "knowledge", output: "knowledge")
	// rate = 0.05 * 2^tier  CostScale: 1.30  Category: "research"
	// note: base raised 0.002 -> 0.05 (25x) to match the ~10x food/wood rate boost
	// plus catch-up — knowledge was ~250x slower than gathering_camp. See XtWUrrKY.
	// =========================================================================

	// tier 0 — primitive_age  rate=0.05
	b = append(b, BuildingDef{
		Name: "Story Circle", Key: "story_circle", Category: "research",
		BaseCost:    map[string]float64{"wood": 20},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.05}},
		BuildTicks:  80,
		RequiredAge: "primitive_age",
		Description: "Elders share stories around the fire. +0.05 knowledge/tick (2 workers).",
		LineageKey:  "knowledge", LineageTier: 0,
		WorkerDomain: "knowledge", WorkerCapacity: 2,
		EpochKey: "stone_era", OutputResource: "knowledge",
	})
	// tier 1 — stone_age  rate=0.1
	b = append(b, BuildingDef{
		Name: "Elders' Hall", Key: "elders_hall", Category: "research",
		BaseCost:    map[string]float64{"wood": 120, "stone": 80},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.1}},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "A hall where tribal elders convene. +0.1 knowledge/tick (2 workers).",
		LineageKey:  "knowledge", LineageTier: 1,
		WorkerDomain: "knowledge", WorkerCapacity: 2,
		EpochKey: "stone_era", OutputResource: "knowledge",
	})
	// tier 2 — bronze_age  rate=0.2
	b = append(b, BuildingDef{
		Name: "Scriptorium", Key: "scriptorium", Category: "research",
		BaseCost:    map[string]float64{"wood": 700, "stone": 400, "gold": 200},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.2}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "Scribes copy and preserve texts. +0.2 knowledge/tick (3 workers).",
		LineageKey:  "knowledge", LineageTier: 2,
		WorkerDomain: "knowledge", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "knowledge",
	})
	// tier 3 — iron_age  rate=0.4
	b = append(b, BuildingDef{
		Name: "Agora", Key: "agora", Category: "research",
		BaseCost:    map[string]float64{"stone": 5000, "gold": 2500, "iron": 1500},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.4}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "Open marketplace of ideas. +0.4 knowledge/tick (3 workers).",
		LineageKey:  "knowledge", LineageTier: 3,
		WorkerDomain: "knowledge", WorkerCapacity: 3,
		EpochKey: "iron_era", OutputResource: "knowledge",
	})
	// tier 4 — classical_age  rate=0.8
	b = append(b, BuildingDef{
		Name: "Library", Key: "library", Category: "research",
		BaseCost:    map[string]float64{"stone": 35000, "gold": 12000, "iron": 8000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.8}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Repository of written knowledge. +0.8 knowledge/tick (4 workers).",
		LineageKey:  "knowledge", LineageTier: 4,
		WorkerDomain: "knowledge", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "knowledge",
	})
	// tier 5 — medieval_age  rate=1.6
	b = append(b, BuildingDef{
		Name: "Monastery Library", Key: "monastery_library", Category: "research",
		BaseCost:    map[string]float64{"stone": 180000, "gold": 60000, "knowledge": 15000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 1.6}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "Monks preserve and copy scholarly works. +1.6 knowledge/tick (4 workers).",
		LineageKey:  "knowledge", LineageTier: 5,
		WorkerDomain: "knowledge", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "knowledge",
	})
	// tier 6 — renaissance_age  rate=3.2
	b = append(b, BuildingDef{
		Name: "University", Key: "university", Category: "research",
		BaseCost:    map[string]float64{"gold": 600000, "steel": 200000, "knowledge": 80000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 3.2}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "Higher learning for the intellectual elite. +3.2 knowledge/tick (5 workers).",
		LineageKey:  "knowledge", LineageTier: 6,
		WorkerDomain: "knowledge", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "knowledge",
	})
	// tier 7 — colonial_age  rate=6.4
	b = append(b, BuildingDef{
		Name: "Natural Philosophy Hall", Key: "natural_philosophy_hall", Category: "research",
		BaseCost:    map[string]float64{"gold": 3.5e6, "steel": 1.5e6, "knowledge": 300000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 6.4}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "Scientific inquiry into the natural world. +6.4 knowledge/tick (5 workers).",
		LineageKey:  "knowledge", LineageTier: 7,
		WorkerDomain: "knowledge", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "knowledge",
	})
	// tier 8 — industrial_age  rate=12.8
	b = append(b, BuildingDef{
		Name: "Research Institute", Key: "research_institute", Category: "research",
		BaseCost:    map[string]float64{"steel": 25e6, "coal": 8e6, "gold": 15e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 12.8}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "Formal scientific research institute. +12.8 knowledge/tick (6 workers).",
		LineageKey:  "knowledge", LineageTier: 8,
		WorkerDomain: "knowledge", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "knowledge",
	})
	// tier 9 — victorian_age  rate=25.6
	b = append(b, BuildingDef{
		Name: "Academy", Key: "academy", Category: "research",
		BaseCost:    map[string]float64{"steel": 175e6, "gold": 90e6, "iron": 60e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 25.6}},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "Victorian era academy of sciences. +25.6 knowledge/tick (6 workers).",
		LineageKey:  "knowledge", LineageTier: 9,
		WorkerDomain: "knowledge", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "knowledge",
	})
	// tier 10 — electric_age  rate=51.2
	b = append(b, BuildingDef{
		Name: "Physics Laboratory", Key: "physics_laboratory", Category: "research",
		BaseCost:    map[string]float64{"steel": 1e9, "electricity": 400e6, "gold": 600e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 51.2}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "Electrified physics research laboratory. +51.2 knowledge/tick (7 workers).",
		LineageKey:  "knowledge", LineageTier: 10,
		WorkerDomain: "knowledge", WorkerCapacity: 7,
		EpochKey: "electric_era", OutputResource: "knowledge",
	})
	// tier 11 — atomic_age  rate=102.4
	b = append(b, BuildingDef{
		Name: "Research Campus", Key: "research_campus", Category: "research",
		BaseCost:    map[string]float64{"steel": 5e9, "electricity": 2e9, "uranium": 300e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 102.4}},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "Multi-disciplinary atomic-age research campus. +102.4 knowledge/tick (7 workers).",
		LineageKey:  "knowledge", LineageTier: 11,
		WorkerDomain: "knowledge", WorkerCapacity: 7,
		EpochKey: "electric_era", OutputResource: "knowledge",
	})
	// tier 12 — modern_age  rate=204.8
	b = append(b, BuildingDef{
		Name: "Think Tank", Key: "think_tank", Category: "research",
		BaseCost:    map[string]float64{"steel": 30e9, "electricity": 10e9, "data": 1e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 204.8}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "Elite group of problem-solvers. +204.8 knowledge/tick (8 workers).",
		LineageKey:  "knowledge", LineageTier: 12,
		WorkerDomain: "knowledge", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "knowledge",
	})
	// tier 13 — information_age  rate=409.6
	b = append(b, BuildingDef{
		Name: "Innovation Hub", Key: "innovation_hub", Category: "research",
		BaseCost:    map[string]float64{"electricity": 80e9, "data": 8e9, "gold": 150e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 409.6}},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "Startup-style innovation accelerator. +409.6 knowledge/tick (8 workers).",
		LineageKey:  "knowledge", LineageTier: 13,
		WorkerDomain: "knowledge", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "knowledge",
	})
	// tier 14 — digital_age  rate=819.2
	b = append(b, BuildingDef{
		Name: "AI Research Lab", Key: "ai_research_lab", Category: "research",
		BaseCost:    map[string]float64{"electricity": 400e9, "data": 50e9, "steel": 600e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 819.2}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Artificial intelligence drives research. +819.2 knowledge/tick (10 workers).",
		LineageKey:  "knowledge", LineageTier: 14,
		WorkerDomain: "knowledge", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "knowledge",
	})
	// tier 15 — cyberpunk_age  rate=1638.4
	b = append(b, BuildingDef{
		Name: "Neuro Research Center", Key: "neuro_research_center", Category: "research",
		BaseCost:    map[string]float64{"data": 200e9, "crypto": 1e12, "electricity": 2e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 1638.4}},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "Neural-interface enhanced research. +1638.4 knowledge/tick (10 workers).",
		LineageKey:  "knowledge", LineageTier: 15,
		WorkerDomain: "knowledge", WorkerCapacity: 10,
		EpochKey: "neon_era", OutputResource: "knowledge",
	})
	// tier 16 — fusion_age  rate=3276.8
	b = append(b, BuildingDef{
		Name: "Theoretical Institute", Key: "theoretical_institute", Category: "research",
		BaseCost:    map[string]float64{"plasma": 5e12, "electricity": 15e12, "steel": 20e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 3276.8}},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "Theoretical physics at fusion-era scale. +3276.8 knowledge/tick (12 workers).",
		LineageKey:  "knowledge", LineageTier: 16,
		WorkerDomain: "knowledge", WorkerCapacity: 12,
		EpochKey: "neon_era", OutputResource: "knowledge",
	})
	// tier 17 — space_age  rate=6553.6
	b = append(b, BuildingDef{
		Name: "Deep Space Observatory", Key: "deep_space_observatory", Category: "research",
		BaseCost:    map[string]float64{"titanium": 80e12, "plasma": 40e12, "electricity": 100e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 6553.6}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "Observes the far reaches of the cosmos. +6553.6 knowledge/tick (12 workers).",
		LineageKey:  "knowledge", LineageTier: 17,
		WorkerDomain: "knowledge", WorkerCapacity: 12,
		EpochKey: "neon_era", OutputResource: "knowledge",
	})
	// tier 18 — interstellar_age  rate=13107.2
	b = append(b, BuildingDef{
		Name: "Xenology Institute", Key: "xenology_institute", Category: "research",
		BaseCost:    map[string]float64{"dark_matter": 100e12, "titanium": 800e12, "plasma": 500e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 13107.2}},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "Studies alien civilisations and xeno-science. +13107.2 knowledge/tick (15 workers).",
		LineageKey:  "knowledge", LineageTier: 18,
		WorkerDomain: "knowledge", WorkerCapacity: 15,
		EpochKey: "cosmic_era", OutputResource: "knowledge",
	})
	// tier 19 — galactic_age  rate=26214.4
	b = append(b, BuildingDef{
		Name: "Cosmic Research Station", Key: "cosmic_research_station", Category: "research",
		BaseCost:    map[string]float64{"antimatter": 200e12, "dark_matter": 1e15, "titanium": 5e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 26214.4}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Galactic-scale scientific research station. +26214.4 knowledge/tick (15 workers).",
		LineageKey:  "knowledge", LineageTier: 19,
		WorkerDomain: "knowledge", WorkerCapacity: 15,
		EpochKey: "cosmic_era", OutputResource: "knowledge",
	})
	// tier 20 — quantum_age  rate=52428.8
	b = append(b, BuildingDef{
		Name: "Reality Academy", Key: "reality_academy", Category: "research",
		BaseCost:    map[string]float64{"quantum_flux": 200e12, "antimatter": 60e15, "dark_matter": 50e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 52428.8}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Learns from the fabric of reality itself. +52428.8 knowledge/tick (20 workers).",
		LineageKey:  "knowledge", LineageTier: 20,
		WorkerDomain: "knowledge", WorkerCapacity: 20,
		EpochKey: "cosmic_era", OutputResource: "knowledge",
	})

	return b
}
