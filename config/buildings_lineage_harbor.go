package config

// buildingsLineageHarbor returns the HARBOR lineage — a small maritime-trade
// infrastructure chain that boosts passive trade-route income.
//
// Unlike the main "trade" lineage (markets/banks that simply produce gold),
// harbours improve the throughput of *trade routes*: every built harbour adds a
// fractional "trade_route_income" bonus that the engine applies to the imports
// of every active route (see game/trade.go, harborRouteBonus). They also produce
// a little gold themselves so an idle player still gets value from them.
//
// 5 tiers across the colonial → digital band — exactly the colonial→industrial
// gap the Trade Expansion targets, extended a couple of ages so the bonus keeps
// scaling. Worker domain is "trade" (same recruits as markets).
//
//	tier 0 harbor             colonial_age      +5%  route income
//	tier 1 harbor_authority   industrial_age    +10% route income
//	tier 2 seaport            modern_age        +15% route income
//	tier 3 container_terminal information_age    +20% route income
//	tier 4 logistics_hub      digital_age       +25% route income
//
// Merged into newProductionBuildings() via NewProductionBuildings().
func buildingsLineageHarbor() []BuildingDef {
	b := []BuildingDef{}

	// tier 0 — colonial_age
	b = append(b, BuildingDef{
		Name: "Harbor", Key: "harbor", Category: "production",
		BaseCost:  map[string]float64{"wood": 3e6, "stone": 2.5e6, "gold": 1.5e6},
		CostScale: 1.45,
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 0.80},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.05},
		},
		BuildTicks:  3000,
		RequiredAge: "colonial_age",
		Description: "A working harbour for trade ships. +0.80 gold/tick and +5% trade-route income (4 workers).",
		LineageKey:  "harbor", LineageTier: 0,
		WorkerDomain: "trade", WorkerCapacity: 4,
		EpochKey: "steel_era", OutputResource: "gold",
	})
	// tier 1 — industrial_age
	b = append(b, BuildingDef{
		Name: "Harbor Authority", Key: "harbor_authority", Category: "production",
		BaseCost:  map[string]float64{"steel": 22e6, "coal": 8e6, "gold": 14e6},
		CostScale: 1.45,
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 1.60},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.10},
		},
		BuildTicks:  3300,
		RequiredAge: "industrial_age",
		Description: "Industrial port authority coordinating dock traffic. +1.60 gold/tick and +10% trade-route income (5 workers).",
		LineageKey:  "harbor", LineageTier: 1,
		WorkerDomain: "trade", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "gold",
	})
	// tier 2 — modern_age
	b = append(b, BuildingDef{
		Name: "Seaport", Key: "seaport", Category: "production",
		BaseCost:  map[string]float64{"steel": 26e9, "electricity": 9e9, "gold": 600e6},
		CostScale: 1.45,
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 25.60},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.15},
		},
		BuildTicks:  3600,
		RequiredAge: "modern_age",
		Description: "A deep-water seaport handling global cargo. +25.60 gold/tick and +15% trade-route income (6 workers).",
		LineageKey:  "harbor", LineageTier: 2,
		WorkerDomain: "trade", WorkerCapacity: 6,
		EpochKey: "digital_era", OutputResource: "gold",
	})
	// tier 3 — information_age
	b = append(b, BuildingDef{
		Name: "Container Terminal", Key: "container_terminal", Category: "production",
		BaseCost:  map[string]float64{"electricity": 70e9, "data": 7e9, "gold": 130e9},
		CostScale: 1.45,
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 51.20},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.20},
		},
		BuildTicks:  3600,
		RequiredAge: "information_age",
		Description: "Automated container terminal moving freight at scale. +51.20 gold/tick and +20% trade-route income (8 workers).",
		LineageKey:  "harbor", LineageTier: 3,
		WorkerDomain: "trade", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "gold",
	})
	// tier 4 — digital_age
	b = append(b, BuildingDef{
		Name: "Logistics Hub", Key: "logistics_hub", Category: "production",
		BaseCost:  map[string]float64{"electricity": 350e9, "data": 45e9, "crypto": 80e9},
		CostScale: 1.45,
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 102.40},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.25},
		},
		BuildTicks:  3600,
		RequiredAge: "digital_age",
		Description: "Algorithmic logistics hub optimising every shipment. +102.40 gold/tick and +25% trade-route income (10 workers).",
		LineageKey:  "harbor", LineageTier: 4,
		WorkerDomain: "trade", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "gold",
	})

	return b
}
