package config

// RouteShare is the share of the town's gross production of an exported
// resource that a route at its gate sends out.
const RouteShare = 0.10

// RouteMargin is how much more value a route brings in than it sends out, by
// the price levels deals use.
const RouteMargin = 1.10

// TradeRouteDef defines a trade route. A running route is a flow: at its gate
// (MinCount of RequiredBld) it sends RouteShare of the town's gross
// production of each export and brings in RouteMargin times that value, split
// evenly over its imports. The flow grows with the buildings past the gate. A
// shortage of an export slows the route.
type TradeRouteDef struct {
	Name        string
	Key         string
	MinAge      string   // minimum age to unlock this route
	RequiredBld string   // building key that must be built (e.g. "market", "port")
	MinCount    int      // minimum count of RequiredBld needed
	Export      []string // what the route takes
	Import      []string // what the route brings
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
}

// ValidPersonalities is the closed set of civilization personalities. Used by
// config validation and the diplomacy drift/war logic.
var ValidPersonalities = map[string]bool{
	"aggressive":   true,
	"peaceful":     true,
	"mercantile":   true,
	"isolationist": true,
}

// BaseTradeRoutes returns all trade route definitions.
func BaseTradeRoutes() []TradeRouteDef {
	defs := []TradeRouteDef{
		{
			Name: "Local Barter", Key: "local_barter",
			MinAge: "bronze_age", RequiredBld: "market", MinCount: 1,
			Export:      []string{"food"},
			Import:      []string{"wood"},
			Description: "Send surplus food to nearby villages for wood.",
		},
		{
			Name: "Stone Trade", Key: "stone_trade",
			MinAge: "iron_age", RequiredBld: "market", MinCount: 2,
			Export:      []string{"wood"},
			Import:      []string{"stone"},
			Description: "Send wood to the quarries for stone.",
		},
		{
			Name: "Gold Caravan", Key: "gold_caravan",
			MinAge: "classical_age", RequiredBld: "market", MinCount: 3,
			Export:      []string{"stone"},
			Import:      []string{"gold"},
			Description: "Send stone caravans out for gold.",
		},
		{
			Name: "Silk Road", Key: "silk_road",
			MinAge: "medieval_age", RequiredBld: "market", MinCount: 2,
			Export:      []string{"gold"},
			Import:      []string{"culture"},
			Description: "Send gold along the Silk Road for culture.",
		},
		{
			Name: "Spice Trade", Key: "spice_trade",
			MinAge: "colonial_age", RequiredBld: "port", MinCount: 1,
			Export:      []string{"gold"},
			Import:      []string{"food", "culture"},
			Description: "Pay gold to distant lands for spiced food and culture.",
		},
		{
			Name: "Colonial Exports", Key: "colonial_exports",
			MinAge: "colonial_age", RequiredBld: "port", MinCount: 2,
			Export:      []string{"food"},
			Import:      []string{"gold"},
			Description: "Sell food to colonial settlements for gold.",
		},
		// --- Colonial → Industrial gap fillers (Trade Expansion) ---
		// This band was thin (only spice_trade + colonial_exports). These six
		// routes give the renaissance→industrial player something to do with
		// surplus stone/iron/coal/knowledge and a reason to build harbours.
		{
			Name: "Mercantile Convoy", Key: "mercantile_convoy",
			MinAge: "renaissance_age", RequiredBld: "exchange", MinCount: 1,
			Export:      []string{"stone", "wood"},
			Import:      []string{"gold"},
			Description: "Run armed convoys of stone and wood between city-states for gold.",
		},
		{
			Name: "Triangular Trade", Key: "triangular_trade",
			MinAge: "colonial_age", RequiredBld: "harbor", MinCount: 1,
			Export:      []string{"food", "gold"},
			Import:      []string{"culture", ResearchTarget},
			Description: "A three-way colonial exchange: send food and gold for culture and research.",
		},
		{
			Name: "Tea Clippers", Key: "tea_clippers",
			MinAge: "colonial_age", RequiredBld: "harbor", MinCount: 2,
			Export:      []string{"gold"},
			Import:      []string{"food", "culture"},
			Description: "Fast clippers trade gold in distant ports for food and culture.",
		},
		{
			Name: "Coal Barges", Key: "coal_barges",
			MinAge: "industrial_age", RequiredBld: "harbor", MinCount: 2,
			Export:      []string{"coal"},
			Import:      []string{"gold", "iron"},
			Description: "Barge coal downriver to the foundries for gold and iron.",
		},
		{
			Name: "Cotton Exchange", Key: "cotton_exchange",
			MinAge: "industrial_age", RequiredBld: "stock_exchange", MinCount: 1,
			Export:      []string{"gold"},
			Import:      []string{"culture", ResearchTarget},
			Description: "Pay gold at the cotton exchange for culture and research.",
		},
		{
			Name: "Steamship Line", Key: "steamship_line",
			MinAge: "industrial_age", RequiredBld: "harbor_authority", MinCount: 2,
			Export:      []string{"steel", "coal"},
			Import:      []string{"gold"},
			Description: "A steamship line hauls steel and coal across the ocean for gold.",
		},
		{
			Name: "Rail Freight", Key: "rail_freight",
			MinAge: "industrial_age", RequiredBld: "iron_works_complex", MinCount: 1,
			Export:      []string{"iron"},
			Import:      []string{"gold", "coal"},
			Description: "Ship iron by rail for gold and coal.",
		},
		{
			Name: "Oil Pipeline", Key: "oil_pipeline",
			MinAge: "victorian_age", RequiredBld: "oil_derrick", MinCount: 2,
			Export:      []string{"oil"},
			Import:      []string{"gold"},
			Description: "Pipe oil to the refineries for gold.",
		},
		{
			Name: "Power Exchange", Key: "power_exchange",
			MinAge: "electric_age", RequiredBld: "power_station", MinCount: 1,
			Export:      []string{"electricity"},
			Import:      []string{"gold"},
			Description: "Sell surplus electricity on the grid for gold.",
		},
		{
			Name: "Data Trade", Key: "data_trade",
			MinAge: "information_age", RequiredBld: "server_farm", MinCount: 1,
			Export:      []string{"data"},
			Import:      []string{"gold"},
			Description: "Sell data on digital marketplaces for gold.",
		},
		{
			Name: "Crypto Market", Key: "crypto_market",
			MinAge: "cyberpunk_age", RequiredBld: "black_market", MinCount: 1,
			Export:      []string{"crypto"},
			Import:      []string{"gold"},
			Description: "Sell crypto on underground exchanges for gold.",
		},
		{
			Name: "Fusion Export", Key: "fusion_export",
			MinAge: "fusion_age", RequiredBld: "fusion_reactor", MinCount: 1,
			Export:      []string{"electricity"},
			Import:      []string{"gold"},
			Description: "Sell fusion electricity to nearby civilizations for gold.",
		},
		{
			Name: "Warp Commerce", Key: "warp_commerce",
			MinAge: "interstellar_age", RequiredBld: "warp_drive_plant", MinCount: 1,
			Export:      []string{"gold"},
			Import:      []string{"dark_matter"},
			Description: "Send gold across the warp gates for dark matter.",
		},
		{
			Name: "Stellar Freight", Key: "stellar_exchange",
			MinAge: "galactic_age", RequiredBld: "galactic_trade_hub", MinCount: 1,
			Export:      []string{"dark_matter"},
			Import:      []string{"gold"},
			Description: "Ship dark matter between star systems for gold.",
		},
		{
			Name: "Quantum Trade", Key: "quantum_trade",
			MinAge: "quantum_age", RequiredBld: "reality_processor", MinCount: 1,
			Export:      []string{"quantum_flux"},
			Import:      []string{"gold"},
			Description: "Send quantum flux across dimensional boundaries for gold.",
		},
	}
	return defs
}

// TradeRouteByKey returns trade routes keyed by route key
func TradeRouteByKey() map[string]TradeRouteDef {
	return indexBy(BaseTradeRoutes(), func(def TradeRouteDef) string { return def.Key })
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
			Backstory: "Fisherfolk and farmers of the great delta, generous with grain and slow to anger. They remember every neighbor who once shared a harvest.",
		},
		// --- Iron Era (founding civ) ---
		{
			Name: "Ironhold Clans", Key: "ironhold_clans",
			MinAge: "medieval_age", Specialty: "iron", TradeBonus: 0.20,
			Personality: "aggressive", Strength: 3,
			Backstory: "Mountain smiths and raiders who measure honor in steel. They respect strength and despise weakness. Cross them and the war-horns sound.",
		},
		// --- Steel Era ---
		{
			Name: "Merchant Guild", Key: "merchant_guild",
			MinAge: "colonial_age", Specialty: "gold", TradeBonus: 0.20,
			Personality: "mercantile", Strength: 2,
			Backstory: "A continent-spanning cartel of traders and financiers whose loyalty follows the ledger. Trade with them often and their coffers open to you.",
		},
		{
			Name: "Artisan League", Key: "artisan_league",
			MinAge: "industrial_age", Specialty: "culture", TradeBonus: 0.15,
			Personality: "peaceful", Strength: 1,
			Backstory: "Master craftspeople and cultural preservationists who lend their guild-hands to allies they admire, asking only that beauty be protected.",
		},
		// --- Electric Era ---
		{
			Name: "Atomic Directorate", Key: "atomic_directorate",
			MinAge: "atomic_age", Specialty: "steel", TradeBonus: 0.20,
			Personality: "isolationist", Strength: 4,
			Backstory: "A secretive technocracy behind sealed borders, hoarding reactor science. They neither court nor provoke. They simply wait and watch.",
		},
		// --- Digital Era ---
		{
			Name: "Tech Consortium", Key: "tech_consortium",
			MinAge: "information_age", Specialty: "data", TradeBonus: 0.20,
			Personality: "mercantile", Strength: 2,
			Backstory: "A coalition of technology firms that trades in information itself. Data is their currency and your trade volume buys their goodwill.",
		},
		// --- Neon Era ---
		{
			Name: "Shadow Syndicate", Key: "shadow_syndicate",
			MinAge: "cyberpunk_age", Specialty: "crypto", TradeBonus: 0.25,
			Personality: "aggressive", Strength: 3,
			Backstory: "An underground network dealing in digital currencies and quieter goods. They lend muscle to friends and launch data raids on enemies.",
		},
		{
			Name: "Plasma Nomads", Key: "plasma_nomads",
			MinAge: "fusion_age", Specialty: "plasma", TradeBonus: 0.22,
			Personality: "peaceful", Strength: 2,
			Backstory: "Wandering fusion-ship caravans with no homeworld and no grudges. To those who welcome them they gift willing crews and bright plasma.",
		},
		{
			Name: "Stellar Federation", Key: "stellar_federation",
			MinAge: "space_age", Specialty: "dark_matter", TradeBonus: 0.20,
			Personality: "isolationist", Strength: 4,
			Backstory: "An interstellar alliance of spacefaring civilizations bound by strict non-interference. Hard to befriend, harder still to anger.",
		},
		// --- Cosmic Era ---
		{
			Name: "Void Reavers", Key: "void_reavers",
			MinAge: "galactic_age", Specialty: "antimatter", TradeBonus: 0.28,
			Personality: "aggressive", Strength: 5,
			Backstory: "Antimatter corsairs who strip dead stars and weaker empires alike. They take what they want. Only overwhelming respect keeps their fleets at bay.",
		},
		{
			Name: "Quantum Collective", Key: "quantum_collective",
			MinAge: "quantum_age", Specialty: "quantum_flux", TradeBonus: 0.30,
			Personality: "isolationist", Strength: 5,
			Backstory: "Beings who exist across multiple dimensions, perceiving your civilization as a curiosity. They are indifferent by nature, but their favor bends reality.",
		},
	}
}

// FactionByKey returns factions keyed by faction key
func FactionByKey() map[string]FactionDef {
	return indexBy(BaseFactions(), func(def FactionDef) string { return def.Key })
}

// A route's upgrades (game.routeMult): a level costs RouteUpgradeMinutes of the
// route's level-0 import, times RouteUpgradeGrowth for each level bought, and
// the flow doubles again at each of the levels game.routeDoublings lists.
const (
	RouteUpgradeMinutes = 10
	RouteUpgradeGrowth  = 1.15
)
