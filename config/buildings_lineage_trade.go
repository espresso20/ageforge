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
		BaseCost:    map[string]float64{"wood": 1000, "stone": 700, "iron": 200},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.05}},
		BuildTicks:  150,
		RequiredAge: "bronze_age",
		Description: "A trading market for barter and coin. +0.05 gold/tick (3 workers).",
		LineageKey:  "trade", LineageTier: 0,
		WorkerDomain: "trade", WorkerCapacity: 3,
		EpochKey: "stone_era", OutputResource: "gold",
	})
	// tier 1 — iron_age  rate=0.10
	b = append(b, BuildingDef{
		Name: "Trading Post", Key: "trading_post", Category: "production",
		BaseCost:    map[string]float64{"stone": 5500, "iron": 2500, "gold": 1500},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.10}},
		BuildTicks:  300,
		RequiredAge: "iron_age",
		Description: "A regional trading post. +0.10 gold/tick (3 workers).",
		LineageKey:  "trade", LineageTier: 1,
		WorkerDomain: "trade", WorkerCapacity: 3,
		EpochKey: "iron_era", OutputResource: "gold",
	})
	// tier 2 — classical_age  rate=0.20
	b = append(b, BuildingDef{
		Name: "Merchant Quarter", Key: "merchant_quarter", Category: "production",
		BaseCost:    map[string]float64{"stone": 35000, "gold": 15000, "iron": 10000},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.20}},
		BuildTicks:  3000,
		RequiredAge: "classical_age",
		Description: "An urban merchant district. +0.20 gold/tick (4 workers).",
		LineageKey:  "trade", LineageTier: 2,
		WorkerDomain: "trade", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "gold",
	})
	// tier 3 — medieval_age  rate=0.40
	b = append(b, BuildingDef{
		Name: "Guildhall", Key: "guildhall", Category: "production",
		BaseCost:    map[string]float64{"stone": 200000, "gold": 70000, "knowledge": 20000},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.40}},
		BuildTicks:  6000,
		RequiredAge: "medieval_age",
		Description: "Merchant guild organises regional trade. +0.40 gold/tick (4 workers).",
		LineageKey:  "trade", LineageTier: 3,
		WorkerDomain: "trade", WorkerCapacity: 4,
		EpochKey: "iron_era", OutputResource: "gold",
	})
	// tier 4 — renaissance_age  rate=0.80
	b = append(b, BuildingDef{
		Name: "Exchange", Key: "exchange", Category: "production",
		BaseCost:    map[string]float64{"gold": 700000, "steel": 250000, "knowledge": 100000},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 0.80}},
		BuildTicks:  12000,
		RequiredAge: "renaissance_age",
		Description: "A commodity exchange for international trade. +0.80 gold/tick (5 workers).",
		LineageKey:  "trade", LineageTier: 4,
		WorkerDomain: "trade", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "gold",
	})
	// tier 5 — colonial_age  rate=1.60
	b = append(b, BuildingDef{
		Name: "Port", Key: "port", Category: "production",
		BaseCost:    map[string]float64{"gold": 4e6, "steel": 2e6},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 1.60}},
		BuildTicks:  18000,
		RequiredAge: "colonial_age",
		Description: "A colonial maritime trade port. +1.60 gold/tick (5 workers).",
		LineageKey:  "trade", LineageTier: 5,
		WorkerDomain: "trade", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "gold",
	})
	// tier 6 — industrial_age  rate=3.20
	b = append(b, BuildingDef{
		Name: "Stock Exchange", Key: "stock_exchange", Category: "production",
		BaseCost:    map[string]float64{"steel": 28e6, "coal": 10e6, "gold": 18e6},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 3.20}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Industrial-era stock exchange. +3.20 gold/tick (6 workers).",
		LineageKey:  "trade", LineageTier: 6,
		WorkerDomain: "trade", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "gold",
	})
	// tier 7 — victorian_age  rate=6.40
	b = append(b, BuildingDef{
		Name: "Bank", Key: "bank", Category: "production",
		BaseCost:    map[string]float64{"steel": 190e6, "gold": 100e6, "iron": 80e6},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 6.40}},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "A Victorian national bank. +6.40 gold/tick (6 workers).",
		LineageKey:  "trade", LineageTier: 7,
		WorkerDomain: "trade", WorkerCapacity: 6,
		EpochKey: "electric_era", OutputResource: "gold",
	})
	// tier 8 — electric_age  rate=12.80
	b = append(b, BuildingDef{
		Name: "Financial District", Key: "financial_district", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.1e9, "electricity": 450e6, "gold": 750e6},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 12.80}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "Electric-age financial district. +12.80 gold/tick (7 workers).",
		LineageKey:  "trade", LineageTier: 8,
		WorkerDomain: "trade", WorkerCapacity: 7,
		EpochKey: "electric_era", OutputResource: "gold",
	})
	// tier 9 — atomic_age  rate=25.60
	b = append(b, BuildingDef{
		Name: "Corporate HQ", Key: "corporate_hq", Category: "production",
		BaseCost:    map[string]float64{"steel": 5.5e9, "electricity": 2.2e9, "gold": 3.5e9},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 25.60}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Multinational corporate headquarters. +25.60 gold/tick (7 workers).",
		LineageKey:  "trade", LineageTier: 9,
		WorkerDomain: "trade", WorkerCapacity: 7,
		EpochKey: "electric_era", OutputResource: "gold",
	})
	// tier 10 — modern_age  rate=51.20
	b = append(b, BuildingDef{
		Name: "Investment Firm", Key: "investment_firm", Category: "production",
		BaseCost:    map[string]float64{"steel": 32e9, "electricity": 12e9, "data": 1.2e9},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 51.20}},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Global investment and wealth management. +51.20 gold/tick (8 workers).",
		LineageKey:  "trade", LineageTier: 10,
		WorkerDomain: "trade", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "gold",
	})
	// tier 11 — information_age  rate=102.40
	b = append(b, BuildingDef{
		Name: "Venture Hub", Key: "venture_hub", Category: "production",
		BaseCost:    map[string]float64{"electricity": 85e9, "data": 9e9, "gold": 160e9},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 102.40}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "Digital venture capital hub. +102.40 gold/tick (8 workers).",
		LineageKey:  "trade", LineageTier: 11,
		WorkerDomain: "trade", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "gold",
	})
	// tier 12 — digital_age  rate=204.80
	b = append(b, BuildingDef{
		Name: "Crypto Exchange", Key: "crypto_exchange", Category: "production",
		BaseCost:    map[string]float64{"electricity": 420e9, "data": 52e9, "crypto": 100e9},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 204.80}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Decentralised digital currency exchange. +204.80 gold/tick (10 workers).",
		LineageKey:  "trade", LineageTier: 12,
		WorkerDomain: "trade", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "gold",
	})
	// tier 13 — cyberpunk_age  rate=409.60
	b = append(b, BuildingDef{
		Name: "Black Market Hub", Key: "black_market", Category: "production",
		BaseCost:    map[string]float64{"data": 200e9, "crypto": 1.05e12, "electricity": 2.1e12},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 409.60}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Underground black market network. +409.60 gold/tick (10 workers).",
		LineageKey:  "trade", LineageTier: 13,
		WorkerDomain: "trade", WorkerCapacity: 10,
		EpochKey: "neon_era", OutputResource: "gold",
	})
	// tier 14 — fusion_age  rate=819.20
	b = append(b, BuildingDef{
		Name: "Energy Exchange", Key: "energy_exchange", Category: "production",
		BaseCost:    map[string]float64{"plasma": 4.2e12, "electricity": 12e12, "steel": 16e12},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 819.20}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Interplanetary energy trading exchange. +819.20 gold/tick (12 workers).",
		LineageKey:  "trade", LineageTier: 14,
		WorkerDomain: "trade", WorkerCapacity: 12,
		EpochKey: "neon_era", OutputResource: "gold",
	})
	// tier 15 — space_age  rate=1638.40
	b = append(b, BuildingDef{
		Name: "Asteroid Market", Key: "asteroid_market", Category: "production",
		BaseCost:    map[string]float64{"titanium": 75e12, "plasma": 35e12, "electricity": 88e12},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 1638.40}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Mineral trading hub in the asteroid belt. +1638.40 gold/tick (12 workers).",
		LineageKey:  "trade", LineageTier: 15,
		WorkerDomain: "trade", WorkerCapacity: 12,
		EpochKey: "neon_era", OutputResource: "gold",
	})
	// tier 16 — interstellar_age  rate=3276.80
	b = append(b, BuildingDef{
		Name: "Galactic Trade Hub", Key: "galactic_trade_hub", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 88e12, "titanium": 700e12, "plasma": 430e12},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 3276.80}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Interstellar trade network hub. +3276.80 gold/tick (15 workers).",
		LineageKey:  "trade", LineageTier: 16,
		WorkerDomain: "trade", WorkerCapacity: 15,
		EpochKey: "cosmic_era", OutputResource: "gold",
	})
	// tier 17 — galactic_age  rate=6553.60
	b = append(b, BuildingDef{
		Name: "Stellar Exchange", Key: "stellar_exchange", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 175e12, "dark_matter": 880e12, "titanium": 4.4e15},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 6553.60}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Galaxy-spanning stellar exchange. +6553.60 gold/tick (15 workers).",
		LineageKey:  "trade", LineageTier: 17,
		WorkerDomain: "trade", WorkerCapacity: 15,
		EpochKey: "cosmic_era", OutputResource: "gold",
	})
	// tier 18 — quantum_age  rate=13107.20
	b = append(b, BuildingDef{
		Name: "Probability Market", Key: "probability_market", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 195e12, "antimatter": 59e15, "dark_matter": 48e15},
		CostScale:   1.40,
		Effects:     []Effect{{Type: "production", Target: "gold", Value: 13107.20}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Trades across all probable timelines. +13107.20 gold/tick (18 workers).",
		LineageKey:  "trade", LineageTier: 18,
		WorkerDomain: "trade", WorkerCapacity: 18,
		EpochKey: "cosmic_era", OutputResource: "gold",
	})

	return b
}
