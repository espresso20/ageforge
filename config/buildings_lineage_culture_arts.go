package config

// buildingsLineageCultureArts returns lineages 10-13:
// culture_arts, metallurgy, energy, hacker.
func buildingsLineageCultureArts() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 10 — CULTURE/ARTS (lineageKey: "culture_arts", no domain, no workers)
	// starts at classical_age (tier 0)
	// Two effects per building: production(culture) + storage(culture cap bonus)
	// CostScale: 1.30  Category: "production"
	// =========================================================================

	// Stage 2A: entertainment venues also lift morale, on the same gentle
	// ~0.0005·1.15^tier ramp as the faith worship buildings. Modest by design.

	// tier 0 — classical_age  rate=0.5  cap=+500
	b = append(b, BuildingDef{
		Name: "Amphitheater", Key: "amphitheater", Category: "production",
		BaseCost:  map[string]float64{"stone": 120000, "gold": 36000, "wood": 45000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 0.5},
			{Type: "storage", Target: "culture", Value: 500},
			{Type: "morale", Value: 0.0005},
		},
		BuildTicks:  2730,
		RequiredAge: "classical_age",
		Description: "Open-air theater and culture hub.",
		LineageKey:  "culture_arts", LineageTier: 0,
		EpochKey: "iron_era", OutputResource: "culture",
	})
	// tier 1 — medieval_age  rate=1.0  cap=+1000
	b = append(b, BuildingDef{
		Name: "Great Hall", Key: "great_hall", Category: "production",
		BaseCost:  map[string]float64{"stone": 630000, "gold": 210000, "iron": 90000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 1},
			{Type: "storage", Target: "culture", Value: 1000},
			{Type: "morale", Value: 0.0006},
		},
		BuildTicks:  3510,
		RequiredAge: "medieval_age",
		Description: "A lord's great hall for feasts and culture.",
		LineageKey:  "culture_arts", LineageTier: 1,
		EpochKey: "iron_era", OutputResource: "culture",
	})
	// tier 2 — renaissance_age  rate=2.0  cap=+2500
	b = append(b, BuildingDef{
		Name: "Art Studio", Key: "art_studio", Category: "production",
		BaseCost:  map[string]float64{"gold": 2000000, "stone": 900000, "knowledge": 240000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 2},
			{Type: "storage", Target: "culture", Value: 2500},
			{Type: "morale", Value: 0.0007},
		},
		BuildTicks:  4680,
		RequiredAge: "renaissance_age",
		Description: "Painters and sculptors create cultural works.",
		LineageKey:  "culture_arts", LineageTier: 2,
		EpochKey: "steel_era", OutputResource: "culture",
	})
	// tier 3 — colonial_age  rate=4.0  cap=+5000
	b = append(b, BuildingDef{
		Name: "Concert Hall", Key: "concert_hall", Category: "production",
		BaseCost:  map[string]float64{"gold": 1.2e07, "stone": 7500000, "steel": 2400000},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 4},
			{Type: "storage", Target: "culture", Value: 5000},
			{Type: "morale", Value: 0.0008},
		},
		BuildTicks:  5460,
		RequiredAge: "colonial_age",
		Description: "Classical music and colonial culture.",
		LineageKey:  "culture_arts", LineageTier: 3,
		EpochKey: "steel_era", OutputResource: "culture",
	})
	// tier 4 — industrial_age  rate=8.0  cap=+10000
	b = append(b, BuildingDef{
		Name: "Opera House", Key: "opera_house", Category: "production",
		BaseCost:  map[string]float64{"steel": 8.4e07, "gold": 5.4e07, "stone": 3.6e07},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 8},
			{Type: "storage", Target: "culture", Value: 10000},
			{Type: "morale", Value: 0.0009},
		},
		BuildTicks:  6240,
		RequiredAge: "industrial_age",
		Description: "Grandest venue for opera and orchestral culture.",
		LineageKey:  "culture_arts", LineageTier: 4,
		EpochKey: "steel_era", OutputResource: "culture",
	})
	// tier 5 — victorian_age  rate=15  cap=+25000
	b = append(b, BuildingDef{
		Name: "Grand Museum", Key: "grand_museum", Category: "production",
		BaseCost:  map[string]float64{"steel": 6e08, "gold": 3.6e08, "stone": 2.4e08},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 15},
			{Type: "storage", Target: "culture", Value: 25000},
			{Type: "morale", Value: 0.001},
		},
		BuildTicks:  7020,
		RequiredAge: "victorian_age",
		Description: "A grand Victorian museum of arts and history.",
		LineageKey:  "culture_arts", LineageTier: 5,
		EpochKey: "electric_era", OutputResource: "culture",
	})
	// tier 6 — electric_age  rate=30  cap=+50000
	b = append(b, BuildingDef{
		Name: "Radio Station", Key: "radio_station", Category: "production",
		BaseCost:  map[string]float64{"steel": 3.3e09, "electricity": 1.4e09, "gold": 2.1e09},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 30},
			{Type: "storage", Target: "culture", Value: 50000},
			{Type: "morale", Value: 0.0012},
		},
		BuildTicks:  7800,
		RequiredAge: "electric_age",
		Description: "Broadcasts culture to the masses.",
		LineageKey:  "culture_arts", LineageTier: 6,
		EpochKey: "electric_era", OutputResource: "culture",
	})
	// tier 7 — atomic_age  rate=60  cap=+100000
	b = append(b, BuildingDef{
		Name: "Cinema", Key: "cinema", Category: "production",
		BaseCost:  map[string]float64{"steel": 1.7e10, "electricity": 6.6e09, "gold": 1.1e10},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 60},
			{Type: "storage", Target: "culture", Value: 100000},
			{Type: "morale", Value: 0.0013},
		},
		BuildTicks:  9360,
		RequiredAge: "atomic_age",
		Description: "Film and cinema spread cultural influence.",
		LineageKey:  "culture_arts", LineageTier: 7,
		EpochKey: "electric_era", OutputResource: "culture",
	})
	// tier 8 — modern_age  rate=120  cap=+250000
	b = append(b, BuildingDef{
		Name: "TV Studio", Key: "tv_studio", Category: "production",
		BaseCost:  map[string]float64{"steel": 1e11, "electricity": 3.9e10, "data": 3.6e09},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 120},
			{Type: "storage", Target: "culture", Value: 250000},
			{Type: "morale", Value: 0.0015},
		},
		BuildTicks:  9360,
		RequiredAge: "modern_age",
		Description: "Television studio broadcasting global culture.",
		LineageKey:  "culture_arts", LineageTier: 8,
		EpochKey: "digital_era", OutputResource: "culture",
	})
	// tier 9 — information_age  rate=250  cap=+500000
	b = append(b, BuildingDef{
		Name: "Media Center", Key: "media_center", Category: "production",
		BaseCost:  map[string]float64{"electricity": 2.7e11, "data": 2.7e10, "gold": 4.8e11},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 250},
			{Type: "storage", Target: "culture", Value: 500000},
			{Type: "morale", Value: 0.0018},
		},
		BuildTicks:  10920,
		RequiredAge: "information_age",
		Description: "Digital media center for global cultural content.",
		LineageKey:  "culture_arts", LineageTier: 9,
		EpochKey: "digital_era", OutputResource: "culture",
	})
	// tier 10 — digital_age  rate=500  cap=+1000000
	b = append(b, BuildingDef{
		Name: "VR Studio", Key: "vr_studio", Category: "production",
		BaseCost:  map[string]float64{"electricity": 1.3e12, "data": 1.6e11, "steel": 1.9e12},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 500},
			{Type: "storage", Target: "culture", Value: 1000000},
			{Type: "morale", Value: 0.002},
		},
		BuildTicks:  12480,
		RequiredAge: "digital_age",
		Description: "Virtual reality cultural experience studio.",
		LineageKey:  "culture_arts", LineageTier: 10,
		EpochKey: "digital_era", OutputResource: "culture",
	})
	// tier 11 — cyberpunk_age  rate=1000  cap=+2500000
	b = append(b, BuildingDef{
		Name: "Holographic Theater", Key: "holographic_theater", Category: "production",
		BaseCost:  map[string]float64{"data": 6.5e11, "crypto": 3.3e12, "electricity": 6.6e12},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 1000},
			{Type: "storage", Target: "culture", Value: 2500000},
			{Type: "morale", Value: 0.0023},
		},
		BuildTicks:   14040,
		RequiredAge:  "cyberpunk_age",
		RequiredTech: "holography",
		Description:  "Full-immersion holographic cultural performances.",
		LineageKey:   "culture_arts", LineageTier: 11,
		EpochKey: "neon_era", OutputResource: "culture",
	})
	// tier 12 — fusion_age  rate=2000  cap=+5000000
	b = append(b, BuildingDef{
		Name: "Neural Art Complex", Key: "neural_art_complex", Category: "production",
		BaseCost:  map[string]float64{"plasma": 1.4e13, "electricity": 4.2e13, "steel": 5.7e13},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 2000},
			{Type: "storage", Target: "culture", Value: 5000000},
			{Type: "morale", Value: 0.0027},
		},
		BuildTicks:  15600,
		RequiredAge: "fusion_age",
		Description: "Neural-linked art creation at fusion scale.",
		LineageKey:  "culture_arts", LineageTier: 12,
		EpochKey: "neon_era", OutputResource: "culture",
	})
	// tier 13 — space_age  rate=4000  cap=+10000000
	b = append(b, BuildingDef{
		Name: "Zero G Gallery", Key: "zero_g_gallery", Category: "production",
		BaseCost:  map[string]float64{"titanium": 2.5e14, "plasma": 1.2e14, "electricity": 3e14},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 4000},
			{Type: "storage", Target: "culture", Value: 1e07},
			{Type: "morale", Value: 0.0031},
		},
		BuildTicks:  17160,
		RequiredAge: "space_age",
		Description: "Zero-gravity orbital art gallery.",
		LineageKey:  "culture_arts", LineageTier: 13,
		EpochKey: "neon_era", OutputResource: "culture",
	})
	// tier 14 — interstellar_age  rate=8000  cap=+25000000
	b = append(b, BuildingDef{
		Name: "Cultural Beacon", Key: "cultural_beacon", Category: "production",
		BaseCost:  map[string]float64{"dark_matter": 2.9e14, "titanium": 2.2e15, "plasma": 1.4e15},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 8000},
			{Type: "storage", Target: "culture", Value: 2.5e07},
			{Type: "morale", Value: 0.0035},
		},
		BuildTicks:  18720,
		RequiredAge: "interstellar_age",
		Description: "A beacon broadcasting culture across star systems.",
		LineageKey:  "culture_arts", LineageTier: 14,
		EpochKey: "cosmic_era", OutputResource: "culture",
	})
	// tier 15 — galactic_age  rate=16000  cap=+50000000
	b = append(b, BuildingDef{
		Name: "Civilization Archive", Key: "civilization_archive", Category: "production",
		BaseCost:  map[string]float64{"antimatter": 5.6e14, "dark_matter": 2.8e15, "titanium": 1.4e16},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 16000},
			{Type: "storage", Target: "culture", Value: 5e07},
			{Type: "morale", Value: 0.0041},
		},
		BuildTicks:  18720,
		RequiredAge: "galactic_age",
		Description: "Archives of every civilization across the galaxy.",
		LineageKey:  "culture_arts", LineageTier: 15,
		EpochKey: "cosmic_era", OutputResource: "culture",
	})
	// tier 16 — quantum_age  rate=32000  cap=+100000000
	b = append(b, BuildingDef{
		Name: "Reality Art Engine", Key: "reality_art_engine", Category: "production",
		BaseCost:  map[string]float64{"quantum_flux": 6.2e14, "antimatter": 1.8e17, "dark_matter": 1.5e17},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 32000},
			{Type: "storage", Target: "culture", Value: 1e08},
			{Type: "morale", Value: 0.0047},
		},
		BuildTicks:  18720,
		RequiredAge: "quantum_age",
		Description: "Reshapes reality as a medium for art.",
		LineageKey:  "culture_arts", LineageTier: 16,
		EpochKey: "cosmic_era", OutputResource: "culture",
	})

	return b
}
