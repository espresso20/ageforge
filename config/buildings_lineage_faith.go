package config

// buildingsLineageFaith returns lineages 5-9:
// knowledge, faith, military, trade, engineering.
// Merged into newProductionBuildings() via init — see buildings_new_merge.go.
func buildingsLineageFaith() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 6 — FAITH (lineageKey: "faith", domain: "faith", output: "faith")
	// rate = 0.002 * 2^tier  CostScale: 1.30  Category: "research"
	// =========================================================================

	// tier 0 — primitive_age  rate=0.002
	b = append(b, BuildingDef{
		Name: "Shrine", Key: "shrine", Category: "research",
		BaseCost:    map[string]float64{"wood": 20},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.002}},
		BuildTicks:  80,
		RequiredAge: "primitive_age",
		Description: "A small spirit shrine. +0.002 faith/tick (2 workers).",
		LineageKey:  "faith", LineageTier: 0,
		WorkerDomain: "faith", WorkerCapacity: 2,
		EpochKey: "stone_era", OutputResource: "faith",
	})
	// tier 1 — stone_age  rate=0.004
	b = append(b, BuildingDef{
		Name: "Standing Stones", Key: "standing_stones", Category: "research",
		BaseCost:    map[string]float64{"wood": 100, "stone": 80},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.004}},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "Monolithic stones with ritual significance. +0.004 faith/tick (2 workers).",
		LineageKey:  "faith", LineageTier: 1,
		WorkerDomain: "faith", WorkerCapacity: 2,
		EpochKey: "stone_era", OutputResource: "faith",
	})
	// tier 2 — bronze_age  rate=0.008
	b = append(b, BuildingDef{
		Name: "Altar", Key: "altar", Category: "research",
		BaseCost:    map[string]float64{"wood": 700, "stone": 350, "gold": 150},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.008}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "A sacred altar for offerings. +0.008 faith/tick (3 workers).",
		LineageKey:  "faith", LineageTier: 2,
		WorkerDomain: "faith", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "faith",
	})
	// tier 3 — iron_age  rate=0.016
	b = append(b, BuildingDef{
		Name: "Temple", Key: "temple", Category: "research",
		BaseCost:    map[string]float64{"stone": 5000, "gold": 2000, "iron": 1000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.016}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "A formal temple for organised worship. +0.016 faith/tick (3 workers).",
		LineageKey:  "faith", LineageTier: 3,
		WorkerDomain: "faith", WorkerCapacity: 3,
		EpochKey: "iron_era", OutputResource: "faith",
	})
	// tier 4 — classical_age  rate=0.032
	b = append(b, BuildingDef{
		Name: "Oracle House", Key: "oracle_house", Category: "research",
		BaseCost:    map[string]float64{"stone": 30000, "gold": 10000, "iron": 5000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.032}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Oracles speak for the gods. +0.032 faith/tick (4 workers).",
		LineageKey:  "faith", LineageTier: 4,
		WorkerDomain: "faith", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "faith",
	})
	// tier 5 — medieval_age  rate=0.064
	b = append(b, BuildingDef{
		Name: "Cathedral", Key: "cathedral", Category: "research",
		BaseCost:    map[string]float64{"stone": 200000, "gold": 65000, "knowledge": 15000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.064}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "A towering medieval cathedral. +0.064 faith/tick (5 workers).",
		LineageKey:  "faith", LineageTier: 5,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "iron_era", OutputResource: "faith",
	})
	// tier 6 — renaissance_age  rate=0.128
	b = append(b, BuildingDef{
		Name: "Basilica", Key: "basilica", Category: "research",
		BaseCost:    map[string]float64{"gold": 650000, "stone": 400000, "knowledge": 100000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.128}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "A grand renaissance basilica. +0.128 faith/tick (5 workers).",
		LineageKey:  "faith", LineageTier: 6,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "faith",
	})
	// tier 7 — colonial_age  rate=0.256
	b = append(b, BuildingDef{
		Name: "Mission", Key: "mission", Category: "research",
		BaseCost:    map[string]float64{"gold": 3e6, "stone": 2e6, "knowledge": 400000},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.256}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "A colonial mission spreading faith. +0.256 faith/tick (5 workers).",
		LineageKey:  "faith", LineageTier: 7,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "faith",
	})
	// tier 8 — industrial_age  rate=0.512
	b = append(b, BuildingDef{
		Name: "Church", Key: "church", Category: "research",
		BaseCost:    map[string]float64{"stone": 20e6, "gold": 10e6, "iron": 8e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 0.512}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "An industrial-age parish church. +0.512 faith/tick (5 workers).",
		LineageKey:  "faith", LineageTier: 8,
		WorkerDomain: "faith", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "faith",
	})
	// tier 9 — victorian_age  rate=1.024
	b = append(b, BuildingDef{
		Name: "Grand Cathedral", Key: "grand_cathedral", Category: "research",
		BaseCost:    map[string]float64{"steel": 170e6, "gold": 90e6, "stone": 120e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 1.024}},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "A vast Victorian grand cathedral. +1.024 faith/tick (6 workers).",
		LineageKey:  "faith", LineageTier: 9,
		WorkerDomain: "faith", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "faith",
	})
	// tier 10 — electric_age  rate=2.048
	b = append(b, BuildingDef{
		Name: "Revival Hall", Key: "revival_hall", Category: "research",
		BaseCost:    map[string]float64{"steel": 1e9, "electricity": 350e6, "gold": 500e6},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 2.048}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "Electric revival meetings spread spiritual fervour. +2.048 faith/tick (6 workers).",
		LineageKey:  "faith", LineageTier: 10,
		WorkerDomain: "faith", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "faith",
	})
	// tier 11 — atomic_age  rate=4.096
	b = append(b, BuildingDef{
		Name: "Spiritual Center", Key: "spiritual_center", Category: "research",
		BaseCost:    map[string]float64{"steel": 5e9, "electricity": 2e9, "gold": 3e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 4.096}},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "Atomic-age spiritual wellness center. +4.096 faith/tick (6 workers).",
		LineageKey:  "faith", LineageTier: 11,
		WorkerDomain: "faith", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "faith",
	})
	// tier 12 — modern_age  rate=8.192
	b = append(b, BuildingDef{
		Name: "Meditation Center", Key: "meditation_center", Category: "research",
		BaseCost:    map[string]float64{"steel": 28e9, "electricity": 10e9, "gold": 20e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 8.192}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "Modern meditation and mindfulness hub. +8.192 faith/tick (7 workers).",
		LineageKey:  "faith", LineageTier: 12,
		WorkerDomain: "faith", WorkerCapacity: 7,
		EpochKey: "digital_era", OutputResource: "faith",
	})
	// tier 13 — information_age  rate=16.384
	b = append(b, BuildingDef{
		Name: "Digital Temple", Key: "digital_temple", Category: "research",
		BaseCost:    map[string]float64{"electricity": 80e9, "data": 7e9, "gold": 130e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 16.384}},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "A virtual spiritual sanctuary. +16.384 faith/tick (7 workers).",
		LineageKey:  "faith", LineageTier: 13,
		WorkerDomain: "faith", WorkerCapacity: 7,
		EpochKey: "digital_era", OutputResource: "faith",
	})
	// tier 14 — digital_age  rate=32.768
	b = append(b, BuildingDef{
		Name: "Cyber Shrine", Key: "cyber_shrine", Category: "research",
		BaseCost:    map[string]float64{"electricity": 380e9, "data": 45e9, "crypto": 200e9},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 32.768}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "A cybernetic devotional shrine. +32.768 faith/tick (8 workers).",
		LineageKey:  "faith", LineageTier: 14,
		WorkerDomain: "faith", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "faith",
	})
	// tier 15 — cyberpunk_age  rate=65.536
	b = append(b, BuildingDef{
		Name: "Neon Sanctuary", Key: "neon_sanctuary", Category: "research",
		BaseCost:    map[string]float64{"data": 170e9, "crypto": 900e9, "electricity": 1.8e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 65.536}},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "Neon-lit cyberpunk sanctuary. +65.536 faith/tick (8 workers).",
		LineageKey:  "faith", LineageTier: 15,
		WorkerDomain: "faith", WorkerCapacity: 8,
		EpochKey: "neon_era", OutputResource: "faith",
	})
	// tier 16 — fusion_age  rate=131.072
	b = append(b, BuildingDef{
		Name: "Quantum Chapel", Key: "quantum_chapel", Category: "research",
		BaseCost:    map[string]float64{"plasma": 4e12, "electricity": 13e12, "steel": 17e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 131.072}},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "A chapel resonating with quantum energies. +131.072 faith/tick (9 workers).",
		LineageKey:  "faith", LineageTier: 16,
		WorkerDomain: "faith", WorkerCapacity: 9,
		EpochKey: "neon_era", OutputResource: "faith",
	})
	// tier 17 — space_age  rate=262.144
	b = append(b, BuildingDef{
		Name: "Orbital Sanctuary", Key: "orbital_sanctuary", Category: "research",
		BaseCost:    map[string]float64{"titanium": 70e12, "plasma": 30e12, "electricity": 90e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 262.144}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "A faith sanctuary in orbital space. +262.144 faith/tick (9 workers).",
		LineageKey:  "faith", LineageTier: 17,
		WorkerDomain: "faith", WorkerCapacity: 9,
		EpochKey: "neon_era", OutputResource: "faith",
	})
	// tier 18 — interstellar_age  rate=524.288
	b = append(b, BuildingDef{
		Name: "Void Monastery", Key: "void_monastery", Category: "research",
		BaseCost:    map[string]float64{"dark_matter": 90e12, "titanium": 720e12, "plasma": 440e12},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 524.288}},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "A monastery floating in the interstellar void. +524.288 faith/tick (10 workers).",
		LineageKey:  "faith", LineageTier: 18,
		WorkerDomain: "faith", WorkerCapacity: 10,
		EpochKey: "cosmic_era", OutputResource: "faith",
	})
	// tier 19 — galactic_age  rate=1048.576
	b = append(b, BuildingDef{
		Name: "Stellar Shrine", Key: "stellar_shrine", Category: "research",
		BaseCost:    map[string]float64{"antimatter": 180e12, "dark_matter": 900e12, "titanium": 4.5e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 1048.576}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "A galactic-scale stellar shrine. +1048.576 faith/tick (10 workers).",
		LineageKey:  "faith", LineageTier: 19,
		WorkerDomain: "faith", WorkerCapacity: 10,
		EpochKey: "cosmic_era", OutputResource: "faith",
	})
	// tier 20 — quantum_age  rate=2097.152
	b = append(b, BuildingDef{
		Name: "Transcendence Hall", Key: "transcendence_hall", Category: "research",
		BaseCost:    map[string]float64{"quantum_flux": 190e12, "antimatter": 58e15, "dark_matter": 47e15},
		CostScale:   1.30,
		Effects:     []Effect{{Type: "production", Target: "faith", Value: 2097.152}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "A hall dedicated to transcendence beyond existence. +2097.152 faith/tick (12 workers).",
		LineageKey:  "faith", LineageTier: 20,
		WorkerDomain: "faith", WorkerCapacity: 12,
		EpochKey: "cosmic_era", OutputResource: "faith",
	})

	return b
}
