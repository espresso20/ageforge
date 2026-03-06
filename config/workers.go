package config

// WorkerClassDef defines a single tier of a worker domain, tied to a specific age.
// Each domain has one class per age it spans (from its start age through quantum_age).
// Food cost and production multiplier scale geometrically per tier:
//   FoodCost = baseFoodCost × 1.5^tier
//   Multiplier = 2.0^tier
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
		fc *= 1.5
		mp *= 2.0
	}
	return WorkerClassDef{Domain: domain, AgeKey: ageKey, ClassName: className, FoodCost: fc, Multiplier: mp}
}

// WorkerClasses returns all worker class definitions across all 12 domains.
// Domains and their start ages / base food costs:
//   food        primitive   0.08  — Forager → Quantum Harvester
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

		// === FOOD DOMAIN (starts primitive_age, base food cost 0.05) ===
		// Base cost lowered from 0.08 to 0.05 so gathering camps can sustain larger populations.
		// At tier 0 (primitive): FoodCost = 0.05; tier 1 (stone): 0.075; tier 2 (bronze): 0.11; etc.
		wc("food", "primitive_age", "Forager", 0.05, 0),
		wc("food", "stone_age", "Farmhand", 0.05, 1),
		wc("food", "bronze_age", "Cultivator", 1.0, 2),
		wc("food", "iron_age", "Laborer", 1.0, 3),
		wc("food", "classical_age", "Peasant", 1.0, 4),
		wc("food", "medieval_age", "Serf", 1.0, 5),
		wc("food", "renaissance_age", "Plowman", 1.0, 6),
		wc("food", "colonial_age", "Colonial Farmer", 1.0, 7),
		wc("food", "industrial_age", "Factory Hand", 1.0, 8),
		wc("food", "victorian_age", "Agricultural Worker", 1.0, 9),
		wc("food", "electric_age", "Electric Farmer", 1.0, 10),
		wc("food", "atomic_age", "Atomic Agronomist", 1.0, 11),
		wc("food", "modern_age", "Modern Farmer", 1.0, 12),
		wc("food", "information_age", "Digital Cultivator", 1.0, 13),
		wc("food", "digital_age", "AI Agronomist", 1.0, 14),
		wc("food", "cyberpunk_age", "Aug Harvester", 1.0, 15),
		wc("food", "fusion_age", "Bio-Farmer", 1.0, 16),
		wc("food", "space_age", "Zero-G Farmer", 1.0, 17),
		wc("food", "interstellar_age", "Stellar Cultivator", 1.0, 18),
		wc("food", "galactic_age", "Galactic Farmer", 1.0, 19),
		wc("food", "quantum_age", "Quantum Harvester", 1.0, 20),

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

		// === FAITH DOMAIN (starts medieval_age, base 2.0) ===
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

		// === ENGINEERING DOMAIN (starts victorian_age, base 8.0) ===
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
// Returns (def, true) if found, or (zero, false) if no class exists for that combination.
func WorkerClassByDomainAndAge(domain, ageKey string) (WorkerClassDef, bool) {
	for _, wc := range WorkerClasses() {
		if wc.Domain == domain && wc.AgeKey == ageKey {
			return wc, true
		}
	}
	return WorkerClassDef{}, false
}

// WorkerDomains returns the list of all unique worker domain keys.
func WorkerDomains() []string {
	return []string{
		"food", "lumber", "masonry", "knowledge",
		"faith", "military", "trade", "engineering",
		"metallurgy", "energy", "hacker", "astronaut",
	}
}
