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
		BaseCost:    map[string]float64{"wood": 15},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 10}},
		BuildTicks:  8,
		RequiredAge: "primitive_age",
		Description: "A crude shelter of sticks and leaves. +10 pop cap.",
		LineageKey:  "housing", LineageTier: 0,
		EpochKey: "stone_era",
	})
	// tier 1 — stone_age
	b = append(b, BuildingDef{
		Name: "Longhouse", Key: "longhouse", Category: "housing",
		BaseCost:    map[string]float64{"wood": 120, "stone": 80},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 25}},
		BuildTicks:  60,
		RequiredAge: "stone_age",
		Description: "A communal longhouse. +25 pop cap.",
		LineageKey:  "housing", LineageTier: 1,
		EpochKey: "stone_era",
	})
	// tier 2 — bronze_age
	b = append(b, BuildingDef{
		Name: "House", Key: "house", Category: "housing",
		BaseCost:    map[string]float64{"wood": 600, "stone": 400},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 50}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "A sturdy family dwelling. +50 pop cap.",
		LineageKey:  "housing", LineageTier: 2,
		EpochKey: "stone_era",
	})
	// tier 3 — iron_age
	b = append(b, BuildingDef{
		Name: "Townhouse", Key: "townhouse", Category: "housing",
		BaseCost:    map[string]float64{"stone": 6000, "iron": 2000},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 80}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "A multi-floor townhouse. +80 pop cap.",
		LineageKey:  "housing", LineageTier: 3,
		EpochKey: "iron_era",
	})
	// tier 4 — classical_age
	b = append(b, BuildingDef{
		Name: "Villa", Key: "villa", Category: "housing",
		BaseCost:    map[string]float64{"stone": 40000, "gold": 10000},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 120}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "An elegant classical villa. +120 pop cap.",
		LineageKey:  "housing", LineageTier: 4,
		EpochKey: "iron_era",
	})
	// tier 5 — medieval_age
	b = append(b, BuildingDef{
		Name: "Manor", Key: "manor", Category: "housing",
		BaseCost:    map[string]float64{"stone": 200000, "gold": 50000},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 240}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "A lord's country manor. +240 pop cap.",
		LineageKey:  "housing", LineageTier: 5,
		EpochKey: "iron_era",
	})
	// tier 6 — renaissance_age
	b = append(b, BuildingDef{
		Name: "Estate", Key: "estate", Category: "housing",
		BaseCost:    map[string]float64{"stone": 600000, "gold": 250000},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 350}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "A grand estate with grounds. +350 pop cap.",
		LineageKey:  "housing", LineageTier: 6,
		EpochKey: "steel_era",
	})
	// tier 7 — colonial_age
	b = append(b, BuildingDef{
		Name: "Settlement Block", Key: "settlement_block", Category: "housing",
		BaseCost:    map[string]float64{"wood": 2e6, "stone": 1.5e6},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 600}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "A colonial settlement block. +600 pop cap.",
		LineageKey:  "housing", LineageTier: 7,
		EpochKey: "steel_era",
	})
	// tier 8 — industrial_age
	b = append(b, BuildingDef{
		Name: "Tenement", Key: "tenement", Category: "housing",
		BaseCost:    map[string]float64{"stone": 12e6, "iron": 8e6},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 1000}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Dense worker housing. +1000 pop cap.",
		LineageKey:  "housing", LineageTier: 8,
		EpochKey: "steel_era",
	})
	// tier 9 — victorian_age
	b = append(b, BuildingDef{
		Name: "Row House", Key: "row_house", Category: "housing",
		BaseCost:    map[string]float64{"steel": 80e6, "stone": 60e6},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 1800}},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "A Victorian terrace of row houses. +1800 pop cap.",
		LineageKey:  "housing", LineageTier: 9,
		EpochKey: "electric_era",
	})
	// tier 10 — electric_age
	b = append(b, BuildingDef{
		Name: "Apartment Block", Key: "apartment_block", Category: "housing",
		BaseCost:    map[string]float64{"steel": 500e6, "electricity": 200e6},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 3200}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "A modern apartment block. +3200 pop cap.",
		LineageKey:  "housing", LineageTier: 10,
		EpochKey: "electric_era",
	})
	// tier 11 — atomic_age
	b = append(b, BuildingDef{
		Name: "Housing Project", Key: "housing_project", Category: "housing",
		BaseCost:    map[string]float64{"steel": 3e9, "electricity": 1e9},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 5500}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Government-built housing towers. +5500 pop cap.",
		LineageKey:  "housing", LineageTier: 11,
		EpochKey: "electric_era",
	})
	// tier 12 — modern_age
	b = append(b, BuildingDef{
		Name: "Tower Block", Key: "tower_block", Category: "housing",
		BaseCost:    map[string]float64{"steel": 20e9, "electricity": 8e9},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 10000}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "A soaring residential tower. +10000 pop cap.",
		LineageKey:  "housing", LineageTier: 12,
		EpochKey: "digital_era",
	})
	// tier 13 — information_age
	b = append(b, BuildingDef{
		Name: "Smart Complex", Key: "smart_complex", Category: "housing",
		BaseCost:    map[string]float64{"steel": 120e9, "electricity": 50e9, "data": 5e9},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 18000}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "AI-managed smart living complex. +18000 pop cap.",
		LineageKey:  "housing", LineageTier: 13,
		EpochKey: "digital_era",
	})
	// tier 14 — digital_age
	b = append(b, BuildingDef{
		Name: "Megaplex", Key: "megaplex", Category: "housing",
		BaseCost:    map[string]float64{"steel": 700e9, "electricity": 300e9, "data": 30e9},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 32000}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "A self-contained urban megaplex. +32000 pop cap.",
		LineageKey:  "housing", LineageTier: 14,
		EpochKey: "digital_era",
	})
	// tier 15 — cyberpunk_age
	b = append(b, BuildingDef{
		Name: "Arcology Pod", Key: "arcology_pod", Category: "housing",
		BaseCost:    map[string]float64{"steel": 5e12, "data": 150e9, "crypto": 500e9},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 55000}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "A self-sustaining arcology pod. +55000 pop cap.",
		LineageKey:  "housing", LineageTier: 15,
		EpochKey: "neon_era",
	})
	// tier 16 — fusion_age
	b = append(b, BuildingDef{
		Name: "Habitat Ring", Key: "habitat_ring", Category: "housing",
		BaseCost:    map[string]float64{"steel": 35e12, "plasma": 5e12, "electricity": 15e12},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 100000}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "A rotating habitat ring in orbit. +100000 pop cap.",
		LineageKey:  "housing", LineageTier: 16,
		EpochKey: "neon_era",
	})
	// tier 17 — space_age
	b = append(b, BuildingDef{
		Name: "Orbital Habitat", Key: "orbital_habitat", Category: "housing",
		BaseCost:    map[string]float64{"titanium": 50e12, "plasma": 30e12, "steel": 200e12},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 180000}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "A vast orbital habitat complex. +180000 pop cap.",
		LineageKey:  "housing", LineageTier: 17,
		EpochKey: "neon_era",
	})
	// tier 18 — interstellar_age
	b = append(b, BuildingDef{
		Name: "Generation Ship", Key: "generation_ship", Category: "housing",
		BaseCost:    map[string]float64{"titanium": 500e12, "dark_matter": 50e12, "plasma": 300e12},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 320000}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "A ship that houses entire generations. +320000 pop cap.",
		LineageKey:  "housing", LineageTier: 18,
		EpochKey: "cosmic_era",
	})
	// tier 19 — galactic_age
	b = append(b, BuildingDef{
		Name: "Dyson Sphere Habitat", Key: "dyson_sphere_habitat", Category: "housing",
		BaseCost:    map[string]float64{"dark_matter": 500e12, "titanium": 5e15, "antimatter": 100e12},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 600000}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Living quarters within a Dyson sphere. +600000 pop cap.",
		LineageKey:  "housing", LineageTier: 19,
		EpochKey: "cosmic_era",
	})
	// tier 20 — quantum_age
	b = append(b, BuildingDef{
		Name: "Reality Fold", Key: "reality_fold", Category: "housing",
		BaseCost:    map[string]float64{"quantum_flux": 100e12, "antimatter": 50e15, "dark_matter": 40e15},
		CostScale:   1.12,
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 1000000}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "A folded-reality habitation zone. +1000000 pop cap.",
		LineageKey:  "housing", LineageTier: 20,
		EpochKey: "cosmic_era",
	})

	return b
}
