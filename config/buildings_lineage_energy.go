package config

// buildingsLineageEnergy returns lineages 10-13:
// culture_arts, metallurgy, energy, hacker.
func buildingsLineageEnergy() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 12 — ENERGY (lineageKey: "energy", domain: "energy")
	// starts at industrial_age (tier 0)
	// CostScale: ProducerRate (1.15)  Category: "production"  (both in the lineageSpec at the foot)
	// Output: coal/electricity transitions → plasma → dark_matter → quantum_flux
	// =========================================================================

	// tier 0 — industrial_age  output=coal  rate=10
	// Resource pivot note: coal_plant (coal) → steam_turbine (electricity+coal bonus)
	// Validator will show HIGH_BOOST on this transition — intentional cross-resource pivot.
	b = append(b, BuildingDef{
		Name: "Coal Plant", Key: "coal_plant",
		BaseCost:       map[string]float64{"steel": 93 * Million, "coal": 34 * Million, "gold": 51 * Million},
		Effects:        []Effect{{Type: "production", Target: "coal", Value: 11500}},
		BuildTicks:     6240,
		RequiredAge:    "industrial_age",
		RequiredTech:   "steam_power",
		Description:    "Industrial coal processing plant.",
		WorkerCapacity: 6,
	})
	// tier 1 — victorian_age  output=electricity  rate=50  (+ some coal)
	b = append(b, BuildingDef{
		Name: "Steam Turbine", Key: "steam_turbine",
		BaseCost: map[string]float64{"steel": 780 * Million, "coal": 380 * Million, "gold": 470 * Million},
		Effects: []Effect{
			{Type: "production", Target: "electricity", Value: 50},
			{Type: "production", Target: "coal", Value: 106000},
		},
		BuildTicks:     7020,
		RequiredAge:    "victorian_age",
		Description:    "A steam turbine hall with its own coal seam.",
		WorkerCapacity: 7,
	})
	// tier 2 — electric_age  output=electricity  rate=100
	b = append(b, BuildingDef{
		Name: "Dynamo Hall", Key: "power_generator",
		BaseCost: map[string]float64{"steel": 4.7 * Billion, "electricity": 1.9 * Billion, "coal": 1.3 * Billion},
		Effects: []Effect{
			{Type: "production", Target: "electricity", Value: 223000},
		},
		BuildTicks:     7800,
		RequiredAge:    "electric_age",
		RequiredTech:   "power_distribution",
		Description:    "Electric power generator.",
		WorkerCapacity: 8,
	})
	// tier 3 — atomic_age  output=electricity  rate=200
	b = append(b, BuildingDef{
		Name: "Breeder Reactor", Key: "nuclear_reactor",
		BaseCost:       map[string]float64{"steel": 25 * Billion, "electricity": 10 * Billion, "uranium": 2.5 * Billion},
		Effects:        []Effect{{Type: "production", Target: "electricity", Value: 1.29 * Million}},
		BuildTicks:     9360,
		RequiredAge:    "atomic_age",
		Description:    "Nuclear fission reactor.",
		WorkerCapacity: 9,
	})
	// tier 4 — modern_age  output=oil + electricity  rate: oil=20 electricity=250
	// electricity raised from 100→250 so upgrading from nuclear_reactor (200 elec) never regresses
	b = append(b, BuildingDef{
		Name: "Oil Refinery", Key: "oil_refinery",
		BaseCost: map[string]float64{"steel": 70 * Billion, "titanium_ore": 70 * Billion, "electricity": 51 * Billion, "oil": 8.5 * Billion},
		Effects: []Effect{
			{Type: "production", Target: "oil", Value: 540000},
			{Type: "production", Target: "electricity", Value: 2.86 * Million},
		},
		BuildTicks:     9360,
		RequiredAge:    "modern_age",
		Description:    "Modern oil refinery and power generation.",
		WorkerCapacity: 10,
	})
	// tier 5 — information_age  output=electricity  rate=400
	b = append(b, BuildingDef{
		Name: "Microgrid Array", Key: "smart_energy_grid",
		BaseCost:       map[string]float64{"electricity": 440 * Billion, "data": 47 * Billion, "steel": 400 * Billion, "titanium_ore": 400 * Billion},
		Effects:        []Effect{{Type: "production", Target: "electricity", Value: 93.09999999999999 * Million}},
		BuildTicks:     10920,
		RequiredAge:    "information_age",
		RequiredTech:   "smart_grid",
		Description:    "AI-optimized microgrid batteries.",
		WorkerCapacity: 11,
	})
	// tier 6 — digital_age  output=electricity  rate=800
	b = append(b, BuildingDef{
		Name: "Quantum Battery Array", Key: "quantum_battery_array",
		BaseCost:       map[string]float64{"electricity": 2.1 * Trillion, "data": 260 * Billion, "steel": 1.5 * Trillion, "titanium_ore": 1.5 * Trillion, "nanobots": 14000},
		Effects:        []Effect{{Type: "production", Target: "electricity", Value: 154 * Million}},
		BuildTicks:     12480,
		RequiredAge:    "digital_age",
		RequiredTech:   "grid_storage",
		Description:    "Quantum-state battery arrays.",
		WorkerCapacity: 12,
	})
	// tier 7 — cyberpunk_age  output=electricity  rate=1600
	b = append(b, BuildingDef{
		Name: "Dark Energy Tap", Key: "dark_energy_tap",
		BaseCost:       map[string]float64{"data": Trillion, "crypto": 5.3 * Trillion, "electricity": 5.5 * Trillion, "dark_matter_crystals": 5.5 * Trillion, "nanobots": 150000},
		Effects:        []Effect{{Type: "production", Target: "electricity", Value: 755 * Million}},
		BuildTicks:     14040,
		RequiredAge:    "cyberpunk_age",
		RequiredTech:   "dark_energy",
		Description:    "Taps dark energy streams for electricity.",
		WorkerCapacity: 13,
	})
	// tier 8 — fusion_age  output=plasma  rate=20
	b = append(b, BuildingDef{
		Name: "Tokamak Array", Key: "fusion_reactor_array",
		BaseCost: map[string]float64{"plasma": 22 * Trillion, "electricity": 34 * Trillion, "dark_matter_crystals": 34 * Trillion, "steel": 89 * Trillion},
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 259 * Million},
			{Type: "production", Target: "electricity", Value: 776 * Million},
		},
		BuildTicks:     15600,
		RequiredAge:    "fusion_age",
		Description:    "Array of fusion reactors producing plasma.",
		WorkerCapacity: 15,
	})
	// tier 9 — space_age  output=plasma + electricity  rate: plasma=40, electricity=2500
	// electricity preserved from fusion_reactor_array (2000) — upgrading must never regress
	b = append(b, BuildingDef{
		Name: "Solar Collector Array", Key: "solar_collector_array",
		BaseCost: map[string]float64{"titanium": 380 * Trillion, "plasma": 190 * Trillion, "electricity": 235 * Trillion, "dark_matter_crystals": 235 * Trillion},
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 2.92 * Billion},
			{Type: "production", Target: "electricity", Value: 7.31 * Billion},
		},
		BuildTicks:     17160,
		RequiredAge:    "space_age",
		Description:    "Orbital solar collectors feeding plasma energy.",
		WorkerCapacity: 16,
	})
	// tier 10 — interstellar_age  output=plasma  rate=80  (+ electricity continuation)
	// Resource pivot: plasma remains primary; electricity secondary continues from solar_collector_array
	// solar_collector_array: plasma:40 + electricity:2500 → pulsar_tap: plasma:80 + electricity:3200
	b = append(b, BuildingDef{
		Name: "Pulsar Tap", Key: "pulsar_tap",
		BaseCost: map[string]float64{"dark_matter": 470 * Trillion, "titanium": 3.6 * Quadrillion, "plasma": 2.2 * Quadrillion},
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 42.5 * Billion},
			{Type: "production", Target: "electricity", Value: 7.94 * Billion},
		},
		BuildTicks:     18720,
		RequiredAge:    "interstellar_age",
		RequiredTech:   "stellar_engineering",
		Description:    "Taps pulsar radiation for plasma and electricity.",
		WorkerCapacity: 18,
	})
	// tier 11 — galactic_age  output=dark_matter  rate=10
	// Resource pivot: plasma → dark_matter. Validator shows LOW boost — intentional cross-resource pivot.
	// dark_matter rate of 10 is deliberately low (dark_matter is a tier-7 resource; starts at 1/tick in metallurgy).
	// quasar_tap → zero_point_generator jumps 5x (dark_matter:10 → quantum_flux:50) — also an intentional pivot.
	b = append(b, BuildingDef{
		Name: "Quasar Tap", Key: "quasar_tap",
		BaseCost:       map[string]float64{"antimatter": 930 * Trillion, "dark_matter": 4.7 * Quadrillion, "titanium": 23 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "dark_matter", Value: 82.3 * Billion}},
		BuildTicks:     18720,
		RequiredAge:    "galactic_age",
		RequiredTech:   "antimatter_synthesis",
		Description:    "Taps quasar jets for dark matter.",
		WorkerCapacity: 20,
	})
	// tier 12 — quantum_age  output=quantum_flux  rate=50
	b = append(b, BuildingDef{
		Name: "Zero Point Generator", Key: "zero_point_generator",
		BaseCost:       map[string]float64{"quantum_flux": 970 * Trillion, "antimatter": 290 * Quadrillion, "dark_matter": 250 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "quantum_flux", Value: 67 * Billion}},
		BuildTicks:     18720,
		RequiredAge:    "quantum_age",
		RequiredTech:   "zero_point_energy",
		Description:    "Generates energy from quantum zero-point fields.",
		WorkerCapacity: 25,
	})

	return finishLineage(b, lineageSpec{
		Key: "energy", Category: "production", Scale: ProducerRate, Domain: "energy", Lines: true,
	})
}
