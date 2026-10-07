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
				{Kind: EffectMechanic, Target: MechanicWonderBuildTicks, Value: -0.25},
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
			Name: "Television", Key: "television", Code: "TELEV", Emblem: "▭",
			Age: "modern_age", Lane: LaneFaith,
			Prerequisites: []string{"cinema"},
			Description:   "One screen in every front room, and the whole country watching it. Opens the Monument of Ages.",
		},
		{
			Name: "Information Theory", Key: "information_theory", Code: "INFOR", Emblem: "∂",
			Age: "modern_age", Lane: LaneKnowledge,
			Prerequisites: []string{"modern_physics"},
			Description:   "Any message is a count of yes and no, and a count can be sent without loss.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.05},
			},
		},
		{
			Name: "Containerization", Key: "containerization", Code: "CONTA", Emblem: "▬",
			Age: "modern_age", Lane: LaneTrade,
			Prerequisites: []string{"corporations"},
			Description:   "One steel box fits every ship, train and truck, and the docks empty in hours.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicRouteIncome, Value: 0.10},
			},
		},
		{
			Name: "Suburbs", Key: "suburbs", Code: "SUBUR", Emblem: "▴",
			Age: "modern_age", Lane: LaneAgriculture,
			Prerequisites: []string{"green_revolution"},
			Description:   "A house, a lawn and a car for every family, an hour from where they work.",
			Effects: []TechEffect{
				{Kind: EffectHousing, Value: 0.05},
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
		{
			Name: "Titanium Alloys", Key: "titanium_alloys", Code: "TITAN", Emblem: "▨",
			Age: "modern_age", Lane: LaneMaterials,
			Prerequisites: []string{"chemical_engineering"},
			Description:   "Light, strong metal for anything that has to fly or last.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "steel", Value: 0.05},
			},
		},
		{
			Name: "Special Forces", Key: "special_forces", Code: "SPECI", Emblem: "⚜",
			Age: "modern_age", Lane: LaneMilitary,
			Prerequisites: []string{"nuclear_deterrence"},
			Description:   "Small teams that are in and out before the war is declared.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicCampaignTicks, Value: -0.15},
			},
		},
		{
			Name: "Advanced Electrics", Key: "electricity_tech", Code: "ADVEL", Emblem: "↭",
			Age: "modern_age", Lane: LaneEnergy,
			Prerequisites: []string{"nuclear_fission"},
			Description:   "High-voltage grids carry power across a continent.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "electricity", Value: 0.05},
			},
		},
		{
			Name: "Satellite Technology", Key: "satellite_tech", Code: "SATEL", Emblem: "✧",
			Age: "modern_age", Lane: LaneSpace,
			Prerequisites: []string{"rocketry", "electricity_tech"},
			Description:   "Orbital satellites for communication and surveillance.",
			Effects: []TechEffect{
				{Kind: EffectFlatOutput, Target: "data", Value: 1.0},
				{Kind: EffectExpeditionReward, Value: 0.10},
			},
		},
		{
			Name: "Computers", Key: "computers", Code: "COMP", Emblem: "⊟",
			Age: "modern_age", Lane: LaneComputing,
			Prerequisites: []string{"electricity_tech"},
			Description:   "Machines that do the arithmetic of a thousand clerks, and never tire.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.05},
				{Kind: EffectResearchTime, Value: -0.03},
			},
		},

		// === INFORMATION AGE ===
		{
			Name: "Social Media", Key: "social_media", Code: "SOCIA", Emblem: "@",
			Age: "information_age", Lane: LaneFaith,
			Prerequisites: []string{"internet"},
			Description:   "Mass digital communication platforms.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "culture", Value: 0.05},
				{Kind: EffectMechanic, Target: MechanicFestivalCost, Value: -0.20},
			},
		},
		{
			Name: "Search Engines", Key: "search_engines", Code: "SEARC", Emblem: "?",
			Age: "information_age", Lane: LaneKnowledge,
			Prerequisites: []string{"internet"},
			Description:   "Everything written down, found in half a second.",
			Effects: []TechEffect{
				{Kind: EffectResearchTime, Value: -0.03},
			},
		},
		{
			Name: "E-commerce", Key: "e_commerce", Code: "ECOMM", Emblem: "€",
			Age: "information_age", Lane: LaneTrade,
			Prerequisites: []string{"containerization", "internet"},
			Description:   "The shop is a page, and the till never closes.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicMarketFee, Value: -0.02},
			},
		},
		{
			// note: the original spec wanted this to cut worker FOOD COST, but the
			// food drain (wc.FoodCost * count in game/villagers.go) has no bonus hook
			// and threading one through WorkerManager isn't a "tiny" engine change.
			// Substituted a supported, clearly-beneficial effect instead: nanobots
			// keep the population healthier (bigger pop cap) and better fed (+food).
			Name: "Medical Nanobots", Key: "medical_nanobots", Code: "MEDIC", Emblem: "✚",
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
			Name: "Internet of Things", Key: "internet_of_things", Code: "IOT", Emblem: "⌘",
			Age: "information_age", Lane: LaneAgriculture,
			Prerequisites: []string{"social_media", "cybersecurity", "medical_nanobots"},
			Description:   "The fridges and the tractors go online and start reporting back. Opens the Smart Farm and the Smart Complex.",
		},
		{
			Name: "Embedded Systems", Key: "embedded_systems", Code: "EMBED", Emblem: "▧",
			Age: "information_age", Lane: LaneCraft,
			Prerequisites: []string{"nanofabrication"},
			Description:   "A small computer inside every machine, minding it.",
			Effects: []TechEffect{
				{Kind: EffectBuildTime, Value: -0.04},
			},
		},
		{
			Name: "Precision Mining", Key: "precision_mining", Code: "PRECI", Emblem: "↧",
			Age: "information_age", Lane: LaneMaterials,
			Prerequisites: []string{"titanium_alloys"},
			Description:   "Sensors read the rock ahead of the drill, and nothing is dug twice.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "steel", Value: 0.05},
			},
		},
		{
			Name: "Cybersecurity", Key: "cybersecurity", Code: "SECUR", Emblem: "⊘",
			Age: "information_age", Lane: LaneMilitary,
			Prerequisites: []string{"computers"},
			Description:   "Defense against digital threats.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
			},
		},
		{
			Name: "Smart Grid", Key: "smart_grid", Code: "SMART", Emblem: "⊹",
			Age: "information_age", Lane: LaneEnergy,
			Prerequisites: []string{"electricity_tech"},
			Description:   "A grid that knows where the power is wanted before the switch is thrown.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "electricity", Value: 0.05},
			},
		},
		{
			Name: "Space Stations", Key: "space_stations", Code: "STATN", Emblem: "✜",
			Age: "information_age", Lane: LaneSpace,
			Prerequisites: []string{"satellite_tech"},
			Description:   "Crews that live in orbit for months and watch the whole world turn.",
			Effects: []TechEffect{
				{Kind: EffectExpeditionReward, Value: 0.08},
			},
		},
		{
			Name: "Internet", Key: "internet", Code: "INTER", Emblem: "※",
			Age: "information_age", Lane: LaneComputing,
			Prerequisites: []string{"computers", "satellite_tech"},
			Description:   "Global network connecting all of humanity.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "data", Value: 0.05},
			},
		},

		// === DIGITAL AGE ===
		{
			Name: "Virtual Reality", Key: "virtual_reality", Code: "VR", Emblem: "◈",
			Age: "digital_age", Lane: LaneFaith,
			Prerequisites: []string{"social_media"},
			Description:   "Anywhere you like, from a chair, and nearly as good.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicMoraleCap, Value: 0.05},
			},
		},
		{
			Name: "Open Science", Key: "open_science", Code: "OPEN", Emblem: "∀",
			Age: "digital_age", Lane: LaneKnowledge,
			Prerequisites: []string{"search_engines"},
			Description:   "Every paper and every dataset, free to read the day it is finished.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.05},
			},
		},
		{
			Name: "Automated Logistics", Key: "automated_logistics", Code: "LOGIS", Emblem: "⇛",
			Age: "digital_age", Lane: LaneTrade,
			Prerequisites: []string{"e_commerce"},
			Description:   "Warehouses that pick, pack and send without a hand on the parcel.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicRouteTicks, Value: -0.15},
			},
		},
		{
			// The Digital Era's Trade capstone.
			Name: "Global Village", Key: "global_village", Code: "GLOBE", Emblem: "⊚",
			Age: "digital_age", Lane: LaneTrade, Capstone: true,
			Prerequisites: []string{"e_commerce", "social_media"},
			Description:   "Everyone is a neighbor now, and neighbors do business.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicDealSlots, Value: 1},
				{Kind: EffectMechanic, Target: MechanicAllianceCost, Value: -0.50},
			},
		},
		{
			Name: "Gene Editing", Key: "gene_editing", Code: "GENE", Emblem: "∽",
			Age: "digital_age", Lane: LaneAgriculture,
			Prerequisites: []string{"medical_nanobots"},
			Description:   "Crops rewritten a letter at a time, for drought, blight and yield.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.05},
			},
		},
		{
			Name: "Self-Replication", Key: "self_replication", Code: "SELF", Emblem: "↺",
			Age: "digital_age", Lane: LaneCraft,
			Prerequisites: []string{"medical_nanobots", "machine_learning"},
			Description:   "Nanobots that build copies of themselves.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "nanobots", Value: 0.10},
				{Kind: EffectBuildTime, Value: -0.05},
			},
		},
		{
			Name: "Nano Alloys", Key: "nano_alloys", Code: "ALLOY", Emblem: "⬢",
			Age: "digital_age", Lane: LaneMaterials,
			Prerequisites: []string{"precision_mining"},
			Description:   "Metal laid down grain by grain, with no flaw to start a crack.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "steel", Value: 0.05},
			},
		},
		{
			Name: "Drone Warfare", Key: "drone_warfare", Code: "DRONE", Emblem: "✣",
			Age: "digital_age", Lane: LaneMilitary,
			Prerequisites: []string{"cybersecurity"},
			Description:   "The pilot is a thousand miles away, and home for dinner.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
			},
		},
		{
			Name: "Grid Storage", Key: "grid_storage", Code: "STORE", Emblem: "▮",
			Age: "digital_age", Lane: LaneEnergy,
			Prerequisites: []string{"smart_grid"},
			Description:   "Batteries the size of buildings hold the noon sun for the evening.",
			Effects: []TechEffect{
				{Kind: EffectStorage, Value: 0.05},
			},
		},
		{
			Name: "Reusable Launchers", Key: "reusable_launchers", Code: "REUSE", Emblem: "⇅",
			Age: "digital_age", Lane: LaneSpace,
			Prerequisites: []string{"space_stations"},
			Description:   "The rocket lands where it took off, and flies again next week.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicExpeditionTicks, Value: -0.10},
			},
		},
		{
			Name: "Machine Learning", Key: "machine_learning", Code: "LEARN", Emblem: "∇",
			Age: "digital_age", Lane: LaneComputing,
			Prerequisites: []string{"internet", "cybersecurity"},
			Description:   "Algorithms that learn and improve autonomously.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "data", Value: 0.05},
				{Kind: EffectResearchTime, Value: -0.03},
			},
		},
		{
			Name: "Cloud Computing", Key: "cloud_computing", Code: "CLOUD", Emblem: "⌒",
			Age: "digital_age", Lane: LaneComputing,
			Prerequisites: []string{"internet"},
			Description:   "Distributed computing at global scale.",
			Effects: []TechEffect{
				{Kind: EffectStorage, Value: 0.08},
			},
		},
		{
			// The Digital Era's Computing capstone.
			Name: "General AI", Key: "general_ai", Code: "AGI", Emblem: "⊨",
			Age: "digital_age", Lane: LaneComputing, Capstone: true,
			Prerequisites: []string{"machine_learning", "cloud_computing"},
			Description:   "A machine that can be handed any question, and asks better ones back.",
			Effects: []TechEffect{
				{Kind: EffectResearchTime, Value: -0.06},
			},
		},

		// === CYBERPUNK AGE ===
		{
			// Mid-age unlock (Pacing v2): it stands behind Cybernetics and
			// Blockchain, so it comes during the saving-up for the wonder.
			Name: "Holography", Key: "holography", Code: "HOLOG", Emblem: "◬",
			Age: "cyberpunk_age", Lane: LaneFaith,
			Prerequisites: []string{"cybernetics", "blockchain"},
			Description:   "Light learns to lie convincingly, and every wall becomes an ad. Opens the Holographic Theater.",
		},
		{
			Name: "Neural Interface", Key: "neural_interface", Code: "NEURA", Emblem: "ψ",
			Age: "cyberpunk_age", Lane: LaneKnowledge,
			Prerequisites: []string{"machine_learning"},
			Description:   "Direct brain-computer interface technology.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.04},
			},
		},
		{
			// No bonus on crypto: only the Neon Citadel makes any, so there is
			// nothing for one to raise in this age. Crypto is bought at the
			// market, which is what the fee cut helps.
			Name: "Blockchain", Key: "blockchain", Code: "BLOCK", Emblem: "⋈",
			Age: "cyberpunk_age", Lane: LaneTrade,
			Prerequisites: []string{"cybersecurity", "cloud_computing"},
			Description:   "Decentralized trustless systems.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicMarketFee, Value: -0.02},
			},
		},
		{
			Name: "Synthetic Food", Key: "synthetic_food", Code: "SYNTH", Emblem: "◒",
			Age: "cyberpunk_age", Lane: LaneAgriculture,
			Prerequisites: []string{"gene_editing"},
			Description:   "Protein grown in a vat, shaped like whatever sells.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.04},
			},
		},
		{
			// The Neon Citadel's keystone, with Holography behind it.
			Name: "Cybernetics", Key: "cybernetics", Code: "CYBER", Emblem: "Ø",
			Age: "cyberpunk_age", Lane: LaneCraft,
			Prerequisites: []string{"neural_interface"},
			Description:   "Mechanical augmentation of the human body.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
			},
		},
		{
			Name: "Dark Crystal Mining", Key: "dark_crystal_mining", Code: "CRYST", Emblem: "♢",
			Age: "cyberpunk_age", Lane: LaneMaterials,
			Prerequisites: []string{"nano_alloys"},
			Description:   "Crystals that bend light the wrong way, cut from the deepest rock.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "dark_matter_crystals", Value: 0.04},
			},
		},
		{
			Name: "Augmented Soldiers", Key: "augmented_soldiers", Code: "AUGMT", Emblem: "✠",
			Age: "cyberpunk_age", Lane: LaneMilitary,
			Prerequisites: []string{"cybernetics", "drone_warfare"},
			Description:   "Soldiers rebuilt to see in the dark and carry twice the load.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
				{Kind: EffectMechanic, Target: MechanicSoldierStorage, Value: 0.10},
			},
		},
		{
			Name: "Dark Energy", Key: "dark_energy", Code: "DARKE", Emblem: "◕",
			Age: "cyberpunk_age", Lane: LaneEnergy,
			Prerequisites: []string{"grid_storage"},
			Description:   "Power drawn from the pressure that pushes the universe apart.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "electricity", Value: 0.04},
			},
		},
		{
			Name: "Lunar Outposts", Key: "lunar_outposts", Code: "LUNAR", Emblem: "☽",
			Age: "cyberpunk_age", Lane: LaneSpace,
			Prerequisites: []string{"reusable_launchers"},
			Description:   "A permanent crew on the Moon, and a harbor for everything going further.",
			Effects: []TechEffect{
				{Kind: EffectExpeditionReward, Value: 0.08},
			},
		},
		{
			Name: "Darknets", Key: "darknets", Code: "DARKN", Emblem: "▼",
			Age: "cyberpunk_age", Lane: LaneComputing,
			Prerequisites: []string{"cloud_computing"},
			Description:   "Networks under the network, where nobody is who they say.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "data", Value: 0.04},
			},
		},

		// === FUSION AGE ===
		{
			Name: "Neural Art", Key: "neural_art", Code: "ART", Emblem: "❂",
			Age: "fusion_age", Lane: LaneFaith,
			Prerequisites: []string{"holography"},
			Description:   "Art played straight into the mind, with no canvas in between.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "culture", Value: 0.04},
			},
		},
		{
			Name: "Unified Theory", Key: "unified_theory", Code: "UNIFY", Emblem: "∮",
			Age: "fusion_age", Lane: LaneKnowledge,
			Prerequisites: []string{"neural_interface", "plasma_physics"},
			Description:   "One set of equations for the very large and the very small.",
			Effects: []TechEffect{
				{Kind: EffectResearchTime, Value: -0.03},
			},
		},
		{
			// Mid-age unlock (Pacing v2), paced with Plasma Physics (see
			// there).
			Name: "Maglev Transit", Key: "maglev_transit", Code: "MAGLV", Emblem: "⇒",
			Age: "fusion_age", Lane: LaneTrade,
			Prerequisites: []string{"superconductors"},
			Description:   "Superconducting rails float the freight across the city at the speed of a mild panic. Opens the Energy Exchange.",
		},
		{
			Name: "Closed Biospheres", Key: "closed_biospheres", Code: "BIOSP", Emblem: "◠",
			Age: "fusion_age", Lane: LaneAgriculture,
			Prerequisites: []string{"synthetic_food"},
			Description:   "A sealed dome that feeds, waters and airs everyone inside it.",
			Effects: []TechEffect{
				{Kind: EffectHousing, Value: 0.05},
			},
		},
		{
			Name: "Molecular Assembly", Key: "molecular_assembly", Code: "MOLEC", Emblem: "⁂",
			Age: "fusion_age", Lane: LaneCraft,
			Prerequisites: []string{"cybernetics", "self_replication"},
			Description:   "Parts grown to shape, molecule by molecule, with nothing to cut away.",
			Effects: []TechEffect{
				{Kind: EffectBuildTime, Value: -0.05},
			},
		},
		{
			// Paced with Plasma Physics (see there).
			Name: "Superconductors", Key: "superconductors", Code: "SUPER", Emblem: "℧",
			Age: "fusion_age", Lane: LaneMaterials,
			Prerequisites: []string{"plasma_physics"},
			Description:   "Zero-resistance materials lose nothing between the reactor and the store.",
			Effects: []TechEffect{
				{Kind: EffectStorage, Value: 0.08},
			},
		},
		{
			Name: "Plasma Weapons", Key: "plasma_weapons", Code: "BEAMS", Emblem: "☇",
			Age: "fusion_age", Lane: LaneMilitary,
			Prerequisites: []string{"augmented_soldiers", "plasma_physics"},
			Description:   "A bolt of contained star, aimed.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
			},
		},
		{
			Name: "Fusion Power", Key: "fusion_power", Code: "FUSN", Emblem: "✹",
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
			Name: "Plasma Physics", Key: "plasma_physics", Code: "PLASM", Emblem: "≀",
			Age: "fusion_age", Lane: LaneEnergy,
			Prerequisites: []string{"fusion_power"},
			Description:   "Mastery of superheated matter states.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "plasma", Value: 0.04},
			},
		},
		{
			Name: "Fusion Drives", Key: "fusion_drives", Code: "DRIVE", Emblem: "⇑",
			Age: "fusion_age", Lane: LaneSpace,
			Prerequisites: []string{"fusion_power", "lunar_outposts"},
			Description:   "A torch that burns for weeks: the outer planets in a season.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicExpeditionTicks, Value: -0.10},
			},
		},
		{
			Name: "Quantum Networking", Key: "quantum_networking", Code: "QNET", Emblem: "⊶",
			Age: "fusion_age", Lane: LaneComputing,
			Prerequisites: []string{"darknets"},
			Description:   "Two machines that share a state, and a line nobody can tap.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "data", Value: 0.04},
			},
		},

		// === SPACE AGE ===
		{
			Name: "Overview Effect", Key: "overview_effect", Code: "OVIEW", Emblem: "♁",
			Age: "space_age", Lane: LaneFaith,
			Prerequisites: []string{"neural_art"},
			Description:   "Everyone who has seen the whole world from outside comes home changed.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicMoraleCap, Value: 0.05},
			},
		},
		{
			Name: "Deep Space Astronomy", Key: "deep_space_astronomy", Code: "ASTRO", Emblem: "☄",
			Age: "space_age", Lane: LaneKnowledge,
			Prerequisites: []string{"unified_theory"},
			Description:   "Telescopes beyond the air and the glare, looking back to the first light.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "knowledge", Value: 0.04},
			},
		},
		{
			Name: "Asteroid Claims", Key: "asteroid_claims", Code: "CLAIM", Emblem: "◊",
			Age: "space_age", Lane: LaneTrade,
			Prerequisites: []string{"maglev_transit"},
			Description:   "A rock, a registry number and a company that owns what is inside it.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicRouteIncome, Value: 0.10},
			},
		},
		{
			// The Neon Era's Trade capstone.
			Name: "Stellar Cartography", Key: "stellar_cartography", Code: "CHART", Emblem: "✦",
			Age: "space_age", Lane: LaneTrade, Capstone: true,
			Prerequisites: []string{"asteroid_claims", "deep_space_astronomy"},
			Description:   "Every star within reach, charted with its worlds and its hazards.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicExpeditionTicks, Value: -0.20},
				{Kind: EffectExpeditionReward, Value: 0.10},
			},
		},
		{
			Name: "Hydroponics", Key: "hydroponics", Code: "HYDRO", Emblem: "⋎",
			Age: "space_age", Lane: LaneAgriculture,
			Prerequisites: []string{"closed_biospheres"},
			Description:   "Roots in running water under lamps: a harvest every month, anywhere.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.04},
			},
		},
		{
			Name: "Zero-G Manufacturing", Key: "zero_g_manufacturing", Code: "ZEROG", Emblem: "∅",
			Age: "space_age", Lane: LaneCraft,
			Prerequisites: []string{"orbital_mechanics", "superconductors"},
			Description:   "Space-based manufacturing for perfect materials.",
			Effects: []TechEffect{
				{Kind: EffectBuildTime, Value: -0.06},
			},
		},
		{
			Name: "Asteroid Refining", Key: "asteroid_refining", Code: "REFIN", Emblem: "◣",
			Age: "space_age", Lane: LaneMaterials,
			Prerequisites: []string{"space_mining"},
			Description:   "Ore smelted where it is mined, with the Sun for a furnace.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "titanium", Value: 0.04},
			},
		},
		{
			Name: "Orbital Defense", Key: "orbital_defense", Code: "ODEF", Emblem: "▽",
			Age: "space_age", Lane: LaneMilitary,
			Prerequisites: []string{"plasma_weapons", "orbital_mechanics"},
			Description:   "Nothing crosses the sky without being seen, and nothing lands without leave.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicRaidLoss, Value: -0.10},
			},
		},
		{
			Name: "Orbital Solar", Key: "orbital_solar", Code: "SOLAR", Emblem: "✷",
			Age: "space_age", Lane: LaneEnergy,
			Prerequisites: []string{"plasma_physics"},
			Description:   "Mirrors in permanent daylight, beaming their catch down.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "electricity", Value: 0.04},
				{Kind: EffectOutput, Target: "plasma", Value: 0.04},
			},
		},
		{
			Name: "Orbital Mechanics", Key: "orbital_mechanics", Code: "ORBIT", Emblem: "☊",
			Age: "space_age", Lane: LaneSpace,
			Prerequisites: []string{"rocketry", "plasma_physics"},
			Description:   "Advanced spaceflight and orbital dynamics.",
			Effects: []TechEffect{
				{Kind: EffectExpeditionReward, Value: 0.10},
			},
		},
		{
			Name: "Space Mining", Key: "space_mining", Code: "SPACE", Emblem: "◮",
			Age: "space_age", Lane: LaneSpace,
			Prerequisites: []string{"orbital_mechanics"},
			Description:   "Asteroid and lunar resource extraction.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "titanium", Value: 0.04},
				{Kind: EffectOutput, Target: "steel", Value: 0.04},
			},
		},
		{
			// The Neon Era's Space capstone.
			Name: "Space Elevator", Key: "space_elevator", Code: "ELEV", Emblem: "↥",
			Age: "space_age", Lane: LaneSpace, Capstone: true,
			Prerequisites: []string{"zero_g_manufacturing", "space_mining"},
			Description:   "A cable from the ground to orbit: freight goes up for the price of the electricity.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.04},
				{Kind: EffectStorage, Value: 0.08},
			},
		},
		{
			Name: "Orbital Relays", Key: "orbital_relays", Code: "RELAY", Emblem: "↹",
			Age: "space_age", Lane: LaneComputing,
			Prerequisites: []string{"quantum_networking", "orbital_mechanics"},
			Description:   "A ring of relays, and no corner of the system out of touch.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "data", Value: 0.04},
			},
		},

		// === INTERSTELLAR AGE ===
		{
			Name: "Void Contemplation", Key: "void_contemplation", Code: "VOID", Emblem: "○",
			Age: "interstellar_age", Lane: LaneFaith,
			Prerequisites: []string{"overview_effect"},
			Description:   "Years between the stars teach a crew to sit with the dark, and to bargain with it.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicAppeaseCost, Value: -0.20},
			},
		},
		{
			Name: "Xenology", Key: "xenology", Code: "XENO", Emblem: "ξ",
			Age: "interstellar_age", Lane: LaneKnowledge,
			Prerequisites: []string{"deep_space_astronomy"},
			Description:   "The study of life that owes nothing to ours.",
			Effects: []TechEffect{
				{Kind: EffectResearchTime, Value: -0.03},
			},
		},
		{
			Name: "Interstellar Trade", Key: "interstellar_trade", Code: "TRADE", Emblem: "⊛",
			Age: "interstellar_age", Lane: LaneTrade,
			Prerequisites: []string{"asteroid_claims", "warp_drive"},
			Description:   "Cargo that outruns the news of its own departure.",
		},
		{
			Name: "Protein Synthesis", Key: "protein_synthesis", Code: "PROTN", Emblem: "∾",
			Age: "interstellar_age", Lane: LaneAgriculture,
			Prerequisites: []string{"hydroponics"},
			Description:   "Food built from its elements, to any recipe, with no field at all.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.04},
				{Kind: EffectHousing, Value: 0.04},
			},
		},
		{
			Name: "Hull Printing", Key: "hull_printing", Code: "HULL", Emblem: "▱",
			Age: "interstellar_age", Lane: LaneCraft,
			Prerequisites: []string{"zero_g_manufacturing"},
			Description:   "A ship's hull laid down in one piece, in the dark, by machines.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.02},
				{Kind: EffectBuildTime, Value: -0.04},
			},
		},
		{
			Name: "Stellar Core Mining", Key: "stellar_core_mining", Code: "CORE", Emblem: "☉",
			Age: "interstellar_age", Lane: LaneMaterials,
			Prerequisites: []string{"asteroid_refining"},
			Description:   "Dead stars are mostly metal, if you can stand the heat of getting there.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "titanium", Value: 0.04},
			},
		},
		{
			Name: "Fleet Doctrine", Key: "fleet_doctrine", Code: "FLEET", Emblem: "➤",
			Age: "interstellar_age", Lane: LaneMilitary,
			Prerequisites: []string{"orbital_defense"},
			Description:   "How to fight a war where the order arrives after the battle.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
			},
		},
		{
			Name: "Stellar Engineering", Key: "stellar_engineering", Code: "STELL", Emblem: "✫",
			Age: "interstellar_age", Lane: LaneEnergy,
			Prerequisites: []string{"warp_drive"},
			Description:   "Harnessing and shaping stars themselves.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "plasma", Value: 0.04},
				{Kind: EffectOutput, Target: "electricity", Value: 0.04},
			},
		},
		{
			Name: "Warp Drive", Key: "warp_drive", Code: "WARP", Emblem: "≫",
			Age: "interstellar_age", Lane: LaneSpace,
			Prerequisites: []string{"space_mining", "zero_g_manufacturing"},
			Description:   "Faster-than-light propulsion.",
			Effects: []TechEffect{
				{Kind: EffectExpeditionReward, Value: 0.10},
				{Kind: EffectOutput, Target: "dark_matter", Value: 0.04},
			},
		},
		{
			Name: "Galactic Network", Key: "galactic_network", Code: "GNET", Emblem: "✺",
			Age: "interstellar_age", Lane: LaneComputing,
			Prerequisites: []string{"orbital_relays"},
			Description:   "Every colony on one network, whatever the distance.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "data", Value: 0.04},
			},
		},

		// === GALACTIC AGE ===
		{
			Name: "Galactic Memory", Key: "galactic_memory", Code: "MEMRY", Emblem: "✪",
			Age: "galactic_age", Lane: LaneFaith,
			Prerequisites: []string{"void_contemplation"},
			Description:   "Every song, story and quarrel of every world, kept where none can be lost.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "culture", Value: 0.04},
			},
		},
		{
			Name: "Cosmology", Key: "cosmology", Code: "COSMO", Emblem: "∞",
			Age: "galactic_age", Lane: LaneKnowledge,
			Prerequisites: []string{"xenology"},
			Description:   "Where everything came from, and how long it has left.",
			Effects: []TechEffect{
				{Kind: EffectResearchTime, Value: -0.03},
			},
		},
		{
			Name: "Federation Charter", Key: "federation_charter", Code: "FED", Emblem: "⊎",
			Age: "galactic_age", Lane: LaneTrade,
			Prerequisites: []string{"interstellar_trade", "xenology"},
			Description:   "One law of trade and passage for every signatory, whatever they breathe.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicAllianceBonus, Value: 0.25},
				{Kind: EffectMechanic, Target: MechanicDealSlots, Value: 1},
			},
		},
		{
			Name: "Matter Conversion", Key: "matter_conversion", Code: "MATTR", Emblem: "⇌",
			Age: "galactic_age", Lane: LaneAgriculture,
			Prerequisites: []string{"protein_synthesis"},
			Description:   "Rock in, bread out. And walls, and air.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.04},
				{Kind: EffectHousing, Value: 0.04},
			},
		},
		{
			Name: "Dyson Engineering", Key: "dyson_engineering", Code: "DYSON", Emblem: "◌",
			Age: "galactic_age", Lane: LaneCraft,
			Prerequisites: []string{"stellar_engineering", "hull_printing"},
			Description:   "Building at the scale of a star's whole output.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.02},
			},
		},
		{
			Name: "Neutron Mining", Key: "neutron_mining", Code: "NEUTN", Emblem: "⊝",
			Age: "galactic_age", Lane: LaneMaterials,
			Prerequisites: []string{"stellar_core_mining"},
			Description:   "A teaspoon of the stuff outweighs a mountain, and it is all ore.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "antimatter", Value: 0.04},
			},
		},
		{
			Name: "Armada Command", Key: "armada_command", Code: "ARMAD", Emblem: "✯",
			Age: "galactic_age", Lane: LaneMilitary,
			Prerequisites: []string{"fleet_doctrine"},
			Description:   "A thousand ships on one mind's orders.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
			},
		},
		{
			Name: "Antimatter Synthesis", Key: "antimatter_synthesis", Code: "ANTIM", Emblem: "⊖",
			Age: "galactic_age", Lane: LaneEnergy,
			Prerequisites: []string{"galactic_navigation"},
			Description:   "Controlled production of antimatter.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "antimatter", Value: 0.04},
			},
		},
		{
			Name: "Galactic Navigation", Key: "galactic_navigation", Code: "GNAV", Emblem: "✵",
			Age: "galactic_age", Lane: LaneSpace,
			Prerequisites: []string{"warp_drive", "stellar_engineering"},
			Description:   "Charting paths across the galaxy.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "dark_matter", Value: 0.04},
				{Kind: EffectExpeditionReward, Value: 0.10},
			},
		},
		{
			Name: "Terraforming", Key: "terraforming", Code: "TERRA", Emblem: "◓",
			Age: "galactic_age", Lane: LaneSpace,
			Prerequisites: []string{"warp_drive"},
			Description:   "A dead world given air and seas, and a few centuries to settle.",
			Effects: []TechEffect{
				{Kind: EffectHousing, Value: 0.06},
			},
		},
		{
			Name: "Mind Uploading", Key: "mind_uploading", Code: "MIND", Emblem: "⇪",
			Age: "galactic_age", Lane: LaneComputing,
			Prerequisites: []string{"galactic_network"},
			Description:   "A person, copied out of the body and still arguing.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "data", Value: 0.04},
			},
		},

		// === QUANTUM AGE ===
		{
			Name: "Reality Art", Key: "reality_art", Code: "RART", Emblem: "❈",
			Age: "quantum_age", Lane: LaneFaith,
			Prerequisites: []string{"galactic_memory"},
			Description:   "Works made of what might have happened, shown beside what did.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "culture", Value: 0.04},
			},
		},
		{
			Name: "Quantum Mechanics", Key: "quantum_mechanics", Code: "QUANT", Emblem: "ℚ",
			Age: "quantum_age", Lane: LaneKnowledge,
			Prerequisites: []string{"antimatter_synthesis"},
			Description:   "Mastery of quantum phenomena at all scales.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "quantum_flux", Value: 0.04},
			},
		},
		{
			// The Cosmic Era's Knowledge capstone.
			Name: "Timeless Archive", Key: "timeless_archive", Code: "ARCHV", Emblem: "☰",
			Age: "quantum_age", Lane: LaneKnowledge, Capstone: true,
			Prerequisites: []string{"cosmology", "quantum_mechanics"},
			Description:   "Everything ever known, and every answer already looked up.",
			Effects: []TechEffect{
				{Kind: EffectResearchTime, Value: -0.06},
			},
		},
		{
			Name: "Probability Markets", Key: "probability_markets", Code: "PMKT", Emblem: "‰",
			Age: "quantum_age", Lane: LaneTrade,
			Prerequisites: []string{"federation_charter"},
			Description:   "A price on every outcome, and a buyer for the ones that never happen.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicMarketFee, Value: -0.02},
			},
		},
		{
			Name: "Quantum Cultivation", Key: "quantum_cultivation", Code: "QCULT", Emblem: "❃",
			Age: "quantum_age", Lane: LaneAgriculture,
			Prerequisites: []string{"matter_conversion"},
			Description:   "Every harvest that could have been, and you pick the best one.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "food", Value: 0.04},
			},
		},
		{
			Name: "Reality Manipulation", Key: "reality_manipulation", Code: "REALI", Emblem: "≋",
			Age: "quantum_age", Lane: LaneCraft,
			Prerequisites: []string{"quantum_mechanics"},
			Description:   "Bending the fabric of spacetime.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "quantum_flux", Value: 0.04},
			},
		},
		{
			// The Cosmic Era's Craft capstone.
			Name: "Reality Engineering", Key: "reality_engineering", Code: "RENG", Emblem: "⊠",
			Age: "quantum_age", Lane: LaneCraft, Capstone: true,
			Prerequisites: []string{"reality_manipulation", "dyson_engineering"},
			Description:   "The plans are drawn, and the building has always been there.",
			Effects: []TechEffect{
				{Kind: EffectBuildCost, Value: -0.04},
				{Kind: EffectBuildTime, Value: -0.08},
			},
		},
		{
			Name: "Quantum Metallurgy", Key: "quantum_metallurgy", Code: "QMET", Emblem: "◪",
			Age: "quantum_age", Lane: LaneMaterials,
			Prerequisites: []string{"neutron_mining"},
			Description:   "Metals that are only there when they are needed.",
			Effects: []TechEffect{
				{Kind: EffectOutput, Target: "antimatter", Value: 0.04},
			},
		},
		{
			Name: "Probability Warfare", Key: "probability_warfare", Code: "PWAR", Emblem: "⚄",
			Age: "quantum_age", Lane: LaneMilitary,
			Prerequisites: []string{"armada_command"},
			Description:   "The battle is fought in every way at once, and you keep the one you won.",
			Effects: []TechEffect{
				{Kind: EffectMilitaryPower, Value: 0.10},
				{Kind: EffectMechanic, Target: MechanicCampaignReward, Value: 0.15},
			},
		},
		{
			Name: "Zero-Point Energy", Key: "zero_point_energy", Code: "ZPE", Emblem: "∘",
			Age: "quantum_age", Lane: LaneEnergy,
			Prerequisites: []string{"antimatter_synthesis"},
			Description:   "Power from the hum of empty space, which never runs down.",
		},
		{
			Name: "Wormholes", Key: "wormholes", Code: "WORM", Emblem: "⌀",
			Age: "quantum_age", Lane: LaneSpace,
			Prerequisites: []string{"galactic_navigation"},
			Description:   "A door in one sky that opens on another.",
			Effects: []TechEffect{
				{Kind: EffectMechanic, Target: MechanicRouteTicks, Value: -0.20},
				{Kind: EffectMechanic, Target: MechanicExpeditionTicks, Value: -0.15},
			},
		},
		{
			Name: "Quantum Computing", Key: "quantum_computing", Code: "QCOMP", Emblem: "ℂ",
			Age: "quantum_age", Lane: LaneComputing,
			Prerequisites: []string{"clockwork_automation", "quantum_mechanics"},
			Description:   "Quantum processing raises game speed.",
			Effects: []TechEffect{
				{Kind: EffectGameSpeed, Value: 0.15},
			},
		},

		// === TRANSCENDENT AGE ===
		{
			Name: "Transcendence", Key: "transcendence", Code: "TRANS", Emblem: "Ω",
			Age: "transcendent_age", Lane: LaneFaith,
			Prerequisites: []string{"reality_manipulation"},
			Description:   "A civilization beyond physical limits.",
			Effects: []TechEffect{
				{Kind: EffectAllOutput, Value: 0.05},
			},
		},
		{
			Name: "Omniversal Exchange", Key: "omniversal_exchange", Code: "OMNIX", Emblem: "⇔",
			Age: "transcendent_age", Lane: LaneTrade,
			Prerequisites: []string{"probability_markets"},
			Description:   "A market between every world that is and every world that might have been.",
		},
		{
			Name: "Singularity Engineering", Key: "singularity_engineering", Code: "SING", Emblem: "⊡",
			Age: "transcendent_age", Lane: LaneCraft,
			Prerequisites: []string{"reality_manipulation"},
			Description:   "A whole economy folded into a point that does the work.",
		},
		{
			Name: "Omniversal Command", Key: "omniversal_command", Code: "OMNIC", Emblem: "♚",
			Age: "transcendent_age", Lane: LaneMilitary,
			Prerequisites: []string{"probability_warfare"},
			Description:   "One council for every army in every version of events.",
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
