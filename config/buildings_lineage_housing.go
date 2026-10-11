package config

// buildingsLineageHousing returns all production, housing, research, and military buildings
// for the 13 lineage chains introduced in Phase 10 of the economy redesign.
// Storage buildings and wonders remain in baseBuildingsRaw().
// All fields are set inline; no separate buildingMeta() entries needed for these buildings.
func buildingsLineageHousing() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 1 — HOUSING (lineageKey: "housing", no workers, no output)
	// CostScale: 1.12  Category: "housing"
	// Effect: capacity / population
	// =========================================================================

	// tier 0 — primitive_age
	// BuildTicks lowered from 80 → 8 so the player can get pop cap before first workers starve.
	b = append(b, BuildingDef{
		Name: "Hut", Key: "hut", Category: "housing",
		BaseCost:    map[string]float64{"wood": 14},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 10}},
		BuildTicks:  8,
		RequiredAge: "primitive_age",
		Description: "A crude shelter of sticks and leaves.",
		LineageKey:  "housing", LineageTier: 0,
		EpochKey: "stone_era",
	})
	// tier 1 — stone_age
	b = append(b, BuildingDef{
		Name: "Longhouse", Key: "longhouse", Category: "housing",
		BaseCost:    map[string]float64{"wood": 110, "stone": 74},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 25}},
		BuildTicks:  60,
		RequiredAge: "stone_age",
		Description: "A communal longhouse.",
		LineageKey:  "housing", LineageTier: 1,
		EpochKey: "stone_era",
	})
	// tier 2 — bronze_age
	b = append(b, BuildingDef{
		Name: "House", Key: "house", Category: "housing",
		BaseCost:    map[string]float64{"wood": 550, "stone": 370},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 50}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "A sturdy family dwelling.",
		LineageKey:  "housing", LineageTier: 2,
		EpochKey: "stone_era",
	})
	// tier 3 — iron_age
	b = append(b, BuildingDef{
		Name: "Townhouse", Key: "townhouse", Category: "housing",
		BaseCost:    map[string]float64{"stone": 5500, "iron": 1800},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 80}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "A multi-floor townhouse.",
		LineageKey:  "housing", LineageTier: 3,
		EpochKey: "iron_era",
	})
	// tier 4 — classical_age
	b = append(b, BuildingDef{
		Name: "Villa", Key: "villa", Category: "housing",
		BaseCost:    map[string]float64{"stone": 37000, "gold": 9200},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 120}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "An elegant classical villa.",
		LineageKey:  "housing", LineageTier: 4,
		EpochKey: "iron_era",
	})
	// tier 5 — medieval_age
	b = append(b, BuildingDef{
		Name: "Manor", Key: "manor", Category: "housing",
		BaseCost:    map[string]float64{"stone": 180000, "gold": 46000},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 240}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "A lord's country manor.",
		LineageKey:  "housing", LineageTier: 5,
		EpochKey: "iron_era",
	})
	// tier 6 — renaissance_age
	b = append(b, BuildingDef{
		Name: "Estate", Key: "estate", Category: "housing",
		BaseCost:    map[string]float64{"stone": 550000, "gold": 230000},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 480}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "A grand estate with grounds.",
		LineageKey:  "housing", LineageTier: 6,
		EpochKey: "steel_era",
	})
	// tier 7 — colonial_age
	b = append(b, BuildingDef{
		Name: "Settlement Block", Key: "settlement_block", Category: "housing",
		BaseCost:    map[string]float64{"wood": 1800000, "stone": 1400000},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 960}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "A colonial settlement block.",
		LineageKey:  "housing", LineageTier: 7,
		EpochKey: "steel_era",
	})
	// tier 8 — industrial_age
	b = append(b, BuildingDef{
		Name: "Tenement", Key: "tenement", Category: "housing",
		BaseCost:    map[string]float64{"stone": 1.1e07, "iron": 7400000},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 1920}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "Dense worker housing.",
		LineageKey:  "housing", LineageTier: 8,
		EpochKey: "steel_era",
	})
	// tier 9 — victorian_age
	b = append(b, BuildingDef{
		Name: "Row House", Key: "row_house", Category: "housing",
		BaseCost:    map[string]float64{"steel": 7.4e07, "stone": 5.5e07},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 3840}},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "A Victorian terrace of row houses.",
		LineageKey:  "housing", LineageTier: 9,
		EpochKey: "electric_era",
	})
	// tier 10 — electric_age
	b = append(b, BuildingDef{
		Name: "Apartment Block", Key: "apartment_block", Category: "housing",
		BaseCost:    map[string]float64{"steel": 4.6e08, "electricity": 1.8e08},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 7680}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "A modern apartment block.",
		LineageKey:  "housing", LineageTier: 10,
		EpochKey: "electric_era",
	})
	// tier 11 — atomic_age
	b = append(b, BuildingDef{
		Name: "Housing Project", Key: "housing_project", Category: "housing",
		BaseCost:    map[string]float64{"steel": 2.8e09, "electricity": 9.2e08},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 15360}},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "Government-built housing towers.",
		LineageKey:  "housing", LineageTier: 11,
		EpochKey: "electric_era",
	})
	// tier 12 — modern_age
	b = append(b, BuildingDef{
		Name: "Tower Block", Key: "tower_block", Category: "housing",
		BaseCost:    map[string]float64{"steel": 1.8e10, "electricity": 7.4e09},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 30720}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "A soaring residential tower.",
		LineageKey:  "housing", LineageTier: 12,
		EpochKey: "digital_era",
	})
	// tier 13 — information_age
	b = append(b, BuildingDef{
		Name: "Smart Complex", Key: "smart_complex", Category: "housing",
		BaseCost:     map[string]float64{"steel": 1.1e11, "electricity": 4.6e10, "data": 4.6e09},
		CostScale:    1.13,
		Effects:      []Effect{{Type: "capacity", Target: "population", Value: 61440}},
		BuildTicks:   3600,
		RequiredAge:  "information_age",
		RequiredTech: "internet_of_things",
		Description:  "AI-managed smart living complex.",
		LineageKey:   "housing", LineageTier: 13,
		EpochKey: "digital_era",
	})
	// tier 14 — digital_age
	b = append(b, BuildingDef{
		Name: "Megaplex", Key: "megaplex", Category: "housing",
		BaseCost:    map[string]float64{"steel": 6.5e11, "electricity": 2.8e11, "data": 2.8e10},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 122880}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "A self-contained urban megaplex.",
		LineageKey:  "housing", LineageTier: 14,
		EpochKey: "digital_era",
	})
	// tier 15 — cyberpunk_age
	b = append(b, BuildingDef{
		Name: "Arcology Pod", Key: "arcology_pod", Category: "housing",
		BaseCost:    map[string]float64{"steel": 4.6e12, "data": 1.4e11, "crypto": 4.6e11},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 245760}},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "A self-sustaining arcology pod.",
		LineageKey:  "housing", LineageTier: 15,
		EpochKey: "neon_era",
	})
	// tier 16 — fusion_age
	b = append(b, BuildingDef{
		Name: "Habitat Ring", Key: "habitat_ring", Category: "housing",
		BaseCost:    map[string]float64{"steel": 3.2e13, "plasma": 4.6e12, "electricity": 1.4e13},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 491520}},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "A rotating habitat ring in orbit.",
		LineageKey:  "housing", LineageTier: 16,
		EpochKey: "neon_era",
	})
	// tier 17 — space_age
	b = append(b, BuildingDef{
		Name: "Orbital Habitat", Key: "orbital_habitat", Category: "housing",
		BaseCost:    map[string]float64{"titanium": 4.6e13, "plasma": 2.8e13, "steel": 1.8e14},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 983040}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "A vast orbital habitat complex.",
		LineageKey:  "housing", LineageTier: 17,
		EpochKey: "neon_era",
	})
	// tier 18 — interstellar_age
	b = append(b, BuildingDef{
		Name: "Generation Ship", Key: "generation_ship", Category: "housing",
		BaseCost:    map[string]float64{"titanium": 4.6e14, "dark_matter": 4.6e13, "plasma": 2.8e14},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 1966080}},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "A ship that houses entire generations.",
		LineageKey:  "housing", LineageTier: 18,
		EpochKey: "cosmic_era",
	})
	// tier 19 — galactic_age
	b = append(b, BuildingDef{
		Name: "Dyson Sphere Habitat", Key: "dyson_sphere_habitat", Category: "housing",
		BaseCost:    map[string]float64{"dark_matter": 4.6e14, "titanium": 4.6e15, "antimatter": 9.2e13},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 3932160}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Living quarters within a Dyson sphere.",
		LineageKey:  "housing", LineageTier: 19,
		EpochKey: "cosmic_era",
	})
	// tier 20 — quantum_age
	b = append(b, BuildingDef{
		Name: "Reality Fold", Key: "reality_fold", Category: "housing",
		BaseCost:    map[string]float64{"quantum_flux": 9.2e13, "antimatter": 4.6e16, "dark_matter": 3.7e16},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 7864320}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "A folded-reality habitation zone.",
		LineageKey:  "housing", LineageTier: 20,
		EpochKey: "cosmic_era",
	})
	// tier 21 — transcendent_age
	b = append(b, BuildingDef{
		Name: "Transcendent Nexus", Key: "transcendent_nexus", Category: "housing",
		BaseCost:    map[string]float64{"quantum_flux": 9.2e14, "antimatter": 4.6e17, "dark_matter": 3.7e17},
		CostScale:   1.13,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 1.572864e07}},
		BuildTicks:  3600,
		RequiredAge: "transcendent_age",
		Description: "A transcendent dimensional habitation nexus.",
		LineageKey:  "housing", LineageTier: 21,
		EpochKey: "cosmic_era",
	})

	return b
}
