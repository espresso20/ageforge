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

	// tier 0 — primitive_age  rate=0.2
	b = append(b, BuildingDef{
		Name: "Story Circle", Key: "story_circle", Category: "research",
		BaseCost:    map[string]float64{"wood": 60},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 0.2}},
		BuildTicks:  75,
		RequiredAge: "primitive_age",
		Description: "Elders share stories around the fire.",
		LineageKey:  "knowledge", LineageTier: 0,
		Scholars: 2,
		EpochKey: "stone_era",
	})
	// tier 1 — stone_age  rate=0.6
	b = append(b, BuildingDef{
		Name: "Elders' Hall", Key: "elders_hall", Category: "research",
		BaseCost:    map[string]float64{"wood": 360, "stone": 240},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 0.6}},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "A hall where tribal elders convene.",
		LineageKey:  "knowledge", LineageTier: 1,
		Scholars: 2,
		EpochKey: "stone_era",
	})
	// tier 2 — bronze_age  rate=2.0
	b = append(b, BuildingDef{
		Name: "Scriptorium", Key: "scriptorium", Category: "research",
		BaseCost:    map[string]float64{"wood": 2100, "stone": 1200, "gold": 600},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 2}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "Scribes copy and preserve texts.",
		LineageKey:  "knowledge", LineageTier: 2,
		Scholars: 3,
		EpochKey: "stone_era",
	})
	// tier 3 — iron_age  rate=1.6
	b = append(b, BuildingDef{
		Name: "Agora", Key: "agora", Category: "research",
		BaseCost:    map[string]float64{"stone": 7500, "marble": 7500, "gold": 7500, "iron": 2250, "iron_ore": 2250},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 1.6}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "Open marketplace of ideas.",
		LineageKey:  "knowledge", LineageTier: 3,
		Scholars: 3,
		EpochKey: "iron_era",
	})
	// tier 4 — classical_age  rate=3.2
	b = append(b, BuildingDef{
		Name: "Library", Key: "library", Category: "research",
		BaseCost:    map[string]float64{"stone": 55000, "marble": 55000, "gold": 36000, "iron": 12000, "iron_ore": 12000},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 3.2}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Repository of written knowledge.",
		LineageKey:  "knowledge", LineageTier: 4,
		Scholars: 4,
		EpochKey: "iron_era",
	})
	// tier 5 — medieval_age  rate=1.6
	b = append(b, BuildingDef{
		Name: "Monastery Library", Key: "monastery_library", Category: "research",
		BaseCost:    map[string]float64{"stone": 810000, "gold": 270000},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 42.4}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "Monks preserve and copy scholarly works.",
		LineageKey:  "knowledge", LineageTier: 5,
		Scholars: 4,
		EpochKey: "iron_era",
	})
	// tier 6 — renaissance_age  rate=3.2
	b = append(b, BuildingDef{
		Name: "University", Key: "university", Category: "research",
		BaseCost:    map[string]float64{"gold": 2700000, "steel": 900000},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 76.3}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "Higher learning for the intellectual elite.",
		LineageKey:  "knowledge", LineageTier: 6,
		Scholars: 5,
		EpochKey: "steel_era",
	})
	// tier 7 — colonial_age  rate=6.4
	b = append(b, BuildingDef{
		Name: "Natural Philosophy Hall", Key: "natural_philosophy_hall", Category: "research",
		BaseCost:    map[string]float64{"gold": 16400000, "steel": 6700000},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 452}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "Scientific inquiry into the natural world.",
		LineageKey:  "knowledge", LineageTier: 7,
		Scholars: 5,
		EpochKey: "steel_era",
	})
	// tier 8 — industrial_age  rate=12.8
	b = append(b, BuildingDef{
		Name: "Research Institute", Key: "research_institute", Category: "research",
		BaseCost:    map[string]float64{"steel": 7.5e07, "coal": 2.4e07, "gold": 4.5e07},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 12.8}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "Formal scientific research institute.",
		LineageKey:  "knowledge", LineageTier: 8,
		Scholars: 6,
		EpochKey: "steel_era",
	})
	// tier 9 — victorian_age  rate=25.6
	b = append(b, BuildingDef{
		Name: "Academy", Key: "academy", Category: "research",
		BaseCost:    map[string]float64{"steel": 5.3e08, "gold": 2.7e08, "iron": 1.8e08},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 25.6}},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "Victorian era academy of sciences.",
		LineageKey:  "knowledge", LineageTier: 9,
		Scholars: 6,
		EpochKey: "electric_era",
	})
	// tier 10 — electric_age  rate=51.2
	b = append(b, BuildingDef{
		Name: "Physics Laboratory", Key: "physics_laboratory", Category: "research",
		BaseCost:    map[string]float64{"steel": 3e09, "electricity": 1.2e09, "gold": 1.8e09},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 51.2}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "Electrified physics research laboratory.",
		LineageKey:  "knowledge", LineageTier: 10,
		Scholars: 7,
		EpochKey: "electric_era",
	})
	// tier 11 — atomic_age  rate=102.4
	b = append(b, BuildingDef{
		Name: "Research Campus", Key: "research_campus", Category: "research",
		BaseCost:    map[string]float64{"steel": 1.5e10, "electricity": 6e09, "uranium": 9e08},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 102.4}},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "Multi-disciplinary atomic-age research campus.",
		LineageKey:  "knowledge", LineageTier: 11,
		Scholars: 7,
		EpochKey: "electric_era",
	})
	// tier 12 — modern_age  rate=204.8
	b = append(b, BuildingDef{
		Name: "Think Tank", Key: "think_tank", Category: "research",
		BaseCost:    map[string]float64{"steel": 4.5e10, "titanium_ore": 4.5e10, "electricity": 3e10, "data": 3e09},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 204.8}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "Elite group of problem-solvers.",
		LineageKey:  "knowledge", LineageTier: 12,
		Scholars: 8,
		EpochKey: "digital_era",
	})
	// tier 13 — information_age  rate=409.6
	b = append(b, BuildingDef{
		Name: "Innovation Hub", Key: "innovation_hub", Category: "research",
		BaseCost:    map[string]float64{"electricity": 2.4e11, "data": 2.4e10, "gold": 4.5e11},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 409.6}},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "Startup-style innovation accelerator.",
		LineageKey:  "knowledge", LineageTier: 13,
		Scholars: 8,
		EpochKey: "digital_era",
	})
	// tier 14 — digital_age  rate=819.2
	b = append(b, BuildingDef{
		Name: "AI Research Lab", Key: "ai_research_lab", Category: "research",
		BaseCost:    map[string]float64{"electricity": 1.2e12, "data": 1.5e11, "steel": 9e11, "titanium_ore": 9e11},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 819.2}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Artificial intelligence drives research.",
		LineageKey:  "knowledge", LineageTier: 14,
		Scholars: 10,
		EpochKey: "digital_era",
	})
	// tier 15 — cyberpunk_age  rate=1638.4
	b = append(b, BuildingDef{
		Name: "Neuro Research Center", Key: "neuro_research_center", Category: "research",
		BaseCost:    map[string]float64{"data": 6e11, "crypto": 3e12, "electricity": 3e12, "dark_matter_crystals": 3e12},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 1638.4}},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "Neural-interface enhanced research.",
		LineageKey:  "knowledge", LineageTier: 15,
		Scholars: 10,
		EpochKey: "neon_era",
	})
	// tier 16 — fusion_age  rate=3276.8
	b = append(b, BuildingDef{
		Name: "Theoretical Institute", Key: "theoretical_institute", Category: "research",
		BaseCost:    map[string]float64{"plasma": 1.5e13, "electricity": 2.25e13, "dark_matter_crystals": 2.25e13, "steel": 6e13},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 3276.8}},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "Theoretical physics at fusion-era scale.",
		LineageKey:  "knowledge", LineageTier: 16,
		Scholars: 12,
		EpochKey: "neon_era",
	})
	// tier 17 — space_age  rate=6553.6
	b = append(b, BuildingDef{
		Name: "Deep Space Observatory", Key: "deep_space_observatory", Category: "research",
		BaseCost:    map[string]float64{"titanium": 2.4e14, "plasma": 1.2e14, "electricity": 1.5e14, "dark_matter_crystals": 1.5e14},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 6553.6}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "Observes the far reaches of the cosmos.",
		LineageKey:  "knowledge", LineageTier: 17,
		Scholars: 12,
		EpochKey: "neon_era",
	})
	// tier 18 — interstellar_age  rate=13107.2
	b = append(b, BuildingDef{
		Name: "Xenology Institute", Key: "xenology_institute", Category: "research",
		BaseCost:    map[string]float64{"dark_matter": 3e14, "titanium": 2.4e15, "plasma": 1.5e15},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 13107.2}},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "Studies alien civilizations and xeno-science.",
		LineageKey:  "knowledge", LineageTier: 18,
		Scholars: 15,
		EpochKey: "cosmic_era",
	})
	// tier 19 — galactic_age  rate=26214.4
	b = append(b, BuildingDef{
		Name: "Cosmic Research Station", Key: "cosmic_research_station", Category: "research",
		BaseCost:    map[string]float64{"antimatter": 6e14, "dark_matter": 3e15, "titanium": 1.5e16},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 26214.4}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Galactic-scale scientific research station.",
		LineageKey:  "knowledge", LineageTier: 19,
		Scholars: 15,
		EpochKey: "cosmic_era",
	})
	// tier 20 — quantum_age  rate=52428.8
	b = append(b, BuildingDef{
		Name: "Reality Academy", Key: "reality_academy", Category: "research",
		BaseCost:    map[string]float64{"quantum_flux": 6e14, "antimatter": 1.8e17, "dark_matter": 1.5e17},
		CostScale:   1.5,
		Effects:     []Effect{{Type: "research", Value: 52428.8}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Learns from the fabric of reality itself.",
		LineageKey:  "knowledge", LineageTier: 20,
		Scholars: 20,
		EpochKey: "cosmic_era",
	})

	// A knowledge building cannot be built past the count its gate asks for
	// (KnowledgeCap): an age's research has a ceiling.
	for i := range b {
		b[i].MaxCount = KnowledgeCap(b[i].Key)
	}
	return b
}
