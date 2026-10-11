package config

// EpochDef defines a meta-progression era spanning multiple ages.
// There are 7 epochs. Most span 3 ages; the final Cosmic Era spans 4.
// At the boundary between epochs the engine rolls a special epoch event
// (good or challenging) that shapes the era's permanent flavour. A catastrophe
// is not a transition event: on entering an era from the Iron Era on, a hidden
// roll decides whether a doom is fated to strike somewhere inside it
// (game/fate.go).
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
			Description:    "Humanity's first steps with wood, stone and fire.",
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
			Description:    "Steam and steel carry trade around the globe.",
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
			Description:    "Data becomes the thing everyone wants.",
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
			"Nations turn the atom on each other. Cities become glass."
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
	// Rates are the event's timed changes to a resource's rate, as sizes
	// (EventRate: a share of the town's own income of the resource, see
	// event_size.go). The engine turns each into an amount per tick when the
	// event fires. The flavor text states the share; the log line under it
	// states the amount. A Target of "" stands for the era's primary
	// resource (Resource Drought).
	Rates []Effect
}

// epochEventAge is the age epoch events are timed for. They fire on entering
// an era, and the first era they can fire in, the Iron Era, already runs at
// PacingStretch, so every one does.
const epochEventAge = "iron_age"

// stretchEpochEvents re-times epoch events for the pacing curve: each typed
// Duration is the base-curve value, stretched by StretchTicks, and "{dur}"
// in the flavor text becomes the stretched duration.
func stretchEpochEvents(defs []EpochEventDef) []EpochEventDef {
	for i := range defs {
		defs[i].FlavorText = withDuration(defs[i].FlavorText, defs[i].Duration)
	}
	return defs
}

// GoodEpochEvents returns the 10 good epoch transition events (minor/major/legendary).
func GoodEpochEvents() []EpochEventDef {
	return stretchEpochEvents([]EpochEventDef{
		// --- Minor (any culture level) ---
		{
			Key: "age_of_plenty", Name: "Age of Plenty", Type: "good_minor",
			FlavorText: "Harvests overflow and the rivers run clear. All production +100% for {dur}.",
			Duration:   562, // ~7 min on the base curve
		},
		{
			Key: "population_surge", Name: "Population Surge", Type: "good_minor",
			FlavorText: "A generation of plenty brings new hands. Your population grows by 15%.",
			Duration:   0, // instant: +15% workers added
		},
		{
			Key: "ancient_cache", Name: "Ancient Cache", Type: "good_minor",
			FlavorText: "Explorers open a sealed vault of ancient stores. Every unlocked resource gains 40% of its storage.",
			Duration:   0, // instant: fills 40% of every resource's storage
		},
		{
			Key: "trade_winds", Name: "Trade Winds", Type: "good_minor",
			FlavorText: "A steady wind fills the sails. Gold production +50% for {dur}.",
			Duration:   374, // ~5 min on the base curve
			Rates:      []Effect{{Type: EventRate, Target: "gold", Value: 0.5}},
		},
		{
			Key: "cultural_festival", Name: "Cultural Festival", Type: "good_minor",
			FlavorText: "A grand festival brings everyone together. Culture +30% and faith +20% of what you hold, then culture production +50% and faith production +50% for {dur}.",
			Duration:   374, // ~5 min on the base curve
			Rates: []Effect{
				{Type: EventRate, Target: "culture", Value: 0.5},
				{Type: EventRate, Target: "faith", Value: 0.5},
			},
		},
		// --- Major (medium culture required) ---
		{
			Key: "grand_discovery", Name: "The Grand Discovery", Type: "good_major",
			FlavorText: "Scholars make a breakthrough. Up to 3 techs you can research now are completed for free.",
			Duration:   0, // instant: complete 3 free techs
		},
		{
			Key: "worker_innovation", Name: "Worker Innovation", Type: "good_major",
			FlavorText: "A better way of working spreads through the workshops. All production +10%, permanently.",
			Duration:   0, // instant: permanent +10% production_all
		},
		{
			Key: "architects_gift", Name: "The Architect's Gift", Type: "good_major",
			FlavorText: "A master architect gives away his designs. You get 10 free copies of the building you have the most of.",
			Duration:   0, // instant: 10 free buildings
		},
		{
			Key: "peaceful_century", Name: "Peaceful Century", Type: "good_major",
			FlavorText: "An era of peace settles in. All production +20% for {dur}.",
			Duration:   749, // ~10 min on the base curve
		},
		// --- Legendary (high culture, rare) ---
		{
			Key: "epoch_blessing", Name: "Epoch Blessing", Type: "good_legendary",
			FlavorText: "The heavens smile on your civilization. All production +15%, permanently.",
			Duration:   0, // instant: permanent +15% production_all; recorded in history
		},
	})
}

// ChallengingEpochEvents returns the 8 bad (non-catastrophe) epoch transition events.
func ChallengingEpochEvents() []EpochEventDef {
	return stretchEpochEvents([]EpochEventDef{
		{
			Key: "the_famine", Name: "The Famine", Type: "bad_challenging",
			FlavorText: "Crops wither and the granaries run empty. Food production -30% for {dur}.",
			Duration:   312,
			Rates:      []Effect{{Type: EventRate, Target: "food", Value: -0.3}},
		},
		{
			Key: "merchant_betrayal", Name: "Merchant Betrayal", Type: "bad_challenging",
			FlavorText: "Your trading partners vanish with 50% of your gold. Gold production -30% for {dur}.",
			Duration:   187,
			Rates:      []Effect{{Type: EventRate, Target: "gold", Value: -0.3}},
		},
		{
			Key: "the_great_fire", Name: "The Great Fire", Type: "bad_challenging",
			FlavorText: "Flames sweep the city. Up to 8 buildings burn down.",
			Duration:   0, // instant: 8 random buildings destroyed
		},
		{
			Key: "epidemic", Name: "Epidemic", Type: "bad_challenging",
			FlavorText: "A plague moves through your population. 20% of your workers die, and food production -20% for {dur}.",
			Duration:   468,
			Rates:      []Effect{{Type: EventRate, Target: "food", Value: -0.2}},
		},
		{
			Key: "resource_drought", Name: "Resource Drought", Type: "bad_challenging",
			FlavorText: "The epoch's main building material runs short. Its production -40% for {dur}.",
			Duration:   234,
			Rates:      []Effect{{Type: EventRate, Target: "", Value: -0.4}},
		},
		{
			Key: "political_instability", Name: "Political Instability", Type: "bad_challenging",
			FlavorText: "Rival courts tear at the throne. You lose 60% of your faith, and knowledge production -30% for {dur}.",
			Duration:   156,
			Rates:      []Effect{{Type: EventRate, Target: "knowledge", Value: -0.3}},
		},
		{
			Key: "economic_crash", Name: "Economic Crash", Type: "bad_challenging",
			FlavorText: "Markets implode. You lose 50% of your gold, and gold production -40% for {dur}.",
			Duration:   562,
			Rates:      []Effect{{Type: EventRate, Target: "gold", Value: -0.4}},
		},
		{
			Key: "the_dark_age", Name: "The Dark Age", Type: "bad_challenging",
			FlavorText: "Your scholars fall silent. Current research is canceled with no refund, you lose 80% of your knowledge, and knowledge production -40% for {dur}.",
			Duration:   374,
			Rates:      []Effect{{Type: EventRate, Target: "knowledge", Value: -0.4}},
		},
	})
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

// CatastropheGateEpoch is the first epoch in which a civilizational catastrophe
// can strike: the epoch that contains the Iron Age. Nothing is ever fated in
// an earlier epoch (in practice only the Stone Era, which every run starts
// in), and the dev console's /catastrophe is refused there. Good and
// challenging epoch events are unaffected by the gate.
const CatastropheGateEpoch = "iron_era"

// CatastropheAllowed reports whether a catastrophe may occur in the given epoch
// (random roll or voluntary invoke). Unknown epoch keys are not allowed.
func CatastropheAllowed(epochKey string) bool {
	byKey := EpochByKey()
	ep, ok := byKey[epochKey]
	if !ok {
		return false
	}
	return ep.Order >= byKey[CatastropheGateEpoch].Order
}

// FateAllowed reports whether a doom can be fated inside epochKey: past the
// Iron gate, the Cosmic Era included (its doom is the Reality Tear; the Last
// Passage at prestige is separate, game/last_passage.go). Every epoch rolls a
// fate on entry; before the gate it can only hold a false prophet.
func FateAllowed(epochKey string) bool {
	return CatastropheAllowed(epochKey)
}

// NextEpoch returns the epoch that follows epochKey in order, or ok=false when
// epochKey is the final epoch or unknown.
func NextEpoch(epochKey string) (EpochDef, bool) {
	all := Epochs()
	for i, e := range all {
		if e.Key == epochKey && i+1 < len(all) {
			return all[i+1], true
		}
	}
	return EpochDef{}, false
}

// LastPassageKey is the catastrophe key of the Last Passage, the Cosmic Era's
// passage. The Cosmic Era has no next epoch, so its passage is prestige itself:
// the end of the run can bring one last catastrophe (game/last_passage.go).
const LastPassageKey = "last_passage"

// LastPassageInfo returns the display name and flavor text of the Last Passage.
func LastPassageInfo() (name, flavor string) {
	return "The Last Passage",
		"The civilization reaches the end of its road, and something waiting there reaches back."
}

// IsFinalEpoch reports whether epochKey is the last epoch, whose passage is
// prestige rather than an epoch transition. Unknown keys are not final.
func IsFinalEpoch(epochKey string) bool {
	if _, ok := EpochByKey()[epochKey]; !ok {
		return false
	}
	_, hasNext := NextEpoch(epochKey)
	return !hasNext
}
