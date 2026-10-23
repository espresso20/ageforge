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
		BaseCost:     map[string]float64{"stone": 11500, "marble": 11500, "iron": 5500, "iron_ore": 5500},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "iron", Value: 15.6}},
		BuildTicks:   400,
		RequiredAge:  "iron_age",
		RequiredTech: "iron_smelting",
		Description:  "A furnace that smelts iron.",
		LineageKey:   "metallurgy", LineageTier: 0, LineKey: "metallurgy_iron",
		WorkerDomain: "metallurgy", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 1 — classical_age  output=iron  rate=0.20
	b = append(b, BuildingDef{
		Name: "Forge", Key: "forge", Category: "production",
		BaseCost:     map[string]float64{"stone": 75000, "marble": 75000, "gold": 51000},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "iron", Value: 30.4}},
		BuildTicks:   600,
		RequiredAge:  "classical_age",
		RequiredTech: "metal_casting",
		Description:  "A proper forge for working iron.",
		LineageKey:   "metallurgy", LineageTier: 1, LineKey: "metallurgy_iron",
		WorkerDomain: "metallurgy", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 2 — medieval_age  output=iron  rate=0.40
	b = append(b, BuildingDef{
		Name: "Ironmonger", Key: "ironmonger", Category: "production",
		BaseCost:    map[string]float64{"stone": 850000, "gold": 280000, "iron": 55000, "iron_ore": 55000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 143}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "Specialist iron trade and metalworking.",
		LineageKey:  "metallurgy", LineageTier: 2, LineKey: "metallurgy_iron",
		WorkerDomain: "metallurgy", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 3 — renaissance_age  output=steel  rate=0.50
	// Resource pivot: iron → steel. Validator shows LOW boost (iron:0.40 → steel:0.50) — intentional.
	// Steel is a higher-tier resource; the rate reset to 0.50 is correct for the new tier.
	b = append(b, BuildingDef{
		Name: "Foundry", Key: "foundry", Category: "production",
		BaseCost:     map[string]float64{"gold": 2800000, "steel": 930000, "coal": 340000},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "steel", Value: 261}},
		BuildTicks:   2400,
		RequiredAge:  "renaissance_age",
		RequiredTech: "blast_furnace",
		Description:  "Crucible furnaces that turn out steel.",
		LineageKey:   "metallurgy", LineageTier: 3, LineKey: "metallurgy_steel",
		WorkerDomain: "metallurgy", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 4 — colonial_age  output=steel  rate=1.0
	b = append(b, BuildingDef{
		Name: "Colonial Steelworks", Key: "iron_works", Category: "production",
		BaseCost:     map[string]float64{"gold": 16 * Million, "steel": 8500000},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "steel", Value: 2590}},
		BuildTicks:   3600,
		RequiredAge:  "colonial_age",
		RequiredTech: "coke_smelting",
		Description:  "Colonial iron works processing steel.",
		LineageKey:   "metallurgy", LineageTier: 4, LineKey: "metallurgy_steel",
		WorkerDomain: "metallurgy", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 5 — industrial_age  output=steel  rate=2.0
	b = append(b, BuildingDef{
		Name: "Steel Mill", Key: "steel_mill", Category: "production",
		BaseCost:    map[string]float64{"steel": 120 * Million, "coal": 51 * Million, "gold": 68 * Million},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 41500}},
		BuildTicks:  5000,
		RequiredAge: "industrial_age",
		Description: "Industrial-scale steel production.",
		LineageKey:  "metallurgy", LineageTier: 5, LineKey: "metallurgy_steel",
		WorkerDomain: "metallurgy", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 6 — victorian_age  output=steel  rate=4.0
	b = append(b, BuildingDef{
		Name: "Bessemer Plant", Key: "bessemer_plant", Category: "production",
		BaseCost:  map[string]float64{"steel": 830 * Million, "coal": 420 * Million, "gold": 510 * Million},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "steel", Value: 101000},
			{Type: "production", Target: "iron", Value: 20800},
		},
		BuildTicks:  7020,
		RequiredAge: "victorian_age",
		Description: "Bessemer converter for mass steel and iron production.",
		LineageKey:  "metallurgy", LineageTier: 6, LineKey: "metallurgy_steel",
		WorkerDomain: "metallurgy", WorkerCapacity: 8,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 7 — electric_age  output=steel  rate=8.0
	b = append(b, BuildingDef{
		Name: "Electric Arc Furnace", Key: "electric_arc_furnace", Category: "production",
		BaseCost:  map[string]float64{"steel": 5.1 * Billion, "electricity": 2.1 * Billion, "gold": 3 * Billion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "steel", Value: 659000},
			{Type: "production", Target: "iron", Value: 9010},
		},
		BuildTicks:  7800,
		RequiredAge: "electric_age",
		Description: "Electric arc furnace for high-grade steel and iron.",
		LineageKey:  "metallurgy", LineageTier: 7, LineKey: "metallurgy_steel",
		WorkerDomain: "metallurgy", WorkerCapacity: 9,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 8 — atomic_age  output=steel  rate=16.0
	b = append(b, BuildingDef{
		Name: "Advanced Alloy Plant", Key: "advanced_alloy_plant", Category: "production",
		BaseCost:  map[string]float64{"steel": 25 * Billion, "electricity": 11 * Billion, "uranium": 1.7 * Billion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "steel", Value: 2550000},
			{Type: "production", Target: "iron", Value: 54200},
		},
		BuildTicks:  9360,
		RequiredAge: "atomic_age",
		Description: "Atomic-era advanced alloy manufacturing.",
		LineageKey:  "metallurgy", LineageTier: 8, LineKey: "metallurgy_steel",
		WorkerDomain: "metallurgy", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 9 — modern_age  output=titanium  rate=0.50 (reset for new metal)
	// Resource pivot: steel → titanium. Validator shows 0.03x (steel:16 → titanium:0.50) — intentional.
	// Titanium is a tier-5 resource; starting rate 0.50 matches the iron/steel pivot pattern.
	b = append(b, BuildingDef{
		Name: "Titanium Smelter", Key: "titanium_smelter", Category: "production",
		BaseCost:  map[string]float64{"steel": 70 * Billion, "titanium_ore": 70 * Billion, "electricity": 55 * Billion, "data": 5.1 * Billion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "titanium", Value: 0.5},
			{Type: "production", Target: "steel", Value: 17100000},
		},
		BuildTicks:  9360,
		RequiredAge: "modern_age",
		Description: "Refines titanium and steel for precision work.",
		LineageKey:  "metallurgy", LineageTier: 9, LineKey: "metallurgy_titanium",
		WorkerDomain: "metallurgy", WorkerCapacity: 11,
		EpochKey: "digital_era", OutputResource: "titanium",
	})
	// tier 10 — information_age  output=titanium  rate=1.0
	b = append(b, BuildingDef{
		Name: "Aerospace Foundry", Key: "aerospace_foundry", Category: "production",
		BaseCost:  map[string]float64{"electricity": 400 * Billion, "data": 42 * Billion, "steel": 370 * Billion, "titanium_ore": 370 * Billion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "titanium", Value: 1},
			{Type: "production", Target: "steel", Value: 141000000},
		},
		BuildTicks:  10920,
		RequiredAge: "information_age",
		Description: "Precision aerospace-grade titanium and steel foundry.",
		LineageKey:  "metallurgy", LineageTier: 10, LineKey: "metallurgy_titanium",
		WorkerDomain: "metallurgy", WorkerCapacity: 12,
		EpochKey: "digital_era", OutputResource: "titanium",
	})
	// tier 11 — digital_age  output=titanium  rate=2.0
	b = append(b, BuildingDef{
		Name: "Nano Alloy Plant", Key: "nano_alloy_plant", Category: "production",
		BaseCost:  map[string]float64{"electricity": 1.9 * Trillion, "data": 240 * Billion, "steel": 1.4 * Trillion, "titanium_ore": 1.4 * Trillion, "nanobots": 17000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "titanium", Value: 2},
			{Type: "production", Target: "steel", Value: 202000000},
		},
		BuildTicks:  12480,
		RequiredAge: "digital_age",
		Description: "Nano-scale titanium and steel alloy production.",
		LineageKey:  "metallurgy", LineageTier: 11, LineKey: "metallurgy_titanium",
		WorkerDomain: "metallurgy", WorkerCapacity: 13,
		EpochKey: "digital_era", OutputResource: "titanium",
	})
	// tier 12 — cyberpunk_age  output=dark_matter  rate=1.0 (new material)
	// Resource pivot: titanium → dark_matter. Validator shows 0.50x (titanium:2.0 → dark_matter:1.0) — intentional.
	// dark_matter is a tier-7 resource; starting at 1.0/tick is correct.
	b = append(b, BuildingDef{
		Name: "Dark Matter Refinery", Key: "dark_matter_refinery", Category: "production",
		BaseCost:  map[string]float64{"data": 930 * Billion, "crypto": 4.9 * Trillion, "electricity": 4.85 * Trillion, "dark_matter_crystals": 4.85 * Trillion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "dark_matter", Value: 1},
			{Type: "production", Target: "steel", Value: 493000000},
		},
		BuildTicks:  14040,
		RequiredAge: "cyberpunk_age",
		Description: "Distills dark matter and steel from the void.",
		LineageKey:  "metallurgy", LineageTier: 12, LineKey: "metallurgy_dark_matter",
		WorkerDomain: "metallurgy", WorkerCapacity: 14,
		EpochKey: "neon_era", OutputResource: "dark_matter",
	})
	// tier 13 — fusion_age  output=dark_matter  rate=2.0
	b = append(b, BuildingDef{
		Name: "Exotic Matter Forge", Key: "exotic_matter_forge", Category: "production",
		BaseCost:  map[string]float64{"plasma": 20 * Trillion, "electricity": 32 * Trillion, "dark_matter_crystals": 32 * Trillion, "steel": 85 * Trillion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "dark_matter", Value: 2},
			{Type: "production", Target: "steel", Value: 1.93 * Billion},
		},
		BuildTicks:  15600,
		RequiredAge: "fusion_age",
		Description: "Plasma-forged exotic matter and steel manufacturing.",
		LineageKey:  "metallurgy", LineageTier: 13, LineKey: "metallurgy_dark_matter",
		WorkerDomain: "metallurgy", WorkerCapacity: 16,
		EpochKey: "neon_era", OutputResource: "dark_matter",
	})
	// tier 14 — space_age  output=titanium (rate set by the Payback Rule)
	// note: was dark_matter at 4/tick, a resource that only unlocks in the
	// Interstellar Age. Titanium unlocks in the Space Age and nearly every
	// Space building costs it, yet nothing there made it. The refinery doesn't
	// cost titanium either (its 88T went to plasma), so it can start the
	// supply, like the Bronze Age smithy and iron. With no market to buy from,
	// it makes the age's steel as well.
	b = append(b, BuildingDef{
		Name: "Orbital Refinery", Key: "orbital_refinery", Category: "production",
		BaseCost:  map[string]float64{"plasma": 470 * Trillion, "electricity": 235 * Trillion, "dark_matter_crystals": 235 * Trillion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "titanium", Value: 10.8 * Billion},
			{Type: "production", Target: "steel", Value: 2.3 * Billion},
		},
		BuildTicks:  17160,
		RequiredAge: "space_age",
		Description: "Zero-gravity orbital titanium and steel refinery.",
		LineageKey:  "metallurgy", LineageTier: 14, LineKey: "metallurgy_titanium",
		WorkerDomain: "metallurgy", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "titanium",
	})
	// tier 15 — interstellar_age  output=antimatter  rate=2.0 (new material)
	// Resource pivot: titanium → antimatter. Validator shows 0.50x (titanium:4.0 → antimatter:2.0) — intentional.
	// antimatter is a tier-8 resource; starting at 2.0/tick matches cosmic-era rarity.
	b = append(b, BuildingDef{
		Name: "Antimatter Forge", Key: "antimatter_forge", Category: "production",
		BaseCost:  map[string]float64{"dark_matter": 440 * Trillion, "titanium": 3.5 * Quadrillion, "plasma": 2.2 * Quadrillion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "antimatter", Value: 2},
			{Type: "production", Target: "titanium", Value: 132 * Billion},
		},
		BuildTicks:  18720,
		RequiredAge: "interstellar_age",
		Description: "Forges antimatter and titanium in magnetic containment.",
		LineageKey:  "metallurgy", LineageTier: 15, LineKey: "metallurgy_antimatter",
		WorkerDomain: "metallurgy", WorkerCapacity: 20,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 16 — galactic_age  output=antimatter  rate=4.0
	b = append(b, BuildingDef{
		Name: "Stellar Metallurgy", Key: "stellar_metallurgy", Category: "production",
		BaseCost:  map[string]float64{"dark_matter": 6.63 * Quadrillion, "titanium": 33.1 * Quadrillion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "antimatter", Value: 5.94 * Billion},
			{Type: "production", Target: "titanium", Value: 242 * Billion},
		},
		BuildTicks:  18720,
		RequiredAge: "galactic_age",
		Description: "Stellar-scale antimatter and titanium metallurgy.",
		LineageKey:  "metallurgy", LineageTier: 16, LineKey: "metallurgy_antimatter",
		WorkerDomain: "metallurgy", WorkerCapacity: 22,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 17 — quantum_age  output=quantum_flux  rate=4.0
	// Resource pivot: antimatter → quantum_flux. Validator shows 1.00x (antimatter:4.0 → quantum_flux:4.0) — intentional.
	// quantum_flux is the terminal resource; matching antimatter rate is the intended cap.
	b = append(b, BuildingDef{
		Name: "Quantum Metal Works", Key: "quantum_metal_works", Category: "production",
		BaseCost:     map[string]float64{"quantum_flux": 950 * Trillion, "antimatter": 280 * Quadrillion, "dark_matter": 240 * Quadrillion},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "quantum_flux", Value: 64.9 * Billion}},
		BuildTicks:   18720,
		RequiredAge:  "quantum_age",
		RequiredTech: "quantum_metallurgy",
		Description:  "Quantum-state metalworking across dimensions.",
		LineageKey:   "metallurgy", LineageTier: 17, LineKey: "metallurgy_quantum_flux",
		WorkerDomain: "metallurgy", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	return b
}
