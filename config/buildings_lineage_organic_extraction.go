package config

// buildingsLineageOrganicExtraction returns all production, housing, research, and military buildings
// for the 13 lineage chains introduced in Phase 10 of the economy redesign.
// Storage buildings and wonders remain in baseBuildingsRaw().
// All fields are set inline; no separate buildingMeta() entries needed for these buildings.
func buildingsLineageOrganicExtraction() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 3 — ORGANIC EXTRACTION (lineageKey: "organic_extraction", domain: "lumber")
	// rate = 0.20 * 2^tier  CostScale: ProducerRate (1.15)  Category: "production"  (both in the lineageSpec below)
	// Output transitions: wood(0-5) → coal(6-8) → oil(9-13) → nanobots(14-16) → quantum_flux(17-20)
	// =========================================================================

	// tier 0 — primitive_age  rate=0.20  output=wood
	// note: cost matched to gathering_camp's raw inputs (wood:20 @ 1.12) so the two
	// primitive extractors normalize to the same ~16 wood @ 1.15 — a wood camp must
	// not cost ~2.9x a gathering camp. See Ety5GDPw.
	b = append(b, BuildingDef{
		Name: "Wood Camp", Key: "wood_camp",
		BaseCost:       map[string]float64{"wood": 16},
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 0.569}},
		BuildTicks:     20,
		RequiredAge:    "primitive_age",
		Description:    "A basic camp for collecting wood.",
		WorkerCapacity: 3,
	})
	// tier 1 — stone_age  rate=0.40  output=wood
	b = append(b, BuildingDef{
		Name: "Woodcutter Camp", Key: "woodcutter_camp",
		BaseCost:       map[string]float64{"wood": 360, "stone": 180},
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 5.76}},
		BuildTicks:     50,
		RequiredAge:    "stone_age",
		Description:    "Choppers fell trees with stone axes.",
		WorkerCapacity: 4,
	})
	// tier 2 — bronze_age  rate=0.80  output=wood
	b = append(b, BuildingDef{
		Name: "Lumber Mill", Key: "lumber_mill",
		BaseCost:       map[string]float64{"wood": 2100, "stone": 1200, "iron": 450},
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 7.91}},
		BuildTicks:     150,
		RequiredAge:    "bronze_age",
		Description:    "Bronze-saw lumber processing.",
		WorkerCapacity: 5,
	})
	// tier 3 — iron_age  rate=1.60  output=wood
	b = append(b, BuildingDef{
		Name: "Timber Yard", Key: "timber_yard",
		BaseCost:       map[string]float64{"stone": 6000, "marble": 6000, "iron": 3750, "iron_ore": 3750, "wood": 9000},
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 19.2}},
		BuildTicks:     300,
		RequiredAge:    "iron_age",
		Description:    "Iron-saw timber processing yard.",
		WorkerCapacity: 5,
	})
	// tier 4 — classical_age  rate=3.20  output=wood
	b = append(b, BuildingDef{
		Name: "Wood Workshop", Key: "wood_workshop",
		BaseCost:       map[string]float64{"stone": 45000, "marble": 45000, "gold": 30000, "iron": 12000, "iron_ore": 12000},
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 54.4}},
		BuildTicks:     600,
		RequiredAge:    "classical_age",
		Description:    "Skilled carpenters working at full tilt.",
		WorkerCapacity: 6,
	})
	// tier 5 — medieval_age  rate=6.40  output=wood
	b = append(b, BuildingDef{
		Name: "Sawmill", Key: "sawmill",
		BaseCost:       map[string]float64{"stone": 742000, "gold": 263000},
		Effects:        []Effect{{Type: "production", Target: "wood", Value: 6.4}},
		BuildTicks:     1200,
		RequiredAge:    "medieval_age",
		Description:    "Water-wheel-powered sawmill.",
		WorkerCapacity: 6,
	})
	// tier 6 — renaissance_age  rate=12.80  output=coal
	b = append(b, BuildingDef{
		Name: "Coal Mine", Key: "coal_mine",
		BaseCost:       map[string]float64{"gold": 2.57 * Million, "steel": 909000},
		Effects:        []Effect{{Type: "production", Target: "coal", Value: 106}},
		BuildTicks:     2400,
		RequiredAge:    "renaissance_age",
		Description:    "Early coal extraction for industry.",
		WorkerCapacity: 7,
	})
	// tier 7 — colonial_age  rate=25.60  output=coal
	b = append(b, BuildingDef{
		Name: "Coal Works", Key: "coal_works",
		BaseCost: map[string]float64{"gold": 9 * Million, "steel": 4.5 * Million},
		Effects: []Effect{
			{Type: "production", Target: "coal", Value: 25.6},
		},
		BuildTicks:     3600,
		RequiredAge:    "colonial_age",
		Description:    "Organized coal extraction and processing.",
		WorkerCapacity: 8,
	})
	// tier 8 — industrial_age  rate=51.20  output=coal
	b = append(b, BuildingDef{
		Name: "Steam Colliery", Key: "steam_coal_plant",
		BaseCost:       map[string]float64{"steel": 66 * Million, "coal": 24 * Million, "gold": 36 * Million},
		Effects:        []Effect{{Type: "production", Target: "coal", Value: 8130}},
		BuildTicks:     3600,
		RequiredAge:    "industrial_age",
		Description:    "Steam-powered coal extraction plant.",
		WorkerCapacity: 10,
	})
	// tier 9 — victorian_age  rate=102.40  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Derrick", Key: "oil_derrick",
		BaseCost:       map[string]float64{"steel": 510 * Million, "iron": 240 * Million, "gold": 300 * Million},
		Effects:        []Effect{{Type: "production", Target: "oil", Value: 46000}},
		BuildTicks:     3600,
		RequiredAge:    "victorian_age",
		Description:    "Early oil extraction derrick.",
		WorkerCapacity: 10,
	})
	// tier 10 — electric_age  rate=204.80  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Field", Key: "oil_field",
		BaseCost:       map[string]float64{"steel": 2.7 * Billion, "electricity": 1.1 * Billion, "oil": 300 * Million},
		Effects:        []Effect{{Type: "production", Target: "oil", Value: 69100}},
		BuildTicks:     3600,
		RequiredAge:    "electric_age",
		Description:    "Electrified oil field operations.",
		WorkerCapacity: 12,
	})
	// tier 11 — atomic_age  rate=409.60  output=oil
	b = append(b, BuildingDef{
		Name: "Petroleum Refinery", Key: "petroleum_refinery",
		BaseCost:       map[string]float64{"steel": 15 * Billion, "electricity": 6 * Billion, "uranium": 900 * Million},
		Effects:        []Effect{{Type: "production", Target: "oil", Value: 409.6}},
		BuildTicks:     3600,
		RequiredAge:    "atomic_age",
		Description:    "Advanced petroleum refinery.",
		WorkerCapacity: 12,
	})
	// tier 12 — modern_age  rate=819.20  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Platform", Key: "oil_platform",
		BaseCost: map[string]float64{"steel": 42 * Billion, "titanium_ore": 42 * Billion, "electricity": 30 * Billion, "data": 2.4 * Billion},
		Effects: []Effect{
			{Type: "production", Target: "oil", Value: 649000},
		},
		BuildTicks:     3600,
		RequiredAge:    "modern_age",
		Description:    "Offshore AI-monitored oil platform.",
		WorkerCapacity: 14,
	})
	// tier 13 — information_age  rate=1638.40  output=oil
	b = append(b, BuildingDef{
		Name: "Smart Refinery", Key: "smart_refinery",
		BaseCost:       map[string]float64{"electricity": 230 * Billion, "data": 21 * Billion, "steel": 210 * Billion, "titanium_ore": 210 * Billion},
		Effects:        []Effect{{Type: "production", Target: "oil", Value: 1638.4}},
		BuildTicks:     3600,
		RequiredAge:    "information_age",
		Description:    "AI-optimized smart petroleum refinery.",
		WorkerCapacity: 15,
	})
	// tier 14 — digital_age  rate=3276.80  output=nanobots
	b = append(b, BuildingDef{
		Name: "Bio Fabrication Lab", Key: "bio_fabrication_lab",
		BaseCost:       map[string]float64{"electricity": 1.1 * Trillion, "data": 140 * Billion, "steel": 850 * Billion, "titanium_ore": 850 * Billion},
		Effects:        []Effect{{Type: "production", Target: "nanobots", Value: 3276.8}},
		BuildTicks:     3600,
		RequiredAge:    "digital_age",
		Description:    "Digital-biological nanofabrication.",
		WorkerCapacity: 16,
	})
	// tier 15 — cyberpunk_age  rate=6553.60  output=nanobots
	b = append(b, BuildingDef{
		Name: "Nanobot Vat", Key: "nanobot_vat",
		BaseCost:       map[string]float64{"data": 540 * Billion, "crypto": 3.6 * Trillion, "electricity": 3.75 * Trillion, "dark_matter_crystals": 3.75 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "nanobots", Value: 6553.6}},
		BuildTicks:     3600,
		RequiredAge:    "cyberpunk_age",
		Description:    "Vat-grown nanobot manufacturing.",
		WorkerCapacity: 18,
	})
	// tier 16 — fusion_age  rate=13107.20  output=nanobots
	b = append(b, BuildingDef{
		Name: "Molecular Synthesizer", Key: "molecular_synthesizer",
		BaseCost:       map[string]float64{"plasma": 12 * Trillion, "electricity": 21 * Trillion, "dark_matter_crystals": 21 * Trillion, "steel": 54 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "nanobots", Value: 13107.2}},
		BuildTicks:     3600,
		RequiredAge:    "fusion_age",
		Description:    "Plasma-powered molecular synthesis.",
		WorkerCapacity: 20,
	})
	// tier 17 — space_age  rate=26214.40  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Quantum Organic Extractor", Key: "quantum_organic_extractor",
		BaseCost:       map[string]float64{"titanium": 230 * Trillion, "plasma": 110 * Trillion, "electricity": 135 * Trillion, "dark_matter_crystals": 135 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "quantum_flux", Value: 26214.4}},
		BuildTicks:     3600,
		RequiredAge:    "space_age",
		Description:    "Quantum-state organic matter extraction.",
		WorkerCapacity: 20,
	})
	// tier 18 — interstellar_age  rate=52428.80  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Reality Matter Weaver", Key: "reality_matter_weaver",
		BaseCost:       map[string]float64{"dark_matter": 270 * Trillion, "titanium": 2.3 * Quadrillion, "plasma": 1.4 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "quantum_flux", Value: 52428.8}},
		BuildTicks:     3600,
		RequiredAge:    "interstellar_age",
		Description:    "Weaves reality matter into quantum flux.",
		WorkerCapacity: 25,
	})
	// tier 19 — galactic_age  rate=104857.60  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Cosmic Organic Works", Key: "cosmic_organic_works",
		BaseCost:       map[string]float64{"antimatter": 540 * Trillion, "dark_matter": 2.7 * Quadrillion, "titanium": 14 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "quantum_flux", Value: 104857.6}},
		BuildTicks:     3600,
		RequiredAge:    "galactic_age",
		Description:    "Galactic-scale cosmic organic works.",
		WorkerCapacity: 25,
	})
	// tier 20 — quantum_age  rate=209715.20  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Reality Harvester", Key: "reality_harvester",
		BaseCost: map[string]float64{"antimatter": 674 * Quadrillion},
		Effects: []Effect{
			{Type: "production", Target: "quantum_flux", Value: 44.9 * Billion},
		},
		BuildTicks:     3600,
		RequiredAge:    "quantum_age",
		Description:    "Harvests raw quantum flux from reality itself.",
		WorkerCapacity: 30,
	})

	// Fill the lineage rows now: the nano foundry below is appended after the call,
	// so it keeps its explicit fields and is never given a lineage.
	b = finishLineage(b, lineageSpec{
		Key: "organic_extraction", Category: "production", Scale: ProducerRate, Domain: "lumber", Lines: true,
	})

	// =========================================================================
	// NANOBOT PRODUCER (standalone — not part of the lineage tier chain).
	// Nanobots unlock as a resource in the Modern Age, but the lineage's first
	// nanobot output (Bio Fabrication Lab) doesn't arrive until the Digital Age,
	// two ages later. This dedicated foundry closes that gap so nanobots have a
	// real production path the moment they're unlocked. No LineageKey/OutputResource
	// so it is never remapped on age transition — it always makes nanobots.
	// Costed to the Modern tier (cf. power_grid_hub); engineering domain workers.
	// =========================================================================
	b = append(b, BuildingDef{
		Name: "Nano Foundry", Key: "nano_foundry", Category: "production",
		BaseCost:     map[string]float64{"steel": 85 * Billion, "titanium_ore": 85 * Billion, "electricity": 68 * Billion, "data": 8.5 * Billion},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "nanobots", Value: 80}},
		BuildTicks:   3600,
		RequiredAge:  "modern_age",
		Description:  "Molecular assembly line printing self-organizing nanobots.",
		WorkerDomain: "engineering", WorkerCapacity: 12,
	})

	return b
}
