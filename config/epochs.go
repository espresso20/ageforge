package config

// EpochDef defines a meta-progression era spanning 3 ages.
// 7 epochs × 3 ages = 21 of the 22 ages; Transcendent Age belongs to Cosmic Era.
type EpochDef struct {
	Name            string
	Key             string
	Order           int
	Ages            []string // age keys in this epoch (3 per epoch, 4 for Cosmic)
	Icon            string   // display symbol
	Color           string   // tview color tag for UI
	PrimaryResource string   // dominant construction/structural resource
	EnergyResource  string   // dominant energy/fuel resource
	CatastropheKey  string   // flavor key for the catastrophe event in this epoch
	Description     string
}

// Epochs returns all 7 epoch definitions in order.
func Epochs() []EpochDef {
	return []EpochDef{
		{
			Name: "Stone Era", Key: "stone_era", Order: 0,
			Ages:            []string{"primitive_age", "stone_age", "bronze_age"},
			Icon:            "◈", Color: "white",
			PrimaryResource: "wood", EnergyResource: "food",
			CatastropheKey: "meteor_impact",
			Description:    "Humanity's first steps — wood, stone, and fire.",
		},
		{
			Name: "Iron Era", Key: "iron_era", Order: 1,
			Ages:            []string{"iron_age", "classical_age", "medieval_age"},
			Icon:            "⚔", Color: "red",
			PrimaryResource: "iron", EnergyResource: "coal",
			CatastropheKey: "barbarian_invasion",
			Description:    "Empires of iron and faith rise and fall.",
		},
		{
			Name: "Steel Era", Key: "steel_era", Order: 2,
			Ages:            []string{"renaissance_age", "colonial_age", "industrial_age"},
			Icon:            "⚙", Color: "yellow",
			PrimaryResource: "steel", EnergyResource: "coal",
			CatastropheKey: "industrial_collapse",
			Description:    "Steam, steel, and global ambition.",
		},
		{
			Name: "Electric Era", Key: "electric_era", Order: 3,
			Ages:            []string{"victorian_age", "electric_age", "atomic_age"},
			Icon:            "⚡", Color: "lightblue",
			PrimaryResource: "steel", EnergyResource: "electricity",
			CatastropheKey: "nuclear_meltdown",
			Description:    "Electricity and the atom reshape civilization.",
		},
		{
			Name: "Digital Era", Key: "digital_era", Order: 4,
			Ages:            []string{"modern_age", "information_age", "digital_age"},
			Icon:            "▣", Color: "blue",
			PrimaryResource: "data", EnergyResource: "electricity",
			CatastropheKey: "digital_collapse",
			Description:    "Data flows become the rivers of power.",
		},
		{
			Name: "Neon Era", Key: "neon_era", Order: 5,
			Ages:            []string{"cyberpunk_age", "fusion_age", "space_age"},
			Icon:            "◉", Color: "cyan",
			PrimaryResource: "plasma", EnergyResource: "plasma",
			CatastropheKey: "solar_event",
			Description:    "Augmented reality and the conquest of the solar system.",
		},
		{
			Name: "Cosmic Era", Key: "cosmic_era", Order: 6,
			Ages:            []string{"interstellar_age", "galactic_age", "quantum_age", "transcendent_age"},
			Icon:            "✦", Color: "magenta",
			PrimaryResource: "dark_matter", EnergyResource: "antimatter",
			CatastropheKey: "reality_fracture",
			Description:    "Between stars and beyond time itself.",
		},
	}
}

// EpochByKey returns a map of key -> EpochDef.
func EpochByKey() map[string]EpochDef {
	m := make(map[string]EpochDef)
	for _, e := range Epochs() {
		m[e.Key] = e
	}
	return m
}

// EpochForAge returns the epoch key for a given age key.
func EpochForAge(ageKey string) string {
	for _, e := range Epochs() {
		for _, a := range e.Ages {
			if a == ageKey {
				return e.Key
			}
		}
	}
	return "stone_era" // fallback
}
