package mapgen

// Sol. The renderers draw the solar system from SolSpec, a literal: eight
// planets with their real moon counts, the belt, the dwarf planet and a comet.

// StarSpec is a star: three hues from the heart out, and its radius at the
// full zoom in 16ths of a pixel.
type StarSpec struct {
	Core, Mid, Rim Hue
	Size           int
}

// PlanetSpec is one planet. Sizes are in 16ths of a pixel at the full zoom.
type PlanetSpec struct {
	Size     int   // radius
	Close    int   // radius of its close-up disc, percent of the home world's
	Orbit    int   // distance from the star, percent of the system's reach
	Moons    int   // how many
	Base     Paint // the face, without bands
	Bands    []Paint
	Blotch   Paint // weight 0: none; else the odd fleck on the face
	Cap      bool  // a polar cap
	Spot     bool  // a great storm
	Ring     bool
	World    bool // seas, land and cloud
	Home     bool // the home world: the town's, with its air
	Outpost  int  // a station by it in the squeezed views: 0 none, 1 a hull pixel, 2 a glyph
	Stations int  // stations in its close-up
}

// SystemSpec is a whole system.
type SystemSpec struct {
	Star    StarSpec
	Planets []PlanetSpec // inner to outer
	Belts   []int        // orbits of the asteroid belts
	Dwarf   bool         // a dwarf planet far out
	Comet   bool
}

// ThingOf is the thing planet i is drawn as: the home world is "home", the
// others "planet" numbered by Ref from 1.
func (sp *SystemSpec) ThingOf(i int) Thing {
	if sp.Planets[i].Home {
		return Thing{Kind: KindTerrain, Key: "home"}
	}
	return Thing{Kind: KindTerrain, Key: "planet", Ref: i + 1}
}

// The sun is a terrain thing of its own; its halo is the same thing with a
// Ref, which is not a body.
var (
	sunThing     = Thing{Kind: KindTerrain, Key: "sun"}
	sunHaloThing = Thing{Kind: KindTerrain, Key: "sun", Ref: 1}
)

// SolSpec is the home system as the sky has it: eight planets with their real
// moon counts, Saturn ringed, a station by the home world and one at Mars.
var SolSpec = SystemSpec{
	Star: StarSpec{Core: HWhite, Mid: HSun, Rim: HCore, Size: 88},
	Planets: []PlanetSpec{
		{Size: 22, Close: 55, Orbit: 18, Base: P(HRock, 255)},
		{Size: 32, Close: 90, Orbit: 28, Bands: []Paint{P(HSun, 250), P(HSun, 205), P(HSun, 250), P(HSun, 205), P(HSun, 250), P(HSun, 205)}},
		{Size: 35, Close: 100, Orbit: 39, Moons: 1, World: true, Home: true, Outpost: 1, Stations: 2},
		{Size: 27, Close: 70, Orbit: 50, Moons: 2, Base: P(HRed, 255), Blotch: P(HRed, 200), Cap: true, Outpost: 2},
		{Size: 68, Close: 110, Orbit: 68, Moons: 4, Spot: true, Bands: []Paint{P(HRock, 235), P(HAmber, 255), P(HSand, 255), P(HRock, 215), P(HAmber, 245), P(HSand, 240), P(HRock, 225), P(HAmber, 255), P(HRock, 235)}},
		{Size: 58, Close: 85, Orbit: 79, Moons: 5, Ring: true, Bands: []Paint{P(HSand, 235), P(HGold, 250), P(HSand, 255), P(HGold, 235), P(HSand, 250), P(HGold, 255), P(HSand, 240), P(HGold, 240), P(HSand, 230)}},
		{Size: 42, Close: 90, Orbit: 89, Moons: 4, Base: P(HCool, 255)},
		{Size: 42, Close: 90, Orbit: 98, Moons: 2, Bands: []Paint{P(HPanel, 255), P(HCyan, 215), P(HPanel, 255), P(HPanel, 255), P(HPanel, 255), P(HCyan, 215), P(HPanel, 255), P(HPanel, 255), P(HPanel, 255)}},
	},
	Belts: []int{59},
	Dwarf: true,
	Comet: true,
}
