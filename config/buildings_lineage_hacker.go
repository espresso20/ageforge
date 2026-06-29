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
		BaseCost:    map[string]float64{"electricity": 95e9, "data": 10e9, "steel": 180e9},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 2.0}},
		BuildTicks:  300000,
		RequiredAge: "information_age",
		Description: "Large server farm processing data. +2.0 data/tick (8 workers).",
		LineageKey:  "hacker", LineageTier: 0,
		WorkerDomain: "hacker", WorkerCapacity: 8,
		EpochKey: "digital_era", OutputResource: "data",
	})
	// tier 1 — digital_age  rate=4.0
	b = append(b, BuildingDef{
		Name: "Data Center", Key: "data_center", Category: "production",
		BaseCost:    map[string]float64{"electricity": 470e9, "data": 58e9, "steel": 685e9, "nanobots": 3000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 4.0}},
		BuildTicks:  500000,
		RequiredAge: "digital_age",
		Description: "Hyper-scale data center. +4.0 data/tick (10 workers).",
		LineageKey:  "hacker", LineageTier: 1,
		WorkerDomain: "hacker", WorkerCapacity: 10,
		EpochKey: "digital_era", OutputResource: "data",
	})
	// tier 2 — cyberpunk_age  rate=8.0
	b = append(b, BuildingDef{
		Name: "Cyber Hub", Key: "cyber_hub", Category: "production",
		BaseCost:    map[string]float64{"data": 225e9, "crypto": 1.18e12, "electricity": 2.35e12, "nanobots": 35000},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 8.0}},
		BuildTicks:  1000000,
		RequiredAge: "cyberpunk_age",
		Description: "Cyberpunk underground hacker hub. +8.0 data/tick (12 workers).",
		LineageKey:  "hacker", LineageTier: 2,
		WorkerDomain: "hacker", WorkerCapacity: 12,
		EpochKey: "neon_era", OutputResource: "data",
	})
	// tier 3 — fusion_age  rate=16.0
	b = append(b, BuildingDef{
		Name: "Quantum Server Farm", Key: "quantum_server_farm", Category: "production",
		BaseCost:    map[string]float64{"plasma": 5.1e12, "electricity": 15e12, "steel": 20e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 16.0}},
		BuildTicks:  1500000,
		RequiredAge: "fusion_age",
		Description: "Quantum-computing server farm. +16.0 data/tick (14 workers).",
		LineageKey:  "hacker", LineageTier: 3,
		WorkerDomain: "hacker", WorkerCapacity: 14,
		EpochKey: "neon_era", OutputResource: "data",
	})
	// tier 4 — space_age  rate=32.0
	b = append(b, BuildingDef{
		Name: "Orbital Data Relay", Key: "orbital_data_relay", Category: "production",
		BaseCost:    map[string]float64{"titanium": 86e12, "plasma": 42e12, "electricity": 106e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 32.0}},
		BuildTicks:  2000000,
		RequiredAge: "space_age",
		Description: "Orbital relay node for stellar data networks. +32.0 data/tick (16 workers).",
		LineageKey:  "hacker", LineageTier: 4,
		WorkerDomain: "hacker", WorkerCapacity: 16,
		EpochKey: "neon_era", OutputResource: "data",
	})
	// tier 5 — interstellar_age  rate=64.0
	b = append(b, BuildingDef{
		Name: "Galactic Network Node", Key: "galactic_network_node", Category: "production",
		BaseCost:    map[string]float64{"dark_matter": 107e12, "titanium": 830e12, "plasma": 515e12},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 64.0}},
		BuildTicks:  2500000,
		RequiredAge: "interstellar_age",
		Description: "Interstellar galactic network node. +64.0 data/tick (18 workers).",
		LineageKey:  "hacker", LineageTier: 5,
		WorkerDomain: "hacker", WorkerCapacity: 18,
		EpochKey: "cosmic_era", OutputResource: "data",
	})
	// tier 6 — galactic_age  rate=128.0
	b = append(b, BuildingDef{
		Name: "Consciousness Upload Hub", Key: "consciousness_upload_hub", Category: "production",
		BaseCost:    map[string]float64{"antimatter": 215e12, "dark_matter": 1.08e15, "titanium": 5.4e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 128.0}},
		BuildTicks:  3000000,
		RequiredAge: "galactic_age",
		Description: "Uploads and processes consciousness data. +128.0 data/tick (20 workers).",
		LineageKey:  "hacker", LineageTier: 6,
		WorkerDomain: "hacker", WorkerCapacity: 20,
		EpochKey: "cosmic_era", OutputResource: "data",
	})
	// tier 7 — quantum_age  rate=256.0
	b = append(b, BuildingDef{
		Name: "Reality Processor", Key: "reality_processor", Category: "production",
		BaseCost:    map[string]float64{"quantum_flux": 235e12, "antimatter": 70e15, "dark_matter": 60e15},
		CostScale:   1.35,
		Effects:     []Effect{{Type: "production", Target: "data", Value: 256.0}},
		BuildTicks:  5000000,
		RequiredAge: "quantum_age",
		Description: "Processes data from the very structure of reality. +256.0 data/tick (25 workers).",
		LineageKey:  "hacker", LineageTier: 7,
		WorkerDomain: "hacker", WorkerCapacity: 25,
		EpochKey: "cosmic_era", OutputResource: "data",
	})

	return b
}
