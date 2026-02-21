package config

// MilestoneDef defines an achievement/milestone with a permanent reward
type MilestoneDef struct {
	Name        string
	Key         string
	Description string
	Category    string // "settlement", "scholar", "builder", "military", "ages"
	Hidden      bool   // hidden milestones only revealed when close to completion
	// Conditions (any that are set must be met)
	MinTick        int               // minimum game tick
	MinAge         string            // must be in this age or later
	MinResources   map[string]float64 // resource amounts required (checked live)
	MinBuildings   map[string]int     // building counts required
	MinPopulation  int               // total population required
	MinTechCount   int               // number of techs researched
	RequiredTechs  []string          // specific techs that must be researched
	// Rewards
	Rewards []Effect
}

// MilestoneChainDef defines a chain of milestones that grants a bonus when all are completed
type MilestoneChainDef struct {
	Name          string
	Key           string
	Category      string
	MilestoneKeys []string
	Title         string  // civilization title unlocked on completion
	BoostValue    float64 // tick_speed bonus
	BoostDuration int     // duration in ticks
}

// TitleDef defines a civilization title earned by reaching a milestone count
type TitleDef struct {
	Title         string
	MinMilestones int
}

// MilestoneChains returns all milestone chain definitions
func MilestoneChains() []MilestoneChainDef {
	return []MilestoneChainDef{
		{
			Name:     "Settlement Chain",
			Key:      "settlement_chain",
			Category: "settlement",
			MilestoneKeys: []string{
				"first_shelter", "small_village", "bustling_town",
				"growing_city", "metropolis", "megalopolis",
			},
			Title:         "The Founders",
			BoostValue:    3.0,
			BoostDuration: 180,
		},
		{
			Name:     "Scholar Chain",
			Key:      "scholar_chain",
			Category: "scholar",
			MilestoneKeys: []string{
				"knowledge_seeker", "first_research", "tech_pioneer",
				"scholars_haven", "renaissance_mind", "tech_master",
			},
			Title:         "The Enlightened",
			BoostValue:    3.0,
			BoostDuration: 180,
		},
		{
			Name:     "Builder Chain",
			Key:      "builder_chain",
			Category: "builder",
			MilestoneKeys: []string{
				"stone_mason", "master_builder", "wonder_builder",
			},
			Title:         "The Architects",
			BoostValue:    1.5,
			BoostDuration: 90,
		},
		{
			Name:     "Military Chain",
			Key:      "military_chain",
			Category: "military",
			MilestoneKeys: []string{
				"war_machine",
			},
			Title:         "The Conquerors",
			BoostValue:    0.5,
			BoostDuration: 30,
		},
		{
			Name:     "Ancient Ages Chain",
			Key:      "ancient_ages_chain",
			Category: "ages",
			MilestoneKeys: []string{
				"bronze_pioneer", "iron_forged", "classical_scholar",
				"medieval_lord", "enlightened",
			},
			Title:         "The Ancients",
			BoostValue:    2.5,
			BoostDuration: 150,
		},
	}
}

// MilestoneChainByKey returns a map of key -> MilestoneChainDef
func MilestoneChainByKey() map[string]MilestoneChainDef {
	m := make(map[string]MilestoneChainDef)
	for _, c := range MilestoneChains() {
		m[c.Key] = c
	}
	return m
}

// MilestoneTitles returns fallback titles based on milestone count (sorted ascending)
func MilestoneTitles() []TitleDef {
	return []TitleDef{
		{Title: "Aspiring", MinMilestones: 3},
		{Title: "Rising Power", MinMilestones: 8},
		{Title: "Established", MinMilestones: 15},
		{Title: "Dominant Force", MinMilestones: 22},
		{Title: "Legend", MinMilestones: 30},
	}
}

// Milestones returns all milestone definitions
func Milestones() []MilestoneDef {
	return []MilestoneDef{
		// === SETTLEMENT ===
		{
			Name: "First Shelter", Key: "first_shelter",
			Description: "Build your first hut.",
			Category: "settlement",
			MinBuildings: map[string]int{"hut": 1},
			Rewards: []Effect{
				{Type: "instant_resource", Target: "food", Value: 10},
			},
		},
		{
			Name: "Small Village", Key: "small_village",
			Description: "Reach a population of 5.",
			Category: "settlement",
			MinPopulation: 5,
			Rewards: []Effect{
				{Type: "instant_resource", Target: "wood", Value: 20},
			},
		},
		{
			Name: "Bustling Town", Key: "bustling_town",
			Description: "Reach a population of 20.",
			Category: "settlement",
			MinPopulation: 20,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.1},
			},
		},
		{
			Name: "Growing City", Key: "growing_city",
			Description: "Reach a population of 50.",
			Category: "settlement",
			MinPopulation: 50,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "food_rate", Value: 0.2},
			},
		},
		{
			Name: "Metropolis", Key: "metropolis",
			Description: "Reach a population of 100.",
			Category: "settlement", Hidden: true,
			MinPopulation: 100,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},
		{
			Name: "Megalopolis", Key: "megalopolis",
			Description: "Reach a population of 500.",
			Category: "settlement", Hidden: true,
			MinPopulation: 500,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.5},
			},
		},

		// === SCHOLAR ===
		{
			Name: "Knowledge Seeker", Key: "knowledge_seeker",
			Description: "Accumulate 50 knowledge.",
			Category: "scholar",
			MinResources: map[string]float64{"knowledge": 50},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.1},
			},
		},
		{
			Name: "First Research", Key: "first_research",
			Description: "Complete your first technology.",
			Category: "scholar",
			MinTechCount: 1,
			Rewards: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 15},
			},
		},
		{
			Name: "Tech Pioneer", Key: "tech_pioneer",
			Description: "Research 5 technologies.",
			Category: "scholar",
			MinTechCount: 5,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.1},
			},
		},
		{
			Name: "Scholar's Haven", Key: "scholars_haven",
			Description: "Have 5 scholars working.",
			Category: "scholar",
			MinPopulation: 10,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.2},
			},
		},
		{
			Name: "Renaissance Mind", Key: "renaissance_mind",
			Description: "Research 15 technologies.",
			Category: "scholar", Hidden: true,
			MinTechCount: 15,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.2},
			},
		},
		{
			Name: "Tech Master", Key: "tech_master",
			Description: "Research 30 technologies.",
			Category: "scholar", Hidden: true,
			MinTechCount: 30,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.5},
			},
		},

		// === BUILDER ===
		{
			Name: "Stone Mason", Key: "stone_mason",
			Description: "Build 3 stone pits.",
			Category: "builder",
			MinBuildings: map[string]int{"stone_pit": 3},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "stone_rate", Value: 0.1},
			},
		},
		{
			Name: "Master Builder", Key: "master_builder",
			Description: "Build 20 structures total.",
			Category: "builder",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.05},
			},
		},
		{
			Name: "Wonder Builder", Key: "wonder_builder",
			Description: "Build a Wonder.",
			Category: "builder", Hidden: true,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.1},
			},
		},

		// === MILITARY ===
		{
			Name: "War Machine", Key: "war_machine",
			Description: "Have 10 soldiers.",
			Category: "military",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.2},
			},
		},

		// === AGES ===
		{
			Name: "Bronze Age Pioneer", Key: "bronze_pioneer",
			Description: "Advance to the Bronze Age.",
			Category: "ages",
			MinAge: "bronze_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "iron", Value: 20},
				{Type: "instant_resource", Target: "gold", Value: 20},
			},
		},
		{
			Name: "Iron Forged", Key: "iron_forged",
			Description: "Advance to the Iron Age.",
			Category: "ages",
			MinAge: "iron_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "coal", Value: 30},
				{Type: "permanent_bonus", Target: "iron_rate", Value: 0.15},
			},
		},
		{
			Name: "Classical Scholar", Key: "classical_scholar",
			Description: "Advance to the Classical Age.",
			Category: "ages",
			MinAge: "classical_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 100},
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.2},
			},
		},
		{
			Name: "Medieval Lord", Key: "medieval_lord",
			Description: "Advance to the Medieval Age.",
			Category: "ages",
			MinAge: "medieval_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "faith", Value: 30},
				{Type: "instant_resource", Target: "steel", Value: 15},
			},
		},
		{
			Name: "Enlightened", Key: "enlightened",
			Description: "Advance to the Renaissance Age.",
			Category: "ages",
			MinAge: "renaissance_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "culture", Value: 50},
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.3},
			},
		},
		{
			Name: "Colonial Power", Key: "colonial_power",
			Description: "Advance to the Colonial Age.",
			Category: "ages", Hidden: true,
			MinAge: "colonial_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.3},
				{Type: "permanent_bonus", Target: "expedition_reward", Value: 0.2},
			},
		},
		{
			Name: "Industrial Revolution", Key: "industrial_revolution",
			Description: "Advance to the Industrial Age.",
			Category: "ages", Hidden: true,
			MinAge: "industrial_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.25},
			},
		},
		{
			Name: "Victorian Innovation", Key: "victorian_innovation",
			Description: "Advance to the Victorian Age.",
			Category: "ages", Hidden: true,
			MinAge: "victorian_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.2},
			},
		},
		{
			Name: "Electric Dawn", Key: "electric_dawn",
			Description: "Advance to the Electric Age.",
			Category: "ages", Hidden: true,
			MinAge: "electric_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.2},
			},
		},
		{
			Name: "Atomic Pioneer", Key: "atomic_pioneer",
			Description: "Advance to the Atomic Age.",
			Category: "ages", Hidden: true,
			MinAge: "atomic_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.25},
			},
		},
		{
			Name: "Modern Era", Key: "modern_era",
			Description: "Advance to the Modern Age.",
			Category: "ages", Hidden: true,
			MinAge: "modern_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.5},
			},
		},
		{
			Name: "Information Pioneer", Key: "information_pioneer",
			Description: "Advance to the Information Age.",
			Category: "ages", Hidden: true,
			MinAge: "information_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 1.0},
			},
		},
		{
			Name: "Digital Native", Key: "digital_native",
			Description: "Advance to the Digital Age.",
			Category: "ages", Hidden: true,
			MinAge: "digital_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.3},
			},
		},
		{
			Name: "Cyberpunk", Key: "cyberpunk_milestone",
			Description: "Advance to the Cyberpunk Age.",
			Category: "ages", Hidden: true,
			MinAge: "cyberpunk_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gather_rate", Value: 0.2},
			},
		},
		{
			Name: "Fusion Pioneer", Key: "fusion_pioneer",
			Description: "Advance to the Fusion Age.",
			Category: "ages", Hidden: true,
			MinAge: "fusion_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.5},
			},
		},
		{
			Name: "Space Explorer", Key: "space_explorer",
			Description: "Advance to the Space Age.",
			Category: "ages", Hidden: true,
			MinAge: "space_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.5},
				{Type: "permanent_bonus", Target: "expedition_reward", Value: 0.5},
			},
		},
		{
			Name: "Star Voyager", Key: "star_voyager",
			Description: "Advance to the Interstellar Age.",
			Category: "ages", Hidden: true,
			MinAge: "interstellar_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.5},
			},
		},
		{
			Name: "Galactic Emperor", Key: "galactic_emperor",
			Description: "Advance to the Galactic Age.",
			Category: "ages", Hidden: true,
			MinAge: "galactic_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 1.0},
			},
		},
		{
			Name: "Quantum Master", Key: "quantum_master",
			Description: "Advance to the Quantum Age.",
			Category: "ages", Hidden: true,
			MinAge: "quantum_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 1.0},
			},
		},
		{
			Name: "Transcended", Key: "transcended",
			Description: "Advance to the Transcendent Age.",
			Category: "ages", Hidden: true,
			MinAge: "transcendent_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 2.0},
			},
		},
	}
}

// MilestoneByKey returns a map of key -> MilestoneDef
func MilestoneByKey() map[string]MilestoneDef {
	m := make(map[string]MilestoneDef)
	for _, ms := range Milestones() {
		m[ms.Key] = ms
	}
	return m
}

// MilestoneCategoryOrder returns the display order of milestone categories
func MilestoneCategoryOrder() []string {
	return []string{"settlement", "builder", "scholar", "military", "ages"}
}

// MilestoneCategoryNames returns display names for categories
func MilestoneCategoryNames() map[string]string {
	return map[string]string{
		"settlement": "Settlement",
		"builder":    "Builder",
		"scholar":    "Scholar",
		"military":   "Military",
		"ages":       "Ages",
	}
}
