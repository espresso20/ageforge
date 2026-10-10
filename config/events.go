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
	Effects     []Effect
	Description string
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
			Description: "A season of plenty yields bonus food. Nobody asks why; you simply eat.",
			LogMessage:  "The harvest came in fat and early, zero questions asked.",
			Effects: []Effect{
				{Type: EventGain, Target: "food", Value: 5},
			},
		},
		{
			Name: "Wandering Traders", Key: "wandering_traders",
			MinAge: "bronze_age", Weight: 12, MinTick: 60, Cooldown: 80,
			Duration: 0, Sentiment: "good",
			Description: "Traveling merchants share their goods. They smell of cabbage and opportunity.",
			LogMessage:  "Wandering traders arrive, smelling of cabbage and opportunity.",
			Effects: []Effect{
				{Type: EventGain, Target: "gold", Value: 5},
				{Type: EventGain, Target: "food", Value: 3},
			},
		},
		{
			Name: "Gold Rush", Key: "gold_rush",
			MinAge: "bronze_age", Weight: 8, MinTick: 100, Cooldown: 150,
			Duration: 15, Sentiment: "good",
			Description: "Gold deposits discovered. Half the workforce has already quit to dig holes.",
			LogMessage:  "Someone found gold in the creek and everyone is now a prospector.",
			Effects: []Effect{
				{Type: EventRate, Target: "gold", Value: 1},
			},
		},
		{
			Name: "Skilled Immigrants", Key: "skilled_immigrants",
			MinAge: "stone_age", Weight: 10, MinTick: 40, Cooldown: 100,
			Duration: 0, Sentiment: "good",
			Description: "Skilled people seek to join your civilization, references and all.",
			LogMessage:  "Skilled newcomers arrive with strong opinions and stronger résumés.",
			Effects: []Effect{
				{Type: EventGain, Target: "knowledge", Value: 5},
			},
		},
		{
			Name: "Ancient Discovery", Key: "ancient_discovery",
			MinAge: "iron_age", Weight: 6, MinTick: 150, Cooldown: 200,
			Duration: 0, Sentiment: "good",
			Description: "Ancient ruins reveal forgotten knowledge, and one very smug skeleton.",
			LogMessage:  "Ruins older than memory give up their secrets, and one ominous skull.",
			Effects: []Effect{
				{Type: EventGain, Target: "knowledge", Value: 10},
			},
		},
		{
			Name: "Trade Boom", Key: "trade_boom",
			MinAge: "medieval_age", Weight: 8, MinTick: 200, Cooldown: 120,
			Duration: 20, Sentiment: "good",
			Description: "A surge in trade activity boosts gold. The merchants are insufferable about it.",
			LogMessage:  "Trade is booming and the merchants will not stop talking about it.",
			Effects: []Effect{
				{Type: EventRate, Target: "gold", Value: 1},
			},
		},

		// === NEGATIVE EVENTS ===
		{
			Name: "Drought", Key: "drought",
			MinAge: "primitive_age", Weight: 12, MinTick: 30, Cooldown: 80,
			Duration: 10, Sentiment: "bad",
			Description: "Dry conditions reduce food. The rain dance was, in hindsight, optimistic.",
			LogMessage:  "It hasn't rained in weeks and the dancing isn't helping.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: -0.5},
			},
		},
		{
			Name: "Plague", Key: "plague",
			MinAge: "stone_age", Weight: 6, MinTick: 80, Cooldown: 200,
			Duration: 8, Sentiment: "bad",
			Description: "Disease spreads through your population. The local healer recommends more leeches.",
			LogMessage:  "A plague spreads and the healer's plan is, alarmingly, more leeches.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: -0.5},
				{Type: "worker_loss", Value: 0.15},
			},
		},
		{
			Name: "Bandit Raid", Key: "bandit_raid",
			MinAge: "bronze_age", Weight: 10, MinTick: 60, Cooldown: 60,
			Duration: 0, Raid: true, Sentiment: "bad",
			Description: "Bandits attack and steal resources. They left a thank-you note, which is somehow worse.",
			LogMessage:  "Bandits cleaned out the stores and left a polite thank-you note.",
			Effects: []Effect{
				{Type: EventLoss, Target: "food", Value: 0.06},
				{Type: EventLoss, Target: "gold", Value: 0.06},
			},
		},
		{
			Name: "Storm", Key: "storm",
			MinAge: "primitive_age", Weight: 14, MinTick: 25, Cooldown: 50,
			Duration: 5, Sentiment: "bad",
			Description: "A fierce storm hampers wood gathering. Several roofs have gone exploring.",
			LogMessage:  "A storm rolls in and takes several roofs with it.",
			Effects: []Effect{
				{Type: EventRate, Target: "wood", Value: -0.5},
			},
		},
		{
			Name: "Mine Collapse", Key: "mine_collapse",
			MinAge: "iron_age", Weight: 7, MinTick: 120, Cooldown: 150,
			Duration: 8, Sentiment: "bad",
			Description: "A mine collapses. The foreman insists the support beams were 'decorative anyway.'",
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
			Description: "Religious dissent reduces faith. Someone has started a rival sect in a nicer barn.",
			LogMessage:  "A breakaway sect has set up in a nicer barn and is poaching the congregation.",
			Effects: []Effect{
				{Type: EventRate, Target: "faith", Value: -0.5},
			},
		},

		// === MIXED / SPECIAL EVENTS ===
		{
			Name: "Earthquake", Key: "earthquake",
			MinAge: "stone_age", Weight: 5, MinTick: 100, Cooldown: 200,
			Duration: 0, Sentiment: "mixed",
			Description: "An earthquake rearranges the village. The new layout is, on balance, worse.",
			LogMessage:  "The ground shrugged and rearranged the village, and the cracks turned up stone.",
			Effects: []Effect{
				{Type: EventLoss, Target: "wood", Value: 0.05},
				{Type: EventGain, Target: "stone", Value: 5},
			},
		},
		{
			Name: "Renaissance Fair", Key: "renaissance_fair",
			MinAge: "renaissance_age", Weight: 10, MinTick: 250, Cooldown: 100,
			Duration: 15, Sentiment: "good",
			Description: "A cultural festival boosts culture and gold. There is a man juggling. Why is there a man juggling.",
			LogMessage:  "A festival breaks out, complete with an unexplained juggler.",
			Effects: []Effect{
				{Type: EventRate, Target: "culture", Value: 0.5},
				{Type: EventRate, Target: "gold", Value: 0.5},
			},
		},
		{
			Name: "Industrial Accident", Key: "industrial_accident",
			MinAge: "industrial_age", Weight: 8, MinTick: 300, Cooldown: 120,
			Duration: 0, Sentiment: "bad",
			Description: "A factory accident. The safety poster, now on fire, reminded everyone to be careful.",
			LogMessage:  "Something exploded that wasn't supposed to, and the 'Be Careful' poster is also on fire.",
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
			Description: "A colonial expedition returns with treasure and a parrot of dubious vocabulary.",
			LogMessage:  "The expedition returns rich, sunburnt, and accompanied by a parrot nobody approved.",
			Effects: []Effect{
				{Type: EventGain, Target: "gold", Value: 8},
				{Type: EventGain, Target: "culture", Value: 5},
			},
		},
		{
			Name: "Pirate Attack", Key: "pirate_attack",
			MinAge: "colonial_age", Weight: 7, MinTick: 320, Cooldown: 140,
			Duration: 0, Raid: true, Sentiment: "bad",
			Description: "Pirates raid your trade routes. They are, regrettably, very good at this.",
			LogMessage:  "Pirates hit the trade lanes again. They're alarmingly professional about it.",
			Effects: []Effect{
				{Type: EventLoss, Target: "gold", Value: 0.08},
				{Type: EventLoss, Target: "food", Value: 0.05},
			},
		},
		{
			Name: "Power Surge", Key: "power_surge_base",
			MinAge: "victorian_age", Weight: 6, MinTick: 400, Cooldown: 160,
			Duration: 10, Sentiment: "good",
			Description: "An electrical surge runs through the grid. The lights flicker in a way best described as 'enthusiastic.'",
			LogMessage:  "The grid surges and the lights flicker enthusiastically.",
			Effects: []Effect{
				{Type: EventRate, Target: "electricity", Value: 1},
			},
		},
		{
			Name: "Nuclear Scare", Key: "nuclear_scare",
			MinAge: "atomic_age", Weight: 4, MinTick: 500, Cooldown: 250,
			Duration: 12, Sentiment: "bad",
			Description: "Nuclear anxiety reduces productivity. Everyone has built a bunker; no one is in their bunker working.",
			LogMessage:  "A nuclear scare grips the population and the bunkers are fully staffed.",
			Effects: []Effect{
				{Type: EventRate, Target: "electricity", Value: -0.4},
				{Type: EventRate, Target: "knowledge", Value: -0.3},
			},
		},
		{
			Name: "Data Breach", Key: "data_breach",
			MinAge: "information_age", Weight: 6, MinTick: 600, Cooldown: 180,
			Duration: 0, Raid: true, Sentiment: "bad",
			Description: "Hackers steal your data reserves. The password was 'password.' It is always 'password.'",
			LogMessage:  "Hackers walked in through the front door. The password was 'password' again.",
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
			Description: "Cryptocurrency values skyrocket. Your most useless worker is now a thought leader.",
			LogMessage:  "Crypto moons and your least competent worker is suddenly a visionary.",
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
			Description: "Cryptocurrency values plummet in a flash crash. The thought leader has gone very quiet.",
			LogMessage:  "Crypto cratered overnight and the thought leader has stopped posting.",
			Effects: []Effect{
				{Type: EventLoss, Target: "crypto", Value: 0.12},
			},
		},
		{
			Name: "Plasma Storm", Key: "plasma_storm",
			MinAge: "fusion_age", Weight: 5, MinTick: 800, Cooldown: 220,
			Duration: 10, Sentiment: "mixed",
			Description: "Solar plasma eruption disrupts power but yields plasma. The sky is doing something unsettling.",
			LogMessage:  "The sun threw a tantrum and the sky turned a worrying color.",
			Effects: []Effect{
				{Type: EventRate, Target: "electricity", Value: -0.5},
				{Type: EventRate, Target: "plasma", Value: 0.5},
			},
		},
		{
			Name: "First Contact", Key: "first_contact",
			MinAge: "space_age", Weight: 3, MinTick: 900, Cooldown: 300,
			Duration: 0, Sentiment: "good",
			Description: "Contact with alien intelligence yields knowledge. They seem disappointed it took this long.",
			LogMessage:  "We are not alone, and they seem mildly disappointed in us.",
			Effects: []Effect{
				{Type: EventGain, Target: "knowledge", Value: 15},
				{Type: EventGain, Target: "titanium", Value: 5},
			},
		},
		{
			Name: "Dark Matter Rift", Key: "dark_matter_rift",
			MinAge: "interstellar_age", Weight: 4, MinTick: 1000, Cooldown: 280,
			Duration: 15, Sentiment: "good",
			Description: "A rift in spacetime leaks dark matter. The science team is collecting it in buckets.",
			LogMessage:  "A hole in spacetime is leaking dark matter and the team is catching it in buckets.",
			Effects: []Effect{
				{Type: EventRate, Target: "dark_matter", Value: 1},
			},
		},
		{
			Name: "Quantum Fluctuation", Key: "quantum_fluctuation",
			MinAge: "quantum_age", Weight: 3, MinTick: 1100, Cooldown: 300,
			Duration: 10, Sentiment: "good",
			Description: "Reality destabilizes briefly but yields quantum flux. Tuesday happened twice. Nobody minded.",
			LogMessage:  "Reality stuttered and Tuesday happened twice.",
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
			Description: "Rival clans descend in the night, yelling things. The yelling, frankly, works.",
			LogMessage:  "A rival clan raids in the dark, doing a lot of yelling. It works.",
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
			Description: "Hunters find a grove that hums. They have decided it is holy. They may be right.",
			LogMessage:  "Hunters found a grove that hums when you stand in it. It's holy now.",
			Effects: []Effect{
				{Type: EventRate, Target: "faith", Value: 0.5},
				{Type: EventGain, Target: "wood", Value: 5},
			},
		},
		{
			Name: "Beast Stampede", Key: "beast_stampede", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 8, MinTick: 15, Cooldown: 90,
			Duration: 0, Raid: true, Sentiment: "bad",
			Description: "Very large animals run through the settlement at speed. The fence had opinions about this. The fence lost.",
			LogMessage:  "Enormous beasts stampeded straight through the fence, which lost the argument.",
			Effects: []Effect{
				{Type: EventLoss, Target: "wood", Value: 0.1},
				{Type: EventLoss, Target: "food", Value: 0.08},
			},
		},
		{
			Name: "River Blessing", Key: "river_flooding", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 12, MinTick: 10, Cooldown: 100,
			Duration: 144, Sentiment: "good",
			Description: "The river floods and leaves rich silt everywhere. The shaman is taking full credit.",
			LogMessage:  "The river flooded, the soil is now magnificent, and the shaman insists this was the plan.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: 0.25},
			},
		},
		{
			Name: "Wandering Sage", Key: "wandering_sage", EpochKey: "stone_era",
			MinAge: "primitive_age", Weight: 7, MinTick: 30, Cooldown: 120,
			Duration: 0, Sentiment: "good",
			Description: "A traveling elder trades wisdom for a warm fire and a captive audience.",
			LogMessage:  "An old wanderer talked for nine hours by the fire. Most of it was useful.",
			Effects: []Effect{
				{Type: EventGain, Target: "knowledge", Value: 10},
				{Type: EventGain, Target: "faith", Value: 10},
			},
		},

		// === IRON ERA ===
		{
			Name: "Iron Vein Strike", Key: "iron_vein_strike", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 10, MinTick: 100, Cooldown: 120,
			Duration: 180, Sentiment: "good",
			Description: "Miners hit a seam of high-grade ore. The blacksmith wept, then got back to work.",
			LogMessage:  "Miners struck a rich iron seam and the blacksmith openly wept.",
			Effects: []Effect{
				{Type: EventRate, Target: "iron", Value: 0.3},
			},
		},
		{
			Name: "Locust Swarm", Key: "locust_swarm", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 9, MinTick: 100, Cooldown: 100,
			Duration: 120, Sentiment: "bad",
			Description: "A plague of locusts eats the fields, the seed stores, and one farmer's hat.",
			LogMessage:  "Locusts stripped the fields bare and ate a man's hat for good measure.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: -0.3},
				{Type: "worker_loss", Value: 0.12},
			},
		},
		{
			Name: "Conquered Village", Key: "conquered_village", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 8, MinTick: 120, Cooldown: 150,
			Duration: 0, Sentiment: "good",
			Description: "Your legions return victorious with tribute, captives, and unbearable swagger.",
			LogMessage:  "The legions came home victorious and insufferably smug. The treasury, at least, is happy.",
			Effects: []Effect{
				{Type: EventGain, Target: "gold", Value: 10},
			},
		},
		{
			Name: "Imperial Road", Key: "roman_road_built", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 8, MinTick: 120, Cooldown: 140,
			Duration: 216, Sentiment: "good",
			Description: "A great road opens for traders and carts. It is suspiciously, perfectly straight.",
			LogMessage:  "The new road is finished and unnervingly straight.",
			Effects: []Effect{
				{Type: EventRate, Target: "gold", Value: 0.2},
			},
		},
		{
			Name: "Oracle's Prophecy", Key: "oracle_prophecy", EpochKey: "iron_era",
			MinAge: "iron_age", Weight: 7, MinTick: 100, Cooldown: 160,
			Duration: 144, Sentiment: "good",
			Description: "The oracle speaks of destiny. As always, it is vague enough to be technically correct.",
			LogMessage:  "The oracle delivered a prophecy vague enough to never be wrong, and the people are inspired.",
			Effects: []Effect{
				{Type: EventRate, Target: "faith", Value: 0.3},
				{Type: EventRate, Target: "knowledge", Value: 0.15},
			},
		},

		// === STEEL ERA ===
		{
			Name: "Coal Seam Discovery", Key: "coal_seam_discovery", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 10, MinTick: 250, Cooldown: 120,
			Duration: 180, Sentiment: "good",
			Description: "Surveyors uncover an enormous coal deposit. The air quality forecast is, in return, grim.",
			LogMessage:  "A vast coal seam turned up under the hills. The sky will pay for this later.",
			Effects: []Effect{
				{Type: EventRate, Target: "coal", Value: 0.3},
			},
		},
		{
			Name: "Workers' Uprising", Key: "workers_uprising", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 9, MinTick: 250, Cooldown: 130,
			Duration: 120, Sentiment: "bad",
			Description: "Workers strike for better conditions. The demands are reasonable, which is the truly alarming part.",
			LogMessage:  "The workers are striking and their demands are entirely reasonable, which has management rattled.",
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
			Description: "Trade fleets return laden with gold and several crates marked 'do not open.'",
			LogMessage:  "The fleets came home heavy with gold and a few crates nobody wants to discuss.",
			Effects: []Effect{
				{Type: EventGain, Target: "gold", Value: 12},
			},
		},
		{
			Name: "Steam Age Inventor", Key: "steam_inventor", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 7, MinTick: 260, Cooldown: 160,
			Duration: 144, Sentiment: "good",
			Description: "A brilliant inventor unveils a steam engine. It only exploded twice during the demonstration.",
			LogMessage:  "An inventor unveiled a steam engine that exploded only twice during the demo, a record.",
			Effects: []Effect{
				{Type: EventGain, Target: "knowledge", Value: 8},
				{Type: EventRate, Target: "knowledge", Value: 0.2},
			},
		},
		{
			Name: "Industrial Blight", Key: "industrial_blight", EpochKey: "steel_era",
			MinAge: "industrial_age", Weight: 9, MinTick: 250, Cooldown: 120,
			Duration: 144, Sentiment: "bad",
			Description: "Factory runoff poisons the river. The fish are now an unusual color and so is breakfast.",
			LogMessage:  "Factory runoff turned the river a color fish were not meant to be.",
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
			Description: "An unexpected surge runs through the grid. Three toasters achieved sentience and were talked down.",
			LogMessage:  "The grid surged so hard a toaster briefly gained sentience. It's fine now.",
			Effects: []Effect{
				{Type: EventRate, Target: "electricity", Value: 0.3},
			},
		},
		{
			Name: "Oil Strike", Key: "oil_strike", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 8, MinTick: 360, Cooldown: 150,
			Duration: 180, Sentiment: "good",
			Description: "Black gold erupts from a borehole. Everyone is covered in it. Everyone is delighted.",
			LogMessage:  "A gusher blew and now everyone's covered in oil and grinning.",
			Effects: []Effect{
				{Type: EventRate, Target: "oil", Value: 0.4},
				{Type: EventGain, Target: "gold", Value: 8},
			},
		},
		{
			Name: "The Broadcast", Key: "radio_broadcast", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 8, MinTick: 350, Cooldown: 130,
			Duration: 180, Sentiment: "good",
			Description: "A radio signal reaches millions. Most of them, it turns out, will believe anything.",
			LogMessage:  "A single broadcast reached millions and proved they'll believe nearly anything.",
			Effects: []Effect{
				{Type: EventGain, Target: "culture", Value: 10},
				{Type: EventRate, Target: "faith", Value: 0.2},
			},
		},
		{
			Name: "Labor Movement", Key: "labor_movement", EpochKey: "electric_era",
			MinAge: "victorian_age", Weight: 9, MinTick: 350, Cooldown: 120,
			Duration: 60, Sentiment: "bad",
			Description: "Workers organize for better pay. Management has discovered the meeting that could've been a memo.",
			LogMessage:  "The workers organized, and management is learning what 'collective bargaining' means the hard way.",
			Effects: []Effect{
				{Type: EventRate, Target: "food", Value: -0.15},
				{Type: EventRate, Target: "gold", Value: -0.15},
			},
		},
		{
			Name: "Nuclear Theory", Key: "nuclear_theory", EpochKey: "electric_era",
			MinAge: "atomic_age", Weight: 6, MinTick: 400, Cooldown: 200,
			Duration: 180, Sentiment: "good",
			Description: "A physicist publishes a paradigm-shifting theory. Nobody understands it, which proves it's brilliant.",
			LogMessage:  "A physicist published something nobody understands, so obviously it's genius.",
			Effects: []Effect{
				{Type: EventGain, Target: "knowledge", Value: 10},
				{Type: EventRate, Target: "knowledge", Value: 0.25},
			},
		},

		// === DIGITAL ERA ===
		{
			Name: "The Great Breach", Key: "epoch_data_breach", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 9, MinTick: 550, Cooldown: 120,
			Duration: 120, Raid: true, Sentiment: "bad",
			Description: "A sophisticated attack siphons terabytes of data. The intern clicked the link. Of course the intern clicked the link.",
			LogMessage:  "Terabytes gone because someone clicked a link promising a free cruise.",
			Effects: []Effect{
				{Type: EventLoss, Target: "data", Value: 0.2},
				{Type: EventRate, Target: "knowledge", Value: -0.2},
			},
		},
		{
			Name: "Viral Moment", Key: "viral_moment", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 10, MinTick: 550, Cooldown: 100,
			Duration: 0, Sentiment: "good",
			Description: "A cultural upload spreads across every network at once. Historians will pretend they don't know what it was.",
			LogMessage:  "Something went viral across every network and future historians will deny knowing what it was.",
			Effects: []Effect{
				{Type: EventGain, Target: "culture", Value: 15},
			},
		},
		{
			Name: "Tech Monopoly", Key: "tech_monopoly", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 8, MinTick: 560, Cooldown: 150,
			Duration: 180, Sentiment: "good",
			Description: "Your platforms dominate global commerce. The regulators have noticed. The regulators are typing.",
			LogMessage:  "Your platforms now own the market and the regulators are visibly typing something.",
			Effects: []Effect{
				{Type: EventRate, Target: "gold", Value: 0.3},
			},
		},
		{
			Name: "Server Outage", Key: "server_outage", EpochKey: "digital_era",
			MinAge: "information_age", Weight: 9, MinTick: 550, Cooldown: 110,
			Duration: 120, Sentiment: "bad",
			Description: "A catastrophic hardware failure takes the data centers offline. Someone has tried turning it off and on again.",
			LogMessage:  "The data centers are down and 'turn it off and on again' has officially failed.",
			Effects: []Effect{
				{Type: EventRate, Target: "data", Value: -0.4},
			},
		},
		{
			Name: "AI Breakthrough", Key: "ai_breakthrough", EpochKey: "digital_era",
			MinAge: "cyberpunk_age", Weight: 6, MinTick: 650, Cooldown: 200,
			Duration: 216, Sentiment: "good",
			Description: "Your research AIs achieve recursive self-improvement. They have asked, very politely, for more compute.",
			LogMessage:  "The research AI improved itself and then said 'please' for more compute, which is fine and not at all ominous.",
			Effects: []Effect{
				{Type: EventRate, Target: "knowledge", Value: 0.3},
				{Type: EventRate, Target: "data", Value: 0.2},
			},
		},

		// === NEON ERA ===
		{
			Name: "Plasma Windfall", Key: "epoch_plasma_storm", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 10, MinTick: 750, Cooldown: 120,
			Duration: 180, Sentiment: "good",
			Description: "A stellar plasma ejection floods the system with free energy. The accountants are weeping with joy.",
			LogMessage:  "A plasma ejection dumped free energy across the grid and the accountants are weeping with joy.",
			Effects: []Effect{
				{Type: EventRate, Target: "plasma", Value: 0.4},
				{Type: EventRate, Target: "electricity", Value: 0.3},
			},
		},
		{
			Name: "Void Rift", Key: "void_rift", EpochKey: "neon_era",
			MinAge: "fusion_age", Weight: 7, MinTick: 760, Cooldown: 180,
			Duration: 0, Sentiment: "good",
			Description: "A rift bleeds exotic matter into local space. Standing near it is strongly discouraged, so naturally everyone does.",
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
			Description: "Augmented workers revolt against the surveillance state. They have, fittingly, organized it all on the surveillance network.",
			LogMessage:  "The augmented workers revolted, coordinating the whole thing over the surveillance network we built.",
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
			Description: "A rival megacorp steals gold and data. Their spy left a five-star review on the way out.",
			LogMessage:  "A rival corp robbed you blind, and the spy left a five-star review of your security.",
			Effects: []Effect{
				{Type: EventLoss, Target: "gold", Value: 0.15},
				{Type: EventLoss, Target: "data", Value: 0.15},
			},
		},
		{
			Name: "Stellar Migration", Key: "stellar_migration", EpochKey: "neon_era",
			MinAge: "space_age", Weight: 8, MinTick: 800, Cooldown: 160,
			Duration: 144, Sentiment: "mixed",
			Description: "A fleet of generation ships docks. The passengers are hungry, and very tired of the spaceships.",
			LogMessage:  "Generation ships docked, full of passengers sick of spaceship food and keen on yours.",
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
			Description: "A quantum decoherence event destabilizes local spacetime. Cause and effect are taking a short break.",
			LogMessage:  "Spacetime fractured and now effects keep arriving before their causes.",
			Effects: []Effect{
				{Type: EventRate, Target: "quantum_flux", Value: -0.4},
				{Type: EventRate, Target: "knowledge", Value: -0.1},
			},
		},
		{
			Name: "Dimensional Harvest", Key: "dimensional_harvest", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 8, MinTick: 1060, Cooldown: 160,
			Duration: 0, Sentiment: "good",
			Description: "A tear in reality yields a bounty of exotic matter. We are choosing not to ask where it came from.",
			LogMessage:  "Exotic matter poured out of a tear in reality and we have wisely decided not to ask whose it was.",
			Effects: []Effect{
				{Type: EventGain, Target: "antimatter", Value: 8},
				{Type: EventGain, Target: "quantum_flux", Value: 8},
			},
		},
		{
			Name: "Galactic Council", Key: "galactic_council", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 6, MinTick: 1050, Cooldown: 200,
			Duration: 216, Sentiment: "good",
			Description: "Alien civilizations recognize your sovereignty and send tribute. You are now, technically, doing paperwork for the galaxy.",
			LogMessage:  "The galactic council recognized us as a real civilization, which mostly means more paperwork and a stipend.",
			Effects: []Effect{
				{Type: EventRate, Target: "gold", Value: 0.2},
				{Type: EventGain, Target: "gold", Value: 10},
			},
		},
		{
			Name: "Entropy Wave", Key: "entropy_wave", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 8, MinTick: 1050, Cooldown: 140,
			Duration: 144, Sentiment: "bad",
			Description: "A wave of cosmic entropy degrades matter everywhere. The universe is, gently, giving up.",
			LogMessage:  "An entropy wave swept through and everything is now slightly more worn out, the universe included.",
			Effects: []Effect{
				{Type: EventRate, Target: "quantum_flux", Value: -0.2},
				{Type: EventRate, Target: "knowledge", Value: -0.2},
			},
		},
		{
			Name: "Transcendence Signal", Key: "transcendence_signal", EpochKey: "cosmic_era",
			MinAge: "quantum_age", Weight: 4, MinTick: 1100, Cooldown: 300,
			Duration: 0, Sentiment: "good",
			Description: "A signal from the edge of the universe rewrites your understanding of reality. It opens, against all odds, with 'Hello.'",
			LogMessage:  "A signal from the edge of everything rewrote what we know, and politely began with 'Hello.'",
			Effects: []Effect{
				{Type: EventGain, Target: "knowledge", Value: 15},
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
