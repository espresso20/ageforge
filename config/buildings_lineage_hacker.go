package config

// buildingsLineageHacker returns lineages 10-13:
// culture_arts, metallurgy, energy, hacker.
func buildingsLineageHacker() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 13 — HACKER/DIGITAL (lineageKey: "hacker", domain: "hacker", output: "data")
	// starts at information_age (tier 0)
	// rate = 2.0 * 2^tier  CostScale: 1.35  Category: "production"
	// =========================================================================

	// tier 0 — information_age  rate=2.0
	b = append(b, BuildingDef{
		Name: "Server Farm", Key: "server_farm", Category: "production",
		BaseCost:    map[string]float64{"electricity": 4e11, "data": 4.2e10, "steel": 7.6e11},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 8540000}},
		BuildTicks:  10920,
		RequiredAge: "information_age",
		Description: "Large server farm processing data.",
		LineageKey:  "hacker", LineageTier: 0,
		WorkerDomain: "hacker", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "data",
	})
	// tier 1 — digital_age  rate=4.0
	b = append(b, BuildingDef{
		Name: "Data Center", Key: "data_center", Category: "production",
		BaseCost:    map[string]float64{"electricity": 2e12, "data": 2.5e11, "steel": 2.9e12, "nanobots": 13000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 1.83e07}},
		BuildTicks:  12480,
		RequiredAge: "digital_age",
		Description: "Hyper-scale data center.",
		LineageKey:  "hacker", LineageTier: 1,
		WorkerDomain: "hacker", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "data",
	})
	// tier 2 — cyberpunk_age  rate=8.0
	b = append(b, BuildingDef{
		Name: "Cyber Hub", Key: "cyber_hub", Category: "production",
		BaseCost:    map[string]float64{"data": 9.5e11, "crypto": 5e12, "electricity": 9.9e12, "nanobots": 150000},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 5.239999999999999e07}},
		BuildTicks:  14040,
		RequiredAge: "cyberpunk_age",
		Description: "Cyberpunk underground hacker hub.",
		LineageKey:  "hacker", LineageTier: 2,
		WorkerDomain: "hacker", WorkerCapacity: 12,
		EpochKey: "neon_era", OutputResource: "data",
	})
	// tier 3 — fusion_age  rate=16.0
	b = append(b, BuildingDef{
		Name: "Quantum Server Farm", Key: "quantum_server_farm", Category: "production",
		BaseCost:    map[string]float64{"plasma": 2.2e13, "electricity": 6.4e13, "steel": 8.5e13},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 16}},
		BuildTicks:  15600,
		RequiredAge: "fusion_age",
		Description: "Quantum-computing server farm.",
		LineageKey:  "hacker", LineageTier: 3,
		WorkerDomain: "hacker", WorkerCapacity: 14,
		EpochKey: "neon_era", OutputResource: "data",
	})
	// tier 4 — space_age  rate=32.0
	b = append(b, BuildingDef{
		Name: "Orbital Data Relay", Key: "orbital_data_relay", Category: "production",
		BaseCost:    map[string]float64{"titanium": 3.6e14, "plasma": 1.8e14, "electricity": 4.5e14},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 32}},
		BuildTicks:  17160,
		RequiredAge: "space_age",
		Description: "Orbital relay node for stellar data networks.",
		LineageKey:  "hacker", LineageTier: 4,
		WorkerDomain: "hacker", WorkerCapacity: 16,
		EpochKey: "neon_era", OutputResource: "data",
	})
	// tier 5 — interstellar_age  rate=64.0
	b = append(b, BuildingDef{
		Name: "Galactic Network Node", Key: "galactic_network_node", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 4.5e14, "titanium": 3.5e15, "plasma": 2.2e15},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 64}},
		BuildTicks:  18720,
		RequiredAge: "interstellar_age",
		Description: "Interstellar galactic network node.",
		LineageKey:  "hacker", LineageTier: 5,
		WorkerDomain: "hacker", WorkerCapacity: 18,
		EpochKey: "cosmic_era", OutputResource: "data",
	})
	// tier 6 — galactic_age  rate=128.0
	b = append(b, BuildingDef{
		Name: "Consciousness Upload Hub", Key: "consciousness_upload_hub", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 9.1e14, "dark_matter": 4.6e15, "titanium": 2.3e16},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 128}},
		BuildTicks:  18720,
		RequiredAge: "galactic_age",
		Description: "Uploads and processes consciousness data.",
		LineageKey:  "hacker", LineageTier: 6,
		WorkerDomain: "hacker", WorkerCapacity: 20,
		EpochKey: "cosmic_era", OutputResource: "data",
	})
	// tier 7 — quantum_age  rate=256.0
	b = append(b, BuildingDef{
		Name: "Reality Processor", Key: "reality_processor", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 9.9e14, "antimatter": 3e17, "dark_matter": 2.5e17},
		CostScale:   1.15,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 256}},
		BuildTicks:  18720,
		RequiredAge: "quantum_age",
		Description: "Processes data from the very structure of reality.",
		LineageKey:  "hacker", LineageTier: 7,
		WorkerDomain: "hacker", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "data",
	})

	return b
}
