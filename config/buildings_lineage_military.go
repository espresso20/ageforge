package config

// buildingsLineageMilitary returns lineages 5-9:
// knowledge, faith, military, trade, engineering.
// Merged into newProductionBuildings() via init — see buildings_new_merge.go.
//
// A military building comes with its soldiers (BuildingDef.Soldiers a copy):
// they are people who live in the homes and eat, not a stock the building
// trains. The number is derived from the reference town (soldier_rule.go) when
// the table is built, so these entries carry no soldier figure and no effect.
func buildingsLineageMilitary() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 7 — MILITARY (lineageKey: "military"; no worker slots)
	// CostScale: 1.35  Category: "military"
	// =========================================================================

	// tier 0 — stone_age
	b = append(b, BuildingDef{
		Name: "War Camp", Key: "war_camp", Category: "military",
		BaseCost:    map[string]float64{"wood": 760, "stone": 420},
		CostScale:   1.15,
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "A fortified war camp.",
		LineageKey:  "military", LineageTier: 0,
		EpochKey: "stone_era",
	})
	// tier 1 — bronze_age
	b = append(b, BuildingDef{
		Name: "Barracks", Key: "barracks", Category: "military",
		BaseCost:     map[string]float64{"wood": 3800, "stone": 2500, "iron": 850},
		CostScale:    1.15,
		BuildTicks:   500,
		RequiredAge:  "bronze_age",
		RequiredTech: "military_tactics",
		Description:  "Trains and houses soldiers.",
		LineageKey:   "military", LineageTier: 1,
		EpochKey: "stone_era",
	})
	// tier 2 — iron_age
	b = append(b, BuildingDef{
		Name: "Hunting Lodge", Key: "hunting_lodge", Category: "military",
		BaseCost:    map[string]float64{"wood": 110},
		CostScale:   1.15,
		BuildTicks:  80,
		RequiredAge: "iron_age",
		Description: "A gathering place for hunters turned military post.",
		LineageKey:  "military", LineageTier: 2,
		EpochKey: "iron_era",
	})
	// tier 3 — iron_age
	b = append(b, BuildingDef{
		Name: "Legion Fort", Key: "legion_fort", Category: "military",
		BaseCost:     map[string]float64{"stone": 15000, "marble": 15000, "iron": 7500, "iron_ore": 7500, "gold": 8500},
		CostScale:    1.15,
		BuildTicks:   300,
		RequiredAge:  "iron_age",
		RequiredTech: "siege_warfare",
		Description:  "A fortified Roman-style legion camp.",
		LineageKey:   "military", LineageTier: 3,
		EpochKey: "iron_era",
	})
	// tier 4 — classical_age
	b = append(b, BuildingDef{
		Name: "Military Academy", Key: "military_academy", Category: "military",
		BaseCost:    map[string]float64{"stone": 85000, "marble": 85000, "gold": 64000, "iron": 21000, "iron_ore": 21000},
		CostScale:   1.15,
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Trains elite military officers.",
		LineageKey:  "military", LineageTier: 4,
		EpochKey: "iron_era",
	})
	// tier 5 — medieval_age
	b = append(b, BuildingDef{
		Name: "Castle Keep", Key: "castle_keep", Category: "military",
		BaseCost:    map[string]float64{"stone": 930000, "iron": 150000, "iron_ore": 150000, "gold": 210000},
		CostScale:   1.15,
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "A fortified stone keep.",
		LineageKey:  "military", LineageTier: 5,
		EpochKey: "iron_era",
	})
	// tier 6 — renaissance_age
	b = append(b, BuildingDef{
		Name: "Fortress", Key: "fortress", Category: "military",
		BaseCost:    map[string]float64{"stone": 3000000, "gold": 1300000, "steel": 640000},
		CostScale:   1.15,
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "A star-fort capable of holding a large garrison.",
		LineageKey:  "military", LineageTier: 6,
		EpochKey: "steel_era",
	})
	// tier 7 — colonial_age
	b = append(b, BuildingDef{
		Name: "Fort", Key: "fort", Category: "military",
		BaseCost:    map[string]float64{"gold": 1.7e07, "steel": 8500000},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "A colonial frontier fort.",
		LineageKey:  "military", LineageTier: 7,
		EpochKey: "steel_era",
	})
	// tier 8 — industrial_age
	b = append(b, BuildingDef{
		Name: "Military Base", Key: "military_base", Category: "military",
		BaseCost:    map[string]float64{"steel": 1.3e08, "coal": 4.2e07, "gold": 6.4e07},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "An industrial-era military base.",
		LineageKey:  "military", LineageTier: 8,
		EpochKey: "steel_era",
	})
	// tier 9 — victorian_age
	b = append(b, BuildingDef{
		Name: "Garrison", Key: "garrison", Category: "military",
		BaseCost:    map[string]float64{"steel": 8.5e08, "iron": 4.2e08, "gold": 5.1e08},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "A Victorian-era garrison town.",
		LineageKey:  "military", LineageTier: 9,
		EpochKey: "electric_era",
	})
	// tier 10 — electric_age
	b = append(b, BuildingDef{
		Name: "Command Post", Key: "command_post", Category: "military",
		BaseCost:    map[string]float64{"steel": 5.1e09, "electricity": 2.1e09, "gold": 3e09},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "An electrified command and control post.",
		LineageKey:  "military", LineageTier: 10,
		EpochKey: "electric_era",
	})
	// tier 11 — atomic_age
	b = append(b, BuildingDef{
		Name: "Bunker Complex", Key: "bunker_complex", Category: "military",
		BaseCost:    map[string]float64{"steel": 2.5e10, "stone": 3.4e10, "electricity": 8.5e09},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "A hardened atomic-era bunker complex.",
		LineageKey:  "military", LineageTier: 11,
		EpochKey: "electric_era",
	})
	// tier 12 — modern_age
	b = append(b, BuildingDef{
		Name: "Special Ops HQ", Key: "special_ops_hq", Category: "military",
		BaseCost:    map[string]float64{"steel": 7.5e10, "titanium_ore": 7.5e10, "electricity": 5.1e10, "data": 4.2e09},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "Headquarters for special operations forces.",
		LineageKey:  "military", LineageTier: 12,
		EpochKey: "digital_era",
	})
	// tier 13 — information_age
	b = append(b, BuildingDef{
		Name: "Cyber Command", Key: "cyber_command", Category: "military",
		BaseCost:     map[string]float64{"electricity": 3.8e11, "data": 4.2e10, "gold": 6.8e11},
		CostScale:    1.15,
		BuildTicks:   3600,
		RequiredAge:  "information_age",
		RequiredTech: "cybersecurity",
		Description:  "Cyber warfare command center.",
		LineageKey:   "military", LineageTier: 13,
		EpochKey: "digital_era",
	})
	// tier 14 — digital_age
	b = append(b, BuildingDef{
		Name: "Drone Warfare Center", Key: "drone_warfare_center", Category: "military",
		BaseCost:    map[string]float64{"electricity": 1.9e12, "data": 2.3e11, "steel": 1.4e12, "titanium_ore": 1.4e12},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Autonomous drone warfare command.",
		LineageKey:  "military", LineageTier: 14,
		EpochKey: "digital_era",
	})
	// tier 15 — cyberpunk_age
	b = append(b, BuildingDef{
		Name: "Combat Aug Center", Key: "combat_aug_center", Category: "military",
		BaseCost:     map[string]float64{"data": 8.9e11, "crypto": 4.7e12, "electricity": 4.65e12, "dark_matter_crystals": 4.65e12},
		CostScale:    1.15,
		BuildTicks:   3600,
		RequiredAge:  "cyberpunk_age",
		RequiredTech: "cybernetics",
		Description:  "Cybernetic augmentation for soldiers.",
		LineageKey:   "military", LineageTier: 15,
		EpochKey: "neon_era",
	})
	// tier 16 — fusion_age
	b = append(b, BuildingDef{
		Name: "Plasma Command", Key: "plasma_command", Category: "military",
		BaseCost:    map[string]float64{"plasma": 2.1e13, "electricity": 2.95e13, "dark_matter_crystals": 2.95e13, "steel": 7.6e13},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "Plasma-weapon equipped military command.",
		LineageKey:  "military", LineageTier: 16,
		EpochKey: "neon_era",
	})
	// tier 17 — space_age
	b = append(b, BuildingDef{
		Name: "Space Force Base", Key: "space_force_base", Category: "military",
		BaseCost:    map[string]float64{"titanium": 3.4e14, "plasma": 1.6e14, "electricity": 2e14, "dark_matter_crystals": 2e14},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "An orbital space force base.",
		LineageKey:  "military", LineageTier: 17,
		EpochKey: "neon_era",
	})
	// tier 18 — interstellar_age
	b = append(b, BuildingDef{
		Name: "Fleet Command", Key: "fleet_command", Category: "military",
		BaseCost:    map[string]float64{"dark_matter": 4e14, "titanium": 3.2e15, "plasma": 1.9e15},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "Commands a full interstellar fleet.",
		LineageKey:  "military", LineageTier: 18,
		EpochKey: "cosmic_era",
	})
	// tier 19 — galactic_age
	b = append(b, BuildingDef{
		Name: "Stellar Armada HQ", Key: "stellar_armada_hq", Category: "military",
		BaseCost:    map[string]float64{"antimatter": 8e14, "dark_matter": 4e15, "titanium": 2e16},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Headquarters for the galactic armada.",
		LineageKey:  "military", LineageTier: 19,
		EpochKey: "cosmic_era",
	})
	// tier 20 — quantum_age
	b = append(b, BuildingDef{
		Name: "Probability War Room", Key: "probability_war_room", Category: "military",
		BaseCost:    map[string]float64{"quantum_flux": 8.9e14, "antimatter": 2.6e17, "dark_matter": 2.2e17},
		CostScale:   1.15,
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Wages war across probability timelines.",
		LineageKey:  "military", LineageTier: 20,
		EpochKey: "cosmic_era",
	})
	// tier 21 — transcendent_age
	b = append(b, BuildingDef{
		Name: "Omniversal War Council", Key: "omniversal_war_council", Category: "military",
		BaseCost:     map[string]float64{"quantum_flux": 8.9e15, "antimatter": 2.6e18, "dark_matter": 2.2e18},
		CostScale:    1.15,
		BuildTicks:   3600,
		RequiredAge:  "transcendent_age",
		RequiredTech: "omniversal_command",
		Description:  "Commands forces across the omniverse.",
		LineageKey:   "military", LineageTier: 21,
		EpochKey: "cosmic_era",
	})

	return b
}
