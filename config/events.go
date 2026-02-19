package config

// EventDef defines a random event that can occur during gameplay
type EventDef struct {
	Name        string
	Key         string
	MinAge      string // earliest age this can trigger
	Weight      int    // relative probability (higher = more common)
	MinTick     int    // earliest tick this can trigger
	Cooldown    int    // minimum ticks between occurrences
	Duration    int    // how many ticks the effect lasts (0 = instant)
	Sentiment   string // "good", "bad", or "mixed"
	Effects     []Effect
	Description string
	LogMessage  string // what shows in the game log
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
			LogMessage:  "A bountiful harvest! +25 food.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "food", Value: 25},
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
			Description: "Disease spreads through your population.",
			LogMessage:  "Plague! Food drain increased for 8 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -1.0},
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
			Description: "A mine collapse reduces iron and coal production.",
			LogMessage:  "Mine collapse! Iron and coal production reduced for 8 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "iron", Value: -0.5},
				{Type: "production", Target: "coal", Value: -0.3},
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
			Description: "A factory accident destroys some steel and oil.",
			LogMessage:  "Industrial accident! Lost steel and oil.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "steel", Value: 10},
				{Type: "steal_resource", Target: "oil", Value: 15},
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

// EventByKey returns a map of key -> EventDef
func EventByKey() map[string]EventDef {
	m := make(map[string]EventDef)
	for _, e := range RandomEvents() {
		m[e.Key] = e
	}
	return m
}
