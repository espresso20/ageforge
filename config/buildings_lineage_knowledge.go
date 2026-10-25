package config

// newProductionBuildings2 returns lineages 5-9:
// knowledge, faith, military, trade, engineering.
// Merged into newProductionBuildings() via init — see buildings_new_merge.go.
func buildingsLineageKnowledge() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 5 — KNOWLEDGE (lineageKey: "knowledge", domain: "knowledge", output: "knowledge")
	// rate = 0.05 * 2^tier  CostScale: KnowledgeRate (1.5)  Category: "research"  (both in the lineageSpec at the foot)
	// note: base raised 0.002 -> 0.05 (25x) to match the ~10x food/wood rate boost
	// plus catch-up — knowledge was ~250x slower than gathering_camp. See XtWUrrKY.
	// =========================================================================

	// tier 0 — primitive_age  rate=0.2
	b = append(b, BuildingDef{
		Name: "Story Circle", Key: "story_circle",
		BaseCost:    map[string]float64{"wood": 60},
		Effects:     []Effect{{Type: "research", Value: 0.2}},
		BuildTicks:  75,
		RequiredAge: "primitive_age",
		Description: "Elders share stories around the fire.",
		Scholars:    2,
	})
	// tier 1 — stone_age  rate=0.6
	b = append(b, BuildingDef{
		Name: "Elders' Hall", Key: "elders_hall",
		BaseCost:    map[string]float64{"wood": 360, "stone": 240},
		Effects:     []Effect{{Type: "research", Value: 0.6}},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "A hall where tribal elders convene.",
		Scholars:    2,
	})
	// tier 2 — bronze_age  rate=2.0
	b = append(b, BuildingDef{
		Name: "Scriptorium", Key: "scriptorium",
		BaseCost:    map[string]float64{"wood": 2100, "stone": 1200, "gold": 600},
		Effects:     []Effect{{Type: "research", Value: 2}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "Scribes copy and preserve texts.",
		Scholars:    3,
	})
	// tier 3 — iron_age  rate=1.6
	b = append(b, BuildingDef{
		Name: "Agora", Key: "agora",
		BaseCost:    map[string]float64{"stone": 7500, "marble": 7500, "gold": 7500, "iron": 2250, "iron_ore": 2250},
		Effects:     []Effect{{Type: "research", Value: 1.6}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "Open marketplace of ideas.",
		Scholars:    3,
	})
	// tier 4 — classical_age  rate=3.2
	b = append(b, BuildingDef{
		Name: "Library", Key: "library",
		BaseCost:    map[string]float64{"stone": 55000, "marble": 55000, "gold": 36000, "iron": 12000, "iron_ore": 12000},
		Effects:     []Effect{{Type: "research", Value: 3.2}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Repository of written knowledge.",
		Scholars:    4,
	})
	// tier 5 — medieval_age  rate=1.6
	b = append(b, BuildingDef{
		Name: "Monastery Library", Key: "monastery_library",
		BaseCost:    map[string]float64{"stone": 810000, "gold": 270000},
		Effects:     []Effect{{Type: "research", Value: 42.4}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "Monks preserve and copy scholarly works.",
		Scholars:    4,
	})
	// tier 6 — renaissance_age  rate=3.2
	b = append(b, BuildingDef{
		Name: "University", Key: "university",
		BaseCost:    map[string]float64{"gold": 2.7 * Million, "steel": 900000},
		Effects:     []Effect{{Type: "research", Value: 76.3}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "Higher learning for the intellectual elite.",
		Scholars:    5,
	})
	// tier 7 — colonial_age  rate=6.4
	b = append(b, BuildingDef{
		Name: "Natural Philosophy Hall", Key: "natural_philosophy_hall",
		BaseCost:    map[string]float64{"gold": 16.4 * Million, "steel": 6.7 * Million},
		Effects:     []Effect{{Type: "research", Value: 452}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "Scientific inquiry into the natural world.",
		Scholars:    5,
	})
	// tier 8 — industrial_age  rate=12.8
	b = append(b, BuildingDef{
		Name: "Research Institute", Key: "research_institute",
		BaseCost:    map[string]float64{"steel": 75 * Million, "coal": 24 * Million, "gold": 45 * Million},
		Effects:     []Effect{{Type: "research", Value: 12.8}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "Formal scientific research institute.",
		Scholars:    6,
	})
	// tier 9 — victorian_age  rate=25.6
	b = append(b, BuildingDef{
		Name: "Academy", Key: "academy",
		BaseCost:    map[string]float64{"steel": 530 * Million, "gold": 270 * Million, "iron": 180 * Million},
		Effects:     []Effect{{Type: "research", Value: 25.6}},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "Victorian era academy of sciences.",
		Scholars:    6,
	})
	// tier 10 — electric_age  rate=51.2
	b = append(b, BuildingDef{
		Name: "Physics Laboratory", Key: "physics_laboratory",
		BaseCost:    map[string]float64{"steel": 3 * Billion, "electricity": 1.2 * Billion, "gold": 1.8 * Billion},
		Effects:     []Effect{{Type: "research", Value: 51.2}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "Electrified physics research laboratory.",
		Scholars:    7,
	})
	// tier 11 — atomic_age  rate=102.4
	b = append(b, BuildingDef{
		Name: "Research Campus", Key: "research_campus",
		BaseCost:    map[string]float64{"steel": 15 * Billion, "electricity": 6 * Billion, "uranium": 900 * Million},
		Effects:     []Effect{{Type: "research", Value: 102.4}},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "Multi-disciplinary atomic-age research campus.",
		Scholars:    7,
	})
	// tier 12 — modern_age  rate=204.8
	b = append(b, BuildingDef{
		Name: "Think Tank", Key: "think_tank",
		BaseCost:    map[string]float64{"steel": 45 * Billion, "titanium_ore": 45 * Billion, "electricity": 30 * Billion, "data": 3 * Billion},
		Effects:     []Effect{{Type: "research", Value: 204.8}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "Elite group of problem-solvers.",
		Scholars:    8,
	})
	// tier 13 — information_age  rate=409.6
	b = append(b, BuildingDef{
		Name: "Innovation Hub", Key: "innovation_hub",
		BaseCost:    map[string]float64{"electricity": 240 * Billion, "data": 24 * Billion, "gold": 450 * Billion},
		Effects:     []Effect{{Type: "research", Value: 409.6}},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "Startup-style innovation accelerator.",
		Scholars:    8,
	})
	// tier 14 — digital_age  rate=819.2
	b = append(b, BuildingDef{
		Name: "AI Research Lab", Key: "ai_research_lab",
		BaseCost:    map[string]float64{"electricity": 1.2 * Trillion, "data": 150 * Billion, "steel": 900 * Billion, "titanium_ore": 900 * Billion},
		Effects:     []Effect{{Type: "research", Value: 819.2}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Artificial intelligence drives research.",
		Scholars:    10,
	})
	// tier 15 — cyberpunk_age  rate=1638.4
	b = append(b, BuildingDef{
		Name: "Neuro Research Center", Key: "neuro_research_center",
		BaseCost:    map[string]float64{"data": 600 * Billion, "crypto": 3 * Trillion, "electricity": 3 * Trillion, "dark_matter_crystals": 3 * Trillion},
		Effects:     []Effect{{Type: "research", Value: 1638.4}},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "Neural-interface enhanced research.",
		Scholars:    10,
	})
	// tier 16 — fusion_age  rate=3276.8
	b = append(b, BuildingDef{
		Name: "Theoretical Institute", Key: "theoretical_institute",
		BaseCost:    map[string]float64{"plasma": 15 * Trillion, "electricity": 22.5 * Trillion, "dark_matter_crystals": 22.5 * Trillion, "steel": 60 * Trillion},
		Effects:     []Effect{{Type: "research", Value: 3276.8}},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "Theoretical physics at fusion-era scale.",
		Scholars:    12,
	})
	// tier 17 — space_age  rate=6553.6
	b = append(b, BuildingDef{
		Name: "Deep Space Observatory", Key: "deep_space_observatory",
		BaseCost:    map[string]float64{"titanium": 240 * Trillion, "plasma": 120 * Trillion, "electricity": 150 * Trillion, "dark_matter_crystals": 150 * Trillion},
		Effects:     []Effect{{Type: "research", Value: 6553.6}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "Observes the far reaches of the cosmos.",
		Scholars:    12,
	})
	// tier 18 — interstellar_age  rate=13107.2
	b = append(b, BuildingDef{
		Name: "Xenology Institute", Key: "xenology_institute",
		BaseCost:    map[string]float64{"dark_matter": 300 * Trillion, "titanium": 2.4 * Quadrillion, "plasma": 1.5 * Quadrillion},
		Effects:     []Effect{{Type: "research", Value: 13107.2}},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "Studies alien civilizations and xeno-science.",
		Scholars:    15,
	})
	// tier 19 — galactic_age  rate=26214.4
	b = append(b, BuildingDef{
		Name: "Cosmic Research Station", Key: "cosmic_research_station",
		BaseCost:    map[string]float64{"antimatter": 600 * Trillion, "dark_matter": 3 * Quadrillion, "titanium": 15 * Quadrillion},
		Effects:     []Effect{{Type: "research", Value: 26214.4}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Galactic-scale scientific research station.",
		Scholars:    15,
	})
	// tier 20 — quantum_age  rate=52428.8
	b = append(b, BuildingDef{
		Name: "Reality Academy", Key: "reality_academy",
		BaseCost:    map[string]float64{"quantum_flux": 600 * Trillion, "antimatter": 180 * Quadrillion, "dark_matter": 150 * Quadrillion},
		Effects:     []Effect{{Type: "research", Value: 52428.8}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Learns from the fabric of reality itself.",
		Scholars:    20,
	})

	// A knowledge building researches at full rate up to the count its gate
	// asks for (KnowledgeCap); each copy past it researches at half rate. An
	// age's research has a soft ceiling, and nothing refuses a copy.
	for i := range b {
		b[i].FullRateCopies = KnowledgeCap(b[i].Key)
	}
	return finishLineage(b, lineageSpec{
		Key: "knowledge", Category: "research", Scale: KnowledgeRate,
	})
}
