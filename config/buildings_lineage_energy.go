package config

// buildingsLineageEnergy returns lineages 10-13:
// culture_arts, metallurgy, energy, hacker.
func buildingsLineageEnergy() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 12 — ENERGY (lineageKey: "energy", domain: "energy")
	// starts at industrial_age (tier 0)
	// CostScale: 1.35  Category: "production"
	// Output: coal/electricity transitions → plasma → dark_matter → quantum_flux
	// =========================================================================

	// tier 0 — industrial_age  output=coal  rate=10
	// Resource pivot note: coal_plant (coal) → steam_turbine (electricity+coal bonus)
	// Validator will show HIGH_BOOST on this transition — intentional cross-resource pivot.
	b = append(b, BuildingDef{
		Name: "Coal Plant", Key: "coal_plant", Category: "production",
		BaseCost:     map[string]float64{"steel": 9.3e07, "coal": 3.4e07, "gold": 5.1e07},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "coal", Value: 11500}},
		BuildTicks:   6240,
		RequiredAge:  "industrial_age",
		RequiredTech: "steam_power",
		Description:  "Industrial coal processing plant.",
		LineageKey:   "energy", LineageTier: 0,
		WorkerDomain: "energy", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 1 — victorian_age  output=electricity  rate=50  (+ some coal)
	b = append(b, BuildingDef{
		Name: "Steam Turbine", Key: "steam_turbine", Category: "production",
		BaseCost:  map[string]float64{"steel": 7.8e08, "coal": 3.8e08, "gold": 4.7e08},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "electricity", Value: 50},
			{Type: "production", Target: "coal", Value: 106000},
		},
		BuildTicks:  7020,
		RequiredAge: "victorian_age",
		Description: "A steam turbine hall with its own coal seam.",
		LineageKey:  "energy", LineageTier: 1,
		WorkerDomain: "energy", WorkerCapacity: 7,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 2 — electric_age  output=electricity  rate=100
	b = append(b, BuildingDef{
		Name: "Dynamo Hall", Key: "power_generator", Category: "production",
		BaseCost:     map[string]float64{"steel": 4.7e09, "electricity": 1.9e09, "coal": 1.3e09},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "electricity", Value: 240000}},
		BuildTicks:   7800,
		RequiredAge:  "electric_age",
		RequiredTech: "power_distribution",
		Description:  "Electric power generator.",
		LineageKey:   "energy", LineageTier: 2,
		WorkerDomain: "energy", WorkerCapacity: 8,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 3 — atomic_age  output=electricity  rate=200
	b = append(b, BuildingDef{
		Name: "Breeder Reactor", Key: "nuclear_reactor", Category: "production",
		BaseCost:    map[string]float64{"steel": 2.5e10, "electricity": 1e10, "uranium": 2.5e09},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 1290000}},
		BuildTicks:  9360,
		RequiredAge: "atomic_age",
		Description: "Nuclear fission reactor.",
		LineageKey:  "energy", LineageTier: 3,
		WorkerDomain: "energy", WorkerCapacity: 9,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 4 — modern_age  output=oil + electricity  rate: oil=20 electricity=250
	// electricity raised from 100→250 so upgrading from nuclear_reactor (200 elec) never regresses
	b = append(b, BuildingDef{
		Name: "Oil Refinery", Key: "oil_refinery", Category: "production",
		BaseCost:  map[string]float64{"steel": 1.4e11, "electricity": 5.1e10, "oil": 8.5e09},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "oil", Value: 546000},
			{Type: "production", Target: "electricity", Value: 2500000},
		},
		BuildTicks:  9360,
		RequiredAge: "modern_age",
		Description: "Modern oil refinery and power generation.",
		LineageKey:  "energy", LineageTier: 4,
		WorkerDomain: "energy", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "oil",
	})
	// tier 5 — information_age  output=electricity  rate=400
	b = append(b, BuildingDef{
		Name: "Microgrid Array", Key: "smart_energy_grid", Category: "production",
		BaseCost:     map[string]float64{"electricity": 4.4e11, "data": 4.7e10, "steel": 8e11},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "electricity", Value: 9.309999999999999e07}},
		BuildTicks:   10920,
		RequiredAge:  "information_age",
		RequiredTech: "smart_grid",
		Description:  "AI-optimized microgrid batteries.",
		LineageKey:   "energy", LineageTier: 5,
		WorkerDomain: "energy", WorkerCapacity: 11,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 6 — digital_age  output=electricity  rate=800
	b = append(b, BuildingDef{
		Name: "Quantum Battery Array", Key: "quantum_battery_array", Category: "production",
		BaseCost:     map[string]float64{"electricity": 2.1e12, "data": 2.6e11, "steel": 3e12, "nanobots": 14000},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "electricity", Value: 1.54e08}},
		BuildTicks:   12480,
		RequiredAge:  "digital_age",
		RequiredTech: "grid_storage",
		Description:  "Quantum-state battery arrays.",
		LineageKey:   "energy", LineageTier: 6,
		WorkerDomain: "energy", WorkerCapacity: 12,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 7 — cyberpunk_age  output=electricity  rate=1600
	b = append(b, BuildingDef{
		Name: "Dark Energy Tap", Key: "dark_energy_tap", Category: "production",
		BaseCost:     map[string]float64{"data": 1e12, "crypto": 5.3e12, "electricity": 1.1e13, "nanobots": 150000},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "electricity", Value: 7.55e08}},
		BuildTicks:   14040,
		RequiredAge:  "cyberpunk_age",
		RequiredTech: "dark_energy",
		Description:  "Taps dark energy streams for electricity.",
		LineageKey:   "energy", LineageTier: 7,
		WorkerDomain: "energy", WorkerCapacity: 13,
		EpochKey: "neon_era", OutputResource: "electricity",
	})
	// tier 8 — fusion_age  output=plasma  rate=20
	b = append(b, BuildingDef{
		Name: "Tokamak Array", Key: "fusion_reactor_array", Category: "production",
		BaseCost:  map[string]float64{"plasma": 2.2e13, "electricity": 6.8e13, "steel": 8.9e13},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 2.59e08},
			{Type: "production", Target: "electricity", Value: 7.76e08},
		},
		BuildTicks:  15600,
		RequiredAge: "fusion_age",
		Description: "Array of fusion reactors producing plasma.",
		LineageKey:  "energy", LineageTier: 8,
		WorkerDomain: "energy", WorkerCapacity: 15,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 9 — space_age  output=plasma + electricity  rate: plasma=40, electricity=2500
	// electricity preserved from fusion_reactor_array (2000) — upgrading must never regress
	b = append(b, BuildingDef{
		Name: "Solar Collector Array", Key: "solar_collector_array", Category: "production",
		BaseCost:  map[string]float64{"titanium": 3.8e14, "plasma": 1.9e14, "electricity": 4.7e14},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 2.92e09},
			{Type: "production", Target: "electricity", Value: 7.31e09},
		},
		BuildTicks:  17160,
		RequiredAge: "space_age",
		Description: "Orbital solar collectors feeding plasma energy.",
		LineageKey:  "energy", LineageTier: 9,
		WorkerDomain: "energy", WorkerCapacity: 16,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 10 — interstellar_age  output=plasma  rate=80  (+ electricity continuation)
	// Resource pivot: plasma remains primary; electricity secondary continues from solar_collector_array
	// solar_collector_array: plasma:40 + electricity:2500 → pulsar_tap: plasma:80 + electricity:3200
	b = append(b, BuildingDef{
		Name: "Pulsar Tap", Key: "pulsar_tap", Category: "production",
		BaseCost:  map[string]float64{"dark_matter": 4.7e14, "titanium": 3.6e15, "plasma": 2.2e15},
		CostScale: 1.15,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 4.25e10},
			{Type: "production", Target: "electricity", Value: 7.94e09},
		},
		BuildTicks:   18720,
		RequiredAge:  "interstellar_age",
		RequiredTech: "stellar_engineering",
		Description:  "Taps pulsar radiation for plasma and electricity.",
		LineageKey:   "energy", LineageTier: 10,
		WorkerDomain: "energy", WorkerCapacity: 18,
		EpochKey: "cosmic_era", OutputResource: "plasma",
	})
	// tier 11 — galactic_age  output=dark_matter  rate=10
	// Resource pivot: plasma → dark_matter. Validator shows LOW boost — intentional cross-resource pivot.
	// dark_matter rate of 10 is deliberately low (dark_matter is a tier-7 resource; starts at 1/tick in metallurgy).
	// quasar_tap → zero_point_generator jumps 5x (dark_matter:10 → quantum_flux:50) — also an intentional pivot.
	b = append(b, BuildingDef{
		Name: "Quasar Tap", Key: "quasar_tap", Category: "production",
		BaseCost:     map[string]float64{"antimatter": 9.3e14, "dark_matter": 4.7e15, "titanium": 2.3e16},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "dark_matter", Value: 8.23e10}},
		BuildTicks:   18720,
		RequiredAge:  "galactic_age",
		RequiredTech: "antimatter_synthesis",
		Description:  "Taps quasar jets for dark matter.",
		LineageKey:   "energy", LineageTier: 11,
		WorkerDomain: "energy", WorkerCapacity: 20,
		EpochKey: "cosmic_era", OutputResource: "dark_matter",
	})
	// tier 12 — quantum_age  output=quantum_flux  rate=50
	b = append(b, BuildingDef{
		Name: "Zero Point Generator", Key: "zero_point_generator", Category: "production",
		BaseCost:     map[string]float64{"quantum_flux": 9.7e14, "antimatter": 2.9e17, "dark_matter": 2.5e17},
		CostScale:    1.15,
		Effects:      []Effect{{Type: "production", Target: "quantum_flux", Value: 6.7e10}},
		BuildTicks:   18720,
		RequiredAge:  "quantum_age",
		RequiredTech: "zero_point_energy",
		Description:  "Generates energy from quantum zero-point fields.",
		LineageKey:   "energy", LineageTier: 12,
		WorkerDomain: "energy", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	return b
}
