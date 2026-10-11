package config

// buildingsLineageMilitary returns lineages 5-9:
// knowledge, faith, military, trade, engineering.
// Merged into newProductionBuildings() via init — see buildings_new_merge.go.
//
// Military rework (Stage 1): every military building produces and stores the
// `soldiers` resource. Each building's {capacity, military} value below is its
// soldier storage cap, and it produces soldiers at max(0.1, cap/50) per tick
// (worker-scaled) so a fully-worked building fills its own soldier cap in ~50
// ticks. The loop at the bottom of this function turns each marker into those
// two effects, so the per-tier caps stay the single source of truth.
func buildingsLineageMilitary() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 7 — MILITARY (lineageKey: "military", domain: "military")
	// Effect type: "capacity", Target: "military", Value: 10 * 2^tier
	// CostScale: 1.35  Category: "military"
	// =========================================================================

	// tier 0 — stone_age  soldiers=10
	b = append(b, BuildingDef{
		Name: "War Camp", Key: "war_camp", Category: "military",
		BaseCost:    map[string]float64{"wood": 760, "stone": 420},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 10}, {Type: "production", Target: "soldiers", Value: 0.2}},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "A fortified war camp.",
		LineageKey:  "military", LineageTier: 0,
		WorkerDomain: "military", WorkerCapacity: 3,
		EpochKey: "stone_era",
	})
	// tier 1 — bronze_age  soldiers=20
	b = append(b, BuildingDef{
		Name: "Barracks", Key: "barracks", Category: "military",
		BaseCost:     map[string]float64{"wood": 3800, "stone": 2500, "iron": 850},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "storage", Target: "soldiers", Value: 20}, {Type: "production", Target: "soldiers", Value: 0.4}},
		BuildTicks:   500,
		RequiredAge:  "bronze_age",
		RequiredTech: "military_tactics",
		Description:  "Trains and houses soldiers.",
		LineageKey:   "military", LineageTier: 1,
		WorkerDomain: "military", WorkerCapacity: 4,
		EpochKey: "stone_era",
	})
	// tier 2 — iron_age  soldiers=40
	b = append(b, BuildingDef{
		Name: "Hunting Lodge", Key: "hunting_lodge", Category: "military",
		BaseCost:    map[string]float64{"wood": 110},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 40}, {Type: "production", Target: "soldiers", Value: 0.8}},
		BuildTicks:  80,
		RequiredAge: "iron_age",
		Description: "A gathering place for hunters turned military post.",
		LineageKey:  "military", LineageTier: 2,
		WorkerDomain: "military", WorkerCapacity: 5,
		EpochKey: "iron_era",
	})
	// tier 3 — iron_age  soldiers=80
	b = append(b, BuildingDef{
		Name: "Legion Fort", Key: "legion_fort", Category: "military",
		BaseCost:     map[string]float64{"stone": 30000, "iron": 15000, "gold": 8500},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "storage", Target: "soldiers", Value: 80}, {Type: "production", Target: "soldiers", Value: 1.6}},
		BuildTicks:   300,
		RequiredAge:  "iron_age",
		RequiredTech: "siege_warfare",
		Description:  "A fortified Roman-style legion camp.",
		LineageKey:   "military", LineageTier: 3,
		WorkerDomain: "military", WorkerCapacity: 6,
		EpochKey: "iron_era",
	})
	// tier 4 — classical_age  soldiers=160
	b = append(b, BuildingDef{
		Name: "Military Academy", Key: "military_academy", Category: "military",
		BaseCost:    map[string]float64{"stone": 170000, "gold": 64000, "iron": 42000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 160}, {Type: "production", Target: "soldiers", Value: 3.2}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Trains elite military officers.",
		LineageKey:  "military", LineageTier: 4,
		WorkerDomain: "military", WorkerCapacity: 6,
		EpochKey: "iron_era",
	})
	// tier 5 — medieval_age  soldiers=320
	b = append(b, BuildingDef{
		Name: "Castle Keep", Key: "castle_keep", Category: "military",
		BaseCost:    map[string]float64{"stone": 930000, "iron": 300000, "gold": 210000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 320}, {Type: "production", Target: "soldiers", Value: 6.4}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "A fortified stone keep.",
		LineageKey:  "military", LineageTier: 5,
		WorkerDomain: "military", WorkerCapacity: 7,
		EpochKey: "iron_era",
	})
	// tier 6 — renaissance_age  soldiers=640
	b = append(b, BuildingDef{
		Name: "Fortress", Key: "fortress", Category: "military",
		BaseCost:    map[string]float64{"stone": 3000000, "gold": 1300000, "steel": 640000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 640}, {Type: "production", Target: "soldiers", Value: 12.8}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "A star-fort capable of holding a large garrison.",
		LineageKey:  "military", LineageTier: 6,
		WorkerDomain: "military", WorkerCapacity: 7,
		EpochKey: "steel_era",
	})
	// tier 7 — colonial_age  soldiers=1280
	b = append(b, BuildingDef{
		Name: "Fort", Key: "fort", Category: "military",
		BaseCost:    map[string]float64{"gold": 1.7e07, "steel": 8500000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 1280}, {Type: "production", Target: "soldiers", Value: 25.6}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "A colonial frontier fort.",
		LineageKey:  "military", LineageTier: 7,
		WorkerDomain: "military", WorkerCapacity: 8,
		EpochKey: "steel_era",
	})
	// tier 8 — industrial_age  soldiers=2560
	b = append(b, BuildingDef{
		Name: "Military Base", Key: "military_base", Category: "military",
		BaseCost:    map[string]float64{"steel": 1.3e08, "coal": 4.2e07, "gold": 6.4e07},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 2560}, {Type: "production", Target: "soldiers", Value: 51.2}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "An industrial-era military base.",
		LineageKey:  "military", LineageTier: 8,
		WorkerDomain: "military", WorkerCapacity: 10,
		EpochKey: "steel_era",
	})
	// tier 9 — victorian_age  soldiers=5120
	b = append(b, BuildingDef{
		Name: "Garrison", Key: "garrison", Category: "military",
		BaseCost:    map[string]float64{"steel": 8.5e08, "iron": 4.2e08, "gold": 5.1e08},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 5120}, {Type: "production", Target: "soldiers", Value: 102.4}},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "A Victorian-era garrison town.",
		LineageKey:  "military", LineageTier: 9,
		WorkerDomain: "military", WorkerCapacity: 10,
		EpochKey: "electric_era",
	})
	// tier 10 — electric_age  soldiers=10240
	b = append(b, BuildingDef{
		Name: "Command Post", Key: "command_post", Category: "military",
		BaseCost:    map[string]float64{"steel": 5.1e09, "electricity": 2.1e09, "gold": 3e09},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 10240}, {Type: "production", Target: "soldiers", Value: 204.8}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "An electrified command and control post.",
		LineageKey:  "military", LineageTier: 10,
		WorkerDomain: "military", WorkerCapacity: 12,
		EpochKey: "electric_era",
	})
	// tier 11 — atomic_age  soldiers=20480
	b = append(b, BuildingDef{
		Name: "Bunker Complex", Key: "bunker_complex", Category: "military",
		BaseCost:    map[string]float64{"steel": 2.5e10, "stone": 3.4e10, "electricity": 8.5e09},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 20480}, {Type: "production", Target: "soldiers", Value: 409.6}},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "A hardened atomic-era bunker complex.",
		LineageKey:  "military", LineageTier: 11,
		WorkerDomain: "military", WorkerCapacity: 12,
		EpochKey: "electric_era",
	})
	// tier 12 — modern_age  soldiers=40960
	b = append(b, BuildingDef{
		Name: "Special Ops HQ", Key: "special_ops_hq", Category: "military",
		BaseCost:    map[string]float64{"steel": 1.5e11, "electricity": 5.1e10, "data": 4.2e09},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 40960}, {Type: "production", Target: "soldiers", Value: 819.2}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "Headquarters for special operations forces.",
		LineageKey:  "military", LineageTier: 12,
		WorkerDomain: "military", WorkerCapacity: 14,
		EpochKey: "digital_era",
	})
	// tier 13 — information_age  soldiers=81920
	b = append(b, BuildingDef{
		Name: "Cyber Command", Key: "cyber_command", Category: "military",
		BaseCost:     map[string]float64{"electricity": 3.8e11, "data": 4.2e10, "gold": 6.8e11},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "storage", Target: "soldiers", Value: 81920}, {Type: "production", Target: "soldiers", Value: 1638.4}},
		BuildTicks:   3600,
		RequiredAge:  "information_age",
		RequiredTech: "cybersecurity",
		Description:  "Cyber warfare command center.",
		LineageKey:   "military", LineageTier: 13,
		WorkerDomain: "military", WorkerCapacity: 15,
		EpochKey: "digital_era",
	})
	// tier 14 — digital_age  soldiers=163840
	b = append(b, BuildingDef{
		Name: "Drone Warfare Center", Key: "drone_warfare_center", Category: "military",
		BaseCost:    map[string]float64{"electricity": 1.9e12, "data": 2.3e11, "steel": 2.8e12},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 163840}, {Type: "production", Target: "soldiers", Value: 3276.8}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Autonomous drone warfare command.",
		LineageKey:  "military", LineageTier: 14,
		WorkerDomain: "military", WorkerCapacity: 16,
		EpochKey: "digital_era",
	})
	// tier 15 — cyberpunk_age  soldiers=327680
	b = append(b, BuildingDef{
		Name: "Combat Aug Center", Key: "combat_aug_center", Category: "military",
		BaseCost:     map[string]float64{"data": 8.9e11, "crypto": 4.7e12, "electricity": 9.3e12},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "storage", Target: "soldiers", Value: 327680}, {Type: "production", Target: "soldiers", Value: 6553.6}},
		BuildTicks:   3600,
		RequiredAge:  "cyberpunk_age",
		RequiredTech: "cybernetics",
		Description:  "Cybernetic augmentation for soldiers.",
		LineageKey:   "military", LineageTier: 15,
		WorkerDomain: "military", WorkerCapacity: 18,
		EpochKey: "neon_era",
	})
	// tier 16 — fusion_age  soldiers=655360
	b = append(b, BuildingDef{
		Name: "Plasma Command", Key: "plasma_command", Category: "military",
		BaseCost:    map[string]float64{"plasma": 2.1e13, "electricity": 5.9e13, "steel": 7.6e13},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 655360}, {Type: "production", Target: "soldiers", Value: 13107.2}},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "Plasma-weapon equipped military command.",
		LineageKey:  "military", LineageTier: 16,
		WorkerDomain: "military", WorkerCapacity: 20,
		EpochKey: "neon_era",
	})
	// tier 17 — space_age  soldiers=1310720
	b = append(b, BuildingDef{
		Name: "Space Force Base", Key: "space_force_base", Category: "military",
		BaseCost:    map[string]float64{"titanium": 3.4e14, "plasma": 1.6e14, "electricity": 4e14},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 1310720}, {Type: "production", Target: "soldiers", Value: 26214.4}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "An orbital space force base.",
		LineageKey:  "military", LineageTier: 17,
		WorkerDomain: "military", WorkerCapacity: 20,
		EpochKey: "neon_era",
	})
	// tier 18 — interstellar_age  soldiers=2621440
	b = append(b, BuildingDef{
		Name: "Fleet Command", Key: "fleet_command", Category: "military",
		BaseCost:    map[string]float64{"dark_matter": 4e14, "titanium": 3.2e15, "plasma": 1.9e15},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 2621440}, {Type: "production", Target: "soldiers", Value: 52428.8}},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "Commands a full interstellar fleet.",
		LineageKey:  "military", LineageTier: 18,
		WorkerDomain: "military", WorkerCapacity: 25,
		EpochKey: "cosmic_era",
	})
	// tier 19 — galactic_age  soldiers=5242880
	b = append(b, BuildingDef{
		Name: "Stellar Armada HQ", Key: "stellar_armada_hq", Category: "military",
		BaseCost:    map[string]float64{"antimatter": 8e14, "dark_matter": 4e15, "titanium": 2e16},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 5242880}, {Type: "production", Target: "soldiers", Value: 104857.6}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Headquarters for the galactic armada.",
		LineageKey:  "military", LineageTier: 19,
		WorkerDomain: "military", WorkerCapacity: 25,
		EpochKey: "cosmic_era",
	})
	// tier 20 — quantum_age  soldiers=10485760
	b = append(b, BuildingDef{
		Name: "Probability War Room", Key: "probability_war_room", Category: "military",
		BaseCost:    map[string]float64{"quantum_flux": 8.9e14, "antimatter": 2.6e17, "dark_matter": 2.2e17},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "storage", Target: "soldiers", Value: 1.048576e07}, {Type: "production", Target: "soldiers", Value: 209715.2}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Wages war across probability timelines.",
		LineageKey:  "military", LineageTier: 20,
		WorkerDomain: "military", WorkerCapacity: 30,
		EpochKey: "cosmic_era",
	})
	// tier 21 — transcendent_age  soldiers=20971520
	b = append(b, BuildingDef{
		Name: "Omniversal War Council", Key: "omniversal_war_council", Category: "military",
		BaseCost:     map[string]float64{"quantum_flux": 8.9e15, "antimatter": 2.6e18, "dark_matter": 2.2e18},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "storage", Target: "soldiers", Value: 2.097152e07}, {Type: "production", Target: "soldiers", Value: 419430.4}},
		BuildTicks:   3600,
		RequiredAge:  "transcendent_age",
		RequiredTech: "omniversal_command",
		Description:  "Commands forces across the omniverse.",
		LineageKey:   "military", LineageTier: 21,
		WorkerDomain: "military", WorkerCapacity: 35,
		EpochKey: "cosmic_era",
	})

	// Each military building holds soldiers (a storage effect) and trains them
	// (a production effect): a fully staffed one fills its own store in about
	// 50 ticks. Both numbers are written on the building.
	return b
}
