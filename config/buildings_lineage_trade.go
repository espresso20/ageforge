package config

// buildingsLineageTrade returns lineages 5-9:
// knowledge, faith, military, trade, engineering.
// Merged into newProductionBuildings() via init — see buildings_new_merge.go.
func buildingsLineageTrade() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 8 — TRADE (lineageKey: "trade", domain: "trade", output: "gold")
	// starts at bronze_age (tier 0)
	// rate = 0.05 * 2^tier  CostScale: 1.40  Category: "production"
	// =========================================================================

	// tier 0 — bronze_age  rate=0.05
	b = append(b, BuildingDef{
		Name: "Market", Key: "market", Category: "production",
		BaseCost:    map[string]float64{"wood": 5900, "stone": 4100, "iron": 1200},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 5.47}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "A trading market for barter and coin.",
		LineageKey:  "trade", LineageTier: 0, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 3,
		OutputResource: "gold",
	})
	// tier 1 — iron_age  rate=0.10
	b = append(b, BuildingDef{
		Name: "Trading Post", Key: "trading_post", Category: "production",
		// note: no gold. The trading post is the Iron Age's only gold
		// producer and its only trade building (the market needs one), and
		// the Bronze Age market is optional, so a gold price locked players
		// out of gold for the whole age. The Payback Rule re-derives its rate.
		BaseCost:    map[string]float64{"stone": 16000, "marble": 16000, "iron": 7500, "iron_ore": 7500},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 23.9}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "A regional trading post.",
		LineageKey:  "trade", LineageTier: 1, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 3,
		OutputResource: "gold",
	})
	// tier 2 — classical_age  rate=0.20
	b = append(b, BuildingDef{
		Name: "Merchant Quarter", Key: "merchant_quarter", Category: "production",
		BaseCost:    map[string]float64{"stone": 105000, "marble": 105000, "gold": 88000, "iron": 29500, "iron_ore": 29500},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 112}},
		BuildTicks:  600,
		RequiredAge: "classical_age",
		Description: "An urban merchant district.",
		LineageKey:  "trade", LineageTier: 2, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 4,
		OutputResource: "gold",
	})
	// tier 3 — medieval_age  rate=0.40
	b = append(b, BuildingDef{
		Name: "Guildhall", Key: "guildhall", Category: "production",
		BaseCost:    map[string]float64{"stone": 1.91 * Million, "gold": 653000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 405}},
		BuildTicks:  1200,
		RequiredAge: "medieval_age",
		Description: "Merchant guilds organize regional trade.",
		LineageKey:  "trade", LineageTier: 3, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 4,
		OutputResource: "gold",
	})
	// tier 4 — renaissance_age  rate=0.80
	b = append(b, BuildingDef{
		Name: "Exchange", Key: "exchange", Category: "production",
		BaseCost:    map[string]float64{"gold": 6.21 * Million, "steel": 2.27 * Million},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 1380}},
		BuildTicks:  2400,
		RequiredAge: "renaissance_age",
		Description: "A commodity exchange for international trade.",
		LineageKey:  "trade", LineageTier: 4, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 5,
		OutputResource: "gold",
	})
	// tier 5 — colonial_age  rate=1.60
	b = append(b, BuildingDef{
		Name: "Port", Key: "port", Category: "production",
		BaseCost:    map[string]float64{"gold": 23 * Million, "steel": 12 * Million},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 8210}},
		BuildTicks:  3600,
		RequiredAge: "colonial_age",
		Description: "A colonial maritime trade port.",
		LineageKey:  "trade", LineageTier: 5, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 5,
		OutputResource: "gold",
	})
	// tier 6 — industrial_age  rate=3.20
	b = append(b, BuildingDef{
		Name: "Stock Exchange", Key: "stock_exchange", Category: "production",
		BaseCost:    map[string]float64{"steel": 160 * Million, "coal": 59 * Million, "gold": 110 * Million},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 36100}},
		BuildTicks:  3600,
		RequiredAge: "industrial_age",
		Description: "Industrial-era stock exchange.",
		LineageKey:  "trade", LineageTier: 6, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 6,
		OutputResource: "gold",
	})
	// tier 7 — victorian_age  rate=6.40
	b = append(b, BuildingDef{
		Name: "Bank", Key: "bank", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.1 * Billion, "gold": 590 * Million, "iron": 470 * Million},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 128000}},
		BuildTicks:  3600,
		RequiredAge: "victorian_age",
		Description: "A Victorian national bank.",
		LineageKey:  "trade", LineageTier: 7, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 6,
		OutputResource: "gold",
	})
	// tier 8 — electric_age  rate=12.80
	b = append(b, BuildingDef{
		Name: "Financial District", Key: "financial_district", Category: "production",
		BaseCost:    map[string]float64{"steel": 6.5 * Billion, "electricity": 2.6 * Billion, "gold": 4.4 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 768000}},
		BuildTicks:  3600,
		RequiredAge: "electric_age",
		Description: "Electric-age financial district.",
		LineageKey:  "trade", LineageTier: 8, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 7,
		OutputResource: "gold",
	})
	// tier 9 — atomic_age  rate=25.60
	b = append(b, BuildingDef{
		Name: "Corporate HQ", Key: "corporate_hq", Category: "production",
		BaseCost:  map[string]float64{"steel": 32 * Billion, "electricity": 13 * Billion, "gold": 21 * Billion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 2.39 * Million},
		},
		BuildTicks:  3600,
		RequiredAge: "atomic_age",
		Description: "Multinational corporate headquarters.",
		LineageKey:  "trade", LineageTier: 9, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 7,
		OutputResource: "gold",
	})
	// tier 10 — modern_age  rate=51.20
	b = append(b, BuildingDef{
		Name: "Investment Firm", Key: "investment_firm", Category: "production",
		BaseCost:    map[string]float64{"steel": 95 * Billion, "titanium_ore": 95 * Billion, "electricity": 70 * Billion, "data": 7 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 6.16 * Million}},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "Global investment and wealth management.",
		LineageKey:  "trade", LineageTier: 10, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 8,
		OutputResource: "gold",
	})
	// tier 11 — information_age  rate=102.40
	b = append(b, BuildingDef{
		Name: "Venture Hub", Key: "venture_hub", Category: "production",
		BaseCost:    map[string]float64{"electricity": 500 * Billion, "data": 53 * Billion, "gold": 940 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 211 * Million}},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "Digital venture capital hub.",
		LineageKey:  "trade", LineageTier: 11, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 8,
		OutputResource: "gold",
	})
	// tier 12 — digital_age  rate=204.80
	b = append(b, BuildingDef{
		Name: "Crypto Exchange", Key: "crypto_exchange", Category: "production",
		BaseCost:    map[string]float64{"electricity": 2.5 * Trillion, "data": 310 * Billion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 204.8}},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Decentralized digital currency exchange.",
		LineageKey:  "trade", LineageTier: 12, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 10,
		OutputResource: "gold",
	})
	// tier 13 — cyberpunk_age  rate=409.60
	b = append(b, BuildingDef{
		Name: "Black Market Hub", Key: "black_market", Category: "production",
		BaseCost:  map[string]float64{"data": 1.84 * Trillion, "electricity": 18.4 * Trillion},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "crypto", Value: 337 * Million},
			{Type: "production", Target: "gold", Value: 409.6},
		},
		BuildTicks:  3600,
		RequiredAge: "cyberpunk_age",
		Description: "Underground black market network.",
		LineageKey:  "trade", LineageTier: 13, LineKey: "trade_crypto",
		WorkerDomain: "trade", WorkerCapacity: 10,
		OutputResource: "crypto",
	})
	// tier 14 — fusion_age  rate=819.20
	b = append(b, BuildingDef{
		Name: "Energy Exchange", Key: "energy_exchange", Category: "production",
		BaseCost:     map[string]float64{"plasma": 25 * Trillion, "electricity": 35 * Trillion, "dark_matter_crystals": 35 * Trillion, "steel": 94 * Trillion},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "gold", Value: 819.2}},
		BuildTicks:   3600,
		RequiredAge:  "fusion_age",
		RequiredTech: "maglev_transit",
		Description:  "Interplanetary energy trading exchange.",
		LineageKey:   "trade", LineageTier: 14, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 12,
		OutputResource: "gold",
	})
	// tier 15 — space_age  rate=1638.40
	b = append(b, BuildingDef{
		Name: "Asteroid Market", Key: "asteroid_market", Category: "production",
		BaseCost:    map[string]float64{"titanium": 440 * Trillion, "plasma": 210 * Trillion, "electricity": 260 * Trillion, "dark_matter_crystals": 260 * Trillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 1638.4}},
		BuildTicks:  3600,
		RequiredAge: "space_age",
		Description: "Mineral trading hub in the asteroid belt.",
		LineageKey:  "trade", LineageTier: 15, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 12,
		OutputResource: "gold",
	})
	// tier 16 — interstellar_age  rate=3276.80
	b = append(b, BuildingDef{
		Name: "Galactic Trade Hub", Key: "galactic_trade_hub", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 520 * Trillion, "titanium": 4.1 * Quadrillion, "plasma": 2.5 * Quadrillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 3276.8}},
		BuildTicks:  3600,
		RequiredAge: "interstellar_age",
		Description: "Interstellar trade network hub.",
		LineageKey:  "trade", LineageTier: 16, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 15,
		OutputResource: "gold",
	})
	// tier 17 — galactic_age  rate=6553.60
	b = append(b, BuildingDef{
		Name: "Stellar Exchange", Key: "stellar_exchange", Category: "production",
		BaseCost:    map[string]float64{"antimatter": Quadrillion, "dark_matter": 5.2 * Quadrillion, "titanium": 26 * Quadrillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 6553.6}},
		BuildTicks:  3600,
		RequiredAge: "galactic_age",
		Description: "Galaxy-spanning stellar exchange.",
		LineageKey:  "trade", LineageTier: 17, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 15,
		OutputResource: "gold",
	})
	// tier 18 — quantum_age  rate=13107.20
	b = append(b, BuildingDef{
		Name: "Probability Market", Key: "probability_market", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 1.1 * Quadrillion, "antimatter": 350 * Quadrillion, "dark_matter": 280 * Quadrillion},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 13107.2}},
		BuildTicks:  3600,
		RequiredAge: "quantum_age",
		Description: "Trades across all probable timelines.",
		LineageKey:  "trade", LineageTier: 18, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 18,
		OutputResource: "gold",
	})
	// tier 19 — transcendent_age  rate=26214.40
	b = append(b, BuildingDef{
		Name: "Omniversal Bazaar", Key: "omniversal_bazaar", Category: "production",
		BaseCost:     map[string]float64{"quantum_flux": 11 * Quadrillion, "antimatter": 3.5 * Quintillion, "dark_matter": 2.8 * Quintillion},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "gold", Value: 26214.4}},
		BuildTicks:   3600,
		RequiredAge:  "transcendent_age",
		RequiredTech: "omniversal_exchange",
		Description:  "Omniversal trading bazaar beyond spacetime.",
		LineageKey:   "trade", LineageTier: 19, LineKey: "trade_gold",
		WorkerDomain: "trade", WorkerCapacity: 20,
		OutputResource: "gold",
	})

	return b
}
