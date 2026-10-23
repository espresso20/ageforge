package config

// buildingsLineageFaith returns lineages 5-9:
// knowledge, faith, military, trade, engineering.
// Merged into newProductionBuildings() via init — see buildings_new_merge.go.
func buildingsLineageFaith() []BuildingDef {
	// Stage 2A: every worship building lifts morale. Value ramps gently by tier
	// (~0.0005·1.15^tier, hand-rounded) so later, costlier temples give a bit more
	// morale — but values stay modest: a few noticeably lift morale over a minute
	// or two, yet you must keep investing. Morale is background pressure, not a
	// resource to trivially max.
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 6 — FAITH (lineageKey: "faith", domain: "faith", output: "faith")
	// rate = 0.002 * 2^tier  CostScale: 1.30  Category: "research"
	// =========================================================================

	// tier 0 — primitive_age  rate=0.002
	b = append(b, BuildingDef{
		Name: "Shrine", Key: "shrine", Category: "research",
		BaseCost:  map[string]float64{"wood": 60},
		CostScale: 1.15,
		// Stage 1 proof-of-plumbing: small flat morale lift. Stage 2 will broaden
		// morale effects to the full era-appropriate set; keep this value modest.
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.002}, {Type: "morale", Value: 0.0006}},
		BuildTicks:  75,
		RequiredAge: "primitive_age",
		Description: "A small spirit shrine.",
		LineageKey:  "faith", LineageTier: 0,
		WorkerDomain: "faith", WorkerCapacity: 2,
		EpochKey: "stone_era", OutputResource: "faith",
	})
	// tier 1 — stone_age  rate=0.004
	b = append(b, BuildingDef{
		Name: "Standing Stones", Key: "standing_stones", Category: "research",
		BaseCost:     map[string]float64{"wood": 300, "stone": 240},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "faith", Value: 0.004}, {Type: "morale", Value: 0.0006}},
		BuildTicks:   200,
		RequiredAge:  "stone_age",
		RequiredTech: "ritual",
		Description:  "Monolithic stones with ritual significance.",
		LineageKey:   "faith", LineageTier: 1,
		WorkerDomain: "faith", WorkerCapacity: 2,
		EpochKey: "stone_era", OutputResource: "faith",
	})
	// tier 2 — bronze_age  rate=0.008
	b = append(b, BuildingDef{
		Name: "Altar", Key: "altar", Category: "research",
		BaseCost:     map[string]float64{"wood": 2100, "stone": 1100, "gold": 450},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "faith", Value: 0.008}, {Type: "morale", Value: 0.0007}},
		BuildTicks:   150,
		RequiredAge:  "bronze_age",
		RequiredTech: "calendar",
		Description:  "A sacred altar for offerings.",
		LineageKey:   "faith", LineageTier: 2,
		WorkerDomain: "faith", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "faith",
	})
	// tier 3 — iron_age  rate=0.016
	b = append(b, BuildingDef{
		Name: "Temple", Key: "temple", Category: "research",
		BaseCost:  map[string]float64{"stone": 7500, "marble": 7500, "gold": 6000, "iron": 1500, "iron_ore": 1500},
		CostScale: 1.15,
		// Stage 1 proof-of-plumbing: small flat morale lift (see shrine). Stage 2
		// broadens this to the full set; keep the value modest for now.
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.016}, {Type: "morale", Value: 0.0006}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "A formal temple for organized worship.",
		LineageKey:  "faith", LineageTier: 3,
		WorkerDomain: "faith", WorkerCapacity: 3,
		EpochKey: "iron_era", OutputResource: "faith",
	})
	// tier 4 — classical_age  rate=0.032
	b = append(b, BuildingDef{
		Name: "Oracle House", Key: "oracle_house", Category: "research",
		BaseCost:    map[string]float64{"stone": 45000, "marble": 45000, "gold": 30000, "iron": 7500, "iron_ore": 7500},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.032}, {Type: "morale", Value: 0.0009}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Oracles speak for the gods.",
		LineageKey:  "faith", LineageTier: 4,
		WorkerDomain: "faith", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "faith",
	})
	// tier 5 — medieval_age  rate=0.064
	b = append(b, BuildingDef{
		Name: "Cathedral", Key: "cathedral", Category: "research",
		BaseCost:     map[string]float64{"stone": 870000, "gold": 290000},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "faith", Value: 0.064}, {Type: "morale", Value: 0.001}},
		BuildTicks:   1200,
		RequiredAge:  "medieval_age",
		RequiredTech: "theology",
		Description:  "A towering medieval cathedral.",
		LineageKey:   "faith", LineageTier: 5,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "faith",
	})
	// tier 6 — renaissance_age  rate=0.128
	b = append(b, BuildingDef{
		Name: "Basilica", Key: "basilica", Category: "research",
		BaseCost:    map[string]float64{"gold": 3.02 * Million, "stone": 1.81 * Million},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.128}, {Type: "morale", Value: 0.0012}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "A grand Renaissance basilica.",
		LineageKey:  "faith", LineageTier: 6,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "faith",
	})
	// tier 7 — colonial_age  rate=0.256
	b = append(b, BuildingDef{
		Name: "Mission", Key: "mission", Category: "research",
		BaseCost:    map[string]float64{"gold": 14.9 * Million, "stone": 12.4 * Million},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.256}, {Type: "morale", Value: 0.0013}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "A colonial mission spreading faith.",
		LineageKey:  "faith", LineageTier: 7,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "faith",
	})
	// tier 8 — industrial_age  rate=0.512
	b = append(b, BuildingDef{
		Name: "Church", Key: "church", Category: "research",
		BaseCost:    map[string]float64{"stone": 60 * Million, "gold": 30 * Million, "iron": 12 * Million, "iron_ore": 12 * Million},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.512}, {Type: "morale", Value: 0.0015}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "An industrial-age parish church.",
		LineageKey:  "faith", LineageTier: 8,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "faith",
	})
	// tier 9 — victorian_age  rate=1.024
	b = append(b, BuildingDef{
		Name: "Grand Cathedral", Key: "grand_cathedral", Category: "research",
		BaseCost:    map[string]float64{"steel": 510 * Million, "gold": 270 * Million, "stone": 360 * Million},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 1.024}, {Type: "morale", Value: 0.0018}},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "A vast Victorian grand cathedral.",
		LineageKey:  "faith", LineageTier: 9,
		WorkerDomain: "faith", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "faith",
	})
	// tier 10 — electric_age  rate=2.048
	b = append(b, BuildingDef{
		Name: "Revival Hall", Key: "revival_hall", Category: "research",
		BaseCost:    map[string]float64{"steel": 3 * Billion, "electricity": 1.1 * Billion, "gold": 1.5 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 2.048}, {Type: "morale", Value: 0.002}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "Electric revival meetings spread spiritual fervor.",
		LineageKey:  "faith", LineageTier: 10,
		WorkerDomain: "faith", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "faith",
	})
	// tier 11 — atomic_age  rate=4.096
	b = append(b, BuildingDef{
		Name: "Spiritual Center", Key: "spiritual_center", Category: "research",
		BaseCost:    map[string]float64{"steel": 15 * Billion, "electricity": 6 * Billion, "gold": 9 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 4.096}, {Type: "morale", Value: 0.0023}},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "Atomic-age spiritual wellness center.",
		LineageKey:  "faith", LineageTier: 11,
		WorkerDomain: "faith", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "faith",
	})
	// tier 12 — modern_age  rate=8.192
	b = append(b, BuildingDef{
		Name: "Meditation Center", Key: "meditation_center", Category: "research",
		BaseCost:    map[string]float64{"steel": 42 * Billion, "titanium_ore": 42 * Billion, "electricity": 30 * Billion, "gold": 60 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 8.192}, {Type: "morale", Value: 0.0027}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "Modern meditation and mindfulness hub.",
		LineageKey:  "faith", LineageTier: 12,
		WorkerDomain: "faith", WorkerCapacity: 7,
		EpochKey: "digital_era", OutputResource: "faith",
	})
	// tier 13 — information_age  rate=16.384
	b = append(b, BuildingDef{
		Name: "Digital Temple", Key: "digital_temple", Category: "research",
		BaseCost:    map[string]float64{"electricity": 240 * Billion, "data": 21 * Billion, "gold": 390 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 16.384}, {Type: "morale", Value: 0.0031}},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "A virtual spiritual sanctuary.",
		LineageKey:  "faith", LineageTier: 13,
		WorkerDomain: "faith", WorkerCapacity: 7,
		EpochKey: "digital_era", OutputResource: "faith",
	})
	// tier 14 — digital_age  rate=32.768
	b = append(b, BuildingDef{
		Name: "Cyber Shrine", Key: "cyber_shrine", Category: "research",
		BaseCost:    map[string]float64{"electricity": 1.1 * Trillion, "data": 140 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 32.768}, {Type: "morale", Value: 0.0035}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "A cybernetic devotional shrine.",
		LineageKey:  "faith", LineageTier: 14,
		WorkerDomain: "faith", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "faith",
	})
	// tier 15 — cyberpunk_age  rate=65.536
	b = append(b, BuildingDef{
		Name: "Neon Sanctuary", Key: "neon_sanctuary", Category: "research",
		BaseCost:    map[string]float64{"data": 510 * Billion, "crypto": 2.7 * Trillion, "electricity": 2.7 * Trillion, "dark_matter_crystals": 2.7 * Trillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 65.536}, {Type: "morale", Value: 0.0041}},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "Neon-lit cyberpunk sanctuary.",
		LineageKey:  "faith", LineageTier: 15,
		WorkerDomain: "faith", WorkerCapacity: 8,
		EpochKey: "neon_era", OutputResource: "faith",
	})
	// tier 16 — fusion_age  rate=131.072
	b = append(b, BuildingDef{
		Name: "Quantum Chapel", Key: "quantum_chapel", Category: "research",
		BaseCost:    map[string]float64{"plasma": 12 * Trillion, "electricity": 19.5 * Trillion, "dark_matter_crystals": 19.5 * Trillion, "steel": 51 * Trillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 131.072}, {Type: "morale", Value: 0.0047}},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "A chapel resonating with quantum energies.",
		LineageKey:  "faith", LineageTier: 16,
		WorkerDomain: "faith", WorkerCapacity: 9,
		EpochKey: "neon_era", OutputResource: "faith",
	})
	// tier 17 — space_age  rate=262.144
	b = append(b, BuildingDef{
		Name: "Orbital Sanctuary", Key: "orbital_sanctuary", Category: "research",
		BaseCost:    map[string]float64{"titanium": 210 * Trillion, "plasma": 90 * Trillion, "electricity": 135 * Trillion, "dark_matter_crystals": 135 * Trillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 262.144}, {Type: "morale", Value: 0.0054}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "A faith sanctuary in orbital space.",
		LineageKey:  "faith", LineageTier: 17,
		WorkerDomain: "faith", WorkerCapacity: 9,
		EpochKey: "neon_era", OutputResource: "faith",
	})
	// tier 18 — interstellar_age  rate=524.288
	b = append(b, BuildingDef{
		Name: "Void Monastery", Key: "void_monastery", Category: "research",
		BaseCost:    map[string]float64{"dark_matter": 270 * Trillion, "titanium": 2.2 * Quadrillion, "plasma": 1.3 * Quadrillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 524.288}, {Type: "morale", Value: 0.0062}},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "A monastery floating in the interstellar void.",
		LineageKey:  "faith", LineageTier: 18,
		WorkerDomain: "faith", WorkerCapacity: 10,
		EpochKey: "cosmic_era", OutputResource: "faith",
	})
	// tier 19 — galactic_age  rate=1048.576
	b = append(b, BuildingDef{
		Name: "Stellar Shrine", Key: "stellar_shrine", Category: "research",
		BaseCost:    map[string]float64{"antimatter": 540 * Trillion, "dark_matter": 2.7 * Quadrillion, "titanium": 14 * Quadrillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 1048.576}, {Type: "morale", Value: 0.0071}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "A galactic-scale stellar shrine.",
		LineageKey:  "faith", LineageTier: 19,
		WorkerDomain: "faith", WorkerCapacity: 10,
		EpochKey: "cosmic_era", OutputResource: "faith",
	})
	// tier 20 — quantum_age  rate=2097.152
	b = append(b, BuildingDef{
		Name: "Transcendence Hall", Key: "transcendence_hall", Category: "research",
		BaseCost:    map[string]float64{"quantum_flux": 570 * Trillion, "antimatter": 170 * Quadrillion, "dark_matter": 140 * Quadrillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 2097.152}, {Type: "morale", Value: 0.0082}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "A hall dedicated to transcendence beyond existence.",
		LineageKey:  "faith", LineageTier: 20,
		WorkerDomain: "faith", WorkerCapacity: 12,
		EpochKey: "cosmic_era", OutputResource: "faith",
	})

	return b
}
