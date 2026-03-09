package config

// MilestoneDef defines an achievement with a one-time permanent reward.
// All non-zero conditions must be satisfied simultaneously for the milestone
// to complete. Conditions are checked every tick by the engine.
//
// Hidden milestones are suppressed from the UI until either:
//   - They are completed
//   - Progress > 50% of any numeric condition
//   - (For MinAge milestones) the player is in the preceding age
type MilestoneDef struct {
	Name        string
	Key         string
	Description string
	Category    string // "settlement", "scholar", "builder", "military", "trade", "faith", "epoch", "ages"
	Hidden      bool   // hidden milestones only revealed when close to completion or done
	// Conditions — all non-zero fields must be satisfied at the same time
	MinTick       int                // game tick must have reached this value
	MinAge        string             // player must be in this age or any later age
	MinResources  map[string]float64 // each resource must currently hold at least this amount
	MinBuildings  map[string]int     // each building type must have at least this many built
	MinPopulation int                // state.Workers.TotalPop must be ≥ this
	MinTechCount  int                // total number of completed research techs must be ≥ this
	RequiredTechs []string           // all listed tech keys must be researched
	// Rewards — applied once when the milestone completes
	Rewards []Effect
}

// MilestoneChainDef defines a set of milestones that, when all completed,
// grants a civilization Title and a temporary tick-speed boost.
// The boost is applied via engine.InjectEvent with type "tick_speed".
type MilestoneChainDef struct {
	Name          string
	Key           string
	Category      string
	MilestoneKeys []string // all of these must be completed to finish the chain
	Title         string   // civilization title shown in the status bar
	BoostValue    float64  // tick_speed multiplier bonus (e.g. 3.0 = +3x speed for BoostDuration ticks)
	BoostDuration int      // how many ticks the speed boost lasts
}

// TitleDef is a fallback title awarded based purely on total milestones completed.
// Chain completions take precedence over these generic titles in the UI.
type TitleDef struct {
	Title         string
	MinMilestones int // minimum completed milestone count to earn this title
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
				"grand_architect", "wonder_collector",
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
				"first_soldiers", "war_machine", "iron_legion",
				"fortress_state", "military_superpower",
			},
			Title:         "The Conquerors",
			BoostValue:    0.5,
			BoostDuration: 60,
		},
		{
			Name:     "Trade Chain",
			Key:      "trade_chain",
			Category: "trade",
			MilestoneKeys: []string{
				"first_market", "merchant_guild", "trade_empire",
			},
			Title:         "The Merchants",
			BoostValue:    1.0,
			BoostDuration: 90,
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

// MilestoneTitles returns the fallback title ladder sorted by MinMilestones ascending.
// The engine applies the highest-matching title unless a chain title is active.
func MilestoneTitles() []TitleDef {
	return []TitleDef{
		{Title: "Aspiring", MinMilestones: 5},
		{Title: "Rising Power", MinMilestones: 12},
		{Title: "Established", MinMilestones: 24},
		{Title: "Dominant Force", MinMilestones: 40},
		{Title: "Legend", MinMilestones: 56},
		{Title: "Transcendent", MinMilestones: 62},
	}
}

// Milestones returns all milestone definitions.
// Total: 74 milestones across settlement, builder, scholar, military, trade, faith, epoch, and ages categories.
func Milestones() []MilestoneDef {
	return []MilestoneDef{

		// =================================================================
		// === SETTLEMENT (population / housing progression) ===
		// =================================================================

		// first_shelter: build 1 hut — tutorial trigger, unchanged
		{
			Name: "First Shelter", Key: "first_shelter",
			Description:  "Build your first hut.",
			Category:     "settlement",
			MinBuildings: map[string]int{"hut": 1},
			Rewards: []Effect{
				{Type: "instant_resource", Target: "food", Value: 10},
			},
		},
		// small_village: raised to 1,000 — requires real hut investment + recruiting
		{
			Name: "Small Village", Key: "small_village",
			Description:   "Reach a population of 1,000.",
			Category:      "settlement",
			MinPopulation: 1000,
			Rewards: []Effect{
				{Type: "instant_resource", Target: "wood", Value: 50},
			},
		},
		// bustling_town: raised to 25,000
		{
			Name: "Bustling Town", Key: "bustling_town",
			Description:   "Reach a population of 25,000.",
			Category:      "settlement",
			MinPopulation: 25000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "food_rate", Value: 0.05},
			},
		},
		// growing_city: raised to 250,000; bronze_age gate maintained
		{
			Name: "Growing City", Key: "growing_city",
			Description:   "Reach a population of 250,000.",
			Category:      "settlement",
			MinAge:        "bronze_age",
			MinPopulation: 250000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "food_rate", Value: 0.10},
			},
		},
		// metropolis: raised to 5M pop; iron_age gated
		{
			Name: "Metropolis", Key: "metropolis",
			Description: "Reach a population of 5,000,000.",
			Category:    "settlement", Hidden: true,
			MinAge:        "iron_age",
			MinPopulation: 5000000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},
		// megalopolis: raised to 500M pop
		{
			Name: "Megalopolis", Key: "megalopolis",
			Description: "Reach a population of 500,000,000.",
			Category:    "settlement", Hidden: true,
			MinAge:        "classical_age",
			MinPopulation: 500000000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		// NEW: urban_sprawl — 500M pop, medieval_age gated
		{
			Name: "Urban Sprawl", Key: "urban_sprawl",
			Description: "Reach a population of 500,000,000.",
			Category:    "settlement", Hidden: true,
			MinAge:        "medieval_age",
			MinPopulation: 500000000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},
		// NEW: global_city — 5B pop, industrial_age gated
		{
			Name: "Global City", Key: "global_city",
			Description: "Reach a population of 5,000,000,000.",
			Category:    "settlement", Hidden: true,
			MinAge:        "industrial_age",
			MinPopulation: 5000000000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},

		// =================================================================
		// === BUILDER (construction / storage / wonders) ===
		// =================================================================

		// NEW: first_storehouse — build 1 stash; earliest storage milestone
		{
			Name: "First Storehouse", Key: "first_storehouse",
			Description:  "Build your first Stash.",
			Category:     "builder",
			MinBuildings: map[string]int{"stash": 1},
			Rewards: []Effect{
				{Type: "instant_resource", Target: "food", Value: 20},
				{Type: "instant_resource", Target: "wood", Value: 20},
			},
		},
		// NEW: storage_network — 3 storage pits; stone/bronze age
		{
			Name: "Storage Network", Key: "storage_network",
			Description:  "Build 3 Storage Pits.",
			Category:     "builder",
			MinAge:       "stone_age",
			MinBuildings: map[string]int{"storage_pit": 3},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "food_rate", Value: 0.05},
			},
		},
		// NEW: granary_keeper — 3 granaries; bronze/iron age
		{
			Name: "Granary Keeper", Key: "granary_keeper",
			Description:  "Build 3 Granaries.",
			Category:     "builder",
			MinAge:       "bronze_age",
			MinBuildings: map[string]int{"granary": 3},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "food_rate", Value: 0.05},
			},
		},
		// stone_mason: raised to 10 stone pits
		{
			Name: "Stone Mason", Key: "stone_mason",
			Description:  "Build 10 Stone Pits.",
			Category:     "builder",
			MinAge:       "stone_age",
			MinBuildings: map[string]int{"stone_pit": 10},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "stone_rate", Value: 0.10},
			},
		},
		// NEW: lumber_operation — 5 wood camps; stone age
		{
			Name: "Lumber Operation", Key: "lumber_operation",
			Description:  "Build 5 Wood Camps.",
			Category:     "builder",
			MinAge:       "stone_age",
			MinBuildings: map[string]int{"wood_camp": 5},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "wood_rate", Value: 0.10},
			},
		},
		// early_builder: raised to 150 buildings
		{
			Name: "Early Builder", Key: "early_builder",
			Description: "Build 150 structures total.",
			Category:    "builder",
			MinAge:      "bronze_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.03},
			},
		},
		// NEW: mining_syndicate — 5 stone pits + 1 iron mine; iron age
		{
			Name: "Mining Syndicate", Key: "mining_syndicate",
			Description:  "Build 5 Stone Pits and an Iron Mine.",
			Category:     "builder",
			MinAge:       "iron_age",
			MinBuildings: map[string]int{"stone_pit": 5, "iron_mine": 1},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "iron_rate", Value: 0.10},
			},
		},
		// NEW: forge_master — 5 smithies; iron age
		{
			Name: "Forge Master", Key: "forge_master",
			Description:  "Build 5 Smithies.",
			Category:     "builder",
			MinAge:       "iron_age",
			MinBuildings: map[string]int{"smithy": 5},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "iron_rate", Value: 0.05},
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.03},
			},
		},
		// seasoned_builder: raised to 500 buildings
		{
			Name: "Seasoned Builder", Key: "seasoned_builder",
			Description: "Build 500 structures total.",
			Category:    "builder",
			MinAge:      "iron_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.03},
			},
		},
		// master_builder: raised to 1,500 buildings
		{
			Name:        "Master Builder", Key: "master_builder",
			Description: "Build 1,500 structures total.",
			Category:    "builder",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.05},
			},
		},
		// wonder_builder: 1 wonder; reward trimmed from +10% to +5%
		{
			Name: "Wonder Builder", Key: "wonder_builder",
			Description: "Complete your first Wonder.",
			Category:    "builder", Hidden: true,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// grand_architect: raised to 5,000 buildings
		{
			Name: "Grand Architect", Key: "grand_architect",
			Description: "Build 5,000 structures total.",
			Category:    "builder", Hidden: true,
			MinAge:      "medieval_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.05},
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// wonder_collector: raised to 5 wonders
		{
			Name: "Wonder Collector", Key: "wonder_collector",
			Description: "Construct 5 Wonders.",
			Category:    "builder", Hidden: true,
			MinAge:      "colonial_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},
		// wonder_empire: raised to 12 wonders
		{
			Name: "Wonder Empire", Key: "wonder_empire",
			Description: "Construct 12 Wonders.",
			Category:    "builder", Hidden: true,
			MinAge:      "modern_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},

		// =================================================================
		// === SCHOLAR (knowledge / research progression) ===
		// =================================================================

		// knowledge_seeker: raised to 1,000 knowledge
		{
			Name: "Knowledge Seeker", Key: "knowledge_seeker",
			Description:  "Accumulate 1,000 knowledge.",
			Category:     "scholar",
			MinResources: map[string]float64{"knowledge": 1000},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.05},
			},
		},
		// first_research: 1 tech — tutorial unlock, unchanged
		{
			Name: "First Research", Key: "first_research",
			Description:  "Complete your first technology.",
			Category:     "scholar",
			MinTechCount: 1,
			Rewards: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 50},
			},
		},
		// tech_pioneer: raised to 12 techs
		{
			Name: "Tech Pioneer", Key: "tech_pioneer",
			Description:  "Research 12 technologies.",
			Category:     "scholar",
			MinTechCount: 12,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.05},
			},
		},
		// scholars_haven: raised to 20 knowledge workers + 1 library
		{
			Name: "Scholar's Haven", Key: "scholars_haven",
			Description:  "Staff 20 knowledge workers and build a Library.",
			Category:     "scholar",
			MinBuildings: map[string]int{"library": 1},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.10},
			},
		},
		// deep_thinker: raised to 18 techs
		{
			Name: "Deep Thinker", Key: "deep_thinker",
			Description:  "Research 18 technologies.",
			Category:     "scholar",
			MinAge:       "bronze_age",
			MinTechCount: 18,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.05},
			},
		},
		// philosophes: raised to 25 techs
		{
			Name: "Philosophes", Key: "philosophes",
			Description:  "Research 25 technologies.",
			Category:     "scholar",
			MinAge:       "classical_age",
			MinTechCount: 25,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.05},
			},
		},
		// renaissance_mind: raised to 30 techs; renaissance age gated
		{
			Name: "Renaissance Mind", Key: "renaissance_mind",
			Description: "Research 30 technologies.",
			Category:    "scholar", Hidden: true,
			MinAge:       "renaissance_age",
			MinTechCount: 30,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.10},
			},
		},
		// NEW: grand_library_built — build Great Library wonder; classical age
		{
			Name: "Grand Library Built", Key: "grand_library_built",
			Description:  "Construct the Great Library.",
			Category:     "scholar", Hidden: true,
			MinAge:       "classical_age",
			MinBuildings: map[string]int{"great_library": 1},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.15},
			},
		},
		// tech_master: raised to 48 techs; industrial age gated
		{
			Name: "Tech Master", Key: "tech_master",
			Description: "Research 48 technologies.",
			Category:    "scholar", Hidden: true,
			MinAge:       "industrial_age",
			MinTechCount: 48,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.15},
			},
		},
		// NEW: tech_ascendant — all 52 techs; quantum age gated
		{
			Name: "Tech Ascendant", Key: "tech_ascendant",
			Description: "Research all 52 technologies.",
			Category:    "scholar", Hidden: true,
			MinAge:       "quantum_age",
			MinTechCount: 52,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.20},
			},
		},

		// =================================================================
		// === MILITARY (soldiers / fortifications) ===
		// =================================================================

		// NEW: first_soldiers — 5 soldiers; iron age (earliest military domain age)
		{
			Name: "First Soldiers", Key: "first_soldiers",
			Description: "Train 5 soldiers.",
			Category:    "military",
			MinAge:      "iron_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.05},
			},
		},
		// war_machine: raised to 75 soldiers
		{
			Name: "War Machine", Key: "war_machine",
			Description: "Train 75 soldiers.",
			Category:    "military",
			MinAge:      "iron_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.10},
			},
		},
		// standing_army: 10 barracks + 1 legion fort; classical age
		{
			Name: "Standing Army", Key: "standing_army",
			Description:  "Build 10 Barracks and a Legion Fort.",
			Category:     "military",
			MinAge:       "classical_age",
			MinBuildings: map[string]int{"barracks": 10, "legion_fort": 1},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.05},
			},
		},
		// iron_legion: raised to 300 soldiers + 5 barracks; classical age
		{
			Name: "Iron Legion", Key: "iron_legion",
			Description:  "Train 300 soldiers and build 5 Barracks.",
			Category:     "military", Hidden: true,
			MinAge:       "classical_age",
			MinBuildings: map[string]int{"barracks": 5},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.10},
			},
		},
		// fortress_state: raised to 10 castle keeps; medieval age
		{
			Name: "Fortress State", Key: "fortress_state",
			Description:  "Build 10 Castle Keeps.",
			Category:     "military", Hidden: true,
			MinAge:       "medieval_age",
			MinBuildings: map[string]int{"castle_keep": 10},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.10},
			},
		},
		// military_superpower: raised to 1,000 soldiers; industrial age
		{
			Name: "Military Superpower", Key: "military_superpower",
			Description: "Field 1,000 soldiers.",
			Category:    "military", Hidden: true,
			MinAge:      "industrial_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.15},
			},
		},

		// =================================================================
		// === TRADE (markets / commerce) ===
		// =================================================================

		// NEW: first_market — 1 market; bronze age
		{
			Name: "First Market", Key: "first_market",
			Description:  "Build your first Market.",
			Category:     "trade",
			MinAge:       "bronze_age",
			MinBuildings: map[string]int{"market": 1},
			Rewards: []Effect{
				{Type: "instant_resource", Target: "gold", Value: 25},
			},
		},
		// merchant_guild: raised to 10 markets; iron age
		{
			Name: "Merchant Guild", Key: "merchant_guild",
			Description:  "Operate 10 Markets.",
			Category:     "trade",
			MinAge:       "iron_age",
			MinBuildings: map[string]int{"market": 10},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.05},
			},
		},
		// trade_empire: raised to 10 trading posts + 3 merchant quarters; medieval age
		{
			Name: "Trade Empire", Key: "trade_empire",
			Description:  "Build 10 Trading Posts and 3 Merchant Quarters.",
			Category:     "trade", Hidden: true,
			MinAge:       "medieval_age",
			MinBuildings: map[string]int{"trading_post": 10, "merchant_quarter": 3},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.10},
			},
		},
		// NEW: guildhall_master — 5 guildhalls; renaissance age
		{
			Name: "Guildhall Master", Key: "guildhall_master",
			Description:  "Establish 5 Guildhalls.",
			Category:     "trade", Hidden: true,
			MinAge:       "renaissance_age",
			MinBuildings: map[string]int{"guildhall": 5},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.10},
			},
		},
		// NEW: colonial_trade — 1 port + 3 colonial warehouses; colonial age
		{
			Name: "Colonial Trade Network", Key: "colonial_trade",
			Description: "Build a Port and 3 Colonial Warehouses.",
			Category:    "trade", Hidden: true,
			MinAge:      "colonial_age",
			MinBuildings: map[string]int{
				"port":               1,
				"colonial_warehouse": 3,
			},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.10},
				{Type: "permanent_bonus", Target: "expedition_reward", Value: 0.10},
			},
		},
		// gold_hoard: raised to 50,000 gold; renaissance age
		{
			Name: "Gold Hoard", Key: "gold_hoard",
			Description:  "Accumulate 50,000 gold.",
			Category:     "trade", Hidden: true,
			MinAge:       "renaissance_age",
			MinResources: map[string]float64{"gold": 50000},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.10},
			},
		},

		// =================================================================
		// === FAITH (shrines / temples) ===
		// =================================================================

		// NEW: first_shrine — 1 shrine; primitive age
		{
			Name: "First Shrine", Key: "first_shrine",
			Description:  "Build your first Shrine.",
			Category:     "faith",
			MinBuildings: map[string]int{"shrine": 1},
			Rewards: []Effect{
				{Type: "instant_resource", Target: "faith", Value: 10},
			},
		},
		// devout_settlement: raised to 10 shrines; stone age
		{
			Name: "Devout Settlement", Key: "devout_settlement",
			Description:  "Operate 10 Shrines.",
			Category:     "faith",
			MinAge:       "stone_age",
			MinBuildings: map[string]int{"shrine": 10},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "faith_rate", Value: 0.05},
			},
		},
		// temple_city: raised to 15 temples; iron age
		{
			Name: "Temple City", Key: "temple_city",
			Description:  "Build 15 Temples.",
			Category:     "faith", Hidden: true,
			MinAge:       "iron_age",
			MinBuildings: map[string]int{"temple": 15},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "faith_rate", Value: 0.10},
			},
		},
		// cathedral_age: raised to 5 cathedrals; medieval age
		{
			Name: "Cathedral Age", Key: "cathedral_age",
			Description:  "Construct 5 Cathedrals.",
			Category:     "faith", Hidden: true,
			MinAge:       "medieval_age",
			MinBuildings: map[string]int{"cathedral": 5},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "faith_rate", Value: 0.10},
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.05},
			},
		},

		// =================================================================
		// === EPOCH / LONGEVITY ===
		// =================================================================

		// NEW: first_farmers — 3 gathering camps + min tick 30 (~60s)
		{
			Name: "First Farmers", Key: "first_farmers",
			Description:  "Build 3 Gathering Camps and survive 30 ticks.",
			Category:     "epoch",
			MinTick:      30,
			MinBuildings: map[string]int{"gathering_camp": 3},
			Rewards: []Effect{
				{Type: "instant_resource", Target: "food", Value: 30},
			},
		},
		// survivor: raised to 3,000 ticks (~100 min at 2s/tick)
		{
			Name: "Survivor", Key: "survivor",
			Description: "Survive 3,000 ticks.",
			Category:    "epoch",
			MinTick:     3000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// enduring_civilization: raised to 15,000 ticks
		{
			Name: "Enduring Civilization", Key: "enduring_civilization",
			Description: "Survive 15,000 ticks.",
			Category:    "epoch", Hidden: true,
			MinTick: 15000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// NEW: age_hopper — reach classical age (5th age)
		{
			Name: "Age Hopper", Key: "age_hopper",
			Description: "Advance through 5 ages.",
			Category:    "epoch",
			MinAge:      "classical_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// industrial_titan: raised to 3,000 coal and 1,500 iron; industrial age
		{
			Name: "Industrial Titan", Key: "industrial_titan",
			Description:  "Stockpile 3,000 coal and 1,500 iron.",
			Category:     "epoch", Hidden: true,
			MinAge:       "industrial_age",
			MinResources: map[string]float64{"coal": 3000, "iron": 1500},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},
		// power_grid: raised to 15 coal plants + 3 steam turbines; victorian age
		{
			Name: "Power Grid", Key: "power_grid",
			Description:  "Build 15 Coal Plants and 3 Steam Turbines.",
			Category:     "epoch", Hidden: true,
			MinAge:       "victorian_age",
			MinBuildings: map[string]int{"coal_plant": 15, "steam_turbine": 3},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},

		// =================================================================
		// === AGES (advance to each age) ===
		// Rewards calibrated: early ages give instant resources,
		// later ages give permanent_bonus capped at +20% per milestone.
		// =================================================================

		{
			Name: "Bronze Age Pioneer", Key: "bronze_pioneer",
			Description: "Advance to the Bronze Age.",
			Category:    "ages",
			MinAge:      "bronze_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "iron", Value: 30},
				{Type: "instant_resource", Target: "gold", Value: 30},
			},
		},
		{
			Name: "Iron Forged", Key: "iron_forged",
			Description: "Advance to the Iron Age.",
			Category:    "ages",
			MinAge:      "iron_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "coal", Value: 40},
				{Type: "permanent_bonus", Target: "iron_rate", Value: 0.10},
			},
		},
		{
			Name: "Classical Scholar", Key: "classical_scholar",
			Description: "Advance to the Classical Age.",
			Category:    "ages",
			MinAge:      "classical_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 150},
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.10},
			},
		},
		{
			Name: "Medieval Lord", Key: "medieval_lord",
			Description: "Advance to the Medieval Age.",
			Category:    "ages",
			MinAge:      "medieval_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "faith", Value: 50},
				{Type: "instant_resource", Target: "steel", Value: 25},
			},
		},
		{
			Name: "Enlightened", Key: "enlightened",
			Description: "Advance to the Renaissance Age.",
			Category:    "ages",
			MinAge:      "renaissance_age",
			Rewards: []Effect{
				{Type: "instant_resource", Target: "culture", Value: 75},
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.15},
			},
		},
		{
			Name: "Colonial Power", Key: "colonial_power",
			Description: "Advance to the Colonial Age.",
			Category:    "ages", Hidden: true,
			MinAge: "colonial_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.15},
				{Type: "permanent_bonus", Target: "expedition_reward", Value: 0.10},
			},
		},
		{
			Name: "Industrial Revolution", Key: "industrial_revolution",
			Description: "Advance to the Industrial Age.",
			Category:    "ages", Hidden: true,
			MinAge: "industrial_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},
		{
			Name: "Victorian Innovation", Key: "victorian_innovation",
			Description: "Advance to the Victorian Age.",
			Category:    "ages", Hidden: true,
			MinAge: "victorian_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},
		{
			Name: "Electric Dawn", Key: "electric_dawn",
			Description: "Advance to the Electric Age.",
			Category:    "ages", Hidden: true,
			MinAge: "electric_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},
		{
			Name: "Atomic Pioneer", Key: "atomic_pioneer",
			Description: "Advance to the Atomic Age.",
			Category:    "ages", Hidden: true,
			MinAge: "atomic_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},
		{
			Name: "Modern Era", Key: "modern_era",
			Description: "Advance to the Modern Age.",
			Category:    "ages", Hidden: true,
			MinAge: "modern_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Information Pioneer", Key: "information_pioneer",
			Description: "Advance to the Information Age.",
			Category:    "ages", Hidden: true,
			MinAge: "information_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.20},
			},
		},
		{
			Name: "Digital Native", Key: "digital_native",
			Description: "Advance to the Digital Age.",
			Category:    "ages", Hidden: true,
			MinAge: "digital_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Cyberpunk", Key: "cyberpunk_milestone",
			Description: "Advance to the Cyberpunk Age.",
			Category:    "ages", Hidden: true,
			MinAge: "cyberpunk_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gather_rate", Value: 0.15},
			},
		},
		{
			Name: "Fusion Pioneer", Key: "fusion_pioneer",
			Description: "Advance to the Fusion Age.",
			Category:    "ages", Hidden: true,
			MinAge: "fusion_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Space Explorer", Key: "space_explorer",
			Description: "Advance to the Space Age.",
			Category:    "ages", Hidden: true,
			MinAge: "space_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
				{Type: "permanent_bonus", Target: "expedition_reward", Value: 0.20},
			},
		},
		{
			Name: "Star Voyager", Key: "star_voyager",
			Description: "Advance to the Interstellar Age.",
			Category:    "ages", Hidden: true,
			MinAge: "interstellar_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Galactic Emperor", Key: "galactic_emperor",
			Description: "Advance to the Galactic Age.",
			Category:    "ages", Hidden: true,
			MinAge: "galactic_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Quantum Master", Key: "quantum_master",
			Description: "Advance to the Quantum Age.",
			Category:    "ages", Hidden: true,
			MinAge: "quantum_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Transcended", Key: "transcended",
			Description: "Advance to the Transcendent Age.",
			Category:    "ages", Hidden: true,
			MinAge: "transcendent_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.50},
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
	return []string{"settlement", "builder", "scholar", "military", "trade", "faith", "epoch", "ages"}
}

// MilestoneCategoryNames returns display names for categories
func MilestoneCategoryNames() map[string]string {
	return map[string]string{
		"settlement": "Settlement",
		"builder":    "Builder",
		"scholar":    "Scholar",
		"military":   "Military",
		"trade":      "Trade",
		"faith":      "Faith",
		"epoch":      "Epoch",
		"ages":       "Ages",
	}
}
