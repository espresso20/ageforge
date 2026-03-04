package config

// newProductionBuildings2 returns lineages 5-9:
// knowledge, faith, military, trade, engineering.
// Merged into newProductionBuildings() via init — see buildings_new_merge.go.
func newProductionBuildings2() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 5 — KNOWLEDGE (lineageKey: "knowledge", domain: "knowledge", output: "knowledge")
	// rate = 0.002 * 2^tier  CostScale: 1.30  Category: "research"
	// =========================================================================

	// tier 0 — primitive_age  rate=0.002
	b = append(b, BuildingDef{
		Name: "Story Circle", Key: "story_circle", Category: "research",
		BaseCost:    map[string]float64{"wood": 20},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.002}},
		BuildTicks:  80,
		RequiredAge: "primitive_age",
		Description: "Elders share stories around the fire. +0.002 knowledge/tick (2 workers).",
		LineageKey: "knowledge", LineageTier: 0,
		WorkerDomain: "knowledge", WorkerCapacity: 2,
		EpochKey: "stone_era", OutputResource: "knowledge",
	})
	// tier 1 — stone_age  rate=0.004
	b = append(b, BuildingDef{
		Name: "Elders' Hall", Key: "elders_hall", Category: "research",
		BaseCost:    map[string]float64{"wood": 120, "stone": 80},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.004}},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "A hall where tribal elders convene. +0.004 knowledge/tick (2 workers).",
		LineageKey: "knowledge", LineageTier: 1,
		WorkerDomain: "knowledge", WorkerCapacity: 2,
		EpochKey: "stone_era", OutputResource: "knowledge",
	})
	// tier 2 — bronze_age  rate=0.008
	b = append(b, BuildingDef{
		Name: "Scriptorium", Key: "scriptorium", Category: "research",
		BaseCost:    map[string]float64{"wood": 700, "stone": 400, "gold": 200},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.008}},
		BuildTicks:  500,
		RequiredAge: "bronze_age",
		Description: "Scribes copy and preserve texts. +0.008 knowledge/tick (3 workers).",
		LineageKey: "knowledge", LineageTier: 2,
		WorkerDomain: "knowledge", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "knowledge",
	})
	// tier 3 — iron_age  rate=0.016
	b = append(b, BuildingDef{
		Name: "Agora", Key: "agora", Category: "research",
		BaseCost:    map[string]float64{"stone": 5000, "gold": 2500, "iron": 1500},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.016}},
		BuildTicks:  1000,
		RequiredAge: "iron_age",
		Description: "Open marketplace of ideas. +0.016 knowledge/tick (3 workers).",
		LineageKey: "knowledge", LineageTier: 3,
		WorkerDomain: "knowledge", WorkerCapacity: 3,
		EpochKey: "iron_era", OutputResource: "knowledge",
	})
	// tier 4 — classical_age  rate=0.032
	b = append(b, BuildingDef{
		Name: "Library", Key: "library", Category: "research",
		BaseCost:    map[string]float64{"stone": 35000, "gold": 12000, "iron": 8000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.032}},
		BuildTicks:  3000,
		RequiredAge: "classical_age",
		Description: "Repository of written knowledge. +0.032 knowledge/tick (4 workers).",
		LineageKey: "knowledge", LineageTier: 4,
		WorkerDomain: "knowledge", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "knowledge",
	})
	// tier 5 — medieval_age  rate=0.064
	b = append(b, BuildingDef{
		Name: "Monastery Library", Key: "monastery_library", Category: "research",
		BaseCost:    map[string]float64{"stone": 180000, "gold": 60000, "knowledge": 15000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.064}},
		BuildTicks:  6000,
		RequiredAge: "medieval_age",
		Description: "Monks preserve and copy scholarly works. +0.064 knowledge/tick (4 workers).",
		LineageKey: "knowledge", LineageTier: 5,
		WorkerDomain: "knowledge", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "knowledge",
	})
	// tier 6 — renaissance_age  rate=0.128
	b = append(b, BuildingDef{
		Name: "University", Key: "university", Category: "research",
		BaseCost:    map[string]float64{"gold": 600000, "steel": 200000, "knowledge": 80000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.128}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "Higher learning for the intellectual elite. +0.128 knowledge/tick (5 workers).",
		LineageKey: "knowledge", LineageTier: 6,
		WorkerDomain: "knowledge", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "knowledge",
	})
	// tier 7 — colonial_age  rate=0.256
	b = append(b, BuildingDef{
		Name: "Natural Philosophy Hall", Key: "natural_philosophy_hall", Category: "research",
		BaseCost:    map[string]float64{"gold": 3.5e6, "steel": 1.5e6, "knowledge": 300000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.256}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "Scientific inquiry into the natural world. +0.256 knowledge/tick (5 workers).",
		LineageKey: "knowledge", LineageTier: 7,
		WorkerDomain: "knowledge", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "knowledge",
	})
	// tier 8 — industrial_age  rate=0.512
	b = append(b, BuildingDef{
		Name: "Research Institute", Key: "research_institute", Category: "research",
		BaseCost:    map[string]float64{"steel": 25e6, "coal": 8e6, "gold": 15e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 0.512}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Formal scientific research institute. +0.512 knowledge/tick (6 workers).",
		LineageKey: "knowledge", LineageTier: 8,
		WorkerDomain: "knowledge", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "knowledge",
	})
	// tier 9 — victorian_age  rate=1.024
	b = append(b, BuildingDef{
		Name: "Academy", Key: "academy", Category: "research",
		BaseCost:    map[string]float64{"steel": 175e6, "gold": 90e6, "iron": 60e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 1.024}},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "Victorian era academy of sciences. +1.024 knowledge/tick (6 workers).",
		LineageKey: "knowledge", LineageTier: 9,
		WorkerDomain: "knowledge", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "knowledge",
	})
	// tier 10 — electric_age  rate=2.048
	b = append(b, BuildingDef{
		Name: "Physics Laboratory", Key: "physics_laboratory", Category: "research",
		BaseCost:    map[string]float64{"steel": 1e9, "electricity": 400e6, "gold": 600e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 2.048}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "Electrified physics research laboratory. +2.048 knowledge/tick (7 workers).",
		LineageKey: "knowledge", LineageTier: 10,
		WorkerDomain: "knowledge", WorkerCapacity: 7,
		EpochKey: "electric_era", OutputResource: "knowledge",
	})
	// tier 11 — atomic_age  rate=4.096
	b = append(b, BuildingDef{
		Name: "Research Campus", Key: "research_campus", Category: "research",
		BaseCost:    map[string]float64{"steel": 5e9, "electricity": 2e9, "uranium": 300e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 4.096}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Multi-disciplinary atomic-age research campus. +4.096 knowledge/tick (7 workers).",
		LineageKey: "knowledge", LineageTier: 11,
		WorkerDomain: "knowledge", WorkerCapacity: 7,
		EpochKey: "electric_era", OutputResource: "knowledge",
	})
	// tier 12 — modern_age  rate=8.192
	b = append(b, BuildingDef{
		Name: "Think Tank", Key: "think_tank", Category: "research",
		BaseCost:    map[string]float64{"steel": 30e9, "electricity": 10e9, "data": 1e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 8.192}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Elite group of problem-solvers. +8.192 knowledge/tick (8 workers).",
		LineageKey: "knowledge", LineageTier: 12,
		WorkerDomain: "knowledge", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "knowledge",
	})
	// tier 13 — information_age  rate=16.384
	b = append(b, BuildingDef{
		Name: "Innovation Hub", Key: "innovation_hub", Category: "research",
		BaseCost:    map[string]float64{"electricity": 80e9, "data": 8e9, "gold": 150e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 16.384}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "Startup-style innovation accelerator. +16.384 knowledge/tick (8 workers).",
		LineageKey: "knowledge", LineageTier: 13,
		WorkerDomain: "knowledge", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "knowledge",
	})
	// tier 14 — digital_age  rate=32.768
	b = append(b, BuildingDef{
		Name: "AI Research Lab", Key: "ai_research_lab", Category: "research",
		BaseCost:    map[string]float64{"electricity": 400e9, "data": 50e9, "steel": 600e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 32.768}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Artificial intelligence drives research. +32.768 knowledge/tick (10 workers).",
		LineageKey: "knowledge", LineageTier: 14,
		WorkerDomain: "knowledge", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "knowledge",
	})
	// tier 15 — cyberpunk_age  rate=65.536
	b = append(b, BuildingDef{
		Name: "Neuro Research Center", Key: "neuro_research_center", Category: "research",
		BaseCost:    map[string]float64{"data": 200e9, "crypto": 1e12, "electricity": 2e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 65.536}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Neural-interface enhanced research. +65.536 knowledge/tick (10 workers).",
		LineageKey: "knowledge", LineageTier: 15,
		WorkerDomain: "knowledge", WorkerCapacity: 10,
		EpochKey: "neon_era", OutputResource: "knowledge",
	})
	// tier 16 — fusion_age  rate=131.072
	b = append(b, BuildingDef{
		Name: "Theoretical Institute", Key: "theoretical_institute", Category: "research",
		BaseCost:    map[string]float64{"plasma": 5e12, "electricity": 15e12, "steel": 20e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 131.072}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Theoretical physics at fusion-era scale. +131.072 knowledge/tick (12 workers).",
		LineageKey: "knowledge", LineageTier: 16,
		WorkerDomain: "knowledge", WorkerCapacity: 12,
		EpochKey: "neon_era", OutputResource: "knowledge",
	})
	// tier 17 — space_age  rate=262.144
	b = append(b, BuildingDef{
		Name: "Deep Space Observatory", Key: "deep_space_observatory", Category: "research",
		BaseCost:    map[string]float64{"titanium": 80e12, "plasma": 40e12, "electricity": 100e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 262.144}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Observes the far reaches of the cosmos. +262.144 knowledge/tick (12 workers).",
		LineageKey: "knowledge", LineageTier: 17,
		WorkerDomain: "knowledge", WorkerCapacity: 12,
		EpochKey: "neon_era", OutputResource: "knowledge",
	})
	// tier 18 — interstellar_age  rate=524.288
	b = append(b, BuildingDef{
		Name: "Xenology Institute", Key: "xenology_institute", Category: "research",
		BaseCost:    map[string]float64{"dark_matter": 100e12, "titanium": 800e12, "plasma": 500e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 524.288}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Studies alien civilisations and xeno-science. +524.288 knowledge/tick (15 workers).",
		LineageKey: "knowledge", LineageTier: 18,
		WorkerDomain: "knowledge", WorkerCapacity: 15,
		EpochKey: "cosmic_era", OutputResource: "knowledge",
	})
	// tier 19 — galactic_age  rate=1048.576
	b = append(b, BuildingDef{
		Name: "Cosmic Research Station", Key: "cosmic_research_station", Category: "research",
		BaseCost:    map[string]float64{"antimatter": 200e12, "dark_matter": 1e15, "titanium": 5e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 1048.576}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Galactic-scale scientific research station. +1048.576 knowledge/tick (15 workers).",
		LineageKey: "knowledge", LineageTier: 19,
		WorkerDomain: "knowledge", WorkerCapacity: 15,
		EpochKey: "cosmic_era", OutputResource: "knowledge",
	})
	// tier 20 — quantum_age  rate=2097.152
	b = append(b, BuildingDef{
		Name: "Reality Academy", Key: "reality_academy", Category: "research",
		BaseCost:    map[string]float64{"quantum_flux": 200e12, "antimatter": 60e15, "dark_matter": 50e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "knowledge", Value: 2097.152}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Learns from the fabric of reality itself. +2097.152 knowledge/tick (20 workers).",
		LineageKey: "knowledge", LineageTier: 20,
		WorkerDomain: "knowledge", WorkerCapacity: 20,
		EpochKey: "cosmic_era", OutputResource: "knowledge",
	})

	// =========================================================================
	// LINEAGE 6 — FAITH (lineageKey: "faith", domain: "faith", output: "faith")
	// rate = 0.002 * 2^tier  CostScale: 1.30  Category: "research"
	// =========================================================================

	// tier 0 — primitive_age  rate=0.002
	b = append(b, BuildingDef{
		Name: "Shrine", Key: "shrine", Category: "research",
		BaseCost:    map[string]float64{"wood": 20},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.002}},
		BuildTicks:  80,
		RequiredAge: "primitive_age",
		Description: "A small spirit shrine. +0.002 faith/tick (2 workers).",
		LineageKey: "faith", LineageTier: 0,
		WorkerDomain: "faith", WorkerCapacity: 2,
		EpochKey: "stone_era", OutputResource: "faith",
	})
	// tier 1 — stone_age  rate=0.004
	b = append(b, BuildingDef{
		Name: "Standing Stones", Key: "standing_stones", Category: "research",
		BaseCost:    map[string]float64{"wood": 100, "stone": 80},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.004}},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "Monolithic stones with ritual significance. +0.004 faith/tick (2 workers).",
		LineageKey: "faith", LineageTier: 1,
		WorkerDomain: "faith", WorkerCapacity: 2,
		EpochKey: "stone_era", OutputResource: "faith",
	})
	// tier 2 — bronze_age  rate=0.008
	b = append(b, BuildingDef{
		Name: "Altar", Key: "altar", Category: "research",
		BaseCost:    map[string]float64{"wood": 700, "stone": 350, "gold": 150},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.008}},
		BuildTicks:  500,
		RequiredAge: "bronze_age",
		Description: "A sacred altar for offerings. +0.008 faith/tick (3 workers).",
		LineageKey: "faith", LineageTier: 2,
		WorkerDomain: "faith", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "faith",
	})
	// tier 3 — iron_age  rate=0.016
	b = append(b, BuildingDef{
		Name: "Temple", Key: "temple", Category: "research",
		BaseCost:    map[string]float64{"stone": 5000, "gold": 2000, "iron": 1000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.016}},
		BuildTicks:  1000,
		RequiredAge: "iron_age",
		Description: "A formal temple for organised worship. +0.016 faith/tick (3 workers).",
		LineageKey: "faith", LineageTier: 3,
		WorkerDomain: "faith", WorkerCapacity: 3,
		EpochKey: "iron_era", OutputResource: "faith",
	})
	// tier 4 — classical_age  rate=0.032
	b = append(b, BuildingDef{
		Name: "Oracle House", Key: "oracle_house", Category: "research",
		BaseCost:    map[string]float64{"stone": 30000, "gold": 10000, "iron": 5000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.032}},
		BuildTicks:  3000,
		RequiredAge: "classical_age",
		Description: "Oracles speak for the gods. +0.032 faith/tick (4 workers).",
		LineageKey: "faith", LineageTier: 4,
		WorkerDomain: "faith", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "faith",
	})
	// tier 5 — medieval_age  rate=0.064
	b = append(b, BuildingDef{
		Name: "Cathedral", Key: "cathedral", Category: "research",
		BaseCost:    map[string]float64{"stone": 200000, "gold": 65000, "knowledge": 15000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.064}},
		BuildTicks:  6000,
		RequiredAge: "medieval_age",
		Description: "A towering medieval cathedral. +0.064 faith/tick (5 workers).",
		LineageKey: "faith", LineageTier: 5,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "faith",
	})
	// tier 6 — renaissance_age  rate=0.128
	b = append(b, BuildingDef{
		Name: "Basilica", Key: "basilica", Category: "research",
		BaseCost:    map[string]float64{"gold": 650000, "stone": 400000, "knowledge": 100000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.128}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "A grand renaissance basilica. +0.128 faith/tick (5 workers).",
		LineageKey: "faith", LineageTier: 6,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "faith",
	})
	// tier 7 — colonial_age  rate=0.256
	b = append(b, BuildingDef{
		Name: "Mission", Key: "mission", Category: "research",
		BaseCost:    map[string]float64{"gold": 3e6, "stone": 2e6, "knowledge": 400000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.256}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "A colonial mission spreading faith. +0.256 faith/tick (5 workers).",
		LineageKey: "faith", LineageTier: 7,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "faith",
	})
	// tier 8 — industrial_age  rate=0.512
	b = append(b, BuildingDef{
		Name: "Church", Key: "church", Category: "research",
		BaseCost:    map[string]float64{"stone": 20e6, "gold": 10e6, "iron": 8e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.512}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "An industrial-age parish church. +0.512 faith/tick (5 workers).",
		LineageKey: "faith", LineageTier: 8,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "faith",
	})
	// tier 9 — victorian_age  rate=1.024
	b = append(b, BuildingDef{
		Name: "Grand Cathedral", Key: "grand_cathedral", Category: "research",
		BaseCost:    map[string]float64{"steel": 170e6, "gold": 90e6, "stone": 120e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 1.024}},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "A vast Victorian grand cathedral. +1.024 faith/tick (6 workers).",
		LineageKey: "faith", LineageTier: 9,
		WorkerDomain: "faith", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "faith",
	})
	// tier 10 — electric_age  rate=2.048
	b = append(b, BuildingDef{
		Name: "Revival Hall", Key: "revival_hall", Category: "research",
		BaseCost:    map[string]float64{"steel": 1e9, "electricity": 350e6, "gold": 500e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 2.048}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "Electric revival meetings spread spiritual fervour. +2.048 faith/tick (6 workers).",
		LineageKey: "faith", LineageTier: 10,
		WorkerDomain: "faith", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "faith",
	})
	// tier 11 — atomic_age  rate=4.096
	b = append(b, BuildingDef{
		Name: "Spiritual Center", Key: "spiritual_center", Category: "research",
		BaseCost:    map[string]float64{"steel": 5e9, "electricity": 2e9, "gold": 3e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 4.096}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Atomic-age spiritual wellness center. +4.096 faith/tick (6 workers).",
		LineageKey: "faith", LineageTier: 11,
		WorkerDomain: "faith", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "faith",
	})
	// tier 12 — modern_age  rate=8.192
	b = append(b, BuildingDef{
		Name: "Meditation Center", Key: "meditation_center", Category: "research",
		BaseCost:    map[string]float64{"steel": 28e9, "electricity": 10e9, "gold": 20e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 8.192}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Modern meditation and mindfulness hub. +8.192 faith/tick (7 workers).",
		LineageKey: "faith", LineageTier: 12,
		WorkerDomain: "faith", WorkerCapacity: 7,
		EpochKey: "digital_era", OutputResource: "faith",
	})
	// tier 13 — information_age  rate=16.384
	b = append(b, BuildingDef{
		Name: "Digital Temple", Key: "digital_temple", Category: "research",
		BaseCost:    map[string]float64{"electricity": 80e9, "data": 7e9, "gold": 130e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 16.384}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "A virtual spiritual sanctuary. +16.384 faith/tick (7 workers).",
		LineageKey: "faith", LineageTier: 13,
		WorkerDomain: "faith", WorkerCapacity: 7,
		EpochKey: "digital_era", OutputResource: "faith",
	})
	// tier 14 — digital_age  rate=32.768
	b = append(b, BuildingDef{
		Name: "Cyber Shrine", Key: "cyber_shrine", Category: "research",
		BaseCost:    map[string]float64{"electricity": 380e9, "data": 45e9, "crypto": 200e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 32.768}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "A cybernetic devotional shrine. +32.768 faith/tick (8 workers).",
		LineageKey: "faith", LineageTier: 14,
		WorkerDomain: "faith", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "faith",
	})
	// tier 15 — cyberpunk_age  rate=65.536
	b = append(b, BuildingDef{
		Name: "Neon Sanctuary", Key: "neon_sanctuary", Category: "research",
		BaseCost:    map[string]float64{"data": 170e9, "crypto": 900e9, "electricity": 1.8e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 65.536}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Neon-lit cyberpunk sanctuary. +65.536 faith/tick (8 workers).",
		LineageKey: "faith", LineageTier: 15,
		WorkerDomain: "faith", WorkerCapacity: 8,
		EpochKey: "neon_era", OutputResource: "faith",
	})
	// tier 16 — fusion_age  rate=131.072
	b = append(b, BuildingDef{
		Name: "Quantum Chapel", Key: "quantum_chapel", Category: "research",
		BaseCost:    map[string]float64{"plasma": 4e12, "electricity": 13e12, "steel": 17e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 131.072}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "A chapel resonating with quantum energies. +131.072 faith/tick (9 workers).",
		LineageKey: "faith", LineageTier: 16,
		WorkerDomain: "faith", WorkerCapacity: 9,
		EpochKey: "neon_era", OutputResource: "faith",
	})
	// tier 17 — space_age  rate=262.144
	b = append(b, BuildingDef{
		Name: "Orbital Sanctuary", Key: "orbital_sanctuary", Category: "research",
		BaseCost:    map[string]float64{"titanium": 70e12, "plasma": 30e12, "electricity": 90e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 262.144}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "A faith sanctuary in orbital space. +262.144 faith/tick (9 workers).",
		LineageKey: "faith", LineageTier: 17,
		WorkerDomain: "faith", WorkerCapacity: 9,
		EpochKey: "neon_era", OutputResource: "faith",
	})
	// tier 18 — interstellar_age  rate=524.288
	b = append(b, BuildingDef{
		Name: "Void Monastery", Key: "void_monastery", Category: "research",
		BaseCost:    map[string]float64{"dark_matter": 90e12, "titanium": 720e12, "plasma": 440e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 524.288}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "A monastery floating in the interstellar void. +524.288 faith/tick (10 workers).",
		LineageKey: "faith", LineageTier: 18,
		WorkerDomain: "faith", WorkerCapacity: 10,
		EpochKey: "cosmic_era", OutputResource: "faith",
	})
	// tier 19 — galactic_age  rate=1048.576
	b = append(b, BuildingDef{
		Name: "Stellar Shrine", Key: "stellar_shrine", Category: "research",
		BaseCost:    map[string]float64{"antimatter": 180e12, "dark_matter": 900e12, "titanium": 4.5e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 1048.576}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "A galactic-scale stellar shrine. +1048.576 faith/tick (10 workers).",
		LineageKey: "faith", LineageTier: 19,
		WorkerDomain: "faith", WorkerCapacity: 10,
		EpochKey: "cosmic_era", OutputResource: "faith",
	})
	// tier 20 — quantum_age  rate=2097.152
	b = append(b, BuildingDef{
		Name: "Transcendence Hall", Key: "transcendence_hall", Category: "research",
		BaseCost:    map[string]float64{"quantum_flux": 190e12, "antimatter": 58e15, "dark_matter": 47e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 2097.152}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "A hall dedicated to transcendence beyond existence. +2097.152 faith/tick (12 workers).",
		LineageKey: "faith", LineageTier: 20,
		WorkerDomain: "faith", WorkerCapacity: 12,
		EpochKey: "cosmic_era", OutputResource: "faith",
	})

	// =========================================================================
	// LINEAGE 7 — MILITARY (lineageKey: "military", domain: "military")
	// Effect type: "capacity", Target: "military", Value: 10 * 2^tier
	// CostScale: 1.35  Category: "military"
	// =========================================================================

	// tier 0 — primitive_age  soldiers=10
	b = append(b, BuildingDef{
		Name: "Hunting Lodge", Key: "hunting_lodge", Category: "military",
		BaseCost:    map[string]float64{"wood": 25},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 10}},
		BuildTicks:  80,
		RequiredAge: "primitive_age",
		Description: "A gathering place for hunters. +10 military cap (3 workers).",
		LineageKey: "military", LineageTier: 0,
		WorkerDomain: "military", WorkerCapacity: 3,
		EpochKey: "stone_era",
	})
	// tier 1 — stone_age  soldiers=20
	b = append(b, BuildingDef{
		Name: "War Camp", Key: "war_camp", Category: "military",
		BaseCost:    map[string]float64{"wood": 180, "stone": 100},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 20}},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "A fortified war camp. +20 military cap (4 workers).",
		LineageKey: "military", LineageTier: 1,
		WorkerDomain: "military", WorkerCapacity: 4,
		EpochKey: "stone_era",
	})
	// tier 2 — bronze_age  soldiers=40
	b = append(b, BuildingDef{
		Name: "Barracks", Key: "barracks", Category: "military",
		BaseCost:    map[string]float64{"wood": 900, "stone": 600, "iron": 200},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 40}},
		BuildTicks:  500,
		RequiredAge: "bronze_age",
		Description: "Trains and houses soldiers. +40 military cap (5 workers).",
		LineageKey: "military", LineageTier: 2,
		WorkerDomain: "military", WorkerCapacity: 5,
		EpochKey: "stone_era",
	})
	// tier 3 — iron_age  soldiers=80
	b = append(b, BuildingDef{
		Name: "Legion Fort", Key: "legion_fort", Category: "military",
		BaseCost:    map[string]float64{"stone": 7000, "iron": 3500, "gold": 2000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 80}},
		BuildTicks:  1000,
		RequiredAge: "iron_age",
		Description: "A fortified roman-style legion camp. +80 military cap (6 workers).",
		LineageKey: "military", LineageTier: 3,
		WorkerDomain: "military", WorkerCapacity: 6,
		EpochKey: "iron_era",
	})
	// tier 4 — classical_age  soldiers=160
	b = append(b, BuildingDef{
		Name: "Military Academy", Key: "military_academy", Category: "military",
		BaseCost:    map[string]float64{"stone": 40000, "gold": 15000, "iron": 10000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 160}},
		BuildTicks:  3000,
		RequiredAge: "classical_age",
		Description: "Trains elite military officers. +160 military cap (6 workers).",
		LineageKey: "military", LineageTier: 4,
		WorkerDomain: "military", WorkerCapacity: 6,
		EpochKey: "iron_era",
	})
	// tier 5 — medieval_age  soldiers=320
	b = append(b, BuildingDef{
		Name: "Castle Keep", Key: "castle_keep", Category: "military",
		BaseCost:    map[string]float64{"stone": 220000, "iron": 70000, "gold": 50000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 320}},
		BuildTicks:  6000,
		RequiredAge: "medieval_age",
		Description: "A fortified stone keep. +320 military cap (7 workers).",
		LineageKey: "military", LineageTier: 5,
		WorkerDomain: "military", WorkerCapacity: 7,
		EpochKey: "iron_era",
	})
	// tier 6 — renaissance_age  soldiers=640
	b = append(b, BuildingDef{
		Name: "Fortress", Key: "fortress", Category: "military",
		BaseCost:    map[string]float64{"stone": 700000, "gold": 300000, "steel": 150000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 640}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "A star-fort capable of holding a large garrison. +640 military cap (7 workers).",
		LineageKey: "military", LineageTier: 6,
		WorkerDomain: "military", WorkerCapacity: 7,
		EpochKey: "steel_era",
	})
	// tier 7 — colonial_age  soldiers=1280
	b = append(b, BuildingDef{
		Name: "Fort", Key: "fort", Category: "military",
		BaseCost:    map[string]float64{"gold": 4e6, "steel": 2e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 1280}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "A colonial frontier fort. +1280 military cap (8 workers).",
		LineageKey: "military", LineageTier: 7,
		WorkerDomain: "military", WorkerCapacity: 8,
		EpochKey: "steel_era",
	})
	// tier 8 — industrial_age  soldiers=2560
	b = append(b, BuildingDef{
		Name: "Military Base", Key: "military_base", Category: "military",
		BaseCost:    map[string]float64{"steel": 30e6, "coal": 10e6, "gold": 15e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 2560}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "An industrial-era military base. +2560 military cap (10 workers).",
		LineageKey: "military", LineageTier: 8,
		WorkerDomain: "military", WorkerCapacity: 10,
		EpochKey: "steel_era",
	})
	// tier 9 — victorian_age  soldiers=5120
	b = append(b, BuildingDef{
		Name: "Garrison", Key: "garrison", Category: "military",
		BaseCost:    map[string]float64{"steel": 200e6, "iron": 100e6, "gold": 120e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 5120}},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "A Victorian-era garrison town. +5120 military cap (10 workers).",
		LineageKey: "military", LineageTier: 9,
		WorkerDomain: "military", WorkerCapacity: 10,
		EpochKey: "electric_era",
	})
	// tier 10 — electric_age  soldiers=10240
	b = append(b, BuildingDef{
		Name: "Command Post", Key: "command_post", Category: "military",
		BaseCost:    map[string]float64{"steel": 1.2e9, "electricity": 500e6, "gold": 700e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 10240}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "An electrified command and control post. +10240 military cap (12 workers).",
		LineageKey: "military", LineageTier: 10,
		WorkerDomain: "military", WorkerCapacity: 12,
		EpochKey: "electric_era",
	})
	// tier 11 — atomic_age  soldiers=20480
	b = append(b, BuildingDef{
		Name: "Bunker Complex", Key: "bunker_complex", Category: "military",
		BaseCost:    map[string]float64{"steel": 6e9, "stone": 8e9, "electricity": 2e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 20480}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "A hardened atomic-era bunker complex. +20480 military cap (12 workers).",
		LineageKey: "military", LineageTier: 11,
		WorkerDomain: "military", WorkerCapacity: 12,
		EpochKey: "electric_era",
	})
	// tier 12 — modern_age  soldiers=40960
	b = append(b, BuildingDef{
		Name: "Special Ops HQ", Key: "special_ops_hq", Category: "military",
		BaseCost:    map[string]float64{"steel": 35e9, "electricity": 12e9, "data": 1e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 40960}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Headquarters for special operations forces. +40960 military cap (14 workers).",
		LineageKey: "military", LineageTier: 12,
		WorkerDomain: "military", WorkerCapacity: 14,
		EpochKey: "digital_era",
	})
	// tier 13 — information_age  soldiers=81920
	b = append(b, BuildingDef{
		Name: "Cyber Command", Key: "cyber_command", Category: "military",
		BaseCost:    map[string]float64{"electricity": 90e9, "data": 10e9, "gold": 160e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 81920}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "Cyber warfare command centre. +81920 military cap (15 workers).",
		LineageKey: "military", LineageTier: 13,
		WorkerDomain: "military", WorkerCapacity: 15,
		EpochKey: "digital_era",
	})
	// tier 14 — digital_age  soldiers=163840
	b = append(b, BuildingDef{
		Name: "Drone Warfare Center", Key: "drone_warfare_center", Category: "military",
		BaseCost:    map[string]float64{"electricity": 450e9, "data": 55e9, "steel": 650e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 163840}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Autonomous drone warfare command. +163840 military cap (16 workers).",
		LineageKey: "military", LineageTier: 14,
		WorkerDomain: "military", WorkerCapacity: 16,
		EpochKey: "digital_era",
	})
	// tier 15 — cyberpunk_age  soldiers=327680
	b = append(b, BuildingDef{
		Name: "Combat Aug Center", Key: "combat_aug_center", Category: "military",
		BaseCost:    map[string]float64{"data": 210e9, "crypto": 1.1e12, "electricity": 2.2e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 327680}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Cybernetic augmentation for soldiers. +327680 military cap (18 workers).",
		LineageKey: "military", LineageTier: 15,
		WorkerDomain: "military", WorkerCapacity: 18,
		EpochKey: "neon_era",
	})
	// tier 16 — fusion_age  soldiers=655360
	b = append(b, BuildingDef{
		Name: "Plasma Command", Key: "plasma_command", Category: "military",
		BaseCost:    map[string]float64{"plasma": 5e12, "electricity": 14e12, "steel": 18e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 655360}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Plasma-weapon equipped military command. +655360 military cap (20 workers).",
		LineageKey: "military", LineageTier: 16,
		WorkerDomain: "military", WorkerCapacity: 20,
		EpochKey: "neon_era",
	})
	// tier 17 — space_age  soldiers=1310720
	b = append(b, BuildingDef{
		Name: "Space Force Base", Key: "space_force_base", Category: "military",
		BaseCost:    map[string]float64{"titanium": 80e12, "plasma": 38e12, "electricity": 95e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 1310720}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "An orbital space force base. +1310720 military cap (20 workers).",
		LineageKey: "military", LineageTier: 17,
		WorkerDomain: "military", WorkerCapacity: 20,
		EpochKey: "neon_era",
	})
	// tier 18 — interstellar_age  soldiers=2621440
	b = append(b, BuildingDef{
		Name: "Fleet Command", Key: "fleet_command", Category: "military",
		BaseCost:    map[string]float64{"dark_matter": 95e12, "titanium": 760e12, "plasma": 460e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 2621440}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Commands a full interstellar fleet. +2621440 military cap (25 workers).",
		LineageKey: "military", LineageTier: 18,
		WorkerDomain: "military", WorkerCapacity: 25,
		EpochKey: "cosmic_era",
	})
	// tier 19 — galactic_age  soldiers=5242880
	b = append(b, BuildingDef{
		Name: "Stellar Armada HQ", Key: "stellar_armada_hq", Category: "military",
		BaseCost:    map[string]float64{"antimatter": 190e12, "dark_matter": 950e12, "titanium": 4.8e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 5242880}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Headquarters for the galactic armada. +5242880 military cap (25 workers).",
		LineageKey: "military", LineageTier: 19,
		WorkerDomain: "military", WorkerCapacity: 25,
		EpochKey: "cosmic_era",
	})
	// tier 20 — quantum_age  soldiers=10485760
	b = append(b, BuildingDef{
		Name: "Probability War Room", Key: "probability_war_room", Category: "military",
		BaseCost:    map[string]float64{"quantum_flux": 210e12, "antimatter": 62e15, "dark_matter": 52e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 10485760}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Wages war across probability timelines. +10485760 military cap (30 workers).",
		LineageKey: "military", LineageTier: 20,
		WorkerDomain: "military", WorkerCapacity: 30,
		EpochKey: "cosmic_era",
	})

	// =========================================================================
	// LINEAGE 8 — TRADE (lineageKey: "trade", domain: "trade", output: "gold")
	// starts at bronze_age (tier 0)
	// rate = 0.05 * 2^tier  CostScale: 1.40  Category: "production"
	// =========================================================================

	// tier 0 — bronze_age  rate=0.05
	b = append(b, BuildingDef{
		Name: "Market", Key: "market", Category: "production",
		BaseCost:    map[string]float64{"wood": 1000, "stone": 700, "iron": 200},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.05}},
		BuildTicks:  500,
		RequiredAge: "bronze_age",
		Description: "A trading market for barter and coin. +0.05 gold/tick (3 workers).",
		LineageKey: "trade", LineageTier: 0,
		WorkerDomain: "trade", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "gold",
	})
	// tier 1 — iron_age  rate=0.10
	b = append(b, BuildingDef{
		Name: "Trading Post", Key: "trading_post", Category: "production",
		BaseCost:    map[string]float64{"stone": 5500, "iron": 2500, "gold": 1500},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.10}},
		BuildTicks:  1000,
		RequiredAge: "iron_age",
		Description: "A regional trading post. +0.10 gold/tick (3 workers).",
		LineageKey: "trade", LineageTier: 1,
		WorkerDomain: "trade", WorkerCapacity: 3,
		EpochKey: "iron_era", OutputResource: "gold",
	})
	// tier 2 — classical_age  rate=0.20
	b = append(b, BuildingDef{
		Name: "Merchant Quarter", Key: "merchant_quarter", Category: "production",
		BaseCost:    map[string]float64{"stone": 35000, "gold": 15000, "iron": 10000},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.20}},
		BuildTicks:  3000,
		RequiredAge: "classical_age",
		Description: "An urban merchant district. +0.20 gold/tick (4 workers).",
		LineageKey: "trade", LineageTier: 2,
		WorkerDomain: "trade", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "gold",
	})
	// tier 3 — medieval_age  rate=0.40
	b = append(b, BuildingDef{
		Name: "Guildhall", Key: "guildhall", Category: "production",
		BaseCost:    map[string]float64{"stone": 200000, "gold": 70000, "knowledge": 20000},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.40}},
		BuildTicks:  6000,
		RequiredAge: "medieval_age",
		Description: "Merchant guild organises regional trade. +0.40 gold/tick (4 workers).",
		LineageKey: "trade", LineageTier: 3,
		WorkerDomain: "trade", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "gold",
	})
	// tier 4 — renaissance_age  rate=0.80
	b = append(b, BuildingDef{
		Name: "Exchange", Key: "exchange", Category: "production",
		BaseCost:    map[string]float64{"gold": 700000, "steel": 250000, "knowledge": 100000},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.80}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "A commodity exchange for international trade. +0.80 gold/tick (5 workers).",
		LineageKey: "trade", LineageTier: 4,
		WorkerDomain: "trade", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "gold",
	})
	// tier 5 — colonial_age  rate=1.60
	b = append(b, BuildingDef{
		Name: "Port", Key: "port", Category: "production",
		BaseCost:    map[string]float64{"gold": 4e6, "steel": 2e6},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 1.60}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "A colonial maritime trade port. +1.60 gold/tick (5 workers).",
		LineageKey: "trade", LineageTier: 5,
		WorkerDomain: "trade", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "gold",
	})
	// tier 6 — industrial_age  rate=3.20
	b = append(b, BuildingDef{
		Name: "Stock Exchange", Key: "stock_exchange", Category: "production",
		BaseCost:    map[string]float64{"steel": 28e6, "coal": 10e6, "gold": 18e6},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 3.20}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Industrial-era stock exchange. +3.20 gold/tick (6 workers).",
		LineageKey: "trade", LineageTier: 6,
		WorkerDomain: "trade", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "gold",
	})
	// tier 7 — victorian_age  rate=6.40
	b = append(b, BuildingDef{
		Name: "Bank", Key: "bank", Category: "production",
		BaseCost:    map[string]float64{"steel": 190e6, "gold": 100e6, "iron": 80e6},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 6.40}},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "A Victorian national bank. +6.40 gold/tick (6 workers).",
		LineageKey: "trade", LineageTier: 7,
		WorkerDomain: "trade", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "gold",
	})
	// tier 8 — electric_age  rate=12.80
	b = append(b, BuildingDef{
		Name: "Financial District", Key: "financial_district", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.1e9, "electricity": 450e6, "gold": 750e6},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 12.80}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "Electric-age financial district. +12.80 gold/tick (7 workers).",
		LineageKey: "trade", LineageTier: 8,
		WorkerDomain: "trade", WorkerCapacity: 7,
		EpochKey: "electric_era", OutputResource: "gold",
	})
	// tier 9 — atomic_age  rate=25.60
	b = append(b, BuildingDef{
		Name: "Corporate HQ", Key: "corporate_hq", Category: "production",
		BaseCost:    map[string]float64{"steel": 5.5e9, "electricity": 2.2e9, "gold": 3.5e9},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 25.60}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Multinational corporate headquarters. +25.60 gold/tick (7 workers).",
		LineageKey: "trade", LineageTier: 9,
		WorkerDomain: "trade", WorkerCapacity: 7,
		EpochKey: "electric_era", OutputResource: "gold",
	})
	// tier 10 — modern_age  rate=51.20
	b = append(b, BuildingDef{
		Name: "Investment Firm", Key: "investment_firm", Category: "production",
		BaseCost:    map[string]float64{"steel": 32e9, "electricity": 12e9, "data": 1.2e9},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 51.20}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Global investment and wealth management. +51.20 gold/tick (8 workers).",
		LineageKey: "trade", LineageTier: 10,
		WorkerDomain: "trade", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "gold",
	})
	// tier 11 — information_age  rate=102.40
	b = append(b, BuildingDef{
		Name: "Venture Hub", Key: "venture_hub", Category: "production",
		BaseCost:    map[string]float64{"electricity": 85e9, "data": 9e9, "gold": 160e9},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 102.40}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "Digital venture capital hub. +102.40 gold/tick (8 workers).",
		LineageKey: "trade", LineageTier: 11,
		WorkerDomain: "trade", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "gold",
	})
	// tier 12 — digital_age  rate=204.80
	b = append(b, BuildingDef{
		Name: "Crypto Exchange", Key: "crypto_exchange", Category: "production",
		BaseCost:    map[string]float64{"electricity": 420e9, "data": 52e9, "crypto": 100e9},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 204.80}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Decentralised digital currency exchange. +204.80 gold/tick (10 workers).",
		LineageKey: "trade", LineageTier: 12,
		WorkerDomain: "trade", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "gold",
	})
	// tier 13 — cyberpunk_age  rate=409.60
	b = append(b, BuildingDef{
		Name: "Black Market Hub", Key: "black_market", Category: "production",
		BaseCost:    map[string]float64{"data": 200e9, "crypto": 1.05e12, "electricity": 2.1e12},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 409.60}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Underground black market network. +409.60 gold/tick (10 workers).",
		LineageKey: "trade", LineageTier: 13,
		WorkerDomain: "trade", WorkerCapacity: 10,
		EpochKey: "neon_era", OutputResource: "gold",
	})
	// tier 14 — fusion_age  rate=819.20
	b = append(b, BuildingDef{
		Name: "Energy Exchange", Key: "energy_exchange", Category: "production",
		BaseCost:    map[string]float64{"plasma": 4.2e12, "electricity": 12e12, "steel": 16e12},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 819.20}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Interplanetary energy trading exchange. +819.20 gold/tick (12 workers).",
		LineageKey: "trade", LineageTier: 14,
		WorkerDomain: "trade", WorkerCapacity: 12,
		EpochKey: "neon_era", OutputResource: "gold",
	})
	// tier 15 — space_age  rate=1638.40
	b = append(b, BuildingDef{
		Name: "Asteroid Market", Key: "asteroid_market", Category: "production",
		BaseCost:    map[string]float64{"titanium": 75e12, "plasma": 35e12, "electricity": 88e12},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 1638.40}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Mineral trading hub in the asteroid belt. +1638.40 gold/tick (12 workers).",
		LineageKey: "trade", LineageTier: 15,
		WorkerDomain: "trade", WorkerCapacity: 12,
		EpochKey: "neon_era", OutputResource: "gold",
	})
	// tier 16 — interstellar_age  rate=3276.80
	b = append(b, BuildingDef{
		Name: "Galactic Trade Hub", Key: "galactic_trade_hub", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 88e12, "titanium": 700e12, "plasma": 430e12},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 3276.80}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Interstellar trade network hub. +3276.80 gold/tick (15 workers).",
		LineageKey: "trade", LineageTier: 16,
		WorkerDomain: "trade", WorkerCapacity: 15,
		EpochKey: "cosmic_era", OutputResource: "gold",
	})
	// tier 17 — galactic_age  rate=6553.60
	b = append(b, BuildingDef{
		Name: "Stellar Exchange", Key: "stellar_exchange", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 175e12, "dark_matter": 880e12, "titanium": 4.4e15},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 6553.60}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Galaxy-spanning stellar exchange. +6553.60 gold/tick (15 workers).",
		LineageKey: "trade", LineageTier: 17,
		WorkerDomain: "trade", WorkerCapacity: 15,
		EpochKey: "cosmic_era", OutputResource: "gold",
	})
	// tier 18 — quantum_age  rate=13107.20
	b = append(b, BuildingDef{
		Name: "Probability Market", Key: "probability_market", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 195e12, "antimatter": 59e15, "dark_matter": 48e15},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 13107.20}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Trades across all probable timelines. +13107.20 gold/tick (18 workers).",
		LineageKey: "trade", LineageTier: 18,
		WorkerDomain: "trade", WorkerCapacity: 18,
		EpochKey: "cosmic_era", OutputResource: "gold",
	})

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
		BuildTicks:  500,
		RequiredAge: "bronze_age",
		Description: "A forge smelting iron tools. +0.10 iron/tick (4 workers).",
		LineageKey: "engineering", LineageTier: 0,
		WorkerDomain: "engineering", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "iron",
	})
	// tier 1 — iron_age  output=iron  rate=0.20
	b = append(b, BuildingDef{
		Name: "Ironworks", Key: "ironworks", Category: "production",
		BaseCost:    map[string]float64{"stone": 6000, "iron": 3000, "coal": 1500},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.20}},
		BuildTicks:  1000,
		RequiredAge: "iron_age",
		Description: "Organised iron working and tools. +0.20 iron/tick (5 workers).",
		LineageKey: "engineering", LineageTier: 1,
		WorkerDomain: "engineering", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 2 — classical_age  output=iron  rate=0.40
	b = append(b, BuildingDef{
		Name: "Aqueduct", Key: "aqueduct", Category: "production",
		BaseCost:    map[string]float64{"stone": 35000, "gold": 10000, "iron": 8000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.40}},
		BuildTicks:  3000,
		RequiredAge: "classical_age",
		Description: "Roman engineering feats improve iron output. +0.40 iron/tick (5 workers).",
		LineageKey: "engineering", LineageTier: 2,
		WorkerDomain: "engineering", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 3 — medieval_age  output=iron  rate=0.80
	b = append(b, BuildingDef{
		Name: "Workshop", Key: "workshop", Category: "production",
		BaseCost:    map[string]float64{"stone": 190000, "gold": 60000, "iron": 30000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.80}},
		BuildTicks:  6000,
		RequiredAge: "medieval_age",
		Description: "A skilled craftsman's workshop. +0.80 iron/tick (6 workers).",
		LineageKey: "engineering", LineageTier: 3,
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
		LineageKey: "engineering", LineageTier: 4,
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
		LineageKey: "engineering", LineageTier: 5,
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
		LineageKey: "engineering", LineageTier: 6,
		WorkerDomain: "engineering", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 7 — victorian_age  output=steel  rate=4.0  (+ small electricity bonus)
	b = append(b, BuildingDef{
		Name: "Steam Works", Key: "steam_works", Category: "production",
		BaseCost:    map[string]float64{"steel": 195e6, "coal": 100e6, "gold": 120e6},
		CostScale:   1.35,
		Effects: []Effect{
			{Type: "production", Target: "steel", Value: 4.0},
			{Type: "production", Target: "electricity", Value: 5.0},
		},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "Steam-powered steelworks with electricity generation. +4.0 steel, +5.0 electricity/tick (9 workers).",
		LineageKey: "engineering", LineageTier: 7,
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
		LineageKey: "engineering", LineageTier: 8,
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
		LineageKey: "engineering", LineageTier: 9,
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
		LineageKey: "engineering", LineageTier: 10,
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
		LineageKey: "engineering", LineageTier: 11,
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
		LineageKey: "engineering", LineageTier: 12,
		WorkerDomain: "engineering", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 13 — cyberpunk_age  output=plasma+electricity  rate: plasma=5, electricity=500
	b = append(b, BuildingDef{
		Name: "Augmentation Foundry", Key: "augmentation_foundry", Category: "production",
		BaseCost:    map[string]float64{"data": 230e9, "crypto": 1.2e12, "electricity": 2.4e12},
		CostScale:   1.35,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 5},
			{Type: "production", Target: "electricity", Value: 500},
		},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Cyberpunk foundry producing plasma and power. +5 plasma, +500 electricity/tick (15 workers).",
		LineageKey: "engineering", LineageTier: 13,
		WorkerDomain: "engineering", WorkerCapacity: 15,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 14 — fusion_age  output=plasma  rate=10
	b = append(b, BuildingDef{
		Name: "Fusion Reactor", Key: "fusion_reactor", Category: "production",
		BaseCost:    map[string]float64{"plasma": 5e12, "electricity": 16e12, "steel": 22e12},
		CostScale:   1.35,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 10},
			{Type: "production", Target: "electricity", Value: 1000},
		},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Fusion reactor generating plasma and electricity. +10 plasma, +1000 electricity/tick.",
		LineageKey: "engineering", LineageTier: 14,
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
		LineageKey: "engineering", LineageTier: 15,
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
		LineageKey: "engineering", LineageTier: 16,
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
		LineageKey: "engineering", LineageTier: 17,
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
		LineageKey: "engineering", LineageTier: 18,
		WorkerDomain: "engineering", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	return b
}
