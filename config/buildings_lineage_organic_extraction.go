package config

// buildingsLineageOrganicExtraction returns all production, housing, research, and military buildings
// for the 13 lineage chains introduced in Phase 10 of the economy redesign.
// Storage buildings and wonders remain in baseBuildingsRaw().
// All fields are set inline; no separate buildingMeta() entries needed for these buildings.
func buildingsLineageOrganicExtraction() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 3 — ORGANIC EXTRACTION (lineageKey: "organic_extraction", domain: "lumber")
	// rate = 0.20 * 2^tier  CostScale: 1.12 (tier 0), 1.30 (tiers 1+)  Category: "production"
	// Output transitions: wood(0-5) → coal(6-8) → oil(9-13) → nanobots(14-16) → quantum_flux(17-20)
	// =========================================================================

	// tier 0 — primitive_age  rate=0.20  output=wood
	// note: cost matched to gathering_camp's raw inputs (wood:20 @ 1.12) so the two
	// primitive extractors normalize to the same ~16 wood @ 1.15 — a wood camp must
	// not cost ~2.9x a gathering camp. See Ety5GDPw.
	b = append(b, BuildingDef{
		Name: "Wood Camp", Key: "wood_camp", Category: "production",
		BaseCost:    map[string]float64{"wood": 16},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 0.569}},
		BuildTicks:  20,
		RequiredAge: "primitive_age",
		Description: "A basic camp for collecting wood.",
		LineageKey:  "organic_extraction", LineageTier: 0,
		WorkerDomain: "lumber", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "wood",
	})
	// tier 1 — stone_age  rate=0.40  output=wood
	b = append(b, BuildingDef{
		Name: "Woodcutter Camp", Key: "woodcutter_camp", Category: "production",
		BaseCost:    map[string]float64{"wood": 360, "stone": 180},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 5.76}},
		BuildTicks:  50,
		RequiredAge: "stone_age",
		Description: "Choppers fell trees with stone axes.",
		LineageKey:  "organic_extraction", LineageTier: 1,
		WorkerDomain: "lumber", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "wood",
	})
	// tier 2 — bronze_age  rate=0.80  output=wood
	b = append(b, BuildingDef{
		Name: "Lumber Mill", Key: "lumber_mill", Category: "production",
		BaseCost:    map[string]float64{"wood": 2100, "stone": 1200, "iron": 450},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 7.91}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "Bronze-saw lumber processing.",
		LineageKey:  "organic_extraction", LineageTier: 2,
		WorkerDomain: "lumber", WorkerCapacity: 5,
		EpochKey: "stone_era", OutputResource: "wood",
	})
	// tier 3 — iron_age  rate=1.60  output=wood
	b = append(b, BuildingDef{
		Name: "Timber Yard", Key: "timber_yard", Category: "production",
		BaseCost:    map[string]float64{"stone": 12000, "iron": 7500, "wood": 9000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 19.2}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "Iron-saw timber processing yard.",
		LineageKey:  "organic_extraction", LineageTier: 3,
		WorkerDomain: "lumber", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "wood",
	})
	// tier 4 — classical_age  rate=3.20  output=wood
	b = append(b, BuildingDef{
		Name: "Wood Workshop", Key: "wood_workshop", Category: "production",
		BaseCost:    map[string]float64{"stone": 90000, "gold": 30000, "iron": 24000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 54.4}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Skilled carpenters working at full tilt.",
		LineageKey:  "organic_extraction", LineageTier: 4,
		WorkerDomain: "lumber", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "wood",
	})
	// tier 5 — medieval_age  rate=6.40  output=wood
	b = append(b, BuildingDef{
		Name: "Sawmill", Key: "sawmill", Category: "production",
		BaseCost:    map[string]float64{"stone": 480000, "gold": 170000, "knowledge": 45000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 6.4}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "Water-wheel-powered sawmill.",
		LineageKey:  "organic_extraction", LineageTier: 5,
		WorkerDomain: "lumber", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "wood",
	})
	// tier 6 — renaissance_age  rate=12.80  output=coal
	b = append(b, BuildingDef{
		Name: "Coal Mine", Key: "coal_mine", Category: "production",
		BaseCost:    map[string]float64{"gold": 1700000, "steel": 600000, "knowledge": 240000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "coal", Value: 106}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "Early coal extraction for industry.",
		LineageKey:  "organic_extraction", LineageTier: 6,
		WorkerDomain: "lumber", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 7 — colonial_age  rate=25.60  output=coal
	b = append(b, BuildingDef{
		Name: "Coal Works", Key: "coal_works", Category: "production",
		BaseCost:    map[string]float64{"gold": 9000000, "steel": 4500000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "coal", Value: 25.6}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "Organized coal extraction and processing.",
		LineageKey:  "organic_extraction", LineageTier: 7,
		WorkerDomain: "lumber", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 8 — industrial_age  rate=51.20  output=coal
	b = append(b, BuildingDef{
		Name: "Steam Colliery", Key: "steam_coal_plant", Category: "production",
		BaseCost:    map[string]float64{"steel": 6.6e07, "coal": 2.4e07, "gold": 3.6e07},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "coal", Value: 8130}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "Steam-powered coal extraction plant.",
		LineageKey:  "organic_extraction", LineageTier: 8,
		WorkerDomain: "lumber", WorkerCapacity: 10,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 9 — victorian_age  rate=102.40  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Derrick", Key: "oil_derrick", Category: "production",
		BaseCost:    map[string]float64{"steel": 5.1e08, "iron": 2.4e08, "gold": 3e08},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "oil", Value: 46000}},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "Early oil extraction derrick.",
		LineageKey:  "organic_extraction", LineageTier: 9,
		WorkerDomain: "lumber", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "oil",
	})
	// tier 10 — electric_age  rate=204.80  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Field", Key: "oil_field", Category: "production",
		BaseCost:    map[string]float64{"steel": 2.7e09, "electricity": 1.1e09, "oil": 3e08},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "oil", Value: 69100}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "Electrified oil field operations.",
		LineageKey:  "organic_extraction", LineageTier: 10,
		WorkerDomain: "lumber", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "oil",
	})
	// tier 11 — atomic_age  rate=409.60  output=oil
	b = append(b, BuildingDef{
		Name: "Petroleum Refinery", Key: "petroleum_refinery", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.5e10, "electricity": 6e09, "uranium": 9e08},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "oil", Value: 409.6}},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "Advanced petroleum refinery.",
		LineageKey:  "organic_extraction", LineageTier: 11,
		WorkerDomain: "lumber", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "oil",
	})
	// tier 12 — modern_age  rate=819.20  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Platform", Key: "oil_platform", Category: "production",
		BaseCost:    map[string]float64{"steel": 8.4e10, "electricity": 3e10, "data": 2.4e09},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "oil", Value: 655000}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "Offshore AI-monitored oil platform.",
		LineageKey:  "organic_extraction", LineageTier: 12,
		WorkerDomain: "lumber", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "oil",
	})
	// tier 13 — information_age  rate=1638.40  output=oil
	b = append(b, BuildingDef{
		Name: "Smart Refinery", Key: "smart_refinery", Category: "production",
		BaseCost:    map[string]float64{"electricity": 2.3e11, "data": 2.1e10, "steel": 4.2e11},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "oil", Value: 1638.4}},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "AI-optimized smart petroleum refinery.",
		LineageKey:  "organic_extraction", LineageTier: 13,
		WorkerDomain: "lumber", WorkerCapacity: 15,
		EpochKey: "digital_era", OutputResource: "oil",
	})
	// tier 14 — digital_age  rate=3276.80  output=nanobots
	b = append(b, BuildingDef{
		Name: "Bio Fabrication Lab", Key: "bio_fabrication_lab", Category: "production",
		BaseCost:    map[string]float64{"electricity": 1.1e12, "data": 1.4e11, "steel": 1.7e12},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "nanobots", Value: 3276.8}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Digital-biological nanofabrication.",
		LineageKey:  "organic_extraction", LineageTier: 14,
		WorkerDomain: "lumber", WorkerCapacity: 16,
		EpochKey: "digital_era", OutputResource: "nanobots",
	})
	// tier 15 — cyberpunk_age  rate=6553.60  output=nanobots
	b = append(b, BuildingDef{
		Name: "Nanobot Vat", Key: "nanobot_vat", Category: "production",
		BaseCost:    map[string]float64{"data": 5.4e11, "crypto": 3.6e12, "electricity": 7.5e12},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "nanobots", Value: 6553.6}},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "Vat-grown nanobot manufacturing.",
		LineageKey:  "organic_extraction", LineageTier: 15,
		WorkerDomain: "lumber", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "nanobots",
	})
	// tier 16 — fusion_age  rate=13107.20  output=nanobots
	b = append(b, BuildingDef{
		Name: "Molecular Synthesizer", Key: "molecular_synthesizer", Category: "production",
		BaseCost:    map[string]float64{"plasma": 1.2e13, "electricity": 4.2e13, "steel": 5.4e13},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "nanobots", Value: 13107.2}},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "Plasma-powered molecular synthesis.",
		LineageKey:  "organic_extraction", LineageTier: 16,
		WorkerDomain: "lumber", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "nanobots",
	})
	// tier 17 — space_age  rate=26214.40  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Quantum Organic Extractor", Key: "quantum_organic_extractor", Category: "production",
		BaseCost:    map[string]float64{"titanium": 2.3e14, "plasma": 1.1e14, "electricity": 2.7e14},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 26214.4}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "Quantum-state organic matter extraction.",
		LineageKey:  "organic_extraction", LineageTier: 17,
		WorkerDomain: "lumber", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "quantum_flux",
	})
	// tier 18 — interstellar_age  rate=52428.80  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Reality Matter Weaver", Key: "reality_matter_weaver", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 2.7e14, "titanium": 2.3e15, "plasma": 1.4e15},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 52428.8}},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "Weaves reality matter into quantum flux.",
		LineageKey:  "organic_extraction", LineageTier: 18,
		WorkerDomain: "lumber", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})
	// tier 19 — galactic_age  rate=104857.60  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Cosmic Organic Works", Key: "cosmic_organic_works", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 5.4e14, "dark_matter": 2.7e15, "titanium": 1.4e16},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 104857.6}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Galactic-scale cosmic organic works.",
		LineageKey:  "organic_extraction", LineageTier: 19,
		WorkerDomain: "lumber", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})
	// tier 20 — quantum_age  rate=209715.20  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Reality Harvester", Key: "reality_harvester", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 5.4e14, "antimatter": 1.7e17, "dark_matter": 1.4e17},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 3.81e10}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Harvests raw quantum flux from reality itself.",
		LineageKey:  "organic_extraction", LineageTier: 20,
		WorkerDomain: "lumber", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
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
		BaseCost:     map[string]float64{"steel": 1.7e11, "electricity": 6.8e10, "data": 8.5e09},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "nanobots", Value: 80}},
		BuildTicks:   3600,
		RequiredAge:  "modern_age",
		Description:  "Molecular assembly line printing self-organizing nanobots.",
		WorkerDomain: "engineering", WorkerCapacity: 12,
		EpochKey: "digital_era",
	})

	return b
}
