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
	techs := rawTechnologies()
	// The wonders alone decide the kinds, and they are all in the raw table:
	// reading it instead of BaseBuildings keeps this off the building
	// normalizers (TestTechKindsNeedOnlyTheRawWonders).
	kinds := TechKinds(techs, baseBuildingsRaw())
	return fillTechArt(normalizeResearchCosts(normalizeResearchTicks(techs, kinds), kinds))
}

// rawTechnologies is the tech table as it is written: no cost, no research
// time, and art only where a tech sets its own.
func rawTechnologies() []TechDef {
	return []TechDef{
		// === PRIMITIVE AGE ===
		// Three roots: speech, fire and tools. No tech is needed to leave it.
		{
			Name: "Language", Key: "language", Code: "LANG", Emblem: "♪",
			Age: "primitive_age", Lane: LaneKnowledge,
			Description: "Shared words carry what one person learns to the next.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.10},
			},
		},
		{
			Name: "Fire Mastery", Key: "fire_mastery", Emblem: "△",
			Age: "primitive_age", Lane: LaneAgriculture,
			Description: "Control of fire improves food preservation and warmth.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.10},
				{Kind: EffectHousing, Value: 0.05},
			},
		},
		{
			Name: "Tool Making", Key: "tool_making", Code: "TOOLS", Emblem: "⚒",
			Age: "primitive_age", Lane: LaneCraft,
			Description: "Stone tools make every worker more productive.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.10},
				{Kind: EffectOutput, Target: "wood", Value: 0.10},
				{Kind: EffectMechanic, Target: MechanicGatherAmount, Value: 2},
			},
		},

		// === STONE AGE ===
		{
			Name: "Ritual", Key: "ritual", Code: "RITE", Emblem: "∴",
			Age: "stone_age", Lane: LaneFaith,
			Prerequisites: []string{"language"},
			Description:   "Shared rites give the tribe its first holy places.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "faith", Value: 0.10},
			},
		},
		{
			Name: "Primitive Writing", Key: "primitive_writing", Code: "WRITE", Emblem: "§",
			Age: "stone_age", Lane: LaneKnowledge,
			Prerequisites: []string{"language"},
			Description:   "Early symbols enable knowledge transfer.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.10},
			},
		},
		{
			Name: "Pottery", Key: "pottery", Code: "POTS", Emblem: "∪",
			Age: "stone_age", Lane: LaneTrade,
			Prerequisites: []string{"fire_mastery"},
			Description:   "Clay vessels for storage and trade.",
			Effects: []TechEffect{
				{Kind: EffectStorage, Value: 0.10},
			},
		},
		{
			Name: "Animal Husbandry", Key: "animal_husbandry", Code: "HERDS", Emblem: "♞",
			Age: "stone_age", Lane: LaneAgriculture,
			Prerequisites: []string{"fire_mastery"},
			Description:   "Domesticating animals for food and labor.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.10},
			},
		},
		{
			Name: "Woodworking", Key: "woodworking", Code: "WOOD", Emblem: "⊤",
			Age: "stone_age", Lane: LaneCraft,
			Prerequisites: []string{"tool_making"},
			Description:   "Joints, pegs and planks turn timber into more than firewood.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "wood", Value: 0.10},
			},
		},
		{
			Name: "Stoneworking", Key: "stoneworking", Emblem: "◆",
			Age: "stone_age", Lane: LaneMaterials,
			Prerequisites: []string{"tool_making"},
			Description:   "Cutting and shaping stone for construction.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "stone", Value: 0.10},
			},
		},

		// === BRONZE AGE ===
		{
			Name: "Calendar", Key: "calendar", Code: "CALEN", Emblem: "◔",
			Age: "bronze_age", Lane: LaneFaith,
			Prerequisites: []string{"ritual"},
			Description:   "Counting the days fixes the feasts and the seasons.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "faith", Value: 0.10},
			},
		},
		{
			Name: "Map Making", Key: "map_making", Code: "MAPS", Emblem: "⊕",
			Age: "bronze_age", Lane: LaneKnowledge,
			Prerequisites: []string{"primitive_writing"},
			Description:   "Drawn maps record where things are and how to reach them.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.10},
			},
		},
		{
			Name: "Currency", Key: "currency", Code: "COIN", Emblem: "¤",
			Age: "bronze_age", Lane: LaneTrade,
			Prerequisites: []string{"primitive_writing"},
			Description:   "Standardized money raises gold output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.10},
				{Kind: EffectMechanic, Target: MechanicMarketFee, Value: -0.03},
			},
		},
		{
			Name: "Boatbuilding", Key: "boatbuilding", Code: "BOATS", Emblem: "◡",
			Age: "bronze_age", Lane: LaneTrade,
			Prerequisites: []string{"woodworking"},
			Description:   "Hulls and oars bring in the catch and carry goods along the coast.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.05},
				{Kind: EffectMechanic, Target: MechanicRouteIncome, Value: 0.10},
			},
		},
		{
			Name: "Agriculture", Key: "agriculture", Code: "FARMS", Emblem: "♠",
			Age: "bronze_age", Lane: LaneAgriculture,
			Prerequisites: []string{"animal_husbandry"},
			Description:   "Systematic farming adds steady food output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.10},
			},
		},
		{
			Name: "The Wheel", Key: "the_wheel", Emblem: "⊗",
			Age: "bronze_age", Lane: LaneCraft,
			Prerequisites: []string{"woodworking"},
			Description:   "Carts move what backs could not, to the building site and to market.",
			Effects: []TechEffect{
				{Kind: EffectBuildTime, Value: -0.05},
			},
		},
		{
			Name: "Masonry", Key: "masonry", Emblem: "▦",
			Age: "bronze_age", Lane: LaneCraft,
			Prerequisites: []string{"stoneworking"},
			Description:   "Advanced stone construction techniques.",
			Effects: []TechEffect{
				{Kind: EffectStorage, Value: 0.10},
			},
		},
		{
			Name: "Bronze Working", Key: "bronze_working", Emblem: "◐",
			Age: "bronze_age", Lane: LaneMaterials,
			Prerequisites: []string{"stoneworking"},
			Description:   "Alloying copper and tin creates durable tools.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "stone", Value: 0.10},
				{Kind: EffectOutput, Target: "iron", Value: 0.10},
			},
		},
		{
			Name: "Military Tactics", Key: "military_tactics", Code: "TACTI", Emblem: "†",
			Age: "bronze_age", Lane: LaneMilitary,
			Prerequisites: []string{"bronze_working"},
			Description:   "Organized warfare and defense strategies.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.15},
			},
		},

		// === IRON AGE ===
		{
			Name: "Priesthood", Key: "priesthood", Emblem: "Ψ",
			Age: "iron_age", Lane: LaneFaith,
			Prerequisites: []string{"calendar"},
			Description:   "A standing priesthood keeps the rites and the people's spirits.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicMoraleCap, Value: 0.05},
			},
		},
		{
			Name: "Mathematics", Key: "mathematics", Code: "MATH", Emblem: "π",
			Age: "iron_age", Lane: LaneKnowledge,
			Prerequisites: []string{"primitive_writing"},
			Description:   "Advanced calculation raises knowledge output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.08},
			},
		},
		{
			// The tree's one either-or group: by map or by boat.
			Name: "Exploration", Key: "exploration", Emblem: "↗",
			Age: "iron_age", Lane: LaneTrade,
			AnyOf:       []string{"map_making", "boatbuilding"},
			Description: "Maps or boats, and the will to see what lies past the next ridge.",
		},
		{
			Name: "Irrigation", Key: "irrigation", Emblem: "≈",
			Age: "iron_age", Lane: LaneAgriculture,
			Prerequisites: []string{"agriculture"},
			Description:   "Channels bring the river to the fields.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.08},
				{Kind: EffectHousing, Value: 0.05},
			},
		},
		{
			Name: "Road Building", Key: "road_building", Code: "ROADS", Emblem: "═",
			Age: "iron_age", Lane: LaneCraft,
			Prerequisites: []string{"masonry", "the_wheel"},
			Description:   "Paved roads improve trade and movement.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.08},
				{Kind: EffectMechanic, Target: MechanicRouteTicks, Value: -0.15},
			},
		},
		{
			Name: "Iron Smelting", Key: "iron_smelting", Emblem: "■",
			Age: "iron_age", Lane: LaneMaterials,
			Prerequisites: []string{"bronze_working"},
			Description:   "Hotter furnaces raise iron output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "iron", Value: 0.08},
			},
		},
		{
			Name: "Siege Warfare", Key: "siege_warfare", Emblem: "✕",
			Age: "iron_age", Lane: LaneMilitary,
			Prerequisites: []string{"military_tactics"},
			Description:   "Siege engines and fortification techniques.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.15},
			},
		},

		// === CLASSICAL AGE ===
		{
			Name: "Drama", Key: "drama", Emblem: "♫",
			Age: "classical_age", Lane: LaneFaith,
			Prerequisites: []string{"priesthood"},
			Description:   "Plays and choruses give a city its festival days.",
		},
		{
			Name: "Philosophy", Key: "philosophy", Emblem: "Φ",
			Age: "classical_age", Lane: LaneKnowledge,
			Prerequisites: []string{"mathematics"},
			Description:   "Systematic inquiry into fundamental questions.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.08},
			},
		},
		{
			Name: "Envoys", Key: "envoys", Code: "ENVOY", Emblem: "⇄",
			Age: "classical_age", Lane: LaneTrade,
			Prerequisites: []string{"exploration"},
			Description:   "Trusted messengers speak for you in other courts.",
		},
		{
			Name: "The Plough", Key: "the_plough", Code: "PLOW", Emblem: "≡",
			Age: "classical_age", Lane: LaneAgriculture,
			Prerequisites: []string{"irrigation"},
			Description:   "An iron share turns heavier soil than a digging stick.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.08},
			},
		},
		{
			Name: "Civil Engineering", Key: "civil_engineering", Emblem: "∩",
			Age: "classical_age", Lane: LaneCraft,
			Prerequisites: []string{"road_building"},
			Description:   "Large-scale construction and infrastructure.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.03},
			},
		},
		{
			Name: "Metal Casting", Key: "metal_casting", Code: "CAST", Emblem: "◘",
			Age: "classical_age", Lane: LaneMaterials,
			Prerequisites: []string{"iron_smelting"},
			Description:   "Molten metal poured into moulds makes the same part a hundred times.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "iron", Value: 0.08},
			},
		},
		{
			Name: "Imperial Legions", Key: "imperial_legions", Code: "LEGIO", Emblem: "⚑",
			Age: "classical_age", Lane: LaneMilitary,
			Prerequisites: []string{"siege_warfare", "iron_smelting"},
			Description:   "Professional standing armies with superior discipline.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.15},
				{Kind: EffectMechanic, Target: MechanicRaidLoss, Value: -0.10},
			},
		},

		// === MEDIEVAL AGE ===
		{
			Name: "Theology", Key: "theology", Emblem: "Θ",
			Age: "medieval_age", Lane: LaneFaith,
			Prerequisites: []string{"philosophy"},
			Description:   "Organized religion provides faith and social cohesion.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "faith", Value: 0.08},
			},
		},
		{
			Name: "Alchemy", Key: "alchemy", Emblem: "☿",
			Age: "medieval_age", Lane: LaneKnowledge,
			Prerequisites: []string{"philosophy"},
			Description:   "Proto-chemistry yields material insights.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.08},
			},
		},
		{
			// The Iron Era's Knowledge capstone.
			Name: "Scholasticism", Key: "scholasticism", Code: "SCHOL", Emblem: "Σ",
			Age: "medieval_age", Lane: LaneKnowledge, Capstone: true,
			Prerequisites: []string{"alchemy", "theology"},
			Description:   "The schools set every question out, argue it and write the answer down.",
			Effects: []TechEffect{
				{Kind: EffectResearchTime, Value: -0.06},
			},
		},
		{
			Name: "Banking", Key: "banking", Code: "BANK", Emblem: "%",
			Age: "medieval_age", Lane: LaneTrade,
			Prerequisites: []string{"currency", "mathematics"},
			Description:   "Financial institutions raise gold output and gold storage.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.08},
				{Kind: EffectMechanic, Target: MechanicMarketFee, Value: -0.03},
			},
		},
		{
			Name: "Feudalism", Key: "feudalism", Emblem: "⌂",
			Age: "medieval_age", Lane: LaneAgriculture,
			Prerequisites: []string{"the_plough"},
			Description:   "Feudal land grants house more workers.",
			Effects: []TechEffect{
				{Kind: EffectHousing, Value: 0.08},
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
		{
			// The Iron Era's Craft capstone.
			Name: "Guilds", Key: "guilds", Code: "GUILD", Emblem: "♜",
			Age: "medieval_age", Lane: LaneCraft, Capstone: true,
			Prerequisites: []string{"civil_engineering", "metal_casting"},
			Description:   "Masters, journeymen and set prices: every trade builds to one standard.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.04},
				{Kind: EffectBuildTime, Value: -0.08},
			},
		},
		{
			Name: "Steel Forging", Key: "steel_forging", Emblem: "▣",
			Age: "medieval_age", Lane: LaneMaterials,
			Prerequisites: []string{"iron_smelting"},
			Description:   "Refining iron into steel for superior tools and weapons.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "steel", Value: 0.25},
				{Kind: EffectOutput, Target: "iron", Value: 0.08},
			},
		},
		{
			Name: "Fortification", Key: "fortification", Code: "FORTS", Emblem: "╬",
			Age: "medieval_age", Lane: LaneMilitary,
			Prerequisites: []string{"imperial_legions"},
			Description:   "Curtain walls and gatehouses make a raid cost more than it takes.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicRaidLoss, Value: -0.15},
				{Kind: EffectMilitaryPower, Value: 0.10},
			},
		},

		// === RENAISSANCE AGE ===
		{
			Name: "Patronage", Key: "patronage", Code: "PATRN", Emblem: "♛",
			Age: "renaissance_age", Lane: LaneFaith,
			Prerequisites: []string{"banking"},
			Description:   "Wealthy patrons fund arts and science.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "culture", Value: 0.06},
			},
		},
		{
			Name: "Printing Press", Key: "printing_press", Emblem: "¶",
			Age: "renaissance_age", Lane: LaneKnowledge,
			Prerequisites: []string{"alchemy", "theology"},
			Description:   "Printed books carry one idea to a thousand readers at once.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.06},
				{Kind: EffectResearchTime, Value: -0.03},
			},
		},
		{
			Name: "Navigation", Key: "navigation", Emblem: "✶",
			Age: "renaissance_age", Lane: LaneTrade,
			Prerequisites: []string{"exploration", "mathematics"},
			Description:   "Star, compass and log line take ships out of sight of land.",
			Effects: []TechEffect{
				{Kind: EffectExpeditionReward, Value: 0.10},
			},
		},
		{
			Name: "Crop Rotation", Key: "crop_rotation", Code: "CROPS", Emblem: "↻",
			Age: "renaissance_age", Lane: LaneAgriculture,
			Prerequisites: []string{"feudalism"},
			Description:   "Fields take turns at wheat, roots and rest, and none wears out.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.06},
			},
		},
		{
			Name: "Architecture", Key: "architecture", Code: "ARCH", Emblem: "∧",
			Age: "renaissance_age", Lane: LaneCraft,
			Prerequisites: []string{"civil_engineering"},
			Description:   "Drawn plans and worked proportions, before the first stone is laid.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.03},
				{Kind: EffectMechanic, Target: MechanicWonderBuildTicks, Value: -0.50},
			},
		},
		{
			Name: "Blast Furnace", Key: "blast_furnace", Emblem: "◭",
			Age: "renaissance_age", Lane: LaneMaterials,
			Prerequisites: []string{"steel_forging"},
			Description:   "A taller stack and a harder blast turn out metal by the ton.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "steel", Value: 0.06},
			},
		},
		{
			Name: "Gunpowder", Key: "gunpowder", Code: "POWDR", Emblem: "✸",
			Age: "renaissance_age", Lane: LaneMilitary,
			Prerequisites: []string{"alchemy", "siege_warfare"},
			Description:   "Explosive weapons raise military power.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.12},
			},
		},

		// === COLONIAL AGE ===
		{
			Name: "Baroque Arts", Key: "baroque_arts", Emblem: "❦",
			Age: "colonial_age", Lane: LaneFaith,
			Prerequisites: []string{"patronage"},
			Description:   "Music and ornament on a scale built to overwhelm.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicFestivalTicks, Value: 0.25},
			},
		},
		{
			Name: "Scientific Method", Key: "scientific_method", Emblem: "⊢",
			Age: "colonial_age", Lane: LaneKnowledge,
			Prerequisites: []string{"printing_press"},
			Description:   "Guess, test, write it down, and let someone else try to break it.",
			Effects: []TechEffect{
				{Kind: EffectResearchTime, Value: -0.04},
			},
		},
		{
			Name: "Cartography", Key: "cartography", Emblem: "⊞",
			Age: "colonial_age", Lane: LaneTrade,
			Prerequisites: []string{"navigation"},
			Description:   "Detailed maps bring expeditions home sooner and richer.",
			Effects: []TechEffect{
				{Kind: EffectExpeditionReward, Value: 0.10},
				{Kind: EffectMechanic, Target: MechanicExpeditionTicks, Value: -0.10},
			},
		},
		{
			Name: "Mercantilism", Key: "mercantilism", Code: "MERC", Emblem: "£",
			Age: "colonial_age", Lane: LaneTrade,
			Prerequisites: []string{"banking", "navigation"},
			Description:   "National trade policies maximize wealth.",
		},
		{
			Name: "Embassies", Key: "embassies", Emblem: "⚐",
			Age: "colonial_age", Lane: LaneTrade,
			Prerequisites: []string{"envoys"},
			Description:   "A resident envoy in every court, and a house to keep them in.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicGiftOpinion, Value: 0.50},
			},
		},
		{
			Name: "New World Crops", Key: "new_world_crops", Code: "MAIZE", Emblem: "✿",
			Age: "colonial_age", Lane: LaneAgriculture,
			Prerequisites: []string{"crop_rotation"},
			Description:   "Maize, potatoes and beans cross the ocean and feed twice the mouths.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.06},
				{Kind: EffectHousing, Value: 0.04},
			},
		},
		{
			Name: "Surveying", Key: "surveying", Code: "SURVY", Emblem: "∠",
			Age: "colonial_age", Lane: LaneCraft,
			Prerequisites: []string{"architecture"},
			Description:   "Chain, level and theodolite: the ground is measured before it is built on.",
			Effects: []TechEffect{
				{Kind: EffectBuildTime, Value: -0.05},
				{Kind: EffectBuildCost, Value: -0.02},
			},
		},
		{
			Name: "Coke Smelting", Key: "coke_smelting", Emblem: "●",
			Age: "colonial_age", Lane: LaneMaterials,
			Prerequisites: []string{"blast_furnace"},
			Description:   "Coal baked into coke burns hot enough to smelt without a forest.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "steel", Value: 0.06},
				{Kind: EffectOutput, Target: "coal", Value: 0.06},
			},
		},
		{
			Name: "Colonialism", Key: "colonialism", Code: "COLNY", Emblem: "⚔",
			Age: "colonial_age", Lane: LaneMilitary,
			Prerequisites: []string{"cartography", "gunpowder"},
			Description:   "Overseas territorial expansion.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.12},
			},
		},

		// === INDUSTRIAL AGE ===
		{
			Name: "Romanticism", Key: "romanticism", Emblem: "♥",
			Age: "industrial_age", Lane: LaneFaith,
			Prerequisites: []string{"baroque_arts"},
			Description:   "Feeling over reason, on the stage and on the page.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "culture", Value: 0.06},
			},
		},
		{
			Name: "Encyclopedia", Key: "encyclopedia", Emblem: "Æ",
			Age: "industrial_age", Lane: LaneKnowledge,
			Prerequisites: []string{"scientific_method"},
			Description:   "Everything known, set in order and put in print.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.06},
			},
		},
		{
			Name: "Railroads", Key: "railroads", Code: "RAIL", Emblem: "‡",
			Age: "industrial_age", Lane: LaneTrade,
			Prerequisites: []string{"steam_power", "road_building"},
			Description:   "Rail networks connect your civilization.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicRouteTicks, Value: -0.15},
			},
		},
		{
			Name: "Geographic Societies", Key: "geographic_societies", Code: "GEOG", Emblem: "◎",
			Age: "industrial_age", Lane: LaneTrade,
			Prerequisites: []string{"cartography"},
			Description:   "Learned societies fund the expeditions and publish what they find.",
		},
		{
			// The Steel Era's Trade capstone.
			Name: "Concert of Nations", Key: "concert_of_nations", Code: "CONCT", Emblem: "⚖",
			Age: "industrial_age", Lane: LaneTrade, Capstone: true,
			Prerequisites: []string{"embassies", "geographic_societies"},
			Description:   "The great powers settle their quarrels at a table and keep the peace by treaty.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicAllianceBonus, Value: 0.25},
			},
		},
		{
			Name: "Seed Drill", Key: "seed_drill", Code: "DRILL", Emblem: "∷",
			Age: "industrial_age", Lane: LaneAgriculture,
			Prerequisites: []string{"new_world_crops"},
			Description:   "Seed sown in rows at an even depth, and none thrown to the birds.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.06},
			},
		},
		{
			Name: "Industrialization", Key: "industrialization", Emblem: "⚙",
			Age: "industrial_age", Lane: LaneCraft,
			Prerequisites: []string{"steam_power"},
			Description:   "Factory systems raise all production.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.05},
			},
		},
		{
			Name: "Clockwork Automation", Key: "clockwork_automation", Emblem: "✲",
			Age: "industrial_age", Lane: LaneCraft,
			Prerequisites: []string{"chronometry"},
			Description:   "Mechanical automation raises game speed.",
			Effects: []TechEffect{
				{Kind: EffectGameSpeed, Value: 0.10},
			},
		},
		{
			// The Steel Era's Craft capstone.
			Name: "Interchangeable Parts", Key: "interchangeable_parts", Code: "PARTS", Emblem: "❖",
			Age: "industrial_age", Lane: LaneCraft, Capstone: true,
			Prerequisites: []string{"industrialization", "clockwork_automation"},
			Description:   "Every part made to one gauge fits every machine of its kind.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.04},
				{Kind: EffectMechanic, Target: MechanicUpgradeCost, Value: -0.15},
			},
		},
		{
			Name: "Steam Pumps", Key: "steam_pumps", Code: "PUMPS", Emblem: "⇕",
			Age: "industrial_age", Lane: LaneMaterials,
			Prerequisites: []string{"coke_smelting"},
			Description:   "Engines drain the deep workings, and the mines go further down.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "iron_ore", Value: 0.06},
				{Kind: EffectOutput, Target: "coal", Value: 0.06},
			},
		},
		{
			Name: "Rifling", Key: "rifling", Code: "RIFLE", Emblem: "✛",
			Age: "industrial_age", Lane: LaneMilitary,
			Prerequisites: []string{"gunpowder"},
			Description:   "Precision firearms improve military effectiveness.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.12},
			},
		},
		{
			Name: "Steam Power", Key: "steam_power", Emblem: "≈",
			Age: "industrial_age", Lane: LaneEnergy,
			Prerequisites: []string{"steel_forging"},
			Description:   "Steam engines drive the mills and the mines.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "steel", Value: 0.06},
				{Kind: EffectOutput, Target: "coal", Value: 0.06},
			},
		},

		// === VICTORIAN AGE ===
		{
			Name: "Museums", Key: "museums", Code: "MUSEM", Emblem: "Π",
			Age: "victorian_age", Lane: LaneFaith,
			Prerequisites: []string{"romanticism"},
			Description:   "The nation's treasures behind glass, open to anyone on a Sunday.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "culture", Value: 0.05},
			},
		},
		{
			Name: "Public Education", Key: "public_education", Code: "EDUC", Emblem: "✎",
			Age: "victorian_age", Lane: LaneKnowledge,
			Prerequisites: []string{"encyclopedia"},
			Description:   "Every child in a schoolroom, and every schoolroom teaching the same lessons.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.05},
				{Kind: EffectResearchTime, Value: -0.03},
			},
		},
		{
			Name: "Telecommunications", Key: "telecommunications", Code: "TELEG", Emblem: "∿",
			Age: "victorian_age", Lane: LaneTrade,
			Prerequisites: []string{"electrification"},
			Description:   "Telegraph and early telephone networks.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicDealRefreshTicks, Value: -0.30},
				{Kind: EffectMechanic, Target: MechanicGiftCost, Value: -0.25},
			},
		},
		{
			Name: "Sanitation", Key: "sanitation", Emblem: "⊔",
			Age: "victorian_age", Lane: LaneAgriculture,
			Prerequisites: []string{"seed_drill"},
			Description:   "Sewers, clean water and paved streets let a city grow without sickening.",
			Effects: []TechEffect{
				{Kind: EffectHousing, Value: 0.06},
			},
		},
		{
			Name: "Mass Production", Key: "mass_production", Emblem: "▥",
			Age: "victorian_age", Lane: LaneCraft,
			Prerequisites: []string{"industrialization"},
			Description:   "Standard goods, made in long runs by the thousand.",
			Effects: []TechEffect{
				{Kind: EffectBuildTime, Value: -0.08},
			},
		},
		{
			Name: "Geology", Key: "geology", Code: "GEOL", Emblem: "▤",
			Age: "victorian_age", Lane: LaneMaterials,
			Prerequisites: []string{"steam_pumps"},
			Description:   "The strata are read like a book, and the seam is found before the shaft is sunk.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "iron_ore", Value: 0.05},
				{Kind: EffectOutput, Target: "coal", Value: 0.05},
			},
		},
		{
			Name: "General Staff", Key: "general_staff", Code: "STAFF", Emblem: "★",
			Age: "victorian_age", Lane: LaneMilitary,
			Prerequisites: []string{"rifling"},
			Description:   "Officers whose whole work is to plan the war before it is fought.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicCampaignTicks, Value: -0.15},
			},
		},
		{
			Name: "Electrification", Key: "electrification", Code: "ELEC", Emblem: "ϟ",
			Age: "victorian_age", Lane: LaneEnergy,
			Prerequisites: []string{"industrialization"},
			Description:   "Electric power reaches homes and factories.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "electricity", Value: 0.05},
			},
		},

		// === ELECTRIC AGE ===
		{
			Name: "Radio", Key: "radio", Emblem: "♪",
			Age: "electric_age", Lane: LaneFaith,
			Prerequisites: []string{"telecommunications"},
			Description:   "Wireless communication reaches the masses.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicFestivalCooldownTicks, Value: -0.20},
			},
		},
		{
			Name: "Modern Physics", Key: "modern_physics", Code: "PHYS", Emblem: "ħ",
			Age: "electric_age", Lane: LaneKnowledge,
			Prerequisites: []string{"public_education"},
			Description:   "Relativity and the quantum: the old certainties, measured and found short.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.05},
			},
		},
		{
			Name: "Wire Transfers", Key: "wire_transfers", Emblem: "↯",
			Age: "electric_age", Lane: LaneTrade,
			Prerequisites: []string{"telecommunications"},
			Description:   "Money goes down a telegraph wire and arrives before the letter that announces it.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicMarketFee, Value: -0.02},
			},
		},
		{
			Name: "Fertilizers", Key: "fertilizers", Code: "FERT", Emblem: "❋",
			Age: "electric_age", Lane: LaneAgriculture,
			Prerequisites: []string{"sanitation"},
			Description:   "Nitrogen fixed from the air feeds fields that manure never could.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.05},
			},
		},
		{
			Name: "Assembly Line", Key: "assembly_line", Emblem: "⇉",
			Age: "electric_age", Lane: LaneCraft,
			Prerequisites: []string{"mass_production"},
			Description:   "The work moves to the worker, one step at a time.",
			Effects: []TechEffect{
				{Kind: EffectBuildTime, Value: -0.05},
			},
		},
		{
			Name: "Chemical Engineering", Key: "chemical_engineering", Code: "CHEM", Emblem: "∆",
			Age: "electric_age", Lane: LaneMaterials,
			Prerequisites: []string{"mass_production"},
			Description:   "Industrial chemistry and synthetic materials.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "oil", Value: 0.05},
				{Kind: EffectOutput, Target: "steel", Value: 0.05},
			},
		},
		{
			Name: "Mechanized Warfare", Key: "mechanized_warfare", Code: "MECH", Emblem: "▰",
			Age: "electric_age", Lane: LaneMilitary,
			Prerequisites: []string{"general_staff"},
			Description:   "Engines and armor take the place of the horse and the charge.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
			},
		},
		{
			Name: "Power Distribution", Key: "power_distribution", Code: "GRID", Emblem: "#",
			Age: "electric_age", Lane: LaneEnergy,
			Prerequisites: []string{"electrification"},
			Description:   "AC power grids span entire regions.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "electricity", Value: 0.05},
			},
		},
		{
			Name: "Aviation", Key: "aviation", Emblem: "✈",
			Age: "electric_age", Lane: LaneSpace,
			Prerequisites: []string{"mass_production"},
			Description:   "Powered flight shrinks every journey.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicExpeditionTicks, Value: -0.10},
			},
		},

		// === ATOMIC AGE ===
		{
			Name: "Cinema", Key: "cinema", Code: "FILM", Emblem: "►",
			Age: "atomic_age", Lane: LaneFaith,
			Prerequisites: []string{"radio"},
			Description:   "A whole town in the dark, watching the same story.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "culture", Value: 0.05},
			},
		},
		{
			// The Electric Era's Knowledge capstone.
			Name: "Big Science", Key: "big_science", Code: "BIGSC", Emblem: "⚛",
			Age: "atomic_age", Lane: LaneKnowledge, Capstone: true,
			Prerequisites: []string{"modern_physics"},
			Description:   "Laboratories the size of towns, with budgets to match.",
			Effects: []TechEffect{
				{Kind: EffectResearchTime, Value: -0.06},
			},
		},
		{
			Name: "Corporations", Key: "corporations", Code: "CORP", Emblem: "©",
			Age: "atomic_age", Lane: LaneTrade,
			Prerequisites: []string{"mercantilism", "wire_transfers"},
			Description:   "Firms that outlive their founders and trade on every continent.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "gold", Value: 0.05},
			},
		},
		{
			Name: "Green Revolution", Key: "green_revolution", Emblem: "❧",
			Age: "atomic_age", Lane: LaneAgriculture,
			Prerequisites: []string{"fertilizers"},
			Description:   "New strains of wheat and rice double the harvest of the same field.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.05},
				{Kind: EffectHousing, Value: 0.04},
			},
		},
		{
			Name: "Prefabrication", Key: "prefabrication", Code: "PREFB", Emblem: "◫",
			Age: "atomic_age", Lane: LaneCraft,
			Prerequisites: []string{"assembly_line"},
			Description:   "Buildings made in a factory and bolted together on site.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.02},
				{Kind: EffectStorage, Value: 0.05},
			},
		},
		{
			Name: "Plastics", Key: "plastics", Emblem: "⬡",
			Age: "atomic_age", Lane: LaneMaterials,
			Prerequisites: []string{"chemical_engineering"},
			Description:   "Oil turned into anything, in any shape, by the ton.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "oil", Value: 0.05},
			},
		},
		{
			Name: "Nuclear Deterrence", Key: "nuclear_deterrence", Code: "DETER", Emblem: "☠",
			Age: "atomic_age", Lane: LaneMilitary,
			Prerequisites: []string{"nuclear_fission", "rocketry"},
			Description:   "Mutually assured destruction maintains peace.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
				{Kind: EffectMechanic, Target: MechanicRaidLoss, Value: -0.10},
			},
		},
		{
			// The Electric Era's Military capstone.
			Name: "Military-Industrial Complex", Key: "military_industrial_complex", Code: "MIC", Emblem: "▩",
			Age: "atomic_age", Lane: LaneMilitary, Capstone: true,
			Prerequisites: []string{"mechanized_warfare", "nuclear_deterrence"},
			Description:   "Armies, factories and laboratories on one budget, in peace as in war.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicSoldierStorage, Value: 0.20},
				{Kind: EffectMechanic, Target: MechanicCampaignReward, Value: 0.20},
			},
		},
		{
			Name: "Nuclear Fission", Key: "nuclear_fission", Code: "FISSN", Emblem: "◉",
			Age: "atomic_age", Lane: LaneEnergy,
			Prerequisites: []string{"power_distribution", "chemical_engineering"},
			Description:   "Splitting the atom for energy and weapons.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "uranium", Value: 0.05},
				{Kind: EffectOutput, Target: "electricity", Value: 0.05},
			},
		},
		{
			// Mid-age unlock (Pacing v2): it stands behind the age's keystone
			// and Rocketry, so it comes after them, and opens the Nuclear
			// Plant in what was a quiet stretch.
			Name: "Civilian Reactors", Key: "civilian_reactors", Code: "REACT", Emblem: "▣",
			Age: "atomic_age", Lane: LaneEnergy,
			Prerequisites: []string{"nuclear_deterrence"},
			Description:   "The reactors built for the arms race find steadier work on the grid. Opens the Nuclear Plant.",
		},
		{
			// Flight comes first: Rocketry stands on Aviation, no longer on
			// Rifling and Chemical Engineering, which keeps the military
			// chain optional.
			Name: "Rocketry", Key: "rocketry", Code: "ROCKT", Emblem: "▲",
			Age: "atomic_age", Lane: LaneSpace,
			Prerequisites: []string{"aviation"},
			Description:   "Rockets raise military power and expedition rewards.",
			Effects: []TechEffect{
				{Kind: EffectExpeditionReward, Value: 0.10},
				{Kind: EffectMilitaryPower, Value: 0.10},
			},
		},

		// === MODERN AGE ===
		{
			Name: "Advanced Electrics", Key: "electricity_tech", Code: "ADVEL", Emblem: "ϟ",
			Age: "modern_age", Lane: LaneEnergy,
			Prerequisites: []string{"nuclear_fission"},
			Description:   "Advanced electrical systems raise all production.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "electricity", Value: 0.05},
			},
		},
		{
			Name: "Computers", Key: "computers", Code: "COMP",
			Age: "modern_age", Lane: LaneComputing,
			Prerequisites: []string{"electricity_tech"},
			Description:   "Digital computing raises knowledge output.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.05},
				{Kind: EffectResearchTime, Value: -0.03},
			},
		},
		{
			Name: "Satellite Technology", Key: "satellite_tech", Emblem: "✧",
			Age: "modern_age", Lane: LaneSpace,
			Prerequisites: []string{"rocketry", "electricity_tech"},
			Description:   "Orbital satellites for communication and surveillance.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 1.0},
				{Kind: EffectExpeditionReward, Value: 0.10},
			},
		},
		{
			Name: "Nanofabrication", Key: "nanofabrication", Code: "NANO", Emblem: "◇",
			Age: "modern_age", Lane: LaneCraft,
			Prerequisites: []string{"computers"},
			Description:   "Nanobot swarms assemble structures atom-by-atom, cutting construction costs.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.03},
			},
		},

		// === INFORMATION AGE ===
		{
			Name: "Internet", Key: "internet",
			Age: "information_age", Lane: LaneComputing,
			Prerequisites: []string{"computers", "satellite_tech"},
			Description:   "Global network connecting all of humanity.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "data", Value: 0.05},
			},
		},
		{
			Name: "Cybersecurity", Key: "cybersecurity", Code: "SECUR",
			Age: "information_age", Lane: LaneMilitary,
			Prerequisites: []string{"computers"},
			Description:   "Defense against digital threats.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
			},
		},
		{
			Name: "Social Media", Key: "social_media",
			Age: "information_age", Lane: LaneFaith,
			Prerequisites: []string{"internet"},
			Description:   "Mass digital communication platforms.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "culture", Value: 0.05},
				{Kind: EffectMechanic, Target: MechanicFestivalCost, Value: -0.20},
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
				{Kind: EffectHousing, Value: 0.05},
				{Kind: EffectOutput, Target: "food", Value: 0.05},
			},
		},
		{
			// Mid-age unlock (Pacing v2): it stands behind three of the
			// age's techs, so it comes after them.
			Name: "Internet of Things", Key: "internet_of_things", Code: "IOT",
			Age: "information_age", Lane: LaneAgriculture,
			Prerequisites: []string{"social_media", "cybersecurity", "medical_nanobots"},
			Description:   "The fridges and the tractors go online and start reporting back. Opens the Smart Farm and the Smart Complex.",
		},

		// === DIGITAL AGE ===
		{
			Name: "Machine Learning", Key: "machine_learning",
			Age: "digital_age", Lane: LaneComputing,
			Prerequisites: []string{"internet", "cybersecurity"},
			Description:   "Algorithms that learn and improve autonomously.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "data", Value: 0.05},
				{Kind: EffectResearchTime, Value: -0.03},
			},
		},
		{
			Name: "Cloud Computing", Key: "cloud_computing",
			Age: "digital_age", Lane: LaneComputing,
			Prerequisites: []string{"internet"},
			Description:   "Distributed computing at global scale.",
			Effects: []TechEffect{
				{Kind: EffectStorage, Value: 0.08},
			},
		},
		{
			Name: "Self-Replication", Key: "self_replication",
			Age: "digital_age", Lane: LaneCraft,
			Prerequisites: []string{"medical_nanobots", "machine_learning"},
			Description:   "Nanobots that build copies of themselves.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "nanobots", Value: 0.10},
				{Kind: EffectBuildTime, Value: -0.05},
			},
		},

		// === CYBERPUNK AGE ===
		{
			Name: "Neural Interface", Key: "neural_interface",
			Age: "cyberpunk_age", Lane: LaneKnowledge,
			Prerequisites: []string{"machine_learning"},
			Description:   "Direct brain-computer interface technology.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.04},
			},
		},
		{
			Name: "Blockchain", Key: "blockchain",
			Age: "cyberpunk_age", Lane: LaneTrade,
			Prerequisites: []string{"cybersecurity", "cloud_computing"},
			Description:   "Decentralized trustless systems.",
			Effects: []TechEffect{
				// No bonus on crypto yet: only the Neon Citadel makes any,
				// so there is nothing for one to raise in this age. Crypto
				// is bought at the market, which is what the fee cut helps.
				{Kind: EffectMechanic, Target: MechanicMarketFee, Value: -0.02},
			},
		},
		{
			// The Neon Citadel's keystone, with Holography behind it.
			Name: "Cybernetics", Key: "cybernetics",
			Age: "cyberpunk_age", Lane: LaneCraft,
			Prerequisites: []string{"neural_interface"},
			Description:   "Mechanical augmentation of the human body.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
			},
		},
		{
			// Mid-age unlock (Pacing v2): it stands behind Cybernetics and
			// Blockchain, so it comes during the saving-up for the wonder.
			Name: "Holography", Key: "holography",
			Age: "cyberpunk_age", Lane: LaneFaith,
			Prerequisites: []string{"cybernetics", "blockchain"},
			Description:   "Light learns to lie convincingly, and every wall becomes an ad. Opens the Holographic Theater.",
		},

		// === FUSION AGE ===
		{
			Name: "Fusion Power", Key: "fusion_power",
			Age: "fusion_age", Lane: LaneEnergy,
			Prerequisites: []string{"nuclear_fission", "cybernetics"},
			Description:   "Controlled fusion adds electricity and plasma.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "electricity", Value: 0.04},
				{Kind: EffectOutput, Target: "plasma", Value: 0.04},
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
				{Kind: EffectOutput, Target: "plasma", Value: 0.04},
			},
		},
		{
			// Paced with Plasma Physics (see there).
			Name: "Superconductors", Key: "superconductors",
			Age: "fusion_age", Lane: LaneMaterials,
			Prerequisites: []string{"plasma_physics"},
			Description:   "Zero-resistance materials raise all production and storage.",
			Effects: []TechEffect{
				{Kind: EffectStorage, Value: 0.08},
			},
		},
		{
			// Mid-age unlock (Pacing v2), paced with Plasma Physics (see
			// there).
			Name: "Maglev Transit", Key: "maglev_transit",
			Age: "fusion_age", Lane: LaneTrade,
			Prerequisites: []string{"superconductors"},
			Description:   "Superconducting rails float the freight across the city at the speed of a mild panic. Opens the Energy Exchange.",
		},

		// === SPACE AGE ===
		{
			Name: "Orbital Mechanics", Key: "orbital_mechanics",
			Age: "space_age", Lane: LaneSpace,
			Prerequisites: []string{"rocketry", "plasma_physics"},
			Description:   "Advanced spaceflight and orbital dynamics.",
			Effects: []TechEffect{
				{Kind: EffectExpeditionReward, Value: 0.10},
			},
		},
		{
			Name: "Space Mining", Key: "space_mining",
			Age: "space_age", Lane: LaneSpace,
			Prerequisites: []string{"orbital_mechanics"},
			Description:   "Asteroid and lunar resource extraction.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "titanium", Value: 0.04},
				{Kind: EffectOutput, Target: "steel", Value: 0.04},
			},
		},
		{
			Name: "Zero-G Manufacturing", Key: "zero_g_manufacturing",
			Age: "space_age", Lane: LaneCraft,
			Prerequisites: []string{"orbital_mechanics", "superconductors"},
			Description:   "Space-based manufacturing for perfect materials.",
			Effects: []TechEffect{
				{Kind: EffectBuildTime, Value: -0.06},
			},
		},

		// === INTERSTELLAR AGE ===
		{
			Name: "Warp Drive", Key: "warp_drive",
			Age: "interstellar_age", Lane: LaneSpace,
			Prerequisites: []string{"space_mining", "zero_g_manufacturing"},
			Description:   "Faster-than-light propulsion.",
			Effects: []TechEffect{
				{Kind: EffectExpeditionReward, Value: 0.10},
				{Kind: EffectOutput, Target: "dark_matter", Value: 0.04},
			},
		},
		{
			Name: "Stellar Engineering", Key: "stellar_engineering",
			Age: "interstellar_age", Lane: LaneEnergy,
			Prerequisites: []string{"warp_drive"},
			Description:   "Harnessing and shaping stars themselves.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "plasma", Value: 0.04},
				{Kind: EffectOutput, Target: "electricity", Value: 0.04},
			},
		},

		// === GALACTIC AGE ===
		{
			Name: "Galactic Navigation", Key: "galactic_navigation",
			Age: "galactic_age", Lane: LaneSpace,
			Prerequisites: []string{"warp_drive", "stellar_engineering"},
			Description:   "Charting paths across the galaxy.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "dark_matter", Value: 0.04},
				{Kind: EffectExpeditionReward, Value: 0.10},
			},
		},
		{
			Name: "Antimatter Synthesis", Key: "antimatter_synthesis",
			Age: "galactic_age", Lane: LaneEnergy,
			Prerequisites: []string{"galactic_navigation"},
			Description:   "Controlled production of antimatter.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "antimatter", Value: 0.04},
			},
		},

		// === QUANTUM AGE ===
		{
			Name: "Quantum Mechanics", Key: "quantum_mechanics",
			Age: "quantum_age", Lane: LaneKnowledge,
			Prerequisites: []string{"antimatter_synthesis"},
			Description:   "Mastery of quantum phenomena at all scales.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "quantum_flux", Value: 0.04},
			},
		},
		{
			Name: "Reality Manipulation", Key: "reality_manipulation",
			Age: "quantum_age", Lane: LaneCraft,
			Prerequisites: []string{"quantum_mechanics"},
			Description:   "Bending the fabric of spacetime.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "quantum_flux", Value: 0.04},
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
				{Kind: EffectAllOutput, Value: 0.05},
			},
		},
	}
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
