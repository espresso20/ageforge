package mapmodel

import "github.com/espresso20/ageforge/config"

// city.go is the Earth arc's city roster: every kind of thing the city
// builds over the land from the Modern Age to the Fusion Age, the age that
// brings it and the last age it stands in. It works like the mover roster:
// an age shows (draws, lists in a legend, names on an inspect line) only
// what it, or an earlier age, brought, so the map never shows what a later
// age will build (the no-spoilers rule), and older features give way as the
// city grows over them.

// CityFeature is one kind of thing the city builds.
type CityFeature uint8

const (
	FeatNone CityFeature = iota
	FeatSuburbs
	FeatPark
	FeatHighway
	FeatGlassTower
	FeatServerFarm
	FeatDish
	FeatFiber
	FeatGarden
	FeatOffice
	FeatScrub
	FeatSmog
	FeatConcrete
	FeatLandfill
	FeatReserve
	FeatDataGlow
	FeatMegablock
	FeatAlley
	FeatCorpTower
	FeatArcology
	FeatNeonSign
	FeatVent
	FeatScrap
	FeatToxic
	FeatCanal
	FeatSkyRail
	FeatMaglevLine
	FeatReactor
	FeatConduit
	FeatLaunchTower
	FeatTether
	FeatCooling
	NumCityFeatures
)

// CityFeatureInfo describes one kind of city feature.
type CityFeatureInfo struct {
	Key   string // "server_farm"
	Name  string // legend label: "server farm"
	Title string // inspect title: "Server farm"
	Sym   Sym    // its symbol, SymNone for a line or a haze the styles draw themselves
	// From is the index of the age that brings it; Until the last age it
	// stands in (-1: to the end of time).
	From, Until int
	Lines       []string // what inspecting it says; each cell keeps one
}

// In reports whether the feature stands in the age with index age.
func (i CityFeatureInfo) In(age int) bool {
	return i.From >= 0 && age >= i.From && (i.Until < 0 || age <= i.Until)
}

// Line is the inspect line for the feature's cell number n.
func (i CityFeatureInfo) Line(n int) string {
	if len(i.Lines) == 0 {
		return ""
	}
	return i.Lines[Hash(int64(n), HashStr(i.Key))%uint64(len(i.Lines))]
}

type cityFeatureDef struct {
	key, name, title string
	sym              Sym
	from, until      string // age keys; until "" stands to the end
	lines            []string
}

var cityFeatureDefs = [NumCityFeatures]cityFeatureDef{
	FeatSuburbs: {key: "suburbs", name: "suburbs", title: "Suburbs", sym: SymSuburb, from: "modern_age", until: "modern_age",
		lines: []string{"Houses with gardens, a drive and a car.", "Commuter houses, quiet by day."}},
	FeatPark: {key: "park", name: "fenced park", title: "Park", sym: SymPark, from: "modern_age", until: "information_age",
		lines: []string{"Still green, fenced into its block.", "A park, railed in on every side."}},
	FeatHighway: {key: "highway", name: "highway", title: "Highway", from: "modern_age", until: "digital_age",
		lines: []string{"Four lanes each way, out of town.", "The highway, never quite empty."}},
	FeatGlassTower: {key: "glass_tower", name: "glass tower", title: "Glass tower", sym: SymGlassTower,
		from: "modern_age", until: "digital_age",
		lines: []string{"Glass and steel, downtown.", "Offices, forty floors of them."}},
	FeatServerFarm: {key: "server_farm", name: "server farm", title: "Server farm", sym: SymServerFarm,
		from: "information_age", until: "digital_age",
		lines: []string{"Racks humming behind blank walls.", "Cooling fans, day and night."}},
	FeatDish: {key: "dish", name: "satellite dish", title: "Satellite dish", sym: SymDish,
		from: "information_age", until: "digital_age",
		lines: []string{"Listening to the sky.", "A dish, pointed at a satellite."}},
	FeatFiber: {key: "fiber", name: "fiber line", title: "Fiber line", from: "information_age", until: "digital_age",
		lines: []string{"Fiber under the street, carrying data.", "Cable trenches, freshly dug."}},
	FeatGarden: {key: "garden", name: "rooftop garden", title: "Rooftop garden", sym: SymGarden,
		from: "information_age", until: "information_age",
		lines: []string{"What is left of the green, up on a roof.", "A garden on the roof, among the vents."}},
	FeatOffice: {key: "office_block", name: "office block", title: "Office block", from: "information_age", until: "digital_age",
		lines: []string{"Screens on every floor.", "Desks, glass and air conditioning."}},
	FeatScrub: {key: "scrub", name: "dry ground", title: "Dry ground", from: "information_age", until: "information_age",
		lines: []string{"Scrub, browned by the smog.", "The grass gone dry past the city."}},
	FeatSmog: {key: "smog", name: "smog", title: "Smog", from: "information_age", until: "fusion_age",
		lines: []string{"The air turning brown at the edges.", "Haze over the outskirts."}},
	FeatConcrete: {key: "concrete", name: "concrete", title: "Concrete", from: "digital_age", until: "digital_age",
		lines: []string{"Paved over, where the fields were.", "Car parks and loading bays."}},
	FeatLandfill: {key: "landfill", name: "landfill", title: "Landfill", sym: SymLandfill, from: "digital_age", until: "digital_age",
		lines: []string{"The city's rubbish, piled high.", "Old screens and packaging, a hill of it."}},
	FeatReserve: {key: "reserve", name: "the reserve", title: "The Reserve", sym: SymPark, from: "digital_age", until: "digital_age",
		lines: []string{"The last green, behind a wall.", "Old trees, kept like a museum."}},
	FeatDataGlow: {key: "data_glow", name: "data glow", title: "Data glow", from: "digital_age", until: "digital_age",
		lines: []string{"Servers warm enough to light the street.", "The glow of a data hall."}},
	FeatMegablock: {key: "megablock", name: "megablock", title: "Megablock", from: "cyberpunk_age", until: "fusion_age",
		lines: []string{"Towers on towers, lit all night.", "A hundred floors, and no one knows the neighbors."}},
	FeatAlley: {key: "alley", name: "alley", title: "Alley", from: "cyberpunk_age", until: "fusion_age",
		lines: []string{"Dark, wet and crowded.", "Steam, noodle stalls and wires overhead."}},
	FeatCorpTower: {key: "corp_tower", name: "megacorp tower", title: "Megacorp tower", from: "cyberpunk_age", until: "fusion_age",
		lines: []string{"Its initials in neon, a block high.", "A megacorp's tower, watching the street."}},
	FeatArcology: {key: "arcology", name: "arcology", title: "Arcology", sym: SymArcology, from: "cyberpunk_age", until: "fusion_age",
		lines: []string{"A city in one building.", "Terraces stacked to the clouds."}},
	FeatNeonSign: {key: "neon_sign", name: "neon sign", title: "Neon sign", sym: SymNeonSign, from: "cyberpunk_age", until: "fusion_age",
		lines: []string{"Buzzing, and flickering now and then.", "A holo-ad, selling something."}},
	FeatVent: {key: "vent", name: "steam vent", title: "Steam vent", sym: SymVent, from: "cyberpunk_age", until: "cyberpunk_age",
		lines: []string{"Steam from the undercity.", "A vent, breathing hot air."}},
	FeatScrap: {key: "scrap", name: "scrap heap", title: "Scrap heap", sym: SymScrap, from: "cyberpunk_age", until: "cyberpunk_age",
		lines: []string{"Rusting where it was dumped.", "Scrap, picked over by scavengers."}},
	FeatToxic: {key: "toxic_ground", name: "toxic ground", title: "Toxic ground", from: "cyberpunk_age", until: "cyberpunk_age",
		lines: []string{"Nothing grows here now.", "Poisoned ground, fenced off and forgotten."}},
	FeatCanal: {key: "toxic_canal", name: "toxic canal", title: "Toxic canal", from: "cyberpunk_age", until: "cyberpunk_age",
		lines: []string{"The water glows a sickly green.", "A canal, thick with runoff."}},
	FeatSkyRail: {key: "sky_rail", name: "sky rail", title: "Sky rail", from: "cyberpunk_age", until: "cyberpunk_age",
		lines: []string{"An elevated line, high over the street."}},
	FeatMaglevLine: {key: "maglev_skyway", name: "maglev skyway", title: "Maglev skyway", from: "fusion_age",
		lines: []string{"An elevated maglev line, humming with power."}},
	FeatReactor: {key: "reactor", name: "fusion reactor", title: "Fusion reactor", sym: SymReactor, from: "fusion_age",
		lines: []string{"Plasma held in a ring of magnets.", "A star in a bottle, powering the city."}},
	FeatConduit: {key: "conduit", name: "power conduit", title: "Power conduit", from: "fusion_age",
		lines: []string{"Plasma power, carried to the core.", "A conduit, pulsing with light."}},
	FeatLaunchTower: {key: "launch_tower", name: "launch tower", title: "Launch tower", sym: SymLaunchTower, from: "fusion_age",
		lines: []string{"A rocket on its pad, waiting.", "A launch gantry, lit for the night."}},
	FeatTether: {key: "tether", name: "space elevator", title: "Space elevator", from: "fusion_age",
		lines: []string{"A tether rising out of sight.", "Climbers ride it up out of the sky."}},
	FeatCooling: {key: "cooling", name: "cooling channel", title: "Cooling channel", from: "fusion_age",
		lines: []string{"Water cooling the reactors, glowing blue.", "A channel of clean, cold water."}},
}

var cityFeatureTable = buildCityFeatures()

func buildCityFeatures() [NumCityFeatures]CityFeatureInfo {
	idx := map[string]int{}
	for i, k := range config.AgeOrder() {
		idx[k] = i
	}
	age := func(k string, none int) int {
		if k == "" {
			return none
		}
		if i, ok := idx[k]; ok {
			return i
		}
		return -1
	}
	var out [NumCityFeatures]CityFeatureInfo
	out[FeatNone].From, out[FeatNone].Until = -1, -1
	for f := CityFeature(1); f < NumCityFeatures; f++ {
		d := cityFeatureDefs[f]
		out[f] = CityFeatureInfo{Key: d.key, Name: d.name, Title: d.title, Sym: d.sym,
			From: age(d.from, -1), Until: age(d.until, -1), Lines: d.lines}
		if d.until != "" && out[f].Until < 0 {
			out[f].From = -1 // an unknown age key: never stands (TestCityRoster fails on it)
		}
	}
	return out
}

// Info returns a feature's roster entry.
func (f CityFeature) Info() CityFeatureInfo {
	if f >= NumCityFeatures {
		return cityFeatureTable[FeatNone]
	}
	return cityFeatureTable[f]
}

// Built reports whether the age with index age, or an earlier one, has
// brought feature f (it may have given way since).
func Built(f CityFeature, age int) bool {
	i := f.Info()
	return i.From >= 0 && age >= i.From
}

// FeaturesAt lists the features standing in the age with index age.
func FeaturesAt(age int) []CityFeature {
	var out []CityFeature
	for f := CityFeature(1); f < NumCityFeatures; f++ {
		if cityFeatureTable[f].In(age) {
			out = append(out, f)
		}
	}
	return out
}
