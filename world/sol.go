package world

// Sol. The home system as the sky has it: eight planets with their real moon
// counts, the belt, the dwarf planet and a comet. A literal, moved here from
// mapgen; the painters read it through mapgen/tone.go.

// StarSpec is a star: three tones from the heart out, and its radius at the
// full zoom in 16ths of a pixel.
type StarSpec struct {
	Core, Mid, Rim Tone
	Size           int
}

// PlanetSpec is one planet. Sizes are in 16ths of a pixel at the full zoom.
type PlanetSpec struct {
	Size     int   // radius
	Close    int   // radius of its close-up disc, percent of the home world's
	Orbit    int   // distance from the star, percent of the system's reach
	Moons    int   // how many
	Base     Shade // the face, without bands
	Bands    []Shade
	Blotch   Shade // weight 0: none; else the odd fleck on the face
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

// OrbitPeriod is the ticks planet p takes to go round once, on the animation
// clock the painters share: the angle moves 16 of 4096 every max(10, orbit^2/37)
// ticks.
func (p PlanetSpec) OrbitPeriod() int { return 256 * max(10, p.Orbit*p.Orbit/37) }

// OrbitPhase is where planet i starts on its orbit, in 4096ths of a turn, for
// a seed.
func OrbitPhase(seed int64, i int) int { return ((int(seed%1000) + i*382) % 1000) * Turn / 1000 }

// SolSpec is the home system as the sky has it: eight planets with their real
// moon counts, Saturn ringed, a station by the home world and one at Mars.
var SolSpec = SystemSpec{
	Star: StarSpec{Core: ToneWhite, Mid: ToneSun, Rim: ToneCore, Size: 88},
	Planets: []PlanetSpec{
		{Size: 22, Close: 55, Orbit: 18, Base: S(ToneRock, 255)},
		{Size: 32, Close: 90, Orbit: 28, Bands: []Shade{S(ToneSun, 250), S(ToneSun, 205), S(ToneSun, 250), S(ToneSun, 205), S(ToneSun, 250), S(ToneSun, 205)}},
		{Size: 35, Close: 100, Orbit: 39, Moons: 1, World: true, Home: true, Outpost: 1, Stations: 2},
		{Size: 27, Close: 70, Orbit: 50, Moons: 2, Base: S(ToneRed, 255), Blotch: S(ToneRed, 200), Cap: true, Outpost: 2},
		{Size: 68, Close: 110, Orbit: 68, Moons: 4, Spot: true, Bands: []Shade{S(ToneRock, 235), S(ToneAmber, 255), S(ToneSand, 255), S(ToneRock, 215), S(ToneAmber, 245), S(ToneSand, 240), S(ToneRock, 225), S(ToneAmber, 255), S(ToneRock, 235)}},
		{Size: 58, Close: 85, Orbit: 79, Moons: 5, Ring: true, Bands: []Shade{S(ToneSand, 235), S(ToneGold, 250), S(ToneSand, 255), S(ToneGold, 235), S(ToneSand, 250), S(ToneGold, 255), S(ToneSand, 240), S(ToneGold, 240), S(ToneSand, 230)}},
		{Size: 42, Close: 90, Orbit: 89, Moons: 4, Base: S(ToneCool, 255)},
		{Size: 42, Close: 90, Orbit: 98, Moons: 2, Bands: []Shade{S(TonePanel, 255), S(ToneCyan, 215), S(TonePanel, 255), S(TonePanel, 255), S(TonePanel, 255), S(ToneCyan, 215), S(TonePanel, 255), S(TonePanel, 255), S(TonePanel, 255)}},
	},
	Belts: []int{59},
	Dwarf: true,
	Comet: true,
}
