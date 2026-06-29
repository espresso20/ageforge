package config

// AwakeningDef defines a one-time "Age Awakening" — a pivotal moment that fires
// exactly once per civilisation cycle (prestige run) the FIRST time the player
// advances INTO that epoch's trigger age. There are 7 awakenings, one per epoch,
// marking each epoch transition as a defining event in the run.
//
// Unlike the random epoch-transition roll (GoodEpochEvents / ChallengingEpochEvents,
// which is a gamble gated on faith/culture), an awakening is deterministic: it always
// fires, always grants a modest thematic boost, and never carries a downside.
//
// The effect is delivered through the existing ActiveEvent / InjectEvent mechanism —
// a TEMPORARY boost that decays after Duration ticks. Two effect shapes are used,
// both of which active events already feed into recalculateRates:
//   - Type "production_all", Value v   → +v multiplier on ALL positive rates while active
//   - Type "production", Target res, Value v → +v FLAT additive to that resource's rate
//
// No new effect engine is introduced; these are the same Effect types regular and
// epoch events already use.
type AwakeningDef struct {
	Key        string   // unique key, also used for the ActiveEvent key + fired-set tracking
	EpochKey   string   // epoch this awakening belongs to (one per epoch)
	TriggerAge string   // age key whose FIRST entry fires this awakening
	Name       string   // display name shown in the log/splash
	FlavorText string   // 1-2 line dry/evocative description, matches the game's tone
	Duration   int      // ticks the temporary effect lasts (ticks are ~2s of real time)
	Effects    []Effect // temporary boost applied via InjectEvent
}

// Awakenings returns all 7 age-awakening definitions, one per epoch, in epoch order.
//
// Trigger ages are chosen as the first "meaningful" age of each epoch. The Stone Era
// triggers on the Stone Age rather than the Primitive Age (the tutorial/starting age),
// and the Steel Era's Steam Breakthrough triggers on the Industrial Age where steam
// power thematically arrives. The remaining five trigger on their epoch's first age.
//
// Duration tuning (1 tick ≈ 2s real time):
//
//	~250 ticks ≈ 8 min, ~500 ticks ≈ 16 min, ~200 ticks ≈ 6.5 min.
func Awakenings() []AwakeningDef {
	return []AwakeningDef{
		{
			Key:      "awakening_pottery_mastery",
			EpochKey: "stone_era", TriggerAge: "stone_age",
			Name:       "Pottery Mastery",
			FlavorText: "Clay yields to patient hands. Sealed vessels hold the harvest through the lean months — and the surplus, for once, keeps.",
			Duration:   250,
			Effects: []Effect{
				{Type: "production", Target: "food", Value: 1.0},
				{Type: "production", Target: "stone", Value: 0.5},
			},
		},
		{
			Key:      "awakening_metallurgy",
			EpochKey: "iron_era", TriggerAge: "iron_age",
			Name:       "Discovery of Metallurgy",
			FlavorText: "The forge runs hotter than any fire before it. Ore that once defied you now bleeds into ingots — and the smiths cannot smelt fast enough.",
			Duration:   500,
			Effects: []Effect{
				{Type: "production", Target: "iron", Value: 2.0},
			},
		},
		{
			Key:      "awakening_steam_breakthrough",
			EpochKey: "steel_era", TriggerAge: "industrial_age",
			Name:       "Steam Breakthrough",
			FlavorText: "Pressure, piston, purpose. The first engine coughs, catches, and roars — and every workshop in the land suddenly works twice as hard.",
			Duration:   200,
			Effects: []Effect{
				{Type: "production_all", Value: 0.25},
			},
		},
		{
			Key:      "awakening_electrification",
			EpochKey: "electric_era", TriggerAge: "victorian_age",
			Name:       "Electrification",
			FlavorText: "Night surrenders. The grid hums to life, lamps bloom across the skyline, and machines that never sleep take up the long shift.",
			Duration:   300,
			Effects: []Effect{
				{Type: "production", Target: "electricity", Value: 2.0},
				{Type: "production_all", Value: 0.10},
			},
		},
		{
			Key:      "awakening_information_age",
			EpochKey: "digital_era", TriggerAge: "modern_age",
			Name:       "Information Age Dawns",
			FlavorText: "Knowledge stops being scarce. The networks wake, the archives open, and insight compounds faster than anyone can read it.",
			Duration:   300,
			Effects: []Effect{
				{Type: "production", Target: "data", Value: 2.0},
				{Type: "production", Target: "knowledge", Value: 1.0},
			},
		},
		{
			Key:      "awakening_cybernetic",
			EpochKey: "neon_era", TriggerAge: "cyberpunk_age",
			Name:       "Cybernetic Awakening",
			FlavorText: "Flesh and circuit reach an accord. Augmented crews never tire, never blink — and the city's output climbs to a neon-lit fever pitch.",
			Duration:   250,
			Effects: []Effect{
				{Type: "production_all", Value: 0.20},
			},
		},
		{
			Key:      "awakening_first_contact",
			EpochKey: "cosmic_era", TriggerAge: "interstellar_age",
			Name:       "First Contact Signal",
			FlavorText: "A pattern threads through the static — too regular to be noise, too strange to be us. Whatever sent it, your engineers cannot stop listening.",
			Duration:   400,
			Effects: []Effect{
				{Type: "production", Target: "dark_matter", Value: 1.5},
				{Type: "production_all", Value: 0.10},
			},
		},
	}
}

// AwakeningByKey returns a map of awakening key -> AwakeningDef.
func AwakeningByKey() map[string]AwakeningDef {
	m := make(map[string]AwakeningDef)
	for _, a := range Awakenings() {
		m[a.Key] = a
	}
	return m
}

// AwakeningForAge returns the awakening whose TriggerAge matches ageKey, and true,
// or a zero AwakeningDef and false if no awakening triggers on that age.
func AwakeningForAge(ageKey string) (AwakeningDef, bool) {
	for _, a := range Awakenings() {
		if a.TriggerAge == ageKey {
			return a, true
		}
	}
	return AwakeningDef{}, false
}
