package config

// ExchangeRateDef defines a one-directional spot exchange between two resources.
// The actual exchange rate seen in-game is BaseRate adjusted by market pressure
// (repeated trades drive rates down for a while — see game/trade.go).
type ExchangeRateDef struct {
	From     string  // resource key being sold
	To       string  // resource key being bought
	BaseRate float64 // units of To received per 1 unit of From at zero market pressure
	MinAge   string  // earliest age at which this exchange is unlocked
}

// TradeRouteDef defines a passive recurring trade route. Once started, the
// engine automatically exports Export resources and adds Import resources every
// TicksPerRun ticks. The route stops if Export resources are insufficient.
//
// TicksPerRun is typed for the base curve and stretched for the route's
// MinAge (StretchTicks), so an age runs as many cycles as it did before the
// one-week curve. Every route opens from the Bronze Age on, where the stretch
// is the same in every later age (a config test holds that).
type TradeRouteDef struct {
	Name        string
	Key         string
	MinAge      string             // minimum age to unlock this route
	RequiredBld string             // building key that must be built (e.g. "market", "port")
	MinCount    int                // minimum count of RequiredBld needed
	TicksPerRun int                // ticks per trade cycle (at 1x speed)
	Export      map[string]float64 // resources consumed from player each cycle
	Import      map[string]float64 // resources added to player each cycle
	Description string
}

// FactionDef defines an NPC diplomatic civilization. Civilizations become
// discoverable through age/epoch-gated first-contact events once the player
// reaches MinAge. When allied, trade involving the civ's Specialty resource
// gets a TradeBonus multiplier.
//
// Personality shapes passive opinion drift and worker-lending / war behaviour:
//   - "aggressive":    opinion trends down over time; provocable into war
//   - "peaceful":      opinion trends up; lends you workers when standing is high
//   - "mercantile":    opinion responds to your trade activity; trade-focused
//   - "isolationist":  opinion stays near neutral; slow to befriend or anger
//
// Backstory is flavour shown on first contact and in the diplomacy overlay.
type FactionDef struct {
	Name        string
	Key         string
	MinAge      string  // age at which this civ is first discoverable (first-contact gate)
	Specialty   string  // resource key the civ specialises in (used for TradeBonus)
	TradeBonus  float64 // fractional bonus (0.20 = +20%) applied to specialty-resource trades when allied
	Personality string  // "aggressive" | "peaceful" | "mercantile" | "isolationist"
	Strength    int     // 1-5; scales war raid severity (higher = harsher raids)
	Backstory   string  // flavour shown on first contact and in the overlay
	Description string
}

// ValidPersonalities is the closed set of civilization personalities. Used by
// config validation and the diplomacy drift/war logic.
var ValidPersonalities = map[string]bool{
	"aggressive":   true,
	"peaceful":     true,
	"mercantile":   true,
	"isolationist": true,
}

// BaseExchangeRates returns all exchange rate definitions
func BaseExchangeRates() []ExchangeRateDef {
	return []ExchangeRateDef{
		// Bronze Age basics
		{From: "food", To: "wood", BaseRate: 1.0, MinAge: "bronze_age"},
		{From: "wood", To: "food", BaseRate: 1.0, MinAge: "bronze_age"},
		{From: "food", To: "stone", BaseRate: 0.8, MinAge: "bronze_age"},
		{From: "stone", To: "food", BaseRate: 1.25, MinAge: "bronze_age"},
		{From: "wood", To: "stone", BaseRate: 0.9, MinAge: "bronze_age"},
		{From: "stone", To: "wood", BaseRate: 1.1, MinAge: "bronze_age"},

		// Gold exchanges
		{From: "gold", To: "food", BaseRate: 50, MinAge: "bronze_age"},
		{From: "gold", To: "wood", BaseRate: 40, MinAge: "bronze_age"},
		{From: "gold", To: "stone", BaseRate: 30, MinAge: "bronze_age"},

		// Iron Age
		{From: "iron", To: "gold", BaseRate: 2.0, MinAge: "iron_age"},
		{From: "gold", To: "iron", BaseRate: 0.4, MinAge: "iron_age"},
		{From: "iron", To: "stone", BaseRate: 3.0, MinAge: "iron_age"},

		// Medieval
		{From: "gold", To: "knowledge", BaseRate: 5.0, MinAge: "medieval_age"},
		{From: "gold", To: "culture", BaseRate: 3.0, MinAge: "medieval_age"},
		{From: "faith", To: "culture", BaseRate: 2.0, MinAge: "medieval_age"},

		// Colonial
		{From: "gold", To: "coal", BaseRate: 10, MinAge: "colonial_age"},
		{From: "coal", To: "gold", BaseRate: 0.08, MinAge: "colonial_age"},

		// Industrial
		{From: "steel", To: "gold", BaseRate: 5.0, MinAge: "industrial_age"},
		{From: "oil", To: "gold", BaseRate: 3.0, MinAge: "industrial_age"},

		// Electric
		{From: "electricity", To: "gold", BaseRate: 0.5, MinAge: "electric_age"},

		// Modern Age (data unlocks here)
		// note: was information_age. Data unlocks in the Modern Age, where ten
		// buildings (think_tank among them, which the Information Age requires)
		// cost it and nothing but a 1/tick tech produced it; the market is the
		// Modern Age's way in.
		{From: "data", To: "gold", BaseRate: 5.0, MinAge: "modern_age"},
		{From: "gold", To: "data", BaseRate: 0.15, MinAge: "modern_age"},

		// Cyberpunk
		{From: "crypto", To: "gold", BaseRate: 20.0, MinAge: "cyberpunk_age"},
		{From: "gold", To: "crypto", BaseRate: 0.04, MinAge: "cyberpunk_age"},

		// Space Age+
		{From: "dark_matter", To: "gold", BaseRate: 50.0, MinAge: "space_age"},
		{From: "quantum_flux", To: "gold", BaseRate: 100.0, MinAge: "quantum_age"},
	}
}

// ExchangeRateByKey returns exchange rates keyed by "from:to" (e.g. "gold:food").
// This compound key format avoids ambiguity between directional rate pairs.
func ExchangeRateByKey() map[string]ExchangeRateDef {
	out := make(map[string]ExchangeRateDef)
	for _, def := range BaseExchangeRates() {
		out[def.From+":"+def.To] = def
	}
	return out
}

// BaseTradeRoutes returns all trade route definitions, with TicksPerRun
// stretched for each route's MinAge.
func BaseTradeRoutes() []TradeRouteDef {
	defs := []TradeRouteDef{
		{
			Name: "Local Barter", Key: "local_barter",
			MinAge: "bronze_age", RequiredBld: "market", MinCount: 1,
			TicksPerRun: 26,
			Export:      map[string]float64{"food": 10},
			Import:      map[string]float64{"wood": 8},
			Description: "Send surplus food to nearby villages for wood.",
		},
		{
			Name: "Stone Trade", Key: "stone_trade",
			MinAge: "iron_age", RequiredBld: "market", MinCount: 2,
			TicksPerRun: 31,
			Export:      map[string]float64{"wood": 15},
			Import:      map[string]float64{"stone": 12},
			Description: "Send wood to the quarries for stone.",
		},
		{
			Name: "Gold Caravan", Key: "gold_caravan",
			MinAge: "classical_age", RequiredBld: "market", MinCount: 3,
			TicksPerRun: 39,
			Export:      map[string]float64{"stone": 50},
			Import:      map[string]float64{"gold": 5},
			Description: "Send stone caravans out for gold.",
		},
		{
			Name: "Silk Road", Key: "silk_road",
			MinAge: "medieval_age", RequiredBld: "market", MinCount: 2,
			TicksPerRun: 52,
			Export:      map[string]float64{"gold": 30},
			Import:      map[string]float64{"culture": 80},
			Description: "Send gold along the Silk Road for culture.",
		},
		{
			Name: "Spice Trade", Key: "spice_trade",
			MinAge: "colonial_age", RequiredBld: "port", MinCount: 1,
			TicksPerRun: 47,
			Export:      map[string]float64{"gold": 100},
			Import:      map[string]float64{"food": 200, "culture": 50},
			Description: "Pay gold to distant lands for spiced food and culture.",
		},
		{
			Name: "Colonial Exports", Key: "colonial_exports",
			MinAge: "colonial_age", RequiredBld: "port", MinCount: 2,
			TicksPerRun: 39,
			Export:      map[string]float64{"food": 500},
			Import:      map[string]float64{"gold": 150},
			Description: "Sell food to colonial settlements for gold.",
		},
		// --- Colonial → Industrial gap fillers (Trade Expansion) ---
		// This band was thin (only spice_trade + colonial_exports). These six
		// routes give the renaissance→industrial player something to do with
		// surplus stone/iron/coal/knowledge and a reason to build harbours.
		{
			Name: "Mercantile Convoy", Key: "mercantile_convoy",
			MinAge: "renaissance_age", RequiredBld: "exchange", MinCount: 1,
			TicksPerRun: 42,
			Export:      map[string]float64{"stone": 300, "wood": 200},
			Import:      map[string]float64{"gold": 90},
			Description: "Run armed convoys of stone and wood between city-states for gold.",
		},
		{
			Name: "Triangular Trade", Key: "triangular_trade",
			MinAge: "colonial_age", RequiredBld: "harbor", MinCount: 1,
			TicksPerRun: 47,
			Export:      map[string]float64{"food": 400, "gold": 60},
			Import:      map[string]float64{"culture": 120, "knowledge": 80},
			Description: "A three-way colonial exchange: send food and gold for culture and knowledge.",
		},
		{
			Name: "Tea Clippers", Key: "tea_clippers",
			MinAge: "colonial_age", RequiredBld: "harbor", MinCount: 2,
			TicksPerRun: 52,
			Export:      map[string]float64{"gold": 250},
			Import:      map[string]float64{"food": 600, "culture": 90},
			Description: "Fast clippers trade gold in distant ports for food and culture.",
		},
		{
			Name: "Coal Barges", Key: "coal_barges",
			MinAge: "industrial_age", RequiredBld: "harbor", MinCount: 2,
			TicksPerRun: 36,
			Export:      map[string]float64{"coal": 300},
			Import:      map[string]float64{"gold": 220, "iron": 150},
			Description: "Barge coal downriver to the foundries for gold and iron.",
		},
		{
			Name: "Cotton Exchange", Key: "cotton_exchange",
			MinAge: "industrial_age", RequiredBld: "stock_exchange", MinCount: 1,
			TicksPerRun: 42,
			Export:      map[string]float64{"gold": 400},
			Import:      map[string]float64{"culture": 200, "knowledge": 150},
			Description: "Pay gold at the cotton exchange for culture and knowledge.",
		},
		{
			Name: "Steamship Line", Key: "steamship_line",
			MinAge: "industrial_age", RequiredBld: "harbor_authority", MinCount: 2,
			TicksPerRun: 47,
			Export:      map[string]float64{"steel": 250, "coal": 200},
			Import:      map[string]float64{"gold": 900},
			Description: "A steamship line hauls steel and coal across the ocean for gold.",
		},
		{
			Name: "Rail Freight", Key: "rail_freight",
			MinAge: "industrial_age", RequiredBld: "iron_works_complex", MinCount: 1,
			TicksPerRun: 31,
			Export:      map[string]float64{"iron": 200},
			Import:      map[string]float64{"gold": 100, "coal": 50},
			Description: "Ship iron by rail for gold and coal.",
		},
		{
			Name: "Oil Pipeline", Key: "oil_pipeline",
			MinAge: "victorian_age", RequiredBld: "oil_derrick", MinCount: 2,
			TicksPerRun: 39,
			Export:      map[string]float64{"oil": 100},
			Import:      map[string]float64{"gold": 300},
			Description: "Pipe oil to the refineries for gold.",
		},
		{
			Name: "Power Exchange", Key: "power_exchange",
			MinAge: "electric_age", RequiredBld: "power_station", MinCount: 1,
			TicksPerRun: 26,
			Export:      map[string]float64{"electricity": 500},
			Import:      map[string]float64{"gold": 200},
			Description: "Sell surplus electricity on the grid for gold.",
		},
		{
			Name: "Data Trade", Key: "data_trade",
			MinAge: "information_age", RequiredBld: "server_farm", MinCount: 1,
			TicksPerRun: 26,
			Export:      map[string]float64{"data": 100},
			Import:      map[string]float64{"gold": 500},
			Description: "Sell data on digital marketplaces for gold.",
		},
		{
			Name: "Crypto Market", Key: "crypto_market",
			MinAge: "cyberpunk_age", RequiredBld: "black_market", MinCount: 1,
			TicksPerRun: 21,
			Export:      map[string]float64{"crypto": 50},
			Import:      map[string]float64{"gold": 1000},
			Description: "Sell crypto on underground exchanges for gold.",
		},
		{
			Name: "Fusion Export", Key: "fusion_export",
			MinAge: "fusion_age", RequiredBld: "fusion_reactor", MinCount: 1,
			TicksPerRun: 31,
			Export:      map[string]float64{"electricity": 200},
			Import:      map[string]float64{"gold": 1000},
			Description: "Sell fusion electricity to nearby civilizations for gold.",
		},
		{
			Name: "Warp Commerce", Key: "warp_commerce",
			MinAge: "interstellar_age", RequiredBld: "warp_drive_plant", MinCount: 1,
			TicksPerRun: 39,
			Export:      map[string]float64{"gold": 500},
			Import:      map[string]float64{"dark_matter": 200},
			Description: "Send gold across the warp gates for dark matter.",
		},
		{
			Name: "Stellar Freight", Key: "stellar_exchange",
			MinAge: "galactic_age", RequiredBld: "galactic_trade_hub", MinCount: 1,
			TicksPerRun: 52,
			Export:      map[string]float64{"dark_matter": 100},
			Import:      map[string]float64{"gold": 2000},
			Description: "Ship dark matter between star systems for gold.",
		},
		{
			Name: "Quantum Trade", Key: "quantum_trade",
			MinAge: "quantum_age", RequiredBld: "reality_processor", MinCount: 1,
			TicksPerRun: 26,
			Export:      map[string]float64{"quantum_flux": 50},
			Import:      map[string]float64{"gold": 5000},
			Description: "Send quantum flux across dimensional boundaries for gold.",
		},
	}
	return defs
}

// TradeRouteByKey returns trade routes keyed by route key
func TradeRouteByKey() map[string]TradeRouteDef {
	out := make(map[string]TradeRouteDef)
	for _, def := range BaseTradeRoutes() {
		out[def.Key] = def
	}
	return out
}

// BaseFactions returns all NPC civilization definitions. The roster of 11 civs
// spans all 7 epochs: the two founding civs are discoverable in the early eras,
// with more civilizations encountered each epoch as the player advances.
//
// The original 6 factions are preserved (merchant_guild, artisan_league,
// tech_consortium, shadow_syndicate, stellar_federation, quantum_collective)
// and are now the backbone of the roster, each given a personality + backstory.
func BaseFactions() []FactionDef {
	return []FactionDef{
		// --- Stone Era (founding civ) ---
		{
			Name: "Riverlands Tribes", Key: "riverlands_tribes",
			MinAge: "bronze_age", Specialty: "food", TradeBonus: 0.15,
			Personality: "peaceful", Strength: 1,
			Backstory:   "Fisherfolk and farmers of the great delta, generous with grain and slow to anger. They remember every neighbor who once shared a harvest.",
			Description: "Settled farming clans who prize hospitality above all.",
		},
		// --- Iron Era (founding civ) ---
		{
			Name: "Ironhold Clans", Key: "ironhold_clans",
			MinAge: "medieval_age", Specialty: "iron", TradeBonus: 0.20,
			Personality: "aggressive", Strength: 3,
			Backstory:   "Mountain smiths and raiders who measure honor in steel. They respect strength and despise weakness. Cross them and the war-horns sound.",
			Description: "Warlike highland clans forged around the anvil.",
		},
		// --- Steel Era ---
		{
			Name: "Merchant Guild", Key: "merchant_guild",
			MinAge: "colonial_age", Specialty: "gold", TradeBonus: 0.20,
			Personality: "mercantile", Strength: 2,
			Backstory:   "A continent-spanning cartel of traders and financiers whose loyalty follows the ledger. Trade with them often and their coffers open to you.",
			Description: "A powerful guild of traders and financiers.",
		},
		{
			Name: "Artisan League", Key: "artisan_league",
			MinAge: "industrial_age", Specialty: "culture", TradeBonus: 0.15,
			Personality: "peaceful", Strength: 1,
			Backstory:   "Master craftspeople and cultural preservationists who lend their guild-hands to allies they admire, asking only that beauty be protected.",
			Description: "Master craftspeople and cultural preservationists.",
		},
		// --- Electric Era ---
		{
			Name: "Atomic Directorate", Key: "atomic_directorate",
			MinAge: "atomic_age", Specialty: "steel", TradeBonus: 0.20,
			Personality: "isolationist", Strength: 4,
			Backstory:   "A secretive technocracy behind sealed borders, hoarding reactor science. They neither court nor provoke. They simply wait and watch.",
			Description: "An insular technocracy guarding the secrets of the atom.",
		},
		// --- Digital Era ---
		{
			Name: "Tech Consortium", Key: "tech_consortium",
			MinAge: "information_age", Specialty: "data", TradeBonus: 0.20,
			Personality: "mercantile", Strength: 2,
			Backstory:   "A coalition of technology firms that trades in information itself. Data is their currency and your trade volume buys their goodwill.",
			Description: "A coalition of technology companies and innovators.",
		},
		// --- Neon Era ---
		{
			Name: "Shadow Syndicate", Key: "shadow_syndicate",
			MinAge: "cyberpunk_age", Specialty: "crypto", TradeBonus: 0.25,
			Personality: "aggressive", Strength: 3,
			Backstory:   "An underground network dealing in digital currencies and quieter goods. They lend muscle to friends and launch data raids on enemies.",
			Description: "An underground network dealing in digital currencies.",
		},
		{
			Name: "Plasma Nomads", Key: "plasma_nomads",
			MinAge: "fusion_age", Specialty: "plasma", TradeBonus: 0.22,
			Personality: "peaceful", Strength: 2,
			Backstory:   "Wandering fusion-ship caravans with no homeworld and no grudges. To those who welcome them they gift willing crews and bright plasma.",
			Description: "Stateless caravans drifting between the inner worlds.",
		},
		{
			Name: "Stellar Federation", Key: "stellar_federation",
			MinAge: "space_age", Specialty: "dark_matter", TradeBonus: 0.20,
			Personality: "isolationist", Strength: 4,
			Backstory:   "An interstellar alliance of spacefaring civilizations bound by strict non-interference. Hard to befriend, harder still to anger.",
			Description: "An interstellar alliance of spacefaring civilizations.",
		},
		// --- Cosmic Era ---
		{
			Name: "Void Reavers", Key: "void_reavers",
			MinAge: "galactic_age", Specialty: "antimatter", TradeBonus: 0.28,
			Personality: "aggressive", Strength: 5,
			Backstory:   "Antimatter corsairs who strip dead stars and weaker empires alike. They take what they want. Only overwhelming respect keeps their fleets at bay.",
			Description: "Antimatter corsairs feared across the galactic rim.",
		},
		{
			Name: "Quantum Collective", Key: "quantum_collective",
			MinAge: "quantum_age", Specialty: "quantum_flux", TradeBonus: 0.30,
			Personality: "isolationist", Strength: 5,
			Backstory:   "Beings who exist across multiple dimensions, perceiving your civilization as a curiosity. They are indifferent by nature, but their favor bends reality.",
			Description: "Beings who exist across multiple dimensions.",
		},
	}
}

// FactionByKey returns factions keyed by faction key
func FactionByKey() map[string]FactionDef {
	out := make(map[string]FactionDef)
	for _, def := range BaseFactions() {
		out[def.Key] = def
	}
	return out
}
