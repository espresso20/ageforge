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
	// Flavor is an optional one-line quip shown in the completion toast/log
	// alongside the mechanical reward text. Purely cosmetic — never affects
	// conditions or rewards. Empty = no quip (falls back to plain announcement).
	Flavor   string
	Category string // "settlement", "scholar", "builder", "military", "trade", "faith", "epoch", "ages"
	Hidden   bool   // hidden milestones only revealed when close to completion or done
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
	Flavor        string   // optional quip shown in the chain-complete announcement (cosmetic)
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
			Flavor:        "From one sad hut to a sprawling city. The founding myth will leave out the hut.",
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
			Flavor:        "Your scholars have learned enough to be insufferable at every dinner party.",
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
			Flavor:        "You build monuments faster than rivals can finish being impressed by them.",
			BoostValue:    2.5,
			BoostDuration: 150,
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
			Flavor:        "Your neighbours have stopped sending letters and started sending apologies.",
			BoostValue:    2.5,
			BoostDuration: 150,
		},
		{
			Name:     "Trade Chain",
			Key:      "trade_chain",
			Category: "trade",
			MilestoneKeys: []string{
				"first_market", "merchant_guild", "caravan_network",
				"merchant_princes", "trade_empire", "maritime_empire",
			},
			Title:         "The Merchants",
			Flavor:        "Every road leads to your market, and every road charges a modest toll.",
			BoostValue:    2.5,
			BoostDuration: 150,
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
			Flavor:        "You marched through five ages and only set fire to most of them.",
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
			Flavor:       "A roof. Four walls. Civilization has, technically, begun.",
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
			Flavor:        "Five thousand souls, and already someone wants to form a committee.",
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
			Flavor:        "A proper town now, with a market, a tavern, and at least one feud.",
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
			Flavor:        "Half a million people. The traffic has invented itself spontaneously.",
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
			Flavor:      "Ten million citizens. You have officially lost track of all their names.",
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
			Flavor:      "A billion people. The census takers have requested early retirement.",
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
			Flavor:      "The city has no edges anymore. It just sort of keeps being the city.",
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
			Flavor:      "Ten billion. The planet has filed a formal complaint about the crowding.",
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
			Flavor:       "A place to put your things, so they stop becoming someone else's things.",
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
			Flavor:       "Ten pits of carefully hoarded surplus. The hoarding instinct, finally, pays off.",
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
			Flavor:       "Twenty-five granaries. The mice consider this a personal invitation.",
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
			Flavor:       "Fifty pits of honest stone. Your masons can finally stop improvising with mud.",
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
			Flavor:       "Twenty-five camps and a quiet word of apology owed to the forest.",
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
			Flavor:      "Five hundred buildings. You've stopped naming them and started numbering them.",
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
			Flavor:       "Stone and iron in industrial quantities. The hills are getting visibly nervous.",
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
			Flavor:       "Fifteen forges roaring at once. The whole valley now smells faintly of ambition.",
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
			Flavor:      "Two thousand buildings. You could get lost in your own civilization, and frequently do.",
			Category:    "builder",
			MinAge:      "iron_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.03},
			},
		},
		// master_builder: raised to 5,000 buildings
		{
			Name: "Master Builder", Key: "master_builder",
			Description: "Build 5,000 structures total.",
			Flavor:      "Five thousand structures. Future archaeologists will assume you were showing off.",
			Category:    "builder",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.05},
			},
		},
		// wonder_builder: 1 wonder; reward trimmed from +10% to +5%
		{
			Name: "Wonder Builder", Key: "wonder_builder",
			Description: "Complete your first Wonder.",
			Flavor:      "You built a Wonder. Your neighbors are impressed. One is drafting a strongly worded letter.",
			Category:    "builder", Hidden: true,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// grand_architect: raised to 20,000 buildings
		{
			Name: "Grand Architect", Key: "grand_architect",
			Description: "Build 20,000 structures total.",
			Flavor:      "Twenty thousand buildings. The mapmakers have unionized and gone home.",
			Category:    "builder", Hidden: true,
			MinAge: "medieval_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "build_cost", Value: -0.05},
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// wonder_collector: raised to 8 wonders
		{
			Name: "Wonder Collector", Key: "wonder_collector",
			Description: "Construct 8 Wonders.",
			Flavor:      "Eight Wonders. Tourists from rival empires now visit just to feel inadequate.",
			Category:    "builder", Hidden: true,
			MinAge: "colonial_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},
		// wonder_empire: raised to 15 wonders
		{
			Name: "Wonder Empire", Key: "wonder_empire",
			Description: "Construct 15 Wonders.",
			Flavor:      "Fifteen Wonders. At this point you're just collecting them, like a very expensive hobby.",
			Category:    "builder", Hidden: true,
			MinAge: "modern_age",
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
			Flavor:       "Ten thousand units of knowledge, and the faint, dawning awareness of how little you know.",
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
			Flavor:       "Your first real invention. Someone will improve it tomorrow and take all the credit.",
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
			Flavor:       "Fifteen breakthroughs. The wheel was one of them, eventually.",
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
			Flavor:       "Three libraries, fifty scholars, and a fierce ongoing dispute about quiet hours.",
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
			Flavor:       "Twenty-five technologies. Your scholars have begun ending sentences with 'well, actually.'",
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
			Flavor:       "Thirty-five technologies. The philosophers now argue about things on purpose.",
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
			Flavor:      "Forty-two technologies. The answer to everything, apparently, requires a sequel.",
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
			Description: "Construct 5 Great Libraries.",
			Flavor:      "Five Great Libraries. The collected wisdom of the age, and overdue fines to match.",
			Category:    "scholar", Hidden: true,
			MinAge:       "classical_age",
			MinBuildings: map[string]int{"great_library": 5},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.15},
			},
		},
		// tech_master: raised to 50 techs; industrial age gated.
		// Capstone broadened — keeps research_speed, adds a touch of production_all.
		{
			Name: "Tech Master", Key: "tech_master",
			Description: "Research 50 technologies.",
			Flavor:      "Fifty technologies mastered. You now understand the universe well enough to be properly worried.",
			Category:    "scholar", Hidden: true,
			MinAge:       "industrial_age",
			MinTechCount: 50,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "research_speed", Value: 0.10},
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// NEW: tech_ascendant — all 52 techs; quantum age gated
		{
			Name: "Tech Ascendant", Key: "tech_ascendant",
			Description: "Research all 52 technologies.",
			Flavor:      "Every technology, researched. The tech tree is bald. You did this.",
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
			Flavor:      "Five soldiers. Technically an army, if you squint and don't ask them to march in step.",
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
			Flavor:      "Two hundred and fifty soldiers. The neighbours have started being noticeably more polite.",
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
			Flavor:       "A standing army that actually stands where you tell it. Discipline is its own miracle.",
			Category:     "military",
			MinAge:       "classical_age",
			MinBuildings: map[string]int{"barracks": 10},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.05},
			},
		},
		// iron_legion: 500 soldiers + 10 barracks; classical age.
		// Broadened to production_all so the chain helps the whole economy.
		{
			Name: "Iron Legion", Key: "iron_legion",
			Description: "Train 500 soldiers and build 10 Barracks.",
			Flavor:      "Five hundred soldiers in iron. The blacksmiths request you stop, just for a week.",
			Category:    "military", Hidden: true,
			MinAge:       "classical_age",
			MinBuildings: map[string]int{"barracks": 10},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// fortress_state: 20 castle keeps; medieval age. Dual payout —
		// keeps a military_power bonus but adds broad production_all.
		{
			Name: "Fortress State", Key: "fortress_state",
			Description: "Build 20 Castle Keeps.",
			Flavor:      "Twenty castle keeps. Your realm is now less a country and more a very pointed suggestion.",
			Category:    "military", Hidden: true,
			MinAge:       "medieval_age",
			MinBuildings: map[string]int{"castle_keep": 20},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "military_power", Value: 0.10},
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// military_superpower: 2,000 soldiers; industrial age.
		// Capstone broadened to production_all (was military_power).
		{
			Name: "Military Superpower", Key: "military_superpower",
			Description: "Field 2,000 soldiers.",
			Flavor:      "Two thousand troops. Diplomacy is now mostly other people agreeing with you, quickly.",
			Category:    "military", Hidden: true,
			MinAge: "industrial_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},

		// =================================================================
		// === TRADE (markets / commerce) ===
		// =================================================================

		// NEW: first_market — 1 market; bronze age
		{
			Name: "First Market", Key: "first_market",
			Description:  "Build your first Market.",
			Flavor:       "A market. Now your people can argue over prices instead of just taking things.",
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
			Flavor:       "Eight markets and a guild that already has strong opinions about everyone else's.",
			Category:     "trade",
			MinAge:       "iron_age",
			MinBuildings: map[string]int{"market": 8},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.05},
			},
		},
		// caravan_network: 5 trading posts; classical age (trading_post unlocks iron age)
		{
			Name: "Caravan Network", Key: "caravan_network",
			Description:  "Operate 5 Trading Posts.",
			Flavor:       "Five trading posts. The caravans now have a route, a schedule, and a complicated rivalry.",
			Category:     "trade",
			MinAge:       "classical_age",
			MinBuildings: map[string]int{"trading_post": 5},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.10},
			},
		},
		// merchant_princes: 12 trading posts + 4 merchant quarters; medieval age.
		// Broadened to production_all so the mid-chain payout helps the whole economy.
		{
			Name: "Merchant Princes", Key: "merchant_princes",
			Description: "Build 12 Trading Posts and 4 Merchant Quarters.",
			Flavor:      "Merchant princes now. They wear nicer hats than you and they know it.",
			Category:    "trade", Hidden: true,
			MinAge:       "medieval_age",
			MinBuildings: map[string]int{"trading_post": 12, "merchant_quarter": 4},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// trade_empire: capstone scaled to 30 trading posts + 12 merchant quarters;
		// renaissance age. Broadened to production_all (was gold_rate).
		{
			Name: "Trade Empire", Key: "trade_empire",
			Description: "Build 30 Trading Posts and 12 Merchant Quarters.",
			Flavor:      "A trade empire. Somewhere, a coin is changing hands in your name right now. And now. And now.",
			Category:    "trade", Hidden: true,
			MinAge:       "renaissance_age",
			MinBuildings: map[string]int{"trading_post": 30, "merchant_quarter": 12},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},
		// maritime_empire: Trade Chain capstone (5→6) — rewards the harbour
		// lineage. 5 harbours + 5 ports + 2 seaports; gated at modern_age because
		// seaport (tier 2 harbour) only unlocks then. Broad production_all payout
		// to match the rest of the late chain.
		{
			Name: "Maritime Empire", Key: "maritime_empire",
			Description: "Build 5 Harbours, 5 Ports, and 2 Seaports.",
			Flavor:      "The sea is now a road, and you are charging tolls on it.",
			Category:    "trade", Hidden: true,
			MinAge:       "modern_age",
			MinBuildings: map[string]int{"harbor": 5, "port": 5, "seaport": 2},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
				{Type: "permanent_bonus", Target: "gold_rate", Value: 0.10},
			},
		},
		// guildhall_master — 10 guildhalls; renaissance age
		{
			Name: "Guildhall Master", Key: "guildhall_master",
			Description: "Establish 10 Guildhalls.",
			Flavor:      "Ten guildhalls, each convinced it secretly runs the city. One of them is right.",
			Category:    "trade", Hidden: true,
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
			Flavor:      "A colonial trade network spanning oceans, and a warehouse full of things best not inventoried.",
			Category:    "trade", Hidden: true,
			MinAge: "colonial_age",
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
			Description: "Accumulate 1,000,000 gold.",
			Flavor:      "A million gold. You could swim in it, if your advisors would stop fainting at the suggestion.",
			Category:    "trade", Hidden: true,
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
			Flavor:       "A shrine. The people now have somewhere official to ask for better weather.",
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
			Flavor:       "Twenty-five shrines. The priests have begun, gently, competing for foot traffic.",
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
			Description: "Build 50 Temples.",
			Flavor:      "Fifty temples. A holy city, with surprisingly aggressive parking.",
			Category:    "faith", Hidden: true,
			MinAge:       "iron_age",
			MinBuildings: map[string]int{"temple": 50},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "faith_rate", Value: 0.10},
			},
		},
		// cathedral_age: raised to 10 cathedrals; medieval age
		{
			Name: "Cathedral Age", Key: "cathedral_age",
			Description: "Construct 10 Cathedrals.",
			Flavor:      "Ten cathedrals, each taking a generation to build. The architects left detailed notes for their grandchildren.",
			Category:    "faith", Hidden: true,
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
			Flavor:       "You've stopped wandering and started farming. Bold move. We'll see how it goes.",
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
			Flavor:      "Ten thousand ticks survived. Whatever you're doing, it's apparently working.",
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
			Flavor:      "Fifty thousand ticks. Other civilizations are now studying you in their history classes.",
			Category:    "epoch", Hidden: true,
			MinTick: 50000,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// age_hopper — reach classical age (5th age) AND have 10+ techs researched
		{
			Name: "Age Hopper", Key: "age_hopper",
			Description:  "Advance through 5 ages and research 10 technologies.",
			Flavor:       "Five ages, ten techs, and a frankly cavalier attitude toward the passage of time.",
			Category:     "epoch",
			MinAge:       "classical_age",
			MinTechCount: 10,
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.05},
			},
		},
		// industrial_titan: raised to 10,000 coal and 5,000 iron_ore; industrial age
		{
			Name: "Industrial Titan", Key: "industrial_titan",
			Description: "Stockpile 10,000 coal and 5,000 iron ore.",
			Flavor:      "Coal and ore by the mountain. The sky is a colour now, and that colour is 'industry.'",
			Category:    "epoch", Hidden: true,
			MinAge:       "industrial_age",
			MinResources: map[string]float64{"coal": 10000, "iron_ore": 5000},
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},
		// power_grid: raised to 50 coal plants + 10 steam turbines; victorian age
		{
			Name: "Power Grid", Key: "power_grid",
			Description: "Build 50 Coal Plants and 10 Steam Turbines.",
			Flavor:      "The lights stay on all night now. Nobody is entirely sure this is an improvement.",
			Category:    "epoch", Hidden: true,
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
			Flavor:      "You discovered that mixing two soft metals makes a hard one. Revolutionary. Slightly suspicious.",
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
			Flavor:      "Iron. Harder, cheaper, everywhere. The bronze merchants are taking it badly.",
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
			Flavor:      "An age of philosophy, democracy, and columns. So many columns.",
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
			Flavor:      "Castles, knights, and the firm conviction that bathing causes illness.",
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
			Flavor:      "Art, science, and a renewed enthusiasm for painting people who didn't ask to be painted.",
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
			Flavor:      "You have discovered distant lands. The distant lands point out they were never lost.",
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
			Flavor:      "The machines do the work now. The workers have been promoted to operating the machines, for less.",
			Category:    "ages", Hidden: true,
			MinAge: "industrial_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},
		{
			Name: "Victorian Innovation", Key: "victorian_innovation",
			Description: "Advance to the Victorian Age.",
			Flavor:      "An age of progress, propriety, and an absolutely staggering amount of soot.",
			Category:    "ages", Hidden: true,
			MinAge: "victorian_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.10},
			},
		},
		{
			Name: "Electric Dawn", Key: "electric_dawn",
			Description: "Advance to the Electric Age.",
			Flavor:      "You have caught lightning and put it in the walls. This will surely have no downsides.",
			Category:    "ages", Hidden: true,
			MinAge: "electric_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},
		{
			Name: "Atomic Pioneer", Key: "atomic_pioneer",
			Description: "Advance to the Atomic Age.",
			Flavor:      "You have split the atom. The atom is taking it personally.",
			Category:    "ages", Hidden: true,
			MinAge: "atomic_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.15},
			},
		},
		{
			Name: "Modern Era", Key: "modern_era",
			Description: "Advance to the Modern Age.",
			Flavor:      "The modern age, where everything is convenient and nobody can find the time.",
			Category:    "ages", Hidden: true,
			MinAge: "modern_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Information Pioneer", Key: "information_pioneer",
			Description: "Advance to the Information Age.",
			Flavor:      "All the knowledge of humankind, in everyone's pocket. They use it to argue.",
			Category:    "ages", Hidden: true,
			MinAge: "information_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "knowledge_rate", Value: 0.20},
			},
		},
		{
			Name: "Digital Native", Key: "digital_native",
			Description: "Advance to the Digital Age.",
			Flavor:      "Everything is digital now, including the problems, which are also harder to turn off.",
			Category:    "ages", Hidden: true,
			MinAge: "digital_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Cyberpunk", Key: "cyberpunk_milestone",
			Description: "Advance to the Cyberpunk Age.",
			Flavor:      "High tech, low life, and neon everywhere. Even the rain has a brand sponsor.",
			Category:    "ages", Hidden: true,
			MinAge: "cyberpunk_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "gather_rate", Value: 0.15},
			},
		},
		{
			Name: "Fusion Pioneer", Key: "fusion_pioneer",
			Description: "Advance to the Fusion Age.",
			Flavor:      "You've built a tiny sun and put it in a box. The fire department has no notes, only fear.",
			Category:    "ages", Hidden: true,
			MinAge: "fusion_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Space Explorer", Key: "space_explorer",
			Description: "Advance to the Space Age.",
			Flavor:      "You have left the planet. The planet is, frankly, relieved to get some space.",
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
			Flavor:      "Other stars now. The commute is appalling but the views are unmatched.",
			Category:    "ages", Hidden: true,
			MinAge: "interstellar_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Galactic Emperor", Key: "galactic_emperor",
			Description: "Advance to the Galactic Age.",
			Flavor:      "An empire across the galaxy. The paperwork now arrives at the speed of light, which is still too slow.",
			Category:    "ages", Hidden: true,
			MinAge: "galactic_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Quantum Master", Key: "quantum_master",
			Description: "Advance to the Quantum Age.",
			Flavor:      "You now operate on the quantum level, where your civilization both exists and doesn't until observed.",
			Category:    "ages", Hidden: true,
			MinAge: "quantum_age",
			Rewards: []Effect{
				{Type: "permanent_bonus", Target: "production_all", Value: 0.20},
			},
		},
		{
			Name: "Transcended", Key: "transcended",
			Description: "Advance to the Transcendent Age.",
			Flavor:      "You have transcended matter, time, and need. You still, somehow, check the production graph.",
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
