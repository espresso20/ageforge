package config

// buildingsLineageHousing returns all production, housing, research, and military buildings
// for the 13 lineage chains introduced in Phase 10 of the economy redesign.
// Storage buildings and wonders remain in baseBuildingsRaw().
// A row sets its own fields; the ones every row shares come from the lineageSpec at the foot,
// so no separate buildingMeta() entries are needed for these buildings.
func buildingsLineageHousing() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 1 — HOUSING (lineageKey: "housing", no workers, no output)
	// CostScale: 1.13  Category: "housing"  (both in the lineageSpec at the foot)
	// Effect: capacity / population
	// =========================================================================

	// tier 0 — primitive_age
	// BuildTicks lowered from 80 → 8 so the player can get pop cap before first workers starve.
	b = append(b, BuildingDef{
		Name: "Hut", Key: "hut",
		BaseCost:    map[string]float64{"wood": 14},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 10}},
		BuildTicks:  8,
		RequiredAge: "primitive_age",
		Description: "A crude shelter of sticks and leaves.",
	})
	// tier 1 — stone_age
	b = append(b, BuildingDef{
		Name: "Longhouse", Key: "longhouse",
		BaseCost:    map[string]float64{"wood": 110, "stone": 74},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 25}},
		BuildTicks:  60,
		RequiredAge: "stone_age",
		Description: "A communal longhouse.",
	})
	// tier 2 — bronze_age
	b = append(b, BuildingDef{
		Name: "House", Key: "house",
		BaseCost:    map[string]float64{"wood": 550, "stone": 370},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 50}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "A sturdy family dwelling.",
	})
	// tier 3 — iron_age
	b = append(b, BuildingDef{
		Name: "Townhouse", Key: "townhouse",
		BaseCost:    map[string]float64{"stone": 5500, "iron": 1800},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 80}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "A multi-floor townhouse.",
	})
	// tier 4 — classical_age
	b = append(b, BuildingDef{
		Name: "Villa", Key: "villa",
		BaseCost:    map[string]float64{"stone": 37000, "gold": 9200},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 120}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "An elegant classical villa.",
	})
	// tier 5 — medieval_age
	b = append(b, BuildingDef{
		Name: "Manor", Key: "manor",
		BaseCost:    map[string]float64{"stone": 180000, "gold": 46000},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 240}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "A lord's country manor.",
	})
	// tier 6 — renaissance_age
	b = append(b, BuildingDef{
		Name: "Estate", Key: "estate",
		BaseCost:    map[string]float64{"stone": 550000, "gold": 230000},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 480}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "A grand estate with grounds.",
	})
	// tier 7 — colonial_age
	b = append(b, BuildingDef{
		Name: "Settlement Block", Key: "settlement_block",
		BaseCost:    map[string]float64{"stone": 5.55 * Million},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 960}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "A colonial settlement block.",
	})
	// tier 8 — industrial_age
	b = append(b, BuildingDef{
		Name: "Tenement", Key: "tenement",
		BaseCost:    map[string]float64{"stone": 11 * Million, "iron": 7.4 * Million},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 1920}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "Dense worker housing.",
	})
	// tier 9 — victorian_age
	b = append(b, BuildingDef{
		Name: "Row House", Key: "row_house",
		BaseCost:    map[string]float64{"steel": 74 * Million, "stone": 55 * Million},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 3840}},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "A Victorian terrace of row houses.",
	})
	// tier 10 — electric_age
	b = append(b, BuildingDef{
		Name: "Apartment Block", Key: "apartment_block",
		BaseCost:    map[string]float64{"steel": 460 * Million, "electricity": 180 * Million},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 7680}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "A modern apartment block.",
	})
	// tier 11 — atomic_age
	b = append(b, BuildingDef{
		Name: "Housing Project", Key: "housing_project",
		BaseCost:    map[string]float64{"steel": 2.8 * Billion, "electricity": 920 * Million},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 15360}},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "Government-built housing towers.",
	})
	// tier 12 — modern_age
	b = append(b, BuildingDef{
		Name: "Tower Block", Key: "tower_block",
		BaseCost:    map[string]float64{"steel": 18 * Billion, "electricity": 7.4 * Billion},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 30720}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "A soaring residential tower.",
	})
	// tier 13 — information_age
	b = append(b, BuildingDef{
		Name: "Smart Complex", Key: "smart_complex",
		BaseCost:     map[string]float64{"steel": 110 * Billion, "electricity": 46 * Billion, "data": 4.6 * Billion},
		Effects:      []Effect{{Type: "capacity", Target: "population", Value: 61440}},
		BuildTicks:   3600,
		RequiredAge:  "information_age",
		RequiredTech: "internet_of_things",
		Description:  "AI-managed smart living complex.",
	})
	// tier 14 — digital_age
	b = append(b, BuildingDef{
		Name: "Megaplex", Key: "megaplex",
		BaseCost:    map[string]float64{"steel": 650 * Billion, "electricity": 280 * Billion, "data": 28 * Billion},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 122880}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "A self-contained urban megaplex.",
	})
	// tier 15 — cyberpunk_age
	b = append(b, BuildingDef{
		Name: "Arcology Pod", Key: "arcology_pod",
		BaseCost:    map[string]float64{"steel": 4.6 * Trillion, "data": 140 * Billion, "crypto": 460 * Billion},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 245760}},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "A self-sustaining arcology pod.",
	})
	// tier 16 — fusion_age
	b = append(b, BuildingDef{
		Name: "Habitat Ring", Key: "habitat_ring",
		BaseCost:    map[string]float64{"steel": 32 * Trillion, "plasma": 4.6 * Trillion, "electricity": 14 * Trillion},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 491520}},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "A rotating habitat ring in orbit.",
	})
	// tier 17 — space_age
	b = append(b, BuildingDef{
		Name: "Orbital Habitat", Key: "orbital_habitat",
		BaseCost:    map[string]float64{"titanium": 46 * Trillion, "plasma": 28 * Trillion, "steel": 180 * Trillion},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 983040}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "A vast orbital habitat complex.",
	})
	// tier 18 — interstellar_age
	b = append(b, BuildingDef{
		Name: "Generation Ship", Key: "generation_ship",
		BaseCost:    map[string]float64{"titanium": 460 * Trillion, "dark_matter": 46 * Trillion, "plasma": 280 * Trillion},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 1966080}},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "A ship that houses entire generations.",
	})
	// tier 19 — galactic_age
	b = append(b, BuildingDef{
		Name: "Dyson Sphere Habitat", Key: "dyson_sphere_habitat",
		BaseCost:    map[string]float64{"dark_matter": 460 * Trillion, "titanium": 4.6 * Quadrillion, "antimatter": 92 * Trillion},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 3932160}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Living quarters within a Dyson sphere.",
	})
	// tier 20 — quantum_age
	b = append(b, BuildingDef{
		Name: "Reality Fold", Key: "reality_fold",
		BaseCost:    map[string]float64{"quantum_flux": 92 * Trillion, "antimatter": 46 * Quadrillion, "dark_matter": 37 * Quadrillion},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 7864320}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "A folded-reality habitation zone.",
	})
	// tier 21 — transcendent_age
	b = append(b, BuildingDef{
		Name: "Transcendent Nexus", Key: "transcendent_nexus",
		BaseCost:    map[string]float64{"quantum_flux": 920 * Trillion, "antimatter": 460 * Quadrillion, "dark_matter": 370 * Quadrillion},
		Effects:     []Effect{{Type: "capacity", Target: "population", Value: 15.72864 * Million}},
		BuildTicks:  3600,
		RequiredAge: "transcendent_age",
		Description: "A transcendent dimensional habitation nexus.",
	})

	return finishLineage(b, lineageSpec{
		Key: "housing", Category: "housing", Scale: 1.13,
	})
}
