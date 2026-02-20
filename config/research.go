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
		// === PRIMITIVE AGE === (~1 min each)
		{
			Name: "Tool Making", Key: "tool_making",
			Age: "primitive_age", Cost: 10, ResearchTicks: 30,
			Description: "Primitive stone tools improve gathering efficiency.",
			Effects: []Effect{
				{Type: "bonus", Target: "gather_rate", Value: 0.15},
			},
		},
		{
			Name: "Fire Mastery", Key: "fire_mastery",
			Age: "primitive_age", Cost: 15, ResearchTicks: 40,
			Prerequisites: []string{"tool_making"},
			Description: "Control of fire improves food preservation and warmth.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 0.1},
			},
		},

		// === STONE AGE === (~2 min each)
		{
			Name: "Stoneworking", Key: "stoneworking",
			Age: "stone_age", Cost: 25, ResearchTicks: 50,
			Prerequisites: []string{"tool_making"},
			Description: "Cutting and shaping stone for construction.",
			Effects: []Effect{
				{Type: "bonus", Target: "stone_rate", Value: 0.2},
			},
		},
		{
			Name: "Animal Husbandry", Key: "animal_husbandry",
			Age: "stone_age", Cost: 30, ResearchTicks: 55,
			Prerequisites: []string{"fire_mastery"},
			Description: "Domesticating animals for food and labor.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 0.2},
			},
		},
		{
			Name: "Pottery", Key: "pottery",
			Age: "stone_age", Cost: 20, ResearchTicks: 45,
			Prerequisites: []string{"fire_mastery"},
			Description: "Clay vessels for storage and trade.",
			Effects: []Effect{
				{Type: "storage", Target: "all", Value: 25},
			},
		},
		{
			Name: "Primitive Writing", Key: "primitive_writing",
			Age: "stone_age", Cost: 40, ResearchTicks: 60,
			Prerequisites: []string{"pottery"},
			Description: "Early symbols enable knowledge transfer.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 0.1},
			},
		},

		// === BRONZE AGE === (~3 min each)
		{
			Name: "Bronze Working", Key: "bronze_working",
			Age: "bronze_age", Cost: 60, ResearchTicks: 75,
			Prerequisites: []string{"stoneworking"},
			Description: "Alloying copper and tin creates durable tools.",
			Effects: []Effect{
				{Type: "bonus", Target: "iron_rate", Value: 0.2},
				{Type: "bonus", Target: "gather_rate", Value: 0.1},
			},
		},
		{
			Name: "Agriculture", Key: "agriculture",
			Age: "bronze_age", Cost: 50, ResearchTicks: 70,
			Prerequisites: []string{"animal_husbandry"},
			Description: "Systematic farming dramatically increases food output.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 0.5},
			},
		},
		{
			Name: "Currency", Key: "currency",
			Age: "bronze_age", Cost: 70, ResearchTicks: 80,
			Prerequisites: []string{"primitive_writing"},
			Description: "Standardized money enables trade.",
			Effects: []Effect{
				{Type: "bonus", Target: "gold_rate", Value: 0.3},
			},
		},
		{
			Name: "Masonry", Key: "masonry",
			Age: "bronze_age", Cost: 55, ResearchTicks: 70,
			Prerequisites: []string{"stoneworking"},
			Description: "Advanced stone construction techniques.",
			Effects: []Effect{
				{Type: "storage", Target: "all", Value: 50},
			},
		},
		{
			Name: "Military Tactics", Key: "military_tactics",
			Age: "bronze_age", Cost: 80, ResearchTicks: 90,
			Prerequisites: []string{"bronze_working"},
			Description: "Organized warfare and defense strategies.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 0.2},
			},
		},

		// === IRON AGE === (~4 min each)
		{
			Name: "Iron Smelting", Key: "iron_smelting",
			Age: "iron_age", Cost: 120, ResearchTicks: 110,
			Prerequisites: []string{"bronze_working"},
			Description: "Smelting iron ore into usable metal.",
			Effects: []Effect{
				{Type: "bonus", Target: "iron_rate", Value: 0.4},
				{Type: "production", Target: "iron", Value: 0.2},
			},
		},
		{
			Name: "Road Building", Key: "road_building",
			Age: "iron_age", Cost: 100, ResearchTicks: 95,
			Prerequisites: []string{"masonry"},
			Description: "Paved roads improve trade and movement.",
			Effects: []Effect{
				{Type: "bonus", Target: "gold_rate", Value: 0.2},
				{Type: "bonus", Target: "gather_rate", Value: 0.1},
			},
		},
		{
			Name: "Mathematics", Key: "mathematics",
			Age: "iron_age", Cost: 150, ResearchTicks: 120,
			Prerequisites: []string{"primitive_writing", "currency"},
			Description: "Advanced calculation enables engineering.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 0.2},
			},
		},
		{
			Name: "Siege Warfare", Key: "siege_warfare",
			Age: "iron_age", Cost: 140, ResearchTicks: 105,
			Prerequisites: []string{"military_tactics"},
			Description: "Siege engines and fortification techniques.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 0.3},
			},
		},

		// === CLASSICAL AGE === (~5 min each)
		{
			Name: "Philosophy", Key: "philosophy",
			Age: "classical_age", Cost: 200, ResearchTicks: 150,
			Prerequisites: []string{"mathematics", "primitive_writing"},
			Description: "Systematic inquiry into fundamental questions.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 0.3},
				{Type: "production", Target: "culture", Value: 0.2},
			},
		},
		{
			Name: "Civil Engineering", Key: "civil_engineering",
			Age: "classical_age", Cost: 180, ResearchTicks: 130,
			Prerequisites: []string{"masonry", "road_building"},
			Description: "Large-scale construction and infrastructure.",
			Effects: []Effect{
				{Type: "storage", Target: "all", Value: 100},
				{Type: "bonus", Target: "build_cost", Value: -0.05},
			},
		},
		{
			Name: "Imperial Legions", Key: "imperial_legions",
			Age: "classical_age", Cost: 220, ResearchTicks: 160,
			Prerequisites: []string{"siege_warfare", "iron_smelting"},
			Description: "Professional standing armies with superior discipline.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 0.4},
			},
		},

		// === MEDIEVAL AGE === (~7 min each)
		{
			Name: "Steel Forging", Key: "steel_forging",
			Age: "medieval_age", Cost: 250, ResearchTicks: 200,
			Prerequisites: []string{"iron_smelting"},
			Description: "Refining iron into steel for superior tools and weapons.",
			Effects: []Effect{
				{Type: "production", Target: "steel", Value: 0.1},
				{Type: "bonus", Target: "iron_rate", Value: 0.3},
			},
		},
		{
			Name: "Theology", Key: "theology",
			Age: "medieval_age", Cost: 200, ResearchTicks: 180,
			Prerequisites: []string{"philosophy"},
			Description: "Organized religion provides faith and social cohesion.",
			Effects: []Effect{
				{Type: "production", Target: "faith", Value: 0.3},
			},
		},
		{
			Name: "Banking", Key: "banking",
			Age: "medieval_age", Cost: 300, ResearchTicks: 210,
			Prerequisites: []string{"currency", "mathematics"},
			Description: "Financial institutions multiply gold generation.",
			Effects: []Effect{
				{Type: "bonus", Target: "gold_rate", Value: 0.5},
				{Type: "storage", Target: "gold", Value: 100},
			},
		},
		{
			Name: "Feudalism", Key: "feudalism",
			Age: "medieval_age", Cost: 220, ResearchTicks: 170,
			Prerequisites: []string{"military_tactics"},
			Description: "Feudal system improves population management.",
			Effects: []Effect{
				{Type: "capacity", Target: "population", Value: 5},
			},
		},
		{
			Name: "Alchemy", Key: "alchemy",
			Age: "medieval_age", Cost: 280, ResearchTicks: 220,
			Prerequisites: []string{"mathematics"},
			Description: "Proto-chemistry yields material insights.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 0.15},
				{Type: "production", Target: "gold", Value: 0.1},
			},
		},
		{
			Name: "Chronometry", Key: "chronometry",
			Age: "medieval_age", Cost: 200, ResearchTicks: 190,
			Description: "Precise timekeeping accelerates all activity.",
			Effects: []Effect{
				{Type: "bonus", Target: "tick_speed", Value: 0.05},
			},
		},

		// === RENAISSANCE AGE === (~10 min each)
		{
			Name: "Printing Press", Key: "printing_press",
			Age: "renaissance_age", Cost: 500, ResearchTicks: 300,
			Prerequisites: []string{"theology", "alchemy"},
			Description: "Mass production of books accelerates knowledge.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 0.4},
				{Type: "production", Target: "culture", Value: 0.3},
			},
		},
		{
			Name: "Navigation", Key: "navigation",
			Age: "renaissance_age", Cost: 450, ResearchTicks: 260,
			Prerequisites: []string{"mathematics", "road_building"},
			Description: "Ocean navigation opens new trade routes.",
			Effects: []Effect{
				{Type: "bonus", Target: "gold_rate", Value: 0.5},
				{Type: "bonus", Target: "expedition_reward", Value: 0.3},
			},
		},
		{
			Name: "Gunpowder", Key: "gunpowder",
			Age: "renaissance_age", Cost: 550, ResearchTicks: 320,
			Prerequisites: []string{"alchemy", "siege_warfare"},
			Description: "Explosive weaponry transforms warfare.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 0.5},
			},
		},
		{
			Name: "Patronage", Key: "patronage",
			Age: "renaissance_age", Cost: 400, ResearchTicks: 250,
			Prerequisites: []string{"banking"},
			Description: "Wealthy patrons fund arts and science.",
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 0.5},
				{Type: "production", Target: "knowledge", Value: 0.12},
			},
		},

		// === COLONIAL AGE === (~14 min each)
		{
			Name: "Cartography", Key: "cartography",
			Age: "colonial_age", Cost: 800, ResearchTicks: 400,
			Prerequisites: []string{"navigation"},
			Description: "Detailed maps enable global exploration.",
			Effects: []Effect{
				{Type: "bonus", Target: "expedition_reward", Value: 0.5},
				{Type: "bonus", Target: "gold_rate", Value: 0.5},
			},
		},
		{
			Name: "Mercantilism", Key: "mercantilism",
			Age: "colonial_age", Cost: 750, ResearchTicks: 380,
			Prerequisites: []string{"banking", "navigation"},
			Description: "National trade policies maximize wealth.",
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 2.0},
				{Type: "bonus", Target: "gold_rate", Value: 0.3},
			},
		},
		{
			Name: "Colonialism", Key: "colonialism",
			Age: "colonial_age", Cost: 900, ResearchTicks: 440,
			Prerequisites: []string{"cartography", "gunpowder"},
			Description: "Overseas territorial expansion.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 2.0},
				{Type: "bonus", Target: "military_power", Value: 0.3},
			},
		},

		// === INDUSTRIAL AGE === (~18 min each)
		{
			Name: "Steam Power", Key: "steam_power",
			Age: "industrial_age", Cost: 1000, ResearchTicks: 520,
			Prerequisites: []string{"steel_forging"},
			Description: "Steam engines revolutionize production.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.3},
			},
		},
		{
			Name: "Industrialization", Key: "industrialization",
			Age: "industrial_age", Cost: 1200, ResearchTicks: 600,
			Prerequisites: []string{"steam_power"},
			Description: "Factory systems massively increase output.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.5},
				{Type: "production", Target: "steel", Value: 0.5},
			},
		},
		{
			Name: "Railroads", Key: "railroads",
			Age: "industrial_age", Cost: 900, ResearchTicks: 500,
			Prerequisites: []string{"steam_power", "road_building"},
			Description: "Rail networks connect your civilization.",
			Effects: []Effect{
				{Type: "bonus", Target: "gold_rate", Value: 1.0},
				{Type: "storage", Target: "all", Value: 200},
			},
		},
		{
			Name: "Rifling", Key: "rifling",
			Age: "industrial_age", Cost: 800, ResearchTicks: 480,
			Prerequisites: []string{"gunpowder"},
			Description: "Precision firearms improve military effectiveness.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 0.5},
			},
		},
		{
			Name: "Clockwork Automation", Key: "clockwork_automation",
			Age: "industrial_age", Cost: 500, ResearchTicks: 540,
			Prerequisites: []string{"chronometry"},
			Description: "Mechanical automation speeds up every process.",
			Effects: []Effect{
				{Type: "bonus", Target: "tick_speed", Value: 0.10},
			},
		},

		// === VICTORIAN AGE === (~23 min each)
		{
			Name: "Electrification", Key: "electrification",
			Age: "victorian_age", Cost: 1800, ResearchTicks: 700,
			Prerequisites: []string{"industrialization"},
			Description: "Electrical power begins to transform society.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 1.0},
				{Type: "bonus", Target: "production_all", Value: 0.2},
			},
		},
		{
			Name: "Telecommunications", Key: "telecommunications",
			Age: "victorian_age", Cost: 1500, ResearchTicks: 660,
			Prerequisites: []string{"electrification"},
			Description: "Telegraph and early telephone networks.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 0.4},
				{Type: "bonus", Target: "gold_rate", Value: 0.5},
			},
		},
		{
			Name: "Mass Production", Key: "mass_production",
			Age: "victorian_age", Cost: 2000, ResearchTicks: 740,
			Prerequisites: []string{"industrialization", "railroads"},
			Description: "Assembly line manufacturing.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.4},
				{Type: "production", Target: "steel", Value: 1.0},
			},
		},

		// === ELECTRIC AGE === (~32 min each)
		{
			Name: "Power Distribution", Key: "power_distribution",
			Age: "electric_age", Cost: 3000, ResearchTicks: 950,
			Prerequisites: []string{"electrification"},
			Description: "AC power grids span entire regions.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 3.0},
				{Type: "bonus", Target: "production_all", Value: 0.3},
			},
		},
		{
			Name: "Radio", Key: "radio",
			Age: "electric_age", Cost: 2500, ResearchTicks: 900,
			Prerequisites: []string{"telecommunications"},
			Description: "Wireless communication reaches the masses.",
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 2.0},
				{Type: "bonus", Target: "knowledge_rate", Value: 0.4},
			},
		},
		{
			Name: "Chemical Engineering", Key: "chemical_engineering",
			Age: "electric_age", Cost: 2800, ResearchTicks: 980,
			Prerequisites: []string{"mass_production"},
			Description: "Industrial chemistry and synthetic materials.",
			Effects: []Effect{
				{Type: "production", Target: "oil", Value: 1.0},
				{Type: "bonus", Target: "production_all", Value: 0.2},
			},
		},

		// === ATOMIC AGE === (~45 min each)
		{
			Name: "Nuclear Fission", Key: "nuclear_fission",
			Age: "atomic_age", Cost: 5000, ResearchTicks: 1350,
			Prerequisites: []string{"power_distribution", "chemical_engineering"},
			Description: "Splitting the atom for energy and weapons.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 5.0},
				{Type: "production", Target: "uranium", Value: 0.5},
			},
		},
		{
			Name: "Rocketry", Key: "rocketry",
			Age: "atomic_age", Cost: 4000, ResearchTicks: 1200,
			Prerequisites: []string{"rifling", "chemical_engineering"},
			Description: "Rocket technology enables space exploration.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 1.0},
				{Type: "bonus", Target: "expedition_reward", Value: 0.5},
			},
		},
		{
			Name: "Nuclear Deterrence", Key: "nuclear_deterrence",
			Age: "atomic_age", Cost: 6000, ResearchTicks: 1500,
			Prerequisites: []string{"nuclear_fission", "rocketry"},
			Description: "Mutually assured destruction maintains peace.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 1.5},
			},
		},

		// === MODERN AGE === (~1.1 hr each)
		{
			Name: "Electricity", Key: "electricity_tech",
			Age: "modern_age", Cost: 8000, ResearchTicks: 1800,
			Prerequisites: []string{"nuclear_fission"},
			Description: "Advanced electrical systems transform civilization.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.5},
				{Type: "production", Target: "electricity", Value: 5.0},
			},
		},
		{
			Name: "Computers", Key: "computers",
			Age: "modern_age", Cost: 10000, ResearchTicks: 2000,
			Prerequisites: []string{"electricity_tech"},
			Description: "Digital computing accelerates all research.",
			Effects: []Effect{
				{Type: "bonus", Target: "knowledge_rate", Value: 0.8},
			},
		},
		{
			Name: "Satellite Technology", Key: "satellite_tech",
			Age: "modern_age", Cost: 12000, ResearchTicks: 1900,
			Prerequisites: []string{"rocketry", "electricity_tech"},
			Description: "Orbital satellites for communication and surveillance.",
			Effects: []Effect{
				{Type: "production", Target: "data", Value: 1.0},
				{Type: "bonus", Target: "knowledge_rate", Value: 0.6},
			},
		},

		// === INFORMATION AGE === (~1.5 hr each)
		{
			Name: "Internet", Key: "internet",
			Age: "information_age", Cost: 20000, ResearchTicks: 2600,
			Prerequisites: []string{"computers", "satellite_tech"},
			Description: "Global network connecting all of humanity.",
			Effects: []Effect{
				{Type: "production", Target: "data", Value: 3.0},
				{Type: "bonus", Target: "knowledge_rate", Value: 1.2},
			},
		},
		{
			Name: "Cybersecurity", Key: "cybersecurity",
			Age: "information_age", Cost: 18000, ResearchTicks: 2400,
			Prerequisites: []string{"computers"},
			Description: "Defense against digital threats.",
			Effects: []Effect{
				{Type: "bonus", Target: "military_power", Value: 1.0},
				{Type: "storage", Target: "data", Value: 5000},
			},
		},
		{
			Name: "Social Media", Key: "social_media",
			Age: "information_age", Cost: 15000, ResearchTicks: 2300,
			Prerequisites: []string{"internet"},
			Description: "Mass digital communication platforms.",
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 5.0},
				{Type: "production", Target: "gold", Value: 5.0},
			},
		},

		// === DIGITAL AGE === (~1.8 hr each)
		{
			Name: "Machine Learning", Key: "machine_learning",
			Age: "digital_age", Cost: 35000, ResearchTicks: 3400,
			Prerequisites: []string{"internet", "cybersecurity"},
			Description: "Algorithms that learn and improve autonomously.",
			Effects: []Effect{
				{Type: "production", Target: "data", Value: 5.0},
				{Type: "bonus", Target: "production_all", Value: 0.5},
			},
		},
		{
			Name: "Cloud Computing", Key: "cloud_computing",
			Age: "digital_age", Cost: 30000, ResearchTicks: 3200,
			Prerequisites: []string{"internet"},
			Description: "Distributed computing at global scale.",
			Effects: []Effect{
				{Type: "production", Target: "data", Value: 8.0},
				{Type: "storage", Target: "all", Value: 10000},
			},
		},

		// === CYBERPUNK AGE === (~2.8 hr each)
		{
			Name: "Neural Interface", Key: "neural_interface",
			Age: "cyberpunk_age", Cost: 60000, ResearchTicks: 4800,
			Prerequisites: []string{"machine_learning"},
			Description: "Direct brain-computer interface technology.",
			Effects: []Effect{
				{Type: "bonus", Target: "gather_rate", Value: 0.3},
				{Type: "bonus", Target: "knowledge_rate", Value: 2.0},
			},
		},
		{
			Name: "Blockchain", Key: "blockchain",
			Age: "cyberpunk_age", Cost: 50000, ResearchTicks: 4500,
			Prerequisites: []string{"cybersecurity", "cloud_computing"},
			Description: "Decentralized trustless systems.",
			Effects: []Effect{
				{Type: "production", Target: "crypto", Value: 2.0},
				{Type: "bonus", Target: "gold_rate", Value: 2.0},
			},
		},
		{
			Name: "Cybernetics", Key: "cybernetics",
			Age: "cyberpunk_age", Cost: 55000, ResearchTicks: 5000,
			Prerequisites: []string{"neural_interface"},
			Description: "Mechanical augmentation of the human body.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.5},
				{Type: "bonus", Target: "military_power", Value: 1.0},
			},
		},

		// === FUSION AGE === (~3.7 hr each)
		{
			Name: "Fusion Power", Key: "fusion_power",
			Age: "fusion_age", Cost: 100000, ResearchTicks: 6500,
			Prerequisites: []string{"nuclear_fission", "cybernetics"},
			Description: "Controlled nuclear fusion for limitless energy.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 20.0},
				{Type: "production", Target: "plasma", Value: 1.0},
			},
		},
		{
			Name: "Plasma Physics", Key: "plasma_physics",
			Age: "fusion_age", Cost: 90000, ResearchTicks: 6200,
			Prerequisites: []string{"fusion_power"},
			Description: "Mastery of superheated matter states.",
			Effects: []Effect{
				{Type: "production", Target: "plasma", Value: 3.0},
				{Type: "bonus", Target: "production_all", Value: 0.3},
			},
		},
		{
			Name: "Superconductors", Key: "superconductors",
			Age: "fusion_age", Cost: 110000, ResearchTicks: 7000,
			Prerequisites: []string{"fusion_power"},
			Description: "Zero-resistance materials revolutionize technology.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.5},
				{Type: "storage", Target: "all", Value: 50000},
			},
		},

		// === SPACE AGE === (~5 hr each)
		{
			Name: "Orbital Mechanics", Key: "orbital_mechanics",
			Age: "space_age", Cost: 200000, ResearchTicks: 8500,
			Prerequisites: []string{"rocketry", "plasma_physics"},
			Description: "Advanced spaceflight and orbital dynamics.",
			Effects: []Effect{
				{Type: "production", Target: "titanium", Value: 1.0},
				{Type: "bonus", Target: "expedition_reward", Value: 1.0},
			},
		},
		{
			Name: "Space Mining", Key: "space_mining",
			Age: "space_age", Cost: 180000, ResearchTicks: 8200,
			Prerequisites: []string{"orbital_mechanics"},
			Description: "Asteroid and lunar resource extraction.",
			Effects: []Effect{
				{Type: "production", Target: "titanium", Value: 3.0},
				{Type: "production", Target: "iron", Value: 20.0},
			},
		},
		{
			Name: "Zero-G Manufacturing", Key: "zero_g_manufacturing",
			Age: "space_age", Cost: 220000, ResearchTicks: 9200,
			Prerequisites: []string{"orbital_mechanics", "superconductors"},
			Description: "Space-based manufacturing for perfect materials.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.5},
				{Type: "production", Target: "steel", Value: 10.0},
			},
		},

		// === INTERSTELLAR AGE === (~7 hr each)
		{
			Name: "Warp Drive", Key: "warp_drive",
			Age: "interstellar_age", Cost: 400000, ResearchTicks: 12000,
			Prerequisites: []string{"space_mining", "zero_g_manufacturing"},
			Description: "Faster-than-light propulsion.",
			Effects: []Effect{
				{Type: "production", Target: "dark_matter", Value: 1.0},
				{Type: "bonus", Target: "expedition_reward", Value: 2.0},
			},
		},
		{
			Name: "Stellar Engineering", Key: "stellar_engineering",
			Age: "interstellar_age", Cost: 450000, ResearchTicks: 13000,
			Prerequisites: []string{"warp_drive"},
			Description: "Harnessing and shaping stars themselves.",
			Effects: []Effect{
				{Type: "production", Target: "plasma", Value: 10.0},
				{Type: "production", Target: "electricity", Value: 100.0},
			},
		},

		// === GALACTIC AGE === (~9 hr each)
		{
			Name: "Galactic Navigation", Key: "galactic_navigation",
			Age: "galactic_age", Cost: 800000, ResearchTicks: 16000,
			Prerequisites: []string{"warp_drive", "stellar_engineering"},
			Description: "Charting paths across the galaxy.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 0.5},
				{Type: "production", Target: "dark_matter", Value: 5.0},
			},
		},
		{
			Name: "Antimatter Synthesis", Key: "antimatter_synthesis",
			Age: "galactic_age", Cost: 900000, ResearchTicks: 18000,
			Prerequisites: []string{"galactic_navigation"},
			Description: "Controlled production of antimatter.",
			Effects: []Effect{
				{Type: "production", Target: "antimatter", Value: 2.0},
				{Type: "bonus", Target: "production_all", Value: 0.3},
			},
		},

		// === QUANTUM AGE === (~12 hr each)
		{
			Name: "Quantum Mechanics", Key: "quantum_mechanics",
			Age: "quantum_age", Cost: 1500000, ResearchTicks: 22000,
			Prerequisites: []string{"antimatter_synthesis"},
			Description: "Mastery of quantum phenomena at all scales.",
			Effects: []Effect{
				{Type: "production", Target: "quantum_flux", Value: 2.0},
				{Type: "bonus", Target: "production_all", Value: 1.0},
			},
		},
		{
			Name: "Reality Manipulation", Key: "reality_manipulation",
			Age: "quantum_age", Cost: 2000000, ResearchTicks: 25000,
			Prerequisites: []string{"quantum_mechanics"},
			Description: "Bending the fabric of spacetime.",
			Effects: []Effect{
				{Type: "production", Target: "quantum_flux", Value: 5.0},
				{Type: "bonus", Target: "production_all", Value: 1.0},
			},
		},
		{
			Name: "Quantum Computing", Key: "quantum_computing",
			Age: "quantum_age", Cost: 1500000, ResearchTicks: 20000,
			Prerequisites: []string{"clockwork_automation"},
			Description: "Quantum processing collapses wait times.",
			Effects: []Effect{
				{Type: "bonus", Target: "tick_speed", Value: 0.15},
			},
		},

		// === TRANSCENDENT AGE === (~18 hr)
		{
			Name: "Transcendence", Key: "transcendence",
			Age: "transcendent_age", Cost: 5000000, ResearchTicks: 32000,
			Prerequisites: []string{"reality_manipulation"},
			Description: "Ascension beyond physical limitations.",
			Effects: []Effect{
				{Type: "bonus", Target: "production_all", Value: 2.0},
				{Type: "production", Target: "quantum_flux", Value: 10.0},
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
