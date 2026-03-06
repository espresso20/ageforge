package config

// buildingsLineageOrganicExtraction returns all production, housing, research, and military buildings
// for the 13 lineage chains introduced in Phase 10 of the economy redesign.
// Storage buildings and wonders remain in baseBuildingsRaw().
// All fields are set inline; no separate buildingMeta() entries needed for these buildings.
func buildingsLineageOrganicExtraction() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 3 — ORGANIC EXTRACTION (lineageKey: "organic_extraction", domain: "lumber")
	// rate = 0.04 * 2^tier  CostScale: 1.30  Category: "production"
	// Output transitions: wood(0-5) → coal(6-8) → oil(9-13) → nanobots(14-16) → quantum_flux(17-20)
	// =========================================================================

	// tier 0 — primitive_age  rate=0.20  output=wood
	b = append(b, BuildingDef{
		Name: "Wood Camp", Key: "wood_camp", Category: "production",
		BaseCost:    map[string]float64{"wood": 15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 0.20}},
		BuildTicks:  20,
		RequiredAge: "primitive_age",
		Description: "A basic camp for collecting wood. +0.20 wood/tick (3 workers).",
		LineageKey:  "organic_extraction", LineageTier: 0,
		WorkerDomain: "lumber", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "wood",
	})
	// tier 1 — stone_age  rate=0.40  output=wood
	b = append(b, BuildingDef{
		Name: "Woodcutter Camp", Key: "woodcutter_camp", Category: "production",
		BaseCost:    map[string]float64{"wood": 120, "stone": 60},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 0.40}},
		BuildTicks:  50,
		RequiredAge: "stone_age",
		Description: "Choppers fell trees with stone axes. +0.40 wood/tick (4 workers).",
		LineageKey:  "organic_extraction", LineageTier: 1,
		WorkerDomain: "lumber", WorkerCapacity: 4,
		EpochKey: "stone_era", OutputResource: "wood",
	})
	// tier 2 — bronze_age  rate=0.80  output=wood
	b = append(b, BuildingDef{
		Name: "Lumber Mill", Key: "lumber_mill", Category: "production",
		BaseCost:    map[string]float64{"wood": 700, "stone": 400, "iron": 150},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 0.80}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "Bronze-saw lumber processing. +0.80 wood/tick (5 workers).",
		LineageKey:  "organic_extraction", LineageTier: 2,
		WorkerDomain: "lumber", WorkerCapacity: 5,
		EpochKey: "stone_era", OutputResource: "wood",
	})
	// tier 3 — iron_age  rate=0.32  output=wood
	b = append(b, BuildingDef{
		Name: "Timber Yard", Key: "timber_yard", Category: "production",
		BaseCost:    map[string]float64{"stone": 4000, "iron": 2500, "wood": 3000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 0.32}},
		BuildTicks:  1000,
		RequiredAge: "iron_age",
		Description: "Iron-saw timber processing yard. +0.32 wood/tick (5 workers).",
		LineageKey:  "organic_extraction", LineageTier: 3,
		WorkerDomain: "lumber", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "wood",
	})
	// tier 4 — classical_age  rate=0.64  output=wood
	b = append(b, BuildingDef{
		Name: "Wood Workshop", Key: "wood_workshop", Category: "production",
		BaseCost:    map[string]float64{"stone": 30000, "gold": 10000, "iron": 8000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 0.64}},
		BuildTicks:  3000,
		RequiredAge: "classical_age",
		Description: "Skilled carpenters maximise wood output. +0.64 wood/tick (6 workers).",
		LineageKey:  "organic_extraction", LineageTier: 4,
		WorkerDomain: "lumber", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "wood",
	})
	// tier 5 — medieval_age  rate=1.28  output=wood
	b = append(b, BuildingDef{
		Name: "Sawmill", Key: "sawmill", Category: "production",
		BaseCost:    map[string]float64{"stone": 160000, "gold": 55000, "knowledge": 15000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "wood", Value: 1.28}},
		BuildTicks:  6000,
		RequiredAge: "medieval_age",
		Description: "Water-wheel-powered sawmill. +1.28 wood/tick (6 workers).",
		LineageKey:  "organic_extraction", LineageTier: 5,
		WorkerDomain: "lumber", WorkerCapacity: 6,
		EpochKey: "iron_era", OutputResource: "wood",
	})
	// tier 6 — renaissance_age  rate=2.56  output=coal
	b = append(b, BuildingDef{
		Name: "Coal Mine", Key: "coal_mine", Category: "production",
		BaseCost:    map[string]float64{"gold": 550000, "steel": 200000, "knowledge": 80000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "coal", Value: 2.56}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "Early coal extraction for industry. +2.56 coal/tick (7 workers).",
		LineageKey:  "organic_extraction", LineageTier: 6,
		WorkerDomain: "lumber", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 7 — colonial_age  rate=5.12  output=coal
	b = append(b, BuildingDef{
		Name: "Coal Works", Key: "coal_works", Category: "production",
		BaseCost:    map[string]float64{"gold": 3e6, "steel": 1.5e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "coal", Value: 5.12}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "Organised coal extraction and processing. +5.12 coal/tick (8 workers).",
		LineageKey:  "organic_extraction", LineageTier: 7,
		WorkerDomain: "lumber", WorkerCapacity: 8,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 8 — industrial_age  rate=10.24  output=coal
	b = append(b, BuildingDef{
		Name: "Steam Coal Plant", Key: "steam_coal_plant", Category: "production",
		BaseCost:    map[string]float64{"steel": 22e6, "coal": 8e6, "gold": 12e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "coal", Value: 10.24}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Steam-powered coal extraction plant. +10.24 coal/tick (10 workers).",
		LineageKey:  "organic_extraction", LineageTier: 8,
		WorkerDomain: "lumber", WorkerCapacity: 10,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 9 — victorian_age  rate=20.48  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Derrick", Key: "oil_derrick", Category: "production",
		BaseCost:    map[string]float64{"steel": 170e6, "iron": 80e6, "gold": 100e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "oil", Value: 20.48}},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "Early oil extraction derrick. +20.48 oil/tick (10 workers).",
		LineageKey:  "organic_extraction", LineageTier: 9,
		WorkerDomain: "lumber", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "oil",
	})
	// tier 10 — electric_age  rate=40.96  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Field", Key: "oil_field", Category: "production",
		BaseCost:    map[string]float64{"steel": 900e6, "electricity": 350e6, "oil": 100e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "oil", Value: 40.96}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "Electrified oil field operations. +40.96 oil/tick (12 workers).",
		LineageKey:  "organic_extraction", LineageTier: 10,
		WorkerDomain: "lumber", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "oil",
	})
	// tier 11 — atomic_age  rate=81.92  output=oil
	b = append(b, BuildingDef{
		Name: "Petroleum Refinery", Key: "petroleum_refinery", Category: "production",
		BaseCost:    map[string]float64{"steel": 5e9, "electricity": 2e9, "uranium": 300e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "oil", Value: 81.92}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Advanced petroleum refinery. +81.92 oil/tick (12 workers).",
		LineageKey:  "organic_extraction", LineageTier: 11,
		WorkerDomain: "lumber", WorkerCapacity: 12,
		EpochKey: "electric_era", OutputResource: "oil",
	})
	// tier 12 — modern_age  rate=163.84  output=oil
	b = append(b, BuildingDef{
		Name: "Oil Platform", Key: "oil_platform", Category: "production",
		BaseCost:    map[string]float64{"steel": 28e9, "electricity": 10e9, "data": 800e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "oil", Value: 163.84}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Offshore AI-monitored oil platform. +163.84 oil/tick (14 workers).",
		LineageKey:  "organic_extraction", LineageTier: 12,
		WorkerDomain: "lumber", WorkerCapacity: 14,
		EpochKey: "digital_era", OutputResource: "oil",
	})
	// tier 13 — information_age  rate=327.68  output=oil
	b = append(b, BuildingDef{
		Name: "Smart Refinery", Key: "smart_refinery", Category: "production",
		BaseCost:    map[string]float64{"electricity": 75e9, "data": 7e9, "steel": 140e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "oil", Value: 327.68}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "AI-optimised smart petroleum refinery. +327.68 oil/tick (15 workers).",
		LineageKey:  "organic_extraction", LineageTier: 13,
		WorkerDomain: "lumber", WorkerCapacity: 15,
		EpochKey: "digital_era", OutputResource: "oil",
	})
	// tier 14 — digital_age  rate=655.36  output=nanobots
	b = append(b, BuildingDef{
		Name: "Bio Fabrication Lab", Key: "bio_fabrication_lab", Category: "production",
		BaseCost:    map[string]float64{"electricity": 380e9, "data": 45e9, "steel": 550e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "nanobots", Value: 655.36}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Digital-biological nanofabrication. +655.36 nanobots/tick (16 workers).",
		LineageKey:  "organic_extraction", LineageTier: 14,
		WorkerDomain: "lumber", WorkerCapacity: 16,
		EpochKey: "digital_era", OutputResource: "nanobots",
	})
	// tier 15 — cyberpunk_age  rate=1310.72  output=nanobots
	b = append(b, BuildingDef{
		Name: "Nanobot Vat", Key: "nanobot_vat", Category: "production",
		BaseCost:    map[string]float64{"data": 180e9, "crypto": 1.2e12, "electricity": 2.5e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "nanobots", Value: 1310.72}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Vat-grown nanobot manufacturing. +1310.72 nanobots/tick (18 workers).",
		LineageKey:  "organic_extraction", LineageTier: 15,
		WorkerDomain: "lumber", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "nanobots",
	})
	// tier 16 — fusion_age  rate=2621.44  output=nanobots
	b = append(b, BuildingDef{
		Name: "Molecular Synthesizer", Key: "molecular_synthesizer", Category: "production",
		BaseCost:    map[string]float64{"plasma": 4e12, "electricity": 14e12, "steel": 18e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "nanobots", Value: 2621.44}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Plasma-powered molecular synthesis. +2621.44 nanobots/tick (20 workers).",
		LineageKey:  "organic_extraction", LineageTier: 16,
		WorkerDomain: "lumber", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "nanobots",
	})
	// tier 17 — space_age  rate=5242.88  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Quantum Organic Extractor", Key: "quantum_organic_extractor", Category: "production",
		BaseCost:    map[string]float64{"titanium": 75e12, "plasma": 35e12, "electricity": 90e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 5242.88}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Quantum-state organic matter extraction. +5242.88 quantum_flux/tick (20 workers).",
		LineageKey:  "organic_extraction", LineageTier: 17,
		WorkerDomain: "lumber", WorkerCapacity: 20,
		EpochKey: "neon_era", OutputResource: "quantum_flux",
	})
	// tier 18 — interstellar_age  rate=10485.76  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Reality Matter Weaver", Key: "reality_matter_weaver", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 90e12, "titanium": 750e12, "plasma": 450e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 10485.76}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Weaves reality matter into quantum flux. +10485.76 quantum_flux/tick (25 workers).",
		LineageKey:  "organic_extraction", LineageTier: 18,
		WorkerDomain: "lumber", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})
	// tier 19 — galactic_age  rate=20971.52  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Cosmic Organic Works", Key: "cosmic_organic_works", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 180e12, "dark_matter": 900e12, "titanium": 4.5e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 20971.52}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Galactic-scale cosmic organic works. +20971.52 quantum_flux/tick (25 workers).",
		LineageKey:  "organic_extraction", LineageTier: 19,
		WorkerDomain: "lumber", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})
	// tier 20 — quantum_age  rate=41943.04  output=quantum_flux
	b = append(b, BuildingDef{
		Name: "Reality Harvester", Key: "reality_harvester", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 180e12, "antimatter": 55e15, "dark_matter": 45e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 41943.04}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Harvests raw quantum flux from reality itself. +41943.04 quantum_flux/tick (30 workers).",
		LineageKey:  "organic_extraction", LineageTier: 20,
		WorkerDomain: "lumber", WorkerCapacity: 30,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	return b
}
