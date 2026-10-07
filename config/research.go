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
	// Cost is the knowledge paid when research starts. It is not typed on a
	// tech: normalizeResearchCosts sets it from the age's knowledge budget
	// and the tech's kind (pacing.go).
	Cost float64
	// Prerequisites are the tech keys that must all be researched first.
	Prerequisites []string
	// AnyOf is an either-or group on top of Prerequisites: at least one of
	// these must be researched too. Empty means no such group. A tech has
	// one group at most, of two keys or more.
	AnyOf []string
	// Capstone marks a tech that ends its lane for an era.
	Capstone    bool
	Effects     []TechEffect // applied permanently when research finishes
	Description string
	// ResearchTicks is the game ticks research takes at 1x, before research
	// speed and Era Mastery. It is not typed on a tech either:
	// normalizeResearchTicks sets it from the age's research cap and the
	// tech's kind (pacing.go).
	ResearchTicks int
}

// Technologies returns every tech definition, ordered loosely by age.
// Use TechByKey() for random access or TechsByAge() to group by age.
//
// No tech carries a cost or a research time here. Both follow from the tech's
// age and its kind (pacing.go: normalizeResearchCosts, normalizeResearchTicks),
// and its kind follows from the wonders (TechKinds), so moving a keystone or
// adding a tech re-prices its age with nothing to retype.
func Technologies() []TechDef {
	techs := []TechDef{
		// === PRIMITIVE AGE ===
		{
			Name: "Tool Making", Key: "tool_making", Code: "TOOLS",
			Age: "primitive_age", Lane: LaneCraft,
			Description: "Stone tools make every worker more productive.",
			Effects: []TechEffect{
				{Kind: EffectWorkerOutput, Value: 0.15},
			},
		},
		{
			// A root, with Tool Making: fire needs no tools.
			Name: "Fire Mastery", Key: "fire_mastery", Emblem: "△",
			Age: "primitive_age", Lane: LaneAgriculture,
			Description: "Control of fire improves food preservation and warmth.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "food", Value: 0.1},
			},
		},

		// === STONE AGE ===
		{
			Name: "Stoneworking", Key: "stoneworking", Emblem: "◆",
			Age: "stone_age", Lane: LaneMaterials,
			Prerequisites: []string{"tool_making"},
			Description:   "Cutting and shaping stone for construction.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "stone", Value: 0.2},
			},
		},
		{
			Name: "Animal Husbandry", Key: "animal_husbandry", Code: "HERDS", Emblem: "♞",
			Age: "stone_age", Lane: LaneAgriculture,
			Prerequisites: []string{"fire_mastery"},
			Description:   "Domesticating animals for food and labor.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "food", Value: 0.2},
			},
		},
		{
			Name: "Pottery", Key: "pottery", Code: "POTS", Emblem: "∪",
			Age: "stone_age", Lane: LaneTrade,
			Prerequisites: []string{"fire_mastery"},
			Description:   "Clay vessels for storage and trade.",
			Effects: []TechEffect{
				{Kind: EffectFlatStorage, Target: AllResources, Value: 25},
			},
		},
		{
			// Writing no longer follows Pottery, which leads to storage and trade.
			// It is a root until the tech it will follow (Language) exists.
			Name: "Primitive Writing", Key: "primitive_writing", Code: "WRITE",
			Age: "stone_age", Lane: LaneKnowledge,
			Description: "Early symbols enable knowledge transfer.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.1},
			},
		},

		// === BRONZE AGE ===
		{
			Name: "Bronze Working", Key: "bronze_working",
			Age: "bronze_age", Lane: LaneMaterials,
			Prerequisites: []string{"stoneworking"},
			Description:   "Alloying copper and tin creates durable tools.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "stone", Value: 0.2},
				{Kind: EffectWorkerOutput, Value: 0.1},
			},
		},
		{
			Name: "Agriculture", Key: "agriculture", Code: "FARMS",
			Age: "bronze_age", Lane: LaneAgriculture,
			Prerequisites: []string{"animal_husbandry"},
			Description:   "Systematic farming adds steady food output.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "food", Value: 0.5},
			},
		},
		{
			Name: "Currency", Key: "currency", Code: "COIN", Emblem: "¤",
			Age: "bronze_age", Lane: LaneTrade,
			Prerequisites: []string{"primitive_writing"},
			Description:   "Standardized money raises gold output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.3},
			},
		},
		{
			Name: "Masonry", Key: "masonry", Emblem: "▦",
			Age: "bronze_age", Lane: LaneCraft,
			Prerequisites: []string{"stoneworking"},
			Description:   "Advanced stone construction techniques.",
			Effects: []TechEffect{
				{Kind: EffectFlatStorage, Target: AllResources, Value: 50},
			},
		},
		{
			Name: "Military Tactics", Key: "military_tactics", Code: "TACTI",
			Age: "bronze_age", Lane: LaneMilitary,
			Prerequisites: []string{"bronze_working"},
			Description:   "Organized warfare and defense strategies.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.2},
			},
		},

		// === IRON AGE ===
		{
			Name: "Iron Smelting", Key: "iron_smelting", Emblem: "■",
			Age: "iron_age", Lane: LaneMaterials,
			Prerequisites: []string{"bronze_working"},
			Description:   "Hotter furnaces raise iron output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "iron", Value: 0.4},
				{Kind: EffectFlatOutput, Target: "iron", Value: 0.2},
			},
		},
		{
			// Roads will need The Wheel too, once that tech exists.
			Name: "Road Building", Key: "road_building", Code: "ROADS", Emblem: "═",
			Age: "iron_age", Lane: LaneCraft,
			Prerequisites: []string{"masonry"},
			Description:   "Paved roads improve trade and movement.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.2},
				{Kind: EffectWorkerOutput, Value: 0.1},
			},
		},
		{
			Name: "Mathematics", Key: "mathematics", Code: "MATH", Emblem: "π",
			Age: "iron_age", Lane: LaneKnowledge,
			Prerequisites: []string{"primitive_writing"},
			Description:   "Advanced calculation raises knowledge output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.2},
			},
		},
		{
			Name: "Siege Warfare", Key: "siege_warfare", Emblem: "✕",
			Age: "iron_age", Lane: LaneMilitary,
			Prerequisites: []string{"military_tactics"},
			Description:   "Siege engines and fortification techniques.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.3},
			},
		},

		// === CLASSICAL AGE ===
		{
			Name: "Philosophy", Key: "philosophy", Emblem: "Φ",
			Age: "classical_age", Lane: LaneKnowledge,
			Prerequisites: []string{"mathematics"},
			Description:   "Systematic inquiry into fundamental questions.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.3},
				{Kind: EffectFlatOutput, Target: "culture", Value: 0.2},
			},
		},
		{
			Name: "Civil Engineering", Key: "civil_engineering", Emblem: "∩",
			Age: "classical_age", Lane: LaneCraft,
			Prerequisites: []string{"road_building"},
			Description:   "Large-scale construction and infrastructure.",
			Effects: []TechEffect{
				{Kind: EffectFlatStorage, Target: AllResources, Value: 100},
				{Kind: EffectBuildCost, Value: -0.05},
			},
		},
		{
			Name: "Imperial Legions", Key: "imperial_legions", Code: "LEGIO", Emblem: "⚑",
			Age: "classical_age", Lane: LaneMilitary,
			Prerequisites: []string{"siege_warfare", "iron_smelting"},
			Description:   "Professional standing armies with superior discipline.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.4},
			},
		},

		// === MEDIEVAL AGE ===
		{
			Name: "Steel Forging", Key: "steel_forging", Emblem: "▣",
			Age: "medieval_age", Lane: LaneMaterials,
			Prerequisites: []string{"iron_smelting"},
			Description:   "Refining iron into steel for superior tools and weapons.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "steel", Value: 0.25},
				{Kind: EffectOutput, Target: "iron", Value: 0.3},
			},
		},
		{
			Name: "Theology", Key: "theology", Emblem: "Θ",
			Age: "medieval_age", Lane: LaneFaith,
			Prerequisites: []string{"philosophy"},
			Description:   "Organized religion provides faith and social cohesion.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "faith", Value: 0.3},
			},
		},
		{
			Name: "Banking", Key: "banking", Code: "BANK",
			Age: "medieval_age", Lane: LaneTrade,
			Prerequisites: []string{"currency", "mathematics"},
			Description:   "Financial institutions raise gold output and gold storage.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.5},
				{Kind: EffectFlatStorage, Target: "gold", Value: 100},
			},
		},
		{
			// No longer behind Military Tactics: it is a farming tech. It is a root
			// until the tech it will follow (The Plough) exists.
			Name: "Feudalism", Key: "feudalism", Emblem: "⌂",
			Age: "medieval_age", Lane: LaneAgriculture,
			Description: "Feudal land grants house more workers.",
			Effects: []TechEffect{
				{Kind: EffectFlatHousing, Value: 5},
			},
		},
		{
			Name: "Alchemy", Key: "alchemy", Emblem: "☿",
			Age: "medieval_age", Lane: LaneKnowledge,
			Prerequisites: []string{"philosophy"},
			Description:   "Proto-chemistry yields material insights.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.15},
				{Kind: EffectFlatOutput, Target: "gold", Value: 0.1},
			},
		},
		{
			Name: "Chronometry", Key: "chronometry", Emblem: "⊙",
			Age: "medieval_age", Lane: LaneCraft,
			Description: "Precise timekeeping raises game speed.",
			Effects: []TechEffect{
				{Kind: EffectGameSpeed, Value: 0.05},
			},
		},

		// === RENAISSANCE AGE ===
		{
			Name: "Printing Press", Key: "printing_press",
			Age: "renaissance_age", Lane: LaneKnowledge,
			Prerequisites: []string{"theology", "alchemy"},
			Description:   "Printed books raise knowledge output and culture.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.4},
				{Kind: EffectFlatOutput, Target: "culture", Value: 0.3},
			},
		},
		{
			// It will need Exploration too, once that tech exists.
			Name: "Navigation", Key: "navigation",
			Age: "renaissance_age", Lane: LaneTrade,
			Prerequisites: []string{"mathematics"},
			Description:   "Ocean navigation raises gold output and expedition rewards.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.5},
				{Kind: EffectExpeditionReward, Value: 0.3},
			},
		},
		{
			Name: "Gunpowder", Key: "gunpowder",
			Age: "renaissance_age", Lane: LaneMilitary,
			Prerequisites: []string{"alchemy", "siege_warfare"},
			Description:   "Explosive weapons raise military power.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.5},
			},
		},
		{
			Name: "Patronage", Key: "patronage",
			Age: "renaissance_age", Lane: LaneFaith,
			Prerequisites: []string{"banking"},
			Description:   "Wealthy patrons fund arts and science.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "culture", Value: 0.5},
				{Kind: EffectFlatOutput, Target: "knowledge", Value: 0.12},
			},
		},

		// === COLONIAL AGE ===
		{
			Name: "Cartography", Key: "cartography",
			Age: "colonial_age", Lane: LaneTrade,
			Prerequisites: []string{"navigation"},
			Description:   "Detailed maps raise expedition rewards and gold output.",
			Effects: []TechEffect{
				{Kind: EffectExpeditionReward, Value: 0.5},
				{Kind: EffectOutput, Target: "gold", Value: 0.5},
			},
		},
		{
			Name: "Mercantilism", Key: "mercantilism",
			Age: "colonial_age", Lane: LaneTrade,
			Prerequisites: []string{"banking", "navigation"},
			Description:   "National trade policies maximize wealth.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "gold", Value: 2.0},
				{Kind: EffectOutput, Target: "gold", Value: 0.3},
			},
		},
		{
			Name: "Colonialism", Key: "colonialism",
			Age: "colonial_age", Lane: LaneMilitary,
			Prerequisites: []string{"cartography", "gunpowder"},
			Description:   "Overseas territorial expansion.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "food", Value: 2.0},
				{Kind: EffectMilitaryPower, Value: 0.3},
			},
		},

		// === INDUSTRIAL AGE ===
		{
			Name: "Steam Power", Key: "steam_power", Emblem: "≈",
			Age: "industrial_age", Lane: LaneEnergy,
			Prerequisites: []string{"steel_forging"},
			Description:   "Steam engines raise all production.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.3},
			},
		},
		{
			Name: "Industrialization", Key: "industrialization",
			Age: "industrial_age", Lane: LaneCraft,
			Prerequisites: []string{"steam_power"},
			Description:   "Factory systems raise all production and add steel.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.5},
				{Kind: EffectFlatOutput, Target: "steel", Value: 0.5},
			},
		},
		{
			Name: "Railroads", Key: "railroads",
			Age: "industrial_age", Lane: LaneTrade,
			Prerequisites: []string{"steam_power", "road_building"},
			Description:   "Rail networks connect your civilization.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 1.0},
				{Kind: EffectFlatStorage, Target: AllResources, Value: 200},
			},
		},
		{
			Name: "Rifling", Key: "rifling",
			Age: "industrial_age", Lane: LaneMilitary,
			Prerequisites: []string{"gunpowder"},
			Description:   "Precision firearms improve military effectiveness.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.5},
			},
		},
		{
			Name: "Clockwork Automation", Key: "clockwork_automation",
			Age: "industrial_age", Lane: LaneCraft,
			Prerequisites: []string{"chronometry"},
			Description:   "Mechanical automation raises game speed.",
			Effects: []TechEffect{
				{Kind: EffectGameSpeed, Value: 0.10},
			},
		},

		// === VICTORIAN AGE ===
		{
			Name: "Electrification", Key: "electrification", Code: "ELEC", Emblem: "ϟ",
			Age: "victorian_age", Lane: LaneEnergy,
			Prerequisites: []string{"industrialization"},
			Description:   "Electric power reaches homes and factories.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "electricity", Value: 1.0},
				{Kind: EffectAllOutput, Value: 0.2},
			},
		},
		{
			Name: "Telecommunications", Key: "telecommunications", Code: "TELEG", Emblem: "∿",
			Age: "victorian_age", Lane: LaneTrade,
			Prerequisites: []string{"electrification"},
			Description:   "Telegraph and early telephone networks.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.4},
				{Kind: EffectOutput, Target: "gold", Value: 0.5},
			},
		},
		{
			Name: "Mass Production", Key: "mass_production", Emblem: "▥",
			Age: "victorian_age", Lane: LaneCraft,
			Prerequisites: []string{"industrialization"},
			Description:   "Assembly line manufacturing.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.4},
				{Kind: EffectFlatOutput, Target: "steel", Value: 1.0},
			},
		},

		// === ELECTRIC AGE ===
		{
			Name: "Power Distribution", Key: "power_distribution", Code: "GRID",
			Age: "electric_age", Lane: LaneEnergy,
			Prerequisites: []string{"electrification"},
			Description:   "AC power grids span entire regions.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "electricity", Value: 3.0},
				{Kind: EffectAllOutput, Value: 0.3},
			},
		},
		{
			Name: "Radio", Key: "radio", Emblem: "♪",
			Age: "electric_age", Lane: LaneFaith,
			Prerequisites: []string{"telecommunications"},
			Description:   "Wireless communication reaches the masses.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "culture", Value: 2.0},
				{Kind: EffectOutput, Target: "knowledge", Value: 0.4},
			},
		},
		{
			Name: "Chemical Engineering", Key: "chemical_engineering", Code: "CHEM", Emblem: "∆",
			Age: "electric_age", Lane: LaneMaterials,
			Prerequisites: []string{"mass_production"},
			Description:   "Industrial chemistry and synthetic materials.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "oil", Value: 1.0},
				{Kind: EffectAllOutput, Value: 0.2},
			},
		},

		// === ATOMIC AGE ===
		{
			Name: "Nuclear Fission", Key: "nuclear_fission", Code: "FISSN", Emblem: "◉",
			Age: "atomic_age", Lane: LaneEnergy,
			Prerequisites: []string{"power_distribution", "chemical_engineering"},
			Description:   "Splitting the atom for energy and weapons.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "electricity", Value: 5.0},
				{Kind: EffectFlatOutput, Target: "uranium", Value: 0.5},
			},
		},
		{
			// No longer behind Rifling and Chemical Engineering: flight comes first.
			// It is a root until the tech it will follow (Aviation) exists.
			Name: "Rocketry", Key: "rocketry", Code: "ROCKT", Emblem: "▲",
			Age: "atomic_age", Lane: LaneSpace,
			Description: "Rockets raise military power and expedition rewards.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 1.0},
				{Kind: EffectExpeditionReward, Value: 0.5},
			},
		},
		{
			Name: "Nuclear Deterrence", Key: "nuclear_deterrence", Code: "DETER", Emblem: "☠",
			Age: "atomic_age", Lane: LaneMilitary,
			Prerequisites: []string{"nuclear_fission", "rocketry"},
			Description:   "Mutually assured destruction maintains peace.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 1.5},
			},
		},
		{
			// Mid-age unlock (Pacing v2): it stands behind the age's other
			// three techs, so it comes after them, and opens the Nuclear
			// Plant in what was a quiet stretch.
			Name: "Civilian Reactors", Key: "civilian_reactors", Code: "REACT", Emblem: "▣",
			Age: "atomic_age", Lane: LaneEnergy,
			Prerequisites: []string{"nuclear_deterrence"},
			Description:   "The reactors built for the arms race find steadier work on the grid. Opens the Nuclear Plant.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "electricity", Value: 5.0},
				{Kind: EffectFlatOutput, Target: "uranium", Value: 0.5},
			},
		},

		// === MODERN AGE ===
		{
			Name: "Advanced Electrics", Key: "electricity_tech", Code: "ADVEL", Emblem: "ϟ",
			Age: "modern_age", Lane: LaneEnergy,
			Prerequisites: []string{"nuclear_fission"},
			Description:   "Advanced electrical systems raise all production.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.5},
				{Kind: EffectFlatOutput, Target: "electricity", Value: 5.0},
			},
		},
		{
			Name: "Computers", Key: "computers", Code: "COMP",
			Age: "modern_age", Lane: LaneComputing,
			Prerequisites: []string{"electricity_tech"},
			Description:   "Digital computing raises knowledge output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.8},
			},
		},
		{
			Name: "Satellite Technology", Key: "satellite_tech", Emblem: "✧",
			Age: "modern_age", Lane: LaneSpace,
			Prerequisites: []string{"rocketry", "electricity_tech"},
			Description:   "Orbital satellites for communication and surveillance.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 1.0},
				{Kind: EffectOutput, Target: "knowledge", Value: 0.6},
			},
		},
		{
			Name: "Nanofabrication", Key: "nanofabrication", Code: "NANO", Emblem: "◇",
			Age: "modern_age", Lane: LaneCraft,
			Prerequisites: []string{"computers"},
			Description:   "Nanobot swarms assemble structures atom-by-atom, cutting construction costs.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.08},
			},
		},

		// === INFORMATION AGE ===
		{
			Name: "Internet", Key: "internet",
			Age: "information_age", Lane: LaneComputing,
			Prerequisites: []string{"computers", "satellite_tech"},
			Description:   "Global network connecting all of humanity.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 3.0},
				{Kind: EffectOutput, Target: "knowledge", Value: 1.2},
			},
		},
		{
			Name: "Cybersecurity", Key: "cybersecurity", Code: "SECUR",
			Age: "information_age", Lane: LaneMilitary,
			Prerequisites: []string{"computers"},
			Description:   "Defense against digital threats.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 1.0},
				{Kind: EffectFlatStorage, Target: "data", Value: 5000},
			},
		},
		{
			Name: "Social Media", Key: "social_media",
			Age: "information_age", Lane: LaneFaith,
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
			Age: "information_age", Lane: LaneAgriculture,
			Prerequisites: []string{"nanofabrication"},
			Description:   "Bloodstream nanobots keep workers healthy, adding housing and food.",
			Effects: []TechEffect{
				{Kind: EffectFlatHousing, Value: 10},
				{Kind: EffectFlatOutput, Target: "food", Value: 8.0},
			},
		},
		{
			// Mid-age unlock (Pacing v2): it stands behind three of the
			// age's techs, so it comes after them.
			Name: "Internet of Things", Key: "internet_of_things", Code: "IOT",
			Age: "information_age", Lane: LaneAgriculture,
			Prerequisites: []string{"social_media", "cybersecurity", "medical_nanobots"},
			Description:   "The fridges and the tractors go online and start reporting back. Opens the Smart Farm and the Smart Complex.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 3.0},
				{Kind: EffectFlatOutput, Target: "food", Value: 8.0},
			},
		},

		// === DIGITAL AGE ===
		{
			Name: "Machine Learning", Key: "machine_learning",
			Age: "digital_age", Lane: LaneComputing,
			Prerequisites: []string{"internet", "cybersecurity"},
			Description:   "Algorithms that learn and improve autonomously.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 5.0},
				{Kind: EffectAllOutput, Value: 0.5},
			},
		},
		{
			Name: "Cloud Computing", Key: "cloud_computing",
			Age: "digital_age", Lane: LaneComputing,
			Prerequisites: []string{"internet"},
			Description:   "Distributed computing at global scale.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 8.0},
				{Kind: EffectFlatStorage, Target: AllResources, Value: 10000},
			},
		},
		{
			Name: "Self-Replication", Key: "self_replication",
			Age: "digital_age", Lane: LaneCraft,
			Prerequisites: []string{"medical_nanobots", "machine_learning"},
			Description:   "Nanobots that build copies of themselves.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "nanobots", Value: 200.0},
			},
		},

		// === CYBERPUNK AGE ===
		{
			Name: "Neural Interface", Key: "neural_interface",
			Age: "cyberpunk_age", Lane: LaneKnowledge,
			Prerequisites: []string{"machine_learning"},
			Description:   "Direct brain-computer interface technology.",
			Effects: []TechEffect{
				{Kind: EffectWorkerOutput, Value: 0.3},
				{Kind: EffectOutput, Target: "knowledge", Value: 2.0},
			},
		},
		{
			Name: "Blockchain", Key: "blockchain",
			Age: "cyberpunk_age", Lane: LaneTrade,
			Prerequisites: []string{"cybersecurity", "cloud_computing"},
			Description:   "Decentralized trustless systems.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "crypto", Value: 2.0},
				{Kind: EffectOutput, Target: "gold", Value: 2.0},
			},
		},
		{
			// The Neon Citadel's keystone, with Holography behind it.
			Name: "Cybernetics", Key: "cybernetics",
			Age: "cyberpunk_age", Lane: LaneCraft,
			Prerequisites: []string{"neural_interface"},
			Description:   "Mechanical augmentation of the human body.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.5},
				{Kind: EffectMilitaryPower, Value: 1.0},
			},
		},
		{
			// Mid-age unlock (Pacing v2): it stands behind Cybernetics and
			// Blockchain, so it comes during the saving-up for the wonder.
			Name: "Holography", Key: "holography",
			Age: "cyberpunk_age", Lane: LaneFaith,
			Prerequisites: []string{"cybernetics", "blockchain"},
			Description:   "Light learns to lie convincingly, and every wall becomes an ad. Opens the Holographic Theater.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "culture", Value: 5.0},
				{Kind: EffectFlatOutput, Target: "crypto", Value: 2.0},
			},
		},

		// === FUSION AGE ===
		{
			Name: "Fusion Power", Key: "fusion_power",
			Age: "fusion_age", Lane: LaneEnergy,
			Prerequisites: []string{"nuclear_fission", "cybernetics"},
			Description:   "Controlled fusion adds electricity and plasma.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "electricity", Value: 20.0},
				{Kind: EffectFlatOutput, Target: "plasma", Value: 1.0},
			},
		},
		{
			// Pacing v2: Plasma Physics, Superconductors and Maglev Transit
			// run one after another (each needs the one before), so the
			// age's research is spread along it. The age used to go 26
			// hours from its last tech to its wonder.
			Name: "Plasma Physics", Key: "plasma_physics",
			Age: "fusion_age", Lane: LaneEnergy,
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
			Age: "fusion_age", Lane: LaneMaterials,
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
			Age: "fusion_age", Lane: LaneTrade,
			Prerequisites: []string{"superconductors"},
			Description:   "Superconducting rails float the freight across the city at the speed of a mild panic. Opens the Energy Exchange.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "plasma", Value: 1.0},
				{Kind: EffectFlatOutput, Target: "gold", Value: 5.0},
			},
		},

		// === SPACE AGE ===
		{
			Name: "Orbital Mechanics", Key: "orbital_mechanics",
			Age: "space_age", Lane: LaneSpace,
			Prerequisites: []string{"rocketry", "plasma_physics"},
			Description:   "Advanced spaceflight and orbital dynamics.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "titanium", Value: 1.0},
				{Kind: EffectExpeditionReward, Value: 1.0},
			},
		},
		{
			Name: "Space Mining", Key: "space_mining",
			Age: "space_age", Lane: LaneSpace,
			Prerequisites: []string{"orbital_mechanics"},
			Description:   "Asteroid and lunar resource extraction.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "titanium", Value: 3.0},
				{Kind: EffectFlatOutput, Target: "iron", Value: 20.0},
			},
		},
		{
			Name: "Zero-G Manufacturing", Key: "zero_g_manufacturing",
			Age: "space_age", Lane: LaneCraft,
			Prerequisites: []string{"orbital_mechanics", "superconductors"},
			Description:   "Space-based manufacturing for perfect materials.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.5},
				{Kind: EffectFlatOutput, Target: "steel", Value: 10.0},
			},
		},

		// === INTERSTELLAR AGE ===
		{
			Name: "Warp Drive", Key: "warp_drive",
			Age: "interstellar_age", Lane: LaneSpace,
			Prerequisites: []string{"space_mining", "zero_g_manufacturing"},
			Description:   "Faster-than-light propulsion.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "dark_matter", Value: 1.0},
				{Kind: EffectExpeditionReward, Value: 2.0},
			},
		},
		{
			Name: "Stellar Engineering", Key: "stellar_engineering",
			Age: "interstellar_age", Lane: LaneEnergy,
			Prerequisites: []string{"warp_drive"},
			Description:   "Harnessing and shaping stars themselves.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "plasma", Value: 10.0},
				{Kind: EffectFlatOutput, Target: "electricity", Value: 100.0},
			},
		},

		// === GALACTIC AGE ===
		{
			Name: "Galactic Navigation", Key: "galactic_navigation",
			Age: "galactic_age", Lane: LaneSpace,
			Prerequisites: []string{"warp_drive", "stellar_engineering"},
			Description:   "Charting paths across the galaxy.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.5},
				{Kind: EffectFlatOutput, Target: "dark_matter", Value: 5.0},
			},
		},
		{
			Name: "Antimatter Synthesis", Key: "antimatter_synthesis",
			Age: "galactic_age", Lane: LaneEnergy,
			Prerequisites: []string{"galactic_navigation"},
			Description:   "Controlled production of antimatter.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "antimatter", Value: 2.0},
				{Kind: EffectAllOutput, Value: 0.3},
			},
		},

		// === QUANTUM AGE ===
		{
			Name: "Quantum Mechanics", Key: "quantum_mechanics",
			Age: "quantum_age", Lane: LaneKnowledge,
			Prerequisites: []string{"antimatter_synthesis"},
			Description:   "Mastery of quantum phenomena at all scales.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "quantum_flux", Value: 2.0},
				{Kind: EffectAllOutput, Value: 1.0},
			},
		},
		{
			Name: "Reality Manipulation", Key: "reality_manipulation",
			Age: "quantum_age", Lane: LaneCraft,
			Prerequisites: []string{"quantum_mechanics"},
			Description:   "Bending the fabric of spacetime.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "quantum_flux", Value: 5.0},
				{Kind: EffectAllOutput, Value: 1.0},
			},
		},
		{
			Name: "Quantum Computing", Key: "quantum_computing", Code: "QCOMP",
			Age: "quantum_age", Lane: LaneComputing,
			Prerequisites: []string{"clockwork_automation", "quantum_mechanics"},
			Description:   "Quantum processing raises game speed.",
			Effects: []TechEffect{
				{Kind: EffectGameSpeed, Value: 0.15},
			},
		},

		// === TRANSCENDENT AGE ===
		{
			Name: "Transcendence", Key: "transcendence",
			Age: "transcendent_age", Lane: LaneFaith,
			Prerequisites: []string{"reality_manipulation"},
			Description:   "A civilization beyond physical limits.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 2.0},
				{Kind: EffectFlatOutput, Target: "quantum_flux", Value: 10.0},
			},
		},
	}
	// The wonders alone decide the kinds, and they are all in the raw table:
	// reading it instead of BaseBuildings keeps this off the building
	// normalizers (TestTechKindsNeedOnlyTheRawWonders).
	kinds := TechKinds(techs, baseBuildingsRaw())
	return fillTechArt(normalizeResearchCosts(normalizeResearchTicks(techs, kinds), kinds))
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
