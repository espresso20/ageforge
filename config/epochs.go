package config

// EpochDef defines a meta-progression era spanning multiple ages.
// There are 7 epochs. Most span 3 ages; the final Cosmic Era spans 4.
// At the boundary between epochs the engine rolls a special epoch event
// (good/bad/catastrophe) that shapes the era's permanent flavour.
type EpochDef struct {
	Name            string
	Key             string
	Order           int      // 0-indexed position in the epoch sequence
	Ages            []string // age keys that belong to this epoch (3 per epoch; Cosmic has 4)
	Icon            string   // single Unicode character shown in the UI (e.g. "◈", "⚡")
	Color           string   // tview dynamic color name used for epoch labels (e.g. "gold", "cyan")
	PrimaryResource string   // dominant structural/construction resource for this era
	EnergyResource  string   // dominant energy/fuel resource for this era
	CatastropheKey  string   // key used to look up the catastrophe event definition for this epoch
	Description     string
}

// Epochs returns all 7 epoch definitions in order.
func Epochs() []EpochDef {
	return []EpochDef{
		{
			Name: "Stone Era", Key: "stone_era", Order: 0,
			Ages: []string{"primitive_age", "stone_age", "bronze_age"},
			Icon: "◈", Color: "white",
			PrimaryResource: "wood", EnergyResource: "food",
			CatastropheKey: "meteor_impact",
			Description:    "Humanity's first steps — wood, stone, and fire.",
		},
		{
			Name: "Iron Era", Key: "iron_era", Order: 1,
			Ages: []string{"iron_age", "classical_age", "medieval_age"},
			Icon: "⚔", Color: "red",
			PrimaryResource: "iron", EnergyResource: "coal",
			CatastropheKey: "barbarian_invasion",
			Description:    "Empires of iron and faith rise and fall.",
		},
		{
			Name: "Steel Era", Key: "steel_era", Order: 2,
			Ages: []string{"renaissance_age", "colonial_age", "industrial_age"},
			Icon: "⚙", Color: "yellow",
			PrimaryResource: "steel", EnergyResource: "coal",
			CatastropheKey: "industrial_collapse",
			Description:    "Steam, steel, and global ambition.",
		},
		{
			Name: "Electric Era", Key: "electric_era", Order: 3,
			Ages: []string{"victorian_age", "electric_age", "atomic_age"},
			Icon: "⚡", Color: "lightblue",
			PrimaryResource: "steel", EnergyResource: "electricity",
			CatastropheKey: "nuclear_meltdown",
			Description:    "Electricity and the atom reshape civilization.",
		},
		{
			Name: "Digital Era", Key: "digital_era", Order: 4,
			Ages: []string{"modern_age", "information_age", "digital_age"},
			Icon: "▣", Color: "blue",
			PrimaryResource: "data", EnergyResource: "electricity",
			CatastropheKey: "digital_collapse",
			Description:    "Data flows become the rivers of power.",
		},
		{
			Name: "Neon Era", Key: "neon_era", Order: 5,
			Ages: []string{"cyberpunk_age", "fusion_age", "space_age"},
			Icon: "◉", Color: "cyan",
			PrimaryResource: "plasma", EnergyResource: "plasma",
			CatastropheKey: "solar_event",
			Description:    "Augmented reality and the conquest of the solar system.",
		},
		{
			Name: "Cosmic Era", Key: "cosmic_era", Order: 6,
			Ages: []string{"interstellar_age", "galactic_age", "quantum_age", "transcendent_age"},
			Icon: "✦", Color: "magenta",
			PrimaryResource: "dark_matter", EnergyResource: "antimatter",
			CatastropheKey: "reality_fracture",
			Description:    "Between stars and beyond time itself.",
		},
	}
}

// LegacyBonusForEpoch returns the permanent per-resource production bonuses granted by
// succumbing to the catastrophe in a given epoch. Keys are resource keys, values are
// fractional multipliers (e.g. 0.20 = +20%). These are additive with other rate bonuses.
func LegacyBonusForEpoch(epochKey string) map[string]float64 {
	switch epochKey {
	case "stone_era":
		return map[string]float64{"wood": 0.20, "stone": 0.20}
	case "iron_era":
		return map[string]float64{"iron": 0.20}
	case "steel_era":
		return map[string]float64{"steel": 0.25, "coal": 0.25}
	case "electric_era":
		return map[string]float64{"electricity": 0.25, "uranium": 0.25}
	case "digital_era":
		return map[string]float64{"data": 0.30, "titanium_ore": 0.30}
	case "neon_era":
		return map[string]float64{"plasma": 0.30, "dark_matter_crystals": 0.30}
	case "cosmic_era":
		return map[string]float64{"dark_matter": 0.35}
	}
	return nil
}

// CatastropheInfo returns the display name and flavor text for an epoch's catastrophe.
func CatastropheInfo(epochKey string) (name, flavor string) {
	switch epochKey {
	case "stone_era":
		return "The Great Meteor",
			"A celestial body has struck your settlement. The sky burns. Your people scatter."
	case "iron_era":
		return "The Great Plague",
			"A devastating plague sweeps your cities. The streets fall silent."
	case "steel_era":
		return "The World War",
			"Industrial warfare tears civilization apart. The factories are ash."
	case "electric_era":
		return "The Nuclear Exchange",
			"Nations unleash the atom. Cities become glass."
	case "digital_era":
		return "The Great Hack",
			"Every system falls silent. The AIs turn on their creators."
	case "neon_era":
		return "Corporate Armageddon",
			"The megacorps end the world with a fusion bomb."
	case "cosmic_era":
		return "The Reality Tear",
			"Exotic matter destabilizes spacetime. Reality cracks open."
	}
	return "Unknown Catastrophe", "Something terrible has happened."
}

// EpochEventDef defines a major transition event that fires exactly once per epoch
// (at the first age advance that crosses into a new epoch). These are separate from
// the regular random event pool (RandomEvents in events.go).
//
// Type values and their selection criteria:
//
//	"good_minor"      — always eligible; lower culture gates these out first
//	"good_major"      — requires medium culture fill %
//	"good_legendary"  — requires high culture fill % (rare)
//	"bad_challenging" — bad event; more likely at low culture
//
// Duration == 0 means the effect is instant (one-time apply); Duration > 0 means
// the effect persists in ActiveEvents for that many ticks.
type EpochEventDef struct {
	Key        string // unique key used for EpochEventRecord.EventKey
	Name       string
	FlavorText string // dramatic one-liner shown in the age splash and epoch overlay
	Type       string // "good_minor" | "good_major" | "good_legendary" | "bad_challenging"
	Duration   int    // ticks the effect lasts; 0 = instant
}

// GoodEpochEvents returns the 10 good epoch transition events (minor/major/legendary).
func GoodEpochEvents() []EpochEventDef {
	return []EpochEventDef{
		// --- Minor (any culture level) ---
		{
			Key: "age_of_plenty", Name: "Age of Plenty", Type: "good_minor",
			FlavorText: "Harvests overflow, rivers run clear. Production surges across all domains.",
			Duration:   216, // ~7 min real
		},
		{
			Key: "population_surge", Name: "Population Surge", Type: "good_minor",
			FlavorText: "A generation of plenty — workers flock to your banner.",
			Duration:   0, // instant: +15% workers added
		},
		{
			Key: "ancient_cache", Name: "Ancient Cache", Type: "good_minor",
			FlavorText: "Explorers uncover a sealed vault of ancient stores.",
			Duration:   0, // instant: fills 40% of every resource's storage
		},
		{
			Key: "trade_winds", Name: "Trade Winds", Type: "good_minor",
			FlavorText: "A favorable wind opens all trade routes and multiplies gold income.",
			Duration:   144, // ~5 min
		},
		{
			Key: "cultural_festival", Name: "Cultural Festival", Type: "good_minor",
			FlavorText: "A grand festival unites your people. Culture and faith surge.",
			Duration:   144, // ~5 min
		},
		// --- Major (medium culture required) ---
		{
			Key: "grand_discovery", Name: "The Grand Discovery", Type: "good_major",
			FlavorText: "Scholars make a breakthrough. Three technologies complete themselves.",
			Duration:   0, // instant: complete 3 free techs
		},
		{
			Key: "worker_innovation", Name: "Worker Innovation", Type: "good_major",
			FlavorText: "A new method transforms your workforce — output climbs permanently.",
			Duration:   0, // instant: permanent +10% production_all
		},
		{
			Key: "architects_gift", Name: "The Architect's Gift", Type: "good_major",
			FlavorText: "A master architect offers designs freely. Ten buildings rise without cost.",
			Duration:   0, // instant: 10 free buildings
		},
		{
			Key: "peaceful_century", Name: "Peaceful Century", Type: "good_major",
			FlavorText: "An era of peace descends. No disasters. Production climbs.",
			Duration:   288, // ~10 min
		},
		// --- Legendary (high culture, rare) ---
		{
			Key: "epoch_blessing", Name: "Epoch Blessing", Type: "good_legendary",
			FlavorText: "The heavens smile on your civilization. A permanent golden age begins.",
			Duration:   0, // instant: permanent +15% production_all; recorded in history
		},
	}
}

// ChallengingEpochEvents returns the 8 bad (non-catastrophe) epoch transition events.
func ChallengingEpochEvents() []EpochEventDef {
	return []EpochEventDef{
		{
			Key: "the_famine", Name: "The Famine", Type: "bad_challenging",
			FlavorText: "Crops wither. Granaries empty. Your people grow desperate.",
			Duration:   120,
		},
		{
			Key: "merchant_betrayal", Name: "Merchant Betrayal", Type: "bad_challenging",
			FlavorText: "Your trading partners vanish with the gold. Markets collapse.",
			Duration:   72,
		},
		{
			Key: "the_great_fire", Name: "The Great Fire", Type: "bad_challenging",
			FlavorText: "Flames sweep the city. Buildings crumble before the dawn.",
			Duration:   0, // instant: 8 random buildings destroyed
		},
		{
			Key: "epidemic", Name: "Epidemic", Type: "bad_challenging",
			FlavorText: "A plague moves swiftly through your population. Workers fall silent.",
			Duration:   180,
		},
		{
			Key: "resource_drought", Name: "Resource Drought", Type: "bad_challenging",
			FlavorText: "The era's vital resource dries up. Supply chains buckle.",
			Duration:   90,
		},
		{
			Key: "political_instability", Name: "Political Instability", Type: "bad_challenging",
			FlavorText: "Factions tear at the throne. Faith collapses. The military is paralyzed.",
			Duration:   60,
		},
		{
			Key: "economic_crash", Name: "Economic Crash", Type: "bad_challenging",
			FlavorText: "Markets implode. Gold vanishes. Building costs skyrocket.",
			Duration:   216,
		},
		{
			Key: "the_dark_age", Name: "The Dark Age", Type: "bad_challenging",
			FlavorText: "Knowledge fades. Research halts. Your scholars fall silent.",
			Duration:   144,
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

// EpochEventByKey returns a map of event key -> EpochEventDef across all epoch event pools.
func EpochEventByKey() map[string]EpochEventDef {
	m := make(map[string]EpochEventDef)
	for _, ev := range GoodEpochEvents() {
		m[ev.Key] = ev
	}
	for _, ev := range ChallengingEpochEvents() {
		m[ev.Key] = ev
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
