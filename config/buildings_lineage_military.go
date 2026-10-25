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
	// CostScale: ProducerRate (1.15)  Category: "military"  (both in the lineageSpec at the foot)
	// =========================================================================

	// tier 0 — stone_age
	b = append(b, BuildingDef{
		Name: "War Camp", Key: "war_camp",
		BaseCost:    map[string]float64{"wood": 760, "stone": 420},
		BuildTicks:  200,
		RequiredAge: "stone_age",
		Description: "A fortified war camp.",
	})
	// tier 1 — bronze_age
	b = append(b, BuildingDef{
		Name: "Barracks", Key: "barracks",
		BaseCost:     map[string]float64{"wood": 3800, "stone": 2500, "iron": 850},
		BuildTicks:   500,
		RequiredAge:  "bronze_age",
		RequiredTech: "military_tactics",
		Description:  "Houses soldiers.",
	})
	// tier 2 — iron_age
	b = append(b, BuildingDef{
		Name: "Hunting Lodge", Key: "hunting_lodge",
		BaseCost:    map[string]float64{"wood": 110},
		BuildTicks:  80,
		RequiredAge: "iron_age",
		Description: "A gathering place for hunters turned military post.",
	})
	// tier 3 — iron_age
	b = append(b, BuildingDef{
		Name: "Legion Fort", Key: "legion_fort",
		BaseCost:     map[string]float64{"stone": 15000, "marble": 15000, "iron": 7500, "iron_ore": 7500, "gold": 8500},
		BuildTicks:   300,
		RequiredAge:  "iron_age",
		RequiredTech: "siege_warfare",
		Description:  "A fortified Roman-style legion camp.",
	})
	// tier 4 — classical_age
	b = append(b, BuildingDef{
		Name: "Military Academy", Key: "military_academy",
		BaseCost:    map[string]float64{"stone": 85000, "marble": 85000, "gold": 64000, "iron": 21000, "iron_ore": 21000},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "Trains elite military officers.",
	})
	// tier 5 — medieval_age
	b = append(b, BuildingDef{
		Name: "Castle Keep", Key: "castle_keep",
		BaseCost:    map[string]float64{"stone": 930000, "iron": 150000, "iron_ore": 150000, "gold": 210000},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "A fortified stone keep.",
	})
	// tier 6 — renaissance_age
	b = append(b, BuildingDef{
		Name: "Fortress", Key: "fortress",
		BaseCost:    map[string]float64{"stone": 3 * Million, "gold": 1.3 * Million, "steel": 640000},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "A star-fort capable of holding a large garrison.",
	})
	// tier 7 — colonial_age
	b = append(b, BuildingDef{
		Name: "Fort", Key: "fort",
		BaseCost:    map[string]float64{"gold": 17 * Million, "steel": 8.5 * Million},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "A colonial frontier fort.",
	})
	// tier 8 — industrial_age
	b = append(b, BuildingDef{
		Name: "Military Base", Key: "military_base",
		BaseCost:    map[string]float64{"steel": 130 * Million, "coal": 42 * Million, "gold": 64 * Million},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "An industrial-era military base.",
	})
	// tier 9 — victorian_age
	b = append(b, BuildingDef{
		Name: "Garrison", Key: "garrison",
		BaseCost:    map[string]float64{"steel": 850 * Million, "iron": 420 * Million, "gold": 510 * Million},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "A Victorian-era garrison town.",
	})
	// tier 10 — electric_age
	b = append(b, BuildingDef{
		Name: "Command Post", Key: "command_post",
		BaseCost:    map[string]float64{"steel": 5.1 * Billion, "electricity": 2.1 * Billion, "gold": 3 * Billion},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "An electrified command and control post.",
	})
	// tier 11 — atomic_age
	b = append(b, BuildingDef{
		Name: "Bunker Complex", Key: "bunker_complex",
		BaseCost:    map[string]float64{"steel": 25 * Billion, "stone": 34 * Billion, "electricity": 8.5 * Billion},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "A hardened atomic-era bunker complex.",
	})
	// tier 12 — modern_age
	b = append(b, BuildingDef{
		Name: "Special Ops HQ", Key: "special_ops_hq",
		BaseCost:    map[string]float64{"steel": 75 * Billion, "titanium_ore": 75 * Billion, "electricity": 51 * Billion, "data": 4.2 * Billion},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "Headquarters for special operations forces.",
	})
	// tier 13 — information_age
	b = append(b, BuildingDef{
		Name: "Cyber Command", Key: "cyber_command",
		BaseCost:     map[string]float64{"electricity": 380 * Billion, "data": 42 * Billion, "gold": 680 * Billion},
		BuildTicks:   3600,
		RequiredAge:  "information_age",
		RequiredTech: "cybersecurity",
		Description:  "Cyber warfare command center.",
	})
	// tier 14 — digital_age
	b = append(b, BuildingDef{
		Name: "Drone Warfare Center", Key: "drone_warfare_center",
		BaseCost:    map[string]float64{"electricity": 1.9 * Trillion, "data": 230 * Billion, "steel": 1.4 * Trillion, "titanium_ore": 1.4 * Trillion},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Autonomous drone warfare command.",
	})
	// tier 15 — cyberpunk_age
	b = append(b, BuildingDef{
		Name: "Combat Aug Center", Key: "combat_aug_center",
		BaseCost:     map[string]float64{"data": 890 * Billion, "crypto": 4.7 * Trillion, "electricity": 4.65 * Trillion, "dark_matter_crystals": 4.65 * Trillion},
		BuildTicks:   3600,
		RequiredAge:  "cyberpunk_age",
		RequiredTech: "cybernetics",
		Description:  "Cybernetic augmentation for soldiers.",
	})
	// tier 16 — fusion_age
	b = append(b, BuildingDef{
		Name: "Plasma Command", Key: "plasma_command",
		BaseCost:    map[string]float64{"plasma": 21 * Trillion, "electricity": 29.5 * Trillion, "dark_matter_crystals": 29.5 * Trillion, "steel": 76 * Trillion},
		BuildTicks:  3600,
		RequiredAge: "fusion_age",
		Description: "Plasma-weapon equipped military command.",
	})
	// tier 17 — space_age
	b = append(b, BuildingDef{
		Name: "Space Force Base", Key: "space_force_base",
		BaseCost:    map[string]float64{"titanium": 340 * Trillion, "plasma": 160 * Trillion, "electricity": 200 * Trillion, "dark_matter_crystals": 200 * Trillion},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "An orbital space force base.",
	})
	// tier 18 — interstellar_age
	b = append(b, BuildingDef{
		Name: "Fleet Command", Key: "fleet_command",
		BaseCost:    map[string]float64{"dark_matter": 400 * Trillion, "titanium": 3.2 * Quadrillion, "plasma": 1.9 * Quadrillion},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "Commands a full interstellar fleet.",
	})
	// tier 19 — galactic_age
	b = append(b, BuildingDef{
		Name: "Stellar Armada HQ", Key: "stellar_armada_hq",
		BaseCost:    map[string]float64{"antimatter": 800 * Trillion, "dark_matter": 4 * Quadrillion, "titanium": 20 * Quadrillion},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Headquarters for the galactic armada.",
	})
	// tier 20 — quantum_age
	b = append(b, BuildingDef{
		Name: "Probability War Room", Key: "probability_war_room",
		BaseCost:    map[string]float64{"quantum_flux": 890 * Trillion, "antimatter": 260 * Quadrillion, "dark_matter": 220 * Quadrillion},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Wages war across probability timelines.",
	})
	// tier 21 — transcendent_age
	b = append(b, BuildingDef{
		Name: "Omniversal War Council", Key: "omniversal_war_council",
		BaseCost:     map[string]float64{"quantum_flux": 8.9 * Quadrillion, "antimatter": 2.6 * Quintillion, "dark_matter": 2.2 * Quintillion},
		BuildTicks:   3600,
		RequiredAge:  "transcendent_age",
		RequiredTech: "omniversal_command",
		Description:  "Commands forces across the omniverse.",
	})

	return finishLineage(b, lineageSpec{
		Key: "military", Category: "military", Scale: ProducerRate,
	})
}
