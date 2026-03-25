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
		BaseCost:    map[string]float64{"steel": 22e6, "coal": 8e6, "gold": 12e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "coal", Value: 10}},
		BuildTicks:  25000,
		RequiredAge: "industrial_age",
		Description: "Industrial coal processing plant. +10 coal/tick (6 workers).",
		LineageKey:  "energy", LineageTier: 0,
		WorkerDomain: "energy", WorkerCapacity: 6,
		EpochKey: "steel_era", OutputResource: "coal",
	})
	// tier 1 — victorian_age  output=electricity  rate=50  (+ some coal)
	b = append(b, BuildingDef{
		Name: "Steam Turbine", Key: "steam_turbine", Category: "production",
		BaseCost:  map[string]float64{"steel": 185e6, "coal": 90e6, "gold": 110e6},
		CostScale: 1.35,
		Effects: []Effect{
			{Type: "production", Target: "electricity", Value: 50},
			{Type: "production", Target: "coal", Value: 5},
		},
		BuildTicks:  50000,
		RequiredAge: "victorian_age",
		Description: "Steam turbine generating electricity from coal. +50 electricity, +5 coal/tick (7 workers).",
		LineageKey:  "energy", LineageTier: 1,
		WorkerDomain: "energy", WorkerCapacity: 7,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 2 — electric_age  output=electricity  rate=100
	b = append(b, BuildingDef{
		Name: "Power Generator", Key: "power_generator", Category: "production",
		BaseCost:    map[string]float64{"steel": 1.1e9, "electricity": 450e6, "coal": 300e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 100}},
		BuildTicks:  75000,
		RequiredAge: "electric_age",
		Description: "Electric power generator. +100 electricity/tick (8 workers).",
		LineageKey:  "energy", LineageTier: 2,
		WorkerDomain: "energy", WorkerCapacity: 8,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 3 — atomic_age  output=electricity  rate=200
	b = append(b, BuildingDef{
		Name: "Nuclear Reactor", Key: "nuclear_reactor", Category: "production",
		BaseCost:    map[string]float64{"steel": 5.8e9, "electricity": 2.4e9, "uranium": 600e6},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 200}},
		BuildTicks:  100000,
		RequiredAge: "atomic_age",
		Description: "Nuclear fission reactor. +200 electricity/tick (9 workers).",
		LineageKey:  "energy", LineageTier: 3,
		WorkerDomain: "energy", WorkerCapacity: 9,
		EpochKey: "electric_era", OutputResource: "electricity",
	})
	// tier 4 — modern_age  output=oil + electricity  rate: oil=20 electricity=250
	// electricity raised from 100→250 so upgrading from nuclear_reactor (200 elec) never regresses
	b = append(b, BuildingDef{
		Name: "Oil Refinery", Key: "oil_refinery", Category: "production",
		BaseCost:  map[string]float64{"steel": 33e9, "electricity": 12e9, "oil": 2e9},
		CostScale: 1.35,
		Effects: []Effect{
			{Type: "production", Target: "oil", Value: 20},
			{Type: "production", Target: "electricity", Value: 250},
		},
		BuildTicks:  150000,
		RequiredAge: "modern_age",
		Description: "Modern oil refinery and power generation. +20 oil, +250 electricity/tick (10 workers).",
		LineageKey:  "energy", LineageTier: 4,
		WorkerDomain: "energy", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "oil",
	})
	// tier 5 — information_age  output=electricity  rate=400
	b = append(b, BuildingDef{
		Name: "Smart Energy Grid", Key: "smart_energy_grid", Category: "production",
		BaseCost:    map[string]float64{"electricity": 105e9, "data": 11e9, "steel": 190e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 400}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "AI-optimised smart energy grid. +400 electricity/tick (11 workers).",
		LineageKey:  "energy", LineageTier: 5,
		WorkerDomain: "energy", WorkerCapacity: 11,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 6 — digital_age  output=electricity  rate=800
	b = append(b, BuildingDef{
		Name: "Quantum Battery Array", Key: "quantum_battery_array", Category: "production",
		BaseCost:    map[string]float64{"electricity": 490e9, "data": 62e9, "steel": 720e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 800}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Quantum-state battery arrays. +800 electricity/tick (12 workers).",
		LineageKey:  "energy", LineageTier: 6,
		WorkerDomain: "energy", WorkerCapacity: 12,
		EpochKey: "digital_era", OutputResource: "electricity",
	})
	// tier 7 — cyberpunk_age  output=electricity  rate=1600
	b = append(b, BuildingDef{
		Name: "Dark Energy Tap", Key: "dark_energy_tap", Category: "production",
		BaseCost:    map[string]float64{"data": 240e9, "crypto": 1.25e12, "electricity": 2.5e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "electricity", Value: 1600}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Taps dark energy streams for electricity. +1600 electricity/tick (13 workers).",
		LineageKey:  "energy", LineageTier: 7,
		WorkerDomain: "energy", WorkerCapacity: 13,
		EpochKey: "neon_era", OutputResource: "electricity",
	})
	// tier 8 — fusion_age  output=plasma  rate=20
	b = append(b, BuildingDef{
		Name: "Fusion Reactor Array", Key: "fusion_reactor_array", Category: "production",
		BaseCost:  map[string]float64{"plasma": 5.2e12, "electricity": 16e12, "steel": 21e12},
		CostScale: 1.35,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 20},
			{Type: "production", Target: "electricity", Value: 2000},
		},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Array of fusion reactors producing plasma. +20 plasma, +2000 electricity/tick (15 workers).",
		LineageKey:  "energy", LineageTier: 8,
		WorkerDomain: "energy", WorkerCapacity: 15,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 9 — space_age  output=plasma + electricity  rate: plasma=40, electricity=2500
	// electricity preserved from fusion_reactor_array (2000) — upgrading must never regress
	b = append(b, BuildingDef{
		Name: "Solar Collector Array", Key: "solar_collector_array", Category: "production",
		BaseCost: map[string]float64{"titanium": 90e12, "plasma": 45e12, "electricity": 112e12},
		CostScale: 1.35,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 40},
			{Type: "production", Target: "electricity", Value: 2500},
		},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Orbital solar collectors feeding plasma energy. +40 plasma, +2500 electricity/tick (16 workers).",
		LineageKey:  "energy", LineageTier: 9,
		WorkerDomain: "energy", WorkerCapacity: 16,
		EpochKey: "neon_era", OutputResource: "plasma",
	})
	// tier 10 — interstellar_age  output=plasma  rate=80  (+ electricity continuation)
	// Resource pivot: plasma remains primary; electricity secondary continues from solar_collector_array
	// solar_collector_array: plasma:40 + electricity:2500 → pulsar_tap: plasma:80 + electricity:3200
	b = append(b, BuildingDef{
		Name: "Pulsar Tap", Key: "pulsar_tap", Category: "production",
		BaseCost: map[string]float64{"dark_matter": 110e12, "titanium": 840e12, "plasma": 520e12},
		CostScale: 1.35,
		Effects: []Effect{
			{Type: "production", Target: "plasma", Value: 80},
			{Type: "production", Target: "electricity", Value: 3200},
		},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Taps pulsar radiation for plasma and electricity. +80 plasma, +3200 electricity/tick (18 workers).",
		LineageKey:  "energy", LineageTier: 10,
		WorkerDomain: "energy", WorkerCapacity: 18,
		EpochKey: "cosmic_era", OutputResource: "plasma",
	})
	// tier 11 — galactic_age  output=dark_matter  rate=10
	// Resource pivot: plasma → dark_matter. Validator shows LOW boost — intentional cross-resource pivot.
	// dark_matter rate of 10 is deliberately low (dark_matter is a tier-7 resource; starts at 1/tick in metallurgy).
	// quasar_tap → zero_point_generator jumps 5x (dark_matter:10 → quantum_flux:50) — also an intentional pivot.
	b = append(b, BuildingDef{
		Name: "Quasar Tap", Key: "quasar_tap", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 220e12, "dark_matter": 1.1e15, "titanium": 5.5e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "dark_matter", Value: 10}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Taps quasar jets for dark matter. +10 dark_matter/tick (20 workers).",
		LineageKey:  "energy", LineageTier: 11,
		WorkerDomain: "energy", WorkerCapacity: 20,
		EpochKey: "cosmic_era", OutputResource: "dark_matter",
	})
	// tier 12 — quantum_age  output=quantum_flux  rate=50
	b = append(b, BuildingDef{
		Name: "Zero Point Generator", Key: "zero_point_generator", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 230e12, "antimatter": 68e15, "dark_matter": 58e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "quantum_flux", Value: 50}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Generates energy from quantum zero-point fields. +50 quantum_flux/tick (25 workers).",
		LineageKey:  "energy", LineageTier: 12,
		WorkerDomain: "energy", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "quantum_flux",
	})

	return b
}
