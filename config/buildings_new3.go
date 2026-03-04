package config

// newProductionBuildings3 returns lineages 10-13:
// culture_arts, metallurgy, energy, hacker.
func newProductionBuildings3() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 10 — CULTURE/ARTS (lineageKey: "culture_arts", no domain, no workers)
	// starts at classical_age (tier 0)
	// Two effects per building: production(culture) + storage(culture cap bonus)
	// CostScale: 1.30  Category: "production"
	// =========================================================================

	// tier 0 — classical_age  rate=0.5  cap=+500
	b = append(b, BuildingDef{
		Name: "Amphitheater", Key: "amphitheater", Category: "production",
		BaseCost:    map[string]float64{"stone": 40000, "gold": 12000, "wood": 15000},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 0.5},
			{Type: "storage", Target: "culture", Value: 500},
		},
		BuildTicks:     3000,
		RequiredAge:    "classical_age",
		Description:    "Open-air theatre and culture hub. +0.5 culture/tick, +500 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 0,
		EpochKey: "iron_era", OutputResource: "culture",
	})
	// tier 1 — medieval_age  rate=1.0  cap=+1000
	b = append(b, BuildingDef{
		Name: "Great Hall", Key: "great_hall", Category: "production",
		BaseCost:    map[string]float64{"stone": 210000, "gold": 70000, "iron": 30000},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 1.0},
			{Type: "storage", Target: "culture", Value: 1000},
		},
		BuildTicks:     6000,
		RequiredAge:    "medieval_age",
		Description:    "A lord's great hall for feasts and culture. +1.0 culture/tick, +1000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 1,
		EpochKey: "iron_era", OutputResource: "culture",
	})
	// tier 2 — renaissance_age  rate=2.0  cap=+2500
	b = append(b, BuildingDef{
		Name: "Art Studio", Key: "art_studio", Category: "production",
		BaseCost:    map[string]float64{"gold": 650000, "stone": 300000, "knowledge": 80000},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 2.0},
			{Type: "storage", Target: "culture", Value: 2500},
		},
		BuildTicks:     12000,
		RequiredAge:    "renaissance_age",
		Description:    "Painters and sculptors create cultural works. +2.0 culture/tick, +2500 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 2,
		EpochKey: "steel_era", OutputResource: "culture",
	})
	// tier 3 — colonial_age  rate=4.0  cap=+5000
	b = append(b, BuildingDef{
		Name: "Concert Hall", Key: "concert_hall", Category: "production",
		BaseCost:    map[string]float64{"gold": 4e6, "stone": 2.5e6, "steel": 800000},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 4.0},
			{Type: "storage", Target: "culture", Value: 5000},
		},
		BuildTicks:     18000,
		RequiredAge:    "colonial_age",
		Description:    "Classical music and colonial culture. +4.0 culture/tick, +5000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 3,
		EpochKey: "steel_era", OutputResource: "culture",
	})
	// tier 4 — industrial_age  rate=8.0  cap=+10000
	b = append(b, BuildingDef{
		Name: "Opera House", Key: "opera_house", Category: "production",
		BaseCost:    map[string]float64{"steel": 28e6, "gold": 18e6, "stone": 12e6},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 8.0},
			{Type: "storage", Target: "culture", Value: 10000},
		},
		BuildTicks:     25000,
		RequiredAge:    "industrial_age",
		Description:    "Grandest venue for opera and orchestral culture. +8.0 culture/tick, +10000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 4,
		EpochKey: "steel_era", OutputResource: "culture",
	})
	// tier 5 — victorian_age  rate=15  cap=+25000
	b = append(b, BuildingDef{
		Name: "Grand Museum", Key: "grand_museum", Category: "production",
		BaseCost:    map[string]float64{"steel": 200e6, "gold": 120e6, "stone": 80e6},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 15},
			{Type: "storage", Target: "culture", Value: 25000},
		},
		BuildTicks:     50000,
		RequiredAge:    "victorian_age",
		Description:    "A grand Victorian museum of arts and history. +15 culture/tick, +25000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 5,
		EpochKey: "electric_era", OutputResource: "culture",
	})
	// tier 6 — electric_age  rate=30  cap=+50000
	b = append(b, BuildingDef{
		Name: "Radio Station", Key: "radio_station", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.1e9, "electricity": 450e6, "gold": 700e6},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 30},
			{Type: "storage", Target: "culture", Value: 50000},
		},
		BuildTicks:     75000,
		RequiredAge:    "electric_age",
		Description:    "Broadcasts culture to the masses. +30 culture/tick, +50000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 6,
		EpochKey: "electric_era", OutputResource: "culture",
	})
	// tier 7 — atomic_age  rate=60  cap=+100000
	b = append(b, BuildingDef{
		Name: "Cinema", Key: "cinema", Category: "production",
		BaseCost:    map[string]float64{"steel": 5.5e9, "electricity": 2.2e9, "gold": 3.5e9},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 60},
			{Type: "storage", Target: "culture", Value: 100000},
		},
		BuildTicks:     100000,
		RequiredAge:    "atomic_age",
		Description:    "Film and cinema spread cultural influence. +60 culture/tick, +100000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 7,
		EpochKey: "electric_era", OutputResource: "culture",
	})
	// tier 8 — modern_age  rate=120  cap=+250000
	b = append(b, BuildingDef{
		Name: "TV Studio", Key: "tv_studio", Category: "production",
		BaseCost:    map[string]float64{"steel": 34e9, "electricity": 13e9, "data": 1.2e9},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 120},
			{Type: "storage", Target: "culture", Value: 250000},
		},
		BuildTicks:     150000,
		RequiredAge:    "modern_age",
		Description:    "Television studio broadcasting global culture. +120 culture/tick, +250000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 8,
		EpochKey: "digital_era", OutputResource: "culture",
	})
	// tier 9 — information_age  rate=250  cap=+500000
	b = append(b, BuildingDef{
		Name: "Media Center", Key: "media_center", Category: "production",
		BaseCost:    map[string]float64{"electricity": 90e9, "data": 9e9, "gold": 160e9},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 250},
			{Type: "storage", Target: "culture", Value: 500000},
		},
		BuildTicks:     300000,
		RequiredAge:    "information_age",
		Description:    "Digital media center for global cultural content. +250 culture/tick, +500000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 9,
		EpochKey: "digital_era", OutputResource: "culture",
	})
	// tier 10 — digital_age  rate=500  cap=+1000000
	b = append(b, BuildingDef{
		Name: "VR Studio", Key: "vr_studio", Category: "production",
		BaseCost:    map[string]float64{"electricity": 440e9, "data": 54e9, "steel": 640e9},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 500},
			{Type: "storage", Target: "culture", Value: 1000000},
		},
		BuildTicks:     500000,
		RequiredAge:    "digital_age",
		Description:    "Virtual reality cultural experience studio. +500 culture/tick, +1000000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 10,
		EpochKey: "digital_era", OutputResource: "culture",
	})
	// tier 11 — cyberpunk_age  rate=1000  cap=+2500000
	b = append(b, BuildingDef{
		Name: "Holographic Theater", Key: "holographic_theater", Category: "production",
		BaseCost:    map[string]float64{"data": 215e9, "crypto": 1.1e12, "electricity": 2.2e12},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 1000},
			{Type: "storage", Target: "culture", Value: 2500000},
		},
		BuildTicks:     1000000,
		RequiredAge:    "cyberpunk_age",
		Description:    "Full-immersion holographic cultural performances. +1000 culture/tick, +2500000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 11,
		EpochKey: "neon_era", OutputResource: "culture",
	})
	// tier 12 — fusion_age  rate=2000  cap=+5000000
	b = append(b, BuildingDef{
		Name: "Neural Art Complex", Key: "neural_art_complex", Category: "production",
		BaseCost:    map[string]float64{"plasma": 4.5e12, "electricity": 14e12, "steel": 19e12},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 2000},
			{Type: "storage", Target: "culture", Value: 5000000},
		},
		BuildTicks:     1500000,
		RequiredAge:    "fusion_age",
		Description:    "Neural-linked art creation at fusion scale. +2000 culture/tick, +5000000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 12,
		EpochKey: "neon_era", OutputResource: "culture",
	})
	// tier 13 — space_age  rate=4000  cap=+10000000
	b = append(b, BuildingDef{
		Name: "Zero G Gallery", Key: "zero_g_gallery", Category: "production",
		BaseCost:    map[string]float64{"titanium": 82e12, "plasma": 40e12, "electricity": 98e12},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 4000},
			{Type: "storage", Target: "culture", Value: 10000000},
		},
		BuildTicks:     2000000,
		RequiredAge:    "space_age",
		Description:    "Zero-gravity orbital art gallery. +4000 culture/tick, +10000000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 13,
		EpochKey: "neon_era", OutputResource: "culture",
	})
	// tier 14 — interstellar_age  rate=8000  cap=+25000000
	b = append(b, BuildingDef{
		Name: "Cultural Beacon", Key: "cultural_beacon", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 95e12, "titanium": 740e12, "plasma": 450e12},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 8000},
			{Type: "storage", Target: "culture", Value: 25000000},
		},
		BuildTicks:     2500000,
		RequiredAge:    "interstellar_age",
		Description:    "A beacon broadcasting culture across star systems. +8000 culture/tick, +25000000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 14,
		EpochKey: "cosmic_era", OutputResource: "culture",
	})
	// tier 15 — galactic_age  rate=16000  cap=+50000000
	b = append(b, BuildingDef{
		Name: "Civilization Archive", Key: "civilization_archive", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 185e12, "dark_matter": 920e12, "titanium": 4.6e15},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 16000},
			{Type: "storage", Target: "culture", Value: 50000000},
		},
		BuildTicks:     3000000,
		RequiredAge:    "galactic_age",
		Description:    "Archives of all civilisations across the galaxy. +16000 culture/tick, +50000000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 15,
		EpochKey: "cosmic_era", OutputResource: "culture",
	})
	// tier 16 — quantum_age  rate=32000  cap=+100000000
	b = append(b, BuildingDef{
		Name: "Reality Art Engine", Key: "reality_art_engine", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 205e12, "antimatter": 61e15, "dark_matter": 51e15},
		CostScale:   1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 32000},
			{Type: "storage", Target: "culture", Value: 100000000},
		},
		BuildTicks:     5000000,
		RequiredAge:    "quantum_age",
		Description:    "Reshapes reality as a medium for art. +32000 culture/tick, +100000000 culture cap.",
		LineageKey:     "culture_arts", LineageTier: 16,
		EpochKey: "cosmic_era", OutputResource: "culture",
	})

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
		BuildTicks:  1000,
		RequiredAge: "iron_age",
		Description: "Smelts iron ore into usable iron. +0.10 iron/tick (4 workers).",
		LineageKey: "metallurgy", LineageTier: 0,
		WorkerDomain: "metallurgy", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 1 — classical_age  output=iron  rate=0.20
	b = append(b, BuildingDef{
		Name: "Forge", Key: "forge", Category: "production",
		BaseCost:    map[string]float64{"stone": 36000, "gold": 12000, "coal": 5000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.20}},
		BuildTicks:  3000,
		RequiredAge: "classical_age",
		Description: "A proper forge for working iron. +0.20 iron/tick (4 workers).",
		LineageKey: "metallurgy", LineageTier: 1,
		WorkerDomain: "metallurgy", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 2 — medieval_age  output=iron  rate=0.40
	b = append(b, BuildingDef{
		Name: "Ironmonger", Key: "ironmonger", Category: "production",
		BaseCost:    map[string]float64{"stone": 200000, "gold": 65000, "iron": 25000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "iron", Value: 0.40}},
		BuildTicks:  6000,
		RequiredAge: "medieval_age",
		Description: "Specialist iron trade and metalworking. +0.40 iron/tick (5 workers).",
		LineageKey: "metallurgy", LineageTier: 2,
		WorkerDomain: "metallurgy", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "iron",
	})
	// tier 3 — renaissance_age  output=steel  rate=0.50
	b = append(b, BuildingDef{
		Name: "Foundry", Key: "foundry", Category: "production",
		BaseCost:    map[string]float64{"gold": 650000, "steel": 220000, "coal": 80000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 0.50}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "Produces steel by alloying iron and carbon. +0.50 steel/tick (5 workers).",
		LineageKey: "metallurgy", LineageTier: 3,
		WorkerDomain: "metallurgy", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 4 — colonial_age  output=steel  rate=1.0
	b = append(b, BuildingDef{
		Name: "Iron Works", Key: "iron_works", Category: "production",
		BaseCost:    map[string]float64{"gold": 3.8e6, "steel": 2e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 1.0}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "Colonial iron works processing steel. +1.0 steel/tick (6 workers).",
		LineageKey: "metallurgy", LineageTier: 4,
		WorkerDomain: "metallurgy", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 5 — industrial_age  output=steel  rate=2.0
	b = append(b, BuildingDef{
		Name: "Steel Mill", Key: "steel_mill", Category: "production",
		BaseCost:    map[string]float64{"steel": 28e6, "coal": 12e6, "gold": 16e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 2.0}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Industrial-scale steel production. +2.0 steel/tick (7 workers).",
		LineageKey: "metallurgy", LineageTier: 5,
		WorkerDomain: "metallurgy", WorkerCapacity: 7,
		EpochKey: "steel_era", OutputResource: "steel",
	})
	// tier 6 — victorian_age  output=steel  rate=4.0
	b = append(b, BuildingDef{
		Name: "Bessemer Plant", Key: "bessemer_plant", Category: "production",
		BaseCost:    map[string]float64{"steel": 195e6, "coal": 100e6, "gold": 120e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 4.0}},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "Bessemer converter for mass steel production. +4.0 steel/tick (8 workers).",
		LineageKey: "metallurgy", LineageTier: 6,
		WorkerDomain: "metallurgy", WorkerCapacity: 8,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 7 — electric_age  output=steel  rate=8.0
	b = append(b, BuildingDef{
		Name: "Electric Arc Furnace", Key: "electric_arc_furnace", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.2e9, "electricity": 500e6, "gold": 700e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 8.0}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "Electric arc furnace for high-grade steel. +8.0 steel/tick (9 workers).",
		LineageKey: "metallurgy", LineageTier: 7,
		WorkerDomain: "metallurgy", WorkerCapacity: 9,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 8 — atomic_age  output=steel  rate=16.0
	b = append(b, BuildingDef{
		Name: "Advanced Alloy Plant", Key: "advanced_alloy_plant", Category: "production",
		BaseCost:    map[string]float64{"steel": 6e9, "electricity": 2.5e9, "uranium": 400e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "steel", Value: 16.0}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Atomic-era advanced alloy manufacturing. +16.0 steel/tick (10 workers).",
		LineageKey: "metallurgy", LineageTier: 8,
		WorkerDomain: "metallurgy", WorkerCapacity: 10,
		EpochKey: "electric_era", OutputResource: "steel",
	})
	// tier 9 — modern_age  output=titanium  rate=0.50 (reset for new metal)
	b = append(b, BuildingDef{
		Name: "Titanium Smelter", Key: "titanium_smelter", Category: "production",
		BaseCost:    map[string]float64{"steel": 34e9, "electricity": 13e9, "data": 1.2e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "titanium", Value: 0.50}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Smelts titanium ore into refined titanium. +0.50 titanium/tick (11 workers).",
		LineageKey: "metallurgy", LineageTier: 9,
		WorkerDomain: "metallurgy", WorkerCapacity: 11,
		EpochKey: "digital_era", OutputResource: "titanium",
	})
	// tier 10 — information_age  output=titanium  rate=1.0
	b = append(b, BuildingDef{
		Name: "Aerospace Foundry", Key: "aerospace_foundry", Category: "production",
		BaseCost:    map[string]float64{"electricity": 95e9, "data": 10e9, "steel": 175e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "titanium", Value: 1.0}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "Precision aerospace-grade titanium foundry. +1.0 titanium/tick (12 workers).",
		LineageKey: "metallurgy", LineageTier: 10,
		WorkerDomain: "metallurgy", WorkerCapacity: 12,
		EpochKey: "digital_era", OutputResource: "titanium",
	})
	// tier 11 — digital_age  output=titanium  rate=2.0
	b = append(b, BuildingDef{
		Name: "Nano Alloy Plant", Key: "nano_alloy_plant", Category: "production",
		BaseCost:    map[string]float64{"electricity": 460e9, "data": 57e9, "steel": 670e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "titanium", Value: 2.0}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Nano-scale titanium alloy production. +2.0 titanium/tick (13 workers).",
		LineageKey: "metallurgy", LineageTier: 11,
		WorkerDomain: "metallurgy", WorkerCapacity: 13,
		EpochKey: "digital_era", OutputResource: "titanium",
	})
	// tier 12 — cyberpunk_age  output=dark_matter  rate=1.0 (new material)
	b = append(b, BuildingDef{
		Name: "Dark Matter Refinery", Key: "dark_matter_refinery", Category: "production",
		BaseCost:    map[string]float64{"data": 220e9, "crypto": 1.15e12, "electricity": 2.3e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 1.0}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Refines dark matter crystals into usable dark matter. +1.0 dark_matter/tick (14 workers).",
		LineageKey: "metallurgy", LineageTier: 12,
		WorkerDomain: "metallurgy", WorkerCapacity: 14,
		EpochKey: "neon_era", OutputResource: "dark_matter",
	})
	// tier 13 — fusion_age  output=dark_matter  rate=2.0
	b = append(b, BuildingDef{
		Name: "Exotic Matter Forge", Key: "exotic_matter_forge", Category: "production",
		BaseCost:    map[string]float64{"plasma": 4.8e12, "electricity": 15e12, "steel": 20e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 2.0}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Plasma-forged exotic matter manufacturing. +2.0 dark_matter/tick (16 workers).",
		LineageKey: "metallurgy", LineageTier: 13,
		WorkerDomain: "metallurgy", WorkerCapacity: 16,
		EpochKey: "neon_era", OutputResource: "dark_matter",
	})
	// tier 14 — space_age  output=dark_matter  rate=4.0
	b = append(b, BuildingDef{
		Name: "Orbital Refinery", Key: "orbital_refinery", Category: "production",
		BaseCost:    map[string]float64{"titanium": 88e12, "plasma": 44e12, "electricity": 110e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 4.0}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Zero-gravity orbital dark matter refinery. +4.0 dark_matter/tick (18 workers).",
		LineageKey: "metallurgy", LineageTier: 14,
		WorkerDomain: "metallurgy", WorkerCapacity: 18,
		EpochKey: "neon_era", OutputResource: "dark_matter",
	})
	// tier 15 — interstellar_age  output=antimatter  rate=2.0 (new material)
	b = append(b, BuildingDef{
		Name: "Antimatter Forge", Key: "antimatter_forge", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 105e12, "titanium": 820e12, "plasma": 510e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "antimatter", Value: 2.0}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Forges antimatter from dark matter reactions. +2.0 antimatter/tick (20 workers).",
		LineageKey: "metallurgy", LineageTier: 15,
		WorkerDomain: "metallurgy", WorkerCapacity: 20,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 16 — galactic_age  output=antimatter  rate=4.0
	b = append(b, BuildingDef{
		Name: "Stellar Metallurgy", Key: "stellar_metallurgy", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 210e12, "dark_matter": 1.05e15, "titanium": 5.2e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "antimatter", Value: 4.0}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Stellar-scale antimatter metallurgy. +4.0 antimatter/tick (22 workers).",
		LineageKey: "metallurgy", LineageTier: 16,
		WorkerDomain: "metallurgy", WorkerCapacity: 22,
		EpochKey: "cosmic_era", OutputResource: "antimatter",
	})
	// tier 17 — quantum_age  output=quantum_flux  rate=4.0
	b = append(b, BuildingDef{
		Name: "Quantum Metal Works", Key: "quantum_metal_works", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 225e12, "antimatter": 67e15, "dark_matter": 57e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 4.0}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Quantum-state metalworking across dimensions. +4.0 quantum_flux/tick (25 workers).",
		LineageKey: "metallurgy", LineageTier: 17,
		WorkerDomain: "metallurgy", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	// =========================================================================
	// LINEAGE 12 — ENERGY (lineageKey: "energy", domain: "energy")
	// starts at industrial_age (tier 0)
	// CostScale: 1.35  Category: "production"
	// Output: coal/electricity transitions → plasma → dark_matter → quantum_flux
	// =========================================================================

	// tier 0 — industrial_age  output=coal  rate=10
	b = append(b, BuildingDef{
		Name: "Coal Plant", Key: "coal_plant", Category: "production",
		BaseCost:    map[string]float64{"steel": 22e6, "coal": 8e6, "gold": 12e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "coal", Value: 10}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Industrial coal processing plant. +10 coal/tick (6 workers).",
		LineageKey: "energy", LineageTier: 0,
		WorkerDomain: "energy", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 1 — victorian_age  output=electricity  rate=50  (+ some coal)
	b = append(b, BuildingDef{
		Name: "Steam Turbine", Key: "steam_turbine", Category: "production",
		BaseCost:    map[string]float64{"steel": 185e6, "coal": 90e6, "gold": 110e6},
		CostScale:   1.35,
		Effects: []Effect{
			{Type: "production", Target: "electricity", Value: 50},
			{Type: "production", Target: "coal", Value: 5},
		},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "Steam turbine generating electricity from coal. +50 electricity, +5 coal/tick (7 workers).",
		LineageKey: "energy", LineageTier: 1,
		WorkerDomain: "energy", WorkerCapacity: 7,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 2 — electric_age  output=electricity  rate=100
	b = append(b, BuildingDef{
		Name: "Power Generator", Key: "power_generator", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.1e9, "electricity": 450e6, "coal": 300e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 100}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "Electric power generator. +100 electricity/tick (8 workers).",
		LineageKey: "energy", LineageTier: 2,
		WorkerDomain: "energy", WorkerCapacity: 8,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 3 — atomic_age  output=electricity  rate=200
	b = append(b, BuildingDef{
		Name: "Nuclear Reactor", Key: "nuclear_reactor", Category: "production",
		BaseCost:    map[string]float64{"steel": 5.8e9, "electricity": 2.4e9, "uranium": 600e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 200}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Nuclear fission reactor. +200 electricity/tick (9 workers).",
		LineageKey: "energy", LineageTier: 3,
		WorkerDomain: "energy", WorkerCapacity: 9,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 4 — modern_age  output=oil + electricity  rate: oil=20 electricity=100
	b = append(b, BuildingDef{
		Name: "Oil Refinery", Key: "oil_refinery", Category: "production",
		BaseCost:    map[string]float64{"steel": 33e9, "electricity": 12e9, "oil": 2e9},
		CostScale:   1.35,
		Effects: []Effect{
			{Type: "production", Target: "oil", Value: 20},
			{Type: "production", Target: "electricity", Value: 100},
		},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Modern oil refinery and power generation. +20 oil, +100 electricity/tick (10 workers).",
		LineageKey: "energy", LineageTier: 4,
		WorkerDomain: "energy", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "oil",
	})
	// tier 5 — information_age  output=electricity  rate=400
	b = append(b, BuildingDef{
		Name: "Smart Energy Grid", Key: "smart_energy_grid", Category: "production",
		BaseCost:    map[string]float64{"electricity": 105e9, "data": 11e9, "steel": 190e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 400}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "AI-optimised smart energy grid. +400 electricity/tick (11 workers).",
		LineageKey: "energy", LineageTier: 5,
		WorkerDomain: "energy", WorkerCapacity: 11,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 6 — digital_age  output=electricity  rate=800
	b = append(b, BuildingDef{
		Name: "Quantum Battery Array", Key: "quantum_battery_array", Category: "production",
		BaseCost:    map[string]float64{"electricity": 490e9, "data": 62e9, "steel": 720e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 800}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Quantum-state battery arrays. +800 electricity/tick (12 workers).",
		LineageKey: "energy", LineageTier: 6,
		WorkerDomain: "energy", WorkerCapacity: 12,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 7 — cyberpunk_age  output=electricity  rate=1600
	b = append(b, BuildingDef{
		Name: "Dark Energy Tap", Key: "dark_energy_tap", Category: "production",
		BaseCost:    map[string]float64{"data": 240e9, "crypto": 1.25e12, "electricity": 2.5e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 1600}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Taps dark energy streams for electricity. +1600 electricity/tick (13 workers).",
		LineageKey: "energy", LineageTier: 7,
		WorkerDomain: "energy", WorkerCapacity: 13,
		EpochKey: "neon_era", OutputResource: "electricity",
	})
	// tier 8 — fusion_age  output=plasma  rate=20
	b = append(b, BuildingDef{
		Name: "Fusion Reactor Array", Key: "fusion_reactor_array", Category: "production",
		BaseCost:    map[string]float64{"plasma": 5.2e12, "electricity": 16e12, "steel": 21e12},
		CostScale:   1.35,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 20},
			{Type: "production", Target: "electricity", Value: 2000},
		},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Array of fusion reactors producing plasma. +20 plasma, +2000 electricity/tick (15 workers).",
		LineageKey: "energy", LineageTier: 8,
		WorkerDomain: "energy", WorkerCapacity: 15,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 9 — space_age  output=plasma  rate=40
	b = append(b, BuildingDef{
		Name: "Solar Collector Array", Key: "solar_collector_array", Category: "production",
		BaseCost:    map[string]float64{"titanium": 90e12, "plasma": 45e12, "electricity": 112e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "plasma", Value: 40}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Orbital solar collectors feeding plasma energy. +40 plasma/tick (16 workers).",
		LineageKey: "energy", LineageTier: 9,
		WorkerDomain: "energy", WorkerCapacity: 16,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 10 — interstellar_age  output=plasma  rate=80
	b = append(b, BuildingDef{
		Name: "Pulsar Tap", Key: "pulsar_tap", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 110e12, "titanium": 840e12, "plasma": 520e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "plasma", Value: 80}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Taps pulsar radiation for plasma energy. +80 plasma/tick (18 workers).",
		LineageKey: "energy", LineageTier: 10,
		WorkerDomain: "energy", WorkerCapacity: 18,
		EpochKey: "cosmic_era", OutputResource: "plasma",
	})
	// tier 11 — galactic_age  output=dark_matter  rate=10
	b = append(b, BuildingDef{
		Name: "Quasar Tap", Key: "quasar_tap", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 220e12, "dark_matter": 1.1e15, "titanium": 5.5e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 10}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Taps quasar jets for dark matter. +10 dark_matter/tick (20 workers).",
		LineageKey: "energy", LineageTier: 11,
		WorkerDomain: "energy", WorkerCapacity: 20,
		EpochKey: "cosmic_era", OutputResource: "dark_matter",
	})
	// tier 12 — quantum_age  output=quantum_flux  rate=50
	b = append(b, BuildingDef{
		Name: "Zero Point Generator", Key: "zero_point_generator", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 230e12, "antimatter": 68e15, "dark_matter": 58e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 50}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Generates energy from quantum zero-point fields. +50 quantum_flux/tick (25 workers).",
		LineageKey: "energy", LineageTier: 12,
		WorkerDomain: "energy", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	// =========================================================================
	// LINEAGE 13 — HACKER/DIGITAL (lineageKey: "hacker", domain: "hacker", output: "data")
	// starts at information_age (tier 0)
	// rate = 2.0 * 2^tier  CostScale: 1.35  Category: "production"
	// =========================================================================

	// tier 0 — information_age  rate=2.0
	b = append(b, BuildingDef{
		Name: "Server Farm", Key: "server_farm", Category: "production",
		BaseCost:    map[string]float64{"electricity": 95e9, "data": 10e9, "steel": 180e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 2.0}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "Large server farm processing data. +2.0 data/tick (8 workers).",
		LineageKey: "hacker", LineageTier: 0,
		WorkerDomain: "hacker", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "data",
	})
	// tier 1 — digital_age  rate=4.0
	b = append(b, BuildingDef{
		Name: "Data Center", Key: "data_center", Category: "production",
		BaseCost:    map[string]float64{"electricity": 470e9, "data": 58e9, "steel": 685e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 4.0}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Hyper-scale data center. +4.0 data/tick (10 workers).",
		LineageKey: "hacker", LineageTier: 1,
		WorkerDomain: "hacker", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "data",
	})
	// tier 2 — cyberpunk_age  rate=8.0
	b = append(b, BuildingDef{
		Name: "Cyber Hub", Key: "cyber_hub", Category: "production",
		BaseCost:    map[string]float64{"data": 225e9, "crypto": 1.18e12, "electricity": 2.35e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 8.0}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Cyberpunk underground hacker hub. +8.0 data/tick (12 workers).",
		LineageKey: "hacker", LineageTier: 2,
		WorkerDomain: "hacker", WorkerCapacity: 12,
		EpochKey: "neon_era", OutputResource: "data",
	})
	// tier 3 — fusion_age  rate=16.0
	b = append(b, BuildingDef{
		Name: "Quantum Server Farm", Key: "quantum_server_farm", Category: "production",
		BaseCost:    map[string]float64{"plasma": 5.1e12, "electricity": 15e12, "steel": 20e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 16.0}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Quantum-computing server farm. +16.0 data/tick (14 workers).",
		LineageKey: "hacker", LineageTier: 3,
		WorkerDomain: "hacker", WorkerCapacity: 14,
		EpochKey: "neon_era", OutputResource: "data",
	})
	// tier 4 — space_age  rate=32.0
	b = append(b, BuildingDef{
		Name: "Orbital Data Relay", Key: "orbital_data_relay", Category: "production",
		BaseCost:    map[string]float64{"titanium": 86e12, "plasma": 42e12, "electricity": 106e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 32.0}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Orbital relay node for stellar data networks. +32.0 data/tick (16 workers).",
		LineageKey: "hacker", LineageTier: 4,
		WorkerDomain: "hacker", WorkerCapacity: 16,
		EpochKey: "neon_era", OutputResource: "data",
	})
	// tier 5 — interstellar_age  rate=64.0
	b = append(b, BuildingDef{
		Name: "Galactic Network Node", Key: "galactic_network_node", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 107e12, "titanium": 830e12, "plasma": 515e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 64.0}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Interstellar galactic network node. +64.0 data/tick (18 workers).",
		LineageKey: "hacker", LineageTier: 5,
		WorkerDomain: "hacker", WorkerCapacity: 18,
		EpochKey: "cosmic_era", OutputResource: "data",
	})
	// tier 6 — galactic_age  rate=128.0
	b = append(b, BuildingDef{
		Name: "Consciousness Upload Hub", Key: "consciousness_upload_hub", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 215e12, "dark_matter": 1.08e15, "titanium": 5.4e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 128.0}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Uploads and processes consciousness data. +128.0 data/tick (20 workers).",
		LineageKey: "hacker", LineageTier: 6,
		WorkerDomain: "hacker", WorkerCapacity: 20,
		EpochKey: "cosmic_era", OutputResource: "data",
	})
	// tier 7 — quantum_age  rate=256.0
	b = append(b, BuildingDef{
		Name: "Reality Processor", Key: "reality_processor", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 235e12, "antimatter": 70e15, "dark_matter": 60e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 256.0}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Processes data from the very structure of reality. +256.0 data/tick (25 workers).",
		LineageKey: "hacker", LineageTier: 7,
		WorkerDomain: "hacker", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "data",
	})

	return b
}
