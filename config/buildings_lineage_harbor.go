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
		Name: "Harbor", Key: "harbor",
		BaseCost: map[string]float64{"stone": 54.3 * Million, "gold": 54.3 * Million},
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 23300},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.05},
		},
		BuildTicks:     3000,
		RequiredAge:    "colonial_age",
		RequiredTech:   "mercantilism",
		Description:    "A working harbor for trade ships.",
		WorkerCapacity: 4,
	})
	// tier 1 — industrial_age
	b = append(b, BuildingDef{
		Name: "Harbor Authority", Key: "harbor_authority",
		BaseCost: map[string]float64{"steel": 180 * Million, "coal": 64 * Million, "gold": 110 * Million},
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 38500},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.1},
		},
		BuildTicks:     3300,
		RequiredAge:    "industrial_age",
		Description:    "Industrial port authority coordinating dock traffic.",
		WorkerCapacity: 5,
	})
	// tier 2 — modern_age
	b = append(b, BuildingDef{
		Name: "Seaport", Key: "seaport",
		BaseCost: map[string]float64{"steel": 105 * Billion, "titanium_ore": 105 * Billion, "electricity": 72 * Billion, "gold": 4.8 * Billion},
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 4.37 * Million},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.15},
		},
		BuildTicks:     3600,
		RequiredAge:    "modern_age",
		RequiredTech:   "containerization",
		Description:    "A deep-water seaport handling global cargo.",
		WorkerCapacity: 6,
	})
	// tier 3 — information_age
	b = append(b, BuildingDef{
		Name: "Container Terminal", Key: "container_terminal",
		BaseCost: map[string]float64{"electricity": 560 * Billion, "data": 56 * Billion, "gold": Trillion},
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 228 * Million},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.2},
		},
		BuildTicks:     3600,
		RequiredAge:    "information_age",
		RequiredTech:   "e_commerce",
		Description:    "Automated container terminal moving freight at scale.",
		WorkerCapacity: 8,
	})
	// tier 4 — digital_age
	b = append(b, BuildingDef{
		Name: "Logistics Hub", Key: "logistics_hub",
		BaseCost: map[string]float64{"electricity": 2.8 * Trillion, "data": 360 * Billion},
		Effects: []Effect{
			{Type: "production", Target: "gold", Value: 102.4},
			{Type: "trade_route_income", Target: "trade_route_income", Value: 0.25},
		},
		BuildTicks:     3600,
		RequiredAge:    "digital_age",
		RequiredTech:   "automated_logistics",
		Description:    "Algorithmic logistics hub routing every shipment.",
		WorkerCapacity: 10,
	})

	return finishLineage(b, lineageSpec{
		Key: "harbor", Category: "production", Scale: ProducerRate, Domain: "trade",
	})
}
