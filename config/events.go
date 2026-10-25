package config

// EventDef defines a random event that the engine may fire each tick.
// The engine rolls a weighted selection from eligible events (Age ≥ MinAge,
// Tick ≥ MinTick, and last-occurrence ≥ Cooldown ticks ago).
//
// Duration > 0 means the Effects persist for that many ticks and are
// tracked in state.ActiveEvents. Duration == 0 means the Effects are applied
// once.
//
// EpochKey non-empty means the event is only eligible within that epoch
// (used for epoch-specific flavour). Empty = available in all epochs.
type EventDef struct {
	Name      string
	Key       string
	EpochKey  string // epoch restriction; "" = universal
	MinAge    string // earliest age key that can trigger this event
	Weight    int    // relative probability weight; higher = more frequent
	MinTick   int    // earliest game tick this event can fire
	Cooldown  int    // minimum ticks that must have passed since last occurrence
	Duration  int    // ticks the effect lasts; 0 = instant (no ActiveEvent entry)
	Sentiment string // "good", "bad", or "mixed" — used for UI coloring
	// Effects are sizes, not amounts (event_size.go): EventGain is minutes
	// of the town's income, EventLoss a share of what it holds, EventRate a
	// share of its income per tick while the event lasts. "worker_loss" is
	// a share of the workers. The engine turns each into an amount for the
	// town the event happens to when it fires.
	Effects []Effect
	// LogMessage is the event's own sentence in the game log. It carries no
	// numbers: the engine writes what the event did after it, from the
	// amounts it applied (what was gained and lost, each rate and how long
	// it lasts in the age the event fired in).
	LogMessage string
	// Raid marks an attack by outsiders (bandits, pirates, rival clans, beasts,
	// thieves and spies). The army's garrison blunts a Raid event's
	// losses of stock and of workers (see config/defense.go). Disasters
	// and unrest (earthquakes, plague, uprisings) are not raids: soldiers do not
	// stop those.
	Raid bool
}

// RandomEvents returns all random event definitions
func RandomEvents() []EventDef {
	return []EventDef{
		// === BENEFICIAL EVENTS ===
		{
			Name: "Bountiful Harvest", Key: "bountiful_harvest",
			MinAge: "primitive_age", Weight: 15, MinTick: 20, Cooldown: 50,
			Duration: 0, Sentiment: "good",
			LogMessage: "The harvest came in fat and early, zero questions asked.",
			Effects: []Effect{
				{Type: EventGain, Target: "food", Value: 5},
			},
		},
		{
			Name: "Wandering Traders", Key: "wandering_traders",
			MinAge: "bronze_age", Weight: 12, MinTick: 60, Cooldown: 80,
			Duration: 0, Sentiment: "good",
			LogMessage: "Wandering traders arrive, smelling of cabbage and opportunity.",
			Effects: []Effect{
				{Type: EventGain, Target: "gold", Value: 5},
				{Type: EventGain, Target: "food", Value: 3},
			},
		},
		{
			Name: "Gold Rush", Key: "gold_rush",
			MinAge: "bronze_age", Weight: 8, MinTick: 100, Cooldown: 150,
			Duration: 15, Sentiment: "good",
			LogMessage: "Someone found gold in the creek and everyone is now a prospector.",
			Effects: []Effect{
				{Type: EventRate, Target: "gold", Value: 1},
			},
		},
		{
			Name: "Skilled Immigrants", Key: "skilled_immigrants",
			MinAge: "stone_age", Weight: 10, MinTick: 40, Cooldown: 100,
			Duration: 0, Sentiment: "good",
			LogMessage: "Skilled newcomers arrive with strong opinions and stronger résumés.",
			Effects: []Effect{
				{Type: EventGain, Target: ResearchTarget, Value: 5},
			},
		},
		{
			Name: "Ancient Discovery", Key: "ancient_discovery",
			MinAge: "iron_age", Weight: 6, MinTick: 150, Cooldown: 200,
			Duration: 0, Sentiment: "good",
			LogMessage: "Ruins older than memory give up their secrets, and one ominous skull.",
			Effects: []Effect{
				{Type: EventGain, Target: ResearchTarget, Value: 10},
			},
		},
		{
			Name: "Trade Boom", Key: "trade_boom",
			MinAge: "medieval_age", Weight: 8, MinTick: 200, Cooldown: 120,
			Duration: 20, Sentiment: "good",
			LogMessage: "Trade is booming and the merchants will not stop talking about it.",
			Effects: []Effect{
				{Type: EventRate, Target: "gold", Value: 1},
			},
		},

		// === NEGATIVE EVENTS ===
		{
			Name: "Drought", Key: "drought",
			MinAge: "primitive_age", Weight: 12, MinTick: 30, Cooldown: 80,
			Duration: 10, Sentiment: "bad",
			LogMessage: "It hasn't rained in weeks and the dancing isn't helping.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: -0.5},
			},
		},
		{
			Name: "Plague", Key: "plague",
			MinAge: "stone_age", Weight: 6, MinTick: 80, Cooldown: 200,
			Duration: 8, Sentiment: "bad",
			LogMessage: "A plague spreads and the healer's plan is, alarmingly, more leeches.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: -0.5},
				{Type: "worker_loss", Value: 0.15},
			},
		},
		{
			Name: "Bandit Raid", Key: "bandit_raid",
			MinAge: "bronze_age", Weight: 10, MinTick: 60, Cooldown: 60,
			Duration: 0, Raid: true, Sentiment: "bad",
			LogMessage: "Bandits cleaned out the stores and left a polite thank-you note.",
			Effects: []Effect{
				{Type: EventLoss, Target: "food", Value: 0.06},
				{Type: EventLoss, Target: "gold", Value: 0.06},
			},
		},
		{
			Name: "Storm", Key: "storm",
			MinAge: "primitive_age", Weight: 14, MinTick: 25, Cooldown: 50,
			Duration: 5, Sentiment: "bad",
			LogMessage: "A storm rolls in and takes several roofs with it.",
			Effects: []Effect{
				{Type: EventRate, Target: "wood", Value: -0.5},
			},
		},
		{
			Name: "Mine Collapse", Key: "mine_collapse",
			MinAge: "iron_age", Weight: 7, MinTick: 120, Cooldown: 150,
			Duration: 8, Sentiment: "bad",
			// Iron only: the event fires from the Iron Age and coal is locked
			// until the Renaissance Age, so its old coal penalty did nothing
			// for three ages while the log claimed it.
			LogMessage: "A shaft caves in. The foreman maintains the support beams were decorative.",
			Effects: []Effect{
				{Type: EventRate, Target: "iron", Value: -0.5},
				{Type: "worker_loss", Value: 0.05},
			},
		},
		{
			Name: "Heresy", Key: "heresy",
			MinAge: "medieval_age", Weight: 5, MinTick: 200, Cooldown: 180,
			Duration: 12, Sentiment: "bad",
			LogMessage: "A breakaway sect has set up in a nicer barn and is poaching the congregation.",
			Effects: []Effect{
				{Type: EventRate, Target: "faith", Value: -0.5},
			},
		},

		// === MIXED / SPECIAL EVENTS ===
		{
			Name: "Earthquake", Key: "earthquake",
			MinAge: "stone_age", Weight: 5, MinTick: 100, Cooldown: 200,
			Duration: 0, Sentiment: "mixed",
			LogMessage: "The ground shrugged and rearranged the village, and the cracks turned up stone.",
			Effects: []Effect{
				{Type: EventLoss, Target: "wood", Value: 0.05},
				{Type: EventGain, Target: "stone", Value: 5},
			},
		},
		{
			Name: "Renaissance Fair", Key: "renaissance_fair",
			MinAge: "renaissance_age", Weight: 10, MinTick: 250, Cooldown: 100,
			Duration: 15, Sentiment: "good",
			LogMessage: "A festival breaks out, complete with an unexplained juggler.",
			Effects: []Effect{
				{Type: EventRate, Target: "culture", Value: 0.5},
				{Type: EventRate, Target: "gold", Value: 0.5},
			},
		},
		{
			Name: "Industrial Accident", Key: "industrial_accident",
			MinAge: "industrial_age", Weight: 8, MinTick: 300, Cooldown: 120,
			Duration: 0, Sentiment: "bad",
			LogMessage: "Something exploded that wasn't supposed to, and the 'Be Careful' poster is also on fire.",
			Effects: []Effect{
				{Type: EventLoss, Target: "steel", Value: 0.06},
				{Type: EventLoss, Target: "oil", Value: 0.06},
				{Type: "worker_loss", Value: 0.07},
			},
		},

		// === COLONIAL+ NEW EVENTS ===
		{
			Name: "Colonial Windfall", Key: "colonial_windfall",
			MinAge: "colonial_age", Weight: 8, MinTick: 300, Cooldown: 150,
			Duration: 0, Sentiment: "good",
			LogMessage: "The expedition returns rich, sunburnt, and accompanied by a parrot nobody approved.",
			Effects: []Effect{
				{Type: EventGain, Target: "gold", Value: 8},
				{Type: EventGain, Target: "culture", Value: 5},
			},
		},
		{
			Name: "Pirate Attack", Key: "pirate_attack",
			MinAge: "colonial_age", Weight: 7, MinTick: 320, Cooldown: 140,
			Duration: 0, Raid: true, Sentiment: "bad",
			LogMessage: "Pirates hit the trade lanes again. They're alarmingly professional about it.",
			Effects: []Effect{
				{Type: EventLoss, Target: "gold", Value: 0.08},
				{Type: EventLoss, Target: "food", Value: 0.05},
			},
		},
		{
			Name: "Power Surge", Key: "power_surge_base",
			MinAge: "victorian_age", Weight: 6, MinTick: 400, Cooldown: 160,
			Duration: 10, Sentiment: "good",
			LogMessage: "The grid surges and the lights flicker enthusiastically.",
			Effects: []Effect{
				{Type: EventRate, Target: "electricity", Value: 1},
			},
		},
		{
			Name: "Nuclear Scare", Key: "nuclear_scare",
			MinAge: "atomic_age", Weight: 4, MinTick: 500, Cooldown: 250,
			Duration: 12, Sentiment: "bad",
			LogMessage: "A nuclear scare grips the population and the bunkers are fully staffed.",
			Effects: []Effect{
				{Type: EventRate, Target: "electricity", Value: -0.4},
				{Type: EventRate, Target: ResearchTarget, Value: -0.3},
			},
		},
		{
			Name: "Data Breach", Key: "data_breach",
			MinAge: "information_age", Weight: 6, MinTick: 600, Cooldown: 180,
			Duration: 0, Raid: true, Sentiment: "bad",
			LogMessage: "Hackers walked in through the front door. The password was 'password' again.",
			Effects: []Effect{
				{Type: EventLoss, Target: "data", Value: 0.1},
				{Type: EventLoss, Target: "gold", Value: 0.08},
			},
		},
		{
			Name: "Crypto Boom", Key: "crypto_boom",
			// From the Fusion Age: the first age a town makes crypto. A
			// boom is a share of what the town makes, and in the Cyberpunk
			// Age that is nothing.
			MinAge: "fusion_age", Weight: 7, MinTick: 700, Cooldown: 200,
			Duration: 15, Sentiment: "good",
			LogMessage: "Crypto moons and your least competent worker is suddenly a visionary.",
			Effects: []Effect{
				{Type: EventRate, Target: "crypto", Value: 1},
			},
		},
		{
			Name: "Crypto Winter", Key: "crypto_winter",
			MinAge: "cyberpunk_age", Weight: 8, MinTick: 700, Cooldown: 300,
			// Instant: its one effect is a loss. (It was typed for 14 ticks
			// and sat in the active events with nothing to show for them.)
			Duration: 0, Sentiment: "bad",
			LogMessage: "Crypto cratered overnight and the thought leader has stopped posting.",
			Effects: []Effect{
				{Type: EventLoss, Target: "crypto", Value: 0.12},
			},
		},
		{
			Name: "Plasma Storm", Key: "plasma_storm",
			MinAge: "fusion_age", Weight: 5, MinTick: 800, Cooldown: 220,
			Duration: 10, Sentiment: "mixed",
			LogMessage: "The sun threw a tantrum and the sky turned a worrying color.",
			Effects: []Effect{
				{Type: EventRate, Target: "electricity", Value: -0.5},
				{Type: EventRate, Target: "plasma", Value: 0.5},
			},
		},
		{
			Name: "First Contact", Key: "first_contact",
			MinAge: "space_age", Weight: 3, MinTick: 900, Cooldown: 300,
			Duration: 0, Sentiment: "good",
			LogMessage: "We are not alone, and they seem mildly disappointed in us.",
			Effects: []Effect{
				{Type: EventGain, Target: ResearchTarget, Value: 15},
				{Type: EventGain, Target: "titanium", Value: 5},
			},
		},
		{
			Name: "Dark Matter Rift", Key: "dark_matter_rift",
			MinAge: "interstellar_age", Weight: 4, MinTick: 1000, Cooldown: 280,
			Duration: 15, Sentiment: "good",
			LogMessage: "A hole in spacetime is leaking dark matter and the team is catching it in buckets.",
			Effects: []Effect{
				{Type: EventRate, Target: "dark_matter", Value: 1},
			},
		},
		{
			Name: "Quantum Fluctuation", Key: "quantum_fluctuation",
			MinAge: "quantum_age", Weight: 3, MinTick: 1100, Cooldown: 300,
			Duration: 10, Sentiment: "good",
			LogMessage: "Reality stuttered and Tuesday happened twice.",
			Effects: []Effect{
				{Type: EventRate, Target: "quantum_flux", Value: 1},
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
			Duration: 60, Raid: true, Sentiment: "bad",
			LogMessage: "A rival clan raids in the dark, doing a lot of yelling. It works.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: -0.2},
				{Type: EventLoss, Target: "food", Value: 0.08},
				{Type: "worker_loss", Value: 0.1},
			},
		},
		{
			Name: "Humming Grove", Key: "sacred_grove_found", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 10, MinTick: 20, Cooldown: 100,
			Duration: 120, Sentiment: "good",
			LogMessage: "Hunters found a grove that hums when you stand in it. It's holy now.",
			Effects: []Effect{
				{Type: EventRate, Target: "faith", Value: 0.5},
				{Type: EventGain, Target: "wood", Value: 5},
			},
		},
		{
			Name: "Beast Stampede", Key: "beast_stampede", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 8, MinTick: 15, Cooldown: 90,
			Duration: 0, Raid: true, Sentiment: "bad",
			LogMessage: "Enormous beasts stampeded straight through the fence, which lost the argument.",
			Effects: []Effect{
				{Type: EventLoss, Target: "wood", Value: 0.1},
				{Type: EventLoss, Target: "food", Value: 0.08},
			},
		},
		{
			Name: "River Blessing", Key: "river_flooding", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 12, MinTick: 10, Cooldown: 100,
			Duration: 144, Sentiment: "good",
			LogMessage: "The river flooded, the soil is now magnificent, and the shaman insists this was the plan.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: 0.25},
			},
		},
		{
			Name: "Wandering Sage", Key: "wandering_sage", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 7, MinTick: 30, Cooldown: 120,
			Duration: 0, Sentiment: "good",
			LogMessage: "An old wanderer talked for nine hours by the fire. Most of it was useful.",
			Effects: []Effect{
				{Type: EventGain, Target: ResearchTarget, Value: 10},
				{Type: EventGain, Target: "faith", Value: 10},
			},
		},

		// === IRON ERA ===
		{
			Name: "Iron Vein Strike", Key: "iron_vein_strike", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 10, MinTick: 100, Cooldown: 120,
			Duration: 180, Sentiment: "good",
			LogMessage: "Miners struck a rich iron seam and the blacksmith openly wept.",
			Effects: []Effect{
				{Type: EventRate, Target: "iron", Value: 0.3},
			},
		},
		{
			Name: "Locust Swarm", Key: "locust_swarm", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 9, MinTick: 100, Cooldown: 100,
			Duration: 120, Sentiment: "bad",
			LogMessage: "Locusts stripped the fields bare and ate a man's hat for good measure.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: -0.3},
				{Type: "worker_loss", Value: 0.12},
			},
		},
		{
			Name: "Conquered Village", Key: "conquered_village", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 8, MinTick: 120, Cooldown: 150,
			Duration: 0, Sentiment: "good",
			LogMessage: "The legions came home victorious and insufferably smug. The treasury, at least, is happy.",
			Effects: []Effect{
				{Type: EventGain, Target: "gold", Value: 10},
			},
		},
		{
			Name: "Imperial Road", Key: "roman_road_built", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 8, MinTick: 120, Cooldown: 140,
			Duration: 216, Sentiment: "good",
			LogMessage: "The new road is finished and unnervingly straight.",
			Effects: []Effect{
				{Type: EventRate, Target: "gold", Value: 0.2},
			},
		},
		{
			Name: "Oracle's Prophecy", Key: "oracle_prophecy", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 7, MinTick: 100, Cooldown: 160,
			Duration: 144, Sentiment: "good",
			LogMessage: "The oracle delivered a prophecy vague enough to never be wrong, and the people are inspired.",
			Effects: []Effect{
				{Type: EventRate, Target: "faith", Value: 0.3},
				{Type: EventRate, Target: ResearchTarget, Value: 0.15},
			},
		},

		// === STEEL ERA ===
		{
			Name: "Coal Seam Discovery", Key: "coal_seam_discovery", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 10, MinTick: 250, Cooldown: 120,
			Duration: 180, Sentiment: "good",
			LogMessage: "A vast coal seam turned up under the hills. The sky will pay for this later.",
			Effects: []Effect{
				{Type: EventRate, Target: "coal", Value: 0.3},
			},
		},
		{
			Name: "Workers' Uprising", Key: "workers_uprising", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 9, MinTick: 250, Cooldown: 130,
			Duration: 120, Sentiment: "bad",
			LogMessage: "The workers are striking and their demands are entirely reasonable, which has management rattled.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: -0.15},
				{Type: EventLoss, Target: "faith", Value: 0.15},
				{Type: "worker_loss", Value: 0.08},
			},
		},
		{
			Name: "Colonial Bounty", Key: "colonial_bounty", EpochKey: "steel_era",
			MinAge: "colonial_age", Weight: 8, MinTick: 280, Cooldown: 150,
			Duration: 0, Sentiment: "good",
			LogMessage: "The fleets came home heavy with gold and a few crates nobody wants to discuss.",
			Effects: []Effect{
				{Type: EventGain, Target: "gold", Value: 12},
			},
		},
		{
			Name: "Steam Age Inventor", Key: "steam_inventor", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 7, MinTick: 260, Cooldown: 160,
			Duration: 144, Sentiment: "good",
			LogMessage: "An inventor unveiled a steam engine that exploded only twice during the demo, a record.",
			Effects: []Effect{
				{Type: EventGain, Target: ResearchTarget, Value: 8},
				{Type: EventRate, Target: ResearchTarget, Value: 0.2},
			},
		},
		{
			Name: "Industrial Blight", Key: "industrial_blight", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 9, MinTick: 250, Cooldown: 120,
			Duration: 144, Sentiment: "bad",
			LogMessage: "Factory runoff turned the river a color fish were not meant to be.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: -0.2},
				{Type: EventLoss, Target: "faith", Value: 0.1},
			},
		},

		// === ELECTRIC ERA ===
		{
			Name: "Grid Surge", Key: "epoch_power_surge", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 10, MinTick: 350, Cooldown: 120,
			Duration: 144, Sentiment: "good",
			LogMessage: "The grid surged so hard a toaster briefly gained sentience. It's fine now.",
			Effects: []Effect{
				{Type: EventRate, Target: "electricity", Value: 0.3},
			},
		},
		{
			Name: "Oil Strike", Key: "oil_strike", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 8, MinTick: 360, Cooldown: 150,
			Duration: 180, Sentiment: "good",
			LogMessage: "A gusher blew and now everyone's covered in oil and grinning.",
			Effects: []Effect{
				{Type: EventRate, Target: "oil", Value: 0.4},
				{Type: EventGain, Target: "gold", Value: 8},
			},
		},
		{
			Name: "The Broadcast", Key: "radio_broadcast", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 8, MinTick: 350, Cooldown: 130,
			Duration: 180, Sentiment: "good",
			LogMessage: "A single broadcast reached millions and proved they'll believe nearly anything.",
			Effects: []Effect{
				{Type: EventGain, Target: "culture", Value: 10},
				{Type: EventRate, Target: "faith", Value: 0.2},
			},
		},
		{
			Name: "Labor Movement", Key: "labor_movement", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 9, MinTick: 350, Cooldown: 120,
			Duration: 60, Sentiment: "bad",
			LogMessage: "The workers organized, and management is learning what 'collective bargaining' means the hard way.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: -0.15},
				{Type: EventRate, Target: "gold", Value: -0.15},
			},
		},
		{
			Name: "Nuclear Theory", Key: "nuclear_theory", EpochKey: "electric_era",
			MinAge: "atomic_age", Weight: 6, MinTick: 400, Cooldown: 200,
			Duration: 180, Sentiment: "good",
			LogMessage: "A physicist published something nobody understands, so obviously it's genius.",
			Effects: []Effect{
				{Type: EventGain, Target: ResearchTarget, Value: 10},
				{Type: EventRate, Target: ResearchTarget, Value: 0.25},
			},
		},

		// === DIGITAL ERA ===
		{
			Name: "The Great Breach", Key: "epoch_data_breach", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 9, MinTick: 550, Cooldown: 120,
			Duration: 120, Raid: true, Sentiment: "bad",
			LogMessage: "Terabytes gone because someone clicked a link promising a free cruise.",
			Effects: []Effect{
				{Type: EventLoss, Target: "data", Value: 0.2},
				{Type: EventRate, Target: ResearchTarget, Value: -0.2},
			},
		},
		{
			Name: "Viral Moment", Key: "viral_moment", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 10, MinTick: 550, Cooldown: 100,
			Duration: 0, Sentiment: "good",
			LogMessage: "Something went viral across every network and future historians will deny knowing what it was.",
			Effects: []Effect{
				{Type: EventGain, Target: "culture", Value: 15},
			},
		},
		{
			Name: "Tech Monopoly", Key: "tech_monopoly", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 8, MinTick: 560, Cooldown: 150,
			Duration: 180, Sentiment: "good",
			LogMessage: "Your platforms now own the market and the regulators are visibly typing something.",
			Effects: []Effect{
				{Type: EventRate, Target: "gold", Value: 0.3},
			},
		},
		{
			Name: "Server Outage", Key: "server_outage", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 9, MinTick: 550, Cooldown: 110,
			Duration: 120, Sentiment: "bad",
			LogMessage: "The data centers are down and 'turn it off and on again' has officially failed.",
			Effects: []Effect{
				{Type: EventRate, Target: "data", Value: -0.4},
			},
		},
		{
			Name: "AI Breakthrough", Key: "ai_breakthrough", EpochKey: "digital_era",
			MinAge: "cyberpunk_age", Weight: 6, MinTick: 650, Cooldown: 200,
			Duration: 216, Sentiment: "good",
			LogMessage: "The research AI improved itself and then said 'please' for more compute, which is fine and not at all ominous.",
			Effects: []Effect{
				{Type: EventRate, Target: ResearchTarget, Value: 0.3},
				{Type: EventRate, Target: "data", Value: 0.2},
			},
		},

		// === NEON ERA ===
		{
			Name: "Plasma Windfall", Key: "epoch_plasma_storm", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 10, MinTick: 750, Cooldown: 120,
			Duration: 180, Sentiment: "good",
			LogMessage: "A plasma ejection dumped free energy across the grid and the accountants are weeping with joy.",
			Effects: []Effect{
				{Type: EventRate, Target: "plasma", Value: 0.4},
				{Type: EventRate, Target: "electricity", Value: 0.3},
			},
		},
		{
			Name: "Void Rift", Key: "void_rift", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 7, MinTick: 760, Cooldown: 180,
			Duration: 0, Sentiment: "good",
			// Dark matter crystals, not dark matter: dark matter is locked
			// until the Interstellar Age, one era on, so the grant went into
			// a store the player could not see.
			LogMessage: "A void rift opened and is leaking dark matter; the 'do not stand here' sign is being roundly ignored.",
			Effects: []Effect{
				{Type: EventGain, Target: "dark_matter_crystals", Value: 10},
			},
		},
		{
			Name: "Neural Uprising", Key: "neural_uprising", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 9, MinTick: 750, Cooldown: 130,
			Duration: 120, Sentiment: "bad",
			LogMessage: "The augmented workers revolted, coordinating the whole thing over the surveillance network we built.",
			Effects: []Effect{
				{Type: EventLoss, Target: "food", Value: 0.12},
				{Type: EventRate, Target: "food", Value: -0.15},
				{Type: "worker_loss", Value: 0.2},
			},
		},
		{
			Name: "Corporate Espionage", Key: "corporate_espionage", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 8, MinTick: 760, Cooldown: 140,
			Duration: 0, Raid: true, Sentiment: "bad",
			LogMessage: "A rival corp robbed you blind, and the spy left a five-star review of your security.",
			Effects: []Effect{
				{Type: EventLoss, Target: "gold", Value: 0.15},
				{Type: EventLoss, Target: "data", Value: 0.15},
			},
		},
		{
			Name: "Stellar Migration", Key: "stellar_migration", EpochKey: "neon_era",
			MinAge: "space_age", Weight: 8, MinTick: 800, Cooldown: 160,
			Duration: 144, Sentiment: "mixed",
			LogMessage: "Generation ships docked, full of passengers sick of spaceship food and keen on yours.",
			Effects: []Effect{
				{Type: EventGain, Target: "food", Value: 5},
				{Type: EventRate, Target: "food", Value: -0.15},
			},
		},

		// === COSMIC ERA ===
		{
			Name: "Reality Fracture", Key: "reality_fracture", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 9, MinTick: 1050, Cooldown: 150,
			Duration: 120, Sentiment: "bad",
			LogMessage: "Spacetime fractured and now effects keep arriving before their causes.",
			Effects: []Effect{
				{Type: EventRate, Target: "quantum_flux", Value: -0.4},
				{Type: EventRate, Target: ResearchTarget, Value: -0.1},
			},
		},
		{
			Name: "Dimensional Harvest", Key: "dimensional_harvest", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 8, MinTick: 1060, Cooldown: 160,
			Duration: 0, Sentiment: "good",
			LogMessage: "Exotic matter poured out of a tear in reality and we have wisely decided not to ask whose it was.",
			Effects: []Effect{
				{Type: EventGain, Target: "antimatter", Value: 8},
				{Type: EventGain, Target: "quantum_flux", Value: 8},
			},
		},
		{
			Name: "Galactic Council", Key: "galactic_council", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 6, MinTick: 1050, Cooldown: 200,
			Duration: 216, Sentiment: "good",
			LogMessage: "The galactic council recognized us as a real civilization, which mostly means more paperwork and a stipend.",
			Effects: []Effect{
				{Type: EventRate, Target: "gold", Value: 0.2},
				{Type: EventGain, Target: "gold", Value: 10},
			},
		},
		{
			Name: "Entropy Wave", Key: "entropy_wave", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 8, MinTick: 1050, Cooldown: 140,
			Duration: 144, Sentiment: "bad",
			LogMessage: "An entropy wave swept through and everything is now slightly more worn out, the universe included.",
			Effects: []Effect{
				{Type: EventRate, Target: "quantum_flux", Value: -0.2},
				{Type: EventRate, Target: ResearchTarget, Value: -0.2},
			},
		},
		{
			Name: "Transcendence Signal", Key: "transcendence_signal", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 4, MinTick: 1100, Cooldown: 300,
			Duration: 0, Sentiment: "good",
			LogMessage: "A signal from the edge of everything rewrote what we know, and politely began with 'Hello.'",
			Effects: []Effect{
				{Type: EventGain, Target: ResearchTarget, Value: 15},
				{Type: EventGain, Target: "culture", Value: 10},
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
