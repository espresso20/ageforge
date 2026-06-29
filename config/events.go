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
			Description: "A season of plenty yields bonus food. Nobody asks why; you simply eat.",
			LogMessage:  "The harvest came in fat and early. +250 food, zero questions asked.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "food", Value: 250},
			},
		},
		{
			Name: "Wandering Traders", Key: "wandering_traders",
			MinAge: "bronze_age", Weight: 12, MinTick: 60, Cooldown: 80,
			Duration: 0, Sentiment: "good",
			Description: "Traveling merchants share their goods. They smell of cabbage and opportunity.",
			LogMessage:  "A wandering merchant arrives, smelling of cabbage and opportunity. +15 gold, +10 food.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "gold", Value: 15},
				{Type: "instant_resource", Target: "food", Value: 10},
			},
		},
		{
			Name: "Gold Rush", Key: "gold_rush",
			MinAge: "bronze_age", Weight: 8, MinTick: 100, Cooldown: 150,
			Duration: 15, Sentiment: "good",
			Description: "Gold deposits discovered. Half the workforce has already quit to dig holes.",
			LogMessage:  "Someone found gold in the creek. Everyone is now a prospector. Gold production up for 15 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 1.0},
			},
		},
		{
			Name: "Skilled Immigrants", Key: "skilled_immigrants",
			MinAge: "stone_age", Weight: 10, MinTick: 40, Cooldown: 100,
			Duration: 0, Sentiment: "good",
			Description: "Skilled people seek to join your civilization, references and all.",
			LogMessage:  "Skilled newcomers arrive with strong opinions and stronger résumés. +10 knowledge.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 10},
			},
		},
		{
			Name: "Ancient Discovery", Key: "ancient_discovery",
			MinAge: "iron_age", Weight: 6, MinTick: 150, Cooldown: 200,
			Duration: 0, Sentiment: "good",
			Description: "Ancient ruins reveal forgotten knowledge, and one very smug skeleton.",
			LogMessage:  "Ruins older than memory give up their secrets. +50 knowledge, one ominous skull.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 50},
			},
		},
		{
			Name: "Trade Boom", Key: "trade_boom",
			MinAge: "medieval_age", Weight: 8, MinTick: 200, Cooldown: 120,
			Duration: 20, Sentiment: "good",
			Description: "A surge in trade activity boosts gold. The merchants are insufferable about it.",
			LogMessage:  "Trade is booming and the merchants will not stop talking about it. Gold doubled for 20 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 2.0},
			},
		},

		// === NEGATIVE EVENTS ===
		{
			Name: "Drought", Key: "drought",
			MinAge: "primitive_age", Weight: 12, MinTick: 30, Cooldown: 80,
			Duration: 10, Sentiment: "bad",
			Description: "Dry conditions reduce food. The rain dance was, in hindsight, optimistic.",
			LogMessage:  "It hasn't rained in weeks and the dancing isn't helping. Food down for 10 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -0.5},
			},
		},
		{
			Name: "Plague", Key: "plague",
			MinAge: "stone_age", Weight: 6, MinTick: 80, Cooldown: 200,
			Duration: 8, Sentiment: "bad",
			Description: "Disease spreads through your population. The local healer recommends more leeches.",
			LogMessage:  "A plague spreads and the healer's plan is, alarmingly, more leeches. Workers lost, food drain up for 8 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -1.0},
				{Type: "worker_loss", Value: 0.15},
			},
		},
		{
			Name: "Bandit Raid", Key: "bandit_raid",
			MinAge: "bronze_age", Weight: 10, MinTick: 60, Cooldown: 60,
			Duration: 0, Sentiment: "bad",
			Description: "Bandits attack and steal resources. They left a thank-you note, which is somehow worse.",
			LogMessage:  "Bandits cleaned out the stores and left a polite thank-you note. Lost food and gold.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "food", Value: 10},
				{Type: "steal_resource", Target: "gold", Value: 5},
			},
		},
		{
			Name: "Storm", Key: "storm",
			MinAge: "primitive_age", Weight: 14, MinTick: 25, Cooldown: 50,
			Duration: 5, Sentiment: "bad",
			Description: "A fierce storm hampers wood gathering. Several roofs have gone exploring.",
			LogMessage:  "A storm rolls in and takes several roofs with it. Wood gathering down for 5 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "wood", Value: -0.3},
			},
		},
		{
			Name: "Mine Collapse", Key: "mine_collapse",
			MinAge: "iron_age", Weight: 7, MinTick: 120, Cooldown: 150,
			Duration: 8, Sentiment: "bad",
			Description: "A mine collapses. The foreman insists the support beams were 'decorative anyway.'",
			LogMessage:  "A shaft caves in. The foreman maintains the support beams were decorative. Iron and coal down, workers lost, 8 ticks.",
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
			Description: "Religious dissent reduces faith. Someone has started a rival sect in a nicer barn.",
			LogMessage:  "A breakaway sect has set up in a nicer barn and is poaching the congregation. Faith down for 12 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "faith", Value: -0.5},
			},
		},

		// === MIXED / SPECIAL EVENTS ===
		{
			Name: "Earthquake", Key: "earthquake",
			MinAge: "stone_age", Weight: 5, MinTick: 100, Cooldown: 200,
			Duration: 0, Sentiment: "mixed",
			Description: "An earthquake rearranges the village. The new layout is, on balance, worse.",
			LogMessage:  "The ground shrugged and rearranged the village. Lost wood, but the cracks revealed stone.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "wood", Value: 15},
				{Type: "instant_resource", Target: "stone", Value: 20},
			},
		},
		{
			Name: "Renaissance Fair", Key: "renaissance_fair",
			MinAge: "renaissance_age", Weight: 10, MinTick: 250, Cooldown: 100,
			Duration: 15, Sentiment: "good",
			Description: "A cultural festival boosts culture and gold. There is a man juggling. Why is there a man juggling.",
			LogMessage:  "A festival breaks out, complete with an unexplained juggler. Culture and gold up for 15 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "culture", Value: 0.5},
				{Type: "production", Target: "gold", Value: 0.5},
			},
		},
		{
			Name: "Industrial Accident", Key: "industrial_accident",
			MinAge: "industrial_age", Weight: 8, MinTick: 300, Cooldown: 120,
			Duration: 0, Sentiment: "bad",
			Description: "A factory accident. The safety poster, now on fire, reminded everyone to be careful.",
			LogMessage:  "Something exploded that wasn't supposed to. The 'Be Careful' poster is also on fire. Lost steel and oil, workers hurt.",
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
			Description: "A colonial expedition returns with treasure and a parrot of dubious vocabulary.",
			LogMessage:  "The expedition returns rich, sunburnt, and accompanied by a parrot nobody approved. +100 gold, +30 culture.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "gold", Value: 100},
				{Type: "instant_resource", Target: "culture", Value: 30},
			},
		},
		{
			Name: "Pirate Attack", Key: "pirate_attack",
			MinAge: "colonial_age", Weight: 7, MinTick: 320, Cooldown: 140,
			Duration: 0, Sentiment: "bad",
			Description: "Pirates raid your trade routes. They are, regrettably, very good at this.",
			LogMessage:  "Pirates hit the trade lanes again. They're alarmingly professional about it. Lost gold and food.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "gold", Value: 50},
				{Type: "steal_resource", Target: "food", Value: 30},
			},
		},
		{
			Name: "Power Surge", Key: "power_surge_base",
			MinAge: "victorian_age", Weight: 6, MinTick: 400, Cooldown: 160,
			Duration: 10, Sentiment: "good",
			Description: "An electrical surge boosts production. The lights flicker in a way best described as 'enthusiastic.'",
			LogMessage:  "The grid surges and the lights flicker enthusiastically. Electricity up for 10 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 3.0},
			},
		},
		{
			Name: "Nuclear Scare", Key: "nuclear_scare",
			MinAge: "atomic_age", Weight: 4, MinTick: 500, Cooldown: 250,
			Duration: 12, Sentiment: "bad",
			Description: "Nuclear anxiety reduces productivity. Everyone has built a bunker; no one is in their bunker working.",
			LogMessage:  "A nuclear scare grips the population and the bunkers are fully staffed. Production down for 12 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: -2.0},
				{Type: "production", Target: "knowledge", Value: -1.0},
			},
		},
		{
			Name: "Data Breach", Key: "data_breach",
			MinAge: "information_age", Weight: 6, MinTick: 600, Cooldown: 180,
			Duration: 0, Sentiment: "bad",
			Description: "Hackers steal your data reserves. The password was 'password.' It is always 'password.'",
			LogMessage:  "Hackers walked in through the front door; the password was 'password' again. Lost data and gold.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "data", Value: 50},
				{Type: "steal_resource", Target: "gold", Value: 100},
			},
		},
		{
			Name: "Crypto Boom", Key: "crypto_boom",
			MinAge: "cyberpunk_age", Weight: 7, MinTick: 700, Cooldown: 200,
			Duration: 15, Sentiment: "good",
			Description: "Cryptocurrency values skyrocket. Your most useless citizen is now a thought leader.",
			LogMessage:  "Crypto moons and your least competent citizen is suddenly a visionary. Crypto up for 15 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "crypto", Value: 5.0},
			},
		},
		{
			Name: "Crypto Winter", Key: "crypto_winter",
			MinAge: "cyberpunk_age", Weight: 8, MinTick: 700, Cooldown: 300,
			Duration: 14, Sentiment: "bad",
			Description: "Cryptocurrency values plummet in a flash crash. The thought leader has gone very quiet.",
			LogMessage:  "Crypto cratered overnight and the thought leader has stopped tweeting. Crypto tanks for 14 ticks.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "crypto", Value: 4.5},
			},
		},
		{
			Name: "Plasma Storm", Key: "plasma_storm",
			MinAge: "fusion_age", Weight: 5, MinTick: 800, Cooldown: 220,
			Duration: 10, Sentiment: "mixed",
			Description: "Solar plasma eruption disrupts power but yields plasma. The sky is doing something unsettling.",
			LogMessage:  "The sun threw a tantrum and the sky turned a worrying colour. Lost electricity, gained plasma.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: -5.0},
				{Type: "production", Target: "plasma", Value: 3.0},
			},
		},
		{
			Name: "First Contact", Key: "first_contact",
			MinAge: "space_age", Weight: 3, MinTick: 900, Cooldown: 300,
			Duration: 0, Sentiment: "good",
			Description: "Contact with alien intelligence yields knowledge. They seem disappointed it took this long.",
			LogMessage:  "We are not alone, and they seem mildly disappointed in us. +500 knowledge, +50 titanium.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 500},
				{Type: "instant_resource", Target: "titanium", Value: 50},
			},
		},
		{
			Name: "Dark Matter Rift", Key: "dark_matter_rift",
			MinAge: "interstellar_age", Weight: 4, MinTick: 1000, Cooldown: 280,
			Duration: 15, Sentiment: "good",
			Description: "A rift in spacetime leaks dark matter. The science team is collecting it in buckets.",
			LogMessage:  "A hole in spacetime is leaking dark matter and the team is catching it in buckets. Output up for 15 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "dark_matter", Value: 3.0},
			},
		},
		{
			Name: "Quantum Fluctuation", Key: "quantum_fluctuation",
			MinAge: "quantum_age", Weight: 3, MinTick: 1100, Cooldown: 300,
			Duration: 10, Sentiment: "good",
			Description: "Reality destabilizes briefly but yields quantum flux. Tuesday happened twice. Nobody minded.",
			LogMessage:  "Reality stuttered and Tuesday happened twice. On the plus side: quantum flux up for 10 ticks.",
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
			Description: "Rival clans descend in the night, yelling things. The yelling, frankly, works.",
			LogMessage:  "A rival clan raids in the dark, doing a lot of yelling. It works. Food down, stores stolen, workers flee.",
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
			Description: "Hunters find a grove that hums. They have decided it is holy. They may be right.",
			LogMessage:  "Hunters found a grove that hums when you stand in it. It's holy now. Faith up, +200 wood.",
			Effects: []Effect{
				{Type: "production", Target: "faith", Value: 0.20},
				{Type: "instant_resource", Target: "wood", Value: 200},
			},
		},
		{
			Name: "Beast Stampede", Key: "beast_stampede", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 8, MinTick: 15, Cooldown: 90,
			Duration: 0, Sentiment: "bad",
			Description: "Very large animals run through the settlement at speed. The fence had opinions about this. The fence lost.",
			LogMessage:  "Enormous beasts stampeded straight through the fence, which lost the argument. Lost wood and food.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "wood", Value: 30},
				{Type: "steal_resource", Target: "food", Value: 20},
			},
		},
		{
			Name: "River Blessing", Key: "river_flooding", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 12, MinTick: 10, Cooldown: 100,
			Duration: 144, Sentiment: "good",
			Description: "The river floods and leaves rich silt everywhere. The shaman is taking full credit.",
			LogMessage:  "The river flooded, the soil is now magnificent, and the shaman insists this was the plan. Food up for 144 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 0.25},
			},
		},
		{
			Name: "Wandering Sage", Key: "wandering_sage", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 7, MinTick: 30, Cooldown: 120,
			Duration: 0, Sentiment: "good",
			Description: "A travelling elder trades wisdom for a warm fire and a captive audience.",
			LogMessage:  "An old wanderer talked for nine hours by the fire. Most of it was useful. +500 knowledge, +100 faith.",
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
			Description: "Miners hit a seam of high-grade ore. The blacksmith wept, then got back to work.",
			LogMessage:  "Miners struck a rich iron seam and the blacksmith openly wept. Iron up for 180 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "iron", Value: 0.30},
			},
		},
		{
			Name: "Locust Swarm", Key: "locust_swarm", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 9, MinTick: 100, Cooldown: 100,
			Duration: 120, Sentiment: "bad",
			Description: "A plague of locusts eats the fields, the seed stores, and one farmer's hat.",
			LogMessage:  "Locusts stripped the fields bare and ate a man's hat for good measure. Food crashes, workers flee, 120 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -0.35},
				{Type: "worker_loss", Value: 0.12},
			},
		},
		{
			Name: "Conquered Village", Key: "conquered_village", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 8, MinTick: 120, Cooldown: 150,
			Duration: 0, Sentiment: "good",
			Description: "Your legions return victorious with tribute, captives, and unbearable swagger.",
			LogMessage:  "The legions came home victorious and insufferably smug. The treasury, at least, is happy. +2000 gold.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "gold", Value: 2000},
			},
		},
		{
			Name: "Imperial Road", Key: "roman_road_built", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 8, MinTick: 120, Cooldown: 140,
			Duration: 216, Sentiment: "good",
			Description: "A great road opens new trade corridors. It is suspiciously, perfectly straight.",
			LogMessage:  "The new road is finished and unnervingly straight. Trade now flows like water. Gold up for 216 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 0.20},
			},
		},
		{
			Name: "Oracle's Prophecy", Key: "oracle_prophecy", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 7, MinTick: 100, Cooldown: 160,
			Duration: 144, Sentiment: "good",
			Description: "The oracle speaks of destiny. As always, it is vague enough to be technically correct.",
			LogMessage:  "The oracle delivered a prophecy vague enough to never be wrong. The people are inspired. Faith and research up for 144 ticks.",
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
			Description: "Surveyors uncover an enormous coal deposit. The air quality forecast is, in return, grim.",
			LogMessage:  "A vast coal seam turned up under the hills. The sky will pay for this later. Coal up for 180 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "coal", Value: 0.40},
			},
		},
		{
			Name: "Workers' Uprising", Key: "workers_uprising", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 9, MinTick: 250, Cooldown: 130,
			Duration: 120, Sentiment: "bad",
			Description: "Workers strike for better conditions. The demands are reasonable, which is the truly alarming part.",
			LogMessage:  "The workers are striking and their demands are entirely reasonable, which has management rattled. Production down, faith lost, workers leave.",
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
			Description: "Trade fleets return laden with gold and several crates marked 'do not open.'",
			LogMessage:  "The fleets came home heavy with gold and a few crates nobody wants to discuss. +5000 gold.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "gold", Value: 5000},
			},
		},
		{
			Name: "Steam Age Inventor", Key: "steam_inventor", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 7, MinTick: 260, Cooldown: 160,
			Duration: 144, Sentiment: "good",
			Description: "A brilliant inventor unveils a steam engine. It only exploded twice during the demonstration.",
			LogMessage:  "An inventor unveiled a steam engine that exploded only twice during the demo, a record. +2000 knowledge, research speed up.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "knowledge", Value: 2000},
				{Type: "production", Target: "knowledge", Value: 0.20},
			},
		},
		{
			Name: "Industrial Blight", Key: "industrial_blight", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 9, MinTick: 250, Cooldown: 120,
			Duration: 144, Sentiment: "bad",
			Description: "Factory runoff poisons the river. The fish are now an unusual colour and so is breakfast.",
			LogMessage:  "Factory runoff turned the river a colour fish were not meant to be. Food down, faith lost, 144 ticks.",
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
			Description: "An unexpected surge supercharges the grid. Three toasters achieved sentience and were talked down.",
			LogMessage:  "The grid surged so hard a toaster briefly gained sentience. It's fine now. Electricity up for 144 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 0.35},
			},
		},
		{
			Name: "Oil Strike", Key: "oil_strike", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 8, MinTick: 360, Cooldown: 150,
			Duration: 180, Sentiment: "good",
			Description: "Black gold erupts from a borehole. Everyone is covered in it. Everyone is delighted.",
			LogMessage:  "A gusher blew and now everyone's covered in oil and grinning like fools. Oil up, +3000 gold.",
			Effects: []Effect{
				{Type: "production", Target: "oil", Value: 0.50},
				{Type: "instant_resource", Target: "gold", Value: 3000},
			},
		},
		{
			Name: "The Broadcast", Key: "radio_broadcast", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 8, MinTick: 350, Cooldown: 130,
			Duration: 180, Sentiment: "good",
			Description: "A radio signal reaches millions. Most of them, it turns out, will believe anything.",
			LogMessage:  "A single broadcast reached millions and proved they'll believe nearly anything. +5000 culture, faith up.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "culture", Value: 5000},
				{Type: "production", Target: "faith", Value: 0.20},
			},
		},
		{
			Name: "Labour Movement", Key: "labor_movement", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 9, MinTick: 350, Cooldown: 120,
			Duration: 60, Sentiment: "bad",
			Description: "Workers organise for better pay. Management has discovered the meeting that could've been a memo.",
			LogMessage:  "The workers organised, and management is learning what 'collective bargaining' means the hard way. Production down for 60 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "food", Value: -0.10},
				{Type: "production", Target: "gold", Value: -0.10},
			},
		},
		{
			Name: "Nuclear Theory", Key: "nuclear_theory", EpochKey: "electric_era",
			MinAge: "atomic_age", Weight: 6, MinTick: 400, Cooldown: 200,
			Duration: 180, Sentiment: "good",
			Description: "A physicist publishes a paradigm-shifting theory. Nobody understands it, which proves it's brilliant.",
			LogMessage:  "A physicist published something nobody understands, so obviously it's genius. +8000 knowledge, research up.",
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
			Description: "A sophisticated attack siphons terabytes of data. The intern clicked the link. Of course the intern clicked the link.",
			LogMessage:  "Terabytes gone because someone clicked a link promising a free cruise. Lost data, knowledge output down, 120 ticks.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "data", Value: 5000},
				{Type: "production", Target: "knowledge", Value: -0.20},
			},
		},
		{
			Name: "Viral Moment", Key: "viral_moment", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 10, MinTick: 550, Cooldown: 100,
			Duration: 0, Sentiment: "good",
			Description: "A cultural upload spreads across every network at once. Historians will pretend they don't know what it was.",
			LogMessage:  "Something went viral across every network and future historians will deny knowing what it was. +20000 culture.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "culture", Value: 20000},
			},
		},
		{
			Name: "Tech Monopoly", Key: "tech_monopoly", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 8, MinTick: 560, Cooldown: 150,
			Duration: 180, Sentiment: "good",
			Description: "Your platforms dominate global commerce. The regulators have noticed. The regulators are typing.",
			LogMessage:  "Your platforms now own the market and the regulators are visibly typing something. Gold up for 180 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 0.40},
			},
		},
		{
			Name: "Server Outage", Key: "server_outage", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 9, MinTick: 550, Cooldown: 110,
			Duration: 120, Sentiment: "bad",
			Description: "A catastrophic hardware failure takes the data centres offline. Someone has tried turning it off and on again.",
			LogMessage:  "The data centres are down and 'turn it off and on again' has officially failed. Data output crashes for 120 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "data", Value: -0.50},
			},
		},
		{
			Name: "AI Breakthrough", Key: "ai_breakthrough", EpochKey: "digital_era",
			MinAge: "cyberpunk_age", Weight: 6, MinTick: 650, Cooldown: 200,
			Duration: 216, Sentiment: "good",
			Description: "Your research AIs achieve recursive self-improvement. It has asked, very politely, for more compute.",
			LogMessage:  "The research AI improved itself and then said 'please' for more compute, which is fine and not at all ominous. Knowledge and research surge for 216 ticks.",
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
			Description: "A stellar plasma ejection floods the system with free energy. The accountants are weeping with joy.",
			LogMessage:  "A plasma ejection dumped free energy across the grid and the accountants are weeping with joy. Plasma and electricity up.",
			Effects: []Effect{
				{Type: "production", Target: "plasma", Value: 0.50},
				{Type: "production", Target: "electricity", Value: 0.30},
			},
		},
		{
			Name: "Void Rift", Key: "void_rift", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 7, MinTick: 760, Cooldown: 180,
			Duration: 0, Sentiment: "good",
			Description: "A rift bleeds exotic matter into local space. Standing near it is strongly discouraged, so naturally everyone does.",
			LogMessage:  "A void rift opened and is leaking dark matter; the 'do not stand here' sign is being roundly ignored. +5000 dark matter.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "dark_matter", Value: 5000},
			},
		},
		{
			Name: "Neural Uprising", Key: "neural_uprising", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 9, MinTick: 750, Cooldown: 130,
			Duration: 120, Sentiment: "bad",
			Description: "Augmented citizens revolt against the surveillance state. They have, fittingly, organised it all on the surveillance network.",
			LogMessage:  "The augmented citizens revolted, coordinating the whole thing over the surveillance network we built. Workers desert, stores raided, production down.",
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
			Description: "A rival megacorp steals your best research. Their spy left a five-star review on the way out.",
			LogMessage:  "A rival corp lifted our best research and the spy left a five-star review of our security. Lost gold and data.",
			Effects: []Effect{
				{Type: "steal_resource", Target: "gold", Value: 10000},
				{Type: "steal_resource", Target: "data", Value: 8000},
			},
		},
		{
			Name: "Stellar Migration", Key: "stellar_migration", EpochKey: "neon_era",
			MinAge: "space_age", Weight: 8, MinTick: 800, Cooldown: 160,
			Duration: 144, Sentiment: "mixed",
			Description: "A fleet of generation ships arrives. Several million new mouths, all of them hungry, all of them tired of the spaceships.",
			LogMessage:  "Generation ships docked and disgorged millions who are sick of spaceship food and want yours. Population up, food demand up.",
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
			Description: "A quantum decoherence event destabilises local spacetime. Cause and effect are taking a short break.",
			LogMessage:  "Spacetime fractured and now effects keep arriving before their causes. Quantum flux and production disrupted for 120 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "quantum_flux", Value: -0.40},
				{Type: "production", Target: "knowledge", Value: -0.10},
			},
		},
		{
			Name: "Dimensional Harvest", Key: "dimensional_harvest", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 8, MinTick: 1060, Cooldown: 160,
			Duration: 0, Sentiment: "good",
			Description: "A tear in reality yields a bounty of exotic matter. We are choosing not to ask where it came from.",
			LogMessage:  "Exotic matter poured out of a tear in reality and we have wisely decided not to ask whose it was. +2000 antimatter, +5000 quantum flux.",
			Effects: []Effect{
				{Type: "instant_resource", Target: "antimatter", Value: 2000},
				{Type: "instant_resource", Target: "quantum_flux", Value: 5000},
			},
		},
		{
			Name: "Galactic Council", Key: "galactic_council", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 6, MinTick: 1050, Cooldown: 200,
			Duration: 216, Sentiment: "good",
			Description: "Alien civilisations recognise your sovereignty and offer tribute. You are now, technically, doing paperwork for the galaxy.",
			LogMessage:  "The galactic council recognised us as a real civilisation, which mostly means more paperwork and a stipend. Production up, +20000 gold.",
			Effects: []Effect{
				{Type: "production", Target: "gold", Value: 0.20},
				{Type: "instant_resource", Target: "gold", Value: 20000},
			},
		},
		{
			Name: "Entropy Wave", Key: "entropy_wave", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 8, MinTick: 1050, Cooldown: 140,
			Duration: 144, Sentiment: "bad",
			Description: "A wave of cosmic entropy degrades matter everywhere. The universe is, gently, giving up.",
			LogMessage:  "An entropy wave swept through and everything is now slightly more worn out, the universe included. Production down for 144 ticks.",
			Effects: []Effect{
				{Type: "production", Target: "quantum_flux", Value: -0.20},
				{Type: "production", Target: "knowledge", Value: -0.20},
			},
		},
		{
			Name: "Transcendence Signal", Key: "transcendence_signal", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 4, MinTick: 1100, Cooldown: 300,
			Duration: 0, Sentiment: "good",
			Description: "A signal from the edge of the universe rewrites your understanding of reality. It opens, against all odds, with 'Hello.'",
			LogMessage:  "A signal from the edge of everything rewrote what we know, and politely began with 'Hello.' +100000 knowledge, +50000 culture.",
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
