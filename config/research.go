package config

// TechDef defines a technology in the tech tree
type TechDef struct {
	Name          string
	Key           string
	Age           string   // minimum age to research
	Cost          float64  // knowledge cost
	Prerequisites []string // tech keys that must be researched first
	Effects       []Effect
	Description   string
	ResearchTicks int // how many ticks to complete (0 = instant)
}

// Technologies returns all tech tree definitions
// Organized by age, with branching prerequisites
func Technologies() []TechDef {
	return []TechDef{
		// === PRIMITIVE AGE ===
		{
			Name: "Tool Making", Key: "tool_making",
			Age: "primitive_age", Cost: 10, ResearchTicks: 5,
			Description: "Primitive stone tools improve gathering efficiency.",
			Effects: []Effect{
				{Type: "bonus", Target: "gather_rate", Value: 0.15},
			},
		},
		{
			Name: "Fire Mastery", Key: "fire_mastery",
			Age: "primitive_age", Cost: 15, ResearchTicks: 8,
			Prerequisites: []string{"tool_making"},
			Description: "Control of fire improves food preservation and warmth.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 0.1},
			},
		},

		// === STONE AGE ===
		{
			Name: "Stoneworking", Key: "stoneworking",
			Age: "stone_age", Cost: 25, ResearchTicks: 10,
			Prerequisites: []string{"tool_making"},
			Description: "Cutting and shaping stone for construction.",
			Effects: []Effect{
				{Type: "bonus", Target: "stone_rate", Value: 0.2},
			},
		},
		{
			Name: "Animal Husbandry", Key: "animal_husbandry",
			Age: "stone_age", Cost: 30, ResearchTicks: 12,
			Prerequisites: []string{"fire_mastery"},
			Description: "Domesticating animals for food and labor.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 0.2},
			},
		},
		{
			Name: "Pottery", Key: "pottery",
			Age: "stone_age", Cost: 20, ResearchTicks: 8,
			Prerequisites: []string{"fire_mastery"},
			Description: "Clay vessels for storage and trade.",
			Effects: []Effect{
				{Type: "storage", Target: "all", Value: 25},
			},
		},
		{
			Name: "Primitive Writing", Key: "primitive_writing",
			Age: "stone_age", Cost: 40, ResearchTicks: 15,
			Prerequisites: []string{"pottery"},
			Description: "Early symbols enable knowledge transfer.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 0.3},
			},
		},

		// === BRONZE AGE ===
		{
			Name: "Bronze Working", Key: "bronze_working",
			Age: "bronze_age", Cost: 60, ResearchTicks: 15,
			Prerequisites: []string{"stoneworking"},
			Description: "Alloying copper and tin creates durable tools.",
			Effects: []Effect{
				{Type: "bonus", Target: "iron_rate", Value: 0.2},
				{Type: "bonus", Target: "gather_rate", Value: 0.1},
			},
		},
		{
			Name: "Agriculture", Key: "agriculture",
			Age: "bronze_age", Cost: 50, ResearchTicks: 12,
			Prerequisites: []string{"animal_husbandry"},
			Description: "Systematic farming dramatically increases food output.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 0.5},
			},
		},
		{
			Name: "Currency", Key: "currency",
			Age: "bronze_age", Cost: 70, ResearchTicks: 15,
			Prerequisites: []string{"primitive_writing"},
			Description: "Standardized money enables trade.",
			Effects: []Effect{
				{Type: "bonus", Target: "gold_rate", Value: 0.3},
			},
		},
		{
			Name: "Masonry", Key: "masonry",
			Age: "bronze_age", Cost: 55, ResearchTicks: 12,
			Prerequisites: []string{"stoneworking"},
			Description: "Advanced stone construction techniques.",
			Effects: []Effect{
				{Type: "storage", Target: "all", Value: 50},
			},
		},
		{
			Name: "Military Tactics", Key: "military_tactics",
			Age: "bronze_age", Cost: 80, ResearchTicks: 18,
			Prerequisites: []string{"bronze_working"},
			Description: "Organized warfare and defense strategies.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 0.2},
			},
		},

		// === IRON AGE ===
		{
			Name: "Iron Smelting", Key: "iron_smelting",
			Age: "iron_age", Cost: 120, ResearchTicks: 20,
			Prerequisites: []string{"bronze_working"},
			Description: "Smelting iron ore into usable metal.",
			Effects: []Effect{
				{Type: "bonus", Target: "iron_rate", Value: 0.4},
				{Type: "production", Target: "iron", Value: 0.2},
			},
		},
		{
			Name: "Road Building", Key: "road_building",
			Age: "iron_age", Cost: 100, ResearchTicks: 15,
			Prerequisites: []string{"masonry"},
			Description: "Paved roads improve trade and movement.",
			Effects: []Effect{
				{Type: "bonus", Target: "gold_rate", Value: 0.2},
				{Type: "bonus", Target: "gather_rate", Value: 0.1},
			},
		},
		{
			Name: "Mathematics", Key: "mathematics",
			Age: "iron_age", Cost: 150, ResearchTicks: 20,
			Prerequisites: []string{"primitive_writing", "currency"},
			Description: "Advanced calculation enables engineering.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 0.5},
			},
		},
		{
			Name: "Siege Warfare", Key: "siege_warfare",
			Age: "iron_age", Cost: 140, ResearchTicks: 18,
			Prerequisites: []string{"military_tactics"},
			Description: "Siege engines and fortification techniques.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 0.3},
			},
		},

		// === MEDIEVAL AGE ===
		{
			Name: "Steel Forging", Key: "steel_forging",
			Age: "medieval_age", Cost: 250, ResearchTicks: 25,
			Prerequisites: []string{"iron_smelting"},
			Description: "Refining iron into steel for superior tools and weapons.",
			Effects: []Effect{
				{Type: "production", Target: "steel", Value: 0.1},
				{Type: "bonus", Target: "iron_rate", Value: 0.3},
			},
		},
		{
			Name: "Theology", Key: "theology",
			Age: "medieval_age", Cost: 200, ResearchTicks: 20,
			Prerequisites: []string{"primitive_writing"},
			Description: "Organized religion provides faith and social cohesion.",
			Effects: []Effect{
				{Type: "production", Target: "faith", Value: 0.3},
			},
		},
		{
			Name: "Banking", Key: "banking",
			Age: "medieval_age", Cost: 300, ResearchTicks: 22,
			Prerequisites: []string{"currency", "mathematics"},
			Description: "Financial institutions multiply gold generation.",
			Effects: []Effect{
				{Type: "bonus", Target: "gold_rate", Value: 0.5},
				{Type: "storage", Target: "gold", Value: 100},
			},
		},
		{
			Name: "Feudalism", Key: "feudalism",
			Age: "medieval_age", Cost: 220, ResearchTicks: 18,
			Prerequisites: []string{"military_tactics"},
			Description: "Feudal system improves population management.",
			Effects: []Effect{
				{Type: "capacity", Target: "population", Value: 5},
			},
		},
		{
			Name: "Alchemy", Key: "alchemy",
			Age: "medieval_age", Cost: 280, ResearchTicks: 25,
			Prerequisites: []string{"mathematics"},
			Description: "Proto-chemistry yields material insights.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 0.4},
				{Type: "production", Target: "gold", Value: 0.1},
			},
		},

		// === RENAISSANCE AGE ===
		{
			Name: "Printing Press", Key: "printing_press",
			Age: "renaissance_age", Cost: 500, ResearchTicks: 30,
			Prerequisites: []string{"theology", "alchemy"},
			Description: "Mass production of books accelerates knowledge.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 1.0},
				{Type: "production", Target: "culture", Value: 0.3},
			},
		},
		{
			Name: "Navigation", Key: "navigation",
			Age: "renaissance_age", Cost: 450, ResearchTicks: 25,
			Prerequisites: []string{"mathematics", "road_building"},
			Description: "Ocean navigation opens new trade routes.",
			Effects: []Effect{
				{Type: "bonus", Target: "gold_rate", Value: 0.5},
				{Type: "bonus", Target: "expedition_reward", Value: 0.3},
			},
		},
		{
			Name: "Gunpowder", Key: "gunpowder",
			Age: "renaissance_age", Cost: 550, ResearchTicks: 28,
			Prerequisites: []string{"alchemy", "siege_warfare"},
			Description: "Explosive weaponry transforms warfare.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 0.5},
			},
		},
		{
			Name: "Patronage", Key: "patronage",
			Age: "renaissance_age", Cost: 400, ResearchTicks: 22,
			Prerequisites: []string{"banking"},
			Description: "Wealthy patrons fund arts and science.",
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 0.5},
				{Type: "production", Target: "knowledge", Value: 0.3},
			},
		},

		// === INDUSTRIAL AGE ===
		{
			Name: "Steam Power", Key: "steam_power",
			Age: "industrial_age", Cost: 1000, ResearchTicks: 35,
			Prerequisites: []string{"steel_forging"},
			Description: "Steam engines revolutionize production.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.3},
			},
		},
		{
			Name: "Industrialization", Key: "industrialization",
			Age: "industrial_age", Cost: 1200, ResearchTicks: 40,
			Prerequisites: []string{"steam_power"},
			Description: "Factory systems massively increase output.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.5},
				{Type: "production", Target: "steel", Value: 0.5},
			},
		},
		{
			Name: "Railroads", Key: "railroads",
			Age: "industrial_age", Cost: 900, ResearchTicks: 30,
			Prerequisites: []string{"steam_power", "road_building"},
			Description: "Rail networks connect your civilization.",
			Effects: []Effect{
				{Type: "bonus", Target: "gold_rate", Value: 1.0},
				{Type: "storage", Target: "all", Value: 200},
			},
		},
		{
			Name: "Rifling", Key: "rifling",
			Age: "industrial_age", Cost: 800, ResearchTicks: 25,
			Prerequisites: []string{"gunpowder"},
			Description: "Precision firearms improve military effectiveness.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 0.5},
			},
		},

		// === MODERN AGE ===
		{
			Name: "Electricity", Key: "electricity_tech",
			Age: "modern_age", Cost: 2000, ResearchTicks: 45,
			Prerequisites: []string{"industrialization"},
			Description: "Electrical power transforms civilization.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.5},
				{Type: "production", Target: "electricity", Value: 1.0},
			},
		},
		{
			Name: "Computers", Key: "computers",
			Age: "modern_age", Cost: 3000, ResearchTicks: 50,
			Prerequisites: []string{"electricity_tech", "mathematics"},
			Description: "Digital computing accelerates all research.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 2.0},
			},
		},
		{
			Name: "Nuclear Power", Key: "nuclear_power",
			Age: "modern_age", Cost: 5000, ResearchTicks: 60,
			Prerequisites: []string{"electricity_tech"},
			Description: "Harnessing atomic energy for massive power generation.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 5.0},
				{Type: "bonus", Target: "production_all", Value: 0.3},
			},
		},
		{
			Name: "Rocketry", Key: "rocketry",
			Age: "modern_age", Cost: 4000, ResearchTicks: 55,
			Prerequisites: []string{"rifling", "industrialization"},
			Description: "Rocket technology enables space exploration.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 1.0},
				{Type: "bonus", Target: "expedition_reward", Value: 0.5},
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

// TechsByAge returns techs grouped by age
func TechsByAge() map[string][]TechDef {
	m := make(map[string][]TechDef)
	for _, t := range Technologies() {
		m[t.Age] = append(m[t.Age], t)
	}
	return m
}
