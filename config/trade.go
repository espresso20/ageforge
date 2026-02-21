package config

// ExchangeRateDef defines base exchange rates between resource pairs
type ExchangeRateDef struct {
	From     string  // resource key
	To       string  // resource key
	BaseRate float64 // how much "To" you get per 1 "From"
	MinAge   string  // earliest age this exchange is available
}

// TradeRouteDef defines a passive trade route unlocked by buildings
type TradeRouteDef struct {
	Name        string
	Key         string
	MinAge      string
	RequiredBld string             // building key required (market, port, etc.)
	MinCount    int                // minimum building count needed
	TicksPerRun int                // ticks per trade cycle
	Export      map[string]float64 // resources consumed per cycle
	Import      map[string]float64 // resources gained per cycle
	Description string
}

// FactionDef defines an NPC faction
type FactionDef struct {
	Name        string
	Key         string
	MinAge      string
	Specialty   string  // resource key they're good at
	TradeBonus  float64 // fractional bonus on trades with them when allied
	Description string
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

		// Information Age
		{From: "data", To: "gold", BaseRate: 5.0, MinAge: "information_age"},
		{From: "gold", To: "data", BaseRate: 0.15, MinAge: "information_age"},

		// Cyberpunk
		{From: "crypto", To: "gold", BaseRate: 20.0, MinAge: "cyberpunk_age"},
		{From: "gold", To: "crypto", BaseRate: 0.04, MinAge: "cyberpunk_age"},

		// Space Age+
		{From: "dark_matter", To: "gold", BaseRate: 50.0, MinAge: "space_age"},
		{From: "quantum_flux", To: "gold", BaseRate: 100.0, MinAge: "quantum_age"},
	}
}

// ExchangeRateByKey returns a map from "from:to" -> ExchangeRateDef
func ExchangeRateByKey() map[string]ExchangeRateDef {
	out := make(map[string]ExchangeRateDef)
	for _, def := range BaseExchangeRates() {
		out[def.From+":"+def.To] = def
	}
	return out
}

// BaseTradeRoutes returns all trade route definitions
func BaseTradeRoutes() []TradeRouteDef {
	return []TradeRouteDef{
		{
			Name: "Local Barter", Key: "local_barter",
			MinAge: "bronze_age", RequiredBld: "market", MinCount: 1,
			TicksPerRun: 10,
			Export:       map[string]float64{"food": 10},
			Import:       map[string]float64{"wood": 8},
			Description:  "Trade surplus food for wood with nearby villages.",
		},
		{
			Name: "Stone Trade", Key: "stone_trade",
			MinAge: "iron_age", RequiredBld: "market", MinCount: 2,
			TicksPerRun: 12,
			Export:       map[string]float64{"wood": 15},
			Import:       map[string]float64{"stone": 12},
			Description:  "Exchange timber for quarried stone.",
		},
		{
			Name: "Gold Caravan", Key: "gold_caravan",
			MinAge: "classical_age", RequiredBld: "market", MinCount: 3,
			TicksPerRun: 15,
			Export:       map[string]float64{"stone": 50},
			Import:       map[string]float64{"gold": 5},
			Description:  "Send stone caravans in exchange for gold.",
		},
		{
			Name: "Silk Road", Key: "silk_road",
			MinAge: "medieval_age", RequiredBld: "market", MinCount: 2,
			TicksPerRun: 20,
			Export:       map[string]float64{"gold": 30},
			Import:       map[string]float64{"culture": 80},
			Description:  "Trade along the fabled Silk Road for cultural riches.",
		},
		{
			Name: "Spice Trade", Key: "spice_trade",
			MinAge: "colonial_age", RequiredBld: "port", MinCount: 1,
			TicksPerRun: 18,
			Export:       map[string]float64{"gold": 100},
			Import:       map[string]float64{"food": 200, "culture": 50},
			Description:  "Import exotic spices and cultural goods from distant lands.",
		},
		{
			Name: "Colonial Exports", Key: "colonial_exports",
			MinAge: "colonial_age", RequiredBld: "port", MinCount: 2,
			TicksPerRun: 15,
			Export:       map[string]float64{"food": 500},
			Import:       map[string]float64{"gold": 150},
			Description:  "Export food supplies to colonial settlements.",
		},
		{
			Name: "Rail Freight", Key: "rail_freight",
			MinAge: "industrial_age", RequiredBld: "train_station", MinCount: 1,
			TicksPerRun: 12,
			Export:       map[string]float64{"iron": 200},
			Import:       map[string]float64{"gold": 100, "coal": 50},
			Description:  "Ship iron ore by rail for gold and coal.",
		},
		{
			Name: "Oil Pipeline", Key: "oil_pipeline",
			MinAge: "victorian_age", RequiredBld: "oil_well", MinCount: 2,
			TicksPerRun: 15,
			Export:       map[string]float64{"oil": 100},
			Import:       map[string]float64{"gold": 300},
			Description:  "Pipe crude oil to refineries for gold.",
		},
		{
			Name: "Power Exchange", Key: "power_exchange",
			MinAge: "electric_age", RequiredBld: "power_grid", MinCount: 1,
			TicksPerRun: 10,
			Export:       map[string]float64{"electricity": 500},
			Import:       map[string]float64{"gold": 200},
			Description:  "Sell surplus electricity on the power grid.",
		},
		{
			Name: "Data Trade", Key: "data_trade",
			MinAge: "information_age", RequiredBld: "fiber_hub", MinCount: 1,
			TicksPerRun: 10,
			Export:       map[string]float64{"data": 100},
			Import:       map[string]float64{"gold": 500},
			Description:  "Monetize data through digital marketplaces.",
		},
		{
			Name: "Crypto Market", Key: "crypto_market",
			MinAge: "cyberpunk_age", RequiredBld: "black_market", MinCount: 1,
			TicksPerRun: 8,
			Export:       map[string]float64{"crypto": 50},
			Import:       map[string]float64{"gold": 1000},
			Description:  "Trade cryptocurrency on underground exchanges.",
		},
		{
			Name: "Fusion Export", Key: "fusion_export",
			MinAge: "fusion_age", RequiredBld: "fusion_reactor", MinCount: 1,
			TicksPerRun: 12,
			Export:       map[string]float64{"electricity": 200},
			Import:       map[string]float64{"gold": 1000},
			Description:  "Export fusion energy to nearby civilizations.",
		},
		{
			Name: "Warp Commerce", Key: "warp_commerce",
			MinAge: "space_age", RequiredBld: "warp_gate", MinCount: 1,
			TicksPerRun: 15,
			Export:       map[string]float64{"gold": 500},
			Import:       map[string]float64{"dark_matter": 200},
			Description:  "Trade across warp gates for exotic matter.",
		},
		{
			Name: "Stellar Exchange", Key: "stellar_exchange",
			MinAge: "galactic_age", RequiredBld: "galactic_hub", MinCount: 1,
			TicksPerRun: 20,
			Export:       map[string]float64{"dark_matter": 100},
			Import:       map[string]float64{"gold": 2000},
			Description:  "Conduct interstellar trade at galactic scale.",
		},
		{
			Name: "Quantum Trade", Key: "quantum_trade",
			MinAge: "quantum_age", RequiredBld: "quantum_computer", MinCount: 1,
			TicksPerRun: 10,
			Export:       map[string]float64{"quantum_flux": 50},
			Import:       map[string]float64{"gold": 5000},
			Description:  "Trade quantum flux across dimensional boundaries.",
		},
	}
}

// TradeRouteByKey returns trade routes keyed by route key
func TradeRouteByKey() map[string]TradeRouteDef {
	out := make(map[string]TradeRouteDef)
	for _, def := range BaseTradeRoutes() {
		out[def.Key] = def
	}
	return out
}

// BaseFactions returns all NPC faction definitions
func BaseFactions() []FactionDef {
	return []FactionDef{
		{
			Name: "Merchant Guild", Key: "merchant_guild",
			MinAge: "colonial_age", Specialty: "gold", TradeBonus: 0.20,
			Description: "A powerful guild of traders and financiers.",
		},
		{
			Name: "Artisan League", Key: "artisan_league",
			MinAge: "industrial_age", Specialty: "culture", TradeBonus: 0.15,
			Description: "Master craftspeople and cultural preservationists.",
		},
		{
			Name: "Tech Consortium", Key: "tech_consortium",
			MinAge: "information_age", Specialty: "data", TradeBonus: 0.20,
			Description: "A coalition of technology companies and innovators.",
		},
		{
			Name: "Shadow Syndicate", Key: "shadow_syndicate",
			MinAge: "cyberpunk_age", Specialty: "crypto", TradeBonus: 0.25,
			Description: "An underground network dealing in digital currencies.",
		},
		{
			Name: "Stellar Federation", Key: "stellar_federation",
			MinAge: "space_age", Specialty: "dark_matter", TradeBonus: 0.20,
			Description: "An interstellar alliance of spacefaring civilizations.",
		},
		{
			Name: "Quantum Collective", Key: "quantum_collective",
			MinAge: "quantum_age", Specialty: "quantum_flux", TradeBonus: 0.30,
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
