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
		BaseCost:  map[string]float64{"wood": 2.4e07, "stone": 2e07, "gold": 1.2e07},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 26900},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.05},
		},
		BuildTicks:   3000,
		RequiredAge:  "colonial_age",
		RequiredTech: "mercantilism",
		Description:  "A working harbor for trade ships.",
		LineageKey:   "harbor", LineageTier: 0,
		WorkerDomain: "trade", WorkerCapacity: 4,
		EpochKey: "steel_era", OutputResource: "gold",
	})
	// tier 1 — industrial_age
	b = append(b, BuildingDef{
		Name: "Harbor Authority", Key: "harbor_authority", Category: "production",
		BaseCost:  map[string]float64{"steel": 1.8e08, "coal": 6.4e07, "gold": 1.1e08},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 38500},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.1},
		},
		BuildTicks:  3300,
		RequiredAge: "industrial_age",
		Description: "Industrial port authority coordinating dock traffic.",
		LineageKey:  "harbor", LineageTier: 1,
		WorkerDomain: "trade", WorkerCapacity: 5,
		EpochKey: "steel_era", OutputResource: "gold",
	})
	// tier 2 — modern_age
	b = append(b, BuildingDef{
		Name: "Seaport", Key: "seaport", Category: "production",
		BaseCost:  map[string]float64{"steel": 2.1e11, "electricity": 7.2e10, "gold": 4.8e09},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 4600000},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.15},
		},
		BuildTicks:   3600,
		RequiredAge:  "modern_age",
		RequiredTech: "containerization",
		Description:  "A deep-water seaport handling global cargo.",
		LineageKey:   "harbor", LineageTier: 2,
		WorkerDomain: "trade", WorkerCapacity: 6,
		EpochKey: "digital_era", OutputResource: "gold",
	})
	// tier 3 — information_age
	b = append(b, BuildingDef{
		Name: "Container Terminal", Key: "container_terminal", Category: "production",
		BaseCost:  map[string]float64{"electricity": 5.6e11, "data": 5.6e10, "gold": 1e12},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 2.28e08},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.2},
		},
		BuildTicks:   3600,
		RequiredAge:  "information_age",
		RequiredTech: "e_commerce",
		Description:  "Automated container terminal moving freight at scale.",
		LineageKey:   "harbor", LineageTier: 3,
		WorkerDomain: "trade", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "gold",
	})
	// tier 4 — digital_age
	b = append(b, BuildingDef{
		Name: "Logistics Hub", Key: "logistics_hub", Category: "production",
		BaseCost:  map[string]float64{"electricity": 2.8e12, "data": 3.6e11},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 102.4},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.25},
		},
		BuildTicks:   3600,
		RequiredAge:  "digital_age",
		RequiredTech: "automated_logistics",
		Description:  "Algorithmic logistics hub routing every shipment.",
		LineageKey:   "harbor", LineageTier: 4,
		WorkerDomain: "trade", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "gold",
	})

	return b
}
