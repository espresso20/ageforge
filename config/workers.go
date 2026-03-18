package config

// WorkerClassDef defines a single tier of a worker domain, tied to a specific age.
// Each domain has one class per age it spans (from its start age through quantum_age).
// Food cost and production multiplier scale geometrically per tier:
//   FoodCost = baseFoodCost × 1.12^tier
//   Multiplier = 2.0^tier
// With 1.12 the tier-20 cost is ~9.6× base (vs 3325× with the old 1.5 multiplier).
// food domain (base 0.06): tier 0 → 0.060, tier 10 → 0.186, tier 20 → 0.579
// lumber/masonry/knowledge (base 1.0): tier 0 → 1.000, tier 9 → 2.773, tier 19 → 7.690
type WorkerClassDef struct {
	Domain     string  // domain key (e.g. "food", "knowledge", "metallurgy")
	AgeKey     string  // age at which this class tier is active
	ClassName  string  // display name shown in UI
	FoodCost   float64 // food consumed per tick per worker of this class
	Multiplier float64 // production multiplier vs building base rate
}

// wc is an internal helper that creates a WorkerClassDef with geometrically
// computed food cost and multiplier based on tier index within the domain.
func wc(domain, ageKey, className string, baseFoodCost float64, tier int) WorkerClassDef {
	fc := baseFoodCost
	mp := 1.0
	for i := 0; i < tier; i++ {
		fc *= 1.12 // was 1.5 — softer scaling so tier-20 costs ~9x base not 3325x
		mp *= 2.0
	}
	return WorkerClassDef{Domain: domain, AgeKey: ageKey, ClassName: className, FoodCost: fc, Multiplier: mp}
}

// WorkerClasses returns all worker class definitions across all 12 domains.
// Domains and their start ages / base food costs:
//   food        primitive   0.06  — Forager → Quantum Harvester
//   lumber      stone       1.0   — Gatherer → Cosmic Extractor
//   masonry     stone       1.0   — Quarryman → Crystal Miner
//   knowledge   primitive   1.0   — Shaman → Quantum Theorist
//   faith       medieval    2.0   — Acolyte → Quantum Sage
//   military    iron        2.0   — Soldier → Quantum Soldier
//   trade       bronze      1.0   — Peddler → Quantum Dealer
//   engineering victorian   8.0   — Tinker → Quantum Engineer
//   metallurgy  iron        2.0   — Smelter → Quantum Smelter
//   energy      victorian   8.0   — Stoker → Zero-Point Engineer
//   hacker      information 16.0  — Script Kiddie → Quantum Hacker
//   astronaut   space       32.0  — Cadet → Quantum Astronaut
// Culture has no worker domain — it auto-produces passively.
func WorkerClasses() []WorkerClassDef {
	return []WorkerClassDef{

		// === FOOD DOMAIN (starts primitive_age, base food cost 0.06) ===
		// FoodCost = 0.06 × 1.12^tier:
		//   tier 0  primitive:    0.060  Forager
		//   tier 1  stone:        0.067  Farmhand
		//   tier 2  bronze:       0.075  Cultivator
		//   tier 3  iron:         0.084  Laborer
		//   tier 4  classical:    0.095  Peasant
		//   tier 5  medieval:     0.106  Serf
		//   tier 6  renaissance:  0.119  Plowman
		//   tier 7  colonial:     0.133  Colonial Farmer
		//   tier 8  industrial:   0.149  Factory Hand
		//   tier 9  victorian:    0.167  Agricultural Worker
		//   tier 10 electric:     0.187  Electric Farmer
		//   tier 11 atomic:       0.209  Atomic Agronomist
		//   tier 12 modern:       0.234  Modern Farmer
		//   tier 13 information:  0.263  Digital Cultivator
		//   tier 14 digital:      0.294  AI Agronomist
		//   tier 15 cyberpunk:    0.330  Aug Harvester
		//   tier 16 fusion:       0.369  Bio-Farmer
		//   tier 17 space:        0.414  Zero-G Farmer
		//   tier 18 interstellar: 0.463  Stellar Cultivator
		//   tier 19 galactic:     0.519  Galactic Farmer
		//   tier 20 quantum:      0.581  Quantum Harvester
		wc("food", "primitive_age", "Forager", 0.06, 0),
		wc("food", "stone_age", "Farmhand", 0.06, 1),
		wc("food", "bronze_age", "Cultivator", 0.06, 2),
		wc("food", "iron_age", "Laborer", 0.06, 3),
		wc("food", "classical_age", "Peasant", 0.06, 4),
		wc("food", "medieval_age", "Serf", 0.06, 5),
		wc("food", "renaissance_age", "Plowman", 0.06, 6),
		wc("food", "colonial_age", "Colonial Farmer", 0.06, 7),
		wc("food", "industrial_age", "Factory Hand", 0.06, 8),
		wc("food", "victorian_age", "Agricultural Worker", 0.06, 9),
		wc("food", "electric_age", "Electric Farmer", 0.06, 10),
		wc("food", "atomic_age", "Atomic Agronomist", 0.06, 11),
		wc("food", "modern_age", "Modern Farmer", 0.06, 12),
		wc("food", "information_age", "Digital Cultivator", 0.06, 13),
		wc("food", "digital_age", "AI Agronomist", 0.06, 14),
		wc("food", "cyberpunk_age", "Aug Harvester", 0.06, 15),
		wc("food", "fusion_age", "Bio-Farmer", 0.06, 16),
		wc("food", "space_age", "Zero-G Farmer", 0.06, 17),
		wc("food", "interstellar_age", "Stellar Cultivator", 0.06, 18),
		wc("food", "galactic_age", "Galactic Farmer", 0.06, 19),
		wc("food", "quantum_age", "Quantum Harvester", 0.06, 20),

		// === LUMBER / ORGANIC EXTRACTION DOMAIN (starts stone_age, base 1.0) ===
		wc("lumber", "stone_age", "Gatherer", 1.0, 0),
		wc("lumber", "bronze_age", "Woodcutter", 1.0, 1),
		wc("lumber", "iron_age", "Lumberjack", 1.0, 2),
		wc("lumber", "classical_age", "Sawyer", 1.0, 3),
		wc("lumber", "medieval_age", "Forester", 1.0, 4),
		wc("lumber", "renaissance_age", "Colonial Logger", 1.0, 5),
		wc("lumber", "colonial_age", "Mill Worker", 1.0, 6),
		wc("lumber", "industrial_age", "Coal Extractor", 1.0, 7),
		wc("lumber", "victorian_age", "Steam Logger", 1.0, 8),
		wc("lumber", "electric_age", "Electric Forester", 1.0, 9),
		wc("lumber", "atomic_age", "Fuel Extractor", 1.0, 10),
		wc("lumber", "modern_age", "Petroleum Worker", 1.0, 11),
		wc("lumber", "information_age", "Digital Forester", 1.0, 12),
		wc("lumber", "digital_age", "Bio-Extractor", 1.0, 13),
		wc("lumber", "cyberpunk_age", "Nano-Harvester", 1.0, 14),
		wc("lumber", "fusion_age", "Organic Engineer", 1.0, 15),
		wc("lumber", "space_age", "Biofield Harvester", 1.0, 16),
		wc("lumber", "interstellar_age", "Quantum Extractor", 1.0, 17),
		wc("lumber", "galactic_age", "Galactic Forester", 1.0, 18),
		wc("lumber", "quantum_age", "Cosmic Extractor", 1.0, 19),

		// === MASONRY / GEOLOGICAL EXTRACTION DOMAIN (starts stone_age, base 1.0) ===
		wc("masonry", "stone_age", "Quarryman", 1.0, 0),
		wc("masonry", "bronze_age", "Stone Cutter", 1.0, 1),
		wc("masonry", "iron_age", "Miner", 1.0, 2),
		wc("masonry", "classical_age", "Iron Extractor", 1.0, 3),
		wc("masonry", "medieval_age", "Medieval Miner", 1.0, 4),
		wc("masonry", "renaissance_age", "Renaissance Quarryman", 1.0, 5),
		wc("masonry", "colonial_age", "Colonial Miner", 1.0, 6),
		wc("masonry", "industrial_age", "Industrial Miner", 1.0, 7),
		wc("masonry", "victorian_age", "Victorian Quarryman", 1.0, 8),
		wc("masonry", "electric_age", "Electric Miner", 1.0, 9),
		wc("masonry", "atomic_age", "Uranium Miner", 1.0, 10),
		wc("masonry", "modern_age", "Modern Geologist", 1.0, 11),
		wc("masonry", "information_age", "Data Miner", 1.0, 12),
		wc("masonry", "digital_age", "Digital Excavator", 1.0, 13),
		wc("masonry", "cyberpunk_age", "Cyber Miner", 1.0, 14),
		wc("masonry", "fusion_age", "Plasma Driller", 1.0, 15),
		wc("masonry", "space_age", "Space Miner", 1.0, 16),
		wc("masonry", "interstellar_age", "Asteroid Miner", 1.0, 17),
		wc("masonry", "galactic_age", "Dark Matter Extractor", 1.0, 18),
		wc("masonry", "quantum_age", "Crystal Miner", 1.0, 19),

		// === KNOWLEDGE DOMAIN (starts primitive_age, base 1.0) ===
		wc("knowledge", "primitive_age", "Shaman", 1.0, 0),
		wc("knowledge", "stone_age", "Elder", 1.0, 1),
		wc("knowledge", "bronze_age", "Scribe", 1.0, 2),
		wc("knowledge", "iron_age", "Scholar", 1.0, 3),
		wc("knowledge", "classical_age", "Philosopher", 1.0, 4),
		wc("knowledge", "medieval_age", "Friar", 1.0, 5),
		wc("knowledge", "renaissance_age", "Academician", 1.0, 6),
		wc("knowledge", "colonial_age", "Naturalist", 1.0, 7),
		wc("knowledge", "industrial_age", "Engineer-Scientist", 1.0, 8),
		wc("knowledge", "victorian_age", "Victorian Scholar", 1.0, 9),
		wc("knowledge", "electric_age", "Research Fellow", 1.0, 10),
		wc("knowledge", "atomic_age", "Nuclear Scientist", 1.0, 11),
		wc("knowledge", "modern_age", "Modern Researcher", 1.0, 12),
		wc("knowledge", "information_age", "Data Scientist", 1.0, 13),
		wc("knowledge", "digital_age", "AI Researcher", 1.0, 14),
		wc("knowledge", "cyberpunk_age", "Cyber-Scholar", 1.0, 15),
		wc("knowledge", "fusion_age", "Fusion Theorist", 1.0, 16),
		wc("knowledge", "space_age", "Orbital Researcher", 1.0, 17),
		wc("knowledge", "interstellar_age", "Stellar Scientist", 1.0, 18),
		wc("knowledge", "galactic_age", "Galactic Researcher", 1.0, 19),
		wc("knowledge", "quantum_age", "Quantum Theorist", 1.0, 20),

		// === FAITH DOMAIN (starts primitive_age for early shrine/altar buildings, base tiered) ===
		// Early tiers added so primitive-through-classical faith workers get correct class names
		// instead of falling back to the tier-0 Acolyte (FoodCost=2.0) via the domain-start fallback.
		WorkerClassDef{Domain: "faith", AgeKey: "primitive_age", ClassName: "Devotee", FoodCost: 0.08, Multiplier: 1.0},
		WorkerClassDef{Domain: "faith", AgeKey: "stone_age", ClassName: "Believer", FoodCost: 0.12, Multiplier: 1.0},
		WorkerClassDef{Domain: "faith", AgeKey: "bronze_age", ClassName: "Worshipper", FoodCost: 0.18, Multiplier: 1.2},
		WorkerClassDef{Domain: "faith", AgeKey: "iron_age", ClassName: "Celebrant", FoodCost: 0.27, Multiplier: 1.2},
		WorkerClassDef{Domain: "faith", AgeKey: "classical_age", ClassName: "Initiate", FoodCost: 0.40, Multiplier: 1.5},
		wc("faith", "medieval_age", "Acolyte", 2.0, 0),
		wc("faith", "renaissance_age", "Monk", 2.0, 1),
		wc("faith", "colonial_age", "Missionary", 2.0, 2),
		wc("faith", "industrial_age", "Revivalist", 2.0, 3),
		wc("faith", "victorian_age", "Parish Priest", 2.0, 4),
		wc("faith", "electric_age", "Evangelical", 2.0, 5),
		wc("faith", "atomic_age", "Atomic Priest", 2.0, 6),
		wc("faith", "modern_age", "Modern Shepherd", 2.0, 7),
		wc("faith", "information_age", "Digital Devotee", 2.0, 8),
		wc("faith", "digital_age", "Virtual Cleric", 2.0, 9),
		wc("faith", "cyberpunk_age", "Cyber Cleric", 2.0, 10),
		wc("faith", "fusion_age", "Plasma Prophet", 2.0, 11),
		wc("faith", "space_age", "Star Preacher", 2.0, 12),
		wc("faith", "interstellar_age", "Interstellar Mystic", 2.0, 13),
		wc("faith", "galactic_age", "Galactic High Priest", 2.0, 14),
		wc("faith", "quantum_age", "Quantum Sage", 2.0, 15),

		// === MILITARY DOMAIN (starts iron_age, base 2.0) ===
		wc("military", "iron_age", "Soldier", 2.0, 0),
		wc("military", "classical_age", "Legionary", 2.0, 1),
		wc("military", "medieval_age", "Knight", 2.0, 2),
		wc("military", "renaissance_age", "Musketeer", 2.0, 3),
		wc("military", "colonial_age", "Colonial Marine", 2.0, 4),
		wc("military", "industrial_age", "Industrial Rifleman", 2.0, 5),
		wc("military", "victorian_age", "Victorian Guard", 2.0, 6),
		wc("military", "electric_age", "Electric Trooper", 2.0, 7),
		wc("military", "atomic_age", "Atomic Soldier", 2.0, 8),
		wc("military", "modern_age", "Modern Soldier", 2.0, 9),
		wc("military", "information_age", "Information Warrior", 2.0, 10),
		wc("military", "digital_age", "Digital Soldier", 2.0, 11),
		wc("military", "cyberpunk_age", "Cyber Warrior", 2.0, 12),
		wc("military", "fusion_age", "Plasma Trooper", 2.0, 13),
		wc("military", "space_age", "Space Marine", 2.0, 14),
		wc("military", "interstellar_age", "Interstellar Commando", 2.0, 15),
		wc("military", "galactic_age", "Galactic Guardian", 2.0, 16),
		wc("military", "quantum_age", "Quantum Soldier", 2.0, 17),

		// === TRADE DOMAIN (starts bronze_age, base 1.0) ===
		wc("trade", "bronze_age", "Peddler", 1.0, 0),
		wc("trade", "iron_age", "Merchant", 1.0, 1),
		wc("trade", "classical_age", "Trader", 1.0, 2),
		wc("trade", "medieval_age", "Nobleman", 1.0, 3),
		wc("trade", "renaissance_age", "Banker", 1.0, 4),
		wc("trade", "colonial_age", "Colonial Merchant", 1.0, 5),
		wc("trade", "industrial_age", "Industrialist", 1.0, 6),
		wc("trade", "victorian_age", "Victorian Trader", 1.0, 7),
		wc("trade", "electric_age", "Electric Broker", 1.0, 8),
		wc("trade", "atomic_age", "Atomic Trader", 1.0, 9),
		wc("trade", "modern_age", "Corporate Trader", 1.0, 10),
		wc("trade", "information_age", "Digital Trader", 1.0, 11),
		wc("trade", "digital_age", "Crypto Broker", 1.0, 12),
		wc("trade", "cyberpunk_age", "Cyber Dealer", 1.0, 13),
		wc("trade", "fusion_age", "Plasma Merchant", 1.0, 14),
		wc("trade", "space_age", "Space Trader", 1.0, 15),
		wc("trade", "interstellar_age", "Interstellar Broker", 1.0, 16),
		wc("trade", "galactic_age", "Galactic Merchant", 1.0, 17),
		wc("trade", "quantum_age", "Quantum Dealer", 1.0, 18),

		// === ENGINEERING DOMAIN (starts bronze_age for early smithy/workshop buildings, base tiered) ===
		// Early tiers added so bronze-through-industrial engineering workers get correct class names
		// instead of falling back to the tier-0 Tinker (FoodCost=8.0) via the domain-start fallback.
		WorkerClassDef{Domain: "engineering", AgeKey: "bronze_age", ClassName: "Apprentice", FoodCost: 0.50, Multiplier: 1.0},
		WorkerClassDef{Domain: "engineering", AgeKey: "iron_age", ClassName: "Craftsman", FoodCost: 0.75, Multiplier: 1.0},
		WorkerClassDef{Domain: "engineering", AgeKey: "classical_age", ClassName: "Artisan", FoodCost: 1.10, Multiplier: 1.2},
		WorkerClassDef{Domain: "engineering", AgeKey: "medieval_age", ClassName: "Engineer", FoodCost: 1.65, Multiplier: 1.2},
		WorkerClassDef{Domain: "engineering", AgeKey: "renaissance_age", ClassName: "Master Eng.", FoodCost: 2.50, Multiplier: 1.5},
		WorkerClassDef{Domain: "engineering", AgeKey: "colonial_age", ClassName: "Mechanic", FoodCost: 3.75, Multiplier: 1.5},
		WorkerClassDef{Domain: "engineering", AgeKey: "industrial_age", ClassName: "Machinist", FoodCost: 5.60, Multiplier: 2.0},
		wc("engineering", "victorian_age", "Tinker", 8.0, 0),
		wc("engineering", "electric_age", "Electrical Engineer", 8.0, 1),
		wc("engineering", "atomic_age", "Nuclear Engineer", 8.0, 2),
		wc("engineering", "modern_age", "Systems Engineer", 8.0, 3),
		wc("engineering", "information_age", "Software Engineer", 8.0, 4),
		wc("engineering", "digital_age", "AI Engineer", 8.0, 5),
		wc("engineering", "cyberpunk_age", "Cyber Engineer", 8.0, 6),
		wc("engineering", "fusion_age", "Plasma Engineer", 8.0, 7),
		wc("engineering", "space_age", "Space Engineer", 8.0, 8),
		wc("engineering", "interstellar_age", "Warp Engineer", 8.0, 9),
		wc("engineering", "galactic_age", "Galactic Engineer", 8.0, 10),
		wc("engineering", "quantum_age", "Quantum Engineer", 8.0, 11),

		// === METALLURGY DOMAIN (starts iron_age, base 2.0) ===
		wc("metallurgy", "iron_age", "Smelter", 2.0, 0),
		wc("metallurgy", "classical_age", "Ironworker", 2.0, 1),
		wc("metallurgy", "medieval_age", "Medieval Smith", 2.0, 2),
		wc("metallurgy", "renaissance_age", "Renaissance Metallurgist", 2.0, 3),
		wc("metallurgy", "colonial_age", "Foundry Worker", 2.0, 4),
		wc("metallurgy", "industrial_age", "Factory Worker", 2.0, 5),
		wc("metallurgy", "victorian_age", "Steam Smelter", 2.0, 6),
		wc("metallurgy", "electric_age", "Electric Smelter", 2.0, 7),
		wc("metallurgy", "atomic_age", "Atomic Metallurgist", 2.0, 8),
		wc("metallurgy", "modern_age", "Modern Metallurgist", 2.0, 9),
		wc("metallurgy", "information_age", "Digital Foundry Worker", 2.0, 10),
		wc("metallurgy", "digital_age", "Digital Smelter", 2.0, 11),
		wc("metallurgy", "cyberpunk_age", "Cyber Forge Worker", 2.0, 12),
		wc("metallurgy", "fusion_age", "Plasma Metallurgist", 2.0, 13),
		wc("metallurgy", "space_age", "Stellar Foundry Worker", 2.0, 14),
		wc("metallurgy", "interstellar_age", "Stellar Smelter", 2.0, 15),
		wc("metallurgy", "galactic_age", "Galactic Metallurgist", 2.0, 16),
		wc("metallurgy", "quantum_age", "Quantum Smelter", 2.0, 17),

		// === ENERGY DOMAIN (starts victorian_age, base 8.0) ===
		wc("energy", "victorian_age", "Stoker", 8.0, 0),
		wc("energy", "electric_age", "Power Worker", 8.0, 1),
		wc("energy", "atomic_age", "Reactor Technician", 8.0, 2),
		wc("energy", "modern_age", "Power Engineer", 8.0, 3),
		wc("energy", "information_age", "Grid Operator", 8.0, 4),
		wc("energy", "digital_age", "Digital Power Manager", 8.0, 5),
		wc("energy", "cyberpunk_age", "Cyber Energy Worker", 8.0, 6),
		wc("energy", "fusion_age", "Fusion Technician", 8.0, 7),
		wc("energy", "space_age", "Solar Engineer", 8.0, 8),
		wc("energy", "interstellar_age", "Dark Energy Worker", 8.0, 9),
		wc("energy", "galactic_age", "Antimatter Specialist", 8.0, 10),
		wc("energy", "quantum_age", "Zero-Point Engineer", 8.0, 11),

		// === HACKER DOMAIN (starts information_age, base 16.0) ===
		wc("hacker", "information_age", "Script Kiddie", 16.0, 0),
		wc("hacker", "digital_age", "Coder", 16.0, 1),
		wc("hacker", "cyberpunk_age", "Black Hat", 16.0, 2),
		wc("hacker", "fusion_age", "AI Hacker", 16.0, 3),
		wc("hacker", "space_age", "Orbital Hacker", 16.0, 4),
		wc("hacker", "interstellar_age", "Interstellar Netrunner", 16.0, 5),
		wc("hacker", "galactic_age", "Galactic Hacker", 16.0, 6),
		wc("hacker", "quantum_age", "Quantum Hacker", 16.0, 7),

		// === ASTRONAUT DOMAIN (starts space_age, base 32.0) ===
		wc("astronaut", "space_age", "Cadet", 32.0, 0),
		wc("astronaut", "interstellar_age", "Interstellar Pilot", 32.0, 1),
		wc("astronaut", "galactic_age", "Galactic Explorer", 32.0, 2),
		wc("astronaut", "quantum_age", "Quantum Astronaut", 32.0, 3),
	}
}

// WorkerClassByDomainAndAge returns the WorkerClassDef for the given domain and age.
// First tries an exact age match; if none found, falls back to the highest-tier class
// whose age order is ≤ the current age (so lumber workers in primitive_age still show
// "Gatherer"). If the current age is before the domain's start age, returns tier-0.
func WorkerClassByDomainAndAge(domain, ageKey string) (WorkerClassDef, bool) {
	classes := WorkerClasses()

	// Exact match first.
	for _, wc := range classes {
		if wc.Domain == domain && wc.AgeKey == ageKey {
			return wc, true
		}
	}

	// Fallback: pick the highest-order age entry for this domain that is ≤ current age.
	ages := AgeByKey()
	currentOrder := 1<<30 // treat unknown age as very high so all classes qualify
	if a, ok := ages[ageKey]; ok {
		currentOrder = a.Order
	}

	bestOrder := -1
	bestIdx := -1
	for i, wc := range classes {
		if wc.Domain != domain {
			continue
		}
		if a, ok := ages[wc.AgeKey]; ok && a.Order <= currentOrder && a.Order > bestOrder {
			bestOrder = a.Order
			bestIdx = i
		}
	}
	if bestIdx >= 0 {
		return classes[bestIdx], true
	}

	// Current age is before domain's start — return tier-0 for this domain.
	for _, wc := range classes {
		if wc.Domain == domain {
			return wc, true
		}
	}
	return WorkerClassDef{}, false
}

// WorkerDomains returns the canonical ordered list of all 12 worker domain keys.
// This is used to initialise WorkerState on new games and to drive iteration
// order in the UI — keep this list in sync with the domains in WorkerClasses().
func WorkerDomains() []string {
	return []string{
		"food", "lumber", "masonry", "knowledge",
		"faith", "military", "trade", "engineering",
		"metallurgy", "energy", "hacker", "astronaut",
	}
}
