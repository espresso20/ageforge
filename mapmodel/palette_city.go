package mapmodel

import (
	"github.com/espresso20/ageforge/config"
	"github.com/espresso20/ageforge/theme"
	"github.com/espresso20/ageforge/vibe"
)

// palette_city.go is the Earth arc's look, age by age, the Modern Age to the
// Fusion Age: the city eats the land (Modern, Information, Digital), then
// the megacity rises (Cyberpunk) and is powered (Fusion). Each age has a
// palette, a signature structure, a signature mover and one ambient effect,
// so every age reads at a glance; both map styles draw from this table, and
// the greenery fade and the smog are rules the tests hold them to.

// CityLook is one Earth-arc age's look.
type CityLook struct {
	Key  string // age key
	Age  int    // age index
	Name string // what the city has become
	// Palette is the age's colours, most prominent first.
	Palette []theme.CityHue
	// Structure is its signature structure, Mover its signature mover and
	// Ambient its ambient effect, in the words the docs use.
	Structure string
	Mover     Mover
	Ambient   string
	// Greenery is the share of the Modern Age's greenery the age keeps.
	Greenery float64
	// Smog is how thick the air is: 0 clear, 1 choking.
	Smog float64
	// Night is a city that reads as night by day: dark towers lit by their
	// own neon and plasma.
	Night bool
	// Rain is the age's ambient rain (the megacity's acid streaks).
	Rain bool

	// the entries Name, Structure and Ambient were written as, read again
	// while VibeMode is on
	nameE, structureE, ambientE vibe.Entry
}

type cityLookDef struct {
	key                      string
	name, structure, ambient vibe.Entry
	palette                  []theme.CityHue
	mover                    Mover
	greenery, smog           float64
	night, rain              bool
}

// cityLookDefs is the Earth arc, oldest first.
var cityLookDefs = []cityLookDef{
	{key: "modern_age", name: vibe.E("map.city_look.modern_age.name", "the glass city"), structure: vibe.E("map.city_look.modern_age.structure", "highways, a glass downtown and parks fenced into blocks"),
		mover: MoverCar, ambient: vibe.E("map.city_look.modern_age.ambient", "headlights streaming down the highways at night"), greenery: 1,
		palette: []theme.CityHue{theme.CityGlass, theme.CityConcrete, theme.CityAsphalt, theme.CityPark, theme.CityRoof}},
	{key: "information_age", name: vibe.E("map.city_look.information_age.name", "the wired city"), structure: vibe.E("map.city_look.information_age.structure", "server farms, fiber lines and satellite dishes"),
		mover: MoverNewsHeli, ambient: vibe.E("map.city_look.information_age.ambient", "screens flickering in the windows"), greenery: 0.5, smog: 0.25,
		palette: []theme.CityHue{theme.CityServer, theme.CityConcrete, theme.CityData, theme.CityScreen, theme.CitySmog}},
	{key: "digital_age", name: vibe.E("map.city_look.digital_age.name", "the paved city"), structure: vibe.E("map.city_look.digital_age.structure", "the walled reserve, the last green, in concrete and landfill"),
		mover: MoverDrone, ambient: vibe.E("map.city_look.digital_age.ambient", "the data glow pulsing under the servers"), greenery: 0.15, smog: 0.5,
		palette: []theme.CityHue{theme.CityConcreteDark, theme.CityLandfill, theme.CityGlow, theme.CitySmog, theme.CityReserve}},
	{key: "cyberpunk_age", name: vibe.E("map.city_look.cyberpunk_age.name", "the megacity"), structure: vibe.E("map.city_look.cyberpunk_age.structure", "megacorp towers carrying their initials, and arcologies"),
		mover: MoverSkyTrain, ambient: vibe.E("map.city_look.cyberpunk_age.ambient", "acid rain and flickering neon"), smog: 0.8, night: true, rain: true,
		palette: []theme.CityHue{theme.CityNight, theme.CityTower, theme.CityNeonMagenta, theme.CityNeonCyan, theme.CityNeonYellow, theme.CityToxic}},
	{key: "fusion_age", name: vibe.E("map.city_look.fusion_age.name", "the powered city"), structure: vibe.E("map.city_look.fusion_age.structure", "fusion reactors in their plasma rings, and the space elevator"),
		mover: MoverClimber, ambient: vibe.E("map.city_look.fusion_age.ambient", "plasma pulsing through the conduits"), smog: 0.3, night: true,
		palette: []theme.CityHue{theme.CityNight, theme.CityElectric, theme.CityPlasma, theme.CityConduit, theme.CityNeonCyan}},
}

var cityLookTable = buildCityLooks()

func buildCityLooks() []CityLook {
	idx := map[string]int{}
	for i, k := range config.AgeOrder() {
		idx[k] = i
	}
	out := make([]CityLook, 0, len(cityLookDefs))
	for _, d := range cityLookDefs {
		a, ok := idx[d.key]
		if !ok {
			continue // an unknown age key (TestCityLooks fails on it)
		}
		out = append(out, CityLook{Key: d.key, Age: a, Name: d.name.Plain(), Palette: d.palette, Structure: d.structure.Plain(),
			Mover: d.mover, Ambient: d.ambient.Plain(), Greenery: d.greenery, Smog: d.smog, Night: d.night, Rain: d.rain,
			nameE: d.name, structureE: d.structure, ambientE: d.ambient})
	}
	return out
}

// CityLooks lists the Earth arc's looks, oldest first.
func CityLooks() []CityLook {
	out := append([]CityLook(nil), cityLookTable...)
	for i := range out {
		out[i] = out[i].told()
	}
	return out
}

// told is the look with its words as they read now: itself while VibeMode is
// off, else read again from the entries the table was written as.
func (l CityLook) told() CityLook {
	if vibe.On() {
		l.Name, l.Structure, l.Ambient = l.nameE.String(), l.structureE.String(), l.ambientE.String()
	}
	return l
}

// CityLookAt is the look of the age with index age, if it is an Earth-arc
// age.
func CityLookAt(age int) (CityLook, bool) {
	for _, l := range cityLookTable {
		if l.Age == age {
			return l.told(), true
		}
	}
	return CityLook{}, false
}

// Greenery is the greenery fade, the share of the Modern Age's greenery
// (forest, grass, parks and gardens) the age with index age keeps: all of
// it up to the Modern Age, about half in the Information Age (parks and
// rooftop gardens), the walled reserve's 15% in the Digital Age, and none
// from the Cyberpunk Age on.
func Greenery(age int) float64 {
	if len(cityLookTable) == 0 || age < cityLookTable[0].Age {
		return 1
	}
	if l, ok := CityLookAt(age); ok {
		return l.Greenery
	}
	return 0 // past the Earth arc: the city is everywhere
}

// SmogLevel is how thick the air is in the age with index age: clear
// before the Information Age, thickening to the megacity's and thinning
// once fusion cleans it.
func SmogLevel(age int) float64 {
	if l, ok := CityLookAt(age); ok {
		return l.Smog
	}
	return 0
}
