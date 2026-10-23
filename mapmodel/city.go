package mapmodel

import (
	"github.com/espresso20/ageforge/config"

	"github.com/espresso20/ageforge/vibe"
)

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

	// def is the table entry these were written as, read again while VibeMode is on
	def *cityFeatureDef
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
	key         string
	name, title vibe.Entry
	sym         Sym
	from, until string // age keys; until "" stands to the end
	lines       []vibe.Entry
}

var cityFeatureDefs = [NumCityFeatures]cityFeatureDef{
	FeatSuburbs: {key: "suburbs", name: vibe.E("map.city.suburbs.name", "suburbs"), title: vibe.E("map.city.suburbs.title", "Suburbs"), sym: SymSuburb, from: "modern_age", until: "modern_age",
		lines: []vibe.Entry{vibe.E("map.city.suburbs.line.01", "Houses with gardens, a drive and a car."), vibe.E("map.city.suburbs.line.02", "Commuter houses, quiet by day.")}},
	FeatPark: {key: "park", name: vibe.E("map.city.park.name", "fenced park"), title: vibe.E("map.city.park.title", "Park"), sym: SymPark, from: "modern_age", until: "information_age",
		lines: []vibe.Entry{vibe.E("map.city.park.line.01", "Still green, fenced into its block."), vibe.E("map.city.park.line.02", "A park, railed in on every side.")}},
	FeatHighway: {key: "highway", name: vibe.E("map.city.highway.name", "highway"), title: vibe.E("map.city.highway.title", "Highway"), from: "modern_age", until: "digital_age",
		lines: []vibe.Entry{vibe.E("map.city.highway.line.01", "Four lanes each way, out of town."), vibe.E("map.city.highway.line.02", "The highway, never quite empty.")}},
	FeatGlassTower: {key: "glass_tower", name: vibe.E("map.city.glass_tower.name", "glass tower"), title: vibe.E("map.city.glass_tower.title", "Glass tower"), sym: SymGlassTower,
		from: "modern_age", until: "digital_age",
		lines: []vibe.Entry{vibe.E("map.city.glass_tower.line.01", "Glass and steel, downtown."), vibe.E("map.city.glass_tower.line.02", "Offices, forty floors of them.")}},
	FeatServerFarm: {key: "server_farm", name: vibe.E("map.city.server_farm.name", "server farm"), title: vibe.E("map.city.server_farm.title", "Server farm"), sym: SymServerFarm,
		from: "information_age", until: "digital_age",
		lines: []vibe.Entry{vibe.E("map.city.server_farm.line.01", "Racks humming behind blank walls."), vibe.E("map.city.server_farm.line.02", "Cooling fans, day and night.")}},
	FeatDish: {key: "dish", name: vibe.E("map.city.dish.name", "satellite dish"), title: vibe.E("map.city.dish.title", "Satellite dish"), sym: SymDish,
		from: "information_age", until: "digital_age",
		lines: []vibe.Entry{vibe.E("map.city.dish.line.01", "Listening to the sky."), vibe.E("map.city.dish.line.02", "A dish, pointed at a satellite.")}},
	FeatFiber: {key: "fiber", name: vibe.E("map.city.fiber.name", "fiber line"), title: vibe.E("map.city.fiber.title", "Fiber line"), from: "information_age", until: "digital_age",
		lines: []vibe.Entry{vibe.E("map.city.fiber.line.01", "Fiber under the street, carrying data."), vibe.E("map.city.fiber.line.02", "Cable trenches, freshly dug.")}},
	FeatGarden: {key: "garden", name: vibe.E("map.city.garden.name", "rooftop garden"), title: vibe.E("map.city.garden.title", "Rooftop garden"), sym: SymGarden,
		from: "information_age", until: "information_age",
		lines: []vibe.Entry{vibe.E("map.city.garden.line.01", "What is left of the green, up on a roof."), vibe.E("map.city.garden.line.02", "A garden on the roof, among the vents.")}},
	FeatOffice: {key: "office_block", name: vibe.E("map.city.office_block.name", "office block"), title: vibe.E("map.city.office_block.title", "Office block"), from: "information_age", until: "digital_age",
		lines: []vibe.Entry{vibe.E("map.city.office_block.line.01", "Screens on every floor."), vibe.E("map.city.office_block.line.02", "Desks, glass and air conditioning.")}},
	FeatScrub: {key: "scrub", name: vibe.E("map.city.scrub.name", "dry ground"), title: vibe.E("map.city.scrub.title", "Dry ground"), from: "information_age", until: "information_age",
		lines: []vibe.Entry{vibe.E("map.city.scrub.line.01", "Scrub, browned by the smog."), vibe.E("map.city.scrub.line.02", "The grass gone dry past the city.")}},
	FeatSmog: {key: "smog", name: vibe.E("map.city.smog.name", "smog"), title: vibe.E("map.city.smog.title", "Smog"), from: "information_age", until: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.smog.line.01", "The air turning brown at the edges."), vibe.E("map.city.smog.line.02", "Haze over the outskirts.")}},
	FeatConcrete: {key: "concrete", name: vibe.E("map.city.concrete.name", "concrete"), title: vibe.E("map.city.concrete.title", "Concrete"), from: "digital_age", until: "digital_age",
		lines: []vibe.Entry{vibe.E("map.city.concrete.line.01", "Paved over, where the fields were."), vibe.E("map.city.concrete.line.02", "Car parks and loading bays.")}},
	FeatLandfill: {key: "landfill", name: vibe.E("map.city.landfill.name", "landfill"), title: vibe.E("map.city.landfill.title", "Landfill"), sym: SymLandfill, from: "digital_age", until: "digital_age",
		lines: []vibe.Entry{vibe.E("map.city.landfill.line.01", "The city's rubbish, piled high."), vibe.E("map.city.landfill.line.02", "Old screens and packaging, a hill of it.")}},
	FeatReserve: {key: "reserve", name: vibe.E("map.city.reserve.name", "the reserve"), title: vibe.E("map.city.reserve.title", "The Reserve"), sym: SymPark, from: "digital_age", until: "digital_age",
		lines: []vibe.Entry{vibe.E("map.city.reserve.line.01", "The last green, behind a wall."), vibe.E("map.city.reserve.line.02", "Old trees, kept like a museum.")}},
	FeatDataGlow: {key: "data_glow", name: vibe.E("map.city.data_glow.name", "data glow"), title: vibe.E("map.city.data_glow.title", "Data glow"), from: "digital_age", until: "digital_age",
		lines: []vibe.Entry{vibe.E("map.city.data_glow.line.01", "Servers warm enough to light the street."), vibe.E("map.city.data_glow.line.02", "The glow of a data hall.")}},
	FeatMegablock: {key: "megablock", name: vibe.E("map.city.megablock.name", "megablock"), title: vibe.E("map.city.megablock.title", "Megablock"), from: "cyberpunk_age", until: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.megablock.line.01", "Towers on towers, lit all night."), vibe.E("map.city.megablock.line.02", "A hundred floors, and no one knows the neighbors.")}},
	FeatAlley: {key: "alley", name: vibe.E("map.city.alley.name", "alley"), title: vibe.E("map.city.alley.title", "Alley"), from: "cyberpunk_age", until: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.alley.line.01", "Dark, wet and crowded."), vibe.E("map.city.alley.line.02", "Steam, noodle stalls and wires overhead.")}},
	FeatCorpTower: {key: "corp_tower", name: vibe.E("map.city.corp_tower.name", "megacorp tower"), title: vibe.E("map.city.corp_tower.title", "Megacorp tower"), from: "cyberpunk_age", until: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.corp_tower.line.01", "Its initials in neon, a block high."), vibe.E("map.city.corp_tower.line.02", "A megacorp's tower, watching the street.")}},
	FeatArcology: {key: "arcology", name: vibe.E("map.city.arcology.name", "arcology"), title: vibe.E("map.city.arcology.title", "Arcology"), sym: SymArcology, from: "cyberpunk_age", until: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.arcology.line.01", "A city in one building."), vibe.E("map.city.arcology.line.02", "Terraces stacked to the clouds.")}},
	FeatNeonSign: {key: "neon_sign", name: vibe.E("map.city.neon_sign.name", "neon sign"), title: vibe.E("map.city.neon_sign.title", "Neon sign"), sym: SymNeonSign, from: "cyberpunk_age", until: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.neon_sign.line.01", "Buzzing, and flickering now and then."), vibe.E("map.city.neon_sign.line.02", "A holo-ad, selling something.")}},
	FeatVent: {key: "vent", name: vibe.E("map.city.vent.name", "steam vent"), title: vibe.E("map.city.vent.title", "Steam vent"), sym: SymVent, from: "cyberpunk_age", until: "cyberpunk_age",
		lines: []vibe.Entry{vibe.E("map.city.vent.line.01", "Steam from the undercity."), vibe.E("map.city.vent.line.02", "A vent, breathing hot air.")}},
	FeatScrap: {key: "scrap", name: vibe.E("map.city.scrap.name", "scrap heap"), title: vibe.E("map.city.scrap.title", "Scrap heap"), sym: SymScrap, from: "cyberpunk_age", until: "cyberpunk_age",
		lines: []vibe.Entry{vibe.E("map.city.scrap.line.01", "Rusting where it was dumped."), vibe.E("map.city.scrap.line.02", "Scrap, picked over by scavengers.")}},
	FeatToxic: {key: "toxic_ground", name: vibe.E("map.city.toxic_ground.name", "toxic ground"), title: vibe.E("map.city.toxic_ground.title", "Toxic ground"), from: "cyberpunk_age", until: "cyberpunk_age",
		lines: []vibe.Entry{vibe.E("map.city.toxic_ground.line.01", "Nothing grows here now."), vibe.E("map.city.toxic_ground.line.02", "Poisoned ground, fenced off and forgotten.")}},
	FeatCanal: {key: "toxic_canal", name: vibe.E("map.city.toxic_canal.name", "toxic canal"), title: vibe.E("map.city.toxic_canal.title", "Toxic canal"), from: "cyberpunk_age", until: "cyberpunk_age",
		lines: []vibe.Entry{vibe.E("map.city.toxic_canal.line.01", "The water glows a sickly green."), vibe.E("map.city.toxic_canal.line.02", "A canal, thick with runoff.")}},
	FeatSkyRail: {key: "sky_rail", name: vibe.E("map.city.sky_rail.name", "sky rail"), title: vibe.E("map.city.sky_rail.title", "Sky rail"), from: "cyberpunk_age", until: "cyberpunk_age",
		lines: []vibe.Entry{vibe.E("map.city.sky_rail.line.01", "An elevated line, high over the street.")}},
	FeatMaglevLine: {key: "maglev_skyway", name: vibe.E("map.city.maglev_skyway.name", "maglev skyway"), title: vibe.E("map.city.maglev_skyway.title", "Maglev skyway"), from: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.maglev_skyway.line.01", "An elevated maglev line, humming with power.")}},
	FeatReactor: {key: "reactor", name: vibe.E("map.city.reactor.name", "fusion reactor"), title: vibe.E("map.city.reactor.title", "Fusion reactor"), sym: SymReactor, from: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.reactor.line.01", "Plasma held in a ring of magnets."), vibe.E("map.city.reactor.line.02", "A star in a bottle, powering the city.")}},
	FeatConduit: {key: "conduit", name: vibe.E("map.city.conduit.name", "power conduit"), title: vibe.E("map.city.conduit.title", "Power conduit"), from: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.conduit.line.01", "Plasma power, carried to the core."), vibe.E("map.city.conduit.line.02", "A conduit, pulsing with light.")}},
	FeatLaunchTower: {key: "launch_tower", name: vibe.E("map.city.launch_tower.name", "launch tower"), title: vibe.E("map.city.launch_tower.title", "Launch tower"), sym: SymLaunchTower, from: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.launch_tower.line.01", "A rocket on its pad, waiting."), vibe.E("map.city.launch_tower.line.02", "A launch gantry, lit for the night.")}},
	FeatTether: {key: "tether", name: vibe.E("map.city.tether.name", "space elevator"), title: vibe.E("map.city.tether.title", "Space elevator"), from: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.tether.line.01", "A tether rising out of sight."), vibe.E("map.city.tether.line.02", "Climbers ride it up out of the sky.")}},
	FeatCooling: {key: "cooling", name: vibe.E("map.city.cooling.name", "cooling channel"), title: vibe.E("map.city.cooling.title", "Cooling channel"), from: "fusion_age",
		lines: []vibe.Entry{vibe.E("map.city.cooling.line.01", "Water cooling the reactors, glowing blue."), vibe.E("map.city.cooling.line.02", "A channel of clean, cold water.")}},
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
		d := &cityFeatureDefs[f]
		out[f] = CityFeatureInfo{Key: d.key, Name: d.name.Plain(), Title: d.title.Plain(), Sym: d.sym,
			From: age(d.from, -1), Until: age(d.until, -1), Lines: plainLines(d.lines), def: d}
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
	in := cityFeatureTable[f]
	if vibe.On() && in.def != nil { // (FeatNone has no entry)
		in.Name, in.Title, in.Lines = in.def.name.String(), in.def.title.String(), toldLines(in.def.lines)
	}
	return in
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
