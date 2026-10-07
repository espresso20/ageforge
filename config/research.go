package config

// TechDef defines a single technology in the tech tree.
// Technologies are gated by both age (Age field) and prerequisites: every key
// in Prerequisites must be researched, and one key of AnyOf when it has any.
// Only one tech can be researched at a time.
//
// A tech's kind (keystone, spine, capstone, optional) is not a field: it is
// worked out from the wonders (TechKinds), so it cannot go stale.
type TechDef struct {
	Name string
	Key  string
	Age  string // minimum age key required to start research
	// Lane is the tech's column in the tree: one of the TechLanes keys.
	Lane string
	// Code is the tech's letterhead on the tree's small badges: two to
	// TechCodeMax capital letters, unlike any other tech's. Left empty, it
	// is made from the name (TechCodeFor).
	Code string
	// Emblem is the one glyph inside the tech's badge, one cell wide. Left
	// empty, it is the lane's glyph.
	Emblem string
	Cost   float64 // knowledge points consumed on research completion
	// Prerequisites are the tech keys that must all be researched first.
	Prerequisites []string
	// AnyOf is an either-or group on top of Prerequisites: at least one of
	// these must be researched too. Empty means no such group. A tech has
	// one group at most, of two keys or more.
	AnyOf []string
	// Capstone marks a tech that ends its lane for an era.
	Capstone      bool
	Effects       []TechEffect // applied permanently when research finishes
	Description   string
	ResearchTicks int // game ticks to complete (at 1x speed); scales with research_speed bonus
}

// Technologies returns every tech definition, ordered loosely by age.
// Use TechByKey() for random access or TechsByAge() to group by age.
func Technologies() []TechDef {
	return fillTechArt(normalizeResearchTicks([]TechDef{
		// === PRIMITIVE AGE === (~1 min each)
		{
			Name: "Tool Making", Key: "tool_making", Code: "TOOLS",
			Age: "primitive_age", Lane: LaneCraft, Cost: 800, ResearchTicks: 200,
			Description: "Stone tools make every worker more productive.",
			Effects: []TechEffect{
				{Kind: EffectWorkerOutput, Value: 0.15},
			},
		},
		{
			Name: "Fire Mastery", Key: "fire_mastery", Emblem: "△",
			Age: "primitive_age", Lane: LaneAgriculture, Cost: 1000, ResearchTicks: 200,
			Prerequisites: []string{"tool_making"},
			Description:   "Control of fire improves food preservation and warmth.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "food", Value: 0.1},
			},
		},

		// === STONE AGE === (~2 min each)
		{
			Name: "Stoneworking", Key: "stoneworking", Emblem: "◆",
			Age: "stone_age", Lane: LaneMaterials, Cost: 6000, ResearchTicks: 500,
			Prerequisites: []string{"tool_making"},
			Description:   "Cutting and shaping stone for construction.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "stone", Value: 0.2},
			},
		},
		{
			Name: "Animal Husbandry", Key: "animal_husbandry", Code: "HERDS", Emblem: "♞",
			Age: "stone_age", Lane: LaneAgriculture, Cost: 7500, ResearchTicks: 550,
			Prerequisites: []string{"fire_mastery"},
			Description:   "Domesticating animals for food and labor.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "food", Value: 0.2},
			},
		},
		{
			Name: "Pottery", Key: "pottery", Code: "POTS", Emblem: "∪",
			Age: "stone_age", Lane: LaneTrade, Cost: 5000, ResearchTicks: 450,
			Prerequisites: []string{"fire_mastery"},
			Description:   "Clay vessels for storage and trade.",
			Effects: []TechEffect{
				{Kind: EffectFlatStorage, Target: AllResources, Value: 25},
			},
		},
		{
			Name: "Primitive Writing", Key: "primitive_writing", Code: "WRITE",
			Age: "stone_age", Lane: LaneKnowledge, Cost: 10000, ResearchTicks: 600,
			Prerequisites: []string{"pottery"},
			Description:   "Early symbols enable knowledge transfer.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.1},
			},
		},

		// === BRONZE AGE === (~3 min each)
		{
			Name: "Bronze Working", Key: "bronze_working",
			Age: "bronze_age", Lane: LaneMaterials, Cost: 1600, ResearchTicks: 750,
			Prerequisites: []string{"stoneworking"},
			Description:   "Alloying copper and tin creates durable tools.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "stone", Value: 0.2},
				{Kind: EffectWorkerOutput, Value: 0.1},
			},
		},
		{
			Name: "Agriculture", Key: "agriculture", Code: "FARMS",
			Age: "bronze_age", Lane: LaneAgriculture, Cost: 12000, ResearchTicks: 700,
			Prerequisites: []string{"animal_husbandry"},
			Description:   "Systematic farming adds steady food output.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "food", Value: 0.5},
			},
		},
		{
			Name: "Currency", Key: "currency", Code: "COIN", Emblem: "¤",
			Age: "bronze_age", Lane: LaneTrade, Cost: 17500, ResearchTicks: 800,
			Prerequisites: []string{"primitive_writing"},
			Description:   "Standardized money raises gold output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.3},
			},
		},
		{
			Name: "Masonry", Key: "masonry", Emblem: "▦",
			Age: "bronze_age", Lane: LaneCraft, Cost: 13000, ResearchTicks: 700,
			Prerequisites: []string{"stoneworking"},
			Description:   "Advanced stone construction techniques.",
			Effects: []TechEffect{
				{Kind: EffectFlatStorage, Target: AllResources, Value: 50},
			},
		},
		{
			Name: "Military Tactics", Key: "military_tactics", Code: "TACTI",
			Age: "bronze_age", Lane: LaneMilitary, Cost: 20000, ResearchTicks: 900,
			Prerequisites: []string{"bronze_working"},
			Description:   "Organized warfare and defense strategies.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.2},
			},
		},

		// === IRON AGE === (~4 min each)
		{
			Name: "Iron Smelting", Key: "iron_smelting", Emblem: "■",
			Age: "iron_age", Lane: LaneMaterials, Cost: 30000, ResearchTicks: 1100,
			Prerequisites: []string{"bronze_working"},
			Description:   "Hotter furnaces raise iron output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "iron", Value: 0.4},
				{Kind: EffectFlatOutput, Target: "iron", Value: 0.2},
			},
		},
		{
			Name: "Road Building", Key: "road_building", Code: "ROADS", Emblem: "═",
			Age: "iron_age", Lane: LaneCraft, Cost: 25000, ResearchTicks: 950,
			Prerequisites: []string{"masonry"},
			Description:   "Paved roads improve trade and movement.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.2},
				{Kind: EffectWorkerOutput, Value: 0.1},
			},
		},
		{
			Name: "Mathematics", Key: "mathematics", Code: "MATH", Emblem: "π",
			Age: "iron_age", Lane: LaneKnowledge, Cost: 37500, ResearchTicks: 1200,
			Prerequisites: []string{"primitive_writing", "currency"},
			Description:   "Advanced calculation raises knowledge output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.2},
			},
		},
		{
			Name: "Siege Warfare", Key: "siege_warfare", Emblem: "✕",
			Age: "iron_age", Lane: LaneMilitary, Cost: 35000, ResearchTicks: 1005,
			Prerequisites: []string{"military_tactics"},
			Description:   "Siege engines and fortification techniques.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.3},
			},
		},

		// === CLASSICAL AGE === (~5 min each)
		{
			Name: "Philosophy", Key: "philosophy", Emblem: "Φ",
			Age: "classical_age", Lane: LaneKnowledge, Cost: 20000, ResearchTicks: 1500,
			Prerequisites: []string{"mathematics", "primitive_writing"},
			Description:   "Systematic inquiry into fundamental questions.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.3},
				{Kind: EffectFlatOutput, Target: "culture", Value: 0.2},
			},
		},
		{
			Name: "Civil Engineering", Key: "civil_engineering", Emblem: "∩",
			Age: "classical_age", Lane: LaneCraft, Cost: 18000, ResearchTicks: 1300,
			Prerequisites: []string{"masonry", "road_building"},
			Description:   "Large-scale construction and infrastructure.",
			Effects: []TechEffect{
				{Kind: EffectFlatStorage, Target: AllResources, Value: 100},
				{Kind: EffectBuildCost, Value: -0.05},
			},
		},
		{
			Name: "Imperial Legions", Key: "imperial_legions", Code: "LEGIO", Emblem: "⚑",
			Age: "classical_age", Lane: LaneMilitary, Cost: 22000, ResearchTicks: 1600,
			Prerequisites: []string{"siege_warfare", "iron_smelting"},
			Description:   "Professional standing armies with superior discipline.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.4},
			},
		},

		// === MEDIEVAL AGE === (~7 min each)
		{
			Name: "Steel Forging", Key: "steel_forging", Emblem: "▣",
			Age: "medieval_age", Lane: LaneMaterials, Cost: 25000, ResearchTicks: 2000,
			Prerequisites: []string{"iron_smelting"},
			Description:   "Refining iron into steel for superior tools and weapons.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "steel", Value: 0.25},
				{Kind: EffectOutput, Target: "iron", Value: 0.3},
			},
		},
		{
			Name: "Theology", Key: "theology", Emblem: "Θ",
			Age: "medieval_age", Lane: LaneFaith, Cost: 20000, ResearchTicks: 1800,
			Prerequisites: []string{"philosophy"},
			Description:   "Organized religion provides faith and social cohesion.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "faith", Value: 0.3},
			},
		},
		{
			Name: "Banking", Key: "banking", Code: "BANK",
			Age: "medieval_age", Lane: LaneTrade, Cost: 30000, ResearchTicks: 2100,
			Prerequisites: []string{"currency", "mathematics"},
			Description:   "Financial institutions raise gold output and gold storage.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.5},
				{Kind: EffectFlatStorage, Target: "gold", Value: 100},
			},
		},
		{
			Name: "Feudalism", Key: "feudalism", Emblem: "⌂",
			Age: "medieval_age", Lane: LaneAgriculture, Cost: 22000, ResearchTicks: 1700,
			Prerequisites: []string{"military_tactics"},
			Description:   "Feudal land grants house more workers.",
			Effects: []TechEffect{
				{Kind: EffectFlatHousing, Value: 5},
			},
		},
		{
			Name: "Alchemy", Key: "alchemy", Emblem: "☿",
			Age: "medieval_age", Lane: LaneKnowledge, Cost: 28000, ResearchTicks: 2200,
			Prerequisites: []string{"mathematics"},
			Description:   "Proto-chemistry yields material insights.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.15},
				{Kind: EffectFlatOutput, Target: "gold", Value: 0.1},
			},
		},
		{
			Name: "Chronometry", Key: "chronometry", Emblem: "⊙",
			Age: "medieval_age", Lane: LaneCraft, Cost: 20000, ResearchTicks: 1900,
			Description: "Precise timekeeping raises game speed.",
			Effects: []TechEffect{
				{Kind: EffectGameSpeed, Value: 0.05},
			},
		},

		// === RENAISSANCE AGE === (~10 min each)
		{
			Name: "Printing Press", Key: "printing_press",
			Age: "renaissance_age", Lane: LaneKnowledge, Cost: 50000, ResearchTicks: 3000,
			Prerequisites: []string{"theology", "alchemy"},
			Description:   "Printed books raise knowledge output and culture.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.4},
				{Kind: EffectFlatOutput, Target: "culture", Value: 0.3},
			},
		},
		{
			Name: "Navigation", Key: "navigation",
			Age: "renaissance_age", Lane: LaneTrade, Cost: 45000, ResearchTicks: 2600,
			Prerequisites: []string{"mathematics", "road_building"},
			Description:   "Ocean navigation raises gold output and expedition rewards.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.5},
				{Kind: EffectExpeditionReward, Value: 0.3},
			},
		},
		{
			Name: "Gunpowder", Key: "gunpowder",
			Age: "renaissance_age", Lane: LaneMilitary, Cost: 55000, ResearchTicks: 3200,
			Prerequisites: []string{"alchemy", "siege_warfare"},
			Description:   "Explosive weapons raise military power.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.5},
			},
		},
		{
			Name: "Patronage", Key: "patronage",
			Age: "renaissance_age", Lane: LaneFaith, Cost: 40000, ResearchTicks: 2500,
			Prerequisites: []string{"banking"},
			Description:   "Wealthy patrons fund arts and science.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "culture", Value: 0.5},
				{Kind: EffectFlatOutput, Target: "knowledge", Value: 0.12},
			},
		},

		// === COLONIAL AGE === (~14 min each)
		{
			Name: "Cartography", Key: "cartography",
			Age: "colonial_age", Lane: LaneTrade, Cost: 80000, ResearchTicks: 4000,
			Prerequisites: []string{"navigation"},
			Description:   "Detailed maps raise expedition rewards and gold output.",
			Effects: []TechEffect{
				{Kind: EffectExpeditionReward, Value: 0.5},
				{Kind: EffectOutput, Target: "gold", Value: 0.5},
			},
		},
		{
			Name: "Mercantilism", Key: "mercantilism",
			Age: "colonial_age", Lane: LaneTrade, Cost: 75000, ResearchTicks: 3800,
			Prerequisites: []string{"banking", "navigation"},
			Description:   "National trade policies maximize wealth.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "gold", Value: 2.0},
				{Kind: EffectOutput, Target: "gold", Value: 0.3},
			},
		},
		{
			Name: "Colonialism", Key: "colonialism",
			Age: "colonial_age", Lane: LaneMilitary, Cost: 90000, ResearchTicks: 4400,
			Prerequisites: []string{"cartography", "gunpowder"},
			Description:   "Overseas territorial expansion.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "food", Value: 2.0},
				{Kind: EffectMilitaryPower, Value: 0.3},
			},
		},

		// === INDUSTRIAL AGE === (~18 min each)
		{
			Name: "Steam Power", Key: "steam_power", Emblem: "≈",
			Age: "industrial_age", Lane: LaneEnergy, Cost: 100000, ResearchTicks: 5200,
			Prerequisites: []string{"steel_forging"},
			Description:   "Steam engines raise all production.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.3},
			},
		},
		{
			Name: "Industrialization", Key: "industrialization",
			Age: "industrial_age", Lane: LaneCraft, Cost: 120000, ResearchTicks: 6000,
			Prerequisites: []string{"steam_power"},
			Description:   "Factory systems raise all production and add steel.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.5},
				{Kind: EffectFlatOutput, Target: "steel", Value: 0.5},
			},
		},
		{
			Name: "Railroads", Key: "railroads",
			Age: "industrial_age", Lane: LaneTrade, Cost: 90000, ResearchTicks: 5000,
			Prerequisites: []string{"steam_power", "road_building"},
			Description:   "Rail networks connect your civilization.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 1.0},
				{Kind: EffectFlatStorage, Target: AllResources, Value: 200},
			},
		},
		{
			Name: "Rifling", Key: "rifling",
			Age: "industrial_age", Lane: LaneMilitary, Cost: 80000, ResearchTicks: 4800,
			Prerequisites: []string{"gunpowder"},
			Description:   "Precision firearms improve military effectiveness.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.5},
			},
		},
		{
			Name: "Clockwork Automation", Key: "clockwork_automation",
			Age: "industrial_age", Lane: LaneCraft, Cost: 50000, ResearchTicks: 5400,
			Prerequisites: []string{"chronometry"},
			Description:   "Mechanical automation raises game speed.",
			Effects: []TechEffect{
				{Kind: EffectGameSpeed, Value: 0.10},
			},
		},

		// === VICTORIAN AGE === (~23 min each)
		{
			Name: "Electrification", Key: "electrification", Code: "ELEC", Emblem: "ϟ",
			Age: "victorian_age", Lane: LaneEnergy, Cost: 180000, ResearchTicks: 7000,
			Prerequisites: []string{"industrialization"},
			Description:   "Electric power reaches homes and factories.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "electricity", Value: 1.0},
				{Kind: EffectAllOutput, Value: 0.2},
			},
		},
		{
			Name: "Telecommunications", Key: "telecommunications", Code: "TELEG", Emblem: "∿",
			Age: "victorian_age", Lane: LaneTrade, Cost: 150000, ResearchTicks: 6600,
			Prerequisites: []string{"electrification"},
			Description:   "Telegraph and early telephone networks.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.4},
				{Kind: EffectOutput, Target: "gold", Value: 0.5},
			},
		},
		{
			Name: "Mass Production", Key: "mass_production", Emblem: "▥",
			Age: "victorian_age", Lane: LaneCraft, Cost: 200000, ResearchTicks: 7400,
			Prerequisites: []string{"industrialization", "railroads"},
			Description:   "Assembly line manufacturing.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.4},
				{Kind: EffectFlatOutput, Target: "steel", Value: 1.0},
			},
		},

		// === ELECTRIC AGE === (~32 min each)
		{
			Name: "Power Distribution", Key: "power_distribution", Code: "GRID",
			Age: "electric_age", Lane: LaneEnergy, Cost: 300000, ResearchTicks: 9500,
			Prerequisites: []string{"electrification"},
			Description:   "AC power grids span entire regions.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "electricity", Value: 3.0},
				{Kind: EffectAllOutput, Value: 0.3},
			},
		},
		{
			Name: "Radio", Key: "radio", Emblem: "♪",
			Age: "electric_age", Lane: LaneFaith, Cost: 250000, ResearchTicks: 9000,
			Prerequisites: []string{"telecommunications"},
			Description:   "Wireless communication reaches the masses.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "culture", Value: 2.0},
				{Kind: EffectOutput, Target: "knowledge", Value: 0.4},
			},
		},
		{
			Name: "Chemical Engineering", Key: "chemical_engineering", Code: "CHEM", Emblem: "∆",
			Age: "electric_age", Lane: LaneMaterials, Cost: 280000, ResearchTicks: 9800,
			Prerequisites: []string{"mass_production"},
			Description:   "Industrial chemistry and synthetic materials.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "oil", Value: 1.0},
				{Kind: EffectAllOutput, Value: 0.2},
			},
		},

		// === ATOMIC AGE === (~45 min each)
		{
			Name: "Nuclear Fission", Key: "nuclear_fission", Code: "FISSN", Emblem: "◉",
			Age: "atomic_age", Lane: LaneEnergy, Cost: 500000, ResearchTicks: 13050,
			Prerequisites: []string{"power_distribution", "chemical_engineering"},
			Description:   "Splitting the atom for energy and weapons.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "electricity", Value: 5.0},
				{Kind: EffectFlatOutput, Target: "uranium", Value: 0.5},
			},
		},
		{
			Name: "Rocketry", Key: "rocketry", Code: "ROCKT", Emblem: "▲",
			Age: "atomic_age", Lane: LaneSpace, Cost: 400000, ResearchTicks: 12000,
			Prerequisites: []string{"rifling", "chemical_engineering"},
			Description:   "Rockets raise military power and expedition rewards.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 1.0},
				{Kind: EffectExpeditionReward, Value: 0.5},
			},
		},
		{
			Name: "Nuclear Deterrence", Key: "nuclear_deterrence", Code: "DETER", Emblem: "☠",
			Age: "atomic_age", Lane: LaneMilitary, Cost: 600000, ResearchTicks: 15000,
			Prerequisites: []string{"nuclear_fission", "rocketry"},
			Description:   "Mutually assured destruction maintains peace.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 1.5},
			},
		},
		{
			// Mid-age unlock (Pacing v2): the Atomic Age's techs were done
			// in its first third, then nothing new for most of a day. The
			// cost is what holds it to the middle: knowledge runs about
			// 200 million an hour here, so the bot affords it some 15 hours
			// in and finishes it near the middle of the old quiet stretch.
			Name: "Civilian Reactors", Key: "civilian_reactors", Code: "REACT", Emblem: "▣",
			Age: "atomic_age", Lane: LaneEnergy, Cost: 3400000000, ResearchTicks: 15000,
			Prerequisites: []string{"nuclear_deterrence"},
			Description:   "The reactors built for the arms race find steadier work on the grid. Opens the Nuclear Plant.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "electricity", Value: 5.0},
				{Kind: EffectFlatOutput, Target: "uranium", Value: 0.5},
			},
		},

		// === MODERN AGE === (~1.1 hr each)
		{
			Name: "Advanced Electrics", Key: "electricity_tech", Code: "ADVEL", Emblem: "ϟ",
			Age: "modern_age", Lane: LaneEnergy, Cost: 800000, ResearchTicks: 18000,
			Prerequisites: []string{"nuclear_fission"},
			Description:   "Advanced electrical systems raise all production.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.5},
				{Kind: EffectFlatOutput, Target: "electricity", Value: 5.0},
			},
		},
		{
			Name: "Computers", Key: "computers", Code: "COMP",
			Age: "modern_age", Lane: LaneComputing, Cost: 1000000, ResearchTicks: 20000,
			Prerequisites: []string{"electricity_tech"},
			Description:   "Digital computing raises knowledge output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.8},
			},
		},
		{
			Name: "Satellite Technology", Key: "satellite_tech", Emblem: "✧",
			Age: "modern_age", Lane: LaneSpace, Cost: 1200000, ResearchTicks: 19000,
			Prerequisites: []string{"rocketry", "electricity_tech"},
			Description:   "Orbital satellites for communication and surveillance.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 1.0},
				{Kind: EffectOutput, Target: "knowledge", Value: 0.6},
			},
		},
		{
			Name: "Nanofabrication", Key: "nanofabrication", Code: "NANO", Emblem: "◇",
			Age: "modern_age", Lane: LaneCraft, Cost: 1100000, ResearchTicks: 19000,
			Prerequisites: []string{"computers"},
			Description:   "Nanobot swarms assemble structures atom-by-atom, cutting construction costs.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.08},
			},
		},

		// === INFORMATION AGE === (~1.5 hr each)
		{
			Name: "Internet", Key: "internet",
			Age: "information_age", Lane: LaneComputing, Cost: 2000000, ResearchTicks: 26000,
			Prerequisites: []string{"computers", "satellite_tech"},
			Description:   "Global network connecting all of humanity.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 3.0},
				{Kind: EffectOutput, Target: "knowledge", Value: 1.2},
			},
		},
		{
			Name: "Cybersecurity", Key: "cybersecurity", Code: "SECUR",
			Age: "information_age", Lane: LaneMilitary, Cost: 1800000, ResearchTicks: 24000,
			Prerequisites: []string{"computers"},
			Description:   "Defense against digital threats.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 1.0},
				{Kind: EffectFlatStorage, Target: "data", Value: 5000},
			},
		},
		{
			Name: "Social Media", Key: "social_media",
			Age: "information_age", Lane: LaneFaith, Cost: 1500000, ResearchTicks: 23000,
			Prerequisites: []string{"internet"},
			Description:   "Mass digital communication platforms.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "culture", Value: 5.0},
				{Kind: EffectFlatOutput, Target: "gold", Value: 5.0},
			},
		},
		{
			// note: the original spec wanted this to cut worker FOOD COST, but the
			// food drain (wc.FoodCost * count in game/villagers.go) has no bonus hook
			// and threading one through WorkerManager isn't a "tiny" engine change.
			// Substituted a supported, clearly-beneficial effect instead: nanobots
			// keep the population healthier (bigger pop cap) and better fed (+food).
			Name: "Medical Nanobots", Key: "medical_nanobots",
			Age: "information_age", Lane: LaneAgriculture, Cost: 1700000, ResearchTicks: 24000,
			Prerequisites: []string{"nanofabrication"},
			Description:   "Bloodstream nanobots keep workers healthy, adding housing and food.",
			Effects: []TechEffect{
				{Kind: EffectFlatHousing, Value: 10},
				{Kind: EffectFlatOutput, Target: "food", Value: 8.0},
			},
		},
		{
			// Mid-age unlock (Pacing v2): knowledge runs about 280 million
			// an hour here, so it is afforded some 15 hours in and finishes
			// about 18 hours in, after the age's first techs.
			Name: "Internet of Things", Key: "internet_of_things", Code: "IOT",
			Age: "information_age", Lane: LaneAgriculture, Cost: 4500000000, ResearchTicks: 26000,
			Prerequisites: []string{"social_media", "cybersecurity", "medical_nanobots"},
			Description:   "The fridges and the tractors go online and start reporting back. Opens the Smart Farm and the Smart Complex.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 3.0},
				{Kind: EffectFlatOutput, Target: "food", Value: 8.0},
			},
		},

		// === DIGITAL AGE === (~1.8 hr each)
		{
			Name: "Machine Learning", Key: "machine_learning",
			Age: "digital_age", Lane: LaneComputing, Cost: 3500000, ResearchTicks: 34000,
			Prerequisites: []string{"internet", "cybersecurity"},
			Description:   "Algorithms that learn and improve autonomously.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 5.0},
				{Kind: EffectAllOutput, Value: 0.5},
			},
		},
		{
			Name: "Cloud Computing", Key: "cloud_computing",
			Age: "digital_age", Lane: LaneComputing, Cost: 3000000, ResearchTicks: 32000,
			Prerequisites: []string{"internet"},
			Description:   "Distributed computing at global scale.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 8.0},
				{Kind: EffectFlatStorage, Target: AllResources, Value: 10000},
			},
		},
		{
			Name: "Self-Replication", Key: "self_replication",
			Age: "digital_age", Lane: LaneCraft, Cost: 3200000, ResearchTicks: 33000,
			Prerequisites: []string{"medical_nanobots", "machine_learning"},
			Description:   "Nanobots that build copies of themselves.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "nanobots", Value: 200.0},
			},
		},

		// === CYBERPUNK AGE === (~2.8 hr each)
		{
			Name: "Neural Interface", Key: "neural_interface",
			Age: "cyberpunk_age", Lane: LaneKnowledge, Cost: 6000000, ResearchTicks: 48000,
			Prerequisites: []string{"machine_learning"},
			Description:   "Direct brain-computer interface technology.",
			Effects: []TechEffect{
				{Kind: EffectWorkerOutput, Value: 0.3},
				{Kind: EffectOutput, Target: "knowledge", Value: 2.0},
			},
		},
		{
			Name: "Blockchain", Key: "blockchain",
			Age: "cyberpunk_age", Lane: LaneTrade, Cost: 5000000, ResearchTicks: 45000,
			Prerequisites: []string{"cybersecurity", "cloud_computing"},
			Description:   "Decentralized trustless systems.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "crypto", Value: 2.0},
				{Kind: EffectOutput, Target: "gold", Value: 2.0},
			},
		},
		{
			// Pacing v2: priced to finish about 26 hours into the age
			// (knowledge runs about 450 million an hour here), with
			// Holography after it near the end. The Cyberpunk Age's techs
			// used to be done in its first 11 hours.
			Name: "Cybernetics", Key: "cybernetics",
			Age: "cyberpunk_age", Lane: LaneCraft, Cost: 14000000000, ResearchTicks: 50000,
			Prerequisites: []string{"neural_interface"},
			Description:   "Mechanical augmentation of the human body.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.5},
				{Kind: EffectMilitaryPower, Value: 1.0},
			},
		},
		{
			// Mid-age unlock (Pacing v2), paced with Cybernetics: afforded
			// about 14 hours after Cybernetics starts, so it finishes about
			// 37 hours in, during the age's long saving-up for its wonder.
			Name: "Holography", Key: "holography",
			Age: "cyberpunk_age", Lane: LaneFaith, Cost: 6300000000, ResearchTicks: 50000,
			Prerequisites: []string{"cybernetics", "blockchain"},
			Description:   "Light learns to lie convincingly, and every wall becomes an ad. Opens the Holographic Theater.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "culture", Value: 5.0},
				{Kind: EffectFlatOutput, Target: "crypto", Value: 2.0},
			},
		},

		// === FUSION AGE === (~3.7 hr each)
		{
			Name: "Fusion Power", Key: "fusion_power",
			Age: "fusion_age", Lane: LaneEnergy, Cost: 10000000, ResearchTicks: 65000,
			Prerequisites: []string{"nuclear_fission", "cybernetics"},
			Description:   "Controlled fusion adds electricity and plasma.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "electricity", Value: 20.0},
				{Kind: EffectFlatOutput, Target: "plasma", Value: 1.0},
			},
		},
		{
			// Pacing v2: Plasma Physics, Superconductors and Maglev Transit
			// now run one after another (each needs the one before) and are
			// priced in knowledge, which runs about 600 million an hour in
			// this age, to finish about 17, 24 and 32 hours in. The age used
			// to go 26 hours from its last tech to its wonder.
			Name: "Plasma Physics", Key: "plasma_physics",
			Age: "fusion_age", Lane: LaneEnergy, Cost: 9500000000, ResearchTicks: 62000,
			Prerequisites: []string{"fusion_power"},
			Description:   "Mastery of superheated matter states.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "plasma", Value: 3.0},
				{Kind: EffectAllOutput, Value: 0.3},
			},
		},
		{
			// Paced with Plasma Physics (see there).
			Name: "Superconductors", Key: "superconductors",
			Age: "fusion_age", Lane: LaneMaterials, Cost: 4800000000, ResearchTicks: 70000,
			Prerequisites: []string{"plasma_physics"},
			Description:   "Zero-resistance materials raise all production and storage.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.5},
				{Kind: EffectFlatStorage, Target: AllResources, Value: 50000},
			},
		},
		{
			// Mid-age unlock (Pacing v2), paced with Plasma Physics (see
			// there).
			Name: "Maglev Transit", Key: "maglev_transit",
			Age: "fusion_age", Lane: LaneTrade, Cost: 5400000000, ResearchTicks: 70000,
			Prerequisites: []string{"superconductors"},
			Description:   "Superconducting rails float the freight across the city at the speed of a mild panic. Opens the Energy Exchange.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "plasma", Value: 1.0},
				{Kind: EffectFlatOutput, Target: "gold", Value: 5.0},
			},
		},

		// === SPACE AGE === (~5 hr each)
		{
			Name: "Orbital Mechanics", Key: "orbital_mechanics",
			Age: "space_age", Lane: LaneSpace, Cost: 20000000, ResearchTicks: 85000,
			Prerequisites: []string{"rocketry", "plasma_physics"},
			Description:   "Advanced spaceflight and orbital dynamics.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "titanium", Value: 1.0},
				{Kind: EffectExpeditionReward, Value: 1.0},
			},
		},
		{
			Name: "Space Mining", Key: "space_mining",
			Age: "space_age", Lane: LaneSpace, Cost: 18000000, ResearchTicks: 82000,
			Prerequisites: []string{"orbital_mechanics"},
			Description:   "Asteroid and lunar resource extraction.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "titanium", Value: 3.0},
				{Kind: EffectFlatOutput, Target: "iron", Value: 20.0},
			},
		},
		{
			Name: "Zero-G Manufacturing", Key: "zero_g_manufacturing",
			Age: "space_age", Lane: LaneCraft, Cost: 22000000, ResearchTicks: 92000,
			Prerequisites: []string{"orbital_mechanics", "superconductors"},
			Description:   "Space-based manufacturing for perfect materials.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.5},
				{Kind: EffectFlatOutput, Target: "steel", Value: 10.0},
			},
		},

		// === INTERSTELLAR AGE === (~7 hr each)
		{
			Name: "Warp Drive", Key: "warp_drive",
			Age: "interstellar_age", Lane: LaneSpace, Cost: 40000000, ResearchTicks: 120000,
			Prerequisites: []string{"space_mining", "zero_g_manufacturing"},
			Description:   "Faster-than-light propulsion.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "dark_matter", Value: 1.0},
				{Kind: EffectExpeditionReward, Value: 2.0},
			},
		},
		{
			Name: "Stellar Engineering", Key: "stellar_engineering",
			Age: "interstellar_age", Lane: LaneEnergy, Cost: 45000000, ResearchTicks: 130000,
			Prerequisites: []string{"warp_drive"},
			Description:   "Harnessing and shaping stars themselves.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "plasma", Value: 10.0},
				{Kind: EffectFlatOutput, Target: "electricity", Value: 100.0},
			},
		},

		// === GALACTIC AGE === (~9 hr each)
		{
			Name: "Galactic Navigation", Key: "galactic_navigation",
			Age: "galactic_age", Lane: LaneSpace, Cost: 80000000, ResearchTicks: 160000,
			Prerequisites: []string{"warp_drive", "stellar_engineering"},
			Description:   "Charting paths across the galaxy.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.5},
				{Kind: EffectFlatOutput, Target: "dark_matter", Value: 5.0},
			},
		},
		{
			Name: "Antimatter Synthesis", Key: "antimatter_synthesis",
			Age: "galactic_age", Lane: LaneEnergy, Cost: 90000000, ResearchTicks: 180000,
			Prerequisites: []string{"galactic_navigation"},
			Description:   "Controlled production of antimatter.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "antimatter", Value: 2.0},
				{Kind: EffectAllOutput, Value: 0.3},
			},
		},

		// === QUANTUM AGE === (~12 hr each)
		{
			Name: "Quantum Mechanics", Key: "quantum_mechanics",
			Age: "quantum_age", Lane: LaneKnowledge, Cost: 150000000, ResearchTicks: 220000,
			Prerequisites: []string{"antimatter_synthesis"},
			Description:   "Mastery of quantum phenomena at all scales.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "quantum_flux", Value: 2.0},
				{Kind: EffectAllOutput, Value: 1.0},
			},
		},
		{
			Name: "Reality Manipulation", Key: "reality_manipulation",
			Age: "quantum_age", Lane: LaneCraft, Cost: 200000000, ResearchTicks: 250000,
			Prerequisites: []string{"quantum_mechanics"},
			Description:   "Bending the fabric of spacetime.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "quantum_flux", Value: 5.0},
				{Kind: EffectAllOutput, Value: 1.0},
			},
		},
		{
			Name: "Quantum Computing", Key: "quantum_computing", Code: "QCOMP",
			Age: "quantum_age", Lane: LaneComputing, Cost: 150000000, ResearchTicks: 200000,
			Prerequisites: []string{"clockwork_automation"},
			Description:   "Quantum processing raises game speed.",
			Effects: []TechEffect{
				{Kind: EffectGameSpeed, Value: 0.15},
			},
		},

		// === TRANSCENDENT AGE === (~18 hr)
		{
			Name: "Transcendence", Key: "transcendence",
			Age: "transcendent_age", Lane: LaneFaith, Cost: 500000000, ResearchTicks: 320000,
			Prerequisites: []string{"reality_manipulation"},
			Description:   "A civilization beyond physical limits.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 2.0},
				{Kind: EffectFlatOutput, Target: "quantum_flux", Value: 10.0},
			},
		},
	}))
}

// TechByKey returns a map of key -> TechDef
func TechByKey() map[string]TechDef {
	m := make(map[string]TechDef)
	for _, t := range Technologies() {
		m[t.Key] = t
	}
	return m
}

// TechsByAge groups all tech definitions by their minimum age key.
// Used by the research overlay to display techs in age buckets.
func TechsByAge() map[string][]TechDef {
	m := make(map[string][]TechDef)
	for _, t := range Technologies() {
		m[t.Age] = append(m[t.Age], t)
	}
	return m
}
