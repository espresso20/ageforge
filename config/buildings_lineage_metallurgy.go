package config

// buildingsLineageMetallurgy returns lineages 10-13:
// culture_arts, metallurgy, energy, hacker.
func buildingsLineageMetallurgy() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 11 — METALLURGY (lineageKey: "metallurgy", domain: "metallurgy")
	// starts at iron_age (tier 0)
	// iron tiers 0-2: rate = 0.10 * 2^tier
	// steel tiers 3-8: rate = 0.50 * 2^(tier-3)
	// titanium tiers 9-11: rate = 0.50 * 2^(tier-3) continued
	// dark_matter tiers 12-14; antimatter tiers 15-16; quantum_flux tier 17
	// CostScale: 1.35  Category: "production"
	// =========================================================================

	// tier 0 — iron_age  output=iron  rate=0.10
	b = append(b, BuildingDef{
		Name: "Smelter", Key: "smelter", Category: "production",
		BaseCost:    map[string]float64{"stone": 5500, "iron": 2500, "coal": 1000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.10}},
		BuildTicks:  400,
		RequiredAge: "iron_age",
		Description: "Smelts iron ore into usable iron. +0.10 iron/tick (4 workers).",
		LineageKey:  "metallurgy", LineageTier: 0,
		WorkerDomain: "metallurgy", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 1 — classical_age  output=iron  rate=0.20
	b = append(b, BuildingDef{
		Name: "Forge", Key: "forge", Category: "production",
		BaseCost:    map[string]float64{"stone": 36000, "gold": 12000, "coal": 5000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.20}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "A proper forge for working iron. +0.20 iron/tick (4 workers).",
		LineageKey:  "metallurgy", LineageTier: 1,
		WorkerDomain: "metallurgy", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 2 — medieval_age  output=iron  rate=0.40
	b = append(b, BuildingDef{
		Name: "Ironmonger", Key: "ironmonger", Category: "production",
		BaseCost:    map[string]float64{"stone": 200000, "gold": 65000, "iron": 25000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.40}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "Specialist iron trade and metalworking. +0.40 iron/tick (5 workers).",
		LineageKey:  "metallurgy", LineageTier: 2,
		WorkerDomain: "metallurgy", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 3 — renaissance_age  output=steel  rate=0.50
	b = append(b, BuildingDef{
		Name: "Foundry", Key: "foundry", Category: "production",
		BaseCost:    map[string]float64{"gold": 650000, "steel": 220000, "coal": 80000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 0.50}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "Produces steel by alloying iron and carbon. +0.50 steel/tick (5 workers).",
		LineageKey:  "metallurgy", LineageTier: 3,
		WorkerDomain: "metallurgy", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 4 — colonial_age  output=steel  rate=1.0
	b = append(b, BuildingDef{
		Name: "Iron Works", Key: "iron_works", Category: "production",
		BaseCost:    map[string]float64{"gold": 3.8e6, "steel": 2e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 1.0}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "Colonial iron works processing steel. +1.0 steel/tick (6 workers).",
		LineageKey:  "metallurgy", LineageTier: 4,
		WorkerDomain: "metallurgy", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 5 — industrial_age  output=steel  rate=2.0
	b = append(b, BuildingDef{
		Name: "Steel Mill", Key: "steel_mill", Category: "production",
		BaseCost:    map[string]float64{"steel": 28e6, "coal": 12e6, "gold": 16e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 2.0}},
		BuildTicks:  5000,
		RequiredAge: "industrial_age",
		Description: "Industrial-scale steel production. +2.0 steel/tick (7 workers).",
		LineageKey:  "metallurgy", LineageTier: 5,
		WorkerDomain: "metallurgy", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 6 — victorian_age  output=steel  rate=4.0
	b = append(b, BuildingDef{
		Name: "Bessemer Plant", Key: "bessemer_plant", Category: "production",
		BaseCost:    map[string]float64{"steel": 195e6, "coal": 100e6, "gold": 120e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 4.0}},
		BuildTicks:  8000,
		RequiredAge: "victorian_age",
		Description: "Bessemer converter for mass steel production. +4.0 steel/tick (8 workers).",
		LineageKey:  "metallurgy", LineageTier: 6,
		WorkerDomain: "metallurgy", WorkerCapacity: 8,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 7 — electric_age  output=steel  rate=8.0
	b = append(b, BuildingDef{
		Name: "Electric Arc Furnace", Key: "electric_arc_furnace", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.2e9, "electricity": 500e6, "gold": 700e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 8.0}},
		BuildTicks:  12000,
		RequiredAge: "electric_age",
		Description: "Electric arc furnace for high-grade steel. +8.0 steel/tick (9 workers).",
		LineageKey:  "metallurgy", LineageTier: 7,
		WorkerDomain: "metallurgy", WorkerCapacity: 9,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 8 — atomic_age  output=steel  rate=16.0
	b = append(b, BuildingDef{
		Name: "Advanced Alloy Plant", Key: "advanced_alloy_plant", Category: "production",
		BaseCost:    map[string]float64{"steel": 6e9, "electricity": 2.5e9, "uranium": 400e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 16.0}},
		BuildTicks:  20000,
		RequiredAge: "atomic_age",
		Description: "Atomic-era advanced alloy manufacturing. +16.0 steel/tick (10 workers).",
		LineageKey:  "metallurgy", LineageTier: 8,
		WorkerDomain: "metallurgy", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 9 — modern_age  output=titanium  rate=0.50 (reset for new metal)
	b = append(b, BuildingDef{
		Name: "Titanium Smelter", Key: "titanium_smelter", Category: "production",
		BaseCost:    map[string]float64{"steel": 34e9, "electricity": 13e9, "data": 1.2e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "titanium", Value: 0.50}},
		BuildTicks:  20000,
		RequiredAge: "modern_age",
		Description: "Smelts titanium ore into refined titanium. +0.50 titanium/tick (11 workers).",
		LineageKey:  "metallurgy", LineageTier: 9,
		WorkerDomain: "metallurgy", WorkerCapacity: 11,
		EpochKey: "digital_era", OutputResource: "titanium",
	})
	// tier 10 — information_age  output=titanium  rate=1.0
	b = append(b, BuildingDef{
		Name: "Aerospace Foundry", Key: "aerospace_foundry", Category: "production",
		BaseCost:    map[string]float64{"electricity": 95e9, "data": 10e9, "steel": 175e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "titanium", Value: 1.0}},
		BuildTicks:  30000,
		RequiredAge: "information_age",
		Description: "Precision aerospace-grade titanium foundry. +1.0 titanium/tick (12 workers).",
		LineageKey:  "metallurgy", LineageTier: 10,
		WorkerDomain: "metallurgy", WorkerCapacity: 12,
		EpochKey: "digital_era", OutputResource: "titanium",
	})
	// tier 11 — digital_age  output=titanium  rate=2.0
	b = append(b, BuildingDef{
		Name: "Nano Alloy Plant", Key: "nano_alloy_plant", Category: "production",
		BaseCost:    map[string]float64{"electricity": 460e9, "data": 57e9, "steel": 670e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "titanium", Value: 2.0}},
		BuildTicks:  30000,
		RequiredAge: "digital_age",
		Description: "Nano-scale titanium alloy production. +2.0 titanium/tick (13 workers).",
		LineageKey:  "metallurgy", LineageTier: 11,
		WorkerDomain: "metallurgy", WorkerCapacity: 13,
		EpochKey: "digital_era", OutputResource: "titanium",
	})
	// tier 12 — cyberpunk_age  output=dark_matter  rate=1.0 (new material)
	b = append(b, BuildingDef{
		Name: "Dark Matter Refinery", Key: "dark_matter_refinery", Category: "production",
		BaseCost:    map[string]float64{"data": 220e9, "crypto": 1.15e12, "electricity": 2.3e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 1.0}},
		BuildTicks:  30000,
		RequiredAge: "cyberpunk_age",
		Description: "Refines dark matter crystals into usable dark matter. +1.0 dark_matter/tick (14 workers).",
		LineageKey:  "metallurgy", LineageTier: 12,
		WorkerDomain: "metallurgy", WorkerCapacity: 14,
		EpochKey: "neon_era", OutputResource: "dark_matter",
	})
	// tier 13 — fusion_age  output=dark_matter  rate=2.0
	b = append(b, BuildingDef{
		Name: "Exotic Matter Forge", Key: "exotic_matter_forge", Category: "production",
		BaseCost:    map[string]float64{"plasma": 4.8e12, "electricity": 15e12, "steel": 20e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 2.0}},
		BuildTicks:  30000,
		RequiredAge: "fusion_age",
		Description: "Plasma-forged exotic matter manufacturing. +2.0 dark_matter/tick (16 workers).",
		LineageKey:  "metallurgy", LineageTier: 13,
		WorkerDomain: "metallurgy", WorkerCapacity: 16,
		EpochKey: "neon_era", OutputResource: "dark_matter",
	})
	// tier 14 — space_age  output=dark_matter  rate=4.0
	b = append(b, BuildingDef{
		Name: "Orbital Refinery", Key: "orbital_refinery", Category: "production",
		BaseCost:    map[string]float64{"titanium": 88e12, "plasma": 44e12, "electricity": 110e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 4.0}},
		BuildTicks:  30000,
		RequiredAge: "space_age",
		Description: "Zero-gravity orbital dark matter refinery. +4.0 dark_matter/tick (18 workers).",
		LineageKey:  "metallurgy", LineageTier: 14,
		WorkerDomain: "metallurgy", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "dark_matter",
	})
	// tier 15 — interstellar_age  output=antimatter  rate=2.0 (new material)
	b = append(b, BuildingDef{
		Name: "Antimatter Forge", Key: "antimatter_forge", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 105e12, "titanium": 820e12, "plasma": 510e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "antimatter", Value: 2.0}},
		BuildTicks:  30000,
		RequiredAge: "interstellar_age",
		Description: "Forges antimatter from dark matter reactions. +2.0 antimatter/tick (20 workers).",
		LineageKey:  "metallurgy", LineageTier: 15,
		WorkerDomain: "metallurgy", WorkerCapacity: 20,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 16 — galactic_age  output=antimatter  rate=4.0
	b = append(b, BuildingDef{
		Name: "Stellar Metallurgy", Key: "stellar_metallurgy", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 210e12, "dark_matter": 1.05e15, "titanium": 5.2e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "antimatter", Value: 4.0}},
		BuildTicks:  30000,
		RequiredAge: "galactic_age",
		Description: "Stellar-scale antimatter metallurgy. +4.0 antimatter/tick (22 workers).",
		LineageKey:  "metallurgy", LineageTier: 16,
		WorkerDomain: "metallurgy", WorkerCapacity: 22,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 17 — quantum_age  output=quantum_flux  rate=4.0
	b = append(b, BuildingDef{
		Name: "Quantum Metal Works", Key: "quantum_metal_works", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 225e12, "antimatter": 67e15, "dark_matter": 57e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 4.0}},
		BuildTicks:  30000,
		RequiredAge: "quantum_age",
		Description: "Quantum-state metalworking across dimensions. +4.0 quantum_flux/tick (25 workers).",
		LineageKey:  "metallurgy", LineageTier: 17,
		WorkerDomain: "metallurgy", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	return b
}
