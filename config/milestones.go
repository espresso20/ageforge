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
		// small_village: raised to 5,000 — requires real investment
		{
			Name: "Small Village", Key: "small_village",
			Description:   "Reach a population of 5,000.",
			Category:      "settlement",
			MinPopulation: 5000,
			Rewards: []Effect{
				{Type: "instant_resource", Target: "wood", Value: 50},
			},
		},
		// bustling_town: raised to 50,000
		{
			Name: "Bustling Town", Key: "bustling_town",
			Description:   "Reach a population of 50,000.",
			Category:      "settlement",
			MinPopulation: 50000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "food_rate", Value: 0.05},
			},
		},
		// growing_city: raised to 500,000; bronze_age gate maintained
		{
			Name: "Growing City", Key: "growing_city",
			Description:   "Reach a population of 500,000.",
			Category:      "settlement",
			MinAge:        "bronze_age",
			MinPopulation: 500000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "food_rate", Value: 0.10},
			},
		},
		// metropolis: raised to 10M pop; iron_age gated
		{
			Name: "Metropolis", Key: "metropolis",
			Description: "Reach a population of 10,000,000.",
			Category:    "settlement", Hidden: true,
			MinAge:        "iron_age",
			MinPopulation: 10000000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},
		// megalopolis: raised to 1B pop
		{
			Name: "Megalopolis", Key: "megalopolis",
			Description: "Reach a population of 1,000,000,000.",
			Category:    "settlement", Hidden: true,
			MinAge:        "classical_age",
			MinPopulation: 1000000000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		// urban_sprawl — 100M pop, medieval_age gated
		{
			Name: "Urban Sprawl", Key: "urban_sprawl",
			Description: "Reach a population of 100,000,000.",
			Category:    "settlement", Hidden: true,
			MinAge:        "medieval_age",
			MinPopulation: 100000000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},
		// global_city — 10B pop, industrial_age gated
		{
			Name: "Global City", Key: "global_city",
			Description: "Reach a population of 10,000,000,000.",
			Category:    "settlement", Hidden: true,
			MinAge:        "industrial_age",
			MinPopulation: 10000000000,
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
		// storage_network — 10 storage pits; stone/bronze age
		{
			Name: "Storage Network", Key: "storage_network",
			Description:  "Build 10 Storage Pits.",
			Category:     "builder",
			MinAge:       "stone_age",
			MinBuildings: map[string]int{"storage_pit": 10},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "food_rate", Value: 0.05},
			},
		},
		// granary_keeper — 25 granaries; bronze age
		{
			Name: "Granary Keeper", Key: "granary_keeper",
			Description:  "Build 25 Granaries.",
			Category:     "builder",
			MinAge:       "bronze_age",
			MinBuildings: map[string]int{"granary": 25},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "food_rate", Value: 0.05},
			},
		},
		// stone_mason: raised to 50 stone pits
		{
			Name: "Stone Mason", Key: "stone_mason",
			Description:  "Build 50 Stone Pits.",
			Category:     "builder",
			MinAge:       "stone_age",
			MinBuildings: map[string]int{"stone_pit": 50},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "stone_rate", Value: 0.10},
			},
		},
		// lumber_operation — 25 wood camps; stone age
		{
			Name: "Lumber Operation", Key: "lumber_operation",
			Description:  "Build 25 Wood Camps.",
			Category:     "builder",
			MinAge:       "stone_age",
			MinBuildings: map[string]int{"wood_camp": 25},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "wood_rate", Value: 0.10},
			},
		},
		// early_builder: raised to 500 buildings
		{
			Name: "Early Builder", Key: "early_builder",
			Description: "Build 500 structures total.",
			Category:    "builder",
			MinAge:      "bronze_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.03},
			},
		},
		// mining_syndicate — 25 stone pits + 10 iron mines; iron age
		{
			Name: "Mining Syndicate", Key: "mining_syndicate",
			Description:  "Build 25 Stone Pits and 10 Iron Mines.",
			Category:     "builder",
			MinAge:       "iron_age",
			MinBuildings: map[string]int{"stone_pit": 25, "iron_mine": 10},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "iron_rate", Value: 0.10},
			},
		},
		// forge_master — 15 smithies; iron age
		{
			Name: "Forge Master", Key: "forge_master",
			Description:  "Build 15 Smithies.",
			Category:     "builder",
			MinAge:       "iron_age",
			MinBuildings: map[string]int{"smithy": 15},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "iron_rate", Value: 0.05},
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.03},
			},
		},
		// seasoned_builder: raised to 2,000 buildings
		{
			Name: "Seasoned Builder", Key: "seasoned_builder",
			Description: "Build 2,000 structures total.",
			Category:    "builder",
			MinAge:      "iron_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.03},
			},
		},
		// master_builder: raised to 5,000 buildings
		{
			Name:        "Master Builder", Key: "master_builder",
			Description: "Build 5,000 structures total.",
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
		// grand_architect: raised to 20,000 buildings
		{
			Name: "Grand Architect", Key: "grand_architect",
			Description: "Build 20,000 structures total.",
			Category:    "builder", Hidden: true,
			MinAge:      "medieval_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.05},
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// wonder_collector: raised to 8 wonders
		{
			Name: "Wonder Collector", Key: "wonder_collector",
			Description: "Construct 8 Wonders.",
			Category:    "builder", Hidden: true,
			MinAge:      "colonial_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},
		// wonder_empire: raised to 15 wonders
		{
			Name: "Wonder Empire", Key: "wonder_empire",
			Description: "Construct 15 Wonders.",
			Category:    "builder", Hidden: true,
			MinAge:      "modern_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},

		// =================================================================
		// === SCHOLAR (knowledge / research progression) ===
		// =================================================================

		// knowledge_seeker: raised to 10,000 knowledge
		{
			Name: "Knowledge Seeker", Key: "knowledge_seeker",
			Description:  "Accumulate 10,000 knowledge.",
			Category:     "scholar",
			MinResources: map[string]float64{"knowledge": 10000},
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
		// tech_pioneer: raised to 15 techs
		{
			Name: "Tech Pioneer", Key: "tech_pioneer",
			Description:  "Research 15 technologies.",
			Category:     "scholar",
			MinTechCount: 15,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.05},
			},
		},
		// scholars_haven: raised to 50 knowledge workers + 3 libraries
		{
			Name: "Scholar's Haven", Key: "scholars_haven",
			Description:  "Staff 50 knowledge workers and build 3 Libraries.",
			Category:     "scholar",
			MinBuildings: map[string]int{"library": 3},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.10},
			},
		},
		// deep_thinker: raised to 25 techs
		{
			Name: "Deep Thinker", Key: "deep_thinker",
			Description:  "Research 25 technologies.",
			Category:     "scholar",
			MinAge:       "bronze_age",
			MinTechCount: 25,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.05},
			},
		},
		// philosophes: raised to 35 techs
		{
			Name: "Philosophes", Key: "philosophes",
			Description:  "Research 35 technologies.",
			Category:     "scholar",
			MinAge:       "classical_age",
			MinTechCount: 35,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.05},
			},
		},
		// renaissance_mind: raised to 42 techs; renaissance age gated
		{
			Name: "Renaissance Mind", Key: "renaissance_mind",
			Description: "Research 42 technologies.",
			Category:    "scholar", Hidden: true,
			MinAge:       "renaissance_age",
			MinTechCount: 42,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.10},
			},
		},
		// grand_library_built — build 5 Great Libraries; classical age
		{
			Name: "Grand Library Built", Key: "grand_library_built",
			Description:  "Construct 5 Great Libraries.",
			Category:     "scholar", Hidden: true,
			MinAge:       "classical_age",
			MinBuildings: map[string]int{"great_library": 5},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.15},
			},
		},
		// tech_master: raised to 50 techs; industrial age gated
		{
			Name: "Tech Master", Key: "tech_master",
			Description: "Research 50 technologies.",
			Category:    "scholar", Hidden: true,
			MinAge:       "industrial_age",
			MinTechCount: 50,
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
		// war_machine: raised to 250 soldiers
		{
			Name: "War Machine", Key: "war_machine",
			Description: "Train 250 soldiers.",
			Category:    "military",
			MinAge:      "iron_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.10},
			},
		},
		// standing_army: 100 soldiers + 10 barracks; classical age
		{
			Name: "Standing Army", Key: "standing_army",
			Description:  "Train 100 soldiers and build 10 Barracks.",
			Category:     "military",
			MinAge:       "classical_age",
			MinBuildings: map[string]int{"barracks": 10},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.05},
			},
		},
		// iron_legion: raised to 500 soldiers + 10 barracks; classical age
		{
			Name: "Iron Legion", Key: "iron_legion",
			Description:  "Train 500 soldiers and build 10 Barracks.",
			Category:     "military", Hidden: true,
			MinAge:       "classical_age",
			MinBuildings: map[string]int{"barracks": 10},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.10},
			},
		},
		// fortress_state: raised to 20 castle keeps; medieval age
		{
			Name: "Fortress State", Key: "fortress_state",
			Description:  "Build 20 Castle Keeps.",
			Category:     "military", Hidden: true,
			MinAge:       "medieval_age",
			MinBuildings: map[string]int{"castle_keep": 20},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.10},
			},
		},
		// military_superpower: raised to 2,000 soldiers; industrial age
		{
			Name: "Military Superpower", Key: "military_superpower",
			Description: "Field 2,000 soldiers.",
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
		// merchant_guild: raised to 8 active markets; iron age
		{
			Name: "Merchant Guild", Key: "merchant_guild",
			Description:  "Operate 8 Markets.",
			Category:     "trade",
			MinAge:       "iron_age",
			MinBuildings: map[string]int{"market": 8},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.05},
			},
		},
		// trade_empire: raised to 20 trading posts + 8 merchant quarters; medieval age
		{
			Name: "Trade Empire", Key: "trade_empire",
			Description:  "Build 20 Trading Posts and 8 Merchant Quarters.",
			Category:     "trade", Hidden: true,
			MinAge:       "medieval_age",
			MinBuildings: map[string]int{"trading_post": 20, "merchant_quarter": 8},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.10},
			},
		},
		// guildhall_master — 10 guildhalls; renaissance age
		{
			Name: "Guildhall Master", Key: "guildhall_master",
			Description:  "Establish 10 Guildhalls.",
			Category:     "trade", Hidden: true,
			MinAge:       "renaissance_age",
			MinBuildings: map[string]int{"guildhall": 10},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.10},
			},
		},
		// colonial_trade — 3 ports + 5 colonial warehouses; colonial age
		{
			Name: "Colonial Trade Network", Key: "colonial_trade",
			Description: "Build 3 Ports and 5 Colonial Warehouses.",
			Category:    "trade", Hidden: true,
			MinAge:      "colonial_age",
			MinBuildings: map[string]int{
				"port":               3,
				"colonial_warehouse": 5,
			},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.10},
				{Type: "permanent_bonus", Target: "expedition_reward", Value: 0.10},
			},
		},
		// gold_hoard: raised to 1,000,000 gold; renaissance age
		{
			Name: "Gold Hoard", Key: "gold_hoard",
			Description:  "Accumulate 1,000,000 gold.",
			Category:     "trade", Hidden: true,
			MinAge:       "renaissance_age",
			MinResources: map[string]float64{"gold": 1000000},
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
		// devout_settlement: raised to 25 shrines; stone age
		{
			Name: "Devout Settlement", Key: "devout_settlement",
			Description:  "Operate 25 Shrines.",
			Category:     "faith",
			MinAge:       "stone_age",
			MinBuildings: map[string]int{"shrine": 25},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "faith_rate", Value: 0.05},
			},
		},
		// temple_city: raised to 50 temples; iron age
		{
			Name: "Temple City", Key: "temple_city",
			Description:  "Build 50 Temples.",
			Category:     "faith", Hidden: true,
			MinAge:       "iron_age",
			MinBuildings: map[string]int{"temple": 50},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "faith_rate", Value: 0.10},
			},
		},
		// cathedral_age: raised to 10 cathedrals; medieval age
		{
			Name: "Cathedral Age", Key: "cathedral_age",
			Description:  "Construct 10 Cathedrals.",
			Category:     "faith", Hidden: true,
			MinAge:       "medieval_age",
			MinBuildings: map[string]int{"cathedral": 10},
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
		// survivor: raised to 10,000 ticks
		{
			Name: "Survivor", Key: "survivor",
			Description: "Survive 10,000 ticks.",
			Category:    "epoch",
			MinTick:     10000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// enduring_civilization: raised to 50,000 ticks
		{
			Name: "Enduring Civilization", Key: "enduring_civilization",
			Description: "Survive 50,000 ticks.",
			Category:    "epoch", Hidden: true,
			MinTick: 50000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// age_hopper — reach classical age (5th age) AND have 10+ techs researched
		{
			Name: "Age Hopper", Key: "age_hopper",
			Description: "Advance through 5 ages and research 10 technologies.",
			Category:    "epoch",
			MinAge:      "classical_age",
			MinTechCount: 10,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// industrial_titan: raised to 10,000 coal and 5,000 iron_ore; industrial age
		{
			Name: "Industrial Titan", Key: "industrial_titan",
			Description:  "Stockpile 10,000 coal and 5,000 iron ore.",
			Category:     "epoch", Hidden: true,
			MinAge:       "industrial_age",
			MinResources: map[string]float64{"coal": 10000, "iron_ore": 5000},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},
		// power_grid: raised to 50 coal plants + 10 steam turbines; victorian age
		{
			Name: "Power Grid", Key: "power_grid",
			Description:  "Build 50 Coal Plants and 10 Steam Turbines.",
			Category:     "epoch", Hidden: true,
			MinAge:       "victorian_age",
			MinBuildings: map[string]int{"coal_plant": 50, "steam_turbine": 10},
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
