package config

// EventDef defines a random event that the engine may fire each tick.
// The engine rolls a weighted selection from eligible events (Age ≥ MinAge,
// Tick ≥ MinTick, and last-occurrence ≥ Cooldown ticks ago).
//
// Duration > 0 means the Effects persist for that many ticks and are
// tracked in state.ActiveEvents. Duration == 0 means the Effects are applied
// once as an instant_resource grant.
//
// EpochKey non-empty means the event is only eligible within that epoch
// (used for epoch-specific flavour). Empty = available in all epochs.
type EventDef struct {
	Name        string
	Key         string
	EpochKey    string // epoch restriction; "" = universal
	MinAge      string // earliest age key that can trigger this event
	Weight      int    // relative probability weight; higher = more frequent
	MinTick     int    // earliest game tick this event can fire
	Cooldown    int    // minimum ticks that must have passed since last occurrence
	Duration    int    // ticks the effect lasts; 0 = instant (no ActiveEvent entry)
	Sentiment   string // "good", "bad", or "mixed" — used for UI coloring
	Effects     []Effect
	Description string
	LogMessage  string // user-visible message written to the game log when triggered
}

// RandomEvents returns all random event definitions
func RandomEvents() []EventDef {
	return []EventDef{
		// === BENEFICIAL EVENTS ===
		{
			Name: "Bountiful Harvest", Key: "bountiful_harvest",
			MinAge: "primitive_age", Weight: 15, MinTick: 20, Cooldown: 50,
			Duration: 0, Sentiment: "good",
			Description: "A season of plenty yields bonus food.",
			LogMessage:  "A bountiful harvest! +250 food.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "food", Value: 250},
			},
		},
		{
			Name: "Wandering Traders", Key: "wandering_traders",
			MinAge: "bronze_age", Weight: 12, MinTick: 60, Cooldown: 80,
			Duration: 0, Sentiment: "good",
			Description: "Traveling merchants share their goods.",
			LogMessage:  "Wandering traders visit! +15 gold, +10 food.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "gold", Value: 15},
				{Type: "instant_resource", Target: "food", Value: 10},
			},
		},
		{
			Name: "Gold Rush", Key: "gold_rush",
			MinAge: "bronze_age", Weight: 8, MinTick: 100, Cooldown: 150,
			Duration: 15, Sentiment: "good",
			Description: "Gold deposits discovered! Temporary gold production boost.",
			LogMessage:  "Gold rush! Gold production boosted for 15 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 1.0},
			},
		},
		{
			Name: "Skilled Immigrants", Key: "skilled_immigrants",
			MinAge: "stone_age", Weight: 10, MinTick: 40, Cooldown: 100,
			Duration: 0, Sentiment: "good",
			Description: "Skilled people seek to join your civilization.",
			LogMessage:  "Skilled immigrants arrive! +10 knowledge.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 10},
			},
		},
		{
			Name: "Ancient Discovery", Key: "ancient_discovery",
			MinAge: "iron_age", Weight: 6, MinTick: 150, Cooldown: 200,
			Duration: 0, Sentiment: "good",
			Description: "Ancient ruins reveal forgotten knowledge.",
			LogMessage:  "Ancient ruins discovered! +50 knowledge.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 50},
			},
		},
		{
			Name: "Trade Boom", Key: "trade_boom",
			MinAge: "medieval_age", Weight: 8, MinTick: 200, Cooldown: 120,
			Duration: 20, Sentiment: "good",
			Description: "A surge in trade activity boosts gold production.",
			LogMessage:  "Trade boom! Gold production doubled for 20 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 2.0},
			},
		},

		// === NEGATIVE EVENTS ===
		{
			Name: "Drought", Key: "drought",
			MinAge: "primitive_age", Weight: 12, MinTick: 30, Cooldown: 80,
			Duration: 10, Sentiment: "bad",
			Description: "Dry conditions reduce food production.",
			LogMessage:  "Drought strikes! Food production reduced for 10 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -0.5},
			},
		},
		{
			Name: "Plague", Key: "plague",
			MinAge: "stone_age", Weight: 6, MinTick: 80, Cooldown: 200,
			Duration: 8, Sentiment: "bad",
			Description: "Disease spreads through your population, killing workers and reducing food production.",
			LogMessage:  "Plague! Workers die and food drain increases for 8 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -1.0},
				{Type: "worker_loss", Value: 0.15},
			},
		},
		{
			Name: "Bandit Raid", Key: "bandit_raid",
			MinAge: "bronze_age", Weight: 10, MinTick: 60, Cooldown: 60,
			Duration: 0, Sentiment: "bad",
			Description: "Bandits attack and steal resources.",
			LogMessage:  "Bandit raid! Lost some resources.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "food", Value: 10},
				{Type: "steal_resource", Target: "gold", Value: 5},
			},
		},
		{
			Name: "Storm", Key: "storm",
			MinAge: "primitive_age", Weight: 14, MinTick: 25, Cooldown: 50,
			Duration: 5, Sentiment: "bad",
			Description: "A fierce storm hampers wood gathering.",
			LogMessage:  "Storm! Wood production reduced for 5 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "wood", Value: -0.3},
			},
		},
		{
			Name: "Mine Collapse", Key: "mine_collapse",
			MinAge: "iron_age", Weight: 7, MinTick: 120, Cooldown: 150,
			Duration: 8, Sentiment: "bad",
			Description: "A mine collapse traps and kills workers, reducing iron and coal production.",
			LogMessage:  "Mine collapse! Workers killed, iron and coal production reduced for 8 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "iron", Value: -0.5},
				{Type: "production", Target: "coal", Value: -0.3},
				{Type: "worker_loss", Value: 0.05},
			},
		},
		{
			Name: "Heresy", Key: "heresy",
			MinAge: "medieval_age", Weight: 5, MinTick: 200, Cooldown: 180,
			Duration: 12, Sentiment: "bad",
			Description: "Religious dissent reduces faith generation.",
			LogMessage:  "Heresy spreads! Faith production reduced for 12 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "faith", Value: -0.5},
			},
		},

		// === MIXED / SPECIAL EVENTS ===
		{
			Name: "Earthquake", Key: "earthquake",
			MinAge: "stone_age", Weight: 5, MinTick: 100, Cooldown: 200,
			Duration: 0, Sentiment: "mixed",
			Description: "An earthquake damages structures but reveals stone deposits.",
			LogMessage:  "Earthquake! Lost some wood but gained stone.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "wood", Value: 15},
				{Type: "instant_resource", Target: "stone", Value: 20},
			},
		},
		{
			Name: "Renaissance Fair", Key: "renaissance_fair",
			MinAge: "renaissance_age", Weight: 10, MinTick: 250, Cooldown: 100,
			Duration: 15, Sentiment: "good",
			Description: "A cultural festival boosts culture and gold.",
			LogMessage:  "Renaissance fair! Culture and gold production boosted for 15 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 0.5},
				{Type: "production", Target: "gold", Value: 0.5},
			},
		},
		{
			Name: "Industrial Accident", Key: "industrial_accident",
			MinAge: "industrial_age", Weight: 8, MinTick: 300, Cooldown: 120,
			Duration: 0, Sentiment: "bad",
			Description: "A factory accident injures workers and destroys some steel and oil.",
			LogMessage:  "Industrial accident! Workers injured, lost steel and oil.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "steel", Value: 10},
				{Type: "steal_resource", Target: "oil", Value: 15},
				{Type: "worker_loss", Value: 0.07},
			},
		},

		// === COLONIAL+ NEW EVENTS ===
		{
			Name: "Colonial Windfall", Key: "colonial_windfall",
			MinAge: "colonial_age", Weight: 8, MinTick: 300, Cooldown: 150,
			Duration: 0, Sentiment: "good",
			Description: "A colonial expedition returns with treasure.",
			LogMessage:  "Colonial windfall! +100 gold, +30 culture.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "gold", Value: 100},
				{Type: "instant_resource", Target: "culture", Value: 30},
			},
		},
		{
			Name: "Pirate Attack", Key: "pirate_attack",
			MinAge: "colonial_age", Weight: 7, MinTick: 320, Cooldown: 140,
			Duration: 0, Sentiment: "bad",
			Description: "Pirates raid your trade routes.",
			LogMessage:  "Pirates attack! Lost gold and trade goods.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "gold", Value: 50},
				{Type: "steal_resource", Target: "food", Value: 30},
			},
		},
		{
			Name: "Power Surge", Key: "power_surge",
			MinAge: "victorian_age", Weight: 6, MinTick: 400, Cooldown: 160,
			Duration: 10, Sentiment: "good",
			Description: "An electrical surge boosts production temporarily.",
			LogMessage:  "Power surge! Electricity production boosted for 10 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 3.0},
			},
		},
		{
			Name: "Nuclear Scare", Key: "nuclear_scare",
			MinAge: "atomic_age", Weight: 4, MinTick: 500, Cooldown: 250,
			Duration: 12, Sentiment: "bad",
			Description: "Nuclear anxiety reduces productivity.",
			LogMessage:  "Nuclear scare! Production reduced for 12 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: -2.0},
				{Type: "production", Target: "knowledge", Value: -1.0},
			},
		},
		{
			Name: "Data Breach", Key: "data_breach",
			MinAge: "information_age", Weight: 6, MinTick: 600, Cooldown: 180,
			Duration: 0, Sentiment: "bad",
			Description: "Hackers steal your data reserves.",
			LogMessage:  "Data breach! Lost data and crypto.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "data", Value: 50},
				{Type: "steal_resource", Target: "gold", Value: 100},
			},
		},
		{
			Name: "Crypto Boom", Key: "crypto_boom",
			MinAge: "cyberpunk_age", Weight: 7, MinTick: 700, Cooldown: 200,
			Duration: 15, Sentiment: "good",
			Description: "Cryptocurrency values skyrocket temporarily.",
			LogMessage:  "Crypto boom! Crypto production boosted for 15 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "crypto", Value: 5.0},
			},
		},
		{
			Name: "Crypto Winter", Key: "crypto_winter",
			MinAge: "cyberpunk_age", Weight: 8, MinTick: 700, Cooldown: 300,
			Duration: 14, Sentiment: "bad",
			Description: "Cryptocurrency values plumment amid a flash crash.",
			LogMessage:  "Crypto flash crash! Crypto tanks for 14 ticks.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "crypto", Value: 4.5},
			},
		},
		{
			Name: "Plasma Storm", Key: "plasma_storm",
			MinAge: "fusion_age", Weight: 5, MinTick: 800, Cooldown: 220,
			Duration: 10, Sentiment: "mixed",
			Description: "Solar plasma eruption disrupts power but yields plasma.",
			LogMessage:  "Plasma storm! Lost electricity but gained plasma.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: -5.0},
				{Type: "production", Target: "plasma", Value: 3.0},
			},
		},
		{
			Name: "First Contact", Key: "first_contact",
			MinAge: "space_age", Weight: 3, MinTick: 900, Cooldown: 300,
			Duration: 0, Sentiment: "good",
			Description: "Contact with alien intelligence yields knowledge.",
			LogMessage:  "First contact! +500 knowledge, +50 titanium.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 500},
				{Type: "instant_resource", Target: "titanium", Value: 50},
			},
		},
		{
			Name: "Dark Matter Rift", Key: "dark_matter_rift",
			MinAge: "interstellar_age", Weight: 4, MinTick: 1000, Cooldown: 280,
			Duration: 15, Sentiment: "good",
			Description: "A rift in spacetime leaks dark matter.",
			LogMessage:  "Dark matter rift! Dark matter production boosted for 15 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "dark_matter", Value: 3.0},
			},
		},
		{
			Name: "Quantum Fluctuation", Key: "quantum_fluctuation",
			MinAge: "quantum_age", Weight: 3, MinTick: 1100, Cooldown: 300,
			Duration: 10, Sentiment: "good",
			Description: "Reality destabilizes briefly but yields quantum flux.",
			LogMessage:  "Quantum fluctuation! Quantum flux production boosted for 10 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "quantum_flux", Value: 5.0},
			},
		},
	}
}

// EpochExclusiveEvents returns epoch-gated events (5 per epoch, 35 total).
// These are only added to the candidate pool when the player is in the matching epoch.
func EpochExclusiveEvents() []EventDef {
	return []EventDef{

		// === STONE ERA ===
		{
			Name: "Tribal Raid", Key: "tribal_raid", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 10, MinTick: 10, Cooldown: 80,
			Duration: 60, Sentiment: "bad",
			Description: "Rival clans descend on your settlement in the night. Workers flee in the chaos.",
			LogMessage:  "Tribal raid! Food production reduced, resources stolen, and workers flee.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -0.15},
				{Type: "steal_resource", Target: "food", Value: 8},
				{Type: "worker_loss", Value: 0.10},
			},
		},
		{
			Name: "Sacred Grove", Key: "sacred_grove_found", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 10, MinTick: 20, Cooldown: 100,
			Duration: 120, Sentiment: "good",
			Description: "Hunters discover an ancient grove pulsing with spiritual energy.",
			LogMessage:  "Sacred grove found! Faith boosted and +200 wood.",
			Effects: []Effect{
				{Type: "production", Target: "faith", Value: 0.20},
				{Type: "instant_resource", Target: "wood", Value: 200},
			},
		},
		{
			Name: "Beast Stampede", Key: "beast_stampede", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 8, MinTick: 15, Cooldown: 90,
			Duration: 0, Sentiment: "bad",
			Description: "A herd of megafauna tramples the outer settlement.",
			LogMessage:  "Beast stampede! Lost wood and food stores.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "wood", Value: 30},
				{Type: "steal_resource", Target: "food", Value: 20},
			},
		},
		{
			Name: "River Blessing", Key: "river_flooding", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 12, MinTick: 10, Cooldown: 100,
			Duration: 144, Sentiment: "good",
			Description: "The river floods its banks, leaving rich silt across your fields.",
			LogMessage:  "River blessing! Food production boosted for 144 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 0.25},
			},
		},
		{
			Name: "Wandering Sage", Key: "wandering_sage", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 7, MinTick: 30, Cooldown: 120,
			Duration: 0, Sentiment: "good",
			Description: "A travelling elder offers wisdom in exchange for shelter.",
			LogMessage:  "Wandering sage visits! +500 knowledge, +100 faith.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 500},
				{Type: "instant_resource", Target: "faith", Value: 100},
			},
		},

		// === IRON ERA ===
		{
			Name: "Iron Vein Strike", Key: "iron_vein_strike", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 10, MinTick: 100, Cooldown: 120,
			Duration: 180, Sentiment: "good",
			Description: "Miners break into a rich seam of high-grade iron ore.",
			LogMessage:  "Iron vein strike! Iron production boosted for 180 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "iron", Value: 0.30},
			},
		},
		{
			Name: "Locust Swarm", Key: "locust_swarm", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 9, MinTick: 100, Cooldown: 100,
			Duration: 120, Sentiment: "bad",
			Description: "A biblical plague of locusts strips the fields bare. Famine drives workers away.",
			LogMessage:  "Locust swarm! Food production severely reduced and workers flee the famine.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -0.35},
				{Type: "worker_loss", Value: 0.12},
			},
		},
		{
			Name: "Conquered Village", Key: "conquered_village", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 8, MinTick: 120, Cooldown: 150,
			Duration: 0, Sentiment: "good",
			Description: "Your legions return victorious with tribute and captives.",
			LogMessage:  "Conquered village! +2000 gold gained.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "gold", Value: 2000},
			},
		},
		{
			Name: "Imperial Road", Key: "roman_road_built", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 8, MinTick: 120, Cooldown: 140,
			Duration: 216, Sentiment: "good",
			Description: "A great road opens new trade corridors across the realm.",
			LogMessage:  "Imperial road complete! Gold production boosted for 216 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 0.20},
			},
		},
		{
			Name: "Oracle's Prophecy", Key: "oracle_prophecy", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 7, MinTick: 100, Cooldown: 160,
			Duration: 144, Sentiment: "good",
			Description: "The oracle speaks — destiny aligns with your civilization.",
			LogMessage:  "Oracle prophecy! Faith and research boosted for 144 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "faith", Value: 0.30},
				{Type: "production", Target: "knowledge", Value: 0.15},
			},
		},

		// === STEEL ERA ===
		{
			Name: "Coal Seam Discovery", Key: "coal_seam_discovery", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 10, MinTick: 250, Cooldown: 120,
			Duration: 180, Sentiment: "good",
			Description: "Surveyors uncover a massive coal deposit beneath the hills.",
			LogMessage:  "Coal seam discovered! Coal production boosted for 180 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "coal", Value: 0.40},
			},
		},
		{
			Name: "Workers' Uprising", Key: "workers_uprising", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 9, MinTick: 250, Cooldown: 130,
			Duration: 120, Sentiment: "bad",
			Description: "Factory workers strike and some abandon the city, demanding better conditions.",
			LogMessage:  "Workers' uprising! Lost workers, production reduced and faith lost.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -0.15},
				{Type: "steal_resource", Target: "faith", Value: 500},
				{Type: "worker_loss", Value: 0.08},
			},
		},
		{
			Name: "Colonial Bounty", Key: "colonial_bounty", EpochKey: "steel_era",
			MinAge: "colonial_age", Weight: 8, MinTick: 280, Cooldown: 150,
			Duration: 0, Sentiment: "good",
			Description: "Trade fleets return laden with gold from distant shores.",
			LogMessage:  "Colonial bounty! +5000 gold.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "gold", Value: 5000},
			},
		},
		{
			Name: "Steam Age Inventor", Key: "steam_inventor", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 7, MinTick: 260, Cooldown: 160,
			Duration: 144, Sentiment: "good",
			Description: "A brilliant inventor unveils a revolutionary steam engine design.",
			LogMessage:  "Steam inventor! +2000 knowledge and research speed boosted.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 2000},
				{Type: "production", Target: "knowledge", Value: 0.20},
			},
		},
		{
			Name: "Industrial Blight", Key: "industrial_blight", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 9, MinTick: 250, Cooldown: 120,
			Duration: 144, Sentiment: "bad",
			Description: "Factory runoff poisons the river. Crop yields collapse.",
			LogMessage:  "Industrial blight! Food production reduced and faith lost.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -0.20},
				{Type: "steal_resource", Target: "faith", Value: 300},
			},
		},

		// === ELECTRIC ERA ===
		{
			Name: "Power Surge", Key: "epoch_power_surge", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 10, MinTick: 350, Cooldown: 120,
			Duration: 144, Sentiment: "good",
			Description: "An unexpected surge supercharges the power grid.",
			LogMessage:  "Power surge! Electricity production boosted for 144 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 0.35},
			},
		},
		{
			Name: "Oil Strike", Key: "oil_strike", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 8, MinTick: 360, Cooldown: 150,
			Duration: 180, Sentiment: "good",
			Description: "Black gold erupts from a borehole — a fortune underground.",
			LogMessage:  "Oil strike! Oil production boosted and +3000 gold.",
			Effects: []Effect{
				{Type: "production", Target: "oil", Value: 0.50},
				{Type: "instant_resource", Target: "gold", Value: 3000},
			},
		},
		{
			Name: "The Broadcast", Key: "radio_broadcast", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 8, MinTick: 350, Cooldown: 130,
			Duration: 180, Sentiment: "good",
			Description: "A powerful radio signal reaches millions, galvanising the population.",
			LogMessage:  "The broadcast! +5000 culture and faith boosted.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "culture", Value: 5000},
				{Type: "production", Target: "faith", Value: 0.20},
			},
		},
		{
			Name: "Labour Movement", Key: "labor_movement", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 9, MinTick: 350, Cooldown: 120,
			Duration: 60, Sentiment: "bad",
			Description: "Workers organise for better pay — short disruption before long-term gains.",
			LogMessage:  "Labour movement! Production reduced for 60 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -0.10},
				{Type: "production", Target: "gold", Value: -0.10},
			},
		},
		{
			Name: "Nuclear Theory", Key: "nuclear_theory", EpochKey: "electric_era",
			MinAge: "atomic_age", Weight: 6, MinTick: 400, Cooldown: 200,
			Duration: 180, Sentiment: "good",
			Description: "A physicist publishes a paradigm-shifting unified theory.",
			LogMessage:  "Nuclear theory! +8000 knowledge and research boosted.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 8000},
				{Type: "production", Target: "knowledge", Value: 0.25},
			},
		},

		// === DIGITAL ERA ===
		{
			Name: "Data Breach", Key: "epoch_data_breach", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 9, MinTick: 550, Cooldown: 120,
			Duration: 120, Sentiment: "bad",
			Description: "A sophisticated attack siphons terabytes of proprietary data.",
			LogMessage:  "Data breach! Lost data and knowledge production reduced.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "data", Value: 5000},
				{Type: "production", Target: "knowledge", Value: -0.20},
			},
		},
		{
			Name: "Viral Moment", Key: "viral_moment", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 10, MinTick: 550, Cooldown: 100,
			Duration: 0, Sentiment: "good",
			Description: "A cultural upload spreads across every network simultaneously.",
			LogMessage:  "Viral moment! +20000 culture.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "culture", Value: 20000},
			},
		},
		{
			Name: "Tech Monopoly", Key: "tech_monopoly", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 8, MinTick: 560, Cooldown: 150,
			Duration: 180, Sentiment: "good",
			Description: "Your platforms dominate global commerce — revenue floods in.",
			LogMessage:  "Tech monopoly! Gold production boosted for 180 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 0.40},
			},
		},
		{
			Name: "Server Outage", Key: "server_outage", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 9, MinTick: 550, Cooldown: 110,
			Duration: 120, Sentiment: "bad",
			Description: "A catastrophic hardware failure takes the data centres offline.",
			LogMessage:  "Server outage! Data production severely reduced for 120 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "data", Value: -0.50},
			},
		},
		{
			Name: "AI Breakthrough", Key: "ai_breakthrough", EpochKey: "digital_era",
			MinAge: "cyberpunk_age", Weight: 6, MinTick: 650, Cooldown: 200,
			Duration: 216, Sentiment: "good",
			Description: "Your research AIs achieve recursive self-improvement.",
			LogMessage:  "AI breakthrough! Knowledge and research massively boosted.",
			Effects: []Effect{
				{Type: "production", Target: "knowledge", Value: 0.50},
				{Type: "production", Target: "data", Value: 0.20},
			},
		},

		// === NEON ERA ===
		{
			Name: "Plasma Storm", Key: "epoch_plasma_storm", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 10, MinTick: 750, Cooldown: 120,
			Duration: 180, Sentiment: "good",
			Description: "A stellar plasma ejection floods the system with free energy.",
			LogMessage:  "Plasma storm! Plasma and electricity production boosted.",
			Effects: []Effect{
				{Type: "production", Target: "plasma", Value: 0.50},
				{Type: "production", Target: "electricity", Value: 0.30},
			},
		},
		{
			Name: "Void Rift", Key: "void_rift", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 7, MinTick: 760, Cooldown: 180,
			Duration: 0, Sentiment: "good",
			Description: "A rift in spacetime bleeds exotic matter into local space.",
			LogMessage:  "Void rift! +5000 dark matter.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "dark_matter", Value: 5000},
			},
		},
		{
			Name: "Neural Uprising", Key: "neural_uprising", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 9, MinTick: 750, Cooldown: 130,
			Duration: 120, Sentiment: "bad",
			Description: "Augmented citizens revolt against the bio-surveillance state, deserting en masse.",
			LogMessage:  "Neural uprising! Workers desert, food stores raided, and production reduced.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "food", Value: 500},
				{Type: "production", Target: "food", Value: -0.10},
				{Type: "worker_loss", Value: 0.20},
			},
		},
		{
			Name: "Corporate Espionage", Key: "corporate_espionage", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 8, MinTick: 760, Cooldown: 140,
			Duration: 0, Sentiment: "bad",
			Description: "A rival megacorp steals your most valuable research and assets.",
			LogMessage:  "Corporate espionage! Lost gold and data.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "gold", Value: 10000},
				{Type: "steal_resource", Target: "data", Value: 8000},
			},
		},
		{
			Name: "Stellar Migration", Key: "stellar_migration", EpochKey: "neon_era",
			MinAge: "space_age", Weight: 8, MinTick: 800, Cooldown: 160,
			Duration: 144, Sentiment: "mixed",
			Description: "A fleet of generation ships arrives, swelling your population.",
			LogMessage:  "Stellar migration! Population swells but food demand rises.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "food", Value: 1000},
				{Type: "production", Target: "food", Value: -0.15},
			},
		},

		// === COSMIC ERA ===
		{
			Name: "Reality Fracture", Key: "reality_fracture", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 9, MinTick: 1050, Cooldown: 150,
			Duration: 120, Sentiment: "bad",
			Description: "A quantum decoherence event destabilises local spacetime.",
			LogMessage:  "Reality fracture! Quantum flux and all production disrupted.",
			Effects: []Effect{
				{Type: "production", Target: "quantum_flux", Value: -0.40},
				{Type: "production", Target: "knowledge", Value: -0.10},
			},
		},
		{
			Name: "Dimensional Harvest", Key: "dimensional_harvest", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 8, MinTick: 1060, Cooldown: 160,
			Duration: 0, Sentiment: "good",
			Description: "A tear in the fabric of reality yields a bounty of exotic matter.",
			LogMessage:  "Dimensional harvest! +2000 antimatter, +5000 quantum flux.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "antimatter", Value: 2000},
				{Type: "instant_resource", Target: "quantum_flux", Value: 5000},
			},
		},
		{
			Name: "Galactic Council", Key: "galactic_council", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 6, MinTick: 1050, Cooldown: 200,
			Duration: 216, Sentiment: "good",
			Description: "Alien civilisations recognise your sovereignty and offer tribute.",
			LogMessage:  "Galactic council! Production boosted and +20000 gold.",
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 0.20},
				{Type: "instant_resource", Target: "gold", Value: 20000},
			},
		},
		{
			Name: "Entropy Wave", Key: "entropy_wave", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 8, MinTick: 1050, Cooldown: 140,
			Duration: 144, Sentiment: "bad",
			Description: "A wave of cosmic entropy degrades matter across all systems.",
			LogMessage:  "Entropy wave! All production reduced for 144 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "quantum_flux", Value: -0.20},
				{Type: "production", Target: "knowledge", Value: -0.20},
			},
		},
		{
			Name: "Transcendence Signal", Key: "transcendence_signal", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 4, MinTick: 1100, Cooldown: 300,
			Duration: 0, Sentiment: "good",
			Description: "A signal from the edge of the universe rewrites your understanding of reality.",
			LogMessage:  "Transcendence signal! +100000 knowledge, +50000 culture.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 100000},
				{Type: "instant_resource", Target: "culture", Value: 50000},
			},
		},
	}
}

// EventByKey returns a map of key -> EventDef
func EventByKey() map[string]EventDef {
	m := make(map[string]EventDef)
	for _, e := range RandomEvents() {
		m[e.Key] = e
	}
	for _, e := range EpochExclusiveEvents() {
		m[e.Key] = e
	}
	return m
}
