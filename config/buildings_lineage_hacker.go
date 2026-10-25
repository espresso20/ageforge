package config

// buildingsLineageHacker returns lineages 10-13:
// culture_arts, metallurgy, energy, hacker.
func buildingsLineageHacker() []BuildingDef {
	b := []BuildingDef{}

	// =========================================================================
	// LINEAGE 13 — HACKER/DIGITAL (lineageKey: "hacker", domain: "hacker", output: "data")
	// starts at modern_age (tier 0)
	// rate = 2.0 * 2^tier  CostScale: ProducerRate (1.15)  Category: "production"  (both in the lineageSpec at the foot)
	// =========================================================================

	// tier 0 — modern_age  the first maker of data: put on the maker rule's
	// standard (config/maker_rule.go), priced in steel and electricity only
	b = append(b, BuildingDef{
		Name: "Mainframe Hall", Key: "mainframe_hall",
		BaseCost: map[string]float64{"steel": 218 * Billion, "electricity": 80.1 * Billion},
		Effects: []Effect{
			{Type: "production", Target: "data", Value: 541000},
		},
		BuildTicks:     9360,
		RequiredAge:    "modern_age",
		Description:    "Room-sized computers that keep the records.",
		WorkerCapacity: 8,
	})
	// tier 1 — information_age  rate=2.0
	b = append(b, BuildingDef{
		Name: "Server Farm", Key: "server_farm",
		BaseCost: map[string]float64{"electricity": 596 * Billion, "steel": 565 * Billion, "titanium_ore": 565 * Billion},
		Effects: []Effect{
			{Type: "production", Target: "data", Value: 7.9 * Million},
		},
		BuildTicks:     10920,
		RequiredAge:    "information_age",
		Description:    "Large server farm processing data.",
		WorkerCapacity: 8,
	})
	// tier 2 — digital_age  rate=4.0
	b = append(b, BuildingDef{
		Name: "Data Center", Key: "data_center",
		BaseCost:       map[string]float64{"electricity": 2 * Trillion, "data": 250 * Billion, "steel": 1.45 * Trillion, "titanium_ore": 1.45 * Trillion, "nanobots": 13000},
		Effects:        []Effect{{Type: "production", Target: "data", Value: 18.3 * Million}},
		BuildTicks:     12480,
		RequiredAge:    "digital_age",
		Description:    "Hyper-scale data center.",
		WorkerCapacity: 10,
	})
	// tier 3 — cyberpunk_age  rate=8.0
	b = append(b, BuildingDef{
		Name: "Cyber Hub", Key: "cyber_hub",
		BaseCost:       map[string]float64{"data": 950 * Billion, "crypto": 5 * Trillion, "electricity": 4.95 * Trillion, "dark_matter_crystals": 4.95 * Trillion, "nanobots": 150000},
		Effects:        []Effect{{Type: "production", Target: "data", Value: 52.39999999999999 * Million}},
		BuildTicks:     14040,
		RequiredAge:    "cyberpunk_age",
		Description:    "Cyberpunk underground hacker hub.",
		WorkerCapacity: 12,
	})
	// tier 4 — fusion_age  rate=16.0
	b = append(b, BuildingDef{
		Name: "Quantum Server Farm", Key: "quantum_server_farm",
		BaseCost:       map[string]float64{"plasma": 22 * Trillion, "electricity": 32 * Trillion, "dark_matter_crystals": 32 * Trillion, "steel": 85 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "data", Value: 16}},
		BuildTicks:     15600,
		RequiredAge:    "fusion_age",
		Description:    "Quantum-computing server farm.",
		WorkerCapacity: 14,
	})
	// tier 5 — space_age  rate=32.0
	b = append(b, BuildingDef{
		Name: "Orbital Data Relay", Key: "orbital_data_relay",
		BaseCost:       map[string]float64{"titanium": 360 * Trillion, "plasma": 180 * Trillion, "electricity": 225 * Trillion, "dark_matter_crystals": 225 * Trillion},
		Effects:        []Effect{{Type: "production", Target: "data", Value: 32}},
		BuildTicks:     17160,
		RequiredAge:    "space_age",
		Description:    "Orbital relay node for stellar data networks.",
		WorkerCapacity: 16,
	})
	// tier 6 — interstellar_age  rate=64.0
	b = append(b, BuildingDef{
		Name: "Galactic Network Node", Key: "galactic_network_node",
		BaseCost:       map[string]float64{"dark_matter": 450 * Trillion, "titanium": 3.5 * Quadrillion, "plasma": 2.2 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "data", Value: 64}},
		BuildTicks:     18720,
		RequiredAge:    "interstellar_age",
		Description:    "Interstellar galactic network node.",
		WorkerCapacity: 18,
	})
	// tier 7 — galactic_age  rate=128.0
	b = append(b, BuildingDef{
		Name: "Consciousness Upload Hub", Key: "consciousness_upload_hub",
		BaseCost:       map[string]float64{"antimatter": 910 * Trillion, "dark_matter": 4.6 * Quadrillion, "titanium": 23 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "data", Value: 128}},
		BuildTicks:     18720,
		RequiredAge:    "galactic_age",
		Description:    "Uploads and processes consciousness data.",
		WorkerCapacity: 20,
	})
	// tier 8 — quantum_age  rate=256.0
	b = append(b, BuildingDef{
		Name: "Reality Processor", Key: "reality_processor",
		BaseCost:       map[string]float64{"quantum_flux": 990 * Trillion, "antimatter": 300 * Quadrillion, "dark_matter": 250 * Quadrillion},
		Effects:        []Effect{{Type: "production", Target: "data", Value: 256}},
		BuildTicks:     18720,
		RequiredAge:    "quantum_age",
		Description:    "Processes data from the very structure of reality.",
		WorkerCapacity: 25,
	})

	return finishLineage(b, lineageSpec{
		Key: "hacker", Category: "production", Scale: ProducerRate, Domain: "hacker",
	})
}
