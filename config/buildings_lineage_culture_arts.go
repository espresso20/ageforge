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
		BaseCost:  map[string]float64{"stone": 40000, "gold": 12000, "wood": 15000},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 0.5},
			{Type: "storage", Target: "culture", Value: 500},
			{Type: "morale", Value: 0.0005},
		},
		BuildTicks:  3000,
		RequiredAge: "classical_age",
		Description: "Open-air theatre and culture hub. +0.5 culture/tick, +500 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 0,
		EpochKey: "iron_era", OutputResource: "culture",
	})
	// tier 1 — medieval_age  rate=1.0  cap=+1000
	b = append(b, BuildingDef{
		Name: "Great Hall", Key: "great_hall", Category: "production",
		BaseCost:  map[string]float64{"stone": 210000, "gold": 70000, "iron": 30000},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 1.0},
			{Type: "storage", Target: "culture", Value: 1000},
			{Type: "morale", Value: 0.0006},
		},
		BuildTicks:  6000,
		RequiredAge: "medieval_age",
		Description: "A lord's great hall for feasts and culture. +1.0 culture/tick, +1000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 1,
		EpochKey: "iron_era", OutputResource: "culture",
	})
	// tier 2 — renaissance_age  rate=2.0  cap=+2500
	b = append(b, BuildingDef{
		Name: "Art Studio", Key: "art_studio", Category: "production",
		BaseCost:  map[string]float64{"gold": 650000, "stone": 300000, "knowledge": 80000},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 2.0},
			{Type: "storage", Target: "culture", Value: 2500},
			{Type: "morale", Value: 0.0007},
		},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "Painters and sculptors create cultural works. +2.0 culture/tick, +2500 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 2,
		EpochKey: "steel_era", OutputResource: "culture",
	})
	// tier 3 — colonial_age  rate=4.0  cap=+5000
	b = append(b, BuildingDef{
		Name: "Concert Hall", Key: "concert_hall", Category: "production",
		BaseCost:  map[string]float64{"gold": 4e6, "stone": 2.5e6, "steel": 800000},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 4.0},
			{Type: "storage", Target: "culture", Value: 5000},
			{Type: "morale", Value: 0.0008},
		},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "Classical music and colonial culture. +4.0 culture/tick, +5000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 3,
		EpochKey: "steel_era", OutputResource: "culture",
	})
	// tier 4 — industrial_age  rate=8.0  cap=+10000
	b = append(b, BuildingDef{
		Name: "Opera House", Key: "opera_house", Category: "production",
		BaseCost:  map[string]float64{"steel": 28e6, "gold": 18e6, "stone": 12e6},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 8.0},
			{Type: "storage", Target: "culture", Value: 10000},
			{Type: "morale", Value: 0.0009},
		},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Grandest venue for opera and orchestral culture. +8.0 culture/tick, +10000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 4,
		EpochKey: "steel_era", OutputResource: "culture",
	})
	// tier 5 — victorian_age  rate=15  cap=+25000
	b = append(b, BuildingDef{
		Name: "Grand Museum", Key: "grand_museum", Category: "production",
		BaseCost:  map[string]float64{"steel": 200e6, "gold": 120e6, "stone": 80e6},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 15},
			{Type: "storage", Target: "culture", Value: 25000},
			{Type: "morale", Value: 0.0010},
		},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "A grand Victorian museum of arts and history. +15 culture/tick, +25000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 5,
		EpochKey: "electric_era", OutputResource: "culture",
	})
	// tier 6 — electric_age  rate=30  cap=+50000
	b = append(b, BuildingDef{
		Name: "Radio Station", Key: "radio_station", Category: "production",
		BaseCost:  map[string]float64{"steel": 1.1e9, "electricity": 450e6, "gold": 700e6},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 30},
			{Type: "storage", Target: "culture", Value: 50000},
			{Type: "morale", Value: 0.0012},
		},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "Broadcasts culture to the masses. +30 culture/tick, +50000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 6,
		EpochKey: "electric_era", OutputResource: "culture",
	})
	// tier 7 — atomic_age  rate=60  cap=+100000
	b = append(b, BuildingDef{
		Name: "Cinema", Key: "cinema", Category: "production",
		BaseCost:  map[string]float64{"steel": 5.5e9, "electricity": 2.2e9, "gold": 3.5e9},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 60},
			{Type: "storage", Target: "culture", Value: 100000},
			{Type: "morale", Value: 0.0013},
		},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Film and cinema spread cultural influence. +60 culture/tick, +100000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 7,
		EpochKey: "electric_era", OutputResource: "culture",
	})
	// tier 8 — modern_age  rate=120  cap=+250000
	b = append(b, BuildingDef{
		Name: "TV Studio", Key: "tv_studio", Category: "production",
		BaseCost:  map[string]float64{"steel": 34e9, "electricity": 13e9, "data": 1.2e9},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 120},
			{Type: "storage", Target: "culture", Value: 250000},
			{Type: "morale", Value: 0.0015},
		},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Television studio broadcasting global culture. +120 culture/tick, +250000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 8,
		EpochKey: "digital_era", OutputResource: "culture",
	})
	// tier 9 — information_age  rate=250  cap=+500000
	b = append(b, BuildingDef{
		Name: "Media Center", Key: "media_center", Category: "production",
		BaseCost:  map[string]float64{"electricity": 90e9, "data": 9e9, "gold": 160e9},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 250},
			{Type: "storage", Target: "culture", Value: 500000},
			{Type: "morale", Value: 0.0018},
		},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "Digital media center for global cultural content. +250 culture/tick, +500000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 9,
		EpochKey: "digital_era", OutputResource: "culture",
	})
	// tier 10 — digital_age  rate=500  cap=+1000000
	b = append(b, BuildingDef{
		Name: "VR Studio", Key: "vr_studio", Category: "production",
		BaseCost:  map[string]float64{"electricity": 440e9, "data": 54e9, "steel": 640e9},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 500},
			{Type: "storage", Target: "culture", Value: 1000000},
			{Type: "morale", Value: 0.0020},
		},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Virtual reality cultural experience studio. +500 culture/tick, +1000000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 10,
		EpochKey: "digital_era", OutputResource: "culture",
	})
	// tier 11 — cyberpunk_age  rate=1000  cap=+2500000
	b = append(b, BuildingDef{
		Name: "Holographic Theater", Key: "holographic_theater", Category: "production",
		BaseCost:  map[string]float64{"data": 215e9, "crypto": 1.1e12, "electricity": 2.2e12},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 1000},
			{Type: "storage", Target: "culture", Value: 2500000},
			{Type: "morale", Value: 0.0023},
		},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Full-immersion holographic cultural performances. +1000 culture/tick, +2500000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 11,
		EpochKey: "neon_era", OutputResource: "culture",
	})
	// tier 12 — fusion_age  rate=2000  cap=+5000000
	b = append(b, BuildingDef{
		Name: "Neural Art Complex", Key: "neural_art_complex", Category: "production",
		BaseCost:  map[string]float64{"plasma": 4.5e12, "electricity": 14e12, "steel": 19e12},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 2000},
			{Type: "storage", Target: "culture", Value: 5000000},
			{Type: "morale", Value: 0.0027},
		},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Neural-linked art creation at fusion scale. +2000 culture/tick, +5000000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 12,
		EpochKey: "neon_era", OutputResource: "culture",
	})
	// tier 13 — space_age  rate=4000  cap=+10000000
	b = append(b, BuildingDef{
		Name: "Zero G Gallery", Key: "zero_g_gallery", Category: "production",
		BaseCost:  map[string]float64{"titanium": 82e12, "plasma": 40e12, "electricity": 98e12},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 4000},
			{Type: "storage", Target: "culture", Value: 10000000},
			{Type: "morale", Value: 0.0031},
		},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Zero-gravity orbital art gallery. +4000 culture/tick, +10000000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 13,
		EpochKey: "neon_era", OutputResource: "culture",
	})
	// tier 14 — interstellar_age  rate=8000  cap=+25000000
	b = append(b, BuildingDef{
		Name: "Cultural Beacon", Key: "cultural_beacon", Category: "production",
		BaseCost:  map[string]float64{"dark_matter": 95e12, "titanium": 740e12, "plasma": 450e12},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 8000},
			{Type: "storage", Target: "culture", Value: 25000000},
			{Type: "morale", Value: 0.0035},
		},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "A beacon broadcasting culture across star systems. +8000 culture/tick, +25000000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 14,
		EpochKey: "cosmic_era", OutputResource: "culture",
	})
	// tier 15 — galactic_age  rate=16000  cap=+50000000
	b = append(b, BuildingDef{
		Name: "Civilization Archive", Key: "civilization_archive", Category: "production",
		BaseCost:  map[string]float64{"antimatter": 185e12, "dark_matter": 920e12, "titanium": 4.6e15},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 16000},
			{Type: "storage", Target: "culture", Value: 50000000},
			{Type: "morale", Value: 0.0041},
		},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Archives of all civilisations across the galaxy. +16000 culture/tick, +50000000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 15,
		EpochKey: "cosmic_era", OutputResource: "culture",
	})
	// tier 16 — quantum_age  rate=32000  cap=+100000000
	b = append(b, BuildingDef{
		Name: "Reality Art Engine", Key: "reality_art_engine", Category: "production",
		BaseCost:  map[string]float64{"quantum_flux": 205e12, "antimatter": 61e15, "dark_matter": 51e15},
		CostScale: 1.30,
		Effects: []Effect{
			{Type: "production", Target: "culture", Value: 32000},
			{Type: "storage", Target: "culture", Value: 100000000},
			{Type: "morale", Value: 0.0047},
		},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Reshapes reality as a medium for art. +32000 culture/tick, +100000000 culture cap.",
		LineageKey:  "culture_arts", LineageTier: 16,
		EpochKey: "cosmic_era", OutputResource: "culture",
	})

	return b
}
