package config

// buildingsLineageMilitary returns lineages 5-9:
// knowledge, faith, military, trade, engineering.
// Merged into newProductionBuildings() via init — see buildings_new_merge.go.
func buildingsLineageMilitary() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 7 — MILITARY (lineageKey: "military", domain: "military")
	// Effect type: "capacity", Target: "military", Value: 10 * 2^tier
	// CostScale: 1.35  Category: "military"
	// =========================================================================

	// tier 0 — iron_age  soldiers=10
	b = append(b, BuildingDef{
		Name: "Hunting Lodge", Key: "hunting_lodge", Category: "military",
		BaseCost:    map[string]float64{"wood": 25},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 10}},
		BuildTicks:  80,
		RequiredAge: "iron_age",
		Description: "A gathering place for hunters. +10 military cap (3 workers).",
		LineageKey:  "military", LineageTier: 0,
		WorkerDomain: "military", WorkerCapacity: 3,
		EpochKey: "stone_era",
	})
	// tier 1 — stone_age  soldiers=20
	b = append(b, BuildingDef{
		Name: "War Camp", Key: "war_camp", Category: "military",
		BaseCost:    map[string]float64{"wood": 180, "stone": 100},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 20}},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "A fortified war camp. +20 military cap (4 workers).",
		LineageKey:  "military", LineageTier: 1,
		WorkerDomain: "military", WorkerCapacity: 4,
		EpochKey: "stone_era",
	})
	// tier 2 — bronze_age  soldiers=40
	b = append(b, BuildingDef{
		Name: "Barracks", Key: "barracks", Category: "military",
		BaseCost:    map[string]float64{"wood": 900, "stone": 600, "iron": 200},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 40}},
		BuildTicks:  500,
		RequiredAge: "bronze_age",
		Description: "Trains and houses soldiers. +40 military cap (5 workers).",
		LineageKey:  "military", LineageTier: 2,
		WorkerDomain: "military", WorkerCapacity: 5,
		EpochKey: "stone_era",
	})
	// tier 3 — iron_age  soldiers=80
	b = append(b, BuildingDef{
		Name: "Legion Fort", Key: "legion_fort", Category: "military",
		BaseCost:    map[string]float64{"stone": 7000, "iron": 3500, "gold": 2000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 80}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "A fortified roman-style legion camp. +80 military cap (6 workers).",
		LineageKey:  "military", LineageTier: 3,
		WorkerDomain: "military", WorkerCapacity: 6,
		EpochKey: "iron_era",
	})
	// tier 4 — classical_age  soldiers=160
	b = append(b, BuildingDef{
		Name: "Military Academy", Key: "military_academy", Category: "military",
		BaseCost:    map[string]float64{"stone": 40000, "gold": 15000, "iron": 10000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 160}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Trains elite military officers. +160 military cap (6 workers).",
		LineageKey:  "military", LineageTier: 4,
		WorkerDomain: "military", WorkerCapacity: 6,
		EpochKey: "iron_era",
	})
	// tier 5 — medieval_age  soldiers=320
	b = append(b, BuildingDef{
		Name: "Castle Keep", Key: "castle_keep", Category: "military",
		BaseCost:    map[string]float64{"stone": 220000, "iron": 70000, "gold": 50000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 320}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "A fortified stone keep. +320 military cap (7 workers).",
		LineageKey:  "military", LineageTier: 5,
		WorkerDomain: "military", WorkerCapacity: 7,
		EpochKey: "iron_era",
	})
	// tier 6 — renaissance_age  soldiers=640
	b = append(b, BuildingDef{
		Name: "Fortress", Key: "fortress", Category: "military",
		BaseCost:    map[string]float64{"stone": 700000, "gold": 300000, "steel": 150000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 640}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "A star-fort capable of holding a large garrison. +640 military cap (7 workers).",
		LineageKey:  "military", LineageTier: 6,
		WorkerDomain: "military", WorkerCapacity: 7,
		EpochKey: "steel_era",
	})
	// tier 7 — colonial_age  soldiers=1280
	b = append(b, BuildingDef{
		Name: "Fort", Key: "fort", Category: "military",
		BaseCost:    map[string]float64{"gold": 4e6, "steel": 2e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 1280}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "A colonial frontier fort. +1280 military cap (8 workers).",
		LineageKey:  "military", LineageTier: 7,
		WorkerDomain: "military", WorkerCapacity: 8,
		EpochKey: "steel_era",
	})
	// tier 8 — industrial_age  soldiers=2560
	b = append(b, BuildingDef{
		Name: "Military Base", Key: "military_base", Category: "military",
		BaseCost:    map[string]float64{"steel": 30e6, "coal": 10e6, "gold": 15e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 2560}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "An industrial-era military base. +2560 military cap (10 workers).",
		LineageKey:  "military", LineageTier: 8,
		WorkerDomain: "military", WorkerCapacity: 10,
		EpochKey: "steel_era",
	})
	// tier 9 — victorian_age  soldiers=5120
	b = append(b, BuildingDef{
		Name: "Garrison", Key: "garrison", Category: "military",
		BaseCost:    map[string]float64{"steel": 200e6, "iron": 100e6, "gold": 120e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 5120}},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "A Victorian-era garrison town. +5120 military cap (10 workers).",
		LineageKey:  "military", LineageTier: 9,
		WorkerDomain: "military", WorkerCapacity: 10,
		EpochKey: "electric_era",
	})
	// tier 10 — electric_age  soldiers=10240
	b = append(b, BuildingDef{
		Name: "Command Post", Key: "command_post", Category: "military",
		BaseCost:    map[string]float64{"steel": 1.2e9, "electricity": 500e6, "gold": 700e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 10240}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "An electrified command and control post. +10240 military cap (12 workers).",
		LineageKey:  "military", LineageTier: 10,
		WorkerDomain: "military", WorkerCapacity: 12,
		EpochKey: "electric_era",
	})
	// tier 11 — atomic_age  soldiers=20480
	b = append(b, BuildingDef{
		Name: "Bunker Complex", Key: "bunker_complex", Category: "military",
		BaseCost:    map[string]float64{"steel": 6e9, "stone": 8e9, "electricity": 2e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 20480}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "A hardened atomic-era bunker complex. +20480 military cap (12 workers).",
		LineageKey:  "military", LineageTier: 11,
		WorkerDomain: "military", WorkerCapacity: 12,
		EpochKey: "electric_era",
	})
	// tier 12 — modern_age  soldiers=40960
	b = append(b, BuildingDef{
		Name: "Special Ops HQ", Key: "special_ops_hq", Category: "military",
		BaseCost:    map[string]float64{"steel": 35e9, "electricity": 12e9, "data": 1e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 40960}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Headquarters for special operations forces. +40960 military cap (14 workers).",
		LineageKey:  "military", LineageTier: 12,
		WorkerDomain: "military", WorkerCapacity: 14,
		EpochKey: "digital_era",
	})
	// tier 13 — information_age  soldiers=81920
	b = append(b, BuildingDef{
		Name: "Cyber Command", Key: "cyber_command", Category: "military",
		BaseCost:    map[string]float64{"electricity": 90e9, "data": 10e9, "gold": 160e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 81920}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "Cyber warfare command centre. +81920 military cap (15 workers).",
		LineageKey:  "military", LineageTier: 13,
		WorkerDomain: "military", WorkerCapacity: 15,
		EpochKey: "digital_era",
	})
	// tier 14 — digital_age  soldiers=163840
	b = append(b, BuildingDef{
		Name: "Drone Warfare Center", Key: "drone_warfare_center", Category: "military",
		BaseCost:    map[string]float64{"electricity": 450e9, "data": 55e9, "steel": 650e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 163840}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Autonomous drone warfare command. +163840 military cap (16 workers).",
		LineageKey:  "military", LineageTier: 14,
		WorkerDomain: "military", WorkerCapacity: 16,
		EpochKey: "digital_era",
	})
	// tier 15 — cyberpunk_age  soldiers=327680
	b = append(b, BuildingDef{
		Name: "Combat Aug Center", Key: "combat_aug_center", Category: "military",
		BaseCost:    map[string]float64{"data": 210e9, "crypto": 1.1e12, "electricity": 2.2e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 327680}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Cybernetic augmentation for soldiers. +327680 military cap (18 workers).",
		LineageKey:  "military", LineageTier: 15,
		WorkerDomain: "military", WorkerCapacity: 18,
		EpochKey: "neon_era",
	})
	// tier 16 — fusion_age  soldiers=655360
	b = append(b, BuildingDef{
		Name: "Plasma Command", Key: "plasma_command", Category: "military",
		BaseCost:    map[string]float64{"plasma": 5e12, "electricity": 14e12, "steel": 18e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 655360}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Plasma-weapon equipped military command. +655360 military cap (20 workers).",
		LineageKey:  "military", LineageTier: 16,
		WorkerDomain: "military", WorkerCapacity: 20,
		EpochKey: "neon_era",
	})
	// tier 17 — space_age  soldiers=1310720
	b = append(b, BuildingDef{
		Name: "Space Force Base", Key: "space_force_base", Category: "military",
		BaseCost:    map[string]float64{"titanium": 80e12, "plasma": 38e12, "electricity": 95e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 1310720}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "An orbital space force base. +1310720 military cap (20 workers).",
		LineageKey:  "military", LineageTier: 17,
		WorkerDomain: "military", WorkerCapacity: 20,
		EpochKey: "neon_era",
	})
	// tier 18 — interstellar_age  soldiers=2621440
	b = append(b, BuildingDef{
		Name: "Fleet Command", Key: "fleet_command", Category: "military",
		BaseCost:    map[string]float64{"dark_matter": 95e12, "titanium": 760e12, "plasma": 460e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 2621440}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Commands a full interstellar fleet. +2621440 military cap (25 workers).",
		LineageKey:  "military", LineageTier: 18,
		WorkerDomain: "military", WorkerCapacity: 25,
		EpochKey: "cosmic_era",
	})
	// tier 19 — galactic_age  soldiers=5242880
	b = append(b, BuildingDef{
		Name: "Stellar Armada HQ", Key: "stellar_armada_hq", Category: "military",
		BaseCost:    map[string]float64{"antimatter": 190e12, "dark_matter": 950e12, "titanium": 4.8e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 5242880}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Headquarters for the galactic armada. +5242880 military cap (25 workers).",
		LineageKey:  "military", LineageTier: 19,
		WorkerDomain: "military", WorkerCapacity: 25,
		EpochKey: "cosmic_era",
	})
	// tier 20 — quantum_age  soldiers=10485760
	b = append(b, BuildingDef{
		Name: "Probability War Room", Key: "probability_war_room", Category: "military",
		BaseCost:    map[string]float64{"quantum_flux": 210e12, "antimatter": 62e15, "dark_matter": 52e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "capacity", Target: "military", Value: 10485760}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Wages war across probability timelines. +10485760 military cap (30 workers).",
		LineageKey:  "military", LineageTier: 20,
		WorkerDomain: "military", WorkerCapacity: 30,
		EpochKey: "cosmic_era",
	})

	return b
}
