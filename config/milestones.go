package config

// MilestoneDef defines an achievement/milestone with a permanent reward
type MilestoneDef struct {
	Name        string
	Key         string
	Description string
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

// Milestones returns all milestone definitions
func Milestones() []MilestoneDef {
	return []MilestoneDef{
		// === EARLY GAME ===
		{
			Name: "First Shelter", Key: "first_shelter",
			Description: "Build your first hut.",
			MinBuildings: map[string]int{"hut": 1},
			Rewards: []Effect{
				{Type: "instant_resource", Target: "food", Value: 10},
			},
		},
		{
			Name: "Small Village", Key: "small_village",
			Description: "Reach a population of 5.",
			MinPopulation: 5,
			Rewards: []Effect{
				{Type: "instant_resource", Target: "wood", Value: 20},
			},
		},
		{
			Name: "Knowledge Seeker", Key: "knowledge_seeker",
			Description: "Accumulate 50 knowledge.",
			MinResources: map[string]float64{"knowledge": 50},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.1},
			},
		},
		{
			Name: "Stone Mason", Key: "stone_mason",
			Description: "Build 3 stone pits.",
			MinBuildings: map[string]int{"stone_pit": 3},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "stone_rate", Value: 0.1},
			},
		},

		// === MID GAME ===
		{
			Name: "Bronze Age Pioneer", Key: "bronze_pioneer",
			Description: "Advance to the Bronze Age.",
			MinAge: "bronze_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "iron", Value: 20},
				{Type: "instant_resource", Target: "gold", Value: 20},
			},
		},
		{
			Name: "Scholar's Haven", Key: "scholars_haven",
			Description: "Have 5 scholars working.",
			MinPopulation: 10,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.2},
			},
		},
		{
			Name: "Bustling Town", Key: "bustling_town",
			Description: "Reach a population of 20.",
			MinPopulation: 20,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.1},
			},
		},
		{
			Name: "Master Builder", Key: "master_builder",
			Description: "Build 20 structures total.",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.05},
			},
		},
		{
			Name: "First Research", Key: "first_research",
			Description: "Complete your first technology.",
			MinTechCount: 1,
			Rewards: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 15},
			},
		},
		{
			Name: "Tech Pioneer", Key: "tech_pioneer",
			Description: "Research 5 technologies.",
			MinTechCount: 5,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.1},
			},
		},

		// === LATE GAME ===
		{
			Name: "Iron Forged", Key: "iron_forged",
			Description: "Advance to the Iron Age.",
			MinAge: "iron_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "coal", Value: 30},
				{Type: "permanent_bonus", Target: "iron_rate", Value: 0.15},
			},
		},
		{
			Name: "Growing City", Key: "growing_city",
			Description: "Reach a population of 50.",
			MinPopulation: 50,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "food_rate", Value: 0.2},
			},
		},
		{
			Name: "War Machine", Key: "war_machine",
			Description: "Have 10 soldiers.",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.2},
			},
		},
		{
			Name: "Medieval Lord", Key: "medieval_lord",
			Description: "Advance to the Medieval Age.",
			MinAge: "medieval_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "faith", Value: 30},
				{Type: "instant_resource", Target: "steel", Value: 15},
			},
		},
		{
			Name: "Wonder Builder", Key: "wonder_builder",
			Description: "Build a Wonder.",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.1},
			},
		},
		{
			Name: "Renaissance Mind", Key: "renaissance_mind",
			Description: "Research 15 technologies.",
			MinTechCount: 15,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.2},
			},
		},
		{
			Name: "Metropolis", Key: "metropolis",
			Description: "Reach a population of 100.",
			MinPopulation: 100,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},
		{
			Name: "Enlightened", Key: "enlightened",
			Description: "Advance to the Renaissance Age.",
			MinAge: "renaissance_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "culture", Value: 50},
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.3},
			},
		},
		{
			Name: "Industrial Revolution", Key: "industrial_revolution",
			Description: "Advance to the Industrial Age.",
			MinAge: "industrial_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.25},
			},
		},
		{
			Name: "Modern Era", Key: "modern_era",
			Description: "Advance to the Modern Age.",
			MinAge: "modern_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.5},
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
